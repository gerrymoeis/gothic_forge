package providers

import (
	"os"
)

// init registers all available providers with the default registry
func init() {
	// Database providers
	registerDatabaseProviders()

	// Cache providers
	registerCacheProviders()
}

func registerDatabaseProviders() {
	// PostgreSQL (standard database)
	// Uses DATABASE_URL environment variable
	DefaultRegistry.RegisterDatabase("postgresql", NewPostgreSQLProvider())
}

func registerCacheProviders() {
	// Redis (standard cache)
	// Uses REDIS_URL or VALKEY_URL environment variable
	DefaultRegistry.RegisterCache("redis", NewRedisProvider())

	// Valkey (Redis-compatible cache)
	// Uses VALKEY_URL environment variable and optional AIVEN_TOKEN for provisioning
	token := os.Getenv("AIVEN_TOKEN")
	DefaultRegistry.RegisterCache("valkey", NewValkeyProvider(token))
}
