package doctortest

import (
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorBasicFunctionality(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test that doctor command runs without errors for system dependencies check
	err := commands.RunDoctor([]string{"--check-deps", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor system deps check, got: %v", err)
	}
}

func TestDoctorPHPDependencies(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test PHP dependencies check
	err := commands.RunDoctor([]string{"--check-php", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor PHP deps check, got: %v", err)
	}
}

func TestDoctorSSLCertificates(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test SSL certificates check
	err := commands.RunDoctor([]string{"--check-ssl", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor SSL check, got: %v", err)
	}
}

func TestDoctorNetworkDependencies(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test network dependencies check
	err := commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor network check, got: %v", err)
	}
}

func TestDoctorAllChecks(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test all checks (default behavior)
	err := commands.RunDoctor([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor all checks, got: %v", err)
	}
}


func TestDoctorWithoutWorkspace(t *testing.T) {
	// Use temporary directory instead of workspace setup
	_ = t.TempDir()

	// Test that doctor can run system dependency checks even without workspace
	err := commands.RunDoctor([]string{"--check-deps", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor without workspace, got: %v", err)
	}
}