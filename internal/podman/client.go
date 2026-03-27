package podman

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// PodmanClient wraps podman CLI invocations.
type PodmanClient struct{}

// NewPodmanClient creates a new PodmanClient.
func NewPodmanClient() *PodmanClient {
	return &PodmanClient{}
}

// Available checks if podman is installed and functional.
func (c *PodmanClient) Available(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "podman", "version", "--format", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "executable file not found") ||
			strings.Contains(err.Error(), "executable file not found") {
			return ErrPodmanNotFound
		}
		return fmt.Errorf("podman version: %w", err)
	}
	return nil
}

// Run executes a podman command with the given arguments.
func (c *PodmanClient) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "podman", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil // Explicitly don't inherit stdin

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("podman %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// runLines executes a podman command and returns lines of output.
func (c *PodmanClient) runLines(ctx context.Context, args ...string) ([]string, error) {
	output, err := c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	lines := strings.Split(output, "\n")
	return lines, nil
}

// RunWithStdin executes a podman command with stdin data.
func (c *PodmanClient) RunWithStdin(ctx context.Context, args []string, stdinData []byte) (string, error) {
	cmd := exec.CommandContext(ctx, "podman", args...)
	cmd.Stdin = bytes.NewReader(stdinData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("podman %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// ContainerExists checks if a container with the given name exists.
func (c *PodmanClient) ContainerExists(ctx context.Context, name string) (bool, error) {
	_, err := c.Run(ctx, "container", "exists", name)
	if err != nil {
		// podman container exists returns "container does not exist" or just exit 1
		errMsg := err.Error()
		if strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "exit status 1") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// NetworkExists checks if a network with the given name exists.
func (c *PodmanClient) NetworkExists(ctx context.Context, name string) (bool, error) {
	_, err := c.Run(ctx, "network", "exists", name)
	if err != nil {
		// podman network exists returns exit 1 + no stderr message when network doesn't exist.
		// Also handle explicit "network not found" message from some podman versions.
		errMsg := err.Error()
		if strings.Contains(errMsg, "network not found") ||
			strings.Contains(errMsg, "exit status 1") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// VolumeExists checks if a volume with the given name exists.
func (c *PodmanClient) VolumeExists(ctx context.Context, name string) (bool, error) {
	_, err := c.Run(ctx, "volume", "exists", name)
	if err != nil {
		// podman volume exists returns exit 1 + no stderr when volume doesn't exist.
		errMsg := err.Error()
		if strings.Contains(errMsg, "volume does not exist") ||
			strings.Contains(errMsg, "exit status 1") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
