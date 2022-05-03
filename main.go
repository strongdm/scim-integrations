package main

import (
	"fmt"
	"os"
	"scim-integrations/internal/app"
	"scim-integrations/internal/cli"
)

func main() {
	if executed, err := cli.ResolveCommand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	} else if executed {
		os.Exit(0)
	}
	app.Start()
}
