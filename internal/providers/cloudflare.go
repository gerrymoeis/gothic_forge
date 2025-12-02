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

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

// CloudflareProvider implements CDNProvider for Cloudflare Pages.
//
// Cloudflare Pages provides fast, secure static site hosting with automatic
// deployments from Git repositories. This provider supports:
// - Automatic Pages deployment via API
// - Cache invalidation (purge)
// - Zone detection and management
// - Custom domain support
//
// The provider can work in two modes:
// 1. API-based deployment (requires CLOUDFLARE_API_TOKEN)
// 2. Git-based deployment (requires Git push to Cloudflare)
type CloudflareProvider struct {
	apiToken    string
	accountID   string
	projectName string
	zoneID      string
	deployURL   string
	httpClient  *http.Client
}

// NewCloudflareProvider creates a new Cloudflare provider
func NewCloudflareProvider(apiToken string) *CloudflareProvider {
	return &CloudflareProvider{
		apiToken:    apiToken,
		accountID:   os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		projectName: getEnvOrDefault("CLOUDFLARE_PROJECT_NAME", "gothic-forge"),
		zoneID:      os.Getenv("CLOUDFLARE_ZONE_ID"),
		deployURL:   os.Getenv("CLOUDFLARE_DEPLOY_URL"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider name
func (p *CloudflareProvider) Name() string {
	return "cloudflare"
}

// Deploy deploys static assets to Cloudflare Pages
func (p *CloudflareProvider) Deploy(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error) {
	// Check if we already have a deployment URL (existing project)
	if p.deployURL != "" {
		return p.deployToExistingProject(ctx, opts)
	}

	// Check if we have API credentials for creating new project
	if p.apiToken == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN not set. Please set your Cloudflare API token")
	}

	if p.accountID == "" {
		return nil, fmt.Errorf("CLOUDFLARE_ACCOUNT_ID not set. Please set your Cloudflare account ID")
	}

	// Create new project and deploy
	return p.createAndDeploy(ctx, opts)
}

// deployToExistingProject deploys to an existing Cloudflare Pages project
func (p *CloudflareProvider) deployToExistingProject(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error) {
	// For existing projects, we assume deployment happens via Git push
	// or the project is already configured. We just return the deployment info.
	
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = p.projectName
	}

	return &CDNInfo{
		ID:        projectName,
		Name:      projectName,
		URL:       p.deployURL,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"provider":    "cloudflare",
			"deploy_url":  p.deployURL,
			"deploy_type": "existing",
		},
	}, nil
}

// createAndDeploy creates a new Cloudflare Pages project and deploys to it
func (p *CloudflareProvider) createAndDeploy(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error) {
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = p.projectName
	}

	// Step 1: Check if project already exists
	project, err := p.getProject(ctx, projectName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check project: %w", err)
	}

	// Step 2: Create project if it doesn't exist
	if project == nil {
		project, err = p.createProject(ctx, projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to create project: %w", err)
		}
	}

	// Step 3: Create a deployment
	deployment, err := p.createDeployment(ctx, projectName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// Step 4: Wait for deployment to complete
	if err := p.waitForDeployment(ctx, projectName, deployment.ID); err != nil {
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	// Cache the deployment URL
	p.deployURL = deployment.URL

	return &CDNInfo{
		ID:        deployment.ID,
		Name:      projectName,
		URL:       deployment.URL,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"provider":      "cloudflare",
			"project_name":  projectName,
			"deployment_id": deployment.ID,
			"deploy_type":   "api",
		},
	}, nil
}

// Invalidate purges the cache for specified paths
func (p *CloudflareProvider) Invalidate(ctx context.Context, paths []string) error {
	if p.apiToken == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN not set. Cannot invalidate cache without API credentials")
	}

	// Detect zone ID if not set
	if p.zoneID == "" {
		zoneID, err := p.detectZoneID(ctx)
		if err != nil {
			return fmt.Errorf("failed to detect zone ID: %w", err)
		}
		p.zoneID = zoneID
	}

	// Prepare purge request
	payload := map[string]interface{}{
		"files": paths,
	}

	// If no specific paths provided, purge everything
	if len(paths) == 0 {
		payload = map[string]interface{}{
			"purge_everything": true,
		}
	}

	var result struct {
		Success bool   `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/zones/%s/purge_cache", p.zoneID), payload, &result); err != nil {
		return err
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			return fmt.Errorf("cache invalidation failed: %s", result.Errors[0].Message)
		}
		return fmt.Errorf("cache invalidation failed")
	}

	return nil
}

// GetURL returns the deployment URL
func (p *CloudflareProvider) GetURL(ctx context.Context) (string, error) {
	// Check if we have a cached URL
	if p.deployURL != "" {
		return p.deployURL, nil
	}

	// Check environment variable
	deployURL := os.Getenv("CLOUDFLARE_DEPLOY_URL")
	if deployURL == "" {
		return "", fmt.Errorf("CLOUDFLARE_DEPLOY_URL not set. Please deploy your project first or set the URL manually")
	}

	p.deployURL = deployURL
	return deployURL, nil
}

// Cloudflare API types
type cloudflareProject struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	CreatedOn time.Time `json:"created_on"`
	Domains   []string  `json:"domains"`
}

type cloudflareDeployment struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"latest_stage"`
	CreatedOn time.Time `json:"created_on"`
}

type cloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// doAPIRequest performs an HTTP request to the Cloudflare API
func (p *CloudflareProvider) doAPIRequest(ctx context.Context, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, cloudflareAPIBase+path, body)
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
		return fmt.Errorf("cloudflare api %s %s: %s: %s", method, path, resp.Status, string(b))
	}

	if out != nil && resp.StatusCode != 204 {
		// Cloudflare API wraps responses in a result object
		var wrapper struct {
			Result  json.RawMessage `json:"result"`
			Success bool            `json:"success"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		if !wrapper.Success {
			if len(wrapper.Errors) > 0 {
				return fmt.Errorf("api error: %s", wrapper.Errors[0].Message)
			}
			return fmt.Errorf("api request failed")
		}

		if err := json.Unmarshal(wrapper.Result, out); err != nil {
			return fmt.Errorf("failed to decode result: %w", err)
		}
	}

	return nil
}

// getProject retrieves an existing Cloudflare Pages project
func (p *CloudflareProvider) getProject(ctx context.Context, projectName string) (*cloudflareProject, error) {
	var project cloudflareProject
	err := p.doAPIRequest(ctx, "GET", 
		fmt.Sprintf("/accounts/%s/pages/projects/%s", p.accountID, projectName), nil, &project)
	
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("project not found: %s", projectName)
		}
		return nil, err
	}

	return &project, nil
}

// createProject creates a new Cloudflare Pages project
func (p *CloudflareProvider) createProject(ctx context.Context, projectName string) (*cloudflareProject, error) {
	payload := map[string]interface{}{
		"name":              projectName,
		"production_branch": "main",
	}

	var project cloudflareProject
	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/accounts/%s/pages/projects", p.accountID), payload, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// createDeployment creates a new deployment for a Pages project
func (p *CloudflareProvider) createDeployment(ctx context.Context, projectName string, opts CDNDeployOptions) (*cloudflareDeployment, error) {
	// For direct uploads, we would need to upload files here
	// For now, we'll create a deployment record that assumes Git-based deployment
	
	payload := map[string]interface{}{
		"branch": opts.Branch,
	}

	if opts.Branch == "" {
		payload["branch"] = "main"
	}

	var deployment cloudflareDeployment
	if err := p.doAPIRequest(ctx, "POST", 
		fmt.Sprintf("/accounts/%s/pages/projects/%s/deployments", p.accountID, projectName), 
		payload, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// waitForDeployment waits for a deployment to complete
func (p *CloudflareProvider) waitForDeployment(ctx context.Context, projectName, deploymentID string) error {
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
			var deployment cloudflareDeployment
			if err := p.doAPIRequest(ctx, "GET", 
				fmt.Sprintf("/accounts/%s/pages/projects/%s/deployments/%s", 
					p.accountID, projectName, deploymentID), nil, &deployment); err != nil {
				return fmt.Errorf("failed to get deployment status: %w", err)
			}

			switch deployment.Status {
			case "success", "active":
				return nil
			case "failure", "canceled":
				return fmt.Errorf("deployment failed with status: %s", deployment.Status)
			}
			// Continue polling for other statuses (queued, building, deploying, etc.)
		}
	}
}

// detectZoneID attempts to detect the Cloudflare zone ID
func (p *CloudflareProvider) detectZoneID(ctx context.Context) (string, error) {
	// If we have a deploy URL, try to extract the domain
	if p.deployURL == "" {
		return "", fmt.Errorf("no deployment URL available for zone detection")
	}

	// Extract domain from URL
	domain := extractDomain(p.deployURL)
	if domain == "" {
		return "", fmt.Errorf("failed to extract domain from URL: %s", p.deployURL)
	}

	// List zones and find matching one
	var zones []cloudflareZone
	if err := p.doAPIRequest(ctx, "GET", "/zones", nil, &zones); err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	// Find zone that matches the domain
	for _, zone := range zones {
		if strings.HasSuffix(domain, zone.Name) {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("no zone found for domain: %s", domain)
}

// extractDomain extracts the domain from a URL
func extractDomain(urlStr string) string {
	// Remove protocol
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Remove path
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	// Remove port
	if idx := strings.Index(urlStr, ":"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	return urlStr
}
