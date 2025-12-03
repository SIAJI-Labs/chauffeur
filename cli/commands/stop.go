package commands

import (
	"fmt"
	"os"
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
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

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

	logger := lib.NewCommandLogger("stop")

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
				logger.Warn("Could not load project services", fmt.Sprintf("%s: %v", projectSlug, err))
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
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		
		entries, err := os.ReadDir(cfg.ProjectsDir)
		if err != nil {
			return fmt.Errorf("read projects directory: %w", err)
		}
		
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			
			projectSlug := entry.Name()
			projectPath := filepath.Join(cfg.ProjectsDir, projectSlug)
			configPath := filepath.Join(projectPath, "project.yaml")
			if _, err := os.Stat(configPath); err != nil {
				continue
			}
			projectServices, err := manager.ListProjectServices(projectSlug)
			if err != nil {
				logger.Warn("Could not load project services", fmt.Sprintf("%s: %v", projectSlug, err))
				continue
			}
			servicesToStop = append(servicesToStop, projectServices...)
		}
	}

	// Remove duplicate services (shared PHP-FPM for multiple projects)
	seen := make(map[string]bool)
	var uniqueServices []services.Service
	for _, svc := range servicesToStop {
		if !seen[svc.Name] {
			seen[svc.Name] = true
			uniqueServices = append(uniqueServices, svc)
		}
	}
	servicesToStop = uniqueServices

	if len(servicesToStop) == 0 {
		logger.Info("No services to stop.")
		return nil
	}

	if dryRun {
		logger.Info(fmt.Sprintf("Would stop %d services:", len(servicesToStop)))
		for _, service := range servicesToStop {
			status, _ := manager.GetStatus(service)
			logger.Info(fmt.Sprintf("  - %s (%s)", service.Name, status))
		}
		return nil
	}

	// Stop services
	logger.Info(fmt.Sprintf("Stopping %d services...", len(servicesToStop)))
	nginxStopped := false
	for _, service := range servicesToStop {
		// Check if service is already stopped
		status, err := manager.GetStatus(service)
		if err != nil {
			logger.Warn(fmt.Sprintf("Could not check status of %s", service.Name), err.Error())
		} else if status == "stopped" {
			logger.Success(fmt.Sprintf("%s already stopped", service.Name), "")
			if service.Name == "chauf-nginx" {
				nginxStopped = true
			}
			continue
		}

		// Stop with spinner for active processes
		spin := lib.NewSpinner("stop", fmt.Sprintf("Stopping %s", service.Name))

		if err := manager.Stop(service); err != nil {
			spin.Fail("failed to stop")
			logger.Warn(fmt.Sprintf("Failed to stop %s", service.Name), err.Error())
			continue
		}

		// Verify it stopped successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			spin.Fail("stopped but verification failed")
			logger.Warn(fmt.Sprintf("Stopped %s but could not verify status", service.Name), err.Error())
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
				logger.Warn("Failed to clean up port forwarding", err.Error())
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
