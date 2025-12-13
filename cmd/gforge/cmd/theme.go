package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gothicforge3/internal/webawesome"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage Web Awesome themes",
	Long:  "Create, customize, list, and apply Web Awesome themes for your application",
}

var themeInitCmd = &cobra.Command{
	Use:   "init <theme-name>",
	Short: "Initialize a new custom theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		themeName := args[0]
		
		if !isValidName(themeName) {
			return fmt.Errorf("invalid theme name: %s (use letters, numbers, dash, underscore)", themeName)
		}
		
		return initTheme(themeName)
	},
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		return listThemes()
	},
}

var themeApplyCmd = &cobra.Command{
	Use:   "apply <theme-name>",
	Short: "Apply a theme to your application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		themeName := args[0]
		return applyTheme(themeName)
	},
}

var themeCustomizeCmd = &cobra.Command{
	Use:   "customize <theme-name>",
	Short: "Customize an existing theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		themeName := args[0]
		return customizeTheme(themeName)
	},
}

func initTheme(name string) error {
	// Create Web Awesome config
	config := &webawesome.Config{
		Themes: webawesome.ThemeConfig{
			Default:    "default",
			CustomPath: "app/themes",
		},
	}
	
	// Create theme manager
	manager := webawesome.NewThemeManager(config)
	
	// Create new theme based on default
	defaultTheme := webawesome.DefaultTheme()
	newTheme := manager.CreateTheme(name, strings.Title(name), defaultTheme)
	
	// Save theme
	if err := manager.SaveTheme(newTheme); err != nil {
		return fmt.Errorf("failed to save theme: %w", err)
	}
	
	// Create theme directory if it doesn't exist
	themeDir := filepath.Join("app", "themes")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}
	
	// Create CSS file
	cssPath := filepath.Join(themeDir, name+".css")
	css, err := manager.GenerateCSS()
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}
	
	if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
		return fmt.Errorf("failed to write CSS file: %w", err)
	}
	
	fmt.Printf("Created theme: %s\n", name)
	fmt.Printf("  - Theme file: %s\n", filepath.Join(themeDir, name+".json"))
	fmt.Printf("  - CSS file: %s\n", cssPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Customize theme: gforge theme customize %s\n", name)
	fmt.Printf("  2. Apply theme: gforge theme apply %s\n", name)
	fmt.Println("  3. Include CSS in your layout template")
	
	return nil
}

func listThemes() error {
	// Create Web Awesome config
	config := &webawesome.Config{
		Themes: webawesome.ThemeConfig{
			Default:    "default",
			CustomPath: "app/themes",
		},
	}
	
	fmt.Println("Available themes:")
	fmt.Println()
	
	// List built-in themes
	fmt.Println("Built-in themes:")
	builtinThemes := []string{"default", "dark", "light"}
	for _, theme := range builtinThemes {
		fmt.Printf("  • %s\n", theme)
	}
	
	// List custom themes
	customDir := config.Themes.CustomPath
	if _, err := os.Stat(customDir); err == nil {
		entries, err := os.ReadDir(customDir)
		if err != nil {
			return fmt.Errorf("failed to read themes directory: %w", err)
		}
		
		customThemes := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				themeName := strings.TrimSuffix(entry.Name(), ".json")
				customThemes = append(customThemes, themeName)
			}
		}
		
		if len(customThemes) > 0 {
			fmt.Println()
			fmt.Println("Custom themes:")
			for _, theme := range customThemes {
				fmt.Printf("  • %s\n", theme)
			}
		}
	}
	
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gforge theme apply <theme-name>     - Apply a theme")
	fmt.Println("  gforge theme init <theme-name>      - Create new theme")
	fmt.Println("  gforge theme customize <theme-name> - Customize existing theme")
	
	return nil
}

