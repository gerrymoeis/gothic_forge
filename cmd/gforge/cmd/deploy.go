package cmd

import (
  "bufio"
  "context"
  "crypto/rand"
  "encoding/hex"
  "fmt"
  "os"
  "path/filepath"
  "strings"
  "time"

  "gothicforge3/internal/env"
  "gothicforge3/internal/execx"
  "github.com/spf13/cobra"
)

var (
  deployProd   bool
  deployDryRun bool
  deployRun    bool
  deployInstall bool
  deployInitProject bool
  deployProjectName string
  deployServiceName string
  deployTeamSlug    string
  deployLinkInstead bool
)

var deployCmd = &cobra.Command{
  Use:   "deploy",
  Short: "Deploy using omakase stack (Railway, Neon, Valkey, Cloudflare)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    _ = env.Load() // ensure .env is loaded for both normal and --dry-run flows
    if deployDryRun {
      fmt.Println("Deploy (dry-run)")
    } else {
      fmt.Println("Deploy wizard")
    }

    // Check required secrets/env
    required := []string{"RAILWAY_TOKEN", "NEON_TOKEN", "AIVEN_TOKEN", "CF_API_TOKEN"}
    missing := []string{}
    for _, k := range required {
      if os.Getenv(k) == "" { missing = append(missing, k) }
    }
    siteBase := os.Getenv("SITE_BASE_URL")
    apiTok := os.Getenv("RAILWAY_API_TOKEN")

    fmt.Println("  • Checking secrets:")
    for _, k := range required {
      v := os.Getenv(k)
      if v == "" {
        fmt.Printf("    - %s: MISSING\n", k)
      } else {
        fmt.Printf("    - %s: present\n", k)
      }
    }
    if apiTok == "" {
      fmt.Println("    - RAILWAY_API_TOKEN: not set (optional, enables project creation)")
    } else {
      fmt.Println("    - RAILWAY_API_TOKEN: present")
    }
    if siteBase == "" {
      fmt.Println("    - SITE_BASE_URL: not set (will default to '/')")
    } else {
      fmt.Println("    - SITE_BASE_URL:", siteBase)
    }

    // Helpful provider links for sign-up and tokens (show in dry-run only; interactive flow shows links inline per prompt)
    if deployDryRun {
      fmt.Println("  • Provider links:")
      fmt.Println("    - Railway:", "https://railway.app")
      fmt.Println("    - Neon API keys:", "https://neon.tech/docs/manage/api-keys")
      fmt.Println("    - Aiven tokens:", "https://docs.aiven.io/docs/platform/howto/create_authentication_token")
      fmt.Println("    - Cloudflare API tokens:", "https://dash.cloudflare.com/profile/api-tokens")
    }

    // Ensure SEO files exist
    if _, err := os.Stat(filepath.Join("app", "static", "sitemap.xml")); err == nil {
      fmt.Println("  • sitemap.xml: found under app/static")
    } else {
      fmt.Println("  • sitemap.xml: not found (run 'gforge build')")
    }
    if _, err := os.Stat(filepath.Join("app", "static", "robots.txt")); err == nil {
      fmt.Println("  • robots.txt: found under app/static")
    } else {
      fmt.Println("  • robots.txt: not found (run 'gforge build')")
    }

    fmt.Println("  • Preparing build artifacts and static assets")
    fmt.Println("  • Provisioning Neon (Postgres)")
    fmt.Println("  • Provisioning Aiven Valkey")
    fmt.Println("  • Configuring Railway service & env")
    fmt.Println("  • Publishing static assets to Cloudflare Pages")
    if deployProd { fmt.Println("  • Using production settings") }
    if deployDryRun { fmt.Println("  • Dry-run: no external calls executed") }

    if !deployDryRun {
      // Interactive env setup only when not linked (first-time setup). Skip for subsequent deploys.
      ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
      defer cancel()
      if !isRailwayLinkedCLI(ctx) {
        if err := interactiveEnvSetup(); err != nil {
          fmt.Println("────────────────────────────────────────")
          fmt.Println("Env setup aborted:", err)
          return nil
        }
      }
      fmt.Println("  • Running build to refresh static assets")
      if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
        fmt.Println("    → build failed:", err)
      } else {
        fmt.Println("    → build complete")
      }
      // Interactive provider flow (chat-style)
      reader := bufio.NewReader(os.Stdin)
      // Offer to install Railway CLI if missing
      if _, ok := execx.Look("railway"); !ok {
        fmt.Print("  • Railway CLI not found. Install now? [Y/n]: ")
        ans, _ := reader.ReadString('\n')
        ans = strings.ToLower(strings.TrimSpace(ans))
        if ans == "" || ans == "y" || ans == "yes" { deployInstall = true }
      }
      // Offer to create or link if not linked (CLI-based detection)
      ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
      defer cancel2()
      if !isRailwayLinkedCLI(ctx2) {
        fmt.Println("  • No Railway project link detected.")
        fmt.Println("    1) Create new Railway project (init)")
        fmt.Println("    2) Link to existing project")
        fmt.Println("    3) Skip for now")
        fmt.Print("    Select [1/2/3]: ")
        ans, _ := reader.ReadString('\n')
        ans = strings.TrimSpace(ans)
        switch ans {
        case "2":
          deployInitProject = true
          deployLinkInstead = true
        case "3":
          deployInitProject = false
        default: // "1" or empty → init
          deployInitProject = true
          deployLinkInstead = false
        }
      }
      // Confirm deploy (skip confirmation when already linked for seamless updates)
      doDeploy := deployRun // allow --run to auto-confirm
      {
        ctx3, cancel3 := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel3()
        if isRailwayLinkedCLI(ctx3) { doDeploy = true }
      }
      if !doDeploy {
        fmt.Print("  • Proceed with Railway deploy now? [Y/n]: ")
        ans, _ := reader.ReadString('\n')
        ans = strings.ToLower(strings.TrimSpace(ans))
        doDeploy = (ans == "" || ans == "y" || ans == "yes")
      }
      if doDeploy {
        if err := runRailwayDeploy(false); err != nil {
          fmt.Println("  • Railway deploy error:", err)
        } else {
          fmt.Println("────────────────────────────────────────")
          fmt.Println("Deployment steps executed. Review your Railway dashboard for status.")
        }
        return nil
      }
      fmt.Println("────────────────────────────────────────")
      fmt.Println("You can re-run deployment anytime with: gforge deploy --run")
      return nil
    }

    if len(missing) > 0 {
      fmt.Println("────────────────────────────────────────")
      fmt.Println("Some required secrets are missing. Set them with:")
      for _, k := range missing {
        fmt.Printf("  gforge secrets --set %s=...\n", k)
      }
      fmt.Println()
      fmt.Println("Quick links:")
      fmt.Println("  Railway: https://railway.app")
      fmt.Println("  Neon API keys: https://neon.tech/docs/manage/api-keys")
      fmt.Println("  Aiven tokens: https://docs.aiven.io/docs/platform/howto/create_authentication_token")
      fmt.Println("  Cloudflare API tokens: https://dash.cloudflare.com/profile/api-tokens")
      return nil
    }

    fmt.Println("────────────────────────────────────────")
    // Dry-run provider steps
    _ = runRailwayDeploy(true)
    fmt.Println("Deployment flow stub complete. (More integrations to follow)")
    return nil
  },
}

