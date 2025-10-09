package cmd

import (
  "bufio"
  "fmt"
  "os"
  "strings"

  "github.com/spf13/cobra"
)

var (
  oauthGHClientID     string
  oauthGHClientSecret string
  oauthBaseURL        string
  oauthPrintOnly      bool
)

var oauthCmd = &cobra.Command{
  Use:   "oauth",
  Short: "OAuth helpers (GitHub)",
}

var oauthGithubCmd = &cobra.Command{
  Use:   "github",
  Short: "Configure GitHub OAuth env and show callback URL",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()

    // Load existing env map from .env (if present)
    envPath := ".env"
    kv := loadEnvFile(envPath)

    // Determine callback base
    base := strings.TrimSpace(oauthBaseURL)
    if base == "" {
      base = strings.TrimSpace(kv["OAUTH_BASE_URL"])
      if base == "" {
        base = strings.TrimSpace(kv["SITE_BASE_URL"])
      }
      if base == "" {
        base = "http://127.0.0.1:8080" // dev fallback
      }
    }
    base = normalizeBaseURL(base)
    callback := base + "/auth/github/callback"

    // Print-only mode: just display recommended callback and exit
    if oauthPrintOnly {
      fmt.Println("GitHub OAuth callback URL:")
      fmt.Println("  ", callback)
      fmt.Println()
      fmt.Println("Register at https://github.com/settings/developers and set:")
      fmt.Println("  Homepage URL:", base)
      fmt.Println("  Authorization callback URL:", callback)
      return nil
    }

    // Prompt for missing client ID/secret unless provided via flags
    reader := bufio.NewReader(os.Stdin)

    if strings.TrimSpace(oauthGHClientID) == "" {
      cur := strings.TrimSpace(kv["GITHUB_CLIENT_ID"])
      if cur != "" {
        oauthGHClientID = cur
      } else {
        fmt.Print("  • Enter GITHUB_CLIENT_ID: ")
        v, _ := reader.ReadString('\n')
        oauthGHClientID = strings.TrimSpace(v)
      }
    }

    if strings.TrimSpace(oauthGHClientSecret) == "" {
      cur := strings.TrimSpace(kv["GITHUB_CLIENT_SECRET"])
      if cur != "" {
        oauthGHClientSecret = cur
      } else {
        fmt.Print("  • Enter GITHUB_CLIENT_SECRET: ")
        v, _ := reader.ReadString('\n')
        oauthGHClientSecret = strings.TrimSpace(v)
      }
    }

    // Optionally persist OAUTH_BASE_URL when provided explicitly via flag
    if strings.TrimSpace(oauthBaseURL) != "" {
      kv["OAUTH_BASE_URL"] = base
    }

    // Persist values to .env
    up := map[string]string{
      "GITHUB_CLIENT_ID":     strings.TrimSpace(oauthGHClientID),
      "GITHUB_CLIENT_SECRET": strings.TrimSpace(oauthGHClientSecret),
    }
    // Merge explicit base if set
    if strings.TrimSpace(kv["OAUTH_BASE_URL"]) == "" && strings.TrimSpace(oauthBaseURL) != "" {
      up["OAUTH_BASE_URL"] = base
    }

    if _, err := os.Stat(envPath); os.IsNotExist(err) {
      if err := os.WriteFile(envPath, []byte(""), 0o600); err != nil { return err }
    }
    if err := updateEnvFileInPlace(envPath, up); err != nil { return err }

    fmt.Println("  • Saved GitHub OAuth env to .env")
    fmt.Println("  • Homepage URL:", base)
    fmt.Println("  • Authorization callback URL:", callback)
    fmt.Println()
    fmt.Println("Next:")
    fmt.Println("  1) Create a GitHub OAuth App: https://github.com/settings/developers")
    fmt.Println("  2) Use the URLs above. Then re-run your server and visit /auth/github/login")
    return nil
  },
}

func init() {
  oauthGithubCmd.Flags().StringVar(&oauthGHClientID, "client-id", "", "GitHub OAuth client ID")
  oauthGithubCmd.Flags().StringVar(&oauthGHClientSecret, "client-secret", "", "GitHub OAuth client secret")
  oauthGithubCmd.Flags().StringVar(&oauthBaseURL, "base", "", "Base URL to compute callback (overrides SITE_BASE_URL)")
  oauthGithubCmd.Flags().BoolVar(&oauthPrintOnly, "print", false, "Print the callback URL and exit")
  oauthCmd.AddCommand(oauthGithubCmd)
  rootCmd.AddCommand(oauthCmd)
}
