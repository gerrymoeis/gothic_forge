// Package webawesome provides Web Awesome design system integration for Gothic Forge
package webawesome

import (
	"fmt"
	"strings"
	"sync"
)

// Version represents the Web Awesome integration version
const Version = "1.0.0"

// Manager coordinates all Web Awesome functionality
type Manager struct {
	config          *Config
	registry        ComponentRegistry
	registryManager *RegistryManager
	themeProvider   ThemeProvider
	themeManager    *ThemeManager
	migrationTool   MigrationTool
	errorHandler    ErrorHandler
	mu              sync.RWMutex
}

// NewManager creates a new Web Awesome manager with default implementations
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	
	registryManager := NewRegistryManager(config)
	themeProvider := NewThemeProvider(config)
	themeManager := NewThemeManager(config)
	
	return &Manager{
		config:          config,
		registry:        registryManager.GetRegistry(),
		registryManager: registryManager,
		themeProvider:   themeProvider,
		themeManager:    themeManager,
		migrationTool:   NewMigrationTool(),
		errorHandler:    NewDefaultErrorHandler(),
	}
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the manager configuration
func (m *Manager) UpdateConfig(config *Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.config = config
	
	// Update theme provider with new config
	m.themeProvider = NewThemeProvider(config)
	
	return nil
}

// GetRegistry returns the component registry
func (m *Manager) GetRegistry() ComponentRegistry {
	return m.registry
}

// GetRegistryManager returns the registry manager
func (m *Manager) GetRegistryManager() *RegistryManager {
	return m.registryManager
}

// GetThemeProvider returns the theme provider
func (m *Manager) GetThemeProvider() ThemeProvider {
	return m.themeProvider
}

// GetThemeManager returns the theme manager
func (m *Manager) GetThemeManager() *ThemeManager {
	return m.themeManager
}

// GetMigrationTool returns the migration tool
func (m *Manager) GetMigrationTool() MigrationTool {
	return m.migrationTool
}

// GetErrorHandler returns the error handler
func (m *Manager) GetErrorHandler() ErrorHandler {
	return m.errorHandler
}

// Initialize sets up the Web Awesome integration
func (m *Manager) Initialize() error {
	// Validate configuration
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Initialize registry with default components
	if err := m.registryManager.Initialize(); err != nil {
		return fmt.Errorf("registry initialization failed: %w", err)
	}
	
	// Set up default theme if needed
	if m.config.Themes.AutoLoad {
		defaultTheme := DefaultTheme()
		if err := m.themeProvider.ValidateTheme(defaultTheme); err != nil {
			return fmt.Errorf("default theme validation failed: %w", err)
		}
	}
	
	return nil
}

// GetVersion returns the Web Awesome integration version
func GetVersion() string {
	return Version
}

// GetSupportedComponents returns a list of supported Web Awesome components
func GetSupportedComponents() []string {
	return []string{
		"sl-button",
		"sl-input",
		"sl-select",
		"sl-textarea",
		"sl-checkbox",
		"sl-radio",
		"sl-card",
		"sl-dialog",
		"sl-drawer",
		"sl-tab-group",
		"sl-badge",
		"sl-progress-bar",
		"sl-spinner",
		"sl-dropdown",
		"sl-menu",
		"sl-tooltip",
	}
}

// GetComponentMapping returns the DaisyUI to Web Awesome component mapping
func GetComponentMapping() map[string]string {
	tool := NewMigrationTool()
	return tool.componentMappings
}

// ValidateComponent checks if a component configuration is valid
func ValidateComponent(component Component) error {
	if component == nil {
		return NewValidationError("component", "", "Component cannot be nil", "required")
	}
	
	return component.Validate()
}

// BuildAttributes builds HTML attributes from a component's properties
func BuildAttributes(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	
	var result string
	for key, value := range attrs {
		if value != "" {
			result += fmt.Sprintf(` %s="%s"`, key, value)
		} else {
			result += fmt.Sprintf(` %s`, key)
		}
	}
	
	return result
}

// EscapeHTML escapes HTML content for safe rendering
func EscapeHTML(content string) string {
	// This is a basic implementation - in production, use a proper HTML escaping library
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}
	
	result := content
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	return result
}

// Global manager instance (optional - can be used for singleton pattern)
var globalManager *Manager
var globalManagerOnce sync.Once

// GetGlobalManager returns a global manager instance
func GetGlobalManager() *Manager {
	globalManagerOnce.Do(func() {
		globalManager = NewManager(nil)
	})
	return globalManager
}

// SetGlobalManager sets the global manager instance
func SetGlobalManager(manager *Manager) {
	globalManager = manager
}