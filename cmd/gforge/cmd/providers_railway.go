package cmd

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
  "time"

  "gothicforge3/internal/execx"
)

// runRailwayDeploy orchestrates Railway CLI actions.
// - If dryRun is true: prints the planned steps based on available tokens and link state.
// - If false: executes a minimal, safe flow:
//     • verifies CLI availability
//     • if a .railway link exists, runs `railway up --detach`
//     • otherwise, prints guidance to link or use API token
func runRailwayDeploy(dryRun bool) error {
  // Detect CLI (optionally auto-install if missing and user requested install)
  railwayPath, ok := execx.Look("railway")
  if !ok {
    if deployInstall {
      if p, err := ensureRailwayCLI(); err == nil {
        railwayPath, ok = p, true
      } else {
        printRailwayInstallHelp()
        return fmt.Errorf("railway CLI not found: %w", err)
      }
    } else {
      printRailwayInstallHelp()
      return fmt.Errorf("railway CLI not found")
    }
  }

  // Build context for CLI calls
  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
  defer cancel()

  // Read tokens (may be temporarily suppressed during interactive flows)
  projTok := os.Getenv("RAILWAY_TOKEN")
  linked := isRailwayLinkedCLI(ctx)

  if dryRun {
    fmt.Println("  • Railway plan:")
    if projTok != "" {
      fmt.Println("    - Using RAILWAY_TOKEN (project token)")
      if linked {
        fmt.Println("    - Would run: railway up --detach")
      } else {
        fmt.Println("    - Not linked: run 'railway link' once (interactive) or use RAILWAY_API_TOKEN to create/link non-interactively")
      }
    } else {
      fmt.Println("    - No Railway project token set. For linking we use your login session.")
      if !linked && deployInitProject {
        fmt.Println("    - Would run: railway whoami (session)")
        fmt.Println("    - Would run: railway login (if session missing)")
        if deployLinkInstead {
          fmt.Println("    - Would run: railway link (interactive)")
        } else {
          fmt.Println("    - Would run: railway init (interactive)")
          fmt.Println("    - Would run: railway link (if not linked after init)")
        }
      }
      fmt.Println("    - Would run: railway up --detach")
    }
    return nil
  }

  // Execute minimal non-interactive flow
  _ = railwayPath // kept for clarity; execx.Run uses PATH anyway

  // For deploy-only when already linked, we don't need to touch auth here.

  if linked {
    // Deploy current directory to linked service/project (interactive to allow service prompts)
    if err := execx.RunInteractive(ctx, "railway up", railwayPath, "up", "--detach"); err != nil {
      return err
    }
    return nil
  }

  // Not linked
  if deployInitProject {
    // Ensure we have a valid login-backed session for account-level actions.
    if err := ensureLoggedInSession(ctx); err != nil {
      return err
    }
    if deployLinkInstead {
      // Link to existing project (no init). If link succeeds, proceed to up.
      {
        restore := withNoTokens()
        defer restore()
        if err := execx.RunInteractive(ctx, "railway link", railwayPath, "link"); err != nil {
          return fmt.Errorf("railway link failed: %w", err)
        }
      }
    } else {
      // Create a new project
      {
        restore := withNoTokens()
        defer restore()
        if err := execx.RunInteractive(ctx, "railway init", railwayPath, "init"); err != nil {
          return fmt.Errorf("railway init failed: %w", err)
        }
      }
    }
    // Deploy (interactive). If it fails due to linking, try to link once and retry.
    {
      restore := withNoTokens()
      defer restore()
      if err := execx.RunInteractive(ctx, "railway up", railwayPath, "up", "--detach"); err != nil {
        // Attempt link then retry once
        if linkErr := execx.RunInteractive(ctx, "railway link", railwayPath, "link"); linkErr == nil {
          if retryErr := execx.RunInteractive(ctx, "railway up", railwayPath, "up", "--detach"); retryErr == nil {
            return nil
          } else {
            return retryErr
          }
        }
        return err
      }
    }
    return nil
  }

  // Provide actionable guidance if not initializing automatically
  fmt.Println("    Run: railway link    (to select project/environment)")
  fmt.Println("         railway service (to select service)")
  fmt.Println("    Then re-run: gforge deploy --run")
  return nil
}

