package lib

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// PortConflict represents a port conflict information
type PortConflict struct {
	Port         int
	Service      string // "caddy", "nginx", "php-fpm"
	UsedBy       string // Description of what's using the port
	Suggestions  []int  // Alternative port suggestions
}

// PortManager handles port conflict detection and resolution
type PortManager struct {
	startRange         int
	endRange           int
	conflictResolution string // "prompt", "auto", "fail"
}

// NewPortManager creates a new port manager instance
func NewPortManager(startRange, endRange int, conflictResolution string) *PortManager {
	return &PortManager{
		startRange:         startRange,
		endRange:           endRange,
		conflictResolution: conflictResolution,
	}
}

// IsPortAvailable checks if a specific port is available
func (pm *PortManager) IsPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	
	// Try to connect to the port with a timeout
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false // Port is free (connection failed)
	}
	conn.Close()
	return true // Port is in use (connection succeeded)
}

// GetPortUsageDetails provides information about what's using a port
func (pm *PortManager) GetPortUsageDetails(port int) string {
	address := fmt.Sprintf(":%d", port)
	
	// Try to connect and get more details
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()
	
	// Try to read service information (basic implementation)
	// In a more sophisticated implementation, we could use system calls
	// to get process information from the port usage
	
	// For now, provide generic information
	switch port {
	case 80:
		return "system HTTP server (possibly apache/nginx)"
	case 443:
		return "system HTTPS server (possibly apache/nginx)"
	case 8080, 8081:
		return "web server (possibly another development server)"
	case 8443, 8444:
		return "secure web server"
	case 9000:
		return "PHP-FPM or other service"
	default:
		return fmt.Sprintf("unknown service on port %d", port)
	}
}

// FindAvailablePorts finds available ports within the configured range
func (pm *PortManager) FindAvailablePorts(count int, preferredPort int) []int {
	var available []int
	
	// Try preferred port first if specified
	if preferredPort > 0 && !pm.IsPortAvailable(preferredPort) {
		available = append(available, preferredPort)
		if len(available) >= count {
			return available
		}
	}
	
	// Scan the range for available ports
	for port := pm.startRange; port <= pm.endRange; port++ {
		if !pm.IsPortAvailable(port) {
			// Skip if it's the preferred port we already tried
			if port == preferredPort {
				continue
			}
			available = append(available, port)
			if len(available) >= count {
				break
			}
		}
	}
	
	return available
}

// ValidatePortConfiguration checks all service ports for conflicts
func (pm *PortManager) ValidatePortConfiguration(ports map[string]int) []PortConflict {
	var conflicts []PortConflict
	
	for service, port := range ports {
		if pm.IsPortAvailable(port) {
			conflict := PortConflict{
				Port:    port,
				Service: service,
				UsedBy:  pm.GetPortUsageDetails(port),
			}
			
			// Find alternative suggestions
			conflict.Suggestions = pm.FindAvailablePorts(3, port)
			
			conflicts = append(conflicts, conflict)
		}
	}
	
	return conflicts
}

// GetPortFromPrompt prompts user to select an available port
func (pm *PortManager) GetPortFromPrompt(service string, conflictPort int, suggestions []int) (int, error) {
	logger := NewCommandLogger("ports")
	
	logger.Warn("Port conflict detected", fmt.Sprintf("Port %d is already in use", conflictPort))
	logger.Info(fmt.Sprintf("Service: %s", service))
	if usedBy := pm.GetPortUsageDetails(conflictPort); usedBy != "" {
		logger.Info(fmt.Sprintf("Port %d used by: %s", conflictPort, usedBy))
	}
	
	logger.Info("Available options:")
	logger.Info(fmt.Sprintf("  [s] Suggest alternative ports: %v", suggestions))
	logger.Info("  [c] Choose custom port")
	logger.Info("  [a] Abort operation")
	
	fmt.Print("\n[ ports ] Your choice [s/c/a]: ")
	var response string
	fmt.Scanln(&response)
	
	switch strings.ToLower(response) {
	case "s", "suggest":
		if len(suggestions) == 0 {
			suggestions = pm.FindAvailablePorts(3, 0)
		}
		
		if len(suggestions) == 0 {
			return 0, fmt.Errorf("no available ports found in range %d-%d", pm.startRange, pm.endRange)
		}
		
		logger.Info("Suggested ports:")
		for i, port := range suggestions {
			logger.Info(fmt.Sprintf("  [%d] Port %d", i+1, port))
		}
		
		fmt.Printf("[ ports ] Choose port number (1-%d): ", len(suggestions))
		var choice string
		fmt.Scanln(&choice)
		
		choiceIndex, err := strconv.Atoi(choice)
		if err != nil || choiceIndex < 1 || choiceIndex > len(suggestions) {
			return 0, fmt.Errorf("invalid choice: %s", choice)
		}
		
		selectedPort := suggestions[choiceIndex-1]
		logger.Success("Port selected", fmt.Sprintf("%d", selectedPort))
		return selectedPort, nil
		
	case "c", "custom":
		logger.Info("Enter custom port number:")
		fmt.Print("[ ports ] Port: ")
		var customPort string
		fmt.Scanln(&customPort)
		
		port, err := strconv.Atoi(customPort)
		if err != nil {
			return 0, fmt.Errorf("invalid port number: %s", customPort)
		}
		
		if port < 1 || port > 65535 {
			return 0, fmt.Errorf("port must be between 1 and 65535")
		}
		
		if pm.IsPortAvailable(port) {
			return 0, fmt.Errorf("port %d is already in use", port)
		}
		
		logger.Success("Custom port selected", fmt.Sprintf("%d", port))
		return port, nil
		
	case "a", "abort":
		return 0, fmt.Errorf("operation aborted by user")
		
	default:
		return 0, fmt.Errorf("invalid choice: %s", response)
	}
}

