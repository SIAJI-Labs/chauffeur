package commands

import (
	"os"
	"os/exec"
	"testing"

	"github.com/siaji/chauffeur/cli/lib" // Import the lib package for CommandExecutor
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
	// Mock exec.Command to prevent actual iptables calls
	lib.SetCommandExecutor(func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	})
	defer lib.ResetCommandExecutor()

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

// TestHelperProcess is a helper function for mocking exec.Command
// It's not a real test, but a way to intercept exec.Command calls.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		return // Should not happen
	}

	cmd := args[0]
	switch cmd {
	case "sudo":
		// Simulate successful sudo operations
		return
	case "mkcert":
		// Simulate successful mkcert operations
		return
	default:
		// Default to success for other commands
		return
	}
}