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

var errorSign = fmt.Sprintf("\033[31mx\033[0m")

type sinkSDMSCIMImpl struct {
	client scimsdk.Client
}

func NewSinkSDMSCIM() *sinkSDMSCIMImpl {
	return &sinkSDMSCIMImpl{NewSDMSCIMClient()}
}

func NewSDMSCIMClient() scimsdk.Client {
	return scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
}

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
		return nil, fmt.Errorf(formatErrorMessage("An error was occurred creating the user \"%s\": %v", row.User.UserName, err))
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
		return fmt.Errorf(formatErrorMessage("An error was occurred updating the user \"%s\": %v", row.User.UserName, err))
	}
	return nil
}

func (s *sinkSDMSCIMImpl) DeleteUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Delete(ctx, row.User.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred deleting the user \"%s\": %v", row.User.UserName, err))
	}
	return nil
}

func (s *sinkSDMSCIMImpl) FetchGroups(ctx context.Context) ([]*sink.GroupRow, error) {
	iterator := s.client.Groups().List(ctx, nil)
	var result []*sink.GroupRow
	for iterator.Next() {
		group := *iterator.Value()
		result = append(result, groupToSink(&group))
	}
	if iterator.Err() != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error was occurred listing the SDM groups: %v", iterator.Err()))
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
		return nil, fmt.Errorf(formatErrorMessage("An error was occurred creating the group \"%s\": %v", groupName, err))
	}
	group.ID = response.ID
	return groupToSink(response), nil
}

func (s *sinkSDMSCIMImpl) ReplaceGroupMembers(ctx context.Context, group *sink.GroupRow) error {
	sdmMembers := sinkGroupMembersToSDMSCIM(group.Members)
	_, err := s.client.Groups().UpdateReplaceMembers(ctx, group.ID, sdmMembers)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred replacing the %s group members: %v", group.DisplayName, err))
	}
	return nil
}

func (s *sinkSDMSCIMImpl) DeleteGroup(ctx context.Context, group *sink.GroupRow) error {
	_, err := s.client.Groups().Delete(ctx, group.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred deleting the group \"%s\": %v", group.DisplayName, err))
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

func formatGroupName(orgUnitPath string) string {
	orgUnits := strings.Split(orgUnitPath, "/")
	if len(orgUnits) == 0 {
		return ""
	}
	return strings.Join(orgUnits[1:], "_")
}

func formatErrorMessage(message string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", errorSign, fmt.Sprintf(message, args...))
}
