# Manual Deployment Test Guide

This guide walks you through the manual deployment test for Task 21 of the Gothic Forge cleanup specification.

## Prerequisites

Before starting, ensure you have accounts and API keys for the Opinionated Stack:

1. **Leapcell** (Compute)
   - Sign up: https://leapcell.io/signup
   - 20 FREE projects on Hobby tier!
   - GitHub integration required

2. **CockroachDB Serverless** (Database)
   - Sign up: https://cockroachlabs.cloud/signup
   - Create service account: https://cockroachlabs.cloud/service-accounts
   - Get API key

3. **Aiven Valkey** (Cache - Optional)
   - Sign up: https://console.aiven.io/signup
   - Get API token: https://console.aiven.io/profile/tokens

4. **Cloudflare** (CDN/Proxy - Optional)
   - Sign up: https://dash.cloudflare.com/sign-up
   - Create API token: https://dash.cloudflare.com/profile/api-tokens
   - Use template: "Edit Cloudflare Workers" with Pages permissions

## Test Procedure

### Step 1: Create Fresh .env File

1. Backup your existing .env (if any):
   ```bash
   copy .env .env.backup
   ```

2. Remove the current .env:
   ```bash
   del .env
   ```

3. Copy from .env.example:
   ```bash
   copy .env.example .env
   ```

4. Edit .env and fill in the following required values:

   ```env
   # Application Settings
   APP_ENV=production
   SITE_BASE_URL=https://your-app-name.leapcell.dev
   
   # Security (generate with: gforge secrets --gen-jwt)
   JWT_SECRET=<64-character-hex-string>
   
   # Database
   COCKROACH_API_KEY=<your-cockroach-api-key>
   
   # Cache (Optional)
   AIVEN_TOKEN=<your-aiven-token>
   
   # CDN (Optional)
   CLOUDFLARE_API_TOKEN=<your-cloudflare-token>
   CLOUDFLARE_ACCOUNT_ID=<your-account-id>
   CF_PROJECT_NAME=<your-project-name>
   ```

5. Generate JWT_SECRET:
   ```bash
   gforge secrets --gen-jwt
   ```

### Step 2: Run Doctor Command

Run the doctor command with --fix to ensure all tools are installed:

```bash
gforge doctor --fix
```

**Expected Output:**
- ✓ Go version displayed
- ✓ git: present
- ✓ templ: present (or installed)
- ✓ gotailwindcss: present (or installed)
- ✓ wrangler: present (or installed if Cloudflare is configured)
- ✓ docker: present and daemon running
- ✓ Dockerfile: present and valid
- ✓ .env: present
- ✓ .env.example: present
- ✓ Port 8080: available
- ✓ CockroachDB: reachable (if DATABASE_URL set)
- ✓ Valkey/Redis: PING ok (if VALKEY_URL set)
- ✓ Ready for deploy: Yes

**Checklist:**
- [ ] Doctor command completes without errors
- [ ] All required tools are present
- [ ] Docker daemon is running
- [ ] Dockerfile syntax is valid
- [ ] .env file is present with all required keys
- [ ] Readiness summary shows "Ready for deploy: Yes"

### Step 3: Run Deploy Dry-Run

Test the deployment workflow without making actual changes:

```bash
gforge deploy --dry-run
```

**Expected Output:**
- ✓ Opinionated Stack displayed (Leapcell + CockroachDB + Aiven + Cloudflare)
- ✓ Configuration check shows all required keys present
- ✓ Provider links displayed
- ✓ sitemap.xml and robots.txt status shown
- ✓ Build artifacts preparation mentioned
- ✓ CockroachDB provisioning plan shown
- ✓ Aiven Valkey provisioning plan shown (if configured)
- ✓ Leapcell deployment plan shown
- ✓ Cloudflare Pages plan shown (if configured)

**Checklist:**
- [ ] Dry-run completes without errors
- [ ] All provider configurations are validated
- [ ] No missing required secrets
- [ ] Deployment plan is clear and accurate

### Step 4: Run Actual Deployment

Deploy to Leapcell:

```bash
gforge deploy
```

