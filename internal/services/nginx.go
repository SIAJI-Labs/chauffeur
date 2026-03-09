package services

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// NginxService manages the nginx process lifecycle.
type NginxService struct {
	binaryPath string
	configPath string
	pidPath    string
}

// NewNginxService returns a NginxService wired to the given workspace root.
func NewNginxService(root string) *NginxService {
	return &NginxService{
		binaryPath: filepath.Join(root, "nginx", "sbin", "nginx"),
		configPath: filepath.Join(root, "nginx", "etc", "nginx.conf"),
		pidPath:    filepath.Join(root, "nginx", "logs", "nginx.pid"),
	}
}

// IsRunning returns true if nginx is running and the process is alive.
func (n *NginxService) IsRunning() bool {
	pid, err := readPIDFile(n.pidPath)
	if err != nil || pid <= 0 {
		return false
	}
	return processAlive(pid)
}

// PID returns the current nginx PID, or 0 if not running.
func (n *NginxService) PID() int {
	pid, err := readPIDFile(n.pidPath)
	if err != nil {
		return 0
	}
	return pid
}

// TestConfig validates the nginx configuration without applying it.
func (n *NginxService) TestConfig() error {
	out, err := exec.Command(n.binaryPath, "-t", "-c", n.configPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config invalid\n\n%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Start launches nginx. Idempotent if already running.
func (n *NginxService) Start() error {
	if n.IsRunning() {
		return nil
	}
	if err := n.TestConfig(); err != nil {
		return err
	}
	cmd := exec.Command(n.binaryPath, "-c", n.configPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx start: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	// Wait briefly for PID file to appear
	return waitForPID(n.pidPath, 3*time.Second)
}

// Stop sends SIGQUIT for graceful shutdown and waits up to timeout.
func (n *NginxService) Stop(timeout time.Duration) error {
	pid, err := readPIDFile(n.pidPath)
	if err != nil {
		return nil // not running
	}
	if err := syscall.Kill(pid, syscall.SIGQUIT); err != nil {
		return nil // already gone
	}
	return waitForExit(pid, timeout)
}

// Reload sends SIGHUP for a zero-downtime config reload.
// If nginx is not running, starts it instead.
func (n *NginxService) Reload() error {
	if !n.IsRunning() {
		return n.Start()
	}
	if err := n.TestConfig(); err != nil {
		return err
	}
	pid, err := readPIDFile(n.pidPath)
	if err != nil {
		return n.Start()
	}
	return syscall.Kill(pid, syscall.SIGHUP)
}

// Uptime returns how long nginx has been running.
func (n *NginxService) Uptime() time.Duration {
	pid := n.PID()
	if pid <= 0 {
		return 0
	}
	return processUptime(pid)
}

// MemoryMB returns nginx master process RSS in MB (best-effort, 0 on error).
func (n *NginxService) MemoryMB() int {
	pid := n.PID()
	if pid <= 0 {
		return 0
	}
	return processMemoryMB(pid)
}
