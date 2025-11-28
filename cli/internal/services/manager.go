package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

// ServiceType defines the type of service
type ServiceType int

const (
	ServiceTypeGlobal ServiceType = iota
	ServiceTypeProject
)

// Service represents a manageable Chauffeur service
type Service struct {
	Name    string
	Type    ServiceType
	Slug    string // for project services only
	Binary  string
	PIDFile string
	Args    []string
	Env     []string
}

// ServiceManager handles service lifecycle operations
type ServiceManager struct {
	workspaceDir string
	projectsDir  string
}

var newServiceManagerHook func() (*ServiceManager, error)

// NewServiceManager creates a new service manager
func NewServiceManager() (*ServiceManager, error) {
	if newServiceManagerHook != nil {
		return newServiceManagerHook()
	}

	wsDir, err := workspace.Dir()
	if err != nil {
		return nil, fmt.Errorf("get workspace directory: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		// Use default projects directory if config loading fails
		cfg.ProjectsDir = filepath.Join(wsDir, "projects")
	}

	return &ServiceManager{
		workspaceDir: wsDir,
		projectsDir:  cfg.ProjectsDir,
	}, nil
}

// ListGlobalServices returns all global services (nginx)
func (sm *ServiceManager) ListGlobalServices() []Service {
	services := []Service{}

	// Nginx service
	nginxPath := filepath.Join(sm.workspaceDir, "nginx", "sbin", "nginx")
	if _, err := os.Stat(nginxPath); err == nil {
		// Generate custom nginx config if it doesn't exist
		nginxConfigPath := filepath.Join(sm.workspaceDir, "nginx", "etc", "nginx.conf")
		if _, err := os.Stat(nginxConfigPath); os.IsNotExist(err) {
			if err := sm.generateNginxConfig(); err != nil {
				// Continue even if config generation fails
			}
		}

		services = append(services, Service{
			Name:    "chauf-nginx",
			Type:    ServiceTypeGlobal,
			Binary:  nginxPath,
			PIDFile: filepath.Join(sm.workspaceDir, "nginx", "logs", "nginx.pid"),
			Args:    []string{"-g", "daemon off;", "-c", nginxConfigPath},
			Env:     []string{},
		})
	}

	return services
}

// ListProjectServices returns all PHP-FPM services for a given project
func (sm *ServiceManager) ListProjectServices(projectSlug string) ([]Service, error) {
	// Load project config
	configPath := filepath.Join(sm.projectsDir, projectSlug, "project.yaml")
	projectConfig, err := projects.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load project config: %w", err)
	}

	// Check if PHP version is installed
	phpPath := filepath.Join(sm.workspaceDir, "php", projectConfig.PHP, "bin", "php")
	if _, err := os.Stat(phpPath); err != nil {
		return nil, fmt.Errorf("PHP %s not installed for project %s", projectConfig.PHP, projectSlug)
	}

	// PHP-FPM service
	phpFpmPath := filepath.Join(sm.workspaceDir, "php", projectConfig.PHP, "sbin", "php-fpm")
	if _, err := os.Stat(phpFpmPath); err != nil {
		// Try alternative path
		phpFpmPath = filepath.Join(sm.workspaceDir, "php", projectConfig.PHP, "bin", "php-fpm")
		if _, err := os.Stat(phpFpmPath); err != nil {
			return nil, fmt.Errorf("PHP-FPM binary not found for PHP %s", projectConfig.PHP)
		}
	}

	// Check if project has dedicated FPM
	if projectConfig.Runtime.FPM != nil && projectConfig.Runtime.FPM.Dedicated {
		return sm.createProjectSpecificService(projectSlug, projectConfig, phpFpmPath)
	} else {
		return sm.createVersionSpecificService(projectConfig.PHP, phpFpmPath)
	}
}

