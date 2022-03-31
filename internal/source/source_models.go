package source

import "github.com/strongdm/scimsdk/scimsdk"

type User struct {
	ID           string
	UserName     string
	GivenName    string
	FamilyName   string
	Active       bool
	Groups       []string
	SinkObjectID string
}

type UserGroup struct {
	DisplayName  string
	Members      []scimsdk.GroupMember
	SinkObjectID string
}
