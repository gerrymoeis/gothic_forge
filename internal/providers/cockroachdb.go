package providers

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const cockroachAPIBase = "https://cockroachlabs.cloud/api/v1"

// retryConfig defines the configuration for retry logic with exponential backoff
type retryConfig struct {
	maxRetries      int
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	backoffMultiplier float64
}

// defaultRetryConfig returns the default retry configuration
func defaultRetryConfig() retryConfig {
	return retryConfig{
		maxRetries:      3,
		initialBackoff:  1 * time.Second,
		maxBackoff:      10 * time.Second,
		backoffMultiplier: 2.0,
	}
}

// retryWithBackoff executes a function with exponential backoff retry logic.
// It returns the error from the last attempt if all retries fail.
func retryWithBackoff(ctx context.Context, cfg retryConfig, operation func() error) error {
	var lastErr error
	backoff := cfg.initialBackoff

	for attempt := 0; attempt < cfg.maxRetries; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success!
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt < cfg.maxRetries-1 {
			// Check if context is still valid before sleeping
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry (attempt %d/%d): %w", attempt+1, cfg.maxRetries, ctx.Err())
			case <-time.After(backoff):
				// Calculate next backoff with exponential growth
				backoff = time.Duration(float64(backoff) * cfg.backoffMultiplier)
				if backoff > cfg.maxBackoff {
					backoff = cfg.maxBackoff
				}
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", cfg.maxRetries, lastErr)
}

// CockroachDBProvider implements DatabaseProvider for CockroachDB Serverless.
//
// This provider supports automatic cluster provisioning via the CockroachDB Cloud API.
// It can also use existing clusters if DATABASE_URL is already set.
//
// The provider includes:
// - Automatic cluster provisioning via CockroachDB Cloud API
// - Connection string management
// - Database migration support using goose
// - Health checks with timeout
// - Retry logic with exponential backoff for migrations
// - PostgreSQL-compatible operations (CockroachDB is PostgreSQL-compatible)
type CockroachDBProvider struct {
	apiKey         string
	clusterName    string
	region         string
	dbName         string
	dbUser         string
	connectionStr  string
	clusterID      string
	httpClient     *http.Client
}

// NewCockroachDBProvider creates a new CockroachDB provider
func NewCockroachDBProvider(apiKey string) *CockroachDBProvider {
	return &CockroachDBProvider{
		apiKey:      apiKey,
		clusterName: getEnvOrDefault("COCKROACH_CLUSTER_NAME", "gothic-forge-db"),
		region:      getEnvOrDefault("COCKROACH_REGION", "us-east-1"),
		dbName:      getEnvOrDefault("COCKROACH_DB_NAME", "app"),
		dbUser:      getEnvOrDefault("COCKROACH_DB_USER", "gothicforge"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider name
func (p *CockroachDBProvider) Name() string {
	return "cockroachdb"
}

// Provision creates a new CockroachDB Serverless cluster via the CockroachDB Cloud API
func (p *CockroachDBProvider) Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error) {
	// Check if DATABASE_URL is already set (cluster already provisioned)
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		p.connectionStr = connStr
		
		// Test the connection
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		return &DatabaseInfo{
			ID:               "cockroachdb-existing",
			Name:             opts.Name,
			ConnectionString: connStr,
			Region:           p.region,
			CreatedAt:        time.Now(),
			Metadata: map[string]interface{}{
				"provider": "cockroachdb",
				"type":     "serverless",
			},
		}, nil
	}

	// Validate API key
	if p.apiKey == "" {
		return nil, fmt.Errorf("COCKROACH_API_KEY not set. Please set your CockroachDB API key")
	}

	// Use provided name or default
	clusterName := p.clusterName
	if opts.Name != "" {
		clusterName = opts.Name
	}

	// Use provided region or default
	region := p.region
	if opts.Region != "" {
		region = opts.Region
	}

	// Step 1: Create the serverless cluster
	cluster, err := p.createServerlessCluster(ctx, clusterName, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}
	p.clusterID = cluster.ID

	// Step 2: Wait for cluster to be ready
	if err := p.waitForClusterReady(ctx, cluster.ID); err != nil {
		return nil, fmt.Errorf("cluster failed to become ready: %w", err)
	}

	// Step 3: Create SQL user
	password, err := p.createSQLUser(ctx, cluster.ID, p.dbUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL user: %w", err)
	}

	// Step 4: Create database
	if err := p.createDatabase(ctx, cluster.ID, p.dbName); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Step 5: Build connection string
	if len(cluster.Regions) == 0 {
		return nil, fmt.Errorf("cluster has no regions configured")
	}
	
	host := cluster.Regions[0].SQLDns
	pwdEsc := url.QueryEscape(password)
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:26257/%s?sslmode=verify-full",
		p.dbUser, pwdEsc, host, p.dbName)
	
	// Add routing ID if available
	if rid := strings.TrimSpace(cluster.Config.Serverless.RoutingID); rid != "" {
		connStr += "&options=--cluster=" + rid
	}

	p.connectionStr = connStr

	return &DatabaseInfo{
		ID:               cluster.ID,
		Name:             cluster.Name,
		ConnectionString: connStr,
		Region:           region,
		CreatedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"provider":         "cockroachdb",
			"type":             "serverless",
			"cluster_id":       cluster.ID,
			"routing_id":       cluster.Config.Serverless.RoutingID,
			"cockroach_version": cluster.CockroachVersion,
			"state":            cluster.State,
		},
	}, nil
}

