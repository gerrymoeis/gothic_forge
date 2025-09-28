package cmd

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gothicforge/internal/execx"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Enable built-in auth scaffolding (routes, templates, users repo)",
	Long: `Enables the built-in auth scaffolding that is gated by the 'authscaffold' build tag.

This removes build tags from the auth files and regenerates Templ components.

Includes:
- app/routes/auth_routes.go
- app/db/users.go
- app/templates/auth_login.templ
- app/templates/auth_register.templ
- app/templates/auth_logout.templ
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := []string{
			filepath.Join("app", "routes", "auth_routes.go"),
			filepath.Join("app", "db", "users.go"),
			filepath.Join("app", "templates", "auth_login.templ"),
			filepath.Join("app", "templates", "auth_register.templ"),
			filepath.Join("app", "templates", "auth_logout.templ"),
		}
		var changed int
		for _, p := range paths {
			ok, err := removeAuthBuildTag(p)
			if err != nil {
				color.Yellow("skip %s: %v", p, err)
				continue
			}
			if ok {
				color.Green("✔ enabled %s", p)
				changed++
			}
		}
		if changed == 0 {
			color.Yellow("No files updated. They may already be enabled.")
		}
		// Regenerate templ components
		if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
			_ = execx.Run(cmd.Context(), "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false")
		} else {
			color.Yellow("templ not available and auto-install failed: %v", err)
		}
		// best-effort gofmt on modified files
		for _, p := range paths {
			_ = execx.Run(cmd.Context(), "gofmt", "gofmt", "-w", p)
		}
		color.HiGreen("Auth scaffolding enabled. You can now visit /auth/register and /auth/login.")
		return nil
	},
}

func init() {
	addCmd.AddCommand(addAuthCmd)
}

// removeAuthBuildTag strips the leading build tag lines for 'authscaffold' from a file.
// Returns (true, nil) if file updated.
func removeAuthBuildTag(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) == 0 {
		return false, errors.New("empty file")
	}
	start := 0
	// Strip first two lines if they match build constraints
	if start < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[start]), "//go:build authscaffold") {
		start++
	}
	if start < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[start]), "// +build authscaffold") {
		start++
	}
	if start == 0 {
		return false, nil // nothing to do
	}
	// Also remove a following blank line if present for cleanliness
	if start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	out := strings.Join(lines[start:], "\n")
	// Ensure trailing newline
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	// Write back atomically
	f, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return false, err
	}
	defer os.Remove(f.Name())
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(out); err != nil {
		f.Close()
		return false, err
	}
	_ = w.Flush()
	if err := f.Close(); err != nil {
		return false, err
	}
	return true, os.Rename(f.Name(), path)
}
