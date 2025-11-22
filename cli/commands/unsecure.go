package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunUnsecure removes SSL certificate from the current linked project
func RunUnsecure(args []string) error {
	logger := lib.NewCommandLogger("unsecure")

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return logger.Error("Failed to get current directory", err.Error())
	}

	// Load global configuration
	cfg, err := config.Load()
	if err != nil {
		return logger.Error("Failed to load global configuration", err.Error())
	}

	// Find project in current directory
	proj, _, err := projects.FindByPath(cfg.ProjectsDir, wd)
	if err != nil {
		if os.IsNotExist(err) {
			return logger.Error(
				"No linked project found in current directory",
				"Run 'chauf link' first to link this directory",
			)
		}
		return logger.Error("Failed to load project configuration", err.Error())
	}

	// Generate project slug from directory name
	slug := projects.Slugify(filepath.Base(wd))

	// Check if SSL is already disabled
	if !proj.Site.SSL {
		logger.Info(fmt.Sprintf("Project '%s' does not have SSL enabled", slug))
		return nil
	}

	logger.Info(fmt.Sprintf("Removing SSL certificate from project '%s'", slug))

	// Update project configuration to disable SSL
	proj.Site.SSL = false

	// Save updated configuration
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
	if err != nil {
		return logger.Error("Failed to ensure project layout", err.Error())
	}

	if err := projects.WriteConfig(proj, layout.ConfigPath, true); err != nil {
		return logger.Error("Failed to save project configuration", err.Error())
	}

	// Remove SSL certificate files
	if err := removeSSLCertificates(cfg.WorkspaceDir, slug); err != nil {
		logger.Warn("Failed to remove SSL certificate files", err.Error())
	}

	// Restart nginx to apply configuration changes
	restartNginxIfNeeded(logger)

	// Show success message
	logger.Success(fmt.Sprintf("SSL certificate removed successfully from '%s'", slug), "")

	domain := proj.Site.Domain
	if domain == "" {
		domain = slug + ".test"
	}
	logger.Info(fmt.Sprintf("HTTP access: http://%s", domain))

	return nil
}

// removeSSLCertificates removes SSL certificate files for the project
func removeSSLCertificates(workspaceDir, slug string) error {
	certsDir := filepath.Join(workspaceDir, "nginx", "certs")
	certFile := filepath.Join(certsDir, fmt.Sprintf("%s.crt", slug))
	keyFile := filepath.Join(certsDir, fmt.Sprintf("%s.key", slug))

	// Remove certificate files
	if err := os.Remove(certFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate file: %w", err)
	}
	if err := os.Remove(keyFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove private key file: %w", err)
	}

	return nil
}