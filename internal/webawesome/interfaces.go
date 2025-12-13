package webawesome

import (
	"context"
	"io"
)

// Component represents a Web Awesome component wrapper
type Component interface {
	Render(ctx context.Context, w io.Writer) error
	Validate() error
	GetAttributes() map[string]string
}

// ThemeProvider manages Web Awesome themes
type ThemeProvider interface {
	LoadTheme(name string) (*Theme, error)
	GenerateCSS(theme *Theme) (string, error)
	ValidateTheme(theme *Theme) error
}

// ComponentRegistry tracks available components
type ComponentRegistry interface {
	Register(name string, component Component) error
	Get(name string) (Component, error)
	List() []string
	GetUsage() map[string]int
	GetStats() *RegistryStats
	GetComponentInfo(name string) (*ComponentInfo, error)
	Search(query string) []*ComponentInfo
	GetCategories() []ComponentCategory
	GetOptimizationHints() []OptimizationHint
	GetDependencies() map[string][]string
	GetComponentsByCategory(category string) []*ComponentInfo
	Clear() error
}

// MigrationTool handles DaisyUI to Web Awesome migration
type MigrationTool interface {
	ScanTemplates(dir string) (*MigrationReport, error)
	GenerateMigration(report *MigrationReport) (*Migration, error)
	ApplyMigration(migration *Migration) error
}

// ErrorHandler provides centralized error handling for Web Awesome integration
type ErrorHandler interface {
	HandleBuildError(err BuildError) error
	HandleRuntimeError(err RuntimeError) error
	HandleMigrationError(err MigrationError) error
	GetSuggestions(err error) []string
}