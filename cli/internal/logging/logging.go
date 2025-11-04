package logging

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ANSI Color constants
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// CommandLogger provides standardized logging for CLI commands
type CommandLogger struct {
	command string
	colors  bool
	parent  bool  // Whether this is a parent logger (for nested output)
}

// NewCommandLogger creates a new command logger instance
func NewCommandLogger(command string) *CommandLogger {
	return &CommandLogger{
		command: command,
		colors:  isTerminal(os.Stdout),
		parent:  false,
	}
}

// NewChildLogger creates a child logger for nested output
func (l *CommandLogger) NewChildLogger(childCommand string) *CommandLogger {
	return &CommandLogger{
		command: childCommand,
		colors:  l.colors,
		parent:  true,
	}
}

// formatDuration converts duration to human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000.0)
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", float64(d.Nanoseconds())/1000000000.0)
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// isTerminal checks if the file is a terminal
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// colorize applies color formatting if terminal supports it
func (l *CommandLogger) colorize(color, text string) string {
	if l.colors {
		return color + text + colorReset
	}
	return text
}

// blue formats command prefix in blue
func (l *CommandLogger) blue(text string) string {
	return l.colorize(colorBlue, text)
}

// green formats text in green
func (l *CommandLogger) green(text string) string {
	return l.colorize(colorGreen, text)
}

// red formats text in red
func (l *CommandLogger) red(text string) string {
	return l.colorize(colorRed, text)
}

// yellow formats text in yellow
func (l *CommandLogger) yellow(text string) string {
	return l.colorize(colorYellow, text)
}

// gray formats text in gray (secondary info)
func (l *CommandLogger) gray(text string) string {
	return l.colorize(colorGray, text)
}

// bold formats text in bold
func (l *CommandLogger) bold(text string) string {
	return l.colorize(colorBold, text)
}

// prefix returns the command prefix
func (l *CommandLogger) prefix() string {
	if l.parent {
		return l.blue(fmt.Sprintf("  └── [ %s ]", l.command))
	}
	return l.blue(fmt.Sprintf("[ %s ]", l.command))
}

// Prefix returns the command prefix for external use
func (l *CommandLogger) Prefix() string {
	return l.prefix()
}

// uppercasePrefix returns the command prefix in uppercase
func (l *CommandLogger) uppercasePrefix() string {
	if l.parent {
		return l.blue(fmt.Sprintf("  └── [ %s ]", strings.ToUpper(l.command)))
	}
	return l.blue(fmt.Sprintf("[ %s ]", strings.ToUpper(l.command)))
}

// Info prints an informational message
func (l *CommandLogger) Info(message string) {
	fmt.Printf("%s %s\n", l.prefix(), message)
}

// Success prints a success message with optional context
func (l *CommandLogger) Success(message, context string) {
	if context != "" {
		fmt.Printf("%s %s %s (%s)\n", l.prefix(), l.green("✓"), message, context)
	} else {
		fmt.Printf("%s %s %s\n", l.prefix(), l.green("✓"), message)
	}
}

// Fail prints a failure message with error details and returns an error
func (l *CommandLogger) Fail(message, error string) error {
	errorIndent := "  └──"
	if l.parent {
		errorIndent = "      └──"
	}
	
	if error != "" {
		fmt.Printf("%s %s %s\n", l.prefix(), l.red("✗"), message)
		fmt.Printf("%s Error: %s\n", errorIndent, l.gray(error))
	} else {
		fmt.Printf("%s %s %s\n", l.prefix(), l.red("✗"), message)
	}
	return fmt.Errorf("%s: %s", message, error)
}

// Warn prints a warning message with optional context
func (l *CommandLogger) Warn(message, context string) {
	warnIndent := "  ⚠"
	contextIndent := "  └──"
	if l.parent {
		warnIndent = "    ⚠"
		contextIndent = "      └──"
	}
	
	if context != "" {
		fmt.Printf("%s Warning: %s\n", warnIndent, l.yellow(message))
		fmt.Printf("%s %s\n", contextIndent, l.gray(context))
	} else {
		fmt.Printf("%s Warning: %s\n", warnIndent, l.yellow(message))
	}
}

// PrintSummary prints a summary section
func (l *CommandLogger) PrintSummary(items []SummaryItem) {
	fmt.Printf("\n%s Summary:\n", l.prefix())
	for _, item := range items {
		fmt.Printf("  └── %s: %s\n", l.bold(item.Label), item.Value)
	}
	fmt.Printf("\n%s %s\n", l.prefix(), l.green("Complete"))
}

// SummaryItem represents an item in the summary
type SummaryItem struct {
	Label string
	Value string
}

// progressSpinner provides a simple spinner for indeterminate operations
type progressSpinner struct {
	message    string
	startTime  time.Time
	dotCounter int
}

// NewSpinner creates a new progress spinner
func (l *CommandLogger) NewSpinner(message string) *progressSpinner {
	fmt.Printf("%s %s...\n", l.prefix(), message)
	return &progressSpinner{
		message:   message,
		startTime: time.Now(),
	}
}

// Success marks the spinner as successful
func (spin *progressSpinner) Success(message string) {
	duration := time.Since(spin.startTime)
	fmt.Printf("    %s %s (%s)\n", "→", message, formatDuration(duration))
}

// Fail marks the spinner as failed
func (spin *progressSpinner) Fail(message string) {
	duration := time.Since(spin.startTime)
	fmt.Printf("    %s %s (%s)\n", "✗", message, formatDuration(duration))
}

// progressPrinter provides progress bar for determinate operations
type progressPrinter struct {
	label     string
	total     int64
	current   int64
	width     int
	lastPrint time.Time
}

// NewProgressPrinter creates a new progress printer
func (l *CommandLogger) NewProgressPrinter(label string, total int64) *progressPrinter {
	fmt.Printf("    - %s", label)
	return &progressPrinter{
		label:     label,
		total:     total,
		width:     20,
		lastPrint: time.Time{},
	}
}

// Write for progress tracking
func (p *progressPrinter) Write(data []byte) (int, error) {
	p.current += int64(len(data))
	percentage := float64(p.current) / float64(p.total) * 100
	filled := int(percentage / 100 * float64(p.width))
	
	bar := ""
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "."
		}
	}
	
	fmt.Printf("\r    - %s [%s] %.0f%%", p.label, bar, percentage)
	
	if p.current >= p.total {
		fmt.Printf(" (%.1f MiB)\n", float64(p.total)/1024/1024)
	}
	
	return len(data), nil
}
