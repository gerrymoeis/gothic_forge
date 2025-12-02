package providers

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the provider configuration from providers.yaml
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	Compute  ComputeConfig  `yaml:"compute"`
	CDN      CDNConfig      `yaml:"cdn"`
}

// DatabaseConfig contains database provider configuration
type DatabaseConfig struct {
	Provider string                 `yaml:"provider"` // cockroachdb, postgresql, sqlite
	Options  map[string]interface{} `yaml:"options"`
}

// CacheConfig contains cache provider configuration
type CacheConfig struct {
	Provider string                 `yaml:"provider"` // valkey, redis, none
	Options  map[string]interface{} `yaml:"options"`
}

// ComputeConfig contains compute provider configuration
type ComputeConfig struct {
	Provider string                 `yaml:"provider"` // leapcell, docker, aws-ecs
	Options  map[string]interface{} `yaml:"options"`
}

// CDNConfig contains CDN provider configuration
type CDNConfig struct {
	Provider string                 `yaml:"provider"` // cloudflare, none
	Options  map[string]interface{} `yaml:"options"`
}

// LoadConfig loads provider configuration from a YAML file
// Environment variables can override the provider selection:
//   - GFORGE_DB_PROVIDER
//   - GFORGE_CACHE_PROVIDER
//   - GFORGE_COMPUTE_PROVIDER
//   - GFORGE_CDN_PROVIDER
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	if dbProvider := strings.TrimSpace(os.Getenv("GFORGE_DB_PROVIDER")); dbProvider != "" {
		cfg.Database.Provider = dbProvider
	}
	if cacheProvider := strings.TrimSpace(os.Getenv("GFORGE_CACHE_PROVIDER")); cacheProvider != "" {
		cfg.Cache.Provider = cacheProvider
	}
	if computeProvider := strings.TrimSpace(os.Getenv("GFORGE_COMPUTE_PROVIDER")); computeProvider != "" {
		cfg.Compute.Provider = computeProvider
	}
	if cdnProvider := strings.TrimSpace(os.Getenv("GFORGE_CDN_PROVIDER")); cdnProvider != "" {
		cfg.CDN.Provider = cdnProvider
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate database provider
	if c.Database.Provider == "" {
		return fmt.Errorf("database provider is required")
	}
	if _, err := DefaultRegistry.GetDatabase(c.Database.Provider); err != nil {
		return fmt.Errorf("invalid database provider: %w", err)
	}

	// Validate cache provider (optional, can be "none")
	if c.Cache.Provider != "" && c.Cache.Provider != "none" {
		if _, err := DefaultRegistry.GetCache(c.Cache.Provider); err != nil {
			return fmt.Errorf("invalid cache provider: %w", err)
		}
	}

	// Validate compute provider
	if c.Compute.Provider == "" {
		return fmt.Errorf("compute provider is required")
	}
	if _, err := DefaultRegistry.GetCompute(c.Compute.Provider); err != nil {
		return fmt.Errorf("invalid compute provider: %w", err)
	}

	// Validate CDN provider (optional, can be "none")
	if c.CDN.Provider != "" && c.CDN.Provider != "none" {
		if _, err := DefaultRegistry.GetCDN(c.CDN.Provider); err != nil {
			return fmt.Errorf("invalid CDN provider: %w", err)
		}
	}

	return nil
}

// GetString retrieves a string option from provider options
func GetString(options map[string]interface{}, key string, defaultValue string) string {
	if val, ok := options[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetInt retrieves an int option from provider options
func GetInt(options map[string]interface{}, key string, defaultValue int) int {
	if val, ok := options[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetBool retrieves a bool option from provider options
func GetBool(options map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := options[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// DefaultConfig returns the default Opinionated Stack configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Provider: "cockroachdb",
			Options: map[string]interface{}{
				"region": "us-east-1",
				"tier":   "serverless",
			},
		},
		Cache: CacheConfig{
			Provider: "valkey",
			Options: map[string]interface{}{
				"region": "us-east-1",
				"tier":   "startup",
			},
		},
		Compute: ComputeConfig{
			Provider: "leapcell",
			Options: map[string]interface{}{
				"region":   "global",
				"replicas": 1,
			},
		},
		CDN: CDNConfig{
			Provider: "cloudflare",
			Options: map[string]interface{}{
				"zone": "auto",
			},
		},
	}
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
