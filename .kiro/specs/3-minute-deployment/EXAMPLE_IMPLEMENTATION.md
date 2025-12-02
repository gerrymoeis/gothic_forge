# Gothic Forge v10.0 - Implementation Examples

This document shows concrete code examples of how the new features will work.

---

## 1. Provider Abstraction Layer

### Example: Database Provider Interface

```go
// internal/providers/interfaces.go
package providers

import (
    "context"
    "io"
    "time"
)

// DatabaseProvider abstracts database provisioning and management
type DatabaseProvider interface {
    // Name returns the provider name (e.g., "cockroachdb", "postgresql")
    Name() string
    
    // Provision creates a new database instance
    Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error)
    
    // GetConnectionString returns the connection string
    GetConnectionString(ctx context.Context) (string, error)
    
    // RunMigrations executes pending database migrations
    RunMigrations(ctx context.Context, migrationsDir string) error
    
    // Health checks if the database is reachable
    Health(ctx context.Context) error
    
    // Destroy removes the database instance
    Destroy(ctx context.Context) error
}

// ProvisionOptions contains options for provisioning
type ProvisionOptions struct {
    Name     string
    Region   string
    Tier     string
    Tags     map[string]string
    Metadata map[string]interface{}
}

// DatabaseInfo contains information about a provisioned database
type DatabaseInfo struct {
    ID               string
    Name             string
    ConnectionString string
    Host             string
    Port             int
    Database         string
    Username         string
    Password         string
    Region           string
    CreatedAt        time.Time
    Metadata         map[string]interface{}
}
```

### Example: CockroachDB Provider Implementation

```go
// internal/providers/cockroachdb.go
package providers

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cockroachdb/cockroach-cloud-sdk-go/pkg/client"
)

type CockroachDBProvider struct {
    apiKey string
    client *client.Client
}

func NewCockroachDBProvider(apiKey string) *CockroachDBProvider {
    return &CockroachDBProvider{
        apiKey: apiKey,
        client: client.NewClient(apiKey),
    }
}

func (p *CockroachDBProvider) Name() string {
    return "cockroachdb"
}

func (p *CockroachDBProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
    // Create serverless cluster
    cluster, err := p.client.CreateCluster(ctx, &client.CreateClusterRequest{
        Name:   opts.Name,
        Region: opts.Region,
        Plan:   "serverless",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create cluster: %w", err)
    }
    
    // Wait for cluster to be ready
    if err := p.waitForReady(ctx, cluster.ID); err != nil {
        return nil, fmt.Errorf("cluster not ready: %w", err)
    }
    
    // Create database and user
    db, err := p.client.CreateDatabase(ctx, cluster.ID, opts.Name)
    if err != nil {
        return nil, fmt.Errorf("failed to create database: %w", err)
    }
    
    user, err := p.client.CreateUser(ctx, cluster.ID, opts.Name+"_user")
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Build connection string
    connStr := fmt.Sprintf(
        "postgresql://%s:%s@%s:%d/%s?sslmode=verify-full",
        user.Name, user.Password, cluster.Host, cluster.Port, db.Name,
    )
    
    return &DatabaseInfo{
        ID:               cluster.ID,
        Name:             cluster.Name,
        ConnectionString: connStr,
        Host:             cluster.Host,
        Port:             cluster.Port,
        Database:         db.Name,
        Username:         user.Name,
        Password:         user.Password,
        Region:           cluster.Region,
        CreatedAt:        cluster.CreatedAt,
    }, nil
}

func (p *CockroachDBProvider) GetConnectionString(ctx context.Context) (string, error) {
    // Retrieve from environment or saved state
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        return "", fmt.Errorf("DATABASE_URL not set")
    }
    return connStr, nil
}

func (p *CockroachDBProvider) RunMigrations(ctx context.Context, migrationsDir string) error {
    connStr, err := p.GetConnectionString(ctx)
    if err != nil {
        return err
    }
    
    // Use goose for migrations
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    defer db.Close()
    
    return goose.Up(db, migrationsDir)
}

func (p *CockroachDBProvider) Health(ctx context.Context) error {
    connStr, err := p.GetConnectionString(ctx)
    if err != nil {
        return err
    }
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    defer db.Close()
    
    return db.PingContext(ctx)
}

func (p *CockroachDBProvider) Destroy(ctx context.Context) error {
    // Implementation for cleanup
    return nil
}

func (p *CockroachDBProvider) waitForReady(ctx context.Context, clusterID string) error {
    timeout := time.After(5 * time.Minute)
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-timeout:
            return fmt.Errorf("timeout waiting for cluster to be ready")
        case <-ticker.C:
            cluster, err := p.client.GetCluster(ctx, clusterID)
            if err != nil {
                return err
            }
            if cluster.State == "READY" {
                return nil
            }
        }
    }
}
```

