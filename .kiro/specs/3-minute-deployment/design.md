# Gothic Forge v10.0 - Technical Design

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Gothic Forge CLI                         │
│                      (gforge command)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Provider Abstraction Layer                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Database   │  │    Cache     │  │   Compute    │      │
│  │   Provider   │  │   Provider   │  │   Provider   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Concrete Provider Implementations               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Cockroach │  │  Valkey  │  │ Leapcell │  │Cloudflare│   │
│  │   DB     │  │  Redis   │  │  Docker  │  │   CDN    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 1. Provider Abstraction Layer

### 1.1 Core Interfaces

Location: `internal/providers/interfaces.go`

```go
// DatabaseProvider abstracts database provisioning and management
type DatabaseProvider interface {
    // Name returns the provider name (e.g., "cockroachdb", "postgresql")
    Name() string
    
    // Provision creates a new database instance
    // Returns connection string and any metadata
    Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error)
    
    // GetConnectionString returns the connection string for an existing database
    GetConnectionString(ctx context.Context) (string, error)
    
    // RunMigrations executes pending database migrations
    RunMigrations(ctx context.Context, migrationsDir string) error
    
    // Health checks if the database is reachable and healthy
    Health(ctx context.Context) error
    
    // Destroy removes the database instance (for cleanup)
    Destroy(ctx context.Context) error
}

// CacheProvider abstracts cache provisioning and management
type CacheProvider interface {
    Name() string
    Provision(ctx context.Context, opts ProvisionOptions) (*CacheInfo, error)
    GetConnectionString(ctx context.Context) (string, error)
    Health(ctx context.Context) error
    Destroy(ctx context.Context) error
}

// ComputeProvider abstracts application deployment
type ComputeProvider interface {
    Name() string
    Deploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error)
    GetURL(ctx context.Context) (string, error)
    GetLogs(ctx context.Context, opts LogOptions) (io.ReadCloser, error)
    Health(ctx context.Context) error
    Rollback(ctx context.Context, version string) error
    Scale(ctx context.Context, replicas int) error
}

// CDNProvider abstracts CDN/edge deployment
type CDNProvider interface {
    Name() string
    Deploy(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error)
    Invalidate(ctx context.Context, paths []string) error
    GetURL(ctx context.Context) (string, error)
}
```

### 1.2 Provider Registry

Location: `internal/providers/registry.go`

```go
// Registry manages available providers
type Registry struct {
    databases map[string]DatabaseProvider
    caches    map[string]CacheProvider
    computes  map[string]ComputeProvider
    cdns      map[string]CDNProvider
}

// Global registry instance
var DefaultRegistry = NewRegistry()

// Register providers at init time
func init() {
    // Database providers
    DefaultRegistry.RegisterDatabase("cockroachdb", &CockroachDBProvider{})
    DefaultRegistry.RegisterDatabase("postgresql", &PostgreSQLProvider{})
    DefaultRegistry.RegisterDatabase("sqlite", &SQLiteProvider{})
    
    // Cache providers
    DefaultRegistry.RegisterCache("valkey", &ValkeyProvider{})
    DefaultRegistry.RegisterCache("redis", &RedisProvider{})
    
    // Compute providers
    DefaultRegistry.RegisterCompute("leapcell", &LeapcellProvider{})
    DefaultRegistry.RegisterCompute("docker", &DockerProvider{})
    
    // CDN providers
    DefaultRegistry.RegisterCDN("cloudflare", &CloudflareProvider{})
}
```

### 1.3 Provider Configuration

Location: `providers.yaml` (project root)

```yaml
# Gothic Forge Provider Configuration
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
```

## 2. Instant Initialization Flow

### 2.1 Command: `gforge init`

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Detect Environment (2s)                              │
│  - Check if Go is installed                                  │
│  - Check if Git repo exists                                  │
│  - Detect project name from directory                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Generate Configuration (3s)                          │
│  - Create .env with secure defaults                          │
│  - Generate JWT_SECRET (64 chars)                            │
│  - Set SITE_BASE_URL from Git remote                         │
│  - Create providers.yaml with defaults                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Install Dependencies (10s)                           │
│  - Run go mod download                                       │
│  - Install templ, gotailwindcss (if missing)                 │
│  - Generate Templ templates                                  │
│  - Build Tailwind CSS                                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Verify Setup (5s)                                    │
│  - Run gforge doctor                                         │
│  - Build project to verify compilation                       │
│  - Run smoke tests                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 5: Success Message                                      │
│  ✓ Project initialized successfully!                         │
│  → Run: gforge dev (start development server)                │
│  → Run: gforge deploy --live (deploy to production)          │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Implementation: `cmd/gforge/cmd/init.go`


