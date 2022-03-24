package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/idp"
	"scim-integrations/internal/sync"
)

// TODO: Add documentation
// TODO: Add tests
func main() {
	err := flags.ValidateMandatoryFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	if err := validateEnvironment(); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred setting up the environment: "+err.Error())
		os.Exit(-1)
	}
	idp := idp.ByFlag(getFlagName())
	rnr := sync.NewRunner()
	err = rnr.Run(idp)
	if err != nil {
		log.Fatal("An error occurred running IDP sync: " + err.Error())
	}
	fmt.Println("IdP sync ran successfully")
}

func getFlagName() string {
	if *flags.OktaFlag {
		return "okta"
	}
	return "google"
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you need to set the SDM_SCIM_TOKEN env var")
	}
	return nil
}
