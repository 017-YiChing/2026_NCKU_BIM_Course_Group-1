package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'client',
			avatar_url TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open',
			reporter TEXT NOT NULL DEFAULT '',
			reporter_id INTEGER DEFAULT NULL,
			assignee TEXT NOT NULL DEFAULT '',
			assignee_id INTEGER DEFAULT NULL,
			source TEXT NOT NULL DEFAULT 'manual',
			revit_element_id TEXT DEFAULT '',
			ifc_global_id TEXT DEFAULT '',
			model_id INTEGER DEFAULT NULL,
			screenshot TEXT DEFAULT '',
			position TEXT DEFAULT '',
			view_name TEXT DEFAULT '',
			revit_view_id TEXT DEFAULT '',
			model_name TEXT DEFAULT '',
			archived INTEGER NOT NULL DEFAULT 0,
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
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	// 舊庫升級欄位（已存在就跳過）
	for _, col := range []string{
		`ALTER TABLE issues ADD COLUMN view_name TEXT DEFAULT ''`,
		`ALTER TABLE issues ADD COLUMN revit_view_id TEXT DEFAULT ''`,
		`ALTER TABLE issues ADD COLUMN model_name TEXT DEFAULT ''`,
		`ALTER TABLE issues ADD COLUMN archived INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE issues ADD COLUMN reporter_id INTEGER DEFAULT NULL`,
		`ALTER TABLE issues ADD COLUMN assignee_id INTEGER DEFAULT NULL`,
	} {
		if _, err := db.Exec(col); err != nil && !isDupColumnErr(err) {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	// 種子用戶（若表空就插入，demo 密碼皆為 password123）
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if count == 0 {
		seedUsers(db)
	}
	return nil
}

func seedUsers(db *sql.DB) {
	users := []struct {
		email, password, name, role string
	}{
		{"james@example.com", "password123", "James Rose", "owner"},
		{"bob@example.com", "password123", "Bob Smith", "client"},
		{"sally@example.com", "password123", "Sally Smith", "client"},
	}
	for _, u := range users {
		db.Exec(`INSERT INTO users (email, password_hash, name, role) VALUES (?,?,?,?)`,
			u.email, simpleHash(u.password), u.name, u.role)
	}
}

func simpleHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func isDupColumnErr(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate column name")
}
