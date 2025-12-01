package db

import (
  "context"
  "errors"
  "log"
  "os"
  "strconv"
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
  
  // Configure connection pool
  cfg.MaxConns = getEnvInt("DB_MAX_CONNS", 20)
  cfg.MinConns = getEnvInt("DB_MIN_CONNS", 2)
  cfg.MaxConnLifetime = time.Hour
  cfg.MaxConnIdleTime = 30 * time.Minute
  cfg.HealthCheckPeriod = time.Minute
  
  cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
  defer cancel()
  p, err := pgxpool.NewWithConfig(cctx, cfg)
  if err != nil { return err }
  if err := p.Ping(cctx); err != nil { p.Close(); return err }
  pool = p
  
  log.Printf("Database pool configured: max=%d, min=%d, maxLifetime=%v, maxIdleTime=%v, healthCheck=%v",
    cfg.MaxConns, cfg.MinConns, cfg.MaxConnLifetime, cfg.MaxConnIdleTime, cfg.HealthCheckPeriod)
  
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

// getEnvInt retrieves an integer environment variable or returns the default value.
func getEnvInt(key string, defaultVal int32) int32 {
  if v := os.Getenv(key); v != "" {
    if n, err := strconv.ParseInt(v, 10, 32); err == nil {
      return int32(n)
    }
    log.Printf("WARNING: invalid integer value for %s, using default %d", key, defaultVal)
  }
  return defaultVal
}
