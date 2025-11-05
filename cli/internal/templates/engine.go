package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/siaji/chauffeur/cli/internal/projects"
)

// TemplateData represents the data available to nginx templates
type TemplateData struct {
	ProjectSlug    string
	ServerName     string
	ProjectRoot    string
	PHPFpmSocket   string
	LogsDir        string
	SSL            bool
	CaddySSLCert   string
	CaddySSLKey    string
}

// CaddyTemplateData represents the data available to Caddy templates
type CaddyTemplateData struct {
	ServerName     string
	ProjectRoot    string
	PHPFpmSocket   string
	SiteDomain    bool
	SSL            bool
}

// TemplateEngine handles nginx template processing
type TemplateEngine struct {
	templateDir string
	caddyDir   string
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() (*TemplateEngine, error) {
	// Try multiple template directory paths
	var nginxDir, caddyDir string
	
	// Get the template directory relative to the CLI binary location
	exePath, err := os.Executable()
	if err == nil {
		// Try nginx templates
		nginxCandidate := filepath.Join(filepath.Dir(exePath), "..", "..", "templates", "nginx")
		if _, err := os.Stat(nginxCandidate); err == nil {
			nginxDir = nginxCandidate
		}
		
		// Try caddy templates
		caddyCandidate := filepath.Join(filepath.Dir(exePath), "..", "..", "templates", "caddy")
		if _, err := os.Stat(caddyCandidate); err == nil {
			caddyDir = caddyCandidate
		}
	}
	
	// Fallback to development path if installation path doesn't exist
	if nginxDir == "" {
		// Assume we're in development environment
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}
		
		nginxDevCandidate := filepath.Join(wd, "cli", "templates", "nginx")
		if _, err := os.Stat(nginxDevCandidate); err == nil {
			nginxDir = nginxDevCandidate
		}
		
		caddyDevCandidate := filepath.Join(wd, "cli", "templates", "caddy")
		if _, err := os.Stat(caddyDevCandidate); err == nil {
			caddyDir = caddyDevCandidate
		}
	}
	
	return &TemplateEngine{
		templateDir: nginxDir,
		caddyDir:  caddyDir,
	}, nil
}

// RenderNginxTemplate renders a nginx template with the provided data
func (e *TemplateEngine) RenderNginxTemplate(templateName string, data TemplateData) (string, error) {
	if e.templateDir == "" {
		return "", fmt.Errorf("template directory not available")
	}
	
	templatePath := filepath.Join(e.templateDir, templateName)
	if templateName == "" {
		templatePath = filepath.Join(e.templateDir, "general.conf")
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read nginx template %s: %w", templatePath, err)
	}
	
	return e.processTemplate(string(content), data)
}

// RenderCaddyTemplate renders a Caddy template with the provided data
func (e *TemplateEngine) RenderCaddyTemplate(templateName string, data CaddyTemplateData) (string, error) {
	if e.caddyDir == "" {
		return "", fmt.Errorf("template directory not available")
	}
	
	templatePath := filepath.Join(e.caddyDir, templateName)
	if templateName == "" {
		templatePath = filepath.Join(e.caddyDir, "general.conf")
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read caddy template %s: %w", templatePath, err)
	}
	
	return e.processCaddyTemplate(string(content), data), nil
}

// processTemplate applies template variable substitution with simple {{VAR}} syntax
func (e *TemplateEngine) processTemplate(content string, data TemplateData) (string, error) {
	result := content
	
	// Basic variable replacement
	replacements := map[string]string{
		"{{PROJECT_SLUG}}":    data.ProjectSlug,
		"{{SERVER_NAME}}":      data.ServerName,
		"{{PROJECT_ROOT}}":     data.ProjectRoot,
		"{{PHP_FPM_SOCKET}}":   data.PHPFpmSocket,
		"{{LOGS_DIR}}":         data.LogsDir,
		"{{CADDY_SSL_CERT}}":   data.CaddySSLCert,
		"{{CADDY_SSL_KEY}}":    data.CaddySSLKey,
	}
	
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Handle conditional SSL blocks
	if !data.SSL {
		// Remove SSL section (everything between {{#SSL}} and {{/SSL}})
		start := strings.Index(result, "{{#SSL}}")
		end := strings.Index(result, "{{/SSL}}")
		
		if start != -1 && end != -1 {
			beforeSSL := result[:start]
			afterSSL := result[end + len("{{/SSL}}"):]
			result = beforeSSL + afterSSL
		}
	} else {
		// Remove the conditional markers but keep the content
		result = strings.ReplaceAll(result, "{{#SSL}}", "")
		result = strings.ReplaceAll(result, "{{/SSL}}", "")
	}
	
	return result, nil
}

