package synchronizer

import (
	"context"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
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

func (s *UserSynchronizer) Sync(ctx context.Context) error {
	err := s.createUsers(ctx, s.report.IdPUsersToAdd)
	if err != nil {
		return err
	}
	if *flags.DeleteUsersNotInIdPFlag {
		err = s.DeleteUnmatchingSDMUsers(ctx, s.report.SDMUsersNotInIdP)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserSynchronizer) EnrichReport() error {
	sdmUsers, err := sink.FetchUsers(context.Background())
	if err != nil {
		return err
	}
	var existentUsers []source.User
	var newUsers []source.User
	for _, iuser := range s.report.IdPUsers {
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
	usersNotInIdP := removeSDMUsersIntersection(sdmUsers, s.report.IdPUsers)
	s.report.IdPUsersToAdd = newUsers
	s.report.IdPUsersInSDM = existentUsers
	s.report.SDMUsersNotInIdP = usersNotInIdP
	return nil
}

func removeSDMUsersIntersection(sdmUsers []sink.SDMUserRow, existentIdPUsers []source.User) []sink.SDMUserRow {
	var unmatchingUsers []sink.SDMUserRow
	mappedUsers := map[string]bool{}
	for _, user := range existentIdPUsers {
		mappedUsers[user.UserName] = true
	}
	for _, user := range sdmUsers {
		if _, ok := mappedUsers[user.UserName]; !ok {
			unmatchingUsers = append(unmatchingUsers, user)
		}
	}
	return unmatchingUsers
}

func (sync *UserSynchronizer) createUsers(ctx context.Context, idpUsers []source.User) error {
	for _, idpUser := range idpUsers {
		_, err := sink.CreateUser(ctx, idpUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sync *UserSynchronizer) DeleteUnmatchingSDMUsers(ctx context.Context, users []sink.SDMUserRow) error {
	for _, user := range users {
		err := sink.DeleteUser(ctx, user.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
