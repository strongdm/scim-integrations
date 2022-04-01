package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
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

func (sync *GroupSynchronizer) Sync(ctx context.Context, errCh chan error) {
	err := sync.EnrichReport()
	if err != nil {
		errCh <- err
	}
	sync.createGroups(ctx, sync.report.IdPUserGroupsToAdd, errCh)
	err = sync.EnrichReport()
	if err != nil {
		errCh <- err
	}
	sync.replaceGroupMembers(ctx, errCh)
	if *flags.DeleteGroupsNotInIdPFlag {
		sync.deleteDisjointedGroups(ctx, errCh)
	}
}

func (sync *GroupSynchronizer) EnrichReport() error {
	if len(sync.report.SinkGroups) == 0 {
		sdmGroups, err := sdmscim.FetchGroups(context.Background())
		if err != nil {
			return err
		}
		sync.report.SinkGroups = sdmGroups
	}
	newGroups, groupsNotInIdP, existentGroups := sync.removeSDMGroupsIntersection()
	sync.report.IdPUserGroupsToAdd = newGroups
	sync.report.IdPUserGroupsInSink = existentGroups
	sync.report.SinkGroupsNotInIdP = groupsNotInIdP
	return nil
}

func (sync *GroupSynchronizer) removeSDMGroupsIntersection() ([]*source.UserGroup, []*sink.GroupRow, []*source.UserGroup) {
	var newGroups []*source.UserGroup
	var disjointedGroups []*sink.GroupRow
	var existentGroups []*source.UserGroup
	var sdmGroups = sync.report.SinkGroups
	var mappedGroups map[string]bool = map[string]bool{}
	for _, group := range sync.report.IdPUserGroups {
		mappedGroups[group.DisplayName] = true
	}
	for _, group := range sdmGroups {
		if _, ok := mappedGroups[group.DisplayName]; !ok {
			disjointedGroups = append(disjointedGroups, group)
		}
	}
	for _, idpGroup := range sync.report.IdPUserGroups {
		var found bool
		var sinkObjectID string
		for _, group := range sdmGroups {
			if idpGroup.DisplayName == group.DisplayName {
				found = true
				sinkObjectID = group.ID
				break
			}
		}
		sync.enrichGroupMembers()
		if !found {
			newGroups = append(newGroups, idpGroup)
		} else {
			idpGroup.SinkObjectID = sinkObjectID
			existentGroups = append(existentGroups, idpGroup)
		}
	}
	return newGroups, disjointedGroups, existentGroups
}

func (sync *GroupSynchronizer) enrichGroupMembers() {
	users := append(sync.report.IdPUsersToAdd, sync.report.IdPUsersInSink...)
	usersMappedByUsername := map[string]*source.User{}
	for _, idpUser := range users {
		usersMappedByUsername[idpUser.UserName] = idpUser
	}
	for _, idpGroup := range sync.report.IdPUserGroups {
		for idx, member := range idpGroup.Members {
			if _, ok := usersMappedByUsername[member.Email]; ok {
				idpGroup.Members[idx].ID = usersMappedByUsername[member.Email].SinkObjectID
			}
		}
	}
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, sinkGroups []*source.UserGroup, errCh chan error) {
	for _, group := range sinkGroups {
		_, err := sdmscim.CreateGroup(ctx, group)
		if err != nil {
			errCh <- err
		}
	}
}

func (sync *GroupSynchronizer) replaceGroupMembers(ctx context.Context, errCh chan error) {
	for _, group := range sync.report.IdPUserGroupsInSink {
		err := sdmscim.ReplaceGroupMembers(ctx, group)
		if err != nil {
			errCh <- err
		}
	}
}

func (sync *GroupSynchronizer) deleteDisjointedGroups(ctx context.Context, errCh chan error) {
	for _, group := range sync.report.SinkGroupsNotInIdP {
		err := sdmscim.DeleteGroup(ctx, group)
		if err != nil {
			errCh <- err
		}
	}
}
