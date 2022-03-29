package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink/sdmscim"
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

func (sync *GroupSynchronizer) Sync(ctx context.Context) error {
	err := sync.createGroups(ctx, sync.report.IdPUserGroupsToAdd)
	if err != nil {
		return err
	}
	err = sync.replaceGroupMembers(ctx, sync.report.IdPUserGroupsInSDM)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if *flags.DeleteGroupsNotInIdPFlag {
		err = sync.deleteDisjointedGroups(ctx, sync.report.SDMGroupsNotInIdP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) EnrichReport() error {
	sdmGroups, err := sdmscim.FetchGroups(context.Background())
	if err != nil {
		return err
	}
	var existentGroups []source.UserGroup
	var newGroups []source.UserGroup
	for _, iGroup := range sync.report.IdPUserGroups {
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
	groupsNotInIdP := removeSDMGroupsIntersection(sdmGroups, sync.report.IdPUserGroups)
	sync.report.IdPUserGroupsToAdd = newGroups
	sync.report.IdPUserGroupsInSDM = existentGroups
	sync.report.SDMGroupsNotInIdP = groupsNotInIdP
	return nil
}

func removeSDMGroupsIntersection(sdmGroups []sdmscim.GroupRow, existentIdPGroups []source.UserGroup) []sdmscim.GroupRow {
	var disjointedGroups []sdmscim.GroupRow
	mappedGroups := map[string]bool{}
	for _, group := range existentIdPGroups {
		mappedGroups[group.DisplayName] = true
	}
	for _, group := range sdmGroups {
		if _, ok := mappedGroups[group.DisplayName]; !ok {
			disjointedGroups = append(disjointedGroups, group)
		}
	}
	return disjointedGroups
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, sinkGroups []source.UserGroup) error {
	for _, group := range sinkGroups {
		_, err := sdmscim.CreateGroup(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) replaceGroupMembers(ctx context.Context, sinkGroups []source.UserGroup) error {
	for _, group := range sinkGroups {
		err := sdmscim.ReplaceGroupMembers(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) deleteDisjointedGroups(ctx context.Context, sdmGroups []sdmscim.GroupRow) error {
	for _, group := range sdmGroups {
		err := sdmscim.DeleteGroup(ctx, group.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
