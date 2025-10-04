package execx

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

// Run executes a command and streams stdout/stderr to the current process.
func Run(ctx context.Context, name string, args ...string) error {
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// RunInteractive executes a command wiring stdin/stdout/stderr to the current process,
// allowing interactive prompts.
func RunInteractive(ctx context.Context, name string, args ...string) error {
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    return cmd.Run()
}

// RunQuiet executes a command and suppresses stdout/stderr. Useful for probes.
func RunQuiet(ctx context.Context, name string, args ...string) error {
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    cmd.Stdout = io.Discard
    cmd.Stderr = io.Discard
    return cmd.Run()
}

// RunCapture executes a command and returns combined stdout+stderr as a string.
func RunCapture(ctx context.Context, name string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    var buf bytes.Buffer
    cmd.Stdout = &buf
    cmd.Stderr = &buf
    err := cmd.Run()
    return buf.String(), err
}

// Look finds an executable in PATH.
func Look(name string) (string, bool) {
    if p, err := exec.LookPath(name); err == nil {
        return p, true
    }
    return "", false
}

// TimeoutContext returns a context with timeout; 0 means no timeout.
func TimeoutContext(ms int) (context.Context, context.CancelFunc) {
    if ms <= 0 {
        return context.WithCancel(context.Background())
    }
    return context.WithTimeout(context.Background(), time.Duration(ms)*time.Millisecond)
}

// WriteFileIfMissing writes content to a file if it doesn't exist.
func WriteFileIfMissing(path string, data []byte, perm os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

// Download fetches a URL to a local file.
func Download(ctx context.Context, url, dest string) error {
	req, err := httpNewRequestWithContext(ctx, url)
	if err != nil {
		return err
	}
	resp, err := httpDefaultClientDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmtErrorf("download failed: %s -> %d", url, resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// Lightweight indirection to avoid importing net/http directly at call sites
// and to make unit testing easier by swapping these vars.
var (
	httpNewRequestWithContext = func(ctx context.Context, url string) (*http.Request, error) { return http.NewRequestWithContext(ctx, http.MethodGet, url, nil) }
	httpDefaultClientDo       = func(req *http.Request) (*http.Response, error) { return http.DefaultClient.Do(req) }
	fmtErrorf                 = func(format string, a ...any) error { return fmt.Errorf(format, a...) }
)
