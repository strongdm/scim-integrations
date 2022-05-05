package google

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func TestGoogleSourceFetchUsers(t *testing.T) {
	t.Run("should return a list of users when executing the default flow", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.GetGoogleTokenSourceFunc = mockGetGoogleTokenSourceFunc
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsers

		users, err := mock.FetchUsers(context.Background())

		assertT.NotNil(users)
		assertT.Len(users, 2)
		assertT.Nil(err)
	})

	t.Run("should return an empty list of users when executing the default flow", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.GetGoogleTokenSourceFunc = mockGetGoogleTokenSourceFunc
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsersEmpty

		users, err := mock.FetchUsers(context.Background())

		assertT.Nil(users)
		assertT.Empty(users)
		assertT.Nil(err)
	})

	t.Run("should return an error not finding the google credentials file", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsers

		users, err := mock.FetchUsers(context.Background())

		assertT.Nil(users)
		assertT.Contains(err.Error(), "Unable to read service account key")
		assertT.Contains(err.Error(), "no such file")
	})

	t.Run("should return a list of users when passing a context with timeout", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.GetGoogleTokenSourceFunc = mockGetGoogleTokenSourceFunc
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsers

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := mock.FetchUsers(timeoutContext)

		assertT.NotNil(users)
		assertT.Len(users, 2)
		assertT.Nil(err)
	})

	t.Run("should return a context timeout error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsersWhenExceedCTXTimeout
		mock.GetGoogleTokenSourceFunc = mockGetGoogleTokenSourceFunc

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := mock.FetchUsers(timeoutContext)

		assertT.NotNil(timeoutContext.Err())
		assertT.Nil(users)
		assertT.NotNil(err)
		assertT.Contains(timeoutContext.Err().Error(), "context deadline exceeded")
		assertT.Contains(err.Error(), "context deadline exceeded")
	})
}

func TestGoogleSourceExtractGroups(t *testing.T) {
	t.Run("should return a list of users groups when executing the normal flow", func(t *testing.T) {
		assertT := assert.New(t)
		mock := NewMockSourceGoogle()
		mock.InternalGoogleFetchUsersFunc = mockedInternalGoogleFetchUsers
		mock.GetGoogleTokenSourceFunc = mockGetGoogleTokenSourceFunc

		users, err := mock.FetchUsers(context.Background())

		assertT.NotNil(users)
		assertT.Len(users, 2)
		assertT.Nil(err)

		groups := mock.ExtractGroupsFromUsers(users)

		assertT.NotNil(groups)
		assertT.Len(groups, 1)
		assertT.Len(groups[0].Members, 2)
	})
}

func mockedInternalGoogleFetchUsers(_ *admin.Service, nextPageToken string) (*admin.Users, error) {
	response := &admin.Users{
		ServerResponse: googleapi.ServerResponse{
			Header:         nil,
			HTTPStatusCode: 200,
		},
		Users: []*admin.User{
			{Id: "xxx", Name: &admin.UserName{}, OrgUnitPath: "yyy/zzz"},
		},
	}
	if nextPageToken == "" {
		response.NextPageToken = "token"
	}
	return response, nil
}

func mockedInternalGoogleFetchUsersEmpty(_ *admin.Service, _ string) (*admin.Users, error) {
	return &admin.Users{
		ServerResponse: googleapi.ServerResponse{
			Header:         nil,
			HTTPStatusCode: 200,
		},
		Users: []*admin.User{},
	}, nil
}

func mockedInternalGoogleFetchUsersWhenExceedCTXTimeout(_ *admin.Service, _ string) (*admin.Users, error) {
	time.Sleep(time.Millisecond * 2)
	return nil, errors.New("context deadline exceeded")
}

func mockGetGoogleTokenSourceFunc(_ context.Context) (oauth2.TokenSource, error) {
	return &MockToken{}, nil
}
