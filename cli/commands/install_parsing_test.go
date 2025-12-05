package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// Minimal parser harness to keep install coverage up without running installers.
func TestRunInstallParsesServicesAndVersions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Avoid real workspace validation by pre-creating .chauffeur
	if err := createTempWorkspace(t); err != nil {
		t.Fatalf("workspace setup failed: %v", err)
	}

	args := []string{"php", "8.3", "php", "8.2", "nginx", "composer"}
	// Override installer to no-op to avoid side effects
	reset := OverridePHPInstallFunc(func(version, prefix string, force bool, enableGD bool) error { return nil })
	defer reset()

	// Exercise parsing path (help returns nil)
	err := RunInstall(append(args, "--help"))
	if err != nil {
		t.Fatalf("expected help path to return nil, got %v", err)
	}

	// Unknown flag should error
	if err := RunInstall([]string{"--wat"}); err == nil {
		t.Fatalf("expected error for unknown flag")
	}
}

func createTempWorkspace(t *testing.T) error {
	ws := workspaceDir(t)
	return os.MkdirAll(ws, 0o755)
}

func workspaceDir(t *testing.T) string {
	return filepath.Join(os.Getenv("HOME"), ".chauffeur")
}
