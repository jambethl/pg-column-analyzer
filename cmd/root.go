package cmd

import (
	"fmt"
	"log"

	"main/pkg/db"
	"main/pkg/report"

	"github.com/spf13/cobra"
)

var (
	db_name     string
	user_name   string
	password    string
	host        string
	schema_name string
)

var root_cmd = &cobra.Command{
	Use:   "cli",
	Short: "A CLI tool for PostgreSQL column order optimization",
	Run: func(cmd *cobra.Command, arg []string) {
		configure_database()
	},
}

func Execute() {
	if err := root_cmd.Execute(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}

func init() {
	root_cmd.PersistentFlags().StringVarP(&db_name, "database", "d", "", "Database name (required)")
	root_cmd.PersistentFlags().StringVarP(&user_name, "username", "u", "", "Username (required)")
	root_cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password (required)")
	root_cmd.PersistentFlags().StringVarP(&host, "host", "h", "localhost", "Host")
	root_cmd.PersistentFlags().StringVarP(&schema_name, "schema", "s", "public", "Schema name")

	root_cmd.MarkPersistentFlagRequired("database")
	root_cmd.MarkPersistentFlagRequired("username")
	root_cmd.MarkPersistentFlagRequired("password")
}

func configure_database() {
	db_config := db.Config{}

	connection, err := db.Connect(db_config)
	if err != nil {
		log.Fatalf()
	}

	defer connection.Close()

	if err := report.GenerateReport(connection); err != nil {
		log.Fatalf()
	}
}
