package webawesome

import (
	"fmt"
	"strings"
)

// HTMXIntegration provides utilities for seamless HTMX integration with Web Awesome components
type HTMXIntegration struct{}

// NewHTMXIntegration creates a new HTMX integration helper
func NewHTMXIntegration() *HTMXIntegration {
	return &HTMXIntegration{}
}

// PreserveHTMXAttributes ensures HTMX attributes are preserved when rendering Web Awesome components
func (h *HTMXIntegration) PreserveHTMXAttributes(attrs map[string]string) map[string]string {
	preserved := make(map[string]string)
	
	// Copy all existing attributes
	for k, v := range attrs {
		preserved[k] = v
	}
	
	// Ensure HTMX attributes are properly formatted
	htmxAttrs := []string{
		"hx-get", "hx-post", "hx-put", "hx-patch", "hx-delete",
		"hx-target", "hx-swap", "hx-trigger", "hx-indicator",
		"hx-confirm", "hx-prompt", "hx-include", "hx-vals",
		"hx-headers", "hx-params", "hx-encoding", "hx-ext",
		"hx-select", "hx-select-oob", "hx-disinherit",
	}
	
	for _, attr := range htmxAttrs {
		if val, exists := preserved[attr]; exists && val != "" {
			// Ensure proper attribute formatting for Web Awesome components
			preserved[attr] = val
		}
	}
	
	return preserved
}

// AddHTMXSupport adds HTMX initialization script for Web Awesome components
func (h *HTMXIntegration) AddHTMXSupport() string {
	return `
<script>
// HTMX integration for Web Awesome components
document.addEventListener('DOMContentLoaded', function() {
  // Initialize Web Awesome components after HTMX DOM updates
  document.body.addEventListener('htmx:afterSwap', function(evt) {
    // Re-initialize Web Awesome components in the swapped content
    if (evt.detail.target) {
      const webAwesomeElements = evt.detail.target.querySelectorAll('[class*="sl-"], sl-button, sl-input, sl-select, sl-textarea, sl-checkbox, sl-radio, sl-card, sl-dialog, sl-drawer, sl-tab-group, sl-dropdown, sl-menu, sl-tooltip, sl-badge, sl-progress-bar, sl-spinner');
      webAwesomeElements.forEach(function(el) {
        // Trigger component initialization if needed
        if (el.connectedCallback && typeof el.connectedCallback === 'function') {
          el.connectedCallback();
        }
      });
    }
  });
  
  // Handle form serialization for Web Awesome form controls
  document.body.addEventListener('htmx:configRequest', function(evt) {
    const form = evt.detail.elt.closest('form');
    if (form) {
      // Collect values from Web Awesome form components
      const webAwesomeInputs = form.querySelectorAll('sl-input, sl-select, sl-textarea, sl-checkbox, sl-radio');
      webAwesomeInputs.forEach(function(input) {
        if (input.name && input.value !== undefined) {
          evt.detail.parameters[input.name] = input.value;
        }
      });
    }
  });
  
  // Focus management for Web Awesome dialogs with HTMX
  document.body.addEventListener('htmx:afterSwap', function(evt) {
    const dialogs = evt.detail.target.querySelectorAll('sl-dialog[open]');
    dialogs.forEach(function(dialog) {
      // Focus the first focusable element in the dialog
      const focusable = dialog.querySelector('sl-button, sl-input, sl-select, sl-textarea, [tabindex]:not([tabindex="-1"])');
      if (focusable) {
        setTimeout(() => focusable.focus(), 100);
      }
    });
  });
});
</script>`
}

// CreateHTMXButton creates a Web Awesome button with HTMX attributes
func (h *HTMXIntegration) CreateHTMXButton(text, hxGet, hxTarget, variant string) *ButtonComponent {
	button := NewButtonComponent()
	button.SetAttribute("hx-get", hxGet)
	if hxTarget != "" {
		button.SetAttribute("hx-target", hxTarget)
	}
	if variant != "" {
		button.Variant = variant
	}
	// Store button text in a data attribute for now
	button.SetAttribute("button-text", text)
	return button
}

// CreateHTMXForm creates a Web Awesome form with HTMX submission
func (h *HTMXIntegration) CreateHTMXForm(action, method, target string) map[string]string {
	attrs := make(map[string]string)
	
	if method == "" {
		method = "post"
	}
	
	attrs["hx-"+strings.ToLower(method)] = action
	if target != "" {
		attrs["hx-target"] = target
	}
	
	// Add form serialization support
	attrs["hx-encoding"] = "application/x-www-form-urlencoded"
	
	return attrs
}

// CreateHTMXInput creates a Web Awesome input with HTMX live search
func (h *HTMXIntegration) CreateHTMXInput(name, placeholder, hxPost, hxTarget, hxTrigger string) *InputComponent {
	input := NewInputComponent()
	input.SetName(name)
	input.SetPlaceholder(placeholder)
	
	if hxPost != "" {
		input.SetAttribute("hx-post", hxPost)
	}
	if hxTarget != "" {
		input.SetAttribute("hx-target", hxTarget)
	}
	if hxTrigger != "" {
		input.SetAttribute("hx-trigger", hxTrigger)
	} else {
		// Default to live search trigger
		input.SetAttribute("hx-trigger", "keyup changed delay:500ms")
	}
	
	return input
}

// CreateHTMXSelect creates a Web Awesome select with HTMX change handling
func (h *HTMXIntegration) CreateHTMXSelect(name, hxPost, hxTarget string) *SelectComponent {
	select_ := NewSelectComponent()
	select_.SetName(name)
	
	if hxPost != "" {
		select_.SetAttribute("hx-post", hxPost)
	}
	if hxTarget != "" {
		select_.SetAttribute("hx-target", hxTarget)
	}
	
	// Default trigger for select changes
	select_.SetAttribute("hx-trigger", "change")
	
	return select_
}

