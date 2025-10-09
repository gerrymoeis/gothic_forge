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
  deployCheck  bool
  deployInstall bool
  deployInitProject bool
  deployProjectName string
  deployServiceName string
  deployTeamSlug    string
  deployLinkInstead bool
  deployWithValkey  bool
  deployWithPages   bool
  deployNeonRegion  string
  deployNeonProject string
  deployNeonBranch  string
  deployNeonDBName  string
  deployNeonUser    string
  deployNeonPass    string
)

var deployCmd = &cobra.Command{
  Use:   "deploy",
  Short: "Deploy using omakase stack (Railway, Neon, Valkey, Cloudflare)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    _ = env.Load() // ensure .env is loaded for both normal and --dry-run flows
    if deployCheck {
      return runDeployPreflightCheck()
    }
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
      // Phase 2b: Ensure DATABASE_URL (Neon) — prefer auto-provision when NEON_TOKEN is set, else prompt for DSN
      if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
        fmt.Println("  • Neon: configuring database connection")
        var dsn string
        var err error
        if strings.TrimSpace(os.Getenv("NEON_TOKEN")) != "" {
          // Apply Neon overrides via env for this run
          if v := strings.TrimSpace(deployNeonRegion); v != "" { _ = os.Setenv("NEON_REGION", v) }
          if v := strings.TrimSpace(deployNeonProject); v != "" { _ = os.Setenv("NEON_PROJECT_NAME", v) }
          if v := strings.TrimSpace(deployNeonBranch); v != "" { _ = os.Setenv("NEON_BRANCH_NAME", v) }
          if v := strings.TrimSpace(deployNeonDBName); v != "" { _ = os.Setenv("NEON_DB_NAME", v) }
          if v := strings.TrimSpace(deployNeonUser); v != "" { _ = os.Setenv("NEON_DB_USER", v) }
          if v := strings.TrimSpace(deployNeonPass); v != "" { _ = os.Setenv("NEON_DB_PASSWORD", v) }
          dsn, err = neonAutoProvision(ctx, deployDryRun)
        } else {
          dsn, err = neonInteractiveProvision(ctx, deployDryRun)
        }
        if err != nil {
          fmt.Println("    → skipped Neon provisioning:", err)
        } else if strings.TrimSpace(dsn) != "" {
          fmt.Println("  • Running migrations")
          prev := dbMigrate
          dbMigrate = true
          if err := dbCmd.RunE(dbCmd, []string{}); err != nil {
            fmt.Println("    → migrations failed:", err)
          } else {
            fmt.Println("    → migrations complete")
          }
          dbMigrate = prev
        }
      }
      // Phase 3: Ensure REDIS_URL (Valkey) — optional with non-interactive/env/flag gating
      if strings.TrimSpace(os.Getenv("REDIS_URL")) == "" {
        nonInteractive := boolish(os.Getenv("GFORGE_NONINTERACTIVE"))
        wantValkey := deployWithValkey || boolish(os.Getenv("GFORGE_WITH_VALKEY"))
        if nonInteractive {
          if !wantValkey {
            fmt.Println("    → skipping Valkey setup (non-interactive)")
          } else if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
            fmt.Println("  • Valkey: configuring cache connection (non-interactive)")
            if vurl, verr := valkeyAutoProvision(ctx, false); verr != nil {
              fmt.Println("    → skipped Valkey provisioning:", verr)
            } else if strings.TrimSpace(vurl) != "" {
              fmt.Println("    → REDIS_URL configured")
            }
          } else {
            fmt.Println("    → skipping Valkey: non-interactive and AIVEN_TOKEN missing; set REDIS_URL or provide AIVEN_TOKEN")
          }
        } else {
          if !wantValkey {
            readerOpt := bufio.NewReader(os.Stdin)
            fmt.Print("  • Configure Valkey (Redis-compatible) now? [y/N]: ")
            ans, _ := readerOpt.ReadString('\n')
            ans = strings.ToLower(strings.TrimSpace(ans))
            if ans != "y" && ans != "yes" {
              fmt.Println("    → skipping Valkey setup (optional)")
            } else {
              fmt.Println("  • Valkey: configuring cache connection")
              var vurl string
              var verr error
              if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
                vurl, verr = valkeyAutoProvision(ctx, false)
              } else {
                vurl, verr = valkeyInteractiveProvision(ctx, false)
              }
              if verr != nil {
                fmt.Println("    → skipped Valkey provisioning:", verr)
              } else if strings.TrimSpace(vurl) != "" {
                fmt.Println("    → REDIS_URL configured")
              }
            }
          } else {
            fmt.Println("  • Valkey: configuring cache connection (pre-approved)")
            var vurl string
            var verr error
            if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
              vurl, verr = valkeyAutoProvision(ctx, false)
            } else {
              vurl, verr = valkeyInteractiveProvision(ctx, false)
            }
            if verr != nil {
              fmt.Println("    → skipped Valkey provisioning:", verr)
            } else if strings.TrimSpace(vurl) != "" {
              fmt.Println("    → REDIS_URL configured")
            }
          }
        }
      }
      fmt.Println("  • Running build to refresh static assets")
      if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
        fmt.Println("    → build failed:", err)
      } else {
        fmt.Println("    → build complete")
      }
      // Phase 4: Optional Cloudflare Pages deploy
      {
        nonInteractive := boolish(os.Getenv("GFORGE_NONINTERACTIVE"))
        wantPages := deployWithPages || boolish(os.Getenv("GFORGE_WITH_PAGES"))
        if nonInteractive {
          if !wantPages {
            fmt.Println("    → skipping Cloudflare Pages (non-interactive)")
          } else {
            fmt.Println("  • Cloudflare Pages: deploying static export (non-interactive)")
            pagesProject = strings.TrimSpace(os.Getenv("CF_PROJECT_NAME"))
            pagesDeployRun = true
            if err := deployPagesCmd.RunE(deployPagesCmd, []string{}); err != nil {
              fmt.Println("    → pages deploy failed:", err)
            }
          }
        } else {
          if !wantPages {
            readerP := bufio.NewReader(os.Stdin)
            fmt.Print("  • Deploy static export to Cloudflare Pages now? [y/N]: ")
            ans, _ := readerP.ReadString('\n')
            ans = strings.ToLower(strings.TrimSpace(ans))
            wantPages = (ans == "y" || ans == "yes")
          }
          if wantPages {
            fmt.Println("  • Cloudflare Pages: deploying static export")
            pagesProject = strings.TrimSpace(os.Getenv("CF_PROJECT_NAME"))
            pagesDeployRun = true
            if err := deployPagesCmd.RunE(deployPagesCmd, []string{}); err != nil {
              fmt.Println("    → pages deploy failed:", err)
            }
          }
        }
      }
      // Interactive provider flow (chat-style)
      reader := bufio.NewReader(os.Stdin)
      // Railway env sync (push .env variables to linked service)
      {
        kv := loadEnvFile(".env")
        filtered := map[string]string{
          "SITE_BASE_URL":          strings.TrimSpace(kv["SITE_BASE_URL"]),
          "JWT_SECRET":             strings.TrimSpace(kv["JWT_SECRET"]),
          "DATABASE_URL":           strings.TrimSpace(kv["DATABASE_URL"]),
          "REDIS_URL":              strings.TrimSpace(kv["REDIS_URL"]),
          "VALKEY_URL":             strings.TrimSpace(kv["VALKEY_URL"]),
          "VALKEY_TLS_SKIP_VERIFY": strings.TrimSpace(kv["VALKEY_TLS_SKIP_VERIFY"]),
        }
        _ = setRailwayEnv(context.Background(), filtered, false)
      }
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
    // Phase 2b: Show Neon auto-provision plan if NEON_TOKEN is set; otherwise show DSN prompt plan
    if strings.TrimSpace(os.Getenv("NEON_TOKEN")) != "" {
      _, _ = neonAutoProvision(context.Background(), true)
    } else {
      _, _ = neonInteractiveProvision(context.Background(), true)
    }
    {
      nonInteractive := boolish(os.Getenv("GFORGE_NONINTERACTIVE"))
      wantValkey := deployWithValkey || boolish(os.Getenv("GFORGE_WITH_VALKEY"))
      if !wantValkey && nonInteractive {
        fmt.Println("  • Valkey (dry-run): would skip (non-interactive)")
      } else if wantValkey || !nonInteractive {
        if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
          _, _ = valkeyAutoProvision(context.Background(), true)
        } else {
          _, _ = valkeyInteractiveProvision(context.Background(), true)
        }
      }
    }
    // Cloudflare Pages dry-run plan
    {
      nonInteractive := boolish(os.Getenv("GFORGE_NONINTERACTIVE"))
      wantPages := deployWithPages || boolish(os.Getenv("GFORGE_WITH_PAGES"))
      if !wantPages && nonInteractive {
        fmt.Println("  • Cloudflare Pages (dry-run): would skip (non-interactive)")
      } else if wantPages || !nonInteractive {
        fmt.Println("  • Cloudflare Pages (dry-run): would run wrangler pages deploy dist --project-name $CF_PROJECT_NAME")
      }
    }
    // Railway env sync (dry-run): show what would be pushed if linked
    {
      kv := loadEnvFile(".env")
      filtered := map[string]string{
        "SITE_BASE_URL":          strings.TrimSpace(kv["SITE_BASE_URL"]),
        "JWT_SECRET":             strings.TrimSpace(kv["JWT_SECRET"]),
        "DATABASE_URL":           strings.TrimSpace(kv["DATABASE_URL"]),
        "REDIS_URL":              strings.TrimSpace(kv["REDIS_URL"]),
        "VALKEY_URL":             strings.TrimSpace(kv["VALKEY_URL"]),
        "VALKEY_TLS_SKIP_VERIFY": strings.TrimSpace(kv["VALKEY_TLS_SKIP_VERIFY"]),
      }
      _ = setRailwayEnv(context.Background(), filtered, true)
    }
    _ = runRailwayDeploy(true)
    fmt.Println("Deployment flow stub complete. (More integrations to follow)")
    return nil
  },
}

