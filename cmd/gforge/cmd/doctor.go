package cmd

import (
  "context"
  "crypto/tls"
  "database/sql"
  "fmt"
  "net"
  "net/url"
  "os"
  "os/exec"
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

    // Core tools (Git is essential for deployment workflows)
    gitPath, gitOK := execx.Look("git")
    fmt.Printf("  • git: %s\n", pathOrMissing(gitPath, gitOK))
    if gitOK {
      if v, err := execx.RunCapture(context.Background(), "git --version", "git", "--version"); err == nil {
        v = strings.TrimSpace(v)
        if v != "" { fmt.Printf("    → %s\n", v) }
      }
    } else {
      fmt.Println("    ⚠️  Git is required for deployment workflows (Leapcell)")
      if doctorFix {
        fmt.Println("    → Git requires manual installation")
        printGitInstallHelp()
      }
    }

    // Development tools
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

    // Provider CLIs (Opinionated Stack)
    // Wrangler (for Cloudflare Pages)
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

    // Docker (required for Leapcell)
    dockerPath, dockerOK := execx.Look("docker")
    dockerDaemonRunning := false
    fmt.Printf("  • docker: %s\n", pathOrMissing(dockerPath, dockerOK))
    if dockerOK {
      if v, err := execx.RunCapture(context.Background(), "docker --version", "docker", "--version"); err == nil {
        v = strings.TrimSpace(v)
        if v != "" { fmt.Printf("    → %s\n", v) }
      }
      // Check if Docker daemon is running
      ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
      if err := execx.Run(ctx, "docker ps", "docker", "ps"); err == nil {
        fmt.Println("    → Docker daemon: running")
        dockerDaemonRunning = true
      } else {
        fmt.Println("    → Docker daemon: not running (start Docker Desktop)")
      }
      cancel()
    } else if doctorFix {
      fmt.Println("    → Docker requires manual installation")
      printDockerInstallHelp()
    }

    // Dockerfile validation (required for containerized deployments)
    dockerfilePath := "Dockerfile"
    dockerignorePath := ".dockerignore"
    hasDockerfile := false
    if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
      fmt.Printf("  • Dockerfile: missing\n")
      if doctorFix {
        fmt.Println("    → Creating Dockerfile for container deployments")
        if err := ensureDockerFiles(); err != nil {
          fmt.Println("    → failed to create Docker files:", err)
        } else {
          fmt.Println("    → Dockerfile and .dockerignore created")
          hasDockerfile = true
        }
      } else {
        fmt.Println("    ⚠️  Required for Leapcell deployment")
      }
    } else {
      fmt.Printf("  • Dockerfile: present\n")
      hasDockerfile = true
      // Validate Dockerfile syntax only if Docker daemon is running
      if dockerOK && dockerDaemonRunning {
        if err := validateDockerfile(dockerfilePath); err == nil {
          fmt.Println("    → Dockerfile syntax: valid")
        } else {
          fmt.Printf("    ⚠️  Dockerfile syntax check failed: %s\n", err)
        }
      } else if dockerOK && !dockerDaemonRunning {
        fmt.Println("    → Dockerfile syntax: not validated (Docker daemon not running)")
      }
    }
    if hasDockerfile {
      if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
        fmt.Printf("  • .dockerignore: missing (recommended for faster builds)\n")
        if doctorFix {
          // This should have been created by ensureDockerFiles, but handle separately
          fmt.Println("    → Creating .dockerignore")
          if err := ensureDockerFiles(); err != nil {
            fmt.Println("    → failed:", err)
          }
        }
      } else {
        fmt.Printf("  • .dockerignore: present\n")
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
          fmt.Println("  • CockroachDB (Postgres): reachable")
        } else {
          fmt.Println("  • CockroachDB (Postgres):", err)
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

      cockroachTok := strings.TrimSpace(readEnvKey(envPath, "COCKROACH_API_KEY"))
      aivenTok := strings.TrimSpace(readEnvKey(envPath, "AIVEN_TOKEN"))
      // Support both CLOUDFLARE_API_TOKEN (standard) and CF_API_TOKEN (legacy)
      cfTok := strings.TrimSpace(readEnvKey(envPath, "CLOUDFLARE_API_TOKEN"))
      if cfTok == "" {
        cfTok = strings.TrimSpace(readEnvKey(envPath, "CF_API_TOKEN"))
      }
      missing := []string{}
      // Database provider required (CockroachDB only)
      if cockroachTok == "" { missing = append(missing, "COCKROACH_API_KEY") }
      if aivenTok == "" { missing = append(missing, "AIVEN_TOKEN") }
      if cfTok == "" { missing = append(missing, "CLOUDFLARE_API_TOKEN") }
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
  // Read from the canonical .env.example file in the repository
  // This ensures consistency between the template and what gforge generates
  templatePath := ".env.example"
  
  // Check if template exists
  if _, err := os.Stat(templatePath); err == nil {
    // Copy template to target location
    content, err := os.ReadFile(templatePath)
    if err != nil {
      return fmt.Errorf("failed to read .env.example template: %w", err)
    }
    return os.WriteFile(p, content, 0o644)
  }
  
  // Fallback: generate minimal template if .env.example doesn't exist
  const minimalContent = `# Gothic Forge v3 - Environment Template
# Copy this file to .env and fill only what's necessary.
#
# Gothic Forge Opinionated Stack:
# - Cloudflare Pages + Functions (static hosting + edge compute)
# - Leapcell (Go backend compute)
# - CockroachDB Serverless (PostgreSQL-compatible database)
# - Aiven Valkey (Redis-compatible cache)

APP_ENV=development
SITE_BASE_URL=http://127.0.0.1:8080
HTTP_HOST=127.0.0.1
HTTP_PORT=8080
JWT_SECRET=

DATABASE_URL=
VALKEY_URL=

# Deployment tokens
LEAPCELL_APP_URL=
COCKROACH_API_KEY=
AIVEN_TOKEN=
CLOUDFLARE_API_TOKEN=

# For full configuration options, see: https://github.com/yourusername/gothic-forge
`
  return os.WriteFile(p, []byte(minimalContent), 0o644)
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

// printGitInstallHelp displays platform-specific Git installation instructions.
func printGitInstallHelp() {
  fmt.Println("  • Git installation guide:")
  switch runtime.GOOS {
  case "windows":
    fmt.Println("    - Git for Windows: https://git-scm.com/download/win")
    fmt.Println("    - Chocolatey: choco install git")
    fmt.Println("    - Winget: winget install Git.Git")
    fmt.Println("    - Scoop: scoop install git")
  case "darwin":
    fmt.Println("    - Xcode Command Line Tools: xcode-select --install")
    fmt.Println("    - Homebrew: brew install git")
    fmt.Println("    - Official installer: https://git-scm.com/download/mac")
  default: // linux
    fmt.Println("    - Debian/Ubuntu: sudo apt-get install git")
    fmt.Println("    - Fedora: sudo dnf install git")
    fmt.Println("    - Arch: sudo pacman -S git")
    fmt.Println("    - Other distros: https://git-scm.com/download/linux")
  }
  fmt.Println("  • After installation, verify with: git --version")
}

// printDockerInstallHelp displays platform-specific Docker installation instructions.
func printDockerInstallHelp() {
  fmt.Println("  • Docker installation required for Leapcell:")
  switch runtime.GOOS {
  case "windows":
    fmt.Println("    - Docker Desktop (recommended): https://docs.docker.com/desktop/install/windows-install/")
    fmt.Println("    - Requires WSL2 for better performance")
  case "darwin":
    fmt.Println("    - Docker Desktop for Mac: https://docs.docker.com/desktop/install/mac-install/")
    fmt.Println("    - Also available via Homebrew: brew install --cask docker")
  default: // linux
    fmt.Println("    - Docker Engine: https://docs.docker.com/engine/install/")
    fmt.Println("    - Or Docker Desktop: https://docs.docker.com/desktop/install/linux-install/")
  }
}

// validateDockerfile checks if the Dockerfile has valid syntax using docker build --dry-run.
// Returns nil if valid, error if invalid or check fails.
func validateDockerfile(path string) error {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  
  // Try with BuildKit enabled (required for --dry-run in some Docker versions)
  cmd := exec.CommandContext(ctx, "docker", "build", "--dry-run", "-f", path, ".")
  cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
  output, err := cmd.CombinedOutput()
  
  if err != nil {
    outStr := string(output)
    
    // Check if --dry-run is not supported
    if strings.Contains(outStr, "unknown flag") || strings.Contains(outStr, "requires BuildKit") {
      // Fall back to basic syntax check by parsing the Dockerfile
      return validateDockerfileSyntax(path)
    }
    
    // Parse output for actual syntax errors
    if strings.Contains(outStr, "ERROR") || strings.Contains(outStr, "failed to solve") {
      // Extract first error line for brevity
      lines := strings.Split(outStr, "\n")
      for _, line := range lines {
        if strings.Contains(line, "ERROR") {
          return fmt.Errorf("syntax error: %s", strings.TrimSpace(line))
        }
      }
      return fmt.Errorf("syntax error: %s", strings.TrimSpace(outStr[:min(200, len(outStr))]))
    }
    
    // Other errors (context timeout, etc)
    return err
  }
  
  return nil
}

// validateDockerfileSyntax performs basic Dockerfile syntax validation without Docker
func validateDockerfileSyntax(path string) error {
  content, err := os.ReadFile(path)
  if err != nil {
    return err
  }
  
  lines := strings.Split(string(content), "\n")
  for i, line := range lines {
    line = strings.TrimSpace(line)
    
    // Skip empty lines and comments
    if line == "" || strings.HasPrefix(line, "#") {
      continue
    }
    
    // Check for basic Dockerfile instruction format
    upper := strings.ToUpper(line)
    validInstructions := []string{"FROM", "RUN", "CMD", "LABEL", "EXPOSE", "ENV", "ADD", "COPY", "ENTRYPOINT", "VOLUME", "USER", "WORKDIR", "ARG", "ONBUILD", "STOPSIGNAL", "HEALTHCHECK", "SHELL"}
    
    hasValidInstruction := false
    for _, inst := range validInstructions {
      if strings.HasPrefix(upper, inst+" ") || strings.HasPrefix(upper, inst+"\t") || upper == inst {
        hasValidInstruction = true
        break
      }
    }
    
    // Allow continuation lines (backslash at end of previous line)
    if !hasValidInstruction && i > 0 && strings.HasSuffix(strings.TrimSpace(lines[i-1]), "\\") {
      hasValidInstruction = true
    }
    
    if !hasValidInstruction {
      return fmt.Errorf("line %d: invalid instruction or format", i+1)
    }
  }
  
  return nil
}

func min(a, b int) int {
  if a < b {
    return a
  }
  return b
}

// Note: ensureDockerFiles, createDockerfile, and createDockerignore are defined in install.go
// and shared across the cmd package.

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
