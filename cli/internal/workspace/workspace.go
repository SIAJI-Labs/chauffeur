package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const workspaceDirName = ".chauffeur"

// Dir returns the root Chauffeur workspace path (defaults to ~/.chauffeur).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, workspaceDirName), nil
}

// Path joins sub paths under the workspace root.
func Path(parts ...string) (string, error) {
	root, err := Dir()
	if err != nil {
		return "", err
	}
	all := append([]string{root}, parts...)
	return filepath.Join(all...), nil
}

// Ensure sets up required directories if they do not exist.
func Ensure(paths ...string) error {
	root, err := Dir()
	if err != nil {
		return err
	}

	toEnsure := []string{root}
	toEnsure = append(toEnsure, paths...)

	for _, rel := range toEnsure {
		var target string
		if filepath.IsAbs(rel) {
			target = rel
		} else {
			target = filepath.Join(root, rel)
		}

		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("ensure directory %s: %w", target, err)
		}
	}
	return nil
}

// IsInitialized checks if the Chauffeur workspace exists.
func IsInitialized() bool {
	root, err := Dir()
	if err != nil {
		return false
	}

	info, err := os.Stat(root)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// ValidateOrPrompt checks if workspace is initialized, and offers to initialize it if not.
// Returns true if workspace is ready (either existed or was successfully initialized).
func ValidateOrPrompt() (bool, error) {
	if IsInitialized() {
		return true, nil
	}

	root, err := Dir()
	if err != nil {
		return false, fmt.Errorf("determine workspace path: %w", err)
	}

	fmt.Printf("Chauffeur workspace not found at %s\n\n", root)
	fmt.Println("This appears to be your first time using Chauffeur.")
	fmt.Println("The workspace stores your configurations, projects, and installed services.")
	fmt.Printf("\nWould you like to initialize the workspace now? [Y/n]: ")

	response := promptUser()

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "n" || response == "no" {
		fmt.Println("\nTo initialize manually, run: chauf init")
		return false, nil
	}

	// User wants to initialize
	fmt.Println("\nInitializing Chauffeur workspace...")
	if err := Ensure(); err != nil {
		return false, fmt.Errorf("failed to initialize workspace: %w", err)
	}

	fmt.Printf("✅ Workspace created successfully at %s\n", root)
	fmt.Println("\nNext steps:")
	fmt.Println("  chauf install php 8.3    # Install PHP runtime")
	fmt.Println("  chauf link              # Link your current project")
	fmt.Println("  chauf start             # Start services")
	fmt.Println()
	fmt.Println("For more information, run: chauf --help")

	return true, nil
}

// ValidateForCommand checks workspace initialization for commands that need it.
// Returns true if workspace is ready, false if command should not proceed.
func ValidateForCommand(args []string) (bool, error) {
	// Check if any help flag is present
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true, nil // Skip validation for help commands
		}
	}

	return ValidateOrPrompt()
}

// promptUser reads user input from stdin.
func promptUser() string {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "y" // Default to yes on read error
	}
	return strings.TrimSpace(response)
}
