package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "psql_admin_utils",
		Short: "Admin Tool for PostgreSQL database",
		Long: `psql_admin_utils is a CLI tool for admins of PostgreSQL database.
This tool can now change owner of database and all objects inside of that particular database:
schemas, tables, views, sequences, triggers.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.psql_admin_utils.yaml)")
	checkErr(viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")))
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		checkErr(err)

		// Search config in home directory with name ".psql_admin_utils" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".psql_admin_utils")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
