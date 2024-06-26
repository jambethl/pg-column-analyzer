package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"main/pkg/common"

	"main/pkg/db"
	"main/pkg/report"

	"github.com/spf13/cobra"
)

const (
	ColumnListOrderQuery = `
		SELECT ordinal_position, column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = '%s' AND table_name = '%s'
		ORDER BY ordinal_position;`

	ColumnCountQuery = `SELECT COUNT('%s') FROM %s;`

	AllTablesInSchemaQuery = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = '%s';`
)

var (
	dbName     string
	userName   string
	password   string
	host       string
	schemaName string
	port       string
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
	rootCmd.PersistentFlags().StringVarP(&port, "port", "t", "5432", "Port")
}

func configureDatabase() {
	dbConfig := db.Config{
		DBName:   dbName,
		UserName: userName,
		Password: password,
		Host:     host,
		Schema:   schemaName,
		Port:     port,
	}

	connection, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer connection.Close()

	tables, err := fetchTables(connection, schemaName)
	if err != nil {
		log.Fatalf("Failed to fetch tables: %v", err)
	}

	if err := os.MkdirAll("reports", 0o755); err != nil {
		log.Fatalf("Failed to create reports directory: %v", err)
	}

	for _, table := range tables {
		columnList, err := fetchColumns(connection, schemaName, table)
		if err != nil {
			log.Fatalf("Failed to fetch columns for table %s: %v", table, err)
		}

		for i := range columnList {
			columnList[i].EntryCount = calculateTotalEntries(connection, table, columnList[i].ColumnName)
		}

		sort.SliceStable(columnList, func(i, j int) bool {
			return columnList[i].OrdinalPosition < columnList[j].OrdinalPosition
		})
		if err := report.GenerateReport(columnList, table); err != nil {
			log.Fatalf("Failed to generate report for table %s: %v", table, err)
		}
	}
}

func fetchTables(connection *sql.DB, schemaName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := connection.QueryContext(ctx, fmt.Sprintf(AllTablesInSchemaQuery, schemaName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func fetchColumns(connection *sql.DB, schemaName string, tableName string) ([]common.ColumnInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := connection.QueryContext(ctx, fmt.Sprintf(ColumnListOrderQuery, schemaName, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []common.ColumnInfo
	for rows.Next() {
		var colInfo common.ColumnInfo
		if err := rows.Scan(&colInfo.OrdinalPosition, &colInfo.ColumnName, &colInfo.DataType, &colInfo.IsNullable); err != nil {
			return nil, err
		}
		columns = append(columns, colInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func calculateTotalEntries(connection *sql.DB, tableName string, columnName string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(ColumnCountQuery, columnName, tableName)
	var count int
	err := connection.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to calculate total entries for column %s in table %s: %v", columnName, tableName, err)
	}
	return count
}
