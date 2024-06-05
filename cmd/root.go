package cmd

import (
	"fmt"
	"log"
	"main/pkg/types"
	"os"
	"sort"

	"main/pkg/db"
	"main/pkg/report"

	"github.com/spf13/cobra"
)

const ColumnListOrderQuery string = `
SELECT
    ordinal_position,
    column_name,
    data_type,
    is_nullable
FROM
    information_schema.columns
WHERE
    table_schema = 'public'
    AND table_name = '%s'
ORDER BY
    ordinal_position;
`

const AllTablesInSchemaQuery string = `
SELECT
    table_name
FROM
    information_schema.tables
WHERE
    table_schema = 'public';
`

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
	rootCmd.PersistentFlags().StringVarP(&host, "host", "l", "localhost", "Host")
	rootCmd.PersistentFlags().StringVarP(&schemaName, "schema", "s", "public", "Schema name")

	rootCmd.MarkPersistentFlagRequired("database")
	rootCmd.MarkPersistentFlagRequired("username")
	rootCmd.MarkPersistentFlagRequired("password")
}

func configureDatabase() {
	dbConfig := db.Config{
		DBName:   dbName,
		UserName: userName,
		Password: password,
		Host:     host,
		Schema:   schemaName,
	}

	connection, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer connection.Close()

	tables, err := connection.Query(fmt.Sprintf(AllTablesInSchemaQuery))
	if err != nil {
		log.Fatalln(err)
	}

	defer tables.Close()

	dirErr := os.MkdirAll("reports", 0o755)
	if dirErr != nil {
		fmt.Println(dirErr)
		return
	}

	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			log.Fatalln(err)
		}

		columns, err := connection.Query(fmt.Sprintf(ColumnListOrderQuery, tableName))
		if err != nil {
			log.Fatalln(err)
		}
		defer columns.Close()

		var columnList []types.ColumnInfo
		for columns.Next() {
			var colInfo types.ColumnInfo
			if err := columns.Scan(&colInfo.OrdinalPosition, &colInfo.ColumnName, &colInfo.DataType, &colInfo.IsNullable); err != nil {
				log.Fatalln(err)
			}
			columnList = append(columnList, colInfo)
		}

		if err := columns.Err(); err != nil {
			log.Fatalln(err)
		}

		sort.SliceStable(columnList, func(i, j int) bool {
			return columnList[i].OrdinalPosition < columnList[j].OrdinalPosition
		})
		if err := report.GenerateReport(columnList, tableName); err != nil {
			log.Fatalf("Failed to generate report: %v", err)
		}
	}
}
