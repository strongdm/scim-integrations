package source

import (
	"context"
	"scim-integrations/internal/sink"
)

type BaseSource interface {
	FetchUsers(ctx context.Context) ([]*User, error)
	ExtractGroupsFromUsers([]*User) []*UserGroup
}

type User struct {
	ID          string
	UserName    string
	GivenName   string
	FamilyName  string
	Active      bool
	Groups      []UserGroup
	SDMObjectID string
}

type UserGroup struct {
	DisplayName string
	Path        string
	Members     []*sink.GroupMember
	SDMObjectID string
}