---

## 2. Provider Registry

```go
// internal/providers/registry.go
package providers

import (
    "fmt"
    "sync"
)

type Registry struct {
    databases map[string]DatabaseProvider
    caches    map[string]CacheProvider
    computes  map[string]ComputeProvider
    cdns      map[string]CDNProvider
    mu        sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        databases: make(map[string]DatabaseProvider),
        caches:    make(map[string]CacheProvider),
        computes:  make(map[string]ComputeProvider),
        cdns:      make(map[string]CDNProvider),
    }
}

func (r *Registry) RegisterDatabase(name string, provider DatabaseProvider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.databases[name] = provider
}

func (r *Registry) GetDatabase(name string) (DatabaseProvider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    provider, ok := r.databases[name]
    if !ok {
        return nil, fmt.Errorf("database provider %q not found", name)
    }
    return provider, nil
}

func (r *Registry) ListDatabases() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    names := make([]string, 0, len(r.databases))
    for name := range r.databases {
        names = append(names, name)
    }
    return names
}

// Global registry
var DefaultRegistry = NewRegistry()

// Register providers at init time
func init() {
    // Database providers
    DefaultRegistry.RegisterDatabase("cockroachdb", NewCockroachDBProvider(os.Getenv("COCKROACH_API_KEY")))
    DefaultRegistry.RegisterDatabase("postgresql", NewPostgreSQLProvider())
    DefaultRegistry.RegisterDatabase("sqlite", NewSQLiteProvider())
    
    // Cache providers
    DefaultRegistry.RegisterCache("valkey", NewValkeyProvider(os.Getenv("AIVEN_TOKEN")))
    DefaultRegistry.RegisterCache("redis", NewRedisProvider())
    
    // Compute providers
    DefaultRegistry.RegisterCompute("leapcell", NewLeapcellProvider(os.Getenv("LEAPCELL_APP_URL")))
    DefaultRegistry.RegisterCompute("docker", NewDockerProvider())
    
    // CDN providers
    DefaultRegistry.RegisterCDN("cloudflare", NewCloudflareProvider(os.Getenv("CLOUDFLARE_API_TOKEN")))
}
```

---

## 3. `gforge init` Command

