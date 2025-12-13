package webawesome

import (
	"time"

	"github.com/a-h/templ"
)

// BaseComponent provides common functionality for all Web Awesome components
type BaseComponent struct {
	ID        string            `json:"id,omitempty"`
	Class     string            `json:"class,omitempty"`
	Style     string            `json:"style,omitempty"`
	DataAttrs map[string]string `json:"dataAttrs,omitempty"`
	Events    map[string]string `json:"events,omitempty"`
}

// ButtonComponent wraps sl-button
type ButtonComponent struct {
	BaseComponent
	Variant  string             `json:"variant,omitempty"`  // primary, secondary, success, etc.
	Size     string             `json:"size,omitempty"`     // small, medium, large
	Disabled bool               `json:"disabled,omitempty"`
	Loading  bool               `json:"loading,omitempty"`
	Href     string             `json:"href,omitempty"`
	Type     string             `json:"type,omitempty"`     // button, submit, reset
	Children []templ.Component  `json:"-"`                  // Child components
}

// InputComponent wraps sl-input
type InputComponent struct {
	BaseComponent
	Type        string `json:"type,omitempty"`        // text, email, password, etc.
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	Readonly    bool   `json:"readonly,omitempty"`
	MinLength   int    `json:"minLength,omitempty"`
	MaxLength   int    `json:"maxLength,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
}

// SelectComponent wraps sl-select
type SelectComponent struct {
	BaseComponent
	Name        string         `json:"name,omitempty"`
	Value       string         `json:"value,omitempty"`
	Multiple    bool           `json:"multiple,omitempty"`
	Disabled    bool           `json:"disabled,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Options     []SelectOption `json:"options,omitempty"`
}

// SelectOption represents an option in a select component
type SelectOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
	Selected bool   `json:"selected,omitempty"`
}

// TextareaComponent wraps sl-textarea
type TextareaComponent struct {
	BaseComponent
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	Readonly    bool   `json:"readonly,omitempty"`
	Rows        int    `json:"rows,omitempty"`
	Cols        int    `json:"cols,omitempty"`
	MinLength   int    `json:"minLength,omitempty"`
	MaxLength   int    `json:"maxLength,omitempty"`
}

// CheckboxComponent wraps sl-checkbox
type CheckboxComponent struct {
	BaseComponent
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Checked  bool   `json:"checked,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Required bool   `json:"required,omitempty"`
	Label    string `json:"label,omitempty"`
}

// RadioComponent wraps sl-radio
type RadioComponent struct {
	BaseComponent
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Checked  bool   `json:"checked,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Required bool   `json:"required,omitempty"`
	Label    string `json:"label,omitempty"`
}

// CardComponent wraps sl-card
type CardComponent struct {
	BaseComponent
	Header   string            `json:"header,omitempty"`
	Footer   string            `json:"footer,omitempty"`
	Children []templ.Component `json:"-"`
}

// DialogComponent wraps sl-dialog
type DialogComponent struct {
	BaseComponent
	Label    string            `json:"label,omitempty"`
	Open     bool              `json:"open,omitempty"`
	NoHeader bool              `json:"noHeader,omitempty"`
	Children []templ.Component `json:"-"`
}

// DrawerComponent wraps sl-drawer
type DrawerComponent struct {
	BaseComponent
	Label     string            `json:"label,omitempty"`
	Open      bool              `json:"open,omitempty"`
	Placement string            `json:"placement,omitempty"` // top, end, bottom, start
	Children  []templ.Component `json:"-"`
}

// TabGroupComponent wraps sl-tab-group
type TabGroupComponent struct {
	BaseComponent
	Placement string `json:"placement,omitempty"` // top, bottom, start, end
	Tabs      []Tab  `json:"tabs,omitempty"`
}

// Tab represents a single tab in a tab group
type Tab struct {
	Panel    string `json:"panel"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
	Active   bool   `json:"active,omitempty"`
	Content  string `json:"content,omitempty"`
}

// TableComponent wraps sl-table (or provides table structure)
type TableComponent struct {
	BaseComponent
	Headers []string     `json:"headers,omitempty"`
	Rows    [][]string   `json:"rows,omitempty"`
	Caption string       `json:"caption,omitempty"`
}

// BadgeComponent wraps sl-badge
type BadgeComponent struct {
	BaseComponent
	Variant string `json:"variant,omitempty"` // primary, secondary, success, etc.
	Pill    bool   `json:"pill,omitempty"`
	Content string `json:"content,omitempty"`
}

// ProgressBarComponent wraps sl-progress-bar
type ProgressBarComponent struct {
	BaseComponent
	Value       int    `json:"value,omitempty"`       // 0-100
	Indeterminate bool `json:"indeterminate,omitempty"`
	Label       string `json:"label,omitempty"`
}

// SpinnerComponent wraps sl-spinner
type SpinnerComponent struct {
	BaseComponent
	Size string `json:"size,omitempty"` // small, medium, large
}

// DropdownComponent wraps sl-dropdown
type DropdownComponent struct {
	BaseComponent
	Open      bool              `json:"open,omitempty"`
	Placement string            `json:"placement,omitempty"` // top, bottom, left, right
	Distance  int               `json:"distance,omitempty"`
	Skidding  int               `json:"skidding,omitempty"`
	Trigger   templ.Component   `json:"-"`
	Children  []templ.Component `json:"-"`
}

// MenuComponent wraps sl-menu
type MenuComponent struct {
	BaseComponent
	Items []MenuItem `json:"items,omitempty"`
}

// MenuItem represents an item in a menu
type MenuItem struct {
	Label    string `json:"label"`
	Value    string `json:"value,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Checked  bool   `json:"checked,omitempty"`
	Type     string `json:"type,omitempty"` // normal, checkbox
}

// TooltipComponent wraps sl-tooltip
type TooltipComponent struct {
	BaseComponent
	Content   string            `json:"content,omitempty"`
	Placement string            `json:"placement,omitempty"` // top, bottom, left, right
	Trigger   string            `json:"trigger,omitempty"`   // hover, click, focus
	Children  []templ.Component `json:"-"`
}
// OptimizationHint provides suggestions for improving component usage
type OptimizationHint struct {
	Type        string    `json:"type"`        // "unused", "duplicate", "performance"
	Component   string    `json:"component"`   // Component name
	Message     string    `json:"message"`     // Human-readable message
	Severity    string    `json:"severity"`    // "info", "warning", "error"
	Suggestion  string    `json:"suggestion"`  // Recommended action
	CreatedAt   time.Time `json:"createdAt"`
}