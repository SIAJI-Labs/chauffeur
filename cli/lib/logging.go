package lib

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

// LogLevel defines the severity of log messages
type LogLevel int

const (
	InfoLevel LogLevel = iota
	SuccessLevel
	WarnLevel
	ErrorLevel
	DebugLevel
)

// Logger provides flexible logging with command context and hierarchical output
type Logger struct {
	name       string
	command    string
	colors     bool
	level      LogLevel
	parent     bool // Whether this is a parent logger (for nested output)
	timestamps bool // Whether to include timestamps
}

// LoggerConfig holds configuration for logger creation
type LoggerConfig struct {
	Name       string   // Logger name (optional, overrides command)
	Command    string   // Command name for prefix
	Level      LogLevel // Minimum log level to output
	Colors     bool     // Enable color output (auto-detected if not set)
	Timestamps bool     // Include timestamps
	Parent     bool     // Is this a parent logger
}

// NewLogger creates a new logger instance with default configuration
func NewLogger(command string) *Logger {
	config := LoggerConfig{
		Command:    command,
		Colors:     isTerminal(os.Stdout),
		Level:      InfoLevel,
		Timestamps: false,
		Parent:     false,
	}
	return NewLoggerWithConfig(config)
}

// NewLoggerWithConfig creates a new logger with custom configuration
func NewLoggerWithConfig(config LoggerConfig) *Logger {
	if config.Colors == isTerminal(nil) {
		config.Colors = isTerminal(os.Stdout)
	}

	return &Logger{
		name:       config.Name,
		command:    config.Command,
		colors:     config.Colors,
		level:      config.Level,
		parent:     config.Parent,
		timestamps: config.Timestamps,
	}
}

// NewCommandLogger creates a logger optimized for CLI command output (backward compatibility)
func NewCommandLogger(command string) *Logger {
	return NewLogger(command)
}

// NewChildLogger creates a child logger for nested output with indentation
func (l *Logger) NewChildLogger(childCommand string) *Logger {
	return &Logger{
		name:       l.name,
		command:    childCommand,
		colors:     l.colors,
		level:      l.level,
		parent:     true,
		timestamps: l.timestamps,
	}
}

