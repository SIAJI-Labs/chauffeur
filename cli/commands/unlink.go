package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/example"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

// Security: Input validation patterns
var (
	// Safe slug pattern - allows only alphanumeric, hyphens, and underscores
	safeSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// Safe domain pattern for unlink - allows only alphanumeric, hyphens, and dots
	safeUnlinkDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	// Maximum path length to prevent buffer overflow attacks
	maxPathLength = 4096
)

// validateSlug ensures project slugs are safe
func validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}

	if len(slug) > 100 {
		return fmt.Errorf("slug too long")
	}

	if !safeSlugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug format: %s", slug)
	}

	return nil
}

// validateUnlinkDomain ensures domain names are safe
func validateUnlinkDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("domain name too long")
	}

	if !safeUnlinkDomainPattern.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	return nil
}

// validateUnlinkPath ensures project paths are safe and within allowed directories
func validateUnlinkPath(path string) error {
	if path == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	if len(path) > maxPathLength {
		return fmt.Errorf("project path too long")
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

// RunUnlink handles `chauf unlink` command invocations.
func RunUnlink(args []string) error {
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

	var (
		slug    string
		domain  string
		project string
		all     bool
		force   bool
		aliases []string
	)

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printUnlinkUsage()
			return nil
		case "--force":
			force = true
			i++
		case "--slug":
			if i+1 >= len(args) {
				return fmt.Errorf("--slug requires a slug value")
			}
			slug = args[i+1]
			// Security: Validate slug format
			if err := validateSlug(slug); err != nil {
				return fmt.Errorf("security validation failed: %w", err)
			}
			i += 2
		case "--site":
			if i+1 >= len(args) {
				return fmt.Errorf("--site requires a domain value")
			}
			domain = args[i+1]
			// Security: Validate domain format
			if err := validateUnlinkDomain(domain); err != nil {
				return fmt.Errorf("security validation failed: %w", err)
			}
			i += 2
		case "--project":
			if i+1 >= len(args) {
				return fmt.Errorf("--project requires a path value")
			}
			project = args[i+1]
			// Security: Validate project path
			if err := validateUnlinkPath(project); err != nil {
				return fmt.Errorf("security validation failed: %w", err)
			}
			i += 2
		case "--all":
			all = true
			i++
		case "--alias":
			if i+1 >= len(args) {
				return fmt.Errorf("--alias requires a domain value")
			}
			aliasDomain := args[i+1]
			// Security: Validate alias domain format
			if err := validateUnlinkDomain(aliasDomain); err != nil {
				return fmt.Errorf("security validation failed: %w", err)
			}
			aliases = append(aliases, aliasDomain)
			i += 2
		default:
			return fmt.Errorf("unknown flag for unlink: %s", arg)
		}
	}

	logger := lib.NewCommandLogger("unlink")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Handle --all flag
	if all {
		return unlinkAllProjects(logger, cfg, force)
	}

	// Handle --alias flag: remove specific aliases from existing project
	if len(aliases) > 0 {
		return handleRemoveAliases(logger, &cfg, aliases, force)
	}

	// Initialize variables
	var projectPath string
	var projectSlug string

	// Handle default case: no flags provided, use current directory
	if !all && slug == "" && domain == "" && project == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine current directory: %w", err)
		}
		logger.Info(fmt.Sprintf("Default unlink case - all=%v, slug='%s', domain='%s', project='%s'", all, slug, domain, project))
		logger.Info(fmt.Sprintf("Current working directory: %s", cwd))
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return fmt.Errorf("resolve project path: %w", err)
		}

		// Security: Validate current working directory path
		if err := validateUnlinkPath(cwd); err != nil {
			return fmt.Errorf("security validation failed: %w", err)
		}

		// Check if current directory is registered
		_, layout, err := projects.FindByPath(cfg.ProjectsDir, cwd)
		if err != nil {
			return fmt.Errorf("current directory is not a registered project")
		}

		projectPath = cwd
		projectSlug = filepath.Base(layout.Root)
	} else {
		// When flags are provided, validate that exactly one identifier is given
		identifiers := 0
		if slug != "" {
			identifiers++
		}
		if domain != "" {
			identifiers++
		}
		if project != "" {
			identifiers++
		}

		if identifiers != 1 {
			return fmt.Errorf("must provide exactly one of: --slug <slug>, --site <domain>, --project <path>")
		}
	}

	// Determine project to unlink if flags were provided
	if projectPath == "" && projectSlug == "" {
		switch {
		case slug != "":
			// Unlink by slug
			projectSlug = slug
			layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
			if err != nil {
				return fmt.Errorf("access project: %w", err)
			}

			_, err = projects.LoadConfig(layout.ConfigPath)
			if err != nil {
				return fmt.Errorf("project %s is not registered", slug)
			}

		case domain != "":
			// Unlink by site domain
			found := false
			entries, err := os.ReadDir(cfg.ProjectsDir)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no projects registered")
				}
				return fmt.Errorf("read projects directory: %w", err)
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				aSlug := entry.Name()
				layout, err := projects.EnsureLayout(cfg.ProjectsDir, aSlug)
				if err != nil {
					continue
				}

				projCfg, err := projects.LoadConfig(layout.ConfigPath)
				if err != nil {
					continue
				}

				if projCfg.Site != nil && projCfg.Site.Domain == domain {
					projectSlug = aSlug
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("no project registered with domain %s", domain)
			}

		case project != "":
			// Unlink by project directory path
			// Note: project path already validated during argument parsing
			absProject, err := filepath.Abs(project)
			if err != nil {
				return fmt.Errorf("resolve project path: %w", err)
			}
			projectPath = absProject

			// Find the project by path
			_, layout, err := projects.FindByPath(cfg.ProjectsDir, projectPath)
			if err != nil {
				return fmt.Errorf("project at path %s is not registered", project)
			}

			// Extract slug from layout path
			projectSlug = filepath.Base(layout.Root)
		}
	}

	// Determine project to unlink
	switch {
	case slug != "":
		// Unlink by slug
		projectSlug = slug
		layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
		if err != nil {
			return fmt.Errorf("access project: %w", err)
		}

		_, err = projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			return fmt.Errorf("project %s is not registered", slug)
		}

	case domain != "":
		// Unlink by site domain
		found := false
		entries, err := os.ReadDir(cfg.ProjectsDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no projects registered")
			}
			return fmt.Errorf("read projects directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			slug := entry.Name()
			layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
			if err != nil {
				continue
			}

			projCfg, err := projects.LoadConfig(layout.ConfigPath)
			if err != nil {
				continue
			}

			if projCfg.Site != nil && projCfg.Site.Domain == domain {
				projectSlug = slug
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("no project registered with domain %s", domain)
		}

	case project != "":
		// Unlink by project directory path
		absProject, err := filepath.Abs(project)
		if err != nil {
			return fmt.Errorf("resolve project path: %w", err)
		}
		projectPath = absProject

		// Find the project by path
		logger.Info(fmt.Sprintf("Looking for project at: %s", projectPath))
		_, layout, err := projects.FindByPath(cfg.ProjectsDir, projectPath)
		if err != nil {
			return fmt.Errorf("project at path %s is not registered", project)
		}

		// Extract slug from layout path
		projectSlug = filepath.Base(layout.Root)
	}

	// Show project details before unlinking
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, projectSlug)
	if err != nil {
		return fmt.Errorf("access project: %w", err)
	}

	projCfg, err := projects.LoadConfig(layout.ConfigPath)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	if !force {
		logger.PrintSection("Project to unlink")
		logger.Info(fmt.Sprintf("  Slug: %s", projectSlug))
		logger.Info(fmt.Sprintf("  Path: %s", projCfg.Path))
		logger.Info(fmt.Sprintf("  PHP: %s", projCfg.PHP))

		// Show domain information
		if projCfg.Site != nil {
			logger.Info(fmt.Sprintf("  Primary Domain: %s (ssl=%t)", projCfg.Site.Domain, projCfg.Site.SSL))
		}

		// Check if this is the example project and provide special guidance
		isExampleProject := (projectSlug == example.ExampleProjectName)
		if isExampleProject {
			logger.Info("📁 This is the example project created by Chauffeur")
			logger.Info("   It's safe to remove it when you're ready to work on your own projects")
			logger.Info("")
		}

		// Show alias domains if they exist
		if projCfg.Domains != nil && len(projCfg.Domains.Aliases) > 0 {
			logger.Info("  Alias Domains:")
			for _, alias := range projCfg.Domains.Aliases {
				sslStatus := "HTTP"
				if alias.SSL {
					sslStatus = "HTTPS"
				}
				logger.Info(fmt.Sprintf("    - %s (%s)", alias.Domain, sslStatus))
			}
		}

		promptMessage := "This will remove the project registration and all associated configuration"
		if isExampleProject {
			promptMessage = "This will remove the example project and all associated configuration"
		}
		logger.Prompt(promptMessage, "Use --force to skip confirmation")
		if !confirmUnlinkAction(logger, "Continue unlinking this project?") {
			logger.Info("Unlink cancelled by user.")
			return nil
		}
	}

	// Remove nginx configuration
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("initialize template engine: %w", err)
	}

	removed := false
	if err := templateEngine.RemoveNginxConfig(projectSlug); err != nil {
		logger.Warn("Failed to remove nginx configuration", err.Error())
		// Continue with unlink even if nginx removal fails
	} else {
		removed = true
	}

	if err := templateEngine.EnsureCatchallServer(cfg.Nginx.HTTPPort); err != nil {
		logger.Warn("Failed to ensure default nginx catch-all", err.Error())
	}

	if removed {
		restartNginxIfNeeded(logger)
	}

	// Remove SSL certificates if they exist
	if projCfg.Site != nil && projCfg.Site.SSL {
		certBase := projCfg.Site.Domain
		if certBase == "" {
			certBase = projectSlug
		}
		certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
		certPath := filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
		keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))

		// Remove SSL certificates and provide structured logging
		certExisted, wasMkcert, err := lib.RemoveSSLCertificate(certPath, keyPath, certBase)
		provideSSLCertificateRemovalGuidance(logger, certBase, certExisted, wasMkcert, err)
	}

	// Remove the project directory
	projectDir := layout.Root

	// Special handling for example project
	isExampleProject := (projectSlug == example.ExampleProjectName)
	if isExampleProject {
		// Example project is managed by Chauffeur, remove it using helper
		if err := example.RemoveExampleProject(); err != nil {
			logger.Warn("Failed to completely remove example project", err.Error())
		} else {
			logger.Success("Example project removed successfully", "")
			logger.Info("You can recreate it anytime by running:")
			logger.Info("  chauf install nginx php")
			return nil
		}
	} else {
		// Regular project removal
		if err := os.RemoveAll(projectDir); err != nil {
			return fmt.Errorf("remove project directory: %w", err)
		}
		logger.Success("Successfully unlinked project", projectSlug)
	}

	return nil
}

