package projects

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ReloadNginx sends SIGHUP to the nginx process if it is running.
// Returns nil if nginx is not running (nothing to reload).
func ReloadNginx(workspaceRoot string) error {
	pidPath := filepath.Join(workspaceRoot, "nginx", "logs", "nginx.pid")
	pid, err := readPIDFile(pidPath)
	if err != nil || pid == 0 {
		return nil // not running
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return nil // process does not exist
	}
	return proc.Signal(syscall.SIGHUP)
}

// IsNginxRunning returns true if the nginx PID file exists and the process is alive.
func IsNginxRunning(workspaceRoot string) bool {
	pidPath := filepath.Join(workspaceRoot, "nginx", "logs", "nginx.pid")
	pid, err := readPIDFile(pidPath)
	if err != nil || pid == 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// RunMkcert invokes mkcert to generate a SAN certificate for the given domains.
// The cert and key are written to workspaceRoot/nginx/certs/<primaryDomain>.{crt,key}.
func RunMkcert(workspaceRoot, primaryDomain string, domains []string) error {
	certsDir := filepath.Join(workspaceRoot, "nginx", "certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("create certs dir: %w", err)
	}

	certFile := filepath.Join(certsDir, primaryDomain+".crt")
	keyFile := filepath.Join(certsDir, primaryDomain+".key")

	args := []string{
		"-cert-file", certFile,
		"-key-file", keyFile,
	}
	args = append(args, domains...)

	cmd := exec.Command("mkcert", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mkcert: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	// Set permissions explicitly
	_ = os.Chmod(certFile, 0644)
	_ = os.Chmod(keyFile, 0600)
	return nil
}

// MkcertInstalled returns true if mkcert is available on PATH.
func MkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// MkcertCAInstalled returns true if the mkcert CA root is present.
func MkcertCAInstalled() bool {
	cmd := exec.Command("mkcert", "-CAROOT")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	caroot := strings.TrimSpace(string(out))
	if caroot == "" {
		return false
	}
	info, err := os.Stat(caroot)
	return err == nil && info.IsDir()
}

func readPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}
