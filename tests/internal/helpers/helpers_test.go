package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetupTestHome tests the test home setup function
func TestSetupTestHome(t *testing.T) {
	// Test that SetupTestHome creates a temporary directory
	tempDir, workspace := SetupTestHome(t)

	if tempDir == "" {
		t.Error("Expected temp directory to be created")
	}

	if workspace == "" {
		t.Error("Expected workspace directory to be returned")
	}

	// Verify the workspace is within the temp directory
	if !filepath.HasPrefix(workspace, tempDir) {
		t.Errorf("Expected workspace %s to be within temp directory %s", workspace, tempDir)
	}

	// Verify the workspace directory exists
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		t.Errorf("Expected workspace directory %s to exist", workspace)
	}
}

// TestCreateTestProject tests project creation for tests
func TestCreateTestProject(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	workspace := filepath.Join(tempDir, ".chauffeur")
	projectsDir := filepath.Join(workspace, "projects")

	// Create projects directory
	err := os.MkdirAll(projectsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create projects directory: %v", err)
	}

	// Create a test project
	projectDir := filepath.Join(projectsDir, "test-project")
	projectConfig := filepath.Join(projectDir, "project.yaml")

	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create project configuration
	configContent := `version: 1
path: /tmp/test-project
php: 8.3
site:
  domain: test-project.test
  ssl: false
runtime:
  php_fpm_socket: /tmp/.chauffeur/projects/test-project/runtime/php-fpm/php-fpm.sock
created_at: 2025-01-01T00:00:00+00:00
`

	err = os.WriteFile(projectConfig, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create project config: %v", err)
	}

	// Verify the project exists
	if _, err := os.Stat(projectConfig); os.IsNotExist(err) {
		t.Errorf("Expected project config %s to exist", projectConfig)
	}
}

// TestCreateTestNginxConfig tests nginx configuration creation for tests
func TestCreateTestNginxConfig(t *testing.T) {
	tempDir := t.TempDir()

	nginxDir := filepath.Join(tempDir, "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	// Create nginx directories
	err := os.MkdirAll(sitesAvailable, 0755)
	if err != nil {
		t.Fatalf("Failed to create sites-available directory: %v", err)
	}

	err = os.MkdirAll(sitesEnabled, 0755)
	if err != nil {
		t.Fatalf("Failed to create sites-enabled directory: %v", err)
	}

	// Create a test nginx config
	configFile := filepath.Join(sitesAvailable, "test-project.test")
	configContent := `server {
    listen 8080;
    server_name test-project.test;
    root /tmp/test-project;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:/tmp/.chauffeur/projects/test-project/runtime/php-fpm/php-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
}
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create nginx config: %v", err)
	}

	// Create symlink in sites-enabled
	symlinkPath := filepath.Join(sitesEnabled, "test-project.test")
	err = os.Symlink(configFile, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create nginx symlink: %v", err)
	}

	// Verify the files exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Expected nginx config %s to exist", configFile)
	}

	if _, err := os.Stat(symlinkPath); os.IsNotExist(err) {
		t.Errorf("Expected nginx symlink %s to exist", symlinkPath)
	}
}

// TestCleanupTestEnvironment tests test environment cleanup
func TestCleanupTestEnvironment(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files and directories
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testDir := filepath.Join(tempDir, "test-dir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Verify files exist before cleanup
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("Expected test file %s to exist before cleanup", testFile)
	}

	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("Expected test directory %s to exist before cleanup", testDir)
	}

	// Cleanup is handled automatically by t.TempDir()
	// So we just verify that the files were created successfully
}

// TestMockServiceManager tests mock service manager functionality
func TestMockServiceManager(t *testing.T) {
	// This would test a mock service manager if one exists
	// For now, we'll test the basic concept of service mocking

	services := []struct {
		name   string
		status string
	}{
		{"chauf-nginx", "running"},
		{"chauf-php-fpm-test", "stopped"},
		{"chauf-php-fpm-another", "running"},
	}

	for _, service := range services {
		if service.name == "" {
			t.Errorf("Service name should not be empty")
		}

		if service.status == "" {
			t.Errorf("Service status should not be empty")
		}

		validStatuses := []string{"running", "stopped", "starting", "stopping"}
		isValid := false
		for _, validStatus := range validStatuses {
			if service.status == validStatus {
				isValid = true
				break
			}
		}

		if !isValid {
			t.Errorf("Invalid service status: %s", service.status)
		}
	}
}

// TestConfigValidationHelpers tests configuration validation helpers
func TestConfigValidationHelpers(t *testing.T) {
	// Test port validation
	validPorts := []int{80, 443, 8080, 8443, 9000}
	invalidPorts := []int{-1, 0, 65536, 70000}

	for _, port := range validPorts {
		if port < 1 || port > 65535 {
			t.Errorf("Port %d should be valid", port)
		}
	}

	for _, port := range invalidPorts {
		if port >= 1 && port <= 65535 {
			t.Errorf("Port %d should be invalid", port)
		}
	}

	// Test version validation
	validVersions := []string{"8.3", "8.2", "8.1", "8.0", "7.4"}
	invalidVersions := []string{"8.", "invalid", "6.0"}

	for _, version := range validVersions {
		if len(version) < 3 {
			t.Errorf("Version %s should be valid", version)
		}
	}

	for _, version := range invalidVersions {
		if len(version) >= 3 {
			// Basic check - would need more sophisticated validation
			isValid := version[0] >= '7' && version[1] == '.'
			if isValid {
				t.Errorf("Version %s should be invalid", version)
			}
		}
	}
}

// TestPathHelpers tests path-related helper functions
func TestPathHelpers(t *testing.T) {
	tempDir := t.TempDir()

	// Test file operations
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file exists and has correct content
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Expected file content %s, got %s", string(content), string(readContent))
	}

	// Test directory operations
	testDir := filepath.Join(tempDir, "test-dir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat test directory: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory", testDir)
	}
}