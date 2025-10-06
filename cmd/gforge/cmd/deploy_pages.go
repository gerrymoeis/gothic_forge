package cmd

import (
  "context"
  "fmt"
  "strings"

  "github.com/spf13/cobra"
  "gothicforge3/internal/execx"
)

var (
  pagesOutDir    string
  pagesProject   string
  pagesDeployRun bool
)

var deployPagesCmd = &cobra.Command{
  Use:   "pages",
  Short: "Deploy static export to Cloudflare Pages (wrangler)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    // 1) Export static site
    if pagesOutDir == "" { pagesOutDir = "dist" }
    exportOut = pagesOutDir
    if err := exportCmd.RunE(exportCmd, []string{}); err != nil {
      return err
    }

    // 2) Try wrangler CLI, else provide guidance
    if p, ok := execx.Look("wrangler"); ok {
      fmt.Println("wrangler found:", p)
      args := []string{"pages", "deploy", pagesOutDir}
      if strings.TrimSpace(pagesProject) != "" { args = append(args, "--project-name", pagesProject) }
      if pagesDeployRun {
        fmt.Println("Running:", "wrangler "+strings.Join(args, " "))
        ctx := context.Background()
        if strings.TrimSpace(pagesProject) != "" {
          if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir, "--project-name", pagesProject); err != nil { return err }
        } else {
          if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir); err != nil { return err }
        }
      } else {
        fmt.Println("Dry-run. To deploy with wrangler:")
        fmt.Println("  wrangler", strings.Join(args, " "))
      }
      return nil
    }

    // Guidance when wrangler not installed
    fmt.Println("wrangler CLI not found in PATH.")
    fmt.Println("Install:")
    fmt.Println("  - npm: npm i -g wrangler")
    fmt.Println("  - docs: https://developers.cloudflare.com/pages/framework-guides/deploy-a-static-site/")
    fmt.Println("Then run:")
    cmdLine := "wrangler pages deploy " + pagesOutDir
    if strings.TrimSpace(pagesProject) != "" { cmdLine += " --project-name " + pagesProject }
    fmt.Println("  "+cmdLine)
    return nil
  },
}

func init() {
  deployPagesCmd.Flags().StringVar(&pagesOutDir, "out", "dist", "export directory to deploy")
  deployPagesCmd.Flags().StringVar(&pagesProject, "project", "", "Cloudflare Pages project name")
  deployPagesCmd.Flags().BoolVar(&pagesDeployRun, "run", false, "execute wrangler if present (otherwise print instructions)")
  deployCmd.AddCommand(deployPagesCmd)
}
