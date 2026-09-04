// Package runtime contains the shared execution contract used by CLI, TUI,
// and future panel operations. Runtime-specific details stay behind this API.
package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/services"
	"github.com/siegg/chauffeur/internal/workspace"
)

type Engine string

const (
	EngineNative Engine = "native"
	EnginePodman Engine = "podman"
)

type Scope struct {
	Workspace string
	Project   string
	Slug      string
	Version   string
	Dedicated bool
	Service   string
}

type ExecOptions struct {
	Dir    string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type LogOptions struct {
	Follow bool
	Lines  int
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string { return e.Err.Error() }
func (e *ExitCodeError) Unwrap() error { return e.Err }
func (e *ExitCodeError) ExitCode() int { return e.Code }

type ServiceStatus struct {
	Name      string
	Label     string
	Role      string
	State     string
	Healthy   bool
	Evidence  string
	Container string
	Image     string
	Digest    string
	Projects  []string
	PID       int
	StartedAt time.Time
	Uptime    time.Duration
	MemoryMB  int
	Ready     bool
	Network   string
}

// ServiceOperation is the common lifecycle result consumed by CLI and UI
// renderers. Backends may populate runtime-specific fields in Service.
type ServiceOperation struct {
	Service ServiceStatus
	Before  string
	After   string
	Changed bool
	Message string
	Err     error
}

// OperationResult preserves partial outcomes while making aggregate success
// truthful for callers that need to render every service attempt.
type OperationResult struct {
	Action   string
	Services []ServiceOperation
	Err      error
}

type Runtime interface {
	Ensure(ctx context.Context, version string) error
	EnsureProject(ctx context.Context, scope Scope, image string, roots map[string]string) error
	EnsureLinkedProject(ctx context.Context, workspace, configPath, certificateDir string, httpPort, httpsPort int, project ProjectSpec) error
	EnsureWorkspace(ctx context.Context, scope WorkspaceScope) error
	RemoveProject(ctx context.Context, scope Scope) error
	RemoveImage(ctx context.Context, image string, force bool) error
	Start(ctx context.Context, scope Scope) error
	Stop(ctx context.Context, scope Scope) error
	Restart(ctx context.Context, scope Scope) error
	Status(ctx context.Context, scope Scope) ([]ServiceStatus, error)
	Logs(ctx context.Context, scope Scope, opts LogOptions) (string, error)
	Exec(ctx context.Context, scope Scope, command []string, opts ExecOptions) (CommandResult, error)
}

type PHPContainerScope struct {
	Scope Scope
	Image string
	Roots map[string]string
}

type WorkspaceScope struct {
	Workspace      string
	ConfigPath     string
	CertificateDir string
	HTTPPort       int
	HTTPSPort      int
	Roots          map[string]string
	PHPContainers  []PHPContainerScope
	Routes         []NginxRoute
}

type CommandRunner interface {
	Run(ctx context.Context, args ...string) (CommandResult, error)
}

type Podman struct{ Runner CommandRunner }

const NginxImage = "docker.io/library/nginx:alpine"

func (p Podman) run(ctx context.Context, args ...string) (CommandResult, error) {
	if p.Runner == nil {
		return CommandResult{}, fmt.Errorf("podman command runner is not configured")
	}
	return p.Runner.Run(ctx, args...)
}

// Preflight verifies the host can execute rootless Podman before any resource
// creation is attempted. The output is intentionally kept small and stable so
// callers can turn failures into actionable diagnostics.
func (p Podman) Preflight(ctx context.Context) error {
	if result, err := p.run(ctx, "version", "--format", "{{.Version}}"); err != nil {
		return fmt.Errorf("Podman is unavailable: %w; use the native runtime or install Podman", err)
	} else if result.ExitCode != 0 {
		return fmt.Errorf("Podman version check failed with status %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	result, err := p.run(ctx, "info", "--format", "{{.Host.Security.Rootless}}")
	if err != nil {
		return fmt.Errorf("Podman preflight failed: %w; verify rootless Podman is configured", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Podman info failed with status %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	if strings.EqualFold(strings.TrimSpace(result.Stdout), "false") {
		return fmt.Errorf("rootless Podman is unavailable; use rootless mode or switch to the native runtime")
	}
	return nil
}

// EnsureReady verifies Podman and starts a Podman Machine when the host uses
// one (for example macOS). On daemonless Linux, the info check is sufficient.
func (p Podman) EnsureReady(ctx context.Context) error {
	if ready, rootless, _ := p.podmanInfo(ctx); ready {
		return requireRootless(rootless)
	}
	if result, err := p.run(ctx, "machine", "start"); err != nil || result.ExitCode != 0 {
		return fmt.Errorf("Podman is not ready; start Podman or its machine before starting Chauffeur")
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ready, rootless, _ := p.podmanInfo(ctx); ready {
			return requireRootless(rootless)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("Podman machine started but Podman did not become ready within 5 seconds")
		case <-ticker.C:
		}
	}
}

func (p Podman) podmanInfo(ctx context.Context) (ready, rootless bool, err error) {
	result, err := p.run(ctx, "info", "--format", "{{.Host.Security.Rootless}}")
	if err != nil || result.ExitCode != 0 {
		return false, false, err
	}
	return true, !strings.EqualFold(strings.TrimSpace(result.Stdout), "false"), nil
}

func requireRootless(rootless bool) error {
	if !rootless {
		return fmt.Errorf("rootless Podman is unavailable; use rootless mode or switch to the native runtime")
	}
	return nil
}

func (p Podman) Ensure(ctx context.Context, version string) error {
	if _, err := NormalizeVersion(version); err != nil {
		return err
	}
	result, err := p.run(ctx, "container", "exists", FPMContainerName(version))
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("PHP %s FPM container is not available", version)
	}
	return nil
}

func (p Podman) EnsureProject(ctx context.Context, scope Scope, image string, roots map[string]string) error {
	return p.EnsurePHPContainerWithRoots(ctx, scope, image, roots)
}

func (p Podman) RemoveProject(ctx context.Context, scope Scope) error {
	if !scope.Dedicated {
		return nil
	}
	name := DedicatedContainerName(scope.Project)
	result, err := p.run(ctx, "container", "exists", name)
	if err != nil || result.ExitCode != 0 {
		return err
	}
	removed, err := p.run(ctx, "container", "rm", "-f", name)
	if err != nil {
		return err
	}
	if removed.ExitCode != 0 {
		return commandFailure("remove dedicated PHP container "+name, removed)
	}
	return nil
}

func (p Podman) EnsureLinkedProject(ctx context.Context, workspace, configPath, certificateDir string, httpPort, httpsPort int, project ProjectSpec) error {
	scope, err := BuildWorkspaceScope(workspace, configPath, certificateDir, httpPort, httpsPort, []ProjectSpec{project})
	if err != nil {
		return err
	}
	return p.EnsureWorkspace(ctx, scope)
}

func (p Podman) RemoveImage(ctx context.Context, image string, force bool) error {
	args := []string{"image", "rm"}
	if force {
		args = append(args, "--force")
	}
	result, err := p.run(ctx, append(args, image)...)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("remove Podman image %s failed with status %d: %s", image, result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	return nil
}

func (p Podman) EnsureWorkspace(ctx context.Context, workspace WorkspaceScope) error {
	if workspace.Workspace == "" || workspace.ConfigPath == "" {
		return fmt.Errorf("Podman workspace scope is missing workspace or nginx config path")
	}
	if err := p.EnsureReady(ctx); err != nil {
		return err
	}
	if err := p.Preflight(ctx); err != nil {
		return err
	}
	if err := p.validateNginxPorts(ctx, workspace.HTTPPort, workspace.HTTPSPort); err != nil {
		return err
	}
	for _, container := range workspace.PHPContainers {
		for _, hostPath := range container.Roots {
			info, err := os.Stat(hostPath)
			if err != nil || !info.IsDir() {
				return fmt.Errorf("linked project path is unavailable for PHP %s: %s; remove the stale link with `chauf unlink %s --yes` or relink the project", container.Scope.Version, hostPath, hostPath)
			}
		}
	}
	// Validate every image before creating or starting any container. This
	// prevents a stale project from causing partial workspace startup.
	for _, container := range workspace.PHPContainers {
		if result, err := p.run(ctx, "image", "exists", container.Image); err != nil {
			return err
		} else if result.ExitCode != 0 {
			return missingPHPImageError(container.Scope, container.Image)
		}
		if _, err := InspectImage(ctx, p.Runner, container.Image); err != nil {
			return fmt.Errorf("PHP %s image metadata is unavailable for %s: %w", container.Scope.Version, container.Scope.Project, err)
		}
	}
	if result, err := p.run(ctx, "image", "exists", NginxImage); err != nil {
		return err
	} else if result.ExitCode != 0 {
		return fmt.Errorf("nginx image is unavailable: %s; run `podman pull %s`", NginxImage, NginxImage)
	}
	for _, route := range workspace.Routes {
		if route.SSL {
			if _, err := os.Stat(filepath.Join(workspace.CertificateDir, route.CertName+".crt")); err != nil {
				return fmt.Errorf("SSL certificate for %s is unavailable: %w", route.ServerName, err)
			}
			if _, err := os.Stat(filepath.Join(workspace.CertificateDir, route.CertName+".key")); err != nil {
				return fmt.Errorf("SSL key for %s is unavailable: %w", route.ServerName, err)
			}
		}
	}
	for _, container := range workspace.PHPContainers {
		name := scopeContainerName(container.Scope)
		exists, err := p.run(ctx, "container", "exists", name)
		if err != nil {
			return err
		}
		if exists.ExitCode != 0 {
		} else {
			_, _, inspectErr := p.inspectPHPContainer(ctx, name, container.Scope, container.Roots)
			if inspectErr != nil {
				return inspectErr
			}
		}
		if err := p.EnsurePHPContainerWithRoots(ctx, container.Scope, container.Image, container.Roots); err != nil {
			return err
		}
	}
	// 8080/8443 are the nginx container's listen ports. Host ports are only
	// used by the publish arguments below.
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS(workspace.Routes, 8080, 8443)
	if err != nil {
		return err
	}
	if err := os.WriteFile(workspace.ConfigPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("write Podman nginx config: %w", err)
	}
	if err := p.EnsureNginxContainerWithRootsAndPorts(ctx, workspace.ConfigPath, workspace.Roots, workspace.HTTPPort, workspace.HTTPSPort, workspace.CertificateDir); err != nil {
		return err
	}
	state, err := p.run(ctx, "container", "inspect", "chauf-nginx", "--format", "{{.State.Status}}")
	if err != nil {
		return err
	}
	if state.ExitCode == 0 && strings.TrimSpace(state.Stdout) == "running" {
		restarted, restartErr := p.run(ctx, "container", "restart", "chauf-nginx")
		if restartErr != nil {
			return restartErr
		}
		if restarted.ExitCode != 0 {
			return commandFailure("refresh nginx after workspace reconciliation", restarted)
		}
	}
	return nil
}

func (p Podman) validateReverseProxyRoutes(ctx context.Context, routes []NginxRoute) error {
	for _, route := range routes {
		if route.ProxyPort == 0 {
			continue
		}
		result, err := p.run(ctx, "exec", "chauf-nginx", "wget", "-q", "-O", "/dev/null", "-T", "3", fmt.Sprintf("http://%s:%d/", route.Upstream, route.ProxyPort))
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("reverse proxy %s cannot reach %s:%d from Podman nginx; bind the host development server to a container-reachable interface (for example, Vite: `pnpm dev -- --host 0.0.0.0`)", route.ServerName, route.Upstream, route.ProxyPort)
		}
	}
	return nil
}

func (p Podman) validateNginxPorts(ctx context.Context, httpPort, httpsPort int) error {
	for _, port := range []int{httpPort, httpsPort} {
		if port == 0 {
			continue
		}
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = listener.Close()
			continue
		}
		result, runErr := p.run(ctx, "container", "inspect", "chauf-nginx", "--format", "{{.State.Status}}")
		if runErr == nil && result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "running" {
			continue
		}
		return fmt.Errorf("nginx host port %d is unavailable: %w; choose a free port with `chauf config nginx http_port <port>` or `chauf config nginx https_port <port>`", port, err)
	}
	return nil
}

func missingPHPImageError(scope Scope, image string) error {
	return fmt.Errorf("PHP %s image is unavailable for project %s: %s; run `chauf install php %s` or `chauf install php %s --build`, or isolate the project to an available version", scope.Version, scope.Project, image, scope.Version, scope.Version)
}

func commandFailure(action string, result CommandResult) error {
	evidence := strings.TrimSpace(result.Stderr)
	if evidence == "" {
		evidence = strings.TrimSpace(result.Stdout)
	}
	if evidence == "" {
		evidence = "no Podman diagnostics were returned"
	}
	return fmt.Errorf("%s failed with status %d: %s", action, result.ExitCode, evidence)
}

// EnsurePHPContainer creates the version container when the image is already
// available. Image pulls and builds remain explicit operations.
func (p Podman) EnsurePHPContainer(ctx context.Context, scope Scope, image, projectRoot string) error {
	return p.EnsurePHPContainerWithRoots(ctx, scope, image, map[string]string{"/workspace": projectRoot})
}

func (p Podman) EnsurePHPContainerWithRoots(ctx context.Context, scope Scope, image string, roots map[string]string) error {
	if len(roots) == 0 {
		return fmt.Errorf("PHP container requires at least one project mount")
	}
	if err := p.Preflight(ctx); err != nil {
		return err
	}
	if err := EnsureNetwork(ctx, p.Runner); err != nil {
		return err
	}
	if result, err := p.run(ctx, "image", "exists", image); err != nil {
		return err
	} else if result.ExitCode != 0 {
		return missingPHPImageError(scope, image)
	}
	imageMetadata, err := InspectImage(ctx, p.Runner, image)
	if err != nil {
		return fmt.Errorf("PHP %s image metadata is unavailable: %w", scope.Version, err)
	}
	name := scopeContainerName(scope)
	if result, err := p.run(ctx, "container", "exists", name); err != nil {
		return err
	} else if result.ExitCode == 0 {
		state, matches, err := p.inspectPHPContainer(ctx, name, scope, roots, imageMetadata.ID)
		if err != nil {
			return err
		}
		if !matches {
			removed, removeErr := p.run(ctx, "container", "rm", "-f", name)
			if removeErr != nil {
				return removeErr
			}
			if removed.ExitCode != 0 {
				return commandFailure("remove stale PHP container "+name, removed)
			}
		} else {
			if state == "running" {
				return p.checkPHPReady(ctx, name)
			}
			started, startErr := p.run(ctx, "container", "start", name)
			if startErr != nil {
				return startErr
			}
			if started.ExitCode != 0 {
				return commandFailure(fmt.Sprintf("start PHP %s container", scope.Version), started)
			}
			return p.checkPHPReady(ctx, name)
		}
	}
	scopeKind := "shared"
	if scope.Dedicated {
		scopeKind = "dedicated"
	}
	args := []string{"run", "-d", "--userns", "keep-id", "--name", name, "--network", "chauf-net", "--label", "com.siegg.chauffeur.role=php-fpm", "--label", "com.siegg.chauffeur.php.version=" + scope.Version, "--label", "com.siegg.chauffeur.scope=" + scopeKind, "--label", "com.siegg.chauffeur.userns=keep-id"}
	for containerPath, hostPath := range roots {
		args = append(args, "--volume", hostPath+":"+containerPath+":rw")
	}
	args = append(args, image)
	result, err := p.run(ctx, args...)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return commandFailure(fmt.Sprintf("create PHP %s container", scope.Version), result)
	}
	if err := p.waitContainerRunning(ctx, name); err != nil {
		return err
	}
	return p.checkPHPReady(ctx, name)
}

type phpContainerInspection struct {
	State struct {
		Status string `json:"Status"`
	} `json:"State"`
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	Image  string `json:"Image"`
	Mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
	NetworkSettings struct {
		Networks map[string]struct{} `json:"Networks"`
	} `json:"NetworkSettings"`
}

func (p Podman) inspectPHPContainer(ctx context.Context, name string, scope Scope, roots map[string]string, expectedImageID ...string) (string, bool, error) {
	result, err := p.run(ctx, "container", "inspect", name, "--format", "json")
	if err != nil {
		return "", false, err
	}
	if result.ExitCode != 0 {
		return "", false, commandFailure("inspect PHP container "+name, result)
	}
	var entries []phpContainerInspection
	if err := json.Unmarshal([]byte(result.Stdout), &entries); err != nil || len(entries) == 0 {
		return "", false, fmt.Errorf("inspect PHP container %s returned invalid metadata", name)
	}
	entry := entries[0]
	if len(expectedImageID) > 0 && expectedImageID[0] != "" && entry.Image != expectedImageID[0] {
		return entry.State.Status, false, nil
	}
	scopeKind := "shared"
	if scope.Dedicated {
		scopeKind = "dedicated"
	}
	if entry.Config.Labels["com.siegg.chauffeur.role"] != "php-fpm" ||
		entry.Config.Labels["com.siegg.chauffeur.php.version"] != scope.Version ||
		entry.Config.Labels["com.siegg.chauffeur.scope"] != scopeKind || len(entry.Mounts) != len(roots) {
		return entry.State.Status, false, nil
	}
	if entry.Config.Labels["com.siegg.chauffeur.userns"] != "keep-id" {
		return entry.State.Status, false, nil
	}
	if len(entry.NetworkSettings.Networks) > 0 {
		if _, attached := entry.NetworkSettings.Networks["chauf-net"]; !attached {
			return entry.State.Status, false, nil
		}
	}
	for _, mount := range entry.Mounts {
		matched := false
		for destination, source := range roots {
			if filepath.Clean(mount.Source) == filepath.Clean(source) && mount.Destination == destination && mount.RW {
				matched = true
				break
			}
		}
		if !matched {
			return entry.State.Status, false, nil
		}
	}
	return entry.State.Status, true, nil
}

func (p Podman) checkPHPReady(ctx context.Context, name string) error {
	result, err := p.run(ctx, "container", "exec", name, "php", "-v")
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("PHP runtime in %s is not ready: php -v exited with status %d", name, result.ExitCode)
	}
	// php -v only proves the CLI exists. Probe the FPM listener as well so nginx
	// is not started against a container whose pool has not bound port 9000.
	probe, err := p.run(ctx, "container", "exec", name, "php", "-r", "$s=@fsockopen('127.0.0.1',9000,$e,$m,1); exit($s ? 0 : 1);")
	if err != nil {
		return err
	}
	if probe.ExitCode != 0 {
		return fmt.Errorf("PHP-FPM in %s is not ready: port 9000 is not accepting connections", name)
	}
	return nil
}

// EnsureNginxContainer prepares a workspace nginx container. Configuration
// generation remains outside the runtime so link/apply can preview it first.
func (p Podman) EnsureNginxContainer(ctx context.Context, configPath, projectRoot string, hostPort int) error {
	return p.EnsureNginxContainerWithRootsAndPorts(ctx, configPath, map[string]string{"/workspace": projectRoot}, hostPort, 0, "")
}

func (p Podman) EnsureNginxContainerWithRoots(ctx context.Context, configPath string, roots map[string]string, hostPort int) error {
	return p.EnsureNginxContainerWithRootsAndPorts(ctx, configPath, roots, hostPort, 0, "")
}

func (p Podman) EnsureNginxContainerWithRootsAndPorts(ctx context.Context, configPath string, roots map[string]string, httpPort, httpsPort int, certDir string) error {
	if httpPort < 1 || httpPort > 65535 {
		return fmt.Errorf("invalid nginx HTTP port %d", httpPort)
	}
	if httpsPort < 0 || httpsPort > 65535 {
		return fmt.Errorf("invalid nginx HTTPS port %d", httpsPort)
	}
	if httpsPort != 0 && httpPort == httpsPort {
		return fmt.Errorf("nginx HTTP and HTTPS ports must be different")
	}
	if err := p.Preflight(ctx); err != nil {
		return err
	}
	if err := p.validateNginxPorts(ctx, httpPort, httpsPort); err != nil {
		return err
	}
	if err := EnsureNetwork(ctx, p.Runner); err != nil {
		return err
	}
	if result, err := p.run(ctx, "image", "exists", NginxImage); err != nil {
		return err
	} else if result.ExitCode != 0 {
		return fmt.Errorf("nginx image is unavailable: %s; run `podman pull %s`", NginxImage, NginxImage)
	}
	name := "chauf-nginx"
	if result, err := p.run(ctx, "container", "exists", name); err != nil {
		return err
	} else if result.ExitCode == 0 {
		state, err := p.run(ctx, "container", "inspect", name, "--format", "{{.State.Status}}")
		if err != nil {
			return err
		}
		if strings.TrimSpace(state.Stdout) == "running" {
			matches, err := p.nginxContainerMountsMatch(ctx, name, configPath, roots, certDir, httpPort, httpsPort)
			if err != nil {
				return err
			}
			if matches {
				return nil
			}
			removed, err := p.run(ctx, "container", "rm", "-f", name)
			if err != nil {
				return err
			}
			if removed.ExitCode != 0 {
				return commandFailure("remove stale nginx container", removed)
			}
		}
		removed, err := p.run(ctx, "container", "rm", "-f", name)
		if err != nil {
			return err
		}
		if removed.ExitCode != 0 {
			return commandFailure("remove stale nginx container", removed)
		}
	}
	args := []string{"run", "-d", "--name", name, "--network", "chauf-net", "--label", "com.siegg.chauffeur.role=nginx", "--label", "com.siegg.chauffeur.workspace=" + filepath.Base(filepath.Dir(filepath.Dir(configPath))), "--publish", fmt.Sprintf("%d:8080", httpPort), "--volume", configPath + ":/etc/nginx/conf.d/default.conf:ro"}
	if httpsPort > 0 {
		args = append(args, "--publish", fmt.Sprintf("%d:8443", httpsPort))
	}
	if certDir != "" {
		args = append(args, "--volume", certDir+":/etc/nginx/certs:ro")
	}
	for containerPath, hostPath := range roots {
		args = append(args, "--volume", hostPath+":"+containerPath+":ro")
	}
	args = append(args, NginxImage)
	result, err := p.run(ctx, args...)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return commandFailure("create nginx container", result)
	}
	return p.waitContainerRunning(ctx, name)
}

func (p Podman) nginxContainerMountsMatch(ctx context.Context, name, configPath string, roots map[string]string, certDir string, httpPort, httpsPort int) (bool, error) {
	result, err := p.run(ctx, "container", "inspect", name, "--format", "json")
	if err != nil {
		return false, err
	}
	if result.ExitCode != 0 {
		return false, commandFailure("inspect nginx container "+name, result)
	}
	var entries []struct {
		Config struct {
			Image  string            `json:"Image"`
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		NetworkSettings struct {
			Ports map[string][]struct {
				HostPort string `json:"HostPort"`
			} `json:"Ports"`
			Networks map[string]struct{} `json:"Networks"`
		} `json:"NetworkSettings"`
		Mounts []struct {
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
		} `json:"Mounts"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &entries); err != nil || len(entries) == 0 {
		return false, fmt.Errorf("inspect nginx container %s returned invalid metadata", name)
	}
	actual := make(map[string]string, len(entries[0].Mounts))
	for _, mount := range entries[0].Mounts {
		actual[mount.Destination] = filepath.Clean(mount.Source)
	}
	want := map[string]string{"/etc/nginx/conf.d/default.conf": filepath.Clean(configPath)}
	if certDir != "" {
		want["/etc/nginx/certs"] = filepath.Clean(certDir)
	}
	for destination, source := range roots {
		want[destination] = filepath.Clean(source)
	}
	if len(actual) != len(want) {
		return false, nil
	}
	for destination, source := range want {
		if actual[destination] != source {
			return false, nil
		}
	}
	if entries[0].Config.Image != "" && entries[0].Config.Image != NginxImage {
		return false, nil
	}
	wantedWorkspace := filepath.Base(filepath.Dir(filepath.Dir(configPath)))
	if label := entries[0].Config.Labels["com.siegg.chauffeur.workspace"]; label != "" && label != wantedWorkspace {
		return false, nil
	}
	if len(entries[0].NetworkSettings.Networks) > 0 {
		if _, attached := entries[0].NetworkSettings.Networks["chauf-net"]; !attached {
			return false, nil
		}
	}
	if !publishedPortMatches(entries[0].NetworkSettings.Ports, "8080/tcp", httpPort) || (httpsPort > 0 && !publishedPortMatches(entries[0].NetworkSettings.Ports, "8443/tcp", httpsPort)) {
		return false, nil
	}
	return true, nil
}

func publishedPortMatches(ports map[string][]struct {
	HostPort string `json:"HostPort"`
}, containerPort string, hostPort int) bool {
	if hostPort == 0 {
		return true
	}
	for _, binding := range ports[containerPort] {
		if binding.HostPort == fmt.Sprintf("%d", hostPort) {
			return true
		}
	}
	return false
}

func (p Podman) waitContainerRunning(ctx context.Context, name string) error {
	state, err := p.run(ctx, "container", "inspect", name, "--format", "{{.State.Status}}")
	if err != nil {
		return err
	}
	if strings.TrimSpace(state.Stdout) != "running" {
		logs, _ := p.LogsContainer(ctx, name, LogOptions{Lines: 20})
		return fmt.Errorf("container %s is not ready (state %q): %s", name, strings.TrimSpace(state.Stdout), strings.TrimSpace(logs))
	}
	return nil
}

func (p Podman) Start(ctx context.Context, scope Scope) error {
	if err := p.EnsureReady(ctx); err != nil {
		return err
	}
	return p.StartContainer(ctx, scopeContainerName(scope))
}

func (p Podman) StartContainer(ctx context.Context, name string) error {
	if state, err := p.containerState(ctx, name); err == nil && state == "running" {
		return nil
	}
	result, err := p.run(ctx, "container", "start", name)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		if detail := strings.TrimSpace(result.Stderr); detail != "" {
			return fmt.Errorf("start container %s failed with status %d: %s", name, result.ExitCode, detail)
		}
		return fmt.Errorf("start container %s failed with status %d", name, result.ExitCode)
	}
	return nil
}

func (p Podman) Stop(ctx context.Context, scope Scope) error {
	return p.StopContainer(ctx, scopeContainerName(scope))
}

func (p Podman) StopContainer(ctx context.Context, name string) error {
	state, stateErr := p.containerState(ctx, name)
	if stateErr == nil && (state == "" || state == "exited" || state == "stopped" || state == "created") {
		return nil
	}
	result, err := p.run(ctx, "container", "stop", name)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		if detail := strings.TrimSpace(result.Stderr); detail != "" {
			return fmt.Errorf("stop container %s failed with status %d: %s", name, result.ExitCode, detail)
		}
		return fmt.Errorf("stop container %s failed with status %d", name, result.ExitCode)
	}
	return nil
}

func (p Podman) containerState(ctx context.Context, name string) (string, error) {
	result, err := p.run(ctx, "container", "inspect", name, "--format", "{{.State.Status}}")
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", nil
	}
	return strings.TrimSpace(result.Stdout), nil
}

func (p Podman) Restart(ctx context.Context, scope Scope) error {
	return p.RestartContainer(ctx, scopeContainerName(scope))
}

func (p Podman) RestartContainer(ctx context.Context, name string) error {
	result, err := p.run(ctx, "container", "restart", name)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		if detail := strings.TrimSpace(result.Stderr); detail != "" {
			return fmt.Errorf("restart container %s failed with status %d: %s", name, result.ExitCode, detail)
		}
		return fmt.Errorf("restart container %s failed with status %d", name, result.ExitCode)
	}
	return nil
}

func (p Podman) Status(ctx context.Context, scope Scope) ([]ServiceStatus, error) {
	name := scopeContainerName(scope)
	role := "php-fpm"
	image := PHPImage(scope.Version)
	label := fmt.Sprintf("php-fpm %s", scope.Version)
	if scope.Dedicated {
		label += " (dedicated)"
	} else {
		label += " (shared)"
	}
	if scope.Service == "nginx" {
		role = "nginx"
		image = NginxImage
		label = "nginx"
	}
	if result, err := p.run(ctx, "image", "exists", image); err != nil {
		return nil, err
	} else if result.ExitCode != 0 {
		return []ServiceStatus{{Name: name, Label: label, Role: role, State: "image-missing", Image: image, Container: name, Evidence: "image unavailable; install or build it before starting"}}, nil
	}
	statuses, err := p.StatusContainer(ctx, name, role)
	if err != nil {
		return nil, err
	}
	for i := range statuses {
		statuses[i].Image = image
		statuses[i].Label = label
	}
	return statuses, nil
}

func (p Podman) StatusContainer(ctx context.Context, name, role string) ([]ServiceStatus, error) {
	result, err := p.run(ctx, "container", "inspect", name, "--format", "json")
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return []ServiceStatus{{Name: name, Label: name, Role: role, State: "absent", Healthy: false, Container: name, Evidence: strings.TrimSpace(result.Stderr)}}, nil
	}
	var entries []struct {
		State struct {
			Status  string    `json:"Status"`
			Pid     int       `json:"Pid"`
			Started time.Time `json:"StartedAt"`
			Health  struct {
				Status string `json:"Status"`
			} `json:"Healthcheck"`
		} `json:"State"`
		NetworkSettings struct {
			Networks map[string]struct{} `json:"Networks"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &entries); err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("inspect container %s returned invalid metadata", name)
	}
	e := entries[0]
	state := e.State.Status
	ready := state == "running"
	if e.State.Health.Status != "" {
		ready = ready && e.State.Health.Status == "healthy"
	}
	status := ServiceStatus{Name: name, Label: name, Role: role, State: state, Healthy: ready, Ready: ready, Container: name, PID: e.State.Pid, StartedAt: e.State.Started, Evidence: "container inspect state"}
	if !e.State.Started.IsZero() && state == "running" {
		status.Uptime = time.Since(e.State.Started).Round(time.Second)
	}
	for network := range e.NetworkSettings.Networks {
		status.Network = network
		break
	}
	// Memory is intentionally best-effort: stats is unavailable for some
	// rootless backends and must not make status fail.
	if stats, statsErr := p.run(ctx, "stats", "--no-stream", "--format", "{{.MemUsage}}", name); statsErr == nil && stats.ExitCode == 0 {
		status.MemoryMB = parseMemoryMB(stats.Stdout)
	}
	return []ServiceStatus{status}, nil
}

func parseMemoryMB(value string) int {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return 0
	}
	value = strings.ToUpper(strings.TrimSpace(strings.Split(fields[0], "/")[0]))
	for _, suffix := range []struct {
		name   string
		factor float64
	}{{"GIB", 1024}, {"GB", 1024}, {"MIB", 1}, {"MB", 1}, {"KIB", 1.0 / 1024}, {"KB", 1.0 / 1024}} {
		if strings.HasSuffix(value, suffix.name) {
			n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, suffix.name)), 64)
			if err == nil {
				return int(n * suffix.factor)
			}
		}
	}
	return 0
}

