package webawesome

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

// ComponentExamples provides usage examples for all Web Awesome components
type ComponentExamples struct{}

// NewComponentExamples creates a new ComponentExamples instance
func NewComponentExamples() *ComponentExamples {
	return &ComponentExamples{}
}

// CreateLoginForm demonstrates form components usage
func (ce *ComponentExamples) CreateLoginForm() string {
	var html strings.Builder
	ctx := context.Background()

	// Create form container
	html.WriteString(`<form class="login-form">`)

	// Email input
	emailInput := NewInputComponent()
	emailInput.SetType("email")
	emailInput.SetName("email")
	emailInput.SetPlaceholder("Enter your email")
	emailInput.SetRequired(true)
	emailInput.SetID("email-input")
	emailInput.Render(ctx, &html)

	// Password input
	passwordInput := NewInputComponent()
	passwordInput.SetType("password")
	passwordInput.SetName("password")
	passwordInput.SetPlaceholder("Enter your password")
	passwordInput.SetRequired(true)
	passwordInput.SetID("password-input")
	passwordInput.Render(ctx, &html)

	// Remember me checkbox
	rememberCheckbox := NewCheckboxComponent()
	rememberCheckbox.SetName("remember")
	rememberCheckbox.SetValue("1")
	rememberCheckbox.SetLabel("Remember me")
	rememberCheckbox.SetID("remember-checkbox")
	rememberCheckbox.Render(ctx, &html)

	// Submit button
	submitButton := NewButtonComponent()
	submitButton.SetID("submit-button")
	submitButton.Type = "submit"
	submitButton.Variant = "primary"
	submitButton.Children = []templ.Component{templ.Raw("Sign In")}
	submitButton.Render(ctx, &html)

	html.WriteString(`</form>`)
	return html.String()
}

// CreateUserProfileCard demonstrates card and layout components
func (ce *ComponentExamples) CreateUserProfileCard() string {
	var html strings.Builder
	ctx := context.Background()

	// Create profile card
	profileCard := NewCardComponent()
	profileCard.SetID("profile-card")
	profileCard.SetHeader("User Profile")
	profileCard.SetFooter(`<div class="card-actions">
		<sl-button variant="neutral">Edit</sl-button>
		<sl-button variant="primary">Save</sl-button>
	</div>`)

	// Add profile content as children
	profileContent := templ.Raw(`
		<div class="profile-content">
			<div class="avatar">
				<img src="/avatar.jpg" alt="User Avatar" />
			</div>
			<div class="profile-info">
				<h3>John Doe</h3>
				<p>Software Developer</p>
				<p>john.doe@example.com</p>
			</div>
		</div>
	`)
	profileCard.AddChild(profileContent)

	profileCard.Render(ctx, &html)
	return html.String()
}

// CreateSettingsDialog demonstrates dialog component
func (ce *ComponentExamples) CreateSettingsDialog() string {
	var html strings.Builder
	ctx := context.Background()

	// Create settings dialog
	settingsDialog := NewDialogComponent()
	settingsDialog.SetID("settings-dialog")
	settingsDialog.SetLabel("Application Settings")

	// Add settings form content
	settingsContent := templ.Raw(`
		<div class="settings-form">
			<sl-select name="theme" label="Theme">
				<sl-option value="light">Light</sl-option>
				<sl-option value="dark">Dark</sl-option>
				<sl-option value="auto">Auto</sl-option>
			</sl-select>
			
			<sl-checkbox name="notifications">Enable notifications</sl-checkbox>
			<sl-checkbox name="auto-save">Auto-save changes</sl-checkbox>
			
			<div class="dialog-actions">
				<sl-button variant="neutral" onclick="document.getElementById('settings-dialog').hide()">Cancel</sl-button>
				<sl-button variant="primary">Save Settings</sl-button>
			</div>
		</div>
	`)
	settingsDialog.AddChild(settingsContent)

	settingsDialog.Render(ctx, &html)
	return html.String()
}

