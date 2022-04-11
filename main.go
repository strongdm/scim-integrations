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
	src := getSourceByFlag(*flags.IdPFlag)
	snc := synchronizer.NewSynchronizer()
	snk := sdmscim.NewSinkSDMSCIMImpl()
	err = snc.Run(src, snk)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Printf("Sync with %s IdP finished\n", *flags.IdPFlag)
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you need to set the SDM_SCIM_TOKEN env var")
	}
	return nil
}

func getSourceByFlag(name string) source.BaseSource {
	if name == "google" {
		return google.NewGoogleSource()
	}
	return nil
}
