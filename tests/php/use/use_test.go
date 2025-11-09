package use_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestPHPUseUpdatesDefaultVersion(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "8.2")

	if err := commands.RunPHP([]string{"use", "8.2"}); err != nil {
		t.Fatalf("php use returned error: %v", err)
	}

	configPath := filepath.Join(workspace, "config", "chauffeur.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "default: 8.2") {
		t.Fatalf("expected config to contain default: 8.2, got:\n%s", string(data))
	}

	phpShim := filepath.Join(workspace, "bin", "php")
	if _, err := os.Stat(phpShim); err != nil {
		t.Fatalf("expected php shim at %s: %v", phpShim, err)
	}
}
