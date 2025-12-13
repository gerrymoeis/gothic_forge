package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultThemeProvider provides a default implementation of ThemeProvider
type DefaultThemeProvider struct {
	config *Config
	themes map[string]*Theme
}

// NewThemeProvider creates a new theme provider
func NewThemeProvider(config *Config) ThemeProvider {
	return &DefaultThemeProvider{
		config: config,
		themes: make(map[string]*Theme),
	}
}

// LoadTheme loads a theme by name
func (tp *DefaultThemeProvider) LoadTheme(name string) (*Theme, error) {
	// Check if theme is already loaded
	if theme, exists := tp.themes[name]; exists {
		return theme, nil
	}
	
	// Try to load from file
	themePath := filepath.Join(tp.config.Themes.CustomPath, name+".json")
	if _, err := os.Stat(themePath); err == nil {
		theme, err := tp.loadThemeFromFile(themePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load theme from file: %w", err)
		}
		tp.themes[name] = theme
		return theme, nil
	}
	
	// Return default theme if name matches
	if name == "default" || name == "light" {
		theme := DefaultTheme()
		tp.themes[name] = theme
		return theme, nil
	}
	
	// Return dark theme variant
	if name == "dark" {
		theme := DarkTheme()
		tp.themes[name] = theme
		return theme, nil
	}
	
	return nil, fmt.Errorf("theme not found: %s", name)
}

// GenerateCSS generates CSS from a theme
func (tp *DefaultThemeProvider) GenerateCSS(theme *Theme) (string, error) {
	if err := tp.ValidateTheme(theme); err != nil {
		return "", fmt.Errorf("invalid theme: %w", err)
	}
	
	var css strings.Builder
	
	// Write theme header
	css.WriteString(fmt.Sprintf("/* Web Awesome Theme: %s */\n", theme.DisplayName))
	css.WriteString("/* Generated for Gothic Forge */\n\n")
	
	// Write CSS custom properties
	css.WriteString(":root {\n")
	
	// Core color properties - simplified but complete
	css.WriteString("  /* Primary Colors */\n")
	css.WriteString(fmt.Sprintf("  --sl-color-primary-500: %s;\n", theme.Colors.Primary))
	css.WriteString(fmt.Sprintf("  --sl-color-primary-600: %s;\n", theme.Colors.Primary)) // Simplified - same color
	
	css.WriteString("\n  /* Semantic Colors */\n")
	css.WriteString(fmt.Sprintf("  --sl-color-success-500: %s;\n", theme.Colors.Success))
	css.WriteString(fmt.Sprintf("  --sl-color-warning-500: %s;\n", theme.Colors.Warning))
	css.WriteString(fmt.Sprintf("  --sl-color-danger-500: %s;\n", theme.Colors.Danger))
	css.WriteString(fmt.Sprintf("  --sl-color-neutral-500: %s;\n", theme.Colors.Neutral))
	
	css.WriteString("\n  /* Surface Colors */\n")
	css.WriteString(fmt.Sprintf("  --sl-color-neutral-0: %s;\n", theme.Colors.Background))
	css.WriteString(fmt.Sprintf("  --sl-color-neutral-50: %s;\n", theme.Colors.Surface))
	css.WriteString(fmt.Sprintf("  --sl-color-neutral-1000: %s;\n", theme.Colors.OnSurface))
	
	// Typography properties
	css.WriteString("\n  /* Typography */\n")
	css.WriteString(fmt.Sprintf("  --sl-font-sans: %s;\n", theme.Typography.FontFamily))
	css.WriteString(fmt.Sprintf("  --sl-font-size-medium: %s;\n", theme.Typography.FontSize))
	css.WriteString(fmt.Sprintf("  --sl-line-height-normal: %s;\n", theme.Typography.LineHeight))
	css.WriteString(fmt.Sprintf("  --sl-font-weight-normal: %s;\n", theme.Typography.FontWeight))
	
	// Spacing properties
	css.WriteString("\n  /* Spacing */\n")
	css.WriteString(fmt.Sprintf("  --sl-spacing-x-small: %s;\n", theme.Spacing.XSmall))
	css.WriteString(fmt.Sprintf("  --sl-spacing-small: %s;\n", theme.Spacing.Small))
	css.WriteString(fmt.Sprintf("  --sl-spacing-medium: %s;\n", theme.Spacing.Medium))
	css.WriteString(fmt.Sprintf("  --sl-spacing-large: %s;\n", theme.Spacing.Large))
	css.WriteString(fmt.Sprintf("  --sl-spacing-x-large: %s;\n", theme.Spacing.XLarge))
	
	// Border radius properties
	css.WriteString("\n  /* Border Radius */\n")
	css.WriteString(fmt.Sprintf("  --sl-border-radius-small: %s;\n", theme.BorderRadius.Small))
	css.WriteString(fmt.Sprintf("  --sl-border-radius-medium: %s;\n", theme.BorderRadius.Medium))
	css.WriteString(fmt.Sprintf("  --sl-border-radius-large: %s;\n", theme.BorderRadius.Large))
	css.WriteString(fmt.Sprintf("  --sl-border-radius-x-large: %s;\n", theme.BorderRadius.XLarge))
	css.WriteString(fmt.Sprintf("  --sl-border-radius-pill: %s;\n", theme.BorderRadius.Pill))
	
	// Shadow properties
	css.WriteString("\n  /* Shadows */\n")
	css.WriteString(fmt.Sprintf("  --sl-shadow-x-small: %s;\n", theme.Shadows.XSmall))
	css.WriteString(fmt.Sprintf("  --sl-shadow-small: %s;\n", theme.Shadows.Small))
	css.WriteString(fmt.Sprintf("  --sl-shadow-medium: %s;\n", theme.Shadows.Medium))
	css.WriteString(fmt.Sprintf("  --sl-shadow-large: %s;\n", theme.Shadows.Large))
	css.WriteString(fmt.Sprintf("  --sl-shadow-x-large: %s;\n", theme.Shadows.XLarge))
	
	// Custom properties
	if len(theme.Custom) > 0 {
		css.WriteString("\n  /* Custom Properties */\n")
		for key, value := range theme.Custom {
			css.WriteString(fmt.Sprintf("  --%s: %s;\n", key, value))
		}
	}
	
	css.WriteString("}\n")
	
	return css.String(), nil
}

