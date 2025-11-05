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
