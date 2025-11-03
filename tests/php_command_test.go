package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestPHPUseSetsDefault(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	version := "8.2"
	binDir := filepath.Join(tmpHome, ".chauffeur", "php", version, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	binary := filepath.Join(binDir, "php")
	if err := os.WriteFile(binary, []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	output := captureOutput(func() error {
		return commands.RunPHP([]string{"use", version})
	})

	if !strings.Contains(output, "Default PHP version updated") {
		t.Fatalf("expected success message, got %q", output)
	}

	if def := readDefaultPHPVersion(t, tmpHome); def != version {
		t.Fatalf("expected default %s, got %s", version, def)
	}

	shimPath := filepath.Join(tmpHome, ".chauffeur", "bin", "php")
	data, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("read shim: %v", err)
	}
	if !strings.Contains(string(data), filepath.Join("php", version, "bin", "php")) {
		t.Fatalf("shim does not reference version %s", version)
	}
}

func TestPHPUseMissingVersion(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	errMsg := captureError(func() error {
		return commands.RunPHP([]string{"use", "8.1"})
	})

	if !strings.Contains(errMsg, "not installed") {
		t.Fatalf("expected not installed message, got %q", errMsg)
	}
}

func TestPHPUseAlreadyDefault(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	version := "8.3"
	binDir := filepath.Join(tmpHome, ".chauffeur", "php", version, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	binary := filepath.Join(binDir, "php")
	if err := os.WriteFile(binary, []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	writeConfigWithDefault(t, tmpHome, version)

	msg := captureOutput(func() error {
		return commands.RunPHP([]string{"use", version})
	})

	if !strings.Contains(msg, "already the default") {
		t.Fatalf("expected already default message, got %q", msg)
	}
}

func TestPHPCommandPassthrough(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	binDir := filepath.Join(tmpHome, ".chauffeur", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	stub := filepath.Join(binDir, "php")
	script := "#!/usr/bin/env bash\necho stub php \"$@\"\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	output := captureOutput(func() error {
		return commands.RunPHP([]string{"-v"})
	})

	if !strings.Contains(output, "stub php -v") {
		t.Fatalf("expected passthrough output, got %q", output)
	}
}

func readDefaultPHPVersion(t *testing.T, home string) string {
	t.Helper()

	cfgFile := filepath.Join(home, ".chauffeur", "config", "chauffeur.yaml")
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inPHP := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "php:" {
			inPHP = true
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inPHP = false
		}
		if inPHP && strings.HasPrefix(trimmed, "default:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan config: %v", err)
	}
	return ""
}

func writeConfigWithDefault(t *testing.T, home, version string) {
	t.Helper()

	root := filepath.Join(home, ".chauffeur")
	cfgDir := filepath.Join(root, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	projectsDir := filepath.Join(root, "projects")
	content := fmt.Sprintf(`version: %d
telemetry: %t
workspace_dir: %s
caddy:
  enable: %t
  http_port: %d
  https_port: %d
nginx:
  enable: %t
php:
  default: %s
projects_dir: %s
`, 1, false, root, true, 80, 443, true, version, projectsDir)

	cfgFile := filepath.Join(cfgDir, "chauffeur.yaml")
	if err := os.WriteFile(cfgFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