func init() {
  deployCmd.Flags().BoolVar(&deployProd, "prod", false, "use production settings")
  deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "show steps without executing")
  deployCmd.Flags().BoolVar(&deployInstall, "install-tools", false, "attempt to auto-install missing provider CLIs (e.g., Railway)")
  deployCmd.Flags().BoolVar(&deployRun, "run", false, "execute provider CLIs (Railway, etc.) after build")
  deployCmd.Flags().BoolVar(&deployInitProject, "init-project", false, "create/link Railway project if missing (requires RAILWAY_API_TOKEN)")
  deployCmd.Flags().StringVar(&deployProjectName, "project-name", "gothic-forge-v3", "Railway project name to create/use")
  deployCmd.Flags().StringVar(&deployServiceName, "service-name", "web", "Railway service name to create/use for this directory")
  deployCmd.Flags().StringVar(&deployTeamSlug, "team", "", "Railway team slug (optional)")
  rootCmd.AddCommand(deployCmd)
}

// interactiveEnvSetup ensures .env exists (copying from .env.example if present),
// prompts for missing values, and writes them back.
func interactiveEnvSetup() error {
  envPath := ".env"
  examplePath := ".env.example"
  if _, err := os.Stat(envPath); os.IsNotExist(err) {
    if b, err2 := os.ReadFile(examplePath); err2 == nil {
      if err3 := os.WriteFile(envPath, b, 0o600); err3 != nil { return err3 }
      fmt.Println("  • Created .env from .env.example")
    } else {
      // Create minimal .env if no example
      if err3 := os.WriteFile(envPath, []byte("APP_ENV=production\n"), 0o600); err3 != nil { return err3 }
      fmt.Println("  • Created minimal .env (APP_ENV=production)")
    }
  }

  kv := loadEnvFile(envPath)
  reader := bufio.NewReader(os.Stdin)

  // Ensure APP_ENV
  curEnv := strings.ToLower(strings.TrimSpace(kv["APP_ENV"]))
  if curEnv == "" {
    kv["APP_ENV"] = "production"
    fmt.Println("  • APP_ENV was empty → set to 'production'")
  } else if curEnv != "production" {
    fmt.Printf("  • APP_ENV is '%s' → switching to 'production' for deployment\n", curEnv)
    kv["APP_ENV"] = "production"
  }

  // Required / recommended keys (structured prompts with provider links)
  // 1) SITE_BASE_URL first
  if strings.TrimSpace(kv["SITE_BASE_URL"]) == "" {
    fmt.Printf("  • Enter %s (leave blank to skip): ", "SITE_BASE_URL")
    val, _ := reader.ReadString('\n')
    kv["SITE_BASE_URL"] = strings.TrimSpace(val)
  }
  // 2) SESSION_SECRET (with generator)
  if strings.TrimSpace(kv["SESSION_SECRET"]) == "" {
    fmt.Print("  • Generate SESSION_SECRET now? [Y/n]: ")
    ans, _ := reader.ReadString('\n')
    ans = strings.ToLower(strings.TrimSpace(ans))
    if ans == "" || ans == "y" || ans == "yes" {
      kv["SESSION_SECRET"] = genSecret()
      fmt.Println("    → SESSION_SECRET generated")
    } else {
      fmt.Printf("  • Enter %s (leave blank to skip): ", "SESSION_SECRET")
      val, _ := reader.ReadString('\n')
      kv["SESSION_SECRET"] = strings.TrimSpace(val)
    }
  }
  // 3) Provider tokens with links shown inline
  type tok struct{ key, label, link string }
  tokens := []tok{
    {"RAILWAY_API_TOKEN", "Railway API tokens", "https://railway.app/account/tokens"},
    {"RAILWAY_TOKEN", "Railway Project token", "https://railway.app"},
    {"NEON_TOKEN", "Neon API keys", "https://neon.tech/docs/manage/api-keys"},
    {"AIVEN_TOKEN", "Aiven tokens", "https://docs.aiven.io/docs/platform/howto/create_authentication_token"},
    {"CF_API_TOKEN", "Cloudflare API tokens", "https://dash.cloudflare.com/profile/api-tokens"},
  }
  for _, t := range tokens {
    if strings.TrimSpace(kv[t.key]) != "" { continue }
    fmt.Printf("  • %s: %s\n", t.label, t.link)
    fmt.Printf("  • Enter %s (leave blank to skip): ", t.key)
    val, _ := reader.ReadString('\n')
    kv[t.key] = strings.TrimSpace(val)
  }

  // If SITE_BASE_URL looks like a dev default, offer to change
  if sb := strings.TrimSpace(kv["SITE_BASE_URL"]); sb == "" || sb == "http://127.0.0.1:8080" {
    fmt.Print("  • SITE_BASE_URL looks dev-like. Provide production URL (https://...)? [leave blank to keep]: ")
    val, _ := reader.ReadString('\n')
    val = strings.TrimSpace(val)
    if val != "" {
      kv["SITE_BASE_URL"] = normalizeBaseURL(val)
    }
  }

  // Normalize SITE_BASE_URL if present
  if sb := strings.TrimSpace(kv["SITE_BASE_URL"]); sb != "" {
    kv["SITE_BASE_URL"] = normalizeBaseURL(sb)
  }

  // Prefer rewriting from .env.example template to preserve full structure
  if fileStartsWithWizardHeader(envPath) {
    if _, err := os.Stat(examplePath); err == nil {
      if err := rewriteEnvFromExample(envPath, examplePath, kv); err == nil {
        fmt.Println("  • Wrote .env using .env.example structure (preserved comments & layout)")
        return nil
      }
    }
  }
  if err := updateEnvFileInPlace(envPath, kv); err != nil { return err }
  fmt.Println("  • Wrote .env with updated values (preserved existing layout)")
  return nil
}

