package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/internal/projects"
)

// TestTemplateData tests TemplateData structure
func TestTemplateData(t *testing.T) {
	data := TemplateData{
		ProjectSlug:  "test-project",
		ServerName:   "test-project.test",
		ProjectRoot:  "/path/to/project",
		PHPFpmSocket: "/path/to/php-fpm.sock",
		LogsDir:      "/path/to/logs",
		SSL:          true,
		HTTPPort:     8080,
		HTTPSPort:    8443,
		SSLCertPath:  "/path/to/cert.pem",
		SSLKeyPath:   "/path/to/key.pem",
	}

	if data.ProjectSlug != "test-project" {
		t.Errorf("Expected project slug 'test-project', got '%s'", data.ProjectSlug)
	}

	if data.ServerName != "test-project.test" {
		t.Errorf("Expected server name 'test-project.test', got '%s'", data.ServerName)
	}

	if !data.SSL {
		t.Error("Expected SSL to be true")
	}

	if data.HTTPPort != 8080 {
		t.Errorf("Expected HTTP port 8080, got %d", data.HTTPPort)
	}
}

// TestNginxConfigOptions tests NginxConfigOptions structure and defaults
func TestNginxConfigOptions(t *testing.T) {
	// Test with custom values
	opts := NginxConfigOptions{
		HTTPPort:    9000,
		HTTPSPort:   9443,
		SSLCertPath: "/custom/cert.pem",
		SSLKeyPath:  "/custom/key.pem",
	}

	if opts.HTTPPort != 9000 {
		t.Errorf("Expected HTTP port 9000, got %d", opts.HTTPPort)
	}

	if opts.HTTPSPort != 9443 {
		t.Errorf("Expected HTTPS port 9443, got %d", opts.HTTPSPort)
	}

	// Test applyDefaults with zero values
	zeroOpts := NginxConfigOptions{}
	zeroOpts.applyDefaults()

	if zeroOpts.HTTPPort != 8080 {
		t.Errorf("Expected default HTTP port 8080, got %d", zeroOpts.HTTPPort)
	}

	if zeroOpts.HTTPSPort != 8443 {
		t.Errorf("Expected default HTTPS port 8443, got %d", zeroOpts.HTTPSPort)
	}
}

// TestNewTemplateEngine tests template engine creation
func TestNewTemplateEngine(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Errorf("Expected NewTemplateEngine to succeed: %v", err)
	}

	if engine == nil {
		t.Error("Expected template engine to be created")
	}
}

