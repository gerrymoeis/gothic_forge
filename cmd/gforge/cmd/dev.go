package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"gothicforge/internal/execx"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run dev server with hot reload (templ + air + gotailwindcss)",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
			<-ch
			color.Yellow("\nshutting down...")
			cancel()
		}()

		// templ watch (auto-install if missing)
		go func() {
			templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest")
			if err != nil {
				color.Yellow("templ not available and auto-install failed: %v", err)
				return
			}

			_ = execx.Run(ctx, "templ", templPath, "generate", "-watch", "-include-version=false", "-include-timestamp=false")
		}()

		// Tailwind CSS via gotailwindcss (pure Go). Poll for changes to input CSS and rebuild.
		go func() {
			gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest")
			if err != nil {
				color.Yellow("gotailwindcss not available and auto-install failed: %v", err)
				return
			}
			input := "./app/static/tailwind.input.css"
			output := "./app/static/styles.css"
			var lastMod time.Time
			// Ensure initial build
			func() {
				_ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", output, input)
			}()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				fi, err := os.Stat(input)
				if err == nil {
					if fi.ModTime().After(lastMod) {
						_ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", output, input)
						lastMod = fi.ModTime()
					}
				}
				time.Sleep(1 * time.Second)
			}
		}()

		// server (prefer Air; auto-install if missing)
		go func() {
			if airPath, err := ensureTool("air", "github.com/air-verse/air@latest"); err == nil {
				_ = execx.Run(ctx, "air", airPath, "-c", ".air.toml")
				return
			}
			color.Yellow("air not available; running `go run ./cmd/server` (no auto-restart)")
			_ = execx.Run(ctx, "server", "go", "run", "./cmd/server")
		}()

		// Block until canceled
		<-ctx.Done()
		time.Sleep(300 * time.Millisecond)
		return nil
	},
}

// ensureTool ensures a CLI tool is available. If missing, it runs `go install <module>`
// and returns the absolute path to the installed binary (from GOBIN or GOPATH/bin).
func ensureTool(name, module string) (string, error) {
	if p, ok := execx.Look(name); ok {
		return p, nil
	}
	color.Yellow("%s not found. Installing %s...", name, module)
	c, cancel := execx.TimeoutContext(0)
	defer cancel()
	if err := execx.Run(c, "go install", "go", "install", module); err != nil {
		return "", fmt.Errorf("install failed for %s: %w", name, err)
	}
	if binDir := goBinDir(); binDir != "" {
		bin := filepath.Join(binDir, exeName(name))
		if _, err := os.Stat(bin); err == nil {
			return bin, nil
		}
	}
	if p, ok := execx.Look(name); ok {
		return p, nil
	}
	return "", fmt.Errorf("%s installed but not found in PATH; ensure GOBIN/GOPATH/bin is on PATH", name)
}

func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func goEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func goBinDir() string {
	if bin := goEnv("GOBIN"); bin != "" {
		return bin
	}
	gopath := goEnv("GOPATH")
	if gopath == "" {
		return ""
	}
	return filepath.Join(gopath, "bin")
}
