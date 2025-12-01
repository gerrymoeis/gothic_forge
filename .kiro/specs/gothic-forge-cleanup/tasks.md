# Implementation Plan

## Phase 1: Critical Fixes

- [x] 1. Fix deploy.go truncation and compilation errors





  - Complete the truncated `interactiveEnvSetup()` function in `cmd/gforge/cmd/deploy.go`
  - Ensure the function properly handles SITE_BASE_URL prompting
  - Verify the file compiles without errors
  - Test basic deploy command execution
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2. Remove Railway provider code





  - Delete `isRailwayLinkedCLI()` function from `cmd/gforge/cmd/deploy.go`
  - Delete `setRailwayEnv()` function from `cmd/gforge/cmd/deploy.go`
  - Delete `runRailwayDeploy()` function from `cmd/gforge/cmd/deploy.go`
  - Remove all Railway-specific flags from deploy command
  - Remove Railway references from deploy command help text
  - Search and remove any remaining Railway references in cmd/gforge/cmd/
  - _Requirements: 2.1, 2.4_

- [x] 3. Remove Neon provider code





  - Delete `neonAutoProvision()` function from `cmd/gforge/cmd/deploy.go`
  - Delete `neonInteractiveProvision()` function from `cmd/gforge/cmd/deploy.go`
  - Remove Neon-specific flags from deploy command (--neon-region, --neon-project, etc.)
  - Remove Neon references from deploy command logic
  - Update database provisioning to only use CockroachDB
  - _Requirements: 2.2, 2.5_

- [x] 4. Remove Back4app provider code





  - Search for `back4appGuidedSetup` function and delete it
  - Remove Back4app case from provider switch statement in deploy command
  - Remove B4A_APP_URL references
  - Delete any Back4app-specific helper functions
  - _Requirements: 2.3, 2.4_

## Phase 2: Code Quality Improvements

- [x] 5. Implement graceful server shutdown




  - Modify `cmd/server/main.go` to use `http.Server` struct instead of `http.ListenAndServe`
  - Add signal handling for SIGINT and SIGTERM
  - Implement graceful shutdown with 30-second timeout
  - Add database connection pool cleanup on shutdown
  - Add shutdown logging (start, duration, completion)
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 5.1 Write unit tests for graceful shutdown



  - Test signal handling triggers shutdown
  - Test in-flight requests complete before shutdown
  - Test timeout forces shutdown after 30 seconds
  - Test database pool closes cleanly
  - _Requirements: 5.1, 5.2, 5.3, 5.6_

- [x] 6. Configure database connection pooling





  - Update `internal/db/pg.go` Connect() function
  - Set MaxConns from DB_MAX_CONNS env (default: 20)
  - Set MinConns from DB_MIN_CONNS env (default: 2)
  - Set MaxConnLifetime to 1 hour
  - Set MaxConnIdleTime to 30 minutes
  - Set HealthCheckPeriod to 1 minute
  - Add helper function `getEnvInt()` for parsing integer env vars
  - Add logging of pool configuration on startup
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 6.1 Write unit tests for database pool configuration



  - Test default values when env vars not set
  - Test custom values from environment variables
  - Test invalid env values fall back to defaults
  - Test pool configuration logging
  - _Requirements: 6.1, 6.2, 6.5, 6.6_
-

- [x] 7. Fix template error handling in templates.go




  - Update `app/templates/templates.go` Index() function
  - Replace `_, _ = io.WriteString(...)` with proper error checking
  - Add helper function `writeHTML(w io.Writer, html string) error`
  - Propagate errors up through ComponentFunc
  - Add template name to error messages
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
-

- [x] 8. Fix template error handling in db_posts.go




  - Update `app/templates/db_posts.go` DBPostsList() function
  - Update `app/templates/db_posts.go` DBPostsForm() function
  - Replace `_, _ = io.WriteString(...)` with proper error checking
  - Use the writeHTML helper function
  - Ensure errors propagate to HTTP handlers
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 8.1 Write integration tests for template error handling



  - Test template rendering with simulated write errors
  - Test error propagation to HTTP layer
  - Test HTTP 500 response on template errors
  - Test error logging includes template context
  - _Requirements: 4.2, 4.3, 4.5_

## Phase 3: Standardization
-

- [x] 9. Standardize environment variable names




  - Create `getEnvWithDeprecation()` helper in `internal/env/env.go`
  - Update all REDIS_URL references to use VALKEY_URL with deprecation
  - Update all CF_API_TOKEN references to use CLOUDFLARE_API_TOKEN with deprecation
  - Update all CF_ACCOUNT_ID references to use CLOUDFLARE_ACCOUNT_ID with deprecation
  - Remove B4A_APP_URL references (replaced with LEAPCELL_APP_URL)
  - Add deprecation warnings to logs when legacy names are used
  - _Requirements: 3.1, 3.2, 3.3, 3.4_
- [x] 10. Update .env.example with standardized names














- [ ] 10. Update .env.example with standardized names

  - Replace REDIS_URL with VALKEY_URL
  - Replace CF_API_TOKEN with CLOUDFLARE_API_TOKEN
  - Replace CF_ACCOUNT_ID with CLOUDFLARE_ACCOUNT_ID
  - Add LEAPCELL_APP_URL
  - Remove RAILWAY_TOKEN, RAILWAY_API_TOKEN
  - Remove NEON_TOKEN and all NEON_* variables
  - Remove B4A_APP_URL
  - Add comments explaining the Opinionated Stack
  - _Requirements: 3.5, 7.1_

