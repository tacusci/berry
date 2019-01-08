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

package robots

import (
	"bytes"
	"fmt"

	"github.com/coocood/freecache"
	"github.com/tacusci/berrycms/db"
)

var RobotsCache *freecache.Cache

func Generate() error {

	sb := bytes.Buffer{}

	sb.WriteString("User-agent: *\n")
	//block indexing admin pages
	//NOTE: if the admin pages URI has been hidden, we're deliberately omitting this from robots.txt
	sb.WriteString("Disallow: /admin\n")

	pt := db.PagesTable{}
	rows, err := pt.Select(db.Conn, "route", "roleprotected = '1'")

	if err != nil {
		return err
	}

	var pageRouteToDisallow string

	for rows.Next() {
		err := rows.Scan(&pageRouteToDisallow)
		if err != nil {
			return err
		}

		_, err = sb.WriteString(fmt.Sprintf("Disallow: %s\n", pageRouteToDisallow))
		if err != nil {
			return err
		}
	}

	robotsToCache := sb.Bytes()
	Cache(&robotsToCache)

	return nil
}

func Cache(val *[]byte) error {
	RobotsCache = freecache.NewCache(len(*val))
	key := []byte("robots")
	err := RobotsCache.Set(key, *val, 0)
	if err != nil {
		return err
	}
	return nil
}
