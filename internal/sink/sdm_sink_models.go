package sink

import "github.com/strongdm/scimsdk/scimsdk"

type SDMUserRow struct {
	*scimsdk.User
	Groups []SDMGroupRow
}

type SDMGroupRow struct {
	ID          string
	DisplayName string
	Members     []SDMGroupMember
}

type SDMGroupMember struct {
	ID    string
	Email string
}
