package flags

import (
	"errors"
	"flag"
)

var OktaFlag = flag.Bool("okta", false, "use Okta as IdP")
var GoogleFlag = flag.Bool("google", false, "use Google as IdP")
var DeleteUnmatchingGroupsFlag = flag.Bool("delete-unmatching-groups", false, "delete groups present in SDM but not in matchers.yml")
var DeleteUnmatchingUsersFlag = flag.Bool("delete-unmatching-users", false, "delete users present in SDM but not in the selected IdP or assigned to any group in matchers.yml")
var JsonFlag = flag.Bool("json", false, "dump a JSON report for debugging")
var PlanFlag = flag.Bool("plan", false, "do not apply changes just show initial queries")

func init() {
	flag.Parse()
}

func ValidateMandatoryFlags() error {
	if (!*OktaFlag && !*GoogleFlag) || (*OktaFlag && *GoogleFlag) {
		return errors.New("you need to specify one Identity Provider (IdP): Okta or Google\nUse -okta or -google")
	}
	return nil
}
