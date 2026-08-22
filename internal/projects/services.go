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
	nginxBin := filepath.Join(workspaceRoot, "nginx", "sbin", "nginx")
	nginxConfig := filepath.Join(workspaceRoot, "nginx", "etc", "nginx.conf")
	if _, statErr := os.Stat(nginxBin); statErr == nil {
		cmd := exec.Command(nginxBin, "-t", "-c", nginxConfig)
		if output, testErr := cmd.CombinedOutput(); testErr != nil {
			return fmt.Errorf("nginx configuration test failed: %w\n%s", testErr, strings.TrimSpace(string(output)))
		}
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

// MkcertCAInstalled returns true if the mkcert CA root is present AND
// trusted by the system OpenSSL bundle (i.e. usable by PHP/curl).
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
	if err != nil || !info.IsDir() {
		return false
	}

	// Check CA is in system trust: look for the PEM in /etc/ssl/certs or
	// /etc/ca-certificates/extracted/cadir (Arch/Debian patterns).
	caPEM := filepath.Join(caroot, "rootCA.pem")
	if _, err := os.Stat(caPEM); err != nil {
		return false
	}

	// Check if the CA appears in the OpenSSL cert bundle or trust dirs.
	certBundle := "/etc/ssl/cert.pem"
	if data, err := os.ReadFile(certBundle); err == nil {
		if strings.Contains(string(data), "mkcert") {
			return true
		}
	}

	// Fallback: check if mkcert symlink exists in /etc/ssl/certs (Arch pattern)
	if entries, err := os.ReadDir("/etc/ssl/certs"); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "mkcert_") {
				return true
			}
		}
	}

	// Fallback: check Debian/Ubuntu trust paths
	trustAnchors := []string{
		"/usr/local/share/ca-certificates/",
		"/usr/share/ca-certificates/",
		"/etc/ca-certificates/trust/",
	}
	for _, dir := range trustAnchors {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if strings.Contains(e.Name(), "mkcert") {
					return true
				}
			}
		}
	}

	return false
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
