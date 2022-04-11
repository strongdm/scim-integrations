package google

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
	"io/ioutil"
	"net/http"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
)

const fetchPageSize = 500

// DefaultGoogleCustomer refer to customer field in: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list
const DefaultGoogleCustomer = "my_customer"

type SourceGoogle interface {
	FetchUsers(context.Context) ([]*source.User, error)
	ExtractGroupsFromUsers([]*source.User) []*source.UserGroup
	InternalGoogleFetchUsers(*admin.Service, string) (*admin.Users, error)
	TokenFromFile(string) (*oauth2.Token, error)
	GetGoogleConfig() (*oauth2.Config, error)
}

type SourceGoogleImpl struct{}

func NewGoogleSource() SourceGoogle {
	return &SourceGoogleImpl{}
}

func (g *SourceGoogleImpl) FetchUsers(ctx context.Context) ([]*source.User, error) {
	return internalFetchUsers(ctx, g)
}

func internalFetchUsers(ctx context.Context, src SourceGoogle) ([]*source.User, error) {
	client, err := prepareGoogleHTTPClient(src)
	if err != nil {
		return nil, err
	}
	svc, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	var nextPageToken string
	var users []*source.User
	for {
		response, err := src.InternalGoogleFetchUsers(svc, nextPageToken)
		if err != nil {
			return nil, err
		}
		users = append(users, googleUsersToSCIMUser(response.Users)...)
		if response.NextPageToken == "" {
			break
		}
		nextPageToken = response.NextPageToken
	}
	return users, nil
}

func (*SourceGoogleImpl) ExtractGroupsFromUsers(users []*source.User) []*source.UserGroup {
	var groups []*source.UserGroup
	mappedGroupMembers := map[string][]*sink.GroupMember{}
	for _, user := range users {
		for _, userGroup := range user.Groups {
			if _, ok := mappedGroupMembers[userGroup]; !ok {
				mappedGroupMembers[userGroup] = []*sink.GroupMember{
					{
						ID:    user.ID,
						Email: user.UserName,
					},
				}
			} else {
				mappedGroupMembers[userGroup] = append(mappedGroupMembers[userGroup], &sink.GroupMember{
					ID:    user.ID,
					Email: user.UserName,
				})
			}
		}
	}
	for groupName, members := range mappedGroupMembers {
		groups = append(groups, &source.UserGroup{DisplayName: groupName, Members: members})
	}
	return groups
}

func (*SourceGoogleImpl) InternalGoogleFetchUsers(service *admin.Service, nextPageToken string) (*admin.Users, error) {
	return service.Users.List().Query(*flags.QueryFlag).Customer(DefaultGoogleCustomer).PageToken(nextPageToken).MaxResults(fetchPageSize).Do()
}

func googleUsersToSCIMUser(googleUsers []*admin.User) []*source.User {
	var users []*source.User
	for _, googleUser := range googleUsers {
		users = append(users, &source.User{
			ID:         googleUser.Id,
			UserName:   googleUser.PrimaryEmail,
			GivenName:  googleUser.Name.GivenName,
			FamilyName: googleUser.Name.FamilyName,
			Active:     !googleUser.Suspended,
			Groups:     []string{googleUser.OrgUnitPath},
		})
	}
	return users
}

func prepareGoogleHTTPClient(src SourceGoogle) (*http.Client, error) {
	config, err := src.GetGoogleConfig()
	if err != nil {
		return nil, err
	}
	token, err := src.TokenFromFile("token.json")
	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), token), nil
}

func (*SourceGoogleImpl) TokenFromFile(filePath string) (*oauth2.Token, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

func (*SourceGoogleImpl) GetGoogleConfig() (*oauth2.Config, error) {
	credentialsBytes, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.New("Unable to read client secret file: " + err.Error())
	}
	config, err := google.ConfigFromJSON(credentialsBytes, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, errors.New("Unable to parse client secret file to config: " + err.Error())
	}
	return config, nil
}
