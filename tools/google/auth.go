package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
)

func main() {
	token, getTokenErr := tokenFromFile("token.json")
	if getTokenErr == nil && token != nil {
		fmt.Println("You are trying to authenticate again. Delete you current token.json file and run the script again.")
	}
	_, err := setupTokenFile()
	if err != nil {
		log.Fatalf("An error occurred when preparing the Google HTTP client: %v", err)
	}
	if getTokenErr != nil {
		fmt.Println("Authentication process finished!")
	}
}

func setupTokenFile() (bool, error) {
	config, err := getGoogleConfig()
	if err != nil {
		return false, err
	}
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return false, err
		}
		err = saveToken(tokFile, tok)
		if err != nil {
			return false, err
		}
	}
	return true, nil
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

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.New("Unable to read authorization code: " + err.Error())
	}
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.New("Unable to retrieve token from web: " + err.Error())
	}
	return tok, nil
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New("Unable to cache oauth token: " + err.Error())
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(token)
	if err != nil {
		return errors.New("Unable to chache oauth token: " + err.Error())
	}
	return nil
}