// TestDetectTemplateType tests template type detection
func TestDetectTemplateType(t *testing.T) {
	tempDir := t.TempDir()

	// Test Laravel detection
	laravelDir := filepath.Join(tempDir, "laravel")
	os.MkdirAll(filepath.Join(laravelDir, "public"), 0755)
	os.WriteFile(filepath.Join(laravelDir, "artisan"), []byte("#!/usr/bin/env php"), 0755)

	// Test WordPress detection
	wordpressDir := filepath.Join(tempDir, "wordpress")
	os.MkdirAll(filepath.Join(wordpressDir, "wp-includes"), 0755)
	os.MkdirAll(filepath.Join(wordpressDir, "wp-admin"), 0755)
	os.WriteFile(filepath.Join(wordpressDir, "wp-config.php"), []byte("<?php"), 0644)

	// Test general project
	generalDir := filepath.Join(tempDir, "general")
	os.MkdirAll(generalDir, 0755)
	os.WriteFile(filepath.Join(generalDir, "index.html"), []byte("<html>"), 0644)

	engine := &TemplateEngine{}

	testCases := []struct {
		name       string
		path       string
		expected   string
	}{
		{"laravel project", laravelDir, "laravel"},
		{"wordpress project", wordpressDir, "wordpress"},
		{"general project", generalDir, "general"},
		{"empty path", "", "general"},
		{"nonexistent path", "/nonexistent/path", "general"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.DetectTemplateType(tc.path)
			if result != tc.expected {
				t.Errorf("Expected template type '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestProcessTemplate tests template variable replacement
func TestProcessTemplate(t *testing.T) {
	engine := &TemplateEngine{}

	data := TemplateData{
		ProjectSlug:  "test-project",
		ServerName:   "test-project.test",
		ProjectRoot:  "/path/to/project",
		PHPFpmSocket: "/path/to/php-fpm.sock",
		LogsDir:      "/path/to/logs",
		HTTPPort:     8080,
		HTTPSPort:    8443,
		SSLCertPath:  "/path/to/cert.pem",
		SSLKeyPath:   "/path/to/key.pem",
	}

	templateContent := `server {
    listen {{NGINX_HTTP_PORT}};
    server_name {{SERVER_NAME}};
    root {{PROJECT_ROOT}};

    location ~ \.php$ {
        fastcgi_pass unix:{{PHP_FPM_SOCKET}};
        access_log {{LOGS_DIR}}/access.log;
    }

    {{#SSL}}
    ssl_certificate {{SSL_CERT_PATH}};
    ssl_certificate_key {{SSL_KEY_PATH}};
    {{/SSL}}
}`

	result, err := engine.processTemplate(templateContent, data)
	if err != nil {
		t.Errorf("Expected processTemplate to succeed: %v", err)
	}

	expectedReplacements := []string{
		"listen 8080;",
		"server_name test-project.test;",
		"root /path/to/project;",
		"fastcgi_pass unix:/path/to/php-fpm.sock;",
		"access_log /path/to/logs/access.log;",
	}

	for _, expected := range expectedReplacements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s'", expected)
		}
	}
}

// TestProcessTemplateWithoutSSL tests template processing with SSL disabled
func TestProcessTemplateWithoutSSL(t *testing.T) {
	engine := &TemplateEngine{}

	data := TemplateData{
		ProjectSlug:  "test-project",
		ServerName:   "test-project.test",
		ProjectRoot:  "/path/to/project",
		PHPFpmSocket: "/path/to/php-fpm.sock",
		LogsDir:      "/path/to/logs",
		SSL:          false,
	}

	templateContent := `server {
    listen 8080;
    server_name {{SERVER_NAME}};

    {{#SSL}}
    ssl_certificate {{SSL_CERT_PATH}};
    ssl_certificate_key {{SSL_KEY_PATH}};
    {{/SSL}}
}`

	result, err := engine.processTemplate(templateContent, data)
	if err != nil {
		t.Errorf("Expected processTemplate to succeed: %v", err)
	}

	// SSL section should be removed
	if strings.Contains(result, "ssl_certificate") {
		t.Error("Expected SSL section to be removed when SSL is disabled")
	}

	if strings.Contains(result, "{{#SSL}}") || strings.Contains(result, "{{/SSL}}") {
		t.Error("Expected SSL conditional markers to be removed")
	}
}

// TestRenderNginxTemplate tests nginx template rendering
func TestRenderNginxTemplate(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	data := TemplateData{
		ProjectSlug:  "test-project",
		ServerName:   "test-project.test",
		ProjectRoot:  "/path/to/project",
		PHPFpmSocket: "/path/to/php-fpm.sock",
		LogsDir:      "/path/to/logs",
		HTTPPort:     8080,
	}

	// Test with nonexistent template (should fall back to embedded templates)
	result, err := engine.RenderNginxTemplate("general.conf", data)
	if err != nil {
		t.Errorf("Expected RenderNginxTemplate to succeed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty template result")
	}

	// Should contain some of the replaced content
	if !strings.Contains(result, "test-project") {
		t.Error("Expected result to contain project slug")
	}
}

// TestGenerateNginxConfig tests nginx configuration generation
func TestGenerateNginxConfig(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	config := projects.Config{
		Path: "/path/to/project",
		Site: &projects.Site{
			Domain: "test-project.test",
			SSL:    false,
		},
		Runtime: projects.Runtime{
			PHPFPM: "/path/to/php-fpm.sock",
		},
	}

	layout := projects.Layout{
		Root:       "test-project",
		SocketPath: "/path/to/php-fpm.sock",
		LogsDir:    "/path/to/logs",
	}

	opts := NginxConfigOptions{
		HTTPPort:  8080,
		HTTPSPort: 8443,
	}

	testCases := []struct {
		name         string
		templateType string
	}{
		{"general template", "general"},
		{"laravel template", "laravel"},
		{"wordpress template", "wordpress"},
		{"unknown template", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.GenerateNginxConfig(config, layout, tc.templateType, opts)
			if err != nil {
				t.Errorf("Expected GenerateNginxConfig to succeed: %v", err)
			}

			if result == "" {
				t.Error("Expected non-empty configuration result")
			}

			// Should contain some basic nginx configuration
			if !strings.Contains(result, "server") {
				t.Error("Expected result to contain nginx server configuration")
			}
		})
	}
}

// TestGenerateBasicConfig tests basic configuration generation
func TestGenerateBasicConfig(t *testing.T) {
	engine := &TemplateEngine{}

	config := projects.Config{
		Path: "/path/to/project",
		Site: &projects.Site{
			Domain: "test-project.test",
			SSL:    false,
		},
		Runtime: projects.Runtime{
			PHPFPM: "/path/to/php-fpm.sock",
		},
	}

	layout := projects.Layout{
		Root:       "test-project",
		SocketPath: "/path/to/php-fpm.sock",
		LogsDir:    "/path/to/logs",
	}

	opts := NginxConfigOptions{
		HTTPPort:  8080,
		HTTPSPort: 8443,
	}

	result := engine.generateBasicConfig(config, layout, opts)

	if result == "" {
		t.Error("Expected non-empty basic configuration")
	}

	expectedContent := []string{
		"test-project",
		"8080",
		"server_name",
		"location ~ \\.php$",
		"fastcgi_pass",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected basic config to contain '%s'", expected)
		}
	}
}

// TestGenerateBasicConfigWithSSL tests basic configuration generation with SSL
func TestGenerateBasicConfigWithSSL(t *testing.T) {
	engine := &TemplateEngine{}

	config := projects.Config{
		Path: "/path/to/project",
		Site: &projects.Site{
			Domain: "test-project.test",
			SSL:    true,
		},
		Runtime: projects.Runtime{
			PHPFPM: "/path/to/php-fpm.sock",
		},
	}

	layout := projects.Layout{
		Root:       "test-project",
		SocketPath: "/path/to/php-fpm.sock",
		LogsDir:    "/path/to/logs",
	}

	opts := NginxConfigOptions{
		HTTPPort:    8080,
		HTTPSPort:   8443,
		SSLCertPath: "/path/to/cert.pem",
		SSLKeyPath:  "/path/to/key.pem",
	}

	result := engine.generateBasicConfig(config, layout, opts)

	// Should contain SSL configuration
	expectedSSLContent := []string{
		"ssl_certificate",
		"ssl_certificate_key",
		"TLSv1.2 TLSv1.3",
		"8443 ssl",
	}

	for _, expected := range expectedSSLContent {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected SSL config to contain '%s'", expected)
		}
	}
}

