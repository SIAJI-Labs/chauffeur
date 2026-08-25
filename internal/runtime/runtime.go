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
	"strings"
)

type Engine string

const (
	EngineNative Engine = "native"
	EnginePodman Engine = "podman"
)

type Scope struct {
	Workspace string
	Project   string
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
}

type Runtime interface {
	Ensure(ctx context.Context, version string) error
	EnsureProject(ctx context.Context, scope Scope, image string, roots map[string]string) error
	EnsureLinkedProject(ctx context.Context, workspace, configPath, certificateDir string, httpPort, httpsPort int, project ProjectSpec) error
	EnsureWorkspace(ctx context.Context, scope WorkspaceScope) error
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
		if err := p.EnsurePHPContainerWithRoots(ctx, container.Scope, container.Image, container.Roots); err != nil {
			return err
		}
	}
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS(workspace.Routes, 8080, workspace.HTTPSPort)
	if err != nil {
		return err
	}
	if err := os.WriteFile(workspace.ConfigPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("write Podman nginx config: %w", err)
	}
	return p.EnsureNginxContainerWithRootsAndPorts(ctx, workspace.ConfigPath, workspace.Roots, workspace.HTTPPort, workspace.HTTPSPort, workspace.CertificateDir)
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
	if _, err := InspectImage(ctx, p.Runner, image); err != nil {
		return fmt.Errorf("PHP %s image metadata is unavailable: %w", scope.Version, err)
	}
	name := scopeContainerName(scope)
	if result, err := p.run(ctx, "container", "exists", name); err != nil {
		return err
	} else if result.ExitCode == 0 {
		state, matches, err := p.inspectPHPContainer(ctx, name, scope, roots)
		if err != nil {
			return err
		}
		if !matches {
			if state == "running" {
				return fmt.Errorf("PHP container %s has incorrect labels or mounts while running; stop it and retry to reconcile the selected project runtime", name)
			}
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
	args := []string{"run", "-d", "--name", name, "--network", "chauf-net", "--label", "com.siegg.chauffeur.role=php-fpm", "--label", "com.siegg.chauffeur.php.version=" + scope.Version, "--label", "com.siegg.chauffeur.scope=" + scopeKind}
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
	Mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
}

func (p Podman) inspectPHPContainer(ctx context.Context, name string, scope Scope, roots map[string]string) (string, bool, error) {
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
	scopeKind := "shared"
	if scope.Dedicated {
		scopeKind = "dedicated"
	}
	if entry.Config.Labels["com.siegg.chauffeur.role"] != "php-fpm" ||
		entry.Config.Labels["com.siegg.chauffeur.php.version"] != scope.Version ||
		entry.Config.Labels["com.siegg.chauffeur.scope"] != scopeKind || len(entry.Mounts) != len(roots) {
		return entry.State.Status, false, nil
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
	return p.StartContainer(ctx, scopeContainerName(scope))
}

func (p Podman) StartContainer(ctx context.Context, name string) error {
	result, err := p.run(ctx, "container", "start", name)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("start container %s failed with status %d", name, result.ExitCode)
	}
	return nil
}

func (p Podman) Stop(ctx context.Context, scope Scope) error {
	return p.StopContainer(ctx, scopeContainerName(scope))
}

func (p Podman) StopContainer(ctx context.Context, name string) error {
	result, err := p.run(ctx, "container", "stop", name)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("stop container %s failed with status %d", name, result.ExitCode)
	}
	return nil
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
	result, err := p.run(ctx, "container", "inspect", name, "--format", "{{.State.Status}}")
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return []ServiceStatus{{Name: name, Label: name, Role: role, State: "absent", Healthy: false, Container: name, Evidence: strings.TrimSpace(result.Stderr)}}, nil
	}
	state := strings.TrimSpace(result.Stdout)
	return []ServiceStatus{{Name: name, Label: name, Role: role, State: state, Healthy: state == "running", Container: name, Evidence: "container inspect state"}}, nil
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
	ExecFunc func(context.Context, Scope, []string, ExecOptions) (CommandResult, error)
}

func (n Native) Ensure(context.Context, string) error { return nil }
func (n Native) EnsureProject(context.Context, Scope, string, map[string]string) error {
	return nil
}
func (n Native) EnsureLinkedProject(context.Context, string, string, string, int, int, ProjectSpec) error {
	return nil
}
func (n Native) RemoveImage(context.Context, string, bool) error {
	return fmt.Errorf("native runtime does not manage images")
}
func (n Native) EnsureWorkspace(context.Context, WorkspaceScope) error  { return nil }
func (n Native) Start(context.Context, Scope) error                     { return nil }
func (n Native) Stop(context.Context, Scope) error                      { return nil }
func (n Native) Restart(context.Context, Scope) error                   { return nil }
func (n Native) Status(context.Context, Scope) ([]ServiceStatus, error) { return nil, nil }
func (n Native) Logs(context.Context, Scope, LogOptions) (string, error) {
	return "", fmt.Errorf("native runtime log adapter is not configured")
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
	cmd.Env = opts.Env
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
		return DedicatedContainerName(filepath.Base(scope.Project))
	}
	return FPMContainerName(scope.Version)
}

// ContainerName exposes the stable resource name used by lifecycle operations.
func ContainerName(scope Scope) string { return scopeContainerName(scope) }
