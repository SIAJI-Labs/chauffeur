package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type portForwardState struct {
	HTTP  int `json:"http_port"`
	HTTPS int `json:"https_port"`
}

func portForwardStatePath(root string) string {
	return filepath.Join(root, "system", "port-forwarding.json")
}

// IsPortForwardingActive returns true if iptables redirect rules for the given
// ports are currently in place. Falls back to the state file when iptables
// requires root (exit code 4 = permission denied).
func IsPortForwardingActive(root string, httpPort, httpsPort int) bool {
	httpOK := redirectRuleExists(80, httpPort)
	httpsOK := redirectRuleExists(443, httpsPort)
	if httpOK && httpsOK {
		return true
	}

	// Fall back to state file (written after user confirms setup).
	state, err := loadPortForwardState(root)
	if err != nil {
		return false
	}
	return state.HTTP == httpPort && state.HTTPS == httpsPort
}

// MarkPortForwardingConfigured persists the active port mapping to the state file.
// Call this after confirming the user has applied the iptables rules.
func MarkPortForwardingConfigured(root string, httpPort, httpsPort int) error {
	return savePortForwardState(root, portForwardState{HTTP: httpPort, HTTPS: httpsPort})
}

// ClearPortForwardingState removes the state file (call on uninstall or when
// rules are removed).
func ClearPortForwardingState(root string) {
	_ = os.Remove(portForwardStatePath(root))
}

// PortForwardingCommands returns the iptables commands needed to redirect
// standard ports (80/443) to the chauffeur nginx ports.
func PortForwardingCommands(httpPort, httpsPort int) []string {
	return []string{
		fmt.Sprintf("sudo iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 80 -j REDIRECT --to-ports %d", httpPort),
		fmt.Sprintf("sudo iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports %d", httpsPort),
	}
}

// PortForwardingRemoveCommands returns the iptables commands to remove the rules.
func PortForwardingRemoveCommands(httpPort, httpsPort int) []string {
	return []string{
		fmt.Sprintf("sudo iptables -t nat -D OUTPUT -p tcp -d 127.0.0.1 --dport 80 -j REDIRECT --to-ports %d", httpPort),
		fmt.Sprintf("sudo iptables -t nat -D OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports %d", httpsPort),
	}
}

// redirectRuleExists checks if an iptables OUTPUT redirect rule is in place.
// Returns false (not an error) when iptables requires root.
func redirectRuleExists(fromPort, toPort int) bool {
	cmd := exec.Command("sudo", "-n", "iptables", "-t", "nat", "-C", "OUTPUT",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(fromPort),
		"-j", "REDIRECT",
		"--to-ports", strconv.Itoa(toPort),
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func loadPortForwardState(root string) (portForwardState, error) {
	data, err := os.ReadFile(portForwardStatePath(root))
	if err != nil {
		return portForwardState{}, err
	}
	var s portForwardState
	return s, json.Unmarshal(data, &s)
}

func savePortForwardState(root string, s portForwardState) error {
	path := portForwardStatePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
