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
	DATABASE_USER_PROMPT     = "Enter your database user: "
	DATABASE_NAME_PROMT      = "Enter your database name: "
	DATABASE_PASSWORD_PROMPT = "Enter your database password: "
	DATABASE_HOST_PROMPT     = "Enter your database host: "
	DATABASE_SCHEMA_PROMPT   = "Enter your database schema name (e.g. public): "
)

const COLUMN_LIST_ORDER_QUERY string = `
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

const ALL_TABLES_IN_SCHEMA_QUERY string = `
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

var user_input_reader *bufio.Reader

func init() {
	user_input_reader = bufio.NewReader(os.Stdin)
}

func main() {
	default_flag := flag.Bool("defaultConfig", false, "Whether to use default Postgres connection properties")

	flag.Parse()

	default_value := *default_flag

	var connection_string string
	if default_value {
		fmt.Println("Running with default Postgres connection properties")
		connection_string = "user=postgres dbname=postgres sslmode=disable password=123 host=localhost"
	} else {

		dbUser, _ := prompt_user_input(DATABASE_USER_PROMPT)
		dbName, _ := prompt_user_input(DATABASE_NAME_PROMT)
		dbPwd, _ := prompt_user_input(DATABASE_PASSWORD_PROMPT)
		dbHost, _ := prompt_user_input(DATABASE_HOST_PROMPT)

		connection_string = fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s", dbUser, dbName, dbPwd, dbHost)
	}

	db, err := sqlx.Connect("postgres", connection_string)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected")
	}

	defer db.Close()

	db_schema, _ := prompt_user_input(DATABASE_SCHEMA_PROMPT)

	tables, err := db.Queryx(fmt.Sprintf(ALL_TABLES_IN_SCHEMA_QUERY, strings.TrimSuffix(db_schema, "\n")))
	if err != nil {
		log.Fatalln(err)
	}

	defer tables.Close()

	dir_err := os.MkdirAll("reports", 0o755)
	if dir_err != nil {
		fmt.Println(dir_err)
		return
	}

	for tables.Next() {
		var table_name string
		if err := tables.Scan(&table_name); err != nil {
			log.Fatalln(err)
		}

		columns, err := db.Queryx(fmt.Sprintf(COLUMN_LIST_ORDER_QUERY, strings.TrimSuffix(db_schema, "\n"), table_name))
		if err != nil {
			log.Fatalln(err)
		}
		defer columns.Close()

		var column_list []ColumnInfo
		for columns.Next() {
			var col_info ColumnInfo
			if err := columns.Scan(&col_info.OrdinalPosition, &col_info.ColumnName, &col_info.DataType, &col_info.IsNullable); err != nil {
				log.Fatalln(err)
			}
			column_list = append(column_list, col_info)
		}

		if err := columns.Err(); err != nil {
			log.Fatalln(err)
		}

		sort.SliceStable(column_list, func(i, j int) bool {
			if column_list[i].IsNullable != column_list[j].IsNullable {
				return column_list[i].IsNullable == "NO"
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
			return typeSize(column_list[i].DataType) > typeSize(column_list[j].DataType)
		})

		report_name := fmt.Sprintf("reports/%s_report.csv", table_name)
		file, err := os.Create(report_name)
		if err != nil {
			log.Fatal("Unable to create file: ", err)
		}

		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable"})

		for _, col := range column_list {
			writer.Write([]string{
				fmt.Sprint(col.OrdinalPosition),
				col.ColumnName,
				col.DataType,
				col.IsNullable,
			})
		}

		fmt.Printf("Report %s generated successfully.\n", report_name)
	}
}

func prompt_user_input(prompt_text string) (string, error) {
	fmt.Print(prompt_text)

	return user_input_reader.ReadString('\n')
}
