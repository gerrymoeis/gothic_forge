package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RequestTiming holds timing information for a request
type RequestTiming struct {
	StartTime time.Time
	DBTime    time.Duration
	CacheTime time.Duration
	TotalTime time.Duration
}

// timingContextKey is the key for storing timing in context
type timingContextKey struct{}

// GetRequestTiming retrieves timing from context
func GetRequestTiming(ctx context.Context) *RequestTiming {
	if timing, ok := ctx.Value(timingContextKey{}).(*RequestTiming); ok {
		return timing
	}
	return nil
}

// SetRequestTiming stores timing in context
func SetRequestTiming(ctx context.Context, timing *RequestTiming) context.Context {
	return context.WithValue(ctx, timingContextKey{}, timing)
}

// TrackDBTime adds database operation time to the request timing
func TrackDBTime(ctx context.Context, duration time.Duration) {
	if timing := GetRequestTiming(ctx); timing != nil {
		timing.DBTime += duration
	}
}

// TrackCacheTime adds cache operation time to the request timing
func TrackCacheTime(ctx context.Context, duration time.Duration) {
	if timing := GetRequestTiming(ctx); timing != nil {
		timing.CacheTime += duration
	}
}

// FormatDuration formats a duration in the most appropriate unit
// This provides human-readable timing information in response headers
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// TelemetryMiddleware tracks request timing and adds timing headers
// This provides visibility into request performance without external tools
func TelemetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create timing context
		timing := &RequestTiming{
			StartTime: time.Now(),
		}
		ctx := SetRequestTiming(r.Context(), timing)
		r = r.WithContext(ctx)

		// Serve the request
		next.ServeHTTP(w, r)

		// Calculate total time
		timing.TotalTime = time.Since(timing.StartTime)

		// Add timing headers
		w.Header().Set("X-Response-Time", FormatDuration(timing.TotalTime))
		
		if timing.DBTime > 0 {
			w.Header().Set("X-DB-Time", FormatDuration(timing.DBTime))
		}
		
		if timing.CacheTime > 0 {
			w.Header().Set("X-Cache-Time", FormatDuration(timing.CacheTime))
		}
	})
}

// WithDBTiming wraps a database operation with timing tracking
func WithDBTiming(ctx context.Context, fn func() error) error {
	start := time.Now()
	err := fn()
	TrackDBTime(ctx, time.Since(start))
	return err
}

// WithCacheTiming wraps a cache operation with timing tracking
func WithCacheTiming(ctx context.Context, fn func() error) error {
	start := time.Now()
	err := fn()
	TrackCacheTime(ctx, time.Since(start))
	return err
}

// TimingStats provides aggregate timing statistics
type TimingStats struct {
	TotalRequests   int64
	AvgResponseTime time.Duration
	AvgDBTime       time.Duration
	AvgCacheTime    time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
}

// PerformanceMetrics tracks aggregate performance metrics
type PerformanceMetrics struct {
	RequestCount    int64
	TotalTime       time.Duration
	TotalDBTime     time.Duration
	TotalCacheTime  time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
}

// UpdateMetrics updates performance metrics with a new request
func (pm *PerformanceMetrics) UpdateMetrics(timing *RequestTiming) {
	pm.RequestCount++
	pm.TotalTime += timing.TotalTime
	pm.TotalDBTime += timing.DBTime
	pm.TotalCacheTime += timing.CacheTime

	if timing.TotalTime > pm.MaxResponseTime {
		pm.MaxResponseTime = timing.TotalTime
	}

	if pm.MinResponseTime == 0 || timing.TotalTime < pm.MinResponseTime {
		pm.MinResponseTime = timing.TotalTime
	}
}

