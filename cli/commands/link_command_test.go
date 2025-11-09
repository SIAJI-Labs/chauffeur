package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
)

func TestRunLink(t *testing.T) {
	// Setup temporary directory for testing
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	defer os.Chdir(origWd)

	// Set up a test project directory
	testProjectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(testProjectDir, 0o755); err != nil {
		t.Fatalf("create test project dir: %v", err)
	}
	if err := os.Chdir(testProjectDir); err != nil {
		t.Fatalf("chdir to test project: %v", err)
	}

	// Setup workspace directory
	workspaceDir := filepath.Join(tmpDir, ".chauffeur")
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}

	// Create a config file
	_ = config.Config{
		Version: 1,
		PHP: config.PHPConfig{
			Default: "8.3",
		},
		ProjectsDir: projectsDir,
	}
	cfgPath := filepath.Join(workspaceDir, "config", "chauffeur.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	t.Setenv("HOME", tmpDir)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		pattern  string
		validate func(t *testing.T, err error)
	}{
		{
			name:    "basic link without flags",
			args:    []string{},
			wantErr: false,
			pattern: "Project linked as test-project",
		},
		{
			name:    "link with site and ssl flags",
			args:    []string{"--site", "test.example", "--ssl"},
			wantErr: false,
			pattern: "ssl=true",
		},
		{
			name:    "link with php version override",
			args:    []string{"--php", "7.4"},
			wantErr: false,
			pattern: "PHP: 7.4",
		},
		{
			name:    "error when ssl without site",
			args:    []string{"--ssl"},
			wantErr: true,
			pattern: "",
			validate: func(t *testing.T, err error) {
				if err == nil || !contains(err.Error(), "--ssl requires --site") {
					t.Errorf("expected error about --ssl requiring --site, got %v", err)
				}
			},
		},
		{
			name:    "error when unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
			pattern: "",
			validate: func(t *testing.T, err error) {
				if err == nil || !contains(err.Error(), "unknown flag for link") {
					t.Errorf("expected error about unknown flag, got %v", err)
				}
			},
		},
		{
			name:    "error when site value missing",
			args:    []string{"--site"},
			wantErr: true,
			pattern: "",
			validate: func(t *testing.T, err error) {
				if err == nil || !contains(err.Error(), "--site requires a domain value") {
					t.Errorf("expected error about missing site value, got %v", err)
				}
			},
		},
		{
			name:    "error when php value missing",
			args:    []string{"--php"},
			wantErr: true,
			pattern: "",
			validate: func(t *testing.T, err error) {
				if err == nil || !contains(err.Error(), "--php requires a version value") {
					t.Errorf("expected error about missing php value, got %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use a fresh project directory for each test
			testProjectDir := filepath.Join(tmpDir, test.name+"-project")
			if err := os.MkdirAll(testProjectDir, 0o755); err != nil {
				t.Fatalf("create test project dir: %v", err)
			}

			// Use a fresh projects directory for each test
			testProjectsDir := filepath.Join(workspaceDir, test.name+"-projects")
			if err := os.MkdirAll(testProjectsDir, 0o755); err != nil {
				t.Fatalf("create projects dir: %v", err)
			}

			// Update environment for this test
			homeDir := filepath.Join(tmpDir, test.name+"-home")
			t.Setenv("HOME", homeDir)
			ensureMockPHP(t, homeDir, "8.3", "7.4")

			// Create fresh config for this test
			_ = config.Config{
				Version: 1,
				PHP: config.PHPConfig{
					Default: "8.3",
				},
				ProjectsDir: testProjectsDir,
			}
			cfgPath := filepath.Join(filepath.Join(tmpDir, test.name+"-home"), ".chauffeur", "config", "chauffeur.yaml")
			if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
				t.Fatalf("create config dir: %v", err)
			}

			// Change to project directory
			if err := os.Chdir(testProjectDir); err != nil {
				t.Fatalf("chdir to test project: %v", err)
			}

			// Store original directory
			origWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("get current dir: %v", err)
			}
			defer os.Chdir(origWd)

			err = RunLink(test.args)

			if test.wantErr {
				if err == nil {
					t.Errorf("RunLink() expected error but got none")
					return
				}
				if test.validate != nil {
					test.validate(t, err)
				}
				return
			}

			if err != nil {
				t.Errorf("RunLink() unexpected error = %v", err)
				return
			}

			if test.pattern != "" {
				output, err := captureOutput(func() error {
					return RunLink(test.args)
				})
				if err != nil {
					t.Errorf("RunLink() failed to capture output: %v", err)
					return
				}
				if !contains(output, test.pattern) {
					t.Errorf("RunLink() expected output to contain %q, got: %q", test.pattern, output)
				}
			}
		})
	}
}

