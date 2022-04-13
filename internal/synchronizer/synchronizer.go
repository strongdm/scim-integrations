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

var createSign = fmt.Sprintf("%s%s%s", colorGreen, "+", colorReset)
var updateSign = fmt.Sprintf("%s%s%s", colorYellow, "~", colorReset)
var deleteSign = fmt.Sprintf("%s%s%s", colorRed, "-", colorReset)

const retryLimitCount = 4

type Synchronizer struct {
	rateLimiter       *RateLimiter
	report            *Report
	userSynchronizer  *UserSynchronizer
	groupSynchronizer *GroupSynchronizer
}

func NewSynchronizer() *Synchronizer {
	report := newReport()
	rateLimiter := NewRateLimiter()
	return &Synchronizer{
		report:            report,
		rateLimiter:       rateLimiter,
		userSynchronizer:  NewUserSynchronizer(report, rateLimiter),
		groupSynchronizer: NewGroupSynchronizer(report, rateLimiter),
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
		fmt.Print("Synchronizing users...\n")
		err := s.userSynchronizer.Sync(context.Background(), snk)
		if err != nil {
			return err
		}
		fmt.Print("\nSynchronizing groups...\n")
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
		fmt.Print(colorGreen, "Groups to create:\n\n", colorReset)
		showItems(s.report.IdPUserGroupsToAdd, createSign, false, describeGroup)
	}
	if len(s.report.IdPUsersToAdd) > 0 {
		fmt.Print(colorGreen, "Users to create:\n\n", colorReset)
		showItems(s.report.IdPUsersToAdd, createSign, true, describeUser)
	}
}

func (s *Synchronizer) showEntitiesToBeUpdated() {
	if len(s.report.IdPUserGroupsToUpdate) > 0 {
		fmt.Print(colorYellow, "Groups to update:\n\n", colorReset)
		showItems(s.report.IdPUserGroupsToUpdate, updateSign, true, describeGroup)
	}
	if len(s.report.IdPUsersToUpdate) > 0 {
		fmt.Print(colorYellow, "Users to update:\n\n", colorReset)
		showItems(s.report.IdPUsersToUpdate, updateSign, true, describeUser)
	}
}

func (s *Synchronizer) showEntitiesToBeDeleted() {
	if len(s.report.SinkGroupsNotInIdP) > 0 {
		fmt.Print(colorRed, "Groups to delete:\n\n", colorReset)
		showItems(s.report.SinkGroupsNotInIdP, deleteSign, false, describeGroup)
	}
	if len(s.report.SinkUsersNotInIdP) > 0 {
		fmt.Print(colorRed, "Users to delete:\n\n")
		showItems(s.report.SinkUsersNotInIdP, deleteSign, false, describeUser)
	}
}

func showItems[T interface{}](list []*T, sign string, showDetails bool, fn func(list *T, sign string, showDetails bool)) {
	for _, item := range list {
		fn(item, sign, showDetails)
	}
	fmt.Print(colorReset)
}

func describeGroup(groupRow *sink.GroupRow, sign string, showDetails bool) {
	if len(groupRow.ID) > 0 {
		fmt.Println("\t", sign, "ID:", groupRow.ID)
	}
	fmt.Println("\t", sign, "Display Name:", groupRow.DisplayName)
	if len(groupRow.Members) > 0 && showDetails {
		fmt.Println("\t", sign, "Members:")
		for _, member := range groupRow.Members {
			fmt.Println("\t\t", sign, "E-mail:", member.Email)
		}
	}
	fmt.Println()
}

func describeUser(user *sink.UserRow, sign string, showDetails bool) {
	fmt.Println("\t", sign, "ID:", user.User.ID)
	fmt.Println("\t\t", sign, " Display Name:", user.User.GivenName, user.User.FamilyName)
	fmt.Println("\t\t", sign, " User Name:", user.User.UserName)
	if showDetails {
		fmt.Println("\t\t", sign, " Family Name:", user.User.FamilyName)
		fmt.Println("\t\t", sign, " Given Name:", user.User.GivenName)
		fmt.Println("\t\t", sign, " Active:", user.User.Active)
		if len(user.User.GroupNames) > 0 {
			fmt.Println("\t\t", sign, " Groups:")
			for _, group := range user.User.GroupNames {
				fmt.Println("\t\t\t", sign, group)
			}
		}
	}
	fmt.Println()
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

func safeRetry(rateLimiter *RateLimiter, fn func() error, actionDescription string) error {
	var retryCounter int
	err := backoff.Retry(try(rateLimiter, fn, retryCounter, actionDescription), getBackoffConfig())
	return err
}

func try(rateLimiter *RateLimiter, fn func() error, retryCounter int, actionDescription string) func() error {
	return func() error {
		rateLimiter.VerifyLimit()
		err := fn()
		rateLimiter.IncreaseCounter()
		if err != nil {
			if !ErrorIsUnexpected(err) {
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
	}
}

func getBackoffConfig() backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewExponentialBackOff(), retryLimitCount)
}
