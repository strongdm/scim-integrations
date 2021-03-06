package sdmscim

import (
	"context"
	"fmt"
	"scim-integrations/internal/sink"

	scimmodels "github.com/strongdm/scimsdk/models"
)

func (s *sinkSDMSCIMImpl) FetchGroups(ctx context.Context) ([]*sink.GroupRow, error) {
	iterator := s.client.Groups().List(ctx, nil)
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
	sdmMembers := sinkGroupMembersToSDMSCIM(group.Members)
	response, err := s.client.Groups().Create(ctx, scimmodels.CreateGroupBody{
		DisplayName: group.DisplayName,
		Members:     sdmMembers,
	})
	if err != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error occurred creating the group \"%s\": %v", group.DisplayName, err))
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
