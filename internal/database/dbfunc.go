package database

import (
	"database/sql"
	"os"
)

const createTable string = `
	CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			date TEXT(8) NOT NULL,
			title TEXT(50) NOT NULL,
			comment TEXT(100),
			repeat TEXT(128)
	);
	CREATE INDEX IF NOT EXISTS scheduler_date_IDX ON scheduler (date);
`

func OpenDB(dsn string) (*sql.DB, error) {
	_, err := os.Stat(dsn)
	install := err != nil

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	if install {
		if _, err := db.Exec(createTable); err != nil {
			return nil, err
		}
	}
	return db, nil
}
