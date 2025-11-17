package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// ServiceHealthInfo contains detailed health information about a service
type ServiceHealthInfo struct {
	Status       string
	Icon         string
	PID          int
	Uptime       string
	MemoryUsage  string
	CPUUsage     float64
	ProcessCount int
	SocketInfo   string
	ConfigPath   string
}

// getHealthIcon returns appropriate status icon based on service health
func getHealthIcon(isRunning bool, hasWarnings bool, hasErrors bool) string {
	if hasErrors {
		return "🔴"
	}
	if hasWarnings {
		return "⚠️"
	}
	if isRunning {
		return "✅"
	}
	return "🔘"
}

// getServiceHealthInfo collects detailed health information about a service
func getServiceHealthInfo(svc services.Service, status string) ServiceHealthInfo {
	info := ServiceHealthInfo{
		Status:     status,
		ConfigPath: getConfigPath(svc),
		SocketInfo: getSocketInfo(svc),
	}

	isRunning := strings.Contains(status, "running")
	hasWarnings := strings.Contains(status, "warning")
	hasErrors := strings.Contains(status, "error") || strings.Contains(status, "failed")

	info.Icon = getHealthIcon(isRunning, hasWarnings, hasErrors)

	// Extract PID from status if available
	if strings.Contains(status, "pid ") {
		pidStr := strings.TrimPrefix(status, "running (pid ")
		pidStr = strings.TrimSuffix(pidStr, ")")
		if pid, err := strconv.Atoi(pidStr); err == nil {
			info.PID = pid
			info.Uptime = getProcessUptime(pid)
			info.MemoryUsage = getProcessMemory(pid)
			info.CPUUsage = getProcessCPU(pid)
		}
	}

	// Get service-specific information
	info.ProcessCount = getServiceProcessCount(svc)

	return info
}

// getProcessUptime returns the uptime of a process given its PID
func getProcessUptime(pid int) string {
	if pid == 0 {
		return "N/A"
	}

	// Try to get process start time from /proc/<pid>/stat
	if statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		fields := strings.Fields(string(statData))
		if len(fields) >= 22 {
			// Field 22 is the start time in jiffies since boot
			startTimeJiffies, err := strconv.ParseInt(fields[21], 10, 64)
			if err == nil {
				// Get system boot time and uptime
				var bootTime time.Time
				var systemUptimeSeconds float64
				if uptimeData, err := os.ReadFile("/proc/uptime"); err == nil {
					uptimeFields := strings.Fields(string(uptimeData))
					if len(uptimeFields) >= 1 {
						systemUptimeSeconds, _ = strconv.ParseFloat(uptimeFields[0], 64)
						bootTime = time.Now().Add(-time.Duration(systemUptimeSeconds) * time.Second)
					}
				}

				// Get system clock frequency (jiffies per second)
				// Default to 100 which is common on most Linux systems
				clockTick := int64(100)
				if tickData, err := os.ReadFile("/sys/kernel/smp/clk_tick"); err == nil {
					if tick, parseErr := strconv.ParseInt(strings.TrimSpace(string(tickData)), 10, 64); parseErr == nil {
						clockTick = tick
					}
				}

				// Calculate process start time: boot time + (start time jiffies / clock ticks per second)
				startTimeSeconds := float64(startTimeJiffies) / float64(clockTick)
				startTime := bootTime.Add(time.Duration(startTimeSeconds) * time.Second)
				uptime := time.Since(startTime)

				// Format uptime appropriately
				if uptime < time.Minute {
					return fmt.Sprintf("%.0fs", uptime.Seconds())
				} else if uptime < time.Hour {
					return fmt.Sprintf("%.0fm", uptime.Minutes())
				} else if uptime < 24*time.Hour {
					return fmt.Sprintf("%.0fh", uptime.Hours())
				} else {
					return fmt.Sprintf("%.0fd", uptime.Hours()/24)
				}
			}
		}
	}

	return "N/A"
}

// getProcessMemory returns memory usage for a process PID
func getProcessMemory(pid int) string {
	if pid == 0 {
		return "N/A"
	}

	// Try to get memory usage from /proc/<pid>/status
	if statusData, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid)); err == nil {
		lines := strings.Split(string(statusData), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb, err := strconv.Atoi(fields[1])
					if err == nil {
						if kb < 1024 {
							return fmt.Sprintf("%dKB", kb)
						} else if kb < 1024*1024 {
							return fmt.Sprintf("%.1fMB", float64(kb)/1024)
						} else {
							return fmt.Sprintf("%.1fGB", float64(kb)/(1024*1024))
						}
					}
				}
			}
		}
	}

	return "N/A"
}

// getProcessCPU returns CPU usage percentage for a process PID (simplified)
func getProcessCPU(pid int) float64 {
	// This is a simplified implementation. A full implementation would track CPU over time.
	return 0.0
}

