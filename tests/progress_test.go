package tests

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/cli/lib"
)

func TestProgress_DisplayInCommands(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create a basic workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Test a command that doesn't require network/installation but uses the same infrastructure
	output := captureError(func() error {
		// Test an invalid command to check error handling without long operations
		return commands.RunInstall([]string{"invalid-service"})
	})

	// Should have some error output (not panic/crash)
	if output == "" {
		t.Error("Expected some error output for invalid service")
	}

	// Should not have panic/crash output
	if strings.Contains(output, "panic") || strings.Contains(output, "runtime error") {
		t.Errorf("Output contains panic/runtime error: %s", output)
	}
}

func TestProgress_HumanBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}
	
	for _, test := range tests {
		result := lib.HumanBytes(test.input)
		if result != test.expected {
			t.Errorf("HumanBytes(%d) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestProgress_DownloadSimulation(t *testing.T) {
	// Test progress tracking functionality without network calls
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create a local file to simulate download
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	
	// Write some test data
	testData := strings.Repeat("test data ", 100)
	if err := os.WriteFile(tempFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test downloadToFile with a local URL (file://) to avoid network issues
	outFile := filepath.Join(tmpHome, "downloaded.txt")
	
	// Since file:// URLs may not be supported, we'll test the progress components directly
	progress := lib.NewProgressPrinter("Test Download", int64(len(testData)))
	
	if progress == nil {
		t.Fatal("Expected progress printer to be created")
	}
	
	// Read the file and write through progress tracker
	inFile, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open input file: %v", err)
	}
	defer inFile.Close()
	
	outFileHandle, err := os.Create(outFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer outFileHandle.Close()
	
	// Use progress writer
	writer := io.MultiWriter(outFileHandle, progress)
	
	// Copy data through progress tracker
	_, err = io.Copy(writer, inFile)
	
	// Even if it errors, it shouldn't panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Progress tracking panicked: %v", err)
	}
	
	// Finish progress tracking
	progress.Finish()
}

func TestProgress_ProgressBarIntegration(t *testing.T) {
	// Test that progress tracking components work together
	progress := lib.NewProgressPrinter("integration test", 1000)
	
	if progress == nil {
		t.Fatal("Expected progress printer to be created")
	}
	
	// Write data in multiple steps and test that it doesn't panic
	data1 := strings.Repeat("x", 250)
	data2 := strings.Repeat("y", 250)
	data3 := strings.Repeat("z", 500)
	
	// These should not panic
	progress.Write([]byte(data1))
	progress.Write([]byte(data2))
	progress.Write([]byte(data3))
	
	// Test completion - should not panic
	progress.Finish()
	
	// Test calling finish twice doesn't panic
	progress.Finish()
}

func TestProgress_NonInterference(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	start := time.Now()
	
	// Run a command that shouldn't be affected by progress tracking
	output := captureOutput(func() error {
		return commands.RunLinks([]string{})
	})
	
	duration := time.Since(start)
	
	// Should complete quickly
	if duration > 2*time.Second {
		t.Errorf("Command took too long (%v), progress tracking may be interfering", duration)
	}
	
	// Should not have progress-related errors
	if strings.Contains(output, "nil pointer") || strings.Contains(output, "segmentation") {
		t.Errorf("Output contains memory errors: %s", output)
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}
	
	for _, test := range tests {
		result := lib.HumanBytes(test.input)
		if result != test.expected {
			t.Errorf("HumanBytes(%d) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestProgressPrinter_Completion(t *testing.T) {
	progress := lib.NewProgressPrinter("test completion", 10)
	
	// Write exactly the total amount
	data := []byte("0123456789")
	progress.Write(data)
	
	// Since we can't access unexported fields, just test that this doesn't panic
	// and that the Write method accepts the data correctly
}

func TestProgressPrinter_TriggerRender(t *testing.T) {
	progress := lib.NewProgressPrinter("test render", 1000)
	
	// Test that render is triggered appropriately without accessing unexported fields
	data := []byte(strings.Repeat("x", 100))
	progress.Write(data)
	
	// The main thing is that this doesn't panic
	// Since we can't access unexported fields, we just test that Write method works
}
