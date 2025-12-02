# Deployment Errors Package

This package provides structured error types for Gothic Forge deployment operations, offering actionable error messages with suggested fixes.

## Overview

The `DeploymentError` type wraps deployment failures with:
- **Phase**: Which deployment phase failed (e.g., "Database Provisioning", "Health Check")
- **Cause**: The underlying error that triggered the failure
- **Actions**: A list of suggested actions to resolve the error

## Usage

### Basic Usage

```go
import (
    "errors"
    deployerrors "gothicforge3/internal/errors"
)

// Create a custom deployment error
err := deployerrors.NewDeploymentError(
    "Network Configuration",
    errors.New("connection timeout"),
    "Check your network connection",
    "Verify firewall rules",
    "Try again with --retry",
)

fmt.Println(err)
// Output:
// ❌ Deployment failed in Network Configuration phase
//    Cause: connection timeout
//
//    Suggested actions:
//    1. Check your network connection
//    2. Verify firewall rules
//    3. Try again with --retry
```

### Specialized Constructors

The package provides specialized constructors for common deployment scenarios:

#### Database Provisioning Errors

```go
err := deployerrors.NewDatabaseProvisioningError(
    errors.New("API rate limit exceeded"),
    "cockroachdb",
)
```

#### Cache Provisioning Errors

```go
err := deployerrors.NewCacheProvisioningError(
    errors.New("insufficient permissions"),
    "valkey",
)
```

#### Application Deployment Errors

```go
err := deployerrors.NewApplicationDeploymentError(
    errors.New("build failed"),
)
```

#### Health Check Errors

```go
err := deployerrors.NewHealthCheckError(
    errors.New("timeout waiting for /healthz"),
    "https://myapp.example.com",
)
```

#### Migration Errors

```go
err := deployerrors.NewMigrationError(
    errors.New("syntax error in migration file"),
)
```

#### Pre-flight Check Errors

```go
err := deployerrors.NewPreflightCheckError(
    errors.New("missing required environment variable"),
)
```

#### Rollback Errors

```go
err := deployerrors.NewRollbackError(
    errors.New("deployment failed"),
    errors.New("rollback failed"),
)
```

## Error Wrapping

`DeploymentError` implements the `Unwrap()` method, making it compatible with Go's error wrapping:

```go
originalErr := errors.New("disk full")
deployErr := deployerrors.NewApplicationDeploymentError(originalErr)

// Check if the error is a specific type
if errors.Is(deployErr, originalErr) {
    fmt.Println("Found the original error")
}

// Unwrap to get the original error
unwrapped := errors.Unwrap(deployErr)
```

## Features

- **User-Friendly Formatting**: Errors are formatted with emojis and clear structure
- **Actionable Suggestions**: Each error includes specific steps to resolve the issue
- **Error Wrapping**: Compatible with Go's `errors.Is` and `errors.As`
- **Type Safety**: Specialized constructors ensure consistent error messages
- **Documentation Links**: All errors include links to troubleshooting documentation

## Design Philosophy

The error messages follow these principles:

1. **Clear Phase Identification**: Users immediately know which deployment phase failed
2. **Root Cause Transparency**: The underlying error is always visible
3. **Actionable Guidance**: Suggested actions are specific and executable
4. **Progressive Disclosure**: Start with simple fixes, escalate to documentation/support
5. **Consistent Format**: All errors follow the same structure for predictability

## Testing

The package includes comprehensive tests covering:
- Error message formatting
- Error wrapping and unwrapping
- All specialized constructors
- Edge cases (empty actions, nil causes, etc.)

Run tests with:
```bash
go test ./internal/errors/...
```
