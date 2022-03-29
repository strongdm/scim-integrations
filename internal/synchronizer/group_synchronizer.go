package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

type GroupSynchronizer struct {
	report *Report
}

func NewGroupSynchronizer(report *Report) *GroupSynchronizer {
	return &GroupSynchronizer{
		report: report,
	}
}

func (s *GroupSynchronizer) Sync(ctx context.Context) error {
	err := s.createGroups(ctx, s.report.IdPUserGroupsToAdd)
	if err != nil {
		return err
	}
	err = s.replaceGroupMembers(ctx, s.report.IdPUserGroupsInSDM)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if *flags.DeleteGroupsNotInIdPFlag {
		err = s.deleteUnmatchingGroups(ctx, s.report.SDMGroupsNotInIdP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupSynchronizer) EnrichReport() error {
	sdmGroups, err := sink.FetchGroups(context.Background())
	if err != nil {
		return err
	}
	var existentGroups []source.UserGroup
	var newGroups []source.UserGroup
	for _, iGroup := range s.report.IdPUserGroups {
		var found bool
		for _, group := range sdmGroups {
			if iGroup.DisplayName == group.DisplayName {
				found = true
				iGroup.ID = group.ID
				break
			}
		}
		if !found {
			newGroups = append(newGroups, iGroup)
		} else {
			existentGroups = append(existentGroups, iGroup)
		}
	}
	groupsNotInIdP := removeSDMGroupsIntersection(sdmGroups, s.report.IdPUserGroups)
	s.report.IdPUserGroupsToAdd = newGroups
	s.report.IdPUserGroupsInSDM = existentGroups
	s.report.SDMGroupsNotInIdP = groupsNotInIdP
	return nil
}

func removeSDMGroupsIntersection(sdmGroups []sink.SDMGroupRow, existentIdPGroups []source.UserGroup) []sink.SDMGroupRow {
	var unmatchingGroups []sink.SDMGroupRow
	mappedGroups := map[string]bool{}
	for _, group := range existentIdPGroups {
		mappedGroups[group.DisplayName] = true
	}
	for _, group := range sdmGroups {
		if _, ok := mappedGroups[group.DisplayName]; !ok {
			unmatchingGroups = append(unmatchingGroups, group)
		}
	}
	return unmatchingGroups
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, sinkGroups []source.UserGroup) error {
	for _, group := range sinkGroups {
		_, err := sink.CreateGroup(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) replaceGroupMembers(ctx context.Context, sinkGroups []source.UserGroup) error {
	for _, group := range sinkGroups {
		err := sink.ReplaceGroupMembers(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) deleteUnmatchingGroups(ctx context.Context, sdmGroups []sink.SDMGroupRow) error {
	for _, group := range sdmGroups {
		err := sink.DeleteGroup(ctx, group.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
