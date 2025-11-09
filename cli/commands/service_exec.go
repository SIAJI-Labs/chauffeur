package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunServiceCommand executes an installed Chauffeur-managed service with passthrough args.
 *
 * @param name Service identifier (e.g., "nginx", "caddy").
 * @param args Arguments forwarded to the underlying binary.
 * @return error when the service is unknown, missing, or execution fails.
 */
func RunServiceCommand(name string, args []string) error {
	if !IsKnownService(name) {
		return fmt.Errorf("unknown service %s", name)
	}

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
		binaryPath, err := getProjectAwarePHPBinary(prefix)
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
func getProjectAwarePHPBinary(prefix string) (string, error) {
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
		} else {
			// Isolated PHP version is not installed, warn and fall back to default
			fmt.Printf("Warning: Project PHP version %s is not installed. Using default PHP instead.\n", projectCfg.PHP)
			spec, specErr := NewServiceSpec("php", prefix, system.Info{})
			if specErr != nil {
				return "", specErr
			}
			return spec.BinaryPath, nil
		}
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
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}