func TestRunLinks(t *testing.T) {
	// Setup temporary directory for testing
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	defer os.Chdir(origWd)

	// Setup workspace directory
	workspaceDir := filepath.Join(tmpDir, ".chauffeur")
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}

	// Create a config file
	_ = config.Config{
		Version: 1,
		PHP: config.PHPConfig{
			Default: "8.3",
		},
		ProjectsDir: projectsDir,
	}
	cfgPath := filepath.Join(workspaceDir, "config", "chauffeur.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	t.Setenv("HOME", tmpDir)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		pattern  string
		setup    func()
		teardown func()
	}{
		{
			name:    "no projects linked",
			args:    []string{},
			wantErr: false,
			pattern: "No projects linked yet",
		},
		{
			name:    "one project linked",
			args:    []string{},
			wantErr: false,
			pattern: "test-project",
			setup: func() {
				// Create a test project
				testProject := projects.Config{
					Version:   1,
					Path:      "/path/to/test-project",
					PHP:       "8.3",
					Site:      &projects.Site{Domain: "test.test", SSL: false},
					CreatedAt: time.Now().UTC(),
				}
				layout, _ := projects.EnsureLayout(projectsDir, "test-project")
				projects.WriteConfig(testProject, layout.ConfigPath, false)
			},
		},
		{
			name:    "multiple projects linked",
			args:    []string{},
			wantErr: false,
			pattern: "Linked Projects (2)",
			setup: func() {
				// First project
				testProject1 := projects.Config{
					Version:   1,
					Path:      "/path/to/project-one",
					PHP:       "8.3",
					CreatedAt: time.Now().UTC(),
				}
				layout1, _ := projects.EnsureLayout(projectsDir, "project-one")
				projects.WriteConfig(testProject1, layout1.ConfigPath, false)

				// Second project
				testProject2 := projects.Config{
					Version:   1,
					Path:      "/path/to/project-two",
					PHP:       "7.4",
					Site:      &projects.Site{Domain: "two.test", SSL: true},
					CreatedAt: time.Now().UTC(),
				}
				layout2, _ := projects.EnsureLayout(projectsDir, "project-two")
				projects.WriteConfig(testProject2, layout2.ConfigPath, false)
			},
		},
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
			pattern: "Usage:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup()
			}
			if test.teardown != nil {
				defer test.teardown()
			}

			err := RunLinks(test.args)

			if test.wantErr && err == nil {
				t.Errorf("RunLinks() expected error but got none")
				return
			}
			if !test.wantErr && err != nil {
				t.Errorf("RunLinks() unexpected error = %v", err)
				return
			}

			if test.pattern != "" {
				output, err := captureOutput(func() error {
					return RunLinks(test.args)
				})
				if err != nil {
					t.Errorf("RunLinks() failed to capture output: %v", err)
					return
				}
				if !contains(output, test.pattern) {
					t.Errorf("RunLinks() expected output to contain %q, got: %q", test.pattern, output)
				}
			}
		})
	}
}

func TestRunLinkOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	defer os.Chdir(origWd)

	testProjectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(testProjectDir, 0o755); err != nil {
		t.Fatalf("create test project dir: %v", err)
	}
	if err := os.Chdir(testProjectDir); err != nil {
		t.Fatalf("chdir to test project: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, ".chauffeur")
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}

	t.Setenv("HOME", tmpDir)
	ensureMockPHP(t, tmpDir, "8.3")

	// First link should succeed
	if err := RunLink([]string{}); err != nil {
		t.Errorf("First RunLink() failed: %v", err)
	}

	// Second link without --force should fail
	if err := RunLink([]string{}); err == nil || !contains(err.Error(), "use --force to overwrite") {
		t.Errorf("Second RunLink() should fail without --force, got %v", err)
	}

	// Second link with --force should succeed
	if err := RunLink([]string{"--force"}); err != nil {
		t.Errorf("RunLink() with --force failed: %v", err)
	}
}

func captureOutput(fn func() error) (string, error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w // Redirect stderr too

	err := fn()

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	data := make([]byte, 1024)
	n, _ := r.Read(data)
	_ = r.Close()

	return string(data[:n]), err
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s == substr ||
					(len(s) > len(substr) && checkContainsSubstring(s, substr))))
}

func checkContainsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		matched := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func ensureMockPHP(t *testing.T, home string, versions ...string) {
	t.Helper()
	for _, version := range versions {
		binDir := filepath.Join(home, ".chauffeur", "php", version, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatalf("create mock PHP dir: %v", err)
		}
		phpPath := filepath.Join(binDir, "php")
		if err := os.WriteFile(phpPath, []byte("#!/usr/bin/env bash\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("create mock PHP binary: %v", err)
		}
	}
}
