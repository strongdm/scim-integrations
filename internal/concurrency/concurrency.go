package concurrency

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

const fileNotFoundErrMessage = "no such file"

func SetupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			RemoveLockFile()
			os.Exit(-1)
		}
	}()
}

func SetupRecover() {
	if err := recover(); err != nil {
		RemoveLockFile()
		panic(err)
	}
}

func CreateLockFile() error {
	filename := GetLockFileName()
	info, err := os.Stat(filename)
	if info != nil {
		return errors.New("The application is already running. You cannot run in parallel.")
	} else if err != nil && !strings.Contains(err.Error(), fileNotFoundErrMessage) {
		return err
	}
	_, err = os.Create(filename)
	if err != nil {
		return err
	}
	err = os.Chmod(filename, 0700)
	if err != nil {
		return err
	}
	return nil
}

func RemoveLockFile() error {
	filename := GetLockFileName()
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info != nil {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetLockFileName() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/.lock", path)
}
