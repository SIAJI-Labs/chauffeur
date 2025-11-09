package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunLinks handles `chauf links` command invocations.
func RunLinks(args []string) error {
	logger := lib.NewCommandLogger("links")

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printLinksUsage()
			return nil
		default:
			return logger.Error("unknown flag for links", arg)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	projectsDir := cfg.ProjectsDir
	if projectsDir == "" {
		// Fallback to default projects directory
		ws, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determine home directory: %w", err)
		}
		projectsDir = filepath.Join(ws, ".chauffeur", "projects")
	}

	allProjects, err := listAllProjects(projectsDir)
	if err != nil {
		return err
	}

	if len(allProjects) == 0 {
		logger.Info("No projects linked yet.")
		logger.Info("Use 'chauf link' in a project directory to register it.")
		return nil
	}

	logger.PrintSection(fmt.Sprintf("Linked Projects (%d)", len(allProjects)))

	maxPath := 20
	maxDomain := 10
	maxSlug := 10

	// First pass to determine column widths and prepare domain defaults
	projectsWithDomains := make([]linkedProject, len(allProjects))
	for i, project := range allProjects {
		if len(project.Path) > maxPath {
			maxPath = len(project.Path)
		}
		if len(project.Slug) > maxSlug {
			maxSlug = len(project.Slug)
		}

		// Copy the project and set default domain if needed
		projectsWithDomains[i] = project
		if projectsWithDomains[i].Site == nil {
			projectsWithDomains[i].Site = &projects.Site{
				Domain: project.Slug + ".test",
				SSL:    false,
			}
		}

		domain := projectsWithDomains[i].Site.Domain
		if len(domain) > maxDomain {
			maxDomain = len(domain)
		}
	}

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-3s  %-4s  %s",
		maxSlug, "SLUG",
		maxPath, "PATH",
		maxDomain, "DOMAIN",
		"SSL", "PHP", "CREATED")
	divider := fmt.Sprintf("%s  %s  %s  ---  ----  %s",
		strings.Repeat("-", maxSlug),
		strings.Repeat("-", maxPath),
		strings.Repeat("-", maxDomain),
		strings.Repeat("-", 19))

	logger.Info(header)
	logger.Info(divider)

	// Project rows
	for _, project := range projectsWithDomains {
		ssl := " "
		if project.Site.SSL {
			ssl = "*"
		}

		created := project.CreatedAt.Format("2006-01-02 15:04")

		row := fmt.Sprintf("%-*s  %-*s  %-*s  %-3s  %-4s  %s",
			maxSlug, project.Slug,
			maxPath, project.Path,
			maxDomain, project.Site.Domain,
			ssl,
			project.PHP,
			created)
		logger.Info(row)
	}

	return nil
}

type linkedProject struct {
	Slug      string
	Path      string
	PHP       string
	Site      *projects.Site
	CreatedAt time.Time
}

// listAllProjects scans the projects directory and returns all registered projects
func listAllProjects(projectsDir string) ([]linkedProject, error) {
	var allProjects []linkedProject

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return allProjects, nil // No projects directory yet
		}
		return nil, fmt.Errorf("read projects directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		layout, err := projects.EnsureLayout(projectsDir, slug)
		if err != nil {
			continue // Skip invalid directories
		}

		cfg, err := projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			continue // Skip projects with invalid config
		}

		allProjects = append(allProjects, linkedProject{
			Slug:      slug,
			Path:      cfg.Path,
			PHP:       cfg.PHP,
			Site:      cfg.Site,
			CreatedAt: cfg.CreatedAt,
		})
	}

	return allProjects, nil
}

func printLinksUsage() {
	fmt.Print(`Chauffeur Project Listing

Usage:
  chauf links          List all registered projects and their configurations.

Flags:
  --help, -h          Show this help message.
`)
}
