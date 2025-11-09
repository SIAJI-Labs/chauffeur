package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

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

	if phpVer == "" {
		phpVer = cfg.PHP.Default
	}
	if phpVer == "" {
		return fmt.Errorf("no PHP version specified and no default configured")
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
