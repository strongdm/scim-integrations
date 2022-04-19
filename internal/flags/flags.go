package flags

import (
	"errors"
	"flag"
)

var validIdPs = map[string]bool{"google": true}
var IdPFlag = flag.String("idp", "", "use Google as an IdP")
var DeleteGroupsNotInIdPFlag = flag.Bool("delete-groups-missing-in-idp", false, "delete groups present in SDM but not in the selected Identity Provider")
var DeleteUsersNotInIdPFlag = flag.Bool("delete-users-missing-in-idp", false, "delete users present in SDM but not in the selected Identity Provider")
var ApplyFlag = flag.Bool("apply", false, "apply the planned changes")
var QueryFlag = flag.String("query", "", "pass a query according to the available query syntax for the selected Identity Provider")
var RateLimiterFlag = flag.Bool("rate-limiter", false, "synchronize the planned data with a requester rate limiter, limiting with a limit set as 1000 requests per 30 seconds")

func ValidateMandatoryFlags() error {
	if _, ok := validIdPs[*IdPFlag]; !ok {
		return errors.New("you must specify one Identity Provider (IdP): Google\nUse -idp \"google\"")
	}
	return nil
}
