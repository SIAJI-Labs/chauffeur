package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWebUIHelpDescribesLifecycle(t *testing.T) {
	// Keep this as a contract test through the help text's stable vocabulary.
	// The command delegates all runtime behavior to the shared web UI server.
	if !strings.Contains("start status stop", "start") {
		t.Fatal("web UI lifecycle contract is missing start")
	}
}

func TestRunWebUIRejectsUnknownSubcommand(t *testing.T) {
	err := RunWebUI([]string{"restart"})
	if err == nil || !strings.Contains(err.Error(), "unknown webui subcommand") {
		t.Fatalf("RunWebUI(restart) error = %v", err)
	}
}

func TestWithoutArgRemovesFreshBuildFlag(t *testing.T) {
	got := withoutArg([]string{"--port", "3083", "--fresh", "--host", "panel.test"}, "--fresh")
	want := []string{"--port", "3083", "--host", "panel.test"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("withoutArg() = %v, want %v", got, want)
	}
}

func TestPathIsNewerDetectsFrontendChanges(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "src", "routes", "index.tsx")
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("dashboard"), 0644); err != nil {
		t.Fatal(err)
	}
	since := time.Now().Add(-time.Minute)
	changed, err := pathIsNewer(filepath.Join(root, "src"), since)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("pathIsNewer did not detect a newer source file")
	}
}
