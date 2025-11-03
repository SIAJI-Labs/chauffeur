package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestRunServiceCommandSuccess(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	sbinDir := filepath.Join(tmpHome, ".chauffeur", "nginx", "sbin")
	if err := os.MkdirAll(sbinDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	stub := filepath.Join(sbinDir, "nginx")
	script := "#!/usr/bin/env bash\necho stub nginx \"$@\"\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	output := captureOutput(func() error {
		return commands.RunServiceCommand("nginx", []string{"-t"})
	})

	if !strings.Contains(output, "stub nginx -t") {
		t.Fatalf("expected stub output, got %q", output)
	}
}

func TestRunServiceCommandMissingBinary(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	errOutput := captureError(func() error {
		return commands.RunServiceCommand("nginx", nil)
	})

	if !strings.Contains(errOutput, "not installed") {
		t.Fatalf("expected not installed message, got %q", errOutput)
	}
}
