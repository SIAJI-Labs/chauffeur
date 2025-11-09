package basicinfo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestRunInfoDisplaysWorkspaceDetails(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Provide a fake composer binary so info lists it as installed.
	composerPath := filepath.Join(workspace, "bin", "composer")
	if err := os.MkdirAll(filepath.Dir(composerPath), 0o755); err != nil {
		t.Fatalf("mkdir composer dir: %v", err)
	}
	if err := os.WriteFile(composerPath, []byte("#!/usr/bin/env bash\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write composer shim: %v", err)
	}

	commands.OverrideLatestVersionFetcher(func() (string, error) {
		return "v9.9.9", nil
	})
	t.Cleanup(func() { commands.OverrideLatestVersionFetcher(nil) })

	output := helpers.CaptureOutput(t, func() {
		if err := commands.RunInfo(nil); err != nil {
			t.Fatalf("RunInfo returned error: %v", err)
		}
	})

	if !strings.Contains(output, workspace) {
		t.Fatalf("expected output to mention workspace %s, got:\n%s", workspace, output)
	}
	if !strings.Contains(output, "Latest release: v9.9.9") {
		t.Fatalf("expected latest release line, got:\n%s", output)
	}
}
