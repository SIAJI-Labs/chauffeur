package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetRuntimeEngineUpdatesWorkspaceConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", root)
	if err := os.MkdirAll(filepath.Join(root, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "chauffeur.yaml"), []byte(DefaultConfigYAML(root)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SetRuntimeEngine("podman"); err != nil {
		t.Fatal(err)
	}
	if got := Load().Runtime.Engine; got != "podman" {
		t.Fatalf("runtime engine = %q, want podman", got)
	}
}

func TestSetRuntimeEngineMigratesLegacyConfigWithoutRuntimeSection(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHAUFFEUR_HOME", root)
	if err := os.MkdirAll(filepath.Join(root, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "config", "chauffeur.yaml")
	legacy := "workspace: \"" + root + "\"\nphp:\n  default_version: \"8.3\"\n"
	if err := os.WriteFile(path, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SetRuntimeEngine("podman"); err != nil {
		t.Fatal(err)
	}
	if got := Load().Runtime.Engine; got != "podman" {
		t.Fatalf("runtime engine = %q, want podman", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != legacy+"runtime:\n  engine: podman\n" {
		t.Fatalf("migrated config = %q", data)
	}
}
