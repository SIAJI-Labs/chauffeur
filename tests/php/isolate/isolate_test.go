package isolate_test

import (
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestPHPIsolateUpdatesProjectConfig(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "8.2")

	projectDir := helpers.NewProjectDir(t, home, "isolated-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	if err := commands.RunPHP([]string{"isolate", "8.2"}); err != nil {
		t.Fatalf("php isolate failed: %v", err)
	}

	projectConfig := filepath.Join(workspace, "projects", "isolated-app", "project.yaml")
	helpers.AssertFileContains(t, projectConfig, "php: 8.2")
}
