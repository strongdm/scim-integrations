package sink

import (
	"context"
	"errors"
	"os"
	"scim-integrations/internal/source"

	"github.com/strongdm/scimsdk/scimsdk"
	sdm "github.com/strongdm/strongdm-sdk-go"
)

func newSDMSCIMClient() *scimsdk.Client {
	client := scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
	return client
}

func FetchUsers(ctx context.Context) ([]SDMUserRow, error) {
	groups, err := FetchGroups(ctx)
	if err != nil {
		return nil, err
	}
	userGroups := separateGroupsByUser(groups)
	client := newSDMSCIMClient()
	userIterator := client.Users().List(ctx, nil)
	users, err := sdmscimUsersWithGroupsToSink(userIterator, userGroups)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func separateGroupsByUser(groups []SDMGroupRow) map[string][]SDMGroupRow {
	userGroups := map[string][]SDMGroupRow{}
	for _, group := range groups {
		for _, member := range group.Members {
			if userGroups[member.ID] == nil {
				userGroups[member.ID] = []SDMGroupRow{group}
			} else {
				userGroups[member.ID] = append(userGroups[member.ID], group)
			}
		}
	}
	return userGroups
}

func CreateUser(ctx context.Context, user source.User) (*SDMUserRow, error) {
	client := newSDMSCIMClient()
	response, err := client.Users().Create(ctx, scimsdk.CreateUser{
		UserName:   user.UserName,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, err
	}
	return sdmscimUserToSink(response), nil
}

func DeleteUser(ctx context.Context, userID string) error {
	client := newSDMSCIMClient()
	_, err := client.Users().Delete(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func FetchGroups(ctx context.Context) ([]SDMGroupRow, error) {
	client := newSDMSCIMClient()
	groupIterator := client.Groups().List(ctx, nil)
	var result []SDMGroupRow
	for groupIterator.Next() {
		group := *groupIterator.Value()
		result = append(result, *sdmscimGroupToSink(&group))
	}
	if groupIterator.Err() != "" {
		return nil, errors.New(groupIterator.Err())
	}
	return result, nil
}

func AssignGroup(ctx context.Context, user *scimsdk.User, groupID string) error {
	client := newSDMSCIMClient()
	_, err := client.Groups().UpdateAddMembers(ctx, groupID, []scimsdk.GroupMember{
		{
			ID:    user.ID,
			Email: user.UserName,
		},
	})
	var alreadyExistsErr *sdm.AlreadyExistsError
	if err != nil && !errors.As(err, &alreadyExistsErr) {
		return err
	}
	return nil
}

func CreateGroup(ctx context.Context, group source.UserGroup) (*SDMGroupRow, error) {
	client := newSDMSCIMClient()
	response, err := client.Groups().Create(ctx, scimsdk.CreateGroupBody{
		DisplayName: group.DisplayName,
		Members:     group.Members,
	})
	if err != nil {
		return nil, err
	}
	return sdmscimGroupToSink(response), nil
}

func ReplaceGroupMembers(ctx context.Context, group source.UserGroup) error {
	client := newSDMSCIMClient()
	_, err := client.Groups().UpdateReplaceMembers(ctx, group.ID, group.Members)
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(ctx context.Context, groupID string) error {
	client := newSDMSCIMClient()
	_, err := client.Groups().Delete(ctx, groupID)
	if err != nil {
		return err
	}
	return nil
}
