package flags

import (
	"errors"
	"flag"
)

// TODO: add a flag for the RateLimiter
var validIdPs = map[string]bool{"google": true}
var IdPFlag = flag.String("idp", "", "use Google as an IdP")
var UserFlag = flag.String("user", "", "pass the user e-mail")
var DeleteGroupsNotInIdPFlag = flag.Bool("delete-groups-missing-in-idp", false, "delete groups present in SDM but not in the selected Identity Provider")
var DeleteUsersNotInIdPFlag = flag.Bool("delete-users-missing-in-idp", false, "delete users present in SDM but not in the selected Identity Provider")
var PlanFlag = flag.Bool("plan", false, "do not apply changes just show initial queries")
var QueryFlag = flag.String("query", "", "pass a query according to the available query syntax for the selected Identity Provider")
var VerboseFlag = flag.Bool("verbose", false, "show the verbose report output")

func ValidateMandatoryFlags() error {
	if _, ok := validIdPs[*IdPFlag]; !ok {
		return errors.New("you must specify one Identity Provider (IdP): Google\nUse -idp \"google\"")
	}
	if *IdPFlag == "google" {
		if *UserFlag == "" {
			return errors.New("you must specify the user flag passing the service account admin email")
		}
	}
	return nil
}
