package synchronizer

import (
	"context"
	"errors"
	"log"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"time"
)

type Synchronizer struct {
	report            *Report
	userSynchronizer  *UserSynchronizer
	groupSynchronizer *GroupSynchronizer
}

func NewSynchronizer() *Synchronizer {
	report := &Report{}
	sdmSink := sink.NewSDMService()
	return &Synchronizer{
		report:            report,
		userSynchronizer:  NewUserSynchronize(sdmSink, report),
		groupSynchronizer: NewGroupSynchronize(sdmSink, report),
	}
}

func (s *Synchronizer) Run(src source.BaseSource) error {
	err := s.fillReport(src)
	if err != nil {
		return err
	}
	err = s.performSync()
	if err != nil {
		return err
	}
	log.Printf("%d SDM users in IdP", len(s.report.IdPUsersInSDM))
	log.Printf("%d SDM users not in IdP", len(s.report.SDMUsersNotInIdP))
	log.Printf("%d SDM groups in IdP", len(s.report.IdPUserGroupsInSDM))
	log.Printf("%d SDM groups not in IdP", len(s.report.SDMGroupsNotInIdP))
	log.Println(s.report.String())
	return nil
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
	s.userSynchronizer.EnrichReport()
	s.groupSynchronizer.EnrichReport()
	return nil
}

func (s *Synchronizer) performSync() error {
	if !*flags.PlanFlag {
		log.Print("Synchronizing users and groups")
		err := s.userSynchronizer.Sync(context.Background())
		if err != nil {
			return errors.New("error synchronizing users: " + err.Error())
		}
		err = s.groupSynchronizer.Sync(context.Background())
		if err != nil {
			return errors.New("error synchronizing groups: " + err.Error())
		}
	}
	s.report.Complete = time.Now()
	return nil
}