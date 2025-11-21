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

	var slugFlag string
	var siteFlag string
	flagSet := lib.NewFlagSet("links", logger)
	flagSet.StringVar(&slugFlag, "slug", "", "Display detailed information for a specific project by slug.")
	flagSet.StringVar(&siteFlag, "site", "", "Display detailed information for a specific project by site domain.")

	if err := flagSet.Parse(args); err != nil {
		if err == lib.ErrHelpRequested {
			printLinksUsage()
			return nil
		}
		return err
	}

	// Validate mutually exclusive flags
	if slugFlag != "" && siteFlag != "" {
		return fmt.Errorf("flags --slug and --site are mutually exclusive, please provide only one")
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

	// Handle detail view if --slug or --site is provided
	if slugFlag != "" || siteFlag != "" {
		var foundProject *linkedProject
		for i, p := range allProjects {
			if slugFlag != "" && p.Slug == slugFlag {
				foundProject = &allProjects[i]
				break
			}
			if siteFlag != "" && p.Site != nil && p.Site.Domain == siteFlag {
				foundProject = &allProjects[i]
				break
			}
			// Check alias domains for siteFlag
			if siteFlag != "" {
				layout, err := projects.EnsureLayout(projectsDir, p.Slug)
				if err != nil {
					continue
				}
				cfg, err := projects.LoadConfig(layout.ConfigPath)
				if err != nil {
					continue
				}
				if cfg.Domains != nil {
					for _, alias := range cfg.Domains.Aliases {
						if alias.Domain == siteFlag {
							foundProject = &allProjects[i]
							break
						}
					}
				}
			}
			if foundProject != nil {
				break
			}
		}

		if foundProject == nil {
			if slugFlag != "" {
				return fmt.Errorf("project with slug '%s' not found", slugFlag)
			}
			return fmt.Errorf("project with site '%s' not found", siteFlag)
		}

		// Load full config for detailed view (linkedProject doesn't have all details)
		layout, err := projects.EnsureLayout(projectsDir, foundProject.Slug)
		if err != nil {
			return err
		}
		cfg, err := projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			return err
		}

		printProjectDetails(logger, foundProject.Slug, cfg, projectsDir)
		return nil
	}

	logger.PrintSection(fmt.Sprintf("Linked Projects (%d)", len(allProjects)))

	maxPath := 20
	maxSlug := 10
	maxAliasCount := 8 // Default width for "ALIASES" (e.g., "Count (X)")

	// First pass to determine column widths
	for _, project := range allProjects {
		if len(project.Path) > maxPath {
			maxPath = len(project.Path)
		}
		if len(project.Slug) > maxSlug {
			maxSlug = len(project.Slug)
		}
		// Adjust maxAliasCount for the string representation of the count
		aliasCountStr := fmt.Sprintf("%d", project.AliasCount)
		if len(aliasCountStr) > maxAliasCount {
			maxAliasCount = len(aliasCountStr)
		}
	}

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-3s  %-4s  %s",
		maxSlug, "SLUG",
		maxPath, "PATH",
		maxSlug+5, "DOMAIN", // Domain column will be wider
		maxAliasCount, "ALIASES",
		"SSL", "PHP", "CREATED")
	divider := fmt.Sprintf("%s  %s  %s  %s  ---  ----  %s",
		strings.Repeat("-", maxSlug),
		strings.Repeat("-", maxPath),
		strings.Repeat("-", maxSlug+5),
		strings.Repeat("-", maxAliasCount),
		strings.Repeat("-", 19))

	logger.Info(header)
	logger.Info(divider)

	// Project rows
	for _, project := range allProjects {
		ssl := " "
		if project.SSL {
			ssl = "*"
		}

		created := project.CreatedAt.Format("2006-01-02 15:04")

		domainDisplay := ""
		if project.Site != nil {
			domainDisplay = project.Site.Domain
		}

		row := fmt.Sprintf("%-*s  %-*s  %-*s  %-*d  %-3s  %-4s  %s",
			maxSlug, project.Slug,
			maxPath, project.Path,
			maxSlug+5, domainDisplay,
			maxAliasCount, project.AliasCount,
			ssl,
			project.PHP,
			created)
		logger.Info(row)
	}

	return nil
}

