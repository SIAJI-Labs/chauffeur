package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/siaji/chauffeur/cli/lib" // Add this import
)



type portForwardingState struct {
	HTTP  int `json:"http_port"`
	HTTPS int `json:"https_port"`
}

func portForwardingStatePath(workspaceDir string) string {
	return filepath.Join(workspaceDir, "system", "port-forwarding.json")
}

// EnsurePortForwarding configures iptables rules that forward privileged ports (80/443)
// to Chauffeur's user-space nginx listeners.
func EnsurePortForwarding(workspaceDir string, httpPort, httpsPort int) error {
	if workspaceDir == "" {
		return fmt.Errorf("workspace directory is required for port forwarding state")
	}

	// Check if running as root or if sudo is available
	if !isSudoAvailable() {
		return fmt.Errorf("port forwarding requires sudo privileges. Current ports: HTTP=%d, HTTPS=%d. Consider using non-privileged ports (above 1024) to avoid port forwarding.", httpPort, httpsPort)
	}

	statePath := portForwardingStatePath(workspaceDir)
	state, err := loadPortForwardingState(statePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load port forwarding state: %w", err)
	}

	desiredHTTP := normalizedForwardPort(httpPort, 80)
	desiredHTTPS := normalizedForwardPort(httpsPort, 443)

	// Check if we need to make changes
	needsUpdate := state.HTTP != desiredHTTP || state.HTTPS != desiredHTTPS

	if !needsUpdate && desiredHTTP == 0 && desiredHTTPS == 0 {
		// No port forwarding needed and no existing rules
		return nil
	}

	// Remove existing rules if they don't match desired state
	if state.HTTP != 0 && state.HTTP != desiredHTTP {
		if err := removeRedirectRule(80, state.HTTP); err != nil {
			return fmt.Errorf("failed to remove old HTTP redirect rule (80→%d): %w", state.HTTP, err)
		}
		state.HTTP = 0
	}

	if state.HTTPS != 0 && state.HTTPS != desiredHTTPS {
		if err := removeRedirectRule(443, state.HTTPS); err != nil {
			return fmt.Errorf("failed to remove old HTTPS redirect rule (443→%d): %w", state.HTTPS, err)
		}
		state.HTTPS = 0
	}

	// Add new rules if needed
	if desiredHTTP != 0 {
		if err := ensureRedirectRule(80, desiredHTTP); err != nil {
			return fmt.Errorf("failed to add HTTP redirect rule (80→%d): %w", desiredHTTP, err)
		}
		state.HTTP = desiredHTTP
	}

	if desiredHTTPS != 0 {
		if err := ensureRedirectRule(443, desiredHTTPS); err != nil {
			return fmt.Errorf("failed to add HTTPS redirect rule (443→%d): %w", desiredHTTPS, err)
		}
		state.HTTPS = desiredHTTPS
	}

	return savePortForwardingState(statePath, state)
}

// CleanupPortForwarding removes any iptables rules previously managed by Chauffeur.
func CleanupPortForwarding(workspaceDir string) error {
	if workspaceDir == "" {
		return nil
	}

	statePath := portForwardingStatePath(workspaceDir)
	state, err := loadPortForwardingState(statePath)
	if err != nil {
		return nil
	}

	if state.HTTP != 0 {
		_ = removeRedirectRule(80, state.HTTP)
	}
	if state.HTTPS != 0 {
		_ = removeRedirectRule(443, state.HTTPS)
	}

	_ = os.Remove(statePath)
	return nil
}

func normalizedForwardPort(actual, privileged int) int {
	if actual <= 0 || actual == privileged {
		return 0
	}
	return actual
}

func loadPortForwardingState(path string) (portForwardingState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return portForwardingState{}, err
	}

	var state portForwardingState
	if err := json.Unmarshal(data, &state); err != nil {
		return portForwardingState{}, err
	}
	return state, nil
}

