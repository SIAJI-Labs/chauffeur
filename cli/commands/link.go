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
		domain string
		phpVer string
		ssl    bool
		force  bool
		caddyHTTPPort  int
		caddyHTTPSPort int
	)

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
		case "--caddy-http-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--caddy-http-port requires a port value")
			}
			portStr := args[i+1]
			if port, err := strconv.Atoi(portStr); err != nil {
				return fmt.Errorf("invalid HTTP port: %s", portStr)
			} else {
				caddyHTTPPort = port
			}
			i += 2
		case "--caddy-https-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--caddy-https-port requires a port value")
			}
			portStr := args[i+1]
			if port, err := strconv.Atoi(portStr); err != nil {
				return fmt.Errorf("invalid HTTPS port: %s", portStr)
			} else {
				caddyHTTPSPort = port
			}
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
	if caddyHTTPPort > 0 {
		// Validate custom port
		validPort, err := validator.SetPortFromCommand("caddy-http", fmt.Sprintf("%d", caddyHTTPPort))
		if err != nil {
			return err
		}
		cfg.Caddy.HTTPPort = validPort
	}
	
	if caddyHTTPSPort > 0 {
		// Validate custom port
		validPort, err := validator.SetPortFromCommand("caddy-https", fmt.Sprintf("%d", caddyHTTPSPort))
		if err != nil {
			return err
		}
		cfg.Caddy.HTTPSPort = validPort
	}

	// Validate all configured ports
	if err := validator.ValidateAllPorts(); err != nil {
		// If validation fails, check if it's due to conflicts that can be resolved
		if cfg.Ports.ConflictResolution == "fail" {
			return fmt.Errorf("port validation failed: %w", err)
		}
		
		// For auto or prompt modes, validator already handled resolution
		// Reload the config to get any updated ports
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("reload configuration after port resolution: %w", err)
		}
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
		fmt.Printf("PHP %s is not installed. Run 'chauf install php %s' first.\n", phpVer, phpVer)
		return fmt.Errorf("php %s not installed", phpVer)
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
	
	// Generate and write nginx configuration
	if err := templateEngine.WriteNginxConfig(proj, layout, templateType); err != nil {
		templateSpin.Fail("nginx configuration generation failed")
		fmt.Printf("Warning: Failed to generate nginx configuration: %v\n", err)
		// Continue even if nginx generation fails - don't return error
	} else {
		// If nginx succeeded, try Caddy
		if err := templateEngine.WriteCaddyConfig(proj, layout, templateType); err != nil {
			templateSpin.Success("nginx templates generated (caddy failed)")
			fmt.Printf("Warning: Failed to generate Caddy configuration: %v\n", err)
		} else {
			templateSpin.Success("nginx + caddy templates generated")
		}
	}

	fmt.Printf("Project linked as %s\n", slug)
	fmt.Printf("  Path: %s\n", cwd)
	fmt.Printf("  Config: %s\n", layout.ConfigPath)
	fmt.Printf("  PHP: %s\n", phpVer)
	fmt.Printf("  Template: %s\n", templateType)
	
	// Show configured ports
	fmt.Printf("  Caddy HTTP: %d\n", cfg.Caddy.HTTPPort)
	fmt.Printf("  Caddy HTTPS: %d\n", cfg.Caddy.HTTPSPort)
	
	// Access URLs - always show domain (now default to <slug>.test)
	if proj.Site != nil && proj.Site.Domain != "" {
		if strings.Contains(proj.Site.Domain, ".test") {
			fmt.Printf("  Domain: %s (default .test domain)\n", proj.Site.Domain)
		} else {
			fmt.Printf("  Domain: %s (custom domain)\n", proj.Site.Domain)
		}
		
		// Show URLs with actual ports
		httpURL := fmt.Sprintf("http://%s", proj.Site.Domain)
		if cfg.Caddy.HTTPPort != 80 {
			httpURL = fmt.Sprintf("http://%s:%d", proj.Site.Domain, cfg.Caddy.HTTPPort)
		}
		fmt.Printf("  Access: %s\n", httpURL)
		
		if proj.Site.SSL {
			httpsURL := fmt.Sprintf("https://%s", proj.Site.Domain)
			if cfg.Caddy.HTTPSPort != 443 {
				httpsURL = fmt.Sprintf("https://%s:%d", proj.Site.Domain, cfg.Caddy.HTTPSPort)
			}
			fmt.Printf("  Access Secure: %s\n", httpsURL)
		}
		
		fmt.Printf("  Note: Use --site <custom-domain> to override default domain\n")
		if caddyHTTPPort > 0 || caddyHTTPSPort > 0 {
			fmt.Printf("  Custom ports applied via --caddy-http-port/--caddy-https-port flags\n")
		}
	}

	return nil
}

func printLinkUsage() {
	fmt.Print(`Chauffeur Project Linking

Usage:
  chauf link [--site <domain>] [--ssl] [--php <version>] [--caddy-http-port <port>] [--caddy-https-port <port>] [--force]

Flags:
  --site <domain>           Register a local domain for the project (default: <slug>.test).
  --ssl                     Enable internal TLS for the domain.
  --php <version>           Override the PHP version for this project (default: global default).
  --caddy-http-port <port>  Override Caddy HTTP port for this project (default: from config).
  --caddy-https-port <port> Override Caddy HTTPS port for this project (default: from config).
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
