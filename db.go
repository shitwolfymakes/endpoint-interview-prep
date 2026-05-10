package main

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schemaSQL string

//go:embed sql/seed.sql
var seedSQL string

// openDB opens a SQLite database at path and turns on foreign-key enforcement.
// Pass ":memory:" for an in-memory database (used by tests).
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// applySchema creates all tables and indexes if they do not already exist.
func applySchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

// applySeed inserts the deterministic seed data. Tests rely on the exact
// rows in seed.sql, so do not run this against a database that already has
// data.
func applySeed(db *sql.DB) error {
	_, err := db.Exec(seedSQL)
	return err
}
