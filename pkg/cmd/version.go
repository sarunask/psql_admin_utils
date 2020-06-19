package cmd

import (
	"fmt"

	"github.com/sarunask/psql_admin_utils/pkg/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of psql_admin_utils",
	Long:  `All software has versions. This is psql_admin_utils one`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("psql_admin_utils v%s\n", version.Version)
	},
}
