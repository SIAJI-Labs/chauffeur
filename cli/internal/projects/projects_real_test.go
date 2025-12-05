package projects

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureLayoutCreatesDirectories(t *testing.T) {
	base := t.TempDir()
	layout, err := EnsureLayout(base, "demo")
	if err != nil {
		t.Fatalf("EnsureLayout error: %v", err)
	}

	for _, path := range []string{layout.Root, layout.RuntimeDir, layout.LogsDir} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected path to exist: %s", path)
		}
	}
	if filepath.Base(layout.ConfigPath) != "project.yaml" {
		t.Fatalf("unexpected config path: %s", layout.ConfigPath)
	}
}

func TestWriteLoadAndFindProjectConfig(t *testing.T) {
	base := t.TempDir()
	layout, err := EnsureLayout(base, "demo")
	if err != nil {
		t.Fatalf("EnsureLayout error: %v", err)
	}

	cfg := Config{
		Path:      "/tmp/demo",
		PHP:       "8.3",
		Site:      &Site{Domain: "demo.test", SSL: true},
		Domains:   &Domains{Aliases: []DomainAlias{{Domain: "admin.test", SSL: false}}},
		Runtime:   Runtime{PHPFPM: filepath.Join(layout.RuntimeDir, "php-fpm.sock"), FPM: &FPM{Dedicated: true, Socket: filepath.Join(layout.RuntimeDir, "php-fpm.sock")}},
		CreatedAt: time.Now(),
	}

	if err := WriteConfig(cfg, layout.ConfigPath, true); err != nil {
		t.Fatalf("WriteConfig error: %v", err)
	}

	loaded, err := LoadConfig(layout.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if loaded.Site.Domain != "demo.test" || loaded.PHP != "8.3" {
		t.Fatalf("loaded config mismatch: %+v", loaded)
	}

	found, foundLayout, err := FindByPath(base, cfg.Path)
	if err != nil {
		t.Fatalf("FindByPath error: %v", err)
	}
	if found.Runtime.PHPFPM == "" || foundLayout.SocketPath == "" {
		t.Fatalf("expected runtime paths to be set: %+v", found)
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Hello World":     "hello-world",
		"   Trim   ":      "trim",
		"!!!":             "project",
		"MiXeD_Case.Name": "mixed-case-name",
		"double--dash":    "double-dash",
	}
	for input, expected := range tests {
		if got := Slugify(input); got != expected {
			t.Fatalf("Slugify(%q)=%q, want %q", input, got, expected)
		}
	}
}
