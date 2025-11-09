package lib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
)

// PortValidator handles port validation for Chauffeur commands
type PortValidator struct {
	portManager *PortManager
	config      config.Config
	logger      *Logger
}

// NewPortValidator creates a new port validator instance
func NewPortValidator(cfg config.Config) (*PortValidator, error) {
	// Validate port range configuration
	pm := NewPortManager(cfg.Ports.StartRange, cfg.Ports.EndRange, cfg.Ports.ConflictResolution)
	if err := pm.ValidatePortRange(); err != nil {
		return nil, fmt.Errorf("invalid port range: %w", err)
	}
	
	logger := NewCommandLogger("validate")
	
	return &PortValidator{
		portManager: pm,
		config:      cfg,
		logger:      logger,
	}, nil
}

// ValidateAllPorts validates all configured ports for conflicts
func (pv *PortValidator) ValidateAllPorts() error {
	pv.logger.Info("Validating all configured ports...")
	
	// Collect all ports to validate
	ports := map[string]int{
		"caddy-http":  pv.config.Caddy.HTTPPort,
		"caddy-https": pv.config.Caddy.HTTPSPort,
	}
	
	if pv.config.Nginx.HTTPPort > 0 {
		ports["nginx-http"] = pv.config.Nginx.HTTPPort
	}
	if pv.config.Nginx.HTTPSPort > 0 {
		ports["nginx-https"] = pv.config.Nginx.HTTPSPort
	}
	
	// Check for conflicts
	conflicts := pv.portManager.ValidatePortConfiguration(ports)
	
	if len(conflicts) == 0 {
		pv.logger.Success("All ports are available", "")
		return nil
	}
	
	// Handle conflicts based on resolution strategy
	return pv.handlePortConflicts(conflicts)
}

// ValidateSpecificPorts validates specific ports for a service
func (pv *PortValidator) ValidateSpecificPorts(service string, ports []int) error {
	pv.logger.Info(fmt.Sprintf("Validating ports for %s", service))
	
	// Build port map for validation
	portMap := make(map[string]int)
	for i, port := range ports {
		portMap[fmt.Sprintf("%s-%d", service, i)] = port
	}
	
	// Check for conflicts
	conflicts := pv.portManager.ValidatePortConfiguration(portMap)
	
	if len(conflicts) == 0 {
		pv.logger.Success("All ports are available", "")
		return nil
	}
	
	// Handle conflicts based on resolution strategy
	return pv.handlePortConflicts(conflicts)
}

// handlePortConflicts processes port conflicts according to configured strategy
func (pv *PortValidator) handlePortConflicts(conflicts []PortConflict) error {
	switch pv.config.Ports.ConflictResolution {
	case "fail":
		return pv.failOnConflicts(conflicts)
	case "auto":
		return pv.autoResolveConflicts(conflicts)
	case "prompt":
		return pv.promptForResolution(conflicts)
	default:
		return pv.failOnConflicts(conflicts)
	}
}

// failOnConflicts fails immediately when conflicts are detected
func (pv *PortValidator) failOnConflicts(conflicts []PortConflict) error {
	pv.logger.Error("Port conflicts detected", "")
	for _, conflict := range conflicts {
		pv.logger.Error(fmt.Sprintf("  - Port %d (%s)", conflict.Port, conflict.Service), conflict.UsedBy)
	}
	return fmt.Errorf("port conflicts detected - see above for details")
}

// autoResolveConflicts automatically resolves conflicts by finding available ports
func (pv *PortValidator) autoResolveConflicts(conflicts []PortConflict) error {
	pv.logger.Info("Auto-resolving port conflicts...")
	
	resolved := make(map[string]int)
	
	for _, conflict := range conflicts {
		newPort, err := pv.portManager.AutoResolvePortConflict(conflict.Service, conflict.Port)
		if err != nil {
			return fmt.Errorf("failed to auto-resolve conflict for %s: %w", conflict.Service, err)
		}
		resolved[conflict.Service] = newPort
	}
	
	// Update configuration with resolved ports
	return pv.updateConfigWithResolvedPorts(resolved)
}

// promptForResolution prompts user to resolve conflicts
func (pv *PortValidator) promptForResolution(conflicts []PortConflict) error {
	pv.logger.Warn("Port conflicts detected", "Please resolve them to continue")
	
	resolved := make(map[string]int)
	
	for _, conflict := range conflicts {
		newPort, err := pv.portManager.GetPortFromPrompt(conflict.Service, conflict.Port, conflict.Suggestions)
		if err != nil {
			return fmt.Errorf("failed to resolve conflict for %s: %w", conflict.Service, err)
		}
		resolved[conflict.Service] = newPort
	}
	
	// Update configuration with resolved ports
	return pv.updateConfigWithResolvedPorts(resolved)
}

// updateConfigWithResolvedPorts updates the configuration with new ports
func (pv *PortValidator) updateConfigWithResolvedPorts(resolved map[string]int) error {
	pv.logger.Info("Updating configuration with resolved ports...")
	
	// Update config based on service names
	for service, port := range resolved {
		switch strings.ToLower(service) {
		case "caddy", "caddy-http":
			pv.config.Caddy.HTTPPort = port
		case "caddy-https":
			pv.config.Caddy.HTTPSPort = port
		case "nginx", "nginx-http":
			pv.config.Nginx.HTTPPort = port
		case "nginx-https":
			pv.config.Nginx.HTTPSPort = port
		case "php-fpm", "phpfpm":
			// PHP-FPM port would be handled in a separate config section
			pv.logger.Info(fmt.Sprintf("PHP-FPM port updated: port %d", port))
		}
	}
	
	// Save the updated configuration
	if err := config.Save(pv.config); err != nil {
		return fmt.Errorf("failed to save updated configuration: %w", err)
	}
	
	pv.logger.Success("Configuration updated", "ports resolved")
	return nil
}

