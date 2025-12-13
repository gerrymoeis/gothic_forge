package webawesome

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// This file contains examples and test scenarios for the component registry system
// Note: This is not a proper test file but demonstrates usage patterns

// ExampleRegistryUsage demonstrates basic registry operations
func ExampleRegistryUsage() {
	// Create a new registry with configuration
	config := DefaultConfig()
	config.Components.Registry = "./test-registry"
	
	manager := NewManager(config)
	
	// Initialize the registry
	if err := manager.Initialize(); err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		return
	}
	
	registry := manager.GetRegistry()
	
	// Register a custom component
	customButton := &ButtonComponent{
		BaseComponent: BaseComponent{
			ID:    "custom-btn",
			Class: "custom-button",
		},
		Variant: "primary",
		Size:    "medium",
	}
	
	if err := registry.Register("custom-button", customButton); err != nil {
		fmt.Printf("Failed to register component: %v\n", err)
		return
	}
	
	// List all components
	components := registry.List()
	fmt.Printf("Registered components: %v\n", components)
	
	// Get component info
	if info, err := registry.GetComponentInfo("button"); err == nil {
		fmt.Printf("Button component: %s (type: %s, usage: %d)\n", 
			info.Name, info.Type, info.UsageCount)
	}
	
	// Use a component (this updates usage statistics)
	if component, err := registry.Get("button"); err == nil {
		fmt.Printf("Retrieved component: %T\n", component)
	}
	
	// Get usage statistics
	usage := registry.GetUsage()
	fmt.Printf("Usage statistics: %v\n", usage)
}

// ExampleRegistryStats demonstrates statistics and analytics
func ExampleRegistryStats() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	
	// Simulate some usage
	for i := 0; i < 10; i++ {
		registry.Get("button")
		registry.Get("input")
	}
	
	for i := 0; i < 5; i++ {
		registry.Get("card")
	}
	
	// Get comprehensive statistics
	stats := registry.GetStats()
	
	fmt.Printf("Registry Statistics:\n")
	fmt.Printf("  Total Components: %d\n", stats.TotalComponents)
	fmt.Printf("  Total Usage: %d\n", stats.TotalUsage)
	fmt.Printf("  Components by System: %v\n", stats.ComponentsBySystem)
	fmt.Printf("  Components by Type: %v\n", stats.ComponentsByType)
	
	fmt.Printf("  Most Used Components:\n")
	for _, usage := range stats.MostUsed {
		fmt.Printf("    %s: %d uses\n", usage.Name, usage.UsageCount)
	}
	
	if len(stats.UnusedComponents) > 0 {
		fmt.Printf("  Unused Components: %v\n", stats.UnusedComponents)
	}
}

// ExampleRegistryCategories demonstrates component categorization
func ExampleRegistryCategories() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	
	// Get all categories
	categories := registry.GetCategories()
	
	fmt.Printf("Component Categories:\n")
	for _, category := range categories {
		fmt.Printf("  %s (%s): %d components\n", 
			category.Name, category.Description, len(category.Components))
		
		// Show first few components in each category
		for i, component := range category.Components {
			if i >= 3 { // Show first 3
				fmt.Printf("    ... and %d more\n", len(category.Components)-3)
				break
			}
			fmt.Printf("    - %s\n", component)
		}
	}
	
	// Get components in a specific category
	formComponents := registry.GetComponentsByCategory("form")
	fmt.Printf("\nForm Components: %v\n", formComponents)
}

// ExampleRegistrySearch demonstrates search functionality
func ExampleRegistrySearch() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	
	// Search for components
	searchQueries := []string{"button", "input", "form", "layout"}
	
	for _, query := range searchQueries {
		results := registry.Search(query)
		fmt.Printf("Search results for '%s': %d components\n", query, len(results))
		
		for _, result := range results {
			fmt.Printf("  - %s (%s)\n", result.Name, result.Type)
		}
		fmt.Println()
	}
}