// GenerateNginxConfig generates nginx configuration for a project
func (e *TemplateEngine) GenerateNginxConfig(config projects.Config, layout projects.Layout, templateType string) (string, error) {
	// Determine server name
	serverName := layout.Root
	if config.Site != nil && config.Site.Domain != "" {
		serverName = config.Site.Domain
	}
	
	// Prepare template data
	data := TemplateData{
		ProjectSlug:  filepath.Base(layout.Root),
		ServerName:   serverName,
		ProjectRoot:  config.Path,
		PHPFpmSocket: layout.SocketPath,
		LogsDir:      layout.LogsDir,
	}
	
	// Configure SSL if enabled
	if config.Site != nil && config.Site.SSL {
		data.SSL = true
		data.CaddySSLCert = "/etc/ssl/certs/caddy.pem"
		data.CaddySSLKey = "/etc/ssl/private/caddy.key"
	}
	
	// Choose template based on type
	templateName := "general.conf"
	switch templateType {
	case "laravel":
		templateName = "laravel.conf"
	case "wordpress":
		templateName = "wordpress.conf"
	case "general":
		templateName = "general.conf"
	default:
		templateName = "general.conf"
	}
	
	return e.RenderNginxTemplate(templateName, data)
}

// GenerateCaddyConfig generates Caddy configuration for a project
func (e *TemplateEngine) GenerateCaddyConfig(config projects.Config, layout projects.Layout, templateType string) (string, error) {
	// Determine server name
	serverName := layout.Root
	if config.Site != nil && config.Site.Domain != "" {
		serverName = config.Site.Domain
	}
	
	// Prepare template data
	data := CaddyTemplateData{
		ServerName:   serverName,
		ProjectRoot:  config.Path,
		PHPFpmSocket: layout.SocketPath,
		SiteDomain:  config.Site != nil && config.Site.Domain != "",
		SSL:         config.Site != nil && config.Site.SSL,
	}
	
	// Choose template based on type
	templateName := "general.conf"
	switch templateType {
	case "laravel":
		templateName = "laravel.conf"
	case "wordpress":
		templateName = "wordpress.conf"
	case "general":
		templateName = "general.conf"
	default:
		templateName = "general.conf"
	}
	
	content, err := e.RenderCaddyTemplate(templateName, data)
	if err != nil {
		return "", err
	}
	
	// Update to use custom ports (8789 for HTTP, 9879 for HTTPS)
	content = strings.ReplaceAll(content, ":8080", ":8789")
	content = strings.ReplaceAll(content, ":8443", ":9879")
	
	return content, nil
}

// WriteNginxConfig writes the generated nginx configuration to the workspace
func (e *TemplateEngine) WriteNginxConfig(config projects.Config, layout projects.Layout, templateType string) error {
	// Generate the configuration content
	content, err := e.GenerateNginxConfig(config, layout, templateType)
	if err != nil {
		// If template generation fails, create a basic config as fallback
		content = e.generateBasicConfig(config, layout)
	}
	
	// Determine nginx sites-available path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	
	nginxDir := filepath.Join(home, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")
	
	// Ensure directories exist
	for _, dir := range []string{sitesAvailable, sitesEnabled} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create nginx directory %s: %w", dir, err)
		}
	}
	
	// Write the config file with updated ports (8789 for HTTP, 9879 for HTTPS)
	content = strings.ReplaceAll(content, "listen 8080", "listen 8789")
	content = strings.ReplaceAll(content, "listen [::]:8080", "listen [::]:8789")
	content = strings.ReplaceAll(content, "listen 8443", "listen 9879")
	content = strings.ReplaceAll(content, "listen [::]:8443", "listen [::]:9879")
	
	configPath := filepath.Join(sitesAvailable, filepath.Base(layout.Root)+".conf")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write nginx config to %s: %w", configPath, err)
	}
	
	// Create symlink in sites-enabled
	enabledPath := filepath.Join(sitesEnabled, filepath.Base(layout.Root)+".conf")
	
	// Remove existing symlink if it exists
	os.Remove(enabledPath)
	
	// Create new symlink
	if err := os.Symlink(configPath, enabledPath); err != nil {
		return fmt.Errorf("create nginx symlink: %w", err)
	}
	
	return nil
}

