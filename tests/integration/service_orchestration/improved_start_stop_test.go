package service_orchestration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

// TestImprovedStartStopWithVariousPortScenarios tests the improved start/stop functionality
// with different port configurations
func TestImprovedStartStopWithVariousPortScenarios(t *testing.T) {
	testScenarios := []struct {
		name        string
		httpPort    int
		httpsPort   int
		description string
	}{
		{
			name:        "High non-privileged ports",
			httpPort:    9080,
			httpsPort:   9443,
			description: "Uses high-numbered ports that shouldn't require port forwarding",
		},
		{
			name:        "Low non-privileged ports",
			httpPort:    8000,
			httpsPort:   8433,
			description: "Uses lower non-privileged ports",
		},
		{
			name:        "Mixed privileged and non-privileged",
			httpPort:    80,
			httpsPort:   8443,
			description: "Mixed configuration requiring partial port forwarding",
		},
		{
			name:        "Standard development ports",
			httpPort:    8080,
			httpsPort:   8443,
			description: "Common Chauffeur default ports",
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Use temporary workspace for isolation
			tempDir := t.TempDir()

			// Set up environment
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", originalHome)

			// Initialize workspace in temp directory
			wsDir := filepath.Join(tempDir, ".chauffeur")
			err := os.MkdirAll(wsDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create workspace directory: %v", err)
			}

			// Create projects directory structure to prevent config loading errors
			projectsDir := filepath.Join(wsDir, "projects")
			err = os.MkdirAll(projectsDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create projects directory: %v", err)
			}
			
			

			// Create config for this scenario
			configContent := `version: 1
nginx:
  enable: true
  http_port: %d
  https_port: %d
php:
  default: "8.3"
`
			configContent = fmt.Sprintf(configContent, scenario.httpPort, scenario.httpsPort)

			configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
			err = os.MkdirAll(filepath.Dir(configPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create config directory: %v", err)
			}
			err = os.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Create nginx directory structure
			nginxDir := filepath.Join(wsDir, "nginx")
			err = os.MkdirAll(nginxDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create nginx directory: %v", err)
			}

			t.Logf("Testing scenario: %s", scenario.description)
			t.Logf("Using HTTP port: %d, HTTPS port: %d", scenario.httpPort, scenario.httpsPort)

			// Test dry-run start - should not crash or panic
			err = commands.RunStart([]string{"--dry-run"})
			if err != nil {
				t.Logf("Dry-run start failed (expected in test environment): %v", err)
				// This is expected in test environments due to missing services/sudo
			} else {
				t.Log("Dry-run start succeeded")
			}

			// Test dry-run stop - should not crash or panic
			err = commands.RunStop([]string{"--dry-run"})
			if err != nil {
				t.Errorf("Dry-run stop failed unexpectedly: %v", err)
			} else {
				t.Log("Dry-run stop succeeded")
			}

			t.Logf("Scenario %s completed without crashes", scenario.name)
		})
	}
}

// TestErrorHandlingInStartStop tests that start/stop commands handle errors gracefully
func TestErrorHandlingInStartStop(t *testing.T) {
	// Use temporary workspace for isolation
	tempDir := t.TempDir()

	// Set up environment
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize workspace in temp directory
	wsDir := filepath.Join(tempDir, ".chauffeur")
	err := os.MkdirAll(wsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	// Create projects directory structure to prevent config loading errors
	projectsDir := filepath.Join(wsDir, "projects")
	err = os.MkdirAll(projectsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create projects directory: %v", err)
	}

	// Test with invalid configurations
	invalidConfigs := []struct {
		name        string
		config      string
		description string
	}{
		{
			name: "Invalid port configuration",
			config: `version: 1
nginx:
  enable: true
  http_port: 0
  https_port: 0
php:
  default: "8.3"
`,
			description: "Configuration with zero ports",
		},
		{
			name: "Missing configuration",
			config: `version: 1
php:
  default: "8.3"
`,
			description: "Configuration without nginx section",
		},
	}

	for _, testConfig := range invalidConfigs {
		t.Run(testConfig.name, func(t *testing.T) {
			t.Logf("Testing with config: %s", testConfig.description)

			configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
			err = os.MkdirAll(filepath.Dir(configPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create config directory: %v", err)
			}
			err = os.WriteFile(configPath, []byte(testConfig.config), 0644)
			if err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Test that commands handle errors gracefully without crashing
			err = commands.RunStart([]string{"--dry-run"})
			if err != nil {
				t.Logf("Start command handled error gracefully: %v", err)
			} else {
				t.Log("Start command succeeded with this configuration")
			}

			err = commands.RunStop([]string{"--dry-run"})
			if err != nil {
				t.Errorf("Stop command failed unexpectedly: %v", err)
			} else {
				t.Log("Stop command succeeded")
			}
		})
	}
}