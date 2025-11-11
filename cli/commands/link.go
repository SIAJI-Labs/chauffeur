package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

// Security: Input validation patterns
var (
	// Safe domain pattern - allows only alphanumeric, hyphens, and dots for domain names
	safeDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	// Safe port range for validation
	minPort = 1
	maxPort = 65535

	// Reserved ports that should not be used without special privileges
	privilegedPorts = 1024
)

// validateDomain ensures domain names are safe and properly formatted
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("domain name too long")
	}

	if !safeDomainPattern.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	// Ensure domain ends with .test for security
	if !strings.HasSuffix(domain, ".test") {
		return fmt.Errorf("domain must end with .test: %s", domain)
	}

	return nil
}

// validatePHPVersion ensures PHP version is in expected format
func validatePHPVersion(version string) error {
	if version == "" {
		return fmt.Errorf("PHP version cannot be empty")
	}

	// Allow only major.minor format (e.g., "7.4", "8.0", "8.1")
	validVersions := map[string]bool{
		"7.4": true, "8.0": true, "8.1": true, "8.2": true, "8.3": true, "8.4": true,
	}

	if !validVersions[version] {
		return fmt.Errorf("unsupported PHP version: %s", version)
	}

	return nil
}

// validatePort ensures port numbers are within valid range
func validatePort(port int) error {
	if port < minPort || port > maxPort {
		return fmt.Errorf("port must be between %d and %d", minPort, maxPort)
	}

	// Note: Privileged ports (< 1024) are allowed but require root privileges
	// This is just documented here - the actual validation allows them

	return nil
}

// validateProjectPath ensures the project path is safe and within allowed directories
func validateProjectPath(path string) error {
	if path == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	// Prevent path traversal attempts
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Check if path actually exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("project path does not exist: %s", absPath)
	}

	if !info.IsDir() {
		return fmt.Errorf("project path must be a directory: %s", absPath)
	}

	return nil
}

