# Web Awesome Integration for Gothic Forge

This package provides comprehensive Web Awesome component integration for Gothic Forge, enabling the use of Web Awesome components with type-safe Go templates and server-side rendering.

## Features

- **Type-safe Web Awesome component wrappers** - Full Go type safety for all Web Awesome components
- **Templ integration** - Seamless server-side rendering with Templ templates
- **Theme system** - CSS custom properties and theme management
- **Component registry** - Usage tracking and optimization hints
- **Migration tools** - DaisyUI to Web Awesome migration support
- **HTMX and Alpine.js compatibility** - Works seamlessly with existing Gothic Forge stack
- **Comprehensive validation** - Built-in validation for all component properties
- **Helper methods** - Fluent API for easy component configuration

## Components Overview

### Form Components
- **ButtonComponent** (`sl-button`) - Buttons with variants, sizes, and states
- **InputComponent** (`sl-input`) - Text inputs with validation and types
- **SelectComponent** (`sl-select`) - Dropdown selects with options
- **TextareaComponent** (`sl-textarea`) - Multi-line text inputs
- **CheckboxComponent** (`sl-checkbox`) - Checkbox inputs with labels
- **RadioComponent** (`sl-radio`) - Radio button inputs

### Layout Components
- **CardComponent** (`sl-card`) - Cards with headers, content, and footers
- **DialogComponent** (`sl-dialog`) - Modal dialogs with customizable content
- **DrawerComponent** (`sl-drawer`) - Side drawers with placement options
- **TabGroupComponent** (`sl-tab-group`) - Tabbed interfaces with panels
- **TableComponent** (custom) - Data tables with headers and rows

### Data Display Components
- **BadgeComponent** (`sl-badge`) - Status badges with variants and pill style
- **ProgressBarComponent** (`sl-progress-bar`) - Progress indicators with values
- **SpinnerComponent** (`sl-spinner`) - Loading spinners with sizes

### Interactive Components
- **DropdownComponent** (`sl-dropdown`) - Dropdown menus with triggers
- **MenuComponent** (`sl-menu`) - Context menus with items and separators
- **TooltipComponent** (`sl-tooltip`) - Tooltips with placement and triggers

## Quick Start

### Basic Component Usage

```go
package main

import (
    "fmt"
    "github.com/your-org/gothic-forge/internal/webawesome"
)

func main() {
    // Create a primary button
    button := webawesome.NewButtonComponent()
    button.SetID("submit-btn")
    button.Variant = "primary"
    button.Type = "submit"
    button.SetClass("my-custom-class")
    
    // Render to HTML string
    html, err := button.RenderHTML()
    if err != nil {
        panic(err)
    }
    
    fmt.Println(html)
    // Output: <sl-button id="submit-btn" class="my-custom-class" variant="primary" type="submit"></sl-button>
}
```

### Form Components

```go
// Create a complete login form
func CreateLoginForm() string {
    var html strings.Builder
    
    // Email input
    emailInput := webawesome.NewInputComponent()
    emailInput.SetType("email")
    emailInput.SetName("email")
    emailInput.SetPlaceholder("Enter your email")
    emailInput.SetRequired(true)
    
    // Password input
    passwordInput := webawesome.NewInputComponent()
    passwordInput.SetType("password")
    passwordInput.SetName("password")
    passwordInput.SetPlaceholder("Enter your password")
    passwordInput.SetRequired(true)
    
    // Remember me checkbox
    rememberCheckbox := webawesome.NewCheckboxComponent()
    rememberCheckbox.SetName("remember")
    rememberCheckbox.SetValue("1")
    rememberCheckbox.SetLabel("Remember me")
    
    // Submit button
    submitButton := webawesome.NewButtonComponent()
    submitButton.Type = "submit"
    submitButton.Variant = "primary"
    
    // Render all components...
    return html.String()
}
```

### Layout Components

```go
// Create a user profile card
func CreateProfileCard() string {
    profileCard := webawesome.NewCardComponent()
    profileCard.SetID("profile-card")
    profileCard.SetHeader("User Profile")
    profileCard.SetFooter(`<sl-button variant="primary">Edit Profile</sl-button>`)
    
    // Add content as children
    profileContent := templ.Raw(`
        <div class="profile-info">
            <h3>John Doe</h3>
            <p>Software Developer</p>
            <p>john.doe@example.com</p>
        </div>
    `)
    profileCard.AddChild(profileContent)
    
    html, _ := profileCard.RenderHTML()
    return html
}
```

### Interactive Components

