package cmd

import (
  "fmt"
  "net"
  "os"
  "path/filepath"
  "runtime"
  "strings"
  "time"

  "gothicforge3/internal/execx"
  "github.com/spf13/cobra"
)

var (
  doctorVerbose bool
  doctorFix bool
)

var doctorCmd = &cobra.Command{
  Use:   "doctor",
  Short: "Run preflight checks (env, tools, ports)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    fmt.Println("Doctor")

    // Go version
    fmt.Printf("  • Go: %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

    // Tools
    templPath, templOK := execx.Look("templ")
    gwPath, gwOK := execx.Look("gotailwindcss")
    fmt.Printf("  • templ: %s\n", pathOrMissing(templPath, templOK))
    fmt.Printf("  • gotailwindcss: %s\n", pathOrMissing(gwPath, gwOK))
    if doctorFix {
      if !templOK {
        if installedPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
          fmt.Printf("    → installed templ: %s\n", installedPath)
          templOK = true
        }
      }
      if !gwOK {
        if installedPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err == nil {
          fmt.Printf("    → installed gotailwindcss: %s\n", installedPath)
          gwOK = true
        }
      }
    }

    // .env and .env.example presence
    envPath := filepath.Join(".env")
    envExamplePath := filepath.Join(".env.example")
    var hasEnv bool
    if _, err := os.Stat(envPath); os.IsNotExist(err) {
      fmt.Printf("  • .env: not found (use 'gforge secrets' to populate)\n")
    } else {
      fmt.Printf("  • .env: present\n")
      hasEnv = true
    }
    if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
      fmt.Printf("  • .env.example: not found\n")
      if doctorFix {
        if err := writeEnvExample(envExamplePath); err == nil {
          fmt.Println("    → created .env.example")
        } else {
          fmt.Println("    → failed to create .env.example:", err)
        }
      }
    } else {
      fmt.Printf("  • .env.example: present\n")
    }

    // Warn if prod without real SITE_BASE_URL
    if hasEnv {
      appEnv := readEnvKey(envPath, "APP_ENV")
      siteBase := readEnvKey(envPath, "SITE_BASE_URL")
      if strings.ToLower(strings.TrimSpace(appEnv)) == "production" {
        sb := strings.TrimSpace(siteBase)
        if sb == "" || strings.HasPrefix(sb, "http://127.0.0.1") || strings.HasPrefix(sb, "http://localhost") {
          fmt.Println("  • Warning: APP_ENV=production but SITE_BASE_URL looks dev-like (set your public https URL)")
        }
      }
    }

    // Port probe (default dev port)
    devAddr := "127.0.0.1:8080"
    if portFree(devAddr) {
      fmt.Printf("  • Port %s: available\n", devAddr)
    } else {
      fmt.Printf("  • Port %s: in use\n", devAddr)
    }

    // Create sitemap registry template if missing and --fix
    sitemapDir := filepath.Join("app", "sitemap")
    sitemapFile := filepath.Join(sitemapDir, "urls.txt")
    if _, err := os.Stat(sitemapFile); os.IsNotExist(err) {
      if doctorFix {
        if err := os.MkdirAll(sitemapDir, 0o755); err == nil {
          content := "# Add one path or absolute URL per line.\n" +
            "# Lines starting with # are ignored. Examples:\n" +
            "# /about\n" +
            "# /pricing\n" +
            "# https://example.com/docs\n"
          _ = os.WriteFile(sitemapFile, []byte(content), 0o644)
          fmt.Println("  • Created app/sitemap/urls.txt")
        }
      }
    }

    // Readiness summary
    if hasEnv {
      railTok := strings.TrimSpace(readEnvKey(envPath, "RAILWAY_TOKEN"))
      apiTok := strings.TrimSpace(readEnvKey(envPath, "RAILWAY_API_TOKEN"))
      neonTok := strings.TrimSpace(readEnvKey(envPath, "NEON_TOKEN"))
      aivenTok := strings.TrimSpace(readEnvKey(envPath, "AIVEN_TOKEN"))
      cfTok := strings.TrimSpace(readEnvKey(envPath, "CF_API_TOKEN"))
      missing := []string{}
      // At least one Railway token is required to automate deploy
      if railTok == "" && apiTok == "" { missing = append(missing, "RAILWAY_TOKEN or RAILWAY_API_TOKEN") }
      if neonTok == "" { missing = append(missing, "NEON_TOKEN") }
      if aivenTok == "" { missing = append(missing, "AIVEN_TOKEN") }
      if cfTok == "" { missing = append(missing, "CF_API_TOKEN") }
      ready := "Yes"
      if len(missing) > 0 { ready = "No" }
      fmt.Println("────────────────────────────────────────")
      fmt.Println("Readiness summary")
      fmt.Printf("  • Ready for deploy: %s\n", ready)
      if len(missing) > 0 {
        fmt.Println("  • Missing:")
        for _, m := range missing { fmt.Printf("    - %s\n", m) }
      }
      if doctorVerbose {
        fmt.Println("  • Tips: set APP_ENV=production for stricter CSP, enable CSRF middleware")
      }
    }

    fmt.Println("────────────────────────────────────────")
    return nil
  },
}