// ExampleRegistryOptimization demonstrates optimization features
func ExampleRegistryOptimization() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	registryManager := manager.GetRegistryManager()
	
	// Simulate usage patterns
	for i := 0; i < 100; i++ {
		registry.Get("button") // Heavily used
	}
	
	for i := 0; i < 50; i++ {
		registry.Get("input") // Moderately used
	}
	
	// Get optimization hints
	hints := registry.GetOptimizationHints()
	fmt.Printf("Optimization Hints:\n")
	for i, hint := range hints {
		fmt.Printf("  %d. %s\n", i+1, hint)
	}
	
	// Perform optimization
	result, err := registryManager.OptimizeRegistry()
	if err != nil {
		fmt.Printf("Optimization failed: %v\n", err)
		return
	}
	
	fmt.Printf("\nOptimization Results:\n")
	fmt.Printf("  Actions Performed: %d\n", len(result.Actions))
	fmt.Printf("  Total Size Saved: %d bytes\n", result.TotalSizeSaved)
	
	for _, action := range result.Actions {
		fmt.Printf("  - %s: %s\n", action.Type, action.Description)
	}
}

// ExampleRegistryValidation demonstrates validation features
func ExampleRegistryValidation() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registryManager := manager.GetRegistryManager()
	
	// Validate the registry
	result, err := registryManager.ValidateRegistry()
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}
	
	fmt.Printf("Registry Validation Results:\n")
	fmt.Printf("  Valid: %t\n", result.IsValid)
	fmt.Printf("  Issues Found: %d\n", len(result.Issues))
	
	if len(result.Issues) > 0 {
		fmt.Printf("  Issues:\n")
		for _, issue := range result.Issues {
			fmt.Printf("    [%s] %s: %s\n", issue.Severity, issue.Type, issue.Description)
			if issue.Component != "" {
				fmt.Printf("      Component: %s\n", issue.Component)
			}
		}
	}
}

// ExampleRegistryPersistence demonstrates save/load functionality
func ExampleRegistryPersistence() {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "webawesome-registry-test")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)
	
	config := DefaultConfig()
	config.Components.Registry = tempDir
	
	// Create and initialize registry
	manager := NewManager(config)
	manager.Initialize()
	
	registry := manager.GetRegistry()
	registryManager := manager.GetRegistryManager()
	
	// Use some components to generate usage data
	registry.Get("button")
	registry.Get("input")
	registry.Get("card")
	
	// Create a backup
	backupPath := filepath.Join(tempDir, "backup.json")
	if err := registryManager.BackupRegistry(backupPath); err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		return
	}
	
	fmt.Printf("Registry backed up to: %s\n", backupPath)
	
	// Clear the registry
	registry.Clear()
	fmt.Printf("Registry cleared. Components: %d\n", len(registry.List()))
	
	// Restore from backup
	if err := registryManager.RestoreRegistry(backupPath); err != nil {
		fmt.Printf("Restore failed: %v\n", err)
		return
	}
	
	fmt.Printf("Registry restored. Components: %d\n", len(registry.List()))
}

// ExampleRegistryReport demonstrates comprehensive reporting
func ExampleRegistryReport() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	registryManager := manager.GetRegistryManager()
	
	// Simulate some activity
	components := []string{"button", "input", "card", "dialog", "table"}
	for _, comp := range components {
		for i := 0; i < (len(comp) * 3); i++ { // Variable usage
			registry.Get(comp)
		}
	}
	
	// Generate comprehensive report
	report, err := registryManager.GenerateReport()
	if err != nil {
		fmt.Printf("Report generation failed: %v\n", err)
		return
	}
	
	fmt.Printf("Registry Report (Generated: %s)\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Printf("================================\n\n")
	
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Total Components: %d\n", report.Stats.TotalComponents)
	fmt.Printf("  Total Usage: %d\n", report.Stats.TotalUsage)
	
	fmt.Printf("\nCategories:\n")
	for _, category := range report.Categories {
		fmt.Printf("  %s: %d components\n", category.Name, len(category.Components))
	}
	
	fmt.Printf("\nOptimization Hints:\n")
	for i, hint := range report.OptimizationHints {
		fmt.Printf("  %d. %s\n", i+1, hint)
	}
	
	fmt.Printf("\nComponent Details:\n")
	for name, info := range report.ComponentDetails {
		fmt.Printf("  %s: %d uses, last used %s\n", 
			name, info.UsageCount, info.LastUsed.Format("2006-01-02"))
	}
}