func init() {
  deployCmd.Flags().BoolVar(&deployProd, "prod", false, "use production settings")
  deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "show steps without executing")
  deployCmd.Flags().BoolVar(&deployCheck, "check", false, "preflight checks for tools, tokens, env; no writes or external actions")
  deployCmd.Flags().BoolVar(&deployInstall, "install-tools", false, "attempt to auto-install missing provider CLIs (e.g., Railway)")
  deployCmd.Flags().BoolVar(&deployRun, "run", false, "execute provider CLIs (Railway, etc.) after build")
  deployCmd.Flags().BoolVar(&deployInitProject, "init-project", false, "create/link Railway project if missing (requires RAILWAY_API_TOKEN)")
  deployCmd.Flags().StringVar(&deployProjectName, "project-name", "gothic-forge-v3", "Railway project name to create/use")
  deployCmd.Flags().StringVar(&deployServiceName, "service-name", "web", "Railway service name to create/use for this directory")
  deployCmd.Flags().StringVar(&deployTeamSlug, "team", "", "Railway team slug (optional)")
  deployCmd.Flags().BoolVar(&deployWithValkey, "with-valkey", false, "configure Valkey/Redis cache during deploy (optional)")
  deployCmd.Flags().BoolVar(&deployWithPages, "with-pages", false, "deploy static export to Cloudflare Pages (optional)")
  deployCmd.Flags().StringVar(&deployNeonRegion, "neon-region", "", "Neon region id (e.g., aws-us-east-1); overrides NEON_REGION for this run")
  deployCmd.Flags().StringVar(&deployNeonProject, "neon-project", "", "Neon project name; overrides NEON_PROJECT_NAME for this run")
  deployCmd.Flags().StringVar(&deployNeonBranch, "neon-branch", "", "Neon branch name; overrides NEON_BRANCH_NAME for this run")
  deployCmd.Flags().StringVar(&deployNeonDBName, "neon-db", "", "Neon database name; overrides NEON_DB_NAME for this run")
  deployCmd.Flags().StringVar(&deployNeonUser, "neon-user", "", "Neon database role/user; overrides NEON_DB_USER for this run")
  deployCmd.Flags().StringVar(&deployNeonPass, "neon-password", "", "Neon database password; overrides NEON_DB_PASSWORD for this run")
  rootCmd.AddCommand(deployCmd)
}

