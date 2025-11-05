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

// TemplateEngine handles nginx template processing
type TemplateEngine struct {
	templateDir string
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() (*TemplateEngine, error) {
	// Try multiple template directory paths
	var templateDir string
	
	// Get the template directory relative to the CLI binary location
	exePath, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "..", "..", "templates", "nginx")
		if _, err := os.Stat(candidate); err == nil {
			templateDir = candidate
		}
	}
	
	// Fallback to development path if installation path doesn't exist
	if templateDir == "" {
		// Assume we're in development environment
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}
		candidate := filepath.Join(wd, "cli", "templates", "nginx")
		if _, err := os.Stat(candidate); err == nil {
			templateDir = candidate
		} else {
			// Final fallback - create empty config generation (for testing)
			templateDir = ""
		}
	}
	
	return &TemplateEngine{
		templateDir: templateDir,
	}, nil
}

// RenderTemplate renders a nginx template with the provided data
func (e *TemplateEngine) RenderTemplate(templateName string, data TemplateData) (string, error) {
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
	
	return e.RenderTemplate(templateName, data)
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
	
	// Write the config file
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
    listen 8080;
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
    listen 8443 ssl http2;
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
