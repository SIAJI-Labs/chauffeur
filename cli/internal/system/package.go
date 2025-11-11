package system

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// PackageManager represents a system package manager
type PackageManager string

const (
	Pacman  PackageManager = "pacman"
	Apt     PackageManager = "apt"
	Yum     PackageManager = "yum"
	Dnf     PackageManager = "dnf"
	Zypper  PackageManager = "zypper"
	Unknown PackageManager = "unknown"
)

const (
	// DNSProbeDomain is the synthetic domain used to verify dnsmasq-based resolution
	DNSProbeDomain = "chauffeur-dns-probe.test"

	// dnsProbeTimeout bounds how long we wait when probing for DNS resolution
	dnsProbeTimeout = 2 * time.Second
)

const (
	networkManagerConfDir     = "/etc/NetworkManager/conf.d"
	networkManagerDnsConf     = "/etc/NetworkManager/conf.d/90-chauffeur-dns.conf"
	networkManagerDnsmasqDir  = "/etc/NetworkManager/dnsmasq.d"
	networkManagerDnsmasqConf = "/etc/NetworkManager/dnsmasq.d/chauffeur.conf"
	systemdResolvedDropInDir  = "/etc/systemd/resolved.conf.d"
	systemdResolvedDropInPath = "/etc/systemd/resolved.conf.d/90-chauffeur-disable-stub.conf"
	defaultDnsmasqConfigBlock = `# Chauffeur local development resolver
# Redirect all *.test domains to localhost
address=/.test/127.0.0.1
# Only listen locally
listen-address=127.0.0.1
bind-interfaces
`
	defaultNetworkManagerBlock = `[main]
dns=dnsmasq
`
	defaultResolvedDropIn = `[Resolve]
DNSStubListener=no
`
)

// Package represents a system package information
type Package struct {
	Name        string
	Description string
	PackageName string // actual package name for the package manager
}

// Security: Input validation patterns
var (
	// Safe package name pattern - allows only alphanumeric, dots, hyphens, and underscores
	safePackageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Safe command pattern for system commands - restricts to known safe commands
	safeSystemCommands = map[string]bool{
		"pacman": true, "apt": true, "yum": true, "dnf": true, "zypper": true,
		"dpkg": true, "rpm": true, "systemctl": true, "pgrep": true,
		"install": true, "rm": true, "tee": true, "mkdir": true,
	}
)

// validatePackageName ensures package names are safe to prevent command injection
func validatePackageName(pkgName string) error {
	if pkgName == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	if len(pkgName) > 100 {
		return fmt.Errorf("package name too long")
	}

	if !safePackageNamePattern.MatchString(pkgName) {
		return fmt.Errorf("invalid package name format: %s", pkgName)
	}

	return nil
}

// validateSystemPath ensures file paths are safe for system operations
func validateSystemPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Prevent path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Restrict to system configuration directories
	allowedPaths := []string{
		"/etc/", "/usr/", "/opt/", "/var/", "/home/",
	}

	allowed := false
	for _, allowedPath := range allowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("path not in allowed system directories: %s", path)
	}

	return nil
}

// RequiredPackages represents packages needed for Chauffeur functionality
var RequiredPackages = map[PackageManager][]Package{
	Pacman: {
		{
			Name:        "dnsmasq",
			Description: "Lightweight DNS forwarder for local domain resolution",
			PackageName: "dnsmasq",
		},
	},
	Apt: {
		{
			Name:        "dnsmasq",
			Description: "Lightweight DNS forwarder for local domain resolution",
			PackageName: "dnsmasq",
		},
		{
			Name:        "resolvectl",
			Description: "Systemd resolve control utility",
			PackageName: "systemd-resolved",
		},
	},
	// TODO: Add support for other package managers
}

/**
 * DetectPackageManager identifies the package manager available on the system.
 *
 * @return PackageManager type and an error if detection fails.
 */
