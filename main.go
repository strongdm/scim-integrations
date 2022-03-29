package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/source"
	"scim-integrations/internal/synchronizer"
)

// TODO: Add tests
func main() {
	flag.Parse()
	err := flags.ValidateMandatoryFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	if err := validateEnvironment(); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred setting up the environment: "+err.Error())
		os.Exit(-1)
	}
	source := source.ByFlag(getFlagName())
	snc := synchronizer.NewSynchronizer()
	err = snc.Run(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred running Source sync: %s", err.Error())
	}
	log.Println("Source sync ran successfully")
}

func getFlagName() string {
	if *flags.GoogleFlag {
		return "google"
	}
	return ""
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you need to set the SDM_SCIM_TOKEN env var")
	}
	return nil
}
