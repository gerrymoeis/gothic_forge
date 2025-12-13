package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the Web Awesome integration configuration
type Config struct {
	// CDN Configuration
	CDN CDNConfig `json:"cdn" yaml:"cdn"`
	
	// Theme Configuration
	Themes ThemeConfig `json:"themes" yaml:"themes"`
	
	// Component Configuration
	Components ComponentConfig `json:"components" yaml:"components"`
	
	// Build Configuration
	Build BuildConfig `json:"build" yaml:"build"`
	
	// Security Configuration
	Security SecurityConfig `json:"security" yaml:"security"`
}

// CDNConfig configures Web Awesome CDN settings
type CDNConfig struct {
	BaseURL     string `json:"baseUrl" yaml:"baseUrl"`
	Version     string `json:"version" yaml:"version"`
	Fallback    bool   `json:"fallback" yaml:"fallback"`
	LocalPath   string `json:"localPath" yaml:"localPath"`
	Timeout     int    `json:"timeout" yaml:"timeout"` // seconds
}

// ThemeConfig configures theme settings
type ThemeConfig struct {
	Default    string   `json:"default" yaml:"default"`
	Available  []string `json:"available" yaml:"available"`
	CustomPath string   `json:"customPath" yaml:"customPath"`
	AutoLoad   bool     `json:"autoLoad" yaml:"autoLoad"`
}

// ComponentConfig configures component settings
type ComponentConfig struct {
	Registry    string   `json:"registry" yaml:"registry"`
	AutoImport  bool     `json:"autoImport" yaml:"autoImport"`
	Prefixes    []string `json:"prefixes" yaml:"prefixes"`
	Exclude     []string `json:"exclude" yaml:"exclude"`
}

// BuildConfig configures build-time settings
type BuildConfig struct {
	Optimize     bool   `json:"optimize" yaml:"optimize"`
	TreeShaking  bool   `json:"treeShaking" yaml:"treeShaking"`
	Minify       bool   `json:"minify" yaml:"minify"`
	SourceMaps   bool   `json:"sourceMaps" yaml:"sourceMaps"`
	OutputPath   string `json:"outputPath" yaml:"outputPath"`
}

// SecurityConfig configures security settings
type SecurityConfig struct {
	CSPNonce     bool     `json:"cspNonce" yaml:"cspNonce"`
	AllowedHosts []string `json:"allowedHosts" yaml:"allowedHosts"`
	SanitizeHTML bool     `json:"sanitizeHtml" yaml:"sanitizeHtml"`
}

// DefaultConfig returns a default Web Awesome configuration
func DefaultConfig() *Config {
	return &Config{
		CDN: CDNConfig{
			BaseURL:   "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.15.1/cdn",
			Version:   "2.15.1",
			Fallback:  true,
			LocalPath: "/static/webawesome",
			Timeout:   10,
		},
		Themes: ThemeConfig{
			Default:    "light",
			Available:  []string{"light", "dark"},
			CustomPath: "/static/themes",
			AutoLoad:   true,
		},
		Components: ComponentConfig{
			Registry:   "auto",
			AutoImport: true,
			Prefixes:   []string{"sl-"},
			Exclude:    []string{},
		},
		Build: BuildConfig{
			Optimize:    true,
			TreeShaking: true,
			Minify:      true,
			SourceMaps:  false,
			OutputPath:  "/static/webawesome",
		},
		Security: SecurityConfig{
			CSPNonce:     true,
			AllowedHosts: []string{"cdn.jsdelivr.net", "unpkg.com"},
			SanitizeHTML: true,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse JSON (could be extended to support YAML)
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate and set defaults
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	config.SetDefaults()
	
	return &config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate CDN configuration
	if c.CDN.BaseURL == "" {
		return NewValidationError("cdn.baseUrl", c.CDN.BaseURL, "CDN base URL cannot be empty", "required")
	}
	
	if c.CDN.Version == "" {
		return NewValidationError("cdn.version", c.CDN.Version, "CDN version cannot be empty", "required")
	}
	
	if c.CDN.Timeout <= 0 {
		return NewValidationError("cdn.timeout", fmt.Sprintf("%d", c.CDN.Timeout), "CDN timeout must be positive", "invalid_value")
	}
	
	// Validate theme configuration
	if c.Themes.Default == "" {
		return NewValidationError("themes.default", c.Themes.Default, "Default theme cannot be empty", "required")
	}
	
	// Validate build configuration
	if c.Build.OutputPath == "" {
		return NewValidationError("build.outputPath", c.Build.OutputPath, "Build output path cannot be empty", "required")
	}
	
	return nil
}

// SetDefaults sets default values for missing configuration options
func (c *Config) SetDefaults() {
	defaults := DefaultConfig()
	
	// Set CDN defaults
	if c.CDN.BaseURL == "" {
		c.CDN.BaseURL = defaults.CDN.BaseURL
	}
	if c.CDN.Version == "" {
		c.CDN.Version = defaults.CDN.Version
	}
	if c.CDN.LocalPath == "" {
		c.CDN.LocalPath = defaults.CDN.LocalPath
	}
	if c.CDN.Timeout == 0 {
		c.CDN.Timeout = defaults.CDN.Timeout
	}
	
	// Set theme defaults
	if c.Themes.Default == "" {
		c.Themes.Default = defaults.Themes.Default
	}
	if len(c.Themes.Available) == 0 {
		c.Themes.Available = defaults.Themes.Available
	}
	if c.Themes.CustomPath == "" {
		c.Themes.CustomPath = defaults.Themes.CustomPath
	}
	
	// Set component defaults
	if c.Components.Registry == "" {
		c.Components.Registry = defaults.Components.Registry
	}
	if len(c.Components.Prefixes) == 0 {
		c.Components.Prefixes = defaults.Components.Prefixes
	}
	
	// Set build defaults
	if c.Build.OutputPath == "" {
		c.Build.OutputPath = defaults.Build.OutputPath
	}
	
	// Set security defaults
	if len(c.Security.AllowedHosts) == 0 {
		c.Security.AllowedHosts = defaults.Security.AllowedHosts
	}
}

// GetConfigPath returns the default configuration file path for a project
func GetConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".gforge", "webawesome.json")
}

// InitConfig creates a new configuration file with default values
func InitConfig(projectRoot string) error {
	configPath := GetConfigPath(projectRoot)
	config := DefaultConfig()
	return config.SaveConfig(configPath)
}