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
}

// DatabaseConfig contains database provider configuration
type DatabaseConfig struct {
	Provider string                 `yaml:"provider"` // postgresql
	Options  map[string]interface{} `yaml:"options"`
}

// CacheConfig contains cache provider configuration
type CacheConfig struct {
	Provider string                 `yaml:"provider"` // valkey, redis, none
	Options  map[string]interface{} `yaml:"options"`
}

// LoadConfig loads provider configuration from a YAML file
// Environment variables can override the provider selection:
//   - GFORGE_DB_PROVIDER
//   - GFORGE_CACHE_PROVIDER
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

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Provider: "postgresql",
			Options: map[string]interface{}{
				"host":     "localhost",
				"port":     5432,
				"database": "gothic_forge",
			},
		},
		Cache: CacheConfig{
			Provider: "redis",
			Options: map[string]interface{}{
				"host": "localhost",
				"port": 6379,
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
