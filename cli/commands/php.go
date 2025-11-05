package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

// RunPHP routes php subcommands or falls back to executing the default PHP binary.
func RunPHP(args []string) error {
	if len(args) == 0 {
		return runPHPBinary(nil)
	}

	switch args[0] {
	case "--help", "-h":
		printPHPUsage()
		return nil
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("php use requires <version>")
		}
		return runPHPUse(args[1])
	case "isolate":
		if len(args) < 2 {
			return fmt.Errorf("php isolate requires <version>")
		}
		return runPHPIsolate(args[1])
	default:
		return runPHPBinary(args)
	}
}

func runPHPUse(version string) error {
	if !installers.IsPHPVersionSupported(version) {
		return fmt.Errorf("PHP version %s is not supported. Supported versions: %s", version, installers.GetSupportedVersionsList())
	}

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	binary := filepath.Join(prefix, "php", version, "bin", "php")
	if _, err := os.Stat(binary); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("PHP %s is not installed. Run 'chauf install php %s' first.\n", version, version)
			return fmt.Errorf("php %s not installed", version)
		}
		return fmt.Errorf("check php binary: %w", err)
	}

	current, err := config.GetDefaultPHPVersion()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Check if config file actually exists to avoid false positive "already default"
	configFile := filepath.Join(prefix, "config", "chauffeur.yaml")
	if _, err := os.Stat(configFile); err == nil && current == version {
		fmt.Printf("PHP %s is already the default.\n", version)
		return nil
	}

	if err := config.SetDefaultPHPVersion(version); err != nil {
		return fmt.Errorf("update configuration: %w", err)
	}

	if err := installers.UpdateDefaultPHPShim(prefix, version); err != nil {
		return fmt.Errorf("update php shim: %w", err)
	}

	fmt.Printf("Default PHP version updated to %s.\n", version)
	return nil
}

func runPHPIsolate(version string) error {
	if !installers.IsPHPVersionSupported(version) {
		return fmt.Errorf("PHP version %s is not supported. Supported versions: %s", version, installers.GetSupportedVersionsList())
	}

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	binary := filepath.Join(prefix, "php", version, "bin", "php")
	if _, err := os.Stat(binary); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("PHP %s is not installed. Run 'chauf install php %s' first.\n", version, version)
			return fmt.Errorf("php %s not installed", version)
		}
		return fmt.Errorf("check php binary: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine current directory: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	projectCfg, layout, err := projects.FindByPath(cfg.ProjectsDir, cwd)
	if err != nil {
		if errors.Is(err, projects.ErrProjectNotFound) {
			return fmt.Errorf("project is not linked. Run 'chauf link' in this directory first")
		}
		return fmt.Errorf("load project configuration: %w", err)
	}

	projectCfg.PHP = version
	if projectCfg.Runtime.PHPFPM == "" {
		projectCfg.Runtime.PHPFPM = layout.SocketPath
	}

	if err := projects.WriteConfig(projectCfg, layout.ConfigPath, true); err != nil {
		return fmt.Errorf("update project configuration: %w", err)
	}

	// Update nginx template to reflect new PHP version
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("initialize template engine: %w", err)
	}

	// Detect template type based on project structure
	templateType := templateEngine.DetectTemplateType(projectCfg.Path)
	
	// Update nginx configuration with new PHP version
	if err := templateEngine.WriteNginxConfig(projectCfg, layout, templateType); err != nil {
		return fmt.Errorf("update nginx configuration: %w", err)
	}

	fmt.Printf("Project PHP version pinned to %s (config: %s)\n", version, layout.ConfigPath)
	fmt.Printf("Nginx configuration updated for new PHP version %s\n", version)
	return nil
}

func printPHPUsage() {
	fmt.Print(`Chauffeur PHP Commands

Usage:
  chauf php [args...]       Execute the default PHP CLI with passthrough args.
  chauf php use <version>   Set the default PHP version.
  chauf php isolate <version>
                             Pin the current project to a specific PHP version.
`)
}

func runPHPBinary(args []string) error {
	if err := RunServiceCommand("php", args); err != nil {
		return err
	}
	return nil
}