// WithLevel returns a new logger with the specified log level
func (l *Logger) WithLevel(level LogLevel) *Logger {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// WithColors returns a new logger with color output enabled/disabled
func (l *Logger) WithColors(enabled bool) *Logger {
	newLogger := *l
	newLogger.colors = enabled
	return &newLogger
}

// WithTimestamps returns a new logger with timestamp output enabled/disabled
func (l *Logger) WithTimestamps(enabled bool) *Logger {
	newLogger := *l
	newLogger.timestamps = enabled
	return &newLogger
}

// SetName sets the logger name
func (l *Logger) SetName(name string) {
	l.name = name
}

// GetName returns the logger name
func (l *Logger) GetName() string {
	return l.name
}

// GetCommand returns the command name
func (l *Logger) GetCommand() string {
	return l.command
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
	if f == nil {
		f = os.Stdout
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// colorize applies color formatting if terminal supports it
func (l *Logger) colorize(color, text string) string {
	if l.colors {
		return color + text + colorReset
	}
	return text
}

// blue formats text in blue
func (l *Logger) blue(text string) string {
	return l.colorize(colorBlue, text)
}

// green formats text in green
func (l *Logger) green(text string) string {
	return l.colorize(colorGreen, text)
}

// red formats text in red
func (l *Logger) red(text string) string {
	return l.colorize(colorRed, text)
}

// yellow formats text in yellow
func (l *Logger) yellow(text string) string {
	return l.colorize(colorYellow, text)
}

// gray formats text in gray (secondary info)
func (l *Logger) gray(text string) string {
	return l.colorize(colorGray, text)
}

// cyan formats text in cyan (for prompts and highlights)
func (l *Logger) cyan(text string) string {
	return l.colorize("\033[36m", text) // cyan ANSI code
}

// white formats text in white (for high contrast)
func (l *Logger) white(text string) string {
	return l.colorize("\033[37m", text) // white ANSI code
}

// bold formats text in bold
func (l *Logger) bold(text string) string {
	return l.colorize(colorBold, text)
}

// displayName returns the name to use in prefixes (name overrides command)
func (l *Logger) displayName() string {
	if l.name != "" {
		return l.name
	}
	return l.command
}

// prefix returns the formatted prefix for messages
func (l *Logger) prefix() string {
	displayName := l.displayName()
	if l.parent {
		return l.blue(fmt.Sprintf("  └── [ %s ]", displayName))
	}
	return l.blue(fmt.Sprintf("[ %s ]", displayName))
}

// uppercasePrefix returns the command prefix in uppercase (for backward compatibility)
func (l *Logger) uppercasePrefix() string {
	displayName := l.displayName()
	if l.parent {
		return l.blue(fmt.Sprintf("  └── [ %s ]", strings.ToUpper(displayName)))
	}
	return l.blue(fmt.Sprintf("[ %s ]", strings.ToUpper(displayName)))
}

// Prefix returns the command prefix for external use (backward compatibility)
func (l *Logger) Prefix() string {
	return l.prefix()
}

// formatMessage prepends optional timestamp and formats the full message
func (l *Logger) formatMessage(message string) string {
	if l.timestamps {
		timestamp := time.Now().Format("15:04:05")
		return fmt.Sprintf("[%s] %s %s", timestamp, l.prefix(), message)
	}
	return fmt.Sprintf("%s %s", l.prefix(), message)
}

// shouldLog checks if the message should be printed based on log level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// Info prints an informational message
func (l *Logger) Info(message string) {
	if !l.shouldLog(InfoLevel) {
		return
	}
	fmt.Printf("%s\n", l.formatMessage(message))
}

// Debug prints a debug message (only shown when log level is Debug)
func (l *Logger) Debug(message string) {
	if !l.shouldLog(DebugLevel) {
		return
	}
	debugPrefix := "  └──"
	if l.parent {
		debugPrefix = "      └──"
	}
	fmt.Printf("%s Debug: %s\n", debugPrefix, l.gray(message))
}

// Success prints a success message with optional context
func (l *Logger) Success(message, context string) {
	if !l.shouldLog(SuccessLevel) {
		return
	}
	baseMessage := fmt.Sprintf("%s %s", l.green("✓"), message)
	if context != "" {
		baseMessage += fmt.Sprintf(" (%s)", context)
	}
	fmt.Printf("%s\n", l.formatMessage(baseMessage))
}

// Error prints an error message with details and returns an error
func (l *Logger) Error(message, details string) error {
	if !l.shouldLog(ErrorLevel) {
		return fmt.Errorf("%s: %s", message, details)
	}

	errorIndent := "  └──"
	if l.parent {
		errorIndent = "      └──"
	}

	fmt.Printf("%s %s %s\n", l.prefix(), l.red("✗"), message)
	if details != "" {
		fmt.Printf("%s Error: %s\n", errorIndent, l.gray(details))
	}
	return fmt.Errorf("%s: %s", message, details)
}

// Warn prints a warning message with optional context
func (l *Logger) Warn(message, context string) {
	if !l.shouldLog(WarnLevel) {
		return
	}

	warnIndent := "  ⚠"
	contextIndent := "  └──"
	if l.parent {
		warnIndent = "    ⚠"
		contextIndent = "      └──"
	}

	fmt.Printf("%s Warning: %s\n", warnIndent, l.yellow(message))
	if context != "" {
		fmt.Printf("%s %s\n", contextIndent, l.gray(context))
	}
}

// Fail prints a failure message with error details and returns an error (backward compatibility)
func (l *Logger) Fail(message, error string) error {
	return l.Error(message, error)
}

// PrintSection prints a section header
func (l *Logger) PrintSection(title string) {
	fmt.Printf("\n%s %s\n", l.prefix(), l.bold(title))
}

// PrintSummary prints a summary section with items
func (l *Logger) PrintSummary(items []SummaryItem) {
	fmt.Printf("\n%s Summary:\n", l.prefix())
	for _, item := range items {
		fmt.Printf("  └── %s: %s\n", l.bold(item.Label), item.Value)
	}
	fmt.Printf("\n%s %s\n", l.prefix(), l.green("Complete"))
}

// PrintList prints a list of items with optional bullets
func (l *Logger) PrintList(items []string, bullet string) {
	if bullet == "" {
		bullet = "•"
	}
	for _, item := range items {
		fmt.Printf("  %s %s\n", bullet, item)
	}
}

// PrintTable prints a simple two-column table
func (l *Logger) PrintTable(headers []string, rows [][]string) {
	// Simple table printing - could be enhanced for column alignment
	if len(headers) > 0 {
		fmt.Printf("%s %s\n", l.prefix(), l.bold(strings.Join(headers, " | ")))
		fmt.Printf("%s %s\n", l.prefix(), strings.Repeat("-", len(headers[0]+headers[1])+7))
	}
	for _, row := range rows {
		fmt.Printf("%s %s | %s\n", l.prefix(), row[0], row[1])
	}
}

// Prompt prints a user prompt message with readable context
func (l *Logger) Prompt(message, context string) {
	promptIndent := "  →"
	contextIndent := "  └──"
	if l.parent {
		promptIndent = "    →"
		contextIndent = "      └──"
	}

	if context != "" {
		fmt.Printf("%s %s\n", promptIndent, l.cyan(message))
		fmt.Printf("%s %s\n", contextIndent, l.white(context))
	} else {
		fmt.Printf("%s %s\n", promptIndent, l.cyan(message))
	}
}

// SummaryItem represents an item in the summary
type SummaryItem struct {
	Label string
	Value string
}

// Summary creates a new summary item
func Summary(label, value string) SummaryItem {
	return SummaryItem{Label: label, Value: value}
}

// Utility functions for common operations

// LogOperationStart logs the start of an operation
func (l *Logger) LogOperationStart(operation string) {
	l.Info(fmt.Sprintf("Starting %s...", operation))
}

// LogOperationEnd logs the completion of an operation with duration
func (l *Logger) LogOperationEnd(operation string, duration time.Duration, success bool) {
	status := "completed"
	durStr := formatDuration(duration)

	if success {
		l.Success(fmt.Sprintf("%s %s", operation, status), durStr)
	} else {
		l.Error(fmt.Sprintf("%s failed", operation), durStr)
	}
}

// LogFileOperation logs a file operation (create, read, write, delete)
func (l *Logger) LogFileOperation(operation, filePath string, success bool) {
	action := fmt.Sprintf("%s file", operation)
	if success {
		l.Success(action, filePath)
	} else {
		l.Error(action, filePath)
	}
}

// LogNetworkOperation logs a network operation (download, upload, request)
func (l *Logger) LogNetworkOperation(operation, url string, size int64, success bool) {
	if size > 0 {
		sizeStr := fmt.Sprintf("%s (%s)", url, HumanBytes(size))
		if success {
			l.Success(fmt.Sprintf("%s complete", operation), sizeStr)
		} else {
			l.Error(fmt.Sprintf("%s failed", operation), sizeStr)
		}
	} else {
		if success {
			l.Success(fmt.Sprintf("%s complete", operation), url)
		} else {
			l.Error(fmt.Sprintf("%s failed", operation), url)
		}
	}
}
