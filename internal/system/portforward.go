package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const systemServicePath = "/etc/systemd/system/chauffeur-portforward.service"

type portForwardState struct {
	HTTP  int `json:"http_port"`
	HTTPS int `json:"https_port"`
}

func portForwardStatePath(root string) string {
	return filepath.Join(root, "system", "port-forwarding.json")
}

// IsPortForwardingActive returns true if port forwarding is in place.
// Checks (in order): systemd service active → iptables rule → state file fallback.
func IsPortForwardingActive(root string, httpPort, httpsPort int) bool {
	// Preferred: systemd system service (no root needed to query)
	if isPortForwardSystemdActive() {
		return true
	}

	// Fallback: iptables check (requires passwordless sudo -n)
	if redirectRuleExists(80, httpPort) && redirectRuleExists(443, httpsPort) {
		return true
	}

	// Fallback: state file written after manual confirmation
	state, err := loadPortForwardState(root)
	if err != nil {
		return false
	}
	return state.HTTP == httpPort && state.HTTPS == httpsPort
}

// MarkPortForwardingConfigured persists the active port mapping to the state file.
func MarkPortForwardingConfigured(root string, httpPort, httpsPort int) error {
	return savePortForwardState(root, portForwardState{HTTP: httpPort, HTTPS: httpsPort})
}

// ClearPortForwardingState removes the state file (call on uninstall).
func ClearPortForwardingState(root string) {
	_ = os.Remove(portForwardStatePath(root))
}

// PortForwardingSystemdServiceContent returns the systemd system service unit
// that redirects ports 80/443 to the Chauffeur nginx ports via iptables NAT.
func PortForwardingSystemdServiceContent(httpPort, httpsPort int) string {
	return fmt.Sprintf(`[Unit]
Description=Chauffeur Port Forwarding (80->%d, 443->%d)
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 80 -j REDIRECT --to-ports %d
ExecStart=/usr/bin/iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports %d
ExecStop=/usr/bin/iptables -t nat -D OUTPUT -p tcp -d 127.0.0.1 --dport 80 -j REDIRECT --to-ports %d
ExecStop=/usr/bin/iptables -t nat -D OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports %d
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`, httpPort, httpsPort, httpPort, httpsPort, httpPort, httpsPort)
}

// PortForwardingSystemdCommands returns the shell commands to install the
// port-forwarding systemd system service and start it immediately.
// Requires sudo (runs once; survives reboots automatically after that).
func PortForwardingSystemdCommands(httpPort, httpsPort int) []string {
	content := PortForwardingSystemdServiceContent(httpPort, httpsPort)
	// Convert to printf-safe single-line: newlines → \n, escape single quotes
	printfContent := strings.ReplaceAll(content, "\n", "\\n")
	printfContent = strings.ReplaceAll(printfContent, "'", "'\\''")
	return []string{
		fmt.Sprintf("printf '%s' | sudo tee %s", printfContent, systemServicePath),
		"sudo systemctl daemon-reload",
		"sudo systemctl enable --now chauffeur-portforward.service",
	}
}

// PortForwardingRemoveCommands returns the commands to disable and remove the
// port-forwarding service.
func PortForwardingRemoveCommands() []string {
	return []string{
		"sudo systemctl disable --now chauffeur-portforward.service",
		"sudo rm -f " + systemServicePath,
		"sudo systemctl daemon-reload",
	}
}

// PortForwardingCommands returns one-time iptables commands (non-persistent).
// Prefer PortForwardingSystemdCommands for a setup that survives reboots.
func PortForwardingCommands(httpPort, httpsPort int) []string {
	return []string{
		fmt.Sprintf("sudo iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 80 -j REDIRECT --to-ports %d", httpPort),
		fmt.Sprintf("sudo iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports %d", httpsPort),
	}
}

// isPortForwardSystemdActive returns true when the chauffeur-portforward
// system service is currently active (does not require root).
func isPortForwardSystemdActive() bool {
	out, err := exec.Command("systemctl", "is-active", "chauffeur-portforward.service").Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
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
