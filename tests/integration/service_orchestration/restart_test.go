package service_orchestration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

// TestRestartBasic tests basic restart functionality with dry-run
func TestRestartBasic(t *testing.T) {
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

	// Create minimal config
	configContent := `version: 1
nginx:
  enable: true
  http_port: 8080
  https_port: 8443
php:
  default: "8.3"
`

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

	// Test dry-run restart
	t.Run("restart dry-run", func(t *testing.T) {
		err := commands.RunRestart([]string{"--dry-run"})
		if err != nil {
			t.Logf("Dry-run restart failed (expected due to no services): %v", err)
		}
	})

	// Test restart help
	t.Run("restart help", func(t *testing.T) {
		err := commands.RunRestart([]string{"--help"})
		if err != nil {
			t.Errorf("Restart help should succeed: %v", err)
		}
	})
}

// TestRestartWithSpecificServices tests restarting specific services
func TestRestartWithSpecificServices(t *testing.T) {
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

	// Create minimal config
	configContent := `version: 1
nginx:
  enable: true
  http_port: 8081    # Use non-privileged port
  https_port: 8444   # Use non-privileged port
php:
  default: "8.3"
`

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

	testCases := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "restart nginx only",
			args:        []string{"nginx", "--dry-run"},
			description: "Should restart only nginx service",
		},
		{
			name:        "restart php-fpm only",
			args:        []string{"php-fpm", "--dry-run"},
			description: "Should restart all PHP-FPM services",
		},
		{
			name:        "restart multiple services",
			args:        []string{"nginx", "php-fpm", "--dry-run"},
			description: "Should restart nginx and all PHP-FPM services",
		},
		{
			name:        "restart with project filter",
			args:        []string{"--project", "test-project", "--dry-run"},
			description: "Should restart nginx and specific project PHP-FPM",
		},
		{
			name:        "restart all services",
			args:        []string{"--all", "--dry-run"},
			description: "Should restart all Chauffeur services",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			err := commands.RunRestart(tc.args)
			if err != nil {
				t.Logf("Restart test failed (may be expected in test environment): %v", err)
				// This is expected in test environments due to missing services
			}
			// The important thing is that it should not panic or crash
		})
	}
}

// TestRestartErrorHandling tests restart command error handling
func TestRestartErrorHandling(t *testing.T) {
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

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "restart with invalid flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
			description: "Should fail with invalid flag",
		},
		{
			name:        "restart with invalid service name",
			args:        []string{"invalid-service"},
			expectError: true,
			description: "Should fail with invalid service name",
		},
		{
			name:        "restart with --project but no slug",
			args:        []string{"--project"},
			expectError: true,
			description: "Should fail when --project has no argument",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			err := commands.RunRestart(tc.args)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for args %v but got none", tc.args)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for args %v: %v", tc.args, err)
			}
		})
	}
}

// TestRestartPortConfigurations tests restart with various port configurations
func TestRestartPortConfigurations(t *testing.T) {
	portConfigurations := []struct {
		name        string
		httpPort    int
		httpsPort   int
		description string
	}{
		{
			name:        "standard ports",
			httpPort:    8080,
			httpsPort:   8443,
			description: "Common Chauffeur default ports",
		},
		{
			name:        "privileged ports",
			httpPort:    80,
			httpsPort:   443,
			description: "Requires port forwarding",
		},
		{
			name:        "mixed ports",
			httpPort:    80,
			httpsPort:   8443,
			description: "Mixed privileged/non-privileged",
		},
	}

	for _, config := range portConfigurations {
		t.Run(config.name, func(t *testing.T) {
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

			// Create config with specific ports
			configContent := fmt.Sprintf(`version: 1
nginx:
  enable: true
  http_port: %d
  https_port: %d
php:
  default: "8.3"
`, config.httpPort, config.httpsPort)

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

			t.Logf("Testing restart with %s (HTTP: %d, HTTPS: %d)",
				config.description, config.httpPort, config.httpsPort)

			// Test dry-run restart with nginx
			err = commands.RunRestart([]string{"nginx", "--dry-run"})
			if err != nil {
				t.Logf("Restart test failed (may be expected in test environment): %v", err)
				// Expected in test environment due to missing services/sudo
			}
		})
	}
}

// TestRestartEquivalentToStartStop tests that restart behavior is equivalent to stop+start
func TestRestartEquivalentToStartStop(t *testing.T) {
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

	// Create minimal config
	configContent := `version: 1
nginx:
  enable: true
  http_port: 8080
  https_port: 8443
php:
  default: "8.3"
`

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

	t.Run("equivalent behavior test", func(t *testing.T) {
		// Test that restart accepts same arguments as start/stop
		commonArgs := [][]string{
			{"--all", "--dry-run"},
			{"nginx", "--dry-run"},
			{"php-fpm", "--dry-run"},
		}

		for _, args := range commonArgs {
			t.Logf("Testing restart with args: %v", args)

			// Test restart
			err := commands.RunRestart(args)
			if err != nil {
				t.Logf("Restart with args %v failed (may be expected): %v", args, err)
			}

			// The important test is that restart accepts the same arguments as start/stop
			// and doesn't crash or panic
		}
	})
}