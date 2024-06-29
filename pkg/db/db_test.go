package db

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestConnectSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	config := Config{
		DBName:   "testdb",
		UserName: "testuser",
		Password: "testpass",
		Host:     "localhost",
		Port:     "5432",
		Schema:   "public",
	}

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}

	conn, err := Connect(config)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMain(m *testing.M) {
	defer func() {
		sqlOpen = sql.Open
	}()

	m.Run()
}
