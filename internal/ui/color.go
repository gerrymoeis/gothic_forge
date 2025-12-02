package ui

import (
	"fmt"
	"io"
	"os"
)

// ANSI color codes
const (
	// Reset resets all attributes
	Reset = "\033[0m"

	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bold colors
	BoldBlack   = "\033[1;30m"
	BoldRed     = "\033[1;31m"
	BoldGreen   = "\033[1;32m"
	BoldYellow  = "\033[1;33m"
	BoldBlue    = "\033[1;34m"
	BoldMagenta = "\033[1;35m"
	BoldCyan    = "\033[1;36m"
	BoldWhite   = "\033[1;37m"
)

var (
	// ColorEnabled controls whether colors are enabled globally
	// Automatically disabled when output is not a terminal or NO_COLOR is set
	ColorEnabled = isColorEnabled()
)

// isColorEnabled checks if color output should be enabled
func isColorEnabled() bool {
	// Check NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
}

// Colorize wraps text with the given color code
func Colorize(text string, color string) string {
	if !ColorEnabled {
		return text
	}
	return color + text + Reset
}

// Success returns text in green (for success messages)
func Success(text string) string {
	return Colorize(text, Green)
}

// Warning returns text in yellow (for warning messages)
func Warning(text string) string {
	return Colorize(text, Yellow)
}

// Error returns text in red (for error messages)
func Error(text string) string {
	return Colorize(text, Red)
}

// Info returns text in cyan (for informational messages)
func Info(text string) string {
	return Colorize(text, Cyan)
}

// Bold returns text in bold
func Bold(text string) string {
	if !ColorEnabled {
		return text
	}
	return "\033[1m" + text + Reset
}

// Faint returns text in faint/dim style
func Faint(text string) string {
	if !ColorEnabled {
		return text
	}
	return "\033[2m" + text + Reset
}

// PrintSuccess prints a success message in green
func PrintSuccess(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, Success(format)+"\n", args...)
}

// PrintWarning prints a warning message in yellow
func PrintWarning(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, Warning(format)+"\n", args...)
}

// PrintError prints an error message in red
func PrintError(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, Error(format)+"\n", args...)
}

// PrintInfo prints an informational message in cyan
func PrintInfo(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, Info(format)+"\n", args...)
}
