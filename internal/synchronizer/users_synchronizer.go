package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink/sdmscim"
	"scim-integrations/internal/source"
)

type UserSynchronizer struct {
	report *Report
}

func NewUserSynchronizer(report *Report) *UserSynchronizer {
	return &UserSynchronizer{
		report: report,
	}
}

func (sync *UserSynchronizer) Sync(ctx context.Context) error {
	err := sync.createUsers(ctx, sync.report.IdPUsersToAdd)
	if err != nil {
		return err
	}
	if *flags.DeleteUsersNotInIdPFlag {
		err = sync.DeleteDisjointedSDMUsers(ctx, sync.report.SDMUsersNotInIdP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) EnrichReport() error {
	sdmUsers, err := sdmscim.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	var existentUsers []source.User
	var newUsers []source.User
	for _, idpUser := range sync.report.IdPUsers {
		var found bool
		for _, user := range sdmUsers {
			if idpUser.UserName == user.UserName {
				found = true
				break
			}
		}
		if !found {
			newUsers = append(newUsers, idpUser)
		} else {
			existentUsers = append(existentUsers, idpUser)
		}
	}
	usersNotInIdP := removeSDMUsersIntersection(sdmUsers, sync.report.IdPUsers)
	sync.report.IdPUsersToAdd = newUsers
	sync.report.IdPUsersInSDM = existentUsers
	sync.report.SDMUsersNotInIdP = usersNotInIdP
	return nil
}

func removeSDMUsersIntersection(sdmUsers []sdmscim.UserRow, existentIdPUsers []source.User) []sdmscim.UserRow {
	var disjointedUsers []sdmscim.UserRow
	mappedUsers := map[string]bool{}
	for _, user := range existentIdPUsers {
		mappedUsers[user.UserName] = true
	}
	for _, user := range sdmUsers {
		if _, ok := mappedUsers[user.UserName]; !ok {
			disjointedUsers = append(disjointedUsers, user)
		}
	}
	return disjointedUsers
}

func (sync *UserSynchronizer) createUsers(ctx context.Context, idpUsers []source.User) error {
	for _, idpUser := range idpUsers {
		_, err := sdmscim.CreateUser(ctx, idpUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) DeleteDisjointedSDMUsers(ctx context.Context, users []sdmscim.UserRow) error {
	for _, user := range users {
		err := sdmscim.DeleteUser(ctx, user.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
