package webawesome

import (
	"context"
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// RadioOption represents an option for radio button groups
type RadioOption struct {
	Value string
	Label string
}

// CheckboxOption represents an option for checkbox groups
type CheckboxOption struct {
	Value   string
	Label   string
	Checked bool
}

// renderComponent renders a Web Awesome component to HTML string
func renderComponent(component Component) string {
	var buf strings.Builder
	ctx := context.Background()
	
	if err := component.Render(ctx, &buf); err != nil {
		// In case of error, return an error message wrapped in a span
		return `<span class="error">Component render error: ` + err.Error() + `</span>`
	}
	
	return buf.String()
}



// stringOptionsToSelectOptions converts string slice to SelectOption slice
func stringOptionsToSelectOptions(options []string) []SelectOption {
	selectOptions := make([]SelectOption, len(options))
	for i, option := range options {
		selectOptions[i] = SelectOption{
			Value: option,
			Label: option,
		}
	}
	return selectOptions
}

// CreateButtonWithChildren creates a button component with child components
func CreateButtonWithChildren(variant, size, buttonType string, children []templ.Component, attrs ...map[string]string) ButtonComponent {
	return ButtonComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Variant:       variant,
		Size:          size,
		Type:          buttonType,
		Children:      children,
	}
}

// CreateInputWithValidation creates an input component with validation attributes
func CreateInputWithValidation(inputType, name, placeholder string, required bool, minLength, maxLength int, pattern string, attrs ...map[string]string) InputComponent {
	return InputComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Type:          inputType,
		Name:          name,
		Placeholder:   placeholder,
		Required:      required,
		MinLength:     minLength,
		MaxLength:     maxLength,
		Pattern:       pattern,
	}
}

// CreateSelectWithOptions creates a select component with options
func CreateSelectWithOptions(name string, options []SelectOption, multiple bool, attrs ...map[string]string) SelectComponent {
	return SelectComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Name:          name,
		Options:       options,
		Multiple:      multiple,
	}
}

// CreateTextareaWithConstraints creates a textarea with size and length constraints
func CreateTextareaWithConstraints(name, placeholder string, rows, cols, minLength, maxLength int, attrs ...map[string]string) TextareaComponent {
	return TextareaComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Name:          name,
		Placeholder:   placeholder,
		Rows:          rows,
		Cols:          cols,
		MinLength:     minLength,
		MaxLength:     maxLength,
	}
}

// Utility functions for common component patterns

// CreateFormButton creates a button optimized for forms
func CreateFormButton(text, variant, buttonType string, disabled, loading bool, attrs ...map[string]string) ButtonComponent {
	return ButtonComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Variant:       variant,
		Type:          buttonType,
		Disabled:      disabled,
		Loading:       loading,
		Children:      []templ.Component{templ.Raw(text)},
	}
}

// CreateRequiredInput creates an input with required validation
func CreateRequiredInput(inputType, name, placeholder string, attrs ...map[string]string) InputComponent {
	return InputComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Type:          inputType,
		Name:          name,
		Placeholder:   placeholder,
		Required:      true,
	}
}

// CreateEmailInput creates an email input with validation
func CreateEmailInput(name, placeholder string, required bool, attrs ...map[string]string) InputComponent {
	return InputComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Type:          "email",
		Name:          name,
		Placeholder:   placeholder,
		Required:      required,
		Pattern:       `[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`,
	}
}

// CreatePasswordInput creates a password input with strength requirements
func CreatePasswordInput(name, placeholder string, minLength int, attrs ...map[string]string) InputComponent {
	return InputComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Type:          "password",
		Name:          name,
		Placeholder:   placeholder,
		Required:      true,
		MinLength:     minLength,
	}
}

// CreateNumberInput creates a number input with range validation
func CreateNumberInput(name, placeholder string, min, max int, attrs ...map[string]string) InputComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add min/max as data attributes since they're not part of the InputComponent struct
	if base.DataAttrs == nil {
		base.DataAttrs = make(map[string]string)
	}
	if min != 0 {
		base.DataAttrs["min"] = string(rune(min))
	}
	if max != 0 {
		base.DataAttrs["max"] = string(rune(max))
	}
	
	return InputComponent{
		BaseComponent: base,
		Type:          "number",
		Name:          name,
		Placeholder:   placeholder,
	}
}

