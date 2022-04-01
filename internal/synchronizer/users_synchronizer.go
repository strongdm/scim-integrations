package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
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

func (sync *UserSynchronizer) Sync(ctx context.Context, errCh chan error) {
	sync.createUsers(ctx, sync.report.IdPUsersToAdd, errCh)
	err := sync.EnrichReport()
	if err != nil {
		errCh <- err
		return
	}
	sync.updateUsers(ctx, sync.report.IdPUsersToUpdate, errCh)
	if *flags.DeleteUsersNotInIdPFlag {
		sync.DeleteDisjointedSDMUsers(ctx, sync.report.SinkUsersNotInIdP, errCh)
	}
}

func (sync *UserSynchronizer) EnrichReport() error {
	if len(sync.report.SinkUsers) == 0 {
		sdmUsers, err := sdmscim.FetchUsers(context.Background())
		if err != nil {
			return err
		}
		sync.report.SinkUsers = sdmUsers
	}
	newUsers, usersNotInIdP, existentUsers, usersWithUpdatedData := sync.removeSDMUsersIntersection()
	sync.report.IdPUsersToAdd = newUsers
	sync.report.IdPUsersInSink = existentUsers
	sync.report.SinkUsersNotInIdP = usersNotInIdP
	sync.report.IdPUsersToUpdate = usersWithUpdatedData
	return nil
}

func (sync *UserSynchronizer) removeSDMUsersIntersection() ([]*source.User, []*sink.UserRow, []*source.User, []*source.User) {
	var newUsers []*source.User
	var disjointedUsers []*sink.UserRow
	var existentUsers []*source.User
	var usersWithUpdatedData []*source.User
	var mappedUsers map[string]bool = map[string]bool{}
	var idpUsers = sync.report.IdPUsers
	for idpUserIdx, idpUser := range idpUsers {
		var found bool
		var needUpdate bool
		var sinkObjectID string
		for _, row := range sync.report.SinkUsers {
			if idpUser.UserName == row.User.UserName {
				found = true
				needUpdate = userHasOutdatedData(*row, *idpUser)
				sinkObjectID = row.User.ID
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
	for _, row := range sync.report.SinkUsers {
		if _, ok := mappedUsers[row.User.UserName]; !ok {
			disjointedUsers = append(disjointedUsers, row)
		}
	}
	return newUsers, disjointedUsers, existentUsers, usersWithUpdatedData
}

func (sync *UserSynchronizer) createUsers(ctx context.Context, idpUsers []*source.User, errCh chan error) {
	for _, idpUser := range idpUsers {
		_, err := sdmscim.CreateUser(ctx, idpUser)
		if err != nil {
			errCh <- err
		}
	}
}

func (sync *UserSynchronizer) updateUsers(ctx context.Context, idpUsers []*source.User, errCh chan error) {
	for _, idpUser := range idpUsers {
		err := sdmscim.ReplaceUser(ctx, *idpUser)
		if err != nil {
			errCh <- err
		}
	}
}

func (sync *UserSynchronizer) DeleteDisjointedSDMUsers(ctx context.Context, users []*sink.UserRow, errCh chan error) {
	for _, user := range users {
		err := sdmscim.DeleteUser(ctx, *user)
		if err != nil {
			errCh <- err
		}
	}
}

func userHasOutdatedData(row sink.UserRow, idpUser source.User) bool {
	if idpUser.Active != row.User.Active || idpUser.FamilyName != row.User.FamilyName || idpUser.GivenName != row.User.GivenName {
		return true
	}
	return false
}
