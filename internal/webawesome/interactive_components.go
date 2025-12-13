package webawesome

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
)

// mergeBaseAttrs merges multiple attribute maps into a BaseComponent
func mergeBaseAttrs(attrs ...map[string]string) BaseComponent {
	base := BaseComponent{
		DataAttrs: make(map[string]string),
		Events:    make(map[string]string),
	}
	
	for _, attrMap := range attrs {
		for key, value := range attrMap {
			switch key {
			case "id":
				base.ID = value
			case "class":
				if base.Class == "" {
					base.Class = value
				} else {
					base.Class += " " + value
				}
			case "style":
				if base.Style == "" {
					base.Style = value
				} else {
					base.Style += "; " + value
				}
			default:
				// Assume it's a data attribute or event
				if strings.HasPrefix(key, "hx-") || strings.HasPrefix(key, "data-") {
					base.DataAttrs[key] = value
				} else {
					base.Events[key] = value
				}
			}
		}
	}
	
	return base
}

// DropdownComponent implementation

// Render implements the Component interface for DropdownComponent
func (dc *DropdownComponent) Render(ctx context.Context, w io.Writer) error {
	if err := dc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := dc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if dc.Open {
		attrs["open"] = ""
	}
	if dc.Placement != "" {
		attrs["placement"] = dc.Placement
	}
	if dc.Distance > 0 {
		attrs["distance"] = fmt.Sprintf("%d", dc.Distance)
	}
	if dc.Skidding != 0 {
		attrs["skidding"] = fmt.Sprintf("%d", dc.Skidding)
	}

	// Build dropdown content
	var content strings.Builder
	
	// Add trigger element
	if dc.Trigger != nil {
		content.WriteString(`<div slot="trigger">`)
		if err := dc.Trigger.Render(ctx, &content); err != nil {
			return fmt.Errorf("failed to render dropdown trigger: %w", err)
		}
		content.WriteString(`</div>`)
	}
	
	// Add dropdown content (children)
	if len(dc.Children) > 0 {
		for _, child := range dc.Children {
			if err := child.Render(ctx, &content); err != nil {
				return fmt.Errorf("failed to render dropdown child: %w", err)
			}
		}
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-dropdown", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for DropdownComponent
func (dc *DropdownComponent) Validate() error {
	if err := dc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate placement
	if dc.Placement != "" {
		validPlacements := []string{"top", "top-start", "top-end", "bottom", "bottom-start", "bottom-end", "right", "right-start", "right-end", "left", "left-start", "left-end"}
		if err := Validator.ValidateEnum("placement", dc.Placement, validPlacements); err != nil {
			return err
		}
	}

	return nil
}

// GetAttributes implements the Component interface for DropdownComponent
func (dc *DropdownComponent) GetAttributes() map[string]string {
	return MergeAttributes(dc.BaseComponent.GetAttributes())
}

// MenuComponent implementation

// Render implements the Component interface for MenuComponent
func (mc *MenuComponent) Render(ctx context.Context, w io.Writer) error {
	if err := mc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := mc.GetAttributes()

	// Build menu content
	var content strings.Builder
	
	// Add menu items
	for _, item := range mc.Items {
		itemAttrs := make(map[string]string)
		
		if item.Value != "" {
			itemAttrs["value"] = item.Value
		}
		if item.Disabled {
			itemAttrs["disabled"] = ""
		}
		if item.Checked {
			itemAttrs["checked"] = ""
		}
		if item.Type != "" {
			itemAttrs["type"] = item.Type
		}
		
		itemHTML := BuildWebAwesomeTag("sl-menu-item", itemAttrs, item.Label, false)
		content.WriteString(itemHTML)
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-menu", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for MenuComponent
func (mc *MenuComponent) Validate() error {
	if err := mc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate menu items
	if len(mc.Items) == 0 {
		return NewValidationError("items", "", "Menu must have at least one item", "required")
	}

	for i, item := range mc.Items {
		if item.Label == "" {
			return NewValidationError(fmt.Sprintf("items[%d].label", i), item.Label, "Menu item label cannot be empty", "required")
		}
		
		// Validate item type
		if item.Type != "" {
			validTypes := []string{"normal", "checkbox"}
			if err := Validator.ValidateEnum(fmt.Sprintf("items[%d].type", i), item.Type, validTypes); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAttributes implements the Component interface for MenuComponent
func (mc *MenuComponent) GetAttributes() map[string]string {
	return MergeAttributes(mc.BaseComponent.GetAttributes())
}

// TooltipComponent implementation

// Render implements the Component interface for TooltipComponent
func (tc *TooltipComponent) Render(ctx context.Context, w io.Writer) error {
	if err := tc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := tc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if tc.Content != "" {
		attrs["content"] = tc.Content
	}
	if tc.Placement != "" {
		attrs["placement"] = tc.Placement
	}
	if tc.Trigger != "" {
		attrs["trigger"] = tc.Trigger
	}

	// Build tooltip content (the element that triggers the tooltip)
	var content strings.Builder
	
	// Add children content (the trigger element)
	if len(tc.Children) > 0 {
		for _, child := range tc.Children {
			if err := child.Render(ctx, &content); err != nil {
				return fmt.Errorf("failed to render tooltip child: %w", err)
			}
		}
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-tooltip", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for TooltipComponent
func (tc *TooltipComponent) Validate() error {
	if err := tc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate content is provided
	if tc.Content == "" {
		return NewValidationError("content", tc.Content, "Tooltip content cannot be empty", "required")
	}

	// Validate placement
	if tc.Placement != "" {
		validPlacements := []string{"top", "top-start", "top-end", "right", "right-start", "right-end", "bottom", "bottom-start", "bottom-end", "left", "left-start", "left-end"}
		if err := Validator.ValidateEnum("placement", tc.Placement, validPlacements); err != nil {
			return err
		}
	}

	// Validate trigger
	if tc.Trigger != "" {
		validTriggers := []string{"hover", "click", "focus", "manual"}
		if err := Validator.ValidateEnum("trigger", tc.Trigger, validTriggers); err != nil {
			return err
		}
	}

	// Validate children exist (tooltip needs something to attach to)
	if len(tc.Children) == 0 {
		return NewValidationError("children", "", "Tooltip must have child elements to attach to", "required")
	}

	return nil
}

// GetAttributes implements the Component interface for TooltipComponent
func (tc *TooltipComponent) GetAttributes() map[string]string {
	return MergeAttributes(tc.BaseComponent.GetAttributes())
}

// Helper functions for creating interactive components

// CreateSimpleDropdown creates a dropdown with a button trigger
func CreateSimpleDropdown(buttonText string, menuItems []MenuItem, attrs ...map[string]string) DropdownComponent {
	// Create button trigger
	buttonComponent := &ButtonComponent{
		BaseComponent: BaseComponent{},
		Variant:       "neutral",
		Type:          "button",
		Children:      []templ.Component{templ.Raw(buttonText)},
	}
	
	// Create menu
	menuComponent := &MenuComponent{
		BaseComponent: BaseComponent{},
		Items:         menuItems,
	}
	
	return DropdownComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Trigger:       buttonComponent,
		Children:      []templ.Component{menuComponent},
	}
}

// CreateContextMenu creates a context menu with menu items
func CreateContextMenu(items []MenuItem, attrs ...map[string]string) MenuComponent {
	return MenuComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Items:         items,
	}
}

// CreateTooltipButton creates a button with a tooltip
func CreateTooltipButton(buttonText, tooltipText, variant string, attrs ...map[string]string) TooltipComponent {
	// Create button
	buttonComponent := &ButtonComponent{
		BaseComponent: BaseComponent{},
		Variant:       variant,
		Type:          "button",
		Children:      []templ.Component{templ.Raw(buttonText)},
	}
	
	return TooltipComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Content:       tooltipText,
		Placement:     "top",
		Trigger:       "hover",
		Children:      []templ.Component{buttonComponent},
	}
}

// CreateHelpTooltip creates a help icon with tooltip
func CreateHelpTooltip(helpText string, attrs ...map[string]string) TooltipComponent {
	// Create help icon (using text for now, could be replaced with actual icon)
	helpIcon := templ.Raw(`<span class="help-icon">?</span>`)
	
	return TooltipComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Content:       helpText,
		Placement:     "top",
		Trigger:       "hover",
		Children:      []templ.Component{helpIcon},
	}
}

// CreateDropdownMenu creates a dropdown with custom trigger and menu
func CreateDropdownMenu(trigger templ.Component, menuItems []MenuItem, placement string, attrs ...map[string]string) DropdownComponent {
	// Create menu
	menuComponent := &MenuComponent{
		BaseComponent: BaseComponent{},
		Items:         menuItems,
	}
	
	return DropdownComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Trigger:       trigger,
		Placement:     placement,
		Children:      []templ.Component{menuComponent},
	}
}

// CreateCheckboxMenu creates a menu with checkbox items
func CreateCheckboxMenu(items []MenuItem, attrs ...map[string]string) MenuComponent {
	// Set all items to checkbox type
	for i := range items {
		items[i].Type = "checkbox"
	}
	
	return MenuComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Items:         items,
	}
}

// CreateActionMenu creates a menu with action items (no checkboxes)
func CreateActionMenu(items []MenuItem, attrs ...map[string]string) MenuComponent {
	// Ensure all items are normal type
	for i := range items {
		if items[i].Type == "" {
			items[i].Type = "normal"
		}
	}
	
	return MenuComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Items:         items,
	}
}

// CreateTooltipWithPlacement creates a tooltip with specific placement
func CreateTooltipWithPlacement(content, placement, trigger string, children []templ.Component, attrs ...map[string]string) TooltipComponent {
	return TooltipComponent{
		BaseComponent: mergeBaseAttrs(attrs...),
		Content:       content,
		Placement:     placement,
		Trigger:       trigger,
		Children:      children,
	}
}

// HTMX Integration for Interactive Components

// CreateHTMXDropdown creates a dropdown that loads content via HTMX
func CreateHTMXDropdown(buttonText, hxGet, hxTarget string, attrs ...map[string]string) DropdownComponent {
	// Create HTMX button trigger
	base := mergeBaseAttrs(attrs...)
	if base.Events == nil {
		base.Events = make(map[string]string)
	}
	base.Events["hx-get"] = hxGet
	if hxTarget != "" {
		base.Events["hx-target"] = hxTarget
	}
	
	buttonComponent := &ButtonComponent{
		BaseComponent: base,
		Variant:       "neutral",
		Type:          "button",
		Children:      []templ.Component{templ.Raw(buttonText)},
	}
	
	return DropdownComponent{
		BaseComponent: BaseComponent{},
		Trigger:       buttonComponent,
		Children:      []templ.Component{templ.Raw(`<div class="dropdown-content"></div>`)},
	}
}

// CreateHTMXMenu creates a menu that triggers HTMX requests
func CreateHTMXMenu(items []MenuItem, hxTarget string, attrs ...map[string]string) MenuComponent {
	// Add HTMX attributes to menu items via data attributes
	base := mergeBaseAttrs(attrs...)
	if base.DataAttrs == nil {
		base.DataAttrs = make(map[string]string)
	}
	if hxTarget != "" {
		base.DataAttrs["hx-target"] = hxTarget
	}
	
	return MenuComponent{
		BaseComponent: base,
		Items:         items,
	}
}

// Constructor functions for interactive components

// NewDropdownComponent creates a new DropdownComponent
func NewDropdownComponent() *DropdownComponent {
	return &DropdownComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewMenuComponent creates a new MenuComponent
func NewMenuComponent() *MenuComponent {
	return &MenuComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewTooltipComponent creates a new TooltipComponent
func NewTooltipComponent() *TooltipComponent {
	return &TooltipComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// Helper methods for interactive components

// SetTrigger sets the trigger component for DropdownComponent
func (dc *DropdownComponent) SetTrigger(trigger templ.Component) {
	dc.Trigger = trigger
}

// AddChild adds a child component to DropdownComponent
func (dc *DropdownComponent) AddChild(child templ.Component) {
	dc.Children = append(dc.Children, child)
}

// SetPlacement sets the placement for DropdownComponent
func (dc *DropdownComponent) SetPlacement(placement string) {
	dc.Placement = placement
}

// Show opens the dropdown
func (dc *DropdownComponent) Show() {
	dc.Open = true
}

// Hide closes the dropdown
func (dc *DropdownComponent) Hide() {
	dc.Open = false
}

// AddItem adds a menu item to MenuComponent
func (mc *MenuComponent) AddItem(label, value string) {
	item := MenuItem{
		Label: label,
		Value: value,
		Type:  "normal",
	}
	mc.Items = append(mc.Items, item)
}

// AddCheckboxItem adds a checkbox menu item to MenuComponent
func (mc *MenuComponent) AddCheckboxItem(label, value string, checked bool) {
	item := MenuItem{
		Label:   label,
		Value:   value,
		Type:    "checkbox",
		Checked: checked,
	}
	mc.Items = append(mc.Items, item)
}

// AddSeparator adds a separator to MenuComponent (using disabled item)
func (mc *MenuComponent) AddSeparator() {
	item := MenuItem{
		Label:    "---",
		Disabled: true,
	}
	mc.Items = append(mc.Items, item)
}

// SetContent sets the tooltip content for TooltipComponent
func (tc *TooltipComponent) SetContent(content string) {
	tc.Content = content
}

// SetPlacement sets the placement for TooltipComponent
func (tc *TooltipComponent) SetPlacement(placement string) {
	tc.Placement = placement
}

// SetTrigger sets the trigger type for TooltipComponent
func (tc *TooltipComponent) SetTrigger(trigger string) {
	tc.Trigger = trigger
}

// AddChild adds a child component to TooltipComponent
func (tc *TooltipComponent) AddChild(child templ.Component) {
	tc.Children = append(tc.Children, child)
}

// RenderHTML renders the component to an HTML string
func (dc *DropdownComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := dc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (mc *MenuComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := mc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (tc *TooltipComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := tc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}