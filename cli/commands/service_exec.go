package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// Security: Input validation patterns
var (
	// Safe argument pattern - prevents command injection through arguments
	// Allows alphanumeric, spaces, hyphens, underscores, dots, slashes, and common special characters
	safeArgPattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_./:@=+,%]+$`)

	// Maximum argument length to prevent buffer overflow attacks
	maxArgLength = 1000

	// Known safe services that can be executed
	knownServices = map[string]bool{
		"nginx": true, "php": true, "composer": true, "mysql": true,
	}
)

// validateServiceName ensures the service name is known and safe
func validateServiceName(name string) error {
	if !knownServices[name] {
		return fmt.Errorf("unknown service: %s", name)
	}
	return nil
}

// validateArguments validates command arguments to prevent injection
func validateArguments(args []string) error {
	for i, arg := range args {
		// Check argument length
		if len(arg) > maxArgLength {
			return fmt.Errorf("argument %d too long: %s", i, arg[:50]+"...")
		}

		// Check for dangerous patterns that could lead to command injection
		dangerousPatterns := []string{
			";", "&", "|", "`", "$(", "${", ">", "<", ">>", "<<",
			"&&", "||", "\x00", "\n", "\r",
		}

		for _, pattern := range dangerousPatterns {
			if strings.Contains(arg, pattern) {
				return fmt.Errorf("argument %d contains dangerous pattern: %s", i, pattern)
			}
		}

		// Validate against safe pattern (more permissive for file paths)
		if !safeArgPattern.MatchString(arg) {
			return fmt.Errorf("argument %d contains invalid characters: %s", i, arg)
		}
	}
	return nil
}

// validateBinaryPath ensures the binary path is safe and within allowed directories
func validateBinaryPath(binaryPath string) error {
	if binaryPath == "" {
		return fmt.Errorf("binary path cannot be empty")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("invalid binary path: %w", err)
	}

	// Ensure binary is within Chauffeur workspace or system PATH
	workspaceDir, err := workspace.Dir()
	if err != nil {
		return fmt.Errorf("failed to get workspace directory: %w", err)
	}

	// Allow execution only from workspace directories or system bin directories
	allowedPaths := []string{
		workspaceDir + "/",
		"/usr/bin/", "/usr/local/bin/", "/bin/", "/usr/sbin/",
	}

	allowed := false
	for _, allowedPath := range allowedPaths {
		if strings.HasPrefix(absPath, allowedPath) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("binary path not in allowed directories: %s", absPath)
	}

	return nil
}

/**
 * RunServiceCommand executes an installed Chauffeur-managed service with passthrough args.
 *
 * @param name Service identifier (e.g., "nginx", "php").
 * @param args Arguments forwarded to the underlying binary.
 * @return error when the service is unknown, missing, or execution fails.
 */
func RunServiceCommand(name string, args []string) error {
	// Security: Validate service name to prevent command injection
	if err := validateServiceName(name); err != nil {
		return err
	}

	// Security: Validate arguments to prevent command injection
	if err := validateArguments(args); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	logger := lib.NewCommandLogger(name)

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	info, err := system.Detect()
	if err != nil {
		return err
	}

	// Special handling for PHP to support project isolation
	if name == "php" {
		binaryPath, err := getProjectAwarePHPBinary(prefix, logger)
		if err != nil {
			return err
		}
		return runBinaryCommand(binaryPath, args)
	}

	spec, err := NewServiceSpec(name, prefix, info)
	if err != nil {
		return err
	}

	ok, err := spec.available()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("service %s is not installed; run 'chauf install %s' first", name, name)
	}

	return runBinaryCommand(spec.BinaryPath, args)
}

/**
 * getProjectAwarePHPBinary returns the appropriate PHP binary path taking project isolation into account.
 *
 * @param prefix Chauffeur workspace prefix.
 * @return path to PHP binary that should be executed.
 */
func getProjectAwarePHPBinary(prefix string, logger *lib.Logger) (string, error) {
	// Check if current directory is part of a linked project
	cfg, err := config.Load()
	if err != nil {
		// Fall back to default behavior
		spec, specErr := NewServiceSpec("php", prefix, system.Info{})
		if specErr != nil {
			return "", specErr
		}
		return spec.BinaryPath, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		// Fall back to default behavior
		spec, specErr := NewServiceSpec("php", prefix, system.Info{})
		if specErr != nil {
			return "", specErr
		}
		return spec.BinaryPath, nil
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		// Fall back to default behavior
		spec, specErr := NewServiceSpec("php", prefix, system.Info{})
		if specErr != nil {
			return "", specErr
		}
		return spec.BinaryPath, nil
	}

	projectCfg, _, err := projects.FindByPath(cfg.ProjectsDir, cwd)
	if err != nil {
		// Not in a linked project, use default behavior
		spec, specErr := NewServiceSpec("php", prefix, system.Info{})
		if specErr != nil {
			return "", specErr
		}
		return spec.BinaryPath, nil
	}

	// We're in a linked project, check if PHP is isolated
	if projectCfg.PHP != "" {
		// Project has PHP isolation, check if the isolated version is installed
		isolatedBinary := filepath.Join(prefix, "php", projectCfg.PHP, "bin", "php")
		if _, err := os.Stat(isolatedBinary); err == nil {
			// Isolated PHP version is installed, use it
			return isolatedBinary, nil
		}

		// Isolated PHP version is not installed, warn and fall back to default
		if logger != nil {
			logger.Warn("Project PHP version is not installed", fmt.Sprintf("%s not found, falling back to default PHP", projectCfg.PHP))
		}
		spec, specErr := NewServiceSpec("php", prefix, system.Info{})
		if specErr != nil {
			return "", specErr
		}
		return spec.BinaryPath, nil
	}

	// No isolation or no specific version, use default behavior
	spec, specErr := NewServiceSpec("php", prefix, system.Info{})
	if specErr != nil {
		return "", specErr
	}
	return spec.BinaryPath, nil
}

/**
 * runBinaryCommand executes a binary with the given arguments.
 */
func runBinaryCommand(binaryPath string, args []string) error {
	// Security: Validate binary path to prevent command injection
	if err := validateBinaryPath(binaryPath); err != nil {
		return fmt.Errorf("security validation failed for binary path: %w", err)
	}

	// SECURITY: Fixed - Using validated binary path and arguments
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}