// createVersionSpecificService creates a PHP-FPM service shared across all projects using the same PHP version
func (sm *ServiceManager) createVersionSpecificService(phpVersion, phpFpmPath string) ([]Service, error) {
	// Use version-specific runtime directory under PHP installation
	phpVersionDir := filepath.Join(sm.workspaceDir, "php", phpVersion)
	runtimeDir := filepath.Join(phpVersionDir, "runtime", "php-fpm")
	logsDir := filepath.Join(phpVersionDir, "logs")

	// Ensure directories exist
	for _, dir := range []string{runtimeDir, logsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Generate version-specific PHP-FPM config if it doesn't exist
	configPath := filepath.Join(runtimeDir, "php-fpm.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := sm.generateVersionSpecificPHPFPMConfig(phpVersion, runtimeDir, logsDir); err != nil {
			return nil, fmt.Errorf("generate PHP-FPM config: %w", err)
		}
	}

	services := []Service{
		{
			Name:    fmt.Sprintf("chauf-php-fpm-%s", phpVersion),
			Type:    ServiceTypeGlobal, // Treat as global since it's shared
			Slug:    phpVersion,
			Binary:  phpFpmPath,
			PIDFile: filepath.Join(runtimeDir, "php-fpm.pid"),
			Args:    []string{"--nodaemonize", "--fpm-config", configPath},
			Env:     []string{},
		},
	}

	return services, nil
}

// createProjectSpecificService creates a PHP-FPM service specific to a single project
func (sm *ServiceManager) createProjectSpecificService(projectSlug string, projectConfig projects.Config, phpFpmPath string) ([]Service, error) {
	projectDir := filepath.Join(sm.projectsDir, projectSlug)
	runtimeDir := filepath.Join(projectDir, "runtime", "php-fpm")
	logsDir := filepath.Join(projectDir, "logs")

	// Generate PHP-FPM config if it doesn't exist
	configPath := filepath.Join(runtimeDir, "php-fpm.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := sm.generatePHPFPMConfig(projectSlug, projectConfig, runtimeDir, logsDir); err != nil {
			return nil, fmt.Errorf("generate PHP-FPM config: %w", err)
		}
	}

	services := []Service{
		{
			Name:    fmt.Sprintf("chauf-php-fpm-%s", projectSlug),
			Type:    ServiceTypeProject,
			Slug:    projectSlug,
			Binary:  phpFpmPath,
			PIDFile: filepath.Join(runtimeDir, "php-fpm.pid"),
			Args:    []string{"--nodaemonize", "--fpm-config", configPath},
			Env:     []string{},
		},
	}

	return services, nil
}

// OverrideServiceManager allows tests to inject a fake manager.
func OverrideServiceManager(fn func() (*ServiceManager, error)) (reset func()) {
	prev := newServiceManagerHook
	newServiceManagerHook = fn
	return func() { newServiceManagerHook = prev }
}

// generatePHPFPMConfig creates a PHP-FPM pool configuration
func (sm *ServiceManager) generatePHPFPMConfig(projectSlug string, projectConfig projects.Config, runtimeDir, logsDir string) error {
	uid := os.Getuid()
	gid := os.Getgid()

	config := fmt.Sprintf(`[chauf-%s]
user = %d
group = %d
listen = %s
listen.owner = %d
listen.group = %d
listen.mode = 0660
pm = ondemand
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
php_admin_value[error_log] = %s/php-fpm-error.log
php_admin_flag[log_errors] = on
`,
		projectSlug,
		uid,
		gid,
		projectConfig.Runtime.PHPFPM,
		uid,
		gid,
		logsDir,
	)

	return os.WriteFile(filepath.Join(runtimeDir, "php-fpm.conf"), []byte(config), 0644)
}

