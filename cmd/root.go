package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"main/pkg/common"
	_ "main/pkg/common"
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

const ColumnCountQuery string = `
SELECT COUNT('%s') FROM %s;
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
	rootCmd.PersistentFlags().StringVarP(&dbName, "database", "d", "postgres", "Database name")
	rootCmd.PersistentFlags().StringVarP(&userName, "username", "u", "postgres", "Username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "123", "Password")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "l", "localhost", "Host")
	rootCmd.PersistentFlags().StringVarP(&schemaName, "schema", "s", "public", "Schema name")
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

		var columnList []common.ColumnInfo
		for columns.Next() {
			var colInfo common.ColumnInfo
			if err := columns.Scan(&colInfo.OrdinalPosition, &colInfo.ColumnName, &colInfo.DataType, &colInfo.IsNullable); err != nil {
				log.Fatalln(err)
			}
			colInfo.EntryCount = calculateTotalEntries(colInfo.ColumnName, tableName, connection)
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

func calculateTotalEntries(columnName string, tableName string, connection *sql.DB) int {
	var count int
	err := connection.QueryRow(fmt.Sprintf(ColumnCountQuery, columnName, tableName)).Scan(&count)
	if err != nil {
		log.Fatalln(err)
	}
	return count

}