// TestFileHelpers tests file existence helper functions
func TestFileHelpers(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test fileExists
	if !fileExists(testFile) {
		t.Error("Expected fileExists to return true for existing file")
	}

	if fileExists(testDir) {
		t.Error("Expected fileExists to return false for directory")
	}

	if fileExists("/nonexistent/file") {
		t.Error("Expected fileExists to return false for nonexistent path")
	}

	// Test dirExists
	if !dirExists(testDir) {
		t.Error("Expected dirExists to return true for existing directory")
	}

	if dirExists(testFile) {
		t.Error("Expected dirExists to return false for file")
	}

	if dirExists("/nonexistent/directory") {
		t.Error("Expected dirExists to return false for nonexistent path")
	}
}

// TestRemoveNginxConfig tests nginx configuration removal
func TestRemoveNginxConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	nginxDir := filepath.Join(tempDir, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	err := os.MkdirAll(sitesAvailable, 0755)
	if err != nil {
		t.Fatalf("Failed to create sites-available directory: %v", err)
	}

	err = os.MkdirAll(sitesEnabled, 0755)
	if err != nil {
		t.Fatalf("Failed to create sites-enabled directory: %v", err)
	}

	// Create test config files
	configContent := "server { listen 8080; }"
	configPath := filepath.Join(sitesAvailable, "test-project.conf")
	enabledPath := filepath.Join(sitesEnabled, "test-project.conf")

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = os.Symlink(configPath, enabledPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	engine := &TemplateEngine{}

	// Test successful removal
	err = engine.RemoveNginxConfig("test-project")
	if err != nil {
		t.Errorf("Expected RemoveNginxConfig to succeed: %v", err)
	}

	// Verify files are removed
	if _, err := os.Stat(enabledPath); !os.IsNotExist(err) {
		t.Error("Expected enabled config to be removed")
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Expected available config to be removed")
	}

	// Test removal of nonexistent config (should not error)
	err = engine.RemoveNginxConfig("nonexistent-project")
	if err != nil {
		t.Errorf("Expected RemoveNginxConfig to handle nonexistent config: %v", err)
	}

	// Test with empty slug
	err = engine.RemoveNginxConfig("")
	if err == nil {
		t.Error("Expected RemoveNginxConfig to fail with empty slug")
	}
}

// TestEnsureCatchallServer tests catch-all server creation
func TestEnsureCatchallServer(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	engine := &TemplateEngine{}

	err := engine.EnsureCatchallServer(8080)
	if err != nil {
		t.Errorf("Expected EnsureCatchallServer to succeed: %v", err)
	}

	// Verify the catch-all server was created
	nginxDir := filepath.Join(tempDir, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	configPath := filepath.Join(sitesAvailable, "000-default.conf")
	enabledPath := filepath.Join(sitesEnabled, "000-default.conf")

	// Check available config
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Expected catch-all config to exist: %v", err)
	}

	expectedContent := []string{
		"default_server",
		"server_name _",
		"return 404",
		"8080",
	}

	contentStr := string(content)
	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected catch-all config to contain '%s'", expected)
		}
	}

	// Check symlink exists
	if _, err := os.Lstat(enabledPath); os.IsNotExist(err) {
		t.Error("Expected catch-all config to be enabled")
	}

	// Test calling again (should not error)
	err = engine.EnsureCatchallServer(8080)
	if err != nil {
		t.Errorf("Expected EnsureCatchallServer to succeed on second call: %v", err)
	}
}