func unlinkAllProjects(logger *lib.Logger, cfg config.Config, force bool) error {
	entries, err := os.ReadDir(cfg.ProjectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no projects registered")
		}
		return fmt.Errorf("read projects directory: %w", err)
	}

	if len(entries) == 0 {
		logger.Info("No projects registered.")
		return nil
	}

	var allProjectsInfo []unlinkProject
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
		if err != nil {
			continue
		}

		projCfg, err := projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			continue
		}

		allProjectsInfo = append(allProjectsInfo, unlinkProject{
			Slug:      slug,
			Path:      projCfg.Path,
			PHP:       projCfg.PHP,
			Site:      projCfg.Site,
			CreatedAt: projCfg.CreatedAt,
		})
	}

	if len(allProjectsInfo) == 0 {
		logger.Info("No valid projects found.")
		return nil
	}

	logger.PrintSection(fmt.Sprintf("Found %d project(s) to unlink", len(allProjectsInfo)))
	for _, proj := range allProjectsInfo {
		row := fmt.Sprintf("  %s: %s", proj.Slug, proj.Path)
		if proj.Site != nil {
			row += fmt.Sprintf(" (domain: %s", proj.Site.Domain)
			if proj.Site.SSL {
				row += ", ssl=true"
			}
			row += ")"
		}
		row += fmt.Sprintf(" (php: %s)", proj.PHP)
		logger.Info(row)
	}

	if !force {
		logger.Prompt("This will remove ALL registered projects and their configurations", "Use --force to skip confirmation")
		if !confirmUnlinkAction(logger, "Continue unlinking all projects?") {
			logger.Info("Unlink cancelled by user.")
			return nil
		}
	}

	// Remove nginx configurations first
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		logger.Warn("Failed to initialize template engine", err.Error())
		// Continue with unlink even if template engine fails
	} else {
		removedAny := false
		for _, proj := range allProjectsInfo {
			if err := templateEngine.RemoveNginxConfig(proj.Slug); err != nil {
				logger.Warn(fmt.Sprintf("Failed to remove nginx configuration for %s", proj.Slug), err.Error())
				// Continue even if nginx removal fails
				continue
			}
			removedAny = true
		}
		if err := templateEngine.EnsureCatchallServer(cfg.Nginx.HTTPPort); err != nil {
			logger.Warn("Failed to ensure default nginx catch-all", err.Error())
		}
		if removedAny {
			restartNginxIfNeeded(logger)
		}
	}

	for _, proj := range allProjectsInfo {
		layout, err := projects.EnsureLayout(cfg.ProjectsDir, proj.Slug)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to access project %s", proj.Slug), err.Error())
			continue
		}

		// Remove SSL certificates if they exist
		if proj.Site != nil && proj.Site.SSL {
			certBase := proj.Site.Domain
			if certBase == "" {
				certBase = proj.Slug
			}
			certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
			certPath := filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
			keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))

			// Remove SSL certificates and provide structured logging for bulk operations
			certExisted, wasMkcert, err := lib.RemoveSSLCertificate(certPath, keyPath, certBase)
			if certExisted {
				if err != nil {
					logger.Warn(fmt.Sprintf("Failed to remove SSL certificate for %s", proj.Slug), err.Error())
				} else {
					certType := "self-signed"
					if wasMkcert {
						certType = "mkcert (trusted)"
					}
					logger.Info(fmt.Sprintf("✓ SSL certificate removed (%s): %s", certType, certBase))
				}
			}
		}

		if err := os.RemoveAll(layout.Root); err != nil {
			logger.Warn(fmt.Sprintf("Failed to remove project %s", proj.Slug), err.Error())
			continue
		}
	}

	logger.Success(fmt.Sprintf("Successfully unlinked %d project(s)", len(allProjectsInfo)), "")
	return nil
}