// getServiceProcessCount returns the number of processes for a service
func getServiceProcessCount(svc services.Service) int {
	if strings.Contains(svc.Name, "php-fpm") {
		// For PHP-FPM, count all php-fpm processes (master + workers)
		// Use the binary path from the service to get the correct processes
		binaryName := "php-fpm"
		if svc.Binary != "" {
			// Extract just the binary name from the full path
			parts := strings.Split(svc.Binary, "/")
			if len(parts) > 0 {
				binaryName = parts[len(parts)-1]
			}
		}

		cmd := exec.Command("pgrep", "-f", binaryName)
		if output, err := cmd.Output(); err == nil {
			processes := strings.Fields(strings.TrimSpace(string(output)))
			return len(processes)
		}
	} else if strings.Contains(svc.Name, "nginx") {
		// For nginx, count all nginx processes
		cmd := exec.Command("pgrep", "-f", "nginx")
		if output, err := cmd.Output(); err == nil {
			processes := strings.Fields(strings.TrimSpace(string(output)))
			return len(processes)
		}
	}
	return 0
}

// getConfigPath returns the configuration file path for a service
func getConfigPath(svc services.Service) string {
	if strings.Contains(svc.Name, "nginx") {
		return "~/.chauffeur/nginx/etc/nginx.conf"
	} else if strings.Contains(svc.Name, "php-fpm") {
		if svc.Slug != "" {
			return fmt.Sprintf("~/.chauffeur/projects/%s/runtime/php-fpm/php-fpm.conf", svc.Slug)
		}
		return "~/.chauffeur/php/*/etc/php-fpm.conf"
	}
	return "N/A"
}

// getSocketInfo returns socket file information for a service
func getSocketInfo(svc services.Service) string {
	if strings.Contains(svc.Name, "php-fpm") {
		if svc.Slug != "" {
			return fmt.Sprintf("~/.chauffeur/projects/%s/runtime/php-fpm/php-fpm.sock", svc.Slug)
		}
	}
	return "N/A"
}

// formatServiceType returns a formatted service type string
func formatServiceType(serviceType services.ServiceType) string {
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

	// Deduplicate services (remove duplicates with same name)
	servicesToCheck = deduplicateServices(servicesToCheck)

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
		// Enhanced simple view with table formatting
		logger.PrintSection(fmt.Sprintf("Chauffeur Services (%d total)", len(servicesToCheck)))

		// Prepare table data
		tableData := [][]string{}
		for _, service := range servicesToCheck {
			status, _ := manager.GetStatus(service)
			healthInfo := getServiceHealthInfo(service, status)

			serviceType := formatServiceType(service.Type)
			projectDisplay := service.Slug
			if projectDisplay == "" {
				projectDisplay = "—"
			}

			row := []string{
				healthInfo.Icon,           // Status icon
				service.Name,              // Service name
				serviceType,               // Service type
				projectDisplay,            // Project
				healthInfo.Uptime,         // Uptime
				healthInfo.MemoryUsage,    // Memory usage
			}
			tableData = append(tableData, row)
		}

		// Print service table
		headers := []string{"", "SERVICE", "TYPE", "PROJECT", "UPTIME", "MEMORY"}
		logger.PrintServiceTable(headers, tableData)

		// Add legend
		logger.Info("")
		logger.Info("Legend: ✅ Running  ⚠️ Warning  🔴 Error  🔘 Stopped")
	} else {
		// Enhanced detailed view with health information
		logger.PrintSection(fmt.Sprintf("Chauffeur Services - Detailed Status (%d total)", len(servicesToCheck)))

		for i, service := range servicesToCheck {
			status, err := manager.GetStatus(service)
			if err != nil {
				logger.Warn("Status lookup failed", err.Error())
				continue
			}

			healthInfo := getServiceHealthInfo(service, status)

			// Service header
			serviceHeader := fmt.Sprintf("Service: %s %s", healthInfo.Icon, service.Name)
			if i > 0 {
				logger.Info("")
			}
			logger.PrintSection(serviceHeader)

			// Service metadata table
			metaData := [][]string{
				{"Type", formatServiceType(service.Type)},
				{"Status", healthInfo.Status},
				{"Binary", service.Binary},
				{"PID File", service.PIDFile},
			}

			if service.Slug != "" {
				metaData = append(metaData, []string{"Project", service.Slug})
			}

			if healthInfo.PID > 0 {
				metaData = append(metaData, []string{"PID", fmt.Sprintf("%d", healthInfo.PID)})
			}

			logger.PrintEnhancedTable([]string{"PROPERTY", "VALUE"}, metaData, []string{"left", "left"})

			// Health information table
			logger.Info("")
			healthData := [][]string{
				{"Uptime", healthInfo.Uptime},
				{"Memory Usage", healthInfo.MemoryUsage},
				{"Process Count", fmt.Sprintf("%d", healthInfo.ProcessCount)},
			}

			if healthInfo.SocketInfo != "N/A" {
				healthData = append(healthData, []string{"Socket", healthInfo.SocketInfo})
			}

			if healthInfo.ConfigPath != "N/A" {
				healthData = append(healthData, []string{"Config", healthInfo.ConfigPath})
			}

			logger.PrintEnhancedTable([]string{"METRIC", "VALUE"}, healthData, []string{"left", "left"})
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

// deduplicateServices removes duplicate services with the same name, keeping the first occurrence
func deduplicateServices(serviceList []services.Service) []services.Service {
	seen := make(map[string]bool)
	var result []services.Service

	for _, service := range serviceList {
		if !seen[service.Name] {
			seen[service.Name] = true
			result = append(result, service)
		}
	}

	return result
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
