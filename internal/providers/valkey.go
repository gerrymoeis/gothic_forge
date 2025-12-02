package providers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const aivenAPIBase = "https://api.aiven.io/v1"

// retryConfig contains retry configuration
type retryConfig struct {
	maxRetries int
	delay      time.Duration
	maxDelay   time.Duration
}

// defaultRetryConfig returns the default retry configuration
func defaultRetryConfig() retryConfig {
	return retryConfig{
		maxRetries: 3,
		delay:      time.Second,
		maxDelay:   10 * time.Second,
	}
}

// retryWithBackoff executes a function with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, cfg retryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.delay

	for i := 0; i <= cfg.maxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Exponential backoff
				delay *= 2
				if delay > cfg.maxDelay {
					delay = cfg.maxDelay
				}
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d retries: %w", cfg.maxRetries, lastErr)
}

// ValkeyProvider implements CacheProvider for Valkey (Redis-compatible cache via Aiven).
//
// This provider supports automatic Valkey instance provisioning via the Aiven API.
// It can also use existing instances if VALKEY_URL or REDIS_URL is already set.
//
// The provider includes:
// - Automatic Valkey instance provisioning via Aiven API
// - Connection string management
// - Health checks with PING command
// - TLS support for secure connections
// - Retry logic with exponential backoff
// - Redis-compatible operations (Valkey is Redis-compatible)
type ValkeyProvider struct {
	apiToken       string
	project        string
	serviceName    string
	cloud          string
	plan           string
	connectionStr  string
	httpClient     *http.Client
}

// NewValkeyProvider creates a new Valkey provider
func NewValkeyProvider(apiToken string) *ValkeyProvider {
	return &ValkeyProvider{
		apiToken:    apiToken,
		project:     getEnvOrDefault("AIVEN_PROJECT", "gothic-forge"),
		serviceName: getEnvOrDefault("VALKEY_SERVICE_NAME", "gothic-forge-cache"),
		cloud:       getEnvOrDefault("AIVEN_CLOUD", "aws-us-east-1"),
		plan:        getEnvOrDefault("AIVEN_PLAN", "startup-4"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider name
func (p *ValkeyProvider) Name() string {
	return "valkey"
}

// Provision creates a new Valkey instance via the Aiven API
func (p *ValkeyProvider) Provision(ctx context.Context, opts ProvisionOptions) (*CacheInfo, error) {
	// Check if VALKEY_URL or REDIS_URL is already set (instance already provisioned)
	connStr := os.Getenv("VALKEY_URL")
	if connStr == "" {
		connStr = os.Getenv("REDIS_URL") // Support deprecated REDIS_URL
	}

	if connStr != "" {
		p.connectionStr = connStr

		// Test the connection with proper TLS support
		client := createRedisClient(connStr)
		defer client.Close()

		if err := client.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to ping cache: %w", err)
		}

		return &CacheInfo{
			ID:               "valkey-existing",
			Name:             opts.Name,
			ConnectionString: connStr,
			Region:           p.cloud,
			CreatedAt:        time.Now(),
			Metadata: map[string]interface{}{
				"provider": "valkey",
				"type":     "existing",
			},
		}, nil
	}

	// Validate API token
	if p.apiToken == "" {
		return nil, fmt.Errorf("AIVEN_TOKEN not set. Please set your Aiven API token")
	}

	// Use provided name or default
	serviceName := p.serviceName
	if opts.Name != "" {
		serviceName = opts.Name
	}

	// Use provided region or default
	cloud := p.cloud
	if opts.Region != "" {
		cloud = opts.Region
	}

	// Use provided tier or default
	plan := p.plan
	if opts.Tier != "" {
		plan = opts.Tier
	}

	// Step 1: Create the Valkey service
	service, err := p.createValkeyService(ctx, serviceName, cloud, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to create Valkey service: %w", err)
	}

	// Step 2: Wait for service to be ready
	if err := p.waitForServiceReady(ctx, service.ServiceName); err != nil {
		return nil, fmt.Errorf("service failed to become ready: %w", err)
	}

	// Step 3: Get connection info
	connInfo, err := p.getConnectionInfo(ctx, service.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection info: %w", err)
	}

	p.connectionStr = connInfo.ServiceURI

	return &CacheInfo{
		ID:               service.ServiceName,
		Name:             service.ServiceName,
		ConnectionString: connInfo.ServiceURI,
		Host:             connInfo.Host,
		Port:             connInfo.Port,
		Password:         connInfo.Password,
		Region:           cloud,
		CreatedAt:        time.Now(),
		Metadata: map[string]interface{}{
			"provider":     "valkey",
			"type":         "aiven",
			"service_name": service.ServiceName,
			"plan":         service.Plan,
			"cloud":        service.CloudName,
			"state":        service.State,
		},
	}, nil
}

// GetConnectionString returns the Valkey connection string
func (p *ValkeyProvider) GetConnectionString(ctx context.Context) (string, error) {
	// Check if we have a cached connection string
	if p.connectionStr != "" {
		return p.connectionStr, nil
	}

	// Check VALKEY_URL environment variable
	connStr := os.Getenv("VALKEY_URL")
	if connStr == "" {
		// Fall back to deprecated REDIS_URL
		connStr = os.Getenv("REDIS_URL")
	}

	if connStr == "" {
		return "", fmt.Errorf("VALKEY_URL not set. Please provision a Valkey instance and set the connection string")
	}

	p.connectionStr = connStr
	return connStr, nil
}

// Health checks if the cache is reachable using PING command
func (p *ValkeyProvider) Health(ctx context.Context) error {
	connStr, err := p.GetConnectionString(ctx)
	if err != nil {
		return err
	}

	// Create Redis client with proper TLS support
	client := createRedisClient(connStr)
	defer client.Close()

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}

// Destroy removes the Valkey service via the Aiven API
func (p *ValkeyProvider) Destroy(ctx context.Context) error {
	// Check if we have a service name (meaning we provisioned it)
	if p.serviceName == "" {
		return fmt.Errorf("no service name available. This service was not provisioned by this provider or was manually created")
	}

	// Validate API token
	if p.apiToken == "" {
		return fmt.Errorf("AIVEN_TOKEN not set. Cannot delete service without API credentials")
	}

	// Delete the service via API with retry logic
	cfg := defaultRetryConfig()
	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "DELETE", fmt.Sprintf("/project/%s/service/%s", p.project, p.serviceName), nil, nil)
	})

	if err == nil {
		// Success! Clear the service name and connection string
		p.serviceName = ""
		p.connectionStr = ""
	}

	return err
}

