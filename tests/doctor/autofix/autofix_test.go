package doctortest

import (
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorAutoFixOptionRecognition(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test that --fix option is recognized without panicking (using quiet mode)
	err := commands.RunDoctor([]string{"--check-deps", "--fix", "--quiet"})
	// This should process the fix option without executing actual fixes in quiet mode
	if err != nil {
		t.Fatalf("Expected no error for doctor fix option recognition, got: %v", err)
	}
}

func TestDoctorAutoFixCombination(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test combining fix with other flags
	err := commands.RunDoctor([]string{"--check-php", "--fix", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor fix combination, got: %v", err)
	}
}

func TestDoctorFixWithVerbose(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test fix option with verbose mode
	err := commands.RunDoctor([]string{"--check-deps", "--fix", "--verbose"})
	if err != nil {
		t.Fatalf("Expected no error for doctor fix with verbose, got: %v", err)
	}
}