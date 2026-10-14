package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// OpenDB opens (and creates) the SQLite file, runs migrations.
func OpenDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open',
			reporter TEXT NOT NULL DEFAULT '',
			assignee TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT 'manual',
			revit_element_id TEXT DEFAULT '',
			ifc_global_id TEXT DEFAULT '',
			model_id INTEGER DEFAULT NULL,
			screenshot TEXT DEFAULT '',
			position TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			uploaded_at DATETIME NOT NULL DEFAULT (datetime('now')),
			element_count INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS elements (
			model_id INTEGER NOT NULL,
			global_id TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			storey TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (model_id, global_id)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
