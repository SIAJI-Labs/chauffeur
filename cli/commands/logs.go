package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// LogEntry represents a single log entry with metadata
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Service   string
	PID       int
}

// LogCollectionOptions defines options for log collection
type LogCollectionOptions struct {
	ServiceName  string
	Lines        int
	Follow       bool
	Since        string
	Until        string
	Level        string
	ShowContext  bool
	Verbose      bool
	Quiet        bool
}

// LogCollector handles log collection from various services
type LogCollector struct {
	serviceManager *services.ServiceManager
	logger         *lib.Logger
}

// NewLogCollector creates a new log collector instance
func NewLogCollector() (*LogCollector, error) {
	manager, err := services.NewServiceManager()
	if err != nil {
		return nil, fmt.Errorf("create service manager: %w", err)
	}

	return &LogCollector{
		serviceManager: manager,
		logger:         lib.NewCommandLogger("logs"),
	}, nil
}

// RunLogs implements the logs command
func RunLogs(args []string) error {
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

	collector, err := NewLogCollector()
	if err != nil {
		return fmt.Errorf("create log collector: %w", err)
	}

	options, err := parseLogArgs(args)
	if err != nil {
		return err
	}

	// If help was requested, options will be nil
	if options == nil {
		return nil
	}

	// Ensure workspace exists
	if err := workspace.Ensure(); err != nil {
		return collector.logger.Error("ensure workspace", err.Error())
	}

	return collector.CollectLogs(options)
}

// parseLogArgs parses command line arguments for the logs command
func parseLogArgs(args []string) (*LogCollectionOptions, error) {
	options := &LogCollectionOptions{
		Lines: 100, // Default to last 100 lines
	}

	// Collect positional arguments (service names and versions)
	var positionalArgs []string

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printLogsUsage()
			return nil, nil
		case "--follow", "-f":
			options.Follow = true
			i++
		case "--lines", "-n":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--lines requires a number")
			}
			lines, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid line count: %s", args[i+1])
			}
			options.Lines = lines
			i += 2
		case "--since":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--since requires a time specification")
			}
			options.Since = args[i+1]
			i += 2
		case "--until":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--until requires a time specification")
			}
			options.Until = args[i+1]
			i += 2
		case "--level":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--level requires a log level")
			}
			options.Level = strings.ToLower(args[i+1])
			i += 2
		case "--context", "-c":
			options.ShowContext = true
			i++
		case "--verbose", "-v":
			options.Verbose = true
			i++
		case "--quiet", "-q":
			options.Quiet = true
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			positionalArgs = append(positionalArgs, arg)
			i++
		}
	}

	// Process service name and version specification
	if len(positionalArgs) > 0 {
		if len(positionalArgs) == 1 {
			// Single service name (e.g., "nginx" or "php")
			options.ServiceName = positionalArgs[0]
		} else if len(positionalArgs) == 2 {
			// Service name with version (e.g., "php" "7.4")
			options.ServiceName = fmt.Sprintf("%s-%s", positionalArgs[0], positionalArgs[1])
		} else {
			return nil, fmt.Errorf("too many arguments. Usage: chauf logs [service-name] [version]")
		}
	}

	return options, nil
}

// CollectLogs collects and displays logs based on the provided options
func (lc *LogCollector) CollectLogs(options *LogCollectionOptions) error {
	if options.ServiceName == "" {
		return lc.listAvailableServices()
	}

	// Find the service
	matchingServices, err := lc.findServices(options.ServiceName)
	if err != nil {
		return lc.logger.Error("find services", err.Error())
	}

	if len(matchingServices) == 0 {
		return lc.logger.Error("service not found", fmt.Sprintf("no service matching '%s' found", options.ServiceName))
	}

	// If multiple services found and user hasn't specified exact version, show selection
	if len(matchingServices) > 1 && !strings.Contains(options.ServiceName, "-") {
		selectedService, err := lc.selectService(matchingServices, options)
		if err != nil {
			return err
		}
		if selectedService != nil {
			matchingServices = []services.Service{*selectedService}
		}
		// If selectedService is nil, user wants to see all services
	}

	// Collect logs from each matching service
	for _, service := range matchingServices {
		if len(matchingServices) > 1 {
			if !options.Quiet {
				lc.logger.PrintSection(fmt.Sprintf("Logs for %s", service.Name))
			}
		}

		err := lc.collectServiceLogs(service, options)
		if err != nil {
			lc.logger.Warn("failed to collect logs", fmt.Sprintf("%s: %v", service.Name, err))
		}

		if len(matchingServices) > 1 && !options.Quiet {
			fmt.Println() // Add spacing between services
		}
	}

	return nil
}

