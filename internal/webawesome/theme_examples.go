package webawesome

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// This file contains simple examples of using the theme system

// ExampleBasicThemeUsage demonstrates basic theme operations
func ExampleBasicThemeUsage() {
	// Create a manager
	config := DefaultConfig()
	config.Themes.CustomPath = "./themes"
	
	manager := NewManager(config)
	themeManager := manager.GetThemeManager()
	
	fmt.Println("=== Basic Theme Usage ===")
	
	// Get current theme
	currentTheme := themeManager.GetCurrentTheme()
	fmt.Printf("Current theme: %s (%s)\n", currentTheme.Name, currentTheme.DisplayName)
	
	// List available themes
	themes := themeManager.ListAvailableThemes()
	fmt.Printf("Available themes: %v\n", themes)
	
	// Switch to dark theme
	if err := themeManager.SetTheme("dark"); err != nil {
		fmt.Printf("Error switching theme: %v\n", err)
	} else {
		fmt.Println("Switched to dark theme")
	}
	
	// Generate CSS for current theme
	css, err := themeManager.GenerateCSS()
	if err != nil {
		fmt.Printf("Error generating CSS: %v\n", err)
	} else {
		fmt.Printf("Generated CSS (%d characters)\n", len(css))
	}
}

// ExampleCreateCustomTheme demonstrates creating and saving custom themes
func ExampleCreateCustomTheme() {
	config := DefaultConfig()
	config.Themes.CustomPath = "./themes"
	
	manager := NewManager(config)
	themeManager := manager.GetThemeManager()
	
	fmt.Println("\n=== Custom Theme Creation ===")
	
	// Create a custom theme based on default
	baseTheme := DefaultTheme()
	customTheme := themeManager.CreateTheme("my-brand", "My Brand Theme", baseTheme)
	
	// Customize colors
	customTheme.Colors.Primary = "hsl(210 100% 50%)"
	customTheme.Colors.Success = "hsl(120 60% 45%)"
	
	// Add custom properties
	themeManager.SetTheme("my-brand") // Switch to our theme (in memory)
	themeManager.AddCustomProperty("brand-accent", "#ff6b35")
	themeManager.AddCustomProperty("brand-text", "#2c3e50")
	
	// Save the theme
	if err := themeManager.SaveTheme(customTheme); err != nil {
		fmt.Printf("Error saving theme: %v\n", err)
	} else {
		fmt.Printf("Saved custom theme: %s\n", customTheme.Name)
	}
	
	// Generate CSS for the custom theme
	css, _ := themeManager.GenerateThemeCSS("my-brand")
	fmt.Printf("Custom theme CSS preview:\n%s\n", css[:200]+"...")
}

// ExampleThemeCustomization demonstrates theme customization helpers
func ExampleThemeCustomization() {
	config := DefaultConfig()
	manager := NewManager(config)
	themeManager := manager.GetThemeManager()
	
	fmt.Println("\n=== Theme Customization ===")
	
	// Start with default theme
	fmt.Printf("Starting with: %s\n", themeManager.GetCurrentTheme().DisplayName)
	
	// Change primary color
	if err := themeManager.SetPrimaryColor("hsl(270 70% 50%)"); err != nil {
		fmt.Printf("Error setting primary color: %v\n", err)
	} else {
		fmt.Println("Changed primary color to purple")
	}
	
	// Change multiple colors at once
	colors := map[string]string{
		"success": "hsl(150 70% 40%)",
		"warning": "hsl(45 100% 55%)",
		"danger":  "hsl(0 80% 55%)",
	}
	
	if err := themeManager.SetColors(colors); err != nil {
		fmt.Printf("Error setting colors: %v\n", err)
	} else {
		fmt.Println("Updated semantic colors")
	}
	
	// Change font family
	themeManager.SetFontFamily("'Inter', system-ui, sans-serif")
	fmt.Println("Changed font family to Inter")
	
	// Add custom properties
	themeManager.AddCustomProperty("header-height", "4rem")
	themeManager.AddCustomProperty("sidebar-width", "16rem")
	fmt.Println("Added custom layout properties")
	
	// Show final theme info
	theme := themeManager.GetCurrentTheme()
	fmt.Printf("Final theme - Primary: %s, Font: %s\n", 
		theme.Colors.Primary, theme.Typography.FontFamily)
}

