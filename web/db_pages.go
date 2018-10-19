// Copyright (c) 2018, tacusci ltd
//
// Licensed under the GNU GENERAL PUBLIC LICENSE Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.gnu.org/licenses/gpl-3.0.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/tacusci/berrycms/db"
)

//SavedPageHandler handler to contain pointer to core router and the URI string
type SavedPageHandler struct {
	Router *MutableRouter
	route  string
}

//Get handles get requests to URI
func (sph *SavedPageHandler) Get(w http.ResponseWriter, r *http.Request) {
	pt := db.PagesTable{}
	//JUST FOR LIVE/HOT ROUTE REMAPPING TESTING
	if r.RequestURI == "/addnew" {
		for i := 0; i < 51; i++ {
			pt.Insert(db.Conn, db.Page{
				CreatedDateTime: time.Now().Unix(),
				Title:           fmt.Sprintf("Carbon %d", i),
				Route:           fmt.Sprintf("/carbonite-%d", i),
				Content:         fmt.Sprintf("<h2>Carbonite %d</h2>", i),
				Roleprotected:   true,
			})
		}
		sph.Router.Reload()
	}
	rows, err := pt.Select(db.Conn, "content, route", fmt.Sprintf("route = '%s'", r.RequestURI))
	if err != nil {
		Error(w, err)
		return
	}
	defer rows.Close()
	p := &db.Page{}
	for rows.Next() {
		rows.Scan(&p.Content, &p.Route)
	}

	ctx := plush.NewContext()
	ctx.Set("pagecontent", template.HTML(p.Content))
	Render(w, r, p, ctx)
}

//Post handles post requests to URI
func (sph *SavedPageHandler) Post(w http.ResponseWriter, r *http.Request) {}

//Route get URI route for handler
func (sph *SavedPageHandler) Route() string { return sph.route }

//HandlesGet retrieve whether this handler handles get requests
func (sph *SavedPageHandler) HandlesGet() bool { return true }

//HandlesPost retrieve whether this handler handles post requests
func (sph *SavedPageHandler) HandlesPost() bool { return false }
