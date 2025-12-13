package webawesome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

// ComponentInfo holds metadata about a registered component
type ComponentInfo struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	System       string            `json:"system"` // webawesome, daisyui
	FilePath     string            `json:"filePath"`
	Dependencies []string          `json:"dependencies"`
	Props        []ComponentProp   `json:"props"`
	Examples     []ComponentExample `json:"examples"`
	UsageCount   int              `json:"usageCount"`
	LastUsed     time.Time        `json:"lastUsed"`
}

// ComponentProp describes a component property
type ComponentProp struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"`
}

// ComponentExample provides usage examples for a component
type ComponentExample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Preview     string `json:"preview,omitempty"`
}

// Registry provides a comprehensive implementation of ComponentRegistry
type Registry struct {
	components    map[string]*ComponentInfo
	usage         map[string]int
	factories     map[string]ComponentFactory
	dependencies  map[string][]string
	categories    map[string][]string
	mu            sync.RWMutex
	persistPath   string
	autoSave      bool
	initialized   bool
}

// ComponentFactory is a function that creates a component instance
type ComponentFactory func() Component

// RegistryStats provides statistics about the component registry
type RegistryStats struct {
	TotalComponents    int                    `json:"totalComponents"`
	ComponentsBySystem map[string]int         `json:"componentsBySystem"`
	ComponentsByType   map[string]int         `json:"componentsByType"`
	MostUsed           []ComponentUsageStats  `json:"mostUsed"`
	RecentlyUsed       []ComponentUsageStats  `json:"recentlyUsed"`
	UnusedComponents   []string               `json:"unusedComponents"`
	TotalUsage         int                    `json:"totalUsage"`
	GeneratedAt        time.Time              `json:"generatedAt"`
}

// ComponentUsageStats provides usage statistics for a component
type ComponentUsageStats struct {
	Name       string    `json:"name"`
	UsageCount int       `json:"usageCount"`
	LastUsed   time.Time `json:"lastUsed"`
}

// ComponentCategory represents a category of components
type ComponentCategory struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Components  []string `json:"components"`
}

// RegistryConfig configures the component registry
type RegistryConfig struct {
	PersistPath string `json:"persistPath"`
	AutoSave    bool   `json:"autoSave"`
	AutoLoad    bool   `json:"autoLoad"`
}

// NewRegistry creates a new component registry
func NewRegistry() *Registry {
	return NewRegistryWithConfig(RegistryConfig{
		AutoSave: true,
		AutoLoad: true,
	})
}

// NewRegistryWithConfig creates a new component registry with configuration
func NewRegistryWithConfig(config RegistryConfig) *Registry {
	registry := &Registry{
		components:   make(map[string]*ComponentInfo),
		usage:        make(map[string]int),
		factories:    make(map[string]ComponentFactory),
		dependencies: make(map[string][]string),
		categories:   make(map[string][]string),
		persistPath:  config.PersistPath,
		autoSave:     config.AutoSave,
		initialized:  false,
	}
	
	if config.AutoLoad && config.PersistPath != "" {
		registry.Load()
	}
	
	return registry
}

// Initialize sets up the registry with default Web Awesome components
func (r *Registry) Initialize() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.initialized {
		return nil
	}
	
	// Register default Web Awesome components
	r.registerDefaultComponents()
	
	// Set up component categories
	r.setupCategories()
	
	r.initialized = true
	
	if r.autoSave {
		return r.saveUnsafe()
	}
	
	return nil
}

// Register adds a component to the registry
func (r *Registry) Register(name string, component Component) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Extract component metadata
	info, err := r.extractComponentInfo(name, component)
	if err != nil {
		return fmt.Errorf("failed to extract component info: %w", err)
	}
	
	r.components[name] = info
	
	// Initialize usage tracking
	if _, exists := r.usage[name]; !exists {
		r.usage[name] = 0
	}
	
	if r.autoSave {
		return r.saveUnsafe()
	}
	
	return nil
}

