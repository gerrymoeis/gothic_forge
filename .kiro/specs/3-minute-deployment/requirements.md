# Gothic Forge v10.0 - "Go Live Under 3 Minutes"

## Vision Statement

**"From zero to production in under 3 minutes"** - A complete development-to-deployment workflow that eliminates all friction points and makes going live as simple as running a single command.

## Core Philosophy

1. **Zero Configuration Required**: Sensible defaults for everything
2. **Automatic Provisioning**: All infrastructure created automatically
3. **One Command Deploy**: `gforge init && gforge deploy --live` and you're done
4. **Provider Flexibility**: Easy to swap providers without code changes
5. **Instant Feedback**: Real-time progress with clear error messages

---

## Requirements

### 1. Provider Abstraction Layer

**Goal**: Make Gothic Forge provider-agnostic while keeping the Opinionated Stack as the default.

#### 1.1 Core Interfaces

- **R1.1.1**: Define `DatabaseProvider` interface with methods: `Provision()`, `GetConnectionString()`, `RunMigrations()`, `Health()`
- **R1.1.2**: Define `CacheProvider` interface with methods: `Provision()`, `GetConnectionString()`, `Health()`
- **R1.1.3**: Define `ComputeProvider` interface with methods: `Deploy()`, `GetURL()`, `GetLogs()`, `Health()`
- **R1.1.4**: Define `CDNProvider` interface with methods: `Deploy()`, `Invalidate()`, `GetURL()`

#### 1.2 Built-in Provider Implementations

**Opinionated Stack (Default):**
- **R1.2.1**: Implement `CockroachDBProvider` (database)
- **R1.2.2**: Implement `ValkeyProvider` (cache)
- **R1.2.3**: Implement `LeapcellProvider` (compute)
- **R1.2.4**: Implement `CloudflareProvider` (CDN)

**Alternative Providers (Community/Enterprise):**
- **R1.2.5**: Implement `PostgreSQLProvider` (local dev, self-hosted)
- **R1.2.6**: Implement `SQLiteProvider` (local dev, testing)
- **R1.2.7**: Implement `RedisProvider` (alternative cache)
- **R1.2.8**: Implement `DockerProvider` (local/self-hosted compute)
- **R1.2.9**: Implement `AWSProvider` (enterprise: RDS, ElastiCache, ECS, CloudFront)

#### 1.3 Provider Configuration

- **R1.3.1**: Create `providers.yaml` config file for provider selection
- **R1.3.2**: Support environment variable overrides (e.g., `GFORGE_DB_PROVIDER=postgresql`)
- **R1.3.3**: Auto-detect provider from existing credentials (smart defaults)
- **R1.3.4**: Validate provider compatibility (e.g., warn if mixing incompatible providers)

---

### 2. Instant Initialization (`gforge init`)

**Goal**: Set up a new project in under 30 seconds with zero manual configuration.

#### 2.1 Project Setup

- **R2.1.1**: `gforge init` creates project structure from template
- **R2.1.2**: Auto-generate secure JWT_SECRET (64 chars)
- **R2.1.3**: Auto-detect Git repo and set SITE_BASE_URL from remote
- **R2.1.4**: Create `.env` with all required variables pre-filled
- **R2.1.5**: Run `go mod download` automatically
- **R2.1.6**: Generate Templ templates automatically
- **R2.1.7**: Build Tailwind CSS automatically
- **R2.1.8**: Run `gforge doctor --fix` automatically

#### 2.2 Interactive Setup (Optional)

- **R2.2.1**: `gforge init --interactive` asks for project name, description
- **R2.2.2**: Prompt for provider preferences (or use defaults)
- **R2.2.3**: Offer to create provider accounts (open signup URLs)
- **R2.2.4**: Store API keys securely in `.env`

#### 2.3 Template Selection

- **R2.3.1**: Support multiple starter templates: `minimal`, `blog`, `saas`, `api`
- **R2.3.2**: `gforge init --template=saas` scaffolds auth, billing, dashboard
- **R2.3.3**: Templates include pre-configured providers and example code

---

### 3. One-Command Deployment (`gforge deploy --live`)

**Goal**: Deploy to production in under 2 minutes with automatic provisioning.

#### 3.1 Pre-flight Checks (5 seconds)

