package linkslist_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestLinksPrintsProjectsTable(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	projectsDir := filepath.Join(workspace, "projects")
	writeProject := func(slug, path, domain string) {
		dir := filepath.Join(projectsDir, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir project: %v", err)
		}
		content := "version: 1\npath: " + path + "\nphp: 8.3\nsite:\n  domain: " + domain + "\n  ssl: false\ncreated_at: 2025-11-09T00:00:00Z\n"
		if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content), 0o644); err != nil {
			t.Fatalf("write project config: %v", err)
		}
	}

	writeProject("alpha", "/path/alpha", "alpha.test")
	writeProject("beta", "/path/beta", "beta.test")

	output := helpers.CaptureOutput(t, func() {
		if err := commands.RunLinks(nil); err != nil {
			t.Fatalf("RunLinks failed: %v", err)
		}
	})

	if !containsAll(output, "alpha", "beta") {
		t.Fatalf("expected output to list both projects, got:\n%s", output)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