// RegisterWithFactory adds a component factory to the registry
func (r *Registry) RegisterWithFactory(name string, factory ComponentFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create a component instance to extract metadata
	component := factory()
	info, err := r.extractComponentInfo(name, component)
	if err != nil {
		return fmt.Errorf("failed to extract component info: %w", err)
	}
	
	r.components[name] = info
	r.factories[name] = factory
	
	// Initialize usage tracking
	if _, exists := r.usage[name]; !exists {
		r.usage[name] = 0
	}
	
	if r.autoSave {
		return r.saveUnsafe()
	}
	
	return nil
}

// Get retrieves a component from the registry
func (r *Registry) Get(name string) (Component, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	info, exists := r.components[name]
	if !exists {
		return nil, &RegistryError{
			Type:    "component_not_found",
			Message: "Component not found: " + name,
		}
	}
	
	// Update usage tracking
	r.usage[name]++
	info.UsageCount++
	info.LastUsed = time.Now()
	
	// Create component instance using factory if available
	if factory, hasFactory := r.factories[name]; hasFactory {
		component := factory()
		
		if r.autoSave {
			r.saveUnsafe()
		}
		
		return component, nil
	}
	
	// Return a basic component instance based on type
	component, err := r.createComponentByType(info.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}
	
	if r.autoSave {
		r.saveUnsafe()
	}
	
	return component, nil
}

// List returns all registered component names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

// GetUsage returns usage statistics for all components
func (r *Registry) GetUsage() map[string]int {
	usage := make(map[string]int)
	for name, count := range r.usage {
		usage[name] = count
	}
	return usage
}

// GetComponentInfo returns detailed information about a component
func (r *Registry) GetComponentInfo(name string) (*ComponentInfo, error) {
	info, exists := r.components[name]
	if !exists {
		return nil, &RegistryError{
			Type:    "component_not_found",
			Message: "Component not found: " + name,
		}
	}
	return info, nil
}

// UpdateUsage increments the usage count for a component
func (r *Registry) UpdateUsage(name string) error {
	info, exists := r.components[name]
	if !exists {
		return &RegistryError{
			Type:    "component_not_found",
			Message: "Component not found: " + name,
		}
	}
	
	r.usage[name]++
	info.UsageCount++
	info.LastUsed = time.Now()
	return nil
}

// RegistryError represents errors that occur in the component registry
type RegistryError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *RegistryError) Error() string {
	return e.Message
}

// Enhanced registry methods

// GetStats returns comprehensive statistics about the registry
func (r *Registry) GetStats() *RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stats := &RegistryStats{
		TotalComponents:    len(r.components),
		ComponentsBySystem: make(map[string]int),
		ComponentsByType:   make(map[string]int),
		MostUsed:           []ComponentUsageStats{},
		RecentlyUsed:       []ComponentUsageStats{},
		UnusedComponents:   []string{},
		TotalUsage:         0,
		GeneratedAt:        time.Now(),
	}
	
	var usageStats []ComponentUsageStats
	
	for name, info := range r.components {
		// Count by system
		stats.ComponentsBySystem[info.System]++
		
		// Count by type
		stats.ComponentsByType[info.Type]++
		
		// Track usage
		stats.TotalUsage += info.UsageCount
		
		usageStat := ComponentUsageStats{
			Name:       name,
			UsageCount: info.UsageCount,
			LastUsed:   info.LastUsed,
		}
		usageStats = append(usageStats, usageStat)
		
		// Track unused components
		if info.UsageCount == 0 {
			stats.UnusedComponents = append(stats.UnusedComponents, name)
		}
	}
	
	// Sort by usage count for most used
	sort.Slice(usageStats, func(i, j int) bool {
		return usageStats[i].UsageCount > usageStats[j].UsageCount
	})
	
	// Get top 10 most used
	if len(usageStats) > 10 {
		stats.MostUsed = usageStats[:10]
	} else {
		stats.MostUsed = usageStats
	}
	
	// Sort by last used for recently used
	sort.Slice(usageStats, func(i, j int) bool {
		return usageStats[i].LastUsed.After(usageStats[j].LastUsed)
	})
	
	// Get top 10 recently used
	if len(usageStats) > 10 {
		stats.RecentlyUsed = usageStats[:10]
	} else {
		stats.RecentlyUsed = usageStats
	}
	
	return stats
}