type unlinkProject struct {
	Slug      string
	Path      string
	PHP       string
	Site      *projects.Site
	CreatedAt time.Time
}

// handleRemoveAliases removes specific alias domains from an existing project
func handleRemoveAliases(logger *lib.Logger, cfg *config.Config, aliases []string, force bool) error {
	// Get current working directory to find the project
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine current directory: %w", err)
	}

	// Find existing project
	proj, layout, err := projects.FindByPath(cfg.ProjectsDir, cwd)
	if err != nil {
		return logger.Error(
			"No linked project found in current directory",
			"Use 'chauf link' to create a new project first",
		)
	}

	logger.Info(fmt.Sprintf("Found linked project: %s", layout.Root))

	// Check if project has any aliases
	if proj.Domains == nil || len(proj.Domains.Aliases) == 0 {
		logger.Info("No alias domains configured for this project")
		return nil
	}

	// Remove each alias
	removedAliases := 0
	for _, alias := range aliases {
		if err := proj.RemoveAlias(alias); err != nil {
			if !force {
				return logger.Error(fmt.Sprintf("Failed to remove alias %s", alias), err.Error())
			}
			logger.Warn(fmt.Sprintf("Failed to remove alias %s", alias), err.Error())
			continue
		}
		logger.Success(fmt.Sprintf("Removed alias: %s", alias), "")
		removedAliases++
	}

	if removedAliases == 0 {
		logger.Info("No aliases were removed")
		return nil
	}

	// Save updated configuration
	if err := projects.WriteConfig(proj, layout.ConfigPath, true); err != nil {
		return logger.Error("Failed to save project configuration", err.Error())
	}

	// Regenerate nginx configuration with updated domains
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return logger.Error("Template engine initialization failed", err.Error())
	}

	templateType := templateEngine.DetectTemplateType(cwd)
	nginxOptions := templates.NginxConfigOptions{
		HTTPPort:  cfg.Nginx.HTTPPort,
		HTTPSPort: cfg.Nginx.HTTPSPort,
	}

	// Handle SSL certificate paths if SSL is enabled
	if proj.HasSSLEnabled() {
		certBase := proj.GetPrimaryDomain()
		if certBase == "" {
			certBase = projects.Slugify(filepath.Base(cwd))
		}
		certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
		nginxOptions.SSLCertPath = filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
		nginxOptions.SSLKeyPath = filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))
	}

	// Generate and write nginx configuration
	if err := templateEngine.WriteNginxConfig(proj, layout, templateType, nginxOptions); err != nil {
		return logger.Error("Failed to regenerate nginx configuration", err.Error())
	}

	// Note: nginx should be restarted to apply changes
	logger.Info("Restart nginx to apply changes: chauf restart")

	logger.Success("Aliases removed successfully", fmt.Sprintf("Total removed: %d", removedAliases))
	return nil
}

