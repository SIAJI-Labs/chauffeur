package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInit(t *testing.T) {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Initialize workspace
	err := commands.RunInit([]string{})
	require.NoError(t, err)

	// Verify workspace directory exists
	wsDir := filepath.Join(tmpHome, ".chauffeur")
	assert.DirExists(t, wsDir, "Workspace directory should be created")

	// Verify config file exists
	configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
	assert.FileExists(t, configPath, "Configuration file should be created")

	// Verify configuration content
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Caddy.HTTPPort, "Default Caddy HTTP port should be correct")
	assert.Equal(t, 8443, cfg.Caddy.HTTPSPort, "Default Caddy HTTPS port should be correct")
	assert.Equal(t, 8081, cfg.Nginx.HTTPPort, "Default Nginx HTTP port should be correct")
	assert.Equal(t, 8444, cfg.Nginx.HTTPSPort, "Default Nginx HTTPS port should be correct")
	assert.Equal(t, "prompt", cfg.Ports.ConflictResolution, "Default conflict resolution should be prompt")
	assert.Equal(t, 8080, cfg.Ports.StartRange, "Default start range should be correct")
	assert.Equal(t, 8099, cfg.Ports.EndRange, "Default end range should be correct")

	// Verify subdirectories are created
	expectedDirs := []string{
		"projects",
		"logs",
		"cache",
		"php",
		"nginx/bin",
		"nginx/etc",
		"nginx/sites-available",
		"nginx/sites-enabled",
		"nginx/conf.d",
		"nginx/logs",
		"caddy/bin",
		"caddy/logs",
		"bin",
		"bin/shims",
	}

	for _, dir := range expectedDirs {
		assert.DirExists(t, filepath.Join(wsDir, dir), "Directory should be created: "+dir)
	}

	// Verify .gitignore is created
	gitignorePath := filepath.Join(wsDir, ".gitignore")
	assert.FileExists(t, gitignorePath, ".gitignore should be created")

	// Verify .gitignore content
	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "logs/", "Should ignore logs directory")
	assert.Contains(t, string(content), "cache/", "Should ignore cache directory")
}

func TestRunInitForce(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	wsDir := filepath.Join(tmpHome, ".chauffeur")
	configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")

	// Create initial workspace
	err := commands.RunInit([]string{})
	require.NoError(t, err)

	// Modify the config file
	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Caddy.HTTPPort = 9999
	err = config.Save(cfg)
	require.NoError(t, err)

	// Verify the modification
	loadedCfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 9999, loadedCfg.Caddy.HTTPPort, "Config should be modified")

	// Re-run init with --force
	err = commands.RunInit([]string{"--force"})
	require.NoError(t, err)

	// Verify config is reset to defaults
	loadedCfg, err = config.Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, loadedCfg.Caddy.HTTPPort, "Config should be reset to default")
}

func TestRunInitQuiet(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Run init with --quiet flag
	output := captureOutput(func() error {
		return commands.RunInit([]string{"--quiet"})
	})

	// Should still create everything
	wsDir := filepath.Join(tmpHome, ".chauffeur")
	assert.DirExists(t, wsDir, "Workspace should be created even in quiet mode")

	// Output should be minimal
	assert.Contains(t, output, tmpHome, "Should output workspace path even in quiet mode")
	assert.NotContains(t, output, "Creating directory:", "Should not show directory creation messages")
}

func TestRunInitAlreadyExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Initialize once
	err := commands.RunInit([]string{})
	require.NoError(t, err)

	// Try to initialize again without --force
	output := captureOutput(func() error {
		return commands.RunInit([]string{})
	})

	// Should not fail, but should mention workspace exists
	assert.Contains(t, output, "already initialized", "Should warn about existing workspace")
}

func TestDefaultConfigPorts(t *testing.T) {
	cfg, err := commands.DefaultConfig()
	require.NoError(t, err)

	// Verify default ports are user-space to avoid conflicts
	assert.Equal(t, 8080, cfg.Caddy.HTTPPort, "Caddy HTTP should use user-space port")
	assert.Equal(t, 8443, cfg.Caddy.HTTPSPort, "Caddy HTTPS should use user-space port")
	assert.Equal(t, 8081, cfg.Nginx.HTTPPort, "Nginx HTTP should use different port from Caddy")
	assert.Equal(t, 8444, cfg.Nginx.HTTPSPort, "Nginx HTTPS should use different port from Caddy")

	// Verify port management configuration
	assert.Equal(t, "prompt", cfg.Ports.ConflictResolution, "Default should be prompt")
	assert.Equal(t, 8080, cfg.Ports.CaddyHTTPFallback, "Caddy HTTP fallback should match default")
	assert.Equal(t, 8443, cfg.Ports.CaddyHTTPSFallback, "Caddy HTTPS fallback should match default")
	assert.Equal(t, 8081, cfg.Ports.NginxHTTPFallback, "Nginx HTTP fallback should match default")
	assert.Equal(t, 8444, cfg.Ports.NginxHTTPSFallback, "Nginx HTTPS fallback should match default")
	assert.Equal(t, 9000, cfg.Ports.PHPFPMFallback, "PHP-FPM fallback should be standard")
}

func TestPortConfigurationInWorkspace(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Initialize workspace
	err := commands.RunInit([]string{})
	require.NoError(t, err)

	// Load and verify configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Test that default ports don't conflict with common system ports
	assert.NotEqual(t, 80, cfg.Caddy.HTTPPort, "Caddy HTTP should not use system port 80")
	assert.NotEqual(t, 443, cfg.Caddy.HTTPSPort, "Caddy HTTPS should not use system port 443")
	assert.NotEqual(t, 80, cfg.Nginx.HTTPPort, "Nginx HTTP should not use system port 80")
	assert.NotEqual(t, 443, cfg.Nginx.HTTPSPort, "Nginx HTTPS should not use system port 443")

	// Test that Caddy and Nginx use different ports to avoid conflicts
	assert.NotEqual(t, cfg.Caddy.HTTPPort, cfg.Nginx.HTTPPort, "Caddy and Nginx should use different HTTP ports")
	assert.NotEqual(t, cfg.Caddy.HTTPSPort, cfg.Nginx.HTTPSPort, "Caddy and Nginx should use different HTTPS ports")

	// Test that port range is reasonable for user-space allocation
	assert.GreaterOrEqual(t, cfg.Ports.StartRange, 8000, "Port range should be in user-space")
	assert.LessOrEqual(t, cfg.Ports.EndRange, 9999, "Port range should be in user-space range")
	assert.Greater(t, cfg.Ports.EndRange, cfg.Ports.StartRange, "End range should be greater than start range")
}
