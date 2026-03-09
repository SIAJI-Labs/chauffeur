package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FPMService manages a single PHP-FPM pool (shared or dedicated).
type FPMService struct {
	label      string // display label e.g. "8.3 (shared)" or "my-app (dedicated)"
	binaryPath string
	configPath string
	pidPath    string
	sockPath   string
}

// NewSharedFPM returns a FPMService for the shared pool of a given PHP version.
func NewSharedFPM(root, version string) *FPMService {
	phpDir := filepath.Join(root, "php", version)
	return &FPMService{
		label:      version + " (shared)",
		binaryPath: filepath.Join(phpDir, "sbin", "php-fpm"),
		configPath: filepath.Join(phpDir, "etc", "php-fpm.conf"),
		pidPath:    filepath.Join(phpDir, "runtime", "php-fpm", "php-fpm.pid"),
		sockPath:   filepath.Join(phpDir, "runtime", "php-fpm", "php-fpm.sock"),
	}
}

// NewDedicatedFPM returns a FPMService for a per-project dedicated pool.
func NewDedicatedFPM(root, slug, phpVersion, sockPath string) *FPMService {
	phpDir := filepath.Join(root, "php", phpVersion)
	projectDir := filepath.Join(root, "projects", slug)
	return &FPMService{
		label:      slug + " (dedicated)",
		binaryPath: filepath.Join(phpDir, "sbin", "php-fpm"),
		configPath: filepath.Join(projectDir, "php-fpm.conf"),
		pidPath:    filepath.Join(projectDir, "php-fpm.pid"),
		sockPath:   sockPath,
	}
}

// Label returns the human-readable service name for display.
func (f *FPMService) Label() string { return f.label }

// SockPath returns the unix socket path for this FPM pool.
func (f *FPMService) SockPath() string { return f.sockPath }

// IsRunning returns true if PHP-FPM is running and the process is alive.
func (f *FPMService) IsRunning() bool {
	pid, err := readPIDFile(f.pidPath)
	if err != nil || pid <= 0 {
		return false
	}
	return processAlive(pid)
}

// PID returns the current PHP-FPM master PID, or 0 if not running.
func (f *FPMService) PID() int {
	pid, err := readPIDFile(f.pidPath)
	if err != nil {
		return 0
	}
	return pid
}

// Start launches PHP-FPM. Idempotent if already running.
func (f *FPMService) Start() error {
	if f.IsRunning() {
		return nil
	}
	if _, err := os.Stat(f.binaryPath); err != nil {
		return fmt.Errorf("php-fpm binary not found: %s", f.binaryPath)
	}
	if _, err := os.Stat(f.configPath); err != nil {
		return fmt.Errorf("php-fpm config not found: %s", f.configPath)
	}

	// Ensure runtime directory exists
	if err := os.MkdirAll(filepath.Dir(f.pidPath), 0755); err != nil {
		return fmt.Errorf("create runtime dir: %w", err)
	}

	cmd := exec.Command(f.binaryPath, "--fpm-config", f.configPath, "--daemonize")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("php-fpm start (%s): %w\n%s", f.label, err, strings.TrimSpace(string(out)))
	}
	return waitForPID(f.pidPath, 5*time.Second)
}

// Stop sends SIGTERM for graceful shutdown and waits up to timeout.
func (f *FPMService) Stop(timeout time.Duration) error {
	pid, err := readPIDFile(f.pidPath)
	if err != nil {
		return nil // not running
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return nil // already gone
	}
	return waitForExit(pid, timeout)
}

// Reload sends SIGUSR2 for a graceful worker reload.
// If not running, starts the pool instead.
func (f *FPMService) Reload() error {
	if !f.IsRunning() {
		return f.Start()
	}
	pid, err := readPIDFile(f.pidPath)
	if err != nil {
		return f.Start()
	}
	return syscall.Kill(pid, syscall.SIGUSR2)
}

// Uptime returns how long this PHP-FPM pool has been running.
func (f *FPMService) Uptime() time.Duration {
	pid := f.PID()
	if pid <= 0 {
		return 0
	}
	return processUptime(pid)
}

// MemoryMB returns the master process RSS in MB (best-effort, 0 on error).
func (f *FPMService) MemoryMB() int {
	pid := f.PID()
	if pid <= 0 {
		return 0
	}
	return processMemoryMB(pid)
}
