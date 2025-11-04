package commands

import "strings"

var cliVersion = "dev"

// SetCLIVersion configures the running CLI version for commands that need it.
func SetCLIVersion(v string) {
	if trimmed := strings.TrimSpace(v); trimmed != "" {
		cliVersion = trimmed
	}
}

// getCLIVersion returns the current CLI version string.
func getCLIVersion() string {
	return cliVersion
}
