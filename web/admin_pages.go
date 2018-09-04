package web

import (
	"net/http"

	"github.com/gobuffalo/plush"
	"github.com/tacusci/berrycms/db"
	"github.com/tacusci/logging"
)

//AdminPagesHandler handler to contain pointer to core router and the URI string
type AdminPagesHandler struct {
	Router *MutableRouter
	route  string
}

//Get handles get requests to URI
func (aph *AdminPagesHandler) Get(w http.ResponseWriter, r *http.Request) {
	pages := make([]db.Page, 0)

	pt := db.PagesTable{}
	rows, err := pt.Select(db.Conn, "createddatetime, uuid, title, route", "")
	defer rows.Close()

	if err != nil {
		logging.ErrorAndExit(err.Error())
	}

	for rows.Next() {
		p := db.Page{}
		rows.Scan(&p.CreatedDateTime, &p.UUID, &p.Title, &p.Route)
		pages = append(pages, p)
	}

	pctx := plush.NewContext()
	pctx.Set("unixtostring", UnixToTimeString)
	pctx.Set("title", "Pages")
	pctx.Set("quillenabled", false)
	pctx.Set("pages", pages)

	RenderDefault(w, "admin.pages.html", pctx)
}

//Post handles post requests to URI
func (aph *AdminPagesHandler) Post(w http.ResponseWriter, r *http.Request) {}

//Route get URI route for handler
func (aph *AdminPagesHandler) Route() string { return aph.route }

//HandlesGet retrieve whether this handler handles get requests
func (aph *AdminPagesHandler) HandlesGet() bool { return true }

//HandlesPost retrieve whether this handler handles post requests
func (aph *AdminPagesHandler) HandlesPost() bool { return false }
