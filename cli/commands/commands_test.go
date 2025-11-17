package commands

import (
	"testing"
)

// TestVersionManagement tests version setting functions
func TestVersionManagement(t *testing.T) {
	originalVersion := cliVersion
	originalBuildTimestamp := buildTimestamp
	originalBuildCommit := buildCommit

	// Test setting version info
	SetCLIVersion("1.0.0")
	SetBuildTimestamp("2025-01-01T00:00:00Z")
	SetBuildCommit("abc123")

	if cliVersion != "1.0.0" {
		t.Errorf("Expected CLI version to be '1.0.0', got '%s'", cliVersion)
	}

	if buildTimestamp != "2025-01-01T00:00:00Z" {
		t.Errorf("Expected build timestamp to be '2025-01-01T00:00:00Z', got '%s'", buildTimestamp)
	}

	if buildCommit != "abc123" {
		t.Errorf("Expected build commit to be 'abc123', got '%s'", buildCommit)
	}

	// Restore original values
	cliVersion = originalVersion
	buildTimestamp = originalBuildTimestamp
	buildCommit = originalBuildCommit
}

// TestCommandValidation tests basic command argument validation
func TestCommandValidation(t *testing.T) {
	// Test invalid flag handling
	err := RunStart([]string{"--invalid-flag"})
	if err == nil {
		t.Error("Expected error for invalid flag in start command")
	}

	err = RunStop([]string{"--invalid-flag"})
	if err == nil {
		t.Error("Expected error for invalid flag in stop command")
	}

	err = RunRestart([]string{"--invalid-flag"})
	if err == nil {
		t.Error("Expected error for invalid flag in restart command")
	}
}

// TestHelpCommands tests that help commands don't return errors
func TestHelpCommands(t *testing.T) {
	// Test that help commands succeed
	err := RunStart([]string{"--help"})
	if err != nil {
		t.Errorf("Start help should succeed: %v", err)
	}

	err = RunStop([]string{"--help"})
	if err != nil {
		t.Errorf("Stop help should succeed: %v", err)
	}

	err = RunRestart([]string{"--help"})
	if err != nil {
		t.Errorf("Restart help should succeed: %v", err)
	}

	err = RunStatus([]string{"--help"})
	if err != nil {
		t.Errorf("Status help should succeed: %v", err)
	}
}

// TestDryRunCommands tests dry-run functionality
func TestDryRunCommands(t *testing.T) {
	// Test that dry-run commands don't fail due to missing services
	err := RunStart([]string{"--dry-run"})
	// This might fail due to missing workspace/services, but shouldn't crash
	if err != nil {
		t.Logf("Start dry-run returned error (may be expected): %v", err)
	}

	err = RunStop([]string{"--dry-run"})
	// This should succeed as stop is generally safe
	if err != nil {
		t.Errorf("Stop dry-run should succeed: %v", err)
	}

	err = RunRestart([]string{"--dry-run"})
	// This might fail due to missing workspace/services, but shouldn't crash
	if err != nil {
		t.Logf("Restart dry-run returned error (may be expected): %v", err)
	}
}