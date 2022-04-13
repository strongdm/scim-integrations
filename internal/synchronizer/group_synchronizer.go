package synchronizer

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"strings"
)

type GroupSynchronizer struct {
	report      *Report
	rateLimiter *RateLimiter
}

func NewGroupSynchronizer(report *Report, rateLimiter *RateLimiter) *GroupSynchronizer {
	return &GroupSynchronizer{
		report:      report,
		rateLimiter: rateLimiter,
	}
}

func (sync *GroupSynchronizer) Sync(ctx context.Context, snk sink.BaseSink) error {
	sync.rateLimiter.Start()
	err := sync.EnrichReport(snk)
	if err != nil {
		return err
	}
	err = sync.createGroups(ctx, snk, sync.report.IdPUserGroupsToAdd)
	if err != nil {
		return err
	}
	err = sync.EnrichReport(snk)
	if err != nil {
		return err
	}
	err = sync.updateGroupMembers(ctx, snk)
	if err != nil {
		return err
	}
	if *flags.DeleteGroupsNotInIdPFlag {
		err = sync.deleteDisjointedGroups(ctx, snk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) EnrichReport(snk sink.BaseSink) error {
	if len(sync.report.SinkGroups) == 0 {
		sdmGroups, err := snk.FetchGroups(context.Background())
		if err != nil {
			return err
		}
		sync.report.SinkGroups = sdmGroups
	}
	newGroups, groupsNotInIdP, existentGroups, groupsWithUpdatedData := sync.removeSDMGroupsIntersection()
	sync.report.IdPUserGroupsToAdd = newGroups
	sync.report.IdPUserGroupsToUpdate = groupsWithUpdatedData
	sync.report.IdPUserGroupsInSink = existentGroups
	sync.report.SinkGroupsNotInIdP = groupsNotInIdP
	return nil
}

func (sync *GroupSynchronizer) removeSDMGroupsIntersection() ([]*sink.GroupRow, []*sink.GroupRow, []*sink.GroupRow, []*sink.GroupRow) {
	var newGroups []*sink.GroupRow
	var disjointedGroups []*sink.GroupRow
	var existentGroups []*sink.GroupRow
	var sdmGroups = sync.report.SinkGroups
	var groupsWithUpdatedData []*sink.GroupRow
	var mappedGroups = map[string]bool{}
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
		var isOutdated bool
		for _, group := range sdmGroups {
			if formatSourceGroupName(idpGroup.DisplayName) == group.DisplayName {
				found = true
				sinkObjectID = group.ID
				isOutdated = groupHasOutdatedData(idpGroup, group)
				break
			}
		}
		sync.enrichGroupMembers()
		idpGroup.SDMObjectID = sinkObjectID
		sinkGroup := groupSourceToGroupSink(idpGroup)
		if !found {
			newGroups = append(newGroups, sinkGroup)
		} else {
			if isOutdated {
				groupsWithUpdatedData = append(groupsWithUpdatedData, sinkGroup)
			}
			existentGroups = append(existentGroups, sinkGroup)
		}
	}
	return newGroups, disjointedGroups, existentGroups, groupsWithUpdatedData
}

func groupHasOutdatedData(idpGroup *source.UserGroup, sdmGroup *sink.GroupRow) bool {
	var isOutdated bool
	for _, idpGroupMember := range idpGroup.Members {
		var idpMemberIsInSDM bool
		for _, sdmGroupMember := range sdmGroup.Members {
			if idpGroupMember.Email == sdmGroupMember.Email {
				idpMemberIsInSDM = true
				break
			}
		}
		if !idpMemberIsInSDM {
			isOutdated = true
			break
		}
	}
	return isOutdated
}

func (sync *GroupSynchronizer) enrichGroupMembers() {
	users := append(sync.report.IdPUsersToAdd, sync.report.IdPUsersInSink...)
	usersMappedByUsername := map[string]*sink.UserRow{}
	for _, sdmUser := range users {
		usersMappedByUsername[sdmUser.User.UserName] = sdmUser
	}
	for _, idpGroup := range sync.report.IdPUserGroups {
		for idx, member := range idpGroup.Members {
			if _, ok := usersMappedByUsername[member.Email]; ok {
				idpGroup.Members[idx].SDMObjectID = usersMappedByUsername[member.Email].User.SinkID
			}
		}
	}
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, snk sink.BaseSink, sinkGroups []*sink.GroupRow) error {
	for _, group := range sinkGroups {
		err := safeRetry(sync.rateLimiter, func() error {
			response, err := snk.CreateGroup(ctx, group)
			if err != nil {
				return err
			}
			fmt.Println(createSign, " Group created:", response.DisplayName)
			return nil
		}, "creating a group")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) updateGroupMembers(ctx context.Context, snk sink.BaseSink) error {
	for _, group := range sync.report.IdPUserGroupsToUpdate {
		err := safeRetry(sync.rateLimiter, func() error {
			err := snk.ReplaceGroupMembers(ctx, group)
			if err != nil {
				return err
			}
			fmt.Println(updateSign, " Group updated:", formatSourceGroupName(group.DisplayName))
			fmt.Println("\t", updateSign, " Members:")
			for _, member := range group.Members {
				fmt.Println("\t\t", updateSign, member.Email)
			}
			return nil
		}, "updating group members")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) deleteDisjointedGroups(ctx context.Context, snk sink.BaseSink) error {
	for _, group := range sync.report.SinkGroupsNotInIdP {
		err := safeRetry(sync.rateLimiter, func() error {
			err := snk.DeleteGroup(ctx, group)
			if err != nil {
				return err
			}
			fmt.Println(deleteSign, " Group deleted:", group.DisplayName)
			return nil
		}, "deleting a group")
		if err != nil {
			return err
		}
	}
	return nil
}

func formatSourceGroupName(name string) string {
	orgUnits := strings.Split(name, "/")
	return strings.Join(orgUnits[1:], "_")
}
