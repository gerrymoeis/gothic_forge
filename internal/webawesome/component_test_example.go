package webawesome

import (
	"context"
	"strings"
	"testing"
)

// TestAllComponents demonstrates that all Web Awesome components can be created and rendered
func TestAllComponents(t *testing.T) {
	ctx := context.Background()

	t.Run("Form Components", func(t *testing.T) {
		// Test ButtonComponent
		button := NewButtonComponent()
		button.SetID("test-button")
		button.Variant = "primary"
		button.Type = "button"
		
		var buf strings.Builder
		if err := button.Render(ctx, &buf); err != nil {
			t.Errorf("ButtonComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-button") {
			t.Error("ButtonComponent should render sl-button tag")
		}

		// Test InputComponent
		input := NewInputComponent()
		input.SetName("test-input")
		input.SetType("text")
		input.SetPlaceholder("Test placeholder")
		
		buf.Reset()
		if err := input.Render(ctx, &buf); err != nil {
			t.Errorf("InputComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-input") {
			t.Error("InputComponent should render sl-input tag")
		}

		// Test SelectComponent
		selectComp := NewSelectComponent()
		selectComp.SetName("test-select")
		selectComp.AddOption("value1", "Label 1", false)
		selectComp.AddOption("value2", "Label 2", true)
		
		buf.Reset()
		if err := selectComp.Render(ctx, &buf); err != nil {
			t.Errorf("SelectComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-select") {
			t.Error("SelectComponent should render sl-select tag")
		}

		// Test TextareaComponent
		textarea := NewTextareaComponent()
		textarea.SetName("test-textarea")
		textarea.SetRows(4)
		textarea.SetPlaceholder("Enter text here")
		
		buf.Reset()
		if err := textarea.Render(ctx, &buf); err != nil {
			t.Errorf("TextareaComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-textarea") {
			t.Error("TextareaComponent should render sl-textarea tag")
		}

		// Test CheckboxComponent
		checkbox := NewCheckboxComponent()
		checkbox.SetName("test-checkbox")
		checkbox.SetValue("test-value")
		checkbox.SetLabel("Test Checkbox")
		
		buf.Reset()
		if err := checkbox.Render(ctx, &buf); err != nil {
			t.Errorf("CheckboxComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-checkbox") {
			t.Error("CheckboxComponent should render sl-checkbox tag")
		}

		// Test RadioComponent
		radio := NewRadioComponent()
		radio.SetName("test-radio")
		radio.SetValue("test-value")
		radio.SetLabel("Test Radio")
		
		buf.Reset()
		if err := radio.Render(ctx, &buf); err != nil {
			t.Errorf("RadioComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-radio") {
			t.Error("RadioComponent should render sl-radio tag")
		}
	})

	t.Run("Layout Components", func(t *testing.T) {
		// Test CardComponent
		card := NewCardComponent()
		card.SetID("test-card")
		card.SetHeader("Test Header")
		card.SetFooter("Test Footer")
		
		var buf strings.Builder
		if err := card.Render(ctx, &buf); err != nil {
			t.Errorf("CardComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-card") {
			t.Error("CardComponent should render sl-card tag")
		}

		// Test DialogComponent
		dialog := NewDialogComponent()
		dialog.SetID("test-dialog")
		dialog.SetLabel("Test Dialog")
		
		buf.Reset()
		if err := dialog.Render(ctx, &buf); err != nil {
			t.Errorf("DialogComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-dialog") {
			t.Error("DialogComponent should render sl-dialog tag")
		}

		// Test DrawerComponent
		drawer := NewDrawerComponent()
		drawer.SetID("test-drawer")
		drawer.SetPlacement("end")
		
		buf.Reset()
		if err := drawer.Render(ctx, &buf); err != nil {
			t.Errorf("DrawerComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-drawer") {
			t.Error("DrawerComponent should render sl-drawer tag")
		}

		// Test TabGroupComponent
		tabGroup := NewTabGroupComponent()
		tabGroup.SetID("test-tabs")
		tabGroup.AddTab("tab1", "Tab 1", "Content 1")
		tabGroup.AddTab("tab2", "Tab 2", "Content 2")
		tabGroup.SetActiveTab("tab1")
		
		buf.Reset()
		if err := tabGroup.Render(ctx, &buf); err != nil {
			t.Errorf("TabGroupComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-tab-group") {
			t.Error("TabGroupComponent should render sl-tab-group tag")
		}

		// Test TableComponent
		table := NewTableComponent()
		table.SetID("test-table")
		table.AddHeader("Column 1")
		table.AddHeader("Column 2")
		table.AddRow([]string{"Row 1 Col 1", "Row 1 Col 2"})
		table.AddRow([]string{"Row 2 Col 1", "Row 2 Col 2"})
		
		buf.Reset()
		if err := table.Render(ctx, &buf); err != nil {
			t.Errorf("TableComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "<table") {
			t.Error("TableComponent should render table tag")
		}
	})

	t.Run("Data Display Components", func(t *testing.T) {
		// Test BadgeComponent
		badge := NewBadgeComponent()
		badge.SetID("test-badge")
		badge.SetVariant("success")
		badge.SetContent("Test Badge")
		badge.Pill = true
		
		var buf strings.Builder
		if err := badge.Render(ctx, &buf); err != nil {
			t.Errorf("BadgeComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-badge") {
			t.Error("BadgeComponent should render sl-badge tag")
		}

		// Test ProgressBarComponent
		progress := NewProgressBarComponent()
		progress.SetID("test-progress")
		progress.SetValue(75)
		progress.Label = "Test Progress"
		
		buf.Reset()
		if err := progress.Render(ctx, &buf); err != nil {
			t.Errorf("ProgressBarComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-progress-bar") {
			t.Error("ProgressBarComponent should render sl-progress-bar tag")
		}

		// Test SpinnerComponent
		spinner := NewSpinnerComponent()
		spinner.SetID("test-spinner")
		spinner.SetSize("medium")
		
		buf.Reset()
		if err := spinner.Render(ctx, &buf); err != nil {
			t.Errorf("SpinnerComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-spinner") {
			t.Error("SpinnerComponent should render sl-spinner tag")
		}
	})

	t.Run("Interactive Components", func(t *testing.T) {
		// Test DropdownComponent
		dropdown := NewDropdownComponent()
		dropdown.SetID("test-dropdown")
		dropdown.SetPlacement("bottom")
		
		var buf strings.Builder
		if err := dropdown.Render(ctx, &buf); err != nil {
			t.Errorf("DropdownComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-dropdown") {
			t.Error("DropdownComponent should render sl-dropdown tag")
		}

		// Test MenuComponent
		menu := NewMenuComponent()
		menu.SetID("test-menu")
		menu.AddItem("Item 1", "value1")
		menu.AddItem("Item 2", "value2")
		menu.AddCheckboxItem("Checkbox Item", "checkbox1", true)
		
		buf.Reset()
		if err := menu.Render(ctx, &buf); err != nil {
			t.Errorf("MenuComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-menu") {
			t.Error("MenuComponent should render sl-menu tag")
		}

		// Test TooltipComponent
		tooltip := NewTooltipComponent()
		tooltip.SetID("test-tooltip")
		tooltip.SetContent("Test tooltip content")
		tooltip.SetPlacement("top")
		tooltip.SetTrigger("hover")
		
		buf.Reset()
		if err := tooltip.Render(ctx, &buf); err != nil {
			t.Errorf("TooltipComponent render failed: %v", err)
		}
		if !strings.Contains(buf.String(), "sl-tooltip") {
			t.Error("TooltipComponent should render sl-tooltip tag")
		}
	})
}

// TestComponentValidation tests validation for all components
func TestComponentValidation(t *testing.T) {
	t.Run("Input Validation", func(t *testing.T) {
		// Test invalid input type
		input := NewInputComponent()
		input.SetType("invalid-type")
		input.SetName("test")
		
		if err := input.Validate(); err == nil {
			t.Error("InputComponent should validate input type")
		}

		// Test missing name
		input2 := NewInputComponent()
		input2.SetType("text")
		
		if err := input2.Validate(); err == nil {
			t.Error("InputComponent should require name")
		}
	})

	t.Run("Select Validation", func(t *testing.T) {
		// Test missing name
		selectComp := NewSelectComponent()
		selectComp.AddOption("value1", "Label 1", false)
		
		if err := selectComp.Validate(); err == nil {
			t.Error("SelectComponent should require name")
		}

		// Test no options
		selectComp2 := NewSelectComponent()
		selectComp2.SetName("test")
		
		if err := selectComp2.Validate(); err == nil {
			t.Error("SelectComponent should require at least one option")
		}

		// Test duplicate option values
		selectComp3 := NewSelectComponent()
		selectComp3.SetName("test")
		selectComp3.AddOption("value1", "Label 1", false)
		selectComp3.AddOption("value1", "Label 2", false) // Duplicate value
		
		if err := selectComp3.Validate(); err == nil {
			t.Error("SelectComponent should not allow duplicate option values")
		}
	})

	t.Run("Button Validation", func(t *testing.T) {
		// Test invalid variant
		button := NewButtonComponent()
		button.Variant = "invalid-variant"
		
		if err := button.Validate(); err == nil {
			t.Error("ButtonComponent should validate variant")
		}

		// Test invalid size
		button2 := NewButtonComponent()
		button2.Size = "invalid-size"
		
		if err := button2.Validate(); err == nil {
			t.Error("ButtonComponent should validate size")
		}
	})

	t.Run("Progress Validation", func(t *testing.T) {
		// Test invalid progress value
		progress := NewProgressBarComponent()
		progress.SetValue(150) // Invalid: > 100
		
		if err := progress.Validate(); err == nil {
			t.Error("ProgressBarComponent should validate value range")
		}

		progress2 := NewProgressBarComponent()
		progress2.SetValue(-10) // Invalid: < 0
		
		if err := progress2.Validate(); err == nil {
			t.Error("ProgressBarComponent should validate value range")
		}
	})

	t.Run("TabGroup Validation", func(t *testing.T) {
		// Test no tabs
		tabGroup := NewTabGroupComponent()
		
		if err := tabGroup.Validate(); err == nil {
			t.Error("TabGroupComponent should require at least one tab")
		}

		// Test duplicate panel names
		tabGroup2 := NewTabGroupComponent()
		tabGroup2.AddTab("panel1", "Tab 1", "Content 1")
		tabGroup2.AddTab("panel1", "Tab 2", "Content 2") // Duplicate panel
		
		if err := tabGroup2.Validate(); err == nil {
			t.Error("TabGroupComponent should not allow duplicate panel names")
		}

		// Test multiple active tabs
		tabGroup3 := NewTabGroupComponent()
		tabGroup3.AddTab("panel1", "Tab 1", "Content 1")
		tabGroup3.AddTab("panel2", "Tab 2", "Content 2")
		tabGroup3.Tabs[0].Active = true
		tabGroup3.Tabs[1].Active = true // Multiple active tabs
		
		if err := tabGroup3.Validate(); err == nil {
			t.Error("TabGroupComponent should not allow multiple active tabs")
		}
	})
}

// TestComponentHelpers tests helper methods for components
func TestComponentHelpers(t *testing.T) {
	t.Run("Button Helpers", func(t *testing.T) {
		button := NewButtonComponent()
		button.SetID("test-button")
		button.SetClass("custom-class")
		button.SetAttribute("data-test", "value")
		
		attrs := button.GetAttributes()
		if attrs["id"] != "test-button" {
			t.Error("Button should have correct ID")
		}
		if attrs["class"] != "custom-class" {
			t.Error("Button should have correct class")
		}
	})

	t.Run("Input Helpers", func(t *testing.T) {
		input := NewInputComponent()
		input.SetName("test-input")
		input.SetType("email")
		input.SetPlaceholder("Enter email")
		input.SetRequired(true)
		input.SetDisabled(false)
		
		if input.Name != "test-input" {
			t.Error("Input should have correct name")
		}
		if input.Type != "email" {
			t.Error("Input should have correct type")
		}
		if !input.Required {
			t.Error("Input should be required")
		}
	})

	t.Run("Select Helpers", func(t *testing.T) {
		selectComp := NewSelectComponent()
		selectComp.SetName("test-select")
		selectComp.AddOption("value1", "Label 1", false)
		selectComp.AddOption("value2", "Label 2", true)
		selectComp.SetMultiple(true)
		
		if len(selectComp.Options) != 2 {
			t.Error("Select should have 2 options")
		}
		if !selectComp.Multiple {
			t.Error("Select should be multiple")
		}
		if !selectComp.Options[1].Selected {
			t.Error("Second option should be selected")
		}
	})

	t.Run("Card Helpers", func(t *testing.T) {
		card := NewCardComponent()
		card.SetHeader("Test Header")
		card.SetFooter("Test Footer")
		
		if card.Header != "Test Header" {
			t.Error("Card should have correct header")
		}
		if card.Footer != "Test Footer" {
			t.Error("Card should have correct footer")
		}
	})

	t.Run("TabGroup Helpers", func(t *testing.T) {
		tabGroup := NewTabGroupComponent()
		tabGroup.AddTab("tab1", "Tab 1", "Content 1")
		tabGroup.AddTab("tab2", "Tab 2", "Content 2")
		tabGroup.SetActiveTab("tab2")
		
		if len(tabGroup.Tabs) != 2 {
			t.Error("TabGroup should have 2 tabs")
		}
		if tabGroup.Tabs[0].Active {
			t.Error("First tab should not be active")
		}
		if !tabGroup.Tabs[1].Active {
			t.Error("Second tab should be active")
		}
	})
}

// TestRenderHTML tests HTML rendering for components
func TestRenderHTML(t *testing.T) {
	t.Run("Button RenderHTML", func(t *testing.T) {
		button := NewButtonComponent()
		button.SetID("test-button")
		button.Variant = "primary"
		
		html, err := button.RenderHTML()
		if err != nil {
			t.Errorf("Button RenderHTML failed: %v", err)
		}
		if !strings.Contains(html, "sl-button") {
			t.Error("Button HTML should contain sl-button tag")
		}
		if !strings.Contains(html, `id="test-button"`) {
			t.Error("Button HTML should contain ID attribute")
		}
	})

	t.Run("Input RenderHTML", func(t *testing.T) {
		input := NewInputComponent()
		input.SetName("test-input")
		input.SetType("text")
		
		html, err := input.RenderHTML()
		if err != nil {
			t.Errorf("Input RenderHTML failed: %v", err)
		}
		if !strings.Contains(html, "sl-input") {
			t.Error("Input HTML should contain sl-input tag")
		}
		if !strings.Contains(html, `name="test-input"`) {
			t.Error("Input HTML should contain name attribute")
		}
	})

	t.Run("Card RenderHTML", func(t *testing.T) {
		card := NewCardComponent()
		card.SetID("test-card")
		card.SetHeader("Test Header")
		
		html, err := card.RenderHTML()
		if err != nil {
			t.Errorf("Card RenderHTML failed: %v", err)
		}
		if !strings.Contains(html, "sl-card") {
			t.Error("Card HTML should contain sl-card tag")
		}
		if !strings.Contains(html, "Test Header") {
			t.Error("Card HTML should contain header content")
		}
	})
}