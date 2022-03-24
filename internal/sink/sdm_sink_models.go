package sink

import "github.com/strongdm/scimsdk/scimsdk"

type SDMUserRow struct {
	*scimsdk.User
	Role []scimsdk.Group
}

type SDMRoleRow struct {
	ID   string
	Name string
}

type SinkUserGroup struct {
	ID          string
	DisplayName string
	Members     []scimsdk.GroupMember
}
