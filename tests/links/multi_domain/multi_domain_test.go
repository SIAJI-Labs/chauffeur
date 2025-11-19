package linksmultidomain_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestLinksShowsMultiDomainProjects(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	projectsDir := filepath.Join(workspace, "projects")
	writeMultiDomainProject := func(slug, path, primaryDomain string, aliases []string, sslDomains map[string]bool) {
		dir := filepath.Join(projectsDir, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir project: %v", err)
		}

		var content strings.Builder
		content.WriteString("version: 1\n")
		content.WriteString("path: " + path + "\n")
		content.WriteString("php: 8.3\n")
		content.WriteString("site:\n")
		content.WriteString("  domain: " + primaryDomain + "\n")

		primarySSL := sslDomains[primaryDomain]
		content.WriteString("  ssl: " + formatBool(primarySSL) + "\n")

		if len(aliases) > 0 {
			content.WriteString("domains:\n")
			content.WriteString("  aliases:\n")
			for _, alias := range aliases {
				aliasSSL := sslDomains[alias]
				content.WriteString("    - domain: " + alias + "\n")
				content.WriteString("      ssl: " + formatBool(aliasSSL) + "\n")
			}
		}

		content.WriteString("runtime:\n")
		content.WriteString("  phpfpm: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\n")
		content.WriteString("  fpm:\n")
		content.WriteString("    dedicated: false\n")
		content.WriteString("    socket: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\n")
		content.WriteString("created_at: 2025-11-09T00:00:00Z\n")

		if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content.String()), 0o644); err != nil {
			t.Fatalf("write project config: %v", err)
		}
	}

	// Create a multi-domain project with SSL on some domains
	writeMultiDomainProject("secure-app", "/path/secure-app", "main.test",
		[]string{"admin.test", "api.test", "public.test"},
		map[string]bool{"main.test": true, "admin.test": true, "api.test": false, "public.test": false})

	// Create a single-domain project (backward compatibility)
	writeMultiDomainProject("simple-app", "/path/simple-app", "simple.test",
		[]string{},
		map[string]bool{"simple.test": false})

	output := helpers.CaptureOutput(t, func() {
		if err := commands.RunLinks(nil); err != nil {
			t.Fatalf("RunLinks failed: %v", err)
		}
	})

	// Verify both projects are listed
	if !strings.Contains(output, "secure-app") {
		t.Fatalf("expected output to list secure-app project, got:\n%s", output)
	}

	if !strings.Contains(output, "simple-app") {
		t.Fatalf("expected output to list simple-app project, got:\n%s", output)
	}

	// Verify domains are shown for multi-domain project
	if !strings.Contains(output, "main.test") {
		t.Fatalf("expected output to show main.test domain, got:\n%s", output)
	}

	if !strings.Contains(output, "admin.test") {
		t.Fatalf("expected output to show admin.test alias, got:\n%s", output)
	}

	if !strings.Contains(output, "api.test") {
		t.Fatalf("expected output to show api.test alias, got:\n%s", output)
	}

	if !strings.Contains(output, "public.test") {
		t.Fatalf("expected output to show public.test alias, got:\n%s", output)
	}

	// Verify SSL indicators are shown for SSL-enabled aliases (look for (*) indicators in ALIAS column)
	if !strings.Contains(output, "admin.test (*)") {
		t.Fatalf("expected output to show SSL indicator for admin.test, got:\n%s", output)
	}

	// Verify non-SSL domains don't have indicators
	if strings.Contains(output, "api.test (*)") {
		t.Fatalf("unexpected SSL indicator for api.test (should be non-SSL), got:\n%s", output)
	}

	if strings.Contains(output, "public.test (*)") {
		t.Fatalf("unexpected SSL indicator for public.test (should be non-SSL), got:\n%s", output)
	}

	// Verify primary domain SSL is shown in SSL column (* indicates SSL enabled)
	if !strings.Contains(output, "secure-app") {
		t.Fatalf("expected project to be listed, got:\n%s", output)
	}

	// Check that the SSL column shows * for SSL-enabled project
	lines := strings.Split(output, "\n")
	secureAppFound := false
	for _, line := range lines {
		if strings.Contains(line, "secure-app") {
			secureAppFound = true
			// Should have * in SSL column
			if !strings.Contains(line, "  *  ") {
				t.Fatalf("expected SSL indicator (*) in SSL column for secure-app, got line: %s", line)
			}
			break
		}
	}
	if !secureAppFound {
		t.Fatalf("secure-app project not found in output:\n%s", output)
	}
}

