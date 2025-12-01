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
	deployProd       bool
	deployDryRun     bool
	deployCheck      bool
	deployWithValkey bool
	deployWithPages  bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy using the Opinionated Stack (Leapcell + CockroachDB + Aiven Valkey + Cloudflare)",
	Long: `Deploy your Gothic Forge application using the Opinionated Stack:
  • Compute: Leapcell (20 free projects on Hobby tier)
  • Database: CockroachDB Serverless
  • Cache: Aiven Valkey (Redis-compatible)
  • CDN/Proxy: Cloudflare Pages (optional)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		_ = env.Load() // ensure .env is loaded for both normal and --dry-run flows

		if deployCheck {
			return runDeployPreflightCheck()
		}
		if deployDryRun {
			fmt.Println("Deploy (dry-run) - Opinionated Stack")
			fmt.Println("  • Compute: Leapcell")
			fmt.Println("  • Database: CockroachDB Serverless")
			fmt.Println("  • Cache: Aiven Valkey")
			fmt.Println("  • CDN: Cloudflare (optional)")
		} else {
			fmt.Println("Deploy wizard - Opinionated Stack")
			fmt.Println("  • Compute: Leapcell")
			fmt.Println("  • Database: CockroachDB Serverless")
			fmt.Println("  • Cache: Aiven Valkey")
			fmt.Println("  • CDN: Cloudflare (optional)")
		}

		// Check Opinionated Stack configuration
		missing := []string{}

		// Check database provider (CockroachDB only)
		hasCockroach := strings.TrimSpace(os.Getenv("COCKROACH_API_KEY")) != ""
		if !hasCockroach {
			missing = append(missing, "COCKROACH_API_KEY")
		}

		siteBase := os.Getenv("SITE_BASE_URL")

		fmt.Println("  • Checking Opinionated Stack configuration:")

		// Check database provider first
		if hasCockroach {
			fmt.Println("    - COCKROACH_API_KEY: present (CockroachDB Serverless)")
		} else {
			fmt.Println("    - Database provider: MISSING (need COCKROACH_API_KEY)")
		}

		// Leapcell configuration
		leapURL := os.Getenv("LEAPCELL_APP_URL")
		if leapURL == "" {
			fmt.Println("    - LEAPCELL_APP_URL: not set (will be saved after guided setup)")
		} else {
			fmt.Println("    - LEAPCELL_APP_URL:", leapURL)
		}

		if siteBase == "" {
			fmt.Println("    - SITE_BASE_URL: not set (will default to '/')")
		} else {
			fmt.Println("    - SITE_BASE_URL:", siteBase)
		}

		// Opinionated Stack provider links (show in dry-run only; interactive flow shows links inline per prompt)
		if deployDryRun {
			fmt.Println("  • Opinionated Stack provider links:")
			fmt.Println("    - Leapcell:", "https://leapcell.io/signup")
			fmt.Println("    - Leapcell Docs:", "https://docs.leapcell.io/")
			fmt.Println("    - 🎁 20 FREE projects on Hobby tier!")
			fmt.Println("    - CockroachDB:", "https://cockroachlabs.cloud/signup")
			fmt.Println("    - CockroachDB service accounts:", "https://cockroachlabs.cloud/service-accounts")
			fmt.Println("    - Aiven tokens:", "https://console.aiven.io/profile/tokens")
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
		fmt.Println("  • Provisioning CockroachDB (Serverless)")
		fmt.Println("  • Provisioning Aiven Valkey")
		fmt.Println("  • Guided Leapcell deployment")
		fmt.Println("  • Cloudflare Pages (optional)")
		if deployProd {
			fmt.Println("  • Using production settings")
		}
		if deployDryRun {
			fmt.Println("  • Dry-run: no external calls executed")
		}

		if !deployDryRun {
			// Interactive env setup for first-time deployment
			if err := interactiveEnvSetup(); err != nil {
				fmt.Println("────────────────────────────────────────")
				fmt.Println("Env setup aborted:", err)
				return nil
			}
			// Phase 2b: Ensure DATABASE_URL — CockroachDB is the opinionated Gothic Forge standard
			if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
				var dsn string
				var err error

				// CockroachDB (opinionated standard)
				if strings.TrimSpace(os.Getenv("COCKROACH_API_KEY")) != "" {
					fmt.Println("  • CockroachDB: configuring serverless database (Gothic Forge standard)")
					// Use longer context for CockroachDB API operations
					ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Minute)
					defer cancelDB()
					dsn, err = cockroachInteractiveProvision(ctxDB, deployDryRun)
					if err != nil {
						fmt.Println("    → CockroachDB provisioning failed:", err)
					}
				} else {
					// No API keys set - guide user
					fmt.Println("  • Database: No provider configured")
					fmt.Println("    → CockroachDB Serverless (opinionated Gothic Forge standard)")
					fmt.Println("    → Get API key: https://cockroachlabs.cloud/signup")
					fmt.Println("    → Set in .env: COCKROACH_API_KEY=<your-key>")
				}

				// Note: Migrations are now run automatically by the provider (see providers_cockroachdb.go)
				if err != nil && strings.TrimSpace(dsn) == "" {
					fmt.Println("    ⚠️  Database not configured - skipping")
				}
			}
			// Phase 3: Ensure VALKEY_URL (Valkey) — optional with non-interactive/env/flag gating
			if strings.TrimSpace(os.Getenv("VALKEY_URL")) == "" && strings.TrimSpace(os.Getenv("REDIS_URL")) == "" {
				// Create context for interactive operations
				ctxInteractive, cancelInteractive := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancelInteractive()
				
				nonInteractive := boolish(os.Getenv("GFORGE_NONINTERACTIVE"))
				wantValkey := deployWithValkey || boolish(os.Getenv("GFORGE_WITH_VALKEY"))
				if nonInteractive {
					if !wantValkey {
						fmt.Println("    → skipping Valkey setup (non-interactive)")
					} else if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
						fmt.Println("  • Valkey: configuring cache connection (non-interactive)")
						// Use longer context for Aiven API operations
						ctxValkey, cancelValkey := context.WithTimeout(context.Background(), 20*time.Minute)
						defer cancelValkey()
						if vurl, verr := valkeyAutoProvision(ctxValkey, false); verr != nil {
							fmt.Println("    → skipped Valkey provisioning:", verr)
						} else if strings.TrimSpace(vurl) != "" {
							fmt.Println("    → VALKEY_URL configured")
						}
					} else {
						fmt.Println("    → skipping Valkey: non-interactive and AIVEN_TOKEN missing; set VALKEY_URL or provide AIVEN_TOKEN")
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
								ctxValkey, cancelValkey := context.WithTimeout(context.Background(), 20*time.Minute)
								defer cancelValkey()
								vurl, verr = valkeyAutoProvision(ctxValkey, false)
							} else {
								vurl, verr = valkeyInteractiveProvision(ctxInteractive, false)
							}
							if verr != nil {
								fmt.Println("    → skipped Valkey provisioning:", verr)
							} else if strings.TrimSpace(vurl) != "" {
								fmt.Println("    → VALKEY_URL configured")
							}
						}
					} else {
						fmt.Println("  • Valkey: configuring cache connection (pre-approved)")
						var vurl string
						var verr error
						if strings.TrimSpace(os.Getenv("AIVEN_TOKEN")) != "" {
							ctxValkey, cancelValkey := context.WithTimeout(context.Background(), 20*time.Minute)
							defer cancelValkey()
							vurl, verr = valkeyAutoProvision(ctxValkey, false)
						} else {
							vurl, verr = valkeyInteractiveProvision(ctxInteractive, false)
						}
						if verr != nil {
							fmt.Println("    → skipped Valkey provisioning:", verr)
						} else if strings.TrimSpace(vurl) != "" {
							fmt.Println("    → VALKEY_URL configured")
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
				// For Leapcell compute, default to skipping Pages unless explicitly requested
				if !wantPages {
					fmt.Println("    → skipping Cloudflare Pages (Leapcell compute behind Cloudflare proxy; use --with-pages if needed)")
				} else if nonInteractive {
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
			// Leapcell deployment (Opinionated Stack compute provider)
			reader := bufio.NewReader(os.Stdin)

			fmt.Println("  • Deploying to Leapcell (20 free projects!)")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			if err := runLeapcellDeploy(ctx, reader, false); err != nil {
				fmt.Println("  • Leapcell deployment error:", err)
				fmt.Println("────────────────────────────────────────")
				fmt.Println("Fix the issues above and re-run: gforge deploy")
			} else {
				fmt.Println("────────────────────────────────────────")
				fmt.Println("✓ Deployment complete! Your app is live on Leapcell!")
				fmt.Println("  Next steps:")
				fmt.Println("    - Test your deployment at the URL shown above")
				fmt.Println("    - Check logs: gforge logs")
				fmt.Println("    - Monitor health: curl https://your-app.leapcell.dev/readyz")
			}
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
			fmt.Println("  CockroachDB: https://cockroachlabs.cloud/signup")
			fmt.Println("  CockroachDB service accounts: https://cockroachlabs.cloud/service-accounts")
			fmt.Println("  Aiven tokens: https://console.aiven.io/profile/tokens")
			fmt.Println("  Cloudflare API tokens: https://dash.cloudflare.com/profile/api-tokens")
			return nil
		}

		fmt.Println("────────────────────────────────────────")
		// Dry-run provider steps
		// Phase 2b: Show database provisioning plan - CockroachDB (opinionated standard)
		if strings.TrimSpace(os.Getenv("COCKROACH_API_KEY")) != "" {
			fmt.Println("  • CockroachDB (dry-run): would provision serverless cluster")
			_, _ = cockroachInteractiveProvision(context.Background(), true)
		} else {
			fmt.Println("  • Database (dry-run): No provider configured")
			fmt.Println("    → Set COCKROACH_API_KEY for CockroachDB Serverless")
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

		fmt.Println("Deployment flow stub complete. (More integrations to follow)")
		return nil
	},
}

func init() {
	deployCmd.Flags().BoolVar(&deployProd, "prod", false, "use production settings")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "show steps without executing")
	deployCmd.Flags().BoolVar(&deployCheck, "check", false, "preflight checks for tools, tokens, env; no writes or external actions")
	deployCmd.Flags().BoolVar(&deployWithValkey, "with-valkey", false, "configure Valkey/Redis cache during deploy (optional)")
	deployCmd.Flags().BoolVar(&deployWithPages, "with-pages", false, "deploy static export to Cloudflare Pages (optional)")
	rootCmd.AddCommand(deployCmd)
}

// runDeployPreflightCheck validates tools, tokens, .env, and Pages config without modifying state.
func runDeployPreflightCheck() error {
	fmt.Println("Deploy preflight check")

	// Tools
	wrPath, wrOK := execx.Look("wrangler")
	fmt.Printf("  • wrangler: %s\n", pathOrMissing(wrPath, wrOK))
	if wrOK {
		if v, err := execx.RunCapture(context.Background(), "wrangler --version", "wrangler", "--version"); err == nil {
			v = strings.TrimSpace(v)
			if v != "" {
				fmt.Printf("    → %s\n", v)
			}
		}
	}

	// Env presence
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println("  • .env: not found")
	} else {
		fmt.Println("  • .env: present")
	}
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
	if !jwtOK {
		fmt.Println("  • JWT_SECRET: MISSING or weak (>=32 hex chars recommended)")
	}

	// Provider tokens and config
	cockroachTok := strings.TrimSpace(kv["COCKROACH_API_KEY"]) != ""
	aivenTok := strings.TrimSpace(kv["AIVEN_TOKEN"]) != ""
	// Support both CLOUDFLARE_API_TOKEN (wrangler standard) and CF_API_TOKEN (legacy)
	cfTok := strings.TrimSpace(kv["CLOUDFLARE_API_TOKEN"]) != "" || strings.TrimSpace(kv["CF_API_TOKEN"]) != ""
	cfAcct := strings.TrimSpace(kv["CLOUDFLARE_ACCOUNT_ID"]) != "" || strings.TrimSpace(kv["CF_ACCOUNT_ID"]) != ""
	cfProj := strings.TrimSpace(kv["CF_PROJECT_NAME"]) != ""
	dbSet := strings.TrimSpace(kv["DATABASE_URL"]) != ""
	valkeySet := strings.TrimSpace(kv["VALKEY_URL"]) != "" || strings.TrimSpace(kv["REDIS_URL"]) != ""

	// Gating based on env for optional providers
	wantValkey := boolish(os.Getenv("GFORGE_WITH_VALKEY")) || deployWithValkey
	wantPages := boolish(os.Getenv("GFORGE_WITH_PAGES")) || deployWithPages

	// Compute readiness
	ready := true
	missing := []string{}
	// Core env
	if !jwtOK {
		ready = false
		missing = append(missing, "JWT_SECRET (strong)")
	}
	// DB - CockroachDB only
	if !(dbSet || cockroachTok) {
		ready = false
		missing = append(missing, "DATABASE_URL or COCKROACH_API_KEY")
	}
	// Valkey (optional)
	if wantValkey {
		if !(valkeySet || aivenTok) {
			ready = false
			missing = append(missing, "VALKEY_URL or AIVEN_TOKEN")
		}
	}
	// Pages (optional)
	if wantPages {
		if !(cfTok && cfAcct && cfProj) {
			ready = false
			missing = append(missing, "CLOUDFLARE_API_TOKEN, CLOUDFLARE_ACCOUNT_ID, CF_PROJECT_NAME")
		}
		// Note: wrangler missing does not fail readiness (you can deploy via GitHub Action), but we warn
		if !wrOK {
			fmt.Println("  • Warning: wrangler missing; local pages deploy will be skipped (CI Pages action still works)")
		}
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
		for _, m := range missing {
			fmt.Println("    - ", m)
		}
	}
	return fmt.Errorf("preflight failed")
}

func presentOrMissing(b bool) string {
	if b {
		return "present"
	}
	return "missing"
}

// interactiveEnvSetup ensures .env exists (copying from .env.example if present),
// prompts for missing values, and writes them back.
func interactiveEnvSetup() error {
	envPath := ".env"
	examplePath := ".env.example"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if b, err2 := os.ReadFile(examplePath); err2 == nil {
			if err3 := os.WriteFile(envPath, b, 0o600); err3 != nil {
				return err3
			}
			fmt.Println("  • Created .env from .env.example")
		} else {
			// Create minimal .env if no example
			if err3 := os.WriteFile(envPath, []byte("APP_ENV=production\n"), 0o600); err3 != nil {
				return err3
			}
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
	// 3) Provider tokens with links shown inline (Opinionated Stack only)
	type tok struct{ key, label, link string }
	tokens := []tok{
		{"COCKROACH_API_KEY", "CockroachDB service account", "https://cockroachlabs.cloud/service-accounts"},
		{"AIVEN_TOKEN", "Aiven tokens", "https://console.aiven.io/profile/tokens"},
		{"CLOUDFLARE_API_TOKEN", "Cloudflare API tokens", "https://dash.cloudflare.com/profile/api-tokens"},
	}
	for _, t := range tokens {
		if strings.TrimSpace(kv[t.key]) != "" {
			continue
		}
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
	if err := updateEnvFileInPlace(envPath, kv); err != nil {
		return err
	}
	fmt.Println("  • Wrote .env with updated values (preserved existing layout)")
	return nil
}

func loadEnvFile(path string) map[string]string {
	kv := map[string]string{}
	b, err := os.ReadFile(path)
	if err != nil {
		return kv
	}
	lines := strings.Split(string(b), "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
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
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}



// fileStartsWithWizardHeader detects if the .env was auto-generated by a previous wizard run.
func fileStartsWithWizardHeader(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	r := bufio.NewReader(f)
	line, _ := r.ReadString('\n')
	return strings.HasPrefix(strings.TrimSpace(line), "# Generated by gforge deploy wizard")
}

// rewriteEnvFromExample rewrites envPath using examplePath's structure, substituting
// values from kv where keys match. Comments and blank lines are preserved from example.
func rewriteEnvFromExample(envPath, examplePath string, kv map[string]string) error {
	b, err := os.ReadFile(examplePath)
	if err != nil {
		return err
	}
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
	if v == "" {
		return v
	}
	if !(strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")) {
		// default to https for production URLs
		v = "https://" + v
	}
	if v != "/" {
		v = strings.TrimRight(v, "/")
	}
	return v
}

// updateEnvFileInPlace updates only values for existing keys in .env, preserving
// its current structure and comments. Any missing keys are appended at the end.
func updateEnvFileInPlace(envPath string, kv map[string]string) error {
	b, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	pending := map[string]string{}
	for k, v := range kv {
		pending[k] = v
	}
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
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
