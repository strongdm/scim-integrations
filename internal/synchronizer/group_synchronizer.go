package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

type GroupSynchronizer struct {
	service *sink.SDMSink
	report  *Report
}

func NewGroupSynchronize(service *sink.SDMSink, report *Report) *GroupSynchronizer {
	return &GroupSynchronizer{
		service: service,
		report:  report,
	}
}

func (sync *GroupSynchronizer) Sync(ctx context.Context, newGroups []source.SourceUserGroup, existentGroups []source.SourceUserGroup, unmatchingGroups []sink.SDMGroupRow) error {
	err := sync.createGroups(ctx, newGroups)
	if err != nil {
		return err
	}
	err = sync.replaceGroupMembers(ctx, existentGroups)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if *flags.DeleteUnmatchingGroupsFlag {
		err = sync.deleteUnmatchingGroups(ctx, unmatchingGroups)
		if err != nil {
			return err
		}
	}
	return nil
}

func (synchronizer *GroupSynchronizer) SyncGroupsData(idpGroups []source.SourceUserGroup) error {
	sdmGroups, err := synchronizer.service.FetchGroups(context.Background())
	if err != nil {
		return err
	}
	var existentGroups []source.SourceUserGroup
	var newGroups []source.SourceUserGroup
	for _, iGroup := range idpGroups {
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
	unmatchingGroups := removeSDMGroupsIntersection(sdmGroups, idpGroups)
	synchronizer.report.SourceUserGroupsToAdd = newGroups
	synchronizer.report.SourceMatchingUserGroups = existentGroups
	synchronizer.report.SDMUnmatchingGroups = unmatchingGroups
	return nil
}

func removeSDMGroupsIntersection(sdmGroups []sink.SDMGroupRow, existentIdPGroups []source.SourceUserGroup) []sink.SDMGroupRow {
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

func (sync *GroupSynchronizer) createGroups(ctx context.Context, sinkGroups []source.SourceUserGroup) error {
	for _, group := range sinkGroups {
		_, err := sync.service.CreateGroup(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) replaceGroupMembers(ctx context.Context, sinkGroups []source.SourceUserGroup) error {
	for _, group := range sinkGroups {
		err := sync.service.ReplaceGroupMembers(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) deleteUnmatchingGroups(ctx context.Context, sdmGroups []sink.SDMGroupRow) error {
	for _, group := range sdmGroups {
		err := sync.service.DeleteGroup(ctx, group.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