// CreateNavigationTabs demonstrates tab group component
func (ce *ComponentExamples) CreateNavigationTabs() string {
	var html strings.Builder
	ctx := context.Background()

	// Create tab group
	tabGroup := NewTabGroupComponent()
	tabGroup.SetID("nav-tabs")
	tabGroup.Placement = "top"

	// Add tabs
	tabGroup.AddTab("dashboard", "Dashboard", `
		<div class="dashboard-content">
			<h2>Dashboard</h2>
			<p>Welcome to your dashboard!</p>
		</div>
	`)

	tabGroup.AddTab("profile", "Profile", `
		<div class="profile-content">
			<h2>Profile Settings</h2>
			<p>Manage your profile information.</p>
		</div>
	`)

	tabGroup.AddTab("settings", "Settings", `
		<div class="settings-content">
			<h2>Application Settings</h2>
			<p>Configure your application preferences.</p>
		</div>
	`)

	// Set first tab as active
	tabGroup.SetActiveTab("dashboard")

	tabGroup.Render(ctx, &html)
	return html.String()
}

// CreateDataTable demonstrates table component
func (ce *ComponentExamples) CreateDataTable() string {
	var html strings.Builder
	ctx := context.Background()

	// Create data table
	dataTable := NewTableComponent()
	dataTable.SetID("users-table")
	dataTable.SetCaption("User Management")

	// Add headers
	dataTable.AddHeader("ID")
	dataTable.AddHeader("Name")
	dataTable.AddHeader("Email")
	dataTable.AddHeader("Role")
	dataTable.AddHeader("Actions")

	// Add sample data rows
	dataTable.AddRow([]string{"1", "John Doe", "john@example.com", "Admin", `<sl-button size="small">Edit</sl-button>`})
	dataTable.AddRow([]string{"2", "Jane Smith", "jane@example.com", "User", `<sl-button size="small">Edit</sl-button>`})
	dataTable.AddRow([]string{"3", "Bob Johnson", "bob@example.com", "User", `<sl-button size="small">Edit</sl-button>`})

	dataTable.Render(ctx, &html)
	return html.String()
}

// CreateStatusIndicators demonstrates badge and progress components
func (ce *ComponentExamples) CreateStatusIndicators() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<div class="status-indicators">`)

	// Status badges
	successBadge := NewBadgeComponent()
	successBadge.SetVariant("success")
	successBadge.SetContent("Active")
	successBadge.Pill = true
	successBadge.Render(ctx, &html)

	warningBadge := NewBadgeComponent()
	warningBadge.SetVariant("warning")
	warningBadge.SetContent("Pending")
	warningBadge.Pill = true
	warningBadge.Render(ctx, &html)

	dangerBadge := NewBadgeComponent()
	dangerBadge.SetVariant("danger")
	dangerBadge.SetContent("Error")
	dangerBadge.Pill = true
	dangerBadge.Render(ctx, &html)

	// Progress bars
	html.WriteString(`<div class="progress-section">`)
	
	uploadProgress := NewProgressBarComponent()
	uploadProgress.SetValue(75)
	uploadProgress.Label = "Upload Progress"
	uploadProgress.SetID("upload-progress")
	uploadProgress.Render(ctx, &html)

	loadingProgress := NewProgressBarComponent()
	loadingProgress.SetIndeterminate(true)
	loadingProgress.Label = "Loading..."
	loadingProgress.SetID("loading-progress")
	loadingProgress.Render(ctx, &html)

	html.WriteString(`</div>`)

	// Spinners
	html.WriteString(`<div class="spinner-section">`)
	
	smallSpinner := NewSpinnerComponent()
	smallSpinner.SetSize("small")
	smallSpinner.Render(ctx, &html)

	mediumSpinner := NewSpinnerComponent()
	mediumSpinner.SetSize("medium")
	mediumSpinner.Render(ctx, &html)

	largeSpinner := NewSpinnerComponent()
	largeSpinner.SetSize("large")
	largeSpinner.Render(ctx, &html)

	html.WriteString(`</div>`)
	html.WriteString(`</div>`)

	return html.String()
}

// CreateInteractiveElements demonstrates dropdown, menu, and tooltip components
func (ce *ComponentExamples) CreateInteractiveElements() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<div class="interactive-elements">`)

	// User menu dropdown
	userMenuItems := []MenuItem{
		{Label: "Profile", Value: "profile", Type: "normal"},
		{Label: "Settings", Value: "settings", Type: "normal"},
		{Label: "---", Disabled: true}, // Separator
		{Label: "Logout", Value: "logout", Type: "normal"},
	}

	userDropdown := CreateSimpleDropdown("User Menu", userMenuItems)
	userDropdown.SetID("user-dropdown")
	userDropdown.SetPlacement("bottom-end")
	userDropdown.Render(ctx, &html)

	// Help tooltip
	helpTooltip := CreateHelpTooltip("This is a helpful tooltip that provides additional information.")
	helpTooltip.SetID("help-tooltip")
	helpTooltip.Render(ctx, &html)

	// Action button with tooltip
	actionTooltip := CreateTooltipButton("Save Changes", "Save your changes to the server", "primary")
	actionTooltip.SetID("save-tooltip")
	actionTooltip.Render(ctx, &html)

	html.WriteString(`</div>`)
	return html.String()
}