// ValidateTheme validates a theme configuration
func (tp *DefaultThemeProvider) ValidateTheme(theme *Theme) error {
	if theme == nil {
		return NewValidationError("theme", "", "Theme cannot be nil", "required")
	}
	
	if theme.Name == "" {
		return NewValidationError("theme.name", theme.Name, "Theme name cannot be empty", "required")
	}
	
	// Validate colors
	if err := tp.validateColor("primary", theme.Colors.Primary); err != nil {
		return err
	}
	if err := tp.validateColor("secondary", theme.Colors.Secondary); err != nil {
		return err
	}
	if err := tp.validateColor("success", theme.Colors.Success); err != nil {
		return err
	}
	if err := tp.validateColor("warning", theme.Colors.Warning); err != nil {
		return err
	}
	if err := tp.validateColor("danger", theme.Colors.Danger); err != nil {
		return err
	}
	
	// Validate typography
	if theme.Typography.FontFamily == "" {
		return NewValidationError("theme.typography.fontFamily", theme.Typography.FontFamily, "Font family cannot be empty", "required")
	}
	
	return nil
}

// loadThemeFromFile loads a theme from a JSON file
func (tp *DefaultThemeProvider) loadThemeFromFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, err
	}
	
	return &theme, nil
}

// validateColor validates a color value (supports hex, hsl, rgb)
func (tp *DefaultThemeProvider) validateColor(name, color string) error {
	if color == "" {
		return NewValidationError("theme.colors."+name, color, "Color cannot be empty", "required")
	}
	
	// Basic color validation patterns
	patterns := []string{
		`^#[0-9a-fA-F]{3}$`,                                    // #rgb
		`^#[0-9a-fA-F]{6}$`,                                    // #rrggbb
		`^#[0-9a-fA-F]{8}$`,                                    // #rrggbbaa
		`^rgb\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*\)$`,             // rgb(r,g,b)
		`^rgba\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*,\s*[\d.]+\s*\)$`, // rgba(r,g,b,a)
		`^hsl\(\s*\d+\s*,\s*\d+%\s*,\s*\d+%\s*\)$`,           // hsl(h,s,l)
		`^hsla\(\s*\d+\s*,\s*\d+%\s*,\s*\d+%\s*,\s*[\d.]+\s*\)$`, // hsla(h,s,l,a)
		`^hsl\(\s*[\d.]+\s+[\d.]+%\s+[\d.]+%\s*\)$`,           // hsl(h s l) - modern syntax
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, color); matched {
			return nil
		}
	}
	
	return NewValidationError("theme.colors."+name, color, "Invalid color format", "invalid_format")
}

// Helper functions for color manipulation (simplified implementations)
func lighten(color string, amount float64) string {
	// This is a placeholder - in a real implementation, you'd parse the color
	// and adjust its lightness
	return color
}

func darken(color string, amount float64) string {
	// This is a placeholder - in a real implementation, you'd parse the color
	// and adjust its darkness
	return color
}

// DarkTheme returns a default dark theme
func DarkTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "dark"
	theme.DisplayName = "Dark Theme"
	theme.Description = "Default dark theme for Gothic Forge"
	
	// Adjust colors for dark theme
	theme.Colors.Background = "hsl(240 10% 3.9%)"
	theme.Colors.Surface = "hsl(240 10% 3.9%)"
	theme.Colors.OnSurface = "hsl(0 0% 98%)"
	theme.Colors.Neutral = "hsl(240 3.7% 15.9%)"
	
	return theme
}