**Expected Workflow:**
1. Interactive env setup (if needed)
2. CockroachDB provisioning
   - Cluster creation or connection
   - Database and user creation
   - Migrations run automatically
   - DATABASE_URL saved to .env
3. Aiven Valkey provisioning (optional, prompted)
   - Service creation
   - VALKEY_URL saved to .env
4. Build execution
   - Templ templates compiled
   - Tailwind CSS built
   - Static assets generated
5. Cloudflare Pages deployment (optional, prompted)
   - Static export deployed
6. Leapcell deployment
   - Docker image built
   - Image pushed to Leapcell
   - Application deployed
   - URL displayed

**Checklist:**
- [ ] Deployment completes without errors
- [ ] CockroachDB cluster is created and accessible
- [ ] DATABASE_URL is saved to .env
- [ ] Valkey service is created (if opted in)
- [ ] VALKEY_URL is saved to .env (if opted in)
- [ ] Build completes successfully
- [ ] Leapcell deployment succeeds
- [ ] Application URL is displayed
- [ ] Deployment completion message shown

### Step 5: Verify Application Deployment

Test the deployed application:

1. **Access the application:**
   ```bash
   curl https://your-app-name.leapcell.dev
   ```
   - [ ] Application responds with HTTP 200
   - [ ] HTML content is returned
   - [ ] No error messages in response

2. **Check health endpoint:**
   ```bash
   curl https://your-app-name.leapcell.dev/readyz
   ```
   - [ ] Returns HTTP 200 (all dependencies healthy)
   - [ ] Response includes CockroachDB status
   - [ ] Response includes Valkey status (if configured)
   - [ ] JSON format is correct

3. **Test in browser:**
   - Open https://your-app-name.leapcell.dev in a browser
   - [ ] Page loads correctly
   - [ ] No console errors
   - [ ] Styles are applied (CSS loaded)
   - [ ] JavaScript works (if applicable)

### Step 6: Verify Static Files

Test that static files are served correctly:

1. **CSS files:**
   ```bash
   curl -I https://your-app-name.leapcell.dev/static/styles/output.css
   ```
   - [ ] Returns HTTP 200
   - [ ] Content-Type: text/css

2. **JavaScript files:**
   ```bash
   curl -I https://your-app-name.leapcell.dev/static/app.js
   ```
   - [ ] Returns HTTP 200
   - [ ] Content-Type: application/javascript or text/javascript

3. **SVG files:**
   ```bash
   curl -I https://your-app-name.leapcell.dev/static/favicon.svg
   ```
   - [ ] Returns HTTP 200
   - [ ] Content-Type: image/svg+xml

4. **Other static files:**
   ```bash
   curl -I https://your-app-name.leapcell.dev/static/robots.txt
   curl -I https://your-app-name.leapcell.dev/static/sitemap.xml
   ```
   - [ ] robots.txt returns HTTP 200
   - [ ] sitemap.xml returns HTTP 200

### Step 7: Test Graceful Shutdown

Test that the server shuts down gracefully:

**Note:** This test requires SSH access to the Leapcell container or running the server locally.

**Local Test (Recommended):**

1. Start the server locally:
   ```bash
   go run cmd/server/main.go
   ```

2. In another terminal, make a request:
   ```bash
   curl http://127.0.0.1:8080
   ```

3. Find the process ID:
   ```bash
   # Windows PowerShell
   Get-Process -Name "server" | Select-Object Id
   
   # Or use Task Manager to find the PID
   ```

4. Send SIGTERM signal:
   ```bash
   # Windows PowerShell (requires admin)
   Stop-Process -Id <PID> -Force
   
   # Or use Ctrl+C in the terminal running the server
   ```

5. Check the server logs:
   - [ ] "Shutting down server..." message appears
   - [ ] In-flight requests complete (if any)
   - [ ] Database connections close cleanly
   - [ ] "Server exited" message appears
   - [ ] Shutdown completes within 30 seconds

**Leapcell Test (If SSH Access Available):**

1. SSH into the Leapcell container (if supported)
2. Find the server process:
   ```bash
   ps aux | grep server
   ```
