package flags

import (
	"errors"
	"flag"
)

var validIdPs = map[string]bool{"google": true}
var IdPFlag = flag.String("idp", "", "use Google as an IdP")
var DeleteGroupsNotInIdPFlag = flag.Bool("delete-groups-missing-in-idp", false, "delete groups present in SDM but not in the selected Identity Provider")
var DeleteUsersNotInIdPFlag = flag.Bool("delete-users-missing-in-idp", false, "delete users present in SDM but not in the selected Identity Provider")
var PlanFlag = flag.Bool("plan", false, "do not apply changes just show initial queries")
var VerboseFlag = flag.Bool("verbose", false, "show the verbose report output")
var QueryFlag = flag.String("query", "", "pass a query according to the available query syntax for the selected Identity Provider")

func ValidateMandatoryFlags() error {
	if _, ok := validIdPs[*IdPFlag]; !ok {
		return errors.New("you need to specify one Identity Provider (IdP): Google\nUse -idp \"google\"")
	}
	return nil
}
