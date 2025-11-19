package defaultlink_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestLinkRegistersCurrentDirectory(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "demo-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "demo-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "domain: demo-app.test") {
		t.Fatalf("expected default .test domain in project config:\n%s", content)
	}
	if !strings.Contains(content, `php: "8.3"`) && !strings.Contains(content, "php: 8.3") {
		t.Fatalf("expected php version 8.3 in project config:\n%s", content)
	}
}