// GetConnectionString returns the CockroachDB connection string
func (p *CockroachDBProvider) GetConnectionString(ctx context.Context) (string, error) {
	// Check if we have a cached connection string
	if p.connectionStr != "" {
		return p.connectionStr, nil
	}

	// Check DATABASE_URL environment variable
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return "", fmt.Errorf("DATABASE_URL not set. Please provision a CockroachDB cluster and set the connection string")
	}

	p.connectionStr = connStr
	return connStr, nil
}

// RunMigrations executes database migrations using goose with retry logic
func (p *CockroachDBProvider) RunMigrations(ctx context.Context, migrationsDir string) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set goose dialect to postgres (CockroachDB is PostgreSQL-compatible)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations with retry logic (CockroachDB may have transient errors)
	cfg := defaultRetryConfig()
	return retryWithBackoff(ctx, cfg, func() error {
		return goose.Up(db, migrationsDir)
	})
}

// Health checks if the database is reachable and healthy
func (p *CockroachDBProvider) Health(ctx context.Context) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Run a simple query to verify database is operational
	var result int
	if err := db.QueryRowContext(pingCtx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// Destroy removes the CockroachDB cluster via the CockroachDB Cloud API
// This method will only delete clusters that were provisioned via the API (have a clusterID).
// For safety, it will not attempt to delete clusters that were manually created or
// connected via DATABASE_URL without provisioning through this provider.
//
// The method includes retry logic with exponential backoff to handle transient API failures.
func (p *CockroachDBProvider) Destroy(ctx context.Context) error {
	// Check if we have a cluster ID (meaning we provisioned it)
	if p.clusterID == "" {
		return fmt.Errorf("no cluster ID available. This cluster was not provisioned by this provider or was manually created. Please delete manually at https://cockroachlabs.cloud/")
	}

	// Validate API key
	if p.apiKey == "" {
		return fmt.Errorf("COCKROACH_API_KEY not set. Cannot delete cluster without API credentials")
	}

	// Delete the cluster via API with retry logic
	cfg := defaultRetryConfig()
	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "DELETE", "/clusters/"+p.clusterID, nil, nil)
	})

	if err == nil {
		// Success! Clear the cluster ID and connection string
		p.clusterID = ""
		p.connectionStr = ""
	}

	return err
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CockroachDB API types
type cockroachCluster struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	CockroachVersion string                   `json:"cockroach_version"`
	Plan             string                   `json:"plan"`
	CloudProvider    string                   `json:"cloud_provider"`
	State            string                   `json:"state"`
	Config           cockroachClusterConfig   `json:"config"`
	Regions          []cockroachClusterRegion `json:"regions"`
}

