package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"main/pkg/common"

	"main/pkg/db"
	"main/pkg/report"

	"github.com/spf13/cobra"
)

const (
	ColumnListOrderQuery = `
        SELECT 
            c.ordinal_position,
            c.column_name,
            c.data_type,
            c.is_nullable,
            t.typalign
        FROM 
            information_schema.columns c
        JOIN 
            pg_class pc ON pc.relname = c.table_name
        JOIN 
            pg_attribute a ON a.attrelid = pc.oid AND a.attname = c.column_name
        JOIN 
            pg_type t ON t.oid = a.atttypid
        WHERE 
            c.table_schema = '%s' 
            AND c.table_name = '%s'
        ORDER BY 
            c.ordinal_position;
        `

	ColumnCountQuery = `SELECT COUNT('%s') FROM %s;`

	AllTablesInSchemaQuery = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = '%s';`
)

var alignmentMap = map[string]int{
	"c": -1,
	"s": 2,
	"i": 4,
	"d": 8,
}

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
		generateReportForTable(connection, schemaName, table)
	}
}

func generateReportForTable(connection *sql.DB, schemaName string, table string) {
	columnList, err := fetchColumns(connection, schemaName, table)
	if err != nil {
		log.Fatalf("Failed to fetch columns for table %s: %v", table, err)
	}

	for i := range columnList {
		columnList[i].EntryCount = calculateTotalEntries(connection, table, columnList[i].ColumnName)
	}

	if err := report.GenerateReport(columnList, table); err != nil {
		log.Fatalf("Failed to generate report for table %s: %v", table, err)
	}
}

func fetchTables(connection *sql.DB, schemaName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := connection.QueryContext(ctx, fmt.Sprintf(AllTablesInSchemaQuery, schemaName))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tables, nil
}

func fetchColumns(connection *sql.DB, schemaName string, tableName string) ([]common.ColumnInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := connection.QueryContext(ctx, fmt.Sprintf(ColumnListOrderQuery, schemaName, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []common.ColumnInfo
	for rows.Next() {
		var colInfo common.ColumnInfo
		var typAlignRune string
		if err := rows.Scan(&colInfo.OrdinalPosition, &colInfo.ColumnName, &colInfo.DataType, &colInfo.IsNullable, &typAlignRune); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		alignmentValue, exists := alignmentMap[typAlignRune]
		if !exists {
			return nil, fmt.Errorf("failed to determine alignment value: %w", err)
		}
		colInfo.TypAlign = alignmentValue
		columns = append(columns, colInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return columns, nil
}

func calculateTotalEntries(connection *sql.DB, tableName string, columnName string) int {
	query := fmt.Sprintf(ColumnCountQuery, columnName, tableName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := connection.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to calculate total entries for column %s in table %s: %v", columnName, tableName, err)
	}
	return count
}
