package synchronizer

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
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
		for _, group := range sdmGroups {
			if idpGroup.DisplayName == group.DisplayName {
				found = true
				sinkObjectID = group.ID
				break
			}
		}
		sync.enrichGroupMembers()
		idpGroup.SDMObjectID = sinkObjectID
		if !found {
			newGroups = append(newGroups, idpGroup)
		} else {
			existentGroups = append(existentGroups, idpGroup)
		}
	}
	return newGroups, disjointedGroups, existentGroups
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
				idpGroup.Members[idx].SDMObjectID = usersMappedByUsername[member.Email].User.ID
			}
		}
	}
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, snk sink.BaseSink, sinkGroups []*source.UserGroup) error {
	for _, group := range sinkGroups {
		err := safeRetry(sync.rateLimiter, func() error {
			sdmGroup := groupSourceToGroupSink(group)
			response, err := snk.CreateGroup(ctx, sdmGroup)
			if err != nil {
				return err
			}
			fmt.Println("+ Group created:", response.DisplayName)
			return nil
		}, "creating a group")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) updateGroupMembers(ctx context.Context, snk sink.BaseSink) error {
	for _, sourceGroup := range sync.report.IdPUserGroupsInSink {
		err := safeRetry(sync.rateLimiter, func() error {
			sinkGroup := groupSourceToGroupSink(sourceGroup)
			err := snk.ReplaceGroupMembers(ctx, sinkGroup)
			if err != nil {
				return err
			}
			fmt.Println("~ Group updated:", sinkGroup.DisplayName)
			fmt.Println("\t\t~ Members:")
			for _, member := range sinkGroup.Members {
				fmt.Println("\t\t\t~", member.Email)
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
			fmt.Println("- Group deleted:", group.DisplayName)
			return nil
		}, "deleting a group")
		if err != nil {
			return err
		}
	}
	return nil
}
