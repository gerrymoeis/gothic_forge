package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RegistryManager provides high-level registry management functionality
type RegistryManager struct {
	registry *Registry
	config   *Config
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager(config *Config) *RegistryManager {
	registryConfig := RegistryConfig{
		PersistPath: filepath.Join(config.Components.Registry, "components.json"),
		AutoSave:    true,
		AutoLoad:    true,
	}
	
	return &RegistryManager{
		registry: NewRegistryWithConfig(registryConfig),
		config:   config,
	}
}

// Initialize sets up the registry with default components
func (rm *RegistryManager) Initialize() error {
	return rm.registry.Initialize()
}

// GetRegistry returns the underlying registry
func (rm *RegistryManager) GetRegistry() ComponentRegistry {
	return rm.registry
}

// GetConcreteRegistry returns the underlying registry as concrete type
func (rm *RegistryManager) GetConcreteRegistry() *Registry {
	return rm.registry
}

// GenerateReport creates a comprehensive registry report
func (rm *RegistryManager) GenerateReport() (*RegistryReport, error) {
	stats := rm.registry.GetStats()
	categories := rm.registry.GetCategories()
	hints := rm.registry.GetOptimizationHints()
	dependencies := rm.registry.GetDependencies()
	
	report := &RegistryReport{
		GeneratedAt:      time.Now(),
		Stats:            stats,
		Categories:       categories,
		OptimizationHints: hints,
		Dependencies:     dependencies,
		ComponentDetails: make(map[string]*ComponentInfo),
	}
	
	// Add detailed component information
	for _, name := range rm.registry.List() {
		if info, err := rm.registry.GetComponentInfo(name); err == nil {
			report.ComponentDetails[name] = info
		}
	}
	
	return report, nil
}

// ExportReport exports the registry report to a file
func (rm *RegistryManager) ExportReport(outputPath string) error {
	report, err := rm.GenerateReport()
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}
	
	return nil
}

// OptimizeRegistry performs optimization based on usage patterns
func (rm *RegistryManager) OptimizeRegistry() (*OptimizationResult, error) {
	result := &OptimizationResult{
		PerformedAt: time.Now(),
		Actions:     []OptimizationAction{},
	}
	
	stats := rm.registry.GetStats()
	
	// Remove unused components if configured
	if rm.config.Build.TreeShaking {
		for _, componentName := range stats.UnusedComponents {
			action := OptimizationAction{
				Type:        "remove_unused",
				Component:   componentName,
				Description: fmt.Sprintf("Removed unused component: %s", componentName),
				SizeSaved:   estimateComponentSize(componentName),
			}
			result.Actions = append(result.Actions, action)
			result.TotalSizeSaved += action.SizeSaved
		}
	}
	
	// Suggest caching for heavily used components
	for _, usageStat := range stats.MostUsed {
		if usageStat.UsageCount > 50 {
			action := OptimizationAction{
				Type:        "suggest_caching",
				Component:   usageStat.Name,
				Description: fmt.Sprintf("Consider caching heavily used component: %s (%d uses)", usageStat.Name, usageStat.UsageCount),
				SizeSaved:   0, // No immediate size savings
			}
			result.Actions = append(result.Actions, action)
		}
	}
	
	return result, nil
}

// ValidateRegistry checks the registry for consistency and issues
func (rm *RegistryManager) ValidateRegistry() (*ValidationResult, error) {
	result := &ValidationResult{
		ValidatedAt: time.Now(),
		Issues:      []ValidationIssue{},
		IsValid:     true,
	}
	
	// Check for missing dependencies
	dependencies := rm.registry.GetDependencies()
	for component, deps := range dependencies {
		for _, dep := range deps {
			if _, err := rm.registry.GetComponentInfo(dep); err != nil {
				issue := ValidationIssue{
					Type:        "missing_dependency",
					Component:   component,
					Description: fmt.Sprintf("Component %s depends on missing component: %s", component, dep),
					Severity:    "error",
				}
				result.Issues = append(result.Issues, issue)
				result.IsValid = false
			}
		}
	}
	
	// Check for circular dependencies
	if cycles := rm.findCircularDependencies(dependencies); len(cycles) > 0 {
		for _, cycle := range cycles {
			issue := ValidationIssue{
				Type:        "circular_dependency",
				Component:   cycle[0],
				Description: fmt.Sprintf("Circular dependency detected: %v", cycle),
				Severity:    "warning",
			}
			result.Issues = append(result.Issues, issue)
		}
	}
	
	// Check for components with no usage
	stats := rm.registry.GetStats()
	if len(stats.UnusedComponents) > 0 {
		issue := ValidationIssue{
			Type:        "unused_components",
			Component:   "",
			Description: fmt.Sprintf("%d components are unused and could be removed", len(stats.UnusedComponents)),
			Severity:    "info",
		}
		result.Issues = append(result.Issues, issue)
	}
	
	return result, nil
}