// GetSafePort returns a safe port configuration for a service
func (pv *PortValidator) GetSafePort(service string, preferredPort int) (int, error) {
	if !pv.portManager.IsPortAvailable(preferredPort) {
		// Port is available
		return preferredPort, nil
	}
	
	// Port is in conflict, resolve based on strategy
	switch pv.config.Ports.ConflictResolution {
	case "fail":
		return 0, fmt.Errorf("port %d is already in use for %s", preferredPort, service)
		
	case "auto":
		available := pv.portManager.FindAvailablePorts(1, preferredPort)
		if len(available) == 0 {
			return 0, fmt.Errorf("no available ports found for %s", service)
		}
		return available[0], nil
		
	case "prompt":
		suggestions := pv.portManager.FindAvailablePorts(3, preferredPort)
		return pv.portManager.GetPortFromPrompt(service, preferredPort, suggestions)
		
	default:
		return 0, fmt.Errorf("unknown conflict resolution strategy: %s", pv.config.Ports.ConflictResolution)
	}
}

// ValidateEnvironmentPortConfig validates port settings from environment variables
func (pv *PortValidator) ValidateEnvironmentPortConfig() error {
	pv.logger.Info("Validating environment port configuration...")
	
	envPorts := ReadPortConfigFromEnv()
	if len(envPorts) == 0 {
		pv.logger.Info("No environment port configuration found")
		return nil
	}
	
	// Convert environment port map to validation format
	portMap := make(map[string]int)
	for service, port := range envPorts {
		portMap[fmt.Sprintf("env-%s", service)] = port
	}
	
	// Check for conflicts
	conflicts := pv.portManager.ValidatePortConfiguration(portMap)
	
	if len(conflicts) == 0 {
		pv.logger.Success("Environment port configuration is valid", "")
		return nil
	}
	
	pv.logger.Warn("Environment port conflicts detected", "")
	for _, conflict := range conflicts {
		pv.logger.Warn(fmt.Sprintf("  - Port %d (%s)", conflict.Port, conflict.Service), conflict.UsedBy)
	}
	
	return fmt.Errorf("environment port conflicts - update environment variables or configuration")
}

// ShowPortConfiguration displays current port configuration
func (pv *PortValidator) ShowPortConfiguration() {
	pv.logger.Info("Current port configuration:")
	pv.logger.Info(fmt.Sprintf("  Caddy HTTP: %d", pv.config.Caddy.HTTPPort))
	pv.logger.Info(fmt.Sprintf("  Caddy HTTPS: %d", pv.config.Caddy.HTTPSPort))
	if pv.config.Nginx.HTTPPort > 0 {
		pv.logger.Info(fmt.Sprintf("  Nginx HTTP: %d", pv.config.Nginx.HTTPPort))
	}
	if pv.config.Nginx.HTTPSPort > 0 {
		pv.logger.Info(fmt.Sprintf("  Nginx HTTPS: %d", pv.config.Nginx.HTTPSPort))
	}
	
	pv.logger.Info(fmt.Sprintf("  Port Range: %d-%d", pv.config.Ports.StartRange, pv.config.Ports.EndRange))
	pv.logger.Info(fmt.Sprintf("  Conflict Resolution: %s", pv.config.Ports.ConflictResolution))
	
	// Check if ports are actually available
	ports := map[string]int{
		"caddy-http":  pv.config.Caddy.HTTPPort,
		"caddy-https": pv.config.Caddy.HTTPSPort,
	}
	if pv.config.Nginx.HTTPPort > 0 {
		ports["nginx-http"] = pv.config.Nginx.HTTPPort
	}
	if pv.config.Nginx.HTTPSPort > 0 {
		ports["nginx-https"] = pv.config.Nginx.HTTPSPort
	}
	
	pv.logger.Info("Port availability:")
	for service, port := range ports {
		if pv.portManager.IsPortAvailable(port) {
			pv.logger.Warn(fmt.Sprintf("  %s: Port %d is IN USE", service, port), "")
		} else {
			pv.logger.Info(fmt.Sprintf("  %s: Port %d is available", service, port))
		}
	}
}

// SetPortFromCommand sets a port from a command line flag
func (pv *PortValidator) SetPortFromCommand(service, portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number for %s: %s", service, portStr)
	}
	
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	
	// Validate port is available
	if pv.portManager.IsPortAvailable(port) {
		pv.logger.Error(fmt.Sprintf("Port %d is already in use", port), "")
		usedBy := pv.portManager.GetPortUsageDetails(port)
		if usedBy != "" {
			pv.logger.Error(fmt.Sprintf("Used by: %s", usedBy), "")
		}
		return 0, fmt.Errorf("port %d is not available", port)
	}
	
	pv.logger.Success(fmt.Sprintf("Port %d is available for %s", port, service), "")
	return port, nil
}
