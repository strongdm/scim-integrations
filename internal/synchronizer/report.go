package synchronizer

import (
	"encoding/json"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"time"
)

type Report struct {
	Start    time.Time
	Complete time.Time

	IdPUsers         []*source.User
	IdPUsersToAdd    []*sink.UserRow
	IdPUsersInSink   []*sink.UserRow
	IdPUsersToUpdate []*sink.UserRow

	IdPUserGroups         []*source.UserGroup
	IdPUserGroupsToAdd    []*sink.GroupRow
	IdPUserGroupsInSink   []*sink.GroupRow
	IdPUserGroupsToUpdate []*sink.GroupRow

	SinkUsers          []*sink.UserRow
	SinkGroups         []*sink.GroupRow
	SinkUsersNotInIdP  []*sink.UserRow
	SinkGroupsNotInIdP []*sink.GroupRow
}

func newReport() *Report {
	return &Report{
		IdPUsers:              []*source.User{},
		IdPUsersToAdd:         []*sink.UserRow{},
		IdPUsersInSink:        []*sink.UserRow{},
		IdPUsersToUpdate:      []*sink.UserRow{},
		IdPUserGroups:         []*source.UserGroup{},
		IdPUserGroupsToAdd:    []*sink.GroupRow{},
		IdPUserGroupsInSink:   []*sink.GroupRow{},
		IdPUserGroupsToUpdate: []*sink.GroupRow{},
		SinkUsers:             []*sink.UserRow{},
		SinkGroups:            []*sink.GroupRow{},
		SinkUsersNotInIdP:     []*sink.UserRow{},
		SinkGroupsNotInIdP:    []*sink.GroupRow{},
	}
}

func (rpt *Report) String() string {
	out, err := json.MarshalIndent(rpt, "", "\t")
	if err != nil {
		return fmt.Sprintf("error building JSON report: %s\n\n%s", err, rpt.short())
	}
	return string(out)
}

func (rpt *Report) short() string {
	return fmt.Sprintf("%d IdP users, %d strongDM users in IdP, %d strongDM groups in IdP\n",
		len(rpt.IdPUsers), len(rpt.IdPUsersInSink), len(rpt.IdPUserGroupsInSink))
}

func (rpt *Report) HaveUsersSyncContent() bool {
	haveUsersToAdd := len(rpt.IdPUsersToAdd) > 0
	haveUsersToUpdate := len(rpt.IdPUsersToUpdate) > 0
	haveUsersToDelete := len(rpt.SinkUsersNotInIdP) > 0
	return haveUsersToAdd || haveUsersToUpdate || (*flags.DeleteUsersNotInIdPFlag && haveUsersToDelete)
}

func (rpt *Report) HaveGroupsSyncContent() bool {
	haveGroupsToAdd := len(rpt.IdPUserGroupsToAdd) > 0
	haveGroupsToUpdate := len(rpt.IdPUserGroupsToUpdate) > 0
	haveGroupsToDelete := len(rpt.SinkGroupsNotInIdP) > 0
	return haveGroupsToAdd || haveGroupsToUpdate || (*flags.DeleteGroupsNotInIdPFlag && haveGroupsToDelete)
}
