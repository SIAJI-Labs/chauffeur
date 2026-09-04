package runtime

import (
	"context"
	"testing"
)

type operationTestRuntime struct {
	statuses map[string]ServiceStatus
	errors   map[string]error
	calls    []string
}

func (r *operationTestRuntime) key(scope Scope) string {
	if scope.Service != "" {
		return scope.Service
	}
	return scope.Version
}
func (r *operationTestRuntime) Ensure(context.Context, string) error { return nil }
func (r *operationTestRuntime) EnsureProject(context.Context, Scope, string, map[string]string) error {
	return nil
}
func (r *operationTestRuntime) EnsureLinkedProject(context.Context, string, string, string, int, int, ProjectSpec) error {
	return nil
}
func (r *operationTestRuntime) EnsureWorkspace(context.Context, WorkspaceScope) error { return nil }
func (r *operationTestRuntime) RemoveProject(context.Context, Scope) error            { return nil }
func (r *operationTestRuntime) RemoveImage(context.Context, string, bool) error       { return nil }
func (r *operationTestRuntime) Start(_ context.Context, scope Scope) error {
	r.calls = append(r.calls, "start:"+r.key(scope))
	return r.errors[r.key(scope)]
}
func (r *operationTestRuntime) Stop(_ context.Context, scope Scope) error {
	r.calls = append(r.calls, "stop:"+r.key(scope))
	return r.errors[r.key(scope)]
}
func (r *operationTestRuntime) Restart(_ context.Context, scope Scope) error {
	r.calls = append(r.calls, "restart:"+r.key(scope))
	return r.errors[r.key(scope)]
}
func (r *operationTestRuntime) Status(_ context.Context, scope Scope) ([]ServiceStatus, error) {
	status := r.statuses[r.key(scope)]
	return []ServiceStatus{status}, nil
}
func (r *operationTestRuntime) Logs(context.Context, Scope, LogOptions) (string, error) {
	return "", nil
}
func (r *operationTestRuntime) Exec(context.Context, Scope, []string, ExecOptions) (CommandResult, error) {
	return CommandResult{}, nil
}

func TestExecuteOperationPreservesOrderAndPartialFailure(t *testing.T) {
	runtime := &operationTestRuntime{
		statuses: map[string]ServiceStatus{"8.3": {Label: "php-fpm 8.3 (shared)", State: "stopped"}, "nginx": {Label: "nginx", State: "stopped"}},
		errors:   map[string]error{"nginx": context.Canceled},
	}
	result := ExecuteOperation(context.Background(), runtime, "start", []Scope{{Version: "8.3"}, {Service: "nginx"}})
	if len(result.Services) != 2 || result.Err == nil {
		t.Fatalf("result = %+v; want two results and aggregate error", result)
	}
	if runtime.calls[0] != "start:8.3" || runtime.calls[1] != "start:nginx" {
		t.Fatalf("calls = %#v; want ordered calls", runtime.calls)
	}
	if result.Services[1].Err != context.Canceled {
		t.Fatalf("service error = %v", result.Services[1].Err)
	}
}
