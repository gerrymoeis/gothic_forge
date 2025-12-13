package webawesome

import (
	"fmt"
	"strings"
)

// BuildError represents errors that occur during the build process
type BuildError struct {
	Type        string   `json:"type"`
	Component   string   `json:"component"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
}

func (e *BuildError) Error() string {
	var parts []string
	
	if e.File != "" {
		if e.Line > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", e.File, e.Line))
		} else {
			parts = append(parts, e.File)
		}
	}
	
	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("component '%s'", e.Component))
	}
	
	parts = append(parts, e.Message)
	
	return strings.Join(parts, ": ")
}

// RuntimeError represents errors that occur during runtime
type RuntimeError struct {
	Type      string `json:"type"`
	Component string `json:"component"`
	Context   string `json:"context"`
	Message   string `json:"message"`
	Fallback  string `json:"fallback"`
}

func (e *RuntimeError) Error() string {
	var parts []string
	
	if e.Context != "" {
		parts = append(parts, e.Context)
	}
	
	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("component '%s'", e.Component))
	}
	
	parts = append(parts, e.Message)
	
	return strings.Join(parts, ": ")
}

// MigrationError represents errors that occur during migration
type MigrationError struct {
	Type        string `json:"type"`
	File        string `json:"file"`
	Component   string `json:"component"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
}

func (e *MigrationError) Error() string {
	var parts []string
	
	if e.File != "" {
		parts = append(parts, e.File)
	}
	
	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("component '%s'", e.Component))
	}
	
	parts = append(parts, e.Message)
	
	return strings.Join(parts, ": ")
}

// ValidationError represents validation errors for components or themes
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// DefaultErrorHandler provides a basic implementation of ErrorHandler
type DefaultErrorHandler struct{}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{}
}

// HandleBuildError processes build errors and provides suggestions
func (h *DefaultErrorHandler) HandleBuildError(err BuildError) error {
	// Add contextual suggestions based on error type
	switch err.Type {
	case "invalid_property":
		err.Suggestions = append(err.Suggestions, 
			"Check the Web Awesome documentation for valid property values",
			"Ensure the property name is spelled correctly",
			"Verify the component supports this property",
		)
	case "missing_dependency":
		err.Suggestions = append(err.Suggestions,
			"Ensure Web Awesome CSS and JS are included in your layout template",
			"Check CDN connectivity",
			"Verify the Web Awesome version is compatible",
		)
	case "template_compilation":
		err.Suggestions = append(err.Suggestions,
			"Check for syntax errors in the template",
			"Ensure all component properties are properly typed",
			"Verify imports are correct",
		)
	}
	
	return &err
}

// HandleRuntimeError processes runtime errors and provides fallbacks
func (h *DefaultErrorHandler) HandleRuntimeError(err RuntimeError) error {
	// Set fallback strategies based on error type
	switch err.Type {
	case "cdn_unavailable":
		err.Fallback = "Use local Web Awesome assets or fallback to DaisyUI"
	case "component_initialization":
		err.Fallback = "Render component without JavaScript enhancement"
	case "theme_loading":
		err.Fallback = "Use default theme"
	}
	
	return &err
}

// HandleMigrationError processes migration errors and provides guidance
func (h *DefaultErrorHandler) HandleMigrationError(err MigrationError) error {
	// Add suggestions based on migration error type
	switch err.Type {
	case "incompatible_component":
		err.Suggestion = "This component requires manual migration. Check the migration guide for alternatives."
	case "conflicting_styles":
		err.Suggestion = "Remove conflicting DaisyUI classes before applying Web Awesome components."
	case "breaking_change":
		err.Suggestion = "This change may affect functionality. Review and test thoroughly."
	}
	
	return &err
}

// GetSuggestions provides general suggestions for common errors
func (h *DefaultErrorHandler) GetSuggestions(err error) []string {
	switch e := err.(type) {
	case *BuildError:
		return e.Suggestions
	case *ValidationError:
		return []string{
			"Check the component documentation for valid values",
			"Ensure required properties are provided",
			"Verify property types match expectations",
		}
	case *RegistryError:
		return []string{
			"Check if the component is properly registered",
			"Verify the component name is spelled correctly",
			"Ensure the component system (webawesome/daisyui) is specified",
		}
	default:
		return []string{
			"Check the Gothic Forge documentation",
			"Verify your Web Awesome integration setup",
			"Consider filing an issue if the problem persists",
		}
	}
}

// Common error constructors for convenience

// NewBuildError creates a new build error
func NewBuildError(errorType, component, file string, line int, message string) *BuildError {
	return &BuildError{
		Type:        errorType,
		Component:   component,
		File:        file,
		Line:        line,
		Message:     message,
		Suggestions: []string{},
	}
}

// NewRuntimeError creates a new runtime error
func NewRuntimeError(errorType, component, context, message string) *RuntimeError {
	return &RuntimeError{
		Type:      errorType,
		Component: component,
		Context:   context,
		Message:   message,
		Fallback:  "",
	}
}

// NewMigrationError creates a new migration error
func NewMigrationError(errorType, file, component, message string) *MigrationError {
	return &MigrationError{
		Type:       errorType,
		File:       file,
		Component:  component,
		Message:    message,
		Suggestion: "",
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
	}
}