// AlpineJSIntegration provides utilities for Alpine.js integration
type AlpineJSIntegration struct{}

// NewAlpineJSIntegration creates a new Alpine.js integration helper
func NewAlpineJSIntegration() *AlpineJSIntegration {
	return &AlpineJSIntegration{}
}

// AddAlpineSupport adds Alpine.js initialization for Web Awesome components
func (a *AlpineJSIntegration) AddAlpineSupport() string {
	return `
<script>
// Alpine.js integration for Web Awesome components
document.addEventListener('alpine:init', () => {
  // Custom Alpine directive for Web Awesome component state
  Alpine.directive('wa-state', (el, { expression }, { evaluate }) => {
    const state = evaluate(expression);
    
    // Apply state to Web Awesome components
    if (el.tagName.startsWith('SL-')) {
      Object.keys(state).forEach(key => {
        if (key in el) {
          el[key] = state[key];
        }
      });
    }
  });
  
  // Magic helper for Web Awesome component access
  Alpine.magic('wa', () => {
    return {
      // Helper to get Web Awesome component by ID
      get: (id) => document.getElementById(id),
      
      // Helper to show/hide dialogs
      showDialog: (id) => {
        const dialog = document.getElementById(id);
        if (dialog && dialog.tagName === 'SL-DIALOG') {
          dialog.show();
        }
      },
      
      hideDialog: (id) => {
        const dialog = document.getElementById(id);
        if (dialog && dialog.tagName === 'SL-DIALOG') {
          dialog.hide();
        }
      },
      
      // Helper to open/close drawers
      openDrawer: (id) => {
        const drawer = document.getElementById(id);
        if (drawer && drawer.tagName === 'SL-DRAWER') {
          drawer.open = true;
        }
      },
      
      closeDrawer: (id) => {
        const drawer = document.getElementById(id);
        if (drawer && drawer.tagName === 'SL-DRAWER') {
          drawer.open = false;
        }
      }
    };
  });
});
</script>`
}

// CreateAlpineButton creates a Web Awesome button with Alpine.js integration
func (a *AlpineJSIntegration) CreateAlpineButton(text, alpineClick, variant string) *ButtonComponent {
	button := NewButtonComponent()
	button.SetAttribute("@click", alpineClick)
	if variant != "" {
		button.Variant = variant
	}
	button.SetAttribute("button-text", text)
	return button
}

// EventHandlingPrecedence defines the order of event handling
type EventHandlingPrecedence struct {
	Order []string
}

// NewEventHandlingPrecedence creates event handling precedence rules
func NewEventHandlingPrecedence() *EventHandlingPrecedence {
	return &EventHandlingPrecedence{
		Order: []string{
			"web-awesome",  // Web Awesome component events first
			"alpine",       // Alpine.js directives second
			"htmx",         // HTMX attributes third
			"native",       // Native DOM events last
		},
	}
}

// GetPrecedenceDocumentation returns documentation for event handling precedence
func (e *EventHandlingPrecedence) GetPrecedenceDocumentation() string {
	return `
# Event Handling Precedence in Gothic Forge Web Awesome Integration

## Order of Event Processing

1. **Web Awesome Component Events** - Native component events (e.g., sl-change, sl-input)
2. **Alpine.js Directives** - Alpine.js event handlers (@click, @change, etc.)
3. **HTMX Attributes** - HTMX request attributes (hx-get, hx-post, etc.)
4. **Native DOM Events** - Standard DOM event listeners

## Best Practices

- Use Web Awesome events for component-specific behavior
- Use Alpine.js for reactive UI state management
- Use HTMX for server communication and DOM updates
- Use native events only when the above don't suffice

## Example Usage

` + "```html" + `
<!-- Web Awesome component with all three integration types -->
<sl-button 
  @click="handleClick()"           <!-- Alpine.js -->
  hx-post="/api/action"            <!-- HTMX -->
  hx-target="#result"
  variant="primary">
  Submit
</sl-button>

<!-- The processing order ensures proper behavior -->
<script>
function handleClick() {
  // Alpine.js handler runs first
  console.log('Alpine handler');
  // Then HTMX processes the request
  // Web Awesome handles the visual feedback
}
</script>
` + "```" + `
`
}

// IntegrationHelper provides a unified interface for both HTMX and Alpine.js integration
type IntegrationHelper struct {
	HTMX   *HTMXIntegration
	Alpine *AlpineJSIntegration
}

// NewIntegrationHelper creates a new integration helper
func NewIntegrationHelper() *IntegrationHelper {
	return &IntegrationHelper{
		HTMX:   NewHTMXIntegration(),
		Alpine: NewAlpineJSIntegration(),
	}
}

// GenerateIntegrationScript generates the complete integration script
func (i *IntegrationHelper) GenerateIntegrationScript() string {
	return fmt.Sprintf(`
<!-- Web Awesome + HTMX + Alpine.js Integration -->
%s
%s

<script>
// Unified initialization for all integrations
document.addEventListener('DOMContentLoaded', function() {
  console.log('Gothic Forge Web Awesome integration initialized');
  
  // Ensure proper component initialization order
  setTimeout(() => {
    // Initialize any remaining Web Awesome components
    const uninitializedComponents = document.querySelectorAll('[class*="sl-"]:not([data-wa-initialized])');
    uninitializedComponents.forEach(el => {
      el.setAttribute('data-wa-initialized', 'true');
    });
  }, 100);
});
</script>
`, i.HTMX.AddHTMXSupport(), i.Alpine.AddAlpineSupport())
}