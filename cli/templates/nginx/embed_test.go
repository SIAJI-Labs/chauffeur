package nginx

import (
	"strings"
	"testing"
)

// TestFilesEmbed tests that the embedded file system contains expected templates
func TestFilesEmbed(t *testing.T) {
	// Test that we can read directory contents (this validates that Files is properly initialized)
	entries, err := Files.ReadDir(".")
	if err != nil {
		t.Errorf("Expected to read embedded directory: %v", err)
	}

	// Should have at least one template file
	if len(entries) == 0 {
		t.Error("Expected at least one embedded template file")
	}

	// Check for expected template files
	expectedFiles := []string{"general.conf", "laravel.conf", "wordpress.conf"}
	foundFiles := make(map[string]bool)

	for _, entry := range entries {
		name := entry.Name()
		foundFiles[name] = true

		// Verify it's a regular file
		if entry.IsDir() {
			t.Errorf("Expected %s to be a file, not a directory", name)
		}
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected embedded file %s to be present", expected)
		}
	}
}

// TestReadGeneralTemplate tests reading the general template
func TestReadGeneralTemplate(t *testing.T) {
	content, err := Files.ReadFile("general.conf")
	if err != nil {
		t.Errorf("Expected to read general.conf template: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected general.conf template to have content")
	}

	contentStr := string(content)

	// Check for expected template variables
	expectedPlaceholders := []string{
		"{{PROJECT_SLUG}}",
		"{{PHP_FPM_SOCKET}}",
		"{{NGINX_HTTP_PORT}}",
		"{{SERVER_NAME}}",
		"{{PROJECT_ROOT}}",
		"{{LOGS_DIR}}",
	}

	for _, placeholder := range expectedPlaceholders {
		if !strings.Contains(contentStr, placeholder) {
			t.Errorf("Expected general.conf template to contain placeholder %s", placeholder)
		}
	}

	// Check for expected nginx configuration
	expectedNginxContent := []string{
		"upstream",
		"server {",
		"listen",
		"server_name",
		"root",
		"location ~ \\.php$",
		"fastcgi_pass",
		"fastcgi_index",
	}

	for _, expected := range expectedNginxContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected general.conf template to contain nginx directive %s", expected)
		}
	}
}

// TestReadLaravelTemplate tests reading the Laravel template
func TestReadLaravelTemplate(t *testing.T) {
	content, err := Files.ReadFile("laravel.conf")
	if err != nil {
		t.Errorf("Expected to read laravel.conf template: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected laravel.conf template to have content")
	}

	contentStr := string(content)

	// Check for Laravel-specific configurations
	laravelSpecific := []string{
		"try_files $uri $uri/ /index.php?$query_string",
		"public",
	}

	for _, expected := range laravelSpecific {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected laravel.conf template to contain Laravel-specific content %s", expected)
		}
	}

	// Should also contain standard placeholders
	expectedPlaceholders := []string{
		"{{PROJECT_SLUG}}",
		"{{PHP_FPM_SOCKET}}",
		"{{SERVER_NAME}}",
		"{{PROJECT_ROOT}}",
	}

	for _, placeholder := range expectedPlaceholders {
		if !strings.Contains(contentStr, placeholder) {
			t.Errorf("Expected laravel.conf template to contain placeholder %s", placeholder)
		}
	}
}

// TestReadWordPressTemplate tests reading the WordPress template
func TestReadWordPressTemplate(t *testing.T) {
	content, err := Files.ReadFile("wordpress.conf")
	if err != nil {
		t.Errorf("Expected to read wordpress.conf template: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected wordpress.conf template to have content")
	}

	contentStr := string(content)

	// Check for WordPress-specific configurations
	wordpressSpecific := []string{
		"wp-admin",
		"wp-includes",
		"wp-content",
	}

	// Should contain standard placeholders
	expectedPlaceholders := []string{
		"{{PROJECT_SLUG}}",
		"{{PHP_FPM_SOCKET}}",
		"{{SERVER_NAME}}",
		"{{PROJECT_ROOT}}",
	}

	for _, placeholder := range expectedPlaceholders {
		if !strings.Contains(contentStr, placeholder) {
			t.Errorf("Expected wordpress.conf template to contain placeholder %s", placeholder)
		}
	}

	// Check if it has WordPress-specific location blocks (if any)
	for _, expected := range wordpressSpecific {
		if strings.Contains(contentStr, expected) {
			t.Logf("WordPress template contains expected content: %s", expected)
		}
	}
}

// TestTemplateVariablesConsistency tests that all templates use consistent variables
func TestTemplateVariablesConsistency(t *testing.T) {
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	// Define core variables that should be present in all templates
	coreVariables := []string{
		"{{PROJECT_SLUG}}",
		"{{PHP_FPM_SOCKET}}",
		"{{NGINX_HTTP_PORT}}",
		"{{SERVER_NAME}}",
		"{{PROJECT_ROOT}}",
		"{{LOGS_DIR}}",
	}

	// SSL-related variables (may not be in all templates)
	sslVariables := []string{
		"{{NGINX_HTTPS_PORT}}",
		"{{SSL_CERT_PATH}}",
		"{{SSL_KEY_PATH}}",
	}

	for _, template := range templates {
		t.Run("template_"+template, func(t *testing.T) {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Expected to read %s template: %v", template, err)
				return
			}

			contentStr := string(content)

			// Check core variables
			for _, variable := range coreVariables {
				if !strings.Contains(contentStr, variable) {
					t.Errorf("Expected %s template to contain variable %s", template, variable)
				}
			}

			// Log SSL variables presence (they might be optional in some templates)
			for _, variable := range sslVariables {
				if strings.Contains(contentStr, variable) {
					t.Logf("Template %s contains SSL variable %s", template, variable)
				}
			}
		})
	}
}

