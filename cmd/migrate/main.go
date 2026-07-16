// Command migrate applies (or rolls back) the SQL files in ./migrations
// against the configured MySQL database. It's a deliberately small
// alternative to golang-migrate: no extra dependency, one file per
// migration direction, tracked in a schema_migrations table.
//
// Usage:
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate down   (rolls back the most recent migration)
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"postsapi/internal/config"
	"postsapi/internal/database"
)

const migrationsDir = "migrations"

func main() {
	if len(os.Args) < 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Println("usage: go run ./cmd/migrate [up|down]")
		os.Exit(1)
	}
	direction := os.Args[1]

	cfg := config.Load()
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := ensureMigrationsTable(db); err != nil {
		log.Fatalf("could not prepare schema_migrations table: %v", err)
	}

	if direction == "up" {
		if err := migrateUp(db); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
		return
	}

	if err := migrateDown(db); err != nil {
		log.Fatalf("migrate down failed: %v", err)
	}
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func migrateUp(db *sql.DB) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		version := versionFromFilename(f)

		var exists int
		err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("applying %s: %w", f, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			return err
		}

		log.Printf("applied %s", f)
	}
	return nil
}

func migrateDown(db *sql.DB) error {
	var version string
	err := db.QueryRow(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		log.Println("no migrations to roll back")
		return nil
	}
	if err != nil {
		return err
	}

	downFile := filepath.Join(migrationsDir, version+".down.sql")
	sqlBytes, err := os.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", downFile, err)
	}

	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("applying %s: %w", downFile, err)
	}
	if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = ?`, version); err != nil {
		return err
	}

	log.Printf("rolled back %s", downFile)
	return nil
}

// versionFromFilename turns "migrations/000001_create_posts_table.up.sql"
// into "000001_create_posts_table".
func versionFromFilename(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".up.sql")
}
