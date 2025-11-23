package example

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
//time"

	"github.com/siaji/chauffeur/cli/internal/projects"
)

func TestCreateExampleProject(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create config file with structure expected by Load()
	configDir := filepath.Join(wsDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configFile := filepath.Join(configDir, "chauffeur.yaml")
	configContent := fmt.Sprintf(`workspace_dir: %s
projects_dir: %s
php:
  default: "8.3"
`, wsDir, projectsDir)
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create mock PHP installation
	phpDir := filepath.Join(wsDir, "php", "8.3")
	phpBin := filepath.Join(phpDir, "bin", "php")
	if err := os.MkdirAll(filepath.Dir(phpBin), 0755); err != nil {
		t.Fatalf("Failed to create PHP dir: %v", err)
	}
	if err := os.WriteFile(phpBin, []byte("#!/bin/bash\necho 'PHP 8.3'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock PHP: %v", err)
	}
	
	// Create mock nginx installation
	nginxDir := filepath.Join(wsDir, "nginx")
	nginxBin := filepath.Join(nginxDir, "bin", "nginx")
	if err := os.MkdirAll(filepath.Dir(nginxBin), 0755); err != nil {
		t.Fatalf("Failed to create nginx dir: %v", err)
	}
	if err := os.WriteFile(nginxBin, []byte("#!/bin/bash\necho 'nginx'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock nginx: %v", err)
	}
	
	// Create example project
	if err := CreateExampleProject(); err != nil {
		t.Fatalf("Failed to create example project: %v", err)
	}
	
	// Verify project directory exists
	exampleDir := filepath.Join(projectsDir, ExampleProjectName)
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Errorf("Example project directory was not created")
	}
	
	// Verify index.php exists
	indexPath := filepath.Join(exampleDir, "index.php")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("index.php was not created")
	}
	
	// Verify .gitignore exists
	gitignorePath := filepath.Join(exampleDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Errorf(".gitignore was not created")
	}
}

