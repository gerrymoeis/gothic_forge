package webawesome

import (
	"fmt"
	"strings"
)

// HTMXExamples provides practical examples of HTMX integration with Web Awesome components
type HTMXExamples struct{}

// NewHTMXExamples creates a new HTMX examples helper
func NewHTMXExamples() *HTMXExamples {
	return &HTMXExamples{}
}

// LiveSearchExample demonstrates HTMX live search with Web Awesome input
func (h *HTMXExamples) LiveSearchExample() string {
	return `
<!-- Live Search with Web Awesome Input -->
<div class="search-container">
  <sl-input 
    name="search" 
    placeholder="Search users..." 
    hx-post="/api/search" 
    hx-target="#search-results" 
    hx-trigger="keyup changed delay:500ms"
    hx-indicator="#search-spinner">
  </sl-input>
  
  <sl-spinner id="search-spinner" class="htmx-indicator"></sl-spinner>
  
  <div id="search-results" class="mt-4">
    <!-- Search results will be loaded here -->
  </div>
</div>
`
}

// FormSubmissionExample demonstrates HTMX form submission with Web Awesome components
func (h *HTMXExamples) FormSubmissionExample() string {
	return `
<!-- HTMX Form with Web Awesome Components -->
<form hx-post="/api/contact" hx-target="#form-result" hx-swap="innerHTML">
  <div class="space-y-4">
    <sl-input 
      label="Name" 
      name="name" 
      required>
    </sl-input>
    
    <sl-input 
      label="Email" 
      name="email" 
      type="email" 
      required>
    </sl-input>
    
    <sl-select 
      label="Subject" 
      name="subject" 
      required>
      <sl-option value="general">General Inquiry</sl-option>
      <sl-option value="support">Support</sl-option>
      <sl-option value="sales">Sales</sl-option>
    </sl-select>
    
    <sl-textarea 
      label="Message" 
      name="message" 
      rows="4" 
      required>
    </sl-textarea>
    
    <sl-button 
      type="submit" 
      variant="primary"
      hx-indicator="#submit-spinner">
      <sl-spinner id="submit-spinner" class="htmx-indicator" slot="prefix"></sl-spinner>
      Send Message
    </sl-button>
  </div>
</form>

<div id="form-result" class="mt-4">
  <!-- Form submission result will appear here -->
</div>
`
}

// DialogWithHTMXExample demonstrates HTMX content loading in Web Awesome dialog
func (h *HTMXExamples) DialogWithHTMXExample() string {
	return `
<!-- Dialog with HTMX Content Loading -->
<sl-button 
  variant="primary"
  onclick="document.getElementById('user-dialog').show()"
  hx-get="/api/user/123" 
  hx-target="#dialog-content"
  hx-trigger="click">
  View User Details
</sl-button>

<sl-dialog id="user-dialog" label="User Details">
  <div id="dialog-content">
    <sl-spinner></sl-spinner>
    <p>Loading user details...</p>
  </div>
  
  <sl-button slot="footer" variant="neutral" onclick="document.getElementById('user-dialog').hide()">
    Close
  </sl-button>
</sl-dialog>
`
}

// TabsWithHTMXExample demonstrates HTMX content loading in Web Awesome tabs
func (h *HTMXExamples) TabsWithHTMXExample() string {
	return `
<!-- Tabs with HTMX Content Loading -->
<sl-tab-group>
  <sl-tab slot="nav" panel="overview" active>Overview</sl-tab>
  <sl-tab slot="nav" panel="analytics">Analytics</sl-tab>
  <sl-tab slot="nav" panel="settings">Settings</sl-tab>
  
  <sl-tab-panel name="overview">
    <div hx-get="/api/dashboard/overview" hx-trigger="load">
      <sl-spinner></sl-spinner>
      <p>Loading overview...</p>
    </div>
  </sl-tab-panel>
  
  <sl-tab-panel name="analytics">
    <div hx-get="/api/dashboard/analytics" hx-trigger="load">
      <sl-spinner></sl-spinner>
      <p>Loading analytics...</p>
    </div>
  </sl-tab-panel>
  
  <sl-tab-panel name="settings">
    <div hx-get="/api/dashboard/settings" hx-trigger="load">
      <sl-spinner></sl-spinner>
      <p>Loading settings...</p>
    </div>
  </sl-tab-panel>
</sl-tab-group>
`
}

