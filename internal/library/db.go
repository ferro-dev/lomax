// Package library is lomax's local track database: a per-user SQLite file
// (pure-Go via modernc.org/sqlite, no cgo — see docs/music-cli-plan.md
// section 6) that records which files are managed, so import/retag can
// keep it current and query/verify can read it back. See section 8's
// architecture diagram ("Library DB") and Milestone 4.
package library

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	goose "github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

//go:embed migrations/*.sql
var migrations embed.FS

// driverName is the database/sql driver name modernc.org/sqlite registers
// itself under.
const driverName = "sqlite"

// DB is a handle to lomax's library database.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if necessary) the SQLite database at path and
// applies any pending goose migrations. Callers must Close the returned DB.
func Open(path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("library: create %s: %w", dir, err)
		}
	}

	sqlDB, err := sql.Open(driverName, path)
	if err != nil {
		return nil, fmt.Errorf("library: open %s: %w", path, err)
	}
	// SQLite handles one writer at a time; a single connection avoids
	// SQLITE_BUSY errors from this process's own concurrent goroutines
	// without needing WAL-mode plumbing for what is, at this milestone,
	// low-concurrency CLI usage.
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("library: open %s: %w", path, err)
	}

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger()) // migration progress is an implementation detail, not CLI output
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("library: set migration dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("library: migrate %s: %w", path, err)
	}

	return &DB{sql: sqlDB}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.sql.Close()
}

// defaultStateDirEnv and defaultConfigStateEnv follow the XDG layout
// documented in docs/music-cli-plan.md section 7: the library database is
// process state, so it lives under XDG_STATE_HOME, not XDG_DATA_HOME.
const (
	stateDirEnv = "LOMAX_STATE_DIR"
	dbFileName  = "library.db"
)

// DefaultPath returns the library database's default location:
// $LOMAX_STATE_DIR/library.db if set, else
// $XDG_STATE_HOME/lomax/library.db, else $HOME/.local/state/lomax/library.db.
func DefaultPath() (string, error) {
	if dir := os.Getenv(stateDirEnv); dir != "" {
		return filepath.Join(dir, dbFileName), nil
	}
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "lomax", dbFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("library: determine default database path: %w", err)
	}
	return filepath.Join(home, ".local", "state", "lomax", dbFileName), nil
}
