package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/projects"
	nginxtemplates "github.com/siaji/chauffeur/cli/templates/nginx"
)

// TemplateData represents the data available to nginx templates
type TemplateData struct {
	ProjectSlug  string
	ServerName   string
	ProjectRoot  string
	PHPFpmSocket string
	LogsDir      string
	SSL          bool
	HTTPPort     int
	HTTPSPort    int
	SSLCertPath  string
	SSLKeyPath   string
}

// NginxConfigOptions carries runtime values required when rendering templates.
type NginxConfigOptions struct {
	HTTPPort    int
	HTTPSPort   int
	SSLCertPath string
	SSLKeyPath  string
}

func (o *NginxConfigOptions) applyDefaults() {
	if o.HTTPPort == 0 {
		o.HTTPPort = 8080
	}
	if o.HTTPSPort == 0 {
		o.HTTPSPort = 8443
	}
}

// TemplateEngine handles nginx template processing
type TemplateEngine struct {
	templateDir string
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() (*TemplateEngine, error) {
	var nginxDir string

	// Get the template directory relative to the CLI binary location
	if exePath, err := os.Executable(); err == nil {
		nginxCandidate := filepath.Join(filepath.Dir(exePath), "..", "..", "templates", "nginx")
		if _, err := os.Stat(nginxCandidate); err == nil {
			nginxDir = nginxCandidate
		}
	}

	// Fallback to development path if installation path doesn't exist
	if nginxDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}

		nginxDevCandidate := filepath.Join(wd, "cli", "templates", "nginx")
		if _, err := os.Stat(nginxDevCandidate); err == nil {
			nginxDir = nginxDevCandidate
		}
	}

	return &TemplateEngine{
		templateDir: nginxDir,
	}, nil
}

// DetectTemplateType inspects a project directory and returns the best-fit nginx template.
func (e *TemplateEngine) DetectTemplateType(projectPath string) string {
	if projectPath == "" {
		return "general"
	}

	// Laravel detection: artisan + public directory typically present
	if fileExists(filepath.Join(projectPath, "artisan")) && dirExists(filepath.Join(projectPath, "public")) {
		return "laravel"
	}

	// WordPress detection: wp-config.php or wp-includes/wp-admin directories
	if fileExists(filepath.Join(projectPath, "wp-config.php")) ||
		dirExists(filepath.Join(projectPath, "wp-includes")) ||
		dirExists(filepath.Join(projectPath, "wp-admin")) {
		return "wordpress"
	}

	return "general"
}