// GetStats returns current timing statistics
func (pm *PerformanceMetrics) GetStats() TimingStats {
	if pm.RequestCount == 0 {
		return TimingStats{}
	}

	return TimingStats{
		TotalRequests:   pm.RequestCount,
		AvgResponseTime: time.Duration(int64(pm.TotalTime) / pm.RequestCount),
		AvgDBTime:       time.Duration(int64(pm.TotalDBTime) / pm.RequestCount),
		AvgCacheTime:    time.Duration(int64(pm.TotalCacheTime) / pm.RequestCount),
		MaxResponseTime: pm.MaxResponseTime,
		MinResponseTime: pm.MinResponseTime,
	}
}

// MetricsHandler returns an HTTP handler for the /metrics endpoint
func MetricsHandler(metrics *PerformanceMetrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := metrics.GetStats()
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		
		fmt.Fprintf(w, "# Performance Metrics\n")
		fmt.Fprintf(w, "total_requests %d\n", stats.TotalRequests)
		fmt.Fprintf(w, "avg_response_time_ms %.2f\n", float64(stats.AvgResponseTime.Microseconds())/1000.0)
		fmt.Fprintf(w, "avg_db_time_ms %.2f\n", float64(stats.AvgDBTime.Microseconds())/1000.0)
		fmt.Fprintf(w, "avg_cache_time_ms %.2f\n", float64(stats.AvgCacheTime.Microseconds())/1000.0)
		fmt.Fprintf(w, "max_response_time_ms %.2f\n", float64(stats.MaxResponseTime.Microseconds())/1000.0)
		fmt.Fprintf(w, "min_response_time_ms %.2f\n", float64(stats.MinResponseTime.Microseconds())/1000.0)
	}
}

// TimingHeadersMiddleware is a lightweight version that only adds timing headers
// without tracking aggregate metrics (lower overhead)
func TimingHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create timing context for DB/Cache tracking
		timing := &RequestTiming{StartTime: start}
		ctx := SetRequestTiming(r.Context(), timing)
		r = r.WithContext(ctx)
		
		// Serve request
		next.ServeHTTP(w, r)
		
		// Add timing header
		elapsed := time.Since(start)
		w.Header().Set("X-Response-Time", FormatDuration(elapsed))
		
		// Add DB/Cache timing if tracked
		if timing.DBTime > 0 {
			w.Header().Set("X-DB-Time", FormatDuration(timing.DBTime))
		}
		if timing.CacheTime > 0 {
			w.Header().Set("X-Cache-Time", FormatDuration(timing.CacheTime))
		}
	})
}

// ServerTimingHeader adds Server-Timing header for browser DevTools
// This provides detailed timing breakdown visible in browser developer tools
func ServerTimingHeader(w http.ResponseWriter, timing *RequestTiming) {
	if timing == nil {
		return
	}

	// Build Server-Timing header value
	var timings []string
	
	if timing.DBTime > 0 {
		timings = append(timings, fmt.Sprintf("db;dur=%.2f", float64(timing.DBTime.Microseconds())/1000.0))
	}
	
	if timing.CacheTime > 0 {
		timings = append(timings, fmt.Sprintf("cache;dur=%.2f", float64(timing.CacheTime.Microseconds())/1000.0))
	}
	
	timings = append(timings, fmt.Sprintf("total;dur=%.2f", float64(timing.TotalTime.Microseconds())/1000.0))
	
	if len(timings) > 0 {
		w.Header().Set("Server-Timing", joinStrings(timings, ", "))
	}
}

// joinStrings efficiently joins strings without allocations
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	
	// Calculate total length
	n := len(sep) * (len(parts) - 1)
	for _, s := range parts {
		n += len(s)
	}
	
	// Build string
	b := make([]byte, 0, n)
	b = append(b, parts[0]...)
	for _, s := range parts[1:] {
		b = append(b, sep...)
		b = append(b, s...)
	}
	return string(b)
}

// ParseDuration parses a duration string from headers
func ParseDuration(s string) (time.Duration, error) {
	// Try parsing as float milliseconds first
	if ms, err := strconv.ParseFloat(s, 64); err == nil {
		return time.Duration(ms * float64(time.Millisecond)), nil
	}
	// Fall back to standard duration parsing
	return time.ParseDuration(s)
}
