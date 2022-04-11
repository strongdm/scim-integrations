package sink

import (
	"context"
)

type BaseSink interface {
	FetchUsers(context.Context) ([]*UserRow, error)
	CreateUser(context.Context, *UserRow) (*UserRow, error)
	ReplaceUser(context.Context, UserRow) error
	DeleteUser(context.Context, UserRow) error
	FetchGroups(context.Context) ([]*GroupRow, error)
	CreateGroup(context.Context, *GroupRow) (*GroupRow, error)
	ReplaceGroupMembers(context.Context, *GroupRow) error
	DeleteGroup(context.Context, *GroupRow) error
}

type User struct {
	ID          string
	UserName    string
	DisplayName string
	GivenName   string
	FamilyName  string
	Active      bool
	GroupName   []string
}

type UserRow struct {
	User   *User
	Groups []*GroupRow
}

type GroupRow struct {
	ID          string
	DisplayName string
	Members     []*GroupMember
}

type GroupMember struct {
	ID          string
	Email       string
	SDMObjectID string
}

type ISinkClient interface {
	Users() *ISinkUsersManager
	Groups() *ISinkGroupsManager
}

type ISinkUsersManager interface{}

type ISinkGroupsManager interface{}
