package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type commandFunc func() error
type command struct {
	Name string
	Exec commandFunc
}

var (
	commands  = map[string]*command{exposeMetricsCommand.Name: exposeMetricsCommand}
	boolFlags = []string{
		"apply",
		"enable-rate-limiter",
		"delete-groups-missing-in-idp",
		"delete-users-missing-in-idp",
		"all",
		"add",
		"update",
		"delete",
	}
)

func ResolveCommand() (bool, error) {
	if ok, command := extractCommand(strings.Join(os.Args[1:], " ")); ok && commandIsValid(command) {
		return true, commands[command].Exec()
	} else if ok {
		return false, fmt.Errorf("Command \"%s\" not found", command)
	}
	return false, nil
}

func extractCommand(call string) (bool, string) {
	callWithoutBoolFlags := removeBoolFlags(call)
	compiled, _ := regexp.Compile(fmt.Sprint("-[^ ]* ('([^']|\\')+'|\"([^\"]|\\\")+\"|[^ ]+)"))
	flags := compiled.FindAllString(callWithoutBoolFlags, -1)
	callWithoutFlags := callWithoutBoolFlags
	for _, flagEntry := range flags {
		callWithoutFlags = strings.ReplaceAll(callWithoutFlags, flagEntry, "")
	}
	command := strings.TrimSpace(callWithoutFlags)
	return command != "", command
}

func removeBoolFlags(call string) string {
	var cleanedCall string = call
	for _, flag := range boolFlags {
		if strings.Contains(call, flag) {
			cleanedCall = strings.ReplaceAll(cleanedCall, fmt.Sprintf(" -%s", flag), "")
		}
	}
	return cleanedCall
}

func commandIsValid(command string) bool {
	return commands[command] != nil
}
