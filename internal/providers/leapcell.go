package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const leapcellAPIBase = "https://api.leapcell.io/v1"

// LeapcellProvider implements ComputeProvider for Leapcell deployments.
//
// Leapcell provides simple, fast container deployment with automatic scaling
// and GitHub integration. This provider supports:
// - Docker-based application deployment
// - Automatic scaling
// - Health checks and monitoring
// - Log streaming
// - Rollback capabilities
//
// The provider can work in two modes:
// 1. API-based deployment (requires LEAPCELL_API_KEY)
// 2. Git-based deployment (requires Git push to Leapcell remote)
type LeapcellProvider struct {
	apiKey      string
	appID       string
	appURL      string
	projectName string
	httpClient  *http.Client
}

// NewLeapcellProvider creates a new Leapcell provider
func NewLeapcellProvider(apiKey string) *LeapcellProvider {
	return &LeapcellProvider{
		apiKey:      apiKey,
		appURL:      os.Getenv("LEAPCELL_APP_URL"),
		projectName: getEnvOrDefault("LEAPCELL_PROJECT_NAME", "gothic-forge-app"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider name
func (p *LeapcellProvider) Name() string {
	return "leapcell"
}

// Deploy deploys the application to Leapcell
func (p *LeapcellProvider) Deploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error) {
	// Check if LEAPCELL_APP_URL is already set (app already exists)
	if p.appURL != "" {
		// Use existing app - trigger a new deployment
		return p.deployToExistingApp(ctx, opts)
	}

	// Check if we have API key for creating new app
	if p.apiKey == "" {
		return nil, fmt.Errorf("LEAPCELL_API_KEY not set. Please set your Leapcell API key or LEAPCELL_APP_URL for existing app")
	}

	// Create new app and deploy
	return p.createAndDeploy(ctx, opts)
}

// deployToExistingApp deploys to an existing Leapcell app
func (p *LeapcellProvider) deployToExistingApp(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error) {
	// For existing apps, we assume deployment happens via Git push
	// or the app is already configured. We just return the deployment info.
	
	// Extract app ID from URL if possible
	appID := p.extractAppIDFromURL(p.appURL)
	
	return &DeploymentInfo{
		ID:           appID,
		Name:         opts.Name,
		URL:          p.appURL,
		DashboardURL: "https://leapcell.io/dashboard",
		Version:      "latest",
		Status:       "deployed",
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"provider":     "leapcell",
			"app_url":      p.appURL,
			"deploy_type":  "existing",
		},
	}, nil
}

// createAndDeploy creates a new Leapcell app and deploys to it
func (p *LeapcellProvider) createAndDeploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error) {
	// Step 1: Create the app
	app, err := p.createApp(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}
	
	p.appID = app.ID
	p.appURL = app.URL

	// Step 2: Configure environment variables
	if len(opts.EnvVars) > 0 {
		if err := p.setEnvironmentVariables(ctx, app.ID, opts.EnvVars); err != nil {
			return nil, fmt.Errorf("failed to set environment variables: %w", err)
		}
	}

	// Step 3: Deploy the application
	deployment, err := p.triggerDeployment(ctx, app.ID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger deployment: %w", err)
	}

	// Step 4: Wait for deployment to complete
	if err := p.waitForDeployment(ctx, app.ID, deployment.ID); err != nil {
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	return &DeploymentInfo{
		ID:           deployment.ID,
		Name:         app.Name,
		URL:          app.URL,
		DashboardURL: fmt.Sprintf("https://leapcell.io/dashboard/apps/%s", app.ID),
		Version:      deployment.Version,
		Status:       "deployed",
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"provider":    "leapcell",
			"app_id":      app.ID,
			"app_url":     app.URL,
			"deploy_type": "api",
		},
	}, nil
}

// GetURL returns the application URL
func (p *LeapcellProvider) GetURL(ctx context.Context) (string, error) {
	// Check if we have a cached URL
	if p.appURL != "" {
		return p.appURL, nil
	}

	// Check environment variable
	appURL := os.Getenv("LEAPCELL_APP_URL")
	if appURL == "" {
		return "", fmt.Errorf("LEAPCELL_APP_URL not set. Please deploy your app first or set the URL manually")
	}

	p.appURL = appURL
	return appURL, nil
}

// GetLogs streams application logs from Leapcell
func (p *LeapcellProvider) GetLogs(ctx context.Context, opts LogOptions) (io.ReadCloser, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("LEAPCELL_API_KEY not set. Cannot retrieve logs without API credentials")
	}

	appID := p.appID
	if appID == "" {
		// Try to extract from URL
		appID = p.extractAppIDFromURL(p.appURL)
		if appID == "" {
			return nil, fmt.Errorf("app ID not available. Please deploy first")
		}
	}

	// Build query parameters
	params := ""
	if opts.Tail > 0 {
		params += fmt.Sprintf("?tail=%d", opts.Tail)
	}
	if opts.Follow {
		if params == "" {
			params = "?follow=true"
		} else {
			params += "&follow=true"
		}
	}

	// Create request for log streaming
	req, err := http.NewRequestWithContext(ctx, "GET", 
		fmt.Sprintf("%s/apps/%s/logs%s", leapcellAPIBase, appID, params), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/plain")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("leapcell api GET /apps/%s/logs: %s: %s", appID, resp.Status, string(body))
	}

	return resp.Body, nil
}

