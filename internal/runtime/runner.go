package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type FailureKind string

const (
	FailureExecutable FailureKind = "executable"
	FailureCommand    FailureKind = "command"
	FailureTimeout    FailureKind = "timeout"
)

type RunnerError struct {
	Kind FailureKind
	Err  error
}

func (e *RunnerError) Error() string { return e.Err.Error() }
func (e *RunnerError) Unwrap() error { return e.Err }

// ExecRunner is the production Podman command backend. It is intentionally
// small so tests can replace it with a recording fake.
type ExecRunner struct{}

// Stream runs a long-lived command without buffering its output. It is used
// by follow-mode logs so callers see lines as Podman emits them.
func (ExecRunner) Stream(ctx context.Context, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "podman", args...)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &RunnerError{Kind: FailureCommand, Err: fmt.Errorf("podman %s: %w", strings.Join(args, " "), err)}
	}
	return nil
}

func (ExecRunner) Run(ctx context.Context, args ...string) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, "podman", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		code := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		if code >= 0 {
			return CommandResult{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}, nil
		}
		kind := FailureExecutable
		if ctx.Err() != nil {
			kind = FailureTimeout
		}
		return CommandResult{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}, &RunnerError{Kind: kind, Err: fmt.Errorf("podman %s: %w", strings.Join(args, " "), err)}
	}
	return CommandResult{ExitCode: 0, Stdout: stdout.String(), Stderr: stderr.String()}, nil
}
