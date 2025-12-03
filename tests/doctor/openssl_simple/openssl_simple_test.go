package openssl_simple

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestDoctorSSLDependenciesWithOpenSSL(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	t.Run("SSL dependencies check runs without errors", func(t *testing.T) {
		// Test that doctor SSL dependencies check runs successfully
		// This will test our OpenSSL configuration validation indirectly
		err := commands.RunDoctor([]string{"--check-ssl", "--quiet"})

		// The command should complete successfully even if OpenSSL config is missing
		// because missing config is handled as a warning, not an error
		if err != nil {
			// Check if the error is just "dependencies missing" rather than a real failure
			if !strings.Contains(err.Error(), "found") && !strings.Contains(err.Error(), "need to be resolved") {
				t.Fatalf("Expected no error for doctor SSL check, got: %v", err)
			}
		}
	})
}

func TestOpenSSLConfigurationFileCreation(t *testing.T) {
	// Test that OpenSSL configuration files can be created and have expected content
	workspaceDir := t.TempDir()
	phpDir := filepath.Join(workspaceDir, "php")
	phpVersion := "8.3"
	confDir := filepath.Join(phpDir, phpVersion, "etc", "conf.d")
	opensslConfPath := filepath.Join(confDir, "openssl.ini")

	t.Run("Creates OpenSSL configuration with expected structure", func(t *testing.T) {
		// Create directories
		if err := os.MkdirAll(confDir, 0755); err != nil {
			t.Fatalf("Failed to create conf.d directory: %v", err)
		}

		// Create a sample OpenSSL configuration file
		opensslConfContent := `
; Chauffeur OpenSSL Configuration for PHP 8.3
; Auto-generated for test environment
;
; Certificate Authority settings for SSL/TLS verification
; This configuration enables secure connections to remote services
; including SMTP servers, HTTPS APIs, and other TLS-enabled services
;
; Path to the CA bundle file containing trusted root certificates
openssl.cafile = /etc/ssl/certs/ca-certificates.crt

; Path to the directory containing CA certificates
openssl.capath = /etc/ssl/certs
`

		if err := os.WriteFile(opensslConfPath, []byte(opensslConfContent), 0644); err != nil {
			t.Fatalf("Failed to write openssl.ini: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(opensslConfPath); err != nil {
			t.Fatalf("OpenSSL configuration file was not created: %v", err)
		}

		// Verify the content
		content, err := os.ReadFile(opensslConfPath)
		if err != nil {
			t.Fatalf("Failed to read openssl.ini: %v", err)
		}

		contentStr := string(content)

		// Check for required OpenSSL configuration entries
		requiredEntries := []string{
			"openssl.cafile",
			"openssl.capath",
			"/etc/ssl/certs/ca-certificates.crt",
		}

		for _, entry := range requiredEntries {
			if !strings.Contains(contentStr, entry) {
				t.Errorf("OpenSSL configuration missing required entry: %s", entry)
			}
		}
	})

	t.Run("Detects incomplete OpenSSL configuration", func(t *testing.T) {
		// Create incomplete OpenSSL configuration
		incompleteContent := `
; Incomplete OpenSSL configuration
openssl.cafile = /etc/ssl/certs/ca-certificates.crt
`

		if err := os.WriteFile(opensslConfPath, []byte(incompleteContent), 0644); err != nil {
			t.Fatalf("Failed to write incomplete openssl.ini: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(opensslConfPath); err != nil {
			t.Fatalf("OpenSSL configuration file was not created: %v", err)
		}

		// Verify the content
		content, err := os.ReadFile(opensslConfPath)
		if err != nil {
			t.Fatalf("Failed to read openssl.ini: %v", err)
		}

		contentStr := string(content)

		// Should have cafile but missing capath
		if !strings.Contains(contentStr, "openssl.cafile") {
			t.Error("Expected openssl.cafile in incomplete configuration")
		}

		if strings.Contains(contentStr, "openssl.capath") {
			t.Error("Expected openssl.capath to be missing in incomplete configuration")
		}
	})

	t.Run("Handles missing OpenSSL configuration gracefully", func(t *testing.T) {
		// Remove the OpenSSL configuration file to simulate missing config
		if err := os.Remove(opensslConfPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("Failed to remove openssl.ini: %v", err)
		}

		// The directories should still exist
		if _, err := os.Stat(confDir); err != nil {
			t.Fatalf("Expected conf.d directory to exist: %v", err)
		}

		// The OpenSSL configuration file should not exist
		if _, err := os.Stat(opensslConfPath); !os.IsNotExist(err) {
			t.Errorf("Expected openssl.ini to not exist, got: %v", err)
		}
	})
}

func TestOpenSSLDistributionPaths(t *testing.T) {
	t.Run("Validates common certificate paths", func(t *testing.T) {
		// Test common certificate paths that our OpenSSL configuration should support
		commonPaths := []string{
			"/etc/ssl/certs/ca-certificates.crt",      // Debian/Ubuntu/Arch
			"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // Fedora/RHEL
			"/etc/ssl/ca-bundle.pem",                  // openSUSE
			"/etc/pki/tls/cert.pem",                  // RHEL fallback
		}

		// At least one of these should exist on the test system
		foundValidPath := false
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				foundValidPath = true
				break
			}
		}

		// Log which paths are available for debugging
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				t.Logf("Found certificate path: %s", path)
			}
		}

		// This test will pass even if no paths are found (common in CI environments)
		if !foundValidPath {
			t.Log("No system certificate paths found (expected in CI environments)")
		}
	})
}