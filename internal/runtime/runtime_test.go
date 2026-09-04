package runtime

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
)

type recordingRunner struct {
	args   [][]string
	result CommandResult
}

func (r *recordingRunner) Run(_ context.Context, args ...string) (CommandResult, error) {
	r.args = append(r.args, args)
	return r.result, nil
}

func TestPodmanExecBuildsArgumentSafeCommand(t *testing.T) {
	runner := &recordingRunner{}
	r := Podman{Runner: runner}
	_, err := r.Exec(context.Background(), Scope{Version: "8.3"}, []string{"php", "-r", "echo 'ok';"}, ExecOptions{Dir: "/project path"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"exec", "--workdir", "/project path", "chauf-php83-fpm", "php", "-r", "echo 'ok';"}
	if len(runner.args) != 1 || len(runner.args[0]) != len(want) {
		t.Fatalf("args = %#v, want %#v", runner.args, want)
	}
	for i := range want {
		if runner.args[0][i] != want[i] {
			t.Fatalf("args = %#v, want %#v", runner.args, want)
		}
	}
}

func TestPodmanExecPropagatesCommandExitCode(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{ExitCode: 7, Stderr: "failed"}}
	result, err := (Podman{Runner: runner}).Exec(context.Background(), Scope{Version: "8.3"}, []string{"composer", "validate"}, ExecOptions{})
	if err == nil || result.ExitCode != 7 {
		t.Fatalf("result = %+v, err = %v; want exit code propagation", result, err)
	}
	var exitCoder interface{ ExitCode() int }
	if !errors.As(err, &exitCoder) || exitCoder.ExitCode() != 7 {
		t.Fatalf("err = %v; want exit code 7", err)
	}
}

