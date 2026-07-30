// internal/storage/migrations.go
package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"io"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func runMigrations(db *sql.DB, out io.Writer) error {
	fmt.Fprintln(out, "Running database migrations...")

	source, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	driver, err := migratesqlite3.WithInstance(db, &migratesqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance(
		"iofs",
		source,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer migrator.Close()

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	fmt.Fprintln(out, "Database migrations completed")

	return nil
}