// CreateSearchInput creates a search input with common attributes
func CreateSearchInput(name, placeholder string, attrs ...map[string]string) InputComponent {
	return InputComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Type:          "search",
		Name:          name,
		Placeholder:   placeholder,
	}
}

// CreateMultiSelect creates a multi-select component
func CreateMultiSelect(name string, options []SelectOption, attrs ...map[string]string) SelectComponent {
	return SelectComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Name:          name,
		Options:       options,
		Multiple:      true,
	}
}

// CreateRadioGroup creates a group of radio buttons from options
func CreateRadioGroup(name string, options []RadioOption, selectedValue string) []RadioComponent {
	radios := make([]RadioComponent, len(options))
	for i, option := range options {
		radios[i] = RadioComponent{
			BaseComponent: BaseComponent{},
			Name:          name,
			Value:         option.Value,
			Label:         option.Label,
			Checked:       option.Value == selectedValue,
		}
	}
	return radios
}

// CreateCheckboxGroup creates a group of checkboxes from options
func CreateCheckboxGroup(namePrefix string, options []CheckboxOption) []CheckboxComponent {
	checkboxes := make([]CheckboxComponent, len(options))
	for i, option := range options {
		checkboxes[i] = CheckboxComponent{
			BaseComponent: BaseComponent{},
			Name:          namePrefix + "[]",
			Value:         option.Value,
			Label:         option.Label,
			Checked:       option.Checked,
		}
	}
	return checkboxes
}

// HTMX Integration Helpers

// CreateHTMXButton creates a button with HTMX attributes
func CreateHTMXButton(text, variant string, hxGet, hxTarget string, attrs ...map[string]string) ButtonComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add HTMX attributes
	if base.Events == nil {
		base.Events = make(map[string]string)
	}
	if hxGet != "" {
		base.Events["hx-get"] = hxGet
	}
	if hxTarget != "" {
		base.Events["hx-target"] = hxTarget
	}
	
	return ButtonComponent{
		BaseComponent: base,
		Variant:       variant,
		Type:          "button",
		Children:      []templ.Component{templ.Raw(text)},
	}
}

// CreateHTMXForm creates a form with HTMX attributes
func CreateHTMXForm(hxPost, hxTarget string, attrs ...map[string]string) map[string]string {
	formAttrs := make(map[string]string)
	
	// Merge provided attributes
	for _, attrMap := range attrs {
		for key, value := range attrMap {
			formAttrs[key] = value
		}
	}
	
	// Add HTMX attributes
	if hxPost != "" {
		formAttrs["hx-post"] = hxPost
	}
	if hxTarget != "" {
		formAttrs["hx-target"] = hxTarget
	}
	
	return formAttrs
}

// Alpine.js Integration Helpers

// CreateAlpineButton creates a button with Alpine.js attributes
func CreateAlpineButton(text, variant string, alpineClick string, attrs ...map[string]string) ButtonComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add Alpine.js attributes
	if base.Events == nil {
		base.Events = make(map[string]string)
	}
	if alpineClick != "" {
		base.Events["@click"] = alpineClick
	}
	
	return ButtonComponent{
		BaseComponent: base,
		Variant:       variant,
		Type:          "button",
		Children:      []templ.Component{templ.Raw(text)},
	}
}

// CreateAlpineInput creates an input with Alpine.js model binding
func CreateAlpineInput(inputType, name, placeholder string, alpineModel string, attrs ...map[string]string) InputComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add Alpine.js attributes
	if base.Events == nil {
		base.Events = make(map[string]string)
	}
	if alpineModel != "" {
		base.Events["x-model"] = alpineModel
	}
	
	return InputComponent{
		BaseComponent: base,
		Type:          inputType,
		Name:          name,
		Placeholder:   placeholder,
	}
}
// createTabsFromMap converts a map of tab labels to content into Tab structs
func createTabsFromMap(tabs map[string]string, activeTab string) []Tab {
	var tabList []Tab
	i := 0
	
	for label, content := range tabs {
		panel := fmt.Sprintf("tab-panel-%d", i)
		tab := Tab{
			Panel:   panel,
			Label:   label,
			Content: content,
			Active:  label == activeTab,
		}
		tabList = append(tabList, tab)
		i++
	}
	
	return tabList
}