func (p Podman) Logs(ctx context.Context, scope Scope, opts LogOptions) (string, error) {
	return p.LogsContainer(ctx, scopeContainerName(scope), opts)
}

func (p Podman) LogsContainer(ctx context.Context, name string, opts LogOptions) (string, error) {
	args := []string{"logs"}
	if opts.Follow {
		args = append(args, "--follow")
	}
	if opts.Lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.Lines))
	}
	args = append(args, name)
	result, err := p.run(ctx, args...)
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("read logs failed with status %d", result.ExitCode)
	}
	return result.Stdout, nil
}

// StreamLogs writes Podman logs directly to stdout/stderr. Unlike Logs, this
// does not buffer follow-mode output until the command exits.
func (p Podman) StreamLogs(ctx context.Context, scope Scope, opts LogOptions, stdout, stderr io.Writer) error {
	streamer, ok := p.Runner.(interface {
		Stream(context.Context, io.Writer, io.Writer, ...string) error
	})
	if !ok {
		return fmt.Errorf("Podman runner does not support streaming logs")
	}
	args := []string{"logs", "--follow"}
	if opts.Lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.Lines))
	}
	args = append(args, scopeContainerName(scope))
	return streamer.Stream(ctx, stdout, stderr, args...)
}

func (p Podman) Exec(ctx context.Context, scope Scope, command []string, opts ExecOptions) (CommandResult, error) {
	if len(command) == 0 {
		return CommandResult{}, fmt.Errorf("runtime command cannot be empty")
	}
	args := []string{"exec"}
	if opts.Dir != "" {
		args = append(args, "--workdir", opts.Dir)
	}
	for _, env := range opts.Env {
		args = append(args, "--env", env)
	}
	args = append(args, scopeContainerName(scope))
	args = append(args, command...)
	result, err := p.run(ctx, args...)
	if opts.Stdout != nil {
		_, _ = io.WriteString(opts.Stdout, result.Stdout)
	}
	if opts.Stderr != nil {
		_, _ = io.WriteString(opts.Stderr, result.Stderr)
	}
	if result.ExitCode != 0 {
		return result, &ExitCodeError{Code: result.ExitCode, Err: fmt.Errorf("runtime command exited with status %d", result.ExitCode)}
	}
	return result, err
}

