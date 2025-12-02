# UI Package

The `ui` package provides terminal-based user interface components for the Gothic Forge CLI.

## Components

### Spinner

A terminal spinner component for displaying progress during long-running operations.

### ProgressBar

A terminal progress bar component for tracking quantifiable progress with visual feedback.

#### Features

- Multiple animation styles (15 different character sets)
- Customizable messages
- Success/Fail completion states
- Thread-safe message updates
- Configurable animation speed
- Custom output writers for testing

#### Basic Usage

```go
import "gothicforge3/internal/ui"

// Create and start a spinner
spinner := ui.NewSpinner("Loading data...")
spinner.Start()

// Do some work...
time.Sleep(2 * time.Second)

// Complete with success
spinner.Success("Data loaded successfully")
```

#### Failure Handling

```go
spinner := ui.NewSpinner("Processing request...")
spinner.Start()

err := doSomething()
if err != nil {
    spinner.Fail("Request failed")  // Red ✗
    return err
}

spinner.Success("Request completed")  // Green ✓
```

#### Warning Handling

```go
spinner := ui.NewSpinner("Validating configuration...")
spinner.Start()

warnings := validateConfig()
if len(warnings) > 0 {
    spinner.Warn("Configuration has warnings")  // Yellow ⚠
    for _, w := range warnings {
        fmt.Println(ui.Warning(fmt.Sprintf("  - %s", w)))
    }
}
```

#### Dynamic Message Updates

```go
spinner := ui.NewSpinner("Step 1: Initializing...")
spinner.Start()

// Update message as work progresses
time.Sleep(500 * time.Millisecond)
spinner.UpdateMessage("Step 2: Processing...")

time.Sleep(500 * time.Millisecond)
spinner.UpdateMessage("Step 3: Finalizing...")

spinner.Success("All steps completed")
```

#### Custom Character Sets

```go
// Use a different animation style
spinner := ui.NewSpinnerWithCharset("Loading...", 1) // Circle animation
spinner.Start()
// ...
spinner.Success("Done")
```

Available character sets (index):
- 0: Dots
- 1: Circle
- 2: Bars
- 3: Braille
- 4: Square
- 5: Box
- 6: Dots2
- 7: Line
- 8: Grow
- 9: Blocks
- 10: Arc
- 11: Dots3
- 12: Dots4
- 13: Dots5
- 14: Dots6 (default)

#### Testing

For testing, you can redirect output to a buffer:

```go
import "bytes"

buf := &bytes.Buffer{}
spinner := ui.NewSpinner("Testing...")
spinner.SetWriter(buf)

spinner.Start()
// ... test code ...
spinner.Stop()

output := buf.String()
// Assert on output
```

#### Configuration

```go
spinner := ui.NewSpinner("Loading...")

// Set custom animation interval
spinner.SetInterval(50 * time.Millisecond)

// Set custom output writer
spinner.SetWriter(os.Stderr)

spinner.Start()
```

## Integration Example

Here's how the spinner integrates with Gothic Forge commands:

```go
func deployDatabase(ctx context.Context) error {
    spinner := ui.NewSpinner("Provisioning database...")
    spinner.Start()

    db, err := provisionCockroachDB(ctx)
    if err != nil {
        spinner.Fail("Database provisioning failed")
        return fmt.Errorf("failed to provision database: %w", err)
    }

    spinner.Success(fmt.Sprintf("Database provisioned: %s", db.Name))
    return nil
}
```

---

## ProgressBar

### Features

- Visual progress tracking with percentage display
- Customizable bar width
- Thread-safe operations
- Dynamic description updates
- Clean terminal output with line clearing
- Graceful completion handling

### Basic Usage

```go
import "gothicforge3/internal/ui"

// Create a progress bar
bar := ui.NewProgressBar(100, "Downloading files...")

// Update progress
for i := 0; i < 100; i++ {
    // Do some work...
    bar.Add(1)
}

// Complete the progress bar
bar.Finish()
```

### Setting Progress Directly

```go
bar := ui.NewProgressBar(100, "Processing...")

// Set progress to specific value
bar.Set(50)  // 50%

// Continue updating
bar.Set(75)  // 75%

bar.Finish()
```

### Dynamic Description Updates

```go
bar := ui.NewProgressBar(100, "Phase 1: Initializing...")

bar.Add(25)
bar.UpdateDescription("Phase 2: Processing...")

bar.Add(25)
bar.UpdateDescription("Phase 3: Finalizing...")

bar.Add(50)
bar.Finish()
```

### Custom Width

```go
bar := ui.NewProgressBar(100, "Loading...")
bar.SetWidth(60)  // Wider progress bar

bar.Add(50)
bar.Finish()
```

### Testing

For testing, redirect output to a buffer:

```go
import "bytes"

buf := &bytes.Buffer{}
bar := ui.NewProgressBar(100, "Testing...")
bar.SetWriter(buf)

bar.Add(50)
bar.Finish()

output := buf.String()
// Assert on output
```

