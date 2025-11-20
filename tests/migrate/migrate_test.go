package migratetest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/tests/internal/helpers"
)

func TestMigrateCommandHelp(t *testing.T) {
	// Test migrate help command
	err := commands.RunMigrate([]string{"--help"})
	if err != nil {
		t.Fatalf("Expected no error for migrate help, got: %v", err)
	}
}

func TestMigrateWithoutArguments(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test migrate command without arguments (should require source and destination)
	err := commands.RunMigrate([]string{})
	if err == nil {
		t.Fatal("Expected error for migrate without arguments")
	}

	if !strings.Contains(err.Error(), "source project slug") {
		t.Fatalf("Expected 'source project slug' error, got: %v", err)
	}
}

func TestMigrateWithOnlySource(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test migrate command with only source (should require destination)
	err := commands.RunMigrate([]string{"test-project"})
	if err == nil {
		t.Fatal("Expected error for migrate with only source")
	}

	if !strings.Contains(err.Error(), "destination workspace") {
		t.Fatalf("Expected 'destination workspace' error, got: %v", err)
	}
}

func TestMigrateWithNonExistentProject(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create a fake destination workspace
	destWorkspace := filepath.Join(home, "dest-workspace")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination workspace: %v", err)
	}

	// Create .chauffeur directory in destination to make it look like a workspace
	if err := os.MkdirAll(filepath.Join(destWorkspace, ".chauffeur"), 0755); err != nil {
		t.Fatalf("Failed to create .chauffeur in destination: %v", err)
	}

	// Test migrate with non-existent project
	err := commands.RunMigrate([]string{"non-existent-project", destWorkspace, "--dry-run"})
	if err == nil {
		t.Fatal("Expected error for migrate with non-existent project")
	}

	if !strings.Contains(err.Error(), "not linked") {
		t.Fatalf("Expected 'not linked' error, got: %v", err)
	}
}

func TestMigrateWithInvalidDestination(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Test migrate with non-existent destination
	nonExistentDest := filepath.Join(home, "non-existent")
	err := commands.RunMigrate([]string{"migrate-test-app", nonExistentDest, "--dry-run"})
	if err == nil {
		t.Fatal("Expected error for migrate with non-existent destination")
	}

	if !strings.Contains(err.Error(), "destination path does not exist") {
		t.Fatalf("Expected 'destination path does not exist' error, got: %v", err)
	}
}

func TestMigrateWithNonChauffeurDestination(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create a destination directory that's not a Chauffeur workspace
	destWorkspace := filepath.Join(home, "regular-directory")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Test migrate with non-Chauffeur destination
	err := commands.RunMigrate([]string{"migrate-test-app", destWorkspace, "--dry-run"})
	if err == nil {
		t.Fatal("Expected error for migrate with non-Chauffeur destination")
	}

	if !strings.Contains(err.Error(), "not a Chauffeur workspace") {
		t.Fatalf("Expected 'not a Chauffeur workspace' error, got: %v", err)
	}
}

func TestMigrateDryRun(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create a fake destination workspace
	destWorkspace := filepath.Join(home, "dest-workspace")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination workspace: %v", err)
	}

	// Create .chauffeur directory in destination to make it look like a workspace
	if err := os.MkdirAll(filepath.Join(destWorkspace, ".chauffeur"), 0755); err != nil {
		t.Fatalf("Failed to create .chauffeur in destination: %v", err)
	}

	// Test migrate in dry-run mode
	err := commands.RunMigrate([]string{"migrate-test-app", destWorkspace, "--dry-run", "--no-backup", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for migrate dry-run, got: %v", err)
	}

	// Verify project still exists in source workspace after dry-run
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectPath := filepath.Join(projectsDir, "migrate-test-app")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatal("Project should still exist in source workspace after dry-run")
	}

	// Verify project was not copied to destination workspace
	destProjectsDir := filepath.Join(destWorkspace, ".chauffeur", "projects")
	destProjectPath := filepath.Join(destProjectsDir, "migrate-test-app")
	if _, err := os.Stat(destProjectPath); !os.IsNotExist(err) {
		t.Fatal("Project should not exist in destination workspace after dry-run")
	}
}

func TestMigrateWithBackupOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create a fake destination workspace
	destWorkspace := filepath.Join(home, "dest-workspace")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination workspace: %v", err)
	}

	// Create .chauffeur directory in destination
	if err := os.MkdirAll(filepath.Join(destWorkspace, ".chauffeur"), 0755); err != nil {
		t.Fatalf("Failed to create .chauffeur in destination: %v", err)
	}

	// Test migrate with backup option (dry-run mode to avoid actual migration)
	err := commands.RunMigrate([]string{"migrate-test-app", destWorkspace, "--dry-run", "--backup", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for migrate with backup option dry-run, got: %v", err)
	}

	// Verify project still exists in source workspace after dry-run
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectPath := filepath.Join(projectsDir, "migrate-test-app")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatal("Project should still exist in source workspace after dry-run with backup")
	}
}

func TestMigrateWithNoBackupOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create a fake destination workspace
	destWorkspace := filepath.Join(home, "dest-workspace")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination workspace: %v", err)
	}

	// Create .chauffeur directory in destination
	if err := os.MkdirAll(filepath.Join(destWorkspace, ".chauffeur"), 0755); err != nil {
		t.Fatalf("Failed to create .chauffeur in destination: %v", err)
	}

	// Test migrate with no-backup option (dry-run mode)
	err := commands.RunMigrate([]string{"migrate-test-app", destWorkspace, "--dry-run", "--no-backup", "--force"})
	if err != nil {
		t.Fatalf("Expected no error for migrate with no-backup option dry-run, got: %v", err)
	}

	// Verify project still exists in source workspace after dry-run
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectPath := filepath.Join(projectsDir, "migrate-test-app")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatal("Project should still exist in source workspace after dry-run with no-backup")
	}
}

func TestMigrateWithVerboseOption(t *testing.T) {
	home, workspace := helpers.SetupTestHome(t)

	if err := commands.RunInit(nil); err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	helpers.EnsureFakePHP(t, workspace, "8.3")

	// Create and link a project
	projectDir := helpers.NewProjectDir(t, home, "migrate-test-app")
	helpers.Chdir(t, projectDir)

	if err := commands.RunLink(nil); err != nil {
		t.Fatalf("RunLink failed: %v", err)
	}

	// Create a fake destination workspace
	destWorkspace := filepath.Join(home, "dest-workspace")
	if err := os.MkdirAll(destWorkspace, 0755); err != nil {
		t.Fatalf("Failed to create destination workspace: %v", err)
	}

	// Create .chauffeur directory in destination
	if err := os.MkdirAll(filepath.Join(destWorkspace, ".chauffeur"), 0755); err != nil {
		t.Fatalf("Failed to create .chauffeur in destination: %v", err)
	}

	// Test migrate with verbose option (dry-run mode)
	err := commands.RunMigrate([]string{"migrate-test-app", destWorkspace, "--dry-run", "--no-backup", "--force", "--verbose"})
	if err != nil {
		t.Fatalf("Expected no error for migrate with verbose option dry-run, got: %v", err)
	}

	// Verify project still exists in source workspace after dry-run
	projectsDir := filepath.Join(home, ".chauffeur", "projects")
	projectPath := filepath.Join(projectsDir, "migrate-test-app")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatal("Project should still exist in source workspace after dry-run with verbose")
	}
}

func TestMigrateWithInvalidFlags(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test migrate with invalid flag
	err := commands.RunMigrate([]string{"--invalid-flag"})
	if err == nil {
		t.Fatal("Expected error for migrate with invalid flag")
	}

	if !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("Expected 'unknown flag' error, got: %v", err)
	}
}

func TestMigrateWithTooManyArguments(t *testing.T) {
	_, _ = helpers.SetupTestHome(t)

	// Test migrate with too many arguments
	err := commands.RunMigrate([]string{"source", "dest", "extra-arg"})
	if err == nil {
		t.Fatal("Expected error for migrate with too many arguments")
	}

	if !strings.Contains(err.Error(), "too many arguments") {
		t.Fatalf("Expected 'too many arguments' error, got: %v", err)
	}
}

func TestMigrateWithoutWorkspace(t *testing.T) {
	// Test migrate command without a proper workspace (should fail gracefully)
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWD)

	// Test migrate without workspace (should fail)
	err := commands.RunMigrate([]string{"test-project", "/tmp/dest", "--dry-run"})
	if err == nil {
		t.Fatal("Expected error for migrate without workspace")
	}

	if !strings.Contains(err.Error(), "workspace") {
		t.Fatalf("Expected workspace-related error, got: %v", err)
	}
}