```go
// Create a dropdown menu
func CreateUserMenu() string {
    // Define menu items
    menuItems := []webawesome.MenuItem{
        {Label: "Profile", Value: "profile", Type: "normal"},
        {Label: "Settings", Value: "settings", Type: "normal"},
        {Label: "---", Disabled: true}, // Separator
        {Label: "Logout", Value: "logout", Type: "normal"},
    }
    
    // Create dropdown with button trigger
    dropdown := webawesome.CreateSimpleDropdown("User Menu", menuItems)
    dropdown.SetID("user-dropdown")
    dropdown.SetPlacement("bottom-end")
    
    html, _ := dropdown.RenderHTML()
    return html
}
```

### Data Display Components

```go
// Create a data table
func CreateUsersTable() string {
    table := webawesome.NewTableComponent()
    table.SetID("users-table")
    table.SetCaption("User Management")
    
    // Add headers
    table.AddHeader("ID")
    table.AddHeader("Name")
    table.AddHeader("Email")
    table.AddHeader("Role")
    
    // Add data rows
    table.AddRow([]string{"1", "John Doe", "john@example.com", "Admin"})
    table.AddRow([]string{"2", "Jane Smith", "jane@example.com", "User"})
    
    html, _ := table.RenderHTML()
    return html
}
```

## Advanced Usage

### Component Validation

All components include built-in validation:

```go
input := webawesome.NewInputComponent()
input.SetType("invalid-type") // This will fail validation
input.SetName("") // This will also fail validation

if err := input.Validate(); err != nil {
    fmt.Printf("Validation error: %v\n", err)
    // Handle validation error
}
```

### HTMX Integration

Components work seamlessly with HTMX:

```go
// Create HTMX-enabled dropdown
dropdown := webawesome.CreateHTMXDropdown(
    "Load Content", 
    "/api/content", 
    "#content-area",
)

// Create HTMX-enabled menu
menu := webawesome.CreateHTMXMenu(menuItems, "#main-content")
```

### Theme System

```go
// Load and apply a theme
themeManager := webawesome.NewThemeManager()
theme, err := themeManager.LoadTheme("dark-theme")
if err != nil {
    log.Fatal(err)
}

// Generate CSS for the theme
css, err := themeManager.GenerateCSS(theme)
if err != nil {
    log.Fatal(err)
}

// Apply theme to components
themeManager.ApplyTheme("dark-theme")
```

### Component Registry

Track component usage and get optimization hints:

```go
registry := webawesome.NewComponentRegistry()

// Register components
registry.Register("login-button", button)
registry.Register("user-form", form)

// Get usage statistics
stats := registry.GetStats()
fmt.Printf("Total components: %d\n", stats.TotalComponents)

// Get optimization hints
hints := registry.GetOptimizationHints()
for _, hint := range hints {
    fmt.Printf("Optimization: %s - %s\n", hint.Type, hint.Message)
}
```

## Component Reference

### ButtonComponent

```go
button := webawesome.NewButtonComponent()
button.SetID("my-button")           // Set ID attribute
button.SetClass("custom-class")     // Set CSS class
button.Variant = "primary"          // primary, success, neutral, warning, danger, text
button.Size = "medium"              // small, medium, large
button.Type = "button"              // button, submit, reset
button.Disabled = false             // Enable/disable button
button.Loading = false              // Show loading state
button.Href = "/link"               // Make button a link
```

### InputComponent

```go
input := webawesome.NewInputComponent()
input.SetName("username")           // Required: form field name
input.SetType("text")               // text, email, password, number, etc.
input.SetValue("default")           // Default value
input.SetPlaceholder("Enter text")  // Placeholder text
input.SetRequired(true)             // Make field required
input.SetDisabled(false)            // Enable/disable input
input.SetReadonly(false)            // Make input readonly
input.MinLength = 3                 // Minimum length
input.MaxLength = 50                // Maximum length
input.Pattern = "[A-Za-z]+"         // Validation pattern
```

### SelectComponent

```go
select := webawesome.NewSelectComponent()
select.SetName("country")           // Required: form field name
select.SetPlaceholder("Choose...")  // Placeholder text
select.SetMultiple(false)           // Allow multiple selections
select.SetRequired(true)            // Make field required
select.SetDisabled(false)           // Enable/disable select

// Add options
select.AddOption("us", "United States", false)
select.AddOption("ca", "Canada", true)  // Selected option
select.AddOption("uk", "United Kingdom", false)
```

### CardComponent