```go
// cmd/gforge/cmd/init.go
package cmd

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/spf13/cobra"
    "gothicforge3/internal/ui"
)

var (
    initInteractive bool
    initTemplate    string
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new Gothic Forge project in under 30 seconds",
    Long: `Initialize a new Gothic Forge project with:
  • Secure configuration generation
  • Automatic dependency installation
  • Project verification
  • Ready to deploy in under 30 seconds`,
    RunE: runInit,
}

func init() {
    initCmd.Flags().BoolVar(&initInteractive, "interactive", false, "interactive setup with prompts")
    initCmd.Flags().StringVar(&initTemplate, "template", "minimal", "project template (minimal, blog, saas, api)")
    rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    startTime := time.Now()
    
    fmt.Println("🚀 Gothic Forge v10.0 - Project Initialization")
    fmt.Println()
    
    // Step 1: Detect environment
    spinner := ui.NewSpinner("Detecting environment...")
    spinner.Start()
    
    projectName := detectProjectName()
    gitRemote := detectGitRemote()
    
    spinner.Success(fmt.Sprintf("Environment detected (project: %s)", projectName))
    
    // Step 2: Generate configuration
    spinner = ui.NewSpinner("Generating configuration...")
    spinner.Start()
    
    if err := generateEnvFile(projectName, gitRemote); err != nil {
        spinner.Fail("Failed to generate .env")
        return err
    }
    
    if err := generateProvidersConfig(); err != nil {
        spinner.Fail("Failed to generate providers.yaml")
        return err
    }
    
    spinner.Success("Configuration generated")
    
    // Step 3: Install dependencies
    spinner = ui.NewSpinner("Installing dependencies...")
    spinner.Start()
    
    if err := installDependencies(ctx); err != nil {
        spinner.Fail("Failed to install dependencies")
        return err
    }
    
    spinner.Success("Dependencies installed")
    
    // Step 4: Verify setup
    spinner = ui.NewSpinner("Verifying setup...")
    spinner.Start()
    
    if err := verifySetup(ctx); err != nil {
        spinner.Fail("Setup verification failed")
        return err
    }
    
    spinner.Success("Setup verified")
    
    // Success message
    duration := time.Since(startTime)
    printSuccessMessage(projectName, duration)
    
    return nil
}

func detectProjectName() string {
    // Try Git remote first
    if remote := detectGitRemote(); remote != "" {
        parts := strings.Split(remote, "/")
        if len(parts) > 0 {
            name := parts[len(parts)-1]
            name = strings.TrimSuffix(name, ".git")
            return name
        }
    }
    
    // Fall back to directory name
    wd, _ := os.Getwd()
    return filepath.Base(wd)
}

func detectGitRemote() string {
    cmd := exec.Command("git", "config", "--get", "remote.origin.url")
    output, err := cmd.Output()
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(output))
}

func generateEnvFile(projectName, gitRemote string) error {
    // Generate secure JWT secret
    jwtSecret := generateSecret(64)
    
    // Derive site base URL from Git remote
    siteBaseURL := deriveSiteBaseURL(gitRemote, projectName)
    
    // Create .env content
    envContent := fmt.Sprintf(`# Gothic Forge v10.0 - Environment Configuration
# Generated by: gforge init
# Generated at: %s

# Application
APP_ENV=development
HTTP_HOST=127.0.0.1
HTTP_PORT=8080
SITE_BASE_URL=%s

# Security
JWT_SECRET=%s

# Database (CockroachDB Serverless - Opinionated Stack)
# Get API key: https://cockroachlabs.cloud/service-accounts
COCKROACH_API_KEY=
DATABASE_URL=

# Cache (Aiven Valkey - Opinionated Stack)
# Get token: https://console.aiven.io/profile/tokens
AIVEN_TOKEN=
VALKEY_URL=

# Compute (Leapcell - Opinionated Stack)
# Get app URL: https://leapcell.io/dashboard
LEAPCELL_APP_URL=

# CDN (Cloudflare - Opinionated Stack)
# Get token: https://dash.cloudflare.com/profile/api-tokens
CLOUDFLARE_API_TOKEN=
CLOUDFLARE_ACCOUNT_ID=
CF_PROJECT_NAME=%s

# Logging
LOG_FORMAT=
CORS_ORIGINS=

# SEO
SEO_KEYWORDS=
`,
        time.Now().Format(time.RFC3339),
        siteBaseURL,
        jwtSecret,
        projectName,
    )
    
    return os.WriteFile(".env", []byte(envContent), 0600)
}

func generateProvidersConfig() error {
    providersContent := `# Gothic Forge v10.0 - Provider Configuration
# Customize which providers to use for each service

database:
  provider: cockroachdb  # cockroachdb, postgresql, sqlite
  options:
    region: us-east-1
    tier: serverless

cache:
  provider: valkey  # valkey, redis, none
  options:
    region: us-east-1
    tier: startup

compute:
  provider: leapcell  # leapcell, docker, aws-ecs
  options:
    region: global
    replicas: 1

cdn:
  provider: cloudflare  # cloudflare, none
  options:
    zone: auto
