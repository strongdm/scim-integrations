package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"scim-integrations/internal/concurrency"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink/sdmscim"
	"scim-integrations/internal/source"
	"scim-integrations/internal/source/google"
	"scim-integrations/internal/synchronizer"
)

func Start() {
	defer concurrency.SetupRecoverOnPanic()
	concurrency.SetupInterruptSignalHandler()
	checkEnvironment()
	src := getSourceByFlag()
	snc := synchronizer.NewSynchronizer()
	snk := sdmscim.NewSinkSDMSCIM()
	err := concurrency.CreateLockFile()
	if err != nil {
		showErr(err)
		return
	}
	err = snc.Run(src, snk)
	if err != nil {
		showErr(err)
	}
	removeFileErr := concurrency.RemoveLockFile()
	if removeFileErr != nil {
		showErr(removeFileErr)
	}
}

func checkEnvironment() {
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
}

func showErr(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func validateEnvironment() error {
	if os.Getenv("SDM_SCIM_TOKEN") == "" {
		return errors.New("you must set the SDM_SCIM_TOKEN env var")
	} else if os.Getenv("SDM_SCIM_IDP_GOOGLE_KEY_PATH") == "" {
		return errors.New("you must set the SDM_SCIM_IDP_GOOGLE_KEY_PATH env var")
	} else if os.Getenv("SDM_SCIM_IDP_GOOGLE_SUBJECT_USER") == "" {
		return errors.New("you must set the SDM_SCIM_IDP_GOOGLE_SUBJECT_USER env var")
	}
	return nil
}

func getSourceByFlag() source.BaseSource {
	if *flags.IdPFlag == "google" {
		return google.NewGoogleSource()
	}
	return nil
}
