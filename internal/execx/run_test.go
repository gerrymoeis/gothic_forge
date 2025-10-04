package execx

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b", "c.txt")
	if err := WriteFileIfMissing(p, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil { t.Fatalf("read failed: %v", err) }
	if string(b) != "hello" { t.Fatalf("unexpected content: %q", string(b)) }
	// Second call should be no-op
	if err := WriteFileIfMissing(p, []byte("changed"), 0o644); err != nil {
		t.Fatalf("second write failed: %v", err)
	}
	b2, _ := os.ReadFile(p)
	if string(b2) != "hello" { t.Fatalf("should not overwrite: %q", string(b2)) }
}

func TestDownload_Stubbed(t *testing.T) {
	// Stub network
	oldReq := httpNewRequestWithContext
	oldDo := httpDefaultClientDo
	t.Cleanup(func() {
		httpNewRequestWithContext = oldReq
		httpDefaultClientDo = oldDo
	})
	httpNewRequestWithContext = func(ctx context.Context, url string) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, "http://example.test/file.txt", nil)
	}
	httpDefaultClientDo = func(req *http.Request) (*http.Response, error) {
		body := io.NopCloser(strings.NewReader("payload"))
		return &http.Response{StatusCode: 200, Body: body}, nil
	}

	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")
	if err := Download(context.Background(), "http://ignored.local/file", dest); err != nil {
		t.Fatalf("download failed: %v", err)
	}
	b, err := os.ReadFile(dest)
	if err != nil { t.Fatalf("read failed: %v", err) }
	if string(b) != "payload" { t.Fatalf("unexpected content: %q", string(b)) }
}
