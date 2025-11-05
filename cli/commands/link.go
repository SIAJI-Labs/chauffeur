package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
)

// RunLink handles `chauf link` command invocations.
func RunLink(args []string) error {
	var (
		domain string
		phpVer string
		ssl    bool
		force  bool
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
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("initialize template engine: %w", err)
	}

	// Detect template type based on project structure
	templateType := templateEngine.DetectTemplateType(cwd)
	
	// Generate and write nginx configuration
	if err := templateEngine.WriteNginxConfig(proj, layout, templateType); err != nil {
		fmt.Printf("Warning: Failed to generate nginx configuration: %v\n", err)
		// Continue even if nginx generation fails
	}

	// Generate and write Caddy configuration
	if err := templateEngine.WriteCaddyConfig(proj, layout, templateType); err != nil {
		fmt.Printf("Warning: Failed to generate Caddy configuration: %v\n", err)
		// Continue even if Caddy generation fails
	}

	fmt.Printf("Project linked as %s\n", slug)
	fmt.Printf("  Path: %s\n", cwd)
	fmt.Printf("  Config: %s\n", layout.ConfigPath)
	fmt.Printf("  PHP: %s\n", phpVer)
	fmt.Printf("  Template: %s\n", templateType)
	
	// Access URLs - always show domain (now default to <slug>.test)
	if proj.Site != nil && proj.Site.Domain != "" {
		if strings.Contains(proj.Site.Domain, ".test") {
			fmt.Printf("  Domain: %s (default .test domain)\n", proj.Site.Domain)
		} else {
			fmt.Printf("  Domain: %s (custom domain)\n", proj.Site.Domain)
		}
		fmt.Printf("  Access: http://%s\n", proj.Site.Domain)
		if proj.Site.SSL {
			fmt.Printf("  Access Secure: https://%s\n", proj.Site.Domain)
		}
		fmt.Printf("  Note: Use --site <custom-domain> to override default domain\n")
	}

	return nil
}

func printLinkUsage() {
	fmt.Print(`Chauffeur Project Linking

Usage:
  chauf link [--site <domain>] [--ssl] [--php <version>] [--force]

Flags:
  --site <domain>   Register a local domain for the project (default: <slug>.test).
  --ssl             Enable internal TLS for the domain.
  --php <version>   Override the PHP version for this project (default: global default).
  --force           Overwrite existing project configuration.

Note:
  When --site is not specified, the project is automatically assigned a .test domain
  based on the project directory name (e.g., "my-project" -> "my-project.test").
`)
}
