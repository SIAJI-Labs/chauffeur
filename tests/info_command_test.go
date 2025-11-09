package tests

import (
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInfoDisplaysWorkspaceAndVersions(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := commands.DefaultConfig()
	require.NoError(t, err)
	require.NoError(t, config.Save(cfg))

	commands.SetBuildTimestamp("test-build")
	t.Cleanup(func() { commands.SetBuildTimestamp("unknown") })

	commands.OverrideLatestVersionFetcher(func() (string, error) {
		return "v9.9.9", nil
	})
	defer commands.OverrideLatestVersionFetcher(nil)

	output := captureOutput(func() error {
		return commands.RunInfo(nil)
	})

	assert.Contains(t, output, "Workspace:", "should show workspace path")
	assert.Contains(t, output, "Latest release: v9.9.9", "should report mocked remote version")
	assert.Contains(t, output, "Build timestamp: test-build", "should display build timestamp")
}