// generateVersionSpecificPHPFPMConfig creates a PHP-FPM pool configuration shared across all projects of a PHP version
func (sm *ServiceManager) generateVersionSpecificPHPFPMConfig(phpVersion, runtimeDir, logsDir string) error {
	uid := os.Getuid()
	gid := os.Getgid()

	// Use a shared socket path for the PHP version
	socketPath := filepath.Join(runtimeDir, "php-fpm.sock")

	config := fmt.Sprintf(`[chauf-php-%s]
user = %d
group = %d
listen = %s
listen.owner = %d
listen.group = %d
listen.mode = 0660
pm = ondemand
pm.max_children = 20
pm.start_servers = 4
pm.min_spare_servers = 2
pm.max_spare_servers = 6
php_admin_value[error_log] = %s/php-fpm-error.log
php_admin_flag[log_errors] = on
`,
		phpVersion,
		uid,
		gid,
		socketPath,
		uid,
		gid,
		logsDir,
	)

	return os.WriteFile(filepath.Join(runtimeDir, "php-fpm.conf"), []byte(config), 0644)
}

// GetSocketPath returns the appropriate PHP-FPM socket path for a project based on its FPM configuration
func (sm *ServiceManager) GetSocketPath(projectSlug, phpVersion string) (string, error) {
	// Load project config to check FPM mode
	configPath := filepath.Join(sm.projectsDir, projectSlug, "project.yaml")
	projectConfig, err := projects.LoadConfig(configPath)
	if err != nil {
		// If project config doesn't exist yet, default to shared (for linking process)
		phpVersionDir := filepath.Join(sm.workspaceDir, "php", phpVersion)
		runtimeDir := filepath.Join(phpVersionDir, "runtime", "php-fpm")
		return filepath.Join(runtimeDir, "php-fpm.sock"), nil
	}

	// Check if project has dedicated FPM
	if projectConfig.Runtime.FPM != nil && projectConfig.Runtime.FPM.Dedicated {
		// Use project-specific socket
		projectDir := filepath.Join(sm.projectsDir, projectSlug)
		runtimeDir := filepath.Join(projectDir, "runtime", "php-fpm")
		return filepath.Join(runtimeDir, "php-fpm.sock"), nil
	} else {
		// Use shared socket under PHP version directory
		phpVersionDir := filepath.Join(sm.workspaceDir, "php", phpVersion)
		runtimeDir := filepath.Join(phpVersionDir, "runtime", "php-fpm")
		return filepath.Join(runtimeDir, "php-fpm.sock"), nil
	}
}


// IsRunning checks if a service is currently running
func (sm *ServiceManager) IsRunning(service Service) (bool, error) {
	if _, err := os.Stat(service.PIDFile); os.IsNotExist(err) {
		return false, nil
	}

	pidContent, err := os.ReadFile(service.PIDFile)
	if err != nil {
		return false, err
	}

	pid := strings.TrimSpace(string(pidContent))
	if pid == "" {
		return false, nil
	}

	// Check if the process is actually running
	cmd := exec.Command("kill", "-0", pid)
	err = cmd.Run()
	if err != nil {
		// Process doesn't exist, remove stale PID file
		os.Remove(service.PIDFile)
		return false, nil
	}

	return true, nil
}

// Start starts a service
func (sm *ServiceManager) Start(service Service) error {
	// Check if already running
	running, err := sm.IsRunning(service)
	if err != nil {
		return fmt.Errorf("check service status: %w", err)
	}
	if running {
		return fmt.Errorf("service %s is already running", service.Name)
	}

	// Ensure PID directory exists
	if err := os.MkdirAll(filepath.Dir(service.PIDFile), 0755); err != nil {
		return fmt.Errorf("create PID directory: %w", err)
	}

	// Create the command
	cmd := exec.Command(service.Binary, service.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), service.Env...)

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start service %s: %w", service.Name, err)
	}

	// Write PID file
	pid := cmd.Process.Pid
	if err := os.WriteFile(service.PIDFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		// If writing PID file fails, try to kill the process
		cmd.Process.Kill()
		return fmt.Errorf("write PID file for %s: %w", service.Name, err)
	}

	// Give the process time to start and validate configuration
	time.Sleep(1 * time.Second)

	// Verify it's still running and didn't exit due to config errors
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(service.PIDFile)
		return fmt.Errorf("service %s failed to start", service.Name)
	}

	// Additional check: ensure the process is still alive after another second
	time.Sleep(1 * time.Second)
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(service.PIDFile)
		return fmt.Errorf("service %s failed to start - likely configuration error", service.Name)
	}

	return nil
}

