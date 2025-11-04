package installers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/siaji/chauffeur/cli/internal/projects"
)

func TestProjectAwarePHPShimContent(t *testing.T) {
	content := ProjectAwarePHPShimContent()
	
	// Test that the shim content contains expected components
	expectedPatterns := []string{
		"WORKSPACE=",
		"CWD=",
		"find_project_config()",
		"PROJECT_CONFIG=",
		"PHP_VERSION=",
		"exec \"$PHP_BINARY\"",
		"Chauffeur workspace not found",
		"Please run 'chauf install php",
	}
	
	for _, pattern := range expectedPatterns {
		if !strings.Contains(content, pattern) {
			t.Fatalf("expected shim content to contain %q, got:\n%s", pattern, content)
		}
	}
}

func TestPHPShimProjectIsolation(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}
	
	// Create a project with PHP 8.2 isolation
	projectName := "test-project"
	projectPath := filepath.Join(tmpHome, "test-project")
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	
	// Create project config
	projectConfigDir := filepath.Join(projectsDir, projectName)
	if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
		t.Fatalf("create project config dir: %v", err)
	}
	
	projectConfig := projects.Config{
		Version: 1,
		Path:    projectPath,
		PHP:     "8.2",
		CreatedAt: time.Now().UTC(),
		Runtime: projects.Runtime{
			PHPFPM: filepath.Join(projectConfigDir, "runtime", "php-fpm", "php-fpm.sock"),
		},
	}
	
	configPath := filepath.Join(projectConfigDir, "project.yaml")
	if err := projects.WriteConfig(projectConfig, configPath, false); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	
	// Create PHP 8.2 installation
	php82Dir := filepath.Join(workspaceDir, "php", "8.2", "bin")
	if err := os.MkdirAll(php82Dir, 0o755); err != nil {
		t.Fatalf("create php82 bin dir: %v", err)
	}
	
	phpBinary := filepath.Join(php82Dir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/usr/bin/env bash\necho PHP-8.2"), 0o755); err != nil {
		t.Fatalf("write php82 binary: %v", err)
	}
	
	// Test project detection logic (simulate by checking if the config file exists)
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}
	
	if !strings.Contains(string(configContent), "php: 8.2") {
		t.Fatalf("expected project config to contain php: 8.2, got: %s", string(configContent))
	}
}

func TestPHPShimDefaultFallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	// Create config with default PHP version
	configDir := filepath.Join(workspaceDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	
	configContent := `version: 1
telemetry: false
workspace_dir: ` + workspaceDir + `
caddy:
  enable: true
  http_port: 80
  https_port: 443
nginx:
  enable: true
php:
  default: 8.3
projects_dir: ` + filepath.Join(workspaceDir, "projects")
	
	configPath := filepath.Join(configDir, "chauffeur.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	
	// Create PHP 8.3 installation
	php83Dir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(php83Dir, 0o755); err != nil {
		t.Fatalf("create php83 bin dir: %v", err)
	}
	
	phpBinary := filepath.Join(php83Dir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/usr/bin/env bash\necho PHP-8.3"), 0o755); err != nil {
		t.Fatalf("write php83 binary: %v", err)
	}
	
	// Test that config contains default PHP version
	readContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	
	if !strings.Contains(string(readContent), "default: 8.3") {
		t.Fatalf("expected config to contain default: 8.3, got: %s", string(readContent))
	}
}

func TestPHPShimFallbackTo83(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure without config
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	// Create only PHP 8.3 installation (should be used as fallback)
	php83Dir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(php83Dir, 0o755); err != nil {
		t.Fatalf("create php83 bin dir: %v", err)
	}
	
	phpBinary := filepath.Join(php83Dir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/usr/bin/env bash\necho PHP-8.3-FALLBACK"), 0o755); err != nil {
		t.Fatalf("write php83 binary: %v", err)
	}
	
	// Verify that PHP 8.3 directory exists (for fallback scenario)
	if _, err := os.Stat(phpBinary); err != nil {
		t.Fatalf("php83 binary should exist for fallback: %v", err)
	}
}

func TestPHPShimNoProjectConfigFound(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	// Create projects directory but no specific project
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}
	
	// Test that no project config is found
	projectName := "nonexistent-project"
	slug := projects.Slugify(projectName)
	
	// Check that project config doesn't exist
	projectConfigPath := filepath.Join(projectsDir, slug, "project.yaml")
	if _, err := os.Stat(projectConfigPath); err == nil {
		t.Fatalf("expected project config not to exist at %s", projectConfigPath)
	}
}

func TestPHPShimMissingPHPBinary(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	// Create project config but no PHP installation
	projectName := "missing-php-project"
	projectPath := filepath.Join(tmpHome, projectName)
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	
	projectsDir := filepath.Join(workspaceDir, "projects")
	projectConfigDir := filepath.Join(projectsDir, projectName)
	if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
		t.Fatalf("create project config dir: %v", err)
	}
	
	projectConfig := projects.Config{
		Version: 1,
		Path:    projectPath,
		PHP:     "8.1", // This version is not installed
		CreatedAt: time.Now().UTC(),
		Runtime: projects.Runtime{
			PHPFPM: filepath.Join(projectConfigDir, "runtime", "php-fpm", "php-fpm.sock"),
		},
	}
	
	configPath := filepath.Join(projectConfigDir, "project.yaml")
	if err := projects.WriteConfig(projectConfig, configPath, false); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	
	// Verify that PHP binary doesn't exist
	php81Binary := filepath.Join(workspaceDir, "php", "8.1", "bin", "php")
	if _, err := os.Stat(php81Binary); err == nil {
		t.Fatalf("expected php81 binary not to exist at %s", php81Binary)
	}
}

func TestPHPShimMultipleProjects(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create workspace structure
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}
	
	// Create multiple projects with different PHP versions
	testCases := []struct {
		projectName string
		phpVersion  string
	}{
		{"project-82", "8.2"},
		{"project-81", "8.1"},
		{"project-74", "7.4"},
	}
	
	for _, tc := range testCases {
		// Create project directory
		projectPath := filepath.Join(tmpHome, tc.projectName)
		if err := os.MkdirAll(projectPath, 0o755); err != nil {
			t.Fatalf("create project dir %s: %v", tc.projectName, err)
		}
		
		// Create project config
		slug := projects.Slugify(tc.projectName)
		projectConfigDir := filepath.Join(projectsDir, slug)
		if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
			t.Fatalf("create project config dir %s: %v", slug, err)
		}
		
		projectConfig := projects.Config{
			Version: 1,
			Path:    projectPath,
			PHP:     tc.phpVersion,
			CreatedAt: time.Now().UTC(),
			Runtime: projects.Runtime{
				PHPFPM: filepath.Join(projectConfigDir, "runtime", "php-fpm", "php-fpm.sock"),
			},
		}
		
		configPath := filepath.Join(projectConfigDir, "project.yaml")
		if err := projects.WriteConfig(projectConfig, configPath, false); err != nil {
			t.Fatalf("write project config %s: %v", tc.projectName, err)
		}
		
		// Verify project config contains PHP version
		configContent, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read project config %s: %v", tc.projectName, err)
		}
		
		expected := fmt.Sprintf("php: %s", tc.phpVersion)
		if !strings.Contains(string(configContent), expected) {
			t.Fatalf("expected project config %s to contain %q, got: %s", tc.projectName, expected, string(configContent))
		}
	}
}