func applyTheme(name string) error {
	// Create Web Awesome config
	config := &webawesome.Config{
		Themes: webawesome.ThemeConfig{
			Default:    name,
			CustomPath: "app/themes",
		},
	}
	
	// Create theme manager
	manager := webawesome.NewThemeManager(config)
	
	// Try to load the theme to validate it exists
	if err := manager.SetTheme(name); err != nil {
		return fmt.Errorf("failed to load theme '%s': %w", name, err)
	}
	
	// Generate CSS for the theme
	css, err := manager.GenerateCSS()
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}
	
	// Create themes directory if it doesn't exist
	themeDir := filepath.Join("app", "themes")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}
	
	// Write active theme CSS
	cssPath := filepath.Join(themeDir, "active.css")
	if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
		return fmt.Errorf("failed to write CSS file: %w", err)
	}
	
	// Update config file
	configPath := filepath.Join("app", "themes", "config.json")
	configData := map[string]interface{}{
		"activeTheme": name,
		"lastApplied": "now",
	}
	
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	fmt.Printf("Applied theme: %s\n", name)
	fmt.Printf("  - CSS generated: %s\n", cssPath)
	fmt.Printf("  - Config updated: %s\n", configPath)
	fmt.Println()
	fmt.Println("Make sure to include the CSS in your layout template:")
	fmt.Println(`  <link rel="stylesheet" href="/static/themes/active.css">`)
	
	return nil
}

func customizeTheme(name string) error {
	// Create Web Awesome config
	config := &webawesome.Config{
		Themes: webawesome.ThemeConfig{
			Default:    "default",
			CustomPath: "app/themes",
		},
	}
	
	// Create theme manager
	manager := webawesome.NewThemeManager(config)
	
	// Try to load the theme
	if err := manager.SetTheme(name); err != nil {
		return fmt.Errorf("failed to load theme '%s': %w", name, err)
	}
	
	theme := manager.GetCurrentTheme()
	if theme == nil {
		return fmt.Errorf("theme '%s' not found", name)
	}
	
	// Create customization template
	customPath := filepath.Join("app", "themes", name+"_custom.json")
	
	// Create a customization template with common properties
	customization := map[string]interface{}{
		"name":        name + "_custom",
		"displayName": theme.DisplayName + " (Custom)",
		"description": "Customized version of " + theme.DisplayName,
		"version":     "1.0.0",
		"colors": map[string]string{
			"primary":   theme.Colors.Primary,
			"secondary": theme.Colors.Secondary,
			"success":   theme.Colors.Success,
			"warning":   theme.Colors.Warning,
			"danger":    theme.Colors.Danger,
			"neutral":   theme.Colors.Neutral,
		},
		"typography": map[string]interface{}{
			"fontFamily": theme.Typography.FontFamily,
			"fontSize":   theme.Typography.FontSize,
			"lineHeight": theme.Typography.LineHeight,
		},
		"custom": map[string]string{
			"// Add your custom CSS properties here": "",
			"--custom-border-radius": "8px",
			"--custom-shadow":        "0 2px 4px rgba(0,0,0,0.1)",
		},
	}
	
	customJSON, err := json.MarshalIndent(customization, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal customization: %w", err)
	}
	
	if err := os.WriteFile(customPath, customJSON, 0644); err != nil {
		return fmt.Errorf("failed to write customization file: %w", err)
	}
	
	fmt.Printf("Created customization template: %s\n", name+"_custom")
	fmt.Printf("  - Template file: %s\n", customPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit the customization file to modify colors, typography, and custom properties")
	fmt.Printf("  2. Apply the customized theme: gforge theme apply %s_custom\n", name)
	fmt.Println()
	fmt.Println("Available customization options:")
	fmt.Println("  • colors.primary, secondary, success, warning, danger, info")
	fmt.Println("  • typography.fontFamily, fontSize, lineHeight")
	fmt.Println("  • custom properties (CSS custom properties)")
	
	return nil
}

func init() {
	themeCmd.AddCommand(themeInitCmd)
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeApplyCmd)
	themeCmd.AddCommand(themeCustomizeCmd)
	rootCmd.AddCommand(themeCmd)
}