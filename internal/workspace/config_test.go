package workspace_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siegg/chauffeur/internal/workspace"
)

func TestDefaultConfigYAML(t *testing.T) {
	yaml := workspace.DefaultConfigYAML("/tmp/chauf")

	checks := []string{
		"workspace: /tmp/chauf",
		"http_port: 8080",
		"https_port: 8443",
		`default_version: "8.3"`,
		"tld: test",
		"enabled: true",
		`version: "2"`,
	}
	for _, want := range checks {
		if !strings.Contains(yaml, want) {
			t.Errorf("DefaultConfigYAML missing %q", want)
		}
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Point at a non-existent dir so Load falls back to defaults.
	t.Setenv("CHAUFFEUR_HOME", "/tmp/chauf-nonexistent-xyz")
	cfg := workspace.Load()

	if cfg.Nginx.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080", cfg.Nginx.HTTPPort)
	}
	if cfg.Nginx.HTTPSPort != 8443 {
		t.Errorf("HTTPSPort = %d, want 8443", cfg.Nginx.HTTPSPort)
	}
	if cfg.PHP.DefaultVersion != "8.3" {
		t.Errorf("PHP.DefaultVersion = %q, want 8.3", cfg.PHP.DefaultVersion)
	}
	if cfg.DNS.TLD != "test" {
		t.Errorf("DNS.TLD = %q, want test", cfg.DNS.TLD)
	}
	if !cfg.DNS.Enabled {
		t.Error("DNS.Enabled should be true by default")
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", dir)

	configDir := filepath.Join(dir, "config")
	os.MkdirAll(configDir, 0755)

	yaml := `workspace: ` + dir + `
nginx:
  http_port: 9080
  https_port: 9443
php:
  default_version: "8.1"
dns:
  tld: local
  enabled: false
logging:
  level: debug
  max_size_mb: 5
version: "2"
created_at: "2026-01-01T00:00:00Z"
`
	os.WriteFile(filepath.Join(configDir, "chauffeur.yaml"), []byte(yaml), 0644)

	cfg := workspace.Load()

	if cfg.Nginx.HTTPPort != 9080 {
		t.Errorf("HTTPPort = %d, want 9080", cfg.Nginx.HTTPPort)
	}
	if cfg.Nginx.HTTPSPort != 9443 {
		t.Errorf("HTTPSPort = %d, want 9443", cfg.Nginx.HTTPSPort)
	}
	if cfg.PHP.DefaultVersion != "8.1" {
		t.Errorf("PHP.DefaultVersion = %q, want 8.1", cfg.PHP.DefaultVersion)
	}
	if cfg.DNS.TLD != "local" {
		t.Errorf("DNS.TLD = %q, want local", cfg.DNS.TLD)
	}
	if cfg.DNS.Enabled {
		t.Error("DNS.Enabled should be false")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", cfg.Logging.Level)
	}
	if cfg.Logging.MaxSizeMB != 5 {
		t.Errorf("Logging.MaxSizeMB = %d, want 5", cfg.Logging.MaxSizeMB)
	}
}
