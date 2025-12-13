# Web Awesome Integration - Gothic Forge v10

## Overview

Gothic Forge v10 now includes comprehensive Web Awesome integration, providing modern, accessible UI components alongside the existing DaisyUI system. This integration maintains Gothic Forge's philosophy of zero Node.js dependencies while adding powerful web components.

## Quick Start

### 1. Using Web Awesome Components in Templates

Web Awesome components are now available in all Gothic Forge templates. The main layout automatically includes the necessary CSS and JavaScript.

```html
<!-- Use Web Awesome components directly in your templates -->
<sl-button variant="primary">Click me!</sl-button>
<sl-input label="Name" placeholder="Enter your name"></sl-input>
<sl-card>
  <div slot="header">Card Title</div>
  <p>Card content goes here</p>
</sl-card>
```

### 2. Layout Options

Choose from multiple layout templates:

#### Standard Layout (DaisyUI + Web Awesome)
```go
@Layout("Page Title") {
  <!-- Your content with both DaisyUI and Web Awesome components -->
}
```

#### Web Awesome Only Layout
```go
@WebAwesomeLayout("Page Title") {
  <!-- Pure Web Awesome components -->
}
```

### 3. Theme System

Web Awesome themes automatically sync with DaisyUI themes:

- **Light Theme** → `sl-theme-light`
- **Dark Theme** → `sl-theme-dark`
- **Dim Theme** → `sl-theme-dark`

Themes are automatically applied and persist across page reloads.

## Available Components

### Form Components
- `<sl-button>` - Buttons with variants and states
- `<sl-input>` - Text inputs with validation
- `<sl-select>` - Dropdown selects
- `<sl-textarea>` - Multi-line text inputs
- `<sl-checkbox>` - Checkboxes with labels
- `<sl-radio>` - Radio buttons
- `<sl-radio-group>` - Radio button groups

### Layout Components
- `<sl-card>` - Cards with headers and footers
- `<sl-dialog>` - Modal dialogs
- `<sl-drawer>` - Side drawers
- `<sl-tab-group>` - Tabbed interfaces
- `<sl-divider>` - Content dividers

### Interactive Components
- `<sl-dropdown>` - Dropdown menus
- `<sl-menu>` - Context menus
- `<sl-tooltip>` - Tooltips
- `<sl-alert>` - Alert messages

### Data Display
- `<sl-badge>` - Status badges
- `<sl-progress-bar>` - Progress indicators
- `<sl-spinner>` - Loading spinners
- `<sl-icon>` - Icons from the Web Awesome icon library

## CLI Integration

### Generate Components
```bash
# Generate Web Awesome components
gforge add component Button --system=webawesome
gforge add form ContactForm --system=webawesome
gforge add page Dashboard --components=webawesome
```

### Theme Management
```bash
# Initialize custom theme
gforge theme init my-theme

# List available themes
gforge theme list

# Apply theme
gforge theme apply my-theme
```

## HTMX Integration

Web Awesome components work seamlessly with HTMX with automatic integration:

```html
<sl-button hx-get="/api/data" hx-target="#content" variant="primary">
  Load Data
</sl-button>

<sl-input hx-post="/api/search" hx-trigger="keyup changed delay:500ms" 
          hx-target="#results" placeholder="Search...">
</sl-input>
```

### Automatic Features

- **Component Re-initialization**: Web Awesome components are automatically re-initialized after HTMX DOM updates
- **Form Serialization**: Web Awesome form controls work seamlessly with HTMX form submissions
- **Focus Management**: Proper focus handling for dialogs and interactive elements after HTMX swaps

### Alpine.js Magic Helpers

Use the `$wa` magic helper for component control:

```html
<sl-button @click="$wa.showDialog('my-dialog')" variant="primary">
  Open Dialog
</sl-button>

<sl-dialog id="my-dialog" label="My Dialog">
  <p>Dialog content</p>
  <sl-button slot="footer" @click="$wa.hideDialog('my-dialog')">Close</sl-button>
</sl-dialog>
```

## Examples

### Complete Form
```html
<form class="space-y-4">
  <sl-input label="Name" name="name" required></sl-input>
  <sl-input label="Email" name="email" type="email" required></sl-input>
  <sl-select label="Country" name="country">
    <sl-option value="us">United States</sl-option>
    <sl-option value="ca">Canada</sl-option>
  </sl-select>
  <sl-textarea label="Message" name="message" rows="4"></sl-textarea>
  <sl-button type="submit" variant="primary">Submit</sl-button>
</form>
```

