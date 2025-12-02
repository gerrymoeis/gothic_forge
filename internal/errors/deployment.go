// Package errors provides error types and utilities for Gothic Forge operations.
package errors

import (
	"fmt"
	"strings"
)

// DeploymentError represents an error that occurred during deployment.
// It provides structured information about the failure phase, underlying cause,
// and actionable suggestions for resolution.
type DeploymentError struct {
	// Phase indicates which deployment phase failed (e.g., "Database Provisioning", "Application Deployment")
	Phase string

	// Cause is the underlying error that triggered the failure
	Cause error

	// Actions contains a list of suggested actions the user can take to resolve the error
	Actions []string
}

// Error implements the error interface and formats the deployment error
// with the phase, cause, and suggested actions in a user-friendly format.
func (e *DeploymentError) Error() string {
	var b strings.Builder

	// Header with phase information
	b.WriteString(fmt.Sprintf("❌ Deployment failed in %s phase\n", e.Phase))

	// Underlying cause
	b.WriteString(fmt.Sprintf("   Cause: %s\n\n", e.Cause))

	// Suggested actions (if any)
	if len(e.Actions) > 0 {
		b.WriteString("   Suggested actions:\n")
		for i, action := range e.Actions {
			b.WriteString(fmt.Sprintf("   %d. %s\n", i+1, action))
		}
	}

	return b.String()
}

// Unwrap returns the underlying cause error, allowing error unwrapping
// with errors.Is and errors.As.
func (e *DeploymentError) Unwrap() error {
	return e.Cause
}

// NewDeploymentError creates a new DeploymentError with the specified phase,
// cause, and suggested actions.
func NewDeploymentError(phase string, cause error, actions ...string) *DeploymentError {
	return &DeploymentError{
		Phase:   phase,
		Cause:   cause,
		Actions: actions,
	}
}

// Common deployment error constructors for frequently encountered scenarios

// NewDatabaseProvisioningError creates a deployment error for database provisioning failures.
func NewDatabaseProvisioningError(cause error, provider string) *DeploymentError {
	actions := []string{
		fmt.Sprintf("Check your %s API credentials are valid", strings.ToUpper(provider)),
		"Verify your account has sufficient quota",
		"Try again with: gforge deploy --live --retry",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Database Provisioning", cause, actions...)
}

// NewCacheProvisioningError creates a deployment error for cache provisioning failures.
func NewCacheProvisioningError(cause error, provider string) *DeploymentError {
	actions := []string{
		fmt.Sprintf("Check your %s API credentials are valid", strings.ToUpper(provider)),
		"Verify your account has sufficient quota",
		"Try again with: gforge deploy --live --retry",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Cache Provisioning", cause, actions...)
}

// NewApplicationDeploymentError creates a deployment error for application deployment failures.
func NewApplicationDeploymentError(cause error) *DeploymentError {
	actions := []string{
		"Check your application builds successfully: gforge build",
		"Verify all environment variables are set correctly",
		"Check deployment logs: gforge logs",
		"Try again with: gforge deploy --live --retry",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Application Deployment", cause, actions...)
}

// NewHealthCheckError creates a deployment error for health check failures.
func NewHealthCheckError(cause error, url string) *DeploymentError {
	actions := []string{
		fmt.Sprintf("Check if the application is accessible: curl %s/healthz", url),
		"Verify database and cache connections are working",
		"Check application logs: gforge logs",
		"Increase health check timeout with: --health-timeout=5m",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Health Check", cause, actions...)
}

// NewMigrationError creates a deployment error for database migration failures.
func NewMigrationError(cause error) *DeploymentError {
	actions := []string{
		"Check your migration files for syntax errors",
		"Verify database connection string is correct",
		"Test migrations locally: gforge db migrate",
		"Rollback and try again: gforge db rollback && gforge deploy --live",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Database Migration", cause, actions...)
}

// NewPreflightCheckError creates a deployment error for pre-flight check failures.
func NewPreflightCheckError(cause error) *DeploymentError {
	actions := []string{
		"Run diagnostics: gforge doctor",
		"Verify all required environment variables are set",
		"Check your project builds: gforge build",
		"Fix any issues and try again",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return NewDeploymentError("Pre-flight Check", cause, actions...)
}

// NewRollbackError creates a deployment error for rollback failures.
// This is a critical error as it means both deployment and rollback failed.
func NewRollbackError(deploymentErr, rollbackErr error) *DeploymentError {
	actions := []string{
		"⚠️  CRITICAL: Both deployment and rollback failed",
		"Manual intervention may be required",
		"Check the status of your infrastructure: gforge status",
		"Contact support immediately if production is affected",
		"Get help: https://docs.gothicforge.dev/troubleshooting",
	}
	return &DeploymentError{
		Phase:   "Rollback",
		Cause:   fmt.Errorf("deployment failed: %w, rollback failed: %v", deploymentErr, rollbackErr),
		Actions: actions,
	}
}
