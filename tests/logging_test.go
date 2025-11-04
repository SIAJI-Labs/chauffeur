package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestLogging_CommandOutput(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Test that commands produce reasonable output
	output := captureError(func() error {
		return commands.RunLinks([]string{})
	})

	// Should have some reasonable output (not panic/crash)
	if output == "" {
		t.Error("Expected some output from links command")
	}
	
	// Should not have panic/crash indicators
	if strings.Contains(output, "panic") || strings.Contains(output, "runtime error") {
		t.Errorf("Output contains panic/runtime error: %s", output)
	}
}

func TestLogging_ErrorFormat(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Test error formatting by trying an invalid operation
	output := captureError(func() error {
		return commands.RunLink([]string{"--invalid-flag"})
	})

	// Error messages should be properly formatted
	if output == "" {
		t.Error("Expected error output for invalid flags")
	}
}

func TestLogging_IntegrationTest(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create a basic setup to test logging in a real scenario
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Test the logging from a command that should succeed
	output := captureOutput(func() error {
		return commands.RunLinks([]string{})
	})

	// Should have reasonable output (not panic/crash)
	if output == "" {
		t.Error("Expected some output from links command")
	}
	
	// Should not have panic/crash indicators
	if strings.Contains(output, "panic") || strings.Contains(output, "runtime error") {
		t.Errorf("Output contains panic/runtime error: %s", output)
	}
}

// Test capturing helper functionality similar to existing tests
func TestCapturingHelper(t *testing.T) {
	// Test that our capturing works as expected for logging tests
	output := captureOutput(func() error {
		// This will print to stdout, which we capture
		fmt.Print("test output")
		return nil
	})
	
	if output != "test output" {
		t.Errorf("Expected 'test output', got %q", output)
	}
}

// Test that logging system doesn't interfere with command execution
func TestLogging_NonInterference(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	start := time.Now()
	output := captureOutput(func() error {
		return commands.RunLinks([]string{})
	})
	duration := time.Since(start)

	// Command should complete quickly, logging shouldn't introduce major delays
	if duration > 5*time.Second {
		t.Errorf("Command took too long (%v), logging may be interfering", duration)
	}

	// Should not have any crash/panic output
	if strings.Contains(output, "panic") || strings.Contains(output, "fatal") {
		t.Errorf("Output contains panic/fatal error: %s", output)
	}
}
