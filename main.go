package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink/sdmscim"
	"scim-integrations/internal/source"
	"scim-integrations/internal/source/google"
	"scim-integrations/internal/synchronizer"
)

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
	err = start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Printf("Sync with %s IdP finished\n", *flags.IdPFlag)
}

func start() error {
	src := getSourceByFlag()
	snc := synchronizer.NewSynchronizer()
	snk := sdmscim.NewSinkSDMSCIMImpl()
	setupSignalHandler()
	err := createLockFile()
	if err != nil {
		return err
	}
	err = snc.Run(src, snk)
	err = removeLockFile()
	if err != nil {
		return err
	}
	return nil
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you need to set the SDM_SCIM_TOKEN env var")
	}
	return nil
}

func getSourceByFlag() source.BaseSource {
	if *flags.IdPFlag == "google" {
		return google.NewGoogleSource()
	}
	return nil
}
