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

func (sync *GroupSynchronizer) Sync(ctx context.Context) []error {
	errs := sync.createGroups(ctx, sync.report.IdPUserGroupsToAdd)
	replaceErrs := sync.replaceGroupMembers(ctx)
	errs = append(errs, replaceErrs...)
	if *flags.DeleteGroupsNotInIdPFlag {
		deleteErrs := sync.deleteDisjointedGroups(ctx)
		errs = append(errs, deleteErrs...)
	}
	return errs
}

func (sync *GroupSynchronizer) EnrichReport() error {
	sdmGroups, err := sdmscim.FetchGroups(context.Background())
	if err != nil {
		return err
	}
	newGroups, groupsNotInIdP, existentGroups := sync.removeSDMGroupsIntersection(sdmGroups)
	sync.report.IdPUserGroupsToAdd = newGroups
	sync.report.IdPUserGroupsInSDM = existentGroups
	sync.report.SDMGroupsNotInIdP = groupsNotInIdP
	return nil
}

func (sync *GroupSynchronizer) removeSDMGroupsIntersection(sdmGroups []*sdmscim.GroupRow) ([]*source.UserGroup, []*sdmscim.GroupRow, []*source.UserGroup) {
	var newGroups []*source.UserGroup
	var disjointedGroups []*sdmscim.GroupRow
	var existentGroups []*source.UserGroup
	mappedGroups := map[string]bool{}
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
		users := append(sync.report.IdPUsersToAdd, sync.report.IdPUsersInSDM...)
		for idx, member := range idpGroup.Members {
			for _, idpUser := range users {
				if member.Email == idpUser.UserName {
					idpGroup.Members[idx].ID = idpUser.SinkObjectID
				}
			}
		}
		if !found {
			newGroups = append(newGroups, idpGroup)
		} else {
			idpGroup.SinkObjectID = sinkObjectID
			existentGroups = append(existentGroups, idpGroup)
		}
	}
	return newGroups, disjointedGroups, existentGroups
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, sinkGroups []*source.UserGroup) []error {
	errs := []error{}
	for _, group := range sinkGroups {
		_, err := sdmscim.CreateGroup(ctx, group)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (sync *GroupSynchronizer) replaceGroupMembers(ctx context.Context) []error {
	errs := []error{}
	for _, group := range sync.report.IdPUserGroupsInSDM {
		err := sdmscim.ReplaceGroupMembers(ctx, group)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (sync *GroupSynchronizer) deleteDisjointedGroups(ctx context.Context) []error {
	errs := []error{}
	for _, group := range sync.report.SDMGroupsNotInIdP {
		err := sdmscim.DeleteGroup(ctx, group)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
