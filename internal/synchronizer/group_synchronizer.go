package synchronizer

import (
	"context"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"strings"
)

var errorSign = fmt.Sprintf("\033[31mx\033[0m")

type GroupSynchronizer struct {
	report  *Report
	retrier Retrier
}

func newGroupSynchronizer(report *Report, retrier Retrier) *GroupSynchronizer {
	return &GroupSynchronizer{
		report:  report,
		retrier: retrier,
	}
}

func (sync *GroupSynchronizer) Sync(ctx context.Context, snk sink.BaseSink) error {
	sync.retrier.setEntityScope(GroupScope)
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
		err = sync.deleteMissingGroups(ctx, snk)
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
	var missingGroups []*sink.GroupRow = sync.getMissingGroups()
	var existentGroups []*sink.GroupRow
	var groupsWithUpdatedData []*sink.GroupRow
	sync.enrichGroupMembers()
	for _, idpGroup := range sync.report.IdPUserGroups {
		var found bool
		var isOutdated bool
		var sinkID string
		sinkGroup := groupSourceToGroupSink(idpGroup)
		if found, isOutdated, sinkID = sync.groupExistsInSink(idpGroup); !found {
			newGroups = append(newGroups, sinkGroup)
			continue
		}
		sinkGroup.ID = sinkID
		if isOutdated {
			groupsWithUpdatedData = append(groupsWithUpdatedData, sinkGroup)
		}
		existentGroups = append(existentGroups, sinkGroup)
	}
	return newGroups, missingGroups, existentGroups, groupsWithUpdatedData
}

func (sync *GroupSynchronizer) groupExistsInSink(idpGroup *source.UserGroup) (bool, bool, string) {
	var found, isOutdated bool
	var sinkID string
	for _, group := range sync.report.SinkGroups {
		if found = formatSourceGroupName(idpGroup.DisplayName) == group.DisplayName; found {
			sinkID = group.ID
			isOutdated = groupHasOutdatedData(idpGroup, group)
			break
		}
	}
	return found, isOutdated, sinkID
}

func (sync *GroupSynchronizer) getMissingGroups() []*sink.GroupRow {
	var missingGroups []*sink.GroupRow
	var mappedGroups = map[string]bool{}
	for _, group := range sync.report.IdPUserGroups {
		mappedGroups[formatSourceGroupName(group.DisplayName)] = true
	}
	for _, group := range sync.report.SinkGroups {
		if _, ok := mappedGroups[group.DisplayName]; !ok {
			missingGroups = append(missingGroups, group)
		}
	}
	return missingGroups
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
				idpGroup.Members[idx].SDMObjectID = usersMappedByUsername[member.Email].User.ID
			}
		}
	}
}

func (sync *GroupSynchronizer) createGroups(ctx context.Context, snk sink.BaseSink, sinkGroups []*sink.GroupRow) error {
	for _, group := range sinkGroups {
		err := sync.retrier.Run(func() error {
			members, notRegisteredMembers := getValidAndUnregisteredMembers(group)
			if len(notRegisteredMembers) > 0 {
				informAvoidedMembers(notRegisteredMembers, group.DisplayName)
			}
			group.Members = members
			response, err := snk.CreateGroup(ctx, group)
			if err != nil {
				return err
			} else if response != nil {
				fmt.Println(createSign, "Group created:", response.DisplayName)
				if len(response.Members) > 0 {
					fmt.Println("\t", createSign, "Members:")
					for _, member := range response.Members {
						fmt.Printf("\t\t %s %s\n", createSign, member.Email)
					}
				}
			}
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
		err := sync.retrier.Run(func() error {
			members, notRegisteredMembers := getValidAndUnregisteredMembers(group)
			if len(notRegisteredMembers) > 0 {
				informAvoidedMembers(notRegisteredMembers, group.DisplayName)
			}
			if len(members) == 0 {
				fmt.Fprintf(os.Stderr, "All the users that were planned to add to the group \"%s\" weren't registered. Skipping...", group.DisplayName)
				return nil
			} else if !sync.hasNewMembers(group.DisplayName, members) {
				fmt.Fprintf(os.Stderr, "All the users that were planned to add and were registered are already in the Group \"%s\". Skipping...\n", group.DisplayName)
				return nil
			}
			group.Members = members
			err := snk.ReplaceGroupMembers(ctx, group)
			if err != nil {
				return err
			}
			fmt.Println(updateSign, "Group updated:", formatSourceGroupName(group.DisplayName))
			if len(group.Members) > 0 {
				fmt.Println("\t", updateSign, "Members:")
				for _, member := range group.Members {
					if member.SDMObjectID == "" {
						continue
					}
					fmt.Printf("\t\t %s %s\n", updateSign, member.Email)
				}
			}
			return nil
		}, "updating group members")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) deleteMissingGroups(ctx context.Context, snk sink.BaseSink) error {
	for _, group := range sync.report.SinkGroupsNotInIdP {
		err := sync.retrier.Run(func() error {
			err := snk.DeleteGroup(ctx, group)
			if err != nil {
				return err
			}
			fmt.Println(deleteSign, "Group deleted:", group.DisplayName)
			return nil
		}, "deleting a group")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *GroupSynchronizer) hasNewMembers(name string, members []*sink.GroupMember) bool {
	var sinkGroup *sink.GroupRow
	for _, group := range sync.report.SinkGroups {
		if group.DisplayName == formatSourceGroupName(name) {
			sinkGroup = group
			break
		}
	}
	var foundMembersCount int
	for _, member := range members {
		for _, sinkGroupMember := range sinkGroup.Members {
			if found := sinkGroupMember.Email == member.Email; found {
				foundMembersCount++
			}
		}
	}
	return foundMembersCount != len(members)
}

func formatSourceGroupName(name string) string {
	orgUnits := strings.Split(name, "/")
	if len(orgUnits) == 0 {
		return ""
	}
	return strings.Join(orgUnits[1:], "_")
}

func getValidAndUnregisteredMembers(group *sink.GroupRow) ([]*sink.GroupMember, []*sink.GroupMember) {
	var validMembers []*sink.GroupMember
	var notRegisteredMembers []*sink.GroupMember
	for _, member := range group.Members {
		if member.SDMObjectID == "" {
			notRegisteredMembers = append(notRegisteredMembers, member)
			continue
		}
		validMembers = append(validMembers, member)
	}
	return validMembers, notRegisteredMembers
}

func informAvoidedMembers(members []*sink.GroupMember, groupName string) {
	var emailList []string
	for _, member := range members {
		emailList = append(emailList, member.Email)
	}
	if len(emailList) > 0 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s The member(s): %s won't be added in the %s group because an error occurred registering them.", errorSign, strings.Join(emailList, ", "), groupName))
	}
}