func loadEnvFile(path string) map[string]string {
  kv := map[string]string{}
  b, err := os.ReadFile(path)
  if err != nil { return kv }
  lines := strings.Split(string(b), "\n")
  for _, ln := range lines {
    ln = strings.TrimSpace(ln)
    if ln == "" || strings.HasPrefix(ln, "#") { continue }
    parts := strings.SplitN(ln, "=", 2)
    if len(parts) == 2 {
      kv[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
    }
  }
  return kv
}

// saveEnvFile removed (superseded by updateEnvFileInPlace/rewriteEnvFromExample)

func genSecret() string {
  buf := make([]byte, 32)
  if _, err := rand.Read(buf); err != nil { return "" }
  return hex.EncodeToString(buf)
}

// fileStartsWithWizardHeader detects if the .env was auto-generated by a previous wizard run.
func fileStartsWithWizardHeader(path string) bool {
  f, err := os.Open(path)
  if err != nil { return false }
  defer f.Close()
  r := bufio.NewReader(f)
  line, _ := r.ReadString('\n')
  return strings.HasPrefix(strings.TrimSpace(line), "# Generated by gforge deploy wizard")
}

// rewriteEnvFromExample rewrites envPath using examplePath's structure, substituting
// values from kv where keys match. Comments and blank lines are preserved from example.
func rewriteEnvFromExample(envPath, examplePath string, kv map[string]string) error {
  b, err := os.ReadFile(examplePath)
  if err != nil { return err }
  lines := strings.Split(string(b), "\n")
  // Track keys we substituted
  used := map[string]bool{}
  for i, ln := range lines {
    t := strings.TrimSpace(ln)
    if t == "" || strings.HasPrefix(t, "#") {
      continue
    }
    if idx := strings.Index(ln, "="); idx > 0 {
      key := strings.TrimSpace(ln[:idx])
      if val, ok := kv[key]; ok {
        lines[i] = key + "=" + val
        used[key] = true
      }
    }
  }
  // Append any extra keys not present in example
  extra := []string{}
  for k, v := range kv {
    if !used[k] {
      extra = append(extra, k+"="+v)
    }
  }
  if len(extra) > 0 {
    lines = append(lines, "", "# Added by gforge deploy wizard")
    lines = append(lines, extra...)
  }
  out := strings.Join(lines, "\n")
  return os.WriteFile(envPath, []byte(out), 0o600)
}

// normalizeBaseURL ensures the URL has a scheme and no trailing slash (unless root)
func normalizeBaseURL(val string) string {
  v := strings.TrimSpace(val)
  if v == "" { return v }
  if !(strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")) {
    // default to https for production URLs
    v = "https://" + v
  }
  if v != "/" { v = strings.TrimRight(v, "/") }
  return v
}

// updateEnvFileInPlace updates only values for existing keys in .env, preserving
// its current structure and comments. Any missing keys are appended at the end.
func updateEnvFileInPlace(envPath string, kv map[string]string) error {
  b, err := os.ReadFile(envPath)
  if err != nil { return err }
  lines := strings.Split(string(b), "\n")
  pending := map[string]string{}
  for k, v := range kv { pending[k] = v }
  for i, ln := range lines {
    t := strings.TrimSpace(ln)
    if t == "" || strings.HasPrefix(t, "#") { continue }
    if idx := strings.Index(ln, "="); idx > 0 {
      key := strings.TrimSpace(ln[:idx])
      if val, ok := pending[key]; ok {
        lines[i] = key + "=" + val
        delete(pending, key)
      }
    }
  }
  if len(pending) > 0 {
    lines = append(lines, "", "# Added by gforge deploy wizard")
    for k, v := range pending {
      lines = append(lines, k+"="+v)
    }
  }
  out := strings.Join(lines, "\n")
  return os.WriteFile(envPath, []byte(out), 0o600)
}
