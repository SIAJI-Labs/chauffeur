package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

// Security: Input validation patterns
var (
	// Safe domain pattern - allows only alphanumeric, hyphens, and dots for domain names
	safeDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	// Safe port range for validation
	minPort = 1
	maxPort = 65535

	// Reserved ports that should not be used without special privileges
	privilegedPorts = 1024
)

// validateDomain ensures domain names are safe and properly formatted
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("domain name too long")
	}

	if !safeDomainPattern.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	// Ensure domain ends with .test for security
	if !strings.HasSuffix(domain, ".test") {
		return fmt.Errorf("domain must end with .test: %s", domain)
	}

	return nil
}

// validatePHPVersion ensures PHP version is in expected format
func validatePHPVersion(version string) error {
	if version == "" {
		return fmt.Errorf("PHP version cannot be empty")
	}

	// Allow only major.minor format (e.g., "7.4", "8.0", "8.1")
	validVersions := map[string]bool{
		"7.4": true, "8.0": true, "8.1": true, "8.2": true, "8.3": true, "8.4": true,
	}

	if !validVersions[version] {
		return fmt.Errorf("unsupported PHP version: %s", version)
	}

	return nil
}

// validatePort ensures port numbers are within valid range
func validatePort(port int) error {
	if port < minPort || port > maxPort {
		return fmt.Errorf("port must be between %d and %d", minPort, maxPort)
	}

	// Warn about privileged ports
	if port < privilegedPorts {
		// Note: This is just a warning, not an error
		fmt.Printf("Warning: Port %d is a privileged port (< 1024)\n", port)
	}

	return nil
}

// validateProjectPath ensures the project path is safe and within allowed directories
func validateProjectPath(path string) error {
	if path == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	// Prevent path traversal attempts
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Check if path actually exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("project path does not exist: %s", absPath)
	}

	if !info.IsDir() {
		return fmt.Errorf("project path must be a directory: %s", absPath)
	}

	return nil
}