// WriteCaddyConfig writes the generated Caddy configuration to the workspace
func (e *TemplateEngine) WriteCaddyConfig(config projects.Config, layout projects.Layout, templateType string) error {
	// Generate the configuration content
	content, err := e.GenerateCaddyConfig(config, layout, templateType)
	if err != nil {
		// If template generation fails, create a basic config as fallback
		content = e.generateBasicCaddyConfig(config, layout)
	}
	
	// Update to use custom ports (8789 for HTTP, 9879 for HTTPS)
	content = strings.ReplaceAll(content, ":8080", ":8789")
	content = strings.ReplaceAll(content, ":8443", ":9879")
	
	// Determine Caddy directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	
	caddyDir := filepath.Join(home, ".chauffeur", "caddy")
	
	// Ensure directory exists
	if err := os.MkdirAll(caddyDir, 0o755); err != nil {
		return fmt.Errorf("create caddy directory %s: %w", caddyDir, err)
	}
	
	// Read the current Caddyfile for appending new content
	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	currentContent, err := os.ReadFile(caddyfilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read caddyfile: %w", err)
	}
	
	// Calculate the section delimiters
	projectSlug := filepath.Base(layout.Root)
	startMarker := fmt.Sprintf("# Start of %s configuration", projectSlug)
	endMarker := fmt.Sprintf("# End of %s configuration", projectSlug)
	
	// Remove old configuration if it exists
	if currentContent != nil {
		contentStr := string(currentContent)
		startIdx := strings.Index(contentStr, startMarker)
		endIdx := strings.Index(contentStr, endMarker)
		
		if startIdx != -1 && endIdx != -1 {
			// Remove old section
			beforeSection := contentStr[:startIdx]
			afterSection := contentStr[endIdx + len(endMarker):]
			contentStr = beforeSection + afterSection
			
			// Remove any trailing empty lines
			contentStr = strings.TrimRight(contentStr, "\n") + "\n\n"
			currentContent = []byte(contentStr)
		}
	}
	
	// Prepare new section content
	newSection := fmt.Sprintf("%s\n%s\n%s\n", startMarker, content, endMarker)
	
	// Combine old content with new section
	var fullContent []byte
	if currentContent != nil {
		fullContent = append(currentContent, []byte(newSection)...)
	} else {
		// Start with global config if file didn't exist
		globalConfig := "{\n\tauto_https off\n}\n# Project sites are appended by chauf link\n\n"
		fullContent = append([]byte(globalConfig), []byte(newSection)...)
	}
	
	// Write the complete Caddyfile
	if err := os.WriteFile(caddyfilePath, fullContent, 0o644); err != nil {
		return fmt.Errorf("write caddyfile: %w", err)
	}
	
	return nil
}

// RemoveNginxConfig removes the nginx configuration for a project
func (e *TemplateEngine) RemoveNginxConfig(slug string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	
	nginxDir := filepath.Join(home, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")
	
	configPath := filepath.Join(sitesAvailable, slug+".conf")
	enabledPath := filepath.Join(sitesEnabled, slug+".conf")
	
	// Remove symlink first
	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove nginx enabled config: %w", err)
	}
	
	// Remove config file
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove nginx config: %w", err)
	}
	
	return nil
}