func TestNormalizeVersion(t *testing.T) {
	for input, want := range map[string]string{"8.3": "8.3", "8.3.20": "8.3", " 7.4 ": "7.4"} {
		got, err := NormalizeVersion(input)
		if err != nil || got != want {
			t.Errorf("NormalizeVersion(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestDedicatedScopeUsesProjectSlug(t *testing.T) {
	if got := scopeContainerName(Scope{Version: "8.3", Project: "/tmp/my-shop", Dedicated: true}); got != "chauf-cfpm-my-shop" {
		t.Fatalf("container = %q", got)
	}
}

func TestDedicatedScopeUsesExplicitSlug(t *testing.T) {
	if got := scopeContainerName(Scope{Project: "/tmp/my-shop", Slug: "shop-admin", Dedicated: true}); got != "chauf-cfpm-shop-admin" {
		t.Fatalf("container = %q", got)
	}
}

func TestServiceScopeUsesStableNginxContainer(t *testing.T) {
	if got := scopeContainerName(Scope{Service: "nginx"}); got != "chauf-nginx" {
		t.Fatalf("container = %q, want chauf-nginx", got)
	}
}

func TestEnsureWorkspaceValidatesRequiredPaths(t *testing.T) {
	err := (Podman{Runner: &recordingRunner{}}).EnsureWorkspace(context.Background(), WorkspaceScope{})
	if err == nil {
		t.Fatal("expected incomplete workspace scope error")
	}
}

type workspaceValidationRunner struct {
	args [][]string
}

func (r *workspaceValidationRunner) Run(_ context.Context, args ...string) (CommandResult, error) {
	r.args = append(r.args, args)
	switch {
	case len(args) >= 2 && args[0] == "version":
		return CommandResult{ExitCode: 0, Stdout: "5.0"}, nil
	case len(args) >= 2 && args[0] == "info":
		return CommandResult{ExitCode: 0, Stdout: "true"}, nil
	case len(args) >= 3 && args[0] == "image" && args[1] == "exists":
		if strings.Contains(args[2], ":8.0-fpm") {
			return CommandResult{ExitCode: 1}, nil
		}
		return CommandResult{ExitCode: 0}, nil
	case len(args) >= 3 && args[0] == "image" && args[1] == "inspect":
		return CommandResult{ExitCode: 0, Stdout: `[{"Id":"sha256:test","Labels":{"com.siegg.chauffeur.php.version":"8.3"}}]`}, nil
	default:
		return CommandResult{ExitCode: 0}, nil
	}
}

func TestEnsureWorkspaceValidatesAllImagesBeforeCreatingContainers(t *testing.T) {
	runner := &workspaceValidationRunner{}
	one := t.TempDir()
	two := t.TempDir()
	scope := WorkspaceScope{
		Workspace: "/workspace", ConfigPath: "/workspace/nginx.conf", HTTPPort: 0,
		PHPContainers: []PHPContainerScope{
			{Scope: Scope{Version: "8.3", Project: one}, Image: PHPImage("8.3"), Roots: map[string]string{"/workspace/one": one}},
			{Scope: Scope{Version: "8.0", Project: two}, Image: PHPImage("8.0"), Roots: map[string]string{"/workspace/two": two}},
		},
	}
	err := (Podman{Runner: runner}).EnsureWorkspace(context.Background(), scope)
	if err == nil || !strings.Contains(err.Error(), "chauf install php 8.0 --build") {
		t.Fatalf("error = %v; want stale-image recovery command", err)
	}
	for _, args := range runner.args {
		if len(args) >= 2 && args[0] == "container" && args[1] == "run" {
			t.Fatalf("container was created before all images were validated: %#v", runner.args)
		}
	}
}

func TestCommandFailureIncludesPodmanEvidence(t *testing.T) {
	err := commandFailure("create PHP 8.3 container", CommandResult{ExitCode: 125, Stderr: "invalid mount"})
	if !strings.Contains(err.Error(), "status 125") || !strings.Contains(err.Error(), "invalid mount") {
		t.Fatalf("error = %v; want status and Podman stderr", err)
	}
}

func TestStatusReportsMissingImageBeforeContainerState(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{ExitCode: 1}}
	statuses, err := (Podman{Runner: runner}).Status(context.Background(), Scope{Version: "7.4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].State != "image-missing" || statuses[0].Image != PHPImage("7.4") {
		t.Fatalf("statuses = %+v; want missing PHP image state", statuses)
	}
}

func TestEnsureWorkspaceRejectsMissingProjectMountBeforeImages(t *testing.T) {
	runner := &workspaceValidationRunner{}
	err := (Podman{Runner: runner}).EnsureWorkspace(context.Background(), WorkspaceScope{
		Workspace: "/workspace", ConfigPath: "/workspace/nginx.conf", HTTPPort: 0,
		PHPContainers: []PHPContainerScope{{
			Scope: Scope{Version: "8.3", Project: "/deleted/project"}, Image: PHPImage("8.3"),
			Roots: map[string]string{"/workspace/project": "/deleted/project"},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "chauf unlink /deleted/project --yes") {
		t.Fatalf("error = %v; want stale project recovery command", err)
	}
	for _, args := range runner.args {
		if len(args) > 0 && args[0] == "image" {
			t.Fatalf("image validation ran before stale path validation: %#v", runner.args)
		}
	}
}

func TestInspectPHPContainerDetectsStaleMountsAndLabels(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{Stdout: `[{"State":{"Status":"exited"},"Config":{"Labels":{"com.siegg.chauffeur.role":"php-fpm","com.siegg.chauffeur.php.version":"8.3","com.siegg.chauffeur.scope":"shared"}},"Mounts":[{"Source":"/old/project","Destination":"/workspace","RW":true}]}]`}}
	state, matches, err := (Podman{Runner: runner}).inspectPHPContainer(context.Background(), "chauf-php83-fpm", Scope{Version: "8.3", Project: "/new/project"}, map[string]string{"/workspace": "/new/project"})
	if err != nil {
		t.Fatal(err)
	}
	if state != "exited" || matches {
		t.Fatalf("state=%q matches=%t; want exited and stale metadata", state, matches)
	}
}

func TestValidateNginxPortsReportsOccupiedHostPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	err = (Podman{Runner: &recordingRunner{}}).validateNginxPorts(context.Background(), port, 0)
	if err == nil || !strings.Contains(err.Error(), "chauf config nginx http_port") {
		t.Fatalf("error = %v; want occupied-port recovery command", err)
	}
}

func TestParseMemoryMB(t *testing.T) {
	for input, want := range map[string]int{"12MiB / 1GiB": 12, "1.5GB / 2GB": 1536, "512KB / 1GB": 0} {
		if got := parseMemoryMB(input); got != want {
			t.Fatalf("parseMemoryMB(%q) = %d, want %d", input, got, want)
		}
	}
}

type reverseProxyProbeRunner struct{ exit int }

func (r reverseProxyProbeRunner) Run(_ context.Context, args ...string) (CommandResult, error) {
	if len(args) > 0 && args[0] == "exec" {
		return CommandResult{ExitCode: r.exit, Stderr: "connection refused"}, nil
	}
	return CommandResult{ExitCode: 0}, nil
}

func TestValidateReverseProxyRoutesReportsContainerReachability(t *testing.T) {
	route := []NginxRoute{{ServerName: "arvi-ui.test", Upstream: "host.containers.internal", ProxyPort: 3901}}
	err := (Podman{Runner: reverseProxyProbeRunner{exit: 1}}).validateReverseProxyRoutes(context.Background(), route)
	if err == nil || !strings.Contains(err.Error(), "arvi-ui.test") || !strings.Contains(err.Error(), "container-reachable") {
		t.Fatalf("error = %v; want actionable reverse-proxy reachability error", err)
	}
	if err := (Podman{Runner: reverseProxyProbeRunner{exit: 0}}).validateReverseProxyRoutes(context.Background(), route); err != nil {
		t.Fatalf("reachable route returned error: %v", err)
	}
}

func TestStatusContainerIncludesProcessMetadata(t *testing.T) {
	runner := statusMetadataRunner{}
	statuses, err := (Podman{Runner: runner}).StatusContainer(context.Background(), "chauf-nginx", "nginx")
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].PID != 42 || statuses[0].MemoryMB != 12 || !statuses[0].Healthy {
		t.Fatalf("statuses = %+v; want running metadata", statuses)
	}
}

type streamingTestRunner struct{ args []string }

func (r *streamingTestRunner) Run(context.Context, ...string) (CommandResult, error) {
	return CommandResult{}, nil
}
func (r *streamingTestRunner) Stream(_ context.Context, _ io.Writer, _ io.Writer, args ...string) error {
	r.args = args
	return nil
}

func TestStreamLogsUsesFollowAndContainerTarget(t *testing.T) {
	runner := &streamingTestRunner{}
	if err := (Podman{Runner: runner}).StreamLogs(context.Background(), Scope{Service: "nginx"}, LogOptions{Follow: true, Lines: 25}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	want := []string{"logs", "--follow", "--tail", "25", "chauf-nginx"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("args = %#v, want %#v", runner.args, want)
	}
}

type statusMetadataRunner struct{}

func (statusMetadataRunner) Run(_ context.Context, args ...string) (CommandResult, error) {
	if len(args) > 1 && args[0] == "container" && args[1] == "inspect" {
		return CommandResult{ExitCode: 0, Stdout: `[{"State":{"Status":"running","Pid":42,"StartedAt":"2026-08-30T00:00:00Z"},"NetworkSettings":{"Networks":{"chauf-net":{}}}}]`}, nil
	}
	return CommandResult{ExitCode: 0, Stdout: "12MiB / 1GiB"}, nil
}
