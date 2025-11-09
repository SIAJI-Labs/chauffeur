package purge_test

import (
	"os"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestUninstallPurgeRemovesWorkspace(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	if err := commands.RunUninstall([]string{"--purge"}); err != nil {
		t.Fatalf("RunUninstall --purge failed: %v", err)
	}

	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("expected workspace %s removed, err=%v", workspace, err)
	}
}
