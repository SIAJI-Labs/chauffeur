package doctortest

import (
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorBasicOptionProcessing(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test that basic options are processed correctly
	err := commands.RunDoctor([]string{"--check-deps"})
	if err != nil {
		t.Fatalf("Expected no error for basic option processing, got: %v", err)
	}
}

func TestDoctorVerboseOutput(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test verbose mode doesn't cause errors
	err := commands.RunDoctor([]string{"--check-deps", "--verbose"})
	if err != nil {
		t.Fatalf("Expected no error for doctor verbose mode, got: %v", err)
	}
}

func TestDoctorQuietOutput(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test quiet mode
	err := commands.RunDoctor([]string{"--check-deps", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor quiet mode, got: %v", err)
	}
}

func TestDoctorDNSSpecificCheck(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test DNS specific check
	err := commands.RunDoctor([]string{"--check-dns"})
	if err != nil {
		t.Fatalf("Expected no error for doctor DNS check, got: %v", err)
	}
}

func TestDoctorMultipleSpecificChecks(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test combining multiple specific checks
	err := commands.RunDoctor([]string{"--check-deps", "--check-php", "--check-network"})
	if err != nil {
		t.Fatalf("Expected no error for multiple specific checks, got: %v", err)
	}
}

func TestDoctorPortDetectionWithNoServices(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test port detection when no services are running
	err := commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for network check without services, got: %v", err)
	}
}