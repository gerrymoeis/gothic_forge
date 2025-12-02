package providers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// SQLiteProvider implements DatabaseProvider for SQLite (local development/testing)
type SQLiteProvider struct {
	dbPath string
}

// NewSQLiteProvider creates a new SQLite provider
func NewSQLiteProvider(dbPath string) *SQLiteProvider {
	if dbPath == "" {
		dbPath = "gothic_forge.db"
	}
	return &SQLiteProvider{
		dbPath: dbPath,
	}
}

// Name returns the provider name
func (p *SQLiteProvider) Name() string {
	return "sqlite"
}

// Provision creates a new SQLite database file
func (p *SQLiteProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
	// Ensure directory exists
	dir := filepath.Dir(p.dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create database file (will be created on first connection)
	connStr := fmt.Sprintf("file:%s?cache=shared&mode=rwc", p.dbPath)

	// Test connection
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	absPath, _ := filepath.Abs(p.dbPath)

	return &DatabaseInfo{
		ID:               "sqlite-local",
		Name:             opts.Name,
		ConnectionString: connStr,
		Host:             "localhost",
		Port:             0,
		Database:         absPath,
		Username:         "",
		Password:         "",
		Region:           "local",
		CreatedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"path": absPath,
			"mode": "rwc",
		},
	}, nil
}

// GetConnectionString returns the SQLite connection string
func (p *SQLiteProvider) GetConnectionString(ctx context.Context) (string, error) {
	// Check if DATABASE_URL is set (for compatibility)
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		return connStr, nil
	}

	// Return default connection string
	return fmt.Sprintf("file:%s?cache=shared&mode=rwc", p.dbPath), nil
}

// RunMigrations executes database migrations using goose
func (p *SQLiteProvider) RunMigrations(ctx context.Context, migrationsDir string) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set goose dialect
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Health checks if the database is reachable
func (p *SQLiteProvider) Health(ctx context.Context) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	return db.PingContext(ctx)
}

// Destroy removes the SQLite database file
func (p *SQLiteProvider) Destroy(ctx context.Context) error {
	// Remove database file
	if err := os.Remove(p.dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove database file: %w", err)
	}

	// Remove WAL and SHM files if they exist
	os.Remove(p.dbPath + "-wal")
	os.Remove(p.dbPath + "-shm")

	return nil
}
