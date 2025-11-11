package lib

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// progressPrinter provides real-time progress tracking for downloads and file operations
type progressPrinter struct {
	logger    *Logger
	command   string
	label     string
	total     int64
	current   int64
	width     int
	lastPrint time.Time
	done      bool
	callCount int
}

// progressSpinner provides animated spinner for indeterminate operations
type progressSpinner struct {
	message    string
	command    string
	startTime  time.Time
	enabled    bool
	stop       chan struct{}
	done       chan struct{}
	dotCounter int
}

var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// NewSpinner creates a new animated spinner with command name
func NewSpinner(command, message string) *progressSpinner {
	s := &progressSpinner{
		message:   message,
		command:   command,
		enabled:   isTerminal(os.Stdout),
		startTime: time.Now(),
	}
	if s.enabled {
		s.stop = make(chan struct{})
		s.done = make(chan struct{})
		fmt.Printf("\r[ %s ] %s %s%s", command, message, string(spinnerFrames[0]), gray(""))
		go s.loop(1)
	} else {
		fmt.Printf("[ %s ] %s...\n", command, message)
	}
	return s
}

// Success marks the spinner as successful
func (s *progressSpinner) Success(summary string) {
	if s == nil {
		return
	}
	elapsed := time.Since(s.startTime)
	s.stopSpinner()
	if s.enabled {
		durationStr := formatDuration(elapsed)
		fmt.Printf("\r[ %s ] %s %s (%s)\n", s.command, s.message, green("✓"), gray(durationStr))
		if summary != "" {
			fmt.Printf("  └── %s %s\n", green("✓"), summary)
		}
	} else {
		fmt.Printf("[ %s ] %s %s\n", s.command, s.message, green("✓"))
		if summary != "" {
			fmt.Printf("  └── %s\n", summary)
		}
	}
}

// Fail marks the spinner as failed
func (s *progressSpinner) Fail(summary string) {
	if s == nil {
		return
	}
	elapsed := time.Since(s.startTime)
	s.stopSpinner()
	if s.enabled {
		durationStr := formatDuration(elapsed)
		fmt.Printf("\r[%s] %s %s (%s)\n", s.command, s.message, red("✗"), gray(durationStr))
		if summary != "" {
			fmt.Printf("  └── %s %s\n", red("✗"), summary)
		}
	} else {
		fmt.Printf("[ %s ] %s %s\n", s.command, s.message, red("✗"))
		if summary != "" {
			fmt.Printf("  └── %s\n", summary)
		}
	}
}

func (s *progressSpinner) loop(startIndex int) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	i := startIndex
	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(s.startTime)
			dots := s.getEllipsis()
			durationStr := formatDuration(elapsed)
	// Use the same format as initialization: \r[ %s ] %s %s%s
	fmt.Printf("\r[ %s ] %s %s%s", s.command, s.message, string(spinnerFrames[i%len(spinnerFrames)]), gray(durationStr+dots))
			i++
			s.dotCounter = (s.dotCounter + 1) % 4
		case <-s.stop:
			// Final render without spinner
			elapsed := time.Since(s.startTime)
			fmt.Printf("\r[%s] %s... %s", s.command, s.message, gray(formatDuration(elapsed)))
			close(s.done)
			return
		}
	}
}

func (s *progressSpinner) getEllipsis() string {
	dots := ""
	for i := 0; i < s.dotCounter; i++ {
		dots += "."
	}
	return dots
}

func (s *progressSpinner) stopSpinner() {
	if !s.enabled {
		return
	}
	close(s.stop)
	<-s.done
}

// color helpers using logger - these create temporary loggers for color formatting
func green(text string) string {
	logger := NewLogger("temp")
	return logger.green(text)
}

func red(text string) string {
	logger := NewLogger("temp")
	return logger.red(text)
}

func gray(text string) string {
	logger := NewLogger("temp")
	return logger.gray(text)
}

// progressStep provides step-by-step progress for multi-stage operations
type progressStep struct {
	operation []string
	index     int
	total     int
}

// NewProgressStep creates a step-by-step progress tracker
func NewProgressStep(operation []string, total int) *progressStep {
	return &progressStep{
		operation: operation,
		total:     total,
	}
}

// Begin marks the start of a step
func (ps *progressStep) Begin(stepName string) {
	ps.index++
	fmt.Printf("[%s] Step %d/%d: %s\n", "command-name", ps.index, ps.total, stepName)
}

// Success marks completion of current step
func (ps *progressStep) Success(summary string) {
	fmt.Printf("    %s %s\n", "→", summary)
}

// Fail marks failure of current step
func (ps *progressStep) Fail(summary string) {
	fmt.Printf("    %s %s\n", "✗", summary)
}

// Complete marks overall completion
func (ps *progressStep) Complete(summary string) {
	fmt.Printf("[%s] %s %s\n", "command-name", "✓", summary)
}

// NewProgressPrinter creates a new progress printer with command logger integration
func NewProgressPrinter(label string, total int64) *progressPrinter {
	logger := NewCommandLogger("install")
	return &progressPrinter{
		logger:    logger,
		command:   "install",
		label:     label,
		total:     total,
		width:     32,
		lastPrint: time.Time{},
	}
}

// NewProgressPrinterWithLogger creates a new progress printer with a specific logger and command
func NewProgressPrinterWithLoggerAndCommand(label string, total int64, logger *Logger, command string) *progressPrinter {
	if command == "" {
		command = "install" // fallback
	}

	return &progressPrinter{
		logger:    logger,
		command:   command,
		label:     label,
		total:     total,
		width:     32,
		lastPrint: time.Time{},
	}
}

