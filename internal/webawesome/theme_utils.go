package webawesome

import (
	"fmt"
	"os"
	"path/filepath"
)

// ThemeUtils provides utility functions for theme operations
type ThemeUtils struct{}

// NewThemeUtils creates a new theme utilities instance
func NewThemeUtils() *ThemeUtils {
	return &ThemeUtils{}
}

// WriteThemeCSS writes theme CSS to a file
func (tu *ThemeUtils) WriteThemeCSS(theme *Theme, outputPath string) error {
	provider := &DefaultThemeProvider{}
	css, err := provider.GenerateCSS(theme)
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}
	
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write CSS file
	if err := os.WriteFile(outputPath, []byte(css), 0644); err != nil {
		return fmt.Errorf("failed to write CSS file: %w", err)
	}
	
	return nil
}

// GetThemeCSS returns the CSS string for a theme
func (tu *ThemeUtils) GetThemeCSS(theme *Theme) (string, error) {
	provider := &DefaultThemeProvider{}
	return provider.GenerateCSS(theme)
}

// ValidateTheme validates a theme configuration
func (tu *ThemeUtils) ValidateTheme(theme *Theme) error {
	provider := &DefaultThemeProvider{}
	return provider.ValidateTheme(theme)
}

// CloneTheme creates a deep copy of a theme
func (tu *ThemeUtils) CloneTheme(source *Theme) *Theme {
	if source == nil {
		return nil
	}
	
	clone := &Theme{
		Name:        source.Name,
		DisplayName: source.DisplayName,
		Description: source.Description,
		Version:     source.Version,
		Colors:      source.Colors,      // struct copy
		Typography:  source.Typography,  // struct copy
		Spacing:     source.Spacing,     // struct copy
		BorderRadius: source.BorderRadius, // struct copy
		Shadows:     source.Shadows,     // struct copy
		Custom:      make(map[string]string),
	}
	
	// Copy custom properties
	for k, v := range source.Custom {
		clone.Custom[k] = v
	}
	
	return clone
}

// MergeThemes merges two themes, with the second theme taking precedence
func (tu *ThemeUtils) MergeThemes(base, override *Theme) *Theme {
	if base == nil {
		return tu.CloneTheme(override)
	}
	if override == nil {
		return tu.CloneTheme(base)
	}
	
	merged := tu.CloneTheme(base)
	
	// Override basic properties
	if override.Name != "" {
		merged.Name = override.Name
	}
	if override.DisplayName != "" {
		merged.DisplayName = override.DisplayName
	}
	if override.Description != "" {
		merged.Description = override.Description
	}
	if override.Version != "" {
		merged.Version = override.Version
	}
	
	// Override colors (only non-empty values)
	if override.Colors.Primary != "" {
		merged.Colors.Primary = override.Colors.Primary
	}
	if override.Colors.Secondary != "" {
		merged.Colors.Secondary = override.Colors.Secondary
	}
	if override.Colors.Success != "" {
		merged.Colors.Success = override.Colors.Success
	}
	if override.Colors.Warning != "" {
		merged.Colors.Warning = override.Colors.Warning
	}
	if override.Colors.Danger != "" {
		merged.Colors.Danger = override.Colors.Danger
	}
	if override.Colors.Neutral != "" {
		merged.Colors.Neutral = override.Colors.Neutral
	}
	if override.Colors.Background != "" {
		merged.Colors.Background = override.Colors.Background
	}
	if override.Colors.Surface != "" {
		merged.Colors.Surface = override.Colors.Surface
	}
	
	// Override typography (only non-empty values)
	if override.Typography.FontFamily != "" {
		merged.Typography.FontFamily = override.Typography.FontFamily
	}
	if override.Typography.FontSize != "" {
		merged.Typography.FontSize = override.Typography.FontSize
	}
	if override.Typography.LineHeight != "" {
		merged.Typography.LineHeight = override.Typography.LineHeight
	}
	if override.Typography.FontWeight != "" {
		merged.Typography.FontWeight = override.Typography.FontWeight
	}
	
	// Merge custom properties
	for k, v := range override.Custom {
		merged.Custom[k] = v
	}
	
	return merged
}

