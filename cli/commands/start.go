package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunStart starts Chauffeur services.
 *
 * @param args CLI arguments passed after the start subcommand.
 * @return error when prerequisite checks or start operations fail.
 */
func RunStart(args []string) error {
	var (
		serviceNames []string
		filterProject string
		all           bool
		dryRun        bool
	)

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printStartUsage()
			return nil
		case "--project":
			if i+1 >= len(args) {
				return fmt.Errorf("--project requires a project slug")
			}
			filterProject = args[i+1]
			i += 2
		case "--all":
			all = true
			i++
		case "--dry-run":
			dryRun = true
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for start: %s", arg)
			}
			// If not a flag, treat as service name
			serviceNames = append(serviceNames, arg)
			i++
		}
	}

	// Default to no services if none specified (not --all)
	if len(serviceNames) == 0 && !all && filterProject == "" {
		all = true // Default to starting all services if no specific services mentioned
	}

	// Ensure workspace exists
	if err := workspace.Ensure(); err != nil {
		return fmt.Errorf("ensure workspace: %w", err)
	}

	// Create service manager
	manager, err := services.NewServiceManager()
	if err != nil {
		return fmt.Errorf("create service manager: %w", err)
	}

	// Determine which services to start
	var servicesToStart []services.Service

	if all {
		// Start all global services and all project services
		globalServices := manager.ListGlobalServices()
		servicesToStart = append(servicesToStart, globalServices...)

		// Add all project services (filtered by project if specified)
		projects, err := findAllLinkedProjects()
		if err != nil {
			return fmt.Errorf("find linked projects: %w", err)
		}
		for _, projectSlug := range projects {
			if filterProject != "" && projectSlug != filterProject {
				continue
			}
			projectServices, err := manager.ListProjectServices(projectSlug)
			if err != nil {
				fmt.Printf("Warning: Could not load services for project %s: %v\n", projectSlug, err)
				continue
			}
			servicesToStart = append(servicesToStart, projectServices...)
		}
	} else if len(serviceNames) > 0 {
		// Start specific services
		for _, serviceName := range serviceNames {
			switch serviceName {
			case "nginx", "chauf-nginx":
				globalServices := manager.ListGlobalServices()
				for _, svc := range globalServices {
					if strings.Contains(svc.Name, "nginx") {
						servicesToStart = append(servicesToStart, svc)
					}
				}
			case "caddy", "chauf-caddy":
				globalServices := manager.ListGlobalServices()
				for _, svc := range globalServices {
					if strings.Contains(svc.Name, "caddy") {
						servicesToStart = append(servicesToStart, svc)
					}
				}
			case "php-fpm", "php":
				// Start specific project's PHP-FPM or all if no project filter

				if filterProject != "" {
					// Start specific project's PHP-FPM
					projectServices, err := manager.ListProjectServices(filterProject)
					if err != nil {
						return fmt.Errorf("list project services: %w", err)
					}
					servicesToStart = append(servicesToStart, projectServices...)
				} else {
					// Start all PHP-FPM services
					projects, err := findAllLinkedProjects()
					if err != nil {
						return fmt.Errorf("find linked projects: %w", err)
					}
					for _, projectSlug := range projects {
						projectServices, err := manager.ListProjectServices(projectSlug)
						if err != nil {
							fmt.Printf("Warning: Could not load services for project %s: %v\n", projectSlug, err)
							continue
						}
						servicesToStart = append(servicesToStart, projectServices...)
					}
				}
			default:
				// Check if it's a specific project slug for php-fpm
				projectServices, err := manager.ListProjectServices(serviceName)
				if err != nil {
					return fmt.Errorf("invalid service name: %s (try nginx, caddy, php-fpm, or a project slug)", serviceName)
				}
				servicesToStart = append(servicesToStart, projectServices...)
			}
		}
	} else if filterProject != "" {
		// Filter by project only - start all services for that project
		globalServices := manager.ListGlobalServices()
		servicesToStart = append(servicesToStart, globalServices...)

		projectServices, err := manager.ListProjectServices(filterProject)
		if err != nil {
			return fmt.Errorf("list project services: %w", err)
		}
		servicesToStart = append(servicesToStart, projectServices...)
	}

	if len(servicesToStart) == 0 {
		fmt.Println("No services to start.")
		return nil
	}

	if dryRun {
		fmt.Printf("Would start %d services:\n", len(servicesToStart))
		for _, service := range servicesToStart {
			status, _ := manager.GetStatus(service)
			fmt.Printf("  - %s (%s)\n", service.Name, status)
		}
		return nil
	}

	// Start services
	for _, service := range servicesToStart {
		fmt.Printf("Starting %s...\n", service.Name)

		status, err := manager.GetStatus(service)
		if err != nil {
			fmt.Printf("Warning: Could not check status of %s: %v\n", service.Name, err)
		} else if status != "stopped" {
			fmt.Printf("  ✓ %s is already %s\n", service.Name, status)
			continue
		}

		if err := manager.Start(service); err != nil {
			fmt.Printf("  ✗ Failed to start %s: %v\n", service.Name, err)
			continue
		}

		// Verify it started successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			fmt.Printf("  ⚠ Started %s but could not verify status\n", service.Name)
		} else {
			fmt.Printf("  ✓ %s started successfully (%s)\n", service.Name, status)
		}
	}

	return nil
}

/**
 * findAllLinkedProjects returns a list of all linked project slugs.
 */
func findAllLinkedProjects() ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	entries, err := os.ReadDir(cfg.ProjectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read projects directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(cfg.ProjectsDir, entry.Name())
		configPath := filepath.Join(projectPath, "project.yaml")
		if _, err := os.Stat(configPath); err != nil {
			continue
		}

		projects = append(projects, entry.Name())
	}

	return projects, nil
}

/**
 * printStartUsage renders CLI help for the start command.
 */
func printStartUsage() {
	fmt.Println(`Usage: chauf start [service...] [--project <slug>] [--all] [--dry-run]

Starts Chauffeur services with chauf- prefix to avoid conflicts with system services.

Arguments:
  service           Start specific service(s): nginx, caddy, php-fpm, or project slug.

Flags:
  --project <slug>  Start services for specific project (global + project services).
  --all             Start all services (global + all projects).
  --dry-run         Show what would be started without taking action.
  -h, --help        Show this message.

Examples:
  chauf start                 # Start all Chauffeur services
  chauf start nginx           # Start chauf-nginx only
  chauf start nginx caddy     # Start chauf-nginx and chauf-caddy
  chauf start php-fpm         # Start all chauf-php-fpm-* services
  chauf start --project hja-cms  # Start nginx, caddy, and hja-cms's php-fpm
  chauf start hja-cms         # Start php-fpm for hja-cms project only

Service Names:
  - chauf-nginx              # Global Nginx service
  - chauf-caddy              # Global Caddy service  
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}