func DetectPackageManager() PackageManager {
	// Check for various package managers
	pkgManagers := map[PackageManager][]string{
		Pacman: {"pacman", "--version"},
		Apt:    {"apt", "--version"},
		Yum:    {"yum", "--version"},
		Dnf:    {"dnf", "--version"},
		Zypper: {"zypper", "--version"},
	}

	for pm, args := range pkgManagers {
		if _, err := exec.LookPath(string(args[0])); err == nil {
			// Verify the command works
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err == nil {
				return pm
			}
		}
	}

	return Unknown
}

/**
 * DetectArchPackageManager detects and returns the preferred package manager for Arch Linux.
 * Checks for yay/paru first (AUR helpers), then falls back to pacman.
 *
 * @return package manager command name or empty string if not found
 */
func DetectArchPackageManager() string {
	// Check for AUR helpers first
	aurHelpers := []string{"yay", "paru"}
	for _, helper := range aurHelpers {
		if _, err := exec.LookPath(helper); err == nil {
			return helper
		}
	}

	// Fallback to pacman
	if _, err := exec.LookPath("pacman"); err == nil {
		return "pacman"
	}

	return ""
}

/**
 * IsPackageInstalled checks if a specific package is installed on the system.
 *
 * @param pkg Package to check
 * @return true if package is installed, false otherwise
 */
func IsPackageInstalled(pkg Package) bool {
	pm := DetectPackageManager()

	switch pm {
	case Pacman:
		cmd := exec.Command("pacman", "-Q", pkg.PackageName)
		err := cmd.Run()
		return err == nil

	case Apt:
		cmd := exec.Command("dpkg", "-l", pkg.PackageName)
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run() == nil

	case Yum, Dnf:
		cmd := exec.Command("rpm", "-q", pkg.PackageName)
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run() == nil

	case Zypper:
		cmd := exec.Command("zypper", "search", "-i", pkg.PackageName)
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run() == nil

	default:
		return false
	}
}

/**
 * InstallPackage attempts to install a package using the system package manager.
 *
 * @param pkg Package to install
 * @return error if installation fails
 */
func InstallPackage(pkg Package) error {
	// Security: Validate package name to prevent command injection
	if err := validatePackageName(pkg.PackageName); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	pm := DetectPackageManager()

	switch pm {
	case Pacman:
		// SECURITY: Fixed - Using validated package name with hardcoded command arguments
		args := []string{"pacman", "-S", "--noconfirm", pkg.PackageName}
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case Apt:
		// SECURITY: Fixed - Using validated package name with hardcoded command arguments
		args := []string{"apt", "install", "-y", pkg.PackageName}
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case Yum:
		// SECURITY: Fixed - Using validated package name with hardcoded command arguments
		args := []string{"yum", "install", "-y", pkg.PackageName}
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case Dnf:
		// SECURITY: Fixed - Using validated package name with hardcoded command arguments
		args := []string{"dnf", "install", "-y", pkg.PackageName}
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case Zypper:
		// SECURITY: Fixed - Using validated package name with hardcoded command arguments
		args := []string{"zypper", "install", "-y", pkg.PackageName}
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}
}

/**
 * GetMissingPackages returns a list of required packages that are not installed.
 *
 * @return slice of missing packages
 */
func GetMissingPackages() []Package {
	pm := DetectPackageManager()
	var missing []Package

	if packages, ok := RequiredPackages[pm]; ok {
		for _, pkg := range packages {
			if !IsPackageInstalled(pkg) {
				missing = append(missing, pkg)
			}
		}
	}

	return missing
}

/**
 * PromptUserForPackages presents missing packages to the user and asks for installation consent.
 *
 * @param missing slice of missing packages
 * @return true if user consents to installation, false otherwise
 */
