package providers

import (
	"context"
	"io"
	"time"
)

// DatabaseProvider abstracts database provisioning and management.
type DatabaseProvider interface {
	Name() string
	Provision(ctx context.Context, opts ProvisionOptions) (*DatabaseInfo, error)
	GetConnectionString(ctx context.Context) (string, error)
	RunMigrations(ctx context.Context, migrationsDir string) error
	Health(ctx context.Context) error
	Destroy(ctx context.Context) error
}

// CacheProvider abstracts cache provisioning and management.
type CacheProvider interface {
	Name() string
	Provision(ctx context.Context, opts ProvisionOptions) (*CacheInfo, error)
	GetConnectionString(ctx context.Context) (string, error)
	Health(ctx context.Context) error
	Destroy(ctx context.Context) error
}

// ComputeProvider abstracts application deployment.
type ComputeProvider interface {
	Name() string
	Deploy(ctx context.Context, opts DeployOptions) (*DeploymentInfo, error)
	GetURL(ctx context.Context) (string, error)
	GetLogs(ctx context.Context, opts LogOptions) (io.ReadCloser, error)
	Health(ctx context.Context) error
	Rollback(ctx context.Context, version string) error
	Scale(ctx context.Context, replicas int) error
}

// CDNProvider abstracts CDN/edge deployment.
type CDNProvider interface {
	Name() string
	Deploy(ctx context.Context, opts CDNDeployOptions) (*CDNInfo, error)
	Invalidate(ctx context.Context, paths []string) error
	GetURL(ctx context.Context) (string, error)
}

// ProvisionOptions contains options for provisioning resources
type ProvisionOptions struct {
	Name     string
	Region   string
	Tier     string
	Tags     map[string]string
	Metadata map[string]interface{}
}


// DatabaseInfo contains information about a provisioned database
type DatabaseInfo struct {
	ID               string
	Name             string
	ConnectionString string
	Host             string
	Port             int
	Database         string
	Username         string
	Password         string
	Region           string
	CreatedAt        time.Time
	Metadata         map[string]interface{}
}

// CacheInfo contains information about a provisioned cache
type CacheInfo struct {
	ID               string
	Name             string
	ConnectionString string
	Host             string
	Port             int
	Password         string
	Region           string
	CreatedAt        time.Time
	Metadata         map[string]interface{}
}

// DeployOptions contains options for application deployment
type DeployOptions struct {
	Name        string
	Image       string
	BuildPath   string
	EnvVars     map[string]string
	Replicas    int
	HealthCheck string
	Port        int
	Metadata    map[string]interface{}
}

// DeploymentInfo contains information about a deployment
type DeploymentInfo struct {
	ID           string
	Name         string
	URL          string
	DashboardURL string
	Version      string
	Status       string
	CreatedAt    time.Time
	Metadata     map[string]interface{}
}

// LogOptions contains options for log streaming
type LogOptions struct {
	Follow     bool
	Since      time.Time
	Tail       int
	Timestamps bool
}

// CDNDeployOptions contains options for CDN deployment
type CDNDeployOptions struct {
	ProjectName string
	Directory   string
	Branch      string
	Metadata    map[string]interface{}
}

// CDNInfo contains information about a CDN deployment
type CDNInfo struct {
	ID        string
	Name      string
	URL       string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}
