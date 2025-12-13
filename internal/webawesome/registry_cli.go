package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// RegistryCLI provides command-line interface for registry operations
type RegistryCLI struct {
	manager *RegistryManager
}

// NewRegistryCLI creates a new registry CLI
func NewRegistryCLI(manager *RegistryManager) *RegistryCLI {
	return &RegistryCLI{
		manager: manager,
	}
}

// ListComponents displays all registered components
func (cli *RegistryCLI) ListComponents(format string) error {
	registry := cli.manager.GetRegistry()
	components := registry.List()
	
	if len(components) == 0 {
		fmt.Println("No components registered.")
		return nil
	}
	
	switch format {
	case "json":
		return cli.listComponentsJSON(components)
	case "table":
		return cli.listComponentsTable(components)
	default:
		return cli.listComponentsSimple(components)
	}
}

// ShowStats displays registry statistics
func (cli *RegistryCLI) ShowStats() error {
	registry := cli.manager.GetRegistry()
	stats := registry.GetStats()
	
	fmt.Printf("Registry Statistics\n")
	fmt.Printf("===================\n\n")
	
	fmt.Printf("Total Components: %d\n", stats.TotalComponents)
	fmt.Printf("Total Usage: %d\n", stats.TotalUsage)
	fmt.Printf("Generated At: %s\n\n", stats.GeneratedAt.Format(time.RFC3339))
	
	// Components by system
	fmt.Printf("Components by System:\n")
	for system, count := range stats.ComponentsBySystem {
		fmt.Printf("  %s: %d\n", system, count)
	}
	fmt.Println()
	
	// Components by type
	fmt.Printf("Components by Type:\n")
	for componentType, count := range stats.ComponentsByType {
		fmt.Printf("  %s: %d\n", componentType, count)
	}
	fmt.Println()
	
	// Most used components
	if len(stats.MostUsed) > 0 {
		fmt.Printf("Most Used Components:\n")
		for i, usage := range stats.MostUsed {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("  %s: %d uses\n", usage.Name, usage.UsageCount)
		}
		fmt.Println()
	}
	
	// Unused components
	if len(stats.UnusedComponents) > 0 {
		fmt.Printf("Unused Components (%d):\n", len(stats.UnusedComponents))
		for i, name := range stats.UnusedComponents {
			if i >= 10 { // Show first 10
				fmt.Printf("  ... and %d more\n", len(stats.UnusedComponents)-10)
				break
			}
			fmt.Printf("  %s\n", name)
		}
		fmt.Println()
	}
	
	return nil
}

// ShowComponent displays detailed information about a specific component
func (cli *RegistryCLI) ShowComponent(name string) error {
	registry := cli.manager.GetRegistry()
	info, err := registry.GetComponentInfo(name)
	if err != nil {
		return fmt.Errorf("component not found: %s", name)
	}
	
	fmt.Printf("Component: %s\n", info.Name)
	fmt.Printf("===========%s\n\n", strings.Repeat("=", len(info.Name)))
	
	fmt.Printf("Type: %s\n", info.Type)
	fmt.Printf("System: %s\n", info.System)
	fmt.Printf("Usage Count: %d\n", info.UsageCount)
	fmt.Printf("Last Used: %s\n", info.LastUsed.Format(time.RFC3339))
	
	if info.FilePath != "" {
		fmt.Printf("File Path: %s\n", info.FilePath)
	}
	
	if len(info.Dependencies) > 0 {
		fmt.Printf("\nDependencies:\n")
		for _, dep := range info.Dependencies {
			fmt.Printf("  - %s\n", dep)
		}
	}
	
	if len(info.Props) > 0 {
		fmt.Printf("\nProperties:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tType\tRequired\tDefault\tDescription")
		fmt.Fprintln(w, "----\t----\t--------\t-------\t-----------")
		
		for _, prop := range info.Props {
			required := "No"
			if prop.Required {
				required = "Yes"
			}
			
			defaultVal := ""
			if prop.Default != nil {
				defaultVal = fmt.Sprintf("%v", prop.Default)
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				prop.Name, prop.Type, required, defaultVal, prop.Description)
		}
		w.Flush()
	}
	
	if len(info.Examples) > 0 {
		fmt.Printf("\nExamples:\n")
		for _, example := range info.Examples {
			fmt.Printf("\n%s:\n", example.Name)
			if example.Description != "" {
				fmt.Printf("  Description: %s\n", example.Description)
			}
			fmt.Printf("  Code:\n%s\n", indentCode(example.Code, "    "))
		}
	}
	
	return nil
}

