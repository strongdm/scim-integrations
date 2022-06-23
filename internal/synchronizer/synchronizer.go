package synchronizer

import (
	"context"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/repository"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
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
	report := NewReport()
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
	err = s.groupSynchronizer.EnrichReport(snk)
	if err != nil {
		return err
	}
	return nil
}

func (s *Synchronizer) performSync(snk sink.BaseSink) error {
	if !*flags.ApplyFlag {
		return nil
	}
	err := s.userSynchronizer.Sync(context.Background(), snk)
	if err != nil {
		return err
	}
	err = s.groupSynchronizer.Sync(context.Background(), snk)
	if err != nil {
		return err
	}
	s.report.Succeeded()
	fmt.Println("Sync process completed at", s.report.Complete.String())
	if hasDefinedDatabase() {
		_, err := repository.NewReportRepository().Insert(*reportToRepositoryReportsRow(s.report))
		if err != nil {
			fmt.Fprintln(os.Stderr, "An error occurred when caching a report:", err.Error())
		}
	}
	return nil
}

func hasDefinedDatabase() bool {
	dbFilePath := os.Getenv("SDM_SCIM_REPORTS_DATABASE_PATH")
	return dbFilePath != ""
}