// BackupRegistry creates a backup of the current registry
func (rm *RegistryManager) BackupRegistry(backupPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(backupPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	backup := &RegistryBackup{
		CreatedAt:   time.Now(),
		Components:  make(map[string]*ComponentInfo),
		Usage:       rm.registry.GetUsage(),
		Categories:  rm.registry.GetCategories(),
		Stats:       rm.registry.GetStats(),
	}
	
	// Copy all component information
	for _, name := range rm.registry.List() {
		if info, err := rm.registry.GetComponentInfo(name); err == nil {
			backup.Components[name] = info
		}
	}
	
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	
	return nil
}

// RestoreRegistry restores the registry from a backup
func (rm *RegistryManager) RestoreRegistry(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}
	
	var backup RegistryBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup: %w", err)
	}
	
	// Clear current registry
	if err := rm.registry.Clear(); err != nil {
		return fmt.Errorf("failed to clear registry: %w", err)
	}
	
	// Restore components
	for name, info := range backup.Components {
		// Create a basic component instance
		component, err := rm.registry.createComponentByType(info.Type)
		if err != nil {
			continue // Skip components we can't create
		}
		
		if err := rm.registry.Register(name, component); err != nil {
			return fmt.Errorf("failed to restore component %s: %w", name, err)
		}
	}
	
	return nil
}

// Helper methods

// findCircularDependencies detects circular dependencies in the component graph
func (rm *RegistryManager) findCircularDependencies(dependencies map[string][]string) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)
		
		for _, dep := range dependencies[node] {
			if !visited[dep] {
				if dfs(dep, path) {
					return true
				}
			} else if recStack[dep] {
				// Found a cycle
				cycleStart := -1
				for i, p := range path {
					if p == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
				return true
			}
		}
		
		recStack[node] = false
		return false
	}
	
	for node := range dependencies {
		if !visited[node] {
			dfs(node, []string{})
		}
	}
	
	return cycles
}

// estimateComponentSize estimates the size impact of a component
func estimateComponentSize(componentName string) int {
	// Basic size estimation in bytes
	// In a real implementation, this would analyze actual component sizes
	baseSizes := map[string]int{
		"button":       1024,
		"input":        1536,
		"select":       2048,
		"textarea":     1024,
		"checkbox":     512,
		"radio":        512,
		"card":         2048,
		"dialog":       3072,
		"drawer":       3072,
		"tab-group":    4096,
		"table":        4096,
		"badge":        512,
		"progress-bar": 1024,
		"spinner":      512,
		"dropdown":     2048,
		"menu":         1536,
		"tooltip":      1024,
	}
	
	if size, exists := baseSizes[componentName]; exists {
		return size
	}
	
	return 1024 // Default size
}

// Data structures for registry management

// RegistryReport provides comprehensive information about the registry
type RegistryReport struct {
	GeneratedAt       time.Time                    `json:"generatedAt"`
	Stats             *RegistryStats               `json:"stats"`
	Categories        []ComponentCategory          `json:"categories"`
	OptimizationHints []OptimizationHint           `json:"optimizationHints"`
	Dependencies      map[string][]string          `json:"dependencies"`
	ComponentDetails  map[string]*ComponentInfo    `json:"componentDetails"`
}

// OptimizationResult contains the results of registry optimization
type OptimizationResult struct {
	PerformedAt     time.Time             `json:"performedAt"`
	Actions         []OptimizationAction  `json:"actions"`
	TotalSizeSaved  int                   `json:"totalSizeSaved"`
}

// OptimizationAction represents a single optimization action
type OptimizationAction struct {
	Type        string `json:"type"`
	Component   string `json:"component"`
	Description string `json:"description"`
	SizeSaved   int    `json:"sizeSaved"`
}

// ValidationResult contains the results of registry validation
type ValidationResult struct {
	ValidatedAt time.Time         `json:"validatedAt"`
	IsValid     bool              `json:"isValid"`
	Issues      []ValidationIssue `json:"issues"`
}

// ValidationIssue represents a validation problem
type ValidationIssue struct {
	Type        string `json:"type"`
	Component   string `json:"component"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // error, warning, info
}

// RegistryBackup contains a complete backup of the registry
type RegistryBackup struct {
	CreatedAt  time.Time                    `json:"createdAt"`
	Components map[string]*ComponentInfo    `json:"components"`
	Usage      map[string]int               `json:"usage"`
	Categories []ComponentCategory          `json:"categories"`
	Stats      *RegistryStats               `json:"stats"`
}