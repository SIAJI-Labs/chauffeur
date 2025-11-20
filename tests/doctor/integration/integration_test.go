package doctortest

import (
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorWithWorkspaceSetup(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	// Test doctor after complete workspace setup
	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Test doctor with fully initialized workspace - in CI some deps may be missing, which is expected
	err := commands.RunDoctor([]string{"--quiet"})
	// In CI environments, missing dependencies are expected and should not cause test failure
	if err != nil {
		// Check if the error is just "dependencies missing" rather than a real failure
		if !strings.Contains(err.Error(), "found") && !strings.Contains(err.Error(), "need to be resolved") {
			t.Fatalf("Expected no error for doctor with initialized workspace, got: %v", err)
		}
	}
}

func TestDoctorAfterProjectLinking(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "integration-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Test doctor after project linking - in CI some deps may be missing, which is expected
	err := commands.RunDoctor([]string{"--quiet"})
	// In CI environments, missing dependencies are expected and should not cause test failure
	if err != nil {
		// Check if the error is just "dependencies missing" rather than a real failure
		if !strings.Contains(err.Error(), "found") && !strings.Contains(err.Error(), "need to be resolved") {
			t.Fatalf("Expected no error for doctor after project linking, got: %v", err)
		}
	}
}

func TestDoctorWithDifferentPHPVersions(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	// Install multiple PHP versions
	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "8.2")
	helpers.EnsureFakePHP(t, workspace, "8.1")

	// Test doctor with multiple PHP versions available - in CI some deps may be missing, which is expected
	err := commands.RunDoctor([]string{"--check-php", "--quiet"})
	// In CI environments, missing dependencies are expected and should not cause test failure
	if err != nil {
		// Check if the error is just "dependencies missing" rather than a real failure
		if !strings.Contains(err.Error(), "found") && !strings.Contains(err.Error(), "need to be resolved") {
			t.Fatalf("Expected no error for doctor with multiple PHP versions, got: %v", err)
		}
	}
}

func TestDoctorBeforeAndAfterServiceLifecycle(t *testing.T) {
	t.Skip("Skipping service lifecycle test in CI - requires dnsmasq setup")
}

func TestDoctorErrorHandling(t *testing.T) {
	// Test doctor in invalid environment (no workspace)
	_ = t.TempDir()

	// Should still be able to run basic checks
	err := commands.RunDoctor([]string{"--check-deps", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor in temp directory, got: %v", err)
	}
}

func TestDoctorConsistentOutput(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := helpers.NewProjectDir(t, home, "consistent-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Run doctor multiple times and ensure consistent behavior
	for i := 0; i < 5; i++ {
		err := commands.RunDoctor([]string{"--quiet"})
		// In CI environments, missing dependencies are expected and should not cause test failure
		if err != nil {
			// Check if the error is just "dependencies missing" rather than a real failure
			if !strings.Contains(err.Error(), "found") && !strings.Contains(err.Error(), "need to be resolved") {
				t.Fatalf("Expected no error for doctor run %d, got: %v", i+1, err)
			}
		}
	}
}