// Aiven API types
type aivenService struct {
	ServiceName string                 `json:"service_name"`
	ServiceType string                 `json:"service_type"`
	Plan        string                 `json:"plan"`
	CloudName   string                 `json:"cloud_name"`
	State       string                 `json:"state"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type aivenServiceResponse struct {
	Service aivenService `json:"service"`
}

type aivenConnectionInfo struct {
	ServiceURI string `json:"service_uri"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
}

// doAPIRequest performs an HTTP request to the Aiven API
func (p *ValkeyProvider) doAPIRequest(ctx context.Context, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, aivenAPIBase+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("aiven api %s %s: %s: %s", method, path, resp.Status, string(b))
	}

	if out != nil && resp.StatusCode != 204 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// createValkeyService creates a new Valkey service via the API with retry logic
func (p *ValkeyProvider) createValkeyService(ctx context.Context, serviceName, cloud, plan string) (*aivenService, error) {
	payload := map[string]interface{}{
		"service_name": serviceName,
		"service_type": "redis", // Aiven uses "redis" for Valkey-compatible services
		"plan":         plan,
		"cloud":        cloud,
	}

	var result aivenServiceResponse
	cfg := defaultRetryConfig()

	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "POST", fmt.Sprintf("/project/%s/service", p.project), payload, &result)
	})

	if err != nil {
		return nil, err
	}

	return &result.Service, nil
}

// waitForServiceReady polls the service status until it's ready
func (p *ValkeyProvider) waitForServiceReady(ctx context.Context, serviceName string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Use a smaller retry config for polling operations
	pollRetryConfig := retryConfig{
		maxRetries: 2,
		delay:      500 * time.Millisecond,
		maxDelay:   2 * time.Second,
	}

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for service to be ready")
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var result aivenServiceResponse

			// Use retry logic for the API call to handle transient failures
			err := retryWithBackoff(ctx, pollRetryConfig, func() error {
				return p.doAPIRequest(ctx, "GET", fmt.Sprintf("/project/%s/service/%s", p.project, serviceName), nil, &result)
			})

			if err != nil {
				// If we can't get service status after retries, return error
				return fmt.Errorf("failed to get service status: %w", err)
			}

			if result.Service.State == "RUNNING" {
				return nil
			}

			// Check for failure states
			if strings.Contains(strings.ToUpper(result.Service.State), "FAILED") {
				return fmt.Errorf("service creation failed: %s", result.Service.State)
			}
		}
	}
}

