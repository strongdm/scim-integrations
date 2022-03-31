package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/source"
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
	src := source.ByFlag(*flags.IdPFlag)
	snc := synchronizer.NewSynchronizer()
	snc.Run(src)
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you need to set the SDM_SCIM_TOKEN env var")
	}
	return nil
}
