package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * checkDnsmasqConfiguration validates if dnsmasq is configured for local domain resolution.
 *
 * @param logger Command logger for status reporting.
 * @return error if dnsmasq is not configured and user declines to configure it.
 */
func checkDnsmasqConfiguration(logger *lib.Logger) error {
	logger.Info("Checking dnsmasq configuration for local domain resolution...")

	dnsLogger := logger.NewChildLogger("dns")

	// Check if dnsmasq is available (NetworkManager or standalone)
	if !system.IsDnsmasqAvailable() {
		return dnsLogger.Fail("dnsmasq not available", "Install dnsmasq first")
	}

	// Check if chauffeur.conf exists in either location
	configPaths := []string{
		"/etc/dnsmasq.d/chauffeur.conf",
		"/etc/NetworkManager/dnsmasq.d/chauffeur.conf",
	}

	var (
		configPath   string
		configSource string
	)

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			if strings.Contains(path, "NetworkManager") {
				configSource = "NetworkManager"
			} else {
				configSource = "standalone"
			}
			break
		}
	}

	if configPath != "" {
		dnsLogger.Info(fmt.Sprintf("dnsmasq configuration found at %s (%s)", configPath, configSource))

		resolves, ips, err := system.VerifyLocalDNSResolution()
		if err != nil {
			dnsLogger.Warn("dns probe failed", err.Error())
		}

		if !resolves {
			dnsLogger.Warn(".test domains are not resolving to localhost", "Attempting to restart dnsmasq/NetworkManager")

			if err := system.SetupLocalDNSResolution(); err != nil {
				if system.IsNetworkManagerDnsmasqRunning() {
					dnsLogger.Warn("dnsmasq restart reported warnings", err.Error())
				} else {
					return dnsLogger.Fail("restart dnsmasq", err.Error())
				}
			}

			var retryErr error
			resolves, ips, retryErr = system.VerifyLocalDNSResolution()
			if retryErr != nil {
				return dnsLogger.Fail("verify .test domain resolution", retryErr.Error())
			}

			if !resolves {
				dnsLogger.Warn("Local resolver is not using dnsmasq", "Ensure /etc/resolv.conf points to 127.0.0.1 or configure NetworkManager to use dnsmasq for .test domains")
				return dnsLogger.Fail("local .test domain resolution unavailable", "Update your resolver configuration or follow the README dnsmasq instructions")
			}
		}

		resolvedIPs := "127.0.0.1"
		if len(ips) > 0 {
			resolvedIPs = strings.Join(ips, ", ")
		}
		dnsLogger.Success("dnsmasq configuration validated", fmt.Sprintf("%s resolves to %s", system.DNSProbeDomain, resolvedIPs))
		return nil
	}

	dnsLogger.Warn("dnsmasq configuration not found", "Local .test domains won't resolve")
	dnsLogger.Info("Chauffeur requires dnsmasq configuration to resolve .test domains")
	dnsLogger.Info("Add this configuration to make .test domains resolve to localhost:")

	dnsLogger.PrintSection("Required dnsmasq configuration")
	configLines := []string{
		"sudo install -d -m 755 /etc/dnsmasq.d",
		"sudo tee /etc/dnsmasq.d/chauffeur.conf >/dev/null <<'EOF'",
		"# Chauffeur local development resolver",
		"# Redirect all *.test domains to localhost",
		"address=/.test/127.0.0.1",
		"# Only listen locally",
		"listen-address=127.0.0.1",
		"bind-interfaces",
		"EOF",
	}
	for _, line := range configLines {
		dnsLogger.Info(line)
	}

	dnsLogger.Warn("Do you want to add this configuration now?", "Type 'y' to continue")
	// SENSITIVE: User input confirmation - system configuration consent
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		return dnsLogger.Fail("configuration declined", "Local .test domains will not work without this configuration")
	}

	// Use the new unified setup function
	if err := system.SetupLocalDNSResolution(); err != nil {
		if system.IsNetworkManagerDnsmasqRunning() {
			// Don't fail hard for NetworkManager conflicts
			dnsLogger.Warn("dnsmasq setup completed with warnings", "NetworkManager is managing DNS resolution")
			dnsLogger.Info("Local .test domains should now resolve to localhost (Configuration updated via NetworkManager)")
		} else {
			return dnsLogger.Fail("setup dnsmasq configuration", err.Error())
		}
	} else {
		dnsLogger.Success("dnsmasq configuration completed", "Local .test domains should now resolve to localhost")
	}

	return nil
}

