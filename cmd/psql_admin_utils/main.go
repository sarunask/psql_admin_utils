package main

import (
	"fmt"
	"os"

	"github.com/sarunask/psql_admin_utils/pkg/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
