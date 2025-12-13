package webawesome

import (
	"fmt"
	"html"
	"strings"
)

// AttributeBuilder helps build HTML attributes for components
type AttributeBuilder struct {
	attrs map[string]string
}

// NewAttributeBuilder creates a new attribute builder
func NewAttributeBuilder() *AttributeBuilder {
	return &AttributeBuilder{
		attrs: make(map[string]string),
	}
}

// Add adds an attribute with a value
func (ab *AttributeBuilder) Add(name, value string) *AttributeBuilder {
	if value != "" {
		ab.attrs[name] = value
	}
	return ab
}

// AddBool adds a boolean attribute
func (ab *AttributeBuilder) AddBool(name string, value bool) *AttributeBuilder {
	if value {
		ab.attrs[name] = ""
	}
	return ab
}

// AddIf conditionally adds an attribute
func (ab *AttributeBuilder) AddIf(condition bool, name, value string) *AttributeBuilder {
	if condition && value != "" {
		ab.attrs[name] = value
	}
	return ab
}

// AddBoolIf conditionally adds a boolean attribute
func (ab *AttributeBuilder) AddBoolIf(condition bool, name string) *AttributeBuilder {
	if condition {
		ab.attrs[name] = ""
	}
	return ab
}

// AddClass adds CSS classes
func (ab *AttributeBuilder) AddClass(classes ...string) *AttributeBuilder {
	var validClasses []string
	for _, class := range classes {
		if class != "" {
			validClasses = append(validClasses, class)
		}
	}
	
	if len(validClasses) > 0 {
		existing := ab.attrs["class"]
		if existing != "" {
			ab.attrs["class"] = existing + " " + strings.Join(validClasses, " ")
		} else {
			ab.attrs["class"] = strings.Join(validClasses, " ")
		}
	}
	
	return ab
}

// AddData adds data attributes
func (ab *AttributeBuilder) AddData(name, value string) *AttributeBuilder {
	if value != "" {
		ab.attrs["data-"+name] = value
	}
	return ab
}

// AddAria adds ARIA attributes
func (ab *AttributeBuilder) AddAria(name, value string) *AttributeBuilder {
	if value != "" {
		ab.attrs["aria-"+name] = value
	}
	return ab
}

// Build returns the final attribute string
func (ab *AttributeBuilder) Build() string {
	if len(ab.attrs) == 0 {
		return ""
	}
	
	var parts []string
	for name, value := range ab.attrs {
		if value == "" {
			parts = append(parts, name)
		} else {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, name, html.EscapeString(value)))
		}
	}
	
	return strings.Join(parts, " ")
}

// GetAttributes returns the attributes map
func (ab *AttributeBuilder) GetAttributes() map[string]string {
	result := make(map[string]string)
	for k, v := range ab.attrs {
		result[k] = v
	}
	return result
}

// ComponentValidator provides validation utilities for components
type ComponentValidator struct{}

// NewComponentValidator creates a new component validator
func NewComponentValidator() *ComponentValidator {
	return &ComponentValidator{}
}

// ValidateRequired checks if required fields are present
func (cv *ComponentValidator) ValidateRequired(fields map[string]interface{}) error {
	for name, value := range fields {
		if value == nil || value == "" {
			return NewValidationError(name, fmt.Sprintf("%v", value), "Field is required", "required")
		}
	}
	return nil
}

// ValidateEnum checks if a value is in the allowed list
func (cv *ComponentValidator) ValidateEnum(field, value string, allowed []string) error {
	if value == "" {
		return nil // Empty values are handled by required validation
	}
	
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}
	
	return NewValidationError(field, value, 
		fmt.Sprintf("Value must be one of: %s", strings.Join(allowed, ", ")), 
		"invalid_enum")
}

// ValidateRange checks if a numeric value is within range
func (cv *ComponentValidator) ValidateRange(field string, value, min, max int) error {
	if value < min || value > max {
		return NewValidationError(field, fmt.Sprintf("%d", value),
			fmt.Sprintf("Value must be between %d and %d", min, max),
			"out_of_range")
	}
	return nil
}

// ValidateLength checks string length
func (cv *ComponentValidator) ValidateLength(field, value string, min, max int) error {
	length := len(value)
	if length < min || length > max {
		return NewValidationError(field, value,
			fmt.Sprintf("Length must be between %d and %d characters", min, max),
			"invalid_length")
	}
	return nil
}

// StringUtils provides string manipulation utilities
type StringUtils struct{}

// ToCamelCase converts kebab-case to camelCase
func (su *StringUtils) ToCamelCase(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) <= 1 {
		return s
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	
	return result
}

// ToKebabCase converts camelCase to kebab-case
func (su *StringUtils) ToKebabCase(s string) string {
	var result strings.Builder
	
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	
	return strings.ToLower(result.String())
}

// Sanitize removes potentially dangerous characters
func (su *StringUtils) Sanitize(s string) string {
	// Basic sanitization - remove script tags and dangerous characters
	s = strings.ReplaceAll(s, "<script", "&lt;script")
	s = strings.ReplaceAll(s, "</script>", "&lt;/script&gt;")
	s = strings.ReplaceAll(s, "javascript:", "")
	s = strings.ReplaceAll(s, "vbscript:", "")
	s = strings.ReplaceAll(s, "onload=", "")
	s = strings.ReplaceAll(s, "onerror=", "")
	
	return s
}

// IsEmpty checks if a string is empty or contains only whitespace
func (su *StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Truncate truncates a string to the specified length
func (su *StringUtils) Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	
	if length <= 3 {
		return s[:length]
	}
	
	return s[:length-3] + "..."
}

// Global utility instances
var (
	StringUtil = &StringUtils{}
	Validator  = NewComponentValidator()
)

// Helper functions for common operations

// BuildWebAwesomeTag builds a complete Web Awesome component tag
func BuildWebAwesomeTag(tagName string, attrs map[string]string, content string, selfClosing bool) string {
	builder := NewAttributeBuilder()
	
	for name, value := range attrs {
		builder.Add(name, value)
	}
	
	attrString := builder.Build()
	if attrString != "" {
		attrString = " " + attrString
	}
	
	if selfClosing {
		return fmt.Sprintf("<%s%s />", tagName, attrString)
	}
	
	return fmt.Sprintf("<%s%s>%s</%s>", tagName, attrString, content, tagName)
}

// MergeAttributes merges multiple attribute maps
func MergeAttributes(attrs ...map[string]string) map[string]string {
	result := make(map[string]string)
	
	for _, attrMap := range attrs {
		for key, value := range attrMap {
			if key == "class" && result[key] != "" {
				// Merge CSS classes
				result[key] = result[key] + " " + value
			} else {
				result[key] = value
			}
		}
	}
	
	return result
}

// ExtractBaseAttributes extracts common attributes from BaseComponent
func ExtractBaseAttributes(base BaseComponent) map[string]string {
	attrs := make(map[string]string)
	
	if base.ID != "" {
		attrs["id"] = base.ID
	}
	if base.Class != "" {
		attrs["class"] = base.Class
	}
	if base.Style != "" {
		attrs["style"] = base.Style
	}
	
	// Add data attributes
	for key, value := range base.DataAttrs {
		attrs["data-"+key] = value
	}
	
	// Add event attributes
	for key, value := range base.Events {
		attrs[key] = value
	}
	
	return attrs
}