```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new Gothic Forge project",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := context.Background()
        
        // Step 1: Detect environment
        spinner := NewSpinner("Detecting environment...")
        spinner.Start()
        
        projectName := detectProjectName()
        gitRemote := detectGitRemote()
        
        spinner.Success("Environment detected")
        
        // Step 2: Generate configuration
        spinner = NewSpinner("Generating configuration...")
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
        spinner = NewSpinner("Installing dependencies...")
        spinner.Start()
        
        if err := installDependencies(ctx); err != nil {
            spinner.Fail("Failed to install dependencies")
            return err
        }
        
        spinner.Success("Dependencies installed")
        
        // Step 4: Verify setup
        spinner = NewSpinner("Verifying setup...")
        spinner.Start()
        
        if err := verifySetup(ctx); err != nil {
            spinner.Fail("Setup verification failed")
            return err
        }
        
        spinner.Success("Setup verified")
        
        // Success message
        printSuccessMessage(projectName)
        
        return nil
    },
}
```

## 3. One-Command Deployment Flow

### 3.1 Command: `gforge deploy --live`

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Pre-flight Checks (5s)                              │
│  ✓ Validate environment variables                            │
│  ✓ Check provider credentials                                │
│  ✓ Build project                                             │
│  ✓ Run tests (optional)                                      │
│  ✓ Display deployment plan                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Provision Infrastructure (60s)                      │
│  [████████░░] 80% Provisioning database...                   │
│  ✓ Database: cockroachdb-prod-abc123                         │
│  [██████████] 100% Provisioning cache...                     │
│  ✓ Cache: valkey-prod-xyz789                                 │
│  ✓ Running migrations... (3 applied)                         │
│  ✓ CDN configured: cloudflare-zone-123                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Deploy Application (45s)                            │
│  [████░░░░░░] 40% Building Docker image...                   │
│  [████████░░] 80% Pushing to registry...                     │
│  [██████████] 100% Deploying to Leapcell...                  │
│  ✓ Deployment: gothic-forge-prod-v1                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 4: Health Checks (10s)                                 │
│  ⏳ Waiting for health checks...                             │
│  ✓ /healthz: OK                                              │
│  ✓ /readyz: OK (db: OK, cache: OK)                           │
│  ✓ Smoke tests: PASSED                                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 🎉 Deployment Successful! (2m 15s)                           │
│                                                               │
│  🌐 Production URL: https://your-app.leapcell.dev            │
│  📊 Dashboard: https://leapcell.io/dashboard                 │
│  📝 Logs: gforge logs --follow                               │
│                                                               │
│  Next steps:                                                 │
│   - Test your app: curl https://your-app.leapcell.dev       │
│   - View logs: gforge logs                                   │
│   - Scale up: gforge scale --replicas=3                      │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Implementation: `cmd/gforge/cmd/deploy_live.go`

```go
var deployLiveCmd = &cobra.Command{
    Use:   "deploy --live",
    Short: "Deploy to production in under 3 minutes",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := context.Background()
        startTime := time.Now()
        
        // Phase 1: Pre-flight checks
        if err := runPreflightChecks(ctx); err != nil {
            return fmt.Errorf("pre-flight checks failed: %w", err)
        }
        
        // Phase 2: Provision infrastructure
        infra, err := provisionInfrastructure(ctx)
        if err != nil {
            return fmt.Errorf("infrastructure provisioning failed: %w", err)
        }
        
        // Phase 3: Deploy application
        deployment, err := deployApplication(ctx, infra)
        if err != nil {
            // Rollback on failure
            rollbackInfrastructure(ctx, infra)
            return fmt.Errorf("deployment failed: %w", err)
        }
        
        // Phase 4: Health checks
        if err := waitForHealthy(ctx, deployment); err != nil {
            // Rollback on health check failure
            rollbackDeployment(ctx, deployment)
            return fmt.Errorf("health checks failed: %w", err)
        }
        
        // Success!
        duration := time.Since(startTime)
        printDeploymentSuccess(deployment, duration)
        
        return nil
    },
}
```

## 4. Progress Indicators & UX

### 4.1 Spinner Component

```go
type Spinner struct {
    message string
    spinner *spinner.Spinner
}

func NewSpinner(message string) *Spinner {
    s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
    s.Suffix = " " + message
    return &Spinner{message: message, spinner: s}
}

func (s *Spinner) Start() {
    s.spinner.Start()
}

func (s *Spinner) Success(message string) {
    s.spinner.Stop()
    fmt.Printf("✓ %s\n", message)
}

func (s *Spinner) Fail(message string) {
    s.spinner.Stop()
    fmt.Printf("✗ %s\n", message)
}
```

### 4.2 Progress Bar Component

