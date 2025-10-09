package cmd

import (
  "context"
  "crypto/tls"
  "database/sql"
  "fmt"
  "net"
  "net/url"
  "os"
  "path/filepath"
  "runtime"
  "strconv"
  "strings"
  "time"

  redigo "github.com/gomodule/redigo/redis"
  _ "github.com/jackc/pgx/v5/stdlib"
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

    // Provider CLIs
    // Railway
    railPath, railOK := execx.Look("railway")
    fmt.Printf("  • railway: %s\n", pathOrMissing(railPath, railOK))
    if railOK {
      if v, err := execx.RunCapture(context.Background(), "railway --version", "railway", "--version"); err == nil {
        v = strings.TrimSpace(v)
        if v != "" { fmt.Printf("    → %s\n", v) }
      }
    } else if doctorFix {
      if p, err := ensureRailwayCLI(); err == nil {
        railPath, railOK = p, true
        fmt.Printf("    → installed railway: %s\n", p)
        if v, err := execx.RunCapture(context.Background(), "railway --version", p, "--version"); err == nil {
          v = strings.TrimSpace(v)
          if v != "" { fmt.Printf("    → %s\n", v) }
        }
      } else {
        fmt.Println("    → failed to install railway:", err)
      }
    }

    // Wrangler
    wrPath, wrOK := execx.Look("wrangler")
    fmt.Printf("  • wrangler: %s\n", pathOrMissing(wrPath, wrOK))
    if wrOK {
      if v, err := execx.RunCapture(context.Background(), "wrangler --version", "wrangler", "--version"); err == nil {
        v = strings.TrimSpace(v)
        if v != "" { fmt.Printf("    → %s\n", v) }
      }
    } else if doctorFix {
      if p, err := ensureWranglerCLI(); err == nil {
        wrPath, wrOK = p, true
        fmt.Printf("    → installed wrangler: %s\n", p)
        // Use the binary we just installed if available
        if v, err := execx.RunCapture(context.Background(), "wrangler --version", p, "--version"); err == nil {
          v = strings.TrimSpace(v)
          if v != "" { fmt.Printf("    → %s\n", v) }
        }
      } else {
        fmt.Println("    → failed to install wrangler:", err)
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
      // Connectivity checks (optional)
      dsn := strings.TrimSpace(readEnvKey(envPath, "DATABASE_URL"))
      if dsn != "" {
        if err := probePostgres(dsn); err == nil {
          fmt.Println("  • Neon (Postgres): reachable")
        } else {
          fmt.Println("  • Neon (Postgres):", err)
        }
      }
      vurl := strings.TrimSpace(readEnvKey(envPath, "VALKEY_URL"))
      if vurl == "" { vurl = strings.TrimSpace(readEnvKey(envPath, "REDIS_URL")) }
      if vurl != "" {
        skip := strings.EqualFold(strings.TrimSpace(readEnvKey(envPath, "VALKEY_TLS_SKIP_VERIFY")), "1")
        if err := probeValkey(vurl, skip); err == nil {
          fmt.Println("  • Valkey/Redis: PING ok")
        } else {
          fmt.Println("  • Valkey/Redis:", err)
        }
      }

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

// probePostgres attempts a short ping to DATABASE_URL using pgx stdlib.
func probePostgres(dsn string) error {
  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()
  dbx, err := sql.Open("pgx", dsn)
  if err != nil { return err }
  defer dbx.Close()
  return dbx.PingContext(ctx)
}

// probeValkey attempts to PING the given Redis/Valkey URI with optional TLS skip-verify.
func probeValkey(uri string, skipVerify bool) error {
  u, err := url.Parse(uri)
  if err != nil { return err }
  scheme := strings.ToLower(u.Scheme)
  // Default: use DialURL for plain redis://
  if scheme == "redis" && !skipVerify {
    c, err := redigo.DialURL(uri)
    if err != nil { return err }
    defer c.Close()
    _, err = c.Do("PING")
    return err
  }
  // Compose options similar to server session pool
  opts := []redigo.DialOption{}
  if u.User != nil {
    if pw, ok := u.User.Password(); ok { opts = append(opts, redigo.DialPassword(pw)) }
  }
  if dbStr := strings.TrimPrefix(u.Path, "/"); dbStr != "" {
    if n, e := strconv.Atoi(dbStr); e == nil { opts = append(opts, redigo.DialDatabase(n)) }
  }
  if scheme == "rediss" || skipVerify {
    opts = append(opts, redigo.DialUseTLS(true))
    if skipVerify { opts = append(opts, redigo.DialTLSConfig(&tls.Config{InsecureSkipVerify: true})) }
  }
  host := u.Host
  c, err := redigo.Dial("tcp", host, opts...)
  if err != nil { return err }
  defer c.Close()
  _, err = c.Do("PING")
  return err
}
