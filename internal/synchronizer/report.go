package synchronizer

import (
	"encoding/json"
	"fmt"
	"scim-integrations/internal/sink/sdmscim"
	"scim-integrations/internal/source"
	"time"
)

type Report struct {
	Start    time.Time
	Complete time.Time

	IdPUsers         []*source.User
	IdPUsersToAdd    []*source.User
	IdPUsersInSDM    []*source.User
	IdPUsersToUpdate []*source.User

	IdPUserGroups      []*source.UserGroup
	IdPUserGroupsToAdd []*source.UserGroup
	IdPUserGroupsInSDM []*source.UserGroup

	SDMUsersNotInIdP  []*sdmscim.UserRow
	SDMGroupsNotInIdP []*sdmscim.GroupRow
}

func newReport() *Report {
	return &Report{
		IdPUsers:           []*source.User{},
		IdPUsersToAdd:      []*source.User{},
		IdPUsersInSDM:      []*source.User{},
		IdPUsersToUpdate:   []*source.User{},
		IdPUserGroups:      []*source.UserGroup{},
		IdPUserGroupsToAdd: []*source.UserGroup{},
		IdPUserGroupsInSDM: []*source.UserGroup{},
		SDMUsersNotInIdP:   []*sdmscim.UserRow{},
		SDMGroupsNotInIdP:  []*sdmscim.GroupRow{},
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
		len(rpt.IdPUsers), len(rpt.IdPUsersInSDM), len(rpt.IdPUserGroupsInSDM))
}
