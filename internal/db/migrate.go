package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     INTEGER PRIMARY KEY,
			applied_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var version int
		_, err := fmt.Sscanf(name, "%d", &version)
		if err != nil {
			return fmt.Errorf("parse version from %s: %w", name, err)
		}

		var exists bool
		err = db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == sql.ErrNoRows {
			content, err := migrationsFS.ReadFile("migrations/" + name)
			if err != nil {
				return fmt.Errorf("read migration %s: %w", name, err)
			}

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("begin tx for %s: %w", name, err)
			}

			_, err = tx.Exec(string(content))
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("exec migration %s: %w", name, err)
			}

			_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("record migration %s: %w", name, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit migration %s: %w", name, err)
			}

			slog.Info("applied migration", "version", version, "file", name)
		} else if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
	}

	return nil
}