// ExamplePresetThemes demonstrates using preset themes
func ExamplePresetThemes() {
	fmt.Println("\n=== Preset Themes ===")
	
	presets := GetPresetThemes()
	utils := NewThemeUtils()
	
	for name, theme := range presets {
		_ = name // Use the variable to avoid unused warning
		fmt.Printf("\n%s Theme:\n", theme.DisplayName)
		fmt.Printf("  Primary Color: %s\n", theme.Colors.Primary)
		
		// Generate CSS
		css, err := utils.GetThemeCSS(theme)
		if err != nil {
			fmt.Printf("  Error generating CSS: %v\n", err)
		} else {
			fmt.Printf("  CSS Length: %d characters\n", len(css))
		}
		
		// Validate theme
		if err := utils.ValidateTheme(theme); err != nil {
			fmt.Printf("  Validation Error: %v\n", err)
		} else {
			fmt.Printf("  ✓ Valid theme\n")
		}
	}
}

// ExampleThemeUtils demonstrates theme utility functions
func ExampleThemeUtils() {
	fmt.Println("\n=== Theme Utilities ===")
	
	utils := NewThemeUtils()
	
	// Create a quick theme
	quickTheme := utils.CreateQuickTheme(
		"quick-blue", 
		"Quick Blue Theme", 
		"hsl(210 100% 50%)", 
		"hsl(210 20% 98%)",
	)
	
	fmt.Printf("Created quick theme: %s\n", quickTheme.DisplayName)
	
	// Clone a theme
	cloned := utils.CloneTheme(quickTheme)
	cloned.Name = "cloned-blue"
	cloned.DisplayName = "Cloned Blue Theme"
	
	fmt.Printf("Cloned theme: %s\n", cloned.DisplayName)
	
	// Merge themes
	darkTheme := DarkTheme()
	merged := utils.MergeThemes(quickTheme, darkTheme)
	
	fmt.Printf("Merged theme: %s with background %s\n", 
		merged.DisplayName, merged.Colors.Background)
	
	// Write theme CSS to file
	tempDir := filepath.Join(os.TempDir(), "webawesome-themes")
	cssPath := filepath.Join(tempDir, "quick-blue.css")
	
	if err := utils.WriteThemeCSS(quickTheme, cssPath); err != nil {
		fmt.Printf("Error writing CSS: %v\n", err)
	} else {
		fmt.Printf("Wrote theme CSS to: %s\n", cssPath)
		
		// Clean up
		os.RemoveAll(tempDir)
	}
}

// ExampleThemePreview demonstrates generating theme previews
func ExampleThemePreview() {
	fmt.Println("\n=== Theme Preview ===")
	
	utils := NewThemeUtils()
	
	// Create a sample theme
	theme := CreatePurpleTheme()
	
	// Generate HTML preview
	html := utils.GetThemePreview(theme)
	
	// Write preview to file
	tempDir := filepath.Join(os.TempDir(), "webawesome-preview")
	previewPath := filepath.Join(tempDir, "purple-theme-preview.html")
	
	os.MkdirAll(tempDir, 0755)
	if err := os.WriteFile(previewPath, []byte(html), 0644); err != nil {
		fmt.Printf("Error writing preview: %v\n", err)
	} else {
		fmt.Printf("Generated theme preview: %s\n", previewPath)
		fmt.Printf("Open in browser to see the theme preview\n")
		
		// Clean up after a moment (in real usage, you'd keep this)
		defer os.RemoveAll(tempDir)
	}
}

// ExampleThemeIntegration demonstrates integration with the main manager
func ExampleThemeIntegration() {
	fmt.Println("\n=== Theme Integration ===")
	
	// Create manager with custom config
	config := DefaultConfig()
	config.Themes.Default = "dark"
	config.Themes.CustomPath = "./custom-themes"
	
	manager := NewManager(config)
	
	// Initialize everything
	if err := manager.Initialize(); err != nil {
		fmt.Printf("Initialization failed: %v\n", err)
		return
	}
	
	// Access theme system
	themeManager := manager.GetThemeManager()
	
	// Show current theme (should be dark due to config)
	currentTheme := themeManager.GetCurrentTheme()
	fmt.Printf("Current theme: %s\n", currentTheme.DisplayName)
	
	// Create a brand theme
	brandTheme := themeManager.CreateTheme("company-brand", "Company Brand", nil)
	brandTheme.Colors.Primary = "hsl(200 100% 40%)"
	brandTheme.Colors.Secondary = "hsl(200 20% 80%)"
	
	// Switch to brand theme
	themeManager.SetTheme("company-brand")
	
	// Generate CSS for use in templates
	css, _ := themeManager.GenerateCSS()
	fmt.Printf("Generated brand theme CSS (%d chars)\n", len(css))
	
	fmt.Println("Theme system fully integrated with manager")
}

// RunAllThemeExamples runs all theme examples
func RunAllThemeExamples() {
	examples := []func(){
		ExampleBasicThemeUsage,
		ExampleCreateCustomTheme,
		ExampleThemeCustomization,
		ExamplePresetThemes,
		ExampleThemeUtils,
		ExampleThemePreview,
		ExampleThemeIntegration,
	}
	
	for _, example := range examples {
		example()
	}
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("All theme examples completed!")
}