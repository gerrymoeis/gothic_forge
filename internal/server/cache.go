package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"gothicforge3/internal/env"
)

// CachePoolConfig holds optimized configuration for Valkey/Redis connection pool
type CachePoolConfig struct {
	// MaxIdle is the maximum number of idle connections in the pool
	// Increased from 4 to 10 for better connection reuse under load
	MaxIdle int

	// MaxActive is the maximum number of active connections
	// Set to 50 to prevent connection exhaustion under high load
	MaxActive int

	// IdleTimeout is how long to keep idle connections before closing
	// Increased from 300s to 5min for better connection reuse
	IdleTimeout time.Duration

	// Wait determines if Get() should wait for a connection when MaxActive is reached
	// Set to true to prevent connection errors under load
	Wait bool

	// MaxConnLifetime is the maximum lifetime of a connection
	// Set to 30min to prevent stale connections
	MaxConnLifetime time.Duration

	// ConnectTimeout is the timeout for establishing new connections
	ConnectTimeout time.Duration

	// ReadTimeout is the timeout for read operations
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration
}

// DefaultCachePoolConfig returns production-optimized cache pool settings
func DefaultCachePoolConfig() CachePoolConfig {
	return CachePoolConfig{
		MaxIdle:         10,                // Increased from 4
		MaxActive:       50,                // New: prevent connection exhaustion
		IdleTimeout:     5 * time.Minute,   // Increased from 300s
		Wait:            true,              // New: wait for connections instead of failing
		MaxConnLifetime: 30 * time.Minute,  // New: prevent stale connections
		ConnectTimeout:  5 * time.Second,   // New: fail fast on connection issues
		ReadTimeout:     3 * time.Second,   // New: prevent hanging reads
		WriteTimeout:    3 * time.Second,   // New: prevent hanging writes
	}
}

// CreateOptimizedCachePool creates a Valkey/Redis connection pool with optimized settings
// This replaces the basic pool configuration in server.go with production-ready settings
func CreateOptimizedCachePool(cacheURL string, config CachePoolConfig) (*redigo.Pool, error) {
	if cacheURL == "" {
		return nil, fmt.Errorf("cache URL is empty")
	}

	// Support TLS (rediss://) and optional skip-verify via VALKEY_TLS_SKIP_VERIFY=1
	skipVerify := strings.EqualFold(strings.TrimSpace(env.Get("VALKEY_TLS_SKIP_VERIFY", "")), "1")
	
	u, err := url.Parse(cacheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cache URL: %w", err)
	}

	pool := &redigo.Pool{
		MaxIdle:         config.MaxIdle,
		MaxActive:       config.MaxActive,
		IdleTimeout:     config.IdleTimeout,
		Wait:            config.Wait,
		MaxConnLifetime: config.MaxConnLifetime,
		
		Dial: func() (redigo.Conn, error) {
			return dialWithRetry(u, skipVerify, config, 3)
		},
		
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			// Only ping if connection has been idle for more than 1 minute
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	// Warm up the pool by pre-creating some connections
	go warmupPool(pool, config.MaxIdle/2)

	return pool, nil
}

// dialWithRetry attempts to establish a Redis connection with exponential backoff
func dialWithRetry(u *url.URL, skipVerify bool, config CachePoolConfig, maxRetries int) (redigo.Conn, error) {
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		conn, err := dialRedis(u, skipVerify, config)
		if err == nil {
			return conn, nil
		}
		
		lastErr = err
		
		// Exponential backoff: 100ms, 200ms, 400ms
		if attempt < maxRetries-1 {
			backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
			time.Sleep(backoff)
		}
	}
	
	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, lastErr)
}

// dialRedis establishes a single Redis connection with proper TLS and auth settings
func dialRedis(u *url.URL, skipVerify bool, config CachePoolConfig) (redigo.Conn, error) {
	scheme := strings.ToLower(u.Scheme)
	opts := []redigo.DialOption{
		redigo.DialConnectTimeout(config.ConnectTimeout),
		redigo.DialReadTimeout(config.ReadTimeout),
		redigo.DialWriteTimeout(config.WriteTimeout),
	}

	// Password authentication
	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			opts = append(opts, redigo.DialPassword(pw))
		}
	}

	// Database index from path (e.g., /0)
	if dbStr := strings.TrimPrefix(u.Path, "/"); dbStr != "" {
		if n, err := strconv.Atoi(dbStr); err == nil {
			opts = append(opts, redigo.DialDatabase(n))
		}
	}

	// TLS settings for rediss:// or when skip-verify is enabled
	if scheme == "rediss" || skipVerify {
		opts = append(opts, redigo.DialUseTLS(true))
		if skipVerify {
			opts = append(opts, redigo.DialTLSConfig(&tls.Config{InsecureSkipVerify: true}))
		}
	}

	host := u.Host
	return redigo.Dial("tcp", host, opts...)
}

// warmupPool pre-creates connections to avoid cold start latency
// This runs asynchronously to not block server startup
func warmupPool(pool *redigo.Pool, targetConns int) {
	if targetConns <= 0 {
		return
	}

	log.Printf("Warming up cache pool with %d connections...", targetConns)
	
	conns := make([]redigo.Conn, 0, targetConns)
	
	// Create connections
	for i := 0; i < targetConns; i++ {
		conn := pool.Get()
		if conn.Err() != nil {
			log.Printf("Failed to warm up connection %d: %v", i+1, conn.Err())
			conn.Close()
			continue
		}
		conns = append(conns, conn)
	}
	
	// Return connections to pool
	for _, conn := range conns {
		conn.Close()
	}
	
	log.Printf("Cache pool warmed up with %d/%d connections", len(conns), targetConns)
}

// GetCachePoolStats returns current pool statistics for monitoring
type CachePoolStats struct {
	ActiveCount int // Number of active connections
	IdleCount   int // Number of idle connections
}

// GetPoolStats retrieves current pool statistics
func GetPoolStats(pool *redigo.Pool) CachePoolStats {
	stats := pool.Stats()
	return CachePoolStats{
		ActiveCount: stats.ActiveCount,
		IdleCount:   stats.IdleCount,
	}
}
