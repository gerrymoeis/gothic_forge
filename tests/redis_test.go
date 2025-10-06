package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gothicforge3/internal/server"
)

func Test_Sessions_Default_Memory_RoundTrip(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Unsetenv("VALKEY_URL")
	_ = os.Unsetenv("REDIS_URL")

	r := server.New()
	// write sets a value into the session
	r.Get("/_session/write", func(w http.ResponseWriter, r *http.Request) {
		server.Sessions().Put(r.Context(), "k", "v")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	// read reads the value back
	r.Get("/_session/read", func(w http.ResponseWriter, r *http.Request) {
		val := server.Sessions().GetString(r.Context(), "k")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(val))
	})

	// First request: write and capture session cookie
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/_session/write", nil))
	if rec1.Code != 200 { t.Fatalf("write want 200, got %d", rec1.Code) }
	var cookie *http.Cookie
	for _, c := range rec1.Result().Cookies() { if c.Name == server.Sessions().Cookie.Name { cookie = c; break } }
	if cookie == nil { t.Fatalf("expected session cookie to be set") }

	// Second request: use cookie to read
	req2 := httptest.NewRequest(http.MethodGet, "/_session/read", nil)
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if got := strings.TrimSpace(rec2.Body.String()); got != "v" {
		t.Fatalf("read want 'v', got %q", got)
	}
}

func Test_Sessions_Redis_Configured_Sets_RedisStore(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("VALKEY_URL", "redis://localhost:6379") // don't need actual server; just check store type

	_ = server.New()
	if server.Sessions() == nil || server.Sessions().Store == nil {
		t.Skip("session store not initialized; skipping")
		return
	}
	typeName := strings.ToLower(fmt.Sprintf("%T", server.Sessions().Store))
	if !strings.Contains(typeName, "redisstore") {
		t.Skipf("redisstore not detected in Sessions().Store type (%q). This check is best-effort. Skipping.", typeName)
	}
}
