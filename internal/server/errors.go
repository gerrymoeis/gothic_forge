package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorContext holds reusable error context data
type ErrorContext struct {
	StatusCode int
	Message    string
	Details    string
	Timestamp  time.Time
	RequestID  string
	Path       string
}

// Reset clears the error context for reuse
func (ec *ErrorContext) Reset() {
	ec.StatusCode = 0
	ec.Message = ""
	ec.Details = ""
	ec.Timestamp = time.Time{}
	ec.RequestID = ""
	ec.Path = ""
}

// errorContextPool provides reusable ErrorContext instances
var errorContextPool = sync.Pool{
	New: func() interface{} {
		return &ErrorContext{}
	},
}

// GetErrorContext retrieves an error context from the pool
func GetErrorContext() *ErrorContext {
	ctx := errorContextPool.Get().(*ErrorContext)
	ctx.Reset()
	ctx.Timestamp = time.Now()
	return ctx
}

// PutErrorContext returns an error context to the pool
func PutErrorContext(ctx *ErrorContext) {
	if ctx != nil {
		errorContextPool.Put(ctx)
	}
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// WriteError writes an error response efficiently using buffer pools
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	ctx := GetErrorContext()
	defer PutErrorContext(ctx)
	
	ctx.StatusCode = statusCode
	ctx.Message = message
	
	// Use buffer pool for error formatting
	err := WithSmallBuffer(func(buf *bytes.Buffer) error {
		// Write JSON error response
		buf.WriteString(`{"error":"`)
		buf.WriteString(http.StatusText(statusCode))
		buf.WriteString(`","message":"`)
		buf.WriteString(escapeJSON(message))
		buf.WriteString(`","code":`)
		buf.WriteString(fmt.Sprintf("%d", statusCode))
		buf.WriteString(`}`)
		
		// Set headers
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(statusCode)
		
		// Write response
		_, err := buf.WriteTo(w)
		return err
	})
	
	if err != nil {
		// Fallback to simple error
		http.Error(w, message, statusCode)
	}
	
	// Log error
	logError(ctx)
}

// WriteHTMLError writes an HTML error page efficiently
func WriteHTMLError(w http.ResponseWriter, statusCode int, message string) {
	ctx := GetErrorContext()
	defer PutErrorContext(ctx)
	
	ctx.StatusCode = statusCode
	ctx.Message = message
	
	// Use buffer pool for HTML formatting
	err := WithBuffer(func(buf *bytes.Buffer) error {
		// Write minimal HTML error page
		buf.WriteString("<!DOCTYPE html><html><head><title>")
		buf.WriteString(fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)))
		buf.WriteString("</title></head><body><h1>")
		buf.WriteString(fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)))
		buf.WriteString("</h1><p>")
		buf.WriteString(escapeHTML(message))
		buf.WriteString("</p></body></html>")
		
		// Set headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(statusCode)
		
		// Write response
		_, err := buf.WriteTo(w)
		return err
	})
	
	if err != nil {
		// Fallback to simple error
		http.Error(w, message, statusCode)
	}
	
	// Log error
	logError(ctx)
}

// RecoverMiddleware is an optimized panic recovery middleware
// Uses error context pooling to minimize allocations
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := GetErrorContext()
				defer PutErrorContext(ctx)
				
				ctx.StatusCode = http.StatusInternalServerError
				ctx.Message = "Internal Server Error"
				ctx.Details = fmt.Sprintf("%v", err)
				ctx.Path = r.URL.Path
				
				// Get request ID if available
				if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
					ctx.RequestID = reqID
				}
				
				// Log panic with stack trace
				logPanic(ctx, debug.Stack())
				
				// Write error response
				WriteError(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// logError logs an error efficiently
func logError(ctx *ErrorContext) {
	// Use buffer pool for log formatting
	WithSmallBuffer(func(buf *bytes.Buffer) error {
		buf.WriteString(`{"level":"error","time":"`)
		buf.WriteString(ctx.Timestamp.Format(time.RFC3339))
		buf.WriteString(`","status":`)
		buf.WriteString(fmt.Sprintf("%d", ctx.StatusCode))
		buf.WriteString(`,"message":"`)
		buf.WriteString(escapeJSON(ctx.Message))
		buf.WriteString(`"`)
		
		if ctx.RequestID != "" {
			buf.WriteString(`,"request_id":"`)
			buf.WriteString(ctx.RequestID)
			buf.WriteString(`"`)
		}
		
		if ctx.Path != "" {
			buf.WriteString(`,"path":"`)
			buf.WriteString(ctx.Path)
			buf.WriteString(`"`)
		}
		
		buf.WriteString(`}`)
		log.Println(buf.String())
		return nil
	})
}

// logPanic logs a panic with stack trace
func logPanic(ctx *ErrorContext, stack []byte) {
	log.Printf("PANIC: %s\nPath: %s\nRequestID: %s\nStack:\n%s",
		ctx.Details, ctx.Path, ctx.RequestID, string(stack))
}

// escapeJSON escapes a string for JSON (simple version)
func escapeJSON(s string) string {
	// For production use, consider using json.Marshal or a more robust escaper
	// This is a minimal implementation for common cases
	buf := make([]byte, 0, len(s)+10)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			buf = append(buf, '\\', '\\')
		case '"':
			buf = append(buf, '\\', '"')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			buf = append(buf, c)
		}
	}
	return string(buf)
}

// escapeHTML escapes a string for HTML (simple version)
func escapeHTML(s string) string {
	buf := make([]byte, 0, len(s)+20)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '&':
			buf = append(buf, '&', 'a', 'm', 'p', ';')
		case '<':
			buf = append(buf, '&', 'l', 't', ';')
		case '>':
			buf = append(buf, '&', 'g', 't', ';')
		case '"':
			buf = append(buf, '&', 'q', 'u', 'o', 't', ';')
		case '\'':
			buf = append(buf, '&', '#', '3', '9', ';')
		default:
			buf = append(buf, c)
		}
	}
	return string(buf)
}

// NotFound writes a 404 error response
func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotFound, "Not Found")
}

// MethodNotAllowed writes a 405 error response
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
}

// BadRequest writes a 400 error response
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

// InternalServerError writes a 500 error response
func InternalServerError(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}

// Unauthorized writes a 401 error response
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, http.StatusUnauthorized, message)
}

// Forbidden writes a 403 error response
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, http.StatusForbidden, message)
}
