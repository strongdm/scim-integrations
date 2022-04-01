package source

import (
	"scim-integrations/internal/sink"
)

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
	Members      []*sink.GroupMember
	SinkObjectID string
}
