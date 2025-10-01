package cmd

import (
  "fmt"
  "os"
  "strings"

  "github.com/fatih/color"
  "github.com/spf13/cobra"
  gforgeenv "gothicforge/internal/env"
  "gothicforge/internal/execx"
)

var (
  provApp     string
  provService string
)

// deployProvisionCmd performs minimal provisioning for the omakase stack.
// Phase 1: validate Render CLI and login; validate presence of provider tokens; print next steps.
// Future phases: call Neon/Upstash/Cloudflare APIs to create resources and set envs automatically.
var deployProvisionCmd = &cobra.Command{
  Use:   "provision",
  Short: "Provision the omakase stack (Render; validate Neon/Upstash/Pages tokens)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    _ = gforgeenv.Load()

    app := provApp
    if app == "" {
      app = deployApp
      if app == "" {
        app = "gothic-forge"
      }
    }
    service := provService
    if service == "" {
      service = deployService
      if service == "" {
        service = "web"
      }
    }

    // Ensure Render CLI
    if _, ok := execx.Look("render"); !ok {
      // Try auto-install via go install path; if that fails, instruct manual
      if _, err := ensureTool("render", "github.com/render-oss/cli/cmd/render@latest"); err != nil {
        color.Red("Render CLI not found and auto-install failed.")
        color.Yellow("Install manually from releases: https://render.com/docs/cli")
        return fmt.Errorf("render CLI required")
      }
    }

    // Verify authentication to avoid confusing CLI errors
    if err := execx.Run(cmd.Context(), "render services", "render", "services", "-o", "json", "--confirm"); err != nil {
      color.Red("Render CLI is not authenticated.")
      color.Yellow("Run `render login` (opens browser), then re-run: gforge deploy provision --app %s --service %s", app, service)
      if isInteractive() {
        fmt.Print("Open Render New Web Service page now? [Y/n]: ")
        var ans string
        fmt.Scanln(&ans)
        ans = strings.TrimSpace(strings.ToLower(ans))
        if ans == "" || ans == "y" || ans == "yes" {
          _ = openURL("https://dashboard.render.com/web/new?newUser=true", cmd)
        }
      }
      return fmt.Errorf("render auth required: please run `render login`")
    }

    // Validate environment tokens for providers
    warnMissing := func(k string) {
      if os.Getenv(k) == "" {
        color.Yellow("⚠ %s is not set", k)
      } else {
        color.Green("✔ %s is set", k)
      }
    }

    color.Cyan("Provider tokens & IDs (optional, used for full automation):")
    warnMissing("NEON_API_KEY")
    warnMissing("UPSTASH_API_KEY")
    warnMissing("CLOUDFLARE_API_TOKEN")
    warnMissing("CLOUDFLARE_ACCOUNT_ID")
    // For CI automation with Render CLI
    warnMissing("RENDER_API_KEY")
    // For one-command deploys
    warnMissing("RENDER_SERVICE_ID")

    color.Cyan("Runtime envs (picked up by deploy if present):")
    for _, k := range []string{"BASE_URL", "DATABASE_URL", "REDIS_URL", "CORS_ORIGINS"} {
      if v := os.Getenv(k); v != "" {
        color.Green("✔ %s=%s", k, v)
      } else {
        color.Yellow("⚠ %s not set", k)
      }
    }

    color.HiGreen("Provisioning complete. Next: run `gforge deploy --provider render --app %s --service %s`.", app, service)
    color.HiBlack("Create service if needed: https://dashboard.render.com/web/new?newUser=true")
    return nil
  },
}

func init() {
  deployProvisionCmd.Flags().StringVar(&provApp, "app", "", "Application name (defaults to --app)")
  deployProvisionCmd.Flags().StringVar(&provService, "service", "", "Service name (defaults to --service)")
  deployCmd.AddCommand(deployProvisionCmd)
}
