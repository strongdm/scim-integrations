package synchronizer

import (
	"strings"
)

// TODO: add comments saying what each message means - ideal solution:
// make a map for each error of each entity operation
var commonErrMessages = []string{
	"values are already in use",
	"cannot parse member id",
	"not found",
}

func ErrorIsUnexpected(err error) bool {
	for _, msg := range commonErrMessages {
		if strings.Contains(err.Error(), msg) {
			return false
		}
	}
	return true
}
