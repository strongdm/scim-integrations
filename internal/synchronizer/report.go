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
	IdPUsersToCreate []*sink.UserRow
	IdPUsersInSink   []*sink.UserRow
	IdPUsersToUpdate []*sink.UserRow

	IdPGroups         []*source.UserGroup
	IdPGroupsToCreate []*sink.GroupRow
	IdPGroupsInSink   []*sink.GroupRow
	IdPGroupsToUpdate []*sink.GroupRow

	SinkUsers          []*sink.UserRow
	SinkGroups         []*sink.GroupRow
	SinkUsersNotInIdP  []*sink.UserRow
	SinkGroupsNotInIdP []*sink.GroupRow

	CreatedUsersCount  int
	CreatedGroupsCount int
	UpdatedUsersCount  int
	UpdatedGroupsCount int
	DeletedUsersCount  int
	DeletedGroupsCount int
	Succeed            int
}

func NewReport() *Report {
	return &Report{Succeed: 1}
}

func (rpt *Report) Succeeded() {
	rpt.Complete = time.Now()
	rpt.Succeed = 0
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
		len(rpt.IdPUsers), len(rpt.IdPUsersInSink), len(rpt.IdPGroupsInSink))
}

func (rpt *Report) showPlan() {
	if *flags.AllOperationFlag || *flags.AddOperationFlag {
		rpt.showEntitiesToBeCreated()
	}
	if *flags.AllOperationFlag || *flags.UpdateOperationFlag {
		rpt.showEntitiesToBeUpdated()
	}
	if *flags.AllOperationFlag || *flags.DeleteOperationFlag {
		rpt.showEntitiesToBeDeleted()
	}
}

func (rpt *Report) showEntitiesToBeCreated() {
	if len(rpt.IdPGroupsToCreate) > 0 {
		fmt.Print(colorGreen, "Groups to create:\n\n", colorReset)
		showItems(rpt.IdPGroupsToCreate, createSign, false, rpt.describeGroup)
	}
	if len(rpt.IdPUsersToCreate) > 0 {
		fmt.Print(colorGreen, "Users to create:\n\n", colorReset)
		showItems(rpt.IdPUsersToCreate, createSign, true, rpt.describeUser)
	}
}

func (rpt *Report) showEntitiesToBeUpdated() {
	if len(rpt.IdPGroupsToUpdate) > 0 {
		fmt.Print(colorYellow, "Groups to update:\n\n", colorReset)
		showItems(rpt.IdPGroupsToUpdate, updateSign, true, rpt.describeGroup)
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
	fmt.Printf("\t %s Display Name: %s", sign, groupRow.DisplayName)
	if groupRow.Path != "" {
		fmt.Printf(" (%s)", groupRow.Path)
	}
	fmt.Println()
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
	fmt.Printf("\t\t %s User Name: %s\n", sign, user.User.UserName)
	fmt.Printf("\t\t %s Display Name: %s %s\n", sign, user.User.GivenName, user.User.FamilyName)
	if showDetails {
		fmt.Printf("\t\t %s Given Name: %s\n", sign, user.User.GivenName)
		fmt.Printf("\t\t %s Family Name: %s\n", sign, user.User.FamilyName)
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
