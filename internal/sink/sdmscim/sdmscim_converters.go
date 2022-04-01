package sdmscim

import (
	"errors"
	"fmt"
	"scim-integrations/internal/sink"

	"github.com/strongdm/scimsdk/scimsdk"
)

func usersWithGroupsToSink(iterator scimsdk.UserIterator, userGroups map[string][]*sink.GroupRow) ([]*sink.UserRow, error) {
	var result []*sink.UserRow
	for iterator.Next() {
		user := iterator.Value()
		result = append(result, userToSink(user, userGroups[user.ID]))
	}
	if iterator.Err() != nil {
		return nil, errors.New(fmt.Sprintf("An error was occurred listing the SDM users: %v\n", iterator.Err()))
	}
	return result, nil
}

func userToSink(response *scimsdk.User, userGroups []*sink.GroupRow) *sink.UserRow {
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

func groupToSink(response *scimsdk.Group) *sink.GroupRow {
	return &sink.GroupRow{
		ID:          response.ID,
		DisplayName: response.DisplayName,
		Members:     groupMembersToSink(response.Members),
	}
}

func groupMembersToSink(scimMembers []*scimsdk.GroupMember) []*sink.GroupMember {
	var members []*sink.GroupMember
	for _, member := range scimMembers {
		members = append(members, &sink.GroupMember{ID: member.ID, Email: member.Email})
	}
	return members
}

func sinkGroupMemberListToSDMSCIM(members []*sink.GroupMember) []scimsdk.GroupMember {
	var sdmMembers []scimsdk.GroupMember
	for _, member := range members {
		sdmMembers = append(sdmMembers, sinkGroupMemberToSDMSCIM(*member))
	}
	return sdmMembers
}

func sinkGroupMemberToSDMSCIM(member sink.GroupMember) scimsdk.GroupMember {
	return scimsdk.GroupMember{
		ID:    member.ID,
		Email: member.Email,
	}
}
