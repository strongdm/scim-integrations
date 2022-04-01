package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/source"
	"time"
)

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
	err := s.fillReport(src)
	if err != nil {
		errCh <- errors.New(fmt.Sprintf("An error occurred filling the report data: %v", err))
		close(errCh)
		return
	}
	s.performSync(errCh)
	log.Printf("%d Sink Users", len(s.report.SinkUsers))
	log.Printf("%d Sink Groups", len(s.report.SinkGroups))
	log.Printf("%d Sink Users in IdP", len(s.report.IdPUsersInSink))
	log.Printf("%d Sink Users not in IdP", len(s.report.SinkUsersNotInIdP))
	log.Printf("%d Sink Users to be updated", len(s.report.IdPUsersToUpdate))
	log.Printf("%d Sink Groups in IdP", len(s.report.IdPUserGroupsInSink))
	log.Printf("%d Sink Groups not in IdP", len(s.report.SinkGroupsNotInIdP))
	log.Println(s.report.String())
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
		log.Print("Synchronizing users and groups")
		s.userSynchronizer.Sync(context.Background(), errCh)
		s.groupSynchronizer.Sync(context.Background(), errCh)
	}
	s.report.Complete = time.Now()
}
