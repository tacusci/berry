package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/tacusci/berrycms/db"
	"github.com/tacusci/logging"
)

type SavedPageHandler struct {
	Router *MutableRouter
	route  string
}

func (sph *SavedPageHandler) Get(w http.ResponseWriter, r *http.Request) {
	pt := db.PagesTable{}
	//JUST FOR LIVE/HOT ROUTE REMAPPING TESTING
	if r.RequestURI == "/addnew" {
		pt.Insert(db.Conn, db.Page{
			CreatedDateTime: time.Now().Unix(),
			Title:           "Carbon",
			Route:           "/carbonite",
			Content:         "<h2>Carbonite</h2>",
			Roleprotected:   true,
		})
		sph.Router.Reload()
	}
	rows, err := pt.Select(db.Conn, "content", fmt.Sprintf("route = '%s'", r.RequestURI))
	defer rows.Close()
	if err != nil {
		logging.Error(err.Error())
		w.Write([]byte("<h1>500 Server Error</h1>"))
		return
	}
	p := &db.Page{}
	for rows.Next() {
		rows.Scan(&p.Content)
	}

	ctx := plush.NewContext()
	Render(w, p, ctx)
}

func (sph *SavedPageHandler) Post(w http.ResponseWriter, r *http.Request) {}

func (sph *SavedPageHandler) Route() string { return sph.route }

func (sph *SavedPageHandler) HandlesGet() bool  { return true }
func (sph *SavedPageHandler) HandlesPost() bool { return false }
