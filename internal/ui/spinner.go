package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Spinner provides a simple terminal spinner for long-running operations
type Spinner struct {
	message   string
	frames    []string
	interval  time.Duration
	writer    io.Writer
	active    bool
	quiet     bool // Suppress spinner animation in quiet mode
	mu        sync.Mutex
	stopChan  chan struct{}
	doneChan  chan struct{}
	lastFrame int
}

// Common spinner character sets
var (
	// CharSets contains various spinner animation styles
	CharSets = [][]string{
		{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, // 0: dots
		{"◐", "◓", "◑", "◒"}, // 1: circle
		{"▁", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃"},           // 2: bars
		{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},                               // 3: braille
		{"◴", "◷", "◶", "◵"},                                                   // 4: square
		{"◰", "◳", "◲", "◱"},                                                   // 5: box
		{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"},                               // 6: dots2
		{"|", "/", "-", "\\"},                                                  // 7: line
		{".", "o", "O", "@", "*"},                                              // 8: grow
		{"▖", "▘", "▝", "▗"},                                                   // 9: blocks
		{"◜", "◠", "◝", "◞", "◡", "◟"},                                         // 10: arc
		{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"},                     // 11: dots3
		{"⠄", "⠆", "⠇", "⠋", "⠙", "⠸", "⠰", "⠠", "⠰", "⠸", "⠙", "⠋", "⠇", "⠆"}, // 12: dots4
		{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},                     // 13: dots5
		{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},                     // 14: dots6 (default)
	}
)

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   CharSets[14], // Use dots6 as default (matches design doc)
		interval: 100 * time.Millisecond,
		writer:   os.Stdout,
		quiet:    os.Getenv("NO_COLOR") != "" || !isTerminal(),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// NewSpinnerWithCharset creates a spinner with a specific character set
func NewSpinnerWithCharset(message string, charsetIndex int) *Spinner {
	if charsetIndex < 0 || charsetIndex >= len(CharSets) {
		charsetIndex = 14 // default
	}
	return &Spinner{
		message:  message,
		frames:   CharSets[charsetIndex],
		interval: 100 * time.Millisecond,
		writer:   os.Stdout,
		quiet:    os.Getenv("NO_COLOR") != "" || !isTerminal(),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

// SetWriter sets the output writer for the spinner
func (s *Spinner) SetWriter(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writer = w
}

// SetInterval sets the animation interval
func (s *Spinner) SetInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = interval
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	
	// In quiet mode, just print the message once without animation
	if s.quiet {
		fmt.Fprintf(s.writer, "%s\n", s.message)
		s.mu.Unlock()
		return
	}
	
	s.active = true
	s.mu.Unlock()

	go s.animate()
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	defer close(s.doneChan)

	for {
		select {
		case <-s.stopChan:
			s.clearLine()
			return
		case <-ticker.C:
			s.mu.Lock()
			frame := s.frames[s.lastFrame%len(s.frames)]
			s.lastFrame++
			message := s.message
			writer := s.writer
			s.mu.Unlock()

			s.clearLine()
			fmt.Fprintf(writer, "%s %s", frame, message)
		}
	}
}

// clearLine clears the current line in the terminal
func (s *Spinner) clearLine() {
	fmt.Fprint(s.writer, "\r\033[K")
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stopChan)
	<-s.doneChan

	// Reset channels for potential reuse
	s.stopChan = make(chan struct{})
	s.doneChan = make(chan struct{})
}

// Success stops the spinner and displays a success message in green
func (s *Spinner) Success(message string) {
	s.Stop()
	if s.quiet {
		fmt.Fprintf(s.writer, "[OK] %s\n", message)
	} else {
		fmt.Fprintf(s.writer, "%s %s\n", Success("✓"), message)
	}
}

// Fail stops the spinner and displays a failure message in red
func (s *Spinner) Fail(message string) {
	s.Stop()
	if s.quiet {
		fmt.Fprintf(s.writer, "[FAIL] %s\n", message)
	} else {
		fmt.Fprintf(s.writer, "%s %s\n", Error("✗"), message)
	}
}

// Warn stops the spinner and displays a warning message in yellow
func (s *Spinner) Warn(message string) {
	s.Stop()
	if s.quiet {
		fmt.Fprintf(s.writer, "[WARN] %s\n", message)
	} else {
		fmt.Fprintf(s.writer, "%s %s\n", Warning("⚠"), message)
	}
}

// UpdateMessage updates the spinner message while it's running
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}