// Additional helper functions for component creation

// CreateFormWithValidation creates a form with validation styling
func CreateFormWithValidation(action, method string, hasErrors bool, attrs ...map[string]string) map[string]string {
	formAttrs := make(map[string]string)
	
	// Merge provided attributes
	for _, attrMap := range attrs {
		for key, value := range attrMap {
			formAttrs[key] = value
		}
	}
	
	// Set form attributes
	formAttrs["action"] = action
	formAttrs["method"] = method
	
	// Add validation classes
	if hasErrors {
		if formAttrs["class"] == "" {
			formAttrs["class"] = "form-validation-errors"
		} else {
			formAttrs["class"] += " form-validation-errors"
		}
	}
	
	return formAttrs
}

// CreateResponsiveTable creates a table with responsive classes
func CreateResponsiveTable(headers []string, rows [][]string, caption string, attrs ...map[string]string) TableComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add responsive classes
	if base.Class == "" {
		base.Class = "table-responsive"
	} else {
		base.Class += " table-responsive"
	}
	
	return TableComponent{
		BaseComponent: base,
		Headers:       headers,
		Rows:          rows,
		Caption:       caption,
	}
}

// CreateStatusCard creates a card with status-based styling
func CreateStatusCard(title, content, status string, attrs ...map[string]string) CardComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add status classes
	statusClass := "status-" + status
	if base.Class == "" {
		base.Class = statusClass
	} else {
		base.Class += " " + statusClass
	}
	
	return CardComponent{
		BaseComponent: base,
		Header:        title,
		Children:      []templ.Component{templ.Raw(content)},
	}
}

// CreateLoadingCard creates a card with loading state
func CreateLoadingCard(title string, isLoading bool, content string, attrs ...map[string]string) CardComponent {
	base := mergeBaseAttrs(attrs...)
	
	var children []templ.Component
	if isLoading {
		// Add loading spinner
		spinner := &SpinnerComponent{
			BaseComponent: BaseComponent{Class: "mx-auto my-4"},
			Size:          "medium",
		}
		children = []templ.Component{spinner}
	} else {
		children = []templ.Component{templ.Raw(content)}
	}
	
	return CardComponent{
		BaseComponent: base,
		Header:        title,
		Children:      children,
	}
}

// CreateNotificationBadge creates a notification-style badge
func CreateNotificationBadge(count int, variant string, attrs ...map[string]string) BadgeComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add notification classes
	if base.Class == "" {
		base.Class = "notification-badge"
	} else {
		base.Class += " notification-badge"
	}
	
	content := fmt.Sprintf("%d", count)
	if count > 99 {
		content = "99+"
	}
	
	return BadgeComponent{
		BaseComponent: base,
		Content:       content,
		Variant:       variant,
		Pill:          true,
	}
}

// CreateProgressCard creates a card with progress indicator
func CreateProgressCard(title string, progress int, progressLabel string, attrs ...map[string]string) CardComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Create progress bar component
	progressBar := &ProgressBarComponent{
		BaseComponent: BaseComponent{Class: "mb-2"},
		Value:         progress,
		Label:         progressLabel,
	}
	
	return CardComponent{
		BaseComponent: base,
		Header:        title,
		Children:      []templ.Component{progressBar},
	}
}

// CreateActionCard creates a card with action buttons in footer
func CreateActionCard(title, content string, actions []templ.Component, attrs ...map[string]string) CardComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Build footer with actions
	var footerHTML strings.Builder
	footerHTML.WriteString(`<div class="flex gap-2 justify-end">`)
	for _, action := range actions {
		var actionHTML strings.Builder
		ctx := context.Background()
		if err := action.Render(ctx, &actionHTML); err == nil {
			footerHTML.WriteString(actionHTML.String())
		}
	}
	footerHTML.WriteString(`</div>`)
	
	return CardComponent{
		BaseComponent: base,
		Header:        title,
		Footer:        footerHTML.String(),
		Children:      []templ.Component{templ.Raw(content)},
	}
}

