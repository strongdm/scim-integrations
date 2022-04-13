package synchronizer

import (
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

func userSourceToUserSink(userSource *source.User) *sink.UserRow {
	return &sink.UserRow{
		User: &sink.User{
			ID:         userSource.ID,
			UserName:   userSource.UserName,
			GivenName:  userSource.GivenName,
			FamilyName: userSource.FamilyName,
			Active:     userSource.Active,
			GroupNames: userSource.Groups,
			SinkID:     userSource.SDMObjectID,
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
