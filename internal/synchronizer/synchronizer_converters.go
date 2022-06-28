package synchronizer

import (
	"fmt"
	"scim-integrations/internal/repository"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

func userSourceToUserSink(userSource *source.User) *sink.UserRow {
	groupNames := []string{}
	for _, groupRow := range userSource.Groups {
		groupNames = append(groupNames, fmt.Sprintf("%s (%s)", groupRow.DisplayName, groupRow.Path))
	}
	return &sink.UserRow{
		User: &sink.User{
			UserName:   userSource.UserName,
			GivenName:  userSource.GivenName,
			FamilyName: userSource.FamilyName,
			Active:     userSource.Active,
			GroupNames: groupNames,
		},
	}
}

func groupSourceToGroupSink(groupSource *source.UserGroup) *sink.GroupRow {
	return &sink.GroupRow{
		ID:          groupSource.SDMObjectID,
		DisplayName: groupSource.DisplayName,
		Path:        groupSource.Path,
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
