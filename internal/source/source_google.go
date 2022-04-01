package source

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"scim-integrations/internal/flags"
	"strings"

	"github.com/strongdm/scimsdk/scimsdk"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type GoogleSource struct{}

// DefaultGoogleCustomer refer to customer field in: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list
const DefaultGoogleCustomer = "my_customer"

func NewGoogleSource() *GoogleSource {
	return &GoogleSource{}
}

func (g *GoogleSource) FetchUsers(ctx context.Context) ([]*User, error) {
	client, err := prepareGoogleHTTPClient()
	if err != nil {
		return nil, err
	}
	svc, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	var nextPageToken string
	var users []*User
	for {
		response, err := internalGoogleFetchUsers(svc, nextPageToken)
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

func (*GoogleSource) ExtractGroupsFromUsers(users []*User) []*UserGroup {
	var groups []*UserGroup
	mappedGroupMembers := map[string][]scimsdk.GroupMember{}
	for _, user := range users {
		for _, userGroup := range user.Groups {
			if _, ok := mappedGroupMembers[userGroup]; !ok {
				mappedGroupMembers[userGroup] = []scimsdk.GroupMember{
					{
						ID:    user.ID,
						Email: user.UserName,
					},
				}
			} else {
				mappedGroupMembers[userGroup] = append(mappedGroupMembers[userGroup], scimsdk.GroupMember{
					ID:    user.ID,
					Email: user.UserName,
				})
			}
		}
	}
	for groupName, members := range mappedGroupMembers {
		groups = append(groups, &UserGroup{DisplayName: groupName, Members: members})
	}
	return groups
}

func internalGoogleFetchUsers(service *admin.Service, nextPageToken string) (*admin.Users, error) {
	return service.Users.List().Query(*flags.QueryFlag).Customer(DefaultGoogleCustomer).PageToken(nextPageToken).MaxResults(FetchPageSize).Do()
}

func googleUsersToSCIMUser(googleUsers []*admin.User) []*User {
	var users []*User
	for _, googleUser := range googleUsers {
		users = append(users, &User{
			ID:         googleUser.Id,
			UserName:   googleUser.PrimaryEmail,
			GivenName:  googleUser.Name.GivenName,
			FamilyName: googleUser.Name.FamilyName,
			Active:     !googleUser.Suspended,
			Groups:     getGroups(googleUser.OrgUnitPath),
		})
	}
	return users
}

func getGroups(orgUnitPath string) []string {
	orgUnits := strings.Split(orgUnitPath, "/")
	return []string{orgUnits[len(orgUnits)-1]}
}

func prepareGoogleHTTPClient() (*http.Client, error) {
	config, err := getGoogleConfig()
	if err != nil {
		return nil, err
	}
	token, err := tokenFromFile("token.json")
	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), token), nil
}

func tokenFromFile(filePath string) (*oauth2.Token, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

func getGoogleConfig() (*oauth2.Config, error) {
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
