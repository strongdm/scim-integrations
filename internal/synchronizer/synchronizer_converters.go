package synchronizer

import (
	"scim-integrations/internal/repository"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
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
		StartedAt:          report.Start,
		CompletedAt:        report.Complete,
		CreatedUsersCount:  report.CreatedUsersCount,
		UpdatedUsersCount:  report.UpdatedUsersCount,
		DeletedUsersCount:  report.DeletedUsersCount,
		CreatedGroupsCount: report.CreatedGroupsCount,
		UpdatedGroupsCount: report.UpdatedGroupsCount,
		DeletedGroupsCount: report.DeletedGroupsCount,
		Succeed:            report.Succeed,
	}
}
