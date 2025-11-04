package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestLinkCreatesProjectConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectDir := filepath.Join(t.TempDir(), "my-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	output := captureOutput(func() error {
		return commands.RunLink([]string{"--site", "myproj.test", "--ssl"})
	})

	if !strings.Contains(output, "Project linked as") {
		t.Fatalf("expected success output, got %q", output)
	}

	projectsDir := filepath.Join(tmpHome, ".chauffeur", "projects")
	slugDir := filepath.Join(projectsDir, "my-app")
	configFile := filepath.Join(slugDir, "project.yaml")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)
	assertContains(t, content, "version: 1")
	assertContains(t, content, "path: "+projectDir)
	assertContains(t, content, "php: 8.3")
	assertContains(t, content, "site:")
	assertContains(t, content, "domain: myproj.test")
	assertContains(t, content, "ssl: true")
	assertContains(t, content, "runtime:")
	assertContains(t, content, "php_fpm_socket: "+filepath.Join(slugDir, "runtime", "php-fpm", "php-fpm.sock"))
	assertContains(t, content, "created_at: ")
}

func TestLinkRequiresForceToOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectDir := filepath.Join(t.TempDir(), "site")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("first link should succeed, got %v", err)
	}

	err = commands.RunLink(nil)
	if err == nil {
		t.Fatalf("expected error when linking twice without --force")
	}
	if !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("expected force hint, got %v", err)
	}

	if err := commands.RunLink([]string{"--force"}); err != nil {
		t.Fatalf("link with --force should succeed, got %v", err)
	}
}

func TestLinkRequiresSiteForSSL(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectDir := filepath.Join(t.TempDir(), "ssl-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	err = commands.RunLink([]string{"--ssl"})
	if err == nil {
		t.Fatalf("expected error without --site when using --ssl")
	}
	if !strings.Contains(err.Error(), "requires --site") {
		t.Fatalf("expected requires site error, got %v", err)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q to contain %q", haystack, needle)
	}
}
