package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	var categoryFlag string
	flagSet := lib.NewFlagSet("links", logger)
	flagSet.StringVar(&slugFlag, "slug", "", "Display detailed information for a specific project by slug.")
	flagSet.StringVar(&siteFlag, "site", "", "Display detailed information for a specific project by site domain.")
	flagSet.StringVar(&categoryFlag, "category", "", "Filter projects by category.")

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
	if (slugFlag != "" || siteFlag != "") && categoryFlag != "" {
		return fmt.Errorf("detail view flags (--slug, --site) cannot be combined with --category")
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

	// Filter by category if --category flag is provided
	if categoryFlag != "" {
		var filteredProjects []linkedProject
		for _, project := range allProjects {
			if strings.EqualFold(project.Category, categoryFlag) {
				filteredProjects = append(filteredProjects, project)
			}
		}
		allProjects = filteredProjects

		if len(allProjects) == 0 {
			logger.Info(fmt.Sprintf("No projects found in category '%s'.", categoryFlag))
			return nil
		}
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

	printProjectsByCategory(logger, allProjects)

	
	return nil
}

type linkedProject struct {
	Slug        string
	Path        string
	PHP         string
	Category    string
	Site        *projects.Site
	AliasCount  int // Number of alias domains
	SSL         bool
	CreatedAt   time.Time
	FPMType     string // Shared or Dedicated
	FPMSocket   string // Path to the FPM socket
}

// printProjectsByCategory displays projects grouped by category with formatted tables.
func printProjectsByCategory(logger *lib.Logger, projects []linkedProject) {
	if len(projects) == 0 {
		return
	}

	// Group projects by category
	categories := make(map[string][]linkedProject)
	for _, project := range projects {
		categories[project.Category] = append(categories[project.Category], project)
	}

	// Sort categories alphabetically (put "Uncategorized" last)
	sortedCategories := make([]string, 0, len(categories))
	for category := range categories {
		sortedCategories = append(sortedCategories, category)
	}
	sort.Slice(sortedCategories, func(i, j int) bool {
		if sortedCategories[j] == "Uncategorized" {
			return true
		}
		if sortedCategories[i] == "Uncategorized" {
			return false
		}
		return strings.ToLower(sortedCategories[i]) < strings.ToLower(sortedCategories[j])
	})

	totalProjects := len(projects)
	logger.PrintSection(fmt.Sprintf("Linked Projects (%d projects, %d categories)", totalProjects, len(categories)))

	for _, category := range sortedCategories {
		categoryProjects := categories[category]

		// Print category header with separator
		if len(sortedCategories) > 1 {
			logger.PrintSeparator()
			categoryDisplay := category
			if category == "Uncategorized" {
				categoryDisplay = category + " 📋"
			}
			logger.Info(fmt.Sprintf("📁 %s (%d project%s)", categoryDisplay, len(categoryProjects), func() string {
				if len(categoryProjects) != 1 {
					return "s"
				}
				return ""
			}()))
		}

		// Calculate column widths for this category
		maxPath := 15
		maxSlug := 8
		maxAliasCount := 6

		for _, project := range categoryProjects {
			if len(project.Path) > maxPath {
				maxPath = len(project.Path)
			}
			if len(project.Slug) > maxSlug {
				maxSlug = len(project.Slug)
			}
			aliasCountStr := fmt.Sprintf("%d", project.AliasCount)
			if len(aliasCountStr) > maxAliasCount {
				maxAliasCount = len(aliasCountStr)
			}
		}

		// Print table header
		header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-3s  %-4s  %s",
			maxSlug, "SLUG",
			maxPath, "PATH",
			maxSlug+5, "DOMAIN",
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

		// Print project rows
		for _, project := range categoryProjects {
			ssl := " "
			if project.SSL {
				ssl = "*"
			}

			created := project.CreatedAt.Format("2006-01-02 15:04")

			domainDisplay := ""
			if project.Site != nil {
				domainDisplay = project.Site.Domain
			}

			// Add indicator for example project
			slugDisplay := project.Slug
			if project.Slug == "example-project" {
				slugDisplay = project.Slug + " 📁"
			}

			row := fmt.Sprintf("%-*s  %-*s  %-*s  %-*d  %-3s  %-4s  %s",
				maxSlug, slugDisplay,
				maxPath, project.Path,
				maxSlug+5, domainDisplay,
				maxAliasCount, project.AliasCount,
				ssl,
				project.PHP,
				created)
			logger.Info(row)
		}
	}

	// Print summary if there are multiple categories
	if len(categories) > 1 {
		logger.PrintSeparator()
		var summary []string
		for _, category := range sortedCategories {
			count := len(categories[category])
			categoryDisplay := category
			if category == "Uncategorized" {
				categoryDisplay = category + " 📋"
			}
			summary = append(summary, fmt.Sprintf("%s: %d", categoryDisplay, count))
		}
		logger.Info(fmt.Sprintf("Summary: %s", strings.Join(summary, ", ")))
	}
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
			Category:   cfg.GetCategory(),
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
  chauf links                        List all registered projects grouped by category.
  chauf links --slug <project-slug>  Display detailed information for a specific project.
  chauf links --site <domain>        Display detailed information for a specific project by one of its domains.
  chauf links --category <name>      List projects from a specific category only.

Output:
  Projects are automatically grouped by category and displayed with visual separators.
  The DOMAIN column shows the primary domain for each project.
  The ALIASES column shows the number of alias domains configured for the project.
  An asterisk (*) in the SSL column indicates that SSL/TLS is enabled for at least one domain.
  Uncategorized projects are shown in the "Uncategorized 📋" category.

Categories:
  Projects are automatically categorized when linked using 'chauf link --category <name>'.
  Use 'chauf link --update --category <name> --slug <slug>' to change an existing project's category.
  Each category shows a project count and summary when multiple categories exist.

Flags:
  --slug string       Display detailed information for a specific project by its slug.
  --site string       Display detailed information for a specific project by one of its domains (primary or alias).
  --category string   Filter projects to show only those in the specified category.
  --help, -h          Show this help message.
`)
}