func savePortForwardingState(path string, state portForwardingState) error {
	if state.HTTP == 0 && state.HTTPS == 0 {
		return os.Remove(path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create port forwarding state dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal port forwarding state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write port forwarding state: %w", err)
	}

	return nil
}

func ensureRedirectRule(fromPort, toPort int) error {
	exists, err := redirectRuleExists(fromPort, toPort)
	if err != nil {
		return fmt.Errorf("failed to check if redirect rule exists: %w", err)
	}
	if exists {
		return nil
	}

	// Validate ports
	if fromPort <= 0 || fromPort > 65535 || toPort <= 0 || toPort > 65535 {
		return fmt.Errorf("invalid ports: from=%d, to=%d", fromPort, toPort)
	}

	cmd := lib.CommandExecutor("sudo", "iptables", "-t", "nat", "-A", "OUTPUT",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(fromPort),
		"-j", "REDIRECT",
		"--to-ports", strconv.Itoa(toPort),
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("iptables command failed: %w", err)
	}

	return nil
}

func removeRedirectRule(fromPort, toPort int) error {
	cmd := lib.CommandExecutor("sudo", "iptables", "-t", "nat", "-D", "OUTPUT",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(fromPort),
		"-j", "REDIRECT",
		"--to-ports", strconv.Itoa(toPort),
	)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil
		}
		return err
	}
	return nil
}

func redirectRuleExists(fromPort, toPort int) (bool, error) {
	cmd := lib.CommandExecutor("sudo", "iptables", "-t", "nat", "-C", "OUTPUT",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(fromPort),
		"-j", "REDIRECT",
		"--to-ports", strconv.Itoa(toPort),
	)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, fmt.Errorf("iptables check command failed: %w", err)
	}
	return true, nil
}

// isSudoAvailable checks if sudo privileges are available for port forwarding
func isSudoAvailable() bool {
	// First check if sudo command exists
	if _, err := exec.LookPath("sudo"); err != nil {
		return false
	}

	// Check if we can run a simple sudo command (with -n to avoid password prompt)
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		// If sudo requires password, we'll need to handle it differently
		// For now, assume sudo is available but may need password
		return true
	}

	return true
}

// GetPortForwardingStatus returns the current port forwarding configuration
func GetPortForwardingStatus(workspaceDir string) (httpPort, httpsPort int, err error) {
	if workspaceDir == "" {
		return 0, 0, fmt.Errorf("workspace directory is required")
	}

	statePath := portForwardingStatePath(workspaceDir)
	state, err := loadPortForwardingState(statePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to load port forwarding state: %w", err)
	}

	return state.HTTP, state.HTTPS, nil
}

// ValidatePortConfiguration checks if the port configuration is valid and makes recommendations
func ValidatePortConfiguration(httpPort, httpsPort int) error {
	var warnings []string

	if httpPort == 80 {
		warnings = append(warnings, "HTTP port 80 requires sudo privileges for port forwarding")
	} else if httpPort < 1024 {
		warnings = append(warnings, fmt.Sprintf("HTTP port %d requires sudo privileges (privileged port < 1024)", httpPort))
	} else if httpPort < 8080 {
		warnings = append(warnings, fmt.Sprintf("HTTP port %d may conflict with system services, consider using 8080+", httpPort))
	}

	if httpsPort == 443 {
		warnings = append(warnings, "HTTPS port 443 requires sudo privileges for port forwarding")
	} else if httpsPort < 1024 {
		warnings = append(warnings, fmt.Sprintf("HTTPS port %d requires sudo privileges (privileged port < 1024)", httpsPort))
	} else if httpsPort < 8443 {
		warnings = append(warnings, fmt.Sprintf("HTTPS port %d may conflict with system services, consider using 8443+", httpsPort))
	}

	// Check for port conflicts with common system ports
	if httpPort == httpsPort && httpPort != 0 {
		return fmt.Errorf("HTTP and HTTPS ports cannot be the same (both set to %d)", httpPort)
	}

	// If there are warnings, we don't fail but the caller can choose to display them
	if len(warnings) > 0 {
		// For now, just log them - the caller can decide what to do
		fmt.Fprintf(os.Stderr, "Port configuration warnings:\n")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", warning)
		}
	}

	return nil
}
