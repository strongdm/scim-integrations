package sdmscim

import (
	"context"
	"fmt"
	"scim-integrations/internal/sink"

	scimmodels "github.com/strongdm/scimsdk/models"
)

func (s *sinkSDMSCIMImpl) FetchUsers(ctx context.Context) ([]*sink.UserRow, error) {
	groups, err := s.FetchGroups(ctx)
	if err != nil {
		return nil, err
	}
	userGroups := separateGroupsByUser(groups)
	iterator := s.client.Users().List(ctx, nil)
	users, err := usersWithGroupsToSink(iterator, userGroups)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *sinkSDMSCIMImpl) CreateUser(ctx context.Context, row *sink.UserRow) (*sink.UserRow, error) {
	response, err := s.client.Users().Create(ctx, scimmodels.CreateUser{
		UserName:   row.User.UserName,
		GivenName:  row.User.GivenName,
		FamilyName: row.User.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error occurred creating the user \"%s\": %v", row.User.UserName, err))
	}
	row.User.ID = response.ID
	return userToSink(response, nil), nil
}

func (s *sinkSDMSCIMImpl) ReplaceUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Replace(ctx, row.User.ID, scimmodels.ReplaceUser{
		UserName:   row.User.UserName,
		GivenName:  row.User.GivenName,
		FamilyName: row.User.FamilyName,
		Active:     row.User.Active,
	})
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error occurred updating the user \"%s\": %v", row.User.UserName, err))
	}
	return nil
}

func (s *sinkSDMSCIMImpl) DeleteUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Delete(ctx, row.User.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error occurred deleting the user \"%s\": %v", row.User.UserName, err))
	}
	return nil
}

func separateGroupsByUser(groups []*sink.GroupRow) map[string][]*sink.GroupRow {
	groupRows := map[string][]*sink.GroupRow{}
	for _, group := range groups {
		for _, member := range group.Members {
			if groupRows[member.ID] == nil {
				groupRows[member.ID] = []*sink.GroupRow{group}
			} else {
				groupRows[member.ID] = append(groupRows[member.ID], group)
			}
		}
	}
	return groupRows
}