// getConnectionInfo retrieves connection information for the service
func (p *ValkeyProvider) getConnectionInfo(ctx context.Context, serviceName string) (*aivenConnectionInfo, error) {
	// Define a more complete response structure to capture connection info
	type aivenServiceDetailResponse struct {
		Service struct {
			ServiceName string `json:"service_name"`
			ServiceType string `json:"service_type"`
			State       string `json:"state"`
			ServiceURI  string `json:"service_uri"`
			Components  []struct {
				Component string `json:"component"`
				Host      string `json:"host"`
				Port      int    `json:"port"`
				Route     string `json:"route"`
			} `json:"components"`
			ConnectionInfo struct {
				RedisURI string `json:"redis_uri,omitempty"`
			} `json:"connection_info"`
		} `json:"service"`
	}

	var result aivenServiceDetailResponse

	cfg := defaultRetryConfig()
	err := retryWithBackoff(ctx, cfg, func() error {
		return p.doAPIRequest(ctx, "GET", fmt.Sprintf("/project/%s/service/%s", p.project, serviceName), nil, &result)
	})

	if err != nil {
		return nil, err
	}

	// Extract connection info from service response
	connInfo := &aivenConnectionInfo{
		ServiceURI: result.Service.ServiceURI,
	}

	// Try to get connection URI from connection_info first
	if result.Service.ConnectionInfo.RedisURI != "" {
		connInfo.ServiceURI = result.Service.ConnectionInfo.RedisURI
	}

	// Extract host and port from components if available
	for _, comp := range result.Service.Components {
		if comp.Component == "redis" {
			connInfo.Host = comp.Host
			connInfo.Port = comp.Port
			break
		}
	}

	// If we still don't have a service URI, construct one from components
	if connInfo.ServiceURI == "" && connInfo.Host != "" {
		connInfo.ServiceURI = fmt.Sprintf("redis://%s:%d", connInfo.Host, connInfo.Port)
	}

	// Validate we got a connection string
	if connInfo.ServiceURI == "" {
		return nil, fmt.Errorf("failed to extract connection URI from service response")
	}

	return connInfo, nil
}

// parseRedisAddr extracts the address from a Redis connection string
func parseRedisAddr(connStr string) string {
	// Handle redis:// or rediss:// URLs
	connStr = strings.TrimPrefix(connStr, "redis://")
	connStr = strings.TrimPrefix(connStr, "rediss://")

	// Remove credentials if present (use LastIndex to handle @ in passwords)
	if idx := strings.LastIndex(connStr, "@"); idx != -1 {
		connStr = connStr[idx+1:]
	}

	// Remove path if present
	if idx := strings.Index(connStr, "/"); idx != -1 {
		connStr = connStr[:idx]
	}

	return connStr
}

// parseRedisPassword extracts the password from a Redis connection string
func parseRedisPassword(connStr string) string {
	// Handle redis:// or rediss:// URLs
	connStr = strings.TrimPrefix(connStr, "redis://")
	connStr = strings.TrimPrefix(connStr, "rediss://")

	// Extract credentials if present (format: user:pass@host or :pass@host)
	// Find the last @ to handle passwords with @ in them
	if idx := strings.LastIndex(connStr, "@"); idx != -1 {
		credentials := connStr[:idx]
		// Check if there's a colon (password present)
		if colonIdx := strings.Index(credentials, ":"); colonIdx != -1 {
			return credentials[colonIdx+1:]
		}
	}

	return ""
}

// isTLSConnection checks if the connection string uses TLS (rediss://)
func isTLSConnection(connStr string) bool {
	return strings.HasPrefix(connStr, "rediss://")
}

// createRedisClient creates a Redis client with proper TLS configuration
func createRedisClient(connStr string) *redis.Client {
	opts := &redis.Options{
		Addr: parseRedisAddr(connStr),
	}

	// Extract password if present
	if password := parseRedisPassword(connStr); password != "" {
		opts.Password = password
	}

	// Enable TLS if using rediss://
	if isTLSConnection(connStr) {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// InsecureSkipVerify is set to false by default, which is secure
			// For production, certificates should be properly validated
		}
	}

	return redis.NewClient(opts)
}
