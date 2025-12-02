package providers

import (
	"context"
	"fmt"
	"io"
	"os"
)

// init registers all available providers with the default registry
func init() {
	// Database providers
	registerDatabaseProviders()

	// Cache providers
	registerCacheProviders()

	// Compute providers
	registerComputeProviders()

	// CDN providers
	registerCDNProviders()
}

func registerDatabaseProviders() {
	// SQLite (local development/testing)
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "gothic_forge.db"
	}
	DefaultRegistry.RegisterDatabase("sqlite", NewSQLiteProvider(dbPath))

	// CockroachDB (Opinionated Stack - production)
	// Registered when COCKROACH_API_KEY is available
	if apiKey := os.Getenv("COCKROACH_API_KEY"); apiKey != "" {
		DefaultRegistry.RegisterDatabase("cockroachdb", NewCockroachDBProvider(apiKey))
	}

	// PostgreSQL (alternative production/self-hosted)
	// Uses DATABASE_URL environment variable
	DefaultRegistry.RegisterDatabase("postgresql", NewPostgreSQLProvider())
}

func registerCacheProviders() {
	// Valkey (Opinionated Stack - production)
	// Registered when AIVEN_TOKEN is available
	if token := os.Getenv("AIVEN_TOKEN"); token != "" {
		DefaultRegistry.RegisterCache("valkey", NewValkeyProvider(token))
	}

	// Redis (alternative cache)
	// Uses REDIS_URL or VALKEY_URL environment variable
	DefaultRegistry.RegisterCache("redis", NewRedisProvider())
}

func registerComputeProviders() {
	// Docker (local development)
	DefaultRegistry.RegisterCompute("docker", NewDockerProvider())

	// Leapcell (Opinionated Stack - production)
	// Registered when LEAPCELL_API_KEY or LEAPCELL_APP_URL is available
	if apiKey := os.Getenv("LEAPCELL_API_KEY"); apiKey != "" {
		DefaultRegistry.RegisterCompute("leapcell", NewLeapcellProvider(apiKey))
	} else if os.Getenv("LEAPCELL_APP_URL") != "" {
		// Register with empty API key if only URL is available
		DefaultRegistry.RegisterCompute("leapcell", NewLeapcellProvider(""))
	}
}

func registerCDNProviders() {
	// Cloudflare (Opinionated Stack - production)
	// Registered when CLOUDFLARE_API_TOKEN is available
	if token := os.Getenv("CLOUDFLARE_API_TOKEN"); token != "" {
		DefaultRegistry.RegisterCDN("cloudflare", NewCloudflareProvider(token))
	}
}

// NewPostgreSQLProvider is now implemented in postgresql.go
// This function declaration is removed as the actual implementation is in postgresql.go

// NewValkeyProvider is now implemented in valkey.go
// This function is kept for backwards compatibility but is no longer needed
// as the actual implementation is in valkey.go

// NewRedisProvider is now implemented in redis.go
// This function declaration is removed as the actual implementation is in redis.go

// NewLeapcellProvider is now implemented in leapcell.go
// This function declaration is removed as the actual implementation is in leapcell.go

// NewCloudflareProvider is now implemented in cloudflare.go
// This function declaration is removed as the actual implementation is in cloudflare.go

// Placeholder implementations (to be replaced with real implementations)

type PlaceholderDatabaseProvider struct{ name string }

func (p *PlaceholderDatabaseProvider) Name() string                                                  { return p.name }
func (p *PlaceholderDatabaseProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
	return nil, ErrNotImplemented
}
func (p *PlaceholderDatabaseProvider) GetConnectionString(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}
func (p *PlaceholderDatabaseProvider) RunMigrations(ctx context.Context, migrationsDir string) error {
	return ErrNotImplemented
}
func (p *PlaceholderDatabaseProvider) Health(ctx context.Context) error { return ErrNotImplemented }
func (p *PlaceholderDatabaseProvider) Destroy(ctx context.Context) error { return ErrNotImplemented }

type PlaceholderCacheProvider struct{ name string }

func (p *PlaceholderCacheProvider) Name() string { return p.name }
func (p *PlaceholderCacheProvider) Provision(ctx context.Context, opts ProvisionOptions) (*CacheInfo, error) {
	return nil, ErrNotImplemented
}
func (p *PlaceholderCacheProvider) GetConnectionString(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}
func (p *PlaceholderCacheProvider) Health(ctx context.Context) error  { return ErrNotImplemented }
func (p *PlaceholderCacheProvider) Destroy(ctx context.Context) error { return ErrNotImplemented }

type PlaceholderComputeProvider struct{ name string }

func (p *PlaceholderComputeProvider) Name() string { return p.name }
func (p *PlaceholderComputeProvider) Deploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error) {
	return nil, ErrNotImplemented
}
func (p *PlaceholderComputeProvider) GetURL(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}
func (p *PlaceholderComputeProvider) GetLogs(ctx context.Context, opts LogOptions) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}
func (p *PlaceholderComputeProvider) Health(ctx context.Context) error { return ErrNotImplemented }
func (p *PlaceholderComputeProvider) Rollback(ctx context.Context, version string) error {
	return ErrNotImplemented
}
func (p *PlaceholderComputeProvider) Scale(ctx context.Context, replicas int) error {
	return ErrNotImplemented
}

type PlaceholderCDNProvider struct{ name string }

func (p *PlaceholderCDNProvider) Name() string { return p.name }
func (p *PlaceholderCDNProvider) Deploy(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error) {
	return nil, ErrNotImplemented
}
func (p *PlaceholderCDNProvider) Invalidate(ctx context.Context, paths []string) error {
	return ErrNotImplemented
}
func (p *PlaceholderCDNProvider) GetURL(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}

// ErrNotImplemented is returned when a provider method is not yet implemented
var ErrNotImplemented = fmt.Errorf("provider method not implemented")
