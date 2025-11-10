package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunUnlink handles `chauf unlink` command invocations.
func RunUnlink(args []string) error {
	var (
		slug    string
		domain  string
		project string
		all     bool
		force   bool
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
			i += 2
		case "--site":
			if i+1 >= len(args) {
				return fmt.Errorf("--site requires a domain value")
			}
			domain = args[i+1]
			i += 2
		case "--project":
			if i+1 >= len(args) {
				return fmt.Errorf("--project requires a path value")
			}
			project = args[i+1]
			i += 2
		case "--all":
			all = true
			i++
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

	// Initialize variables
	var projectPath string
	var projectSlug string

	// Handle default case: no flags provided, use current directory
	if !all && slug == "" && domain == "" && project == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine current directory: %w", err)
		}
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return fmt.Errorf("resolve project path: %w", err)
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
		if projCfg.Site != nil {
			logger.Info(fmt.Sprintf("  Domain: %s (ssl=%t)", projCfg.Site.Domain, projCfg.Site.SSL))
		}
		logger.Warn("This will remove the project registration and all associated configuration", "Use --force to skip confirmation")
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

	// Remove the project directory
	projectDir := layout.Root
	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("remove project directory: %w", err)
	}

	logger.Success("Successfully unlinked project", projectSlug)
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
		logger.Warn("This will remove ALL registered projects and their configurations", "Use --force to skip confirmation")
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

func printUnlinkUsage() {
	fmt.Print(`Chauffeur Project Unlinking

Usage:
  chauf unlink                       Unlink current directory (default)
  chauf unlink --slug <slug>          Unlink project by slug
  chauf unlink --site <domain>       Unlink project by domain
  chauf unlink --project <path>       Unlink project by directory path
  chauf unlink --all                 Unlink all registered projects

Flags:
  --force                           Proceed without confirmation.
  --all                             Remove all registered projects.
  --help, -h                        Show this help message.

Examples:
  chauf unlink                       # unlink current project if linked
  chauf unlink --slug myproject
  chauf unlink --site myproject.test
  chauf unlink --project /path/to/project
  chauf unlink --all --force
`)
}

func confirmUnlinkAction(logger *lib.Logger, prompt string) bool {
	logger.Warn(prompt, "Type 'y' to continue, anything else cancels")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func restartNginxIfNeeded(logger *lib.Logger) {
	manager, err := services.NewServiceManager()
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
