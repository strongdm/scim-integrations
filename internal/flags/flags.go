package flags

import (
	"errors"
	"flag"
)

var validIdPs = map[string]bool{"google": true}

// Flags
var IdPFlag = flag.String("idp", "", "use Google as an IdP")
var ApplyFlag = flag.Bool("apply", false, "apply the planned changes")
var QueryFlag = flag.String("query", "", "pass a query according to the available query syntax for the selected Identity Provider")
var RateLimiterFlag = flag.Bool("enable-rate-limiter", false, "synchronize the planned data with a requester rate limiter, limiting with a limit set as 1000 requests per 30 seconds")
var AllOperationFlag = flag.Bool("all", false, "enable the visualization of the planned data for all operations (create, update and delete)")
var AddOperationFlag = flag.Bool("add", false, "enable the visualization of the planned data for the create operation")
var UpdateOperationFlag = flag.Bool("update", false, "enable the visualization of the planned data for the update operation")
var DeleteOperationFlag = flag.Bool("delete", false, "enable the visualization of the planned data for the delete operation")
var DeleteGroupsNotInIdPFlag = flag.Bool("delete-groups-missing-in-idp", false, "delete groups present in SDM but not in the selected Identity Provider")
var DeleteUsersNotInIdPFlag = flag.Bool("delete-users-missing-in-idp", false, "delete users present in SDM but not in the selected Identity Provider")

func ValidateMandatoryFlags() error {
	if _, ok := validIdPs[*IdPFlag]; !ok {
		return errors.New("you must specify one Identity Provider (IdP): Google\nUse -idp \"google\"")
	} else if !*AllOperationFlag && !*AddOperationFlag && !*UpdateOperationFlag && !*DeleteOperationFlag {
		return errors.New("you must specify one of the following flags: \"-all\", \"-add\", \"-update\", \"-delete\"")
	}
	return nil
}