// SearchComponents searches for components matching a query
func (cli *RegistryCLI) SearchComponents(query string) error {
	registry := cli.manager.GetRegistry()
	results := registry.Search(query)
	
	if len(results) == 0 {
		fmt.Printf("No components found matching '%s'\n", query)
		return nil
	}
	
	fmt.Printf("Found %d component(s) matching '%s':\n\n", len(results), query)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tSystem\tUsage\tLast Used")
	fmt.Fprintln(w, "----\t----\t------\t-----\t---------")
	
	for _, info := range results {
		lastUsed := "Never"
		if !info.LastUsed.IsZero() {
			lastUsed = info.LastUsed.Format("2006-01-02")
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			info.Name, info.Type, info.System, info.UsageCount, lastUsed)
	}
	
	w.Flush()
	return nil
}

// ShowCategories displays all component categories
func (cli *RegistryCLI) ShowCategories() error {
	registry := cli.manager.GetRegistry()
	categories := registry.GetCategories()
	
	if len(categories) == 0 {
		fmt.Println("No categories defined.")
		return nil
	}
	
	fmt.Printf("Component Categories\n")
	fmt.Printf("===================\n\n")
	
	for _, category := range categories {
		fmt.Printf("%s (%d components)\n", category.Name, len(category.Components))
		if category.Description != "" {
			fmt.Printf("  Description: %s\n", category.Description)
		}
		
		// Show first few components
		for i, component := range category.Components {
			if i >= 5 { // Show first 5
				fmt.Printf("  ... and %d more\n", len(category.Components)-5)
				break
			}
			fmt.Printf("  - %s\n", component)
		}
		fmt.Println()
	}
	
	return nil
}

// ShowOptimizationHints displays optimization suggestions
func (cli *RegistryCLI) ShowOptimizationHints() error {
	registry := cli.manager.GetRegistry()
	hints := registry.GetOptimizationHints()
	
	if len(hints) == 0 {
		fmt.Println("No optimization hints available. Your registry is well optimized!")
		return nil
	}
	
	fmt.Printf("Optimization Hints\n")
	fmt.Printf("==================\n\n")
	
	for i, hint := range hints {
		fmt.Printf("%d. %s\n", i+1, hint)
	}
	
	return nil
}

// GenerateReport creates and displays a comprehensive registry report
func (cli *RegistryCLI) GenerateReport(outputPath string) error {
	report, err := cli.manager.GenerateReport()
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	if outputPath != "" {
		// Export to file
		if err := cli.manager.ExportReport(outputPath); err != nil {
			return fmt.Errorf("failed to export report: %w", err)
		}
		fmt.Printf("Report exported to: %s\n", outputPath)
		return nil
	}
	
	// Display report summary
	fmt.Printf("Registry Report\n")
	fmt.Printf("===============\n\n")
	
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Printf("Total Components: %d\n", report.Stats.TotalComponents)
	fmt.Printf("Total Usage: %d\n", report.Stats.TotalUsage)
	fmt.Printf("Categories: %d\n", len(report.Categories))
	fmt.Printf("Optimization Hints: %d\n\n", len(report.OptimizationHints))
	
	// Show top categories
	fmt.Printf("Categories:\n")
	for _, category := range report.Categories {
		fmt.Printf("  %s: %d components\n", category.Name, len(category.Components))
	}
	fmt.Println()
	
	// Show optimization hints
	if len(report.OptimizationHints) > 0 {
		fmt.Printf("Optimization Hints:\n")
		for i, hint := range report.OptimizationHints {
			if i >= 3 { // Show first 3
				fmt.Printf("  ... and %d more\n", len(report.OptimizationHints)-3)
				break
			}
			fmt.Printf("  - %s\n", hint)
		}
	}
	
	return nil
}

