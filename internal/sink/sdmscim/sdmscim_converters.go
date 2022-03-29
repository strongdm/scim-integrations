package sdmscim

import (
	"errors"

	"github.com/strongdm/scimsdk/scimsdk"
)

func usersWithGroupsToSink(iterator *scimsdk.UsersIterator, userGroups map[string][]GroupRow) ([]UserRow, error) {
	var result []UserRow
	for iterator.Next() {
		user := iterator.Value()
		result = append(result, UserRow{
			User:   user,
			Groups: userGroups[user.ID],
		})
	}
	if iterator.Err() != "" {
		return nil, errors.New(iterator.Err())
	}
	return result, nil
}

func userToSink(response *scimsdk.User) *UserRow {
	return &UserRow{User: response}
}

func groupToSink(response *scimsdk.Group) *GroupRow {
	return &GroupRow{
		ID:          response.ID,
		DisplayName: response.DisplayName,
		Members:     groupMembersToSink(response.Members),
	}
}

func groupMembersToSink(scimMembers []*scimsdk.GroupMember) []GroupMember {
	var members []GroupMember
	for _, member := range scimMembers {
		members = append(members, GroupMember(*member))
	}
	return members
}