func printUnlinkUsage() {
	logger := lib.NewCommandLogger("unlink")
	logger.PrintBlock(`Chauffeur Project Unlinking

Usage:
  chauf unlink                       Unlink current directory (default)
  chauf unlink --slug <slug>          Unlink project by slug
  chauf unlink --site <domain>       Unlink project by domain
  chauf unlink --project <path>       Unlink project by directory path
  chauf unlink --all                 Unlink all registered projects
  chauf unlink --alias <domain>       Remove alias domains from current project

Flags:
  --slug <slug>                     Remove project by slug.
  --site <domain>                   Remove project by domain.
  --project <path>                  Remove project by directory path.
  --alias <domain>                  Remove alias domain(s) from current project (can be used multiple times).
  --all                             Remove all registered projects.
  --force                           Proceed without confirmation.
  --help, -h                        Show this help message.

Examples:
  chauf unlink                       # unlink current project if linked
  chauf unlink --slug myproject
  chauf unlink --site myproject.test
  chauf unlink --project /path/to/project
  chauf unlink --all --force
  chauf unlink --alias admin.test --alias api.test  # remove specific aliases

Alias Removal:
  Use --alias to remove specific alias domains from a project without
  unlinking the entire project. This automatically regenerates nginx
  configuration and restarts services to apply changes.
`)
}

