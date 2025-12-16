package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	mu     sync.RWMutex
	dbConn *sql.DB
)

// InitSQLite opens a global SQLite connection.
func InitSQLite(ctx context.Context, dbPath string) error {
	mu.RLock()
	if dbConn != nil {
		mu.RUnlock()
		return nil
	}
	mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	dsn := sqliteDSN(dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := conn.PingContext(pingCtx); err != nil {
		_ = conn.Close()
		return fmt.Errorf("ping sqlite: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if dbConn != nil {
		_ = conn.Close()
		return nil
	}
	dbConn = conn
	return nil
}

// DB returns the global SQLite connection.
func DB() (*sql.DB, error) {
	mu.RLock()
	defer mu.RUnlock()
	if dbConn == nil {
		return nil, fmt.Errorf("sqlite not initialized")
	}
	return dbConn, nil
}

func CloseSQLite() error {
	mu.Lock()
	defer mu.Unlock()
	if dbConn == nil {
		return nil
	}
	err := dbConn.Close()
	dbConn = nil
	return err
}

func sqliteDSN(dbPath string) string {
	// Pragmas:
	// - busy_timeout: avoid "database is locked" for short lock contention
	// - foreign_keys: enforce FK constraints (future-proof)
	// - journal_mode=WAL: better concurrency characteristics for a service
	return "file:" + dbPath +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(1)" +
		"&_pragma=journal_mode(WAL)"
}
