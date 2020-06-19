package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/sarunask/psql_admin_utils/pkg/postgres"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	pgHost     string
	pgPort     int
	pgUser     string
	pgPass     string
	pgDatabase string
	pgSchemas  []string
	pgNewOwner string
	pgVerbose  bool
	pgTLS      bool
)

func init() {
	rootCmd.AddCommand(changeOwnerCmd)

	changeOwnerCmd.Flags().StringVar(&pgHost, "host", "", "PostgreSQL DBMS host")
	changeOwnerCmd.Flags().IntVarP(&pgPort, "port", "p", 5432, "PostgreSQL DBMS port")
	changeOwnerCmd.Flags().StringVarP(&pgUser, "dbuser", "U", "postgres", "PostgreSQL DBMS root user")
	changeOwnerCmd.Flags().StringVar(&pgPass, "password", "", "PostgreSQL DBMS root user password")
	changeOwnerCmd.Flags().StringVarP(&pgDatabase, "database", "d", "", "Database name, owner role of which you would like to change")
	changeOwnerCmd.Flags().StringSliceVarP(&pgSchemas, "schemas", "S", []string{}, "Schemas in database you wish to change owner too (including all objects inside)")
	changeOwnerCmd.Flags().StringVarP(&pgNewOwner, "new-owner", "o", "", "Database role, to which you would like to change database ownership")
	changeOwnerCmd.Flags().BoolVar(&pgTLS, "tls", false, "Use TLS to connect to DMBS")
	changeOwnerCmd.Flags().BoolVarP(&pgVerbose, "verbose", "v", false, "Do verbose output, show SQL statments")
	checkErr(viper.BindPFlag("host", changeOwnerCmd.Flags().Lookup("host")))
	checkErr(viper.BindPFlag("port", changeOwnerCmd.Flags().Lookup("port")))
	checkErr(viper.BindPFlag("dbuser", changeOwnerCmd.Flags().Lookup("dbuser")))
	checkErr(viper.BindPFlag("password", changeOwnerCmd.Flags().Lookup("password")))
	checkErr(viper.BindPFlag("database", changeOwnerCmd.Flags().Lookup("database")))
	checkErr(viper.BindPFlag("schemas", changeOwnerCmd.Flags().Lookup("schemas")))
	checkErr(viper.BindPFlag("new-owner", changeOwnerCmd.Flags().Lookup("new-owner")))
	checkErr(viper.BindPFlag("tls", changeOwnerCmd.Flags().Lookup("tls")))
	checkErr(viper.BindPFlag("verbose", changeOwnerCmd.Flags().Lookup("verbose")))
}

var changeOwnerCmd = &cobra.Command{
	Use:     "change_owner",
	Aliases: []string{"chown"},
	Short:   "Command will change owner of database",
	Long: `Command will change owner of database and schemas,
user defined types, tables, sequences, views, materialized views,
indexes and functions within that database`,
	Run: changeOwner,
}

func initVarsFromConfig() {
	pgHost = viper.GetString("host")
	pgPort = viper.GetInt("port")
	pgUser = viper.GetString("dbuser")
	pgPass = viper.GetString("password")
	pgDatabase = viper.GetString("database")
	pgSchemas = viper.GetStringSlice("schemas")
	pgNewOwner = viper.GetString("new-owner")
	pgVerbose = viper.GetBool("verbose")
	pgTLS = viper.GetBool("tls")
	if len(pgHost) == 0 || len(pgDatabase) == 0 ||
		len(pgSchemas) == 0 || len(pgNewOwner) == 0 {
		checkErr(fmt.Errorf("host,database,schemas and new owner are required"))
	}
}

func changeOwner(cmd *cobra.Command, args []string) {
	initVarsFromConfig()
	if len(pgPass) == 0 {
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't read password: %v", err)
			return
		}
		pgPass = string(bytePassword)
	}
	cfg := &postgres.Config{
		Host:         pgHost,
		Port:         pgPort,
		Db:           pgDatabase,
		User:         pgUser,
		Password:     pgPass,
		Schemas:      pgSchemas,
		TLS:          pgTLS,
		Verbose:      pgVerbose,
		WaitDuration: 20 * time.Second,
	}
	pg, err := postgres.New(cfg, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't connect to PostgreSQL: %v", err)
		return
	}
	defer pg.Close()
	err = pg.HealthCheck()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't health check PostgreSQL: %v", err)
		return
	}
	err = pg.ChangeOwnerForDB(pgNewOwner, pgDatabase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't change owner for '%s' and all objects in it to '%s': %v",
			pgDatabase, pgNewOwner, err)
		return
	}
}
