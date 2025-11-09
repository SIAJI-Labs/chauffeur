package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NewProjectDir creates a project directory under home with given name.
func NewProjectDir(t *testing.T, home, name string) string {
	t.Helper()

	projectDir := filepath.Join(home, name)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	return projectDir
}

// AssertFileContains ensures the file at path includes the expected substring.
func AssertFileContains(t *testing.T, path, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), needle) {
		t.Fatalf("expected %s to contain %q, got:\n%s", path, needle, string(data))
	}
}
