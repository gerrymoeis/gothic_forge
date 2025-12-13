package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ThemeManager provides simple theme management
type ThemeManager struct {
	provider ThemeProvider
	config   *Config
	current  *Theme
}

// NewThemeManager creates a new theme manager
func NewThemeManager(config *Config) *ThemeManager {
	return &ThemeManager{
		provider: NewThemeProvider(config),
		config:   config,
		current:  nil,
	}
}

// GetCurrentTheme returns the currently active theme
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	if tm.current == nil {
		// Load default theme
		theme, _ := tm.provider.LoadTheme(tm.config.Themes.Default)
		tm.current = theme
	}
	return tm.current
}

// SetTheme switches to a different theme
func (tm *ThemeManager) SetTheme(name string) error {
	theme, err := tm.provider.LoadTheme(name)
	if err != nil {
		return fmt.Errorf("failed to load theme '%s': %w", name, err)
	}
	
	tm.current = theme
	return nil
}

// CreateTheme creates a new theme from a template
func (tm *ThemeManager) CreateTheme(name, displayName string, baseTheme *Theme) *Theme {
	if baseTheme == nil {
		baseTheme = DefaultTheme()
	}
	
	// Create a copy of the base theme
	newTheme := *baseTheme
	newTheme.Name = name
	newTheme.DisplayName = displayName
	newTheme.Description = fmt.Sprintf("Custom theme based on %s", baseTheme.DisplayName)
	newTheme.Version = "1.0.0"
	
	// Initialize custom properties if nil
	if newTheme.Custom == nil {
		newTheme.Custom = make(map[string]string)
	}
	
	return &newTheme
}

// SaveTheme saves a theme to disk
func (tm *ThemeManager) SaveTheme(theme *Theme) error {
	if err := tm.provider.ValidateTheme(theme); err != nil {
		return fmt.Errorf("theme validation failed: %w", err)
	}
	
	// Ensure themes directory exists
	themesDir := tm.config.Themes.CustomPath
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}
	
	// Save theme to file
	filePath := filepath.Join(themesDir, theme.Name+".json")
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}
	
	return nil
}

// GenerateCSS generates CSS for the current theme
func (tm *ThemeManager) GenerateCSS() (string, error) {
	theme := tm.GetCurrentTheme()
	if theme == nil {
		return "", fmt.Errorf("no theme available")
	}
	
	return tm.provider.GenerateCSS(theme)
}

// GenerateThemeCSS generates CSS for a specific theme
func (tm *ThemeManager) GenerateThemeCSS(themeName string) (string, error) {
	theme, err := tm.provider.LoadTheme(themeName)
	if err != nil {
		return "", err
	}
	
	return tm.provider.GenerateCSS(theme)
}

// ListAvailableThemes returns all available themes
func (tm *ThemeManager) ListAvailableThemes() []string {
	themes := []string{"default", "dark"} // Built-in themes
	
	// Add custom themes from directory
	themesDir := tm.config.Themes.CustomPath
	if entries, err := os.ReadDir(themesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				themes = append(themes, name)
			}
		}
	}
	
	return themes
}

// GetThemeInfo returns information about a theme without loading it fully
func (tm *ThemeManager) GetThemeInfo(name string) (*ThemeInfo, error) {
	theme, err := tm.provider.LoadTheme(name)
	if err != nil {
		return nil, err
	}
	
	return &ThemeInfo{
		Name:        theme.Name,
		DisplayName: theme.DisplayName,
		Description: theme.Description,
		Version:     theme.Version,
		IsBuiltIn:   name == "default" || name == "dark",
	}, nil
}

// Simple theme customization helpers

// SetPrimaryColor updates the primary color of the current theme
func (tm *ThemeManager) SetPrimaryColor(color string) error {
	theme := tm.GetCurrentTheme()
	if theme == nil {
		return fmt.Errorf("no theme available")
	}
	
	// Validate color format
	if err := tm.provider.(*DefaultThemeProvider).validateColor("primary", color); err != nil {
		return err
	}
	
	theme.Colors.Primary = color
	return nil
}

// SetColors updates multiple colors at once
func (tm *ThemeManager) SetColors(colors map[string]string) error {
	theme := tm.GetCurrentTheme()
	if theme == nil {
		return fmt.Errorf("no theme available")
	}
	
	// Validate all colors first
	provider := tm.provider.(*DefaultThemeProvider)
	for name, color := range colors {
		if err := provider.validateColor(name, color); err != nil {
			return fmt.Errorf("invalid color '%s': %w", name, err)
		}
	}
	
	// Apply colors
	for name, color := range colors {
		switch name {
		case "primary":
			theme.Colors.Primary = color
		case "secondary":
			theme.Colors.Secondary = color
		case "success":
			theme.Colors.Success = color
		case "warning":
			theme.Colors.Warning = color
		case "danger":
			theme.Colors.Danger = color
		case "neutral":
			theme.Colors.Neutral = color
		case "background":
			theme.Colors.Background = color
		case "surface":
			theme.Colors.Surface = color
		}
	}
	
	return nil
}

// SetFontFamily updates the font family
func (tm *ThemeManager) SetFontFamily(fontFamily string) {
	theme := tm.GetCurrentTheme()
	if theme != nil {
		theme.Typography.FontFamily = fontFamily
	}
}

// AddCustomProperty adds a custom CSS property
func (tm *ThemeManager) AddCustomProperty(name, value string) {
	theme := tm.GetCurrentTheme()
	if theme != nil {
		if theme.Custom == nil {
			theme.Custom = make(map[string]string)
		}
		theme.Custom[name] = value
	}
}

// ThemeInfo provides basic information about a theme
type ThemeInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Version     string `json:"version"`
	IsBuiltIn   bool   `json:"isBuiltIn"`
}

// Predefined theme helpers

// CreateLightTheme creates a light theme variant
func CreateLightTheme(name, displayName string) *Theme {
	theme := DefaultTheme()
	theme.Name = name
	theme.DisplayName = displayName
	theme.Description = "Light theme variant"
	return theme
}

// CreateDarkTheme creates a dark theme variant
func CreateDarkTheme(name, displayName string) *Theme {
	theme := DarkTheme()
	theme.Name = name
	theme.DisplayName = displayName
	theme.Description = "Dark theme variant"
	return theme
}

// CreateBrandTheme creates a theme with brand colors
func CreateBrandTheme(name, displayName, primaryColor string) *Theme {
	theme := DefaultTheme()
	theme.Name = name
	theme.DisplayName = displayName
	theme.Description = "Brand theme with custom primary color"
	theme.Colors.Primary = primaryColor
	return theme
}