// findServices finds services matching the given name pattern
func (lc *LogCollector) findServices(namePattern string) ([]services.Service, error) {
	var matchingServices []services.Service
	pattern := strings.ToLower(namePattern)
	seen := make(map[string]bool) // Track service names to avoid duplicates

	// Check global services
	globalServices := lc.serviceManager.ListGlobalServices()
	for _, service := range globalServices {
		if strings.Contains(strings.ToLower(service.Name), pattern) {
			if !seen[service.Name] {
				matchingServices = append(matchingServices, service)
				seen[service.Name] = true
			}
		}
	}

	// Check project services
	projects, err := findAllLinkedProjects()
	if err != nil {
		return nil, fmt.Errorf("find linked projects: %w", err)
	}

	for _, projectSlug := range projects {
		projectServices, err := lc.serviceManager.ListProjectServices(projectSlug)
		if err != nil {
			continue // Skip projects that can't be loaded
		}

		for _, service := range projectServices {
			if strings.Contains(strings.ToLower(service.Name), pattern) {
				if !seen[service.Name] {
					matchingServices = append(matchingServices, service)
					seen[service.Name] = true
				}
			}
		}
	}

	return matchingServices, nil
}

// collectServiceLogs collects logs from a specific service
func (lc *LogCollector) collectServiceLogs(service services.Service, options *LogCollectionOptions) error {
	// Determine log file locations based on service type
	logFiles, err := lc.getServiceLogFiles(service)
	if err != nil {
		return fmt.Errorf("get log files: %w", err)
	}

	if len(logFiles) == 0 {
		lc.logger.Info(fmt.Sprintf("no log files found for service: %s", service.Name))
		return nil
	}

	// Collect logs from each file
	for _, logFile := range logFiles {
		if options.Verbose {
			lc.logger.Info(fmt.Sprintf("reading log file: %s", logFile))
		}

		err := lc.readLogFile(logFile, service, options)
		if err != nil {
			lc.logger.Warn("failed to read log file", fmt.Sprintf("%s: %v", logFile, err))
		}
	}

	// If follow is enabled, tail the log files
	if options.Follow {
		return lc.tailLogFiles(logFiles, service, options)
	}

	return nil
}

// getServiceLogFiles returns the log file paths for a service
func (lc *LogCollector) getServiceLogFiles(service services.Service) ([]string, error) {
	var logFiles []string

	if strings.Contains(service.Name, "nginx") {
		// Nginx logs
		wsDir, err := workspace.Dir()
		if err != nil {
			return nil, fmt.Errorf("get workspace directory: %w", err)
		}

		nginxLogDir := filepath.Join(wsDir, "nginx", "logs")

		// Common nginx log files
		potentialLogs := []string{
			"access.log",
			"error.log",
			"nginx-error.log",
		}

		for _, logName := range potentialLogs {
			logPath := filepath.Join(nginxLogDir, logName)
			if _, err := os.Stat(logPath); err == nil {
				logFiles = append(logFiles, logPath)
			}
		}

	} else if strings.Contains(service.Name, "php-fpm") {
		// PHP-FPM logs
		wsDir, err := workspace.Dir()
		if err != nil {
			return nil, fmt.Errorf("get workspace directory: %w", err)
		}

		// Try different log locations based on service configuration
		var logDirs []string

		if service.Slug != "" {
			// Project-specific service
			projectsDir, err := getProjectsDir()
			if err != nil {
				return nil, fmt.Errorf("get projects directory: %w", err)
			}
			logDirs = append(logDirs, filepath.Join(projectsDir, service.Slug, "logs"))
		}

		// Version-specific logs
		if service.Slug != "" {
			phpVersionDir := filepath.Join(wsDir, "php", service.Slug)
			logDirs = append(logDirs, filepath.Join(phpVersionDir, "logs"))
		}

		// Common PHP-FPM log files
		potentialLogs := []string{
			"php-fpm.log",
			"php-fpm-error.log",
			"error.log",
			"slow.log",
		}

		for _, logDir := range logDirs {
			for _, logName := range potentialLogs {
				logPath := filepath.Join(logDir, logName)
				if _, err := os.Stat(logPath); err == nil {
					logFiles = append(logFiles, logPath)
				}
			}
		}
	}

	return logFiles, nil
}

