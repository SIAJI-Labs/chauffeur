package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunServiceCommand executes an installed Chauffeur-managed service with passthrough args.
 *
 * @param name Service identifier (e.g., "nginx", "caddy").
 * @param args Arguments forwarded to the underlying binary.
 * @return error when the service is unknown, missing, or execution fails.
 */
func RunServiceCommand(name string, args []string) error {
	if !IsKnownService(name) {
		return fmt.Errorf("unknown service %s", name)
	}

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	info, err := system.Detect()
	if err != nil {
		return err
	}

	spec, err := newServiceSpec(name, prefix, info)
	if err != nil {
		return err
	}

	ok, err := spec.available()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("service %s is not installed; run 'chauf install %s' first", name, name)
	}

	cmd := exec.Command(spec.binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}