// RemoveCaddyConfig removes the Caddy configuration for a project
func (e *TemplateEngine) RemoveCaddyConfig(slug string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	
	caddyfilePath := filepath.Join(home, ".chauffeur", "caddy", "Caddyfile")
	
	// Calculate the section delimiters
	startMarker := fmt.Sprintf("# Start of %s configuration", slug)
	endMarker := fmt.Sprintf("# End of %s configuration", slug)
	
	// Read current content
	currentContent, err := os.ReadFile(caddyfilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to remove from
		}
		return fmt.Errorf("read caddyfile: %w", err)
	}
	
	// Remove project section
	contentStr := string(currentContent)
	startIdx := strings.Index(contentStr, startMarker)
	endIdx := strings.Index(contentStr, endMarker)
	
	if startIdx != -1 && endIdx != -1 {
		// Remove the section
		beforeSection := contentStr[:startIdx]
		afterSection := contentStr[endIdx + len(endMarker):]
		contentStr = beforeSection + afterSection
		
		// Remove any trailing empty lines
		contentStr = strings.TrimRight(contentStr, "\n") + "\n"
		
		// Write back without the removed section
		if err := os.WriteFile(caddyfilePath, []byte(contentStr), 0o644); err != nil {
			return fmt.Errorf("update caddyfile: %w", err)
		}
	}
	
	return nil
}

// DetectTemplateType attempts to detect the appropriate template type based on project structure
func (e *TemplateEngine) DetectTemplateType(projectPath string) string {
	// Check for Laravel indicators
	if e.fileExists(filepath.Join(projectPath, "artisan")) &&
		e.fileExists(filepath.Join(projectPath, "composer.json")) {
		// Additional Laravel check: look for typical Laravel directories
		if e.dirExists(filepath.Join(projectPath, "app")) ||
			e.dirExists(filepath.Join(projectPath, "config")) ||
			e.dirExists(filepath.Join(projectPath, "storage")) {
			return "laravel"
		}
	}
	
	// Check for WordPress indicators
	if e.fileExists(filepath.Join(projectPath, "wp-config.php")) &&
		e.fileExists(filepath.Join(projectPath, "wp-admin")) &&
		e.fileExists(filepath.Join(projectPath, "wp-includes")) {
		return "wordpress"
	}
	
	// Default to general
	return "general"
}

