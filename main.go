package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ColumnInfo struct {
	OrdinalPosition int
	ColumnName      string
	DataType        string
	IsNullable      string
}

const (
	DatabaseUserPrompt     = "Enter your database user: "
	DatabaseNamePrompt     = "Enter your database name: "
	DatabasePasswordPrompt = "Enter your database password: "
	DatabaseHostPrompt     = "Enter your database host: "
	DatabaseSchemaPrompt   = "Enter your database schema name (e.g. public): "
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
    table_schema = '%s'
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
    table_schema = '%s';
`

const (
	BIGINT = iota
	BIGSERIAL
	INTEGER
	REAL
	SMALLINT
	SMALLSERIAL
	SERIAL
)

var data_type_byte_map = map[int]int{
	BIGINT:      8,
	BIGSERIAL:   8,
	INTEGER:     4,
	REAL:        4,
	SMALLINT:    4,
	SMALLSERIAL: 2,
	SERIAL:      4,
}

var userInputReader *bufio.Reader

func init() {
	userInputReader = bufio.NewReader(os.Stdin)
}

func main() {
	defaultFlag := flag.Bool("defaultConfig", false, "Whether to use default Postgres connection properties")

	flag.Parse()

	defaultValue := *defaultFlag

	var connectionString string
	if defaultValue {
		fmt.Println("Running with default Postgres connection properties")
		connectionString = "user=postgres dbname=postgres sslmode=disable password=123 host=localhost"
	} else {

		dbUser, _ := promptUserInput(DatabaseUserPrompt)
		dbName, _ := promptUserInput(DatabaseNamePrompt)
		dbPwd, _ := promptUserInput(DatabasePasswordPrompt)
		dbHost, _ := promptUserInput(DatabaseHostPrompt)

		connectionString = fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s", dbUser, dbName, dbPwd, dbHost)
	}

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected")
	}

	defer db.Close()

	dbSchema, _ := promptUserInput(DatabaseSchemaPrompt)

	tables, err := db.Queryx(fmt.Sprintf(AllTablesInSchemaQuery, strings.TrimSuffix(dbSchema, "\n")))
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

		columns, err := db.Queryx(fmt.Sprintf(ColumnListOrderQuery, strings.TrimSuffix(dbSchema, "\n"), tableName))
		if err != nil {
			log.Fatalln(err)
		}
		defer columns.Close()

		var columnList []ColumnInfo
		for columns.Next() {
			var colInfo ColumnInfo
			if err := columns.Scan(&colInfo.OrdinalPosition, &colInfo.ColumnName, &colInfo.DataType, &colInfo.IsNullable); err != nil {
				log.Fatalln(err)
			}
			columnList = append(columnList, colInfo)
		}

		if err := columns.Err(); err != nil {
			log.Fatalln(err)
		}

		sort.SliceStable(columnList, func(i, j int) bool {
			if columnList[i].IsNullable != columnList[j].IsNullable {
				return columnList[i].IsNullable == "NO"
			}

			typeSize := func(dataType string) int {
				switch dataType {
				case "bigint":
					return 8
				case "integer":
					return 4
				case "smallint":
					return 2
				case "boolean":
					return 1
				case "real":
					return 4
				case "double precision":
					return 8
				case "data":
					return 4
				case "timestamp without time zone", "timestamp with time zone":
					return 8
				case "text", "varchar", "bytea":
					return 10
				default:
					return 10
				}
			}
			return typeSize(columnList[i].DataType) > typeSize(columnList[j].DataType)
		})

		reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
		file, err := os.Create(reportName)
		if err != nil {
			log.Fatal("Unable to create file: ", err)
		}

		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable"})

		for _, col := range columnList {
			writer.Write([]string{
				fmt.Sprint(col.OrdinalPosition),
				col.ColumnName,
				col.DataType,
				col.IsNullable,
			})
		}

		fmt.Printf("Report %s generated successfully.\n", reportName)
	}
}

func promptUserInput(promptText string) (string, error) {
	fmt.Print(promptText)

	return userInputReader.ReadString('\n')
}
