package cmd

import (
	"fmt"
	"os"
	"strings"

	"gothicforge3/internal/env"
)

// DeployConfig holds all configuration for the Opinionated Stack deployment.
type DeployConfig struct {
	// Provider is always "leapcell" for the Opinionated Stack
	Provider string

	// Database (CockroachDB)
	CockroachAPIKey string
	DatabaseURL     string

	// Cache (Valkey)
	ValkeyURL string

	// Compute (Leapcell)
	LeapcellAppURL string

	// CDN (Cloudflare)
	CloudflareToken string
	CloudflareAcct  string
	CFProjectName   string

	// Application
	SiteBaseURL string
	JWTSecret   string
	AppEnv      string
}

// LoadDeployConfig loads deployment configuration from environment variables,
// supporting both new standardized names and deprecated legacy names.
func LoadDeployConfig() (*DeployConfig, error) {
	cfg := &DeployConfig{
		Provider:        "leapcell",
		CockroachAPIKey: strings.TrimSpace(os.Getenv("COCKROACH_API_KEY")),
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		ValkeyURL:       env.GetWithDeprecation("VALKEY_URL", "REDIS_URL", "cache"),
		LeapcellAppURL:  strings.TrimSpace(os.Getenv("LEAPCELL_APP_URL")),
		CloudflareToken: env.GetWithDeprecation("CLOUDFLARE_API_TOKEN", "CF_API_TOKEN", "Cloudflare"),
		CloudflareAcct:  env.GetWithDeprecation("CLOUDFLARE_ACCOUNT_ID", "CF_ACCOUNT_ID", "Cloudflare"),
		CFProjectName:   strings.TrimSpace(os.Getenv("CF_PROJECT_NAME")),
		SiteBaseURL:     strings.TrimSpace(os.Getenv("SITE_BASE_URL")),
		JWTSecret:       strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AppEnv:          strings.TrimSpace(os.Getenv("APP_ENV")),
	}

	return cfg, cfg.Validate()
}

// Validate checks that all required configuration is present and valid.
func (c *DeployConfig) Validate() error {
	var missing []string

	// Database: require either API key for provisioning or connection string
	if c.CockroachAPIKey == "" && c.DatabaseURL == "" {
		missing = append(missing, "COCKROACH_API_KEY or DATABASE_URL")
	}

	// JWT Secret: require strong secret (>= 32 chars)
	if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
		missing = append(missing, "JWT_SECRET (>=32 chars)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	return nil
}
