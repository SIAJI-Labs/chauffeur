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
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

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
	maxAlias := 8
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

		// Copy the project and calculate domains display
		projectWithConfig, err := projects.LoadConfig(filepath.Join(projectsDir, project.Slug, "project.yaml"))
		if err != nil {
			// Fallback to basic project structure
			projectsWithDomains[i] = project
			if projectsWithDomains[i].Site == nil {
				projectsWithDomains[i].Site = &projects.Site{
					Domain: project.Slug + ".test",
					SSL:    false,
				}
			}
		} else {
			// Use full config to get all domains
			allDomains := projectWithConfig.GetAllDomains()
			primaryDomain := projectWithConfig.GetPrimaryDomain()

			var aliasDomains []string
			for _, d := range allDomains {
				if d.Domain != primaryDomain {
					aliasStr := d.Domain
					if d.SSL {
						aliasStr += " (*)"
					}
					aliasDomains = append(aliasDomains, aliasStr)
				}
			}
			aliasesStr := strings.Join(aliasDomains, ", ")
			if aliasesStr == "" {
				aliasesStr = "-"
			}

			projectsWithDomains[i] = linkedProject{
				Slug:      project.Slug,
				Path:      project.Path,
				PHP:       project.PHP,
				Site:      projectWithConfig.Site,
				Domain:    primaryDomain,
				Aliases:   aliasesStr,
				SSL:       projectWithConfig.HasSSLEnabled(),
				CreatedAt: projectWithConfig.CreatedAt,
			}
		}

		// Check max widths for domain and alias columns
		if len(projectsWithDomains[i].Domain) > maxDomain {
			maxDomain = len(projectsWithDomains[i].Domain)
		}
		if len(projectsWithDomains[i].Aliases) > maxAlias {
			maxAlias = len(projectsWithDomains[i].Aliases)
		}
	}

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-3s  %-4s  %s",
		maxSlug, "SLUG",
		maxPath, "PATH",
		maxDomain, "DOMAIN",
		maxAlias, "ALIAS",
		"SSL", "PHP", "CREATED")
	divider := fmt.Sprintf("%s  %s  %s  %s  ---  ----  %s",
		strings.Repeat("-", maxSlug),
		strings.Repeat("-", maxPath),
		strings.Repeat("-", maxDomain),
		strings.Repeat("-", maxAlias),
		strings.Repeat("-", 19))

	logger.Info(header)
	logger.Info(divider)

	// Project rows
	for _, project := range projectsWithDomains {
		ssl := " "
		if project.SSL {
			ssl = "*"
		}

		created := project.CreatedAt.Format("2006-01-02 15:04")

		// Use domain and aliases from separate fields
		domainDisplay := project.Domain
		if domainDisplay == "" && project.Site != nil {
			domainDisplay = project.Site.Domain
		}

		row := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-3s  %-4s  %s",
			maxSlug, project.Slug,
			maxPath, project.Path,
			maxDomain, domainDisplay,
			maxAlias, project.Aliases,
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
	Domain    string  // Primary domain only
	Aliases   string  // Alias domains as comma-separated string
	SSL       bool    // SSL status (from project config)
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

Output:
  The DOMAIN column shows the primary domain for each project.
  The ALIAS column shows additional domains (comma-separated) that point
  to the same project. An asterisk (*) in the SSL column indicates that
  SSL/TLS is enabled for at least one domain.

Flags:
  --help, -h          Show this help message.
`)
}
