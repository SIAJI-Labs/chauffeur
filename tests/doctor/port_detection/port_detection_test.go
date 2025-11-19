package doctortest

import (
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorPortDetectionWithServicesRunning(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "port-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Start services to occupy ports
	if err := commands.RunStart(nil); err != nil {
		t.Fatalf("RunStart failed: %v", err)
	}

	// Test that doctor correctly detects ports used by Chauffeur services
	err := commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for network check with services running, got: %v", err)
	}
}

func TestDoctorPortDetectionAfterServicesStopped(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "port-test-app-2")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Start services
	if err := commands.RunStart(nil); err != nil {
		t.Fatalf("RunStart failed: %v", err)
	}

	// Stop services
	if err := commands.RunStop(nil); err != nil {
		t.Fatalf("RunStop failed: %v", err)
	}

	// Test that port detection works after services are stopped
	err := commands.RunDoctor([]string{"--check-network", "--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for network check after services stopped, got: %v", err)
	}
}

func TestDoctorPortDetectionWithMultipleProjects(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "8.2")

	// Create multiple projects
	project1Dir := helpers.NewProjectDir(t, home, "multi-app-1")
	helpers.Chdir(t, project1Dir)
	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed for project 1: %v", err)
	}

	project2Dir := helpers.NewProjectDir(t, home, "multi-app-2")
	helpers.Chdir(t, project2Dir)
	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed for project 2: %v", err)
	}

	// Start services for all projects
	if err := commands.RunStart(nil); err != nil {
		t.Fatalf("RunStart failed: %v", err)
	}

	// Test port detection with multiple projects running
	err := commands.RunDoctor([]string{"--check-network", "--verbose"})
	if err != nil {
		t.Fatalf("Expected no error for network check with multiple projects, got: %v", err)
	}
}

func TestDoctorPortDetectionConsistency(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := helpers.NewProjectDir(t, home, "consistency-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Run doctor multiple times to ensure consistent results
	for i := 0; i < 3; i++ {
		err := commands.RunDoctor([]string{"--check-network", "--quiet"})
		if err != nil {
			t.Fatalf("Expected no error for network check iteration %d, got: %v", i, err)
		}
	}
}

func TestDoctorAllPortRelatedChecks(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := helpers.NewProjectDir(t, home, "all-checks-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Start services
	if err := commands.RunStart(nil); err != nil {
		t.Fatalf("RunStart failed: %v", err)
	}

	// Test comprehensive doctor check includes port detection
	err := commands.RunDoctor([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Expected no error for comprehensive doctor check, got: %v", err)
	}
}