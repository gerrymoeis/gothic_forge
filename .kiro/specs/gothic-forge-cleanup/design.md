# Design Document

## Overview

This design addresses the cleanup and standardization of Gothic Forge v3 following the migration to the Opinionated Stack (Cloudflare Proxy, Leapcell, CockroachDB, Aiven Valkey). The implementation will remove dead code, fix critical bugs, improve error handling, and ensure production readiness.

## Architecture

### Current State
- Mixed provider support (Railway, Neon, Back4app, Leapcell)
- Incomplete deploy command with truncated code
- Inconsistent environment variable naming
- Template error handling gaps
- No graceful shutdown mechanism

### Target State
- Single Opinionated Stack (Cloudflare + Leapcell + CockroachDB + Aiven)
- Complete, working deploy command
- Standardized environment variables
- Robust error handling throughout
- Graceful shutdown with proper cleanup

## Components and Interfaces

### 1. Deploy Command (`cmd/gforge/cmd/deploy.go`)

**Changes:**
- Remove all Railway provider code
- Remove all Neon provider code  
- Remove all Back4app provider code
- Complete the truncated `interactiveEnvSetup()` function
- Simplify provider selection (Leapcell only)
- Update environment variable validation

**Interface:**
```go
// Simplified deploy command - Leapcell only
func runLeapcellDeploy(ctx context.Context, reader *bufio.Reader, dryRun bool) error

// CockroachDB provisioning (keep)
func cockroachInteractiveProvision(ctx context.Context, dryRun bool) (string, error)

// Aiven Valkey provisioning (keep)
func valkeyAutoProvision(ctx context.Context, dryRun bool) (string, error)

// Remove these:
// - isRailwayLinkedCLI
// - setRailwayEnv
// - runRailwayDeploy
// - neonAutoProvision
// - neonInteractiveProvision
// - back4appGuidedSetup
```

### 2. Server Graceful Shutdown (`cmd/server/main.go`)

**New Implementation:**
```go
func main() {
    _ = env.Load()
    r := server.New()
    routes.Register(r)
    
    // Create HTTP server with timeouts
    srv := &http.Server{
        Addr:         getAddr(),
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        log.Printf("Gothic Forge v3 listening at http://%s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with 30s timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Close database connections
    db.Close()
    
    // Shutdown HTTP server
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}
```

### 3. Database Connection Pool (`internal/db/pg.go`)

**Enhanced Configuration:**
```go
func Connect(ctx context.Context) error {
    if pool != nil { return nil }
    
    dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
    if dsn == "" {
        return errors.New("DATABASE_URL is empty")
    }
    
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil { return err }
    
    // Configure connection pool
    cfg.MaxConns = getEnvInt("DB_MAX_CONNS", 20)
    cfg.MinConns = getEnvInt("DB_MIN_CONNS", 2)
    cfg.MaxConnLifetime = time.Hour
    cfg.MaxConnIdleTime = 30 * time.Minute
    cfg.HealthCheckPeriod = time.Minute
    
    cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    p, err := pgxpool.NewWithConfig(cctx, cfg)
    if err != nil { return err }
    
    if err := p.Ping(cctx); err != nil { 
        p.Close()
        return err 
    }
    
    pool = p
    log.Printf("Database pool configured: max=%d, min=%d", cfg.MaxConns, cfg.MinConns)
    return nil
}

func getEnvInt(key string, defaultVal int32) int32 {
    if v := os.Getenv(key); v != "" {
        if n, err := strconv.ParseInt(v, 10, 32); err == nil {
            return int32(n)
        }
    }
    return defaultVal
}
```

### 4. Template Error Handling (`app/templates/*.go`)

**Pattern to Apply:**
```go
// Before (ignoring errors):
_, _ = io.WriteString(w, `<section>...</section>`)

// After (proper error handling):
if _, err := io.WriteString(w, `<section>...</section>`); err != nil {
    return fmt.Errorf("template rendering failed: %w", err)
}

// Or use helper:
func writeHTML(w io.Writer, html string) error {
    if _, err := io.WriteString(w, html); err != nil {
        return fmt.Errorf("write HTML failed: %w", err)
    }
    return nil
}
```

### 5. Environment Variable Standardization

**Standard Names:**
```bash
# Cache (Valkey)
VALKEY_URL=redis://...              # Primary
VALKEY_TLS_SKIP_VERIFY=0           # Primary

# Cloudflare
CLOUDFLARE_API_TOKEN=...            # Primary
CLOUDFLARE_ACCOUNT_ID=...           # Primary
CF_PROJECT_NAME=...                 # Keep (project-specific)

# Compute (Leapcell)
LEAPCELL_APP_URL=...                # Primary

# Database (CockroachDB)
COCKROACH_API_KEY=...               # Primary
DATABASE_URL=...                    # Connection string

# Aiven
AIVEN_TOKEN=...                     # Primary
```

**Deprecation Handling:**
```go
func getEnvWithDeprecation(primary, deprecated, description string) string {
    if v := os.Getenv(primary); v != "" {
        return v
    }
    if v := os.Getenv(deprecated); v != "" {
        log.Printf("WARNING: %s is deprecated, use %s instead", deprecated, primary)
        return v
    }
    return ""
}
```

## Data Models

