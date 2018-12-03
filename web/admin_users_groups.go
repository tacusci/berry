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
	"net/http"

	"github.com/gobuffalo/plush"
	"github.com/tacusci/logging"

	"github.com/tacusci/berrycms/db"
)

type AdminUserGroupsHandler struct {
	Router *MutableRouter
	route  string
}

func (ugh *AdminUserGroupsHandler) Get(w http.ResponseWriter, r *http.Request) {
	groups := make([]db.Group, 0)
	groupMemberships := make([]db.GroupMembership, 0)
	users := make([]db.User, 0)

	groupTable := db.GroupTable{}
	rows, err := groupTable.Select(db.Conn, "createddatetime, uuid, title", "")

	if err != nil {
		logging.Error(err.Error())
	}

	for rows.Next() {
		group := db.Group{}
		err := rows.Scan(&group.CreatedDateTime, &group.UUID, &group.Title)

		if err != nil {
			logging.Error(err.Error())
			continue
		}

		groups = append(groups, group)

		groupMembershipTable := db.GroupMembershipTable{}
		rows, err := groupMembershipTable.Select(db.Conn, "createddatetime, groupuuid, useruuid", fmt.Sprintf("groupuuid = '%s'", group.UUID))

		if err != nil {
			logging.Error(err.Error())
			continue
		}

		for rows.Next() {
			groupMembership := db.GroupMembership{}
			err := rows.Scan(&groupMembership.CreatedDateTime, &groupMembership.GroupUUID, &groupMembership.UserUUID)

			if err != nil {
				logging.Error(err.Error())
				continue
			}

			groupMemberships = append(groupMemberships, groupMembership)

			usersTable := db.UsersTable{}
			rows, err := usersTable.Select(db.Conn, "createddatetime, userroleid, uuid, username, authhash, firstname, lastname, email", fmt.Sprintf("uuid = '%s'", groupMembership.UserUUID))

			if err != nil {
				logging.Error(err.Error())
				continue
			}

			for rows.Next() {
				user := db.User{}
				err := rows.Scan(&user.CreatedDateTime, &user.UserroleId, &user.UUID, &user.Username, &user.AuthHash, &user.FirstName, &user.LastName, &user.Email)

				if err != nil {
					logging.Error(err.Error())
					continue
				}

				users = append(users, user)
			}
		}
	}

	pctx := plush.NewContext()
	pctx.Set("unixtostring", UnixToTimeString)
	pctx.Set("title", "Groups")
	pctx.Set("quillenabled", false)

	RenderDefault(w, "admin.users.groups.html", pctx)
}

func (ugh *AdminUserGroupsHandler) Post(w http.ResponseWriter, r *http.Request) {}

func (ugh *AdminUserGroupsHandler) Route() string { return ugh.route }

func (ugh *AdminUserGroupsHandler) HandlesGet() bool { return true }

func (ugh *AdminUserGroupsHandler) HandlesPost() bool { return false }