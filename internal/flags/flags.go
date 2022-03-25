package flags

import (
	"errors"
	"flag"
)

var GoogleFlag = flag.Bool("google", false, "use Google as a Source")
var DeleteUnmatchingGroupsFlag = flag.Bool("delete-unmatching-groups", false, "delete groups present in SDM but not in selected Source data")
var DeleteUnmatchingUsersFlag = flag.Bool("delete-unmatching-users", false, "delete users present in SDM but not in the selected Source data")
var JsonFlag = flag.Bool("json", false, "dump a JSON report for debugging")
var PlanFlag = flag.Bool("plan", false, "do not apply changes just show initial queries")

func init() {
	flag.Parse()
}

func ValidateMandatoryFlags() error {
	if !*GoogleFlag {
		return errors.New("you need to specify one Identity Provider (IdP/Source): Google\nUse -google")
	}
	return nil
}
