package sync

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"

	"github.com/strongdm/scimsdk/scimsdk"
)

type GroupSynchronize struct {
	service *sink.SDMService
	report  *Report
}

func NewGroupSynchronize(service *sink.SDMService, report *Report) *GroupSynchronize {
	return &GroupSynchronize{
		service: service,
		report:  report,
	}
}

func (sync *GroupSynchronize) Sync(ctx context.Context, idpGroups []sink.SinkUserGroup) error {
	sdmGroups, err := sync.service.FetchGroups(ctx)
	if err != nil {
		return err
	}
	newGroups, existentGroups, unmatchingGroups := calculateSDMGroupsIntersection(sdmGroups, idpGroups)
	createdGroups, err := sync.createGroups(ctx, newGroups)
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
	sync.report.SDMGroupsInIdP = existentGroups
	sync.report.SDMGroupsNotInIdP = unmatchingGroups
	sync.report.SDMNewGroups = createdGroups
	return nil
}

func calculateSDMGroupsIntersection(sdmGroups []scimsdk.Group, idpGroups []sink.SinkUserGroup) ([]sink.SinkUserGroup, []sink.SinkUserGroup, []scimsdk.Group) {
	var existentGroups []sink.SinkUserGroup
	var newGroups []sink.SinkUserGroup
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
	return newGroups, existentGroups, unmatchingGroups
}

func removeSDMGroupsIntersection(sdmGroups []scimsdk.Group, existentIdPGroups []sink.SinkUserGroup) []scimsdk.Group {
	var unmatchingGroups []scimsdk.Group
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

func (sync *GroupSynchronize) createGroups(ctx context.Context, sinkGroups []sink.SinkUserGroup) ([]scimsdk.Group, error) {
	var finalGroups []scimsdk.Group
	for _, group := range sinkGroups {
		group, err := sync.service.CreateGroup(ctx, group)
		if err != nil {
			return nil, err
		}
		finalGroups = append(finalGroups, *group)
	}
	return finalGroups, nil
}

func (sync *GroupSynchronize) replaceGroupMembers(ctx context.Context, sinkGroups []sink.SinkUserGroup) error {
	for _, group := range sinkGroups {
		err := sync.service.ReplaceGroupMembers(ctx, group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronize) deleteUnmatchingGroups(ctx context.Context, sdmGroups []scimsdk.Group) error {
	for _, group := range sdmGroups {
		err := sync.service.DeleteGroup(ctx, group.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
