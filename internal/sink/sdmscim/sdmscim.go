package sdmscim

import (
	"context"
	"os"
	"scim-integrations/internal/source"

	"github.com/strongdm/scimsdk/scimsdk"
)

func newSDMSCIMClient() *scimsdk.Client {
	client := scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
	return client
}

func FetchUsers(ctx context.Context) ([]UserRow, error) {
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

func separateGroupsByUser(groups []GroupRow) map[string][]GroupRow {
	userGroups := map[string][]GroupRow{}
	for _, group := range groups {
		for _, member := range group.Members {
			if userGroups[member.ID] == nil {
				userGroups[member.ID] = []GroupRow{group}
			} else {
				userGroups[member.ID] = append(userGroups[member.ID], group)
			}
		}
	}
	return userGroups
}

func CreateUser(ctx context.Context, user source.User) (*UserRow, error) {
	response, err := internalSCIMSDKUsersCreate(ctx, scimsdk.CreateUser{
		UserName:   user.UserName,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, err
	}
	return userToSink(response), nil
}

func DeleteUser(ctx context.Context, userID string) error {
	_, err := internalSCIMSDKUsersDelete(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func FetchGroups(ctx context.Context) ([]GroupRow, error) {
	iterator := internalSCIMSDKGroupsList(ctx, nil)
	var result []GroupRow
	for iterator.Next() {
		group := *iterator.Value()
		result = append(result, *groupToSink(&group))
	}
	if iterator.Err() != nil {
		return nil, iterator.Err()
	}
	return result, nil
}

func CreateGroup(ctx context.Context, group source.UserGroup) (*GroupRow, error) {
	response, err := internalSCIMSDKGroupsCreate(ctx, scimsdk.CreateGroupBody{
		DisplayName: group.DisplayName,
		Members:     group.Members,
	})
	if err != nil {
		return nil, err
	}
	return groupToSink(response), nil
}

func ReplaceGroupMembers(ctx context.Context, group source.UserGroup) error {
	_, err := internalSCIMSDKGroupsUpdateReplaceMembers(ctx, group.ID, group.Members)
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(ctx context.Context, groupID string) error {
	_, err := internalSCIMSDKGroupsDelete(ctx, groupID)
	if err != nil {
		return err
	}
	return nil
}

func internalSCIMSDKGroupsList(ctx context.Context, paginationOpts *scimsdk.PaginationOptions) scimsdk.GroupIterator {
	client := newSDMSCIMClient()
	return client.Groups().List(ctx, paginationOpts)
}

func internalSCIMSDKUsersList(ctx context.Context, paginationOpts *scimsdk.PaginationOptions) scimsdk.UserIterator {
	client := newSDMSCIMClient()
	return client.Users().List(ctx, paginationOpts)
}

func internalSCIMSDKUsersCreate(ctx context.Context, user scimsdk.CreateUser) (*scimsdk.User, error) {
	client := newSDMSCIMClient()
	return client.Users().Create(ctx, user)
}

func internalSCIMSDKUsersDelete(ctx context.Context, userID string) (bool, error) {
	client := newSDMSCIMClient()
	return client.Users().Delete(ctx, userID)
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
