package runtime

import (
	"context"
	"testing"
)

func TestPodmanPreflightRejectsRootfulRuntime(t *testing.T) {
	runner := &sequenceRunner{results: []CommandResult{
		{ExitCode: 0, Stdout: "5.0.0"},
		{ExitCode: 0, Stdout: "false"},
	}}
	err := (Podman{Runner: runner}).Preflight(context.Background())
	if err == nil || runner.calls != 2 {
		t.Fatalf("err = %v, calls = %d; want rootless failure after two checks", err, runner.calls)
	}
}

func TestExecRunnerReturnsExitStatusForCommandFailure(t *testing.T) {
	result, err := (ExecRunner{}).Run(context.Background(), "version", "--invalid-chauffeur-test-flag")
	if err != nil || result.ExitCode == 0 {
		t.Fatalf("result = %#v, err = %v; want command exit status", result, err)
	}
}

type sequenceRunner struct {
	results []CommandResult
	calls   int
}

func (r *sequenceRunner) Run(_ context.Context, _ ...string) (CommandResult, error) {
	result := r.results[r.calls]
	r.calls++
	return result, nil
}
