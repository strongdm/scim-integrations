package synchronizer

import (
	"context"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/repository"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"strings"
	"time"
)

const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorGreen string = "\033[32m"
const colorYellow string = "\033[33m"

var createSign = fmt.Sprintf("%s%s%s", colorGreen, "+", colorReset)
var updateSign = fmt.Sprintf("%s%s%s", colorYellow, "~", colorReset)
var deleteSign = fmt.Sprintf("%s%s%s", colorRed, "-", colorReset)

type Synchronizer struct {
	retrier           Retrier
	report            *Report
	userSynchronizer  *UserSynchronizer
	groupSynchronizer *GroupSynchronizer
}

func NewSynchronizer() *Synchronizer {
	report := &Report{}
	retrier := newRetrier(newRateLimiter())
	return &Synchronizer{
		report:            report,
		retrier:           retrier,
		userSynchronizer:  newUserSynchronizer(report, retrier),
		groupSynchronizer: newGroupSynchronizer(report, retrier),
	}
}

// Run collect the source data and synchronize it with the sink
func (s *Synchronizer) Run(src source.BaseSource, snk sink.BaseSink) error {
	s.report.Start = time.Now()
	fmt.Println("Starting at", s.report.Start.String())
	fmt.Println("Collecting data...")
	err := s.fillReport(src, snk)
	if err != nil {
		return fmt.Errorf("an error occurred filling the report data: %v", err)
	}
	fmt.Println()
	s.report.showPlan()
	err = s.performSync(snk)
	if err != nil {
		return err
	}
	return nil
}

func (s *Synchronizer) fillReport(src source.BaseSource, snk sink.BaseSink) error {
	sourceUsers, err := src.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sourceGroups := src.ExtractGroupsFromUsers(sourceUsers)
	s.report.IdPUsers = sourceUsers
	s.report.IdPGroups = sourceGroups
	err = s.userSynchronizer.EnrichReport(snk)
	if err != nil {
		return err
	}
	s.report.UsersToCreateCount = len(s.report.IdPUsersToCreate)
	s.report.UsersToUpdateCount = len(s.report.IdPUsersToUpdate)
	s.report.UsersToDeleteCount = len(s.report.SinkUsersNotInIdP)
	err = s.groupSynchronizer.EnrichReport(snk)
	if err != nil {
		return err
	}
	s.report.GroupsToCreateCount = len(s.report.IdPGroupsToCreate)
	s.report.GroupsToUpdateCount = len(s.report.IdPGroupsToUpdate)
	s.report.GroupsToDeleteCount = len(s.report.SinkGroupsNotInIdP)
	return nil
}

func (s *Synchronizer) performSync(snk sink.BaseSink) error {
	if *flags.ApplyFlag {
		err := s.userSynchronizer.Sync(context.Background(), snk)
		if err != nil {
			return err
		}
		err = s.groupSynchronizer.Sync(context.Background(), snk)
		if err != nil {
			return err
		}
		s.report.Complete = time.Now()
		fmt.Println("Sync process completed at", s.report.Complete.String())
		if isDockerized() {
			_, err := repository.NewReportRepository().Insert(*reportToRepositoryReportsRow(s.report))
			if err != nil {
				fmt.Fprintln(os.Stderr, "An error occurred when caching a report:", err.Error())
			}
		}
	}
	return nil
}

func isDockerized() bool {
	dockerizedEnv := strings.ToLower(os.Getenv("DOCKERIZED"))
	return dockerizedEnv == "true"
}
