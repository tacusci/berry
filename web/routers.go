package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gobuffalo/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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
	store  *sessions.CookieStore
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

	if mr.store == nil {
		newUUID, err := uuid.NewV4()
		if err != nil {
			logging.ErrorAndExit(err.Error())
		}
		mr.store = sessions.NewCookieStore(newUUID.Bytes())
	}

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

	amw := authMiddleware{Router: mr}
	r.Use(amw.Middleware)

	mr.Swap(r)
}

func (mr *MutableRouter) mapSavedPageRoutes(r *mux.Router) {
	savedPageHandler := &SavedPageHandler{Router: mr}

	pt := db.PagesTable{}
	rows, err := pt.Select(db.Conn, "route", "")
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

type authMiddleware struct {
	Router *MutableRouter
}

func (amw *authMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if amw.HasPermissions(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Access denied", http.StatusForbidden)
		}
	})
}

func (amw *authMiddleware) HasPermissions(r *http.Request) bool {
	var routeIsProtected bool

	routeIsProtected = strings.HasPrefix(r.RequestURI, "/admin")

	if !routeIsProtected {
		pt := db.PagesTable{}
		page, err := pt.SelectByRoute(db.Conn, r.RequestURI)
		if err == nil {
			routeIsProtected = page.Roleprotected
		}
	}

	var isLoggedIn bool

	authSession, err := amw.Router.store.Get(r, "auth")
	if err == nil {
		if authSession.Values["isloggedin"] != nil {
			isLoggedIn = authSession.Values["isloggedin"].(bool)
		}
	}

	if routeIsProtected {
		return isLoggedIn
	} else {
		return true
	}
}