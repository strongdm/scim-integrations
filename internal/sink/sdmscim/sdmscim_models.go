package sdmscim

import "github.com/strongdm/scimsdk/scimsdk"

type UserRow struct {
	*scimsdk.User
	Groups []GroupRow
}

type GroupRow struct {
	ID          string
	DisplayName string
	Members     []GroupMember
}

type GroupMember struct {
	ID    string
	Email string
}

type ISinkClient interface {
	Users() *ISinkUsersManager
	Groups() *ISinkGroupsManager
}

type ISinkUsersManager interface{}

type ISinkGroupsManager interface{}
