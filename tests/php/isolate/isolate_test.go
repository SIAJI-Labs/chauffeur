package isolate_test

import (
	"os"
	"path/filepath"
	"strings"
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
	helpers.AssertFileContains(t, projectConfig, `php: "8.2"`)

	// Test that PHP-FPM socket paths are also updated (critical bug fix)
	expectedSocketPath := filepath.Join(workspace, "php", "8.2", "runtime", "php-fpm", "php-fpm.sock")
	helpers.AssertFileContains(t, projectConfig, `phpfpm: `+expectedSocketPath)
	helpers.AssertFileContains(t, projectConfig, `socket: `+expectedSocketPath)

	// Verify the socket path doesn't contain the old PHP version
	oldSocketPath := filepath.Join(workspace, "php", "8.3", "runtime", "php-fpm", "php-fpm.sock")
	data, err := os.ReadFile(projectConfig)
	if err != nil {
		t.Fatalf("read %s: %v", projectConfig, err)
	}
	if strings.Contains(string(data), oldSocketPath) {
		t.Fatalf("expected %s not to contain %q", projectConfig, oldSocketPath)
	}
}
