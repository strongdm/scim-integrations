package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

type UserSynchronizer struct {
	service *sink.SDMSink
	report  *Report
}

func NewUserSynchronize(service *sink.SDMSink, report *Report) *UserSynchronizer {
	return &UserSynchronizer{
		service: service,
		report:  report,
	}
}

func (sync *UserSynchronizer) Sync(ctx context.Context, newUsers []source.SourceUser, deletedUsers []sink.SDMUserRow) error {
	err := sync.createUsers(ctx, newUsers)
	if err != nil {
		return err
	}
	if *flags.DeleteUnmatchingUsersFlag {
		err = sync.DeleteUnmatchingSDMUsers(ctx, deletedUsers)
		if err != nil {
			return err
		}
	}
	return nil
}

func (synchronizer *UserSynchronizer) SyncUsersData(idpUsers []source.SourceUser) error {
	sdmUsers, err := synchronizer.service.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	var existentUsers []source.SourceUser
	var newUsers []source.SourceUser
	for _, iuser := range idpUsers {
		var found bool
		for _, user := range sdmUsers {
			if iuser.UserName == user.UserName {
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
	synchronizer.report.SourceUsersToAdd = newUsers
	synchronizer.report.SourceMatchingUsers = existentUsers
	synchronizer.report.SDMUnmatchingUsers = unmatchingUsers
	return nil
}

func removeSDMUsersIntersection(sdmUsers []sink.SDMUserRow, existentIdPUsers []source.SourceUser) []sink.SDMUserRow {
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

func (sync *UserSynchronizer) createUsers(ctx context.Context, idpUsers []source.SourceUser) error {
	for _, idpUser := range idpUsers {
		_, err := sync.service.CreateUser(ctx, idpUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) DeleteUnmatchingSDMUsers(ctx context.Context, users []sink.SDMUserRow) error {
	for _, user := range users {
		err := sync.service.DeleteUser(ctx, user.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
