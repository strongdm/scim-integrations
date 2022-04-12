package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorGreen string = "\033[32m"
const colorYellow string = "\033[33m"
const retryLimitCount = 4

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

func (s *Synchronizer) Run(src source.BaseSource, snk sink.BaseSink) error {
	fmt.Println("Collecting data...")
	err := s.fillReport(src, snk)
	if err != nil {
		return fmt.Errorf("an error occurred filling the report data: %v", err)
	}
	fmt.Println()
	s.showEntitiesToBeCreated()
	s.showEntitiesToBeUpdated()
	s.showEntitiesToBeDeleted()
	err = s.performSync(snk)
	if err != nil {
		return err
	}
	s.showVerboseOutput()
	return nil
}

func (s *Synchronizer) fillReport(src source.BaseSource, snk sink.BaseSink) error {
	s.report.Start = time.Now()
	sourceUsers, err := src.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sourceGroups := src.ExtractGroupsFromUsers(sourceUsers)
	s.report.IdPUsers = sourceUsers
	s.report.IdPUserGroups = sourceGroups
	err = s.userSynchronizer.EnrichReport(snk)
	if err != nil {
		return err
	}
	err = s.groupSynchronizer.EnrichReport(snk)
	if err != nil {
		return err
	}
	return nil
}

func (s *Synchronizer) performSync(snk sink.BaseSink) error {
	if !*flags.PlanFlag {
		fmt.Print("Summary:\n\n")
		fmt.Print("Synchronizing users:\n\n")
		err := s.userSynchronizer.Sync(context.Background(), snk)
		if err != nil {
			return err
		}
		fmt.Print("\nSynchronizing groups:\n\n")
		err = s.groupSynchronizer.Sync(context.Background(), snk)
		if err != nil {
			return err
		}
		fmt.Println()
	}
	s.report.Complete = time.Now()
	return nil
}

func (s *Synchronizer) showEntitiesToBeCreated() {
	if len(s.report.IdPUserGroupsToAdd) > 0 {
		fmt.Print(colorGreen, "Groups to create:\n\n")
		for _, groupRow := range s.report.IdPUserGroupsToAdd {
			fmt.Println("\t+ Display Name:", groupRow.DisplayName)
			if groupRow.SDMObjectID != "" {
				fmt.Println("\t+ SDMID:", groupRow.SDMObjectID)
			}
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
	if len(s.report.IdPUsersToAdd) > 0 {
		fmt.Print(colorGreen, "Users to create:\n\n")
		for _, user := range s.report.IdPUsersToAdd {
			fmt.Println("\t+ ID:", user.User.ID)
			fmt.Println("\t\t+ Family Name:", user.User.FamilyName)
			fmt.Println("\t\t+ Given Name:", user.User.GivenName)
			fmt.Println("\t\t+ User Name:", user.User.UserName)
			fmt.Println("\t\t+ Active:", user.User.Active)
			fmt.Println("\t\t+ Groups:")
			for _, group := range user.Groups {
				fmt.Println("\t\t\t+", group)
			}
			if user.User.ID != "" {
				fmt.Println("\t\t+ SDMID:", user.User.ID)
			}
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
}

func (s *Synchronizer) showEntitiesToBeUpdated() {
	if len(s.report.IdPUsersToUpdate) > 0 {
		fmt.Print(colorYellow, "Users to update:\n\n")
		for _, user := range s.report.IdPUsersToUpdate {
			fmt.Println("\t~ ID:", user.User.ID)
			fmt.Println("\t\t~ Family Name:", user.User.FamilyName)
			fmt.Println("\t\t~ Given Name:", user.User.GivenName)
			fmt.Println("\t\t~ User Name:", user.User.UserName)
			fmt.Println("\t\t~ Active:", user.User.Active)
			fmt.Println("\t\t~ Groups:")
			for _, group := range user.Groups {
				fmt.Println("\t\t\t~", group)
			}
			if user.User.ID != "" {
				fmt.Println("\t\t~ SDMID:", user.User.ID)
			}
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
}

func (s *Synchronizer) showEntitiesToBeDeleted() {
	if len(s.report.SinkGroupsNotInIdP) > 0 {
		fmt.Print(colorRed, "Groups to delete:\n\n")
		for _, groupRow := range s.report.SinkGroupsNotInIdP {
			fmt.Println("\t- ID:", groupRow.ID)
			fmt.Println("\t\t- Display Name:", groupRow.DisplayName)
			fmt.Println()
		}
		fmt.Print(colorReset)
	}
	if len(s.report.SinkUsersNotInIdP) > 0 {
		fmt.Print(colorRed, "Users to delete:\n\n")
		for _, userRow := range s.report.SinkUsersNotInIdP {
			user := userRow.User
			fmt.Println("\t- ID:", user.ID)
			fmt.Println("\t\t- Display Name:", user.DisplayName)
			fmt.Println("\t\t- User Name:", user.UserName)
			fmt.Println()
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

func safeRetry(fn func() error, actionDescription string) error {
	var retryCounter int
	err := backoff.Retry(func() error {
		err := fn()
		if err != nil {
			if !sink.ErrorIsUnexpected(err) {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}
			retryCounter++
			if retryCounter < retryLimitCount {
				fmt.Fprintf(os.Stderr, "Failed %s. Retrying the operation for the %dst time\n", actionDescription, retryCounter)
			}
			return errors.New("retry limit exceeded with the following error: " + err.Error())
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), retryLimitCount))
	return err
}
