package sync

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/idp"
	"scim-integrations/internal/sink"

	"github.com/strongdm/scimsdk/scimsdk"
)

type UserSynchronize struct {
	service *sink.SDMService
	report  *Report
}

type roleList []sink.SDMRoleRow
type userList []sink.SDMUserRow

func NewUserSynchronize(service *sink.SDMService, report *Report) *UserSynchronize {
	return &UserSynchronize{
		service: service,
		report:  report,
	}
}

func (sync *UserSynchronize) Sync(ctx context.Context, idpUsers []idp.IdPUser) error {
	sdmUsers, err := sync.service.FetchUsers(ctx)
	if err != nil {
		return err
	}
	newUsers, inexistentUsers, unmatchingUsers := calculateSDMUsersIntersection(sdmUsers, idpUsers)
	matchingUsers, err := sync.createUsers(ctx, newUsers)
	if err != nil {
		return err
	}
	if *flags.DeleteUnmatchingUsersFlag {
		err = sync.DeleteUnmatchingSDMUsers(ctx, unmatchingUsers)
		if err != nil {
			return err
		}
	}
	sync.report.SDMUsersInIdP = inexistentUsers
	sync.report.SDMUsersNotInIdP = inexistentUsers
	sync.report.SDMNewUsers = matchingUsers
	return nil
}

func (sync *UserSynchronize) createUsers(ctx context.Context, idpUsers []idp.IdPUser) ([]scimsdk.User, error) {
	var matchingUsers []scimsdk.User
	for _, idpUser := range idpUsers {
		user, err := sync.service.CreateUser(ctx, idpUser)
		if err != nil {
			return nil, err
		}
		matchingUsers = append(matchingUsers, *user)
	}
	return matchingUsers, nil
}

func (sync *UserSynchronize) DeleteUnmatchingSDMUsers(ctx context.Context, users []sink.SDMUserRow) error {
	for _, user := range users {
		err := sync.service.DeleteUser(ctx, user.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// func (sync *UserSynchronize) createCompositeRole(ctx context.Context, roles roleList, idpGroups []string) (*sink.SDMRoleRow, error) {
// 	compositeRoleName := strings.Join(idpGroups, "_")
// 	compositeRole := roleExists(compositeRoleName, roles)
// 	if compositeRole == nil {
// 		newRole, err := sync.service.CreateRole(ctx, compositeRoleName)
// 		if err != nil {
// 			return nil, err
// 		}
// 		compositeRole = &sink.SDMRoleRow{
// 			ID:   newRole.ID,
// 			Name: newRole.DisplayName,
// 		}
// 	}
// 	return compositeRole, nil
// }

func calculateSDMUsersIntersection(sdmUsers []sink.SDMUserRow, idpUsers []idp.IdPUser) ([]idp.IdPUser, []idp.IdPUser, []sink.SDMUserRow) {
	var existentUsers []idp.IdPUser
	var newUsers []idp.IdPUser
	for _, iuser := range idpUsers {
		var found bool
		for _, user := range sdmUsers {
			if iuser.ID == user.ID {
				found = true
				break
			}
		}
		if !found {
			newUsers = append(newUsers, iuser)
		} else {
			existentUsers = append(existentUsers, iuser)
		}
	}
	unmatchingUsers := removeSDMUsersIntersection(sdmUsers, idpUsers)
	return newUsers, existentUsers, unmatchingUsers
}

func removeSDMUsersIntersection(sdmUsers []sink.SDMUserRow, existentIdPUsers []idp.IdPUser) []sink.SDMUserRow {
	var unmatchingUsers []sink.SDMUserRow
	mappedUsers := map[string]bool{}
	for _, user := range existentIdPUsers {
		mappedUsers[user.ID] = true
	}
	for _, user := range sdmUsers {
		if _, ok := mappedUsers[user.ID]; !ok {
			unmatchingUsers = append(unmatchingUsers, user)
		}
	}
	return unmatchingUsers
}

func (sync *UserSynchronize) syncUserRole(ctx context.Context, user *scimsdk.User, matchingRoles []scimsdk.Group) (*scimsdk.Group, error) {
	role := matchingRoles[0]
	err := sync.service.AssignRole(ctx, user, role.ID)
	if err != nil {
		return nil, err
	}
	return &role, nil
}