// Health checks if the application is healthy
func (p *LeapcellProvider) Health(ctx context.Context) error {
	appURL, err := p.GetURL(ctx)
	if err != nil {
		return err
	}

	// Try to hit the health endpoint
	healthURL := strings.TrimSuffix(appURL, "/") + "/healthz"
	
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// Rollback reverts to a previous deployment version
func (p *LeapcellProvider) Rollback(ctx context.Context, version string) error {
	if p.apiKey == "" {
		return fmt.Errorf("LEAPCELL_API_KEY not set. Cannot rollback without API credentials")
	}

	appID := p.appID
	if appID == "" {
		appID = p.extractAppIDFromURL(p.appURL)
		if appID == "" {
			return fmt.Errorf("app ID not available")
		}
	}

	payload := map[string]string{
		"version": version,
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/apps/%s/rollback", appID), payload, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("rollback failed: %s", result.Message)
	}

	return nil
}

// Scale adjusts the number of replicas
func (p *LeapcellProvider) Scale(ctx context.Context, replicas int) error {
	if p.apiKey == "" {
		return fmt.Errorf("LEAPCELL_API_KEY not set. Cannot scale without API credentials")
	}

	if replicas < 1 {
		return fmt.Errorf("replicas must be at least 1")
	}

	appID := p.appID
	if appID == "" {
		appID = p.extractAppIDFromURL(p.appURL)
		if appID == "" {
			return fmt.Errorf("app ID not available")
		}
	}

	payload := map[string]int{
		"replicas": replicas,
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/apps/%s/scale", appID), payload, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("scaling failed: %s", result.Message)
	}

	return nil
}

// Leapcell API types
type leapcellApp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type leapcellDeployment struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// doAPIRequest performs an HTTP request to the Leapcell API
func (p *LeapcellProvider) doAPIRequest(ctx context.Context, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, leapcellAPIBase+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("leapcell api %s %s: %s: %s", method, path, resp.Status, string(b))
	}

	if out != nil && resp.StatusCode != 204 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// createApp creates a new Leapcell app
func (p *LeapcellProvider) createApp(ctx context.Context, opts DeployOptions) (*leapcellApp, error) {
	payload := map[string]interface{}{
		"name": opts.Name,
	}

	if opts.Port > 0 {
		payload["port"] = opts.Port
	}

	var app leapcellApp
	if err := p.doAPIRequest(ctx, "POST", "/apps", payload, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// setEnvironmentVariables configures environment variables for the app
func (p *LeapcellProvider) setEnvironmentVariables(ctx context.Context, appID string, envVars map[string]string) error {
	payload := map[string]interface{}{
		"env": envVars,
	}

	return p.doAPIRequest(ctx, "PUT", fmt.Sprintf("/apps/%s/env", appID), payload, nil)
}

// triggerDeployment triggers a new deployment
func (p *LeapcellProvider) triggerDeployment(ctx context.Context, appID string, opts DeployOptions) (*leapcellDeployment, error) {
	payload := map[string]interface{}{
		"image": opts.Image,
	}

	if opts.BuildPath != "" {
		payload["build_path"] = opts.BuildPath
	}

	var deployment leapcellDeployment
	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/apps/%s/deployments", appID), payload, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// waitForDeployment waits for a deployment to complete
func (p *LeapcellProvider) waitForDeployment(ctx context.Context, appID, deploymentID string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for deployment to complete")
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var deployment leapcellDeployment
			if err := p.doAPIRequest(ctx, "GET", 
				fmt.Sprintf("/apps/%s/deployments/%s", appID, deploymentID), nil, &deployment); err != nil {
				return fmt.Errorf("failed to get deployment status: %w", err)
			}

			switch deployment.Status {
			case "deployed", "success", "running":
				return nil
			case "failed", "error":
				return fmt.Errorf("deployment failed with status: %s", deployment.Status)
			}
			// Continue polling for other statuses (pending, building, deploying, etc.)
		}
	}
}

// extractAppIDFromURL extracts the app ID from a Leapcell URL
func (p *LeapcellProvider) extractAppIDFromURL(url string) string {
	// Leapcell URLs typically follow pattern: https://<app-id>.leapcell.dev
	// or https://<custom-domain>
	if url == "" {
		return ""
	}

	// Try to extract from subdomain
	parts := strings.Split(url, ".")
	if len(parts) >= 3 && strings.Contains(url, "leapcell.dev") {
		// Extract app ID from subdomain
		subdomain := strings.TrimPrefix(parts[0], "https://")
		subdomain = strings.TrimPrefix(subdomain, "http://")
		return subdomain
	}

	// If we can't extract, return empty string
	return ""
}
