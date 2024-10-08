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
	Port     string
}

var sqlOpen = sql.Open

func Connect(config Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", config.Host, config.Port, config.DBName, config.UserName, config.Password)

	db, err := sqlOpen("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
