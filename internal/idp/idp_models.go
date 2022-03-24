package idp

type IdPUser struct {
	ID         string
	UserName   string
	GivenName  string
	FamilyName string
	Active     bool
	Groups     []string
}