// NewProgressPrinterWithLogger creates a new progress printer with a specific logger
func NewProgressPrinterWithLogger(label string, total int64, logger *Logger) *progressPrinter {
	// Detect command from context of the download operation
	command := "download" // For download operations, use "download" as the command
	if strings.Contains(label, "nginx") {
		command = "nginx"
	} else if strings.Contains(label, "php") {
		command = "php"
	}

	return NewProgressPrinterWithLoggerAndCommand(label, total, logger, command)
}

// Write implements io.Writer to track progress as data is written
func (p *progressPrinter) Write(b []byte) (int, error) {
	n := len(b)
	p.current += int64(n)

	// Throttle updates to prevent spam (max every 250ms)
	// Also update on completion regardless of timing
	if time.Since(p.lastPrint) >= 250*time.Millisecond || p.current == p.total {
		p.render()
	}
	return n, nil
}

// Finish completes the progress display and shows final status
func (p *progressPrinter) Finish() {
	if p == nil || p.done {
		return
	}
	p.renderFinal()
	p.done = true
}

// render updates the progress display on the current line
func (p *progressPrinter) render() {
	if p.total > 0 {
		ratio := float64(p.current) / float64(p.total)
		if ratio > 1 {
			ratio = 1
		}
		filled := int(ratio * float64(p.width))
		if filled > p.width {
			filled = p.width
		}
		bar := strings.Repeat("#", filled) + strings.Repeat(".", p.width-filled)

		// Adjust display for terminal width to prevent wrapping/spam
		abbreviatedLabel := p.abbreviateLabel(p.label)

		// Update the same line (no spam - single line progress)
		fmt.Printf("\r\033[K[%s] %s... [%s] %3.0f%%", p.command, abbreviatedLabel, bar, ratio*100)
	} else {
		fmt.Printf("\r\033[K[%s] %s... %s", p.command, p.label, HumanBytes(p.current))
	}
	// Flush output to ensure immediate display
	p.lastPrint = time.Now()
}

// abbreviateLabel shortens the label to prevent terminal wrapping
func (p *progressPrinter) abbreviateLabel(label string) string {
	// Get terminal width (approximate)
	termWidth := 80 // default assumption
	if width, err := getTerminalWidth(); err == nil && width > 0 {
		termWidth = width
	}

	// Calculate maximum safe length for the label part
	// Format: [command] label... [bar] 100%
	// More aggressive: assume tighter terminal widths
	maxLabelLength := termWidth - 25 // command prefix + progress bar + percentage (more conservative)

	// Ensure minimum reasonable label length
	if maxLabelLength < 10 {
		maxLabelLength = 10
	}

	abbreviated := label
	// Always abbreviate long downloads to prevent wrapping regardless of terminal width
	if len(label) > 25 || len(label) > maxLabelLength {
		// For file downloads, intelligently shorten the filename
		if strings.Contains(label, ".tar.gz") {
			// Special handling for tar.gz files: "Download filename... short.tar.gz"
			parts := strings.Split(label, " ")
			if len(parts) >= 2 {
				// Take the action word and shorten the filename
				action := parts[0] // "Download"
				filename := strings.Join(parts[1:], " ")
				abstracted := p.abbreviateFilename(filename)
				abbreviated = action + " " + abstracted
			}
		} else {
			// Generic truncation for other cases
			abbreviated = label[:maxLabelLength-3] + "..."
		}
	}

	return abbreviated
}

// abbreviateFilename intelligently shortens filenames
func (p *progressPrinter) abbreviateFilename(filename string) string {
	// For files like "nginx_1.25.3_linux_amd64.tar.gz"
	if strings.Contains(filename, ".tar.gz") {
		parts := strings.Split(filename, "_")
		if len(parts) >= 3 {
			// For medium terminals: keep name + version, truncate architecture
			if len(parts[0]) > 0 && len(parts[1]) > 0 {
				return parts[0] + "_" + parts[1] + "_..." + ".tar.gz"
			}
		}
	}

	// Compact filename: name + only version number
	if strings.Contains(filename, ".tar.gz") {
		// Extract just version number (e.g., "2.10.2") from the filename
		if idx := strings.Index(filename, "_"); idx >= 0 {
			afterName := filename[idx+1:]
			if verEnd := strings.Index(afterName, "_"); verEnd > 0 {
				version := afterName[:verEnd]
				// Extract the base name (everything before first underscore)
				nameEnd := idx
				name := filename[:nameEnd]
				return name + "_" + version + "_" + "..." + ".tar.gz"
			}
		}
	}

	// Aggressive truncation for very long names
	if len(filename) > 25 {
		return filename[:15] + "..." + filename[len(filename)-5:]
	}

	return filename
}

// getTerminalWidth attempts to get the terminal width
func getTerminalWidth() (int, error) {
	// Simple implementation - in a real scenario, you could use syscall.Syscall or terminal escape codes
	// For now, return a reasonable default
	return 80, nil
}

// renderFinal displays the final completed progress with newline
func (p *progressPrinter) renderFinal() {
	if p.total > 0 {
		fmt.Printf("\r\033[K[%s] %s... [%s] 100%% (%s)\n", p.command, p.label, strings.Repeat("#", p.width), HumanBytes(p.current))
	} else {
		fmt.Printf("\r\033[K[%s] %s... %s\n", p.command, p.label, HumanBytes(p.current))
	}
}

// HumanBytes formats byte counts in human-readable format
func HumanBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	value := float64(n)
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", n, units[idx])
	}
	return fmt.Sprintf("%.1f %s", value, units[idx])
}