// RunLink handles `chauf link` command invocations.
func RunLink(args []string) error {
	var (
		domain    string
		phpVer    string
		ssl       bool
		force     bool
		httpPort  int
		httpsPort int
	)

	logger := lib.NewCommandLogger("link")

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printLinkUsage()
			return nil
		case "--site":
			if i+1 >= len(args) {
				return fmt.Errorf("--site requires a domain value")
			}
			domain = args[i+1]
			i += 2
		case "--php":
			if i+1 >= len(args) {
				return fmt.Errorf("--php requires a version value")
			}
			phpVer = args[i+1]
			i += 2
		case "--ssl":
			ssl = true
			i++
		case "--force":
			force = true
			i++
		case "--http-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--http-port requires a port value")
			}
			portStr := args[i+1]
			val, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid HTTP port: %s", portStr)
			}
			// Security: Validate port range
			if err := validatePort(val); err != nil {
				return fmt.Errorf("invalid HTTP port: %w", err)
			}
			httpPort = val
			i += 2
		case "--https-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--https-port requires a port value")
			}
			portStr := args[i+1]
			val, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid HTTPS port: %s", portStr)
			}
			// Security: Validate port range
			if err := validatePort(val); err != nil {
				return fmt.Errorf("invalid HTTPS port: %w", err)
			}
			httpsPort = val
			i += 2
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for link: %s", arg)
			}
			return fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	// SSL now works with both explicit and default .test domains

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Create port validator
	validator, err := lib.NewPortValidator(cfg)
	if err != nil {
		return fmt.Errorf("create port validator: %w", err)
	}

	// Handle custom port overrides
	if httpPort > 0 {
		validPort, err := validator.SetPortFromCommand("nginx-http", fmt.Sprintf("%d", httpPort))
		if err != nil {
			return err
		}
		cfg.Nginx.HTTPPort = validPort
	}

	if httpsPort > 0 {
		validPort, err := validator.SetPortFromCommand("nginx-https", fmt.Sprintf("%d", httpsPort))
		if err != nil {
			return err
		}
		cfg.Nginx.HTTPSPort = validPort
	}

	// Validate all configured ports
	if err := validator.ValidateAllPorts(); err != nil {
		// If validation fails, check if it's due to conflicts that can be resolved
		if cfg.Ports.ConflictResolution == "fail" {
			return fmt.Errorf("port validation failed: %w", err)
		}
	}

	// Always reload configuration after validation in case ports were updated
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("reload configuration after port validation: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine current directory: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	// Security: Validate project path to prevent path traversal
	if err := validateProjectPath(cwd); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	if phpVer == "" {
		phpVer = cfg.PHP.Default
	}
	if phpVer == "" {
		return fmt.Errorf("no PHP version specified and no default configured")
	}

	// Security: Validate PHP version format
	if err := validatePHPVersion(phpVer); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// Validate that the requested PHP version is installed
	if !projects.IsPHPVersionInstalled(phpVer) {
		return logger.Error(
			fmt.Sprintf("PHP %s is not installed", phpVer),
			fmt.Sprintf("Run 'chauf install php %s' first", phpVer),
		)
	}

	slug := projects.Slugify(filepath.Base(cwd))
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
	if err != nil {
		return err
	}

	// Set default domain if none provided
	if domain == "" {
		domain = slug + ".test"
	}

	// Security: Validate domain format and safety
	if err := validateDomain(domain); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	proj := projects.Config{
		Version: projects.ConfigVersion,
		Path:    cwd,
		PHP:     phpVer,
		Runtime: projects.Runtime{
			PHPFPM: layout.SocketPath,
		},
		CreatedAt: time.Now().UTC(),
	}

	proj.Site = &projects.Site{
		Domain: domain,
		SSL:    ssl,
	}

	if err := projects.WriteConfig(proj, layout.ConfigPath, force); err != nil {
		return err
	}

	// Generate nginx template
	templateSpin := lib.NewSpinner("link", "Generating templates")
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		templateSpin.Fail("template engine initialization failed")
		return fmt.Errorf("initialize template engine: %w", err)
	}

	// Detect template type based on project structure
	templateType := templateEngine.DetectTemplateType(cwd)

	// Prepare nginx rendering options
	nginxOptions := templates.NginxConfigOptions{
		HTTPPort:  cfg.Nginx.HTTPPort,
		HTTPSPort: cfg.Nginx.HTTPSPort,
	}
	if proj.Site != nil && proj.Site.SSL {
		certBase := proj.Site.Domain
		if certBase == "" {
			certBase = slug
		}
		certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
		nginxOptions.SSLCertPath = filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
		nginxOptions.SSLKeyPath = filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))
	}

	// Generate and write nginx configuration
	if err := templateEngine.WriteNginxConfig(proj, layout, templateType, nginxOptions); err != nil {
		templateSpin.Fail("nginx configuration generation failed")
		logger.Warn("Failed to generate nginx configuration", err.Error())
		// Continue even if nginx generation fails - don't return error
	} else {
		templateSpin.Success("nginx templates generated")
	}

	logger.PrintSection(fmt.Sprintf("Project linked as %s", slug))
	logger.Info(fmt.Sprintf("Path: %s", cwd))
	logger.Info(fmt.Sprintf("Config: %s", layout.ConfigPath))
	logger.Info(fmt.Sprintf("PHP: %s", phpVer))
	logger.Info(fmt.Sprintf("Template: %s", templateType))
	logger.Info(fmt.Sprintf("Nginx HTTP: %d", cfg.Nginx.HTTPPort))
	logger.Info(fmt.Sprintf("Nginx HTTPS: %d", cfg.Nginx.HTTPSPort))

	// Access URLs - always show domain (now default to <slug>.test)
	if proj.Site != nil && proj.Site.Domain != "" {
		if strings.Contains(proj.Site.Domain, ".test") {
			logger.Info(fmt.Sprintf("Domain: %s (default .test domain)", proj.Site.Domain))
		} else {
			logger.Info(fmt.Sprintf("Domain: %s (custom domain)", proj.Site.Domain))
		}

		// Show URLs with actual ports
		httpURL := fmt.Sprintf("http://%s", proj.Site.Domain)
		if cfg.Nginx.HTTPPort != 80 {
			httpURL = fmt.Sprintf("http://%s:%d", proj.Site.Domain, cfg.Nginx.HTTPPort)
		}
		logger.Info(fmt.Sprintf("Access: %s", httpURL))

		if proj.Site.SSL {
			httpsURL := fmt.Sprintf("https://%s", proj.Site.Domain)
			if cfg.Nginx.HTTPSPort != 443 {
				httpsURL = fmt.Sprintf("https://%s:%d", proj.Site.Domain, cfg.Nginx.HTTPSPort)
			}
			logger.Info(fmt.Sprintf("Access Secure: %s", httpsURL))
		}

		logger.Info("Note: Use --site <custom-domain> to override default domain")
		if httpPort > 0 || httpsPort > 0 {
			logger.Info("Custom ports applied via --http-port/--https-port flags")
		}
	}

	return nil
}

func printLinkUsage() {
	fmt.Print(`Chauffeur Project Linking

Usage:
  chauf link [--site <domain>] [--ssl] [--php <version>] [--http-port <port>] [--https-port <port>] [--force]

Flags:
  --site <domain>           Register a local domain for the project (default: <slug>.test).
  --ssl                     Enable internal TLS for the domain.
  --php <version>           Override the PHP version for this project (default: global default).
  --http-port <port>        Override Nginx HTTP port for this project (default: from config).
  --https-port <port>       Override Nginx HTTPS port for this project (default: from config).
  --force                   Overwrite existing project configuration.

Port Management:
  If specified ports are already in use, Chauffeur will:
    - Prompt for alternative ports (default behavior)
    - Auto-resolve to available ports (if "conflict_resolution: auto" in config)
    - Fail with error if "conflict_resolution: fail" in config

Note:
  When --site is not specified, the project is automatically assigned a .test domain
  based on the project directory name (e.g., "my-project" -> "my-project.test").
`)
}
