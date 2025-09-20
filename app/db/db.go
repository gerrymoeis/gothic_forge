package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gothicforge/internal/framework/env"
)

var pool *pgxpool.Pool

// Connect initializes a global pgx pool using DATABASE_URL.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	if pool != nil {
		return pool, nil
	}
	url := env.Get("DATABASE_URL", "")
	if url == "" {
		log.Printf("warn: DATABASE_URL is empty; db not connected")
		return nil, nil
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 60 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute

	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := p.Ping(ctx); err != nil {
		return nil, err
	}
	pool = p
	return pool, nil
}

// Pool returns the global pool (may be nil if not connected).
func Pool() *pgxpool.Pool { return pool }
