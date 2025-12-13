// Web Awesome + HTMX + Alpine.js Integration for Gothic Forge
// This file provides seamless integration between Web Awesome components, HTMX, and Alpine.js

// HTMX integration for Web Awesome components
document.addEventListener('DOMContentLoaded', function() {
  // Initialize Web Awesome components after HTMX DOM updates
  document.body.addEventListener('htmx:afterSwap', function(evt) {
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
  
  console.log('Gothic Forge Web Awesome + HTMX integration initialized');
});

// Alpine.js integration for Web Awesome components
document.addEventListener('alpine:init', () => {
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
  
  console.log('Gothic Forge Web Awesome + Alpine.js integration initialized');
});