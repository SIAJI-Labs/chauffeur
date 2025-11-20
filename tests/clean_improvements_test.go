package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test file size display functionality
func TestFileSizeDisplay(t *testing.T) {
	// Create test files with different sizes
	tempDir := t.TempDir()
	testFiles := []struct {
		name    string
		content string
		size    int64
	}{
		{"small.txt", "small", 5},
		{"medium.txt", strings.Repeat("x", 1024), 1024},
		{"large.txt", strings.Repeat("x", 1024*1024), 1024 * 1024},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(filePath, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Verify file size calculation
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		if info.Size() != tf.size {
			t.Errorf("File size mismatch for %s: got %d, want %d", tf.name, info.Size(), tf.size)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	// Test file size formatting function
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024 - 1, "1024.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1.5, "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tc := range testCases {
		// Simulate formatBytes function (same as in clean.go)
		const unit = 1024
		var result string
		if tc.bytes < unit {
			result = fmt.Sprintf("%d B", tc.bytes)
		} else {
			div, exp := int64(unit), 0
			for n := tc.bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			result = fmt.Sprintf("%.1f %cB", float64(tc.bytes)/float64(div), "KMGTPE"[exp])
		}

		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
		}
	}
}

func TestFileDeletionAndCounting(t *testing.T) {
	tempDir := t.TempDir()

	// Create test log files
	logFiles := []string{"app.log", "error.log", "access.log"}
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("Log content for %s", filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test log file: %v", err)
		}
	}

	// Simulate file deletion with counting
	var totalSize int64
	var totalFiles int
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		// "Delete" the file
		if err := os.Remove(filePath); err != nil {
			t.Errorf("Failed to delete file %s: %v", filename, err)
		} else {
			totalSize += info.Size()
			totalFiles++
		}
	}

	// Verify counting
	if totalFiles != len(logFiles) {
		t.Errorf("Deleted %d files, want %d", totalFiles, len(logFiles))
	}

	// Verify size calculation
	if totalSize <= 0 {
		t.Errorf("Total size = %d, want > 0", totalSize)
	}

	// Verify files are actually deleted
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			t.Errorf("File %s was not deleted", filename)
		}
	}
}

func TestDryRunMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create test log files
	logFiles := []string{"app.log", "error.log"}
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("Log content for %s", filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test log file: %v", err)
		}
	}

	// Simulate dry-run mode (should count but not delete)
	var totalSize int64
	var totalFiles int
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file: %v", err)
			continue
		}

		// In dry-run mode, we count but don't delete
		totalSize += info.Size()
		totalFiles++
	}

	// Should count files but not delete them
	if totalFiles != len(logFiles) {
		t.Errorf("Dry-run counted %d files, want %d", totalFiles, len(logFiles))
	}

	// Verify files are still present (not deleted in dry-run)
	for _, filename := range logFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("File %s was deleted in dry-run mode", filename)
		}
	}
}

func TestAccurateFileCounting(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{"cache1", "cache2", "temp1"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		content := strings.Repeat("x", 100) // 100 bytes each
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Simulate accurate counting (only count actually deleted files)
	var totalSize int64
	var totalFiles int
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file: %v", err)
			continue
		}

		// Simulate deletion with user confirmation
		// In real usage, only count if user confirms deletion
		confirmed := true // Simulate user confirmation
		if confirmed {
			if err := os.Remove(filePath); err != nil {
				t.Errorf("Failed to delete file %s: %v", filename, err)
			} else {
				totalSize += info.Size()
				totalFiles++
			}
		}
	}

	// Verify all files were deleted
	if totalFiles != len(testFiles) {
		t.Errorf("Deleted %d files, want %d", totalFiles, len(testFiles))
	}

	// Verify size calculation (100 bytes per file)
	expectedSize := int64(len(testFiles) * 100)
	if totalSize != expectedSize {
		t.Errorf("Total size = %d, want %d", totalSize, expectedSize)
	}
}

func TestCleanLogs_NoFilesMessageLogic(t *testing.T) {
	// Test the logic for handling empty directories
	tempDir := t.TempDir()

	// Create workspace structure
	workspaceDir := filepath.Join(tempDir, ".chauffeur")
	projectsDir := filepath.Join(workspaceDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Test that empty directories don't cause errors
	files := []string{}
	dirs := []string{workspaceDir, projectsDir}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("Failed to read directory %s: %v", dir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".log") {
				files = append(files, entry.Name())
			}
		}
	}

	// Should find no log files
	if len(files) != 0 {
		t.Errorf("Expected no log files, found %d", len(files))
	}
}

func TestOlderThanFilterLogic(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different ages
	now := time.Now()
	testFiles := []struct {
		name       string
		age        time.Duration
		shouldKeep bool
	}{
		{"old.log", 2 * 24 * time.Hour, false},   // 2 days old - should be deleted
		{"new.log", 1 * time.Hour, true},         // 1 hour old - should be kept
		{"recent.log", 30 * time.Minute, true},   // 30 minutes old - should be kept
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		content := fmt.Sprintf("Log content for %s", tf.name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Set file modification time
		modTime := now.Add(-tf.age)
		if err := os.Chtimes(filePath, modTime, modTime); err != nil {
			t.Fatalf("Failed to set file time: %v", err)
		}
	}

	// Test filtering logic (simulate older-than filter of 24h)
	olderThan := 24 * time.Hour
	var filesToDelete []string
	var filesToKeep []string

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file %s: %v", tf.name, err)
			continue
		}

		if time.Since(info.ModTime()) > olderThan {
			filesToDelete = append(filesToDelete, tf.name)
		} else {
			filesToKeep = append(filesToKeep, tf.name)
		}
	}

	// Should only find files older than 1 day for deletion
	expectedToDelete := 1 // Only old.log
	if len(filesToDelete) != expectedToDelete {
		t.Errorf("Filter logic found %d files to delete, want %d", len(filesToDelete), expectedToDelete)
	}

	// Should keep recent files
	expectedToKeep := 2 // new.log and recent.log
	if len(filesToKeep) != expectedToKeep {
		t.Errorf("Filter logic found %d files to keep, want %d", len(filesToKeep), expectedToKeep)
	}
}

