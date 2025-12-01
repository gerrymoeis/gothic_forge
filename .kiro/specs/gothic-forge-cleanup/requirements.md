# Requirements Document

## Introduction

This specification addresses critical issues, code cleanup, and standardization for Gothic Forge v3 following the migration to the opinionated production stack (Cloudflare Proxy, Leapcell, CockroachDB, Aiven Valkey). The goal is to remove dead code from deprecated providers (Railway, Neon, Back4app), fix critical bugs, improve error handling, and standardize the codebase for production readiness.

## Glossary

- **Gothic Forge**: The Go web framework being cleaned up
- **Opinionated Stack**: The standardized production deployment stack (Cloudflare Proxy + Leapcell + CockroachDB + Aiven Valkey)
- **Dead Code**: Code for deprecated providers (Railway, Neon, Back4app) that is no longer used
- **Deploy Command**: The `gforge deploy` CLI command for deployment automation
- **Template System**: The Templ-based type-safe HTML rendering system
- **Server**: The HTTP server component in `cmd/server/main.go`
- **CLI**: The `gforge` command-line interface tool

## Requirements

### Requirement 1: Fix Critical Deploy Command Bug

**User Story:** As a developer, I want the deploy command to compile and execute successfully, so that I can deploy my application without errors.

#### Acceptance Criteria

1. WHEN the deploy.go file is parsed by the Go compiler, THEN the system SHALL compile without syntax errors
2. WHEN the `interactiveEnvSetup()` function is called, THEN the system SHALL complete execution without truncation errors
3. WHEN a user runs `gforge deploy`, THEN the system SHALL execute the complete deployment workflow
4. WHEN the deploy command validates environment variables, THEN the system SHALL only check for Opinionated Stack providers (CockroachDB, Aiven, Cloudflare, Leapcell)

### Requirement 2: Remove Dead Provider Code

**User Story:** As a maintainer, I want all Railway, Neon, and Back4app provider code removed, so that the codebase only reflects the current Opinionated Stack.

#### Acceptance Criteria

1. WHEN searching the codebase for "railway", THEN the system SHALL return zero matches in implementation files (excluding documentation)
2. WHEN searching the codebase for "neon", THEN the system SHALL return zero matches in implementation files (excluding documentation)
3. WHEN searching the codebase for "back4app", THEN the system SHALL return zero matches in implementation files (excluding documentation)
4. WHEN the deploy command executes, THEN the system SHALL only reference Leapcell as the compute provider
5. WHEN the deploy command provisions a database, THEN the system SHALL only use CockroachDB provisioning logic
6. WHEN stub functions for removed providers are searched, THEN the system SHALL return zero results

### Requirement 3: Standardize Environment Variable Naming

**User Story:** As a developer, I want consistent environment variable naming, so that configuration is predictable and clear.

#### Acceptance Criteria

1. WHEN the system reads cache configuration, THEN the system SHALL use `VALKEY_URL` as the primary variable name
2. WHEN the system reads Cloudflare configuration, THEN the system SHALL use `CLOUDFLARE_API_TOKEN` as the primary variable name
3. WHEN the system reads compute provider configuration, THEN the system SHALL use `LEAPCELL_APP_URL` as the standard variable name
4. WHEN legacy variable names are encountered, THEN the system SHALL provide deprecation warnings in logs
5. WHEN the .env.example file is generated, THEN the system SHALL use only standardized variable names

### Requirement 4: Implement Proper Error Handling in Templates

**User Story:** As a developer, I want template rendering errors to be caught and logged, so that I can debug issues in production.

#### Acceptance Criteria

1. WHEN a template writes HTML content, THEN the system SHALL check and handle io.WriteString errors
2. WHEN a template rendering fails, THEN the system SHALL log the error with context
3. WHEN a template error occurs, THEN the system SHALL return an HTTP 500 error to the client
4. WHEN using templ.ComponentFunc, THEN the system SHALL properly propagate errors up the call stack
5. WHEN template errors are logged, THEN the system SHALL include the template name and line number

### Requirement 5: Add Graceful Server Shutdown

**User Story:** As a system administrator, I want the server to shut down gracefully on SIGTERM/SIGINT, so that in-flight requests complete and connections close cleanly.

#### Acceptance Criteria

1. WHEN the server receives SIGTERM or SIGINT, THEN the system SHALL initiate graceful shutdown
2. WHEN graceful shutdown begins, THEN the system SHALL stop accepting new connections
3. WHEN graceful shutdown is in progress, THEN the system SHALL wait for in-flight requests to complete (up to 30 seconds)
4. WHEN the shutdown timeout expires, THEN the system SHALL force-close remaining connections
5. WHEN the server shuts down, THEN the system SHALL log the shutdown reason and duration
6. WHEN database connections exist during shutdown, THEN the system SHALL close the connection pool cleanly

### Requirement 6: Configure Database Connection Pooling

**User Story:** As a system administrator, I want proper database connection pool configuration, so that the application scales reliably under load.

#### Acceptance Criteria

