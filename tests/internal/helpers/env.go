package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SetupTestHome resets HOME to a temp dir and returns both HOME and workspace paths.
func SetupTestHome(t *testing.T) (home string, workspace string) {
	t.Helper()

	home = t.TempDir()
	t.Setenv("HOME", home)
	workspace = filepath.Join(home, ".chauffeur")

	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	return home, workspace
}

// EnsureFakePHP installs a fake php binary for the requested version so commands think it's installed.
func EnsureFakePHP(t *testing.T, workspace, version string) {
	t.Helper()

	binDir := filepath.Join(workspace, "php", version, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("create fake php dir: %v", err)
	}

	phpPath := filepath.Join(binDir, "php")
	script := []byte("#!/usr/bin/env bash\nexit 0\n")
	if err := os.WriteFile(phpPath, script, 0o755); err != nil {
		t.Fatalf("write fake php binary: %v", err)
	}
}

// WriteExecutable creates a simple executable shell script at the given path.
func WriteExecutable(t *testing.T, path string, script string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if !strings.HasSuffix(script, "\n") {
		script += "\n"
	}
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}

// Chdir changes directory for the duration of the test and restores it afterwards.
func Chdir(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}
