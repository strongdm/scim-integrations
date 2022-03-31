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

func (sync *UserSynchronizer) Sync(ctx context.Context) []error {
	errs := sync.createUsers(ctx, sync.report.IdPUsersToAdd)
	updateErrs := sync.updateUsers(ctx, sync.report.IdPUsersToUpdate)
	errs = append(errs, updateErrs...)
	if *flags.DeleteUsersNotInIdPFlag {
		deleteErrs := sync.DeleteDisjointedSDMUsers(ctx, sync.report.SDMUsersNotInIdP)
		errs = append(errs, deleteErrs...)
	}
	return errs
}

func (sync *UserSynchronizer) EnrichReport() error {
	sdmUsers, err := sdmscim.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	newUsers, usersNotInIdP, existentUsers, usersWithUpdatedData := removeSDMUsersIntersection(sdmUsers, sync.report.IdPUsers)
	sync.report.IdPUsersToAdd = newUsers
	sync.report.IdPUsersInSDM = existentUsers
	sync.report.SDMUsersNotInIdP = usersNotInIdP
	sync.report.IdPUsersToUpdate = usersWithUpdatedData
	return nil
}

func removeSDMUsersIntersection(sdmUsers []*sdmscim.UserRow, idpUsers []*source.User) ([]*source.User, []*sdmscim.UserRow, []*source.User, []*source.User) {
	var newUsers []*source.User
	var disjointedUsers []*sdmscim.UserRow
	var existentUsers []*source.User
	var usersWithUpdatedData []*source.User
	var mappedUsers map[string]bool = map[string]bool{}
	for idpUserIdx, idpUser := range idpUsers {
		var found bool
		var needUpdate bool
		var sinkObjectID string
		for _, sdmUser := range sdmUsers {
			if idpUser.UserName == sdmUser.UserName {
				found = true
				needUpdate = userHasOutdatedData(*sdmUser, *idpUser)
				sinkObjectID = sdmUser.ID
				break
			}
		}
		if !found {
			newUsers = append(newUsers, idpUser)
		} else {
			idpUsers[idpUserIdx].SinkObjectID = sinkObjectID
			if needUpdate {
				usersWithUpdatedData = append(usersWithUpdatedData, idpUsers[idpUserIdx])
			}
			existentUsers = append(existentUsers, idpUsers[idpUserIdx])
		}
	}
	for _, user := range idpUsers {
		mappedUsers[user.UserName] = true
	}
	for _, user := range sdmUsers {
		if _, ok := mappedUsers[user.UserName]; !ok {
			disjointedUsers = append(disjointedUsers, user)
		}
	}
	return newUsers, disjointedUsers, existentUsers, usersWithUpdatedData
}

func (sync *UserSynchronizer) createUsers(ctx context.Context, idpUsers []*source.User) []error {
	errs := []error{}
	for _, idpUser := range idpUsers {
		_, err := sdmscim.CreateUser(ctx, idpUser)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (sync *UserSynchronizer) updateUsers(ctx context.Context, idpUsers []*source.User) []error {
	errs := []error{}
	for _, idpUser := range idpUsers {
		err := sdmscim.ReplaceUser(ctx, *idpUser)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (sync *UserSynchronizer) DeleteDisjointedSDMUsers(ctx context.Context, users []*sdmscim.UserRow) []error {
	errs := []error{}
	for _, user := range users {
		err := sdmscim.DeleteUser(ctx, *user)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func userHasOutdatedData(sdmUser sdmscim.UserRow, idpUser source.User) bool {
	if idpUser.Active != sdmUser.Active || idpUser.FamilyName != sdmUser.Name.FamilyName || idpUser.GivenName != sdmUser.Name.GivenName {
		return true
	}
	return false
}
