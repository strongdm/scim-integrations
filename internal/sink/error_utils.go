package sink

import (
	"strings"
)

var commonErrMessages = []string{
	"values are already in use",
	"cannot parse member id",
}

func ErrorIsUnexpected(err error) bool {
	for _, msg := range commonErrMessages {
		if strings.Contains(err.Error(), msg) {
			return false
		}
	}
	return true
}