// CreateConfirmationDialog creates a dialog for confirmations
func CreateConfirmationDialog(title, message string, onConfirm, onCancel string, open bool, attrs ...map[string]string) DialogComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Create dialog content with buttons
	var contentHTML strings.Builder
	contentHTML.WriteString(fmt.Sprintf(`<p class="mb-4">%s</p>`, message))
	contentHTML.WriteString(`<div class="flex gap-2 justify-end">`)
	contentHTML.WriteString(fmt.Sprintf(`<button type="button" class="btn btn-neutral" onclick="%s">Cancel</button>`, onCancel))
	contentHTML.WriteString(fmt.Sprintf(`<button type="button" class="btn btn-primary" onclick="%s">Confirm</button>`, onConfirm))
	contentHTML.WriteString(`</div>`)
	
	return DialogComponent{
		BaseComponent: base,
		Label:         title,
		Open:          open,
		Children:      []templ.Component{templ.Raw(contentHTML.String())},
	}
}

// CreateFilterDropdown creates a dropdown for filtering options
func CreateFilterDropdown(label string, options []MenuItem, selectedValue string, attrs ...map[string]string) DropdownComponent {
	// Mark selected option
	for i := range options {
		options[i].Checked = options[i].Value == selectedValue
		options[i].Type = "checkbox"
	}
	
	// Create button with current selection
	buttonText := label
	if selectedValue != "" {
		for _, option := range options {
			if option.Value == selectedValue {
				buttonText = option.Label
				break
			}
		}
	}
	
	return CreateSimpleDropdown(buttonText, options, attrs...)
}

// CreateBreadcrumbNav creates a breadcrumb navigation using badges
func CreateBreadcrumbNav(items []BreadcrumbItem, attrs ...map[string]string) []templ.Component {
	var breadcrumbs []templ.Component
	
	for i, item := range items {
		// Create badge for each breadcrumb item
		badge := &BadgeComponent{
			BaseComponent: BaseComponent{Class: "breadcrumb-item"},
			Content:       item.Label,
			Variant:       "neutral",
		}
		
		breadcrumbs = append(breadcrumbs, badge)
		
		// Add separator (except for last item)
		if i < len(items)-1 {
			separator := templ.Raw(`<span class="breadcrumb-separator">/</span>`)
			breadcrumbs = append(breadcrumbs, separator)
		}
	}
	
	return breadcrumbs
}

// BreadcrumbItem represents a breadcrumb navigation item
type BreadcrumbItem struct {
	Label string
	URL   string
	Active bool
}

// CreateStatsCard creates a card displaying statistics
func CreateStatsCard(title, value, description string, trend string, attrs ...map[string]string) CardComponent {
	base := mergeBaseAttrs(attrs...)
	
	// Add stats classes
	if base.Class == "" {
		base.Class = "stats-card"
	} else {
		base.Class += " stats-card"
	}
	
	// Build stats content
	var contentHTML strings.Builder
	contentHTML.WriteString(fmt.Sprintf(`<div class="stat-value">%s</div>`, value))
	if description != "" {
		contentHTML.WriteString(fmt.Sprintf(`<div class="stat-desc">%s</div>`, description))
	}
	if trend != "" {
		contentHTML.WriteString(fmt.Sprintf(`<div class="stat-trend">%s</div>`, trend))
	}
	
	return CardComponent{
		BaseComponent: base,
		Header:        title,
		Children:      []templ.Component{templ.Raw(contentHTML.String())},
	}
}
// createInputWithError creates an input component with error styling
func createInputWithError(name, inputType, placeholder, errorMsg string, required bool) InputComponent {
	class := "form-input"
	if errorMsg != "" {
		class += " error"
	}
	
	return InputComponent{
		BaseComponent: BaseComponent{
			ID:    name,
			Class: class,
		},
		Type:        inputType,
		Name:        name,
		Placeholder: placeholder,
		Required:    required,
	}
}