// ExampleRegistryCLI demonstrates CLI interface usage
func ExampleRegistryCLI() {
	manager := NewManager(DefaultConfig())
	manager.Initialize()
	
	registry := manager.GetRegistry()
	registryManager := manager.GetRegistryManager()
	cli := NewRegistryCLI(registryManager)
	
	// Simulate some usage
	registry.Get("button")
	registry.Get("input")
	registry.Get("card")
	
	fmt.Println("=== CLI Examples ===")
	
	// List components
	fmt.Println("1. List Components:")
	cli.ListComponents("simple")
	fmt.Println()
	
	// Show statistics
	fmt.Println("2. Show Statistics:")
	cli.ShowStats()
	fmt.Println()
	
	// Show component details
	fmt.Println("3. Component Details:")
	cli.ShowComponent("button")
	fmt.Println()
	
	// Search components
	fmt.Println("4. Search Components:")
	cli.SearchComponents("form")
	fmt.Println()
	
	// Show categories
	fmt.Println("5. Show Categories:")
	cli.ShowCategories()
	fmt.Println()
	
	// Show optimization hints
	fmt.Println("6. Optimization Hints:")
	cli.ShowOptimizationHints()
	fmt.Println()
	
	// Validate registry
	fmt.Println("7. Validate Registry:")
	cli.ValidateRegistry()
}

// ExampleRegistryIntegration demonstrates integration with the main manager
func ExampleRegistryIntegration() {
	// Create manager with custom configuration
	config := DefaultConfig()
	config.Components.Registry = "./registry-data"
	config.Components.AutoImport = true
	
	manager := NewManager(config)
	
	// Initialize everything
	if err := manager.Initialize(); err != nil {
		fmt.Printf("Initialization failed: %v\n", err)
		return
	}
	
	// Access different subsystems
	registry := manager.GetRegistry()
	registryManager := manager.GetRegistryManager()
	themeProvider := manager.GetThemeProvider()
	
	fmt.Printf("Manager initialized successfully\n")
	fmt.Printf("Registry components: %d\n", len(registry.List()))
	
	// Use registry with theme system
	theme, _ := themeProvider.LoadTheme("default")
	fmt.Printf("Default theme loaded: %s\n", theme.Name)
	
	// Generate comprehensive report
	report, _ := registryManager.GenerateReport()
	fmt.Printf("Report generated with %d components\n", len(report.ComponentDetails))
	
	// Show integration working
	fmt.Printf("Integration successful - all systems working together\n")
}

// RunAllExamples runs all registry examples
func RunAllExamples() {
	examples := []struct {
		name string
		fn   func()
	}{
		{"Basic Usage", ExampleRegistryUsage},
		{"Statistics", ExampleRegistryStats},
		{"Categories", ExampleRegistryCategories},
		{"Search", ExampleRegistrySearch},
		{"Optimization", ExampleRegistryOptimization},
		{"Validation", ExampleRegistryValidation},
		{"Persistence", ExampleRegistryPersistence},
		{"Reporting", ExampleRegistryReport},
		{"CLI Interface", ExampleRegistryCLI},
		{"Integration", ExampleRegistryIntegration},
	}
	
	for _, example := range examples {
		fmt.Print("\n" + strings.Repeat("=", 50) + "\n")
		fmt.Printf("Example: %s\n", example.name)
		fmt.Print(strings.Repeat("=", 50) + "\n")
		
		example.fn()
		
		fmt.Print("\n" + strings.Repeat("-", 50) + "\n")
		fmt.Printf("Example '%s' completed\n", example.name)
	}
}