- **R3.1.1**: Validate all required env vars are set
- **R3.1.2**: Check provider credentials are valid
- **R3.1.3**: Verify project builds successfully
- **R3.1.4**: Run tests (optional, skip with `--skip-tests`)
- **R3.1.5**: Display deployment plan (what will be created)

#### 3.2 Automatic Provisioning (60 seconds)

- **R3.2.1**: Provision database (CockroachDB cluster or alternative)
- **R3.2.2**: Provision cache (Valkey instance or alternative)
- **R3.2.3**: Run database migrations automatically
- **R3.2.4**: Seed database with initial data (if `seed.sql` exists)
- **R3.2.5**: Configure CDN (Cloudflare proxy or alternative)
- **R3.2.6**: Set up SSL certificates automatically

#### 3.3 Application Deployment (45 seconds)

- **R3.3.1**: Build Docker image (or binary for Leapcell)
- **R3.3.2**: Push to container registry (or deploy directly)
- **R3.3.3**: Deploy to compute provider (Leapcell, AWS ECS, etc.)
- **R3.3.4**: Configure environment variables automatically
- **R3.3.5**: Set up health checks and monitoring
- **R3.3.6**: Configure auto-scaling rules

#### 3.4 Post-Deployment (10 seconds)

- **R3.4.1**: Wait for health checks to pass
- **R3.4.2**: Run smoke tests against production URL
- **R3.4.3**: Display deployment summary with URLs
- **R3.4.4**: Save deployment metadata to `.gforge/deployments.json`
- **R3.4.5**: Open production URL in browser (optional)

---

### 4. Developer Experience Improvements

#### 4.1 Real-Time Feedback

- **R4.1.1**: Show progress bar for each deployment step
- **R4.1.2**: Stream logs in real-time during deployment
- **R4.1.3**: Display estimated time remaining
- **R4.1.4**: Show colorized output (green=success, yellow=warning, red=error)
- **R4.1.5**: Support `--quiet` mode for CI/CD

#### 4.2 Error Handling

- **R4.2.1**: Provide actionable error messages (not just stack traces)
- **R4.2.2**: Suggest fixes for common errors (e.g., "Run: gforge secrets --set JWT_SECRET=...")
- **R4.2.3**: Auto-retry transient failures (network timeouts, rate limits)
- **R4.2.4**: Rollback on deployment failure (restore previous version)
- **R4.2.5**: Save error logs to `.gforge/logs/` for debugging

#### 4.3 Deployment Management

- **R4.3.1**: `gforge deployments` lists all deployments with status
- **R4.3.2**: `gforge rollback` reverts to previous deployment
- **R4.3.3**: `gforge logs` streams production logs
- **R4.3.4**: `gforge status` shows health of all services
- **R4.3.5**: `gforge scale` adjusts compute resources

---

### 5. Batteries-Included Features

#### 5.1 Authentication & Authorization

- **R5.1.1**: `gforge add auth` scaffolds complete auth system
- **R5.1.2**: Support email/password, OAuth (GitHub, Google, etc.)
- **R5.1.3**: Built-in password reset, email verification
- **R5.1.4**: Role-based access control (RBAC) helpers
- **R5.1.5**: Session management with Redis/Valkey

#### 5.2 Background Jobs

- **R5.2.1**: Built-in job queue using Asynq (Redis-backed)
- **R5.2.2**: `gforge add job <name>` scaffolds job handler
- **R5.2.3**: Job scheduling (cron-like syntax)
- **R5.2.4**: Job retry with exponential backoff
- **R5.2.5**: Web UI for job monitoring

#### 5.3 Email & Notifications

- **R5.3.1**: Email provider abstraction (SendGrid, Mailgun, SES, SMTP)
- **R5.3.2**: `gforge add email <template>` scaffolds email templates
- **R5.3.3**: Transactional email helpers (welcome, reset password, etc.)
- **R5.3.4**: Email preview in development (MailHog integration)

#### 5.4 File Storage

- **R5.4.1**: Storage provider abstraction (S3, Cloudflare R2, local)
- **R5.4.2**: `gforge add storage` configures file uploads
- **R5.4.3**: Image processing (resize, crop, optimize)
- **R5.4.4**: Signed URLs for private files
- **R5.4.5**: CDN integration for public files

#### 5.5 API Features

