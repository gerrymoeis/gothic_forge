package db

import (
  "context"
  "errors"
  "os"
  "strings"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Connect initializes a global pgx pool using DATABASE_URL if not already connected.
func Connect(ctx context.Context) error {
  if pool != nil { return nil }
  dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
  if dsn == "" {
    return errors.New("DATABASE_URL is empty")
  }
  cfg, err := pgxpool.ParseConfig(dsn)
  if err != nil { return err }
  // Reasonable defaults; keep conservative for local/dev
  // cfg.MaxConns = 5
  cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  p, err := pgxpool.NewWithConfig(cctx, cfg)
  if err != nil { return err }
  if err := p.Ping(cctx); err != nil { p.Close(); return err }
  pool = p
  return nil
}

// Pool returns the current global pool (may be nil).
func Pool() *pgxpool.Pool { return pool }

// Close closes the global pool.
func Close() { if pool != nil { pool.Close(); pool = nil } }

// Health pings the database using a short timeout. Returns nil if healthy.
func Health(ctx context.Context) error {
  if pool == nil { return errors.New("db not connected") }
  cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  return pool.Ping(cctx)
}
