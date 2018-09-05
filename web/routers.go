package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/tacusci/berrycms/db"
	"github.com/tacusci/berrycms/util"
	"github.com/tacusci/logging"
)

//MutableRouter is a mutex lock for the mux router
type MutableRouter struct {
	Server *http.Server
	mu     sync.Mutex
	Root   *mux.Router
	dw     *util.RecursiveDirWatch
}

//Swap takes a new mux router, locks accessing for old one, replaces it and then unlocks, keeps existing connections
func (mr *MutableRouter) Swap(root *mux.Router) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.Root = root
	mr.Server.Handler = mr.Root
}

//Reload map all admin/default page routes and load saved page routes from DB
func (mr *MutableRouter) Reload() {

	r := mux.NewRouter()

	logging.Debug("Mapping default admin routes...")

	for _, handler := range GetDefaultHandlers(mr) {
		if handler.HandlesGet() {
			logging.Debug(fmt.Sprintf("Mapping default GET route %s", handler.Route()))
			r.HandleFunc(handler.Route(), handler.Get).Methods("GET")
		}

		if handler.HandlesPost() {
			logging.Debug(fmt.Sprintf("Mapping default POST route %s", handler.Route()))
			r.HandleFunc(handler.Route(), handler.Post).Methods("POST")
		}
	}

	mr.mapSavedPageRoutes(r)
	mr.mapStaticDir(r, "static")
	go mr.monitorStatic("static")

	amw := AuthMiddleware{Router: mr}
	r.Use(amw.Middleware)

	mr.Swap(r)
}

func (mr *MutableRouter) mapSavedPageRoutes(r *mux.Router) {
	savedPageHandler := &SavedPageHandler{Router: mr}

	pt := db.PagesTable{}
	rows, err := pt.Select(db.Conn, "route", "")
	defer rows.Close()
	if err != nil {
		logging.Error(err.Error())
		return
	}
	for rows.Next() {
		p := db.Page{}
		rows.Scan(&p.Route)
		logging.Debug(fmt.Sprintf("Mapping database page route %s", p.Route))
		r.HandleFunc(p.Route, savedPageHandler.Get)
	}
}

func (mr *MutableRouter) mapStaticDir(r *mux.Router, sd string) {
	fs, err := ioutil.ReadDir(sd)
	if err != nil {
		logging.Error("Unable to find static folder...")
		return
	}
	for _, f := range fs {
		pathPrefixLocation := fmt.Sprintf("%s%s%s", string(os.PathSeparator), f.Name(), string(os.PathSeparator))
		pathPrefixAddress := fmt.Sprintf("/%s/", f.Name())
		logging.Debug(fmt.Sprintf("Serving dir (%s)'s files...", f.Name()))
		r.PathPrefix(pathPrefixAddress).Handler(http.StripPrefix(pathPrefixAddress, http.FileServer(http.Dir(sd+pathPrefixLocation))))
	}
}

func (mr *MutableRouter) monitorStatic(sd string) {
	mr.dw = &util.RecursiveDirWatch{Change: make(chan bool), Stop: make(chan bool)}
	go mr.dw.WatchDir(sd)
	for {
		if <-mr.dw.Change {
			logging.Debug("Change detected in static dir...")
			break
		}
	}
	mr.dw.Stop <- true
	mr.Reload()
}

//AuthMiddleware authentication struct with auth helper functions
type AuthMiddleware struct {
	Router *MutableRouter
}

//Middleware attaches http handler to middleware
func (amw *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if amw.HasPermissionsForRoute(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Access denied", http.StatusForbidden)
		}
	})
}

//HasPermissionsForRoute checks that requesting client has permissions to access the requested URI
func (amw *AuthMiddleware) HasPermissionsForRoute(r *http.Request) bool {
	var routeIsProtected bool

	routeIsProtected = strings.HasPrefix(r.RequestURI, "/admin")

	if !routeIsProtected {
		pt := db.PagesTable{}
		page, err := pt.SelectByRoute(db.Conn, r.RequestURI)
		if err == nil {
			routeIsProtected = page.Roleprotected
		}
	}

	if routeIsProtected {
		return amw.IsLoggedIn(r)
	}
	return true
}

//IsLoggedIn checks if the requesting client is currently logged in
func (amw *AuthMiddleware) IsLoggedIn(r *http.Request) bool {
	var isLoggedIn bool

	authSessionStore, err := sessionsstore.Get(r, "auth")
	if err == nil {
		if authSessionUUID := authSessionStore.Values["sessionuuid"]; authSessionUUID != nil {
			authSessionsTable := db.AuthSessionsTable{}
			authSession, err := authSessionsTable.SelectBySessionUUID(db.Conn, authSessionUUID.(string))
			if err == nil {
				if len(authSession.UserUUID) > 0 {
					isLoggedIn = true
					authSession.LastActiveDateTime = time.Now().Unix()
					authSessionsTable.Update(db.Conn, authSession)
				} else {
					authSessionsTable.DeleteBySessionUUID(db.Conn, authSessionUUID.(string))
					authSessionStore.Options.MaxAge = -1
				}
			} else {
				errString := err.Error()
				if errString != "sql: no rows in result set" {
					logging.Error(err.Error())
				}
			}
		}
	} else {
		logging.Debug(fmt.Sprintf("Error trying to read existing session \"auth\" -> %s", err.Error()))
	}
	return isLoggedIn
}

//LoggedInUser get user of existing web session
func (amw *AuthMiddleware) LoggedInUser(r *http.Request) (db.User, error) {
	authSessionStore, err := sessionsstore.Get(r, "auth")
	if err == nil {
		authSessionsTable := db.AuthSessionsTable{}
		if authSessionUUID := authSessionStore.Values["sessionuuid"]; authSessionUUID != nil {
			authSession, err := authSessionsTable.SelectBySessionUUID(db.Conn, authSessionUUID.(string))
			if err == nil {
				if len(authSession.UserUUID) > 0 {
					ut := db.UsersTable{}
					loggedInUser, err := ut.SelectByUUID(db.Conn, authSession.UserUUID)
					if err != nil {
						return db.User{}, err
					}
					return loggedInUser, nil
				}
			}
		}
	}
	return db.User{}, err
}

//Error writes HTTP error message to web response and add error message to log
func Error(w http.ResponseWriter, err error) {
	logging.Error(err.Error())
	http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
}