func confirmUnlinkAction(logger *lib.Logger, prompt string) bool {
	logger.Prompt(prompt, "Type 'y' to continue, anything else cancels")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// provideSSLCertificateRemovalGuidance provides structured logging for SSL certificate removal
func provideSSLCertificateRemovalGuidance(logger *lib.Logger, domain string, certExisted bool, wasMkcert bool, err error) {
	logger.PrintSection("SSL Certificate Removal")

	if !certExisted {
		logger.Info(fmt.Sprintf("No SSL certificates found for domain: %s", domain))
		logger.Info("Certificate removal not required")
		return
	}

	if err != nil {
		logger.Warn("SSL certificate removal failed", err.Error())
		logger.Info("Certificate files may still exist on filesystem")
		return
	}

	// Certificate was successfully removed
	logger.Success("SSL certificates removed successfully", fmt.Sprintf("domain: %s", domain))

	if wasMkcert {
		logger.Info("Certificate type: mkcert (trusted)")
		logger.Info("✓ Trusted certificate removed from system")
		logger.Info("✓ Certificate files deleted from workspace")

		// Attempt trust store cleanup
		if mkcertAvailable, _ := lib.CheckMkcertAvailable(); mkcertAvailable {
			logger.Info("Cleaning up trust store...")
			if trustErr := lib.CleanupMkcertTrustStore(logger); trustErr != nil {
				logger.Warn("Trust store cleanup failed", trustErr.Error())
				logger.Info("Note: You may need to manually remove the certificate from your browser trust store")
			} else {
				logger.Success("Trust store cleanup completed", "mkcert certificates removed from system")
			}
		} else {
			logger.Info("Note: mkcert not available - only certificate files removed")
		}
	} else {
		logger.Info("Certificate type: self-signed")
		logger.Info("✓ Self-signed certificate files deleted from workspace")
		logger.Info("✓ No trust store cleanup needed for self-signed certificates")
	}

	logger.Info("Certificate location (removed):")
	certPath := filepath.Join(os.Getenv("HOME"), ".chauffeur", "nginx", "certs", fmt.Sprintf("%s.crt", domain))
	keyPath := filepath.Join(os.Getenv("HOME"), ".chauffeur", "nginx", "certs", fmt.Sprintf("%s.key", domain))
	logger.Info(fmt.Sprintf("  Certificate: %s", certPath))
	logger.Info(fmt.Sprintf("  Private key: %s", keyPath))
}

func restartNginxIfNeeded(logger *lib.Logger) {
	manager, err := newServiceManager()
	if err != nil {
		logger.Warn("Unable to inspect nginx status", err.Error())
		return
	}

	var nginxSvc *services.Service
	for _, svc := range manager.ListGlobalServices() {
		if svc.Name == "chauf-nginx" {
			nginxSvc = &svc
			break
		}
	}

	if nginxSvc == nil {
		logger.Info("Chauffeur nginx is not installed; skipping reload")
		return
	}

	running, err := manager.IsRunning(*nginxSvc)
	if err != nil {
		logger.Warn("Failed to determine nginx status", err.Error())
		return
	}
	if !running {
		logger.Info("Chauffeur nginx is not running; no reload required")
		return
	}

	logger.Info("Reloading nginx to apply configuration changes")
	if err := manager.Restart(*nginxSvc); err != nil {
		logger.Warn("Failed to reload nginx after unlink", err.Error())
		return
	}
	logger.Success("Chauffeur nginx restarted", "")
}
