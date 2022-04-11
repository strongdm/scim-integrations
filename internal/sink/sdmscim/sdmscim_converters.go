package sdmscim

import (
	"errors"
	"fmt"
	scimmodels "github.com/strongdm/scimsdk/models"
	"scim-integrations/internal/sink"
)

func usersWithGroupsToSink(iterator scimmodels.Iterator[scimmodels.User], userGroups map[string][]*sink.GroupRow) ([]*sink.UserRow, error) {
	var result []*sink.UserRow
	for iterator.Next() {
		user := scimmodels.User(*iterator.Value())
		result = append(result, userToSink(&user, userGroups[user.ID]))
	}
	if iterator.Err() != nil {
		return nil, errors.New(fmt.Sprintf("An error was occurred listing the SDM users: %v\n", iterator.Err()))
	}
	return result, nil
}

func userToSink(response *scimmodels.User, userGroups []*sink.GroupRow) *sink.UserRow {
	return &sink.UserRow{
		User: &sink.User{
			ID:          response.ID,
			DisplayName: response.DisplayName,
			UserName:    response.UserName,
			GivenName:   response.Name.GivenName,
			FamilyName:  response.Name.FamilyName,
			Active:      response.Active,
		},
		Groups: userGroups,
	}
}

func groupToSink(response *scimmodels.Group) *sink.GroupRow {
	return &sink.GroupRow{
		ID:          response.ID,
		DisplayName: response.DisplayName,
		Members:     groupMembersToSink(response.Members),
	}
}

func groupMembersToSink(scimMembers []*scimmodels.GroupMember) []*sink.GroupMember {
	var members []*sink.GroupMember
	for _, member := range scimMembers {
		members = append(members, &sink.GroupMember{ID: member.ID, Email: member.Email})
	}
	return members
}

func sinkGroupMemberListToSDMSCIM(members []*sink.GroupMember) []scimmodels.GroupMember {
	var sdmMembers []scimmodels.GroupMember
	for _, member := range members {
		sdmMembers = append(sdmMembers, sinkGroupMemberToSDMSCIM(*member))
	}
	return sdmMembers
}

func sinkGroupMemberToSDMSCIM(member sink.GroupMember) scimmodels.GroupMember {
	return scimmodels.GroupMember{
		ID:    member.SDMObjectID,
		Email: member.Email,
	}
}