func portFree(addr string) bool {
  ln, err := net.Listen("tcp", addr)
  if err != nil { return false }
  defer ln.Close()
  // small delay for OS settle
  time.Sleep(50 * time.Millisecond)
  return true
}

func pathOrMissing(p string, ok bool) string {
  if ok { return p }
  return "(missing)"
}

func writeEnvExample(p string) error {
  const content = `# Gothic Forge v3 - Environment Template
# Copy this file to .env and fill only what's necessary.
# Notes:
# - APP_ENV: development | production
# - In production, CSRF middleware is enabled and CSP is stricter.
# - All development happens in /app; gforge handles tooling and build.

# App
APP_ENV=development
SITE_BASE_URL=http://127.0.0.1:8080
JWT_SECRET=devsecret-change-me

# Server
HTTP_HOST=127.0.0.1
HTTP_PORT=8080
# LOG_FORMAT: empty for human logs, or 'json' for structured logs
LOG_FORMAT=

# Rate limiting (per IP)
RATE_LIMIT_MAX=120
RATE_LIMIT_WINDOW_SECONDS=60

# CORS
# Comma-separated list (e.g., https://example.com,https://app.example.com)
# Use '*' only for local development. When '*' is used, credentials are disabled.
CORS_ORIGINS=

# Service URLs (populated by deploy or your provider)
DATABASE_URL=
VALKEY_URL=

# OAuth (optional)
# If set, GitHub OAuth login will be enabled at /auth/github/login
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
# Base URL used to compute OAuth callback, defaults to SITE_BASE_URL
OAUTH_BASE_URL=

# Deploy provider tokens (used by 'gforge deploy')
RAILWAY_TOKEN=
NEON_TOKEN=
AIVEN_TOKEN=
CF_API_TOKEN=

# Provider signup and token/help links
# Railway: https://railway.app
# Neon (API keys): https://neon.tech/docs/manage/api-keys
# Aiven Console: https://console.aiven.io/
# Aiven API tokens: https://docs.aiven.io/docs/platform/howto/create_authentication_token
# Cloudflare Pages: https://pages.cloudflare.com/
# Cloudflare API tokens: https://dash.cloudflare.com/profile/api-tokens

# Security notes
# - CSRF: automatically enforced when APP_ENV=production
# - CSP: allows unpkg.com & cdn.jsdelivr.net and inline JSON-LD for SEO
# - Sessions: cookie SameSite=Lax, Secure in production
`
  return os.WriteFile(p, []byte(content), 0o644)
}

func readEnvKey(path, key string) string {
  b, err := os.ReadFile(path)
  if err != nil { return "" }
  lines := strings.Split(string(b), "\n")
  for _, ln := range lines {
    t := strings.TrimSpace(ln)
    if t == "" || strings.HasPrefix(t, "#") { continue }
    if idx := strings.Index(ln, "="); idx > 0 {
      k := strings.TrimSpace(ln[:idx])
      if strings.EqualFold(k, key) {
        return strings.TrimSpace(ln[idx+1:])
      }
    }
  }
  return ""
}

func init() {
  doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, "verbose output")
  doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "attempt to fix common issues")
  rootCmd.AddCommand(doctorCmd)
}
