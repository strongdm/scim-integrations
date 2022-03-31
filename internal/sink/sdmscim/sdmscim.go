package sdmscim

import (
	"context"
	"errors"
	"fmt"
	"os"
	"scim-integrations/internal/source"

	"github.com/strongdm/scimsdk/scimsdk"
)

func newSDMSCIMClient() *scimsdk.Client {
	client := scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
	return client
}

func FetchUsers(ctx context.Context) ([]*UserRow, error) {
	groups, err := FetchGroups(ctx)
	if err != nil {
		return nil, err
	}
	userGroups := separateGroupsByUser(groups)
	iterator := internalSCIMSDKUsersList(ctx, nil)
	users, err := usersWithGroupsToSink(iterator, userGroups)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func separateGroupsByUser(groups []*GroupRow) map[string][]*GroupRow {
	userGroups := map[string][]*GroupRow{}
	for _, group := range groups {
		for _, member := range group.Members {
			if userGroups[member.ID] == nil {
				userGroups[member.ID] = []*GroupRow{group}
			} else {
				userGroups[member.ID] = append(userGroups[member.ID], group)
			}
		}
	}
	return userGroups
}

func CreateUser(ctx context.Context, user *source.User) (*UserRow, error) {
	response, err := internalSCIMSDKUsersCreate(ctx, scimsdk.CreateUser{
		UserName:   user.UserName,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error was occurred creating the user \"%s\": %v", user.UserName, err))
	}
	user.SinkObjectID = response.ID
	return userToSink(response), nil
}

func ReplaceUser(ctx context.Context, user source.User) error {
	_, err := internalSCIMSDKUsersReplace(ctx, user)
	if err != nil {
		return errors.New(fmt.Sprintf("An error was occurred updating the user \"%s\": %v", user.UserName, err))
	}
	return nil
}

func DeleteUser(ctx context.Context, user UserRow) error {
	_, err := internalSCIMSDKUsersDelete(ctx, user.ID)
	if err != nil {
		return errors.New(fmt.Sprintf("An error was occurred deleting the user \"%s\": %v", user.UserName, err))
	}
	return nil
}

func FetchGroups(ctx context.Context) ([]*GroupRow, error) {
	iterator := internalSCIMSDKGroupsList(ctx, nil)
	var result []*GroupRow
	for iterator.Next() {
		group := *iterator.Value()
		result = append(result, groupToSink(&group))
	}
	if iterator.Err() != nil {
		return nil, errors.New(fmt.Sprintf("An error was occurred listing the SDM groups: %v", iterator.Err()))
	}
	return result, nil
}

func CreateGroup(ctx context.Context, group *source.UserGroup) (*GroupRow, error) {
	response, err := internalSCIMSDKGroupsCreate(ctx, scimsdk.CreateGroupBody{
		DisplayName: group.DisplayName,
		Members:     group.Members,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error was occurred creating the group \"%s\": %v", group.DisplayName, err))
	}
	return groupToSink(response), nil
}

func ReplaceGroupMembers(ctx context.Context, group *source.UserGroup) error {
	_, err := internalSCIMSDKGroupsUpdateReplaceMembers(ctx, group.SinkObjectID, group.Members)
	if err != nil {
		return errors.New(fmt.Sprintf("An error was occurred replacing the %s group members: %v", group.DisplayName, err))
	}
	return nil
}

func DeleteGroup(ctx context.Context, group *GroupRow) error {
	_, err := internalSCIMSDKGroupsDelete(ctx, group.ID)
	if err != nil {
		return errors.New(fmt.Sprintf("An error was occurred deleting the group \"%s\": %v", group.DisplayName, err))
	}
	return nil
}

func internalSCIMSDKUsersList(ctx context.Context, paginationOpts *scimsdk.PaginationOptions) scimsdk.UserIterator {
	client := newSDMSCIMClient()
	return client.Users().List(ctx, paginationOpts)
}

func internalSCIMSDKUsersCreate(ctx context.Context, user scimsdk.CreateUser) (*scimsdk.User, error) {
	client := newSDMSCIMClient()
	return client.Users().Create(ctx, user)
}

func internalSCIMSDKUsersReplace(ctx context.Context, user source.User) (*scimsdk.User, error) {
	client := newSDMSCIMClient()
	return client.Users().Replace(ctx, user.SinkObjectID, scimsdk.ReplaceUser{
		UserName:   user.UserName,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Active:     user.Active,
	})
}

func internalSCIMSDKUsersDelete(ctx context.Context, userID string) (bool, error) {
	client := newSDMSCIMClient()
	return client.Users().Delete(ctx, userID)
}

func internalSCIMSDKGroupsList(ctx context.Context, paginationOpts *scimsdk.PaginationOptions) scimsdk.GroupIterator {
	client := newSDMSCIMClient()
	return client.Groups().List(ctx, paginationOpts)
}

func internalSCIMSDKGroupsCreate(ctx context.Context, group scimsdk.CreateGroupBody) (*scimsdk.Group, error) {
	client := newSDMSCIMClient()
	return client.Groups().Create(ctx, group)
}

func internalSCIMSDKGroupsUpdateReplaceMembers(ctx context.Context, groupID string, members []scimsdk.GroupMember) (bool, error) {
	client := newSDMSCIMClient()
	return client.Groups().UpdateReplaceMembers(ctx, groupID, members)
}

func internalSCIMSDKGroupsDelete(ctx context.Context, groupID string) (bool, error) {
	client := newSDMSCIMClient()
	return client.Groups().Delete(ctx, groupID)
}