// runDeployPreflightCheck validates tools, tokens, .env, Railway link, and Pages config without modifying state.
func runDeployPreflightCheck() error {
  fmt.Println("Deploy preflight check")

  // Tools
  railPath, railOK := execx.Look("railway")
  wrPath, wrOK := execx.Look("wrangler")
  fmt.Printf("  • railway: %s\n", pathOrMissing(railPath, railOK))
  if railOK {
    if v, err := execx.RunCapture(context.Background(), "railway --version", "railway", "--version"); err == nil {
      v = strings.TrimSpace(v); if v != "" { fmt.Printf("    → %s\n", v) }
    }
  }
  fmt.Printf("  • wrangler: %s\n", pathOrMissing(wrPath, wrOK))
  if wrOK {
    if v, err := execx.RunCapture(context.Background(), "wrangler --version", "wrangler", "--version"); err == nil {
      v = strings.TrimSpace(v); if v != "" { fmt.Printf("    → %s\n", v) }
    }
  }

  // Env presence
  envPath := ".env"
  if _, err := os.Stat(envPath); os.IsNotExist(err) {
    fmt.Println("  • .env: not found")
  } else { fmt.Println("  • .env: present") }
  kv := loadEnvFile(envPath)

  // Core env checks
  siteBase := strings.TrimSpace(kv["SITE_BASE_URL"])
  jwt := strings.TrimSpace(kv["JWT_SECRET"])
  if siteBase == "" || strings.HasPrefix(siteBase, "http://127.0.0.1") || strings.HasPrefix(siteBase, "http://localhost") {
    fmt.Println("  • Warning: SITE_BASE_URL looks dev-like (set your public https URL)")
  } else {
    fmt.Println("  • SITE_BASE_URL:", siteBase)
  }
  jwtOK := (jwt != "" && len(jwt) >= 32 && !strings.EqualFold(jwt, "devsecret-change-me"))
  if !jwtOK { fmt.Println("  • JWT_SECRET: MISSING or weak (>=32 hex chars recommended)") }

  // Provider tokens and config
  railTok := strings.TrimSpace(kv["RAILWAY_TOKEN"]) != ""
  railApiTok := strings.TrimSpace(kv["RAILWAY_API_TOKEN"]) != ""
  neonTok := strings.TrimSpace(kv["NEON_TOKEN"]) != ""
  aivenTok := strings.TrimSpace(kv["AIVEN_TOKEN"]) != ""
  cfTok := strings.TrimSpace(kv["CF_API_TOKEN"]) != ""
  cfAcct := strings.TrimSpace(kv["CF_ACCOUNT_ID"]) != ""
  cfProj := strings.TrimSpace(kv["CF_PROJECT_NAME"]) != ""
  dbSet := strings.TrimSpace(kv["DATABASE_URL"]) != ""
  redisSet := strings.TrimSpace(kv["REDIS_URL"]) != "" || strings.TrimSpace(kv["VALKEY_URL"]) != ""

  // Gating based on env for optional providers
  wantValkey := boolish(os.Getenv("GFORGE_WITH_VALKEY")) || deployWithValkey
  wantPages := boolish(os.Getenv("GFORGE_WITH_PAGES")) || deployWithPages

  // Railway link state (read-only)
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  linked := isRailwayLinkedCLI(ctx)
  cancel()
  if linked {
    fmt.Println("  • Railway: linked")
  } else {
    fmt.Println("  • Railway: not linked (run 'railway link' or allow wizard to init)")
  }
  // Token presence info (advisory)
  fmt.Printf("  • Railway tokens: RAILWAY_TOKEN=%s, RAILWAY_API_TOKEN=%s\n",
    presentOrMissing(railTok), presentOrMissing(railApiTok))

  // Compute readiness
  ready := true
  missing := []string{}
  // Tools
  if !railOK { ready = false; missing = append(missing, "railway CLI") }
  // Core env
  if !jwtOK { ready = false; missing = append(missing, "JWT_SECRET (strong)") }
  // DB
  if !(dbSet || neonTok) { ready = false; missing = append(missing, "DATABASE_URL or NEON_TOKEN") }
  // Valkey (optional)
  if wantValkey {
    if !(redisSet || aivenTok) { ready = false; missing = append(missing, "REDIS_URL or AIVEN_TOKEN") }
  }
  // Pages (optional)
  if wantPages {
    if !(cfTok && cfAcct && cfProj) { ready = false; missing = append(missing, "CF_API_TOKEN, CF_ACCOUNT_ID, CF_PROJECT_NAME") }
    // Note: wrangler missing does not fail readiness (you can deploy via GitHub Action), but we warn
    if !wrOK { fmt.Println("  • Warning: wrangler missing; local pages deploy will be skipped (CI Pages action still works)") }
  }

  fmt.Println("────────────────────────────────────────")
  fmt.Println("Preflight summary")
  if ready {
    fmt.Println("  • Ready for deploy: Yes")
    return nil
  }
  fmt.Println("  • Ready for deploy: No")
  if len(missing) > 0 {
    fmt.Println("  • Missing:")
    for _, m := range missing { fmt.Println("    - ", m) }
  }
  return fmt.Errorf("preflight failed")
}