// CreateCompleteForm demonstrates a comprehensive form with all form components
func (ce *ComponentExamples) CreateCompleteForm() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<form class="complete-form">`)
	html.WriteString(`<h2>User Registration</h2>`)

	// Text inputs
	firstNameInput := NewInputComponent()
	firstNameInput.SetType("text")
	firstNameInput.SetName("firstName")
	firstNameInput.SetPlaceholder("First Name")
	firstNameInput.SetRequired(true)
	firstNameInput.Render(ctx, &html)

	lastNameInput := NewInputComponent()
	lastNameInput.SetType("text")
	lastNameInput.SetName("lastName")
	lastNameInput.SetPlaceholder("Last Name")
	lastNameInput.SetRequired(true)
	lastNameInput.Render(ctx, &html)

	// Email input
	emailInput := NewInputComponent()
	emailInput.SetType("email")
	emailInput.SetName("email")
	emailInput.SetPlaceholder("Email Address")
	emailInput.SetRequired(true)
	emailInput.Render(ctx, &html)

	// Select dropdown
	countrySelect := NewSelectComponent()
	countrySelect.SetName("country")
	countrySelect.SetPlaceholder("Select Country")
	countrySelect.AddOption("us", "United States", false)
	countrySelect.AddOption("ca", "Canada", false)
	countrySelect.AddOption("uk", "United Kingdom", false)
	countrySelect.AddOption("de", "Germany", false)
	countrySelect.Render(ctx, &html)

	// Textarea
	bioTextarea := NewTextareaComponent()
	bioTextarea.SetName("bio")
	bioTextarea.SetPlaceholder("Tell us about yourself...")
	bioTextarea.SetRows(4)
	bioTextarea.Render(ctx, &html)

	// Checkboxes
	html.WriteString(`<div class="checkbox-group">`)
	html.WriteString(`<label>Interests:</label>`)

	techCheckbox := NewCheckboxComponent()
	techCheckbox.SetName("interests")
	techCheckbox.SetValue("technology")
	techCheckbox.SetLabel("Technology")
	techCheckbox.Render(ctx, &html)

	sportsCheckbox := NewCheckboxComponent()
	sportsCheckbox.SetName("interests")
	sportsCheckbox.SetValue("sports")
	sportsCheckbox.SetLabel("Sports")
	sportsCheckbox.Render(ctx, &html)

	musicCheckbox := NewCheckboxComponent()
	musicCheckbox.SetName("interests")
	musicCheckbox.SetValue("music")
	musicCheckbox.SetLabel("Music")
	musicCheckbox.Render(ctx, &html)

	html.WriteString(`</div>`)

	// Radio buttons
	html.WriteString(`<div class="radio-group">`)
	html.WriteString(`<label>Preferred Contact Method:</label>`)

	emailRadio := NewRadioComponent()
	emailRadio.SetName("contact")
	emailRadio.SetValue("email")
	emailRadio.SetLabel("Email")
	emailRadio.SetChecked(true)
	emailRadio.Render(ctx, &html)

	phoneRadio := NewRadioComponent()
	phoneRadio.SetName("contact")
	phoneRadio.SetValue("phone")
	phoneRadio.SetLabel("Phone")
	phoneRadio.Render(ctx, &html)

	smsRadio := NewRadioComponent()
	smsRadio.SetName("contact")
	smsRadio.SetValue("sms")
	smsRadio.SetLabel("SMS")
	smsRadio.Render(ctx, &html)

	html.WriteString(`</div>`)

	// Terms checkbox
	termsCheckbox := NewCheckboxComponent()
	termsCheckbox.SetName("terms")
	termsCheckbox.SetValue("accepted")
	termsCheckbox.SetLabel("I agree to the Terms of Service")
	termsCheckbox.SetRequired(true)
	termsCheckbox.Render(ctx, &html)

	// Submit button
	submitButton := NewButtonComponent()
	submitButton.Type = "submit"
	submitButton.Variant = "primary"
	submitButton.Children = []templ.Component{templ.Raw("Register")}
	submitButton.Render(ctx, &html)

	html.WriteString(`</form>`)
	return html.String()
}

// CreateDashboardLayout demonstrates a complete dashboard layout
func (ce *ComponentExamples) CreateDashboardLayout() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<div class="dashboard-layout">`)

	// Header with navigation
	html.WriteString(`<header class="dashboard-header">`)
	html.WriteString(`<h1>Dashboard</h1>`)

	// User dropdown in header
	userMenuItems := []MenuItem{
		{Label: "Profile", Value: "profile"},
		{Label: "Settings", Value: "settings"},
		{Label: "Logout", Value: "logout"},
	}
	headerDropdown := CreateSimpleDropdown("John Doe", userMenuItems)
	headerDropdown.Render(ctx, &html)

	html.WriteString(`</header>`)

	// Main content area with tabs
	mainTabs := NewTabGroupComponent()
	mainTabs.SetID("dashboard-tabs")

	// Overview tab
	mainTabs.AddTab("overview", "Overview", ce.CreateOverviewContent())

	// Users tab
	mainTabs.AddTab("users", "Users", ce.CreateDataTable())

	// Settings tab
	mainTabs.AddTab("settings", "Settings", ce.CreateSettingsContent())

	mainTabs.SetActiveTab("overview")
	mainTabs.Render(ctx, &html)

	html.WriteString(`</div>`)
	return html.String()
}