func TestLinkExampleProjectIfReady(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create config file
	configDir := filepath.Join(wsDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configFile := filepath.Join(configDir, "chauffeur.yaml")
	configContent := fmt.Sprintf(`workspace_dir: %s
projects_dir: %s
php:
  default: "8.3"
`, wsDir, projectsDir)
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create mock PHP installation
	phpDir := filepath.Join(wsDir, "php", "8.3")
	phpBin := filepath.Join(phpDir, "bin", "php")
	if err := os.MkdirAll(filepath.Dir(phpBin), 0755); err != nil {
		t.Fatalf("Failed to create PHP dir: %v", err)
	}
	if err := os.WriteFile(phpBin, []byte("#!/bin/bash\necho 'PHP 8.3'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock PHP: %v", err)
	}
	
	// Create mock nginx installation
	nginxDir := filepath.Join(wsDir, "nginx")
	nginxBin := filepath.Join(nginxDir, "bin", "nginx")
	if err := os.MkdirAll(filepath.Dir(nginxBin), 0755); err != nil {
		t.Fatalf("Failed to create nginx dir: %v", err)
	}
	if err := os.WriteFile(nginxBin, []byte("#!/bin/bash\necho 'nginx'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock nginx: %v", err)
	}
	
	// Test linking when both services are installed
	if err := LinkExampleProjectIfReady(); err != nil {
		t.Fatalf("Failed to link example project: %v", err)
	}
	
	// Verify project is linked
	projectPath := filepath.Join(projectsDir, ExampleProjectName)
	if _, _, err := projects.FindByPath(projectsDir, projectPath); err != nil {
		t.Errorf("Example project was not linked")
	}
	
	// Verify nginx config exists
	nginxConfigPath := filepath.Join(wsDir, "nginx", "sites-enabled", ExampleProjectName+".conf")
	if _, err := os.Stat(nginxConfigPath); os.IsNotExist(err) {
		t.Errorf("Nginx config was not created")
	}
}

func TestLinkExampleProjectIfReady_PHPNotInstalled(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create config file
	configDir := filepath.Join(wsDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configFile := filepath.Join(configDir, "chauffeur.yaml")
	configContent := fmt.Sprintf(`workspace_dir: %s
projects_dir: %s
php:
  default: "8.3"
`, wsDir, projectsDir)
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create mock nginx installation only (PHP not installed)
	nginxDir := filepath.Join(wsDir, "nginx")
	nginxBin := filepath.Join(nginxDir, "bin", "nginx")
	if err := os.MkdirAll(filepath.Dir(nginxBin), 0755); err != nil {
		t.Fatalf("Failed to create nginx dir: %v", err)
	}
	if err := os.WriteFile(nginxBin, []byte("#!/bin/bash\necho 'nginx'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock nginx: %v", err)
	}
	
	// Test linking when PHP is not installed
	if err := LinkExampleProjectIfReady(); err != nil {
		t.Fatalf("Function should not return error when PHP not installed: %v", err)
	}
	
	// Verify project is created but not linked
	projectPath := filepath.Join(projectsDir, ExampleProjectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Errorf("Example project directory should be created even when PHP not installed")
	}
	
	// Verify project is not linked
	if _, _, err := projects.FindByPath(projectsDir, projectPath); err == nil {
		t.Errorf("Example project should not be linked when PHP is not installed")
	}
}

func TestRemoveExampleProject(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create config file
	configDir := filepath.Join(wsDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configFile := filepath.Join(configDir, "chauffeur.yaml")
	configContent := fmt.Sprintf(`workspace_dir: %s
projects_dir: %s
php:
  default: "8.3"
`, wsDir, projectsDir)
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create mock PHP installation
	phpDir := filepath.Join(wsDir, "php", "8.3")
	phpBin := filepath.Join(phpDir, "bin", "php")
	if err := os.MkdirAll(filepath.Dir(phpBin), 0755); err != nil {
		t.Fatalf("Failed to create PHP dir: %v", err)
	}
	if err := os.WriteFile(phpBin, []byte("#!/bin/bash\necho 'PHP 8.3'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock PHP: %v", err)
	}
	
	// Create and link example project first
	if err := CreateExampleProject(); err != nil {
		t.Fatalf("Failed to create example project: %v", err)
	}
	
	if err := LinkExampleProjectIfReady(); err != nil {
		t.Fatalf("Failed to link example project: %v", err)
	}
	
	// Verify project exists before removal
	exampleDir := filepath.Join(projectsDir, ExampleProjectName)
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Errorf("Example project should exist before removal")
	}
	
	// Remove example project
	if err := RemoveExampleProject(); err != nil {
		t.Fatalf("Failed to remove example project: %v", err)
	}
	
	// Verify project is removed
	if _, err := os.Stat(exampleDir); err == nil {
		t.Errorf("Example project should be removed")
	}
}

func TestIsExampleProjectExists(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Test when project doesn't exist
	if IsExampleProjectExists() {
		t.Errorf("Should return false when example project doesn't exist")
	}
	
	// Create example project directory
	exampleDir := filepath.Join(projectsDir, ExampleProjectName)
	if err := os.MkdirAll(exampleDir, 0755); err != nil {
		t.Fatalf("Failed to create example project: %v", err)
	}
	
	// Test when project exists
	if !IsExampleProjectExists() {
		t.Errorf("Should return true when example project exists")
	}
}

func TestIsExampleProjectLinked(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpDir)
	
	// Initialize workspace
	wsDir := filepath.Join(tmpDir, ".chauffeur")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	
	// Create projects directory
	projectsDir := filepath.Join(wsDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create config file
	configDir := filepath.Join(wsDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configFile := filepath.Join(configDir, "chauffeur.yaml")
	configContent := fmt.Sprintf(`workspace_dir: %s
projects_dir: %s
php:
  default: "8.3"
`, wsDir, projectsDir)
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create mock PHP installation
	phpDir := filepath.Join(wsDir, "php", "8.3")
	phpBin := filepath.Join(phpDir, "bin", "php")
	if err := os.MkdirAll(filepath.Dir(phpBin), 0755); err != nil {
		t.Fatalf("Failed to create PHP dir: %v", err)
	}
	if err := os.WriteFile(phpBin, []byte("#!/bin/bash\necho 'PHP 8.3'\n"), 0755); err != nil {
		t.Fatalf("Failed to create mock PHP: %v", err)
	}
	
	// Test when project is not linked
	if IsExampleProjectLinked() {
		t.Errorf("Should return false when example project is not linked")
	}
	
	// Create and link example project
	if err := LinkExampleProjectIfReady(); err != nil {
		t.Fatalf("Failed to link example project: %v", err)
	}
	
	// Test when project is linked
	if !IsExampleProjectLinked() {
		t.Errorf("Should return true when example project is linked")
	}
}