// withAPITokenPriority temporarily unsets RAILWAY_TOKEN when RAILWAY_API_TOKEN is present
// so that Railway CLI uses the account/team token for actions like whoami/init/link.
// It returns a restore function to reinstate the original environment.
// (removed withAPITokenPriority: no longer used; we prefer explicit withNoTokens/session logic)

// withNoTokens temporarily unsets both RAILWAY_TOKEN and RAILWAY_API_TOKEN so the CLI uses the login session.
func withNoTokens() func() {
  prevProj := os.Getenv("RAILWAY_TOKEN")
  prevApi := os.Getenv("RAILWAY_API_TOKEN")
  _ = os.Unsetenv("RAILWAY_TOKEN")
  _ = os.Unsetenv("RAILWAY_API_TOKEN")
  return func() {
    if prevProj != "" {
      _ = os.Setenv("RAILWAY_TOKEN", prevProj)
    }
    if prevApi != "" {
      _ = os.Setenv("RAILWAY_API_TOKEN", prevApi)
    }
  }
}

// ensureLoggedInSession verifies whoami using the login session; if missing, runs interactive login.
func ensureLoggedInSession(ctx context.Context) error {
  // Suppress tokens to force session-based auth
  restore := withNoTokens()
  defer restore()
  if err := execx.Run(ctx, "railway whoami", "railway", "whoami"); err != nil {
    if err2 := execx.RunInteractive(ctx, "railway login", "railway", "login"); err2 != nil {
      return fmt.Errorf("railway login failed: %w", err2)
    }
    // Re-check
    if err3 := execx.Run(ctx, "railway whoami", "railway", "whoami"); err3 != nil {
      return fmt.Errorf("railway whoami failed after login: %w", err3)
    }
  }
  return nil
}

func printRailwayInstallHelp() {
  fmt.Println("  • Railway CLI not found. Install using one of:")
  switch runtime.GOOS {
  case "windows":
    fmt.Println("    - Scoop: scoop install railway")
    fmt.Println("    - Direct binary (x64/arm64): https://github.com/railwayapp/cli/releases/latest")
  case "darwin":
    fmt.Println("    - Homebrew: brew install railway")
    fmt.Println("    - Direct binary (x64/arm64): https://github.com/railwayapp/cli/releases/latest")
  default: // linux and others
    fmt.Println("    - Direct binary (x64/arm64): https://github.com/railwayapp/cli/releases/latest")
  }
}

func dirExists(p string) bool {
  fi, err := os.Stat(p)
  if err != nil { return false }
  return fi.IsDir()
}

// isRailwayLinkedCLI returns true if the Railway CLI reports a linked project for the current directory.
// We use `railway status --json` and consider it linked when the JSON contains a non-empty project field.
func isRailwayLinkedCLI(ctx context.Context) bool {
  // If local link marker exists, treat as linked immediately.
  if dirExists(".railway") { return true }
  // If a project token is set, we can deploy without re-linking; treat as linked for UX.
  if os.Getenv("RAILWAY_TOKEN") != "" || os.Getenv("RAILWAY_API_TOKEN") != "" { return true }
  out, err := execx.RunCapture(ctx, "railway status --json", "railway", "status", "--json")
  if err != nil {
    // Some CLI versions return an error when not linked; fall back to local marker only.
    return dirExists(".railway")
  }
  s := struct{
    Project any `json:"project"`
  }{}
  if json.Unmarshal([]byte(out), &s) == nil {
    if s.Project != nil { return true }
  }
  if strings.Contains(out, "\"project\"") { return true }
  return dirExists(".railway")
}