// CreateQuickTheme creates a theme with just the essential colors
func (tu *ThemeUtils) CreateQuickTheme(name, displayName, primary, background string) *Theme {
	theme := DefaultTheme()
	theme.Name = name
	theme.DisplayName = displayName
	theme.Description = "Quick theme with custom colors"
	
	if primary != "" {
		theme.Colors.Primary = primary
	}
	if background != "" {
		theme.Colors.Background = background
		theme.Colors.Surface = background
	}
	
	return theme
}

// GetThemePreview generates a simple HTML preview of a theme
func (tu *ThemeUtils) GetThemePreview(theme *Theme) string {
	css, _ := tu.GetThemeCSS(theme)
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s Preview</title>
    <style>%s</style>
    <style>
        body { font-family: var(--sl-font-sans); padding: 2rem; background: var(--sl-color-neutral-0); color: var(--sl-color-neutral-1000); }
        .preview { max-width: 600px; margin: 0 auto; }
        .color-box { display: inline-block; width: 60px; height: 60px; margin: 0.5rem; border-radius: var(--sl-border-radius-medium); }
        .primary { background: var(--sl-color-primary-500); }
        .success { background: var(--sl-color-success-500); }
        .warning { background: var(--sl-color-warning-500); }
        .danger { background: var(--sl-color-danger-500); }
        .neutral { background: var(--sl-color-neutral-500); }
        .sample-button { 
            background: var(--sl-color-primary-500); 
            color: white; 
            border: none; 
            padding: var(--sl-spacing-medium) var(--sl-spacing-large);
            border-radius: var(--sl-border-radius-medium);
            font-family: var(--sl-font-sans);
            cursor: pointer;
            margin: var(--sl-spacing-small);
        }
    </style>
</head>
<body>
    <div class="preview">
        <h1>%s</h1>
        <p>%s</p>
        
        <h2>Colors</h2>
        <div class="color-box primary" title="Primary"></div>
        <div class="color-box success" title="Success"></div>
        <div class="color-box warning" title="Warning"></div>
        <div class="color-box danger" title="Danger"></div>
        <div class="color-box neutral" title="Neutral"></div>
        
        <h2>Components</h2>
        <button class="sample-button">Sample Button</button>
        
        <h2>Typography</h2>
        <p>Font Family: %s</p>
        <p>Font Size: %s</p>
        <p>Line Height: %s</p>
    </div>
</body>
</html>`, 
		theme.DisplayName, 
		css,
		theme.DisplayName,
		theme.Description,
		theme.Typography.FontFamily,
		theme.Typography.FontSize,
		theme.Typography.LineHeight,
	)
	
	return html
}

// Common theme presets

// GetPresetThemes returns a collection of preset themes
func GetPresetThemes() map[string]*Theme {
	return map[string]*Theme{
		"blue":   CreateBlueTheme(),
		"green":  CreateGreenTheme(),
		"purple": CreatePurpleTheme(),
		"orange": CreateOrangeTheme(),
		"red":    CreateRedTheme(),
	}
}

// CreateBlueTheme creates a blue-themed variant
func CreateBlueTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "blue"
	theme.DisplayName = "Blue Theme"
	theme.Description = "A clean blue theme"
	theme.Colors.Primary = "hsl(210 100% 50%)"
	return theme
}

// CreateGreenTheme creates a green-themed variant
func CreateGreenTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "green"
	theme.DisplayName = "Green Theme"
	theme.Description = "A fresh green theme"
	theme.Colors.Primary = "hsl(120 60% 45%)"
	return theme
}

// CreatePurpleTheme creates a purple-themed variant
func CreatePurpleTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "purple"
	theme.DisplayName = "Purple Theme"
	theme.Description = "A vibrant purple theme"
	theme.Colors.Primary = "hsl(270 70% 50%)"
	return theme
}

// CreateOrangeTheme creates an orange-themed variant
func CreateOrangeTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "orange"
	theme.DisplayName = "Orange Theme"
	theme.Description = "A warm orange theme"
	theme.Colors.Primary = "hsl(30 100% 50%)"
	return theme
}

// CreateRedTheme creates a red-themed variant
func CreateRedTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "red"
	theme.DisplayName = "Red Theme"
	theme.Description = "A bold red theme"
	theme.Colors.Primary = "hsl(0 70% 50%)"
	return theme
}