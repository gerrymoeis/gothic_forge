package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProvider implements CacheProvider for standard Redis instances.
//
// This provider supports connecting to existing Redis instances (self-hosted or cloud-managed).
// Unlike ValkeyProvider, it does not provision new instances automatically.
//
// The provider includes:
// - Connection string management
// - Health checks with PING command
// - TLS support for secure connections
// - Redis-compatible operations
//
// Supported connection string formats:
// - redis://[:password@]host:port[/db]
// - rediss://[:password@]host:port[/db] (TLS)
type RedisProvider struct {
	connectionStr string
}

// NewRedisProvider creates a new Redis provider
func NewRedisProvider() *RedisProvider {
	return &RedisProvider{}
}

// Name returns the provider name
func (p *RedisProvider) Name() string {
	return "redis"
}

// Provision validates an existing Redis connection
// Note: This provider does not create new Redis instances.
// It expects REDIS_URL to be set with a connection string to an existing instance.
func (p *RedisProvider) Provision(ctx context.Context, opts ProvisionOptions) (*CacheInfo, error) {
	// Check if REDIS_URL is set
	connStr := os.Getenv("REDIS_URL")
	if connStr == "" {
		return nil, fmt.Errorf("REDIS_URL not set. Please set your Redis connection string (e.g., redis://localhost:6379)")
	}

	p.connectionStr = connStr

	// Test the connection with proper TLS support
	client := p.createRedisClient(connStr)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	// Parse connection details
	addr := p.parseRedisAddr(connStr)
	password := p.parseRedisPassword(connStr)

	// Extract host and port
	host := addr
	port := 6379 // default Redis port
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host = addr[:idx]
		fmt.Sscanf(addr[idx+1:], "%d", &port)
	}

	return &CacheInfo{
		ID:               "redis-existing",
		Name:             opts.Name,
		ConnectionString: connStr,
		Host:             host,
		Port:             port,
		Password:         password,
		Region:           "self-hosted",
		CreatedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"provider": "redis",
			"type":     "existing",
		},
	}, nil
}

// GetConnectionString returns the Redis connection string
func (p *RedisProvider) GetConnectionString(ctx context.Context) (string, error) {
	// Check if we have a cached connection string
	if p.connectionStr != "" {
		return p.connectionStr, nil
	}

	// Check REDIS_URL environment variable
	connStr := os.Getenv("REDIS_URL")
	if connStr == "" {
		return "", fmt.Errorf("REDIS_URL not set. Please set your Redis connection string")
	}

	p.connectionStr = connStr
	return connStr, nil
}

// Health checks if the cache is reachable using PING command
func (p *RedisProvider) Health(ctx context.Context) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	// Create Redis client with proper TLS support
	client := p.createRedisClient(connStr)
	defer client.Close()

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Destroy is a no-op for Redis provider since it doesn't manage instance lifecycle
// Users must manually delete their Redis instances
func (p *RedisProvider) Destroy(ctx context.Context) error {
	// Redis provider doesn't manage instance lifecycle
	// Users must manually delete their Redis instances
	return fmt.Errorf("Redis provider does not support automatic instance deletion. Please manually delete your Redis instance")
}

// parseRedisAddr extracts the address from a Redis connection string
func (p *RedisProvider) parseRedisAddr(connStr string) string {
	// Handle redis:// or rediss:// URLs
	connStr = strings.TrimPrefix(connStr, "redis://")
	connStr = strings.TrimPrefix(connStr, "rediss://")

	// Remove credentials if present (use LastIndex to handle @ in passwords)
	if idx := strings.LastIndex(connStr, "@"); idx != -1 {
		connStr = connStr[idx+1:]
	}

	// Remove path if present
	if idx := strings.Index(connStr, "/"); idx != -1 {
		connStr = connStr[:idx]
	}

	return connStr
}

// parseRedisPassword extracts the password from a Redis connection string
func (p *RedisProvider) parseRedisPassword(connStr string) string {
	// Handle redis:// or rediss:// URLs
	connStr = strings.TrimPrefix(connStr, "redis://")
	connStr = strings.TrimPrefix(connStr, "rediss://")

	// Extract credentials if present (format: user:pass@host or :pass@host)
	// Find the last @ to handle passwords with @ in them
	if idx := strings.LastIndex(connStr, "@"); idx != -1 {
		credentials := connStr[:idx]
		// Check if there's a colon (password present)
		if colonIdx := strings.Index(credentials, ":"); colonIdx != -1 {
			return credentials[colonIdx+1:]
		}
	}

	return ""
}

// isTLSConnection checks if the connection string uses TLS (rediss://)
func (p *RedisProvider) isTLSConnection(connStr string) bool {
	return strings.HasPrefix(connStr, "rediss://")
}

// createRedisClient creates a Redis client with proper TLS configuration
func (p *RedisProvider) createRedisClient(connStr string) *redis.Client {
	opts := &redis.Options{
		Addr: p.parseRedisAddr(connStr),
	}

	// Extract password if present
	if password := p.parseRedisPassword(connStr); password != "" {
		opts.Password = password
	}

	// Enable TLS if using rediss://
	if p.isTLSConnection(connStr) {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// InsecureSkipVerify is set to false by default, which is secure
			// For production, certificates should be properly validated
		}
	}

	return redis.NewClient(opts)
}