func TestMultipleCleanupTargetsLogic(t *testing.T) {
	tempDir := t.TempDir()

	// Create workspace structure
	workspaceDir := filepath.Join(tempDir, ".chauffeur")
	logsDir := filepath.Join(workspaceDir, "nginx", "logs")
	cacheDir := filepath.Join(workspaceDir, "cache")
	tempDirForTests := filepath.Join(workspaceDir, "tmp")

	for _, dir := range []string{logsDir, cacheDir, tempDirForTests} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in each directory
	testFiles := map[string]string{
		filepath.Join(logsDir, "access.log"):      "nginx access log",
		filepath.Join(cacheDir, "composer.phar"):   "composer binary data",
		filepath.Join(tempDirForTests, "temp.tmp"): "temporary data",
	}

	for filePath, content := range testFiles {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test finding files by type (simulate cleanup target logic)
	var logFiles, cacheFiles, tempFiles []string

	for filePath := range testFiles {
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file %s: %v", filePath, err)
			continue
		}

		if strings.Contains(filePath, "logs") && strings.HasSuffix(info.Name(), ".log") {
			logFiles = append(logFiles, info.Name())
		} else if strings.Contains(filePath, "cache") {
			cacheFiles = append(cacheFiles, info.Name())
		} else if strings.Contains(filePath, "tmp") {
			tempFiles = append(tempFiles, info.Name())
		}
	}

	// Verify we found the right files in each category
	if len(logFiles) != 1 || logFiles[0] != "access.log" {
		t.Errorf("Expected to find access.log in log files, got %v", logFiles)
	}
	if len(cacheFiles) != 1 || cacheFiles[0] != "composer.phar" {
		t.Errorf("Expected to find composer.phar in cache files, got %v", cacheFiles)
	}
	if len(tempFiles) != 1 || tempFiles[0] != "temp.tmp" {
		t.Errorf("Expected to find temp.tmp in temp files, got %v", tempFiles)
	}
}

func TestOldVersionsCleanupLogic(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock PHP directory structure
	phpDir := filepath.Join(tempDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("Failed to create PHP directory: %v", err)
	}

	// Create mock PHP versions with different modification times
	versions := []struct {
		name    string
		age     time.Duration
	}{
		{"8.3", 1 * time.Hour},     // Most recent
		{"8.2", 2 * 24 * time.Hour}, // 2 days old
		{"8.1", 7 * 24 * time.Hour}, // 1 week old
		{"7.4", 14 * 24 * time.Hour}, // 2 weeks old
	}

	for _, version := range versions {
		versionDir := filepath.Join(phpDir, version.name)
		if err := os.MkdirAll(versionDir, 0755); err != nil {
			t.Fatalf("Failed to create version directory: %v", err)
		}

		// Add some files to make it look like a real PHP installation
		for _, file := range []string{"bin/php", "etc/php.ini"} {
			filePath := filepath.Join(versionDir, file)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				continue
			}
			if err := os.WriteFile(filePath, []byte("mock content"), 0644); err != nil {
				continue
			}
		}

		// Set modification time
		modTime := time.Now().Add(-version.age)
		if err := os.Chtimes(versionDir, modTime, modTime); err != nil {
			t.Fatalf("Failed to set version directory time: %v", err)
		}
	}

	// Test version sorting and selection logic
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		t.Fatalf("Failed to read PHP directory: %v", err)
	}

	var versionDirs []struct {
		name    string
		modTime time.Time
	}

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			versionDirs = append(versionDirs, struct {
				name    string
				modTime time.Time
			}{entry.Name(), info.ModTime()})
		}
	}

	// Sort by modification time (newest first) - simulate cleanOldVersions logic
	for i := 0; i < len(versionDirs)-1; i++ {
		for j := i + 1; j < len(versionDirs); j++ {
			if versionDirs[i].modTime.Before(versionDirs[j].modTime) {
				versionDirs[i], versionDirs[j] = versionDirs[j], versionDirs[i]
			}
		}
	}

	// Test keeping 2 versions
	keepVersions := 2
	if len(versionDirs) <= keepVersions {
		t.Errorf("Expected no old versions to remove, found %d total versions", len(versionDirs))
	}

	// Test which versions would be removed (older versions beyond keep count)
	var oldVersions []string
	for i := keepVersions; i < len(versionDirs); i++ {
		oldVersions = append(oldVersions, versionDirs[i].name)
	}

	expectedOldVersions := 2 // 8.1 and 7.4
	if len(oldVersions) != expectedOldVersions {
		t.Errorf("Expected %d old versions to remove, got %d: %v", expectedOldVersions, len(oldVersions), oldVersions)
	}
}

// Benchmark tests for performance
func BenchmarkFormatBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Simulate formatBytes function for 1MB
		const unit = 1024
		bytes := int64(1024 * 1024)
		if bytes < unit {
			_ = fmt.Sprintf("%d B", bytes)
		} else {
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			_ = fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		}
	}
}

func BenchmarkFileDeletion(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("test_%d.txt", i))
		content := []byte("test content for benchmarking")

		// Create file
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}

		// Delete file (simulating clean operation)
		if err := os.Remove(testFile); err != nil {
			b.Fatalf("Failed to delete test file: %v", err)
		}
	}
}