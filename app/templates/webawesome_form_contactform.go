package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
    "gothicforge3/internal/webawesome"
)

// ContactformFormData represents the form data structure
type ContactformFormData struct {
    Name        string
    Email       string
    Message     string
    // Add more fields as needed
}

// WebAwesomeContactformForm creates a Web Awesome form component
func WebAwesomeContactformForm(action string, data *ContactformFormData, errors map[string]string) templ.Component {
    if data == nil {
        data = &ContactformFormData{}
    }
    if errors == nil {
        errors = make(map[string]string)
    }
    
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"space-y-6\">")
        
        // Name field
        nameInput := webawesome.NewInputComponent()
        nameInput.SetAttribute("name", "name")
        nameInput.SetAttribute("label", "Name")
        nameInput.SetAttribute("value", data.Name)
        nameInput.SetAttribute("required", "true")
        if err, exists := errors["name"]; exists {
            nameInput.SetAttribute("help-text", err)
            nameInput.SetClass("sl-input--invalid")
        }
        nameHTML, _ := nameInput.RenderHTML()
        _, _ = io.WriteString(w, nameHTML)
        
        // Email field
        emailInput := webawesome.NewInputComponent()
        emailInput.SetAttribute("name", "email")
        emailInput.SetAttribute("type", "email")
        emailInput.SetAttribute("label", "Email")
        emailInput.SetAttribute("value", data.Email)
        emailInput.SetAttribute("required", "true")
        if err, exists := errors["email"]; exists {
            emailInput.SetAttribute("help-text", err)
            emailInput.SetClass("sl-input--invalid")
        }
        emailHTML, _ := emailInput.RenderHTML()
        _, _ = io.WriteString(w, emailHTML)
        
        // Message field
        messageTextarea := webawesome.NewTextareaComponent()
        messageTextarea.SetAttribute("name", "message")
        messageTextarea.SetAttribute("label", "Message")
        messageTextarea.SetAttribute("rows", "4")
        messageTextarea.SetContent(data.Message)
        if err, exists := errors["message"]; exists {
            messageTextarea.SetAttribute("help-text", err)
            messageTextarea.SetClass("sl-textarea--invalid")
        }
        messageHTML, _ := messageTextarea.RenderHTML()
        _, _ = io.WriteString(w, messageHTML)
        
        // Submit button
        submitBtn := webawesome.NewButtonComponent()
        submitBtn.SetAttribute("type", "submit")
        submitBtn.SetAttribute("variant", "primary")
        submitBtn.SetContent("Submit")
        submitHTML, _ := submitBtn.RenderHTML()
        _, _ = io.WriteString(w, submitHTML)
        
        _, _ = io.WriteString(w, "</form>")
        return nil
    })
}

// ContactformFormExample shows usage examples for the Contactform form
func ContactformFormExample() templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<div class=\"space-y-6\">")
        _, _ = io.WriteString(w, "<h3 class=\"text-lg font-semibold\">Contactform Form Examples</h3>")
        
        // Default form
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Default Form</h4>")
        _ = WebAwesomeContactformForm("/submit", nil, nil).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Form with data
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Form with Data</h4>")
        sampleData := &ContactformFormData{
            Name:    "John Doe",
            Email:   "john@example.com",
            Message: "Hello, this is a sample message.",
        }
        _ = WebAwesomeContactformForm("/submit", sampleData, nil).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Form with errors
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Form with Validation Errors</h4>")
        errors := map[string]string{
            "name":  "Name is required",
            "email": "Please enter a valid email address",
        }
        _ = WebAwesomeContactformForm("/submit", nil, errors).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        _, _ = io.WriteString(w, "</div>")
        return nil
    })
}