type linkedProject struct {
	Slug        string
	Path        string
	PHP         string
	Site        *projects.Site
	AliasCount  int // Number of alias domains
	SSL         bool
	CreatedAt   time.Time
	FPMType     string // Shared or Dedicated
	FPMSocket   string // Path to the FPM socket
}

// printProjectDetails displays detailed information for a single project.
func printProjectDetails(logger *lib.Logger, slug string, cfg projects.Config, projectsDir string) {
	logger.PrintSection(fmt.Sprintf("Project: %s", slug))

	logger.Info(fmt.Sprintf("  Slug:        %s", slug))
	logger.Info(fmt.Sprintf("  Path:        %s", cfg.Path))
	logger.Info(fmt.Sprintf("  PHP Version: %s", cfg.PHP))
	logger.Info(fmt.Sprintf("  Created At:  %s", cfg.CreatedAt.Format("2006-01-02 15:04:05 MST")))

	// Primary Domain
	if cfg.Site != nil && cfg.Site.Domain != "" {
		sslStatus := "No SSL"
		if cfg.Site.SSL {
			sslStatus = "SSL Enabled"
		}
		logger.Info(fmt.Sprintf("  Primary Domain: %s (%s)", cfg.Site.Domain, sslStatus))
	}

	// Alias Domains
	if cfg.Domains != nil && len(cfg.Domains.Aliases) > 0 {
		logger.Info("  Alias Domains:")
		for _, alias := range cfg.Domains.Aliases {
			sslStatus := "No SSL"
			if alias.SSL {
				sslStatus = "SSL Enabled"
			}
			logger.Info(fmt.Sprintf("    - %s (%s)", alias.Domain, sslStatus))
		}
	} else {
		logger.Info("  Alias Domains: None")
	}

	// FPM Details
	fpmType := "Shared"
	fpmSocket := ""
	if cfg.Runtime.FPM != nil {
		if cfg.Runtime.FPM.Dedicated {
			fpmType = "Dedicated"
		}
		fpmSocket = cfg.Runtime.FPM.Socket
	}

	if fpmSocket == "" {
		// Attempt to get socket path from layout if not explicitly set in config FPM
		layout, err := projects.EnsureLayout(projectsDir, slug)
		if err == nil {
			fpmSocket = layout.SocketPath
		}
	}

	logger.Info(fmt.Sprintf("  PHP-FPM:     %s", fpmType))
	if fpmSocket != "" {
		logger.Info(fmt.Sprintf("  FPM Socket:  %s", fpmSocket))
	}
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

		fpmType := "Shared"
		fpmSocket := layout.SocketPath // Default to layout socket path
		if cfg.Runtime.FPM != nil {
			if cfg.Runtime.FPM.Dedicated {
				fpmType = "Dedicated"
			}
			if cfg.Runtime.FPM.Socket != "" {
				fpmSocket = cfg.Runtime.FPM.Socket
			}
		}

		aliasCount := 0
		if cfg.Domains != nil {
			aliasCount = len(cfg.Domains.Aliases)
		}

		allProjects = append(allProjects, linkedProject{
			Slug:       slug,
			Path:       cfg.Path,
			PHP:        cfg.PHP,
			Site:       cfg.Site,
			AliasCount: aliasCount,
			SSL:        cfg.HasSSLEnabled(),
			CreatedAt:  cfg.CreatedAt,
			FPMType:    fpmType,
			FPMSocket:  fpmSocket,
		})
	}

	return allProjects, nil
}

func printLinksUsage() {
	fmt.Fprintf(lib.CurrentStdout, `Chauffeur Project Listing

Usage:
  chauf links                        List all registered projects and their configurations.
  chauf links --slug <project-slug>  Display detailed information for a specific project.
  chauf links --site <domain>        Display detailed information for a specific project by one of its domains.

Output:
  The DOMAIN column shows the primary domain for each project.
  The ALIASES column shows the number of alias domains configured for the project.
  An asterisk (*) in the SSL column indicates that SSL/TLS is enabled for at least one domain.

Flags:
  --slug string       Display detailed information for a specific project by its slug.
  --site string       Display detailed information for a specific project by one of its domains (primary or alias).
  --help, -h          Show this help message.
`)
}
