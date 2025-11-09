package commands

import "strings"

var (
	cliVersion     = "dev"
	buildTimestamp = "unknown"
)

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

// SetBuildTimestamp configures the CLI build timestamp for display/logging.
func SetBuildTimestamp(ts string) {
	if trimmed := strings.TrimSpace(ts); trimmed != "" {
		buildTimestamp = trimmed
	}
}

func getBuildTimestamp() string {
	return buildTimestamp
}
