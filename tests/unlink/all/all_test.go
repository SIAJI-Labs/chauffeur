package unlinkall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestUnlinkAllRemovesAllProjects(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	for _, name := range []string{"site-a", "site-b"} {
		projectDir := helpers.NewProjectDir(t, home, name)
		helpers.Chdir(t, projectDir)
		if err := commands.RunLink(nil); err != nil {
			t.Fatalf("RunLink for %s failed: %v", name, err)
		}
	}

	if err := commands.RunUnlink([]string{"--all", "--force"}); err != nil {
		t.Fatalf("RunUnlink --all failed: %v", err)
	}

	for _, name := range []string{"site-a", "site-b"} {
		projectConfig := filepath.Join(workspace, "projects", name, "project.yaml")
		if _, err := os.Stat(projectConfig); err == nil {
			t.Fatalf("project %s still exists after unlink --all", name)
		}
	}
}