// ensureRailwayCLI attempts a lean installation using native package managers available on the host.
// Returns absolute path to the installed binary if successful.
func ensureRailwayCLI() (string, error) {
  // If already in PATH, return it
  if p, ok := execx.Look("railway"); ok { return p, nil }

  arch := runtime.GOARCH
  if arch != "arm64" { arch = "amd64" }
  home, _ := os.UserHomeDir()
  binDir := filepath.Join(home, ".gforge", "bin")
  _ = os.MkdirAll(binDir, 0o755)

  // OS-specific install paths
  switch runtime.GOOS {
  case "windows":
    // Try Scoop if present
    if _, err := exec.LookPath("scoop"); err == nil {
      if err := execx.Run(context.Background(), "scoop install railway", "scoop", "install", "railway"); err == nil {
        if p, ok := execx.Look("railway"); ok { return p, nil }
      }
    }
    // Fallback to GitHub release
    url := fmt.Sprintf("https://github.com/railwayapp/cli/releases/latest/download/railway-windows-%s.exe", arch)
    dest := filepath.Join(binDir, "railway.exe")
    if err := downloadFile(url, dest); err == nil {
      _ = os.Chmod(dest, 0o755)
      // Prepend binDir to PATH for this process
      _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
      return dest, nil
    }
  case "darwin":
    // Prefer Homebrew
    if _, err := exec.LookPath("brew"); err == nil {
      if err := execx.Run(context.Background(), "brew install railway", "brew", "install", "railway"); err == nil {
        if p, ok := execx.Look("railway"); ok { return p, nil }
      }
    }
    // Fallback to GitHub release (try macos and darwin naming)
    for _, osName := range []string{"macos", "darwin"} {
      url := fmt.Sprintf("https://github.com/railwayapp/cli/releases/latest/download/railway-%s-%s", osName, arch)
      dest := filepath.Join(binDir, "railway")
      if err := downloadFile(url, dest); err == nil {
        _ = os.Chmod(dest, 0o755)
        _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
        return dest, nil
      }
    }
  default: // linux and others
    url := fmt.Sprintf("https://github.com/railwayapp/cli/releases/latest/download/railway-linux-%s", arch)
    dest := filepath.Join(binDir, "railway")
    if err := downloadFile(url, dest); err == nil {
      _ = os.Chmod(dest, 0o755)
      _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
      return dest, nil
    }
  }

  return "", errors.New("automatic install failed; please install from GitHub releases or package manager")
}

func downloadFile(url, dest string) error {
  resp, err := http.Get(url)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 { return fmt.Errorf("download failed: %s", resp.Status) }
  f, err := os.Create(dest)
  if err != nil { return err }
  defer f.Close()
  if _, err := io.Copy(f, resp.Body); err != nil { return err }
  return nil
}

// setRailwayEnv sets environment variables on the linked Railway service.
// - dryRun=true will only print the plan.
// - skips empty values.
func setRailwayEnv(ctx context.Context, kv map[string]string, dryRun bool) error {
  // Detect CLI
  if _, ok := execx.Look("railway"); !ok {
    printRailwayInstallHelp()
    return errors.New("railway CLI not found")
  }
  if dryRun {
    fmt.Println("  • Railway (env sync): plan")
    for k, v := range kv {
      if strings.TrimSpace(v) == "" { continue }
      fmt.Printf("    - set %s=***\n", k)
    }
    return nil
  }
  if !isRailwayLinkedCLI(ctx) {
    fmt.Println("  • Railway not linked; skipping env sync. Run 'railway link' or set RAILWAY_TOKEN.")
    return nil
  }
  // Apply variables
  for k, v := range kv {
    v = strings.TrimSpace(v)
    if v == "" { continue }
    if err := execx.RunInteractive(ctx, "railway variables set", "railway", "variables", "set", k+"="+v); err != nil {
      return fmt.Errorf("setting %s failed: %w", k, err)
    }
  }
  return nil
}