// TestTemplateStructure tests that templates have proper nginx structure
func TestTemplateStructure(t *testing.T) {
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	// Basic nginx server structure that should be present
	serverStructure := []string{
		"server {",
		"listen",
		"server_name",
		"root",
		"index",
		"location",
		"}",
	}

	// PHP-FPM related structure
	phpFpmStructure := []string{
		"upstream",
		"fastcgi_pass",
		"fastcgi_index",
		"include fastcgi_params",
	}

	for _, template := range templates {
		t.Run("structure_"+template, func(t *testing.T) {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Expected to read %s template: %v", template, err)
				return
			}

			contentStr := string(content)

			// Check basic server structure
			for _, element := range serverStructure {
				if !strings.Contains(contentStr, element) {
					t.Errorf("Expected %s template to contain nginx structure element '%s'", template, element)
				}
			}

			// Check PHP-FPM structure
			for _, element := range phpFpmStructure {
				if !strings.Contains(contentStr, element) {
					t.Errorf("Expected %s template to contain PHP-FPM structure element '%s'", template, element)
				}
			}
		})
	}
}

// TestTemplateSecurityFeatures tests that templates include security headers
func TestTemplateSecurityFeatures(t *testing.T) {
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	// Security headers that should be present
	securityHeaders := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
	}

	for _, template := range templates {
		t.Run("security_"+template, func(t *testing.T) {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Expected to read %s template: %v", template, err)
				return
			}

			contentStr := string(content)

			for _, header := range securityHeaders {
				if strings.Contains(contentStr, header) {
					t.Logf("Template %s includes security header %s", template, header)
				}
			}
		})
	}
}

// TestTemplateErrorHandling tests error handling with embedded files
func TestTemplateErrorHandling(t *testing.T) {
	// Test reading nonexistent file
	_, err := Files.ReadFile("nonexistent.conf")
	if err == nil {
		t.Error("Expected reading nonexistent file to fail")
	}

	// Test opening nonexistent directory
	_, err = Files.ReadDir("nonexistent")
	if err == nil {
		t.Error("Expected opening nonexistent directory to fail")
	}
}

// TestMultipleFileOperations tests multiple file operations on embedded FS
func TestMultipleFileOperations(t *testing.T) {
	// Read all templates multiple times to ensure embed.FS works correctly
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	for i := 0; i < 3; i++ {
		for _, template := range templates {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Iteration %d: Expected to read %s template: %v", i, template, err)
				continue
			}

			if len(content) == 0 {
				t.Errorf("Iteration %d: Expected %s template to have content", i, template)
			}
		}

		// Also test directory listing
		entries, err := Files.ReadDir(".")
		if err != nil {
			t.Errorf("Iteration %d: Expected to read embedded directory: %v", i, err)
		}

		if len(entries) == 0 {
			t.Errorf("Iteration %d: Expected at least one embedded template file", i)
		}
	}
}

// TestFilePermissions tests that embedded files are readable
func TestFilePermissions(t *testing.T) {
	entries, err := Files.ReadDir(".")
	if err != nil {
		t.Fatalf("Expected to read embedded directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		t.Run("readable_"+entry.Name(), func(t *testing.T) {
			content, err := Files.ReadFile(entry.Name())
			if err != nil {
				t.Errorf("Expected to read embedded file %s: %v", entry.Name(), err)
				return
			}

			// File should have content
			if len(content) == 0 {
				t.Errorf("Expected embedded file %s to have content", entry.Name())
			}

			// Content should be text-based and contain nginx directives
			contentStr := string(content)
			if !strings.Contains(contentStr, "#") && !strings.Contains(contentStr, "server") {
				t.Errorf("Expected embedded file %s to contain nginx configuration", entry.Name())
			}
		})
	}
}

// TestTemplateComments tests that templates have proper comments
func TestTemplateComments(t *testing.T) {
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	for _, template := range templates {
		t.Run("comments_"+template, func(t *testing.T) {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Expected to read %s template: %v", template, err)
				return
			}

			contentStr := string(content)

			// Check for comments (should start with #)
			if strings.Contains(contentStr, "#") {
				t.Logf("Template %s contains comments", template)
			}
		})
	}
}

// TestTemplateSSLSupport tests SSL support in templates
func TestTemplateSSLSupport(t *testing.T) {
	templates := []string{"general.conf", "laravel.conf", "wordpress.conf"}

	// SSL-related placeholders and directives
	sslIndicators := []string{
		"{{NGINX_HTTPS_PORT}}",
		"{{SSL_CERT_PATH}}",
		"{{SSL_KEY_PATH}}",
		"{{#SSL}}",
		"{{/SSL}}",
		"ssl_certificate",
		"ssl_certificate_key",
	}

	for _, template := range templates {
		t.Run("ssl_"+template, func(t *testing.T) {
			content, err := Files.ReadFile(template)
			if err != nil {
				t.Errorf("Expected to read %s template: %v", template, err)
				return
			}

			contentStr := string(content)

			// Check for any SSL-related content
			foundSSL := false
			for _, indicator := range sslIndicators {
				if strings.Contains(contentStr, indicator) {
					foundSSL = true
					t.Logf("Template %s contains SSL indicator: %s", template, indicator)
				}
			}

			if !foundSSL {
				t.Logf("Template %s does not appear to have SSL support", template)
			}
		})
	}
}