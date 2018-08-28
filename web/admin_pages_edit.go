package web

import (
	"fmt"
	"net/http"

	"github.com/tacusci/logging"

	"github.com/gorilla/mux"
	"github.com/tacusci/berrycms/db"

	"github.com/gobuffalo/plush"
)

//AdminPagesHandler contains response functions for pages admin page
type AdminPagesEditHandler struct {
	Router *MutableRouter
	route  string
}

//Get takes the web request and writes response to session
func (apeh *AdminPagesEditHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pt := db.PagesTable{}
	pageToEdit, err := pt.SelectByUUID(db.Conn, vars["uuid"])
	if err != nil {
		logging.Error(err.Error())
		w.Write([]byte("Page to edit not found"))
		return
	}
	pctx := plush.NewContext()
	pctx.Set("title", fmt.Sprintf("Edit Page - %s", pageToEdit.Title))
	pctx.Set("quillenabled", true)
	RenderDefault(w, "admin.pages.edit.html", pctx)
}

func (apeh *AdminPagesEditHandler) Post(w http.ResponseWriter, r *http.Request) {}

func (apeh *AdminPagesEditHandler) Route() string { return apeh.route }

func (apeh *AdminPagesEditHandler) HandlesGet() bool  { return true }
func (apeh *AdminPagesEditHandler) HandlesPost() bool { return false }
