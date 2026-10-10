package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewDB(dbPath string) (*sql.DB, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite single-writer safety
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		phone_number TEXT NOT NULL UNIQUE,
		avatar TEXT DEFAULT '',
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS family_groups (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		parent_id TEXT NOT NULL,
		invite_code TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS member_devices (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		device_uuid TEXT NOT NULL,
		device_name TEXT NOT NULL,
		avatar TEXT DEFAULT '',
		ipv4 TEXT,
		ipv6 TEXT,
		mac_address TEXT,
		battery_level INTEGER DEFAULT 100,
		is_active INTEGER DEFAULT 1,
		paired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME,
		UNIQUE(group_id, device_uuid),
		FOREIGN KEY (group_id) REFERENCES family_groups(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS location_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL,
		group_id TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		accuracy REAL DEFAULT 0.0,
		speed REAL DEFAULT 0.0,
		heading REAL DEFAULT 0.0,
		recorded_at INTEGER NOT NULL,
		synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (device_id) REFERENCES member_devices(id) ON DELETE CASCADE,
		FOREIGN KEY (group_id) REFERENCES family_groups(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_loc_group_time ON location_history(group_id, recorded_at DESC);
	CREATE INDEX IF NOT EXISTS idx_loc_device_time ON location_history(device_id, recorded_at DESC);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_loc_dedup ON location_history(device_id, recorded_at);
	`

	_, err := db.Exec(schema)
	return err
}
