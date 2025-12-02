# Quiet Mode for CI/CD

Gothic Forge CLI supports a `--quiet` (or `-q`) flag for minimal output suitable for CI/CD pipelines.

## Usage

Add the `--quiet` flag to any command:

```bash
gforge --quiet init
gforge --quiet build
gforge --quiet deploy-live
gforge -q test
```

## What Quiet Mode Does

When quiet mode is enabled:

1. **Disables Colors**: All ANSI color codes are suppressed by setting `NO_COLOR=1`
2. **Simplifies Spinners**: Instead of animated spinners, simple text messages are printed
3. **Reduces Progress Bars**: Progress bars show only milestone percentages (25%, 50%, 75%, 100%)
4. **Removes Banner**: The Gothic Forge banner is not displayed
5. **Minimal Output**: Only essential information is printed

## Output Format

### Normal Mode
```
Gothic Forge v3 :: CLI
⠋ Provisioning database...
[████████░░] 80% Provisioning database...
✓ Database provisioned successfully
```

### Quiet Mode
```
Provisioning database...
[25%] Provisioning database...
[50%] Provisioning database...
[75%] Provisioning database...
[100%] Provisioning database...
[OK] Database provisioned successfully
```

## Status Indicators

In quiet mode, status indicators use simple text prefixes:

- `[OK]` - Success (replaces ✓)
- `[FAIL]` - Failure (replaces ✗)
- `[WARN]` - Warning (replaces ⚠)
- `[25%]`, `[50%]`, `[75%]`, `[100%]` - Progress milestones

## CI/CD Integration

### GitHub Actions

```yaml
- name: Deploy to production
  run: gforge --quiet deploy-live
  env:
    COCKROACH_API_KEY: ${{ secrets.COCKROACH_API_KEY }}
    AIVEN_TOKEN: ${{ secrets.AIVEN_TOKEN }}
```

### GitLab CI

```yaml
deploy:
  script:
    - gforge --quiet deploy-live
  variables:
    COCKROACH_API_KEY: $COCKROACH_API_KEY
    AIVEN_TOKEN: $AIVEN_TOKEN
```

### Jenkins

```groovy
stage('Deploy') {
    steps {
        sh 'gforge --quiet deploy-live'
    }
}
```

## Environment Variable

Quiet mode automatically detects when output is not a terminal (e.g., piped to a file or running in CI/CD) and adjusts accordingly. You can also force quiet mode by setting:

```bash
export NO_COLOR=1
```

## Benefits for CI/CD

1. **Cleaner Logs**: No ANSI escape codes cluttering your CI/CD logs
2. **Faster Parsing**: Simple text output is easier to parse and search
3. **Better Compatibility**: Works with all CI/CD systems regardless of terminal support
4. **Reduced Noise**: Only essential information is displayed
5. **Consistent Output**: Predictable format for automated processing

## Example: Full Deployment

```bash
# Normal mode (local development)
gforge deploy-live

# Quiet mode (CI/CD)
gforge --quiet deploy-live
```

### Normal Mode Output
```
Gothic Forge v3 :: CLI
Deploy Live - Opinionated Stack
────────────────────────────────────────

📋 Phase 1: Pre-flight Checks (5s)
────────────────────────────────────────
  → Validating environment variables...
  ✓ Environment variables validated
  → Checking provider credentials...
  ✓ Provider credentials validated
  → Building project...
  ✓ Project built successfully

🏗️  Phase 2: Provision Infrastructure (60s)
────────────────────────────────────────
  → Provisioning database...
  ✓ Database: cockroachdb-prod-abc123
  → Running database migrations...
  ✓ Migrations applied

🚀 Phase 3: Deploy Application (45s)
────────────────────────────────────────
  → Building and deploying to Leapcell...
  ✓ Deployment: gothic-forge-prod-v1
  ✓ URL: https://your-app.leapcell.dev

🏥 Phase 4: Health Checks (10s)
────────────────────────────────────────
  → Waiting for health checks...
  ✓ /healthz: OK
  ✓ Application is healthy

────────────────────────────────────────
🎉 Deployment Successful! (2m 15s)
────────────────────────────────────────

🌐 Production URL: https://your-app.leapcell.dev
📊 Dashboard: https://leapcell.io/dashboard

Next steps:
  • Test your app: curl https://your-app.leapcell.dev/healthz
  • View logs: gforge logs --follow
  • Scale up: gforge scale --replicas=3
```

### Quiet Mode Output
```
Phase 1: Pre-flight Checks (5s)
Validating environment variables...
[OK] Environment variables validated
Checking provider credentials...
[OK] Provider credentials validated
Building project...
[OK] Project built successfully

Phase 2: Provision Infrastructure (60s)
Provisioning database...
[OK] Database: cockroachdb-prod-abc123
Running database migrations...
[OK] Migrations applied

Phase 3: Deploy Application (45s)
Building and deploying to Leapcell...
[OK] Deployment: gothic-forge-prod-v1
[OK] URL: https://your-app.leapcell.dev

Phase 4: Health Checks (10s)
Waiting for health checks...
[OK] /healthz: OK
[OK] Application is healthy

Deployment Successful! (2m 15s)
Production URL: https://your-app.leapcell.dev
```

## Implementation Details

The quiet mode is implemented at multiple levels:

1. **Root Command**: Global `--quiet` flag sets `NO_COLOR=1` environment variable
2. **UI Components**: Spinner and ProgressBar components check for quiet mode
3. **Color Module**: Automatically disables colors when `NO_COLOR` is set
4. **Banner**: Suppressed in quiet mode

All commands automatically respect the quiet mode flag without requiring individual implementation.
