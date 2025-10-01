package cmd

import (
  "errors"
  "fmt"
  "os"
  "runtime"
  "strings"

  "github.com/fatih/color"
  "github.com/spf13/cobra"
  gforgeenv "gothicforge/internal/env"
  "gothicforge/internal/execx"
)

// Deploy flags
var (
  deployProvider string // render
  deployApp      string // app name (provider-specific). Default: gothic-forge
  deployService  string // service name (provider-specific). Default: web
  deployServiceID string // Render service ID (optional; otherwise interactive selection)
)

var deployCmd = &cobra.Command{
  Use:   "deploy",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()

    switch strings.ToLower(strings.TrimSpace(deployProvider)) {
    case "", "render":
      return deployRender(cmd)
    default:
      return fmt.Errorf("unsupported provider: %s (supported: render)", deployProvider)
    }
  },
}

func init() {
  // Register deploy command under root and flags
  deployCmd.Flags().StringVar(&deployProvider, "provider", "render", "deploy provider: render")
  deployCmd.Flags().StringVar(&deployApp, "app", "gothic-forge", "application name (provider-specific)")
  deployCmd.Flags().StringVar(&deployService, "service", "web", "service name (provider-specific)")
  deployCmd.Flags().StringVar(&deployServiceID, "service-id", "", "Render service ID (optional; if empty and interactive, opens selector)")
  rootCmd.AddCommand(deployCmd)
}

// isInteractive reports whether stdin is a TTY and not running in CI.
func isInteractive() bool {
  if os.Getenv("CI") != "" {
    return false
  }
  fi, err := os.Stdin.Stat()
  if err != nil {
    return false
  }
  return (fi.Mode() & os.ModeCharDevice) != 0
}

// openURL tries to open the given URL using the OS default handler.
func openURL(url string, cmd *cobra.Command) error {
  switch runtime.GOOS {
  case "windows":
    return execx.Run(cmd.Context(), "open URL", "cmd", "/c", "start", url)
  case "darwin":
    return execx.Run(cmd.Context(), "open URL", "open", url)
  default:
    return execx.Run(cmd.Context(), "open URL", "xdg-open", url)
  }
}

// ---- Render provider implementation ----

func deployRender(cmd *cobra.Command) error {
  // Load .env so CLI respects developer's .env values
  _ = gforgeenv.Load()

  // Ensure templ code is generated once before building
  if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
    _ = execx.Run(cmd.Context(), "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false")
  } else {
    color.Yellow("templ not available and auto-install failed: %v", err)
  }

  // Ensure Render CLI is available (either via go install or prior install)
  if _, ok := execx.Look("render"); !ok {
    color.Yellow("Render CLI not found. Install it via `gforge tools install deploy` or see https://render.com/docs/cli")
    return errors.New("render CLI required")
  }

  // Non-interactive guard: list services; if this fails, user likely needs `render login`
  if err := execx.Run(cmd.Context(), "render services", "render", "services", "-o", "json", "--confirm"); err != nil {
    color.Red("Render CLI is not authenticated.")
    color.Yellow("Run `render login` in a terminal (opens browser), then re-run: gforge deploy --provider render")
    return fmt.Errorf("render auth required: please run `render login`")
  }

  // Fast-path: if RENDER_SERVICE_ID is provided, trigger a deploy
  sid := strings.TrimSpace(deployServiceID)
  if sid == "" {
    sid = strings.TrimSpace(os.Getenv("RENDER_SERVICE_ID"))
  }
  if sid != "" {
    if err := execx.Run(cmd.Context(), "render deploys create", "render", "deploys", "create", sid, "--wait"); err != nil {
      return err
    }
    color.HiGreen("Deployment triggered for service %s. View status in Render Dashboard.", sid)
    return nil
  }

  // Interactive fallback: open the Render service selector and trigger a deploy
  if isInteractive() {
    if err := execx.Run(cmd.Context(), "render deploys create", "render", "deploys", "create", "--wait"); err != nil {
      color.Red("Interactive deploy failed. If no services exist, create one first:")
      color.HiBlack("%s", "https://dashboard.render.com/web/new?newUser=true")
      return err
    }
    color.HiGreen("Deployment triggered interactively. Check Render for progress and URL.")
    return nil
  }

  // Otherwise provide concise guidance to create the service once via Dashboard
  color.HiGreen("Render is ready. Create a New Web Service (Go native):")
  color.HiBlack("  1) %s", "https://dashboard.render.com/web/new?newUser=true")
  color.HiBlack("  2) Env: FIBER_HOST=0.0.0.0, FIBER_PORT=$PORT, APP_ENV=production")
  if v := os.Getenv("BASE_URL"); v != "" { color.HiBlack("     BASE_URL=%s", v) }
  if v := os.Getenv("DATABASE_URL"); v != "" { color.HiBlack("     DATABASE_URL=%s", v) }
  if v := os.Getenv("REDIS_URL"); v != "" { color.HiBlack("     REDIS_URL=%s", v) }
  color.HiBlack("  3) Build Command: templ generate && go build -o server ./cmd/server")
  color.HiBlack("     (templ generate flags already applied by gforge locally)")
  color.HiBlack("  4) Start Command: ./server")
  if isInteractive() {
    fmt.Print("Open Render New Web Service page now? [Y/n]: ")
    var ans string
    fmt.Scanln(&ans)
    ans = strings.TrimSpace(strings.ToLower(ans))
    if ans == "" || ans == "y" || ans == "yes" {
      _ = openURL("https://dashboard.render.com/web/new?newUser=true", cmd)
    }
  }
  return nil
}