// GetCategories returns all component categories
func (r *Registry) GetCategories() []ComponentCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var categories []ComponentCategory
	
	for name, components := range r.categories {
		category := ComponentCategory{
			Name:        name,
			Description: r.getCategoryDescription(name),
			Components:  components,
		}
		categories = append(categories, category)
	}
	
	// Sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})
	
	return categories
}

// GetComponentsByCategory returns components in a specific category
func (r *Registry) GetComponentsByCategory(category string) []*ComponentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if componentNames, exists := r.categories[category]; exists {
		result := make([]*ComponentInfo, 0, len(componentNames))
		for _, name := range componentNames {
			if info, exists := r.components[name]; exists {
				result = append(result, info)
			}
		}
		return result
	}
	
	return []*ComponentInfo{}
}

// GetDependencies returns the dependency graph for components
func (r *Registry) GetDependencies() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string][]string)
	for name, deps := range r.dependencies {
		result[name] = make([]string, len(deps))
		copy(result[name], deps)
	}
	
	return result
}

// GetOptimizationHints returns suggestions for optimizing component usage
func (r *Registry) GetOptimizationHints() []OptimizationHint {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var hints []OptimizationHint
	
	// Check for unused components
	unusedCount := 0
	for _, info := range r.components {
		if info.UsageCount == 0 {
			unusedCount++
		}
	}
	
	if unusedCount > 0 {
		hints = append(hints, OptimizationHint{
			Type:       "unused",
			Component:  "multiple",
			Message:    fmt.Sprintf("Consider removing %d unused components to reduce bundle size", unusedCount),
			Severity:   "warning",
			Suggestion: "Review and remove unused components",
			CreatedAt:  time.Now(),
		})
	}
	
	// Check for heavily used components that might benefit from caching
	for name, info := range r.components {
		if info.UsageCount > 100 {
			hints = append(hints, OptimizationHint{
				Type:       "performance",
				Component:  name,
				Message:    fmt.Sprintf("Component '%s' is heavily used (%d times) - consider caching", name, info.UsageCount),
				Severity:   "info",
				Suggestion: "Implement component caching for better performance",
				CreatedAt:  time.Now(),
			})
		}
	}
	
	// Check for components with many dependencies
	for name, deps := range r.dependencies {
		if len(deps) > 5 {
			hints = append(hints, OptimizationHint{
				Type:       "complexity",
				Component:  name,
				Message:    fmt.Sprintf("Component '%s' has many dependencies (%d) - consider refactoring", name, len(deps)),
				Severity:   "warning",
				Suggestion: "Refactor to reduce dependencies",
				CreatedAt:  time.Now(),
			})
		}
	}
	
	return hints
}

// Search finds components matching the given criteria
func (r *Registry) Search(query string) []*ComponentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var results []*ComponentInfo
	query = strings.ToLower(query)
	
	for _, info := range r.components {
		// Search in name
		if strings.Contains(strings.ToLower(info.Name), query) {
			results = append(results, info)
			continue
		}
		
		// Search in type
		if strings.Contains(strings.ToLower(info.Type), query) {
			results = append(results, info)
			continue
		}
		
		// Search in system
		if strings.Contains(strings.ToLower(info.System), query) {
			results = append(results, info)
			continue
		}
		
		// Search in properties
		for _, prop := range info.Props {
			if strings.Contains(strings.ToLower(prop.Name), query) ||
				strings.Contains(strings.ToLower(prop.Description), query) {
				results = append(results, info)
				break
			}
		}
	}
	
	// Sort results by usage count (most used first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].UsageCount > results[j].UsageCount
	})
	
	return results
}

// Save persists the registry to disk
func (r *Registry) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveUnsafe()
}

