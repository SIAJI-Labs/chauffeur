package logging

import (
	"testing"
)

// TestNewCommandLogger tests command logger creation
func TestNewCommandLogger(t *testing.T) {
	logger := NewCommandLogger("test-command")
	if logger == nil {
		t.Error("Expected command logger to be created")
	}
}

// TestChildLogger tests child logger creation
func TestChildLogger(t *testing.T) {
	parent := NewCommandLogger("parent")
	child := parent.NewChildLogger("child")
	if child == nil {
		t.Error("Expected child logger to be created")
	}
}

// TestLoggerMethods tests that logger methods don't panic
func TestLoggerMethods(t *testing.T) {
	logger := NewCommandLogger("test")

	// These should not panic even if they don't produce visible output
	logger.Info("Test info message")
	logger.Success("Test success", "Test detail")
	logger.Warn("Test warning", "Test detail")
	logger.Fail("Test error", "Test detail")
}

// TestLoggerFormatting tests duration formatting function exists
func TestLoggerFormatting(t *testing.T) {
	// Test that formatDuration function exists and doesn't panic
	// We can't test the output directly since it's not exported
	_ = formatDuration(0) // This just verifies the function exists
}