```go
card := webawesome.NewCardComponent()
card.SetID("profile-card")
card.SetHeader("Card Title")        // Optional header
card.SetFooter("Card Footer")       // Optional footer

// Add content as children
content := templ.Raw("<p>Card content goes here</p>")
card.AddChild(content)
```

### DialogComponent

```go
dialog := webawesome.NewDialogComponent()
dialog.SetID("settings-dialog")
dialog.SetLabel("Settings")         // Dialog title
dialog.Show()                       // Open dialog
dialog.Hide()                       // Close dialog
dialog.NoHeader = false             // Show/hide header

// Add content
content := templ.Raw("<p>Dialog content</p>")
dialog.AddChild(content)
```

### TabGroupComponent

```go
tabs := webawesome.NewTabGroupComponent()
tabs.SetID("main-tabs")
tabs.Placement = "top"              // top, bottom, start, end

// Add tabs
tabs.AddTab("tab1", "First Tab", "<p>First tab content</p>")
tabs.AddTab("tab2", "Second Tab", "<p>Second tab content</p>")

// Set active tab
tabs.SetActiveTab("tab1")
```

## Error Handling

The package provides comprehensive error handling with detailed error messages:

```go
// Validation errors
if err := component.Validate(); err != nil {
    if validationErr, ok := err.(*webawesome.ValidationError); ok {
        fmt.Printf("Field: %s, Value: %s, Message: %s\n", 
            validationErr.Field, 
            validationErr.Value, 
            validationErr.Message)
    }
}

// Render errors
if err := component.Render(ctx, writer); err != nil {
    log.Printf("Render error: %v", err)
}
```

## Best Practices

1. **Always validate components** before rendering
2. **Use constructor functions** (`NewButtonComponent()`) to ensure proper initialization
3. **Set required fields** (like `name` for form components) before rendering
4. **Use helper methods** for cleaner, more readable code
5. **Handle errors appropriately** in your application
6. **Leverage the component registry** for optimization insights
7. **Use HTMX integration** for dynamic content loading
8. **Apply themes consistently** across your application

## Examples

See `component_examples.go` for comprehensive usage examples including:
- Complete login forms
- User profile cards
- Settings dialogs
- Navigation tabs
- Data tables
- Status indicators
- Interactive elements
- Dashboard layouts

## Testing

The package includes comprehensive tests in `component_test_example.go` covering:
- Component creation and rendering
- Validation logic
- Helper methods
- HTML output verification
- Error handling

Run tests with:
```bash
go test ./internal/webawesome/...
```

## Package Structure

### Core Files

- **`webawesome.go`** - Main package file with Manager and global utilities
- **`interfaces.go`** - Core interfaces for all Web Awesome functionality
- **`types.go`** - Component type definitions and structures
- **`config.go`** - Configuration management and validation
- **`errors.go`** - Error types and handling utilities

### Component Files

- **`components.go`** - Form components (Button, Input, Select, Textarea, Checkbox, Radio)
- **`layout_components.go`** - Layout components (Card, Dialog, Drawer, TabGroup, Table, Badge, ProgressBar, Spinner)
- **`interactive_components.go`** - Interactive components (Dropdown, Menu, Tooltip)
- **`component_examples.go`** - Comprehensive usage examples
- **`component_test_example.go`** - Test suite for all components

### Feature Modules

- **`theme.go`** - Theme data models and default themes
- **`theme_provider.go`** - Theme loading, validation, and CSS generation
- **`theme_manager.go`** - Theme management utilities
- **`theme_utils.go`** - Theme utility functions
- **`registry.go`** - Component registry implementation
- **`registry_manager.go`** - Registry management utilities
- **`migration.go`** - DaisyUI to Web Awesome migration tools
- **`utils.go`** - Utility functions and helpers

### Template Integration

- **`templ_components.templ`** - Templ template definitions
- **`templ_helpers.go`** - Template helper functions

## Integration with Gothic Forge

This package integrates seamlessly with Gothic Forge's existing architecture:

- **Templ Templates**: All components generate type-safe Templ templates
- **HTMX Compatibility**: Preserves all hx-* attributes and behaviors
- **Alpine.js Integration**: Works with existing Alpine.js patterns
- **Security**: Applies Gothic Forge's security middleware and CSP policies
- **Build System**: Integrates with the existing build pipeline
- **CLI Integration**: Enhanced `gforge` commands for scaffolding

## Version

Current version: 1.0.0 (Task 6 Complete)

This package is part of Gothic Forge v10.0 and follows semantic versioning.