### Interactive Dashboard
```html
<sl-tab-group>
  <sl-tab slot="nav" panel="overview">Overview</sl-tab>
  <sl-tab slot="nav" panel="analytics">Analytics</sl-tab>
  
  <sl-tab-panel name="overview">
    <sl-card>
      <div slot="header">Statistics</div>
      <sl-progress-bar value="75" label="Progress"></sl-progress-bar>
      <sl-badge variant="success">Active</sl-badge>
    </sl-card>
  </sl-tab-panel>
  
  <sl-tab-panel name="analytics">
    <sl-alert variant="info" open>
      <sl-icon slot="icon" name="info-circle"></sl-icon>
      Analytics data will appear here.
    </sl-alert>
  </sl-tab-panel>
</sl-tab-group>
```

## Customization

### CSS Custom Properties

Web Awesome uses CSS custom properties that can be overridden:

```css
:root {
  --sl-color-primary-600: #your-brand-color;
  --sl-border-radius-medium: 8px;
  --sl-font-sans: 'Your Font', sans-serif;
}
```

### Theme Switching

Programmatically switch themes:

```javascript
// Switch to dark theme
function setTheme(theme) {
  document.documentElement.classList.remove('sl-theme-light', 'sl-theme-dark');
  document.documentElement.classList.add(`sl-theme-${theme}`);
  localStorage.setItem('webawesome-theme', theme);
}

setTheme('dark');
```

## Migration from DaisyUI

Web Awesome components can be used alongside DaisyUI components during migration:

| DaisyUI | Web Awesome | Notes |
|---------|-------------|-------|
| `btn` | `<sl-button>` | More variants and states |
| `input` | `<sl-input>` | Built-in validation |
| `select` | `<sl-select>` | Better accessibility |
| `card` | `<sl-card>` | Slot-based content |
| `modal` | `<sl-dialog>` | Better focus management |

## Performance

Gothic Forge includes automatic performance optimizations:

- **CDN Delivery**: Assets loaded from jsDelivr CDN with optimal caching
- **Smart Loading**: Critical components load immediately, interactive components load on-demand
- **Asset Preloading**: Critical CSS and JavaScript files are preloaded
- **Resource Deduplication**: Automatic removal of duplicate dependencies
- **Performance Monitoring**: Built-in performance tracking (development mode)

### Optimized Layouts

Use the optimized layout for maximum performance:

```go
@OptimizedLayout("Page Title") {
  <!-- Your content with performance optimizations -->
}
```

### Performance Monitoring

In development, performance metrics are automatically logged:

```javascript
// Access performance metrics
console.log(window.webAwesomeMetrics);
// Output: { totalLoad: 45.2, componentInit: 12.1, themeSwitch: 2.3 }
```

## Browser Support

Web Awesome supports all modern browsers:
- Chrome 63+
- Firefox 67+
- Safari 12.1+
- Edge 79+

## Troubleshooting

### Components Not Styling Correctly
Ensure the Web Awesome CSS is loaded after other stylesheets:
```html
<link rel="stylesheet" href="/static/styles/output.css"/>
<link rel="stylesheet" href="/static/styles/webawesome-overrides.css"/>
```

### Icons Not Loading
Verify the base path is set correctly:
```javascript
if (window.SlSetBasePath) {
  window.SlSetBasePath('https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.15.1/cdn/');
}
```

### Theme Not Applying
Check that theme classes are applied to the HTML element:
```javascript
document.documentElement.classList.add('sl-theme-light');
```

## Resources

- [Web Awesome Documentation](https://shoelace.style/)
- [Gothic Forge Documentation](https://gothicforge.dev/)
- [Component Examples](/templates/webawesome_demo.templ)
- [CLI Reference](https://gothicforge.dev/cli/)

## Support

For issues and questions:
- Check the [Gothic Forge GitHub Issues](https://github.com/gothicforge/issues)
- Review the [Web Awesome Documentation](https://shoelace.style/getting-started/overview)
- Join the [Gothic Forge Community](https://discord.gg/gothicforge)