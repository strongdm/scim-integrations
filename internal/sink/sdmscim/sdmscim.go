package sdmscim

import (
	"context"
	"fmt"
	"os"
	"scim-integrations/internal/sink"
	"strings"

	"github.com/strongdm/scimsdk"
	scimmodels "github.com/strongdm/scimsdk/models"
)

type SinkSDMSCIMImpl struct {
	client scimsdk.Client
}

func NewSinkSDMSCIMImpl() *SinkSDMSCIMImpl {
	return &SinkSDMSCIMImpl{NewSDMSCIMClient()}
}

func NewSDMSCIMClient() scimsdk.Client {
	client := scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
	return client
}

func (s *SinkSDMSCIMImpl) FetchUsers(ctx context.Context) ([]*sink.UserRow, error) {
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

func separateGroupsByUser(groups []*sink.GroupRow) map[string][]*sink.GroupRow {
	userGroups := map[string][]*sink.GroupRow{}
	for _, group := range groups {
		for _, member := range group.Members {
			if userGroups[member.ID] == nil {
				userGroups[member.ID] = []*sink.GroupRow{group}
			} else {
				userGroups[member.ID] = append(userGroups[member.ID], group)
			}
		}
	}
	return userGroups
}

func (s *SinkSDMSCIMImpl) CreateUser(ctx context.Context, row *sink.UserRow) (*sink.UserRow, error) {
	response, err := s.client.Users().Create(ctx, scimmodels.CreateUser{
		UserName:   row.User.UserName,
		GivenName:  row.User.GivenName,
		FamilyName: row.User.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("An error was occurred creating the user \"%s\": %v", row.User.UserName, err)
	}
	row.User.ID = response.ID
	return userToSink(response, nil), nil
}

func (s *SinkSDMSCIMImpl) ReplaceUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Replace(ctx, row.User.ID, scimmodels.ReplaceUser{
		UserName:   row.User.UserName,
		GivenName:  row.User.GivenName,
		FamilyName: row.User.FamilyName,
		Active:     row.User.Active,
	})
	if err != nil {
		return fmt.Errorf("An error was occurred updating the user \"%s\": %v", row.User.UserName, err)
	}
	return nil
}

func (s *SinkSDMSCIMImpl) DeleteUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Delete(ctx, row.User.ID)
	if err != nil {
		return fmt.Errorf("An error was occurred deleting the user \"%s\": %v", row.User.UserName, err)
	}
	return nil
}

func (s *SinkSDMSCIMImpl) FetchGroups(ctx context.Context) ([]*sink.GroupRow, error) {
	iterator := s.client.Groups().List(ctx, nil)
	var result []*sink.GroupRow
	for iterator.Next() {
		group := *iterator.Value()
		result = append(result, groupToSink(&group))
	}
	if iterator.Err() != nil {
		return nil, fmt.Errorf("An error was occurred listing the SDM groups: %v", iterator.Err())
	}
	return result, nil
}

func (s *SinkSDMSCIMImpl) CreateGroup(ctx context.Context, group *sink.GroupRow) (*sink.GroupRow, error) {
	groupName := getGroupName(group.DisplayName)
	response, err := s.client.Groups().Create(ctx, scimmodels.CreateGroupBody{
		DisplayName: groupName,
		Members:     sinkGroupMemberListToSDMSCIM(group.Members),
	})
	if err != nil {
		return nil, fmt.Errorf("An error was occurred creating the group \"%s\": %v", groupName, err)
	}
	group.ID = response.ID
	return groupToSink(response), nil
}

func (s *SinkSDMSCIMImpl) ReplaceGroupMembers(ctx context.Context, group *sink.GroupRow) error {
	_, err := s.client.Groups().UpdateReplaceMembers(ctx, group.ID, sinkGroupMemberListToSDMSCIM(group.Members))
	if err != nil {
		return fmt.Errorf("An error was occurred replacing the %s group members: %v", group.DisplayName, err)
	}
	return nil
}

func (s *SinkSDMSCIMImpl) DeleteGroup(ctx context.Context, group *sink.GroupRow) error {
	_, err := s.client.Groups().Delete(ctx, group.ID)
	if err != nil {
		return fmt.Errorf("An error was occurred deleting the group \"%s\": %v", group.DisplayName, err)
	}
	return nil
}

func getGroupName(orgUnitPath string) string {
	orgUnits := strings.Split(orgUnitPath, "/")
	return orgUnits[len(orgUnits)-1]
}

func treatErr(opErr error, description string) error {
	if opErr != nil {
		err := fmt.Errorf("%v: %v", description, opErr)
		if sink.ErrorIsUnexpected(opErr) {
			return err
		}
		fmt.Fprintln(os.Stderr, err)
	}
	return nil
}
