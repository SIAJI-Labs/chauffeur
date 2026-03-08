package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siegg/chauffeur/internal/commands"
)

func TestRunInit_CreatesStructure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	expectedDirs := []string{
		"bin/shims",
		"config",
		"projects",
		"php",
		"nginx/etc/sites-available",
		"nginx/etc/sites-enabled",
		"nginx/certs",
		"nginx/logs",
		"composer",
		"cache/php",
		"cache/nginx",
		"cache/composer",
		"logs/commands",
		"system",
	}
	for _, d := range expectedDirs {
		if fi, err := os.Stat(filepath.Join(dir, d)); err != nil || !fi.IsDir() {
			t.Errorf("expected directory %q to exist", d)
		}
	}

	expectedFiles := []string{
		"config/chauffeur.yaml",
		"nginx/etc/nginx.conf",
		"nginx/etc/fastcgi_params",
		"nginx/etc/mime.types",
		"nginx/etc/sites-available/default.conf",
		"bin/shims/php",
		"bin/shims/composer",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected file %q to exist", f)
		}
	}
}

func TestRunInit_ShimsAreExecutable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	for _, shim := range []string{"bin/shims/php", "bin/shims/composer"} {
		fi, err := os.Stat(filepath.Join(dir, shim))
		if err != nil {
			t.Fatalf("stat %s: %v", shim, err)
		}
		if fi.Mode()&0111 == 0 {
			t.Errorf("%s is not executable (mode %s)", shim, fi.Mode())
		}
	}
}

func TestRunInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("first RunInit: %v", err)
	}

	// Record config content after first run.
	configPath := filepath.Join(dir, "config", "chauffeur.yaml")
	firstContent, _ := os.ReadFile(configPath)

	// Second run should not overwrite config.
	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("second RunInit: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)
	if string(firstContent) != string(secondContent) {
		t.Error("second RunInit overwrote existing config (expected idempotent)")
	}
}

func TestRunInit_Force(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("first RunInit: %v", err)
	}

	configPath := filepath.Join(dir, "config", "chauffeur.yaml")

	// Corrupt the config.
	os.WriteFile(configPath, []byte("# corrupted\n"), 0644)

	// --force should regenerate it.
	if err := commands.RunInit([]string{"--quiet", "--force"}); err != nil {
		t.Fatalf("RunInit --force: %v", err)
	}

	content, _ := os.ReadFile(configPath)
	if string(content) == "# corrupted\n" {
		t.Error("--force did not overwrite existing config")
	}
	if len(content) < 50 {
		t.Errorf("regenerated config looks too short: %q", content)
	}
}

func TestRunInit_ConfigContainsWorkspaceRoot(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	if err := commands.RunInit([]string{"--quiet"}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "config", "chauffeur.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	if !contains(string(content), "workspace: "+dir) {
		t.Errorf("config does not contain expected workspace path %q\ncontent:\n%s", dir, content)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