### Environment Configuration
```go
type DeployConfig struct {
    // Opinionated Stack only
    Provider        string // Always "leapcell"
    
    // Database
    CockroachAPIKey string
    DatabaseURL     string
    
    // Cache
    ValkeyURL       string
    
    // Compute
    LeapcellAppURL  string
    
    // CDN
    CloudflareToken string
    CloudflareAcct  string
    CFProjectName   string
    
    // App
    SiteBaseURL     string
    JWTSecret       string
    AppEnv          string
}

func LoadDeployConfig() (*DeployConfig, error) {
    cfg := &DeployConfig{
        Provider:        "leapcell",
        CockroachAPIKey: os.Getenv("COCKROACH_API_KEY"),
        DatabaseURL:     os.Getenv("DATABASE_URL"),
        ValkeyURL:       getEnvWithDeprecation("VALKEY_URL", "REDIS_URL", "cache"),
        LeapcellAppURL:  os.Getenv("LEAPCELL_APP_URL"),
        CloudflareToken: getEnvWithDeprecation("CLOUDFLARE_API_TOKEN", "CF_API_TOKEN", "Cloudflare"),
        CloudflareAcct:  getEnvWithDeprecation("CLOUDFLARE_ACCOUNT_ID", "CF_ACCOUNT_ID", "Cloudflare"),
        CFProjectName:   os.Getenv("CF_PROJECT_NAME"),
        SiteBaseURL:     os.Getenv("SITE_BASE_URL"),
        JWTSecret:       os.Getenv("JWT_SECRET"),
        AppEnv:          os.Getenv("APP_ENV"),
    }
    
    return cfg, cfg.Validate()
}

func (c *DeployConfig) Validate() error {
    var missing []string
    
    if c.CockroachAPIKey == "" && c.DatabaseURL == "" {
        missing = append(missing, "COCKROACH_API_KEY or DATABASE_URL")
    }
    
    if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
        missing = append(missing, "JWT_SECRET (>=32 chars)")
    }
    
    if len(missing) > 0 {
        return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
    }
    
    return nil
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Deploy Command Completeness
*For any* execution of the deploy command, the command should complete without syntax errors or truncation issues.
**Validates: Requirements 1.1, 1.2, 1.3**

### Property 2: Dead Code Elimination
*For any* search of the codebase for deprecated provider names (railway, neon, back4app), the search should return zero matches in implementation files.
**Validates: Requirements 2.1, 2.2, 2.3, 2.6**

### Property 3: Environment Variable Consistency
*For any* environment variable read operation, the system should use standardized primary names and log deprecation warnings for legacy names.
**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

### Property 4: Template Error Propagation
*For any* template rendering operation that encounters an error, the error should be propagated to the caller and logged with context.
**Validates: Requirements 4.1, 4.2, 4.3, 4.4**

### Property 5: Graceful Shutdown Completion
*For any* SIGTERM or SIGINT signal received by the server, the server should complete in-flight requests within the timeout period before exiting.
**Validates: Requirements 5.1, 5.2, 5.3, 5.6**

### Property 6: Database Pool Configuration
*For any* database connection pool initialization, the pool should be configured with sensible defaults that can be overridden by environment variables.
**Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**

### Property 7: Documentation Accuracy
*For any* documentation file in the repository, the file should reference only the current Opinionated Stack providers.
**Validates: Requirements 7.1, 12.1, 12.2, 12.3**

### Property 8: MIME Type Correctness
*For any* static file served by the server, the Content-Type header should match the file extension regardless of call order.
**Validates: Requirements 8.1, 8.2, 8.4**

### Property 9: Health Check Completeness
*For any* call to /readyz endpoint, the response should include status for all configured Opinionated Stack dependencies.
**Validates: Requirements 10.1, 10.2, 10.3, 10.4**

### Property 10: CLI Output Consistency
*For any* gforge command execution, the output should follow consistent formatting patterns for success, failure, and progress messages.
**Validates: Requirements 11.1, 11.2, 11.3, 11.4**

## Error Handling

### Template Rendering Errors
- Wrap all `io.WriteString` calls with error checks
- Return errors up the component stack
- Log template name and context on failure
- Return HTTP 500 to client with generic error message

### Deployment Errors
- Validate configuration before starting deployment
- Provide actionable error messages with links to documentation
- Support dry-run mode for validation without side effects
- Log all API calls and responses for debugging

### Database Errors
- Retry transient connection errors (3 attempts with exponential backoff)
- Log connection pool statistics on errors
- Return HTTP 503 for database unavailability
- Close pool cleanly on shutdown

### Graceful Shutdown Errors
- Log if shutdown timeout is exceeded
- Force-close connections after timeout
- Report shutdown duration and reason
- Exit with appropriate status code

## Testing Strategy

### Unit Tests
- Test environment variable standardization with deprecated names
- Test database pool configuration with various env values
- Test deploy config validation with missing/invalid values
- Test graceful shutdown signal handling
- Test template error propagation

### Integration Tests
- Test complete deploy workflow in dry-run mode
- Test server startup and graceful shutdown
- Test health check endpoints with mock dependencies
- Test static file serving with various MIME types
- Test database connection pool under load

### Property-Based Tests
- Generate random environment configurations and validate
- Generate random template content and verify error handling
- Generate random shutdown timings and verify cleanup
- Generate random file extensions and verify MIME types

### Manual Testing Checklist
- Deploy to Leapcell from scratch
- Verify all health checks pass
- Test graceful shutdown with active connections
- Verify static files load correctly
- Check logs for deprecation warnings
