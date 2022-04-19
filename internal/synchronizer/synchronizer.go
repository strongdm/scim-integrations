package synchronizer

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
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
	report := newReport()
	retrier := newRetrier(newRateLimiter())
	return &Synchronizer{
		report:            report,
		retrier:           retrier,
		userSynchronizer:  newUserSynchronizer(report, retrier),
		groupSynchronizer: newGroupSynchronizer(report, retrier),
	}
}

func (s *Synchronizer) Run(src source.BaseSource, snk sink.BaseSink) error {
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
	s.report.showVerboseOutput()
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
	// TODO Change this flag by another named apply
	if !*flags.PlanFlag {
		haveUsersSyncContent := s.report.HaveUsersSyncContent()   // TODO Move this to this class
		haveGroupsSyncContent := s.report.HaveGroupsSyncContent() // TODO Move this to this class
		s.retrier.GetRateLimiter().Start()                        // TODO Move this somewhere inside the code
		if haveUsersSyncContent {
			fmt.Print("Synchronizing users...\n")
			err := s.userSynchronizer.Sync(context.Background(), snk)
			if err != nil {
				return err
			}
		}
		if haveGroupsSyncContent {
			fmt.Print("\nSynchronizing groups...\n")
			err := s.groupSynchronizer.Sync(context.Background(), snk)
			if err != nil {
				return err
			}
		}
		if haveUsersSyncContent || haveGroupsSyncContent {
			fmt.Println()
		}
	}
	s.report.Complete = time.Now()
	return nil
}
