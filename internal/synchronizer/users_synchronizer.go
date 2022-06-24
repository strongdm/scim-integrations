package synchronizer

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

type UserSynchronizer struct {
	report  *Report
	retrier Retrier
}

func newUserSynchronizer(report *Report, retrier Retrier) *UserSynchronizer {
	return &UserSynchronizer{
		report:  report,
		retrier: retrier,
	}
}

// Sync synchronizes the users to be added, updated and deleted.
func (sync *UserSynchronizer) Sync(ctx context.Context, snk sink.BaseSink) error {
	if !sync.haveContentForSync() {
		return nil
	}
	fmt.Println("Synchronizing users...")
	sync.retrier.setEntityScope(UserScope)
	if *flags.AllOperationFlag || *flags.AddOperationFlag {
		createdUsersCount, err := sync.createUsers(ctx, snk, sync.report.IdPUsersToCreate)
		sync.report.CreatedUsersCount = createdUsersCount
		if err != nil {
			return err
		}
		err = sync.EnrichReport(snk)
		if err != nil {
			return err
		}
	}
	if *flags.AllOperationFlag || *flags.UpdateOperationFlag {
		updatedUsersCount, err := sync.updateUsers(ctx, snk, sync.report.IdPUsersToUpdate)
		sync.report.UpdatedUsersCount = updatedUsersCount
		if err != nil {
			return err
		}
	}
	if *flags.AllOperationFlag || *flags.DeleteOperationFlag {
		deletedUsersCount, err := sync.deleteMissingSDMUsers(ctx, snk, sync.report.SinkUsersNotInIdP)
		sync.report.DeletedUsersCount = deletedUsersCount
		if err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}

// EnrichReport calculates users to be added, updated and deleted
func (sync *UserSynchronizer) EnrichReport(snk sink.BaseSink) error {
	sdmUsers, err := snk.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sync.report.SinkUsers = sdmUsers
	newUsers, usersNotInIdP, existentUsers, usersWithUpdatedData := sync.intersectUsers()
	sync.report.IdPUsersToCreate = newUsers
	sync.report.IdPUsersInSink = existentUsers
	sync.report.SinkUsersNotInIdP = usersNotInIdP
	sync.report.IdPUsersToUpdate = usersWithUpdatedData
	return nil
}

func (sync *UserSynchronizer) haveContentForSync() bool {
	canPerformAdd := *flags.AllOperationFlag || *flags.AddOperationFlag
	canPerformUpdate := *flags.AllOperationFlag || *flags.UpdateOperationFlag
	canPerformDelete := *flags.AllOperationFlag || *flags.DeleteOperationFlag
	rpt := sync.report
	return (len(rpt.IdPUsersToCreate) > 0 && canPerformAdd) ||
		(len(rpt.IdPUsersToUpdate) > 0 && canPerformUpdate) ||
		(len(rpt.SinkUsersNotInIdP) > 0 && canPerformDelete)
}

func (sync *UserSynchronizer) intersectUsers() ([]*sink.UserRow, []*sink.UserRow, []*sink.UserRow, []*sink.UserRow) {
	var newUsers []*sink.UserRow
	var missingUsers []*sink.UserRow = sync.getMissingUsers()
	var existentUsers []*sink.UserRow
	var usersWithUpdatedData []*sink.UserRow
	for _, idpUser := range sync.report.IdPUsers {
		var sinkUser *sink.UserRow = userSourceToUserSink(idpUser)
		var found, isOutdated bool
		var sinkID string
		if found, isOutdated, sinkID = sync.userExistsInSink(idpUser); !found {
			newUsers = append(newUsers, sinkUser)
			continue
		}
		sinkUser.User.ID = sinkID
		if isOutdated {
			usersWithUpdatedData = append(usersWithUpdatedData, sinkUser)
		}
		existentUsers = append(existentUsers, sinkUser)
	}
	return newUsers, missingUsers, existentUsers, usersWithUpdatedData
}

func (sync *UserSynchronizer) userExistsInSink(idpUser *source.User) (bool, bool, string) {
	var found bool
	var isOutdated bool
	var sinkID string
	for _, sinkUser := range sync.report.SinkUsers {
		if found = idpUser.UserName == sinkUser.User.UserName; found {
			isOutdated = userHasOutdatedData(*sinkUser, *idpUser)
			sinkID = sinkUser.User.ID
			break
		}
	}
	return found, isOutdated, sinkID
}

func (sync *UserSynchronizer) getMissingUsers() []*sink.UserRow {
	var missingUsers []*sink.UserRow
	var mappedUsers = map[string]bool{}
	for _, user := range sync.report.IdPUsers {
		mappedUsers[user.UserName] = true
	}
	for _, row := range sync.report.SinkUsers {
		if _, ok := mappedUsers[row.User.UserName]; !ok {
			missingUsers = append(missingUsers, row)
		}
	}
	return missingUsers
}

func (sync *UserSynchronizer) createUsers(ctx context.Context, snk sink.BaseSink, sdmUsers []*sink.UserRow) (int, error) {
	var successCount int
	for _, sdmUser := range sdmUsers {
		err := sync.retrier.Run(func() error {
			sdmUserResponse, err := snk.CreateUser(ctx, sdmUser)
			if err != nil {
				return err
			}
			fmt.Println(createSign, "User created:", sdmUserResponse.User.UserName)
			successCount++
			return nil
		}, "creating an user")
		if err != nil {
			return successCount, err
		}
	}
	return successCount, nil
}

func (sync *UserSynchronizer) updateUsers(ctx context.Context, snk sink.BaseSink, sdmUsers []*sink.UserRow) (int, error) {
	var successCount int
	for _, sdmUser := range sdmUsers {
		err := sync.retrier.Run(func() error {
			err := snk.ReplaceUser(ctx, *sdmUser)
			if err != nil {
				return err
			}
			fmt.Println(updateSign, "User updated:", sdmUser.User.UserName)
			successCount++
			return nil
		}, "updating an user")
		if err != nil {
			return successCount, err
		}
	}
	return successCount, nil
}

func (sync *UserSynchronizer) deleteMissingSDMUsers(ctx context.Context, snk sink.BaseSink, users []*sink.UserRow) (int, error) {
	var successCount int
	for _, user := range users {
		err := sync.retrier.Run(func() error {
			err := snk.DeleteUser(ctx, *user)
			if err != nil {
				return err
			}
			fmt.Println(deleteSign, "User deleted:", user.User.UserName)
			return nil
		}, "deleting an user")
		if err != nil {
			return successCount, err
		}
	}
	return successCount, nil
}

func userHasOutdatedData(row sink.UserRow, idpUser source.User) bool {
	if idpUser.Active != row.User.Active || idpUser.FamilyName != row.User.FamilyName || idpUser.GivenName != row.User.GivenName {
		return true
	}
	return false
}