type Native struct {
	Root     string
	ExecFunc func(context.Context, Scope, []string, ExecOptions) (CommandResult, error)
}

func (n Native) Ensure(context.Context, string) error { return nil }
func (n Native) EnsureProject(context.Context, Scope, string, map[string]string) error {
	return nil
}
func (n Native) RemoveProject(context.Context, Scope) error { return nil }
func (n Native) EnsureLinkedProject(context.Context, string, string, string, int, int, ProjectSpec) error {
	return nil
}
func (n Native) RemoveImage(context.Context, string, bool) error {
	return fmt.Errorf("native runtime does not manage images")
}
func (n Native) EnsureWorkspace(context.Context, WorkspaceScope) error { return nil }
func (n Native) root() string {
	if n.Root != "" {
		return n.Root
	}
	return workspace.Root()
}

func (n Native) service(scope Scope) (*services.NginxService, *services.FPMService) {
	if scope.Service == "nginx" {
		return services.NewNginxService(n.root()), nil
	}
	if scope.Dedicated {
		return nil, services.NewDedicatedFPM(n.root(), scope.Slug, scope.Version, "")
	}
	return nil, services.NewSharedFPM(n.root(), scope.Version)
}

func (n Native) Start(_ context.Context, scope Scope) error {
	nginx, fpm := n.service(scope)
	if nginx != nil {
		return nginx.Start()
	}
	return fpm.Start()
}
func (n Native) Stop(_ context.Context, scope Scope) error {
	nginx, fpm := n.service(scope)
	if nginx != nil {
		return nginx.Stop(30 * time.Second)
	}
	return fpm.Stop(30 * time.Second)
}
func (n Native) Restart(_ context.Context, scope Scope) error {
	nginx, fpm := n.service(scope)
	if nginx != nil {
		return nginx.Reload()
	}
	return fpm.Reload()
}
func (n Native) Status(_ context.Context, scope Scope) ([]ServiceStatus, error) {
	nginx, fpm := n.service(scope)
	if nginx != nil {
		return []ServiceStatus{{Name: "nginx", Label: "nginx", Role: "nginx", State: nativeState(nginx.IsRunning()), Healthy: nginx.IsRunning(), Ready: nginx.IsRunning(), PID: nginx.PID(), Uptime: nginx.Uptime(), MemoryMB: nginx.MemoryMB(), Evidence: "native process"}}, nil
	}
	running := fpm.IsRunning()
	return []ServiceStatus{{Name: fpm.Label(), Label: "php-fpm " + fpm.Label(), Role: "php-fpm", State: nativeState(running), Healthy: running, Ready: running, PID: fpm.PID(), Uptime: fpm.Uptime(), MemoryMB: fpm.MemoryMB(), Evidence: "native process"}}, nil
}
func nativeState(running bool) string {
	if running {
		return "running"
	}
	return "stopped"
}
func (n Native) Logs(_ context.Context, scope Scope, _ LogOptions) (string, error) {
	path := filepath.Join(n.root(), "nginx", "logs", "error.log")
	if scope.Service != "nginx" {
		path = filepath.Join(n.root(), "php", scope.Version, "var", "log", "php-fpm.log")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func (n Native) Exec(ctx context.Context, scope Scope, command []string, opts ExecOptions) (CommandResult, error) {
	if n.ExecFunc != nil {
		return n.ExecFunc(ctx, scope, command, opts)
	}
	if len(command) == 0 {
		return CommandResult{}, fmt.Errorf("runtime command cannot be empty")
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = opts.Dir
	if opts.Env != nil {
		cmd.Env = opts.Env
	}
	cmd.Stdin = opts.Stdin
	var stdout, stderr bytes.Buffer
	if opts.Stdout == nil {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = opts.Stdout
	}
	if opts.Stderr == nil {
		cmd.Stderr = &stderr
	} else {
		cmd.Stderr = opts.Stderr
	}
	if err := cmd.Run(); err != nil {
		code := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		return CommandResult{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}, err
	}
	return CommandResult{ExitCode: 0, Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func FPMContainerName(version string) string {
	mm, _ := NormalizeVersion(version)
	return "chauf-php" + strings.ReplaceAll(mm, ".", "") + "-fpm"
}

func DedicatedContainerName(slug string) string { return "chauf-cfpm-" + slug }

func NormalizeVersion(version string) (string, error) {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("invalid PHP version %q", version)
	}
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid PHP version %q", version)
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return "", fmt.Errorf("invalid PHP version %q", version)
			}
		}
	}
	return parts[0] + "." + parts[1], nil
}

func scopeContainerName(scope Scope) string {
	if scope.Service == "nginx" {
		return "chauf-nginx"
	}
	if scope.Dedicated && scope.Project != "" {
		slug := scope.Slug
		if slug == "" {
			slug = filepath.Base(scope.Project)
		}
		return DedicatedContainerName(slug)
	}
	return FPMContainerName(scope.Version)
}

// ContainerName exposes the stable resource name used by lifecycle operations.
func ContainerName(scope Scope) string { return scopeContainerName(scope) }
