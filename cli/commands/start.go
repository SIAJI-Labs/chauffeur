package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		return dnsLogger.Fail("dnsmasq not available", "Install dnsmasq first: sudo pacman -S dnsmasq (Arch) or sudo apt install dnsmasq (Ubuntu/Debian)")
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

		// Test DNS resolution with retries
		if err := verifyDNSResolutionWithRetry(dnsLogger, 3); err != nil {
			// If resolution still fails after retries, attempt to fix it
			dnsLogger.Warn("DNS resolution verification failed", "Attempting to restart dnsmasq service")

			if setupErr := system.SetupLocalDNSResolution(); setupErr != nil {
				if system.IsNetworkManagerDnsmasqRunning() {
					dnsLogger.Warn("NetworkManager dnsmasq restart had issues", setupErr.Error())
					dnsLogger.Info("Trying alternative DNS resolution check...")

					// Give it a moment for NetworkManager to settle
					time.Sleep(2 * time.Second)

					// Try verification one more time
					if retryErr := verifyDNSResolutionWithRetry(dnsLogger, 2); retryErr != nil {
						return dnsLogger.Fail("DNS resolution unavailable", fmt.Sprintf("Final error: %v\nTroubleshooting: Check NetworkManager logs or restart with 'sudo systemctl restart NetworkManager'", retryErr))
					}
				} else {
					return dnsLogger.Fail("failed to restart dnsmasq", fmt.Sprintf("%v\nTry: sudo systemctl restart dnsmasq", setupErr))
				}
			} else {
				// Setup succeeded, verify again
				if retryErr := verifyDNSResolutionWithRetry(dnsLogger, 2); retryErr != nil {
					return dnsLogger.Fail("DNS resolution still failing after setup", fmt.Sprintf("Final error: %v", retryErr))
				}
			}
		}

		dnsLogger.Success("dnsmasq configuration validated", fmt.Sprintf("%s resolves correctly", system.DNSProbeDomain))
		return nil
	}

	return setupDnsmasqConfiguration(dnsLogger)
}

/**
 * verifyDNSResolutionWithRetry attempts to verify DNS resolution with multiple retries.
 */
func verifyDNSResolutionWithRetry(dnsLogger *lib.Logger, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			dnsLogger.Info(fmt.Sprintf("Retrying DNS verification (attempt %d/%d)...", attempt, maxRetries))
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resolves, ips, err := system.VerifyLocalDNSResolution()
		if err != nil {
			lastErr = err
			dnsLogger.Warn(fmt.Sprintf("DNS probe failed (attempt %d/%d)", attempt, maxRetries), err.Error())
			continue
		}

		if resolves {
			resolvedIPs := "127.0.0.1"
			if len(ips) > 0 {
				resolvedIPs = strings.Join(ips, ", ")
			}
			dnsLogger.Info(fmt.Sprintf("DNS resolution working: %s → %s", system.DNSProbeDomain, resolvedIPs))
			return nil
		}

		lastErr = fmt.Errorf("domain %s resolved to %v instead of 127.0.0.1", system.DNSProbeDomain, ips)
		dnsLogger.Warn(fmt.Sprintf("Incorrect DNS resolution (attempt %d/%d)", attempt, maxRetries), lastErr.Error())
	}

	return lastErr
}

/**
 * setupDnsmasqConfiguration prompts user to set up dnsmasq configuration.
 */
func setupDnsmasqConfiguration(dnsLogger *lib.Logger) error {
	dnsLogger.Warn("dnsmasq configuration not found", "Local .test domains won't resolve")
	dnsLogger.Info("Chauffeur requires dnsmasq configuration to resolve .test domains")

	// Detect the best setup approach
	if system.IsNetworkManagerAvailable() {
		dnsLogger.Info("NetworkManager detected - can configure automatically")
		dnsLogger.Info("This will configure NetworkManager's dnsmasq plugin for .test domains")
	} else {
		dnsLogger.Info("Standalone dnsmasq setup required")
		dnsLogger.Info("This will create /etc/dnsmasq.d/chauffeur.conf and restart dnsmasq")
	}

	dnsLogger.Prompt("Do you want to configure dnsmasq now?", "Type 'y' to continue, 'n' to skip")

	// SENSITIVE: User input confirmation - system configuration consent
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		return dnsLogger.Fail("configuration declined", "Local .test domains will not work without this configuration")
	}

	// Use the unified setup function
	if err := system.SetupLocalDNSResolution(); err != nil {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Warn("NetworkManager setup completed with warnings", err.Error())
			dnsLogger.Info("Attempting to verify configuration...")

			// Give NetworkManager time to settle
			time.Sleep(3 * time.Second)

			if verifyErr := verifyDNSResolutionWithRetry(dnsLogger, 3); verifyErr != nil {
				return dnsLogger.Fail("configuration verification failed", fmt.Sprintf("NetworkManager setup had issues: %v", verifyErr))
			}
		} else {
			return dnsLogger.Fail("setup dnsmasq configuration", fmt.Sprintf("%v\nManual setup: Edit /etc/dnsmasq.d/chauffeur.conf and restart dnsmasq", err))
		}
	} else {
		dnsLogger.Success("dnsmasq configuration completed", "Verifying DNS resolution...")

		// Verify the configuration works
		time.Sleep(2 * time.Second)
		if verifyErr := verifyDNSResolutionWithRetry(dnsLogger, 3); verifyErr != nil {
			return dnsLogger.Fail("configuration verification failed", fmt.Sprintf("Setup completed but DNS not working: %v", verifyErr))
		}
	}

	dnsLogger.Success("dnsmasq setup successful", fmt.Sprintf("%s now resolves to localhost", system.DNSProbeDomain))
	return nil
}

/**
 * RunStart starts Chauffeur services.
 *
 * @param args CLI arguments passed after the start subcommand.
 * @return error when prerequisite checks or start operations fail.
 */
func RunStart(args []string) error {
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
			// Validate port configuration first
			if err := system.ValidatePortConfiguration(cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
				logger.Error("Invalid port configuration", err.Error())
				return fmt.Errorf("port configuration validation failed: %w", err)
			}

			forwardingNeeded := (cfg.Nginx.HTTPPort > 0 && cfg.Nginx.HTTPPort != 80) ||
				(cfg.Nginx.HTTPSPort > 0 && cfg.Nginx.HTTPSPort != 443)
			if forwardingNeeded {
				portsLogger := logger.NewChildLogger("ports")

				// Check current port forwarding status
				currentHTTP, currentHTTPS, err := system.GetPortForwardingStatus(cfg.WorkspaceDir)
				if err != nil {
					portsLogger.Warn("Could not check current port forwarding status", err.Error())
				} else {
					portsLogger.Info(fmt.Sprintf("Current port forwarding: HTTP 80→%d, HTTPS 443→%d", currentHTTP, currentHTTPS))
				}

				if err := system.EnsurePortForwarding(cfg.WorkspaceDir, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
					portsLogger.Error("Port forwarding failed", err.Error())
					// Don't return error here - allow nginx to start without port forwarding
					portsLogger.Info(fmt.Sprintf("nginx will start without port forwarding - access directly on ports %d/%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort))
				} else {
					portsLogger.Success("Port forwarding configured", fmt.Sprintf("HTTP 80→%d, HTTPS 443→%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort))
				}
			} else {
				logger.Info("No port forwarding needed (using default privileged ports)")
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
