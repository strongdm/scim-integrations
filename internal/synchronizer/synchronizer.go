package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/source"
	"time"
)

const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorGreen string = "\033[32m"
const colorYellow string = "\033[33m"

type Synchronizer struct {
	report            *Report
	userSynchronizer  *UserSynchronizer
	groupSynchronizer *GroupSynchronizer
}

func NewSynchronizer() *Synchronizer {
	report := newReport()
	return &Synchronizer{
		report:            report,
		userSynchronizer:  NewUserSynchronizer(report),
		groupSynchronizer: NewGroupSynchronizer(report),
	}
}

func (s *Synchronizer) Run(src source.BaseSource, errCh chan error) {
	fmt.Println("Collecting data...")
	err := s.fillReport(src)
	if err != nil {
		errCh <- errors.New(fmt.Sprintf("An error occurred filling the report data: %v", err))
		close(errCh)
		return
	}
	s.performSync(errCh)
	s.showEntitiesToBeCreated()
	s.showEntitiesToBeUpdated()
	s.showEntitiesToBeDeleted()
	s.showVerboseOutput()
	close(errCh)
}

func (s *Synchronizer) fillReport(src source.BaseSource) error {
	s.report.Start = time.Now()
	sourceUsers, err := src.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sourceGroups := src.ExtractGroupsFromUsers(sourceUsers)
	s.report.IdPUsers = sourceUsers
	s.report.IdPUserGroups = sourceGroups
	err = s.userSynchronizer.EnrichReport()
	if err != nil {
		return err
	}
	err = s.groupSynchronizer.EnrichReport()
	if err != nil {
		return err
	}
	return nil
}

func (s *Synchronizer) performSync(errCh chan error) {
	if !*flags.PlanFlag {
		fmt.Print("Synchronizing users and groups\n")
		s.userSynchronizer.Sync(context.Background(), errCh)
		s.groupSynchronizer.Sync(context.Background(), errCh)
	}
	s.report.Complete = time.Now()
}

func (s *Synchronizer) showEntitiesToBeCreated() {
	if len(s.report.IdPUsersToAdd) > 0 {
		fmt.Print(colorGreen, "\nUsers to create:\n\n")
		for _, user := range s.report.IdPUsersToAdd {
			fmt.Println("\t+ ID:", user.ID)
			fmt.Println("\t\t+ Family Name:", user.FamilyName)
			fmt.Println("\t\t+ Given Name:", user.GivenName)
			fmt.Println("\t\t+ User Name:", user.UserName)
			fmt.Println("\t\t+ Active:", user.Active)
			if user.SDMObjectID != "" {
				fmt.Println("\t\t+ SDMID:", user.SDMObjectID)
			}
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
	if len(s.report.IdPUserGroupsToAdd) > 0 {
		fmt.Print(colorGreen, "Groups to create:\n\n")
		for _, groupRow := range s.report.IdPUserGroupsToAdd {
			fmt.Println("\t+ Display Name:", groupRow.DisplayName)
			if groupRow.SDMObjectID != "" {
				fmt.Println("\t+ SDMID:", groupRow.SDMObjectID)
			}
			if len(groupRow.Members) > 0 {
				fmt.Println("\t\t+ Members:")
				for _, member := range groupRow.Members {
					fmt.Println("\t\t\t+ ID:", member.ID)
					fmt.Println("\t\t\t+ E-mail:", member.Email)
					if member.SDMObjectID != "" {
						fmt.Println("\t\t\t+ SDMID:", member.SDMObjectID)
					}
					fmt.Println()
				}
			} else {
				fmt.Println()
			}
		}
		fmt.Print(colorReset)
	}
}

func (s *Synchronizer) showEntitiesToBeUpdated() {
	if len(s.report.IdPUsersToUpdate) > 0 {
		fmt.Println(colorYellow, "Users to update:")
		for _, user := range s.report.IdPUsersToUpdate {
			fmt.Println("\t~ ID:", user.ID)
			fmt.Println("\t\t~ Family Name:", user.FamilyName)
			fmt.Println("\t\t~ Given Name:", user.GivenName)
			fmt.Println("\t\t~ User Name:", user.UserName)
			fmt.Println("\t\t~ Active:", user.Active)
			if user.SDMObjectID != "" {
				fmt.Println("\t\t~ SDMID:", user.SDMObjectID)
			}
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
}

func (s *Synchronizer) showEntitiesToBeDeleted() {
	if len(s.report.SinkUsersNotInIdP) > 0 && *flags.DeleteUsersNotInIdPFlag {
		fmt.Print(colorRed, "Users to delete:\n\n")
		for _, userRow := range s.report.SinkUsersNotInIdP {
			user := userRow.User
			fmt.Println("\t- ID:", user.ID)
			fmt.Println("\t\t- Display Name:", user.DisplayName)
			fmt.Println("\t\t- Family Name:", user.FamilyName)
			fmt.Println("\t\t- Given Name:", user.GivenName)
			fmt.Println("\t\t- User Name:", user.UserName)
			fmt.Println("\t\t- Active:", user.Active)
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
	if len(s.report.SinkGroupsNotInIdP) > 0 && *flags.DeleteGroupsNotInIdPFlag {
		fmt.Println(colorRed, "Groups to delete:")
		for _, groupRow := range s.report.SinkGroupsNotInIdP {
			fmt.Println("\t- ID:", groupRow.ID)
			fmt.Println("\t\t- Display Name:", groupRow.DisplayName)
			if len(groupRow.Members) > 0 {
				fmt.Println("\t\t- Members:")
				for _, member := range groupRow.Members {
					fmt.Println("\t\t\t- ID:", member.ID)
					fmt.Println("\t\t\t- E-mail:", member.Email)
					if member.SDMObjectID != "" {
						fmt.Println("\t\t\t- SDMID:", member.SDMObjectID)
					}
					fmt.Println()
				}
			} else {
				fmt.Println()
			}
		}
		fmt.Print(colorReset)
	}
}

func (s *Synchronizer) showVerboseOutput() {
	if *flags.VerboseFlag {
		fmt.Printf("%d Sink Users\n", len(s.report.SinkUsers))
		fmt.Printf("%d Sink Groups\n", len(s.report.SinkGroups))
		fmt.Printf("%d Sink Users in IdP\n", len(s.report.IdPUsersInSink))
		fmt.Printf("%d Sink Users not in IdP\n", len(s.report.SinkUsersNotInIdP))
		fmt.Printf("%d Sink Users to be updated\n", len(s.report.IdPUsersToUpdate))
		fmt.Printf("%d Sink Groups in IdP\n", len(s.report.IdPUserGroupsInSink))
		fmt.Printf("%d Sink Groups not in IdP\n", len(s.report.SinkGroupsNotInIdP))
		fmt.Println(s.report.String())
	}
}