```go
type ProgressBar struct {
    bar *progressbar.ProgressBar
}

func NewProgressBar(total int, description string) *ProgressBar {
    bar := progressbar.NewOptions(total,
        progressbar.OptionSetDescription(description),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "█",
            SaucerPadding: "░",
            BarStart:      "[",
            BarEnd:        "]",
        }),
    )
    return &ProgressBar{bar: bar}
}

func (p *ProgressBar) Add(n int) {
    p.bar.Add(n)
}
```

## 5. Error Handling & Recovery

### 5.1 Actionable Error Messages

```go
type DeploymentError struct {
    Phase   string
    Cause   error
    Actions []string
}

func (e *DeploymentError) Error() string {
    var b strings.Builder
    b.WriteString(fmt.Sprintf("❌ Deployment failed in %s phase\n", e.Phase))
    b.WriteString(fmt.Sprintf("   Cause: %s\n\n", e.Cause))
    b.WriteString("   Suggested actions:\n")
    for i, action := range e.Actions {
        b.WriteString(fmt.Sprintf("   %d. %s\n", i+1, action))
    }
    return b.String()
}

// Example usage
func provisionDatabase(ctx context.Context) error {
    provider := getProvider("database")
    _, err := provider.Provision(ctx, opts)
    if err != nil {
        return &DeploymentError{
            Phase: "Database Provisioning",
            Cause: err,
            Actions: []string{
                "Check your COCKROACH_API_KEY is valid",
                "Verify your account has sufficient quota",
                "Try again with: gforge deploy --live --retry",
                "Get help: https://docs.gothicforge.dev/troubleshooting",
            },
        }
    }
    return nil
}
```

### 5.2 Automatic Rollback

```go
func deployWithRollback(ctx context.Context) error {
    // Save current state
    previousState := captureCurrentState(ctx)
    
    // Attempt deployment
    err := attemptDeployment(ctx)
    if err != nil {
        log.Println("⚠️  Deployment failed, rolling back...")
        
        // Rollback to previous state
        if rollbackErr := rollbackToState(ctx, previousState); rollbackErr != nil {
            return fmt.Errorf("deployment failed and rollback failed: %w", rollbackErr)
        }
        
        log.Println("✓ Rollback successful, previous version restored")
        return err
    }
    
    return nil
}
```

## 6. Batteries-Included Features

### 6.1 Background Jobs (Asynq Integration)

Location: `internal/jobs/`

```go
// Job interface
type Job interface {
    Type() string
    Handle(ctx context.Context, payload []byte) error
}

// Example: Email job
type SendEmailJob struct{}

func (j *SendEmailJob) Type() string {
    return "email:send"
}

func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
    var email EmailPayload
    json.Unmarshal(payload, &email)
    return sendEmail(ctx, email)
}

// Enqueue a job
func EnqueueEmail(to, subject, body string) error {
    payload, _ := json.Marshal(EmailPayload{To: to, Subject: subject, Body: body})
    return jobQueue.Enqueue("email:send", payload)
}
```

Scaffolding: `gforge add job SendEmail`

### 6.2 File Storage Abstraction

Location: `internal/storage/`

```go
type StorageProvider interface {
    Put(ctx context.Context, key string, data io.Reader) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// S3-compatible implementation
type S3Storage struct {
    client *s3.Client
    bucket string
}

// Cloudflare R2 implementation
type R2Storage struct {
    client *s3.Client
    bucket string
}
```

Scaffolding: `gforge add storage --provider=r2`

### 6.3 Email Provider Abstraction

Location: `internal/email/`

```go
type EmailProvider interface {
    Send(ctx context.Context, email Email) error
}

type Email struct {
    From    string
    To      []string
    Subject string
    HTML    string
    Text    string
}

// SendGrid implementation
type SendGridProvider struct {
    apiKey string
}

// SMTP implementation
type SMTPProvider struct {
    host string
    port int
    user string
    pass string
}
```

Scaffolding: `gforge add email welcome --provider=sendgrid`

## 7. Testing Strategy

### 7.1 Provider Mocks

```go
type MockDatabaseProvider struct {
    mock.Mock
}

func (m *MockDatabaseProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).(*DatabaseInfo), args.Error(1)
}

// Usage in tests
func TestDeployment(t *testing.T) {
    mockDB := new(MockDatabaseProvider)
    mockDB.On("Provision", mock.Anything, mock.Anything).Return(&DatabaseInfo{
        ConnectionString: "postgresql://test",
    }, nil)
    
    // Test deployment with mock
    err := deployWithProvider(mockDB)
    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
}
```

### 7.2 Integration Tests

