package server

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gothicforge3/internal/env"
)

// OptimizedCSPMiddleware creates a CSP middleware with pre-computed headers
// This avoids repeated string joins and allocations on every request
func OptimizedCSPMiddleware() func(http.Handler) http.Handler {
	// Pre-compute CSP strings for dev and prod environments
	isDev := env.Get("APP_ENV", "development") == "development"
	
	var cspHeader string
	if isDev {
		// Development CSP (more permissive)
		cspHeader = "default-src 'self'; " +
			"script-src 'self' https: 'unsafe-eval' 'unsafe-inline'; " +
			"style-src 'self' https: 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' https:; " +
			"connect-src 'self' https:; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'self'"
	} else {
		// Production CSP (stricter, but allows specific CDNs)
		cspHeader = "default-src 'self'; " +
			"script-src 'self' https://unpkg.com https://cdn.jsdelivr.net 'unsafe-inline'; " +
			"style-src 'self' https: 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' https:; " +
			"connect-src 'self' https:; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'self'"
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set pre-computed CSP header (no allocations)
			w.Header().Set("Content-Security-Policy", cspHeader)
			next.ServeHTTP(w, r)
		})
	}
}


// Conditional Middleware Configuration
//
// The following environment variables can be used to disable specific middleware:
//
// - DISABLE_COMPRESSION=1    : Disable response compression
// - DISABLE_CORS=1           : Disable CORS headers
// - DISABLE_CSP=1            : Disable Content-Security-Policy headers
// - DISABLE_CSRF=1           : Disable CSRF protection (production only)
// - DISABLE_HTML_CACHE=1     : Disable HTML caching headers
// - RATE_LIMIT_ENABLED=false : Disable rate limiting
//
// These flags are useful for:
// - Debugging performance issues
// - Testing specific middleware behavior
// - Deploying behind proxies that handle these concerns
// - Reducing overhead in specific deployment scenarios
//
// Example usage:
//   DISABLE_COMPRESSION=1 DISABLE_CORS=1 ./app


// MiddlewareContext holds reusable context data for middleware operations
// This reduces allocations by pooling context objects
type MiddlewareContext struct {
	StartTime time.Time
	RequestID string
	UserAgent string
	Path      string
	Method    string
}

// Reset clears the context for reuse
func (mc *MiddlewareContext) Reset() {
	mc.StartTime = time.Time{}
	mc.RequestID = ""
	mc.UserAgent = ""
	mc.Path = ""
	mc.Method = ""
}

// ctxPool provides reusable MiddlewareContext instances
var ctxPool = sync.Pool{
	New: func() interface{} {
		return &MiddlewareContext{}
	},
}

// GetMiddlewareContext retrieves a context from the pool
func GetMiddlewareContext() *MiddlewareContext {
	ctx := ctxPool.Get().(*MiddlewareContext)
	ctx.Reset()
	return ctx
}

// PutMiddlewareContext returns a context to the pool
func PutMiddlewareContext(ctx *MiddlewareContext) {
	if ctx != nil {
		ctxPool.Put(ctx)
	}
}

// WithMiddlewareContext executes a function with a pooled middleware context
func WithMiddlewareContext(fn func(*MiddlewareContext) error) error {
	ctx := GetMiddlewareContext()
	defer PutMiddlewareContext(ctx)
	return fn(ctx)
}


// OptimizedCompressionMiddleware creates a smart compression middleware
// This improves on chi's default by:
// - Skipping compression for small responses (< 1KB)
// - Skipping already-compressed content types
// - Verifying compression ratio meets minimum threshold
func OptimizedCompressionMiddleware(level int) func(http.Handler) http.Handler {
	// Pre-compute list of content types that should not be compressed
	skipContentTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/webm":      true,
		"audio/mpeg":      true,
		"audio/ogg":       true,
		"application/zip": true,
		"application/gzip": true,
		"application/x-gzip": true,
		"application/pdf": true,
	}
	
	const minSize = 1024 // 1KB minimum for compression
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			
			// Use buffer to capture response
			buf := GetBuffer()
			defer PutBuffer(buf)
			
			// Create a response writer that captures to buffer
			bw := &bufferingWriter{
				ResponseWriter: w,
				buffer:         buf,
				statusCode:     http.StatusOK,
			}
			
			// Serve the request
			next.ServeHTTP(bw, r)
			
			// Check content type - skip if already compressed
			contentType := bw.Header().Get("Content-Type")
			if contentType != "" {
				// Extract base content type (before semicolon)
				if idx := strings.Index(contentType, ";"); idx != -1 {
					contentType = strings.TrimSpace(contentType[:idx])
				}
				if skipContentTypes[contentType] {
					// Write uncompressed
					w.WriteHeader(bw.statusCode)
					buf.WriteTo(w)
					return
				}
			}
			
			// Check size - skip if too small
			if buf.Len() < minSize {
				w.WriteHeader(bw.statusCode)
				buf.WriteTo(w)
				return
			}
			
			// Compress and verify ratio
			compressed := GetBuffer()
			defer PutBuffer(compressed)
			
			gw := gzip.NewWriter(compressed)
			gw.Write(buf.Bytes())
			gw.Close()
			
			// Check compression ratio (must save at least 20%)
			originalSize := buf.Len()
			compressedSize := compressed.Len()
			ratio := float64(compressedSize) / float64(originalSize)
			
			if ratio > 0.80 {
				// Compression not effective, send uncompressed
				w.WriteHeader(bw.statusCode)
				buf.WriteTo(w)
				return
			}
			
			// Send compressed
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", strconv.Itoa(compressedSize))
			w.Header().Del("Content-Length") // Let Go set it
			w.WriteHeader(bw.statusCode)
			compressed.WriteTo(w)
		})
	}
}

// bufferingWriter captures response to a buffer
type bufferingWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (bw *bufferingWriter) WriteHeader(statusCode int) {
	bw.statusCode = statusCode
}

func (bw *bufferingWriter) Write(b []byte) (int, error) {
	return bw.buffer.Write(b)
}
