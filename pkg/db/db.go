package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Config struct {
	DBName   string
	UserName string
	Password string
	Host     string
	Schema   string
}

func Connect(config Config) (*sql.DB, error) {
	conn_str := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", config.Host, config.DBName, config.UserName, config.Password)
	return sql.Open("postrgres", conn_str)
}