// ValidateRegistry validates the registry and displays results
func (cli *RegistryCLI) ValidateRegistry() error {
	result, err := cli.manager.ValidateRegistry()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	fmt.Printf("Registry Validation\n")
	fmt.Printf("==================\n\n")
	
	fmt.Printf("Validated: %s\n", result.ValidatedAt.Format(time.RFC3339))
	fmt.Printf("Status: ")
	if result.IsValid {
		fmt.Printf("✓ Valid\n")
	} else {
		fmt.Printf("✗ Issues Found\n")
	}
	fmt.Printf("Issues: %d\n\n", len(result.Issues))
	
	if len(result.Issues) > 0 {
		// Group issues by severity
		errors := []ValidationIssue{}
		warnings := []ValidationIssue{}
		infos := []ValidationIssue{}
		
		for _, issue := range result.Issues {
			switch issue.Severity {
			case "error":
				errors = append(errors, issue)
			case "warning":
				warnings = append(warnings, issue)
			case "info":
				infos = append(infos, issue)
			}
		}
		
		// Display errors
		if len(errors) > 0 {
			fmt.Printf("Errors (%d):\n", len(errors))
			for _, issue := range errors {
				fmt.Printf("  ✗ %s\n", issue.Description)
				if issue.Component != "" {
					fmt.Printf("    Component: %s\n", issue.Component)
				}
			}
			fmt.Println()
		}
		
		// Display warnings
		if len(warnings) > 0 {
			fmt.Printf("Warnings (%d):\n", len(warnings))
			for _, issue := range warnings {
				fmt.Printf("  ⚠ %s\n", issue.Description)
				if issue.Component != "" {
					fmt.Printf("    Component: %s\n", issue.Component)
				}
			}
			fmt.Println()
		}
		
		// Display info
		if len(infos) > 0 {
			fmt.Printf("Information (%d):\n", len(infos))
			for _, issue := range infos {
				fmt.Printf("  ℹ %s\n", issue.Description)
				if issue.Component != "" {
					fmt.Printf("    Component: %s\n", issue.Component)
				}
			}
		}
	}
	
	return nil
}

// Helper methods

// listComponentsJSON displays components in JSON format
func (cli *RegistryCLI) listComponentsJSON(components []string) error {
	registry := cli.manager.GetRegistry()
	var componentInfos []*ComponentInfo
	
	for _, name := range components {
		if info, err := registry.GetComponentInfo(name); err == nil {
			componentInfos = append(componentInfos, info)
		}
	}
	
	data, err := json.MarshalIndent(componentInfos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}
	
	fmt.Println(string(data))
	return nil
}

// listComponentsTable displays components in table format
func (cli *RegistryCLI) listComponentsTable(components []string) error {
	registry := cli.manager.GetRegistry()
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tSystem\tUsage\tLast Used")
	fmt.Fprintln(w, "----\t----\t------\t-----\t---------")
	
	// Sort components by name
	sort.Strings(components)
	
	for _, name := range components {
		info, err := registry.GetComponentInfo(name)
		if err != nil {
			continue
		}
		
		lastUsed := "Never"
		if !info.LastUsed.IsZero() {
			lastUsed = info.LastUsed.Format("2006-01-02")
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			info.Name, info.Type, info.System, info.UsageCount, lastUsed)
	}
	
	w.Flush()
	return nil
}

// listComponentsSimple displays components in simple list format
func (cli *RegistryCLI) listComponentsSimple(components []string) error {
	registry := cli.manager.GetRegistry()
	
	// Sort components by name
	sort.Strings(components)
	
	fmt.Printf("Registered Components (%d):\n\n", len(components))
	
	for _, name := range components {
		info, err := registry.GetComponentInfo(name)
		if err != nil {
			fmt.Printf("  %s (error loading info)\n", name)
			continue
		}
		
		fmt.Printf("  %s (%s) - %d uses\n", info.Name, info.Type, info.UsageCount)
	}
	
	return nil
}

// indentCode indents code with the given prefix
func indentCode(code, indent string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}