- **R5.5.1**: Auto-generate OpenAPI spec from routes
- **R5.5.2**: Built-in API versioning (`/v1/`, `/v2/`)
- **R5.5.3**: Request validation with struct tags
- **R5.5.4**: Pagination helpers (cursor-based, offset-based)
- **R5.5.5**: API rate limiting per user/API key

#### 5.6 Observability

- **R5.6.1**: OpenTelemetry integration (traces, metrics, logs)
- **R5.6.2**: `gforge add telemetry` configures observability
- **R5.6.3**: Built-in dashboards (Grafana, Prometheus)
- **R5.6.4**: Error tracking (Sentry integration)
- **R5.6.5**: Performance monitoring (APM)

---

### 6. Testing & Quality

#### 6.1 Automated Testing

- **R6.1.1**: `gforge test` runs all tests with coverage
- **R6.1.2**: `gforge test --watch` runs tests on file changes
- **R6.1.3**: Generate test coverage report (HTML, JSON)
- **R6.1.4**: Integration tests run against real providers (with mocks)
- **R6.1.5**: E2E tests using Playwright (optional)

#### 6.2 Code Quality

- **R6.2.1**: `gforge lint` runs golangci-lint with project config
- **R6.2.2**: `gforge fmt` formats all code (gofmt, goimports)
- **R6.2.3**: `gforge vuln` checks for vulnerabilities
- **R6.2.4**: Pre-commit hooks (optional, via `gforge hooks install`)

---

### 7. Documentation & Learning

#### 7.1 Interactive Tutorials

- **R7.1.1**: `gforge tutorial` launches interactive CLI tutorial
- **R7.1.2**: Step-by-step guide for first deployment
- **R7.1.3**: Video tutorials embedded in docs
- **R7.1.4**: Example projects for common use cases

#### 7.2 Documentation

- **R7.2.1**: Auto-generate API docs from code comments
- **R7.2.2**: Provider comparison guide (when to use what)
- **R7.2.3**: Performance benchmarks published
- **R7.2.4**: Migration guides from other frameworks

---

## Success Metrics

### Primary Metric: Time to Production

- **Target**: < 3 minutes from `git clone` to live URL
- **Breakdown**:
  - `gforge init`: < 30 seconds
  - `gforge deploy --live`: < 2 minutes
  - Health checks + verification: < 30 seconds

### Secondary Metrics

- **Developer Satisfaction**: NPS score > 50
- **Adoption Rate**: 1000+ GitHub stars in 6 months
- **Production Usage**: 100+ apps deployed in first year
- **Community Growth**: 50+ contributors

---

## Non-Functional Requirements

### Performance

- **NFR1**: CLI commands respond in < 100ms (excluding network calls)
- **NFR2**: Deployment completes in < 3 minutes for 95th percentile
- **NFR3**: Provider API calls have 3 retries with exponential backoff

### Reliability

- **NFR4**: Deployment success rate > 99% (excluding user errors)
- **NFR5**: Automatic rollback on deployment failure
- **NFR6**: Health checks prevent routing traffic to unhealthy instances

### Security

- **NFR7**: All secrets stored encrypted in `.env` (optional encryption)
- **NFR8**: Provider credentials never logged or exposed
- **NFR9**: TLS 1.3 required for all external connections
- **NFR10**: Security scanning in CI/CD pipeline

### Usability

- **NFR11**: Error messages include actionable next steps
- **NFR12**: Progress indicators for all long-running operations
- **NFR13**: Colorized output for better readability
- **NFR14**: `--help` text includes examples for every command

---

## Out of Scope (v10.0)

- Multi-region deployments (v11.0)
- Blue-green deployments (v11.0)
- Canary deployments (v11.0)
- Custom domain automation (v10.1)
- Database backups/restore (v10.1)
- Monitoring dashboards (v10.1)

---

## Dependencies

- Go 1.22+
- Docker (optional, for local dev)
- Git
- Provider accounts (auto-created or user-provided)

---

## Timeline Estimate

- **Phase 1**: Provider abstraction layer (2 weeks)
- **Phase 2**: `gforge init` improvements (1 week)
- **Phase 3**: `gforge deploy --live` automation (2 weeks)
- **Phase 4**: Batteries-included features (3 weeks)
- **Phase 5**: Testing & documentation (1 week)
- **Phase 6**: Beta testing & refinement (1 week)

**Total**: ~10 weeks to v10.0 release
