package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/idp"
	"scim-integrations/internal/sink"
	"time"

	"github.com/strongdm/scimsdk/scimsdk"
)

type Syncer struct {
	report *Report
}

type Report struct {
	Start         time.Time
	Complete      time.Time
	IdPUsers      []idp.IdPUser
	IdPUserGroups []sink.SinkUserGroup

	SDMUsersInIdP    []idp.IdPUser
	SDMUsersNotInIdP []idp.IdPUser
	SDMNewUsers      []scimsdk.User

	SDMGroupsInIdP    []sink.SinkUserGroup
	SDMGroupsNotInIdP []scimsdk.Group
	SDMNewGroups      []scimsdk.Group
}

func NewRunner() *Syncer {
	return &Syncer{&Report{}}
}

func (syncer *Syncer) Run(idp idp.BaseIdP) error {
	idpUsers, err := idp.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	groups := extractGroupsFromUsers(idpUsers)
	syncer.report.IdPUsers = idpUsers
	syncer.report.IdPUserGroups = groups
	sdmService := sink.NewSDMService()
	if !*flags.PlanFlag {
		log.Print("Synchronizing users and groups")
		err := NewUserSynchronize(sdmService, syncer.report).Sync(context.Background(), idpUsers)
		if err != nil {
			return err
		}
		log.Printf("%d SDM users in IdP", len(syncer.report.SDMGroupsInIdP))
		log.Printf("%d SDM users not in IdP", len(syncer.report.SDMGroupsNotInIdP))
		err = NewGroupSynchronize(sdmService, syncer.report).Sync(context.Background(), groups)
		if err != nil {
			return errors.New("error synchronizing groups: " + err.Error())
		}
		log.Printf("%d SDM groups in IdP", len(syncer.report.SDMGroupsInIdP))
		log.Printf("%d SDM groups not in IdP", len(syncer.report.SDMGroupsNotInIdP))
	}
	syncer.report.Complete = time.Now()
	fmt.Println(syncer.report.String())
	return nil
}

func extractGroupsFromUsers(users []idp.IdPUser) []sink.SinkUserGroup {
	var groups []sink.SinkUserGroup
	mappedGroupMembers := map[string][]scimsdk.GroupMember{}
	for _, user := range users {
		for _, userGroup := range user.Groups {
			if _, ok := mappedGroupMembers[userGroup]; !ok {
				mappedGroupMembers[userGroup] = []scimsdk.GroupMember{
					{
						ID:    user.ID,
						Email: user.UserName,
					},
				}
			} else {
				mappedGroupMembers[userGroup] = append(mappedGroupMembers[userGroup], scimsdk.GroupMember{
					ID:    user.ID,
					Email: user.UserName,
				})
			}
		}
	}
	for groupName, members := range mappedGroupMembers {
		groups = append(groups, sink.SinkUserGroup{DisplayName: groupName, Members: members})
	}
	return groups
}

func (rpt *Report) String() string {
	if !*flags.JsonFlag {
		return rpt.short()
	}
	out, err := json.MarshalIndent(rpt, "", "\t")
	if err != nil {
		return fmt.Sprintf("error building JSON report: %s\n\n%s", err, rpt.short())
	}
	return string(out)
}

func (rpt *Report) short() string {
	return fmt.Sprintf("%d IdP users, %d strongDM users in IdP, %d strongDM groups in IdP\n",
		len(rpt.IdPUsers), len(rpt.SDMUsersInIdP), len(rpt.SDMGroupsInIdP))
}
