// Copyright (c) 2019 tacusci ltd
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
	"strings"

	"github.com/tacusci/berrycms/db"
)

var cache *bytes.Buffer

func Add(val *[]byte) error {
	*val = append(*val, []byte("\n")...)
	_, err := cache.Write(*val)
	if err != nil {
		return err
	}
	return nil
}

func Del(val *[]byte) error {
	existingVal := cache.Bytes()
	Reset()
	cache.Write([]byte(strings.Replace(string(existingVal), string(*val), "", -1)))
	return nil
}

func Generate() error {
	Reset()
	cache.WriteString("User-agent: *\n")
	//block indexing admin pages
	//NOTE: if the admin pages URI has been hidden, we're deliberately omitting this from robots.txt
	cache.WriteString("Disallow: /admin\n")

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

		_, err = cache.WriteString(fmt.Sprintf("Disallow: %s\n", pageRouteToDisallow))
		if err != nil {
			return err
		}
	}

	return nil
}

func Cache() *bytes.Buffer {
	return cache
}

func Reset() {
	if cache == nil {
		cache = &bytes.Buffer{}
	}
	cache.Reset()
}
