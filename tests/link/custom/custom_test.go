package linkcustom_test

import (
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestLinkWithCustomDomainAndSSL(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := helpers.NewProjectDir(t, home, "secure-site")
	helpers.Chdir(t, projectDir)

	args := []string{"--site", "example.test", "--ssl"}
	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "secure-site", "project.yaml")
	helpers.AssertFileContains(t, configPath, "domain: example.test")
	helpers.AssertFileContains(t, configPath, "ssl: true")
}


