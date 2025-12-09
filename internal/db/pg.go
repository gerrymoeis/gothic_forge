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
  
  // Determine workload type from environment (default: balanced)
  workloadStr := strings.ToLower(strings.TrimSpace(os.Getenv("DB_WORKLOAD")))
  workload := ParseWorkloadType(workloadStr)
  
  // Apply optimized pool settings based on workload
  OptimizePool(cfg, workload)
  
  // Allow environment variable overrides for fine-tuning
  if maxConns := getEnvInt("DB_MAX_CONNS", 0); maxConns > 0 {
    cfg.MaxConns = maxConns
  }
  if minConns := getEnvInt("DB_MIN_CONNS", 0); minConns > 0 {
    cfg.MinConns = minConns
  }
  
  // Validate configuration
  if err := ValidatePoolConfig(cfg); err != nil {
    return err
  }
  
  cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
  defer cancel()
  p, err := pgxpool.NewWithConfig(cctx, cfg)
  if err != nil { return err }
  if err := p.Ping(cctx); err != nil { p.Close(); return err }
  pool = p
  
  log.Printf("Database pool configured: workload=%s, max=%d, min=%d, maxLifetime=%v, maxIdleTime=%v, healthCheck=%v, connectTimeout=%v",
    workload.String(), cfg.MaxConns, cfg.MinConns, cfg.MaxConnLifetime, cfg.MaxConnIdleTime, cfg.HealthCheckPeriod, cfg.ConnConfig.ConnectTimeout)
  
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

// SlowQueryThreshold is the duration above which queries are logged as slow
// Default is 100ms, can be overridden via SLOW_QUERY_THRESHOLD_MS environment variable
var SlowQueryThreshold = 100 * time.Millisecond

func init() {
  // Allow configuring slow query threshold via environment
  if thresholdMs := strings.TrimSpace(os.Getenv("SLOW_QUERY_THRESHOLD_MS")); thresholdMs != "" {
    if ms, err := strconv.Atoi(thresholdMs); err == nil && ms > 0 {
      SlowQueryThreshold = time.Duration(ms) * time.Millisecond
    }
  }
}

// QueryWithTiming executes a query and logs if it exceeds the slow query threshold
// This is a helper function for tracking query performance
func QueryWithTiming(ctx context.Context, query string, args ...interface{}) (time.Duration, error) {
  if pool == nil {
    return 0, errors.New("database pool not initialized")
  }
  
  start := time.Now()
  _, err := pool.Exec(ctx, query, args...)
  duration := time.Since(start)
  
  if duration > SlowQueryThreshold {
    log.Printf("SLOW QUERY [%v]: %s (args: %v)", duration, query, args)
  }
  
  return duration, err
}

// QueryRowWithTiming executes a query that returns a single row and logs if slow
func QueryRowWithTiming(ctx context.Context, query string, args ...interface{}) (time.Duration, error) {
  if pool == nil {
    return 0, errors.New("database pool not initialized")
  }
  
  start := time.Now()
  row := pool.QueryRow(ctx, query, args...)
  duration := time.Since(start)
  
  if duration > SlowQueryThreshold {
    log.Printf("SLOW QUERY [%v]: %s (args: %v)", duration, query, args)
  }
  
  return duration, row.Scan()
}
