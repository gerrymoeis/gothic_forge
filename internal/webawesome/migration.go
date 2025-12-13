package webawesome

import (
	"time"
)

// MigrationReport contains the results of scanning templates for DaisyUI usage
type MigrationReport struct {
	ProjectPath    string                    `json:"projectPath"`
	ScanDate       time.Time                `json:"scanDate"`
	TotalFiles     int                      `json:"totalFiles"`
	FilesWithDaisy int                      `json:"filesWithDaisy"`
	Components     []ComponentUsage         `json:"components"`
	Conflicts      []MigrationConflict      `json:"conflicts"`
	Suggestions    []MigrationSuggestion    `json:"suggestions"`
}

// ComponentUsage tracks how DaisyUI components are used in the project
type ComponentUsage struct {
	Component string   `json:"component"`
	Files     []string `json:"files"`
	Count     int      `json:"count"`
	Mapping   string   `json:"mapping"` // Web Awesome equivalent
}

// MigrationConflict represents issues that need manual resolution
type MigrationConflict struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Component   string `json:"component"`
	Issue       string `json:"issue"`
	Suggestion  string `json:"suggestion"`
}

// MigrationSuggestion provides recommendations for migration
type MigrationSuggestion struct {
	Type        string `json:"type"` // replace, enhance, manual
	Component   string `json:"component"`
	Replacement string `json:"replacement"`
	Reason      string `json:"reason"`
	Effort      string `json:"effort"` // low, medium, high
}

// Migration represents a complete migration plan
type Migration struct {
	Report      *MigrationReport      `json:"report"`
	Actions     []MigrationAction     `json:"actions"`
	BackupPath  string               `json:"backupPath"`
	CreatedAt   time.Time            `json:"createdAt"`
}

// MigrationAction represents a single migration step
type MigrationAction struct {
	Type        string            `json:"type"` // replace, add, remove, modify
	File        string            `json:"file"`
	Line        int               `json:"line,omitempty"`
	OldContent  string            `json:"oldContent,omitempty"`
	NewContent  string            `json:"newContent,omitempty"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DefaultMigrationTool provides a default implementation for DaisyUI to Web Awesome migration
type DefaultMigrationTool struct {
	componentMappings map[string]string
}

// NewMigrationTool creates a new migration tool with default component mappings
func NewMigrationTool() *DefaultMigrationTool {
	return &DefaultMigrationTool{
		componentMappings: getDefaultComponentMappings(),
	}
}

// ScanTemplates scans the given directory for DaisyUI component usage
func (mt *DefaultMigrationTool) ScanTemplates(dir string) (*MigrationReport, error) {
	// This is a placeholder implementation
	// In a full implementation, this would:
	// 1. Walk through all template files in the directory
	// 2. Parse templates to find DaisyUI class usage
	// 3. Identify components and their usage patterns
	// 4. Generate conflicts and suggestions
	
	report := &MigrationReport{
		ProjectPath:    dir,
		ScanDate:       time.Now(),
		TotalFiles:     0,
		FilesWithDaisy: 0,
		Components:     []ComponentUsage{},
		Conflicts:      []MigrationConflict{},
		Suggestions:    []MigrationSuggestion{},
	}
	
	return report, nil
}

// GenerateMigration creates a migration plan based on the scan report
func (mt *DefaultMigrationTool) GenerateMigration(report *MigrationReport) (*Migration, error) {
	migration := &Migration{
		Report:     report,
		Actions:    []MigrationAction{},
		BackupPath: "",
		CreatedAt:  time.Now(),
	}
	
	// Generate migration actions based on the report
	for _, usage := range report.Components {
		if mapping, exists := mt.componentMappings[usage.Component]; exists {
			for _, file := range usage.Files {
				action := MigrationAction{
					Type:        "replace",
					File:        file,
					OldContent:  usage.Component,
					NewContent:  mapping,
					Description: "Replace DaisyUI " + usage.Component + " with Web Awesome " + mapping,
					Metadata: map[string]string{
						"component_type": "ui",
						"effort":         "low",
					},
				}
				migration.Actions = append(migration.Actions, action)
			}
		}
	}
	
	return migration, nil
}

// ApplyMigration executes the migration plan
func (mt *DefaultMigrationTool) ApplyMigration(migration *Migration) error {
	// This is a placeholder implementation
	// In a full implementation, this would:
	// 1. Create backups of all files to be modified
	// 2. Apply each migration action
	// 3. Validate the results
	// 4. Provide rollback capability if needed
	
	return nil
}

// getDefaultComponentMappings returns the default mapping from DaisyUI to Web Awesome components
func getDefaultComponentMappings() map[string]string {
	return map[string]string{
		// Buttons
		"btn":         "sl-button",
		"btn-primary": "sl-button[variant='primary']",
		"btn-secondary": "sl-button[variant='secondary']",
		"btn-success": "sl-button[variant='success']",
		"btn-warning": "sl-button[variant='warning']",
		"btn-error":   "sl-button[variant='danger']",
		"btn-sm":      "sl-button[size='small']",
		"btn-lg":      "sl-button[size='large']",
		
		// Form Controls
		"input":       "sl-input",
		"select":      "sl-select",
		"textarea":    "sl-textarea",
		"checkbox":    "sl-checkbox",
		"radio":       "sl-radio",
		
		// Layout
		"card":        "sl-card",
		"modal":       "sl-dialog",
		"drawer":      "sl-drawer",
		"tabs":        "sl-tab-group",
		
		// Data Display
		"table":       "sl-table", // Custom implementation
		"badge":       "sl-badge",
		"progress":    "sl-progress-bar",
		"loading":     "sl-spinner",
		
		// Interactive
		"dropdown":    "sl-dropdown",
		"menu":        "sl-menu",
		"tooltip":     "sl-tooltip",
	}
}