// Load restores the registry from disk
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.persistPath == "" {
		return nil // No persistence configured
	}
	
	if _, err := os.Stat(r.persistPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to load
	}
	
	data, err := os.ReadFile(r.persistPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}
	
	var persistData struct {
		Components   map[string]*ComponentInfo `json:"components"`
		Usage        map[string]int            `json:"usage"`
		Dependencies map[string][]string       `json:"dependencies"`
		Categories   map[string][]string       `json:"categories"`
	}
	
	if err := json.Unmarshal(data, &persistData); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}
	
	r.components = persistData.Components
	r.usage = persistData.Usage
	r.dependencies = persistData.Dependencies
	r.categories = persistData.Categories
	
	return nil
}

// Clear removes all components from the registry
func (r *Registry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.components = make(map[string]*ComponentInfo)
	r.usage = make(map[string]int)
	r.factories = make(map[string]ComponentFactory)
	r.dependencies = make(map[string][]string)
	r.categories = make(map[string][]string)
	
	if r.autoSave {
		return r.saveUnsafe()
	}
	
	return nil
}

// Helper methods

// saveUnsafe saves the registry without locking (assumes caller has lock)
func (r *Registry) saveUnsafe() error {
	if r.persistPath == "" {
		return nil // No persistence configured
	}
	
	// Ensure directory exists
	dir := filepath.Dir(r.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	persistData := struct {
		Components   map[string]*ComponentInfo `json:"components"`
		Usage        map[string]int            `json:"usage"`
		Dependencies map[string][]string       `json:"dependencies"`
		Categories   map[string][]string       `json:"categories"`
		SavedAt      time.Time                 `json:"savedAt"`
	}{
		Components:   r.components,
		Usage:        r.usage,
		Dependencies: r.dependencies,
		Categories:   r.categories,
		SavedAt:      time.Now(),
	}
	
	data, err := json.MarshalIndent(persistData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}
	
	if err := os.WriteFile(r.persistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}
	
	return nil
}

// extractComponentInfo extracts metadata from a component
func (r *Registry) extractComponentInfo(name string, component Component) (*ComponentInfo, error) {
	info := &ComponentInfo{
		Name:         name,
		Type:         r.getComponentType(component),
		System:       "webawesome",
		FilePath:     "",
		Dependencies: r.extractDependencies(component),
		Props:        r.extractProperties(component),
		Examples:     []ComponentExample{},
		UsageCount:   0,
		LastUsed:     time.Now(),
	}
	
	return info, nil
}

// getComponentType determines the type of a component
func (r *Registry) getComponentType(component Component) string {
	switch component.(type) {
	case *ButtonComponent:
		return "button"
	case *InputComponent:
		return "input"
	case *SelectComponent:
		return "select"
	case *TextareaComponent:
		return "textarea"
	case *CheckboxComponent:
		return "checkbox"
	case *RadioComponent:
		return "radio"
	case *CardComponent:
		return "card"
	case *DialogComponent:
		return "dialog"
	case *DrawerComponent:
		return "drawer"
	case *TabGroupComponent:
		return "tab-group"
	case *TableComponent:
		return "table"
	case *BadgeComponent:
		return "badge"
	case *ProgressBarComponent:
		return "progress-bar"
	case *SpinnerComponent:
		return "spinner"
	case *DropdownComponent:
		return "dropdown"
	case *MenuComponent:
		return "menu"
	case *TooltipComponent:
		return "tooltip"
	default:
		return "component"
	}
}

// extractDependencies extracts dependencies from a component
func (r *Registry) extractDependencies(component Component) []string {
	// For now, return basic Web Awesome dependencies
	// In a full implementation, this would analyze the component structure
	return []string{"web-awesome-css", "web-awesome-js"}
}

// extractProperties extracts properties from a component using reflection
func (r *Registry) extractProperties(component Component) []ComponentProp {
	var props []ComponentProp
	
	v := reflect.ValueOf(component)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields and BaseComponent
		if !field.IsExported() || field.Name == "BaseComponent" {
			continue
		}
		
		prop := ComponentProp{
			Name:        field.Name,
			Type:        field.Type.String(),
			Required:    false, // Would need struct tags to determine this
			Description: "", // Would need struct tags or comments
		}
		
		// Extract default value
		fieldValue := v.Field(i)
		if fieldValue.IsValid() && fieldValue.CanInterface() {
			prop.Default = fieldValue.Interface()
		}
		
		props = append(props, prop)
	}
	
	return props
}

