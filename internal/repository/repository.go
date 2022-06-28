package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type table struct {
	name string
	body string
}

var (
	reportTable = "reports"
	errorsTable = "errors"
	tables      = []*table{
		{reportTable, `
			id integer primary key autoincrement,
			start_time timestamp,
			end_time timestamp,
			users_created integer,
			users_updated integer,
			users_deleted integer,
			groups_created integer,
			groups_updated integer,
			groups_deleted integer,
			status integer
		`},
	}
)

func init() {
	if !isDatabaseEnabled() {
		return
	}
	err := setupDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func setupDB() error {
	conn, err := getConnection()
	if err != nil {
		return err
	}
	for _, table := range tables {
		_, err := conn.Exec(createTableQuery(table))
		if err != nil {
			return err
		}
	}
	return nil
}

func createTableQuery(table *table) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`, table.name, table.body)
}

func getConnection() (*sql.DB, error) {
	dbFilePath := os.Getenv("SDM_SCIM_REPORTS_DATABASE_PATH")
	conn, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func execQuery(query string, args ...interface{}) (*sql.Rows, error) {
	conn, err := getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	return rows, nil
}

func exec(query string, args ...interface{}) (sql.Result, error) {
	conn, err := getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	result, err := conn.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func isDatabaseEnabled() bool {
	return os.Getenv("CGO_ENABLED") == "1" && os.Getenv("SDM_SCIM_REPORTS_DATABASE_PATH") != ""
}
