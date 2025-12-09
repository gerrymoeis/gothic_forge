package db

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkloadType defines the type of database workload for optimization
type WorkloadType int

const (
	// WorkloadBalanced is for applications with mixed read/write operations
	WorkloadBalanced WorkloadType = iota
	// WorkloadReadHeavy is for applications with primarily read operations
	WorkloadReadHeavy
	// WorkloadWriteHeavy is for applications with primarily write operations
	WorkloadWriteHeavy
)

// String returns the string representation of WorkloadType
func (w WorkloadType) String() string {
	switch w {
	case WorkloadBalanced:
		return "balanced"
	case WorkloadReadHeavy:
		return "read-heavy"
	case WorkloadWriteHeavy:
		return "write-heavy"
	default:
		return "unknown"
	}
}

// ParseWorkloadType converts a string to WorkloadType
func ParseWorkloadType(s string) WorkloadType {
	switch s {
	case "read-heavy", "read_heavy":
		return WorkloadReadHeavy
	case "write-heavy", "write_heavy":
		return WorkloadWriteHeavy
	default:
		return WorkloadBalanced
	}
}

// OptimizePool applies production-ready pool settings based on workload type
// This function configures the connection pool for optimal performance in production
func OptimizePool(cfg *pgxpool.Config, workload WorkloadType) {
	switch workload {
	case WorkloadReadHeavy:
		// Read-heavy workloads benefit from more connections
		cfg.MaxConns = 30
		cfg.MinConns = 10
	case WorkloadWriteHeavy:
		// Write-heavy workloads need fewer connections to avoid contention
		cfg.MaxConns = 20
		cfg.MinConns = 5
	case WorkloadBalanced:
		// Balanced workloads use moderate connection counts
		cfg.MaxConns = 25
		cfg.MinConns = 5
	}

	// Aggressive connection recycling to prevent stale connections
	// Reduced from 1h to 30min based on benchmark findings
	cfg.MaxConnLifetime = 30 * time.Minute

	// Reduced idle timeout to free up connections faster
	// Reduced from 30min to 5min based on benchmark findings
	cfg.MaxConnIdleTime = 5 * time.Minute

	// More frequent health checks to detect issues faster
	// Reduced from 1min to 30s based on benchmark findings
	cfg.HealthCheckPeriod = 30 * time.Second

	// Add connection and acquire timeouts for better error handling
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	// Note: AcquireTimeout is not directly configurable in pgxpool.Config
	// It's handled at the pool.Acquire() call level
}

// ValidatePoolConfig validates pool configuration parameters
func ValidatePoolConfig(cfg *pgxpool.Config) error {
	if cfg == nil {
		return errors.New("pool config cannot be nil")
	}

	if cfg.MaxConns < 1 {
		return errors.New("MaxConns must be at least 1")
	}

	if cfg.MinConns < 0 {
		return errors.New("MinConns cannot be negative")
	}

	if cfg.MaxConns < cfg.MinConns {
		return errors.New("MaxConns must be greater than or equal to MinConns")
	}

	if cfg.MaxConnLifetime < 0 {
		return errors.New("MaxConnLifetime cannot be negative")
	}

	if cfg.MaxConnIdleTime < 0 {
		return errors.New("MaxConnIdleTime cannot be negative")
	}

	if cfg.HealthCheckPeriod < 0 {
		return errors.New("HealthCheckPeriod cannot be negative")
	}

	return nil
}

// PoolMetrics contains statistics about the connection pool
type PoolMetrics struct {
	AcquireCount      int64         // Total number of successful acquires
	AcquireDuration   time.Duration // Cumulative duration of all acquires
	AcquiredConns     int32         // Number of currently acquired connections
	CanceledAcquires  int64         // Number of acquires canceled by context
	ConstructingConns int32         // Number of connections being constructed
	EmptyAcquires     int64         // Number of acquires that waited for a connection
	IdleConns         int32         // Number of idle connections
	MaxConns          int32         // Maximum number of connections
	TotalConns        int32         // Total number of connections in the pool
}

// GetPoolMetrics returns current pool statistics
// Returns nil if pool is not initialized
func GetPoolMetrics() *PoolMetrics {
	if pool == nil {
		return nil
	}

	stat := pool.Stat()
	return &PoolMetrics{
		AcquireCount:      stat.AcquireCount(),
		AcquireDuration:   stat.AcquireDuration(),
		AcquiredConns:     stat.AcquiredConns(),
		CanceledAcquires:  stat.CanceledAcquireCount(),
		ConstructingConns: stat.ConstructingConns(),
		EmptyAcquires:     stat.EmptyAcquireCount(),
		IdleConns:         stat.IdleConns(),
		MaxConns:          stat.MaxConns(),
		TotalConns:        stat.TotalConns(),
	}
}
