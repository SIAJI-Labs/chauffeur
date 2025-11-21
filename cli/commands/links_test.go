package commands

import (
	"bytes"
	"fmt" // Added for fmt.Sprintf
	"os"
	"path/filepath"
	"strings" // Added for strings.Split and strings.TrimSpace
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/lib"
)

// setupMockWorkspace creates a temporary workspace and mock project configs
func setupMockWorkspace(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir) // Isolate HOME for workspace

	mockChauffeurDir := filepath.Join(tmpDir, ".chauffeur")
	mockProjectsDir := filepath.Join(mockChauffeurDir, "projects")
	mockConfigPath := filepath.Join(mockChauffeurDir, "config", "chauffeur.yaml")

	// Ensure directories exist
	assert.NoError(t, os.MkdirAll(filepath.Join(mockChauffeurDir, "config"), 0o755))
	assert.NoError(t, os.MkdirAll(mockProjectsDir, 0o755))

	// Write mock global config
	mockGlobalConfig := config.Config{
		ProjectsDir: mockProjectsDir,
	}
	data, err := yaml.Marshal(mockGlobalConfig)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(mockConfigPath, data, 0o644))

	// Create mock projects
	createMockProject(t, mockProjectsDir, "test-project-one", projects.Config{
		Path: "/path/to/project-one",
		PHP:  "8.3",
		Site: &projects.Site{Domain: "one.test", SSL: true},
		Domains: &projects.Domains{
			Aliases: []projects.DomainAlias{
				{Domain: "www.one.test", SSL: true},
				{Domain: "api.one.test", SSL: false},
			},
		},
		Runtime: projects.Runtime{
			FPM: &projects.FPM{Dedicated: false, Socket: ""}, // Shared FPM
		},
		CreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
	})

	createMockProject(t, mockProjectsDir, "test-project-two", projects.Config{
		Path: "/path/to/project-two",
		PHP:  "7.4",
		Site: &projects.Site{Domain: "two.test", SSL: false},
		Domains: &projects.Domains{
			Aliases: []projects.DomainAlias{
				{Domain: "admin.two.test", SSL: true},
			},
		},
		Runtime: projects.Runtime{
			FPM: &projects.FPM{Dedicated: true, Socket: "/tmp/two-fpm.sock"}, // Dedicated FPM
		},
		CreatedAt: time.Date(2023, 2, 1, 11, 0, 0, 0, time.UTC),
	})

	createMockProject(t, mockProjectsDir, "test-project-three", projects.Config{
		Path: "/path/to/project-three",
		PHP:  "8.1",
		Site: &projects.Site{Domain: "three.test", SSL: true},
		Domains: &projects.Domains{
			Aliases: []projects.DomainAlias{}, // No aliases
		},
		Runtime: projects.Runtime{
			FPM: &projects.FPM{Dedicated: false, Socket: ""}, // Shared FPM
		},
		CreatedAt: time.Date(2023, 3, 1, 12, 0, 0, 0, time.UTC),
	})

	return mockProjectsDir, func() {
		// Teardown is handled by t.TempDir()
	}
}

// createMockProject helper to write a project config
func createMockProject(t *testing.T, projectsDir, slug string, cfg projects.Config) {
	projectRoot := filepath.Join(projectsDir, slug)
	assert.NoError(t, os.MkdirAll(projectRoot, 0o755))
	configPath := filepath.Join(projectRoot, "project.yaml")

	// Ensure default values for version and CreatedAt
	if cfg.Version == 0 {
		cfg.Version = projects.ConfigVersion
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now().UTC()
	}

	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(configPath, data, 0o644))

	// Ensure runtime dir for socket lookup
	runtimeDir := filepath.Join(projectRoot, "runtime", "php-fpm")
	assert.NoError(t, os.MkdirAll(runtimeDir, 0o755))
}

// getExpectedFPMSocketPath constructs the expected FPM socket path for a given slug
func getExpectedFPMSocketPath(baseProjectsDir, slug string) string {
	// mockProjectsDir is already filepath.Join(tmpDir, ".chauffeur", "projects")
	// So we need to use baseProjectsDir directly and then append slug
	return filepath.Join(baseProjectsDir, slug, "runtime", "php-fpm", "php-fpm.sock")
}

// TestRunLinks_DefaultOutput tests the default table output of `chauf links`
func TestRunLinks_DefaultOutput(t *testing.T) {
	_, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout) // Restore default output

	err := RunLinks([]string{})
	assert.NoError(t, err)

	output := buf.String()
	// t.Logf("Raw Output for Default Links:\n%q", output) // Debug logging

	expectedLines := []string{
		"[ links ] Linked Projects (3)",
		"[ links ] SLUG                PATH                    DOMAIN                   ALIASES   SSL  PHP   CREATED",
		"[ links ] ------------------  ----------------------  -----------------------  --------  ---  ----  -------------------",
		"[ links ] test-project-one    /path/to/project-one    one.test                 2         *    8.3   2023-01-01 10:00",
		"[ links ] test-project-three  /path/to/project-three  three.test               0         *    8.1   2023-03-01 12:00",
		"[ links ] test-project-two    /path/to/project-two    two.test                 1         *    7.4   2023-02-01 11:00",
	}

	actualLines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Len(t, actualLines, len(expectedLines), "Expected number of lines in output mismatch")

	for i, expected := range expectedLines {
		assert.Equal(t, expected, actualLines[i], "Line %d mismatch", i)
	}
}

