package webawesome

import (
	"github.com/a-h/templ"
)

// This file contains examples of how to use Web Awesome components in Gothic Forge

// ExampleButton demonstrates various button configurations
func ExampleButtons() []ButtonComponent {
	return []ButtonComponent{
		// Primary button
		{
			BaseComponent: BaseComponent{ID: "btn-primary", Class: "mr-2"},
			Variant:       "primary",
			Size:          "medium",
			Type:          "button",
			Children:      []templ.Component{templ.Raw("Primary Button")},
		},
		// Secondary button
		{
			BaseComponent: BaseComponent{ID: "btn-secondary", Class: "mr-2"},
			Variant:       "neutral",
			Size:          "medium",
			Type:          "button",
			Children:      []templ.Component{templ.Raw("Secondary Button")},
		},
		// Success button
		{
			BaseComponent: BaseComponent{ID: "btn-success", Class: "mr-2"},
			Variant:       "success",
			Size:          "medium",
			Type:          "button",
			Children:      []templ.Component{templ.Raw("Success Button")},
		},
		// Danger button
		{
			BaseComponent: BaseComponent{ID: "btn-danger", Class: "mr-2"},
			Variant:       "danger",
			Size:          "medium",
			Type:          "button",
			Children:      []templ.Component{templ.Raw("Danger Button")},
		},
		// Loading button
		{
			BaseComponent: BaseComponent{ID: "btn-loading", Class: "mr-2"},
			Variant:       "primary",
			Size:          "medium",
			Type:          "button",
			Loading:       true,
			Children:      []templ.Component{templ.Raw("Loading...")},
		},
		// Disabled button
		{
			BaseComponent: BaseComponent{ID: "btn-disabled"},
			Variant:       "neutral",
			Size:          "medium",
			Type:          "button",
			Disabled:      true,
			Children:      []templ.Component{templ.Raw("Disabled Button")},
		},
	}
}

// ExampleForm demonstrates a complete form with various input types
func ExampleForm() []Component {
	return []Component{
		// Text input
		&InputComponent{
			BaseComponent: BaseComponent{ID: "username", Class: "mb-4"},
			Type:          "text",
			Name:          "username",
			Placeholder:   "Enter your username",
			Required:      true,
		},
		// Email input
		&InputComponent{
			BaseComponent: BaseComponent{ID: "email", Class: "mb-4"},
			Type:          "email",
			Name:          "email",
			Placeholder:   "Enter your email",
			Required:      true,
		},
		// Password input
		&InputComponent{
			BaseComponent: BaseComponent{ID: "password", Class: "mb-4"},
			Type:          "password",
			Name:          "password",
			Placeholder:   "Enter your password",
			Required:      true,
			MinLength:     8,
		},
		// Select dropdown
		&SelectComponent{
			BaseComponent: BaseComponent{ID: "country", Class: "mb-4"},
			Name:          "country",
			Placeholder:   "Select your country",
			Options: []SelectOption{
				{Value: "us", Label: "United States"},
				{Value: "ca", Label: "Canada"},
				{Value: "uk", Label: "United Kingdom"},
				{Value: "au", Label: "Australia"},
			},
		},
		// Textarea
		&TextareaComponent{
			BaseComponent: BaseComponent{ID: "message", Class: "mb-4"},
			Name:          "message",
			Placeholder:   "Enter your message",
			Rows:          4,
		},
		// Checkbox
		&CheckboxComponent{
			BaseComponent: BaseComponent{ID: "terms", Class: "mb-4"},
			Name:          "terms",
			Value:         "accepted",
			Label:         "I agree to the terms and conditions",
			Required:      true,
		},
		// Submit button
		&ButtonComponent{
			BaseComponent: BaseComponent{ID: "submit-btn"},
			Variant:       "primary",
			Type:          "submit",
			Children:      []templ.Component{templ.Raw("Submit Form")},
		},
	}
}

// ExampleCard demonstrates card layouts
func ExampleCards() []CardComponent {
	return []CardComponent{
		// Simple card
		{
			BaseComponent: BaseComponent{ID: "simple-card", Class: "mb-4"},
			Header:        "Simple Card",
			Children:      []templ.Component{templ.Raw("<p>This is a simple card with header and content.</p>")},
		},
		// Card with footer
		{
			BaseComponent: BaseComponent{ID: "card-with-footer", Class: "mb-4"},
			Header:        "Card with Footer",
			Footer:        `<div class="flex justify-end"><button class="btn btn-primary">Action</button></div>`,
			Children:      []templ.Component{templ.Raw("<p>This card has a header, content, and footer with an action button.</p>")},
		},
		// Status card
		{
			BaseComponent: BaseComponent{ID: "status-card", Class: "mb-4 border-success"},
			Header:        "Success Status",
			Children:      []templ.Component{templ.Raw("<p>This card indicates a successful operation.</p>")},
		},
	}
}

