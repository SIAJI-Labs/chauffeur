package logstest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestLogsCommandWithoutArguments(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	// Test logs command without arguments should list available services
	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Should not error when listing services (even if none are running)
	err := commands.RunLogs([]string{"--quiet"})
	if err != nil {
		// Check if it's just a "no services found" type error
		if !strings.Contains(err.Error(), "no service") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error for logs listing, got: %v", err)
		}
	}
}

func TestLogsWithHelp(t *testing.T) {
	// Test logs help command
	err := commands.RunLogs([]string{"--help"})
	if err != nil {
		t.Fatalf("Expected no error for logs help, got: %v", err)
	}
}

func TestLogsWithInvalidService(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test logs with non-existent service
	err := commands.RunLogs([]string{"non-existent-service"})
	if err == nil {
		t.Fatal("Expected error for non-existent service")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Expected 'not found' error, got: %v", err)
	}
}

func TestLogsWithNginxService(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake nginx log file
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	// Create fake access.log
	accessLogPath := filepath.Join(nginxLogDir, "access.log")
	accessLogContent := `127.0.0.1 - - [19/Nov/2025:16:30:00 +0000] "GET / HTTP/1.1" 200 512 "-" "curl/7.68.0"
127.0.0.1 - - [19/Nov/2025:16:30:01 +0000] "GET /favicon.ico HTTP/1.1" 404 209 "-" "curl/7.68.0"
`
	if err := os.WriteFile(accessLogPath, []byte(accessLogContent), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test logs for nginx service
	err := commands.RunLogs([]string{"nginx", "--quiet", "--lines", "1"})
	if err != nil {
		// Should not error even if nginx service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for nginx logs, got: %v", err)
		}
	}
}

func TestLogsWithPHPFPMService(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake project and link it
	projectDir := helpers.NewProjectDir(t, home, "logs-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create fake PHP-FPM log file
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectLogDir := filepath.Join(projectsDir, "logs-test-app", "logs")
	if err := os.MkdirAll(projectLogDir, 0755); err != nil {
		t.Fatalf("Failed to create project log directory: %v", err)
	}

	fpmLogPath := filepath.Join(projectLogDir, "php-fpm.log")
	fpmLogContent := `[19-Nov-2025 16:30:00] NOTICE: fpm is running, pid 12345
[19-Nov-2025 16:30:01] WARNING: [pool www] child 12346 said into stderr: "WARNING: Something happened"
[19-Nov-2025 16:30:02] ERROR: FPM encountered an error
`
	if err := os.WriteFile(fpmLogPath, []byte(fpmLogContent), 0644); err != nil {
		t.Fatalf("Failed to write php-fpm.log: %v", err)
	}

	// Test logs for php-fpm service
	err := commands.RunLogs([]string{"php-fpm", "--quiet", "--lines", "1"})
	if err != nil {
		// Should not error even if PHP-FPM service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for php-fpm logs, got: %v", err)
		}
	}
}

func TestLogsWithLinesOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake log file with multiple lines
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	accessLogPath := filepath.Join(nginxLogDir, "access.log")
	var accessLogContent string
	for i := 1; i <= 10; i++ {
		accessLogContent += fmt.Sprintf("127.0.0.1 - - [19/Nov/2025:16:30:%02d +0000] \"GET /page%d HTTP/1.1\" 200 512\n", i, i)
	}

	if err := os.WriteFile(accessLogPath, []byte(accessLogContent), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test logs with limited lines
	err := commands.RunLogs([]string{"nginx", "--quiet", "--lines", "3"})
	if err != nil {
		// Should not error even if nginx service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for nginx logs with lines option, got: %v", err)
		}
	}
}

func TestLogsWithLevelFilter(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake log file with different log levels
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	errorLogPath := filepath.Join(nginxLogDir, "error.log")
	errorLogContent := `2025/11/19 16:30:00 [info] 12345#0: start worker processes
2025/11/19 16:30:01 [warn] 12345#0: some warning message
2025/11/19 16:30:02 [error] 12345#0: some error message
2025/11/19 16:30:03 [info] 12345#0: another info message
`
	if err := os.WriteFile(errorLogPath, []byte(errorLogContent), 0644); err != nil {
		t.Fatalf("Failed to write error.log: %v", err)
	}

	// Test logs with error level filter
	err := commands.RunLogs([]string{"nginx", "--quiet", "--level", "error"})
	if err != nil {
		// Should not error even if nginx service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for nginx logs with level filter, got: %v", err)
		}
	}
}

func TestLogsWithVerboseOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake log file
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	accessLogPath := filepath.Join(nginxLogDir, "access.log")
	accessLogContent := `127.0.0.1 - - [19/Nov/2025:16:30:00 +0000] "GET / HTTP/1.1" 200 512`
	if err := os.WriteFile(accessLogPath, []byte(accessLogContent), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test logs with verbose option
	err := commands.RunLogs([]string{"nginx", "--verbose", "--quiet", "--lines", "1"})
	if err != nil {
		// Should not error even if nginx service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for nginx logs with verbose option, got: %v", err)
		}
	}
}

func TestLogsWithQuietOption(t *testing.T) {
	_, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Test logs with quiet option (should minimize output)
	err := commands.RunLogs([]string{"--quiet"})
	if err != nil {
		// Check if it's just a "no services found" type error
		if !strings.Contains(err.Error(), "no service") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error for logs quiet, got: %v", err)
		}
	}
}

func TestLogsWithContextOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake log file
	nginxLogDir := filepath.Join(home, ".chauffeur", "nginx", "logs")
	if err := os.MkdirAll(nginxLogDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx log directory: %v", err)
	}

	accessLogPath := filepath.Join(nginxLogDir, "access.log")
	accessLogContent := `127.0.0.1 - - [19/Nov/2025:16:30:00 +0000] "GET / HTTP/1.1" 200 512`
	if err := os.WriteFile(accessLogPath, []byte(accessLogContent), 0644); err != nil {
		t.Fatalf("Failed to write access.log: %v", err)
	}

	// Test logs with context option
	err := commands.RunLogs([]string{"nginx", "--context", "--quiet", "--lines", "1"})
	if err != nil {
		// Should not error even if nginx service isn't running
		if !strings.Contains(err.Error(), "no log files") && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("Expected no error or 'no log files' error for nginx logs with context option, got: %v", err)
		}
	}
}