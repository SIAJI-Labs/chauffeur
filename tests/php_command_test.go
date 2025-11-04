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
	// With project-aware shims, the shim content is static and doesn't contain version references
	// Check that the shim is a proper bash script instead
	if !strings.Contains(string(data), "#!/usr/bin/env bash") {
		t.Fatalf("shim is not a proper bash script")
	}
	if !strings.Contains(string(data), "WORKSPACE=") {
		t.Fatalf("shim doesn't contain project-aware logic")
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

func TestPHPIsolateUpdatesProjectConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install default PHP 8.3 for linking
	defaultPHPBinDir := filepath.Join(tmpHome, ".chauffeur", "php", "8.3", "bin")
	if err := os.MkdirAll(defaultPHPBinDir, 0o755); err != nil {
		t.Fatalf("mkdir default php bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(defaultPHPBinDir, "php"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write default php stub: %v", err)
	}

	version := "7.4"
	phpBinDir := filepath.Join(tmpHome, ".chauffeur", "php", version, "bin")
	if err := os.MkdirAll(phpBinDir, 0o755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(phpBinDir, "php"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write php stub: %v", err)
	}

	projectRoot := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("link project: %v", err)
	}

	projectsDir := filepath.Join(tmpHome, ".chauffeur", "projects")
	configPath := filepath.Join(projectsDir, "proj", "project.yaml")
	initialConfig := readFile(t, configPath)
	initialCreated := lineWithPrefix(initialConfig, "created_at:")

	output := captureOutput(func() error {
		return commands.RunPHP([]string{"isolate", version})
	})

	if !strings.Contains(output, "pinned to "+version) {
		t.Fatalf("expected isolate output to mention version, got %q", output)
	}

	updatedConfig := readFile(t, configPath)
	if !strings.Contains(updatedConfig, "php: "+version) {
		t.Fatalf("expected php version to be updated, got %s", updatedConfig)
	}
	if updatedCreated := lineWithPrefix(updatedConfig, "created_at:"); initialCreated != "" && updatedCreated != initialCreated {
		t.Fatalf("expected created_at to remain unchanged (before=%q, after=%q)", initialCreated, updatedCreated)
	}
}

func TestPHPIsolateRequiresLinkedProject(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	version := "7.4"
	phpBinDir := filepath.Join(tmpHome, ".chauffeur", "php", version, "bin")
	if err := os.MkdirAll(phpBinDir, 0o755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(phpBinDir, "php"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write php stub: %v", err)
	}

	projectRoot := filepath.Join(t.TempDir(), "unlinked")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	errMsg := captureError(func() error {
		return commands.RunPHP([]string{"isolate", version})
	})

	if !strings.Contains(errMsg, "Run 'chauf link'") {
		t.Fatalf("expected instruction to link project, got %q", errMsg)
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(data)
}

func lineWithPrefix(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
