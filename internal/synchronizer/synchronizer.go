package synchronizer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type Report struct {
	Start    time.Time
	Complete time.Time

	SourceUsers         []source.SourceUser
	SourceUsersToAdd    []source.SourceUser
	SourceMatchingUsers []source.SourceUser

	SourceUserGroups         []source.SourceUserGroup
	SourceUserGroupsToAdd    []source.SourceUserGroup
	SourceMatchingUserGroups []source.SourceUserGroup

	SDMUnmatchingUsers  []sink.SDMUserRow
	SDMUnmatchingGroups []sink.SDMGroupRow
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
	sourceUsers, err := src.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sourceGroups := src.ExtractGroupsFromUsers(sourceUsers)
	s.userSynchronizer.SyncUsersData(sourceUsers)
	s.groupSynchronizer.SyncGroupsData(sourceGroups)
	s.report.SourceUsers = sourceUsers
	s.report.SourceUserGroups = sourceGroups
	if !*flags.PlanFlag {
		log.Print("Synchronizing users and groups")
		err := s.userSynchronizer.Sync(context.Background(), s.report.SourceUsersToAdd, s.report.SDMUnmatchingUsers)
		if err != nil {
			return errors.New("error synchronizing users: " + err.Error())
		}
		err = s.groupSynchronizer.Sync(context.Background(), s.report.SourceUserGroupsToAdd, s.report.SourceMatchingUserGroups, s.report.SDMUnmatchingGroups)
		if err != nil {
			return errors.New("error synchronizing groups: " + err.Error())
		}
		s.report.Complete = time.Now()
	}
	log.Printf("%d SDM users in IdP Source", len(s.report.SourceMatchingUsers))
	log.Printf("%d SDM users not in IdP Source", len(s.report.SDMUnmatchingUsers))
	log.Printf("%d SDM groups in IdP Source", len(s.report.SourceMatchingUserGroups))
	log.Printf("%d SDM groups not in IdP Source", len(s.report.SDMUnmatchingGroups))
	log.Println(s.report.String())
	return nil
}

func (rpt *Report) String() string {
	if !*flags.JsonFlag {
		return rpt.short()
	}
	out, err := json.MarshalIndent(rpt, "", "\t")
	if err != nil {
		return fmt.Sprintf("error building JSON report: %s\n\n%s", err, rpt.short())
	}
	return string(out)
}

func (rpt *Report) short() string {
	return fmt.Sprintf("%d IdP users, %d strongDM users in IdP, %d strongDM groups in IdP\n",
		len(rpt.SourceUsers), len(rpt.SourceMatchingUsers), len(rpt.SourceMatchingUserGroups))
}
