package phprun_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestRunPHPWithoutArgsExecutesDefaultBinary(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	logFile := filepath.Join(workspace, "php-run.log")
	script := "#!/usr/bin/env bash\necho ran >> " + logFile + "\n"
	helpers.WriteExecutable(t, filepath.Join(workspace, "bin", "php"), script)

	if err := commands.RunPHP(nil); err != nil {
		t.Fatalf("RunPHP failed: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if string(data) == "" {
		t.Fatalf("expected php shim to run, log empty")
	}
}
