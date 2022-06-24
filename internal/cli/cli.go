package cli

import (
	"fmt"
	"os"
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
		"all",
		"add",
		"update",
		"delete",
		"h",
		"help",
	}
)

func ResolveCommand() (bool, error) {
	if exists, command := extractCommand(os.Args[1:]); exists && commandIsValid(command) {
		return true, commands[command].Exec()
	} else if exists {
		return false, fmt.Errorf("Command \"%s\" not found", command)
	}
	return false, nil
}

func extractCommand(args []string) (bool, string) {
	argsWithoutBoolFlags := removeBoolFlags(args)
	if len(argsWithoutBoolFlags)%2 == 0 {
		return false, ""
	}
	for _, arg := range argsWithoutBoolFlags {
		if _, ok := commands[arg]; ok {
			return true, arg
		}
	}
	return true, ""
}

func removeBoolFlags(args []string) []string {
	var cleanedCall []string = []string{}
	for _, arg := range args {
		found := false
		for _, flag := range boolFlags {
			if flag == strings.Replace(arg, "-", "", 1) {
				found = true
				break
			}
		}
		if !found {
			cleanedCall = append(cleanedCall, arg)
		}
	}
	return cleanedCall
}

func commandIsValid(command string) bool {
	return commands[command] != nil
}