// RunLink handles `chauf link` command invocations.
func RunLink(args []string) error {
	var (
		domain    string
		phpVer    string
		ssl       bool
		force     bool
		httpPort  int
		httpsPort int
	)

	logger := lib.NewCommandLogger("link")

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printLinkUsage()
			return nil
		case "--site":
			if i+1 >= len(args) {
				return fmt.Errorf("--site requires a domain value")
			}
			domain = args[i+1]
			i += 2
		case "--php":
			if i+1 >= len(args) {
				return fmt.Errorf("--php requires a version value")
			}
			phpVer = args[i+1]
			i += 2
		case "--ssl":
			ssl = true
			i++
		case "--force":
			force = true
			i++
		case "--http-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--http-port requires a port value")
			}
			portStr := args[i+1]
			val, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid HTTP port: %s", portStr)
			}
			// Security: Validate port range
			if err := validatePort(val); err != nil {
				return fmt.Errorf("invalid HTTP port: %w", err)
			}
			httpPort = val
			i += 2
		case "--https-port":
			if i+1 >= len(args) {
				return fmt.Errorf("--https-port requires a port value")
			}
			portStr := args[i+1]
			val, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid HTTPS port: %s", portStr)
			}
			// Security: Validate port range
			if err := validatePort(val); err != nil {
				return fmt.Errorf("invalid HTTPS port: %w", err)
			}
			httpsPort = val
			i += 2
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for link: %s", arg)
			}
			return fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	// SSL now works with both explicit and default .test domains

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Create port validator
	validator, err := lib.NewPortValidator(cfg)
	if err != nil {
		return fmt.Errorf("create port validator: %w", err)
	}

	// Handle custom port overrides
	if httpPort > 0 {
		validPort, err := validator.SetPortFromCommand("nginx-http", fmt.Sprintf("%d", httpPort))
		if err != nil {
			return err
		}
		cfg.Nginx.HTTPPort = validPort
	}

	if httpsPort > 0 {
		validPort, err := validator.SetPortFromCommand("nginx-https", fmt.Sprintf("%d", httpsPort))
		if err != nil {
			return err
		}
		cfg.Nginx.HTTPSPort = validPort
	}

	// Validate all configured ports
	if err := validator.ValidateAllPorts(); err != nil {
		// If validation fails, check if it's due to conflicts that can be resolved
		if cfg.Ports.ConflictResolution == "fail" {
			return fmt.Errorf("port validation failed: %w", err)
		}
	}

	// Always reload configuration after validation in case ports were updated
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("reload configuration after port validation: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine current directory: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	// Security: Validate project path to prevent path traversal
	if err := validateProjectPath(cwd); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	if phpVer == "" {
		phpVer = cfg.PHP.Default
	}
	if phpVer == "" {
		return fmt.Errorf("no PHP version specified and no default configured")
	}

	// Security: Validate PHP version format
	if err := validatePHPVersion(phpVer); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// Validate that the requested PHP version is installed
	if !projects.IsPHPVersionInstalled(phpVer) {
		return logger.Error(
			fmt.Sprintf("PHP %s is not installed", phpVer),
			fmt.Sprintf("Run 'chauf install php %s' first", phpVer),
		)
	}

	slug := projects.Slugify(filepath.Base(cwd))
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
	if err != nil {
		return err
	}

	// Set default domain if none provided
	if domain == "" {
		domain = slug + ".test"
	}

	// Security: Validate domain format and safety
	if err := validateDomain(domain); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	proj := projects.Config{
		Version: projects.ConfigVersion,
		Path:    cwd,
		PHP:     phpVer,
		Runtime: projects.Runtime{
			PHPFPM: layout.SocketPath,
		},
		CreatedAt: time.Now().UTC(),
	}

	proj.Site = &projects.Site{
		Domain: domain,
		SSL:    ssl,
	}

	if err := projects.WriteConfig(proj, layout.ConfigPath, force); err != nil {
		return err
	}

	// Early validation: Check mkcert availability if SSL is requested
	if proj.Site != nil && proj.Site.SSL {
		logger.Info("Checking for mkcert availability...")
		mkcertAvailable, _ := lib.CheckMkcertAvailable()
		if !mkcertAvailable {
			logger.Warn("mkcert not found - SSL certificates will be self-signed", "")
			logger.Info("For trusted SSL certificates, install mkcert:")
			logger.Info("  # Arch Linux: sudo pacman -S mkcert")
			logger.Info("  # Ubuntu/Debian: sudo apt install mkcert")
			logger.Info("  # Then: mkcert -install")
			logger.Info("")
			logger.Warn("Continue with self-signed certificates?", "This may cause browser warnings")
			logger.Info("Press Ctrl+C to cancel, or continue to proceed...")
			// Give user a moment to consider
			time.Sleep(2 * time.Second)
		} else {
			logger.Info("mkcert found - will generate trusted certificates")
		}
	}

	// Generate nginx template
	templateSpin := lib.NewSpinner("link", "Generating templates")
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		templateSpin.Fail("template engine initialization failed")
		return fmt.Errorf("initialize template engine: %w", err)
	}

	// Detect template type based on project structure
	templateType := templateEngine.DetectTemplateType(cwd)

	// Prepare nginx rendering options
	nginxOptions := templates.NginxConfigOptions{
		HTTPPort:  cfg.Nginx.HTTPPort,
		HTTPSPort: cfg.Nginx.HTTPSPort,
	}

	// Handle SSL certificate generation first if SSL is enabled
	var certType lib.SSLCertificateType
	if proj.Site != nil && proj.Site.SSL {
		certBase := proj.Site.Domain
		if certBase == "" {
			certBase = slug
		}
		certDir := filepath.Join(cfg.WorkspaceDir, "nginx", "certs")
		certPath := filepath.Join(certDir, fmt.Sprintf("%s.crt", certBase))
		keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", certBase))

		// Set SSL paths for nginx options
		nginxOptions.SSLCertPath = certPath
		nginxOptions.SSLKeyPath = keyPath

		// Generate SSL certificate with dedicated spinner
		sslSpin := lib.NewSpinner("link", "Generating SSL certificates")
		generatedCertType, err := generateSSLCertificate(logger, certPath, keyPath, certBase)
		if err != nil {
			sslSpin.Fail("SSL certificate generation failed")
			return fmt.Errorf("generate SSL certificate: %w", err)
		}
		certType = generatedCertType
		sslSpin.Success("SSL certificates generated")
	}

	// Generate and write nginx configuration (with SSL paths if applicable)
	if err := templateEngine.WriteNginxConfig(proj, layout, templateType, nginxOptions); err != nil {
		templateSpin.Fail("nginx configuration generation failed")
		logger.Warn("Failed to generate nginx configuration", err.Error())
		return fmt.Errorf("generate nginx configuration: %w", err)
	} else {
		templateSpin.Success("nginx templates generated")
	}

	// Provide SSL usage guidance if SSL was generated
	if proj.Site != nil && proj.Site.SSL {
		provideSSLUsageGuidance(logger, proj.Site.Domain, cfg.Nginx.HTTPSPort, certType)
	}

	logger.PrintSection(fmt.Sprintf("Project linked as %s", slug))
	logger.Info(fmt.Sprintf("Path: %s", cwd))
	logger.Info(fmt.Sprintf("Config: %s", layout.ConfigPath))
	logger.Info(fmt.Sprintf("PHP: %s", phpVer))
	logger.Info(fmt.Sprintf("Template: %s", templateType))
	logger.Info(fmt.Sprintf("Nginx HTTP: %d", cfg.Nginx.HTTPPort))
	logger.Info(fmt.Sprintf("Nginx HTTPS: %d", cfg.Nginx.HTTPSPort))

	// Access URLs - always show domain (now default to <slug>.test)
	if proj.Site != nil && proj.Site.Domain != "" {
		if strings.Contains(proj.Site.Domain, ".test") {
			logger.Info(fmt.Sprintf("Domain: %s (default .test domain)", proj.Site.Domain))
		} else {
			logger.Info(fmt.Sprintf("Domain: %s (custom domain)", proj.Site.Domain))
		}

		// Show URLs with actual ports
		httpURL := fmt.Sprintf("http://%s", proj.Site.Domain)
		if cfg.Nginx.HTTPPort != 80 {
			httpURL = fmt.Sprintf("http://%s:%d", proj.Site.Domain, cfg.Nginx.HTTPPort)
		}
		logger.Info(fmt.Sprintf("Access: %s", httpURL))

		if proj.Site.SSL {
			httpsURL := fmt.Sprintf("https://%s", proj.Site.Domain)
			if cfg.Nginx.HTTPSPort != 443 {
				httpsURL = fmt.Sprintf("https://%s:%d", proj.Site.Domain, cfg.Nginx.HTTPSPort)
			}
			logger.Info(fmt.Sprintf("Access Secure: %s", httpsURL))
		}

		logger.Info("Note: Use --site <custom-domain> to override default domain")
		if httpPort > 0 || httpsPort > 0 {
			logger.Info("Custom ports applied via --http-port/--https-port flags")
		}
	}

	return nil
}

