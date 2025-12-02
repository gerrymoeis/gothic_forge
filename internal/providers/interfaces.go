package providers

import (
	"context"
	"time"
)

// DatabaseProvider abstracts database connection and health checking.
// Gothic Forge uses PostgreSQL as the standard database.
type DatabaseProvider interface {
	// Name returns the provider name (e.g., "postgresql")
	Name() string

	// GetConnectionString returns the database connection string
	GetConnectionString(ctx context.Context) (string, error)

	// Health checks if the database is accessible
	Health(ctx context.Context) error
}

// CacheProvider abstracts cache connection and health checking.
// Gothic Forge uses Redis/Valkey for session storage and caching.
type CacheProvider interface {
	// Name returns the provider name (e.g., "redis", "valkey")
	Name() string

	// GetConnectionString returns the cache connection string
	GetConnectionString(ctx context.Context) (string, error)

	// Health checks if the cache is accessible
	Health(ctx context.Context) error
}

// DatabaseInfo contains information about a database connection
type DatabaseInfo struct {
	ID               string
	Name             string
	ConnectionString string
	Host             string
	Port             int
	Database         string
	Username         string
	Password         string
	Region           string
	CreatedAt        time.Time
	Metadata         map[string]interface{}
}

// CacheInfo contains information about a cache connection
type CacheInfo struct {
	ID               string
	Name             string
	ConnectionString string
	Host             string
	Port             int
	Password         string
	Region           string
	CreatedAt        time.Time
	Metadata         map[string]interface{}
}

// ProvisionOptions contains options for provisioning resources
// Note: This is kept for backwards compatibility with existing provider implementations
type ProvisionOptions struct {
	Name     string
	Region   string
	Tier     string
	Tags     map[string]string
	Metadata map[string]interface{}
}
