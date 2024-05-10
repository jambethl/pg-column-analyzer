package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const COLUMN_LIST_ORDER_QUERY string = `
SELECT
    ordinal_position,
    column_name,
    data_type
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

		fmt.Print("Enter your database user: ")

		dbUser, _ := read_user_input()

		fmt.Print("Enter your database name: ")

		dbName, _ := read_user_input()

		fmt.Printf("Enter your database password: ")

		dbPwd, _ := read_user_input()

		fmt.Printf("Enter your database host: ")

		dbHost, _ := read_user_input()

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

	fmt.Printf("Enter your database schema name (e.g. public): ")

	db_schema, _ := read_user_input()

	tables, err := db.Queryx(fmt.Sprintf(ALL_TABLES_IN_SCHEMA_QUERY, strings.TrimSuffix(db_schema, "\n")))
	if err != nil {
		log.Fatalln(err)
	}

	defer tables.Close()

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

		fmt.Println("Ordinal Position\tColumn Name\tData Type") // Just for testing purposes
		for columns.Next() {
			var ordinal_position int
			var column_name string
			var data_type string
			if err := columns.Scan(&ordinal_position, &column_name, &data_type); err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("%d\t\t%s\t\t%s\n", ordinal_position, column_name, data_type) // Just for testing purposes
		}

		if err := columns.Err(); err != nil {
			log.Fatalln(err)
		}

	}
}

func read_user_input() (string, error) {
	return user_input_reader.ReadString('\n')
}
