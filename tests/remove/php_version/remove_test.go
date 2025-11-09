package phpversion_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestRemovePHPVersionDeletesRuntime(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "7.4")

	args := []string{"--force", "php", "7.4"}
	if err := commands.RunRemove(args); err != nil {
		t.Fatalf("RunRemove returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workspace, "php", "7.4")); !os.IsNotExist(err) {
		t.Fatalf("expected PHP 7.4 directory to be removed, err=%v", err)
	}
}
