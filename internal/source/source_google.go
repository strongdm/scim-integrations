package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(filePath string) (*oauth2.Token, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(token)
	if err != nil {
		log.Fatalf("Unable to chache oauth token: %v", err)
	}
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