// RenderNginxTemplate renders a nginx template with the provided data
func (e *TemplateEngine) RenderNginxTemplate(templateName string, data TemplateData) (string, error) {
	var content []byte
	var err error

	if templateName == "" {
		templateName = "general.conf"
	}

	if e.templateDir != "" {
		templatePath := filepath.Join(e.templateDir, templateName)
		content, err = os.ReadFile(templatePath)
	} else {
		content, err = nginxtemplates.Files.ReadFile(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("read nginx template %s: %w", templateName, err)
	}

	return e.processTemplate(string(content), data)
}

// processTemplate applies template variable substitution with simple {{VAR}} syntax
func (e *TemplateEngine) processTemplate(content string, data TemplateData) (string, error) {
	result := content

	// Basic variable replacement
	replacements := map[string]string{
		"{{PROJECT_SLUG}}":     data.ProjectSlug,
		"{{SERVER_NAME}}":      data.ServerName,
		"{{PROJECT_ROOT}}":     data.ProjectRoot,
		"{{PHP_FPM_SOCKET}}":   data.PHPFpmSocket,
		"{{LOGS_DIR}}":         data.LogsDir,
		"{{NGINX_HTTP_PORT}}":  fmt.Sprintf("%d", data.HTTPPort),
		"{{NGINX_HTTPS_PORT}}": fmt.Sprintf("%d", data.HTTPSPort),
		"{{SSL_CERT_PATH}}":    data.SSLCertPath,
		"{{SSL_KEY_PATH}}":     data.SSLKeyPath,
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
			afterSSL := result[end+len("{{/SSL}}"):]
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
func (e *TemplateEngine) GenerateNginxConfig(config projects.Config, layout projects.Layout, templateType string, opts NginxConfigOptions) (string, error) {
	opts.applyDefaults()

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
		HTTPPort:     opts.HTTPPort,
		HTTPSPort:    opts.HTTPSPort,
		SSLCertPath:  opts.SSLCertPath,
		SSLKeyPath:   opts.SSLKeyPath,
	}

	// Configure SSL if enabled
	if config.Site != nil && config.Site.SSL {
		data.SSL = true
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

// WriteNginxConfig writes the generated nginx configuration to the workspace
func (e *TemplateEngine) WriteNginxConfig(config projects.Config, layout projects.Layout, templateType string, opts NginxConfigOptions) error {
	if config.Site != nil && config.Site.SSL {
		if opts.SSLCertPath == "" || opts.SSLKeyPath == "" {
			return fmt.Errorf("ssl enabled but certificate paths were not provided")
		}
	}

	// Generate the configuration content
	content, err := e.GenerateNginxConfig(config, layout, templateType, opts)
	if err != nil {
		// If template generation fails, create a basic config as fallback
		content = e.generateBasicConfig(config, layout, opts)
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

	if err := e.ensureDefaultCatchallServer(sitesAvailable, sitesEnabled, opts.HTTPPort); err != nil {
		return err
	}

	// Ensure cert directory exists when SSL is enabled
	if config.Site != nil && config.Site.SSL && opts.SSLCertPath != "" {
		if err := os.MkdirAll(filepath.Dir(opts.SSLCertPath), 0o755); err != nil {
			return fmt.Errorf("create ssl directory: %w", err)
		}
	}

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

// RemoveNginxConfig deletes the generated nginx configuration for the provided slug.
func (e *TemplateEngine) RemoveNginxConfig(projectSlug string) error {
	if projectSlug == "" {
		return fmt.Errorf("project slug is required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	nginxDir := filepath.Join(home, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	configPath := filepath.Join(sitesAvailable, projectSlug+".conf")
	enabledPath := filepath.Join(sitesEnabled, projectSlug+".conf")

	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove nginx enabled config: %w", err)
	}
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove nginx available config: %w", err)
	}

	return nil
}

// EnsureCatchallServer makes sure the default 404 server exists for the configured HTTP port.
func (e *TemplateEngine) EnsureCatchallServer(httpPort int) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	nginxDir := filepath.Join(home, ".chauffeur", "nginx")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	for _, dir := range []string{sitesAvailable, sitesEnabled} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create nginx directory %s: %w", dir, err)
		}
	}

	return e.ensureDefaultCatchallServer(sitesAvailable, sitesEnabled, httpPort)
}

func (e *TemplateEngine) ensureDefaultCatchallServer(sitesAvailable, sitesEnabled string, httpPort int) error {
	if httpPort == 0 {
		httpPort = 8080
	}

	const defaultName = "000-default.conf"
	availablePath := filepath.Join(sitesAvailable, defaultName)
	enabledPath := filepath.Join(sitesEnabled, defaultName)

	if _, err := os.Stat(availablePath); err == nil {
		// ensure symlink exists
		if _, err := os.Lstat(enabledPath); os.IsNotExist(err) {
			_ = os.Symlink(availablePath, enabledPath)
		}
		return nil
	}

	content := fmt.Sprintf(`# Chauffeur default catch-all server
server {
    listen %d default_server;
    listen [::]:%d default_server;
    server_name _;
    return 404;
}
`, httpPort, httpPort)

	if err := os.WriteFile(availablePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write default nginx catch-all: %w", err)
	}

	if err := os.Symlink(availablePath, enabledPath); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link default nginx catch-all: %w", err)
	}
	return nil
}

// generateBasicConfig creates a basic nginx configuration when templates are not available
func (e *TemplateEngine) generateBasicConfig(config projects.Config, layout projects.Layout, opts NginxConfigOptions) string {
	opts.applyDefaults()

	serverName := filepath.Base(layout.Root)
	if config.Site != nil && config.Site.Domain != "" {
		serverName = config.Site.Domain
	}

	base := fmt.Sprintf(`# Basic nginx configuration for %s
upstream %s_php {
    server unix:%s;
}

server {
    listen %d;
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
}
`,
		serverName,
		filepath.Base(layout.Root),
		layout.SocketPath,
		opts.HTTPPort,
		serverName,
		config.Path,
		layout.LogsDir,
		layout.LogsDir,
		filepath.Base(layout.Root),
	)

	if config.Site != nil && config.Site.SSL {
		base += fmt.Sprintf(`
server {
    listen %d ssl http2;
    server_name %s;
    root %s;
    index index.php index.html;

    access_log %s/nginx-access-ssl.log;
    error_log %s/nginx-error-ssl.log;

    ssl_certificate %s;
    ssl_certificate_key %s;
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
}
`,
			opts.HTTPSPort,
			serverName,
			config.Path,
			layout.LogsDir,
			layout.LogsDir,
			opts.SSLCertPath,
			opts.SSLKeyPath,
			filepath.Base(layout.Root),
		)
	}

	return base
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
