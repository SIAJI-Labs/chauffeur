package doctortest

import (
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorPortDetectionWithServicesRunning(t *testing.T) {
	t.Skip("Skipping port detection test with services running in CI - requires dnsmasq setup")
}

func TestDoctorPortDetectionAfterServicesStopped(t *testing.T) {
	t.Skip("Skipping port detection test after services stopped in CI - requires dnsmasq setup")
}

func TestDoctorPortDetectionWithMultipleProjects(t *testing.T) {
	t.Skip("Skipping port detection test with multiple projects in CI - requires dnsmasq setup")
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
	t.Skip("Skipping all port related checks test in CI - requires dnsmasq setup")
}