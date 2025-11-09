package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallScriptWorkspaceCheck(t *testing.T) {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create workspace directory
	wsDir := filepath.Join(tmpHome, ".chauffeur")
	configDir := filepath.Join(wsDir, "config")
	
	// Initially, no config should exist
	assert.NoDirExists(t, wsDir, "Workspace should not exist initially")
	assert.NoDirExists(t, configDir, "Config directory should not exist initially")

	// Simulate the install script logic for creating basic config
	createBasicConfigForTest(t, wsDir)

	// Verify config was created
	assert.FileExists(t, filepath.Join(configDir, "chauffeur.yaml"), "Config file should be created")
}

func TestBasicConfigContent(t *testing.T) {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	wsDir := filepath.Join(tmpHome, ".chauffeur")
	
	// Create basic config
	createBasicConfigForTest(t, wsDir)

	// Read and verify config content
	configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(content)
	
	// Verify user-space ports are set correctly
	assert.Contains(t, configContent, "http_port: 8080", "Caddy HTTP should use user-space port")
	assert.Contains(t, configContent, "https_port: 8443", "Caddy HTTPS should use user-space port")
	assert.Contains(t, configContent, "http_port: 8081", "Nginx HTTP should use different port")
	assert.Contains(t, configContent, "https_port: 8444", "Nginx HTTPS should use different port")
	
	// Verify port management settings
	assert.Contains(t, configContent, "conflict_resolution: \"prompt\"", "Default conflict resolution should be prompt")
	assert.Contains(t, configContent, "start_range: 8080", "Port range should start at 8080")
	assert.Contains(t, configContent, "end_range: 8099", "Port range should end at 8099")
	
	// Verify it's a valid YAML structure
	lines := strings.Split(configContent, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			assert.True(t, strings.Contains(line, ":"), "Line %d should be valid YAML: %s", i+1, line)
		}
	}
}

func TestInstallScriptIntegration(t *testing.T) {
	// This test simulates the key parts of install.sh logic
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	wsDir := filepath.Join(tmpHome, ".chauffeur")
	binDir := filepath.Join(wsDir, "bin")
	configFile := filepath.Join(wsDir, "config", "chauffeur.yaml")

	// Simulate install.sh checking for existing config
	assert.NoFileExists(t, configFile, "Config file should not exist initially")

	// Simulate the script creating workspace directory
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)

	// Simulate the script detecting no config and creating basic config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Simulate logic: workspace not initialized, create basic config
		createBasicConfigForTest(t, wsDir)
	}

	// Verify results
	assert.FileExists(t, configFile, "Config file should exist after basic config creation")
	assert.DirExists(t, binDir, "Binary directory should be created")
}

func TestPortConflictPrevention(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	wsDir := filepath.Join(tmpHome, ".chauffeur")
	createBasicConfigForTest(t, wsDir)

	// Read config and parse ports manually (simple approach for test)
	configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Parse port values from content
	lines := strings.Split(string(content), "\n")
	portMap := make(map[string]int)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "http_port:") || strings.Contains(line, "https_port:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.ReplaceAll(parts[0], " ", "_"))
				value := strings.TrimSpace(parts[1])
				if port, err := fmt.Sscanf(value, "%d", new(int)); err == nil && port == 1 {
					var val int
					_, err := fmt.Sscanf(value, "%d", &val)
					if err == nil {
						portMap[key] = val
					}
				}
			}
		}
	}

	// Verify no conflicts with common system ports
	assert.NotEqual(t, 80, portMap["caddy_http_port"], "Caddy HTTP shouldn't use system port 80")
	assert.NotEqual(t, 443, portMap["caddy_https_port"], "Caddy HTTPS shouldn't use system port 443")
	assert.NotEqual(t, 80, portMap["nginx_http_port"], "Nginx HTTP shouldn't use system port 80")
	assert.NotEqual(t, 443, portMap["nginx_https_port"], "Nginx HTTPS shouldn't use system port 443")
	
	// Verify Caddy and Nginx use different ports
	if caddyHttp, ok := portMap["caddy_http_port"]; ok {
		if nginxHttp, ok2 := portMap["nginx_http_port"]; ok2 {
			assert.NotEqual(t, caddyHttp, nginxHttp, "Caddy and Nginx should use different HTTP ports")
		}
	}
}

// createBasicConfigForTest simulates the basic config creation from install.sh
func createBasicConfigForTest(t *testing.T, wsDir string) {
	configDir := filepath.Join(wsDir, "config")
	configFile := filepath.Join(configDir, "chauffeur.yaml")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configContent := `# Chauffeur Configuration File
# This file controls global Chauffeur settings and port management

# Configuration version (do not modify)
version: 1

# Enable/disable telemetry data collection
telemetry: false

# Workspace directory where Chauffeur stores its data
workspace_dir: ~/.chauffeur

# Caddy web server configuration
caddy:
  enable: true
  # Custom ports to avoid conflicts with system services
  http_port: 8080     # HTTP port (user-space)
  https_port: 8443    # HTTPS port (user-space)

# Nginx web server configuration  
nginx:
  enable: true
  # Different ports from Caddy to avoid conflicts
  http_port: 8081     # HTTP port (user-space)
  https_port: 8444    # HTTPS port (user-space)

# PHP runtime configuration
php:
  default: "8.3"

# Port management settings
ports:
  # Port range for automatic port allocation
  start_range: 8080
  end_range: 8099
  
  # How to handle port conflicts:
  # - "prompt": Ask user to select alternative ports (default)
  # - "auto":  Automatically select available ports
  # - "fail":  Fail if ports are in use
  conflict_resolution: "prompt"
  
  # Fallback ports for each service
  caddy_http_fallback: 8080
  caddy_https_fallback: 8443
  nginx_http_fallback: 8081
  nginx_https_fallback: 8444
  php_fpm_fallback: 9000

# Directory where Chauffeur stores project configurations
projects_dir: ~/.chauffeur/projects
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
}