// CreateOverviewContent creates content for the overview tab
func (ce *ComponentExamples) CreateOverviewContent() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<div class="overview-content">`)

	// Stats cards
	html.WriteString(`<div class="stats-grid">`)

	// Users card
	usersCard := NewCardComponent()
	usersCard.SetHeader("Total Users")
	usersCard.AddChild(templ.Raw(`
		<div class="stat-content">
			<div class="stat-number">1,234</div>
			<div class="stat-change">+12% from last month</div>
		</div>
	`))
	usersCard.Render(ctx, &html)

	// Revenue card
	revenueCard := NewCardComponent()
	revenueCard.SetHeader("Revenue")
	revenueCard.AddChild(templ.Raw(`
		<div class="stat-content">
			<div class="stat-number">$45,678</div>
			<div class="stat-change">+8% from last month</div>
		</div>
	`))
	revenueCard.Render(ctx, &html)

	html.WriteString(`</div>`)

	// Progress indicators
	html.WriteString(`<div class="progress-section">`)
	html.WriteString(`<h3>Current Tasks</h3>`)

	taskProgress := NewProgressBarComponent()
	taskProgress.SetValue(65)
	taskProgress.Label = "Project Alpha"
	taskProgress.Render(ctx, &html)

	taskProgress2 := NewProgressBarComponent()
	taskProgress2.SetValue(30)
	taskProgress2.Label = "Project Beta"
	taskProgress2.Render(ctx, &html)

	html.WriteString(`</div>`)

	html.WriteString(`</div>`)
	return html.String()
}

// CreateSettingsContent creates content for the settings tab
func (ce *ComponentExamples) CreateSettingsContent() string {
	var html strings.Builder
	ctx := context.Background()

	html.WriteString(`<div class="settings-content">`)
	html.WriteString(`<h3>Application Settings</h3>`)

	// Theme selection
	themeSelect := NewSelectComponent()
	themeSelect.SetName("theme")
	themeSelect.SetPlaceholder("Select Theme")
	themeSelect.AddOption("light", "Light Theme", true)
	themeSelect.AddOption("dark", "Dark Theme", false)
	themeSelect.AddOption("auto", "Auto (System)", false)
	themeSelect.Render(ctx, &html)

	// Notification settings
	html.WriteString(`<div class="notification-settings">`)
	html.WriteString(`<h4>Notifications</h4>`)

	emailNotifications := NewCheckboxComponent()
	emailNotifications.SetName("notifications")
	emailNotifications.SetValue("email")
	emailNotifications.SetLabel("Email notifications")
	emailNotifications.SetChecked(true)
	emailNotifications.Render(ctx, &html)

	pushNotifications := NewCheckboxComponent()
	pushNotifications.SetName("notifications")
	pushNotifications.SetValue("push")
	pushNotifications.SetLabel("Push notifications")
	pushNotifications.Render(ctx, &html)

	html.WriteString(`</div>`)

	// Save button
	saveButton := NewButtonComponent()
	saveButton.Variant = "primary"
	saveButton.Children = []templ.Component{templ.Raw("Save Settings")}
	saveButton.Render(ctx, &html)

	html.WriteString(`</div>`)
	return html.String()
}