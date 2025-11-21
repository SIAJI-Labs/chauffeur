package lib

import (
	"os/exec"
)

// --- CommandExecutor for system calls ---

var CommandExecutor = exec.Command // Exported variable to allow mocking exec.Command

// SetCommandExecutor is a test helper function to mock exec.Command
func SetCommandExecutor(f func(name string, arg ...string) *exec.Cmd) {
	CommandExecutor = f
}

// ResetCommandExecutor resets the command executor to the default exec.Command
func ResetCommandExecutor() {
	CommandExecutor = exec.Command
}

// --- MkcertCommandExecutor for mkcert calls ---

var MkcertCommandExecutor = exec.Command // Exported variable to allow mocking mkcert exec.Command

// SetMkcertCommandExecutor is a test helper function to mock mkcert exec.Command
func SetMkcertCommandExecutor(f func(name string, arg ...string) *exec.Cmd) {
	MkcertCommandExecutor = f
}

// ResetMkcertCommandExecutor resets the mkcert command executor to the default exec.Command
func ResetMkcertCommandExecutor() {
	MkcertCommandExecutor = exec.Command
}
