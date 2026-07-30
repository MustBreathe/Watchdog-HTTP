package storage

import (
	"database/sql"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Open(out io.Writer) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", os.Getenv("WATCHDOG_DATABASE_DSN"))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if err := runMigrations(db, out); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
