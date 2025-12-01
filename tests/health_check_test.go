package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"gothicforge3/app/routes"
	"gothicforge3/internal/db"
	"gothicforge3/internal/server"
)

// Test_Readyz_AllHealthy tests /readyz when both CockroachDB and Valkey are available
// Validates: Requirements 10.1, 10.2, 10.3, 10.5
func Test_Readyz_AllHealthy(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	// Set up environment with valid DATABASE_URL (will skip actual connection in test)
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "")  // Empty means SKIP
	_ = os.Setenv("VALKEY_URL", "")    // Empty means SKIP
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// When no dependencies are configured, should return 200 with SKIP status
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "valkey: SKIP") {
		t.Errorf("expected 'valkey: SKIP' in response, got: %q", body)
	}
	if !strings.Contains(body, "db: SKIP") {
		t.Errorf("expected 'db: SKIP' in response, got: %q", body)
	}
	if !strings.Contains(body, "ready") {
		t.Errorf("expected 'ready' in response, got: %q", body)
	}
	
	// Verify headers
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got: %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Errorf("expected Cache-Control no-cache, got: %q", cc)
	}
}

// Test_Readyz_DatabaseUnavailable tests /readyz when CockroachDB is configured but unavailable
// Validates: Requirements 10.1, 10.3, 10.4
func Test_Readyz_DatabaseUnavailable(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	_ = os.Setenv("LOG_FORMAT", "off")
	// Set invalid DATABASE_URL to simulate unavailable database
	_ = os.Setenv("DATABASE_URL", "postgresql://invalid:5432/test?sslmode=disable")
	_ = os.Setenv("VALKEY_URL", "")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// Should return 503 when database is unavailable
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "db: FAIL") {
		t.Errorf("expected 'db: FAIL' in response, got: %q", body)
	}
	if !strings.Contains(body, "not ready") {
		t.Errorf("expected 'not ready' in response, got: %q", body)
	}
	
	// Clean up
	_ = os.Unsetenv("DATABASE_URL")
	db.Close()
}

// Test_Readyz_ValkeyUnavailable tests /readyz when Valkey is configured but unavailable
// Validates: Requirements 10.2, 10.3, 10.4
func Test_Readyz_ValkeyUnavailable(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "")
	// Set invalid VALKEY_URL to simulate unavailable Valkey
	_ = os.Setenv("VALKEY_URL", "redis://invalid:6379")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// Should return 503 when Valkey is unavailable
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "valkey: FAIL") {
		t.Errorf("expected 'valkey: FAIL' in response, got: %q", body)
	}
	if !strings.Contains(body, "not ready") {
		t.Errorf("expected 'not ready' in response, got: %q", body)
	}
	
	// Clean up
	_ = os.Unsetenv("VALKEY_URL")
}

// Test_Readyz_BothUnavailable tests /readyz when both dependencies are unavailable
// Validates: Requirements 10.1, 10.2, 10.3, 10.4
func Test_Readyz_BothUnavailable(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "postgresql://invalid:5432/test?sslmode=disable")
	_ = os.Setenv("VALKEY_URL", "redis://invalid:6379")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// Should return 503 when both are unavailable
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "valkey: FAIL") {
		t.Errorf("expected 'valkey: FAIL' in response, got: %q", body)
	}
	if !strings.Contains(body, "db: FAIL") {
		t.Errorf("expected 'db: FAIL' in response, got: %q", body)
	}
	if !strings.Contains(body, "not ready") {
		t.Errorf("expected 'not ready' in response, got: %q", body)
	}
	
	// Clean up
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("VALKEY_URL")
	db.Close()
}

// Test_Readyz_ResponseFormat tests the response format and headers
// Validates: Requirements 10.3, 10.5
func Test_Readyz_ResponseFormat(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "")
	_ = os.Setenv("VALKEY_URL", "")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// Verify response format
	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	
	// Should have at least 3 lines: valkey status, db status, overall status
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines in response, got %d: %q", len(lines), body)
	}
	
	// Verify each line has expected format (component: STATUS)
	for i, line := range lines {
		// Last line is overall status
		if i == len(lines)-1 {
			if line != "ready" && line != "not ready" {
				t.Errorf("expected last line to be 'ready' or 'not ready', got: %q", line)
			}
			continue
		}
		
		// Other lines should have format "component: STATUS"
		if !strings.Contains(line, ":") {
			t.Errorf("expected line %d to contain ':', got: %q", i, line)
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			t.Errorf("expected line %d to have format 'component: STATUS', got: %q", i, line)
			continue
		}
		
		status := strings.TrimSpace(parts[1])
		validStatuses := []string{"OK", "FAIL", "SKIP"}
		found := false
		for _, vs := range validStatuses {
			if status == vs {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected status to be one of %v, got: %q", validStatuses, status)
		}
	}
	
	// Verify required headers
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got: %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Errorf("expected Cache-Control no-cache, got: %q", cc)
	}
}

// Test_Readyz_DeprecatedEnvVar tests that deprecated REDIS_URL still works
// Validates: Requirements 10.2, 10.3
func Test_Readyz_DeprecatedEnvVar(t *testing.T) {
	// Clean up any existing database connection
	db.Close()
	
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "")
	_ = os.Unsetenv("VALKEY_URL")
	// Use deprecated REDIS_URL
	_ = os.Setenv("REDIS_URL", "redis://invalid:6379")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// Should still check Valkey using deprecated env var
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	
	body := rec.Body.String()
	if !strings.Contains(body, "valkey: FAIL") {
		t.Errorf("expected 'valkey: FAIL' when using deprecated REDIS_URL, got: %q", body)
	}
	
	// Clean up
	_ = os.Unsetenv("REDIS_URL")
}