// createComponentByType creates a basic component instance by type
func (r *Registry) createComponentByType(componentType string) (Component, error) {
	switch componentType {
	case "button":
		return &ButtonComponent{}, nil
	case "input":
		return &InputComponent{}, nil
	case "select":
		return &SelectComponent{}, nil
	case "textarea":
		return &TextareaComponent{}, nil
	case "checkbox":
		return &CheckboxComponent{}, nil
	case "radio":
		return &RadioComponent{}, nil
	case "card":
		return &CardComponent{}, nil
	case "dialog":
		return &DialogComponent{}, nil
	case "drawer":
		return &DrawerComponent{}, nil
	case "tab-group":
		return &TabGroupComponent{}, nil
	case "table":
		return &TableComponent{}, nil
	case "badge":
		return &BadgeComponent{}, nil
	case "progress-bar":
		return &ProgressBarComponent{}, nil
	case "spinner":
		return &SpinnerComponent{}, nil
	case "dropdown":
		return &DropdownComponent{}, nil
	case "menu":
		return &MenuComponent{}, nil
	case "tooltip":
		return &TooltipComponent{}, nil
	default:
		return nil, &RegistryError{
			Type:    "unknown_component_type",
			Message: "Unknown component type: " + componentType,
		}
	}
}

// registerDefaultComponents registers all default Web Awesome components
func (r *Registry) registerDefaultComponents() {
	// Register component factories for all Web Awesome components
	r.factories["button"] = func() Component { return &ButtonComponent{} }
	r.factories["input"] = func() Component { return &InputComponent{} }
	r.factories["select"] = func() Component { return &SelectComponent{} }
	r.factories["textarea"] = func() Component { return &TextareaComponent{} }
	r.factories["checkbox"] = func() Component { return &CheckboxComponent{} }
	r.factories["radio"] = func() Component { return &RadioComponent{} }
	r.factories["card"] = func() Component { return &CardComponent{} }
	r.factories["dialog"] = func() Component { return &DialogComponent{} }
	r.factories["drawer"] = func() Component { return &DrawerComponent{} }
	r.factories["tab-group"] = func() Component { return &TabGroupComponent{} }
	r.factories["table"] = func() Component { return &TableComponent{} }
	r.factories["badge"] = func() Component { return &BadgeComponent{} }
	r.factories["progress-bar"] = func() Component { return &ProgressBarComponent{} }
	r.factories["spinner"] = func() Component { return &SpinnerComponent{} }
	r.factories["dropdown"] = func() Component { return &DropdownComponent{} }
	r.factories["menu"] = func() Component { return &MenuComponent{} }
	r.factories["tooltip"] = func() Component { return &TooltipComponent{} }
	
	// Create component info for each
	for name, factory := range r.factories {
		component := factory()
		info, _ := r.extractComponentInfo(name, component)
		r.components[name] = info
		r.usage[name] = 0
	}
}

// setupCategories organizes components into categories
func (r *Registry) setupCategories() {
	r.categories["form"] = []string{"button", "input", "select", "textarea", "checkbox", "radio"}
	r.categories["layout"] = []string{"card", "dialog", "drawer", "tab-group", "table"}
	r.categories["feedback"] = []string{"badge", "progress-bar", "spinner"}
	r.categories["interactive"] = []string{"dropdown", "menu", "tooltip"}
	r.categories["all"] = []string{}
	
	// Add all components to "all" category
	for name := range r.components {
		r.categories["all"] = append(r.categories["all"], name)
	}
}

// getCategoryDescription returns a description for a category
func (r *Registry) getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"form":        "Form controls and input components",
		"layout":      "Layout and structural components",
		"feedback":    "Status and feedback components",
		"interactive": "Interactive and overlay components",
		"all":         "All available components",
	}
	
	if desc, exists := descriptions[category]; exists {
		return desc
	}
	
	return "Component category"
}