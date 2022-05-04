package synchronizer

import (
	"scim-integrations/internal/repository"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"time"
)

func userSourceToUserSink(userSource *source.User) *sink.UserRow {
	return &sink.UserRow{
		User: &sink.User{
			UserName:   userSource.UserName,
			GivenName:  userSource.GivenName,
			FamilyName: userSource.FamilyName,
			Active:     userSource.Active,
			GroupNames: userSource.Groups,
		},
	}
}

func groupSourceToGroupSink(groupSource *source.UserGroup) *sink.GroupRow {
	return &sink.GroupRow{
		ID:          groupSource.SDMObjectID,
		DisplayName: groupSource.DisplayName,
		Members:     groupSource.Members,
	}
}

func reportToRepositoryReportsRow(report *Report) *repository.ReportsRow {
	return &repository.ReportsRow{
		StartedAt:           report.Start,
		CompletedAt:         report.Complete,
		UsersToCreateCount:  report.UsersToCreateCount,
		UsersToUpdateCount:  report.UsersToUpdateCount,
		UsersToDeleteCount:  report.UsersToDeleteCount,
		GroupsToCreateCount: report.GroupsToCreateCount,
		GroupsToUpdateCount: report.GroupsToUpdateCount,
		GroupsToDeleteCount: report.GroupsToDeleteCount,
	}
}

func errorToRepositoryErrorRow(err error) *repository.ErrorsRow {
	return &repository.ErrorsRow{
		Err:          err.Error(),
		OccurredTime: time.Now(),
	}
}
