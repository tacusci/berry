package web

import (
	"fmt"
	"github.com/gobuffalo/plush"
	"github.com/tacusci/logging"
	"net/http"

	"github.com/tacusci/berrycms/db"
)

type AdminUserGroupsHandler struct {
	Router *MutableRouter
	route  string
}

func (ugh *AdminUserGroupsHandler) Get(w http.ResponseWriter, r *http.Request) {

	gt := db.GroupTable{}

	groupRows, err := gt.Select(db.Conn, "createddatetime, uuid, title", "")
	if err != nil {
		Error(w, err)
	}

	defer groupRows.Close()

	for groupRows.Next() {
		group := db.Group{}
		groupRows.Scan(&group.CreatedDateTime, &group.UUID, &group.Title)

		gmt := db.GroupMembershipTable{}

		groupMembershipRows, err := gmt.Select(db.Conn, "groupuuid, useruuid", fmt.Sprintf("groupuuid = '%s'", group.UUID))
		if err != nil {
			Error(w, err)
		}

		for groupMembershipRows.Next() {
			groupMembership := db.GroupMembership{}
			groupMembershipRows.Scan(&groupMembership.CreatedDateTime, &groupMembership.UUID, &groupMembership.GroupUUID, &groupMembership.UserUUID)

			logging.Debug(fmt.Sprintf("Found group memmbership for group of UUID: %s", groupMembership.GroupUUID))
		}

		groupMembershipRows.Close()
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
