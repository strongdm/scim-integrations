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

func (s *SinkSDMSCIMImpl) CreateUser(ctx context.Context, row *sink.UserRow) (*sink.UserRow, error) {
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

func (s *SinkSDMSCIMImpl) ReplaceUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Replace(ctx, row.User.SinkID, scimmodels.ReplaceUser{
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

func (s *SinkSDMSCIMImpl) DeleteUser(ctx context.Context, row sink.UserRow) error {
	_, err := s.client.Users().Delete(ctx, row.User.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred deleting the user \"%s\": %v", row.User.UserName, err))
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
		return nil, fmt.Errorf(formatErrorMessage("An error was occurred listing the SDM groups: %v", iterator.Err()))
	}
	return result, nil
}

func (s *SinkSDMSCIMImpl) CreateGroup(ctx context.Context, group *sink.GroupRow) (*sink.GroupRow, error) {
	groupName := formatGroupName(group.DisplayName)
	members, notRegisteredMembers := sinkGroupMemberListToSDMSCIM(group.Members)
	if len(notRegisteredMembers) > 0 {
		informAvoidedMembers(notRegisteredMembers, group.DisplayName)
	}
	if len(members) == 0 {
		fmt.Fprintf(os.Stderr, "All the users that were planned to add to the group %s weren't registered. Skipping...", group.DisplayName)
		return nil, nil
	}
	response, err := s.client.Groups().Create(ctx, scimmodels.CreateGroupBody{
		DisplayName: groupName,
		Members:     members,
	})
	if err != nil {
		return nil, fmt.Errorf(formatErrorMessage("An error was occurred creating the group \"%s\": %v", groupName, err))
	}
	group.ID = response.ID
	return groupToSink(response), nil
}

func (s *SinkSDMSCIMImpl) ReplaceGroupMembers(ctx context.Context, group *sink.GroupRow) error {
	members, notRegisteredMembers := sinkGroupMemberListToSDMSCIM(group.Members)
	if len(notRegisteredMembers) > 0 {
		informAvoidedMembers(notRegisteredMembers, group.DisplayName)
	}
	if len(members) == 0 {
		fmt.Fprintf(os.Stderr, "All the users that were planned to add to the group \"%s\" weren't registered. Skipping...", group.DisplayName)
		return nil
	}
	_, err := s.client.Groups().UpdateReplaceMembers(ctx, group.ID, members)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred replacing the %s group members: %v", group.DisplayName, err))
	}
	return nil
}

func (s *SinkSDMSCIMImpl) DeleteGroup(ctx context.Context, group *sink.GroupRow) error {
	_, err := s.client.Groups().Delete(ctx, group.ID)
	if err != nil {
		return fmt.Errorf(formatErrorMessage("An error was occurred deleting the group \"%s\": %v", group.DisplayName, err))
	}
	return nil
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

func formatGroupName(orgUnitPath string) string {
	orgUnits := strings.Split(orgUnitPath, "/")
	if len(orgUnits) == 0 {
		return ""
	}
	return strings.Join(orgUnits[1:], "_")
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

func formatErrorMessage(message string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", errorSign, fmt.Sprintf(message, args...))
}
