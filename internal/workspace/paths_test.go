package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siegg/chauffeur/internal/workspace"
)

func TestRoot_Default(t *testing.T) {
	os.Unsetenv("CHAUFFEUR_HOME")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".chauffeur")
	if got := workspace.Root(); got != want {
		t.Errorf("Root() = %q, want %q", got, want)
	}
}

func TestRoot_EnvOverride(t *testing.T) {
	t.Setenv("CHAUFFEUR_HOME", "/tmp/my-chauf")
	if got := workspace.Root(); got != "/tmp/my-chauf" {
		t.Errorf("Root() = %q, want /tmp/my-chauf", got)
	}
}

func TestPath(t *testing.T) {
	t.Setenv("CHAUFFEUR_HOME", "/tmp/test-root")
	got := workspace.Path("php", "8.3", "bin")
	want := "/tmp/test-root/php/8.3/bin"
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}