func presentOrMissing(b bool) string {
  if b { return "present" }
  return "missing"
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
  // 2) JWT_SECRET (with generator)
  if strings.TrimSpace(kv["JWT_SECRET"]) == "" {
    fmt.Print("  • Generate JWT_SECRET now? [Y/n]: ")
    ans, _ := reader.ReadString('\n')
    ans = strings.ToLower(strings.TrimSpace(ans))
    if ans == "" || ans == "y" || ans == "yes" {
      kv["JWT_SECRET"] = genSecret()
      fmt.Println("    → JWT_SECRET generated")
    } else {
      fmt.Printf("  • Enter %s (leave blank to skip): ", "JWT_SECRET")
      val, _ := reader.ReadString('\n')
      kv["JWT_SECRET"] = strings.TrimSpace(val)
    }
  }
  // Strengthen: regenerate if weak/dev default
  if js := strings.TrimSpace(kv["JWT_SECRET"]); js == "" || strings.EqualFold(js, "devsecret-change-me") || len(js) < 32 {
    fmt.Print("  • JWT_SECRET is weak or missing. Generate a strong one now? [Y/n]: ")
    ans, _ := reader.ReadString('\n')
    ans = strings.ToLower(strings.TrimSpace(ans))
    if ans == "" || ans == "y" || ans == "yes" {
      kv["JWT_SECRET"] = genSecret()
      fmt.Println("    → JWT_SECRET regenerated")
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

// boolish interprets common truthy values.
func boolish(v string) bool {
  s := strings.TrimSpace(strings.ToLower(v))
  return s == "1" || s == "true" || s == "yes" || s == "y" || s == "on"
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
