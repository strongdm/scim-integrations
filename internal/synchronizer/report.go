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

func (rpt *Report) showPlan() {
	rpt.showEntitiesToBeCreated()
	rpt.showEntitiesToBeUpdated()
	rpt.showEntitiesToBeDeleted()
}

func (rpt *Report) showEntitiesToBeCreated() {
	if len(rpt.IdPUserGroupsToAdd) > 0 {
		fmt.Print(colorGreen, "Groups to create:\n\n", colorReset)
		showItems(rpt.IdPUserGroupsToAdd, createSign, false, rpt.describeGroup)
	}
	if len(rpt.IdPUsersToAdd) > 0 {
		fmt.Print(colorGreen, "Users to create:\n\n", colorReset)
		showItems(rpt.IdPUsersToAdd, createSign, true, rpt.describeUser)
	}
}

func (rpt *Report) showEntitiesToBeUpdated() {
	if len(rpt.IdPUserGroupsToUpdate) > 0 {
		fmt.Print(colorYellow, "Groups to update:\n\n", colorReset)
		showItems(rpt.IdPUserGroupsToUpdate, updateSign, true, rpt.describeGroup)
	}
	if len(rpt.IdPUsersToUpdate) > 0 {
		fmt.Print(colorYellow, "Users to update:\n\n", colorReset)
		showItems(rpt.IdPUsersToUpdate, updateSign, true, rpt.describeUser)
	}
}

func (rpt *Report) showEntitiesToBeDeleted() {
	if len(rpt.SinkGroupsNotInIdP) > 0 {
		fmt.Print(colorRed, "Groups to delete:\n\n", colorReset)
		showItems(rpt.SinkGroupsNotInIdP, deleteSign, false, rpt.describeGroup)
	}
	if len(rpt.SinkUsersNotInIdP) > 0 {
		fmt.Print(colorRed, "Users to delete:\n\n")
		showItems(rpt.SinkUsersNotInIdP, deleteSign, false, rpt.describeUser)
	}
}

func showItems[T interface{}](list []*T, sign string, showDetails bool, fn func(list *T, sign string, showDetails bool)) {
	for _, item := range list {
		fn(item, sign, showDetails)
	}
	fmt.Print(colorReset)
}

func (*Report) describeGroup(groupRow *sink.GroupRow, sign string, showDetails bool) {
	if len(groupRow.ID) > 0 {
		fmt.Printf("\t %s ID: %s\n", sign, groupRow.ID)
	}
	fmt.Printf("\t %s Display Name: %s\n", sign, groupRow.DisplayName)
	if len(groupRow.Members) > 0 && showDetails {
		fmt.Printf("\t\t %s Members:\n", sign)
		for _, member := range groupRow.Members {
			fmt.Printf("\t\t\t %s E-mail: %s\n", sign, member.Email)
		}
	}
	fmt.Println()
}

func (*Report) describeUser(user *sink.UserRow, sign string, showDetails bool) {
	fmt.Printf("\t %s ID: %s\n", sign, user.User.ID)
	fmt.Printf("\t\t %s Display Name: %s %s\n", sign, user.User.GivenName, user.User.FamilyName)
	fmt.Printf("\t\t %s User Name: %s\n", sign, user.User.UserName)
	if showDetails {
		fmt.Printf("\t\t %s Family Name: %s\n", sign, user.User.FamilyName)
		fmt.Printf("\t\t %s Given Name: %s\n", sign, user.User.GivenName)
		fmt.Printf("\t\t %s Active: %v\n", sign, user.User.Active)
		if len(user.User.GroupNames) > 0 {
			fmt.Printf("\t\t %s Groups:\n", sign)
			for _, group := range user.User.GroupNames {
				fmt.Printf("\t\t\t %s %s\n", sign, group)
			}
		}
	}
	fmt.Println()
}

// TODO Method
func (rpt *Report) showVerboseOutput() {
	if *flags.VerboseFlag {
		fmt.Printf("%d Sink Users\n", len(rpt.SinkUsers))
		fmt.Printf("%d Sink Groups\n", len(rpt.SinkGroups))
		fmt.Printf("%d Sink Users in IdP\n", len(rpt.IdPUsersInSink))
		fmt.Printf("%d Sink Users not in IdP\n", len(rpt.SinkUsersNotInIdP))
		fmt.Printf("%d Sink Users to be updated\n", len(rpt.IdPUsersToUpdate))
		fmt.Printf("%d Sink Groups in IdP\n", len(rpt.IdPUserGroupsInSink))
		fmt.Printf("%d Sink Groups not in IdP\n", len(rpt.SinkGroupsNotInIdP))
		fmt.Println(rpt.String())
	}
}