func printLinkUsage() {
	fmt.Print(`Chauffeur Project Linking

Usage:
  chauf link [--site <domain>] [--ssl] [--php <version>] [--http-port <port>] [--https-port <port>] [--force]

Flags:
  --site <domain>           Register a local domain for the project (default: <slug>.test).
  --ssl                     Enable internal TLS for the domain.
  --php <version>           Override the PHP version for this project (default: global default).
  --http-port <port>        Override Nginx HTTP port for this project (default: from config).
  --https-port <port>       Override Nginx HTTPS port for this project (default: from config).
  --force                   Overwrite existing project configuration.

Port Management:
  If specified ports are already in use, Chauffeur will:
    - Prompt for alternative ports (default behavior)
    - Auto-resolve to available ports (if "conflict_resolution: auto" in config)
    - Fail with error if "conflict_resolution: fail" in config

Note:
  When --site is not specified, the project is automatically assigned a .test domain
  based on the project directory name (e.g., "my-project" -> "my-project.test").
`)
}

// Import SSLCertificateType from lib package

// generateSSLCertificate creates an SSL certificate using the best available method
func generateSSLCertificate(logger *lib.Logger, certPath, keyPath, domain string) (lib.SSLCertificateType, error) {
	// Check if certificate already exists
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			// Both certificate and key exist, skip generation
			// Determine type by checking if it's an mkcert certificate
			if lib.IsMkcertCertificate(certPath) {
				logger.Info(fmt.Sprintf("Existing mkcert certificate found (domain: %s)", domain))
				return lib.SSLCertificateTypeMkcert, nil
			}
			logger.Info(fmt.Sprintf("Existing self-signed certificate found (domain: %s)", domain))
			return lib.SSLCertificateTypeSelfSigned, nil
		}
	}

	// Ensure certificate directory exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("create certificate directory: %w", err)
	}

	// Check for mkcert availability (we already did early validation, but need the command)
	_, mkcertCmd := lib.CheckMkcertAvailable()
	if mkcertCmd != "" {
		logger.Info(fmt.Sprintf("Generating trusted SSL certificate using mkcert for domain: %s", domain))

		// Change to workspace directory for mkcert generation
		workspaceDir := filepath.Dir(filepath.Dir(certPath)) // ~/.chauffeur
		if err := os.Chdir(workspaceDir); err != nil {
			logger.Warn("Failed to change to workspace directory", err.Error())
		}

		// Generate certificate using mkcert
		cmd := exec.Command(mkcertCmd, domain)
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.Warn("mkcert generation failed, falling back to self-signed", fmt.Sprintf("error: %v, output: %s", err, string(output)))
			return generateSelfSignedCertificate(logger, certPath, keyPath, domain)
		}

		// Move mkcert-generated files to expected locations
		mkcertCertPath := fmt.Sprintf("%s.pem", domain)
		mkcertKeyPath := fmt.Sprintf("%s-key.pem", domain)

		if err := lib.MoveFile(mkcertCertPath, certPath); err != nil {
			return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("move mkcert certificate: %w", err)
		}
		if err := lib.MoveFile(mkcertKeyPath, keyPath); err != nil {
			return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("move mkcert private key: %w", err)
		}

		logger.Success("Trusted SSL certificate generated successfully", fmt.Sprintf("saved to: %s", certPath))
		return lib.SSLCertificateTypeMkcert, nil
	}

	// Fall back to self-signed certificates (mkcert not available)
	return generateSelfSignedCertificate(logger, certPath, keyPath, domain)
}

