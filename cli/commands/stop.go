package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * RunStop stops Chauffeur services.
 *
 * @param args CLI arguments passed after the stop subcommand.
 * @return error when prerequisite checks or stop operations fail.
 */
func RunStop(args []string) error {
	var (
		projectPath string
		all         bool
		dryRun      bool
	)

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printStopUsage()
			return nil
		case "--project":
			if i+1 >= len(args) {
				return fmt.Errorf("--project requires a path")
			}
			projectPath = args[i+1]
			i += 2
		case "--all":
			all = true
			i++
		case "--dry-run":
			dryRun = true
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for stop: %s", arg)
			}
			// If not a flag, treat as project path
			if projectPath == "" {
				projectPath = arg
				i++
			} else {
				return fmt.Errorf("multiple project paths specified")
			}
		}
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

	// Determine which services to stop
	var servicesToStop []services.Service

	if all {
		// Stop all global services and all project services
		globalServices := manager.ListGlobalServices()
		servicesToStop = append(servicesToStop, globalServices...)

		// Add all project services
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
			servicesToStop = append(servicesToStop, projectServices...)
		}
	} else if projectPath != "" {
		// Stop services for specific project
		absProjectPath, err := filepath.Abs(projectPath)
		if err != nil {
			return fmt.Errorf("resolve project path: %w", err)
		}

		// Find project by path
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		_, projectLayout, err := projects.FindByPath(cfg.ProjectsDir, absProjectPath)
		if err != nil {
			return fmt.Errorf("project not found: %w", err)
		}

		projectSlug := filepath.Base(projectLayout.Root)
		projectServices, err := manager.ListProjectServices(projectSlug)
		if err != nil {
			return fmt.Errorf("list project services: %w", err)
		}

		// Stop project services only (don't stop global services for specific project)
		servicesToStop = append(servicesToStop, projectServices...)
	} else {
		// Default: stop all services (global + all projects)
		globalServices := manager.ListGlobalServices()
		servicesToStop = append(servicesToStop, globalServices...)

		// Add all project services
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
			servicesToStop = append(servicesToStop, projectServices...)
		}
	}

	if len(servicesToStop) == 0 {
		fmt.Println("No services to stop.")
		return nil
	}

	if dryRun {
		fmt.Printf("Would stop %d services:\n", len(servicesToStop))
		for _, service := range servicesToStop {
			status, _ := manager.GetStatus(service)
			fmt.Printf("  - %s (%s)\n", service.Name, status)
		}
		return nil
	}

	// Stop services
	fmt.Printf("Stopping %d services...\n", len(servicesToStop))
	nginxStopped := false
	for _, service := range servicesToStop {
		// Check if service is already stopped
		status, err := manager.GetStatus(service)
		if err != nil {
			fmt.Printf("Warning: Could not check status of %s: %v\n", service.Name, err)
		} else if status == "stopped" {
			fmt.Printf("  ✓ %s is already stopped\n", service.Name)
			if service.Name == "chauf-nginx" {
				nginxStopped = true
			}
			continue
		}

		// Stop with spinner for active processes
		spin := lib.NewSpinner("stop", fmt.Sprintf("Stopping %s", service.Name))

		if err := manager.Stop(service); err != nil {
			spin.Fail("failed to stop")
			fmt.Printf("  ✗ Failed to stop %s: %v\n", service.Name, err)
			continue
		}

		// Verify it stopped successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			spin.Fail("stopped but verification failed")
			fmt.Printf("  ⚠ Stopped %s but could not verify status\n", service.Name)
		} else {
			spin.Success(status + " stopped successfully")
			if service.Name == "chauf-nginx" {
				nginxStopped = true
			}
		}
	}

	if nginxStopped {
		if root, err := workspace.Dir(); err == nil {
			if err := system.CleanupPortForwarding(root); err != nil {
				fmt.Printf("Warning: Failed to clean up port forwarding: %v\n", err)
			}
		}
	}

	return nil
}

/**
 * printStopUsage renders CLI help for the stop command.
 */
func printStopUsage() {
	fmt.Println(`Usage: chauf stop [--project <path>] [--all] [--dry-run]

Stops Chauffeur services with chauf- prefix to avoid conflicts with system services.

Flags:
  --project <path>  Stop services for specific project (default: current directory).
  --all             Stop all services (global + all projects).
  --dry-run         Show what would be stopped without taking action.
  -h, --help        Show this message.

Examples:
  chauf stop                 # Stop global services or services for current project
  chauf stop --project /path/to/project  # Stop services for specific project
  chauf stop --all           # Stop all Chauffeur services

Service Names:
  - chauf-nginx              # Global Nginx service
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}
