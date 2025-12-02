package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gothicforge3/internal/providers"

	"github.com/spf13/cobra"
)

var (
	deployLiveSkipTests bool
)

var deployLiveCmd = &cobra.Command{
	Use:   "deploy-live",
	Short: "Deploy to production in under 3 minutes (Opinionated Stack)",
	Long: `Deploy your Gothic Forge application to production in under 3 minutes.

This command orchestrates the complete deployment flow:
  • Phase 1: Pre-flight checks (5s)
  • Phase 2: Provision infrastructure (60s)
  • Phase 3: Deploy application (45s)
  • Phase 4: Health checks (10s)

Target: Complete in under 3 minutes (180 seconds)

The Opinionated Stack:
  • Database: CockroachDB Serverless
  • Cache: Aiven Valkey
  • Compute: Leapcell
  • CDN: Cloudflare Pages (optional)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !QuietMode {
			banner()
			fmt.Println("Deploy Live - Opinionated Stack")
			fmt.Println("────────────────────────────────────────")
		}

		ctx := context.Background()
		startTime := time.Now()

		// Phase 1: Pre-flight checks (5s)
		if err := runPreflightChecks(ctx); err != nil {
			return fmt.Errorf("pre-flight checks failed: %w", err)
		}

		// Phase 2: Provision infrastructure (60s)
		infra, err := provisionInfrastructure(ctx)
		if err != nil {
			return fmt.Errorf("infrastructure provisioning failed: %w", err)
		}

		// Phase 3: Deploy application (45s)
		deployment, err := deployApplication(ctx, infra)
		if err != nil {
			// Rollback on failure
			if rollbackErr := rollbackInfrastructure(ctx, infra); rollbackErr != nil {
				fmt.Printf("⚠️  Rollback failed: %v\n", rollbackErr)
			}
			return fmt.Errorf("deployment failed: %w", err)
		}

		// Phase 4: Health checks (10s)
		if err := waitForHealthy(ctx, deployment); err != nil {
			// Rollback on health check failure
			if rollbackErr := rollbackDeployment(ctx, deployment); rollbackErr != nil {
				fmt.Printf("⚠️  Rollback failed: %v\n", rollbackErr)
			}
			return fmt.Errorf("health checks failed: %w", err)
		}

		// Success!
		duration := time.Since(startTime)
		printDeploymentSuccess(deployment, duration)

		return nil
	},
}

func init() {
	deployLiveCmd.Flags().BoolVar(&deployLiveSkipTests, "skip-tests", false, "skip running tests before deployment")
	rootCmd.AddCommand(deployLiveCmd)
}

// Infrastructure holds provisioned infrastructure resources
type Infrastructure struct {
	Database *providers.DatabaseInfo
	Cache    *providers.CacheInfo
	CDN      *providers.CDNInfo
}

// runPreflightChecks validates environment and configuration before deployment
func runPreflightChecks(ctx context.Context) error {
	if !QuietMode {
		fmt.Println("\n📋 Phase 1: Pre-flight Checks (5s)")
		fmt.Println("────────────────────────────────────────")
	}

	// Check 1: Validate environment variables
	if !QuietMode {
		fmt.Println("  → Validating environment variables...")
	}
	
	cfg, err := LoadDeployConfig()
	if err != nil {
		return &DeploymentError{
			Phase: "Pre-flight Checks",
			Cause: err,
			Actions: []string{
				"Ensure .env file exists with required variables",
				"Run: gforge init (to generate configuration)",
				"Check: JWT_SECRET is set and >= 32 characters",
				"Check: COCKROACH_API_KEY or DATABASE_URL is set",
			},
		}
	}

	if !QuietMode {
		fmt.Println("  ✓ Environment variables validated")
	}

	// Check 2: Verify provider credentials
	if !QuietMode {
		fmt.Println("  → Checking provider credentials...")
	}

	if cfg.CockroachAPIKey == "" && cfg.DatabaseURL == "" {
		return &DeploymentError{
			Phase: "Pre-flight Checks",
			Cause: fmt.Errorf("no database provider configured"),
			Actions: []string{
				"Set COCKROACH_API_KEY for CockroachDB Serverless",
				"Or set DATABASE_URL for existing database",
				"Get API key: https://cockroachlabs.cloud/service-accounts",
			},
		}
	}

	if !QuietMode {
		fmt.Println("  ✓ Provider credentials validated")
	}

	// Check 3: Build project
	if !QuietMode {
		fmt.Println("  → Building project...")
	}

	if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
		return &DeploymentError{
			Phase: "Pre-flight Checks",
			Cause: fmt.Errorf("project build failed: %w", err),
			Actions: []string{
				"Fix build errors in your code",
				"Run: gforge build (to see detailed errors)",
				"Run: gforge doctor (to check for missing tools)",
			},
		}
	}

	if !QuietMode {
		fmt.Println("  ✓ Project built successfully")
	}

	// Check 4: Run tests (optional)
	if !deployLiveSkipTests {
		if !QuietMode {
			fmt.Println("  → Running tests...")
		}

		testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := testCmd.RunE(testCmd, []string{}); err != nil {
			if !QuietMode {
				fmt.Printf("  ⚠️  Tests failed: %v\n", err)
				fmt.Println("  → Use --skip-tests to deploy anyway")
			}
			return &DeploymentError{
				Phase: "Pre-flight Checks",
				Cause: fmt.Errorf("tests failed: %w", err),
				Actions: []string{
					"Fix failing tests",
					"Run: gforge test (to see detailed test output)",
					"Or use: gforge deploy-live --skip-tests (not recommended)",
				},
			}
		}

		_ = testCtx // Use context

		if !QuietMode {
			fmt.Println("  ✓ Tests passed")
		}
	}

	// Check 5: Display deployment plan
	if !QuietMode {
		fmt.Println("\n  📝 Deployment Plan:")
		fmt.Println("    • Database: CockroachDB Serverless")
		if os.Getenv("AIVEN_TOKEN") != "" {
			fmt.Println("    • Cache: Aiven Valkey")
		}
		fmt.Println("    • Compute: Leapcell")
		if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
			fmt.Println("    • CDN: Cloudflare Pages")
		}
	}

	return nil
}

// provisionInfrastructure provisions all required infrastructure in parallel
func provisionInfrastructure(ctx context.Context) (*Infrastructure, error) {
	if !QuietMode {
		fmt.Println("\n🏗️  Phase 2: Provision Infrastructure (60s)")
		fmt.Println("────────────────────────────────────────")
	}

	infra := &Infrastructure{}
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Provision database (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if !QuietMode {
			fmt.Println("  → Provisioning database...")
		}

		db, err := provisionDatabase(ctx)
		if err != nil {
			errChan <- fmt.Errorf("database provisioning failed: %w", err)
			return
		}
		infra.Database = db

		if !QuietMode {
			fmt.Printf("  ✓ Database: %s\n", db.Name)
		}
	}()

	// Provision cache (parallel, optional)
	if os.Getenv("AIVEN_TOKEN") != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !QuietMode {
				fmt.Println("  → Provisioning cache...")
			}

			cache, err := provisionCache(ctx)
			if err != nil {
				// Cache is optional, log warning but don't fail
				if !QuietMode {
					fmt.Printf("  ⚠️  Cache provisioning failed: %v\n", err)
				}
				return
			}
			infra.Cache = cache

			if !QuietMode {
				fmt.Printf("  ✓ Cache: %s\n", cache.Name)
			}
		}()
	}

	// Provision CDN (parallel, optional)
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !QuietMode {
				fmt.Println("  → Provisioning CDN...")
			}

			cdn, err := provisionCDN(ctx)
			if err != nil {
				// CDN is optional, log warning but don't fail
				if !QuietMode {
					fmt.Printf("  ⚠️  CDN provisioning failed: %v\n", err)
				}
				return
			}
			infra.CDN = cdn

			if !QuietMode {
				fmt.Printf("  ✓ CDN: %s\n", cdn.Name)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Run database migrations
	if infra.Database != nil {
		if !QuietMode {
			fmt.Println("  → Running database migrations...")
		}

		dbProvider, err := providers.DefaultRegistry.GetDatabase("cockroachdb")
		if err != nil {
			return nil, fmt.Errorf("failed to get database provider: %w", err)
		}

		migrationsDir := "app/db/migrations"
		if err := dbProvider.RunMigrations(ctx, migrationsDir); err != nil {
			return nil, &DeploymentError{
				Phase: "Infrastructure Provisioning",
				Cause: fmt.Errorf("migrations failed: %w", err),
				Actions: []string{
					"Check migration files in app/db/migrations/",
					"Verify database connection string is correct",
					"Run migrations manually: goose -dir app/db/migrations postgres $DATABASE_URL up",
				},
			}
		}

		if !QuietMode {
			fmt.Println("  ✓ Migrations applied")
		}
	}

	return infra, nil
}

// provisionDatabase provisions the database using the configured provider
func provisionDatabase(ctx context.Context) (*providers.DatabaseInfo, error) {
	// Check if DATABASE_URL is already set
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// Use existing database
		return &providers.DatabaseInfo{
			Name:             "existing-database",
			ConnectionString: dbURL,
		}, nil
	}

	// Provision new database using CockroachDB
	dbProvider, err := providers.DefaultRegistry.GetDatabase("cockroachdb")
	if err != nil {
		return nil, fmt.Errorf("failed to get database provider: %w", err)
	}

	opts := providers.ProvisionOptions{
		Name:   "gothic-forge-prod",
		Region: "us-east-1",
		Tier:   "serverless",
	}

	db, err := dbProvider.Provision(ctx, opts)
	if err != nil {
		return nil, &DeploymentError{
			Phase: "Database Provisioning",
			Cause: err,
			Actions: []string{
				"Check your COCKROACH_API_KEY is valid",
				"Verify your account has sufficient quota",
				"Get help: https://cockroachlabs.cloud/docs",
			},
		}
	}

	// Update .env with DATABASE_URL
	if err := updateEnvVar("DATABASE_URL", db.ConnectionString); err != nil {
		fmt.Printf("⚠️  Could not update .env: %v\n", err)
	}

	return db, nil
}

// provisionCache provisions the cache using the configured provider
func provisionCache(ctx context.Context) (*providers.CacheInfo, error) {
	// Check if VALKEY_URL is already set
	if cacheURL := os.Getenv("VALKEY_URL"); cacheURL != "" {
		// Use existing cache
		return &providers.CacheInfo{
			Name:             "existing-cache",
			ConnectionString: cacheURL,
		}, nil
	}

	// Provision new cache using Valkey
	cacheProvider, err := providers.DefaultRegistry.GetCache("valkey")
	if err != nil {
		return nil, fmt.Errorf("failed to get cache provider: %w", err)
	}

	opts := providers.ProvisionOptions{
		Name:   "gothic-forge-cache",
		Region: "us-east-1",
		Tier:   "startup",
	}

	cache, err := cacheProvider.Provision(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Update .env with VALKEY_URL
	if err := updateEnvVar("VALKEY_URL", cache.ConnectionString); err != nil {
		fmt.Printf("⚠️  Could not update .env: %v\n", err)
	}

	return cache, nil
}

// provisionCDN provisions the CDN using the configured provider
func provisionCDN(ctx context.Context) (*providers.CDNInfo, error) {
	cdnProvider, err := providers.DefaultRegistry.GetCDN("cloudflare")
	if err != nil {
		return nil, fmt.Errorf("failed to get CDN provider: %w", err)
	}

	opts := providers.CDNDeployOptions{
		ProjectName: os.Getenv("CF_PROJECT_NAME"),
		Directory:   "dist",
		Branch:      "main",
	}

	cdn, err := cdnProvider.Deploy(ctx, opts)
	if err != nil {
		return nil, err
	}

	return cdn, nil
}

// deployApplication deploys the application to the compute provider
func deployApplication(ctx context.Context, infra *Infrastructure) (*providers.DeploymentInfo, error) {
	if !QuietMode {
		fmt.Println("\n🚀 Phase 3: Deploy Application (45s)")
		fmt.Println("────────────────────────────────────────")
	}

	// Get compute provider
	computeProvider, err := providers.DefaultRegistry.GetCompute("leapcell")
	if err != nil {
		return nil, &DeploymentError{
			Phase: "Application Deployment",
			Cause: fmt.Errorf("failed to get compute provider: %w", err),
			Actions: []string{
				"Ensure Leapcell provider is configured",
				"Set LEAPCELL_API_KEY in .env",
				"Get API key: https://leapcell.io/dashboard",
			},
		}
	}

	// Prepare environment variables
	envVars := map[string]string{
		"APP_ENV":       "production",
		"SITE_BASE_URL": os.Getenv("SITE_BASE_URL"),
		"JWT_SECRET":    os.Getenv("JWT_SECRET"),
	}

	// Add database connection
	if infra.Database != nil {
		envVars["DATABASE_URL"] = infra.Database.ConnectionString
	}

	// Add cache connection (optional)
	if infra.Cache != nil {
		envVars["VALKEY_URL"] = infra.Cache.ConnectionString
	}

	// Deploy
	if !QuietMode {
		fmt.Println("  → Building and deploying to Leapcell...")
	}

	opts := providers.DeployOptions{
		Name:        "gothic-forge-prod",
		BuildPath:   ".",
		EnvVars:     envVars,
		Replicas:    1,
		HealthCheck: "/healthz",
		Port:        8080,
	}

	deployment, err := computeProvider.Deploy(ctx, opts)
	if err != nil {
		return nil, &DeploymentError{
			Phase: "Application Deployment",
			Cause: err,
			Actions: []string{
				"Check Leapcell deployment logs",
				"Verify your application builds correctly",
				"Run: gforge build (to test locally)",
			},
		}
	}

	if !QuietMode {
		fmt.Printf("  ✓ Deployment: %s\n", deployment.Name)
		fmt.Printf("  ✓ URL: %s\n", deployment.URL)
	}

	return deployment, nil
}

// waitForHealthy waits for the deployment to become healthy
func waitForHealthy(ctx context.Context, deployment *providers.DeploymentInfo) error {
	if !QuietMode {
		fmt.Println("\n🏥 Phase 4: Health Checks (10s)")
		fmt.Println("────────────────────────────────────────")
	}

	// Get compute provider
	computeProvider, err := providers.DefaultRegistry.GetCompute("leapcell")
	if err != nil {
		return fmt.Errorf("failed to get compute provider: %w", err)
	}

	// Wait for health checks with timeout
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if !QuietMode {
		fmt.Println("  → Waiting for health checks...")
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-healthCtx.Done():
			return &DeploymentError{
				Phase: "Health Checks",
				Cause: fmt.Errorf("health checks timed out after 2 minutes"),
				Actions: []string{
					"Check application logs: gforge logs",
					"Verify /healthz endpoint is implemented",
					"Check for startup errors in your application",
				},
			}
		case <-ticker.C:
			if err := computeProvider.Health(healthCtx); err == nil {
				if !QuietMode {
					fmt.Println("  ✓ /healthz: OK")
					fmt.Println("  ✓ Application is healthy")
				}
				return nil
			}
		}
	}
}

// rollbackInfrastructure cleans up provisioned infrastructure on failure
func rollbackInfrastructure(ctx context.Context, infra *Infrastructure) error {
	if !QuietMode {
		fmt.Println("\n⏪ Rolling back infrastructure...")
	}

	var errors []error

	// Destroy database
	if infra.Database != nil {
		if dbProvider, err := providers.DefaultRegistry.GetDatabase("cockroachdb"); err == nil {
			if err := dbProvider.Destroy(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to destroy database: %w", err))
			}
		}
	}

	// Destroy cache
	if infra.Cache != nil {
		if cacheProvider, err := providers.DefaultRegistry.GetCache("valkey"); err == nil {
			if err := cacheProvider.Destroy(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to destroy cache: %w", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("rollback had errors: %v", errors)
	}

	if !QuietMode {
		fmt.Println("  ✓ Infrastructure rolled back")
	}

	return nil
}

// rollbackDeployment reverts to the previous deployment
func rollbackDeployment(ctx context.Context, deployment *providers.DeploymentInfo) error {
	if !QuietMode {
		fmt.Println("\n⏪ Rolling back deployment...")
	}

	computeProvider, err := providers.DefaultRegistry.GetCompute("leapcell")
	if err != nil {
		return fmt.Errorf("failed to get compute provider: %w", err)
	}

	// Rollback to previous version
	if err := computeProvider.Rollback(ctx, "previous"); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if !QuietMode {
		fmt.Println("  ✓ Deployment rolled back to previous version")
	}

	return nil
}

// printDeploymentSuccess prints the success message with deployment details
func printDeploymentSuccess(deployment *providers.DeploymentInfo, duration time.Duration) {
	fmt.Println("\n────────────────────────────────────────")
	fmt.Printf("🎉 Deployment Successful! (%v)\n", duration.Round(time.Second))
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("\n🌐 Production URL: %s\n", deployment.URL)
	if deployment.DashboardURL != "" {
		fmt.Printf("📊 Dashboard: %s\n", deployment.DashboardURL)
	}
	fmt.Println("\nNext steps:")
	fmt.Printf("  • Test your app: curl %s/healthz\n", deployment.URL)
	fmt.Println("  • View logs: gforge logs --follow")
	fmt.Println("  • Scale up: gforge scale --replicas=3")
	fmt.Println("  • Monitor: gforge status")
	fmt.Println()
}

// DeploymentError represents a deployment error with actionable suggestions
type DeploymentError struct {
	Phase   string
	Cause   error
	Actions []string
}

func (e *DeploymentError) Error() string {
	msg := fmt.Sprintf("❌ Deployment failed in %s phase\n", e.Phase)
	msg += fmt.Sprintf("   Cause: %s\n\n", e.Cause)
	msg += "   Suggested actions:\n"
	for i, action := range e.Actions {
		msg += fmt.Sprintf("   %d. %s\n", i+1, action)
	}
	return msg
}

// updateEnvVar updates a single environment variable in .env file
func updateEnvVar(key, value string) error {
	envPath := ".env"
	kv := loadEnvFile(envPath)
	kv[key] = value
	return updateEnvFileInPlace(envPath, kv)
}