```go
func TestFullDeploymentFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    
    // Use real providers in test mode
    cfg := &DeployConfig{
        DatabaseProvider: "sqlite",  // Use SQLite for tests
        CacheProvider:    "none",    // Skip cache
        ComputeProvider:  "docker",  // Use Docker locally
    }
    
    // Run full deployment
    deployment, err := Deploy(ctx, cfg)
    require.NoError(t, err)
    
    // Verify deployment
    assert.NotEmpty(t, deployment.URL)
    
    // Health check
    resp, err := http.Get(deployment.URL + "/healthz")
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Cleanup
    defer deployment.Destroy(ctx)
}
```

## 8. Performance Optimizations

### 8.1 Parallel Provisioning

```go
func provisionInfrastructure(ctx context.Context) (*Infrastructure, error) {
    var wg sync.WaitGroup
    errChan := make(chan error, 3)
    
    infra := &Infrastructure{}
    
    // Provision database (parallel)
    wg.Add(1)
    go func() {
        defer wg.Done()
        db, err := provisionDatabase(ctx)
        if err != nil {
            errChan <- err
            return
        }
        infra.Database = db
    }()
    
    // Provision cache (parallel)
    wg.Add(1)
    go func() {
        defer wg.Done()
        cache, err := provisionCache(ctx)
        if err != nil {
            errChan <- err
            return
        }
        infra.Cache = cache
    }()
    
    // Provision CDN (parallel)
    wg.Add(1)
    go func() {
        defer wg.Done()
        cdn, err := provisionCDN(ctx)
        if err != nil {
            errChan <- err
            return
        }
        infra.CDN = cdn
    }()
    
    wg.Wait()
    close(errChan)
    
    // Check for errors
    for err := range errChan {
        if err != nil {
            return nil, err
        }
    }
    
    return infra, nil
}
```

### 8.2 Caching Provider Metadata

```go
type ProviderCache struct {
    cache map[string]interface{}
    mu    sync.RWMutex
}

func (c *ProviderCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.cache[key]
    return val, ok
}

func (c *ProviderCache) Set(key string, val interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[key] = val
}
```

## 9. Security Considerations

### 9.1 Secrets Management

```go
// Encrypt secrets in .env (optional)
func encryptEnvFile(path string, passphrase string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    encrypted, err := encrypt(data, passphrase)
    if err != nil {
        return err
    }
    
    return os.WriteFile(path+".enc", encrypted, 0600)
}

// Decrypt on load
func loadEncryptedEnv(path string, passphrase string) error {
    data, err := os.ReadFile(path + ".enc")
    if err != nil {
        return err
    }
    
    decrypted, err := decrypt(data, passphrase)
    if err != nil {
        return err
    }
    
    return godotenv.Unmarshal(string(decrypted))
}
```

### 9.2 Credential Validation

```go
func validateProviderCredentials(ctx context.Context) error {
    providers := []string{"database", "cache", "compute", "cdn"}
    
    for _, p := range providers {
        provider := getProvider(p)
        if err := provider.ValidateCredentials(ctx); err != nil {
            return fmt.Errorf("%s credentials invalid: %w", p, err)
        }
    }
    
    return nil
}
```

## 10. Monitoring & Observability

### 10.1 Deployment Metrics

```go
type DeploymentMetrics struct {
    StartTime         time.Time
    EndTime           time.Time
    Duration          time.Duration
    Phase             string
    ProvisioningTime  time.Duration
    DeploymentTime    time.Duration
    HealthCheckTime   time.Duration
    Success           bool
    Error             error
}

func recordDeploymentMetrics(metrics DeploymentMetrics) {
    // Send to telemetry backend
    telemetry.Record("deployment.duration", metrics.Duration.Seconds())
    telemetry.Record("deployment.success", metrics.Success)
    
    // Log for debugging
    log.Printf("Deployment completed in %v (success: %v)", metrics.Duration, metrics.Success)
}
```

## 11. Documentation Generation

### 11.1 Auto-generate API Docs

```go
// Generate OpenAPI spec from routes
func generateOpenAPISpec() error {
    spec := &openapi.Spec{
        OpenAPI: "3.0.0",
        Info: openapi.Info{
            Title:   "Gothic Forge API",
            Version: "1.0.0",
        },
    }
    
    // Scan routes and generate paths
    routes := scanRoutes()
    for _, route := range routes {
        spec.Paths[route.Path] = generatePathItem(route)
    }
    
    // Write to file
    return writeOpenAPISpec(spec, "docs/openapi.yaml")
}
```

## 12. Future Enhancements (v11.0+)

- Multi-region deployments
- Blue-green deployments
- Canary deployments
- Custom domain automation
- Database backups/restore
- Built-in monitoring dashboards
- A/B testing framework
- Feature flags system
- GraphQL support
- WebSocket scaffolding
