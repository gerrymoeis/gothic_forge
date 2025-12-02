package ui

import (
	"fmt"
	"os"
	"time"
)

// DemoColorizedOutput demonstrates the colorized output features
// This is useful for testing and showcasing the color functionality
func DemoColorizedOutput() {
	fmt.Println("\n=== Gothic Forge Colorized Output Demo ===")
	fmt.Println()

	// Demo basic colors
	fmt.Println("Basic Colors:")
	fmt.Printf("  %s\n", Success("✓ Success message (green)"))
	fmt.Printf("  %s\n", Warning("⚠ Warning message (yellow)"))
	fmt.Printf("  %s\n", Error("✗ Error message (red)"))
	fmt.Printf("  %s\n", Info("ℹ Info message (cyan)"))
	fmt.Println()

	// Demo text styles
	fmt.Println("Text Styles:")
	fmt.Printf("  %s\n", Bold("Bold text"))
	fmt.Printf("  %s\n", Faint("Faint/dim text"))
	fmt.Println()

	// Demo spinner with colors
	fmt.Println("Spinner Demo:")
	spinner := NewSpinner("Processing task...")
	spinner.Start()
	time.Sleep(1 * time.Second)
	spinner.Success("Task completed successfully")

	spinner = NewSpinner("Another task...")
	spinner.Start()
	time.Sleep(1 * time.Second)
	spinner.Fail("Task failed")

	spinner = NewSpinner("Warning task...")
	spinner.Start()
	time.Sleep(1 * time.Second)
	spinner.Warn("Task completed with warnings")
	fmt.Println()

	// Demo progress bar with colors
	fmt.Println("Progress Bar Demo:")
	bar := NewProgressBar(100, "Downloading files...")
	for i := 0; i <= 100; i += 10 {
		bar.Set(i)
		time.Sleep(200 * time.Millisecond)
	}
	bar.Finish()
	fmt.Println()

	// Demo deployment-like output
	fmt.Println("Deployment Simulation:")
	phases := []struct {
		name    string
		success bool
	}{
		{"Pre-flight checks", true},
		{"Provisioning database", true},
		{"Provisioning cache", true},
		{"Deploying application", true},
		{"Running health checks", true},
	}

	for _, phase := range phases {
		s := NewSpinner(phase.name + "...")
		s.Start()
		time.Sleep(500 * time.Millisecond)
		if phase.success {
			s.Success(phase.name + " completed")
		} else {
			s.Fail(phase.name + " failed")
		}
	}

	fmt.Println()
	fmt.Printf("%s Deployment completed in %s\n",
		Success("🎉"),
		Bold("2m 15s"))
	fmt.Printf("  %s Production URL: %s\n",
		Info("🌐"),
		Bold("https://your-app.leapcell.dev"))
	fmt.Println()

	// Demo with NO_COLOR
	fmt.Println("Testing NO_COLOR environment variable:")
	fmt.Printf("  Current ColorEnabled: %v\n", ColorEnabled)
	fmt.Printf("  NO_COLOR env var: %q\n", os.Getenv("NO_COLOR"))
	fmt.Println()
}