1. WHEN the database pool is initialized, THEN the system SHALL set MaxConns to a configurable value (default: 20)
2. WHEN the database pool is initialized, THEN the system SHALL set MinConns to a configurable value (default: 2)
3. WHEN the database pool is initialized, THEN the system SHALL set MaxConnLifetime to 1 hour
4. WHEN the database pool is initialized, THEN the system SHALL set MaxConnIdleTime to 30 minutes
5. WHEN environment variables override pool settings, THEN the system SHALL use the provided values
6. WHEN the pool configuration is invalid, THEN the system SHALL log a warning and use safe defaults

### Requirement 7: Remove Unrequired Documentation Files

**User Story:** As a new user, I want a clean repository with only current documentation, so that I'm not confused by outdated version-specific files.

#### Acceptance Criteria

1. WHEN the repository root is listed, THEN the system SHALL contain only current documentation files (README.md, QUICKSTART.md, CONTRIBUTING.md, SECURITY.md, LICENSE, CODE_OF_CONDUCT.md)
2. WHEN version-specific documentation exists, THEN the system SHALL be removed from the repository
3. WHEN deployment analysis documents exist, THEN the system SHALL be removed from the repository
4. WHEN the .gitignore is updated, THEN the system SHALL prevent future version-specific docs from being committed
5. WHEN users need historical documentation, THEN the system SHALL reference the appropriate git branch in README.md

### Requirement 8: Fix MIME Type Response Writer Edge Cases

**User Story:** As a developer, I want static file serving to work correctly in all scenarios, so that CSS/JS files load properly on all platforms.

#### Acceptance Criteria

1. WHEN http.FileServer calls Write() before WriteHeader(), THEN the system SHALL still set the correct Content-Type
2. WHEN multiple Write() calls occur, THEN the system SHALL only set Content-Type once
3. WHEN the ResponseWriter is wrapped, THEN the system SHALL preserve all http.ResponseWriter interface methods
4. WHEN serving static files, THEN the system SHALL set correct MIME types for .css, .js, .svg, .woff2 files
5. WHEN the MIME type is already set by FileServer, THEN the system SHALL not override it if it's correct

### Requirement 9: Improve Deploy Command User Experience

**User Story:** As a developer deploying for the first time, I want clear guidance and validation, so that I can successfully deploy without confusion.

#### Acceptance Criteria

1. WHEN the deploy command starts, THEN the system SHALL display the current Opinionated Stack (Leapcell + CockroachDB + Aiven + Cloudflare)
2. WHEN required environment variables are missing, THEN the system SHALL provide direct links to obtain API keys
3. WHEN the deploy command validates prerequisites, THEN the system SHALL check only for Leapcell-relevant tools
4. WHEN the user runs deploy in dry-run mode, THEN the system SHALL show exactly what would be executed
5. WHEN the deploy command completes, THEN the system SHALL display the application URL and next steps

### Requirement 10: Add Health Check for Leapcell Deployment

**User Story:** As a system administrator, I want health checks to verify Leapcell-specific requirements, so that I know the deployment is healthy.

#### Acceptance Criteria

1. WHEN the /readyz endpoint is called, THEN the system SHALL check CockroachDB connectivity
2. WHEN the /readyz endpoint is called, THEN the system SHALL check Aiven Valkey connectivity (if configured)
3. WHEN the /readyz endpoint is called, THEN the system SHALL return detailed status for each dependency
4. WHEN a dependency is unavailable, THEN the system SHALL return HTTP 503 with specific failure details
5. WHEN all dependencies are healthy, THEN the system SHALL return HTTP 200 with "ready" status

### Requirement 11: Standardize CLI Command Output

**User Story:** As a developer using the CLI, I want consistent, readable output across all commands, so that I can easily understand what's happening.

#### Acceptance Criteria

1. WHEN any gforge command executes, THEN the system SHALL use consistent formatting (emoji, colors, indentation)
2. WHEN a command succeeds, THEN the system SHALL display a clear success message with next steps
3. WHEN a command fails, THEN the system SHALL display the error and suggest remediation steps
4. WHEN a command performs multiple steps, THEN the system SHALL show progress indicators
5. WHEN verbose mode is enabled, THEN the system SHALL display detailed execution logs

### Requirement 12: Update Documentation for Opinionated Stack

**User Story:** As a new user, I want documentation that accurately reflects the current Opinionated Stack, so that I can deploy successfully on first try.

#### Acceptance Criteria

1. WHEN README.md is read, THEN the system SHALL document only Leapcell as the compute provider
2. WHEN QUICKSTART.md is read, THEN the system SHALL provide step-by-step Leapcell deployment instructions
3. WHEN deployment documentation is read, THEN the system SHALL explain the Cloudflare Proxy + Leapcell architecture
4. WHEN API key documentation is read, THEN the system SHALL provide links for CockroachDB, Aiven, Cloudflare, and Leapcell
5. WHEN troubleshooting documentation is read, THEN the system SHALL address common Leapcell deployment issues
