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
	"github.com/siaji/chauffeur/cli/lib"
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
	case "current":
		return runPHPCurrent()
	case "list":
		return runPHPList()
	default:
		return runPHPBinary(args)
	}
}

func runPHPUse(version string) error {
	logger := lib.NewCommandLogger("php")

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
			return logger.Error(
				fmt.Sprintf("PHP %s is not installed", version),
				fmt.Sprintf("Run 'chauf install php %s' first", version),
			)
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
		logger.Info(fmt.Sprintf("PHP %s is already the default", version))
		return nil
	}

	if err := config.SetDefaultPHPVersion(version); err != nil {
		return fmt.Errorf("update configuration: %w", err)
	}

	if err := installers.UpdateDefaultPHPShim(prefix, version); err != nil {
		return fmt.Errorf("update php shim: %w", err)
	}

	logger.Success(fmt.Sprintf("Default PHP version updated to %s", version), "")
	return nil
}

func runPHPIsolate(version string) error {
	logger := lib.NewCommandLogger("php")

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
			return logger.Error(
				fmt.Sprintf("PHP %s is not installed", version),
				fmt.Sprintf("Run 'chauf install php %s' first", version),
			)
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
	// Always update the PHP-FPM socket path when changing PHP versions
	// to ensure it points to the correct PHP version socket
	newSocketPath := filepath.Join(prefix, "php", version, "runtime", "php-fpm", "php-fpm.sock")
	projectCfg.Runtime.PHPFPM = newSocketPath

	// Also update the FPM socket if it exists
	if projectCfg.Runtime.FPM != nil {
		projectCfg.Runtime.FPM.Socket = newSocketPath
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
	nginxOptions := templates.NginxConfigOptions{
		HTTPPort:  cfg.Nginx.HTTPPort,
		HTTPSPort: cfg.Nginx.HTTPSPort,
	}
	if projectCfg.Site != nil && projectCfg.Site.SSL {
		certBase := projectCfg.Site.Domain
		if certBase == "" {
			certBase = filepath.Base(layout.Root)
		}
		certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
		nginxOptions.SSLCertPath = filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
		nginxOptions.SSLKeyPath = filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))
	}

	if err := templateEngine.WriteNginxConfig(projectCfg, layout, templateType, nginxOptions); err != nil {
		return fmt.Errorf("update nginx configuration: %w", err)
	}

	logger.Success(fmt.Sprintf("Project PHP version pinned to %s", version), layout.ConfigPath)
	logger.Info(fmt.Sprintf("Nginx configuration updated for PHP %s", version))

	// Show access URLs for the isolated project
	domain := "localhost"
	if projectCfg.Site != nil && projectCfg.Site.Domain != "" {
		domain = projectCfg.Site.Domain
	}

	httpURL := fmt.Sprintf("http://%s", domain)
	if cfg.Nginx.HTTPPort != 80 {
		httpURL = fmt.Sprintf("http://%s:%d", domain, cfg.Nginx.HTTPPort)
	}
	logger.Info(fmt.Sprintf("Access: %s", httpURL))

	if projectCfg.Site != nil && projectCfg.Site.SSL {
		httpsURL := fmt.Sprintf("https://%s", domain)
		if cfg.Nginx.HTTPSPort != 443 {
			httpsURL = fmt.Sprintf("https://%s:%d", domain, cfg.Nginx.HTTPSPort)
		}
		logger.Info(fmt.Sprintf("Access Secure: %s", httpsURL))
	}

	return nil
}

func runPHPList() error {
	logger := lib.NewCommandLogger("php")

	supportedVersions := installers.GetSupportedPHPVersions()
	workspaceDir, err := workspace.Dir()
	if err != nil {
		return fmt.Errorf("get workspace directory: %w", err)
	}

	logger.Success("Supported PHP versions", "Checking installation status...")

	for _, version := range supportedVersions {
		phpPath := filepath.Join(workspaceDir, "php", version.Version, "bin", "php")
		status := "not installed"

		if _, err := os.Stat(phpPath); err == nil {
			// Check if it's the default version
			defaultVersion, err := config.GetDefaultPHPVersion()
			if err == nil && defaultVersion == version.Version {
				status = "active"
			} else {
				status = "installed"
			}
		}

		statusIcon := "❌"
		if status == "active" {
			statusIcon = "✅ (active)"
		} else if status == "installed" {
			statusIcon = "✅"
		}

		logger.Info(fmt.Sprintf("PHP %s %s", version.Version, statusIcon))
	}

	return nil
}

func runPHPCurrent() error {
	logger := lib.NewCommandLogger("php current")

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return logger.Error("Failed to get current directory", err.Error())
	}

	// Get global configuration
	globalConfig, err := config.Load()
	if err != nil {
		return logger.Error("Failed to load global configuration", err.Error())
	}
	globalPHP := globalConfig.PHP.Default

	// Try to find project in current directory
	projectsDir := globalConfig.ProjectsDir

	proj, _, err := projects.FindByPath(projectsDir, wd)
	if err != nil {
		// No project found, show only global PHP
		logger.Info("No project detected in current directory")
		logger.Info(fmt.Sprintf("Global PHP: %s (default)", globalPHP))

		workspaceDir, err := workspace.Dir()
		if err != nil {
			return logger.Error("Failed to get workspace directory", err.Error())
		}
		phpBinary := filepath.Join(workspaceDir, "php", globalPHP, "bin", "php")
		logger.Info(fmt.Sprintf("PHP binary: %s", phpBinary))
		return nil
	}

	// Project found
	projectPHP := proj.PHP
	logger.Info(fmt.Sprintf("Project: %s", proj.Path))
	logger.Info(fmt.Sprintf("Project PHP: %s", projectPHP))

	if proj.PHP != globalPHP {
		logger.Info(fmt.Sprintf("Global PHP: %s (default)", globalPHP))
	} else {
		logger.Info(fmt.Sprintf("Global PHP: %s (same as project)", globalPHP))
	}

	// Show PHP binary path
	workspaceDir, err := workspace.Dir()
	if err != nil {
		return logger.Error("Failed to get workspace directory", err.Error())
	}
	phpBinary := filepath.Join(workspaceDir, "php", projectPHP, "bin", "php")
	logger.Info(fmt.Sprintf("PHP binary: %s", phpBinary))

	return nil
}

func printPHPUsage() {
	logger := lib.NewCommandLogger("php")
	logger.PrintBlock(`Chauffeur PHP Commands

Usage:
  chauf php [args...]       Execute the default PHP CLI with passthrough args.
  chauf php use <version>   Set the default PHP version.
  chauf php isolate <version>
                             Pin the current project to a specific PHP version.
  chauf php current         Show current PHP version for directory or global default.
  chauf php list            List all supported PHP versions and their status.
`)
}

func runPHPBinary(args []string) error {
	if err := RunServiceCommand("php", args); err != nil {
		return err
	}
	return nil
}
