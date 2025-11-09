package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * RunStatus shows the status of Chauffeur services.
 *
 * @param args CLI arguments passed after the status subcommand.
 * @return error when status checking fails.
 */
func RunStatus(args []string) error {
	// Create logger for status command
	logger := lib.NewCommandLogger("status")

	var (
		serviceType string
		projectSlug string
		detail      bool
	)

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printStatusUsage()
			return nil
		case "--project":
			if i+1 >= len(args) {
				return logger.Error("--project requires a project slug", "missing argument")
			}
			projectSlug = args[i+1]
			i += 2
		case "--detail", "-v":
			detail = true
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return logger.Error("unknown flag for status", arg)
			}
			if serviceType != "" {
				return logger.Error("multiple service types specified", "only one service type allowed")
			}
			serviceType = arg
			i++
		}
	}

	// Ensure workspace exists
	if err := workspace.Ensure(); err != nil {
		return logger.Error("ensure workspace", err.Error())
	}

	// Create service manager
	manager, err := services.NewServiceManager()
	if err != nil {
		return logger.Error("create service manager", err.Error())
	}

	// Determine which services to show status for
	var servicesToCheck []services.Service

	if projectSlug != "" {
		// Show status for specific project only
		projectServices, err := manager.ListProjectServices(projectSlug)
		if err != nil {
			return logger.Error("list project services", err.Error())
		}
		servicesToCheck = append(servicesToCheck, projectServices...)
	} else if serviceType != "" {
		// Show status for specific service type
		switch serviceType {
		case "nginx", "chauf-nginx":
			globalServices := manager.ListGlobalServices()
			for _, svc := range globalServices {
				if strings.Contains(svc.Name, "nginx") {
					servicesToCheck = append(servicesToCheck, svc)
				}
			}
		case "php-fpm", "php":
			// Show all PHP-FPM services for all projects

			projects, err := findAllLinkedProjects()
			if err != nil {
				return logger.Error("find linked projects", err.Error())
			}
			for _, projSlug := range projects {
				projectServices, err := manager.ListProjectServices(projSlug)
				if err != nil {
					logger.Warn("Could not load services for project", fmt.Sprintf("%s: %v", projSlug, err))
					continue
				}
				servicesToCheck = append(servicesToCheck, projectServices...)
			}
		default:
			return logger.Error("unknown service type", serviceType)
		}
	} else {
		// Show status for all services
		globalServices := manager.ListGlobalServices()
		servicesToCheck = append(servicesToCheck, globalServices...)

		// Add all project services
		projects, err := findAllLinkedProjects()
		if err != nil {
			return logger.Error("find linked projects", err.Error())
		}
		for _, projSlug := range projects {
			projectServices, err := manager.ListProjectServices(projSlug)
			if err != nil {
				logger.Warn("Could not load services for project", fmt.Sprintf("%s: %v", projSlug, err))
				continue
			}
			servicesToCheck = append(servicesToCheck, projectServices...)
		}
	}

	if len(servicesToCheck) == 0 {
		if projectSlug != "" {
			logger.Info(fmt.Sprintf("No services found for project: %s", projectSlug))
		} else if serviceType != "" {
			logger.Info(fmt.Sprintf("No services found for type: %s", serviceType))
		} else {
			logger.Info("No Chauffeur services found.")
			logger.Info("Run 'chauf install <service>' to install services.")
			logger.Info("Run 'chauf start' to start Chauffeur services.")
		}
		return nil
	}

	// Display status
	startTime := time.Now()
	logger.Info(fmt.Sprintf("Checking status of %d Chauffeur services...", len(servicesToCheck)))

	if !detail {
		// Simple view
		logger.PrintSection(fmt.Sprintf("Chauffeur Services (%d total)", len(servicesToCheck)))

		// Separate global from project services
		var globalServices []services.Service
		var projectServices []services.Service

		for _, service := range servicesToCheck {
			if service.Type == services.ServiceTypeGlobal {
				globalServices = append(globalServices, service)
			} else {
				projectServices = append(projectServices, service)
			}
		}

		// Show global services
		if len(globalServices) > 0 {
			globalLogger := logger.NewChildLogger("global")
			globalLogger.PrintSection(fmt.Sprintf("Global Services (%d)", len(globalServices)))
			for _, service := range globalServices {
				status, _ := manager.GetStatus(service)
				icon := "✗"
				if strings.Contains(status, "running") {
					icon = "✓"
				}
				globalLogger.Info(fmt.Sprintf("  %s %s (%s)", icon, service.Name, status))
			}
		}

		if len(projectServices) > 0 {
			projectLogger := logger.NewChildLogger("projects")
			projectLogger.PrintSection(fmt.Sprintf("Project Services (%d)", len(projectServices)))
			for _, service := range projectServices {
				status, _ := manager.GetStatus(service)
				icon := "✗"
				if strings.Contains(status, "running") {
					icon = "✓"
				}
				projectLogger.Info(fmt.Sprintf("  %s %s (%s)", icon, service.Name, status))
			}
		}
	} else {
		// Detailed view
		logger.PrintSection(fmt.Sprintf("Chauffeur Services - Detailed Status (%d total)", len(servicesToCheck)))

		for _, service := range servicesToCheck {
			logger.Info(fmt.Sprintf("Service: %s", service.Name))
			logger.Info(fmt.Sprintf("  Type: %s", serviceTypeToString(service.Type)))
			if service.Slug != "" {
				logger.Info(fmt.Sprintf("  Project: %s", service.Slug))
			}
			logger.Info(fmt.Sprintf("  Binary: %s", service.Binary))
			logger.Info(fmt.Sprintf("  PID File: %s", service.PIDFile))

			status, err := manager.GetStatus(service)
			if err != nil {
				logger.Warn("Status lookup failed", err.Error())
			} else {
				logger.Info(fmt.Sprintf("  Status: %s", status))
			}

			logger.Info("------------------------------")
		}
	}

	// Log completion
	duration := time.Since(startTime)
	// Simple duration formatting
	if duration < time.Second {
		logger.Success("Status check complete", fmt.Sprintf("%.0fms", float64(duration.Nanoseconds())/1000000.0))
	} else if duration < time.Minute {
		logger.Success("Status check complete", fmt.Sprintf("%.1fs", float64(duration.Nanoseconds())/1000000000.0))
	} else {
		logger.Success("Status check complete", fmt.Sprintf("%.0fm %.0fs", duration.Minutes(), float64(duration.Seconds())-float64(int(duration.Seconds()))))
	}

	return nil
}

func serviceTypeToString(serviceType services.ServiceType) string {
	switch serviceType {
	case services.ServiceTypeGlobal:
		return "Global"
	case services.ServiceTypeProject:
		return "Project"
	default:
		return "Unknown"
	}
}

/**
 * printStatusUsage renders CLI help for the status command.
 */
func printStatusUsage() {
	fmt.Println(`Usage: chauf status [service-type] [--project <slug>] [--detail]

Shows the status of Chauffeur services with chauf- prefix.

Arguments:
  service-type        Show status for specific service type (nginx, php-fpm).

Flags:
  --project <slug>     Show status for specific project only.
  --detail, -v         Show detailed status information.
  -h, --help          Show this message.

Examples:
  chauf status                  # Show all Chauffeur services
  chauf status nginx           # Show nginx service status
  chauf status php-fpm         # Show all PHP-FPM services
  chauf status --project hja-cms  # Show services for hja-cms project
  chauf status --detail        # Show detailed information

Service Names:
  - chauf-nginx              # Global Nginx service
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}
