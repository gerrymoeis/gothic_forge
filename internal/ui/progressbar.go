package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// ProgressBar provides a terminal progress bar for tracking operation progress
type ProgressBar struct {
	total       int
	current     int
	description string
	width       int
	writer      io.Writer
	quiet       bool // Suppress progress bar animation in quiet mode
	mu          sync.Mutex
	completed   bool
}

// ProgressBarTheme defines the visual appearance of the progress bar
type ProgressBarTheme struct {
	Saucer        string // Filled portion character
	SaucerPadding string // Empty portion character
	BarStart      string // Left bracket
	BarEnd        string // Right bracket
}

// DefaultTheme is the default progress bar theme
var DefaultTheme = ProgressBarTheme{
	Saucer:        "█",
	SaucerPadding: "░",
	BarStart:      "[",
	BarEnd:        "]",
}

// NewProgressBar creates a new progress bar with the given total and description
func NewProgressBar(total int, description string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		description: description,
		width:       40, // Default width of the bar
		writer:      os.Stdout,
		quiet:       os.Getenv("NO_COLOR") != "" || !isProgressBarTerminal(),
		completed:   false,
	}
}

// isProgressBarTerminal checks if stdout is a terminal
func isProgressBarTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

// SetWriter sets the output writer for the progress bar
func (p *ProgressBar) SetWriter(w io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer = w
}

// SetWidth sets the width of the progress bar
func (p *ProgressBar) SetWidth(width int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if width > 0 {
		p.width = width
	}
}

// Add increments the progress by n units
func (p *ProgressBar) Add(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.completed {
		return
	}

	p.current += n
	if p.current > p.total {
		p.current = p.total
	}

	// In quiet mode, only print at milestones (25%, 50%, 75%, 100%)
	if p.quiet {
		percentage := 0
		if p.total > 0 {
			percentage = (p.current * 100) / p.total
		}
		// Print at 25% intervals
		if percentage%25 == 0 && percentage > 0 {
			fmt.Fprintf(p.writer, "[%d%%] %s\n", percentage, p.description)
		}
	} else {
		p.render()
	}
}

// Set sets the current progress to n
func (p *ProgressBar) Set(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.completed {
		return
	}

	p.current = n
	if p.current > p.total {
		p.current = p.total
	}

	// In quiet mode, only print at milestones (25%, 50%, 75%, 100%)
	if p.quiet {
		percentage := 0
		if p.total > 0 {
			percentage = (p.current * 100) / p.total
		}
		// Print at 25% intervals
		if percentage%25 == 0 && percentage > 0 {
			fmt.Fprintf(p.writer, "[%d%%] %s\n", percentage, p.description)
		}
	} else {
		p.render()
	}
}

// UpdateDescription updates the progress bar description
func (p *ProgressBar) UpdateDescription(description string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.description = description
	if !p.completed {
		p.render()
	}
}

// Finish completes the progress bar and moves to a new line
func (p *ProgressBar) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.completed {
		return
	}

	p.current = p.total
	p.completed = true
	
	if p.quiet {
		fmt.Fprintf(p.writer, "[100%%] %s\n", p.description)
	} else {
		p.render()
		fmt.Fprintln(p.writer) // Move to new line
	}
}

// Clear clears the progress bar from the terminal
func (p *ProgressBar) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprint(p.writer, "\r\033[K")
}

// render draws the progress bar (must be called with lock held)
func (p *ProgressBar) render() {
	// Calculate percentage
	percentage := 0
	if p.total > 0 {
		percentage = (p.current * 100) / p.total
	}

	// Calculate filled width
	filledWidth := 0
	if p.total > 0 {
		filledWidth = (p.current * p.width) / p.total
	}

	// Build the bar
	theme := DefaultTheme
	bar := strings.Builder{}
	bar.WriteString(theme.BarStart)

	// Add filled portion (in green)
	filledPortion := strings.Repeat(theme.Saucer, filledWidth)
	bar.WriteString(Success(filledPortion))

	// Add empty portion
	for i := filledWidth; i < p.width; i++ {
		bar.WriteString(theme.SaucerPadding)
	}

	bar.WriteString(theme.BarEnd)

	// Colorize percentage based on completion
	var percentageStr string
	if percentage == 100 {
		percentageStr = Success(fmt.Sprintf("%3d%%", percentage))
	} else if percentage >= 50 {
		percentageStr = Info(fmt.Sprintf("%3d%%", percentage))
	} else {
		percentageStr = fmt.Sprintf("%3d%%", percentage)
	}

	// Clear line and write progress bar
	fmt.Fprintf(p.writer, "\r\033[K%s %s %s", bar.String(), percentageStr, p.description)
}

// Current returns the current progress value
func (p *ProgressBar) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// Total returns the total progress value
func (p *ProgressBar) Total() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.total
}

// IsComplete returns whether the progress bar has completed
func (p *ProgressBar) IsComplete() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.completed
}