// readLogFile reads and displays content from a log file
func (lc *LogCollector) readLogFile(logFile string, service services.Service, options *LogCollectionOptions) error {
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Read all lines first (for simple filtering)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read log file: %w", err)
	}

	// Apply filters and display
	filteredLines := lc.filterLogLines(lines, options)
	displayLines := lc.getLastLines(filteredLines, options.Lines)

	if len(displayLines) == 0 {
		if !options.Quiet {
			lc.logger.Info(fmt.Sprintf("no matching log entries in %s", logFile))
		}
		return nil
	}

	// Display context if requested
	if options.ShowContext && !options.Quiet {
		fmt.Printf("📄 %s (%d lines)\n", logFile, len(displayLines))
		if len(displayLines) < len(filteredLines) {
			fmt.Printf("   Showing last %d of %d matching lines\n", len(displayLines), len(filteredLines))
		}
		fmt.Println()
	}

	// Print the lines
	for _, line := range displayLines {
		if options.Quiet {
			fmt.Println(line)
		} else {
			lc.printLogLine(line, service, options)
		}
	}

	return nil
}

// filterLogLines applies time and level filters to log lines
func (lc *LogCollector) filterLogLines(lines []string, options *LogCollectionOptions) []string {
	var filtered []string

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Level filter
		if options.Level != "" {
			lowerLine := strings.ToLower(line)
			if !strings.Contains(lowerLine, options.Level) {
				continue
			}
		}

		// Time filters (simplified - could be enhanced with proper time parsing)
		if options.Since != "" || options.Until != "" {
			// Basic time-based filtering - this is a simplified implementation
			// A full implementation would parse timestamps from various log formats
		}

		filtered = append(filtered, line)
	}

	return filtered
}

// getLastLines returns the last N lines from a slice
func (lc *LogCollector) getLastLines(lines []string, count int) []string {
	if count <= 0 || len(lines) <= count {
		return lines
	}

	return lines[len(lines)-count:]
}

// printLogLine prints a log line with optional formatting
func (lc *LogCollector) printLogLine(line string, service services.Service, options *LogCollectionOptions) {
	if options.Quiet {
		fmt.Println(line)
		return
	}

	// Add service prefix if not already present
	if !strings.Contains(line, service.Name) && options.Verbose {
		fmt.Printf("[%s] %s\n", service.Name, line)
	} else {
		fmt.Println(line)
	}
}

