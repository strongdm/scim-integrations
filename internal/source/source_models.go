package source

import "github.com/strongdm/scimsdk/scimsdk"

type User struct {
	ID         string
	UserName   string
	GivenName  string
	FamilyName string
	Active     bool
	Groups     []string
}

type UserGroup struct {
	ID          string
	DisplayName string
	Members     []scimsdk.GroupMember
}
