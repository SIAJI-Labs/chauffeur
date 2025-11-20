package cleantest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestCleanCommandHelp(t *testing.T) {
	// Test clean help command
	err := commands.RunClean([]string{"--help"})
	if err != nil {
		t.Fatalf("Expected no error for clean help, got: %v", err)
	}
}

func TestCleanWithInvalidTarget(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test clean with invalid target
	err := commands.RunClean([]string{"invalid-target"})
	if err == nil {
		t.Fatal("Expected error for invalid clean target")
	}

	if !strings.Contains(err.Error(), "invalid clean target") {
		t.Fatalf("Expected 'invalid clean target' error, got: %v", err)
	}
}

func TestCleanLogsDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake log files
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	// Create fake log files
	logFiles := map[string]string{
		"access.log":    "127.0.0.1 - - [19/Nov/2025:16:30:00 +0000] \"GET / HTTP/1.1\" 200 512",
		"error.log":     "2025/11/19 16:30:00 [error] 12345#0: some error",
		"nginx-error.log": "2025/11/19 16:30:00 [error] 12345#0: nginx error",
	}

	for filename, content := range logFiles {
		logPath := filepath.Join(nginxLogDir, filename)
		if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	// Test clean logs in dry-run mode
	err := commands.RunClean([]string{"logs", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean logs dry-run, got: %v", err)
	}

	// Verify log files still exist after dry-run
	for filename := range logFiles {
		logPath := filepath.Join(nginxLogDir, filename)
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatalf("Log file %s should still exist after dry-run", filename)
		}
	}
}

func TestCleanCacheDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake cache files
	cacheDir := filepath.Join(home, ".chauffeur", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	// Create fake cache files
	cacheFiles := map[string]string{
		"cache1.tmp": "some cache data 1",
		"cache2.tmp": "some cache data 2",
		"opcache.bin": "opcache data",
	}

	for filename, content := range cacheFiles {
		cachePath := filepath.Join(cacheDir, filename)
		if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	// Test clean cache in dry-run mode
	err := commands.RunClean([]string{"cache", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean cache dry-run, got: %v", err)
	}

	// Verify cache files still exist after dry-run
	for filename := range cacheFiles {
		cachePath := filepath.Join(cacheDir, filename)
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			t.Fatalf("Cache file %s should still exist after dry-run", filename)
		}
	}
}

func TestCleanTempDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake temp files
	tempDir := filepath.Join(home, ".chauffeur", "tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create fake temp files
	tempFiles := map[string]string{
		"temp1.tmp":    "temporary data 1",
		"temp2.tmp":    "temporary data 2",
		"session.file": "session data",
	}

	for filename, content := range tempFiles {
		tempPath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	// Test clean temp in dry-run mode
	err := commands.RunClean([]string{"temp", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean temp dry-run, got: %v", err)
	}

	// Verify temp files still exist after dry-run
	for filename := range tempFiles {
		tempPath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(tempPath); os.IsNotExist(err) {
			t.Fatalf("Temp file %s should still exist after dry-run", filename)
		}
	}
}

func TestCleanAllDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create various types of files that would be cleaned
	dirs := []string{
		filepath.Join(home, ".chauffeur", "nginx", "logs"),
		filepath.Join(home, ".chauffeur", "cache"),
		filepath.Join(home, ".chauffeur", "tmp"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create fake files in each directory
	testFiles := []struct {
		path    string
		content string
	}{
		{filepath.Join(home, ".chauffeur", "nginx", "logs", "access.log"), "nginx access log"},
		{filepath.Join(home, ".chauffeur", "cache", "cache.tmp"), "cache data"},
		{filepath.Join(home, ".chauffeur", "tmp", "temp.tmp"), "temp data"},
	}

	for _, file := range testFiles {
		if err := os.WriteFile(file.path, []byte(file.content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", file.path, err)
		}
	}

	// Test clean all in dry-run mode
	err := commands.RunClean([]string{"all", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean all dry-run, got: %v", err)
	}

	// Verify all files still exist after dry-run
	for _, file := range testFiles {
		if _, err := os.Stat(file.path); os.IsNotExist(err) {
			t.Fatalf("File %s should still exist after dry-run", file.path)
		}
	}
}

func TestCleanWithOlderThanOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake log files
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	// Create a log file
	logPath := filepath.Join(nginxLogDir, "access.log")
	if err := os.WriteFile(logPath, []byte("test log content"), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test clean logs with older-than option (dry-run mode to avoid actual deletion)
	err := commands.RunClean([]string{"logs", "--dry-run", "--force", "--older-than", "30d"})
	if err != nil {
		t.Fatalf("Expected no error for clean logs with older-than option, got: %v", err)
	}

	// Verify log file still exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("Log file should still exist after dry-run with older-than option")
	}
}

func TestCleanWithKeepVersionsOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")
	helpers.EnsureFakePHP(t, workspace, "8.2")
	helpers.EnsureFakePHP(t, workspace, "8.1")

	// Test clean old-versions with keep-versions option (dry-run mode)
	err := commands.RunClean([]string{"old-versions", "--dry-run", "--force", "--keep-versions", "1"})
	if err != nil {
		t.Fatalf("Expected no error for clean old-versions with keep-versions option, got: %v", err)
	}

	// Verify PHP versions still exist after dry-run
	phpDir := filepath.Join(home, ".chauffeur", "php")
	versions := []string{"8.3", "8.2", "8.1"}

	for _, version := range versions {
		versionDir := filepath.Join(phpDir, version)
		if _, err := os.Stat(versionDir); os.IsNotExist(err) {
			t.Fatalf("PHP version %s should still exist after dry-run", version)
		}
	}
}

func TestCleanProjectsWithoutUnlinked(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "clean-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Test clean projects (should not remove linked projects)
	err := commands.RunClean([]string{"projects", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean projects dry-run, got: %v", err)
	}

	// Verify project directory still exists after dry-run
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectPath := filepath.Join(projectsDir, "clean-test-app")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatal("Linked project directory should still exist after dry-run")
	}
}

func TestCleanSSLCertsDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake SSL certificates
	sslDir := filepath.Join(home, ".chauffeur", "ssl")
	if err := os.MkdirAll(sslDir, 0755); err != nil {
		t.Fatalf("Failed to create SSL directory: %v", err)
	}

	// Create fake SSL certificates (not linked to any project)
	certFiles := map[string]string{
		"stale.example.com.crt": "-----BEGIN CERTIFICATE-----\nstale cert\n-----END CERTIFICATE-----",
		"stale.example.com.key": "-----BEGIN PRIVATE KEY-----\nstale key\n-----END PRIVATE KEY-----",
	}

	for filename, content := range certFiles {
		certPath := filepath.Join(sslDir, filename)
		if err := os.WriteFile(certPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	// Test clean ssl-certs in dry-run mode
	err := commands.RunClean([]string{"ssl-certs", "--dry-run", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for clean ssl-certs dry-run, got: %v", err)
	}

	// Verify SSL certificates still exist after dry-run
	for filename := range certFiles {
		certPath := filepath.Join(sslDir, filename)
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			t.Fatalf("SSL certificate %s should still exist after dry-run", filename)
		}
	}
}

func TestCleanWithVerboseOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create fake log file
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	logPath := filepath.Join(nginxLogDir, "access.log")
	if err := os.WriteFile(logPath, []byte("test log"), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test clean logs with verbose option (dry-run mode)
	err := commands.RunClean([]string{"logs", "--dry-run", "--force", "--verbose"})
	if err != nil {
		t.Fatalf("Expected no error for clean logs with verbose option, got: %v", err)
	}

	// Verify log file still exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("Log file should still exist after dry-run with verbose option")
	}
}

func TestCleanWithoutWorkspace(t *testing.T) {
	// Test clean command without a proper workspace (should fail gracefully)
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWD)

	// Test clean logs without workspace (should auto-initialize and succeed)
	err := commands.RunClean([]string{"logs", "--dry-run"})
	if err != nil {
		t.Fatalf("Expected no error for clean without workspace (auto-initialization), got: %v", err)
	}
}