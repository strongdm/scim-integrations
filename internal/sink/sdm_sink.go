package sink

import (
	"context"
	"errors"
	"os"
	"scim-integrations/internal/idp"

	"github.com/strongdm/scimsdk/scimsdk"
	sdm "github.com/strongdm/strongdm-sdk-go"
)

type SDMService struct {
	client *scimsdk.Client
}

func NewSDMService() *SDMService {
	client := scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
	return &SDMService{client}
}

func (service *SDMService) FetchUsers(ctx context.Context) ([]SDMUserRow, error) {
	roles, err := service.FetchRoles(ctx)
	userRoles := map[string][]scimsdk.Group{}
	for _, role := range roles {
		for _, member := range role.Members {
			if userRoles[member.ID] == nil {
				userRoles[member.ID] = []scimsdk.Group{role}
			} else {
				userRoles[member.ID] = append(userRoles[member.ID], role)
			}
		}
	}

	// TODO: Review filter behavior with resourceType
	userIterator := service.client.Users().List(ctx, nil /*"type:user"*/)
	if err != nil {
		return nil, err
	}
	var result []SDMUserRow
	for userIterator.Next() {
		user := userIterator.Value()
		result = append(result, SDMUserRow{
			User: user,
			Role: userRoles[user.ID],
		})
	}
	if userIterator.Err() != "" {
		return nil, errors.New(userIterator.Err())
	}
	return result, nil
}

func (service *SDMService) CreateUser(ctx context.Context, user idp.IdPUser) (*scimsdk.User, error) {
	response, err := service.client.Users().Create(ctx, scimsdk.CreateUser{
		UserName:   user.UserName,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Active:     true,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (service *SDMService) DeleteUser(ctx context.Context, userID string) error {
	_, err := service.client.Users().Delete(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (service *SDMService) FetchRoles(ctx context.Context) ([]scimsdk.Group, error) {
	groupIterator := service.client.Groups().List(ctx, nil)
	var result []scimsdk.Group
	for groupIterator.Next() {
		result = append(result, *groupIterator.Value())
	}
	if groupIterator.Err() != "" {
		return nil, errors.New(groupIterator.Err())
	}
	return result, nil
}

func (service *SDMService) AssignRole(ctx context.Context, user *scimsdk.User, roleID string) error {
	_, err := service.client.Groups().UpdateAddMembers(ctx, roleID, []scimsdk.GroupMember{
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

func (service *SDMService) CreateRole(ctx context.Context, group SinkUserGroup) (*scimsdk.Group, error) {
	response, err := service.client.Groups().Create(ctx, scimsdk.CreateGroupBody{
		DisplayName: group.DisplayName,
		Members:     group.Members,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (service *SDMService) ReplaceGroupMembers(ctx context.Context, group SinkUserGroup) error {
	_, err := service.client.Groups().UpdateReplaceMembers(ctx, group.ID, group.Members)
	if err != nil {
		return err
	}
	return nil
}

func (service *SDMService) DeleteRole(ctx context.Context, roleID string) error {
	_, err := service.client.Groups().Delete(ctx, roleID)
	if err != nil {
		return err
	}
	return nil
}
