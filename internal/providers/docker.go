package providers

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// DockerProvider implements ComputeProvider for local Docker deployments
type DockerProvider struct {
	containerName string
}

// NewDockerProvider creates a new Docker provider
func NewDockerProvider() *DockerProvider {
	return &DockerProvider{
		containerName: "gothic-forge-app",
	}
}

// Name returns the provider name
func (p *DockerProvider) Name() string {
	return "docker"
}

// Deploy builds and runs the application in a Docker container
func (p *DockerProvider) Deploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error) {
	// Build Docker image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", opts.Name, opts.BuildPath)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker build failed: %w\nOutput: %s", err, output)
	}

	// Stop and remove existing container if it exists
	stopCmd := exec.CommandContext(ctx, "docker", "stop", p.containerName)
	stopCmd.Run() // Ignore error if container doesn't exist

	rmCmd := exec.CommandContext(ctx, "docker", "rm", p.containerName)
	rmCmd.Run() // Ignore error if container doesn't exist

	// Build docker run command with environment variables
	runArgs := []string{
		"run", "-d",
		"--name", p.containerName,
		"-p", fmt.Sprintf("%d:%d", opts.Port, opts.Port),
	}

	// Add environment variables
	for key, value := range opts.EnvVars {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image name
	runArgs = append(runArgs, opts.Name)

	// Run container
	runCmd := exec.CommandContext(ctx, "docker", runArgs...)
	if output, err := runCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker run failed: %w\nOutput: %s", err, output)
	}

	// Get container ID
	containerID, err := p.getContainerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container ID: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d", opts.Port)

	return &DeploymentInfo{
		ID:           containerID,
		Name:         opts.Name,
		URL:          url,
		DashboardURL: "http://localhost:2375", // Docker API endpoint
		Version:      "latest",
		Status:       "running",
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"container_name": p.containerName,
			"port":           opts.Port,
		},
	}, nil
}

// GetURL returns the application URL
func (p *DockerProvider) GetURL(ctx context.Context) (string, error) {
	// Get port mapping
	cmd := exec.CommandContext(ctx, "docker", "port", p.containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get port mapping: %w", err)
	}

	// Parse output (e.g., "8080/tcp -> 0.0.0.0:8080")
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no port mappings found")
	}

	// Extract port from first line
	parts := strings.Split(lines[0], ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid port mapping format")
	}

	port := strings.TrimSpace(parts[len(parts)-1])
	return fmt.Sprintf("http://localhost:%s", port), nil
}

// GetLogs streams container logs
func (p *DockerProvider) GetLogs(ctx context.Context, opts LogOptions) (io.ReadCloser, error) {
	args := []string{"logs"}

	if opts.Follow {
		args = append(args, "-f")
	}
	if opts.Timestamps {
		args = append(args, "-t")
	}
	if opts.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.Tail))
	}
	if !opts.Since.IsZero() {
		args = append(args, "--since", opts.Since.Format(time.RFC3339))
	}

	args = append(args, p.containerName)

	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start logs command: %w", err)
	}

	return stdout, nil
}

// Health checks if the container is running and healthy
func (p *DockerProvider) Health(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Health.Status}}", p.containerName)
	output, err := cmd.Output()
	if err != nil {
		// If no health check is defined, check if container is running
		statusCmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}", p.containerName)
		statusOutput, statusErr := statusCmd.Output()
		if statusErr != nil {
			return fmt.Errorf("container not found or not running: %w", statusErr)
		}

		status := strings.TrimSpace(string(statusOutput))
		if status != "running" {
			return fmt.Errorf("container status: %s", status)
		}

		return nil
	}

	healthStatus := strings.TrimSpace(string(output))
	if healthStatus != "healthy" && healthStatus != "" {
		return fmt.Errorf("container health status: %s", healthStatus)
	}

	return nil
}

// Rollback reverts to a previous version (not implemented for Docker)
func (p *DockerProvider) Rollback(ctx context.Context, version string) error {
	return fmt.Errorf("rollback not implemented for Docker provider")
}

// Scale adjusts the number of replicas (not applicable for single Docker container)
func (p *DockerProvider) Scale(ctx context.Context, replicas int) error {
	if replicas != 1 {
		return fmt.Errorf("Docker provider only supports 1 replica")
	}
	return nil
}

// getContainerID retrieves the container ID
func (p *DockerProvider) getContainerID(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.Id}}", p.containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
