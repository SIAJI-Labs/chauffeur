package helpers

import (
	"os"
	"os/exec"
	"testing"

	"github.com/siaji/chauffeur/cli/lib"
)

// TestHelperProcess is a helper function for mocking exec.Command.
// It's not a real test, but a way to intercept exec.Command calls.
// It should be named TestHelperProcess and reside in a file ending with _test.go
// This function needs to be outside any _test function to be discoverable by 'go test'
// This is exposed for external test packages to use in their own mock setup.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		return // Should not happen
	}

	cmd := args[0]
	switch cmd {
	case "sudo":
		// Simulate successful sudo operations for iptables
		// Can add specific argument checks if needed.
		// For example, if we need to mock 'iptables -C' to return an error sometimes,
		// we'd check args[1:] here.
		return
	case "mkcert":
		// Simulate successful mkcert operations (-install, -uninstall)
		// Can add specific argument checks if needed
		return
	default:
		// Default to success for other commands
		return
	}
}

// MockSystemCommandExecutor sets up a mock for lib.CommandExecutor during tests.
// It ensures that `sudo` and `iptables` calls are intercepted by TestHelperProcess.
// Call this at the beginning of a test and defer its cleanup function.
func MockSystemCommandExecutor(t *testing.T) {
	lib.SetCommandExecutor(func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	})
	t.Cleanup(lib.ResetCommandExecutor)
}

// MockMkcertCommandExecutor sets up a mock for lib.MkcertCommandExecutor during tests.
// It ensures that `mkcert` calls are intercepted by TestHelperProcess.
// Call this at the beginning of a test and defer its cleanup function.
func MockMkcertCommandExecutor(t *testing.T) {
	lib.SetMkcertCommandExecutor(func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	})
	t.Cleanup(lib.ResetMkcertCommandExecutor)
}

// MockAllExecutors sets up mocks for all external command executors.
func MockAllExecutors(t *testing.T) {
	MockSystemCommandExecutor(t)
	MockMkcertCommandExecutor(t)
}

