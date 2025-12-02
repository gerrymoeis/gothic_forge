package providers

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// PostgreSQLProvider implements DatabaseProvider for PostgreSQL (local/self-hosted).
//
// This provider is designed for local development and self-hosted PostgreSQL instances.
// It does not provision infrastructure automatically - it expects an existing PostgreSQL
// server to be available.
//
// The provider supports:
// - Connection via DATABASE_URL environment variable
// - Connection via individual PG_* environment variables (host, port, user, password, database)
// - Database migration support using goose
// - Health checks with timeout
// - PostgreSQL-specific features and optimizations
//
// Environment Variables:
// - DATABASE_URL: Full PostgreSQL connection string (preferred)
// - PG_HOST: PostgreSQL host (default: localhost)
// - PG_PORT: PostgreSQL port (default: 5432)
// - PG_USER: PostgreSQL user (default: postgres)
// - PG_PASSWORD: PostgreSQL password (default: empty)
// - PG_DATABASE: PostgreSQL database name (default: postgres)
// - PG_SSLMODE: SSL mode (default: prefer)
type PostgreSQLProvider struct {
	host     string
	port     string
	user     string
	password string
	database string
	sslMode  string
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewPostgreSQLProvider creates a new PostgreSQL provider.
// It reads configuration from environment variables.
func NewPostgreSQLProvider() *PostgreSQLProvider {
	return &PostgreSQLProvider{
		host:     getEnvOrDefault("PG_HOST", "localhost"),
		port:     getEnvOrDefault("PG_PORT", "5432"),
		user:     getEnvOrDefault("PG_USER", "postgres"),
		password: os.Getenv("PG_PASSWORD"),
		database: getEnvOrDefault("PG_DATABASE", "postgres"),
		sslMode:  getEnvOrDefault("PG_SSLMODE", "prefer"),
	}
}

// Name returns the provider name
func (p *PostgreSQLProvider) Name() string {
	return "postgresql"
}

// Provision validates the connection to an existing PostgreSQL database.
// Unlike cloud providers, this does not create new infrastructure - it expects
// PostgreSQL to already be running and accessible.
func (p *PostgreSQLProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
	// Get connection string
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Test connection
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set connection pool settings for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping with timeout to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w. Please ensure PostgreSQL is running and credentials are correct", err)
	}

	// Get PostgreSQL version for metadata
	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		// Non-fatal - continue without version info
		version = "unknown"
	}

	// Parse connection string to extract components
	parsedURL, err := url.Parse(connStr)
	if err != nil {
		// If parsing fails, use configured values
		return &DatabaseInfo{
			ID:               "postgresql-local",
			Name:             opts.Name,
			ConnectionString: connStr,
			Host:             p.host,
			Port:             parsePort(p.port),
			Database:         p.database,
			Username:         p.user,
			Password:         p.password,
			Region:           "local",
			CreatedAt:        time.Now(),
			Metadata: map[string]interface{}{
				"provider": "postgresql",
				"type":     "self-hosted",
				"version":  version,
			},
		}, nil
	}

	// Extract host and port from URL
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "5432"
	}

	// Extract database name from path
	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		dbName = p.database
	}

	// Extract username
	username := ""
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
	}

	return &DatabaseInfo{
		ID:               "postgresql-local",
		Name:             opts.Name,
		ConnectionString: connStr,
		Host:             host,
		Port:             parsePort(port),
		Database:         dbName,
		Username:         username,
		Password:         p.password,
		Region:           "local",
		CreatedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"provider": "postgresql",
			"type":     "self-hosted",
			"version":  version,
			"sslmode":  p.sslMode,
		},
	}, nil
}

// GetConnectionString returns the PostgreSQL connection string.
// It first checks for DATABASE_URL, then constructs one from individual PG_* variables.
func (p *PostgreSQLProvider) GetConnectionString(ctx context.Context) (string, error) {
	// Check if DATABASE_URL is set (highest priority)
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		return connStr, nil
	}

	// Construct connection string from individual components
	// Format: postgresql://user:password@host:port/database?sslmode=prefer
	var connStr strings.Builder
	connStr.WriteString("postgresql://")

	// Add user
	if p.user != "" {
		connStr.WriteString(url.QueryEscape(p.user))

		// Add password if present
		if p.password != "" {
			connStr.WriteString(":")
			connStr.WriteString(url.QueryEscape(p.password))
		}

		connStr.WriteString("@")
	}

	// Add host and port
	connStr.WriteString(p.host)
	if p.port != "" {
		connStr.WriteString(":")
		connStr.WriteString(p.port)
	}

	// Add database
	connStr.WriteString("/")
	connStr.WriteString(p.database)

	// Add SSL mode
	if p.sslMode != "" {
		connStr.WriteString("?sslmode=")
		connStr.WriteString(p.sslMode)
	}

	return connStr.String(), nil
}

// RunMigrations executes database migrations using goose
func (p *PostgreSQLProvider) RunMigrations(ctx context.Context, migrationsDir string) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set goose dialect to postgres
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Health checks if the database is reachable and healthy
func (p *PostgreSQLProvider) Health(ctx context.Context) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Run a simple query to verify database is operational
	var result int
	queryCtx, queryCancel := context.WithTimeout(ctx, 5*time.Second)
	defer queryCancel()

	if err := db.QueryRowContext(queryCtx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// Destroy is a no-op for PostgreSQL provider.
// Since this provider doesn't provision infrastructure, it doesn't destroy it either.
// The PostgreSQL server and database must be managed manually.
func (p *PostgreSQLProvider) Destroy(ctx context.Context) error {
	// PostgreSQL provider doesn't provision infrastructure, so nothing to destroy
	// Users must manage their PostgreSQL instances manually
	return nil
}

// parsePort converts a port string to an integer, returning 0 if invalid
func parsePort(portStr string) int {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}