func TestLinksShowsOnlySSLIndicatorsForSSLEnabledDomains(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	projectsDir := filepath.Join(workspace, "projects")
	writePartialSSLProject := func(slug, path, primaryDomain string, aliases []string, sslDomains map[string]bool) {
		dir := filepath.Join(projectsDir, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir project: %v", err)
		}

		var content strings.Builder
		content.WriteString("version: 1\n")
		content.WriteString("path: " + path + "\n")
		content.WriteString("php: 8.3\n")
		content.WriteString("site:\n")
		content.WriteString("  domain: " + primaryDomain + "\n")

		primarySSL := sslDomains[primaryDomain]
		content.WriteString("  ssl: " + formatBool(primarySSL) + "\n")

		if len(aliases) > 0 {
			content.WriteString("domains:\n")
			content.WriteString("  aliases:\n")
			for _, alias := range aliases {
				aliasSSL := sslDomains[alias]
				content.WriteString("    - domain: " + alias + "\n")
				content.WriteString("      ssl: " + formatBool(aliasSSL) + "\n")
			}
		}

		content.WriteString("runtime:\n")
		content.WriteString("  phpfpm: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\n")
		content.WriteString("  fpm:\n")
		content.WriteString("    dedicated: false\n")
		content.WriteString("    socket: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\n")
		content.WriteString("created_at: 2025-11-09T00:00:00Z\n")

		if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content.String()), 0o644); err != nil {
			t.Fatalf("write project config: %v", err)
		}
	}

	// Create a project with mixed SSL settings
	writePartialSSLProject("mixed-ssl", "/path/mixed-ssl", "app.test",
		[]string{"secure.test", "insecure.test", "also-secure.test"},
		map[string]bool{"app.test": true, "secure.test": true, "insecure.test": false, "also-secure.test": true})

	output := helpers.CaptureOutput(t, func() {
		if err := commands.RunLinks(nil); err != nil {
			t.Fatalf("RunLinks failed: %v", err)
		}
	})

	// Verify SSL indicators are only shown for SSL-enabled aliases (in ALIAS column)
	sslAliases := []string{"secure.test", "also-secure.test"}
	for _, alias := range sslAliases {
		if !strings.Contains(output, alias+" (*)") {
			t.Fatalf("expected SSL indicator for %s, got:\n%s", alias, output)
		}
	}

	// Verify non-SSL domain doesn't have SSL indicator
	if strings.Contains(output, "insecure.test (*)") {
		t.Fatalf("unexpected SSL indicator for insecure.test (should be non-SSL), got:\n%s", output)
	}

	if !strings.Contains(output, "insecure.test") {
		t.Fatalf("expected insecure.test to be listed (without SSL indicator), got:\n%s", output)
	}

	// Verify primary domain SSL is shown in SSL column (* indicates SSL enabled)
	lines := strings.Split(output, "\n")
	mixedSslFound := false
	for _, line := range lines {
		if strings.Contains(line, "mixed-ssl") {
			mixedSslFound = true
			// Should have * in SSL column
			if !strings.Contains(line, "  *  ") {
				t.Fatalf("expected SSL indicator (*) in SSL column for mixed-ssl, got line: %s", line)
			}
			break
		}
	}
	if !mixedSslFound {
		t.Fatalf("mixed-ssl project not found in output:\n%s", output)
	}
}

func TestLinksEmptyMultiDomainAliases(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	projectsDir := filepath.Join(workspace, "projects")
	writeEmptyAliasProject := func(slug, path, primaryDomain string) {
		dir := filepath.Join(projectsDir, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir project: %v", err)
		}

		content := "version: 1\npath: " + path + "\nphp: 8.3\nsite:\n  domain: " + primaryDomain + "\n  ssl: false\ndomains:\n  aliases: []\nruntime:\n  phpfpm: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\n  fpm:\n    dedicated: false\n    socket: " + workspace + "/php/8.3/runtime/php-fpm/php-fpm.sock\ncreated_at: 2025-11-09T00:00:00Z\n"

		if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content), 0o644); err != nil {
			t.Fatalf("write project config: %v", err)
		}
	}

	// Create a project with empty aliases array
	writeEmptyAliasProject("empty-aliases", "/path/empty-aliases", "primary.test")

	output := helpers.CaptureOutput(t, func() {
		if err := commands.RunLinks(nil); err != nil {
			t.Fatalf("RunLinks failed: %v", err)
		}
	})

	// Should still show the primary domain
	if !strings.Contains(output, "primary.test") {
		t.Fatalf("expected output to show primary.test domain, got:\n%s", output)
	}

	// Should not show any SSL indicator
	if strings.Contains(output, "primary.test (*)") {
		t.Fatalf("unexpected SSL indicator for primary.test (should be non-SSL), got:\n%s", output)
	}
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}