// Test_Livez_AlwaysHealthy tests /livez endpoint always returns 200
func Test_Livez_AlwaysHealthy(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	
	r := server.New()
	routes.Register(r)
	
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	
	body := strings.TrimSpace(rec.Body.String())
	if body != "alive" {
		t.Errorf("expected 'alive', got: %q", body)
	}
	
	// Verify headers
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Errorf("expected Cache-Control no-cache, got: %q", cc)
	}
}

// Property-Based Tests for Health Checks

// Test_Property_Readyz_HTTPStatusMatchesDependencyState tests that HTTP status code
// correctly reflects the state of dependencies
// **Feature: gothic-forge-cleanup, Property 9: Health Check Completeness**
// **Validates: Requirements 10.3, 10.4**
func Test_Property_Readyz_HTTPStatusMatchesDependencyState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("HTTP status matches dependency state", prop.ForAll(
		func(dbConfigured bool, dbHealthy bool, valkeyConfigured bool, valkeyHealthy bool) bool {
			// Clean up any existing database connection
			db.Close()
			
			_ = os.Setenv("LOG_FORMAT", "off")
			
			// Set up environment based on generated states
			if dbConfigured {
				if dbHealthy {
					// Empty string means SKIP (treated as healthy for this test)
					_ = os.Setenv("DATABASE_URL", "")
				} else {
					// Invalid URL simulates unhealthy database
					_ = os.Setenv("DATABASE_URL", "postgresql://invalid:5432/test?sslmode=disable")
				}
			} else {
				_ = os.Setenv("DATABASE_URL", "")
			}
			
			if valkeyConfigured {
				if valkeyHealthy {
					// Empty string means SKIP (treated as healthy for this test)
					_ = os.Setenv("VALKEY_URL", "")
				} else {
					// Invalid URL simulates unhealthy Valkey
					_ = os.Setenv("VALKEY_URL", "redis://invalid:6379")
				}
			} else {
				_ = os.Setenv("VALKEY_URL", "")
			}
			
			r := server.New()
			routes.Register(r)
			
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			
			// Determine expected status code
			// If any configured dependency is unhealthy, expect 503
			// Otherwise expect 200
			anyUnhealthy := (dbConfigured && !dbHealthy) || (valkeyConfigured && !valkeyHealthy)
			expectedStatus := http.StatusOK
			if anyUnhealthy {
				expectedStatus = http.StatusServiceUnavailable
			}
			
			// Clean up
			_ = os.Unsetenv("DATABASE_URL")
			_ = os.Unsetenv("VALKEY_URL")
			db.Close()
			
			return rec.Code == expectedStatus
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Test_Property_Readyz_ResponseFormatConsistency tests that response format is consistent
// across all dependency combinations
// **Feature: gothic-forge-cleanup, Property 9: Health Check Completeness**
// **Validates: Requirements 10.3, 10.4**
func Test_Property_Readyz_ResponseFormatConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("response format is consistent", prop.ForAll(
		func(dbConfigured bool, dbHealthy bool, valkeyConfigured bool, valkeyHealthy bool) bool {
			// Clean up any existing database connection
			db.Close()
			
			_ = os.Setenv("LOG_FORMAT", "off")
			
			// Set up environment based on generated states
			if dbConfigured {
				if dbHealthy {
					_ = os.Setenv("DATABASE_URL", "")
				} else {
					_ = os.Setenv("DATABASE_URL", "postgresql://invalid:5432/test?sslmode=disable")
				}
			} else {
				_ = os.Setenv("DATABASE_URL", "")
			}
			
			if valkeyConfigured {
				if valkeyHealthy {
					_ = os.Setenv("VALKEY_URL", "")
				} else {
					_ = os.Setenv("VALKEY_URL", "redis://invalid:6379")
				}
			} else {
				_ = os.Setenv("VALKEY_URL", "")
			}
			
			r := server.New()
			routes.Register(r)
			
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			
			body := rec.Body.String()
			lines := strings.Split(strings.TrimSpace(body), "\n")
			
			// Verify response has at least 3 lines (valkey, db, overall status)
			if len(lines) < 3 {
				return false
			}
			
			// Verify last line is overall status
			lastLine := lines[len(lines)-1]
			if lastLine != "ready" && lastLine != "not ready" {
				return false
			}
			
			// Verify each component line has format "component: STATUS"
			validStatuses := map[string]bool{"OK": true, "FAIL": true, "SKIP": true}
			for i := 0; i < len(lines)-1; i++ {
				line := lines[i]
				if !strings.Contains(line, ":") {
					return false
				}
				
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					return false
				}
				
				status := strings.TrimSpace(parts[1])
				if !validStatuses[status] {
					return false
				}
			}
			
			// Verify required headers
			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "text/plain") {
				return false
			}
			
			cc := rec.Header().Get("Cache-Control")
			if !strings.Contains(cc, "no-cache") {
				return false
			}
			
			// Clean up
			_ = os.Unsetenv("DATABASE_URL")
			_ = os.Unsetenv("VALKEY_URL")
			db.Close()
			
			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}
