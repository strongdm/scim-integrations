package source

import "github.com/strongdm/scimsdk/scimsdk"

type SourceUser struct {
	ID         string
	UserName   string
	GivenName  string
	FamilyName string
	Active     bool
	Groups     []string
}

type SourceUserGroup struct {
	ID          string
	DisplayName string
	Members     []scimsdk.GroupMember
}