// TestTemplateEngineEdgeCases tests edge cases and error conditions
func TestTemplateEngineEdgeCases(t *testing.T) {
	engine := &TemplateEngine{}

	// Test processTemplate with empty content
	result, err := engine.processTemplate("", TemplateData{})
	if err != nil {
		t.Errorf("Expected processTemplate to handle empty content: %v", err)
	}

	if result != "" {
		t.Error("Expected empty result for empty template content")
	}

	// Test processTemplate with template containing unknown placeholders
	templateWithUnknown := `server {
        listen {{UNKNOWN_PLACEHOLDER}};
        server_name {{SERVER_NAME}};
    }`

	data := TemplateData{
		ServerName: "test.test",
	}

	result, err = engine.processTemplate(templateWithUnknown, data)
	if err != nil {
		t.Errorf("Expected processTemplate to handle unknown placeholders: %v", err)
	}

	// Unknown placeholder should remain unchanged
	if !strings.Contains(result, "{{UNKNOWN_PLACEHOLDER}}") {
		t.Error("Expected unknown placeholder to remain in result")
	}

	// Known placeholder should be replaced
	if !strings.Contains(result, "test.test") {
		t.Error("Expected known placeholder to be replaced")
	}
}

// TestRenderNginxTemplateErrors tests error handling in template rendering
func TestRenderNginxTemplateErrors(t *testing.T) {
	engine := &TemplateEngine{}

	data := TemplateData{
		ProjectSlug: "test",
	}

	// Test with empty template name (should default to general.conf)
	result, err := engine.RenderNginxTemplate("", data)
	if err != nil {
		t.Errorf("Expected RenderNginxTemplate to handle empty template name: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result for default template")
	}
}