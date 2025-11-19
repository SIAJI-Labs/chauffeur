package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * RunRestart restarts Chauffeur services.
 * This is equivalent to running 'chauf stop' followed by 'chauf start' for the same services.
 *
 * @param args CLI arguments passed after the restart subcommand.
 * @return error when prerequisite checks or restart operations fail.
 */
func RunRestart(args []string) error {
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

	var (
		serviceNames  []string
		filterProject string
		all           bool
		dryRun        bool
	)

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printRestartUsage()
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
				return fmt.Errorf("unknown flag for restart: %s", arg)
			}
			// If not a flag, treat as service name
			serviceNames = append(serviceNames, arg)
			i++
		}
	}

	// Default to all services if none specified (not --all)
	if len(serviceNames) == 0 && !all && filterProject == "" {
		all = true // Default to restarting all services if no specific services mentioned
	}

	// Ensure workspace exists
	if err := workspace.Ensure(); err != nil {
		return fmt.Errorf("ensure workspace: %w", err)
	}

	// Create command logger
	logger := lib.NewCommandLogger("restart")

	// Check dnsmasq configuration (same as start command)
	if err := checkDnsmasqConfiguration(logger); err != nil {
		return err
	}

	// Create service manager
	manager, err := services.NewServiceManager()
	if err != nil {
		return fmt.Errorf("create service manager: %w", err)
	}

	// Determine which services to restart
	var servicesToRestart []services.Service

	if all {
		// Restart all global services and all project services
		globalServices := manager.ListGlobalServices()
		servicesToRestart = append(servicesToRestart, globalServices...)

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
				logger.Warn(fmt.Sprintf("Could not load services for project %s", projectSlug), fmt.Sprintf("error: %v", err))
				continue
			}
			servicesToRestart = append(servicesToRestart, projectServices...)
		}
	} else if len(serviceNames) > 0 {
		// Restart specific services
		for _, serviceName := range serviceNames {
			switch serviceName {
			case "nginx", "chauf-nginx":
				globalServices := manager.ListGlobalServices()
				for _, svc := range globalServices {
					if strings.Contains(svc.Name, "nginx") {
						servicesToRestart = append(servicesToRestart, svc)
					}
				}
			case "php-fpm", "php":
				// Restart specific project's PHP-FPM or all if no project filter

				if filterProject != "" {
					// Restart specific project's PHP-FPM
					projectServices, err := manager.ListProjectServices(filterProject)
					if err != nil {
						return fmt.Errorf("list project services: %w", err)
					}
					servicesToRestart = append(servicesToRestart, projectServices...)
				} else {
					// Restart all PHP-FPM services
					projects, err := findAllLinkedProjects()
					if err != nil {
						return fmt.Errorf("find linked projects: %w", err)
					}
					for _, projectSlug := range projects {
						projectServices, err := manager.ListProjectServices(projectSlug)
						if err != nil {
							logger.Warn(fmt.Sprintf("Could not load services for project %s", projectSlug), fmt.Sprintf("error: %v", err))
							continue
						}
						servicesToRestart = append(servicesToRestart, projectServices...)
					}
				}
			default:
				// Check if it's a specific project slug for php-fpm
				projectServices, err := manager.ListProjectServices(serviceName)
				if err != nil {
					return fmt.Errorf("invalid service name: %s (try nginx, php-fpm, or a project slug)", serviceName)
				}
				servicesToRestart = append(servicesToRestart, projectServices...)
			}
		}
	} else if filterProject != "" {
		// Restart all services for specific project
		globalServices := manager.ListGlobalServices()
		servicesToRestart = append(servicesToRestart, globalServices...)

		projectServices, err := manager.ListProjectServices(filterProject)
		if err != nil {
			return fmt.Errorf("list project services: %w", err)
		}
		servicesToRestart = append(servicesToRestart, projectServices...)
	}

	if len(servicesToRestart) == 0 {
		logger.Info("No services to restart.")
		return nil
	}

	if dryRun {
		logger.Info(fmt.Sprintf("Would restart %d services:", len(servicesToRestart)))
		for _, service := range servicesToRestart {
			status, _ := manager.GetStatus(service)
			logger.Info(fmt.Sprintf("  - %s (%s)", service.Name, status))
		}
		return nil
	}

	// Restart services: stop then start
	logger.Info(fmt.Sprintf("Restarting %d services...", len(servicesToRestart)))

	// First, stop all services
	stopLogger := logger.NewChildLogger("stop")
	for _, service := range servicesToRestart {
		// Check if service is already stopped
		status, err := manager.GetStatus(service)
		if err != nil {
			stopLogger.Warn(fmt.Sprintf("Could not check status of %s", service.Name), err.Error())
		} else if status == "stopped" {
			stopLogger.Info(fmt.Sprintf("%s already stopped", service.Name))
			continue
		}

		// Stop with spinner for active processes
		spin := lib.NewSpinner("stop", fmt.Sprintf("Stopping %s", service.Name))

		if err := manager.Stop(service); err != nil {
			spin.Fail("failed to stop")
			stopLogger.Warn(fmt.Sprintf("Failed to stop %s", service.Name), err.Error())
			continue
		}

		// Verify it stopped successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			spin.Fail("stopped but verification failed")
			stopLogger.Warn(fmt.Sprintf("Stopped %s but could not verify status", service.Name), err.Error())
		} else {
			spin.Success(status + " stopped successfully")
		}
	}

	// Brief pause between stop and start for services to fully shut down
	time.Sleep(1 * time.Second)

	// Then, start all services
	startLogger := logger.NewChildLogger("start")
	for _, service := range servicesToRestart {
		// Check if service is already running
		status, err := manager.GetStatus(service)
		if err != nil {
			startLogger.Warn(fmt.Sprintf("Could not check status of %s", service.Name), err.Error())
		} else if status != "stopped" {
			startLogger.Success(fmt.Sprintf("%s already %s", service.Name, status), "")
			continue
		}

		// Start with spinner for active processes
		spin := lib.NewSpinner("start", fmt.Sprintf("Starting %s", service.Name))

		if err := manager.Start(service); err != nil {
			spin.Fail("failed to start")
			startLogger.Error(fmt.Sprintf("Failed to start %s", service.Name), err.Error())
			continue
		}

		// Verify it started successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			spin.Fail("started but verification failed")
			startLogger.Warn(fmt.Sprintf("Started %s but could not verify status", service.Name), "")
		} else {
			spin.Success(status + " started successfully")
		}
	}

	// Handle port forwarding for nginx if needed
	needsPortForwarding := false
	for _, svc := range servicesToRestart {
		if svc.Name == "chauf-nginx" {
			needsPortForwarding = true
			break
		}
	}

	if needsPortForwarding {
		cfg, cfgErr := config.Load()
		if cfgErr != nil {
			logger.Warn("Port forwarding check skipped", cfgErr.Error())
		} else {
			// Validate port configuration first
			if err := system.ValidatePortConfiguration(cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
				logger.Error("Invalid port configuration", err.Error())
				return fmt.Errorf("port configuration validation failed: %w", err)
			}

			forwardingNeeded := (cfg.Nginx.HTTPPort > 0 && cfg.Nginx.HTTPPort != 80) ||
				(cfg.Nginx.HTTPSPort > 0 && cfg.Nginx.HTTPSPort != 443)
			if forwardingNeeded {
				portsLogger := logger.NewChildLogger("ports")
				if err := system.EnsurePortForwarding(cfg.WorkspaceDir, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
					portsLogger.Warn("Port forwarding not configured", err.Error())
				} else {
					portsLogger.Success("Port forwarding active", fmt.Sprintf("80→%d, 443→%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort))
				}
			}
		}
	}

	logger.Success("Restart completed", fmt.Sprintf("%d services restarted", len(servicesToRestart)))
	return nil
}

/**
 * printRestartUsage renders CLI help for the restart command.
 */
func printRestartUsage() {
	fmt.Println(`Usage: chauf restart [service...] [--project <slug>] [--all] [--dry-run]

Restarts Chauffeur services with chauf- prefix to avoid conflicts with system services.
This is equivalent to running 'chauf stop' followed by 'chauf start' for the same services.

Arguments:
  service           Restart specific service(s): nginx, php-fpm, or project slug.

Flags:
  --project <slug>  Restart nginx and project's PHP-FPM.
  --all             Restart all services (global + all projects).
  --dry-run         Show what would be restarted without taking action.
  -h, --help        Show this message.

Examples:
  chauf restart                 # Restart all Chauffeur services
  chauf restart nginx           # Restart chauf-nginx only
  chauf restart php-fpm         # Restart all chauf-php-fpm-* services
  chauf restart --project hja-cms  # Restart nginx and hja-cms's php-fpm
  chauf restart hja-cms         # Restart php-fpm for hja-cms project only

Service Names:
  - chauf-nginx              # Global Nginx service
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}