- [x] 11. Create DeployConfig struct





  - Create new file `cmd/gforge/cmd/deploy_config.go`
  - Define DeployConfig struct with Opinionated Stack fields
  - Implement LoadDeployConfig() function
  - Implement Validate() method
  - Use DeployConfig in deploy command instead of individual env reads
  - _Requirements: 1.4, 9.1_
-

- [x] 12. Simplify deploy command provider logic




  - Remove --provider flag (always Leapcell)
  - Remove provider switch statement
  - Simplify deploy command to single Leapcell path
  - Update deploy command help text
  - Update dry-run output to show Opinionated Stack
  - _Requirements: 1.4, 2.4, 9.1_

## Phase 4: Documentation Cleanup

- [x] 13. Remove unrequired documentation files





  - Delete DEPLOYMENT_ANALYSIS_v7.1.md
  - Delete PRODUCTION_STACK_v8.0.md
  - Delete RELEASE_NOTES_v7.1.md
  - Delete LEAPCELL_FIX_INSTRUCTIONS.md
  - Delete CHANGELOG_v6.1.md (if outdated)
  - Update .gitignore to prevent future version-specific docs
  --_Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 14. Update README.md for Opinionated Stack













- [ ] 14. Update README.md for Opinionated Stack

  - Remove Railway references from deployment section
  - Remove Neon references from database section
  - Remove Back4app references
  - Update deployment section to focus on Leapcell
  - Update database section to focus on CockroachDB
  - Add clear "Opinionated Stack" section explaining the choices
  - Update provider comparison table (remove Railway, Neon, Back4app)
  - _Requirements: 12.1, 12.3, 12.4_

- [x] 15. Update QUICKSTART.md for Opinionated Stack





  - Remove Railway deployment path
  - Remove Back4app deployment path
  - Simplify to single Leapcell deployment path
  - Update API key links (remove Railway, Neon, Back4app)
  - Update deployment comparison table
  - Add Leapcell-specific troubleshooting section
  - _Requirements: 12.2, 12.4, 12.5_

- [x] 16. Update CONTRIBUTING.md




  - Remove references to deprecated providers
  - Update testing instructions
  - Update deployment testing instructions for Leapcell
  - _Requirements: 12.1_

## Phase 5: Testing & Validation

- [x] 17. Update integration tests




  - Remove Railway-specific tests
  - Remove Neon-specific tests
  - Remove Back4app-specific tests
  - Add Leapcell deployment dry-run test
  - Add CockroachDB provisioning test (dry-run)
  - _Requirements: 2.1, 2.2, 2.3_
-

- [x] 18. Add health check tests








  - Test /readyz with CockroachDB available
  - Test /readyz with CockroachDB unavailable
  - Test /readyz with Valkey available
  - Test /readyz with Valkey unavailable
  - Test /readyz response format and status codes
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_


- [x] 18.1 Write property tests for health checks








  - Generate random dependency states and verify correct HTTP status
  - Generate random dependency combinations and verify response format
  - _Requirements: 10.3, 10.4_

- [x] 19. Verify MIME type handling





  - Test static file serving for .css files
  - Test static file serving for .js files
  - Test static file serving for .svg files
  - Test static file serving for .woff2 files
  - Test Content-Type header is set correctly
  - Test mimeTypeResponseWriter edge cases
  - _Requirements: 8.1, 8.2, 8.4, 8.5_

- [x] 20. Run comprehensive test suite




  - Run `gforge test --with-build --coverage`
  - Run `go vet ./...`
  - Run `golangci-lint run` (if available)
  - Verify all tests pass
  - Verify no lint errors
  - _Requirements: All_

## Phase 6: Final Validation
-

- [x] 21. Manual deployment test



  - Create fresh .env with Opinionated Stack credentials
  - Run `gforge doctor --fix`
  - Run `gforge deploy --dry-run`
  - Run `gforge deploy` (to Leapcell)
  - Verify application deploys successfully
  - Verify health checks pass
  - Verify static files load correctly
  - Test graceful shutdown with `kill -TERM <pid>`
  - _Requirements: All_

- [x] 22. Documentation review





  - Read through README.md for accuracy
  - Read through QUICKSTART.md for accuracy
  - Verify all links work
  - Verify code examples are correct
  - Check for any remaining deprecated provider references
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [x] 23. Code cleanup verification




  - Search codebase for "railway" (case-insensitive)
  - Search codebase for "neon" (case-insensitive, excluding "none")
  - Search codebase for "back4app" (case-insensitive)
  - Verify zero matches in implementation files
  - Verify .gitignore prevents version-specific docs
  - _Requirements: 2.1, 2.2, 2.3, 7.4_
-

- [x] 24. Final checkpoint - Ensure all tests pass




  - Ensure all tests pass, ask the user if questions arise
  - Verify build succeeds: `gforge build`
  - Verify tests pass: `gforge test --with-build`
  - Verify no vet errors: `go vet ./...`
  - Verify deploy dry-run works: `gforge deploy --dry-run`
  - Verify doctor passes: `gforge doctor`
  - _Requirements: All_
