package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	statePath := portForwardingStatePath(workspaceDir)
	state, _ := loadPortForwardingState(statePath)

	desiredHTTP := normalizedForwardPort(httpPort, 80)
	desiredHTTPS := normalizedForwardPort(httpsPort, 443)

	if state.HTTP != 0 && state.HTTP != desiredHTTP {
		_ = removeRedirectRule(80, state.HTTP)
		state.HTTP = 0
	}

	if state.HTTPS != 0 && state.HTTPS != desiredHTTPS {
		_ = removeRedirectRule(443, state.HTTPS)
		state.HTTPS = 0
	}

	if desiredHTTP != 0 {
		if err := ensureRedirectRule(80, desiredHTTP); err != nil {
			return err
		}
		state.HTTP = desiredHTTP
	}

	if desiredHTTPS != 0 {
		if err := ensureRedirectRule(443, desiredHTTPS); err != nil {
			return err
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
		return err
	}
	if exists {
		return nil
	}

	return exec.Command("sudo", "iptables", "-t", "nat", "-A", "OUTPUT",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(fromPort),
		"-j", "REDIRECT",
		"--to-ports", strconv.Itoa(toPort),
	).Run()
}

func removeRedirectRule(fromPort, toPort int) error {
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-D", "OUTPUT",
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
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-C", "OUTPUT",
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
		return false, err
	}
	return true, nil
}
