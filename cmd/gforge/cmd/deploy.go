package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gothicforge/internal/execx"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the app to Fly.io using infra/fly/fly.toml (omakase)",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()

		// 1) Require flyctl
		if _, err := exec.LookPath("flyctl"); err != nil {
			color.Red("flyctl not found. Please install it first:")
			fmt.Println("  Windows:  winget install -e --id Fly.io.Flyctl")
			fmt.Println("  macOS:    brew install flyctl")
			fmt.Println("  Linux:    curl -L https://fly.io/install.sh | sh")
			return errors.New("flyctl is required for deploy")
		}

		// 2) Ensure templ code is generated once before building
		if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
			_ = execx.Run(cmd.Context(), "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false")
		} else {
			color.Yellow("templ not available and auto-install failed: %v", err)
		}

		// 3) Read app name from fly.toml
		flyToml := filepath.Join("infra", "fly", "fly.toml")
		appName := readFlyAppName(flyToml)
		if appName == "" {
			appName = "gothic-forge"
			color.Yellow("No app name found in %s, defaulting to %s", flyToml, appName)
		}

		// 4) Ensure app exists; if not, create
		if err := execx.Run(cmd.Context(), "flyctl status", "flyctl", "status", "-a", appName); err != nil {
			if err2 := execx.Run(cmd.Context(), "flyctl apps create", "flyctl", "apps", "create", appName); err2 != nil {
				return fmt.Errorf("failed to create app %s: %w", appName, err2)
			}
		}
		deployArgs := []string{"deploy", "--remote-only", "-c", flyToml, "-a", appName}
		if err := execx.Run(cmd.Context(), "flyctl deploy", "flyctl", deployArgs...); err != nil {
			return err
		}

		// 6) Open the deployed app in the browser
		_ = execx.Run(cmd.Context(), "flyctl open", "flyctl", "open", "-a", appName)
		color.HiGreen("Deployment complete.")
		return nil
	},
}

// readFlyAppName parses the app name from a Fly.io TOML file.
func readFlyAppName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "app = ") {
			v := strings.TrimPrefix(line, "app = ")
			v = strings.TrimSpace(v)
			v = strings.Trim(v, "\"")
			return v
		}
	}
	if err := s.Err(); err != nil {
		color.Yellow("Failed to read app name from %s: %v", path, err)
	}
	return ""
}

func init() {
	// Register deploy command under root
	rootCmd.AddCommand(deployCmd)
}
