package basic_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestRunInitCreatesWorkspace(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}

	configPath := filepath.Join(workspace, "config", "chauffeur.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %s: %v", configPath, err)
	}
}