// tailLogFiles follows log files in real-time
func (lc *LogCollector) tailLogFiles(logFiles []string, service services.Service, options *LogCollectionOptions) error {
	if len(logFiles) == 0 {
		return fmt.Errorf("no log files to tail")
	}

	// For simplicity, tail the first log file
	// A full implementation would use a proper tailing library that can handle multiple files
	logFile := logFiles[0]

	lc.logger.Info(fmt.Sprintf("following log file: %s", logFile))
	lc.logger.Info("press Ctrl+C to stop")

	cmd := exec.Command("tail", "-f", logFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// listAvailableServices lists all available services that have logs
func (lc *LogCollector) listAvailableServices() error {
	lc.logger.PrintSection("Available Services for Log Viewing")

	var allServices []services.Service

	// Add global services
	globalServices := lc.serviceManager.ListGlobalServices()
	allServices = append(allServices, globalServices...)

	// Add project services
	projects, err := findAllLinkedProjects()
	if err != nil {
		return lc.logger.Error("find linked projects", err.Error())
	}

	for _, projectSlug := range projects {
		projectServices, err := lc.serviceManager.ListProjectServices(projectSlug)
		if err != nil {
			continue
		}
		allServices = append(allServices, projectServices...)
	}

	if len(allServices) == 0 {
		lc.logger.Info("no Chauffeur services found")
		lc.logger.Info("use 'chauf start' to start services")
		return nil
	}

	// Group services by type and show log availability
	serviceTable := [][]string{}

	for _, service := range allServices {
		logFiles, err := lc.getServiceLogFiles(service)
		if err != nil {
			continue
		}

		logStatus := "No logs"
		if len(logFiles) > 0 {
			logStatus = fmt.Sprintf("%d log files", len(logFiles))
		}

		serviceType := "Global"
		if service.Type == services.ServiceTypeProject {
			serviceType = "Project"
		}

		projectDisplay := service.Slug
		if projectDisplay == "" {
			projectDisplay = "—"
		}

		row := []string{
			service.Name,
			serviceType,
			projectDisplay,
			logStatus,
		}
		serviceTable = append(serviceTable, row)
	}

	headers := []string{"SERVICE", "TYPE", "PROJECT", "LOGS"}
	lc.logger.PrintServiceTable(headers, serviceTable)

	lc.logger.Info("")
	lc.logger.Info("Usage:")
	lc.logger.Info("  chauf logs <service-name>     # View logs for a specific service")
	lc.logger.Info("  chauf logs nginx              # View nginx logs")
	lc.logger.Info("  chauf logs php-fpm            # View PHP-FPM logs")
	lc.logger.Info("")
	lc.logger.Info("Options:")
	lc.logger.Info("  --follow, -f                 # Follow log output in real-time")
	lc.logger.Info("  --lines <n>                  # Show last N lines (default: 100)")
	lc.logger.Info("  --level <level>              # Filter by log level")
	lc.logger.Info("  --context, -c                # Show file context information")

	return nil
}

// printLogsUsage renders CLI help for the logs command
func printLogsUsage() {
	fmt.Println(`Usage: chauf logs [service-name] [options]

View and follow logs from Chauffeur services.

Arguments:
  service-name         Name of the service to view logs for (nginx, php-fpm, etc.)
                       If omitted, lists available services.

Options:
  --follow, -f         Follow log output in real-time (like tail -f)
  --lines <n>, -n      Show last N lines (default: 100)
  --since <time>       Show logs since specified time
  --until <time>       Show logs until specified time
  --level <level>      Filter logs by level (error, warning, info, debug)
  --context, -c        Show file context and metadata
  --verbose, -v        Show verbose output with service prefixes
  --quiet, -q          Show only log lines without additional formatting
  -h, --help          Show this message

Examples:
  chauf logs                    # List available services
  chauf logs nginx              # View nginx access/error logs
  chauf logs php-fpm            # View PHP-FPM logs
  chauf logs nginx --follow     # Follow nginx logs in real-time
  chauf logs php-fpm -n 50      # Show last 50 lines of PHP-FPM logs
  chauf logs nginx --level error # Show only nginx error logs

Log Locations:
  Nginx:     ~/.chauffeur/nginx/logs/
  PHP-FPM:   ~/.chauffeur/php/*/logs/ or ~/.chauffeur/projects/*/logs/
`)
}

// selectService provides interactive service selection when multiple services match
func (lc *LogCollector) selectService(matchingServices []services.Service, options *LogCollectionOptions) (*services.Service, error) {
	if options.Quiet {
		// In quiet mode, return nil to show all services
		return nil, nil
	}

	fmt.Printf("\nMultiple services found matching '%s':\n\n", options.ServiceName)

	for i, service := range matchingServices {
		status := "🔴 Stopped"
		// Check if service is running by checking PID file
		if service.PIDFile != "" {
			if _, err := os.Stat(service.PIDFile); err == nil {
				status = "🟢 Running"
			}
		}
		fmt.Printf("  [%d] %s - %s\n", i+1, service.Name, status)
	}

	fmt.Printf("\n[ logs ] Select service to view logs (1-%d), or 'a' for all services: ", len(matchingServices))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "a" || input == "all" {
		// User wants to see all services
		return nil, nil
	}

	// Try to parse as number
	choice, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid choice: %s", input)
	}

	if choice < 1 || choice > len(matchingServices) {
		return nil, fmt.Errorf("invalid choice: %d (must be 1-%d)", choice, len(matchingServices))
	}

	return &matchingServices[choice-1], nil
}

// Helper function to get projects directory (similar to service manager)
func getProjectsDir() (string, error) {
	wsDir, err := workspace.Dir()
	if err != nil {
		return "", fmt.Errorf("get workspace directory: %w", err)
	}

	// Try to load config to get projects directory
	// For now, use default
	return filepath.Join(wsDir, "projects"), nil
}