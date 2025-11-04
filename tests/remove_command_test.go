package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestRemove_MissingArguments(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	output := captureError(func() error {
		return commands.RunRemove([]string{})
	})

	if !strings.Contains(output, "no services specified") {
		t.Errorf("Expected error about missing services, got: %s", output)
	}
}

func TestRemove_UnknownService(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	output := captureError(func() error {
		return commands.RunRemove([]string{"unknown-service"})
	})

	if !strings.Contains(output, "unknown service") {
		t.Errorf("Expected error about unknown service, got: %s", output)
	}
}

func TestRemove_UnknownFlag(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	output := captureError(func() error {
		return commands.RunRemove([]string{"--invalid-flag"})
	})

	if !strings.Contains(output, "unknown flag") {
		t.Errorf("Expected error about unknown flag, got: %s", output)
	}
}

func TestRemove_HelpFlag(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	output := captureOutput(func() error {
		return commands.RunRemove([]string{"--help"})
	})

	if output == "" {
		t.Error("Expected help output, got empty string")
	}

	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage information in help output, got: %s", output)
	}
}

func TestRemove_NotInstalledService(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create workspace but no installed services
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	output := captureOutput(func() error {
		return commands.RunRemove([]string{"nginx"})
	})

	if !strings.Contains(output, "not installed") {
		t.Errorf("Expected warning about service not being installed, got: %s", output)
	}
}

func TestRemove_NoWorkspace(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Don't create workspace
	output := captureError(func() error {
		return commands.RunRemove([]string{"php"})
	})

	if !strings.Contains(output, "workspace not found") {
		t.Errorf("Expected error about missing workspace, got: %s", output)
	}
}

func TestRemove_PHPAllVersions(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock PHP installation
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	phpDir := filepath.Join(workspaceDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}

	// Create fake PHP versions
	for _, version := range []string{"8.3", "8.2", "7.4"} {
		versionDir := filepath.Join(phpDir, version, "bin")
		if err := os.MkdirAll(versionDir, 0755); err != nil {
			t.Fatalf("Failed to create PHP version directory: %v", err)
		}
		phpBinary := filepath.Join(versionDir, "php")
		if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho PHP "+version), 0755); err != nil {
			t.Fatalf("Failed to create fake PHP binary: %v", err)
		}
	}

	// Test with --force to skip confirmation
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"php", "--force"})
	})

	if !strings.Contains(output, "Removed all PHP installations") {
		t.Errorf("Expected success message about removing all PHP, got: %s", output)
	}

	// Verify PHP directory was removed
	if _, err := os.Stat(phpDir); !os.IsNotExist(err) {
		t.Error("Expected PHP directory to be removed")
	}
}

func TestRemove_SpecificPHPVersion(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock PHP installation
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	phpDir := filepath.Join(workspaceDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}

	// Create specific PHP version
	version := "8.3"
	versionDir := filepath.Join(phpDir, version, "bin")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP version directory: %v", err)
	}
	phpBinary := filepath.Join(versionDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho PHP "+version), 0755); err != nil {
		t.Fatalf("Failed to create fake PHP binary: %v", err)
	}

	// Test with --force to skip confirmation
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"php", version, "--force"})
	})

	if !strings.Contains(output, fmt.Sprintf("Removed PHP version (PHP %s)", version)) {
		t.Errorf("Expected success message about removing PHP %s, got: %s", version, output)
	}

	// Verify specific version directory was removed
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Error("Expected specific PHP version directory to be removed")
	}

	// Verify PHP parent directory still exists
	if _, err := os.Stat(phpDir); os.IsNotExist(err) {
		t.Error("Expected PHP parent directory to still exist when only one version is removed")
	}
}

func TestRemove_SpecificPHPVersionNotInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create PHP directory but no specific version
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	phpDir := filepath.Join(workspaceDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}

	version := "8.3"
	output := captureError(func() error {
		return commands.RunRemove([]string{"php", version, "--force"})
	})

	if !strings.Contains(output, fmt.Sprintf("PHP %s is not installed", version)) {
		t.Errorf("Expected error about PHP version not being installed, got: %s", output)
	}
}

func TestRemove_NginxService(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock nginx installation
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	nginxDir := filepath.Join(workspaceDir, "nginx", "sbin")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx directory: %v", err)
	}

	nginxBinary := filepath.Join(nginxDir, "nginx")
	if err := os.WriteFile(nginxBinary, []byte("#!/bin/bash\necho nginx"), 0755); err != nil {
		t.Fatalf("Failed to create fake nginx binary: %v", err)
	}

	// Test with --force to skip confirmation
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"nginx", "--force"})
	})

	if !strings.Contains(output, "Removed service (nginx") {
		t.Errorf("Expected success message about removing nginx, got: %s", output)
	}

	// Verify nginx directory was removed
	nginxParentDir := filepath.Join(workspaceDir, "nginx")
	if _, err := os.Stat(nginxParentDir); !os.IsNotExist(err) {
		t.Error("Expected nginx directory to be removed")
	}
}

