package google

import (
	"context"
	"scim-integrations/internal/source"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type MockSourceGoogle struct {
	FetchUsersFunc               func(context.Context) ([]*source.User, error)
	ExtractGroupsFromUsersFunc   func([]*source.User) []*source.UserGroup
	InternalGoogleFetchUsersFunc func(*admin.Service, string) (*admin.Users, error)
	GetGoogleAdminServiceFunc    func(ctx context.Context) (*admin.Service, error)
	GetGoogleTokenSourceFunc     func(ctx context.Context) (oauth2.TokenSource, error)
}

type MockToken struct{}

func (*MockToken) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func NewMockSourceGoogle() *MockSourceGoogle {
	src := sourceGoogleImpl{}
	mock := MockSourceGoogle{}
	mock.FetchUsersFunc = func(ctx context.Context) ([]*source.User, error) {
		return internalFetchUsers(ctx, &mock)
	}
	mock.ExtractGroupsFromUsersFunc = src.ExtractGroupsFromUsers
	mock.InternalGoogleFetchUsersFunc = src.InternalGoogleFetchUsers
	mock.GetGoogleTokenSourceFunc = src.GetGoogleTokenSource
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

func (src *MockSourceGoogle) GetGoogleAdminService(ctx context.Context) (*admin.Service, error) {
	ts, err := src.GetGoogleTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	return admin.NewService(ctx, option.WithTokenSource(ts))
}

func (src *MockSourceGoogle) GetGoogleTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return src.GetGoogleTokenSourceFunc(ctx)
}