// Stop stops a service
func (sm *ServiceManager) Stop(service Service) error {
	// Check if running
	running, err := sm.IsRunning(service)
	if err != nil {
		return fmt.Errorf("check service status: %w", err)
	}
	if !running {
		return fmt.Errorf("service %s is not running", service.Name)
	}

	// Read PID
	pidContent, err := os.ReadFile(service.PIDFile)
	if err != nil {
		return fmt.Errorf("read PID file: %w", err)
	}

	pid := strings.TrimSpace(string(pidContent))
	if pid == "" {
		return fmt.Errorf("invalid PID file")
	}

	// Send SIGTERM
	cmd := exec.Command("kill", "-TERM", pid)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("send SIGTERM to service %s: %w", service.Name, err)
	}

	// Wait up to 5 seconds for graceful shutdown
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		running, err := sm.IsRunning(service)
		if err != nil || !running {
			break
		}
	}

	// Check if still running and force kill if necessary
	running, _ = sm.IsRunning(service)
	if running {
		cmd = exec.Command("kill", "-KILL", pid)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("force kill service %s: %w", service.Name, err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Remove PID file
	os.Remove(service.PIDFile)

	return nil
}

// Restart stops and starts a service
func (sm *ServiceManager) Restart(service Service) error {
	running, err := sm.IsRunning(service)
	if err != nil {
		return fmt.Errorf("check service status: %w", err)
	}

	if running {
		if err := sm.Stop(service); err != nil {
			return fmt.Errorf("stop service %s: %w", service.Name, err)
		}
		time.Sleep(1 * time.Second) // Brief pause between stop and start
	}

	return sm.Start(service)
}

// GetStatus returns the status of a service
func (sm *ServiceManager) GetStatus(service Service) (string, error) {
	running, err := sm.IsRunning(service)
	if err != nil {
		return "unknown", err
	}

	if !running {
		return "stopped", nil
	}

	// Read PID to provide more detailed status
	pidContent, err := os.ReadFile(service.PIDFile)
	if err != nil {
		return "running", nil
	}

	pid := strings.TrimSpace(string(pidContent))
	return fmt.Sprintf("running (pid %s)", pid), nil
}

// GetServiceStatus returns status and error separately to handle "already running" case gracefully
func (sm *ServiceManager) GetServiceStatus(service Service) (bool, error) {
	return sm.IsRunning(service)
}

// generateNginxConfig creates a basic nginx configuration with custom ports
func (sm *ServiceManager) generateNginxConfig() error {
	nginxDir := filepath.Join(sm.workspaceDir, "nginx")
	etcDir := filepath.Join(nginxDir, "etc")
	logsDir := filepath.Join(nginxDir, "logs")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")

	// Ensure directories exist
	for _, dir := range []string{etcDir, logsDir, sitesAvailable, sitesEnabled} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create nginx directory %s: %w", dir, err)
		}
	}

	config := fmt.Sprintf(`# Chauffeur Nginx Configuration
# Custom ports to avoid conflicts with system services

user %s;
worker_processes auto;
error_log %s/error.log notice;
pid %s/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include       %s/mime.types;
    default_type  application/octet-stream;

    # Log format
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log %s/access.log main;

    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    # Include Chauffeur sites
    include %s/sites-enabled/*;
}
`,
		fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		logsDir,
		logsDir,
		nginxDir,
		logsDir,
		nginxDir,
	)

	configPath := filepath.Join(etcDir, "nginx.conf")
	return os.WriteFile(configPath, []byte(config), 0644)
}
