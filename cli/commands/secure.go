package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunSecure adds SSL certificate to the current linked project
func RunSecure(args []string) error {
	logger := lib.NewCommandLogger("secure")

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

	// Check if SSL is already enabled
	if proj.Site.SSL {
		logger.Info(fmt.Sprintf("Project '%s' already has SSL enabled", slug))
		return nil
	}

	logger.Info(fmt.Sprintf("Adding SSL certificate to project '%s'", slug))

	// Check for mkcert availability
	mkcertAvailable := checkMkcertAvailable()
	if mkcertAvailable {
		logger.Info("mkcert found - will generate trusted certificates")
	} else {
		logger.Info("mkcert not found - will generate self-signed certificates")
	}

	// Generate SSL certificate paths
	certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
	certPath := filepath.Join(certDir, fmt.Sprintf("%s.crt", slug))
	keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", slug))

	// Ensure certs directory exists
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return logger.Error("Failed to create certificates directory", err.Error())
	}

	// Generate SSL certificate
	var certType lib.SSLCertificateType
	if mkcertAvailable {
		certType, err = generateMkcertCertificate(logger, certPath, keyPath, slug)
	} else {
		certType, err = generateSelfSignedCertificateSimple(logger, certPath, keyPath, slug)
	}

	if err != nil {
		return logger.Error("Failed to generate SSL certificate", err.Error())
	}

	// Update project configuration
	proj.Site.SSL = true

	// Save updated configuration
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
	if err != nil {
		return logger.Error("Failed to ensure project layout", err.Error())
	}

	if err := projects.WriteConfig(proj, layout.ConfigPath, true); err != nil {
		return logger.Error("Failed to save project configuration", err.Error())
	}

	// Restart nginx to apply SSL configuration
	restartNginxIfNeeded(logger)

	// Show success message
	logger.Success(fmt.Sprintf("SSL certificate added successfully to '%s'", slug), "")

	if certType == lib.SSLCertificateTypeMkcert {
		logger.Info("Trusted SSL certificate generated (mkcert certificate is automatically trusted by browsers)")
		logger.Info("No browser warnings expected")
	} else {
		logger.Info("Self-signed SSL certificate generated")
		logger.Warn("Browser may show 'Not Secure' warning", "but connection is encrypted")
	}

	domain := proj.Site.Domain
	if domain == "" {
		domain = slug + ".test"
	}
	logger.Info(fmt.Sprintf("Secure access: https://%s", domain))

	return nil
}

// checkMkcertAvailable checks if mkcert is available and installed
func checkMkcertAvailable() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// generateMkcertCertificate generates an mkcert certificate
func generateMkcertCertificate(logger *lib.Logger, certPath, keyPath, slug string) (lib.SSLCertificateType, error) {
	domain := slug + ".test"

	// Check if mkcert is available
	if _, err := exec.LookPath("mkcert"); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("mkcert not available: %w", err)
	}

	// Generate certificate with mkcert
	cmd := exec.Command("mkcert", "-cert-file", certPath, "-key-file", keyPath, domain)
	if output, err := cmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("mkcert failed: %w, output: %s", err, string(output))
	}

	logger.Success("Trusted SSL certificate generated successfully", fmt.Sprintf("saved to: %s", certPath))
	return lib.SSLCertificateTypeMkcert, nil
}

// generateSelfSignedCertificateSimple generates a self-signed certificate
func generateSelfSignedCertificateSimple(logger *lib.Logger, certPath, keyPath, slug string) (lib.SSLCertificateType, error) {
	// Generate private key
	keyCmd := exec.Command("openssl", "genrsa", "-out", keyPath, "2048")
	if output, err := keyCmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("failed to generate private key: %w, output: %s", err, string(output))
	}

	// Generate certificate
	domain := slug + ".test"
	certCmd := exec.Command("openssl", "req", "-new", "-x509", "-key", keyPath, "-out", certPath, "-days", "365", "-subj", fmt.Sprintf("/C=US/ST=State/L=City/O=Organization/CN=%s", domain))
	if output, err := certCmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("failed to generate certificate: %w, output: %s", err, string(output))
	}

	logger.Success("Self-signed SSL certificate generated successfully", fmt.Sprintf("saved to: %s", certPath))
	return lib.SSLCertificateTypeSelfSigned, nil
}