// AlpineJSExamples provides practical examples of Alpine.js integration
type AlpineJSExamples struct{}

// NewAlpineJSExamples creates a new Alpine.js examples helper
func NewAlpineJSExamples() *AlpineJSExamples {
	return &AlpineJSExamples{}
}

// ReactiveFormExample demonstrates Alpine.js reactive form with Web Awesome components
func (a *AlpineJSExamples) ReactiveFormExample() string {
	return `
<!-- Reactive Form with Alpine.js and Web Awesome -->
<div x-data="{ 
  user: { name: '', email: '', role: 'user' }, 
  isValid: false,
  checkValid() {
    this.isValid = this.user.name.length > 0 && this.user.email.includes('@');
  }
}">
  <form @submit.prevent="submitForm()">
    <div class="space-y-4">
      <sl-input 
        label="Name" 
        x-model="user.name"
        @sl-input="checkValid()"
        :required="true">
      </sl-input>
      
      <sl-input 
        label="Email" 
        type="email"
        x-model="user.email"
        @sl-input="checkValid()"
        :required="true">
      </sl-input>
      
      <sl-select 
        label="Role" 
        x-model="user.role">
        <sl-option value="user">User</sl-option>
        <sl-option value="admin">Admin</sl-option>
        <sl-option value="moderator">Moderator</sl-option>
      </sl-select>
      
      <sl-button 
        type="submit" 
        variant="primary"
        :disabled="!isValid">
        Create User
      </sl-button>
    </div>
  </form>
  
  <!-- Live Preview -->
  <sl-card class="mt-4" x-show="user.name || user.email">
    <div slot="header">Preview</div>
    <p><strong>Name:</strong> <span x-text="user.name || 'Not provided'"></span></p>
    <p><strong>Email:</strong> <span x-text="user.email || 'Not provided'"></span></p>
    <p><strong>Role:</strong> <span x-text="user.role"></span></p>
  </sl-card>
</div>
`
}

// DialogControlExample demonstrates Alpine.js dialog control
func (a *AlpineJSExamples) DialogControlExample() string {
	return `
<!-- Dialog Control with Alpine.js -->
<div x-data="{ showDialog: false }">
  <sl-button 
    variant="primary"
    @click="$wa.showDialog('my-dialog'); showDialog = true">
    Open Dialog
  </sl-button>
  
  <sl-dialog 
    id="my-dialog" 
    label="Alpine.js Controlled Dialog"
    :open="showDialog"
    @sl-hide="showDialog = false">
    
    <p>This dialog is controlled by Alpine.js state.</p>
    
    <sl-button 
      slot="footer" 
      variant="neutral"
      @click="$wa.hideDialog('my-dialog'); showDialog = false">
      Close
    </sl-button>
  </sl-dialog>
</div>
`
}

