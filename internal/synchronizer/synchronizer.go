package synchronizer

import (
	"context"
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

func (s *Synchronizer) Run(src source.BaseSource) {
	err := s.fillReport(src)
	if err != nil {
		log.Println("An error occurred filling the report data: ", err)
	}
	errs := s.performSync()
	log.Printf("%d SDM users in IdP", len(s.report.IdPUsersInSDM))
	log.Printf("%d SDM users not in IdP", len(s.report.SDMUsersNotInIdP))
	log.Printf("%d SDM users to be updated", len(s.report.IdPUsersToUpdate))
	log.Printf("%d SDM groups in IdP", len(s.report.IdPUserGroupsInSDM))
	log.Printf("%d SDM groups not in IdP", len(s.report.SDMGroupsNotInIdP))
	log.Println(s.report.String())
	for _, err := range errs {
		log.Println(err)
	}
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

func (s *Synchronizer) performSync() []error {
	errs := []error{}
	if !*flags.PlanFlag {
		log.Print("Synchronizing users and groups")
		userErrs := s.userSynchronizer.Sync(context.Background())
		errs = append(errs, userErrs...)
		err := s.groupSynchronizer.EnrichReport()
		if err != nil {
			errs = append(errs, err)
			return errs
		}
		groupErrs := s.groupSynchronizer.Sync(context.Background())
		errs = append(errs, groupErrs...)
	}
	s.report.Complete = time.Now()
	return errs
}
