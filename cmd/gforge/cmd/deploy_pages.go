package cmd

import (
  "errors"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"

  "github.com/fatih/color"
  "github.com/spf13/cobra"
  gforgeenv "gothicforge/internal/env"
  "gothicforge/internal/execx"
)

var (
  deployPagesDir     string
  deployPagesProject string
)

// deployPagesCmd uploads the static site (dist/) to Cloudflare Pages using Wrangler if available.
// We intentionally avoid Node in the core flow, but the documented Direct Upload
// method for Pages is via Wrangler or dashboard drag-and-drop.
var deployPagesCmd = &cobra.Command{
  Use:   "pages",
  Short: "Deploy static site in dist/ to Cloudflare Pages (via Wrangler)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    _ = gforgeenv.Load()

    // 1) Ensure output directory exists
    dir := deployPagesDir
    if dir == "" {
      dir = "dist"
    }
    if st, err := os.Stat(dir); err != nil || !st.IsDir() {
      color.Red("dist folder not found at %q.", dir)
      color.Yellow("Run `gforge export -o %s` first.", dir)
      return errors.New("dist directory missing")
    }

    // 2) Require Wrangler CLI
    if _, err := exec.LookPath("wrangler"); err != nil {
      color.Red("wrangler not found. Cloudflare Pages Direct Upload requires Wrangler or dashboard drag-and-drop.")
      fmt.Println("Install Wrangler:")
      fmt.Println("  Docs: https://developers.cloudflare.com/workers/wrangler/install-and-update/")
      fmt.Println("  Node/npm required. Alternative: use GitHub Action cloudflare/pages-action@v1")
      return errors.New("wrangler is required for pages deploy")
    }

    // 3) Require project name
    proj := deployPagesProject
    if proj == "" {
      proj = os.Getenv("CLOUDFLARE_PAGES_PROJECT")
      if proj == "" {
        return errors.New("missing --pages-project and CLOUDFLARE_PAGES_PROJECT is not set in .env")
      }
    }

    // 4) Execute wrangler pages deploy
    wrArgs := []string{"pages", "deploy", dir, "--project-name", proj}
    if err := execx.Run(cmd.Context(), "wrangler pages deploy", "wrangler", wrArgs...); err != nil {
      return err
    }
    color.HiGreen("Cloudflare Pages deployment submitted for project %s from %s", proj, filepath.Clean(dir))
    return nil
  },
}

func init() {
  deployPagesCmd.Flags().StringVar(&deployPagesDir, "dir", "dist", "directory to upload (default: dist)")
  deployPagesCmd.Flags().StringVar(&deployPagesProject, "pages-project", "", "Cloudflare Pages project name (required)")
  deployCmd.AddCommand(deployPagesCmd)
}
