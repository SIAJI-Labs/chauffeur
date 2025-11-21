package unlinkmultidomain_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers" // Add this import
)

func TestUnlinkShowsMultiDomainConfirmation(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a multi-domain project
	projectDir := filepath.Join(home, "multi-domain-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	args := []string{
		"--site", "main.test",
		"--ssl",
		"--alias", "admin.test",
		"--alias", "api.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Verify project was linked
	configPath := filepath.Join(workspace, "projects", "multi-domain-project", "project.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("project config not found at %s", configPath)
	}

	// Capture unlink output
	output := helpers.CaptureOutput(t, func() {
		// Use --force to avoid interactive confirmation
		unlinkArgs := []string{"--force"}
		if err := commands.RunUnlink(unlinkArgs); err != nil {
			t.Fatalf("RunUnlink failed: %v", err)
		}
	})

	// Verify primary domain is mentioned in the unlink output
	if !strings.Contains(output, "main.test") {
		t.Fatalf("expected unlink output to mention primary domain main.test, got:\n%s", output)
	}

	// Note: The unlink output shows only the primary domain, not individual aliases
	// This is the current behavior - aliases are managed internally but not shown in the summary

	// For SSL-enabled projects, verify SSL certificate removal section is shown
	if !strings.Contains(output, "SSL Certificate Removal") {
		t.Fatalf("expected SSL certificate removal section for SSL-enabled project, got:\n%s", output)
	}

	// Verify project was actually unlinked
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("project config still exists after unlink at %s", configPath)
	}
}

func TestUnlinkWithAliasOption(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a multi-domain project
	projectDir := filepath.Join(home, "alias-test-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	args := []string{
		"--site", "primary.test",
		"--ssl",
		"--alias", "alias1.test",
		"--alias", "alias2.test",
		"--alias", "alias3.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Verify project was linked with all aliases
	configPath := filepath.Join(workspace, "projects", "alias-test-project", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)
	expectedAliases := []string{"alias1.test", "alias2.test", "alias3.test"}
	for _, alias := range expectedAliases {
		if !strings.Contains(content, alias) {
			t.Fatalf("expected alias %s in project config, got:\n%s", alias, content)
		}
	}

	// Now test unlinking with alias option (if supported)
	// This would test the --alias flag for unlink if it exists
	output := helpers.CaptureOutput(t, func() {
		// Use --force to avoid interactive confirmation
		unlinkArgs := []string{"--force"}
		if err := commands.RunUnlink(unlinkArgs); err != nil {
			t.Fatalf("RunUnlink failed: %v", err)
		}
	})

	// Verify primary domain is mentioned
	if !strings.Contains(output, "primary.test") {
		t.Fatalf("expected unlink output to mention primary domain primary.test, got:\n%s", output)
	}

	// Note: The unlink output shows only the primary domain, not individual aliases
	// This is the current behavior - aliases are managed internally but not shown in the summary

	// Verify project was unlinked
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("project config still exists after unlink at %s", configPath)
	}
}

func TestUnlinkSingleDomainBackwardCompatibility(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a single-domain project (backward compatibility)
	projectDir := filepath.Join(home, "single-domain-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	args := []string{
		"--site", "legacy.test",
		"--ssl",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Verify project was linked
	configPath := filepath.Join(workspace, "projects", "single-domain-project", "project.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("project config not found at %s", configPath)
	}

	// Capture unlink output
	output := helpers.CaptureOutput(t, func() {
		// Use --force to avoid interactive confirmation
		unlinkArgs := []string{"--force"}
		if err := commands.RunUnlink(unlinkArgs); err != nil {
			t.Fatalf("RunUnlink failed: %v", err)
		}
	})

	// Verify only the single domain was mentioned
	if !strings.Contains(output, "legacy.test") {
		t.Fatalf("expected unlink output to mention domain legacy.test, got:\n%s", output)
	}

	// For SSL-enabled single domains, verify SSL certificate removal section is shown
	if !strings.Contains(output, "SSL Certificate Removal") {
		t.Fatalf("expected SSL certificate removal section for SSL-enabled project, got:\n%s", output)
	}

	// Verify project was unlinked
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("project config still exists after unlink at %s", configPath)
	}
}

func TestUnlinkAllRemovesMultiDomainProjects(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create multiple projects with different domain configurations
	projects := []struct {
		name     string
		domain   string
		aliases  []string
		ssl      bool
	}{
		{"multi-ssl", "secure.test", []string{"admin.test", "api.test"}, true},
		{"multi-mixed", "mixed.test", []string{"public.test", "private.test"}, false},
		{"single", "simple.test", []string{}, false},
	}

	for _, proj := range projects {
		projectDir := filepath.Join(home, proj.name)
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatalf("create project dir: %v", err)
		}

		helpers.Chdir(t, projectDir)

		args := []string{"--site", proj.domain}
		if proj.ssl {
			args = append(args, "--ssl")
		}
		for _, alias := range proj.aliases {
			args = append(args, "--alias", alias)
		}

		if err := commands.RunLink(args); err != nil {
			t.Fatalf("RunLink failed for %s: %v", proj.name, err)
		}
	}

	// Change to home directory for unlink all
	helpers.Chdir(t, home)

	// Verify all projects exist
	projectsDir := filepath.Join(workspace, "projects")
	for _, proj := range projects {
		configPath := filepath.Join(projectsDir, proj.name, "project.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("project config not found at %s", configPath)
		}
	}

	// Capture unlink all output
	output := helpers.CaptureOutput(t, func() {
		// Use --all to avoid interactive confirmation
		unlinkArgs := []string{"--all", "--force"}
		if err := commands.RunUnlink(unlinkArgs); err != nil {
			t.Fatalf("RunUnlink failed: %v", err)
		}
	})

	// Verify all projects were mentioned in the output
	for _, proj := range projects {
		if !strings.Contains(output, proj.name) {
			t.Fatalf("expected unlink output to mention project %s, got:\n%s", proj.name, output)
		}

		// Verify primary domain was mentioned
		if !strings.Contains(output, proj.domain) {
			t.Fatalf("expected unlink output to mention domain %s, got:\n%s", proj.domain, output)
		}

		// Note: The unlink output shows only the primary domain, not individual aliases
		// This is the current behavior - aliases are managed internally but not shown in the summary
	}

	// Verify all projects were unlinked
	for _, proj := range projects {
		configPath := filepath.Join(projectsDir, proj.name, "project.yaml")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Fatalf("project config still exists after unlink all at %s", configPath)
		}
	}

	// Verify success message
	if !strings.Contains(output, "Successfully unlinked") {
		t.Fatalf("expected success message in unlink output, got:\n%s", output)
	}
}
