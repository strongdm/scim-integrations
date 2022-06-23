package sdmscim

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"strings"

	scimmodels "github.com/strongdm/scimsdk/models"
)

func (s *sinkSDMSCIMImpl) FetchGroups(ctx context.Context) ([]*sink.GroupRow, error) {
	iterator := s.client.Groups().List(ctx, &scimmodels.PaginationOptions{
		Filter: *flags.SDMGroupsQueryFlag,
	})
	var result []*sink.GroupRow
	for iterator.Next() {
		group := *iterator.Value()
		result = append(result, groupToSink(&group))
	}
	if iterator.Err() != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error occurred listing the SDM groups: %v", iterator.Err()))
	}
	return result, nil
}

func (s *sinkSDMSCIMImpl) CreateGroup(ctx context.Context, group *sink.GroupRow) (*sink.GroupRow, error) {
	groupName := formatGroupName(group.DisplayName)
	sdmMembers := sinkGroupMembersToSDMSCIM(group.Members)
	response, err := s.client.Groups().Create(ctx, scimmodels.CreateGroupBody{
		DisplayName: groupName,
		Members:     sdmMembers,
	})
	if err != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error occurred creating the group \"%s\": %v", groupName, err))
	}
	group.ID = response.ID
	return groupToSink(response), nil
}

func (s *sinkSDMSCIMImpl) ReplaceGroupMembers(ctx context.Context, group *sink.GroupRow) error {
	sdmMembers := sinkGroupMembersToSDMSCIM(group.Members)
	_, err := s.client.Groups().UpdateReplaceMembers(ctx, group.ID, sdmMembers)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error occurred replacing the %s group members: %v", group.DisplayName, err))
	}
	return nil
}

func (s *sinkSDMSCIMImpl) DeleteGroup(ctx context.Context, group *sink.GroupRow) error {
	_, err := s.client.Groups().Delete(ctx, group.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error occurred deleting the group \"%s\": %v", group.DisplayName, err))
	}
	return nil
}

func formatGroupName(orgUnitPath string) string {
	orgUnits := strings.Split(orgUnitPath, "/")
	if len(orgUnits) == 0 {
		return ""
	}
	if len(orgUnits) == 1 {
		return orgUnits[0]
	}
	groupName := strings.Join(orgUnits[1:], "_")
	return groupName
}
