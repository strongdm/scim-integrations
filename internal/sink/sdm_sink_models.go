package sink

import "github.com/strongdm/scimsdk/scimsdk"

type SDMUserRow struct {
	*scimsdk.User
	Groups []scimsdk.Group
}

type SDMGroupRow struct {
	ID   string
	Name string
}

type SinkUserGroup struct {
	ID          string
	DisplayName string
	Members     []scimsdk.GroupMember
}
