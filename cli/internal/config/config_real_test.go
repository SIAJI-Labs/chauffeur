package config

import (
	"os"
	"path/filepath"
	"testing"
)

// helpers for tests
func withTempHome(t *testing.T) (cleanup func()) {
	t.Helper()
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	return func() { os.Setenv("HOME", orig) }
}

func TestLoadCreatesDefaultWhenMissing(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}
	if cfg.Nginx.HTTPPort != 8080 || cfg.PHP.Default == "" {
		t.Fatalf("defaults not applied: %+v", cfg)
	}

	// config file should now exist after Save
	path, err := filePath()
	if err != nil {
		t.Fatalf("filePath error: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected Load not to create file automatically")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig error: %v", err)
	}
	cfg.Nginx.HTTPPort = 9090
	cfg.PHP.Default = "8.2"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.Nginx.HTTPPort != 9090 || loaded.PHP.Default != "8.2" {
		t.Fatalf("roundtrip mismatch: %+v", loaded)
	}
}

func TestSetDefaultPHPVersion(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	if err := SetDefaultPHPVersion("8.1"); err != nil {
		t.Fatalf("SetDefaultPHPVersion error: %v", err)
	}
	got, err := GetDefaultPHPVersion()
	if err != nil {
		t.Fatalf("GetDefaultPHPVersion error: %v", err)
	}
	if got != "8.1" {
		t.Fatalf("expected 8.1, got %s", got)
	}
}

func TestLocalTarballPathSetAndGet(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	if err := SetLocalTarballPath("8.3", "/tmp/php.tar.gz"); err != nil {
		t.Fatalf("SetLocalTarballPath error: %v", err)
	}
	path, err := GetLocalTarballPath("8.3")
	if err != nil {
		t.Fatalf("GetLocalTarballPath error: %v", err)
	}
	if path != "/tmp/php.tar.gz" {
		t.Fatalf("expected stored path, got %s", path)
	}
}

func TestFilePathUsesWorkspaceDir(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	path, err := filePath()
	if err != nil {
		t.Fatalf("filePath error: %v", err)
	}
	if filepath.Base(path) != "chauffeur.yaml" {
		t.Fatalf("unexpected config filename: %s", path)
	}
}