func PromptUserForPackages(missing []Package) bool {
	if len(missing) == 0 {
		return true
	}

	fmt.Printf("The following packages are required for Chauffeur local domain resolution:\n")
	for _, pkg := range missing {
		fmt.Printf("  - %s (%s): %s\n", pkg.Name, pkg.PackageName, pkg.Description)
	}

	fmt.Printf("\nWould you like to install these packages? [y/N]: ")

	// SENSITIVE: User input confirmation - package installation consent
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

/**
 * IsCommandAvailable checks if a command is available in PATH
 *
 * @param cmd command name to check
 * @return true if command is available, false otherwise
 */
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

/**
 * IsNetworkManagerDnsmasqRunning checks if NetworkManager is running dnsmasq
 *
 * @return true if NetworkManager's dnsmasq is running and handling DNS resolution
 */
func IsNetworkManagerDnsmasqRunning() bool {
	// Check if NetworkManager's dnsmasq is running by looking for the process
	cmd := exec.Command("pgrep", "-f", "dnsmasq.*NetworkManager")
	err := cmd.Run()
	return err == nil
}

/**
 * IsDnsmasqAvailable checks if dnsmasq is available for local DNS resolution
 * This checks both standalone dnsmasq and NetworkManager's dnsmasq
 *
 * @return true if dnsmasq is available for resolving local domains
 */
func IsDnsmasqAvailable() bool {
	// Check if NetworkManager's dnsmasq is running
	if IsNetworkManagerDnsmasqRunning() {
		return true
	}

	// Check if standalone dnsmasq command is available
	return IsCommandAvailable("dnsmasq")
}

// enableNetworkManagerDnsmasqIntegration configures NetworkManager to run dnsmasq with Chauffeur settings.
func enableNetworkManagerDnsmasqIntegration() error {
	if !IsNetworkManagerAvailable() {
		return fmt.Errorf("NetworkManager is not available; cannot enable dnsmasq integration automatically")
	}

	// Security: Validate and create NetworkManager directories
	if err := validateSystemPath(networkManagerConfDir); err != nil {
		return fmt.Errorf("security validation failed for %s: %w", networkManagerConfDir, err)
	}
	args := []string{"install", "-d", "-m", "755", networkManagerConfDir}
	if err := exec.Command("sudo", args...).Run(); err != nil {
		return fmt.Errorf("failed to prepare NetworkManager conf.d: %w", err)
	}
	if err := writeFileWithSudo(networkManagerDnsConf, defaultNetworkManagerBlock); err != nil {
		return fmt.Errorf("failed to configure NetworkManager dnsmasq plugin: %w", err)
	}

	if err := validateSystemPath(networkManagerDnsmasqDir); err != nil {
		return fmt.Errorf("security validation failed for %s: %w", networkManagerDnsmasqDir, err)
	}
	args = []string{"install", "-d", "-m", "755", networkManagerDnsmasqDir}
	if err := exec.Command("sudo", args...).Run(); err != nil {
		return fmt.Errorf("failed to prepare NetworkManager dnsmasq.d: %w", err)
	}
	if err := writeFileWithSudo(networkManagerDnsmasqConf, defaultDnsmasqConfigBlock); err != nil {
		return fmt.Errorf("failed to configure NetworkManager dnsmasq rules: %w", err)
	}

	if err := disableSystemdResolvedStub(); err != nil {
		return err
	}

	// SECURITY: Fixed - Using hardcoded command arguments
	args = []string{"systemctl", "reload", "NetworkManager"}
	if err := exec.Command("sudo", args...).Run(); err != nil {
		return fmt.Errorf("failed to reload NetworkManager: %w", err)
	}

	for i := 0; i < 10; i++ {
		if IsNetworkManagerDnsmasqRunning() {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("NetworkManager dnsmasq did not start after reconfiguration; check NetworkManager logs")
}

// disableSystemdResolvedStub ensures systemd-resolved frees 127.0.0.53.
func disableSystemdResolvedStub() error {
	// Security: Validate systemd directory path
	if err := validateSystemPath(systemdResolvedDropInDir); err != nil {
		return fmt.Errorf("security validation failed for %s: %w", systemdResolvedDropInDir, err)
	}

	// SECURITY: Fixed - Using validated path with hardcoded command arguments
	args := []string{"install", "-d", "-m", "755", systemdResolvedDropInDir}
	if err := exec.Command("sudo", args...).Run(); err != nil {
		return fmt.Errorf("failed to prepare systemd-resolved drop-in directory: %w", err)
	}

	if err := writeFileWithSudo(systemdResolvedDropInPath, defaultResolvedDropIn); err != nil {
		return fmt.Errorf("failed to configure systemd-resolved drop-in: %w", err)
	}

	// SECURITY: Fixed - Using hardcoded command arguments
	args = []string{"systemctl", "restart", "systemd-resolved"}
	if err := exec.Command("sudo", args...).Run(); err != nil {
		return fmt.Errorf("failed to restart systemd-resolved: %w", err)
	}

	return nil
}

// CleanupNetworkManagerDnsmasqConfiguration removes Chauffeur NetworkManager overrides.
func CleanupNetworkManagerDnsmasqConfiguration() error {
	var (
		reloadedNM    bool
		restartedStub bool
	)

	if fileExists(networkManagerDnsConf) {
		if err := removeFileWithSudo(networkManagerDnsConf); err != nil {
			return fmt.Errorf("failed to remove NetworkManager dnsmasq conf: %w", err)
		}
		reloadedNM = true
	}

	if fileExists(networkManagerDnsmasqConf) {
		if err := removeFileWithSudo(networkManagerDnsmasqConf); err != nil {
			return fmt.Errorf("failed to remove NetworkManager dnsmasq rules: %w", err)
		}
		reloadedNM = true
	}

	if fileExists(systemdResolvedDropInPath) {
		if err := removeFileWithSudo(systemdResolvedDropInPath); err != nil {
			return fmt.Errorf("failed to remove systemd-resolved drop-in: %w", err)
		}
		restartedStub = true
	}

	if reloadedNM {
		// SECURITY: Fixed - Using hardcoded command arguments
		args := []string{"systemctl", "reload", "NetworkManager"}
		_ = exec.Command("sudo", args...).Run()
	}
	if restartedStub {
		// SECURITY: Fixed - Using hardcoded command arguments
		args := []string{"systemctl", "restart", "systemd-resolved"}
		_ = exec.Command("sudo", args...).Run()
	}

	return nil
}

// IsNetworkManagerAvailable checks whether the NetworkManager service is present/active.
func IsNetworkManagerAvailable() bool {
	if err := exec.Command("systemctl", "is-active", "--quiet", "NetworkManager").Run(); err == nil {
		return true
	}
	if err := exec.Command("systemctl", "is-enabled", "--quiet", "NetworkManager").Run(); err == nil {
		return true
	}
	return false
}

// writeFileWithSudo writes content to a root-owned path via sudo tee.
func writeFileWithSudo(path, content string) error {
	// Security: Validate path to prevent path traversal attacks
	if err := validateSystemPath(path); err != nil {
		return fmt.Errorf("security validation failed for path %s: %w", path, err)
	}

	// SECURITY: Fixed - Using validated path with hardcoded command
	args := []string{"tee", path}
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// fileExists is a helper to check file existence.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// removeFileWithSudo removes a file via sudo rm -f.
func removeFileWithSudo(path string) error {
	// Security: Validate path to prevent path traversal attacks
	if err := validateSystemPath(path); err != nil {
		return fmt.Errorf("security validation failed for path %s: %w", path, err)
	}

	// SECURITY: Fixed - Using validated path with hardcoded command
	args := []string{"rm", "-f", path}
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

/**
 * SetupLocalDNSResolution configures dnsmasq for local .test domains
 * Works with both NetworkManager's dnsmasq and standalone dnsmasq
 *
 * @return error if configuration fails
 */
func SetupLocalDNSResolution() error {
	// First check if dnsmasq is available (NetworkManager or standalone)
	if !IsDnsmasqAvailable() {
		return fmt.Errorf("dnsmasq is not available on this system")
	}

	// Avoid conflicts with systemd-resolved stub listener by configuring NetworkManager when possible
	if !IsNetworkManagerDnsmasqRunning() && IsSystemdResolvedStubActive() {
		if IsNetworkManagerAvailable() {
			if err := enableNetworkManagerDnsmasqIntegration(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("systemd-resolved stub listener is active on 127.0.0.53:53. Disable DNSStubListener or configure NetworkManager dnsmasq before continuing")
		}
	}

	// Check if configuration already exists in either location
	configPaths := []string{
		"/etc/dnsmasq.d/chauffeur.conf",
		"/etc/NetworkManager/dnsmasq.d/chauffeur.conf",
	}

	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			// Security: Validate config path
			if err := validateSystemPath(configPath); err != nil {
				return fmt.Errorf("security validation failed for config path %s: %w", configPath, err)
			}

			// Configuration exists, restart the appropriate service
			// SECURITY: Fixed - Using hardcoded command arguments
			if IsNetworkManagerDnsmasqRunning() {
				args := []string{"systemctl", "reload", "NetworkManager"}
				return exec.Command("sudo", args...).Run()
			} else {
				args := []string{"systemctl", "restart", "dnsmasq"}
				return exec.Command("sudo", args...).Run()
			}
		}
	}

	// Install configuration in the appropriate location
	var configPath string
	var restartCmd *exec.Cmd

	if IsNetworkManagerDnsmasqRunning() {
		configPath = "/etc/NetworkManager/dnsmasq.d/chauffeur.conf"
		// SECURITY: Fixed - Using hardcoded command arguments
		args := []string{"systemctl", "reload", "NetworkManager"}
		restartCmd = exec.Command("sudo", args...)
	} else {
		// Create dnsmasq.d directory if it doesn't exist for standalone dnsmasq
		// Security: Validate directory path
		dnsmasqDir := "/etc/dnsmasq.d"
		if err := validateSystemPath(dnsmasqDir); err != nil {
			return fmt.Errorf("security validation failed for %s: %w", dnsmasqDir, err)
		}
		// SECURITY: Fixed - Using validated path with hardcoded command arguments
		args := []string{"install", "-d", "-m", "755", dnsmasqDir}
		if err := exec.Command("sudo", args...).Run(); err != nil {
			return fmt.Errorf("failed to create dnsmasq.d directory: %w", err)
		}
		configPath = "/etc/dnsmasq.d/chauffeur.conf"
		// SECURITY: Fixed - Using hardcoded command arguments
		args = []string{"systemctl", "restart", "dnsmasq"}
		restartCmd = exec.Command("sudo", args...)
	}

	// Create the configuration file
	// Security: Validate config path before writing
	if err := validateSystemPath(configPath); err != nil {
		return fmt.Errorf("security validation failed for config path %s: %w", configPath, err)
	}

	// SECURITY: Fixed - Using validated path with hardcoded command
	args := []string{"tee", configPath}
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = strings.NewReader(defaultDnsmasqConfigBlock)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write dnsmasq configuration: %w", err)
	}

	// Restart the appropriate service
	// SENSITIVE: Process execution - restarting system service with elevated privileges
	if err := restartCmd.Run(); err != nil {
		// For standalone dnsmasq, don't fail hard on restart errors
		// as it might conflict with NetworkManager's dnsmasq
		if !IsNetworkManagerDnsmasqRunning() {
			return fmt.Errorf("failed to restart dnsmasq service: %w", err)
		}
		// If NetworkManager is handling dnsmasq, the reload is sufficient
	}

	return nil
}

// VerifyLocalDNSResolution checks whether the system resolver maps *.test domains to localhost.
// It returns true when the synthetic probe domain resolves to 127.0.0.1.
func VerifyLocalDNSResolution() (bool, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dnsProbeTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupHost(ctx, DNSProbeDomain)
	if err != nil {
		return false, nil, err
	}

	for _, ip := range ips {
		if ip == "127.0.0.1" {
			return true, ips, nil
		}
	}

	return false, ips, nil
}

// IsSystemdResolvedStubActive returns true when systemd-resolved's stub listener is active on 127.0.0.53:53.
func IsSystemdResolvedStubActive() bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", "systemd-resolved")
	if err := cmd.Run(); err != nil {
		return false
	}

	conn, err := net.DialTimeout("udp", "127.0.0.53:53", 250*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}

	return false
}