func TestRemove_CaddyService(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock caddy installation
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	caddyDir := filepath.Join(workspaceDir, "caddy", "bin")
	if err := os.MkdirAll(caddyDir, 0755); err != nil {
		t.Fatalf("Failed to create caddy directory: %v", err)
	}

	caddyBinary := filepath.Join(caddyDir, "caddy")
	if err := os.WriteFile(caddyBinary, []byte("#!/bin/bash\necho caddy"), 0755); err != nil {
		t.Fatalf("Failed to create fake caddy binary: %v", err)
	}

	// Test with --force to skip confirmation
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"caddy", "--force"})
	})

	if !strings.Contains(output, "Removed service (caddy") {
		t.Errorf("Expected success message about removing caddy, got: %s", output)
	}

	// Verify caddy directory was removed
	caddyParentDir := filepath.Join(workspaceDir, "caddy")
	if _, err := os.Stat(caddyParentDir); !os.IsNotExist(err) {
		t.Error("Expected caddy directory to be removed")
	}
}

func TestRemove_MultipleServices(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock caddy and nginx installations
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	services := []struct {
		name string
		path string
	}{
		{"caddy", filepath.Join(workspaceDir, "caddy", "bin", "caddy")},
		{"nginx", filepath.Join(workspaceDir, "nginx", "sbin", "nginx")},
	}

	for _, service := range services {
		serviceDir := filepath.Dir(service.path)
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			t.Fatalf("Failed to create %s directory: %v", service.name, err)
		}
		if err := os.WriteFile(service.path, []byte("#!/bin/bash\necho "+service.name), 0755); err != nil {
			t.Fatalf("Failed to create fake %s binary: %v", service.name, err)
		}
	}

	// Test removing multiple services with --force
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"caddy", "nginx", "--force"})
	})

	if !strings.Contains(output, "Removed service (caddy") {
		t.Errorf("Expected success message about removing caddy, got: %s", output)
	}

	if !strings.Contains(output, "Removed service (nginx") {
		t.Errorf("Expected success message about removing nginx, got: %s", output)
	}

	// Verify both services were removed
	for _, service := range services {
		serviceParentDir := filepath.Join(workspaceDir, service.name)
		if _, err := os.Stat(serviceParentDir); !os.IsNotExist(err) {
			t.Errorf("Expected %s directory to be removed", service.name)
		}
	}
}

func TestRemove_PHPVersionParsing(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock PHP installation
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	phpDir := filepath.Join(workspaceDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}

	version := "8.3"
	versionDir := filepath.Join(phpDir, version, "bin")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP version directory: %v", err)
	}
	phpBinary := filepath.Join(versionDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho PHP "+version), 0755); err != nil {
		t.Fatalf("Failed to create fake PHP binary: %v", err)
	}

	// Test version parsing: arguments should be ["php", "8.3", "--force"]
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"php", version, "--force"})
	})

	if !strings.Contains(output, fmt.Sprintf("Removed PHP version (PHP %s)", version)) {
		t.Errorf("Expected success message about removing PHP %s, got: %s", version, output)
	}
}

func TestRemove_PHPVersionWithOtherServices(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock PHP and caddy installations
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	phpVersionDir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	caddyDir := filepath.Join(workspaceDir, "caddy", "bin")

	if err := os.MkdirAll(phpVersionDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}
	if err := os.MkdirAll(caddyDir, 0755); err != nil {
		t.Fatalf("Failed to create caddy directory: %v", err)
	}

	// Create fake binaries
	if err := os.WriteFile(filepath.Join(phpVersionDir, "php"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create fake PHP binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(caddyDir, "caddy"), []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("Failed to create fake caddy binary: %v", err)
	}

	// Remove PHP 8.3 and caddy
	output := captureOutput(func() error {
		return commands.RunRemove([]string{"php", "8.3", "caddy", "--force"})
	})

	if !strings.Contains(output, "Removed PHP version (PHP 8.3)") {
		t.Errorf("Expected success message about removing PHP 8.3, got: %s", output)
	}

	if !strings.Contains(output, "Removed service (caddy") {
		t.Errorf("Expected success message about removing caddy, got: %s", output)
	}

	// Verify PHP version and caddy were removed, but PHP parent dir remains
	if _, err := os.Stat(phpVersionDir); !os.IsNotExist(err) {
		t.Error("Expected PHP 8.3 directory to be removed")
	}

	if _, err := os.Stat(caddyDir); !os.IsNotExist(err) {
		t.Error("Expected caddy directory to be removed")
	}

	phpParentDir := filepath.Join(workspaceDir, "php")
	if _, err := os.Stat(phpParentDir); os.IsNotExist(err) {
		t.Error("Expected PHP parent directory to still exist")
	}
}
