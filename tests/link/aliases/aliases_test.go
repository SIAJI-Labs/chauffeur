package linkaliases_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers" // Add this import
)

func TestLinkWithMultipleAliases(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "multi-domain-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test linking with multiple aliases
	args := []string{
		"--site", "main.test",
		"--alias", "admin.test",
		"--alias", "api.test",
		"--alias", "cdn.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "multi-domain-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)

	// Check primary domain
	if !strings.Contains(content, "domain: main.test") {
		t.Fatalf("expected primary domain main.test in project config:\n%s", content)
	}

	// Check all aliases are present
	expectedAliases := []string{"admin.test", "api.test", "cdn.test"}
	for _, alias := range expectedAliases {
		if !strings.Contains(content, "domain: "+alias) {
			t.Fatalf("expected alias %s in project config:\n%s", alias, content)
		}
	}

	// Verify domains structure
	if !strings.Contains(content, "domains:") {
		t.Fatalf("expected domains section in project config:\n%s", content)
	}

	if !strings.Contains(content, "aliases:") {
		t.Fatalf("expected aliases section in project config:\n%s", content)
	}
}

func TestLinkWithAliasesAndSSL(t *testing.T) {
	helpers.MockAllExecutors(t) // Setup all necessary mocks

	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "secure-multi-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test linking with aliases and SSL
	args := []string{
		"--site", "app.test",
		"--ssl", // This should enable SSL for primary domain
		"--alias", "admin.test",
		"--alias", "api.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "secure-multi-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)

	// Check primary domain has SSL enabled
	if !strings.Contains(content, "ssl: true") {
		t.Fatalf("expected primary domain SSL enabled in project config:\n%s", content)
	}

	// Check aliases are present
	if !strings.Contains(content, "domain: admin.test") {
		t.Fatalf("expected admin.test alias in project config:\n%s", content)
	}

	if !strings.Contains(content, "domain: api.test") {
		t.Fatalf("expected api.test alias in project config:\n%s", content)
	}

	// Verify SSL certificates were generated
	certPath := filepath.Join(workspace, "nginx", "certs", "app.test.crt")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Fatalf("expected SSL certificate to be generated at %s", certPath)
	}

	keyPath := filepath.Join(workspace, "nginx", "certs", "app.test.key")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatalf("expected SSL private key to be generated at %s", keyPath)
	}
}



func TestLinkSingleDomainBackwardCompatibility(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "legacy-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test single domain linking (backward compatibility)
	args := []string{
		"--site", "legacy.test",
		"--ssl",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "legacy-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)

	// Check primary domain
	if !strings.Contains(content, "domain: legacy.test") {
		t.Fatalf("expected primary domain legacy.test in project config:\n%s", content)
	}

	// Should NOT have domains section for single domain (backward compatibility)
	if strings.Contains(content, "domains:") {
		t.Fatalf("unexpected domains section in single-domain project config:\n%s", content)
	}

	// Check SSL is enabled
	if !strings.Contains(content, "ssl: true") {
		t.Fatalf("expected SSL enabled in project config:\n%s", content)
	}
}

func TestLinkAliasDomainValidation(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "validation-test")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test invalid alias domain (not .test)
	args := []string{
		"--site", "valid.test",
		"--alias", "invalid.com", // Should fail validation
	}

	err := commands.RunLink(args)
	if err == nil {
		t.Fatalf("expected RunLink to fail with invalid alias domain")
	}

	expectedError := "domain must end with .test"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected error containing %s, got: %v", expectedError, err)
	}
}

func TestLinkWithCustomPHPVersionAndAliases(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.1")

	projectDir := filepath.Join(home, "custom-php-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test linking with custom PHP version and aliases
	args := []string{
		"--site", "custom.test",
		"--php", "8.1",
		"--alias", "admin.test",
		"--alias", "api.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "custom-php-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)

	// Check PHP version
	if !strings.Contains(content, `php: "8.1"`) && !strings.Contains(content, "php: 8.1") {
		t.Fatalf("expected PHP version 8.1 in project config:\n%s", content)
	}

	// Check aliases
	if !strings.Contains(content, "domain: admin.test") {
		t.Fatalf("expected admin.test alias in project config:\n%s", content)
	}

	if !strings.Contains(content, "domain: api.test") {
		t.Fatalf("expected api.test alias in project config:\n%s", content)
	}
}

func TestLinkWithDedicatedFPMAndAliases(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	projectDir := filepath.Join(home, "dedicated-fpm-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	helpers.Chdir(t, projectDir)

	// Test linking with dedicated FPM and aliases
	args := []string{
		"--site", "dedicated.test",
		"--dedicated-fpm",
		"--alias", "worker.test",
		"--alias", "queue.test",
	}

	if err := commands.RunLink(args); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	configPath := filepath.Join(workspace, "projects", "dedicated-fpm-app", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}

	content := string(data)

	// Check dedicated FPM is enabled
	if !strings.Contains(content, "dedicated: true") {
		t.Fatalf("expected dedicated FPM enabled in project config:\n%s", content)
	}

	// Check project-specific socket path
	expectedSocketPath := filepath.Join(workspace, "projects", "dedicated-fpm-app", "runtime", "php-fpm", "php-fpm.sock")
	if !strings.Contains(content, expectedSocketPath) {
		t.Fatalf("expected project-specific socket path in project config:\n%s", content)
	}

	// Check aliases
	if !strings.Contains(content, "domain: worker.test") {
		t.Fatalf("expected worker.test alias in project config:\n%s", content)
	}

	if !strings.Contains(content, "domain: queue.test") {
		t.Fatalf("expected queue.test alias in project config:\n%s", content)
	}
}