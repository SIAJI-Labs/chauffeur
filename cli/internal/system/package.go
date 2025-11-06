package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PackageManager represents a system package manager
type PackageManager string

const (
	Pacman   PackageManager = "pacman"
	Apt      PackageManager = "apt"
	Yum      PackageManager = "yum"
	Dnf      PackageManager = "dnf"
	Zypper   PackageManager = "zypper"
	Unknown  PackageManager = "unknown"
)

// Package represents a system package information
type Package struct {
	Name        string
	Description string
	PackageName string // actual package name for the package manager
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
		Pacman:  {"pacman", "--version"},
		Apt:     {"apt", "--version"},
		Yum:     {"yum", "--version"},
		Dnf:     {"dnf", "--version"},
		Zypper:  {"zypper", "--version"},
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
	pm := DetectPackageManager()
	
	switch pm {
	case Pacman:
		cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg.PackageName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
		
	case Apt:
		cmd := exec.Command("sudo", "apt", "install", "-y", pkg.PackageName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
		
	case Yum:
		cmd := exec.Command("sudo", "yum", "install", "-y", pkg.PackageName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
		
	case Dnf:
		cmd := exec.Command("sudo", "dnf", "install", "-y", pkg.PackageName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
		
	case Zypper:
		cmd := exec.Command("sudo", "zypper", "install", "-y", pkg.PackageName)
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
	
	fmt.Printf("The following packages are required for Caddy local domain resolution:\n")
	for _, pkg := range missing {
		fmt.Printf("  - %s (%s): %s\n", pkg.Name, pkg.PackageName, pkg.Description)
	}
	
	fmt.Printf("\nWould you like to install these packages? [y/N]: ")
	
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
	
	// Check if configuration already exists in either location
	configPaths := []string{
		"/etc/dnsmasq.d/chauffeur.conf",
		"/etc/NetworkManager/dnsmasq.d/chauffeur.conf",
	}
	
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			// Configuration exists, restart the appropriate service
			if IsNetworkManagerDnsmasqRunning() {
				return exec.Command("sudo", "systemctl", "reload", "NetworkManager").Run()
			} else {
				return exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run()
			}
		}
	}
	
	// Install configuration in the appropriate location
	var configPath string
	var restartCmd *exec.Cmd
	
	if IsNetworkManagerDnsmasqRunning() {
		configPath = "/etc/NetworkManager/dnsmasq.d/chauffeur.conf"
		restartCmd = exec.Command("sudo", "systemctl", "reload", "NetworkManager")
	} else {
		// Create dnsmasq.d directory if it doesn't exist for standalone dnsmasq
		if err := exec.Command("sudo", "install", "-d", "-m", "755", "/etc/dnsmasq.d").Run(); err != nil {
			return fmt.Errorf("failed to create dnsmasq.d directory: %w", err)
		}
		configPath = "/etc/dnsmasq.d/chauffeur.conf"
		restartCmd = exec.Command("sudo", "systemctl", "restart", "dnsmasq")
	}
	
	// Create the configuration file
	configContent := `# Chauffeur local development resolver
# Redirect all *.test domains to localhost
address=/.test/127.0.0.1
# Only listen locally
listen-address=127.0.0.1
bind-interfaces
`
	
	// Write configuration using tee
	cmd := exec.Command("sudo", "tee", configPath)
	cmd.Stdin = strings.NewReader(configContent)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write dnsmasq configuration: %w", err)
	}
	
	// Restart the appropriate service
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
