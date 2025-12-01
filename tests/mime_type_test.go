package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gothicforge3/internal/server"
)

// Test_StaticFile_CSS verifies that .css files are served with correct Content-Type
func Test_StaticFile_CSS(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/styles/output.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("CSS file: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/css") {
		t.Errorf("CSS file: expected Content-Type to contain 'text/css', got %q", ct)
	}
}

// Test_StaticFile_JS verifies that .js files are served with correct Content-Type
func Test_StaticFile_JS(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("JS file: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/javascript") {
		t.Errorf("JS file: expected Content-Type to contain 'application/javascript', got %q", ct)
	}
}

// Test_StaticFile_SVG verifies that .svg files are served with correct Content-Type
func Test_StaticFile_SVG(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/favicon.svg", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("SVG file: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "image/svg+xml") {
		t.Errorf("SVG file: expected Content-Type to contain 'image/svg+xml', got %q", ct)
	}
}

// Test_StaticFile_WOFF2 verifies that .woff2 files are served with correct Content-Type
// Note: This test creates a temporary .woff2 file since one doesn't exist in the repo
func Test_StaticFile_WOFF2(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	
	// Create a temporary .woff2 file for testing
	tmpFile := "../app/static/test.woff2"
	if err := os.WriteFile(tmpFile, []byte("fake woff2 content"), 0644); err != nil {
		t.Fatalf("Failed to create test .woff2 file: %v", err)
	}
	defer os.Remove(tmpFile)
	
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/test.woff2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("WOFF2 file: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "font/woff2") {
		t.Errorf("WOFF2 file: expected Content-Type to contain 'font/woff2', got %q", ct)
	}
}

// Test_MimeTypeResponseWriter_WriteBeforeWriteHeader tests the edge case where
// Write() is called before WriteHeader()
func Test_MimeTypeResponseWriter_WriteBeforeWriteHeader(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	
	// Create a temporary CSS file
	tmpFile := "../app/static/test-edge.css"
	if err := os.WriteFile(tmpFile, []byte("body { color: red; }"), 0644); err != nil {
		t.Fatalf("Failed to create test CSS file: %v", err)
	}
	defer os.Remove(tmpFile)
	
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/test-edge.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("Edge case CSS: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/css") {
		t.Errorf("Edge case: Content-Type should be 'text/css', got %q", ct)
	}
}

// Test_MimeTypeResponseWriter_MultipleWrites tests that Content-Type is set correctly
// even with multiple Write() calls
func Test_MimeTypeResponseWriter_MultipleWrites(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	
	r := server.New()
	
	// Use an existing file that will have multiple writes
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("Multiple writes test: want 200, got %d", rec.Code)
	}
	
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/javascript") {
		t.Errorf("Multiple writes: Content-Type should contain 'application/javascript', got %q", ct)
	}
	
	// Verify we got content
	if rec.Body.Len() == 0 {
		t.Error("Multiple writes: expected non-empty body")
	}
}

// Test_StaticFile_CacheHeaders verifies that static files have proper cache headers
func Test_StaticFile_CacheHeaders(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("Cache headers test: want 200, got %d", rec.Code)
	}
	
	cacheControl := rec.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "public") {
		t.Errorf("Cache-Control should contain 'public', got %q", cacheControl)
	}
	if !strings.Contains(cacheControl, "max-age") {
		t.Errorf("Cache-Control should contain 'max-age', got %q", cacheControl)
	}
}

// Test_StaticFile_NotFound verifies 404 handling for non-existent files
func Test_StaticFile_NotFound(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	
	req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Non-existent file: want 404, got %d", rec.Code)
	}
}

// Test_MimeType_AllSupportedExtensions verifies all supported file extensions
func Test_MimeType_AllSupportedExtensions(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	
	tests := []struct {
		ext         string
		contentType string
		content     string
	}{
		{".css", "text/css", "body { }"},
		{".js", "application/javascript", "console.log('test');"},
		{".json", "application/json", "{}"},
		{".svg", "image/svg+xml", "<svg></svg>"},
		{".woff2", "font/woff2", "fake woff2"},
		{".woff", "font/woff", "fake woff"},
		{".ttf", "font/ttf", "fake ttf"},
		{".png", "image/png", "fake png"},
		{".jpg", "image/jpeg", "fake jpg"},
		{".jpeg", "image/jpeg", "fake jpeg"},
		{".gif", "image/gif", "fake gif"},
		{".webp", "image/webp", "fake webp"},
		{".ico", "image/x-icon", "fake ico"},
	}
	
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			// Create temporary file
			tmpFile := "../app/static/test" + tt.ext
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(tmpFile)
			
			r := server.New()
			
			req := httptest.NewRequest(http.MethodGet, "/static/test"+tt.ext, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			
			if rec.Code != http.StatusOK {
				t.Fatalf("File %s: want 200, got %d", tt.ext, rec.Code)
			}
			
			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, tt.contentType) {
				t.Errorf("File %s: expected Content-Type to contain %q, got %q", tt.ext, tt.contentType, ct)
			}
		})
	}
}