`
    
    return os.WriteFile("providers.yaml", []byte(providersContent), 0644)
}

func installDependencies(ctx context.Context) error {
    // Run go mod download
    cmd := exec.CommandContext(ctx, "go", "mod", "download")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("go mod download failed: %w", err)
    }
    
    // Install templ if missing
    if _, err := exec.LookPath("templ"); err != nil {
        cmd := exec.CommandContext(ctx, "go", "install", "github.com/a-h/templ/cmd/templ@latest")
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("templ install failed: %w", err)
        }
    }
    
    // Install gotailwindcss if missing
    if _, err := exec.LookPath("gotailwindcss"); err != nil {
        cmd := exec.CommandContext(ctx, "go", "install", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest")
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("gotailwindcss install failed: %w", err)
        }
    }
    
    // Generate Templ templates
    cmd = exec.CommandContext(ctx, "templ", "generate")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("templ generate failed: %w", err)
    }
    
    // Build Tailwind CSS
    cmd = exec.CommandContext(ctx, "gotailwindcss", "-i", "app/static/tailwind.input.css", "-o", "app/styles/output.css")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("tailwind build failed: %w", err)
    }
    
    return nil
}

func verifySetup(ctx context.Context) error {
    // Run gforge doctor
    cmd := exec.CommandContext(ctx, "go", "run", "./cmd/gforge", "doctor")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("doctor check failed: %w", err)
    }
    
    // Build project
    cmd = exec.CommandContext(ctx, "go", "build", "-o", "server", "./cmd/server")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }
    
    return nil
}

func generateSecret(length int) string {
    bytes := make([]byte, length/2)
    if _, err := rand.Read(bytes); err != nil {
        panic(err)
    }
    return hex.EncodeToString(bytes)
}

func deriveSiteBaseURL(gitRemote, projectName string) string {
    if gitRemote == "" {
        return "http://127.0.0.1:8080"
    }
    
    // Try to derive from GitHub/GitLab URL
    if strings.Contains(gitRemote, "github.com") {
        // Extract username/repo
        parts := strings.Split(gitRemote, "/")
        if len(parts) >= 2 {
            username := parts[len(parts)-2]
            return fmt.Sprintf("https://%s.github.io/%s", username, projectName)
        }
    }
    
    return "http://127.0.0.1:8080"
}

func printSuccessMessage(projectName string, duration time.Duration) {
    fmt.Println()
    fmt.Println("✅ Project initialized successfully!")
    fmt.Printf("   Duration: %v\n", duration.Round(time.Millisecond))
    fmt.Println()
    fmt.Println("📁 Project:", projectName)
    fmt.Println("📝 Configuration: .env, providers.yaml")
    fmt.Println("🔧 Dependencies: installed")
    fmt.Println("✓ Verification: passed")
    fmt.Println()
    fmt.Println("🚀 Next steps:")
    fmt.Println("   1. Start development: gforge dev")
    fmt.Println("   2. Deploy to production: gforge deploy --live")
    fmt.Println()
    fmt.Println("📚 Documentation: https://docs.gothicforge.dev")
    fmt.Println("💬 Community: https://discord.gg/gothicforge")
}
```

---

## 4. `gforge deploy --live` Command

