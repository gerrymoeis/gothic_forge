package server

import (
	"bytes"
	"strings"
	"sync"
)

// Buffer pools for reducing memory allocations in hot paths
// These pools reuse buffers across requests to minimize garbage collection

// bufferPool provides reusable bytes.Buffer instances
// Initial capacity of 4KB should handle most response sizes efficiently
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// stringBuilderPool provides reusable strings.Builder instances
// Used for efficient string concatenation without allocations
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		sb := &strings.Builder{}
		sb.Grow(1024) // Pre-allocate 1KB for common string operations
		return sb
	},
}

// smallBufferPool provides smaller buffers for lightweight operations
// 512 bytes should be sufficient for headers, small JSON, etc.
var smallBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

// GetBuffer retrieves a buffer from the pool
// The buffer is reset and ready for use
func GetBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
// The buffer should not be used after calling this function
func PutBuffer(buf *bytes.Buffer) {
	if buf != nil {
		// Don't pool extremely large buffers to avoid memory bloat
		if buf.Cap() <= 64*1024 { // 64KB limit
			bufferPool.Put(buf)
		}
	}
}

// GetSmallBuffer retrieves a small buffer from the pool
// Use this for lightweight operations like headers or small JSON
func GetSmallBuffer() *bytes.Buffer {
	buf := smallBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutSmallBuffer returns a small buffer to the pool
func PutSmallBuffer(buf *bytes.Buffer) {
	if buf != nil {
		// Don't pool buffers that grew too large
		if buf.Cap() <= 4*1024 { // 4KB limit for small buffers
			smallBufferPool.Put(buf)
		}
	}
}

// GetStringBuilder retrieves a string builder from the pool
// The builder is reset and ready for use
func GetStringBuilder() *strings.Builder {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
// The builder should not be used after calling this function
func PutStringBuilder(sb *strings.Builder) {
	if sb != nil {
		// Don't pool extremely large builders
		if sb.Cap() <= 32*1024 { // 32KB limit
			stringBuilderPool.Put(sb)
		}
	}
}

// PoolStats provides statistics about buffer pool usage
type PoolStats struct {
	BufferPoolSize      int // Current number of buffers in pool
	SmallBufferPoolSize int // Current number of small buffers in pool
	StringBuilderPoolSize int // Current number of string builders in pool
}

// GetPoolStats returns current pool statistics
// This is useful for monitoring memory pool efficiency
func GetPoolStats() PoolStats {
	// Note: sync.Pool doesn't expose size directly
	// This is a best-effort implementation for future monitoring enhancements
	return PoolStats{
		// These will always be 0 since sync.Pool doesn't track size
		// But we keep the structure for future monitoring enhancements
		BufferPoolSize:        0,
		SmallBufferPoolSize:   0,
		StringBuilderPoolSize: 0,
	}
}

// WithBuffer executes a function with a pooled buffer
// The buffer is automatically returned to the pool after use
func WithBuffer(fn func(*bytes.Buffer) error) error {
	buf := GetBuffer()
	defer PutBuffer(buf)
	return fn(buf)
}

// WithSmallBuffer executes a function with a pooled small buffer
// The buffer is automatically returned to the pool after use
func WithSmallBuffer(fn func(*bytes.Buffer) error) error {
	buf := GetSmallBuffer()
	defer PutSmallBuffer(buf)
	return fn(buf)
}

// WithStringBuilder executes a function with a pooled string builder
// The builder is automatically returned to the pool after use
func WithStringBuilder(fn func(*strings.Builder) error) error {
	sb := GetStringBuilder()
	defer PutStringBuilder(sb)
	return fn(sb)
}
