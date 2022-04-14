package synchronizer

import (
	"context"
	"fmt"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

type UserSynchronizer struct {
	rateLimiter *RateLimiter
	report      *Report
}

func NewUserSynchronizer(report *Report, rateLimiter *RateLimiter) *UserSynchronizer {
	return &UserSynchronizer{
		report:      report,
		rateLimiter: rateLimiter,
	}
}

func (sync *UserSynchronizer) Sync(ctx context.Context, snk sink.BaseSink) error {
	sync.rateLimiter.Start()
	err := sync.createUsers(ctx, snk, sync.report.IdPUsersToAdd)
	if err != nil {
		return err
	}
	err = sync.EnrichReport(snk)
	if err != nil {
		return err
	}
	err = sync.updateUsers(ctx, snk, sync.report.IdPUsersToUpdate)
	if err != nil {
		return err
	}
	if *flags.DeleteUsersNotInIdPFlag {
		err = sync.deleteDisjointedSDMUsers(ctx, snk, sync.report.SinkUsersNotInIdP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) EnrichReport(snk sink.BaseSink) error {
	sdmUsers, err := snk.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	sync.report.SinkUsers = sdmUsers
	newUsers, usersNotInIdP, existentUsers, usersWithUpdatedData := sync.removeSDMUsersIntersection()
	sync.report.IdPUsersToAdd = newUsers
	sync.report.IdPUsersInSink = existentUsers
	sync.report.SinkUsersNotInIdP = usersNotInIdP
	sync.report.IdPUsersToUpdate = usersWithUpdatedData
	return nil
}

func (sync *UserSynchronizer) removeSDMUsersIntersection() ([]*sink.UserRow, []*sink.UserRow, []*sink.UserRow, []*sink.UserRow) {
	var newUsers []*sink.UserRow
	var disjointedUsers []*sink.UserRow
	var existentUsers []*sink.UserRow
	var usersWithUpdatedData []*sink.UserRow
	var mappedUsers = map[string]bool{}
	var idpUsers = sync.report.IdPUsers
	for idpUserIdx, idpUser := range idpUsers {
		var found bool
		var isOutdated bool
		var sinkObjectID string
		for _, row := range sync.report.SinkUsers {
			if idpUser.UserName == row.User.UserName {
				found = true
				isOutdated = userHasOutdatedData(*row, *idpUser)
				sinkObjectID = row.User.ID
				break
			}
		}
		idpUsers[idpUserIdx].SDMObjectID = sinkObjectID
		sinkUser := userSourceToUserSink(idpUsers[idpUserIdx])
		if !found {
			newUsers = append(newUsers, sinkUser)
		} else {
			if isOutdated {
				usersWithUpdatedData = append(usersWithUpdatedData, sinkUser)
			}
			existentUsers = append(existentUsers, sinkUser)
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

func (sync *UserSynchronizer) createUsers(ctx context.Context, snk sink.BaseSink, sdmUsers []*sink.UserRow) error {
	for _, sdmUser := range sdmUsers {

		err := safeRetry(sync.rateLimiter, func() error {
			sdmUserResponse, err := snk.CreateUser(ctx, sdmUser)
			if err != nil {
				return err
			}
			fmt.Println(createSign, "User created:", sdmUserResponse.User.UserName)
			return nil
		}, "creating an user")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) updateUsers(ctx context.Context, snk sink.BaseSink, sdmUsers []*sink.UserRow) error {
	for _, sdmUser := range sdmUsers {
		err := safeRetry(sync.rateLimiter, func() error {
			err := snk.ReplaceUser(ctx, *sdmUser)
			if err != nil {
				return err
			}
			fmt.Println(updateSign, "User updated:", sdmUser.User.UserName)
			return nil
		}, "updating an user")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) deleteDisjointedSDMUsers(ctx context.Context, snk sink.BaseSink, users []*sink.UserRow) error {
	for _, user := range users {
		err := safeRetry(sync.rateLimiter, func() error {
			err := snk.DeleteUser(ctx, *user)
			if err != nil {
				return err
			}
			fmt.Println(deleteSign, "User deleted:", user.User.UserName)
			return nil
		}, "deleting an user")
		if err != nil {
			return err
		}
	}
	return nil
}

func userHasOutdatedData(row sink.UserRow, idpUser source.User) bool {
	if idpUser.Active != row.User.Active || idpUser.FamilyName != row.User.FamilyName || idpUser.GivenName != row.User.GivenName {
		return true
	}
	return false
}
