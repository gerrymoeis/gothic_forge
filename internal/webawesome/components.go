package webawesome

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Render implements the Component interface for BaseComponent
func (bc *BaseComponent) Render(ctx context.Context, w io.Writer) error {
	// BaseComponent is not meant to be rendered directly
	return NewValidationError("component", "BaseComponent", "BaseComponent cannot be rendered directly", "invalid_operation")
}

// Validate implements the Component interface for BaseComponent
func (bc *BaseComponent) Validate() error {
	// Basic validation for common fields
	if bc.ID != "" && !isValidID(bc.ID) {
		return NewValidationError("id", bc.ID, "Invalid ID format", "invalid_format")
	}
	return nil
}

// GetAttributes implements the Component interface for BaseComponent
func (bc *BaseComponent) GetAttributes() map[string]string {
	return ExtractBaseAttributes(*bc)
}

// ButtonComponent implementation

// Render implements the Component interface for ButtonComponent
func (bc *ButtonComponent) Render(ctx context.Context, w io.Writer) error {
	if err := bc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := bc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if bc.Variant != "" {
		attrs["variant"] = bc.Variant
	}
	if bc.Size != "" {
		attrs["size"] = bc.Size
	}
	if bc.Disabled {
		attrs["disabled"] = ""
	}
	if bc.Loading {
		attrs["loading"] = ""
	}
	if bc.Href != "" {
		attrs["href"] = bc.Href
	}
	if bc.Type != "" {
		attrs["type"] = bc.Type
	}

	// Render the component
	tag := "sl-button"
	content := ""
	
	// Render children if any
	if len(bc.Children) > 0 {
		var childrenHTML strings.Builder
		for _, child := range bc.Children {
			if err := child.Render(ctx, &childrenHTML); err != nil {
				return fmt.Errorf("failed to render button child: %w", err)
			}
		}
		content = childrenHTML.String()
	}

	html := BuildWebAwesomeTag(tag, attrs, content, false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for ButtonComponent
func (bc *ButtonComponent) Validate() error {
	if err := bc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate variant
	if bc.Variant != "" {
		validVariants := []string{"default", "primary", "success", "neutral", "warning", "danger", "text"}
		if err := Validator.ValidateEnum("variant", bc.Variant, validVariants); err != nil {
			return err
		}
	}

	// Validate size
	if bc.Size != "" {
		validSizes := []string{"small", "medium", "large"}
		if err := Validator.ValidateEnum("size", bc.Size, validSizes); err != nil {
			return err
		}
	}

	// Validate type
	if bc.Type != "" {
		validTypes := []string{"button", "submit", "reset"}
		if err := Validator.ValidateEnum("type", bc.Type, validTypes); err != nil {
			return err
		}
	}

	return nil
}

// GetAttributes implements the Component interface for ButtonComponent
func (bc *ButtonComponent) GetAttributes() map[string]string {
	return MergeAttributes(bc.BaseComponent.GetAttributes())
}

// InputComponent implementation

// Render implements the Component interface for InputComponent
func (ic *InputComponent) Render(ctx context.Context, w io.Writer) error {
	if err := ic.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := ic.GetAttributes()
	
	// Add Web Awesome specific attributes
	if ic.Type != "" {
		attrs["type"] = ic.Type
	}
	if ic.Name != "" {
		attrs["name"] = ic.Name
	}
	if ic.Value != "" {
		attrs["value"] = ic.Value
	}
	if ic.Placeholder != "" {
		attrs["placeholder"] = ic.Placeholder
	}
	if ic.Required {
		attrs["required"] = ""
	}
	if ic.Disabled {
		attrs["disabled"] = ""
	}
	if ic.Readonly {
		attrs["readonly"] = ""
	}
	if ic.MinLength > 0 {
		attrs["minlength"] = fmt.Sprintf("%d", ic.MinLength)
	}
	if ic.MaxLength > 0 {
		attrs["maxlength"] = fmt.Sprintf("%d", ic.MaxLength)
	}
	if ic.Pattern != "" {
		attrs["pattern"] = ic.Pattern
	}

	// Render the component (self-closing)
	html := BuildWebAwesomeTag("sl-input", attrs, "", true)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for InputComponent
func (ic *InputComponent) Validate() error {
	if err := ic.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate type
	if ic.Type != "" {
		validTypes := []string{"text", "email", "password", "search", "tel", "url", "number", "date", "datetime-local", "month", "time", "week"}
		if err := Validator.ValidateEnum("type", ic.Type, validTypes); err != nil {
			return err
		}
	}

	// Validate name is provided for form inputs
	if ic.Name == "" {
		return NewValidationError("name", ic.Name, "Input name is required", "required")
	}

	// Validate length constraints
	if ic.MinLength > 0 && ic.MaxLength > 0 && ic.MinLength > ic.MaxLength {
		return NewValidationError("minLength", fmt.Sprintf("%d", ic.MinLength), "MinLength cannot be greater than MaxLength", "invalid_range")
	}

	return nil
}

// GetAttributes implements the Component interface for InputComponent
func (ic *InputComponent) GetAttributes() map[string]string {
	return MergeAttributes(ic.BaseComponent.GetAttributes())
}

// SelectComponent implementation

// Render implements the Component interface for SelectComponent
func (sc *SelectComponent) Render(ctx context.Context, w io.Writer) error {
	if err := sc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := sc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if sc.Name != "" {
		attrs["name"] = sc.Name
	}
	if sc.Value != "" {
		attrs["value"] = sc.Value
	}
	if sc.Multiple {
		attrs["multiple"] = ""
	}
	if sc.Disabled {
		attrs["disabled"] = ""
	}
	if sc.Required {
		attrs["required"] = ""
	}
	if sc.Placeholder != "" {
		attrs["placeholder"] = sc.Placeholder
	}

	// Build options content
	var content strings.Builder
	for _, option := range sc.Options {
		optionAttrs := make(map[string]string)
		optionAttrs["value"] = option.Value
		if option.Disabled {
			optionAttrs["disabled"] = ""
		}
		if option.Selected {
			optionAttrs["selected"] = ""
		}
		
		optionHTML := BuildWebAwesomeTag("sl-option", optionAttrs, option.Label, false)
		content.WriteString(optionHTML)
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-select", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for SelectComponent
func (sc *SelectComponent) Validate() error {
	if err := sc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate name is provided
	if sc.Name == "" {
		return NewValidationError("name", sc.Name, "Select name is required", "required")
	}

	// Validate options
	if len(sc.Options) == 0 {
		return NewValidationError("options", "", "Select must have at least one option", "required")
	}

	// Validate option values are unique
	values := make(map[string]bool)
	for i, option := range sc.Options {
		if option.Value == "" {
			return NewValidationError(fmt.Sprintf("options[%d].value", i), option.Value, "Option value cannot be empty", "required")
		}
		if values[option.Value] {
			return NewValidationError(fmt.Sprintf("options[%d].value", i), option.Value, "Duplicate option value", "duplicate")
		}
		values[option.Value] = true
	}

	return nil
}

// GetAttributes implements the Component interface for SelectComponent
func (sc *SelectComponent) GetAttributes() map[string]string {
	return MergeAttributes(sc.BaseComponent.GetAttributes())
}

// TextareaComponent implementation

// Render implements the Component interface for TextareaComponent
func (tc *TextareaComponent) Render(ctx context.Context, w io.Writer) error {
	if err := tc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := tc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if tc.Name != "" {
		attrs["name"] = tc.Name
	}
	if tc.Placeholder != "" {
		attrs["placeholder"] = tc.Placeholder
	}
	if tc.Required {
		attrs["required"] = ""
	}
	if tc.Disabled {
		attrs["disabled"] = ""
	}
	if tc.Readonly {
		attrs["readonly"] = ""
	}
	if tc.Rows > 0 {
		attrs["rows"] = fmt.Sprintf("%d", tc.Rows)
	}
	if tc.Cols > 0 {
		attrs["cols"] = fmt.Sprintf("%d", tc.Cols)
	}
	if tc.MinLength > 0 {
		attrs["minlength"] = fmt.Sprintf("%d", tc.MinLength)
	}
	if tc.MaxLength > 0 {
		attrs["maxlength"] = fmt.Sprintf("%d", tc.MaxLength)
	}

	// Render the component with value as content
	html := BuildWebAwesomeTag("sl-textarea", attrs, tc.Value, false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for TextareaComponent
func (tc *TextareaComponent) Validate() error {
	if err := tc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate name is provided
	if tc.Name == "" {
		return NewValidationError("name", tc.Name, "Textarea name is required", "required")
	}

	// Validate dimensions
	if tc.Rows > 0 && tc.Rows < 1 {
		return NewValidationError("rows", fmt.Sprintf("%d", tc.Rows), "Rows must be at least 1", "invalid_range")
	}
	if tc.Cols > 0 && tc.Cols < 1 {
		return NewValidationError("cols", fmt.Sprintf("%d", tc.Cols), "Cols must be at least 1", "invalid_range")
	}

	// Validate length constraints
	if tc.MinLength > 0 && tc.MaxLength > 0 && tc.MinLength > tc.MaxLength {
		return NewValidationError("minLength", fmt.Sprintf("%d", tc.MinLength), "MinLength cannot be greater than MaxLength", "invalid_range")
	}

	return nil
}

// GetAttributes implements the Component interface for TextareaComponent
func (tc *TextareaComponent) GetAttributes() map[string]string {
	return MergeAttributes(tc.BaseComponent.GetAttributes())
}

// CheckboxComponent implementation

// Render implements the Component interface for CheckboxComponent
func (cc *CheckboxComponent) Render(ctx context.Context, w io.Writer) error {
	if err := cc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := cc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if cc.Name != "" {
		attrs["name"] = cc.Name
	}
	if cc.Value != "" {
		attrs["value"] = cc.Value
	}
	if cc.Checked {
		attrs["checked"] = ""
	}
	if cc.Disabled {
		attrs["disabled"] = ""
	}
	if cc.Required {
		attrs["required"] = ""
	}

	// Render the component with label as content
	html := BuildWebAwesomeTag("sl-checkbox", attrs, cc.Label, false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for CheckboxComponent
func (cc *CheckboxComponent) Validate() error {
	if err := cc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate name is provided
	if cc.Name == "" {
		return NewValidationError("name", cc.Name, "Checkbox name is required", "required")
	}

	return nil
}

// GetAttributes implements the Component interface for CheckboxComponent
func (cc *CheckboxComponent) GetAttributes() map[string]string {
	return MergeAttributes(cc.BaseComponent.GetAttributes())
}

// RadioComponent implementation

// Render implements the Component interface for RadioComponent
func (rc *RadioComponent) Render(ctx context.Context, w io.Writer) error {
	if err := rc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := rc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if rc.Name != "" {
		attrs["name"] = rc.Name
	}
	if rc.Value != "" {
		attrs["value"] = rc.Value
	}
	if rc.Checked {
		attrs["checked"] = ""
	}
	if rc.Disabled {
		attrs["disabled"] = ""
	}
	if rc.Required {
		attrs["required"] = ""
	}

	// Render the component with label as content
	html := BuildWebAwesomeTag("sl-radio", attrs, rc.Label, false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for RadioComponent
func (rc *RadioComponent) Validate() error {
	if err := rc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate name is provided
	if rc.Name == "" {
		return NewValidationError("name", rc.Name, "Radio name is required", "required")
	}

	// Validate value is provided
	if rc.Value == "" {
		return NewValidationError("value", rc.Value, "Radio value is required", "required")
	}

	return nil
}

// GetAttributes implements the Component interface for RadioComponent
func (rc *RadioComponent) GetAttributes() map[string]string {
	return MergeAttributes(rc.BaseComponent.GetAttributes())
}

// Helper functions

// isValidID checks if an ID is valid (basic HTML ID validation)
func isValidID(id string) bool {
	if id == "" {
		return false
	}
	
	// Basic validation: starts with letter, contains only letters, numbers, hyphens, underscores
	for i, r := range id {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
				return false
			}
		}
	}
	
	return true
}
// Constructor functions for Web Awesome components

// NewButtonComponent creates a new ButtonComponent
func NewButtonComponent() *ButtonComponent {
	return &ButtonComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewInputComponent creates a new InputComponent
func NewInputComponent() *InputComponent {
	return &InputComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewSelectComponent creates a new SelectComponent
func NewSelectComponent() *SelectComponent {
	return &SelectComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewTextareaComponent creates a new TextareaComponent
func NewTextareaComponent() *TextareaComponent {
	return &TextareaComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewCheckboxComponent creates a new CheckboxComponent
func NewCheckboxComponent() *CheckboxComponent {
	return &CheckboxComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewRadioComponent creates a new RadioComponent
func NewRadioComponent() *RadioComponent {
	return &RadioComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// Helper methods for components

// SetID sets the ID attribute
func (bc *BaseComponent) SetID(id string) {
	bc.ID = id
}

// SetClass sets the class attribute
func (bc *BaseComponent) SetClass(class string) {
	bc.Class = class
}

// SetAttribute sets a data attribute
func (bc *BaseComponent) SetAttribute(name, value string) {
	if bc.DataAttrs == nil {
		bc.DataAttrs = make(map[string]string)
	}
	bc.DataAttrs[name] = value
}

// SetContent sets the content for components that support it
func (bc *ButtonComponent) SetContent(content string) {
	// This would be implemented based on how content is handled
	// For now, we'll store it as a data attribute
	bc.SetAttribute("content", content)
}

// SetContent sets the content for textarea
func (tc *TextareaComponent) SetContent(content string) {
	tc.Value = content
}

// RenderHTML renders the component to an HTML string
func (bc *ButtonComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := bc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (ic *InputComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := ic.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (tc *TextareaComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := tc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (cc *CheckboxComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := cc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (rc *RadioComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := rc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (sc *SelectComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := sc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Additional helper methods for form components

// SetName sets the name attribute for form components
func (ic *InputComponent) SetName(name string) {
	ic.Name = name
}

// SetValue sets the value for form components
func (ic *InputComponent) SetValue(value string) {
	ic.Value = value
}

// SetPlaceholder sets the placeholder for form components
func (ic *InputComponent) SetPlaceholder(placeholder string) {
	ic.Placeholder = placeholder
}

// SetRequired sets the required state for form components
func (ic *InputComponent) SetRequired(required bool) {
	ic.Required = required
}

// SetDisabled sets the disabled state for form components
func (ic *InputComponent) SetDisabled(disabled bool) {
	ic.Disabled = disabled
}

// SetType sets the input type
func (ic *InputComponent) SetType(inputType string) {
	ic.Type = inputType
}

// AddOption adds an option to SelectComponent
func (sc *SelectComponent) AddOption(value, label string, selected bool) {
	option := SelectOption{
		Value:    value,
		Label:    label,
		Selected: selected,
	}
	sc.Options = append(sc.Options, option)
}

// SetMultiple sets the multiple selection state for SelectComponent
func (sc *SelectComponent) SetMultiple(multiple bool) {
	sc.Multiple = multiple
}

// SetName sets the name attribute for SelectComponent
func (sc *SelectComponent) SetName(name string) {
	sc.Name = name
}

// SetValue sets the selected value for SelectComponent
func (sc *SelectComponent) SetValue(value string) {
	sc.Value = value
}

// SetPlaceholder sets the placeholder for SelectComponent
func (sc *SelectComponent) SetPlaceholder(placeholder string) {
	sc.Placeholder = placeholder
}

// SetName sets the name attribute for TextareaComponent
func (tc *TextareaComponent) SetName(name string) {
	tc.Name = name
}

// SetRows sets the rows for TextareaComponent
func (tc *TextareaComponent) SetRows(rows int) {
	tc.Rows = rows
}

// SetCols sets the cols for TextareaComponent
func (tc *TextareaComponent) SetCols(cols int) {
	tc.Cols = cols
}

// SetPlaceholder sets the placeholder for TextareaComponent
func (tc *TextareaComponent) SetPlaceholder(placeholder string) {
	tc.Placeholder = placeholder
}

// SetName sets the name attribute for CheckboxComponent
func (cc *CheckboxComponent) SetName(name string) {
	cc.Name = name
}

// SetValue sets the value for CheckboxComponent
func (cc *CheckboxComponent) SetValue(value string) {
	cc.Value = value
}

// SetChecked sets the checked state for CheckboxComponent
func (cc *CheckboxComponent) SetChecked(checked bool) {
	cc.Checked = checked
}

// SetLabel sets the label for CheckboxComponent
func (cc *CheckboxComponent) SetLabel(label string) {
	cc.Label = label
}

// SetRequired sets the required state for CheckboxComponent
func (cc *CheckboxComponent) SetRequired(required bool) {
	cc.Required = required
}

// SetName sets the name attribute for RadioComponent
func (rc *RadioComponent) SetName(name string) {
	rc.Name = name
}

// SetValue sets the value for RadioComponent
func (rc *RadioComponent) SetValue(value string) {
	rc.Value = value
}

// SetChecked sets the checked state for RadioComponent
func (rc *RadioComponent) SetChecked(checked bool) {
	rc.Checked = checked
}

// SetLabel sets the label for RadioComponent
func (rc *RadioComponent) SetLabel(label string) {
	rc.Label = label
}