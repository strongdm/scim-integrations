package sink

import (
	"errors"

	"github.com/strongdm/scimsdk/scimsdk"
)

func sdmscimUsersWithGroupsToSink(iterator *scimsdk.UsersIterator, userGroups map[string][]SDMGroupRow) ([]SDMUserRow, error) {
	var result []SDMUserRow
	for iterator.Next() {
		user := iterator.Value()
		result = append(result, SDMUserRow{
			User:   user,
			Groups: userGroups[user.ID],
		})
	}
	if iterator.Err() != "" {
		return nil, errors.New(iterator.Err())
	}
	return result, nil
}

func sdmscimUserToSink(response *scimsdk.User) *SDMUserRow {
	return &SDMUserRow{User: response}
}

func sdmscimGroupToSink(response *scimsdk.Group) *SDMGroupRow {
	return &SDMGroupRow{
		ID:          response.ID,
		DisplayName: response.DisplayName,
		Members:     sdmscimGroupMembersToSink(response.Members),
	}
}

func sdmscimGroupMembersToSink(scimMembers []*scimsdk.GroupMember) []SDMGroupMember {
	var members []SDMGroupMember
	for _, member := range scimMembers {
		members = append(members, SDMGroupMember(*member))
	}
	return members
}