3. Send SIGTERM:
   ```bash
   kill -TERM <PID>
   ```
4. Check logs via Leapcell dashboard or `gforge logs`

### Step 8: Verify Database Connection Pool

Check that the database connection pool is configured correctly:

1. Check the server logs for pool configuration:
   ```bash
   gforge logs
   ```
   - [ ] Log message shows: "Database pool configured: max=20, min=2"
   - [ ] Or custom values if DB_MAX_CONNS/DB_MIN_CONNS are set

2. Test database connectivity under load (optional):
   - Use a load testing tool (e.g., Apache Bench, wrk)
   - Verify connections are pooled and reused
   - Check for connection errors

### Step 9: Verify Environment Variable Standardization

Check that environment variables use standardized names:

1. Review .env file:
   - [ ] VALKEY_URL is used (not REDIS_URL)
   - [ ] CLOUDFLARE_API_TOKEN is used (not CF_API_TOKEN)
   - [ ] CLOUDFLARE_ACCOUNT_ID is used (not CF_ACCOUNT_ID)
   - [ ] LEAPCELL_APP_URL is present
   - [ ] No Railway, Neon, or Back4app variables

2. Check for deprecation warnings in logs:
   - If legacy names are used, warnings should appear
   - [ ] Deprecation warnings are logged (if applicable)

### Step 10: Final Verification

Complete the final checks:

1. **Code search for dead providers:**
   ```bash
   # Search for Railway references
   findstr /s /i "railway" cmd\gforge\cmd\*.go
   
   # Search for Neon references
   findstr /s /i "neon" cmd\gforge\cmd\*.go
   
   # Search for Back4app references
   findstr /s /i "back4app" cmd\gforge\cmd\*.go
   ```
   - [ ] No Railway references in implementation files
   - [ ] No Neon references in implementation files
   - [ ] No Back4app references in implementation files

2. **Documentation review:**
   - [ ] README.md mentions only Opinionated Stack
   - [ ] QUICKSTART.md has Leapcell deployment path
   - [ ] No outdated version-specific docs in root

3. **Test suite:**
   ```bash
   gforge test --with-build
   ```
   - [ ] All tests pass
   - [ ] No compilation errors
   - [ ] Coverage is adequate

## Success Criteria

All of the following must be true for the manual deployment test to pass:

- [ ] Fresh .env created with Opinionated Stack credentials
- [ ] `gforge doctor --fix` completes successfully
- [ ] `gforge deploy --dry-run` shows correct deployment plan
- [ ] `gforge deploy` completes without errors
- [ ] Application is accessible at Leapcell URL
- [ ] Health checks return HTTP 200
- [ ] Static files load correctly with proper MIME types
- [ ] Graceful shutdown works (tested locally)
- [ ] Database connection pool is configured
- [ ] Environment variables use standardized names
- [ ] No dead provider code remains
- [ ] All tests pass

## Troubleshooting

### Common Issues

1. **Docker daemon not running:**
   - Start Docker Desktop
   - Wait for it to fully initialize
   - Run `docker ps` to verify

2. **Missing API keys:**
   - Double-check .env file
   - Verify keys are valid (not expired)
   - Check for extra whitespace

3. **CockroachDB provisioning fails:**
   - Verify COCKROACH_API_KEY is correct
   - Check service account permissions
   - Try manual cluster creation first

4. **Leapcell deployment fails:**
   - Ensure Docker is running
   - Check Dockerfile syntax
   - Verify GitHub integration

5. **Health check fails:**
   - Check DATABASE_URL is correct
   - Verify VALKEY_URL is accessible
   - Review server logs for errors

## Cleanup

After testing, you can:

1. Keep the deployment for production use
2. Delete the Leapcell app from the dashboard
3. Delete the CockroachDB cluster (if test-only)
4. Delete the Aiven Valkey service (if test-only)
5. Restore your original .env:
   ```bash
   copy .env.backup .env
   ```

## Notes

- This test validates the entire deployment workflow end-to-end
- It confirms that all cleanup tasks have been completed successfully
- It ensures the Opinionated Stack is working as designed
- Manual testing is required because it involves real external services

