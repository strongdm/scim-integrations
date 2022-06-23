package sdmscim

import (
	"fmt"
	"os"

	"github.com/strongdm/scimsdk"
)

var errorSign = fmt.Sprintf("\033[31mx\033[0m")

type sinkSDMSCIMImpl struct {
	client scimsdk.Client
}

func NewSinkSDMSCIM() *sinkSDMSCIMImpl {
	return &sinkSDMSCIMImpl{NewSDMSCIMClient()}
}

func NewSDMSCIMClient() scimsdk.Client {
	return scimsdk.NewClient(os.Getenv("SDM_SCIM_TOKEN"), nil)
}

func formatErrorMessage(message string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", errorSign, fmt.Sprintf(message, args...))
}