// generateSelfSignedCertificate creates a self-signed SSL certificate using OpenSSL (fallback)
func generateSelfSignedCertificate(logger *lib.Logger, certPath, keyPath, domain string) (lib.SSLCertificateType, error) {
	logger.Info(fmt.Sprintf("Generating self-signed SSL certificate for domain: %s", domain))

	// Generate private key
	keyCmd := exec.Command("openssl", "genrsa", "-out", keyPath, "2048")
	if output, err := keyCmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("generate private key: %w, output: %s", err, string(output))
	}

	// Generate certificate signing request (CSR)
	csrCmd := exec.Command("openssl", "req", "-new", "-key", keyPath, "-out", "/tmp/chauffeur.csr",
		"-subj", fmt.Sprintf("/C=US/ST=State/L=City/O=Chauffeur/OU=Development/CN=%s", domain))
	if output, err := csrCmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("generate CSR: %w, output: %s", err, string(output))
	}

	// Generate self-signed certificate
	certCmd := exec.Command("openssl", "x509", "-req", "-days", "365", "-in", "/tmp/chauffeur.csr", "-signkey", keyPath, "-out", certPath)
	if output, err := certCmd.CombinedOutput(); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("generate certificate: %w, output: %s", err, string(output))
	}

	// Clean up temporary CSR file
	if err := os.Remove("/tmp/chauffeur.csr"); err != nil {
		logger.Warn("Failed to clean up temporary CSR file", err.Error())
	}

	// Set appropriate permissions
	if err := os.Chmod(keyPath, 0o600); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("set private key permissions: %w", err)
	}
	if err := os.Chmod(certPath, 0o644); err != nil {
		return lib.SSLCertificateTypeSelfSigned, fmt.Errorf("set certificate permissions: %w", err)
	}

	logger.Success("Self-signed SSL certificate generated successfully", fmt.Sprintf("saved to: %s", certPath))
	return lib.SSLCertificateTypeSelfSigned, nil
}

// provideSSLUsageGuidance educates users about using SSL certificates for development
func provideSSLUsageGuidance(logger *lib.Logger, domain string, httpsPort int, certType lib.SSLCertificateType) {
	logger.PrintSection("SSL Certificate Usage")

	httpsURL := fmt.Sprintf("https://%s", domain)
	if httpsPort != 443 {
		httpsURL = fmt.Sprintf("https://%s:%d", domain, httpsPort)
	}

	switch certType {
	case lib.SSLCertificateTypeMkcert:
		logger.Success("Trusted SSL certificate generated", "mkcert certificate is automatically trusted by browsers")
		logger.Info("✓ No browser warnings expected")
		logger.Info("✓ Can access directly with curl or browsers")
		logger.Info(fmt.Sprintf("  Direct access: %s", httpsURL))
		logger.Info("Certificate type: mkcert (trusted)")

	case lib.SSLCertificateTypeSelfSigned:
		logger.Warn("Self-signed certificate detected", "Browsers and curl will reject this certificate by default")
		logger.Info("Development Options:")
		logger.Info("  1. Use curl with -k or --insecure flag:")
		logger.Info(fmt.Sprintf("     curl -k %s", httpsURL))

		logger.Info("  2. For browser testing:")
		logger.Info("     - Click 'Advanced' → 'Proceed to domain (unsafe)'")
		logger.Info("     - Or import the certificate into your system trust store")

		logger.Info("  3. For better development experience (optional):")
		logger.Info("     Install mkcert for trusted local development certificates:")
		logger.Info("     # Arch Linux")
		logger.Info("     sudo pacman -S mkcert")
		logger.Info("     # Ubuntu/Debian")
		logger.Info("     sudo apt install mkcert")
		logger.Info("     $ mkcert -install")
		logger.Info(fmt.Sprintf("     $ mkcert %s", domain))

		logger.Info("Certificate type: self-signed (requires special handling)")
	}

	logger.Info("Certificate location:")
	certPath := filepath.Join(os.Getenv("HOME"), ".chauffeur", "nginx", "certs", fmt.Sprintf("%s.crt", domain))
	logger.Info(fmt.Sprintf("  Certificate: %s", certPath))
}
