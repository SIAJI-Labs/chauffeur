package commands

import (
	"fmt"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunStatus shows the status of Chauffeur services.
 *
 * @param args CLI arguments passed after the status subcommand.
 * @return error when status checking fails.
 */
func RunStatus(args []string) error {
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
				return fmt.Errorf("--project requires a project slug")
			}
			projectSlug = args[i+1]
			i += 2
		case "--detail", "-v":
			detail = true
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for status: %s", arg)
			}
			if serviceType != "" {
				return fmt.Errorf("multiple service types specified")
			}
			serviceType = arg
			i++
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

	// Determine which services to show status for
	var servicesToCheck []services.Service

	if projectSlug != "" {
		// Show status for specific project only
		projectServices, err := manager.ListProjectServices(projectSlug)
		if err != nil {
			return fmt.Errorf("list project services: %w", err)
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
		case "caddy", "chauf-caddy":
			globalServices := manager.ListGlobalServices()
			for _, svc := range globalServices {
				if strings.Contains(svc.Name, "caddy") {
					servicesToCheck = append(servicesToCheck, svc)
				}
			}
		case "php-fpm", "php":
			// Show all PHP-FPM services for all projects

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
				servicesToCheck = append(servicesToCheck, projectServices...)
			}
		default:
			return fmt.Errorf("unknown service type: %s", serviceType)
		}
	} else {
		// Show status for all services
		globalServices := manager.ListGlobalServices()
		servicesToCheck = append(servicesToCheck, globalServices...)

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
			servicesToCheck = append(servicesToCheck, projectServices...)
		}
	}

	if len(servicesToCheck) == 0 {
		if projectSlug != "" {
			fmt.Printf("No services found for project: %s\n", projectSlug)
		} else if serviceType != "" {
			fmt.Printf("No services found for type: %s\n", serviceType)
		} else {
			fmt.Println("No Chauffeur services found.")
			fmt.Println("Run 'chauf install <service>' to install services.")
			fmt.Println("Run 'chauf start' to start Chauffeur services.")
		}
		return nil
	}

	// Display status
	if !detail {
		// Simple view
		fmt.Printf("Chauffeur Services (%d total):\n", len(servicesToCheck))
		fmt.Println()

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

		if len(globalServices) > 0 {
			fmt.Println("Global Services:")
			for _, service := range globalServices {
				status, _ := manager.GetStatus(service)
				if strings.Contains(status, "running") {
					fmt.Printf("  ✓ %s (%s)\n", service.Name, status)
				} else {
					fmt.Printf("  ✗ %s (%s)\n", service.Name, status)
				}
			}
			fmt.Println()
		}

		if len(projectServices) > 0 {
			fmt.Println("Project Services:")
			for _, service := range projectServices {
				status, _ := manager.GetStatus(service)
				if strings.Contains(status, "running") {
					fmt.Printf("  ✓ %s (%s)\n", service.Name, status)
				} else {
					fmt.Printf("  ✗ %s (%s)\n", service.Name, status)
				}
			}
		}
	} else {
		// Detailed view
		fmt.Printf("Chauffeur Services - Detailed Status (%d total):\n", len(servicesToCheck))
		fmt.Println()

		for _, service := range servicesToCheck {
			fmt.Printf("Service: %s\n", service.Name)
			fmt.Printf("  Type: %s\n", serviceTypeToString(service.Type))
			if service.Slug != "" {
				fmt.Printf("  Project: %s\n", service.Slug)
			}
			fmt.Printf("  Binary: %s\n", service.Binary)
			fmt.Printf("  PID File: %s\n", service.PIDFile)
			
			status, err := manager.GetStatus(service)
			if err != nil {
				fmt.Printf("  Status: Unknown (error: %v)\n", err)
			} else {
				fmt.Printf("  Status: %s\n", status)
			}

			fmt.Println()
		}
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
  service-type        Show status for specific service type (nginx, caddy, php-fpm).

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
  - chauf-caddy              # Global Caddy service  
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}