### Progress Bar Format

The progress bar displays in the following format:

```
[████████████████████░░░░░░░░░░░░░░░░░░░░]  50% Downloading files...
```

Components:
- `[` and `]`: Bar boundaries
- `█`: Filled portion (progress made)
- `░`: Empty portion (remaining progress)
- `50%`: Percentage complete
- Description text

### Thread Safety

The progress bar is thread-safe and can be updated from multiple goroutines:

```go
bar := ui.NewProgressBar(1000, "Parallel processing...")

var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := 0; j < 100; j++ {
            // Do work...
            bar.Add(1)
        }
    }()
}

wg.Wait()
bar.Finish()
```

### Integration Example

Here's how the progress bar integrates with Gothic Forge deployment:

```go
func provisionInfrastructure(ctx context.Context) error {
    bar := ui.NewProgressBar(100, "Provisioning database...")
    
    // Start provisioning
    bar.Set(20)
    db, err := provisionDatabase(ctx)
    if err != nil {
        return err
    }
    
    bar.Set(60)
    bar.UpdateDescription("Running migrations...")
    
    if err := runMigrations(ctx, db); err != nil {
        return err
    }
    
    bar.Set(100)
    bar.UpdateDescription("Database ready!")
    bar.Finish()
    
    return nil
}
```

---

## Design

The UI components follow the design specified in `.kiro/specs/3-minute-deployment/design.md` and provide:

**Spinner:**
- ✓ Non-blocking animation
- ✓ Clean terminal output with line clearing
- ✓ Success/Fail states with visual indicators (✓/✗)
- ✓ Thread-safe operations
- ✓ Graceful handling of double start/stop
- ✓ 100% test coverage

**ProgressBar:**
- ✓ Visual progress tracking with percentage
- ✓ Customizable bar width and appearance
- ✓ Thread-safe operations
- ✓ Dynamic description updates
- ✓ Clean terminal output with line clearing
- ✓ Graceful completion handling
- ✓ 100% test coverage

## Colorized Output

The package provides colorized terminal output with automatic detection and NO_COLOR support.

### Features

- ✓ Automatic color detection (disabled when output is not a terminal)
- ✓ Respects NO_COLOR environment variable (https://no-color.org/)
- ✓ Green for success, yellow for warnings, red for errors, cyan for info
- ✓ Thread-safe color operations
- ✓ Bold and faint text styles

### Basic Usage

```go
import "gothicforge3/internal/ui"

// Color functions
fmt.Println(ui.Success("✓ Operation completed"))
fmt.Println(ui.Warning("⚠ Warning message"))
fmt.Println(ui.Error("✗ Operation failed"))
fmt.Println(ui.Info("ℹ Information"))

// Text styles
fmt.Println(ui.Bold("Important message"))
fmt.Println(ui.Faint("Less important details"))
```

### Print Functions

Convenience functions for printing colored messages:

```go
ui.PrintSuccess(os.Stdout, "Deployment successful")
ui.PrintWarning(os.Stdout, "Using deprecated API")
ui.PrintError(os.Stdout, "Connection failed: %s", err)
ui.PrintInfo(os.Stdout, "Processing %d items", count)
```

### Color Control

Colors are automatically enabled when:
- Output is a terminal (TTY)
- NO_COLOR environment variable is not set

Colors are automatically disabled when:
- Output is redirected to a file or pipe
- NO_COLOR environment variable is set
- Running in CI/CD environments (typically)

You can manually control colors:

```go
// Disable colors
ui.ColorEnabled = false

// Re-enable colors
ui.ColorEnabled = true
```

### Colorized Components

The Spinner and ProgressBar components now use colorized output:

```go
spinner := ui.NewSpinner("Processing...")
spinner.Start()

// Green checkmark for success
spinner.Success("Task completed")

// Red X for failure
spinner.Fail("Task failed")

// Yellow warning symbol
spinner.Warn("Task completed with warnings")
```

Progress bars show green-filled portions and color-coded percentages:

```go
bar := ui.NewProgressBar(100, "Downloading...")
for i := 0; i <= 100; i += 10 {
    bar.Set(i)
    time.Sleep(100 * time.Millisecond)
}
bar.Finish()
```

### Demo

Run the color demo to see all features in action:

```go
ui.DemoColorizedOutput()
```

### Testing with Colors

For testing, you can disable colors for predictable output:

```go
// Save original state
originalColorEnabled := ui.ColorEnabled
defer func() { ui.ColorEnabled = originalColorEnabled }()

// Disable colors for testing
ui.ColorEnabled = false

// Test output without color codes
result := ui.Success("test")
// result == "test" (no ANSI codes)
```

## Future Enhancements

Potential future additions:
- Multi-line progress bar support
- Progress bar groups for parallel operations
- Estimated time remaining (ETA) display
- Transfer rate display for downloads
- Custom themes for progress bars
