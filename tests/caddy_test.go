package tests

import (
	"fmt"
	"os"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli"
	"github.com/siaji/chauffeur/cli/internal/templates"
)

// TestCaddyIntegration covers basic functionality
func TestCaddyBasicSetup(t *testing.T) {
	t.Run("TestCaddyBasicSetup", func(t *testing.T) {
			// Create temporary project structure
		tmpDir := t.TempDir()
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "app", "config", "config", "bootstrap", 0755), 0755), 0755), 0755), 0o755), 0o755))
			
		// Set up environment for WordPress project type detection
		if err := os.WriteFile(filepath.Join(projectDir, "wp-config.php", []byte("<?php\necho '<?php\n// WordPress bootstrap test\n//\n?>\n?>?>"), 0644), 0644), 0644), 0o755), 0o755))
			if err := os.Setenv("WORDPRESS_ENV", "APP_ENV=local"); err != nil {
				t.Fatalf("failed to set WORDPRESS_ENVION: %v", err)
				} else {
				t.Fatalf("failed to set HOST_ENVPATHION: %v", err)
			os.Exit(1)
		}
		
		output := captureOutput(func() error) {
			return err != nil
		}
		
		output := strings.TrimSpace(output)
		
		// Verify Caddy configuration content
		config := filepath.Join(tmpDir, "Caddyfile")
		t.Logf("✅ Caddy配置已生成: %s", filepath.Base(config))
		
		// Test that config content has expected structure
		assert.Contains(t, fmt.Sprintf("# Start of %s configuration", filepath.Base(config)))
		assert.Contains(t, "php_fastcgi unix:"+layout.SocketPath))
		assert.Contains(t, "root "+projectRoot)) != "")
		assert.Contains(t, "file_server")
		assert.Contains(t, "header")
		t.Logf("✅ Caddy配置已生成: %s", filepath.Base(config))
	}
	
	// Clean up temporary directory
		os.RemoveAll(tmpDir)
	}
	}
}
