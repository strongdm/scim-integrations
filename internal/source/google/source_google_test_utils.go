package google

import (
	"context"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"scim-integrations/internal/source"
)

type MockSourceGoogle struct {
	FetchUsersFunc               func(ctx context.Context) ([]*source.User, error)
	ExtractGroupsFromUsersFunc   func([]*source.User) []*source.UserGroup
	InternalGoogleFetchUsersFunc func(*admin.Service, string) (*admin.Users, error)
	TokenFromFileFunc            func(string) (*oauth2.Token, error)
	GetGoogleConfigFunc          func() (*oauth2.Config, error)
}

func NewMockSourceGoogle() *MockSourceGoogle {
	src := SourceGoogleImpl{}
	mock := MockSourceGoogle{}
	mock.FetchUsersFunc = func(ctx context.Context) ([]*source.User, error) {
		return internalFetchUsers(ctx, &mock)
	}
	mock.ExtractGroupsFromUsersFunc = src.ExtractGroupsFromUsers
	mock.InternalGoogleFetchUsersFunc = src.InternalGoogleFetchUsers
	mock.TokenFromFileFunc = src.TokenFromFile
	mock.GetGoogleConfigFunc = src.GetGoogleConfig
	return &mock
}

func (src *MockSourceGoogle) FetchUsers(ctx context.Context) ([]*source.User, error) {
	return src.FetchUsersFunc(ctx)
}

func (src *MockSourceGoogle) ExtractGroupsFromUsers(users []*source.User) []*source.UserGroup {
	return src.ExtractGroupsFromUsersFunc(users)
}

func (src *MockSourceGoogle) InternalGoogleFetchUsers(srv *admin.Service, nextPageToken string) (*admin.Users, error) {
	return src.InternalGoogleFetchUsersFunc(srv, nextPageToken)
}

func (src *MockSourceGoogle) TokenFromFile(filePath string) (*oauth2.Token, error) {
	return src.TokenFromFileFunc(filePath)
}

func (src *MockSourceGoogle) GetGoogleConfig() (*oauth2.Config, error) {
	return src.GetGoogleConfigFunc()
}
