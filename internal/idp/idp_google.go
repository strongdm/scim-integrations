package idp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"scim-integrations/internal/utils"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type GoogleIdP struct{}

func NewGoogleIdP() GoogleIdP {
	return GoogleIdP{}
}

func (GoogleIdP) FetchUsers(ctx context.Context) ([]IdPUser, error) {
	client, err := prepareGoogleHTTPClient(HTTPClient)
	if err != nil {
		return nil, err
	}
	svc, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	var nextPageToken string
	var users []IdPUser
	for {
		response, err := svc.Users.List().Customer("my_customer").PageToken(nextPageToken).MaxResults(FETCH_PAGE_SIZE).Do()
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

func googleUsersToSCIMUser(googleUsers []*admin.User) []IdPUser {
	var users []IdPUser
	var filteredGoogleUsers []*admin.User = filterGoogleUsersByOrganizationUnit(googleUsers)
	for _, googleUser := range filteredGoogleUsers {
		users = append(users, IdPUser{
			ID:         googleUser.Id,
			UserName:   googleUser.PrimaryEmail,
			GivenName:  googleUser.Name.GivenName,
			FamilyName: googleUser.Name.FamilyName,
			Groups:     getGroups(googleUser.OrgUnitPath),
		})
	}
	return users
}

func filterGoogleUsersByOrganizationUnit(users []*admin.User) []*admin.User {
	var filteredUsers []*admin.User
	var organizations []string = getOrganizationsFilter()
	if organizations == nil {
		return users
	}
	for _, user := range users {
		units := strings.Split(user.OrgUnitPath, "/")
		lastOrgUnit := units[len(units)-1]
		if utils.StringArrayContains(organizations, lastOrgUnit) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func getGroups(orgUnitPath string) []string {
	orgUnits := strings.Split(orgUnitPath, "/")
	return []string{orgUnits[len(orgUnits)-1]}
}

func prepareGoogleHTTPClient(client *http.Client) (*http.Client, error) {
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
	json.NewEncoder(file).Encode(token)
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

func getOrganizationsFilter() []string {
	orgFilter := os.Getenv("SDM_SHIM_GOOGLE_ORGANIZATIONS_FILTER")
	if orgFilter == "" {
		return nil
	}
	return strings.Split(orgFilter, " ")
}
