// Package ui provides user interface components for the Gothic Forge CLI.
//
// The ui package includes terminal-based UI components such as spinners,
// progress bars, and colorized output to provide visual feedback during
// long-running operations.
//
// # Colorized Output
//
// The package provides colorized output support with automatic detection
// of terminal capabilities and respect for the NO_COLOR environment variable.
//
// Colors are automatically disabled when:
//   - The NO_COLOR environment variable is set (https://no-color.org/)
//   - Output is not a terminal (e.g., piped to a file)
//
// Example usage:
//
//	fmt.Println(ui.Success("✓ Operation completed"))
//	fmt.Println(ui.Warning("⚠ Warning: deprecated feature"))
//	fmt.Println(ui.Error("✗ Operation failed"))
//	fmt.Println(ui.Info("ℹ Additional information"))
//
// The package provides these color functions:
//   - Success() - green text for success messages
//   - Warning() - yellow text for warning messages
//   - Error() - red text for error messages
//   - Info() - cyan text for informational messages
//   - Bold() - bold text
//   - Faint() - dim/faint text
//
// Print functions are also available for convenience:
//
//	ui.PrintSuccess(os.Stdout, "Deployment successful")
//	ui.PrintWarning(os.Stdout, "Using deprecated API")
//	ui.PrintError(os.Stdout, "Connection failed")
//	ui.PrintInfo(os.Stdout, "Processing %d items", count)
//
// # Spinner
//
// The Spinner component displays an animated spinner with a message while
// operations are in progress. It can be used to indicate that work is being
// done without blocking the terminal.
//
// Example usage:
//
//	spinner := ui.NewSpinner("Provisioning database...")
//	spinner.Start()
//
//	// Perform some work
//	err := provisionDatabase()
//
//	if err != nil {
//	    spinner.Fail("Database provisioning failed")
//	    return err
//	}
//
//	spinner.Success("Database provisioned successfully")
//
// The spinner supports multiple animation styles through character sets,
// and messages can be updated dynamically while the spinner is running:
//
//	spinner := ui.NewSpinner("Step 1: Initializing...")
//	spinner.Start()
//
//	// Update message as work progresses
//	spinner.UpdateMessage("Step 2: Processing...")
//	// ... more work ...
//	spinner.UpdateMessage("Step 3: Finalizing...")
//
//	spinner.Success("All steps completed")
//
// # Character Sets
//
// The package provides multiple pre-defined character sets for different
// spinner animations. Use NewSpinnerWithCharset to select a specific style:
//
//	spinner := ui.NewSpinnerWithCharset("Loading...", 1) // Use circle animation
//
// Available character sets include dots, circles, bars, braille patterns,
// and more. See the CharSets variable for all available options.
//
// # ProgressBar
//
// The ProgressBar component displays a visual progress bar with percentage
// tracking for operations with quantifiable progress. It provides real-time
// feedback on completion status.
//
// Example usage:
//
//	bar := ui.NewProgressBar(100, "Downloading files...")
//
//	for i := 0; i < 100; i++ {
//	    // Perform work
//	    downloadFile(i)
//	    bar.Add(1)
//	}
//
//	bar.Finish()
//
// The progress bar displays in the format:
//
//	[████████████████████░░░░░░░░░░░░░░░░░░░░]  50% Downloading files...
//
// Progress can be updated incrementally with Add() or set directly with Set():
//
//	bar := ui.NewProgressBar(100, "Processing...")
//	bar.Set(25)  // Jump to 25%
//	bar.Add(25)  // Increment by 25% (now at 50%)
//	bar.Finish() // Complete to 100%
//
// The description can be updated dynamically as work progresses:
//
//	bar := ui.NewProgressBar(100, "Phase 1: Initializing...")
//	bar.Add(33)
//	bar.UpdateDescription("Phase 2: Processing...")
//	bar.Add(33)
//	bar.UpdateDescription("Phase 3: Finalizing...")
//	bar.Add(34)
//	bar.Finish()
//
// Both Spinner and ProgressBar are thread-safe and can be used in
// concurrent operations. They support custom output writers for testing
// and can be configured with various display options.
package ui
