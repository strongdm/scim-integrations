package source

import (
	"context"
	"errors"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func TestGoogleSourceFetchUsers(t *testing.T) {
	t.Run("should return a list of users when executing the default flow", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsers)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)
		monkey.Patch(getGoogleConfig, mockedGetGoogleConfig)

		googleSource := NewGoogleSource()
		users, err := googleSource.FetchUsers(context.Background())

		assert.NotNil(users)
		assert.Len(users, 2)
		assert.Nil(err)

		monkey.UnpatchAll()
	})

	t.Run("should return an empty list of users when executing the default flow", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsersEmpty)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)
		monkey.Patch(getGoogleConfig, mockedGetGoogleConfig)

		googleSource := NewGoogleSource()
		users, err := googleSource.FetchUsers(context.Background())

		assert.Nil(users)
		assert.Empty(users)
		assert.Nil(err)

		monkey.UnpatchAll()
	})

	t.Run("should return an error not finding the google credentials file", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsers)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)

		googleSource := NewGoogleSource()
		users, err := googleSource.FetchUsers(context.Background())

		assert.Nil(users)
		assert.Contains(err.Error(), "credentials.json")
		assert.Contains(err.Error(), "no such file")

		monkey.UnpatchAll()
	})

	t.Run("should return a list of users when passing a context with timeout", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsers)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)
		monkey.Patch(getGoogleConfig, mockedGetGoogleConfig)

		googleSource := NewGoogleSource()
		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := googleSource.FetchUsers(timeoutContext)

		assert.NotNil(users)
		assert.Len(users, 2)
		assert.Nil(err)

		monkey.UnpatchAll()
	})

	t.Run("should return a context timeout error when the context timeout exceeds", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsersWhenExceedCTXTimeout)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)
		monkey.Patch(getGoogleConfig, mockedGetGoogleConfig)

		googleSource := NewGoogleSource()
		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := googleSource.FetchUsers(timeoutContext)

		assert.NotNil(timeoutContext.Err())
		assert.Nil(users)
		assert.NotNil(err)
		assert.Contains(timeoutContext.Err().Error(), "context deadline exceeded")
		assert.Contains(err.Error(), "context deadline exceeded")

		monkey.UnpatchAll()
	})
}

func TestGoogleSourceExtractGroups(t *testing.T) {
	t.Run("should return a list of users groups when executing the normal flow", func(t *testing.T) {
		assert := assert.New(t)
		monkey.Patch(internalGoogleFetchUsers, mockedInternalGoogleFetchUsers)
		monkey.Patch(tokenFromFile, mockedTokenFromFile)
		monkey.Patch(getGoogleConfig, mockedGetGoogleConfig)

		googleSource := NewGoogleSource()
		users, err := googleSource.FetchUsers(context.Background())

		assert.NotNil(users)
		assert.Len(users, 2)
		assert.Nil(err)

		userGroups := googleSource.ExtractGroupsFromUsers(users)

		assert.NotNil(userGroups)
		assert.Len(userGroups, 1)
		assert.Len(userGroups[0].Members, 2)

		monkey.UnpatchAll()
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

func mockedTokenFromFile(filePath string) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func mockedGetGoogleConfig() (*oauth2.Config, error) {
	return &oauth2.Config{}, nil
}