/**
 * RunStart starts Chauffeur services.
 *
 * @param args CLI arguments passed after the start subcommand.
 * @return error when prerequisite checks or start operations fail.
 */
func RunStart(args []string) error {
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

	// Create command logger
	logger := lib.NewCommandLogger("start")

	// Check dnsmasq configuration before starting services
	if err := checkDnsmasqConfiguration(logger); err != nil {
		return err
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
				logger.Warn(fmt.Sprintf("Could not load services for project %s", projectSlug), fmt.Sprintf("error: %v", err))
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
							logger.Warn(fmt.Sprintf("Could not load services for project %s", projectSlug), fmt.Sprintf("error: %v", err))
							continue
						}
						servicesToStart = append(servicesToStart, projectServices...)
					}
				}
			default:
				// Check if it's a specific project slug for php-fpm
				projectServices, err := manager.ListProjectServices(serviceName)
				if err != nil {
					return fmt.Errorf("invalid service name: %s (try nginx, php-fpm, or a project slug)", serviceName)
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
		logger.Info("No services to start.")
		return nil
	}

	needsPortForwarding := false
	for _, svc := range servicesToStart {
		if svc.Name == "chauf-nginx" {
			needsPortForwarding = true
			break
		}
	}

	if needsPortForwarding {
		cfg, cfgErr := config.Load()
		if cfgErr != nil {
			logger.Warn("Port forwarding skipped", cfgErr.Error())
		} else {
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

	if dryRun {
		logger.Info(fmt.Sprintf("Would start %d services:", len(servicesToStart)))
		for _, service := range servicesToStart {
			status, _ := manager.GetStatus(service)
			logger.Info(fmt.Sprintf("  - %s (%s)", service.Name, status))
		}
		return nil
	}

	// Start services
	for _, service := range servicesToStart {
		// Check if service is already running
		status, err := manager.GetStatus(service)
		if err != nil {
			logger.Warn(fmt.Sprintf("Could not check status of %s", service.Name), fmt.Sprintf("error: %v", err))
		} else if status != "stopped" {
			logger.Success(fmt.Sprintf("%s already %s", service.Name, status), "")
			continue
		}

		// Start with spinner for active processes
		spin := lib.NewSpinner("start", fmt.Sprintf("Starting %s", service.Name))

		if err := manager.Start(service); err != nil {
			spin.Fail("failed to start")
			logger.Error(fmt.Sprintf("Failed to start %s", service.Name), err.Error())
			continue
		}

		// Verify it started successfully
		status, err = manager.GetStatus(service)
		if err != nil {
			spin.Fail("started but verification failed")
			logger.Warn(fmt.Sprintf("Started %s but could not verify status", service.Name), "")
		} else {
			spin.Success(status + " started successfully")
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
  service           Start specific service(s): nginx, php-fpm, or project slug.

Flags:
  --project <slug>  Start services for specific project (global + project services).
  --all             Start all services (global + all projects).
  --dry-run         Show what would be started without taking action.
  -h, --help        Show this message.

Examples:
  chauf start                 # Start all Chauffeur services
  chauf start nginx           # Start chauf-nginx only
  chauf start php-fpm         # Start all chauf-php-fpm-* services
  chauf start --project hja-cms  # Start nginx and hja-cms's php-fpm
  chauf start hja-cms         # Start php-fpm for hja-cms project only

Service Names:
  - chauf-nginx              # Global Nginx service
  - chauf-php-fpm-<slug>     # Project-specific PHP-FPM service`)
}
