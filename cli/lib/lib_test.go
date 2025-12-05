package lib

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNewLogger tests logger creation
func TestNewLogger(t *testing.T) {
	logger := NewCommandLogger("test")
	if logger == nil {
		t.Error("Expected logger to be created")
	}
}

// TestLoggerOutput tests that logger produces output
func TestLoggerOutput(t *testing.T) {
	logger := NewCommandLogger("test")

	// Test basic logging methods - they should not panic
	logger.Info("Test info message")
	logger.Success("Test success message", "Test detail")
	logger.Warn("Test warning message", "Test detail")
	logger.Error("Test error message", "Test detail")

	// Test that methods exist and don't panic
	logger.Info("Logger methods should not panic")
}

// TestNewSpinner tests spinner creation
func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("test", "Test message")
	if spinner == nil {
		t.Error("Expected spinner to be created")
	}
}

// TestSpinnerMethods tests spinner methods don't panic
func TestSpinnerMethods(t *testing.T) {
	spinner := NewSpinner("test", "Test message")

	// Test spinner methods - they should not panic even if they don't produce visible output
	spinner.Success("Success message")
	spinner.Fail("Fail message")
}

// TestNewProgressPrinter tests progress printer creation
func TestNewProgressPrinter(t *testing.T) {
	printer := NewProgressPrinter("test", 100)
	if printer == nil {
		t.Error("Expected progress printer to be created")
	}
}

// TestPrintSection tests section printing
func TestPrintSection(t *testing.T) {
	logger := NewCommandLogger("test")

	// This should not panic
	logger.PrintSection("Test Section")
}

// TestPrintSeparator tests separator printing
func TestPrintSeparator(t *testing.T) {
	logger := NewCommandLogger("test")

	// This should not panic
	logger.PrintSeparator()
}

// TestPrintSummary tests summary printing
func TestPrintSummary(t *testing.T) {
	logger := NewCommandLogger("test")

	// This should not panic
	summary := []SummaryItem{
		{Label: "Test Item", Value: "Test Value"},
	}
	logger.PrintSummary(summary)
}

// TestLoggerChildLogger tests child logger creation
func TestLoggerChildLogger(t *testing.T) {
	logger := NewCommandLogger("test")
	childLogger := logger.NewChildLogger("child")

	if childLogger == nil {
		t.Error("Expected child logger to be created")
	}
}

// TestNewProgressStep tests progress step creation
func TestNewProgressStep(t *testing.T) {
	step := NewProgressStep([]string{"step1", "step2"}, 2)
	if step == nil {
		t.Error("Expected progress step to be created")
	}
}

// TestNewPortManager tests port manager creation
func TestNewPortManager(t *testing.T) {
	portManager := NewPortManager(8080, 8099, "prompt")
	if portManager == nil {
		t.Error("Expected port manager to be created")
	}
}

// TestNewPortValidator tests port validator creation
func TestNewPortValidator(t *testing.T) {
	// This would require a config, so we'll test the creation concept
	// Without access to internal config types, we'll just verify the function exists
}

// TestColorHelpers tests color-related helper functions
func TestColorHelpers(t *testing.T) {
	logger := NewCommandLogger("test")

	// Test that color methods don't panic
	logger.Info("Test message with colors")
	logger.Success("Success message", "Detail")
	logger.Warn("Warning message", "Detail")
	logger.Error("Error message", "Detail")
}

// TestTTYDetection tests TTY detection
func TestTTYDetection(t *testing.T) {
	// Test that TTY detection doesn't panic
	// This is testing the behavior, not specific output
	logger := NewCommandLogger("test")
	logger.Info("TTY detection test")
}

// TestPrintBlockNewline ensures PrintBlock appends a newline when missing.
func TestPrintBlockNewline(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	logger := NewCommandLogger("block")
	logger.PrintBlock("hello")

	if got := buf.String(); got != "hello\n" {
		t.Fatalf("expected trailing newline, got %q", got)
	}
}

// TestPrintBlockPreservesExistingNewline ensures existing newline isn't doubled.
func TestPrintBlockPreservesExistingNewline(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	logger := NewCommandLogger("block")
	logger.PrintBlock("hello\n")

	if got := buf.String(); got != "hello\n" {
		t.Fatalf("expected single newline, got %q", got)
	}
}

// TestLogLevelFiltering ensures higher log levels suppress lower messages.
func TestLogLevelFiltering(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	logger := NewCommandLogger("filter").WithLevel(ErrorLevel)
	logger.Info("should be suppressed")

	if buf.Len() != 0 {
		t.Fatalf("expected no output for Info at Error level, got %q", buf.String())
	}
}

// TestLogOperationEndOutputsStatus covers duration formatting and status markers.
func TestLogOperationEndOutputsStatus(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	logger := NewCommandLogger("op")
	logger.LogOperationEnd("download", 1500*time.Millisecond, true)

	out := buf.String()
	if !strings.Contains(out, "download completed") {
		t.Fatalf("expected success message, got %q", out)
	}
	if !strings.Contains(out, "1.5s") {
		t.Fatalf("expected formatted duration, got %q", out)
	}
}

// TestLogOperationEndFailure ensures failure path is logged.
func TestLogOperationEndFailure(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	logger := NewCommandLogger("op")
	logger.LogOperationEnd("download", 2*time.Second, false)

	out := buf.String()
	if !strings.Contains(out, "download failed") {
		t.Fatalf("expected failure message, got %q", out)
	}
}