// AutoResolvePortConflict automatically selects an available port
func (pm *PortManager) AutoResolvePortConflict(service string, preferredPort int) (int, error) {
	logger := NewCommandLogger("ports")
	
	logger.Info(fmt.Sprintf("Auto-resolving port conflict for %s", service))
	if preferredPort > 0 {
		logger.Info(fmt.Sprintf("Preferred port %d is in use, finding alternative...", preferredPort))
	}
	
	available := pm.FindAvailablePorts(1, 0)
	if len(available) == 0 {
		return 0, fmt.Errorf("no available ports found in range %d-%d", pm.startRange, pm.endRange)
	}
	
	selectedPort := available[0]
	logger.Success("Auto-resolved port", fmt.Sprintf("%d for %s", selectedPort, service))
	return selectedPort, nil
}

// ValidatePortRange checks if the configured port range is valid
func (pm *PortManager) ValidatePortRange() error {
	if pm.startRange < 1 || pm.startRange > 65535 {
		return fmt.Errorf("start range must be between 1 and 65535")
	}
	
	if pm.endRange < 1 || pm.endRange > 65535 {
		return fmt.Errorf("end range must be between 1 and 65535")
	}
	
	if pm.startRange >= pm.endRange {
		return fmt.Errorf("start range must be less than end range")
	}
	
	// Check if range is too large (performance concern)
	if pm.endRange-pm.startRange > 1000 {
		return fmt.Errorf("port range too large (max 1000 ports)")
	}
	
	return nil
}

// GetRecommendedPort returns a recommended port based on service type
func (pm *PortManager) GetRecommendedPort(service string) int {
	switch strings.ToLower(service) {
	case "caddy":
		return 8080
	case "nginx":
		return 8081
	case "php-fpm", "phpfpm":
		return 9000
	case "caddy-https":
		return 8443
	case "nginx-https":
		return 8444
	default:
		return pm.startRange
	}
}

// WritePortConfigToEnv writes port configuration to environment variables
// This can be used by CLI commands to communicate port config to subprocesses
func WritePortConfigToEnv(ports map[string]int) error {
	for service, port := range ports {
		envKey := fmt.Sprintf("CHAUF_%s_PORT", strings.ToUpper(service))
		envValue := strconv.Itoa(port)
		
		if err := os.Setenv(envKey, envValue); err != nil {
			return fmt.Errorf("failed to set %s=%s: %w", envKey, envValue, err)
		}
	}
	return nil
}

// ReadPortConfigFromEnv reads port configuration from environment variables
func ReadPortConfigFromEnv() map[string]int {
	ports := make(map[string]int)
	
	// Known service keys
	services := []string{"caddy", "nginx", "php-fpm", "caddy-https", "nginx-https"}
	
	for _, service := range services {
		envKey := fmt.Sprintf("CHAUF_%s_PORT", strings.ToUpper(strings.ReplaceAll(service, "-", "_")))
		if value := os.Getenv(envKey); value != "" {
			if port, err := strconv.Atoi(value); err == nil {
				ports[service] = port
			}
		}
	}
	
	return ports
}
