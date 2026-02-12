package selfupdatedev_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestSelfUpdateDevRebuildsBinary(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	repoDir := filepath.Join(home, "chauffeur-src")
	required := []string{
		filepath.Join(repoDir, ".git"),
		filepath.Join(repoDir, "cli"),
		filepath.Join(repoDir, "docs"),
	}
	for _, dir := range required {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(repoDir, "cli", "main.go"), []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatalf("write main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module github.com/siaji/chauffeur\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "docs", "README.md"), []byte("# Chauffeur Documentation\n"), 0o644); err != nil {
		t.Fatalf("write docs/README.md: %v", err)
	}

	helpers.Chdir(t, repoDir)

	target := filepath.Join(workspace, "bin", "chauf")
	t.Setenv("CHAUF_SELF_UPDATE_TARGET", target)

	reset := commands.OverrideSelfUpdateHooks(
		func(string, string, ...string) (string, error) { return "abcdef\n", nil },
		func(_, target, _ string) error {
			return os.WriteFile(target, []byte("binary"), 0o755)
		},
	)
	defer reset()

	if err := commands.RunSelfUpdate([]string{"--dev"}); err != nil {
		t.Fatalf("RunSelfUpdate --dev failed: %v", err)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected binary at %s: %v", target, err)
	}
}