// TestRunLinks_DetailOutput_Slug tests detailed output using --slug
func TestRunLinks_DetailOutput_Slug(t *testing.T) {
	mockProjectsDir, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	err := RunLinks([]string{"--slug", "test-project-one"})
	assert.NoError(t, err)

	output := buf.String()
	expectedFPMSocketOne := getExpectedFPMSocketPath(mockProjectsDir, "test-project-one")

	assert.Contains(t, output, "Project: test-project-one")
	assert.Contains(t, output, "  Slug:        test-project-one")
	assert.Contains(t, output, "  Path:        /path/to/project-one")
	assert.Contains(t, output, "  PHP Version: 8.3")
	assert.Contains(t, output, "  Created At:  2023-01-01 10:00:00 UTC")
	assert.Contains(t, output, "  Primary Domain: one.test (SSL Enabled)")
	assert.Contains(t, output, "  Alias Domains:")
	assert.Contains(t, output, "    - www.one.test (SSL Enabled)")
	assert.Contains(t, output, "    - api.one.test (No SSL)")
	assert.Contains(t, output, "  PHP-FPM:     Shared")
	assert.Contains(t, output, fmt.Sprintf("  FPM Socket:  %s", expectedFPMSocketOne))
}

// TestRunLinks_DetailOutput_Site tests detailed output using --site
func TestRunLinks_DetailOutput_Site(t *testing.T) {
	mockProjectsDir, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	// Test with an alias domain
	err := RunLinks([]string{"--site", "api.one.test"})
	assert.NoError(t, err)

	output := buf.String()
	expectedFPMSocketOne := getExpectedFPMSocketPath(mockProjectsDir, "test-project-one")

	assert.Contains(t, output, "Project: test-project-one")
	assert.Contains(t, output, "  Slug:        test-project-one")
	assert.Contains(t, output, "  Path:        /path/to/project-one")
	assert.Contains(t, output, "  PHP Version: 8.3")
	assert.Contains(t, output, "  Created At:  2023-01-01 10:00:00 UTC")
	assert.Contains(t, output, "  Primary Domain: one.test (SSL Enabled)")
	assert.Contains(t, output, "  Alias Domains:")
	assert.Contains(t, output, "    - www.one.test (SSL Enabled)")
	assert.Contains(t, output, "    - api.one.test (No SSL)")
	assert.Contains(t, output, "  PHP-FPM:     Shared")
	assert.Contains(t, output, fmt.Sprintf("  FPM Socket:  %s", expectedFPMSocketOne))

	buf.Reset() // Reset buffer for next test in same function

	// Test with primary domain, no aliases
	err = RunLinks([]string{"--site", "three.test"})
	assert.NoError(t, err)

	output = buf.String()
	expectedFPMSocketThree := getExpectedFPMSocketPath(mockProjectsDir, "test-project-three")

	assert.Contains(t, output, "Project: test-project-three")
	assert.Contains(t, output, "  Slug:        test-project-three")
	assert.Contains(t, output, "  Path:        /path/to/project-three")
	assert.Contains(t, output, "  PHP Version: 8.1")
	assert.Contains(t, output, "  Created At:  2023-03-01 12:00:00 UTC")
	assert.Contains(t, output, "  Primary Domain: three.test (SSL Enabled)")
	assert.Contains(t, output, "  Alias Domains: None")
	assert.Contains(t, output, "  PHP-FPM:     Shared")
	assert.Contains(t, output, fmt.Sprintf("  FPM Socket:  %s", expectedFPMSocketThree))
}

// TestRunLinks_DetailOutput_NotFound tests error for non-existent slug/site
func TestRunLinks_DetailOutput_NotFound(t *testing.T) {
	_, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	err := RunLinks([]string{"--slug", "non-existent-project"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project with slug 'non-existent-project' not found")

	err = RunLinks([]string{"--site", "non-existent.test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project with site 'non-existent.test' not found")
}

// TestRunLinks_MutuallyExclusiveFlags tests that --slug and --site are mutually exclusive
func TestRunLinks_MutuallyExclusiveFlags(t *testing.T) {
	_, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	err := RunLinks([]string{"--slug", "test-project-one", "--site", "one.test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flags --slug and --site are mutually exclusive")
}

// TestRunLinks_HelpCommand tests that --help works correctly
func TestRunLinks_HelpCommand(t *testing.T) {
	_, teardown := setupMockWorkspace(t)
	defer teardown()

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	err := RunLinks([]string{"--help"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Chauffeur Project Listing")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "  chauf links                        List all registered projects and their configurations.") // Updated line
	assert.Contains(t, output, "  chauf links --slug <project-slug>")
	assert.Contains(t, output, "  chauf links --site <domain>")
	assert.Contains(t, output, "Flags:")
	assert.Contains(t, output, "--slug string")
	assert.Contains(t, output, "--site string")
}

// TestRunLinks_NoProjects tests behavior when no projects are linked
func TestRunLinks_NoProjects(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mockChauffeurDir := filepath.Join(tmpDir, ".chauffeur")
	mockConfigPath := filepath.Join(mockChauffeurDir, "config", "chauffeur.yaml")
	assert.NoError(t, os.MkdirAll(filepath.Join(mockChauffeurDir, "config"), 0o755))

	// Write mock global config
	mockGlobalConfig := config.Config{
		ProjectsDir: filepath.Join(mockChauffeurDir, "projects"),
	}
	data, err := yaml.Marshal(mockGlobalConfig)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(mockConfigPath, data, 0o644))

	var buf bytes.Buffer
	lib.SetOutput(&buf)
	defer lib.SetOutput(os.Stdout)

	err = RunLinks([]string{})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No projects linked yet.")
	assert.Contains(t, output, "Use 'chauf link' in a project directory to register it.")
}
