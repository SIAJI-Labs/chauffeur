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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Handle --all flag
	if all {
		return unlinkAllProjects(cfg, force)
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
		fmt.Printf("Project to unlink:\n")
		fmt.Printf("  Slug: %s\n", projectSlug)
		fmt.Printf("  Path: %s\n", projCfg.Path)
		fmt.Printf("  PHP: %s\n", projCfg.PHP)
		if projCfg.Site != nil {
			fmt.Printf("  Domain: %s (ssl=%t)\n", projCfg.Site.Domain, projCfg.Site.SSL)
		}
		fmt.Printf("\nWARNING: This will remove the project registration and all associated configuration.\n")
		fmt.Printf("Use --force to proceed without confirmation.\n")
		fmt.Print("Continue? [y/N] ")
		
		// SENSITIVE: Destructive operation - user confirmation for unlinking projects
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove nginx configuration
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("initialize template engine: %w", err)
	}
	
	if err := templateEngine.RemoveNginxConfig(projectSlug); err != nil {
		fmt.Printf("Warning: failed to remove nginx configuration: %v\n", err)
		// Continue with unlink even if nginx removal fails
	}

	if err := templateEngine.RemoveCaddyConfig(projectSlug); err != nil {
		fmt.Printf("Warning: failed to remove Caddy configuration: %v\n", err)
		// Continue with unlink even if Caddy removal fails
	}

	// Remove the project directory
	projectDir := layout.Root
	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("remove project directory: %w", err)
	}

	fmt.Printf("Successfully unlinked project '%s'\n", projectSlug)
	return nil
}

func unlinkAllProjects(cfg config.Config, force bool) error {
	entries, err := os.ReadDir(cfg.ProjectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no projects registered")
		}
		return fmt.Errorf("read projects directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No projects registered.")
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
		fmt.Println("No valid projects found.")
		return nil
	}

	fmt.Printf("Found %d project(s) to unlink:\n\n", len(allProjectsInfo))
	for _, proj := range allProjectsInfo {
		fmt.Printf("  %s: %s", proj.Slug, proj.Path)
		if proj.Site != nil {
			fmt.Printf(" (domain: %s", proj.Site.Domain)
			if proj.Site.SSL {
				fmt.Print(", ssl=true")
			}
			fmt.Print(")")
		}
		fmt.Printf(" (php: %s)\n", proj.PHP)
	}

	if !force {
		fmt.Printf("\nWARNING: This will remove ALL registered projects and their configurations.\n")
		fmt.Printf("Use --force to proceed without confirmation.\n")
		fmt.Print("Continue? [y/N] ")

		// SENSITIVE: Destructive operation - user confirmation for removing all projects
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove nginx configurations first
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		fmt.Printf("Warning: failed to initialize template engine: %v\n", err)
		// Continue with unlink even if template engine fails
	} else {
		for _, proj := range allProjectsInfo {
			if err := templateEngine.RemoveNginxConfig(proj.Slug); err != nil {
				fmt.Printf("Warning: failed to remove nginx configuration for %s: %v\n", proj.Slug, err)
				// Continue even if nginx removal fails
			}

			if err := templateEngine.RemoveCaddyConfig(proj.Slug); err != nil {
				fmt.Printf("Warning: failed to remove Caddy configuration for %s: %v\n", proj.Slug, err)
				// Continue even if Caddy removal fails
			}
		}
	}

	for _, proj := range allProjectsInfo {
		layout, err := projects.EnsureLayout(cfg.ProjectsDir, proj.Slug)
		if err != nil {
			fmt.Printf("Warning: failed to access project %s: %v\n", proj.Slug, err)
			continue
		}

		if err := os.RemoveAll(layout.Root); err != nil {
			fmt.Printf("Warning: failed to remove project %s: %v\n", proj.Slug, err)
			continue
		}
	}

	fmt.Printf("Successfully unlinked %d project(s)\n", len(allProjectsInfo))
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