type cockroachClusterConfig struct {
	Serverless cockroachServerlessConfig `json:"serverless"`
}

type cockroachServerlessConfig struct {
	SpendLimit int    `json:"spend_limit"`
	RoutingID  string `json:"routing_id"`
}

type cockroachClusterRegion struct {
	Name   string `json:"name"`
	SQLDns string `json:"sql_dns"`
	UIDns  string `json:"ui_dns"`
}

// doAPIRequest performs an HTTP request to the CockroachDB Cloud API
func (p *CockroachDBProvider) doAPIRequest(ctx context.Context, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, cockroachAPIBase+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("X-CC-API-KEY", p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cockroachdb api %s %s: %s: %s", method, path, resp.Status, string(b))
	}

	if out != nil && resp.StatusCode != 204 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// createServerlessCluster creates a new serverless cluster via the API with retry logic
func (p *CockroachDBProvider) createServerlessCluster(ctx context.Context, name, region string) (*cockroachCluster, error) {
	payload := map[string]interface{}{
		"name":     name,
		"provider": "AWS",
		"spec": map[string]interface{}{
			"serverless": map[string]interface{}{
				"regions":     []string{region},
				"spend_limit": 0, // Free tier
			},
		},
	}

	var result cockroachCluster
	cfg := defaultRetryConfig()
	
	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "POST", "/clusters", payload, &result)
	})
	
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// waitForClusterReady polls the cluster status until it's ready
func (p *CockroachDBProvider) waitForClusterReady(ctx context.Context, clusterID string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Use a smaller retry config for polling operations
	pollRetryConfig := retryConfig{
		maxRetries:      2,
		initialBackoff:  500 * time.Millisecond,
		maxBackoff:      2 * time.Second,
		backoffMultiplier: 2.0,
	}

	for {
		select {
		case <-timeout:
			return errors.New("timeout waiting for cluster to be ready")
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var cluster cockroachCluster
			
			// Use retry logic for the API call to handle transient failures
			err := retryWithBackoff(ctx, pollRetryConfig, func() error {
				return p.doAPIRequest(ctx, "GET", "/clusters/"+clusterID, nil, &cluster)
			})
			
			if err != nil {
				// If we can't get cluster status after retries, return error
				return fmt.Errorf("failed to get cluster status: %w", err)
			}
			
			if cluster.State == "CREATED" {
				return nil
			}
			
			// Check for failure states
			if strings.Contains(strings.ToUpper(cluster.State), "FAILED") {
				return fmt.Errorf("cluster creation failed: %s", cluster.State)
			}
		}
	}
}

// createSQLUser creates a SQL user and returns the generated password with retry logic
func (p *CockroachDBProvider) createSQLUser(ctx context.Context, clusterID, username string) (string, error) {
	// Generate a strong password
	password := generateStrongPassword(24)
	if len(password) < 12 {
		password = generateStrongPassword(32)
	}

	payload := map[string]string{
		"name":     username,
		"password": password,
	}

	var result struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	cfg := defaultRetryConfig()
	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "POST", "/clusters/"+clusterID+"/sql-users", payload, &result)
	})
	
	if err != nil {
		return "", err
	}

	// Some API responses may omit echoing the password; fallback to our generated one
	if strings.TrimSpace(result.Password) != "" {
		return result.Password, nil
	}
	return password, nil
}

// createDatabase creates a database in the cluster with retry logic
func (p *CockroachDBProvider) createDatabase(ctx context.Context, clusterID, dbName string) error {
	payload := map[string]string{
		"name": dbName,
	}

	cfg := defaultRetryConfig()
	return retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "POST", "/clusters/"+clusterID+"/databases", payload, nil)
	})
}

// generateStrongPassword generates a cryptographically secure password
func generateStrongPassword(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
