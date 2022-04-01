package sink

type User struct {
	ID          string
	UserName    string
	DisplayName string
	GivenName   string
	FamilyName  string
	Active      bool
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
	ID    string
	Email string
}

type ISinkClient interface {
	Users() *ISinkUsersManager
	Groups() *ISinkGroupsManager
}

type ISinkUsersManager interface{}

type ISinkGroupsManager interface{}