```go
// cmd/gforge/cmd/deploy_live.go
package cmd

import (
    "context"
    "fmt"
    "time"
    
    "github.com/spf13/cobra"
    "gothicforge3/internal/providers"
    "gothicforge3/internal/ui"
)

var deployLiveCmd = &cobra.Command{
    Use:   "deploy --live",
    Short: "Deploy to production in under 2 minutes",
    Long: `Fully automated deployment with:
  • Pre-flight checks
  • Infrastructure provisioning
  • Application deployment
  • Health checks
  • Automatic rollback on failure`,
    RunE: runDeployLive,
}

func init() {
    deployCmd.Flags().Bool("live", false, "deploy to production (automated)")
    rootCmd.AddCommand(deployCmd)
}

func runDeployLive(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    startTime := time.Now()
    
    fmt.Println("🚀 Gothic Forge v10.0 - Production Deployment")
    fmt.Println()
    
    // Phase 1: Pre-flight checks
    fmt.Println("Phase 1: Pre-flight Checks")
    if err := runPreflightChecks(ctx); err != nil {
        return fmt.Errorf("pre-flight checks failed: %w", err)
    }
    fmt.Println("✓ Pre-flight checks passed")
    fmt.Println()
    
    // Phase 2: Provision infrastructure
    fmt.Println("Phase 2: Infrastructure Provisioning")
    infra, err := provisionInfrastructure(ctx)
    if err != nil {
        return fmt.Errorf("infrastructure provisioning failed: %w", err)
    }
    fmt.Println("✓ Infrastructure provisioned")
    fmt.Println()
    
    // Phase 3: Deploy application
    fmt.Println("Phase 3: Application Deployment")
    deployment, err := deployApplication(ctx, infra)
    if err != nil {
        // Rollback on failure
        fmt.Println("⚠️  Deployment failed, rolling back...")
        rollbackInfrastructure(ctx, infra)
        return fmt.Errorf("deployment failed: %w", err)
    }
    fmt.Println("✓ Application deployed")
    fmt.Println()
    
    // Phase 4: Health checks
    fmt.Println("Phase 4: Health Checks")
    if err := waitForHealthy(ctx, deployment); err != nil {
        // Rollback on health check failure
        fmt.Println("⚠️  Health checks failed, rolling back...")
        rollbackDeployment(ctx, deployment)
        return fmt.Errorf("health checks failed: %w", err)
    }
    fmt.Println("✓ Health checks passed")
    fmt.Println()
    
    // Success!
    duration := time.Since(startTime)
    printDeploymentSuccess(deployment, duration)
    
    return nil
}

func provisionInfrastructure(ctx context.Context) (*Infrastructure, error) {
    // Load provider configuration
    cfg, err := providers.LoadConfig("providers.yaml")
    if err != nil {
        return nil, err
    }
    
    infra := &Infrastructure{}
    
    // Provision database (with progress bar)
    bar := ui.NewProgressBar(100, "Provisioning database...")
    dbProvider, _ := providers.DefaultRegistry.GetDatabase(cfg.Database.Provider)
    
    bar.Add(20)
    db, err := dbProvider.Provision(ctx, providers.ProvisionOptions{
        Name:   "gothic-forge-prod",
        Region: cfg.Database.Options.Region,
        Tier:   cfg.Database.Options.Tier,
    })
    if err != nil {
        return nil, err
    }
    bar.Add(80)
    
    infra.Database = db
    fmt.Printf("✓ Database: %s (%s)\n", db.Name, db.ID)
    
    // Provision cache
    bar = ui.NewProgressBar(100, "Provisioning cache...")
    cacheProvider, _ := providers.DefaultRegistry.GetCache(cfg.Cache.Provider)
    
    bar.Add(20)
    cache, err := cacheProvider.Provision(ctx, providers.ProvisionOptions{
        Name:   "gothic-forge-cache",
        Region: cfg.Cache.Options.Region,
        Tier:   cfg.Cache.Options.Tier,
    })
    if err != nil {
        return nil, err
    }
    bar.Add(80)
    
    infra.Cache = cache
    fmt.Printf("✓ Cache: %s (%s)\n", cache.Name, cache.ID)
    
    // Run migrations
    spinner := ui.NewSpinner("Running database migrations...")
    spinner.Start()
    
    if err := dbProvider.RunMigrations(ctx, "app/db/migrations"); err != nil {
        spinner.Fail("Migrations failed")
        return nil, err
    }
    
    spinner.Success("Migrations applied")
    
    return infra, nil
}

func printDeploymentSuccess(deployment *Deployment, duration time.Duration) {
    fmt.Println("═══════════════════════════════════════════════════════")
    fmt.Println("🎉 Deployment Successful!")
    fmt.Printf("   Duration: %v\n", duration.Round(time.Second))
    fmt.Println("═══════════════════════════════════════════════════════")
    fmt.Println()
    fmt.Printf("🌐 Production URL: %s\n", deployment.URL)
    fmt.Printf("📊 Dashboard: %s\n", deployment.DashboardURL)
    fmt.Println()
    fmt.Println("Next steps:")
    fmt.Println("  • Test your app: curl", deployment.URL)
    fmt.Println("  • View logs: gforge logs --follow")
    fmt.Println("  • Scale up: gforge scale --replicas=3")
    fmt.Println()
}
```

---

This shows the concrete implementation approach for the key features. The actual implementation would follow these patterns while adding more error handling, tests, and polish.
