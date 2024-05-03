package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const COLUMN_LIST_ORDER_QUERY string = `
SELECT
    ordinal_position,
    column_name,
    data_type
FROM
    information_schema.column
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
	fmt.Print("Enter your database user: ")

	dbUser, _ := read_user_input()

	fmt.Print("Enter your database name: ")

	dbName, _ := read_user_input()

	fmt.Printf("Enter your database password: ")

	dbPwd, _ := read_user_input()

	fmt.Printf("Enter your database host: ")

	dbHost, _ := read_user_input()

	connection_string := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s", dbUser, dbName, dbPwd, dbHost)

	db, err := sqlx.Connect("postgres", connection_string)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	fmt.Printf("Enter your database schema name (e.g. public): ")

	db_schema, _ := read_user_input()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected")
	}
}

func read_user_input() (string, error) {
	return user_input_reader.ReadString('\n')
}
