package db

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	poolMu     sync.Mutex
	cachedPool *pgxpool.Pool
	cachedURL  string
)

// Connect returns a pgx pool when DATABASE_URL is set, otherwise (nil, nil).
// This keeps the starter lightweight and lets tests run without a database.
// NOTE: This creates a NEW pool each call. Prefer ConnectCached for app usage.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return nil, nil
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// ConnectCached returns a singleton pgx pool for the current DATABASE_URL.
// If DATABASE_URL changes at runtime, this will recreate the pool.
func ConnectCached(ctx context.Context) (*pgxpool.Pool, error) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return nil, nil
	}
	poolMu.Lock()
	defer poolMu.Unlock()
	if cachedPool != nil && cachedURL == url {
		return cachedPool, nil
	}
	// Close previous pool if any (best effort)
	if cachedPool != nil {
		cachedPool.Close()
		cachedPool = nil
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	cachedPool = p
	cachedURL = url
	return cachedPool, nil
}