func (e *TemplateEngine) fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (e *TemplateEngine) dirExists(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

// generateBasicConfig creates a basic nginx configuration when templates are not available
func (e *TemplateEngine) generateBasicConfig(config projects.Config, layout projects.Layout) string {
	serverName := filepath.Base(layout.Root)
	if config.Site != nil && config.Site.Domain != "" {
		serverName = config.Site.Domain
	}

	basicConfig := fmt.Sprintf(`# Basic nginx configuration for %s
upstream %s_php {
    server unix:%s;
}

server {
    listen 8789;
    server_name %s;
    root %s;
    index index.php index.html;

    access_log %s/nginx-access.log;
    error_log %s/nginx-error.log;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass %s_php;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
    }

    location ~ /\. {
        deny all;
    }
}`, 
		serverName,
		filepath.Base(layout.Root),
		layout.SocketPath,
		serverName,
		config.Path,
		layout.LogsDir,
		layout.LogsDir,
		filepath.Base(layout.Root),
	)

	// Add SSL section if enabled
	if config.Site != nil && config.Site.SSL {
		basicConfig += fmt.Sprintf(`

server {
    listen 9879 ssl http2;
    server_name %s;
    root %s;
    index index.php index.html;

    access_log %s/nginx-access-ssl.log;
    error_log %s/nginx-error-ssl.log;

    # SSL configuration
    ssl_certificate /etc/ssl/certs/caddy.pem;
    ssl_certificate_key /etc/ssl/private/caddy.key;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass %s_php;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
        fastcgi_param HTTPS on;
    }

    location ~ /\. {
        deny all;
    }
}`, 
			serverName,
			config.Path,
			layout.LogsDir,
			layout.LogsDir,
			filepath.Base(layout.Root),
		)
	}

	return basicConfig
}

// processCaddyTemplate applies template variable substitution with simple {{VAR}} syntax for Caddy
func (e *TemplateEngine) processCaddyTemplate(content string, data CaddyTemplateData) string {
	result := content
	
	// Basic variable replacement
	replacements := map[string]string{
		"{{SERVER_NAME}}":    data.ServerName,
		"{{PROJECT_ROOT}}":   data.ProjectRoot,
		"{{PHP_FPM_SOCKET}}":  data.PHPFpmSocket,
	}
	
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Handle conditional blocks
	if !data.SiteDomain {
		// Remove site-specific section (everything between {{#SITE_DOMAIN}} and {{/SITE_DOMAIN}})
		start := strings.Index(result, "{{#SITE_DOMAIN}}")
		end := strings.Index(result, "{{/SITE_DOMAIN}}")
		
		if start != -1 && end != -1 {
			beforeSiteDomain := result[:start]
			afterSiteDomain := result[end + len("{{/SITE_DOMAIN}}"):]
			result = beforeSiteDomain + afterSiteDomain
		}
	} else {
		// Remove the conditional markers but keep the content
		result = strings.ReplaceAll(result, "{{#SITE_DOMAIN}}", "")
		result = strings.ReplaceAll(result, "{{/SITE_DOMAIN}}", "")
	}
	
	if !data.SSL {
		// Remove SSL section (everything between {{#SSL}} and {{/SSL}})
		start := strings.Index(result, "{{#SSL}}")
		end := strings.Index(result, "{{/SSL}}")
		
		if start != -1 && end != -1 {
			beforeSSL := result[:start]
			afterSSL := result[end + len("{{/SSL}}"):]
			result = beforeSSL + afterSSL
		}
	} else {
		// Remove the conditional markers but keep the content
		result = strings.ReplaceAll(result, "{{#SSL}}", "")
		result = strings.ReplaceAll(result, "{{/SSL}}", "")
	}
	
	return result
}

// generateBasicCaddyConfig creates a basic Caddy configuration when templates are not available
func (e *TemplateEngine) generateBasicCaddyConfig(config projects.Config, layout projects.Layout) string {
	serverName := layout.Root
	if config.Site != nil && config.Site.Domain != "" {
		serverName = config.Site.Domain
	}

	basicConfig := fmt.Sprintf(`# Start of %s configuration
%s:8789 {
	root %s
	php_fastcgi unix:%s {
		root %s
	}

	file_server
	
	# Security headers
	header {
		X-Content-Type-Options nosniff
		X-Frame-Options SAMEORIGIN
		X-XSS-Protection "1; mode=block"
	}

	# Handle static files
	@static {
		path *.css *.js *.png *.jpg *.jpeg *.gif *.ico *.svg
	}
	
	handle @static {
		header Cache-Control "public, max-age=31536000"
		file_server
	}
	
	# Handle PHP files
	@php {
		path *.php
	}
	
	handle @php {
		php_fastcgi unix:%s {
			root %s
		}
		rewrite /* /index.php?{query}
	}
}

%s
# End of %s configuration
`, 
		filepath.Base(layout.Root),
		serverName,
		config.Path,
		layout.SocketPath,
		config.Path,
		"",
		filepath.Base(layout.Root),
		filepath.Base(layout.Root),
	)

	// Add SSL section if enabled
	if config.Site != nil && config.Site.SSL {
	sslSection := fmt.Sprintf(`

%s:9879 {
	root %s
	php_fastcgi unix:%s {
			root %s
	}

		file_server
		tls internal
		
	# Enhanced security headers for HTTPS
		header {
			X-Content-Type-Options nosniff
			X-Frame-Options SAMEORIGIN
			X-XSS-Protection "1; mode=block"
			Strict-Transport-Security "max-age=31536000; includeSubDomains"
		}
	
	# Handle static files
	@static {
			path *.css *.js *.png *.jpg *.jpeg *.gif *.ico *.svg
		}
		
		handle @static {
			header Cache-Control "public, max-age=31536000"
			file_server
		}
		
	# Handle PHP files
		@php {
			path *.php
		}
		
		handle @php {
			php_fastcgi unix:%s {
				root %s
			}
			rewrite /* /index.php?{query}
		}
	}

# End of %s configuration
`, 
			serverName,
			config.Path,
			layout.SocketPath,
			config.Path,
			filepath.Base(layout.Root),
			filepath.Base(layout.Root),
		)
		
		basicConfig += sslSection
	}

	return basicConfig
}
