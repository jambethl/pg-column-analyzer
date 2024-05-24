package cmd

import (
	"fmt"
	"log"

	"main/pkg/db"
	"main/pkg/report"

	"github.com/spf13/cobra"
)

var (
	dbName     string
	userName   string
	password   string
	host       string
	schemaName string
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A CLI tool for PostgreSQL column order optimization",
	Run: func(cmd *cobra.Command, arg []string) {
		configureDatabase()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dbName, "database", "d", "", "Database name (required)")
	rootCmd.PersistentFlags().StringVarP(&userName, "username", "u", "", "Username (required)")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password (required)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "h", "localhost", "Host")
	rootCmd.PersistentFlags().StringVarP(&schemaName, "schema", "s", "public", "Schema name")

	rootCmd.MarkPersistentFlagRequired("database")
	rootCmd.MarkPersistentFlagRequired("username")
	rootCmd.MarkPersistentFlagRequired("password")
}

func configureDatabase() {
	dbConfig := db.Config{}

	connection, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf()
	}

	defer connection.Close()

	if err := report.GenerateReport(connection); err != nil {
		log.Fatalf()
	}
}