// CombinedExample demonstrates HTMX + Alpine.js + Web Awesome working together
func (h *HTMXExamples) CombinedExample() string {
	return `
<!-- Combined HTMX + Alpine.js + Web Awesome Example -->
<div x-data="{ 
  loading: false, 
  users: [], 
  selectedUser: null,
  showDetails: false 
}">
  
  <!-- Search with HTMX -->
  <sl-input 
    placeholder="Search users..." 
    hx-post="/api/users/search" 
    hx-target="#user-list" 
    hx-trigger="keyup changed delay:500ms"
    hx-indicator="#search-spinner"
    @htmx:before-request="loading = true"
    @htmx:after-request="loading = false">
  </sl-input>
  
  <sl-spinner id="search-spinner" x-show="loading"></sl-spinner>
  
  <!-- User List (populated by HTMX) -->
  <div id="user-list" class="mt-4">
    <!-- HTMX will populate this with user cards -->
  </div>
  
  <!-- User Details Dialog (controlled by Alpine.js) -->
  <sl-dialog 
    id="user-details" 
    label="User Details"
    :open="showDetails"
    @sl-hide="showDetails = false; selectedUser = null">
    
    <div x-show="selectedUser">
      <p><strong>Name:</strong> <span x-text="selectedUser?.name"></span></p>
      <p><strong>Email:</strong> <span x-text="selectedUser?.email"></span></p>
      <p><strong>Role:</strong> <span x-text="selectedUser?.role"></span></p>
    </div>
    
    <sl-button 
      slot="footer" 
      variant="neutral"
      @click="showDetails = false">
      Close
    </sl-button>
  </sl-dialog>
</div>

<script>
// Event handling precedence example
document.addEventListener('DOMContentLoaded', function() {
  // 1. Web Awesome component events fire first
  document.addEventListener('sl-change', function(e) {
    console.log('Web Awesome event:', e.detail);
  });
  
  // 2. Alpine.js handles reactive updates
  // 3. HTMX processes server requests
  // 4. Native DOM events fire last
});
</script>
`
}

// GetIntegrationDocumentation returns comprehensive integration documentation
func (h *HTMXExamples) GetIntegrationDocumentation() string {
	return fmt.Sprintf(`
# Web Awesome + HTMX + Alpine.js Integration Guide

## Overview

Gothic Forge provides seamless integration between Web Awesome components, HTMX, and Alpine.js. This integration maintains the simplicity and developer experience that Gothic Forge is known for.

## Key Features

- **Automatic Component Initialization**: Web Awesome components are automatically initialized after HTMX DOM updates
- **Form Serialization**: Web Awesome form controls work seamlessly with HTMX form submissions
- **Focus Management**: Proper focus handling for dialogs and interactive elements
- **Alpine.js Magic Helpers**: Convenient ` + "`$wa`" + ` magic helper for component control
- **Event Handling Precedence**: Clear order of event processing

## Integration Script

The integration is automatically included in Gothic Forge layouts. The script handles:

1. **HTMX Integration**:
   - Component re-initialization after DOM swaps
   - Form serialization for Web Awesome controls
   - Focus management for accessibility

2. **Alpine.js Integration**:
   - Magic helpers for component control
   - Reactive state management
   - Event handling coordination

## Examples

### Live Search
%s

### Form Submission
%s

### Dialog with HTMX
%s

### Tabs with HTMX
%s

### Combined Example
%s

## Best Practices

1. **Use Web Awesome events** for component-specific behavior
2. **Use Alpine.js** for reactive UI state management  
3. **Use HTMX** for server communication and DOM updates
4. **Follow the event precedence order** for predictable behavior

## Event Handling Order

1. Web Awesome component events (sl-change, sl-input, etc.)
2. Alpine.js directives (@click, @change, etc.)
3. HTMX attributes (hx-get, hx-post, etc.)
4. Native DOM events

This ensures that component state is updated before server requests are made.

## Magic Helpers

Alpine.js provides the ` + "`$wa`" + ` magic helper for Web Awesome component control:

` + "```javascript" + `
// Show/hide dialogs
$wa.showDialog('my-dialog')
$wa.hideDialog('my-dialog')

// Open/close drawers
$wa.openDrawer('my-drawer')
$wa.closeDrawer('my-drawer')

// Get component reference
const button = $wa.get('my-button')
` + "```" + `

## Troubleshooting

- **Components not initializing**: Ensure the integration script is loaded
- **Form values not submitting**: Check that form controls have ` + "`name`" + ` attributes
- **Focus issues**: The integration handles focus automatically for dialogs
- **Event conflicts**: Follow the event precedence order

The integration is designed to "just work" with minimal configuration, following Gothic Forge's philosophy of simplicity and developer happiness.
`,
		h.LiveSearchExample(),
		h.FormSubmissionExample(),
		h.DialogWithHTMXExample(),
		h.TabsWithHTMXExample(),
		h.CombinedExample(),
	)
}