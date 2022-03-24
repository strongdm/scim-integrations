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
	response, err := svc.Users.List().Customer("my_customer").MaxResults(FETCH_PAGE_SIZE).Do()
	if err != nil {
		return nil, err
	}
	return googleUsersToSCIMUser(response), nil
}

func getGroups(orgUnitPath string) []string {
	orgUnits := strings.Split(orgUnitPath, "/")
	return []string{orgUnits[len(orgUnits)-1]}
}

func googleUsersToSCIMUser(response *admin.Users) []IdPUser {
	var users []IdPUser
	for _, googleUser := range response.Users {
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
