package webawesome

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
)

// CardComponent implementation

// Render implements the Component interface for CardComponent
func (cc *CardComponent) Render(ctx context.Context, w io.Writer) error {
	if err := cc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := cc.GetAttributes()

	// Build card content
	var content strings.Builder
	
	// Add header if provided
	if cc.Header != "" {
		content.WriteString(fmt.Sprintf(`<div slot="header">%s</div>`, cc.Header))
	}
	
	// Add main content (children)
	if len(cc.Children) > 0 {
		for _, child := range cc.Children {
			if err := child.Render(ctx, &content); err != nil {
				return fmt.Errorf("failed to render card child: %w", err)
			}
		}
	}
	
	// Add footer if provided
	if cc.Footer != "" {
		content.WriteString(fmt.Sprintf(`<div slot="footer">%s</div>`, cc.Footer))
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-card", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for CardComponent
func (cc *CardComponent) Validate() error {
	return cc.BaseComponent.Validate()
}

// GetAttributes implements the Component interface for CardComponent
func (cc *CardComponent) GetAttributes() map[string]string {
	return MergeAttributes(cc.BaseComponent.GetAttributes())
}

// DialogComponent implementation

// Render implements the Component interface for DialogComponent
func (dc *DialogComponent) Render(ctx context.Context, w io.Writer) error {
	if err := dc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := dc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if dc.Label != "" {
		attrs["label"] = dc.Label
	}
	if dc.Open {
		attrs["open"] = ""
	}
	if dc.NoHeader {
		attrs["no-header"] = ""
	}

	// Build dialog content
	var content strings.Builder
	
	// Add children content
	if len(dc.Children) > 0 {
		for _, child := range dc.Children {
			if err := child.Render(ctx, &content); err != nil {
				return fmt.Errorf("failed to render dialog child: %w", err)
			}
		}
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-dialog", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for DialogComponent
func (dc *DialogComponent) Validate() error {
	return dc.BaseComponent.Validate()
}

// GetAttributes implements the Component interface for DialogComponent
func (dc *DialogComponent) GetAttributes() map[string]string {
	return MergeAttributes(dc.BaseComponent.GetAttributes())
}

// DrawerComponent implementation

// Render implements the Component interface for DrawerComponent
func (dc *DrawerComponent) Render(ctx context.Context, w io.Writer) error {
	if err := dc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := dc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if dc.Label != "" {
		attrs["label"] = dc.Label
	}
	if dc.Open {
		attrs["open"] = ""
	}
	if dc.Placement != "" {
		attrs["placement"] = dc.Placement
	}

	// Build drawer content
	var content strings.Builder
	
	// Add children content
	if len(dc.Children) > 0 {
		for _, child := range dc.Children {
			if err := child.Render(ctx, &content); err != nil {
				return fmt.Errorf("failed to render drawer child: %w", err)
			}
		}
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-drawer", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for DrawerComponent
func (dc *DrawerComponent) Validate() error {
	if err := dc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate placement
	if dc.Placement != "" {
		validPlacements := []string{"top", "end", "bottom", "start"}
		if err := Validator.ValidateEnum("placement", dc.Placement, validPlacements); err != nil {
			return err
		}
	}

	return nil
}

// GetAttributes implements the Component interface for DrawerComponent
func (dc *DrawerComponent) GetAttributes() map[string]string {
	return MergeAttributes(dc.BaseComponent.GetAttributes())
}

// TabGroupComponent implementation

// Render implements the Component interface for TabGroupComponent
func (tgc *TabGroupComponent) Render(ctx context.Context, w io.Writer) error {
	if err := tgc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := tgc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if tgc.Placement != "" {
		attrs["placement"] = tgc.Placement
	}

	// Build tab group content
	var content strings.Builder
	
	// Add tabs
	for _, tab := range tgc.Tabs {
		tabAttrs := make(map[string]string)
		tabAttrs["slot"] = "nav"
		tabAttrs["panel"] = tab.Panel
		
		if tab.Disabled {
			tabAttrs["disabled"] = ""
		}
		if tab.Active {
			tabAttrs["active"] = ""
		}
		
		tabHTML := BuildWebAwesomeTag("sl-tab", tabAttrs, tab.Label, false)
		content.WriteString(tabHTML)
	}
	
	// Add tab panels
	for _, tab := range tgc.Tabs {
		panelAttrs := make(map[string]string)
		panelAttrs["name"] = tab.Panel
		
		panelHTML := BuildWebAwesomeTag("sl-tab-panel", panelAttrs, tab.Content, false)
		content.WriteString(panelHTML)
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-tab-group", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for TabGroupComponent
func (tgc *TabGroupComponent) Validate() error {
	if err := tgc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate placement
	if tgc.Placement != "" {
		validPlacements := []string{"top", "bottom", "start", "end"}
		if err := Validator.ValidateEnum("placement", tgc.Placement, validPlacements); err != nil {
			return err
		}
	}

	// Validate tabs
	if len(tgc.Tabs) == 0 {
		return NewValidationError("tabs", "", "TabGroup must have at least one tab", "required")
	}

	// Validate tab panels are unique
	panels := make(map[string]bool)
	activeCount := 0
	
	for i, tab := range tgc.Tabs {
		if tab.Panel == "" {
			return NewValidationError(fmt.Sprintf("tabs[%d].panel", i), tab.Panel, "Tab panel name cannot be empty", "required")
		}
		if tab.Label == "" {
			return NewValidationError(fmt.Sprintf("tabs[%d].label", i), tab.Label, "Tab label cannot be empty", "required")
		}
		if panels[tab.Panel] {
			return NewValidationError(fmt.Sprintf("tabs[%d].panel", i), tab.Panel, "Duplicate tab panel name", "duplicate")
		}
		panels[tab.Panel] = true
		
		if tab.Active {
			activeCount++
		}
	}

	// Ensure only one tab is active
	if activeCount > 1 {
		return NewValidationError("tabs", "", "Only one tab can be active", "invalid_state")
	}

	return nil
}

// GetAttributes implements the Component interface for TabGroupComponent
func (tgc *TabGroupComponent) GetAttributes() map[string]string {
	return MergeAttributes(tgc.BaseComponent.GetAttributes())
}

// TableComponent implementation (custom implementation since Web Awesome doesn't have sl-table)

// Render implements the Component interface for TableComponent
func (tc *TableComponent) Render(ctx context.Context, w io.Writer) error {
	if err := tc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := tc.GetAttributes()
	
	// Add default table classes if not provided
	if attrs["class"] == "" {
		attrs["class"] = "table table-auto"
	} else {
		attrs["class"] += " table table-auto"
	}

	// Build table content
	var content strings.Builder
	
	// Add caption if provided
	if tc.Caption != "" {
		content.WriteString(fmt.Sprintf("<caption>%s</caption>", tc.Caption))
	}
	
	// Add headers
	if len(tc.Headers) > 0 {
		content.WriteString("<thead><tr>")
		for _, header := range tc.Headers {
			content.WriteString(fmt.Sprintf("<th>%s</th>", header))
		}
		content.WriteString("</tr></thead>")
	}
	
	// Add rows
	if len(tc.Rows) > 0 {
		content.WriteString("<tbody>")
		for _, row := range tc.Rows {
			content.WriteString("<tr>")
			for _, cell := range row {
				content.WriteString(fmt.Sprintf("<td>%s</td>", cell))
			}
			content.WriteString("</tr>")
		}
		content.WriteString("</tbody>")
	}

	// Render the component using standard HTML table
	html := BuildWebAwesomeTag("table", attrs, content.String(), false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for TableComponent
func (tc *TableComponent) Validate() error {
	if err := tc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate that if headers are provided, all rows have the same number of columns
	if len(tc.Headers) > 0 {
		headerCount := len(tc.Headers)
		for i, row := range tc.Rows {
			if len(row) != headerCount {
				return NewValidationError(fmt.Sprintf("rows[%d]", i), "", 
					fmt.Sprintf("Row has %d columns but table has %d headers", len(row), headerCount), 
					"column_mismatch")
			}
		}
	}

	return nil
}

// GetAttributes implements the Component interface for TableComponent
func (tc *TableComponent) GetAttributes() map[string]string {
	return MergeAttributes(tc.BaseComponent.GetAttributes())
}

// BadgeComponent implementation

// Render implements the Component interface for BadgeComponent
func (bc *BadgeComponent) Render(ctx context.Context, w io.Writer) error {
	if err := bc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := bc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if bc.Variant != "" {
		attrs["variant"] = bc.Variant
	}
	if bc.Pill {
		attrs["pill"] = ""
	}

	// Render the component
	html := BuildWebAwesomeTag("sl-badge", attrs, bc.Content, false)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for BadgeComponent
func (bc *BadgeComponent) Validate() error {
	if err := bc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate variant
	if bc.Variant != "" {
		validVariants := []string{"primary", "success", "neutral", "warning", "danger"}
		if err := Validator.ValidateEnum("variant", bc.Variant, validVariants); err != nil {
			return err
		}
	}

	return nil
}

// GetAttributes implements the Component interface for BadgeComponent
func (bc *BadgeComponent) GetAttributes() map[string]string {
	return MergeAttributes(bc.BaseComponent.GetAttributes())
}

// ProgressBarComponent implementation

// Render implements the Component interface for ProgressBarComponent
func (pbc *ProgressBarComponent) Render(ctx context.Context, w io.Writer) error {
	if err := pbc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := pbc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if pbc.Value > 0 {
		attrs["value"] = fmt.Sprintf("%d", pbc.Value)
	}
	if pbc.Indeterminate {
		attrs["indeterminate"] = ""
	}
	if pbc.Label != "" {
		attrs["label"] = pbc.Label
	}

	// Render the component (self-closing)
	html := BuildWebAwesomeTag("sl-progress-bar", attrs, "", true)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for ProgressBarComponent
func (pbc *ProgressBarComponent) Validate() error {
	if err := pbc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate value range
	if pbc.Value < 0 || pbc.Value > 100 {
		return NewValidationError("value", fmt.Sprintf("%d", pbc.Value), "Progress value must be between 0 and 100", "out_of_range")
	}

	return nil
}

// GetAttributes implements the Component interface for ProgressBarComponent
func (pbc *ProgressBarComponent) GetAttributes() map[string]string {
	return MergeAttributes(pbc.BaseComponent.GetAttributes())
}

// SpinnerComponent implementation

// Render implements the Component interface for SpinnerComponent
func (sc *SpinnerComponent) Render(ctx context.Context, w io.Writer) error {
	if err := sc.Validate(); err != nil {
		return err
	}

	// Build attributes
	attrs := sc.GetAttributes()
	
	// Add Web Awesome specific attributes
	if sc.Size != "" {
		attrs["size"] = sc.Size
	}

	// Render the component (self-closing)
	html := BuildWebAwesomeTag("sl-spinner", attrs, "", true)
	_, err := io.WriteString(w, html)
	return err
}

// Validate implements the Component interface for SpinnerComponent
func (sc *SpinnerComponent) Validate() error {
	if err := sc.BaseComponent.Validate(); err != nil {
		return err
	}

	// Validate size
	if sc.Size != "" {
		validSizes := []string{"small", "medium", "large"}
		if err := Validator.ValidateEnum("size", sc.Size, validSizes); err != nil {
			return err
		}
	}

	return nil
}

// GetAttributes implements the Component interface for SpinnerComponent
func (sc *SpinnerComponent) GetAttributes() map[string]string {
	return MergeAttributes(sc.BaseComponent.GetAttributes())
}

// Constructor functions for layout components

// NewCardComponent creates a new CardComponent
func NewCardComponent() *CardComponent {
	return &CardComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewDialogComponent creates a new DialogComponent
func NewDialogComponent() *DialogComponent {
	return &DialogComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewDrawerComponent creates a new DrawerComponent
func NewDrawerComponent() *DrawerComponent {
	return &DrawerComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewTabGroupComponent creates a new TabGroupComponent
func NewTabGroupComponent() *TabGroupComponent {
	return &TabGroupComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewTableComponent creates a new TableComponent
func NewTableComponent() *TableComponent {
	return &TableComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewBadgeComponent creates a new BadgeComponent
func NewBadgeComponent() *BadgeComponent {
	return &BadgeComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewProgressBarComponent creates a new ProgressBarComponent
func NewProgressBarComponent() *ProgressBarComponent {
	return &ProgressBarComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// NewSpinnerComponent creates a new SpinnerComponent
func NewSpinnerComponent() *SpinnerComponent {
	return &SpinnerComponent{
		BaseComponent: BaseComponent{
			DataAttrs: make(map[string]string),
			Events:    make(map[string]string),
		},
	}
}

// Helper methods for layout components

// SetHeader sets the header content for CardComponent
func (cc *CardComponent) SetHeader(header string) {
	cc.Header = header
}

// SetFooter sets the footer content for CardComponent
func (cc *CardComponent) SetFooter(footer string) {
	cc.Footer = footer
}

// AddChild adds a child component to CardComponent
func (cc *CardComponent) AddChild(child templ.Component) {
	cc.Children = append(cc.Children, child)
}

// SetLabel sets the label for DialogComponent
func (dc *DialogComponent) SetLabel(label string) {
	dc.Label = label
}

// Show opens the dialog
func (dc *DialogComponent) Show() {
	dc.Open = true
}

// Hide closes the dialog
func (dc *DialogComponent) Hide() {
	dc.Open = false
}

// AddChild adds a child component to DialogComponent
func (dc *DialogComponent) AddChild(child templ.Component) {
	dc.Children = append(dc.Children, child)
}

// SetPlacement sets the placement for DrawerComponent
func (dc *DrawerComponent) SetPlacement(placement string) {
	dc.Placement = placement
}

// Show opens the drawer
func (dc *DrawerComponent) Show() {
	dc.Open = true
}

// Hide closes the drawer
func (dc *DrawerComponent) Hide() {
	dc.Open = false
}

// AddTab adds a tab to TabGroupComponent
func (tgc *TabGroupComponent) AddTab(panel, label, content string) {
	tab := Tab{
		Panel:   panel,
		Label:   label,
		Content: content,
	}
	tgc.Tabs = append(tgc.Tabs, tab)
}

// SetActiveTab sets the active tab by panel name
func (tgc *TabGroupComponent) SetActiveTab(panel string) {
	for i := range tgc.Tabs {
		tgc.Tabs[i].Active = (tgc.Tabs[i].Panel == panel)
	}
}

// AddHeader adds a header to TableComponent
func (tc *TableComponent) AddHeader(header string) {
	tc.Headers = append(tc.Headers, header)
}

// AddRow adds a row to TableComponent
func (tc *TableComponent) AddRow(row []string) {
	tc.Rows = append(tc.Rows, row)
}

// SetCaption sets the caption for TableComponent
func (tc *TableComponent) SetCaption(caption string) {
	tc.Caption = caption
}

// SetContent sets the content for BadgeComponent
func (bc *BadgeComponent) SetContent(content string) {
	bc.Content = content
}

// SetVariant sets the variant for BadgeComponent
func (bc *BadgeComponent) SetVariant(variant string) {
	bc.Variant = variant
}

// SetValue sets the progress value for ProgressBarComponent
func (pbc *ProgressBarComponent) SetValue(value int) {
	pbc.Value = value
}

// SetIndeterminate sets the indeterminate state for ProgressBarComponent
func (pbc *ProgressBarComponent) SetIndeterminate(indeterminate bool) {
	pbc.Indeterminate = indeterminate
}

// SetSize sets the size for SpinnerComponent
func (sc *SpinnerComponent) SetSize(size string) {
	sc.Size = size
}

// RenderHTML renders the component to an HTML string
func (cc *CardComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := cc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (dc *DialogComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := dc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders the component to an HTML string
func (tc *TableComponent) RenderHTML() (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := tc.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}