// ExampleInteractiveComponents demonstrates dropdowns, menus, and tooltips
func ExampleInteractiveComponents() []Component {
	return []Component{
		// Simple dropdown
		&DropdownComponent{
			BaseComponent: BaseComponent{ID: "simple-dropdown", Class: "mb-4"},
			Trigger: &ButtonComponent{
				BaseComponent: BaseComponent{},
				Variant:       "neutral",
				Type:          "button",
				Children:      []templ.Component{templ.Raw("Open Menu")},
			},
			Children: []templ.Component{
				&MenuComponent{
					BaseComponent: BaseComponent{},
					Items: []MenuItem{
						{Label: "Edit", Value: "edit"},
						{Label: "Delete", Value: "delete"},
						{Label: "Share", Value: "share"},
					},
				},
			},
		},
		// Tooltip
		&TooltipComponent{
			BaseComponent: BaseComponent{ID: "tooltip-example", Class: "mb-4"},
			Content:       "This is a helpful tooltip",
			Placement:     "top",
			Trigger:       "hover",
			Children: []templ.Component{
				&ButtonComponent{
					BaseComponent: BaseComponent{},
					Variant:       "neutral",
					Type:          "button",
					Children:      []templ.Component{templ.Raw("Hover for tooltip")},
				},
			},
		},
	}
}

// ExampleDataDisplay demonstrates badges, progress bars, and spinners
func ExampleDataDisplay() []Component {
	return []Component{
		// Badges
		&BadgeComponent{
			BaseComponent: BaseComponent{ID: "primary-badge", Class: "mr-2"},
			Content:       "Primary",
			Variant:       "primary",
		},
		&BadgeComponent{
			BaseComponent: BaseComponent{ID: "success-badge", Class: "mr-2"},
			Content:       "Success",
			Variant:       "success",
			Pill:          true,
		},
		&BadgeComponent{
			BaseComponent: BaseComponent{ID: "warning-badge", Class: "mr-2"},
			Content:       "Warning",
			Variant:       "warning",
		},
		// Progress bar
		&ProgressBarComponent{
			BaseComponent: BaseComponent{ID: "progress-example", Class: "mb-4"},
			Value:         75,
			Label:         "Loading progress",
		},
		// Indeterminate progress
		&ProgressBarComponent{
			BaseComponent: BaseComponent{ID: "indeterminate-progress", Class: "mb-4"},
			Indeterminate: true,
			Label:         "Processing...",
		},
		// Spinner
		&SpinnerComponent{
			BaseComponent: BaseComponent{ID: "spinner-example"},
			Size:          "medium",
		},
	}
}

// ExampleTable demonstrates table creation
func ExampleTable() TableComponent {
	return TableComponent{
		BaseComponent: BaseComponent{ID: "example-table", Class: "w-full"},
		Caption:       "Example Data Table",
		Headers:       []string{"Name", "Email", "Role", "Status"},
		Rows: [][]string{
			{"John Doe", "john@example.com", "Admin", "Active"},
			{"Jane Smith", "jane@example.com", "User", "Active"},
			{"Bob Johnson", "bob@example.com", "User", "Inactive"},
		},
	}
}

// ExampleTabs demonstrates tab group usage
func ExampleTabs() TabGroupComponent {
	return TabGroupComponent{
		BaseComponent: BaseComponent{ID: "example-tabs", Class: "mb-4"},
		Placement:     "top",
		Tabs: []Tab{
			{
				Panel:   "tab1",
				Label:   "Overview",
				Content: "<p>This is the overview tab content.</p>",
				Active:  true,
			},
			{
				Panel:   "tab2",
				Label:   "Details",
				Content: "<p>This is the details tab content.</p>",
			},
			{
				Panel:   "tab3",
				Label:   "Settings",
				Content: "<p>This is the settings tab content.</p>",
			},
		},
	}
}

// ExampleDialog demonstrates modal dialog usage
func ExampleDialog() DialogComponent {
	return DialogComponent{
		BaseComponent: BaseComponent{ID: "example-dialog"},
		Label:         "Confirmation Dialog",
		Open:          false, // Controlled by JavaScript
		Children: []templ.Component{
			templ.Raw(`
				<p>Are you sure you want to delete this item?</p>
				<div class="flex justify-end gap-2 mt-4">
					<button class="btn btn-neutral" onclick="closeDialog()">Cancel</button>
					<button class="btn btn-danger" onclick="confirmDelete()">Delete</button>
				</div>
			`),
		},
	}
}

// ExampleHTMXIntegration demonstrates HTMX integration
func ExampleHTMXComponents() []Component {
	return []Component{
		// HTMX Button
		&ButtonComponent{
			BaseComponent: BaseComponent{
				ID:    "htmx-button",
				Class: "mb-4",
				Events: map[string]string{
					"hx-get":    "/api/data",
					"hx-target": "#result",
					"hx-swap":   "innerHTML",
				},
			},
			Variant:  "primary",
			Type:     "button",
			Children: []templ.Component{templ.Raw("Load Data")},
		},
		// HTMX Form
		&InputComponent{
			BaseComponent: BaseComponent{
				ID: "htmx-input",
				Events: map[string]string{
					"hx-post":    "/api/search",
					"hx-target":  "#search-results",
					"hx-trigger": "keyup changed delay:500ms",
				},
			},
			Type:        "search",
			Name:        "query",
			Placeholder: "Search...",
		},
	}
}

// ExampleAlpineIntegration demonstrates Alpine.js integration
func ExampleAlpineComponents() []Component {
	return []Component{
		// Alpine Button
		&ButtonComponent{
			BaseComponent: BaseComponent{
				ID:    "alpine-button",
				Class: "mb-4",
				Events: map[string]string{
					"@click": "count++",
				},
			},
			Variant:  "primary",
			Type:     "button",
			Children: []templ.Component{templ.Raw("Increment Counter")},
		},
		// Alpine Input
		&InputComponent{
			BaseComponent: BaseComponent{
				ID: "alpine-input",
				Events: map[string]string{
					"x-model": "message",
				},
			},
			Type:        "text",
			Name:        "message",
			Placeholder: "Type something...",
		},
	}
}