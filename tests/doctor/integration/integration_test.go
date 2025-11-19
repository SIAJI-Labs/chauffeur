package doctortest

import (
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

	// Test doctor with fully initialized workspace
	err := commands.RunDoctor([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor with initialized workspace, got: %v", err)
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

	// Test doctor after project linking
	err := commands.RunDoctor([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor after project linking, got: %v", err)
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

	// Test doctor with multiple PHP versions available
	err := commands.RunDoctor([]string{"--check-php", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor with multiple PHP versions, got: %v", err)
	}
}

func TestDoctorBeforeAndAfterServiceLifecycle(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := helpers.NewProjectDir(t, home, "lifecycle-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Test doctor before starting services
	err := commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor before service start, got: %v", err)
	}

	// Start services
	if err := commands.RunStart(nil); err != nil {
		t.Fatalf("RunStart failed: %v", err)
	}

	// Test doctor after starting services
	err = commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor after service start, got: %v", err)
	}

	// Stop services
	if err := commands.RunStop(nil); err != nil {
		t.Fatalf("RunStop failed: %v", err)
	}

	// Test doctor after stopping services
	err = commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for doctor after service stop, got: %v", err)
	}
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
		if err != nil {
			t.Fatalf("Expected no error for doctor run %d, got: %v", i+1, err)
		}
	}
}