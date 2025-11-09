package unlinkcurrent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestUnlinkRemovesCurrentProject(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "site-one")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	if err := commands.RunUnlink([]string{"--force"}); err != nil {
		t.Fatalf("RunUnlink failed: %v", err)
	}

	projectConfig := filepath.Join(workspace, "projects", "site-one", "project.yaml")
	if _, err := os.Stat(projectConfig); !os.IsNotExist(err) {
		t.Fatalf("expected project config removed, err=%v", err)
	}
}
