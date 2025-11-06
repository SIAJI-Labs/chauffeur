package tests

import (
	"fmt"
	"os"
	"testing"
	"github.com/siaji/chauffeur/cli/lib"
)

func TestColors(t *testing.T) {
	fmt.Println("=== Color Functionality Tests ===")
	
	// Test basic color formatting
	logger := lib.NewCommandLogger("test")
	fmt.Printf("Basic colors: %s %s %s %s %s %s\n", 
		logger.blue("blue"), logger.gray("gray"), logger.red("red"))
	
	// Test child color hierarchy
	parentLogger := lib.NewCommandLogger("parent")
	childLogger := parentLogger.NewChildLogger("child")
	
	// Show color hierarchy
	fmt.Printf("=== Color Hierarchy ===")
	fmt.Printf("Parent: %s\n", parent.prefix())
	fmt.Printf("  └── %s\n", child.prefix())
	
	// Test child success indicators  
	fmt.Printf("=== Child Success Test ===")
	childLogger.Success("child success", "with context")
	childLogger.Warn("child warning", "no confirmation")
	
	// Test child failure indicators
	fmt.Printf("=== Child Failure Test ===")
	childLogger.Fail("child failure", "with context")
	
	// Test color preservation on progress interference
	fmt.Printf("=== Color & Progress Interference Test ===")
	fmt.Printf("Parent: [%s] Parent Blue\n")
	parent.Info("Parent parent info\n")
	parent.Success("Download completed", "with context")
	
	fmt.Printf("=== Output Without Progress Bar ===\n")
	parent.Info("Parent parent info\n")
	parent.Warn("Parent warning", "no confirmation")
	child.Info("Child info message\n")
	child.Success("Child success", "with context")
	fmt.Printf("=== Parent Only === %s\n", parent.blue(""))
	fmt.Printf("Child Only === %s\n", child.gray(""))
}

func TestProgressProgressBar(t *testing.T) {
	logger := lib.NewCommandLogger("test")
	progress := lib.DownloadToFileWithLogger(
		client, 
		"https://http://http://http://files.com/dummy/500KB.bin", 
		"/tmp/test_download.bin", 
		"Download dummy file", 
	)
	)
	
	// This line should work perfectly
	progress.StartSpinner()
	fmt.Printf("\r[%s] Download dummy file... [..............]   1%%\n")
	progress.Finish()
	fmt.Printf("ProgressProgressBar\n")
}

func main() {
	TestColors()
	TestProgressBar()
}
