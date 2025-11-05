package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/logging"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

// RunRemove handles `chauf remove <service>` logic for removing installed services.
func RunRemove(args []string) error {
	if len(args) == 0 {
		printRemoveUsage()
		return errors.New("no services specified")
	}

	var (
		force    bool
		services []string
		versions map[string]string
	)

	versions = make(map[string]string)

	// Parse arguments to handle service[version] syntax (similar to install)
	i := 0
	for i < len(args) {
		arg := args[i]
		
		switch arg {
		case "--force":
			force = true
			i++
		case "--help", "-h":
			printRemoveUsage()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for remove: %s", arg)
			}
			
			// Handle service[version] syntax (primarily for PHP)
			if arg == "php" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Next argument looks like a version (not a flag)
				nextArg := args[i+1]
				if isValidPHPVersion(nextArg) {
					versions[arg] = nextArg
					services = append(services, arg)
					i += 2
					continue
				}
			}
			
			services = append(services, arg)
			i++
		}
	}

	if len(services) == 0 {
		return errors.New("no services specified")
	}

	// Initialize logger
	logger := logging.NewCommandLogger("remove")

	// Validate each service
	for _, name := range services {
		if !IsKnownService(name) {
			return fmt.Errorf("unknown service: %s", name)
		}
	}

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	// Check workspace exists
	if _, err := os.Stat(prefix); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("chauffeur workspace not found at %s", prefix)
		}
		return fmt.Errorf("workspace check failed: %w", err)
	}

	logger.Info("Preparing to remove services")

	// Process each service
	for _, name := range services {
		if name == "php" {
			// Handle PHP removal with version support
			version := versions["php"]
			if err := runRemovePHP(version, force, logger); err != nil {
				return err
			}
		} else {
			// Handle other services (nginx, caddy)
			info, err := system.Detect()
			if err != nil {
				return err
			}
			spec, err := newServiceSpec(name, prefix, info)
			if err != nil {
				return err
			}

			ok, err := spec.available()
			if err != nil {
				return fmt.Errorf("check %s availability: %w", spec.name, err)
			}
			if !ok {
				logger.Warn(fmt.Sprintf("%s is not installed", spec.name), "use 'chauf install' to install it first")
				continue
			}

			if err := runRemoveService(spec, force, logger); err != nil {
				return err
			}
		}
	}

	logger.Info("Service removal complete")
	return nil
}

// runRemovePHP handles PHP version-specific removal
func runRemovePHP(version string, force bool, logger *logging.CommandLogger) error {
	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	if version == "" {
		// If no version specified, remove all PHP installations
		phpDir := filepath.Join(prefix, "php")
		if _, err := os.Stat(phpDir); os.IsNotExist(err) {
			logger.Warn("No PHP installations found", "use 'chauf install php' to install PHP first")
			return nil
		}

		if !force {
			logger.Warn("Removing all PHP versions", "this will remove ~/.chauffeur/php/ directory")
			fmt.Printf("This will remove all installed PHP versions. Continue? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				logger.Success("Operation cancelled", "")
				return nil
			}
		}

		if err := os.RemoveAll(phpDir); err != nil {
			return logger.Fail("Remove PHP directory", err.Error())
		}

		logger.Success("Removed all PHP installations", phpDir)
		removePHPSHIMs(prefix)
		return nil
	}

	// Remove specific PHP version
	phpVersionDir := filepath.Join(prefix, "php", version)
	if _, err := os.Stat(phpVersionDir); os.IsNotExist(err) {
		return logger.Fail("Remove PHP version", fmt.Sprintf("PHP %s is not installed", version))
	}

	if !force {
		fmt.Printf("Remove PHP %s? This will delete %s. Continue? (y/N): ", version, phpVersionDir)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			logger.Success("Operation cancelled", "")
			return nil
		}
	}

	if err := os.RemoveAll(phpVersionDir); err != nil {
		return logger.Fail("Remove PHP version", err.Error())
	}

	logger.Success("Removed PHP version", fmt.Sprintf("PHP %s", version))

	// Remove the specific PHP shim
	shimPath := filepath.Join(prefix, "bin", "shims", "php-"+version)
	if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove shim", shimPath)
	}

	// Check if this was the default PHP version and update if needed
	updateDefaultPHPAfterRemoval(prefix, version)

	return nil
}

// runRemoveService handles removal of nginx and caddy
func runRemoveService(spec serviceSpec, force bool, logger *logging.CommandLogger) error {
	ok, err := spec.available()
	if err != nil {
		return err
	}
	if !ok {
		logger.Warn(fmt.Sprintf("%s is not installed", spec.name), "")
		return nil
	}

	// Special handling for caddy with dnsmasq validation
	if spec.name == "caddy" {
		if err := handleCaddyRemoval(spec, force, logger); err != nil {
			return err
		}
		return nil
	}

	if !force {
		fmt.Printf("Remove %s? This will delete %s. Continue? (y/N): ", spec.name, spec.binaryPath)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			logger.Success("Operation cancelled", "")
			return nil
		}
	}

	// For nginx and other services, remove their entire directories
	serviceDir := filepath.Dir(filepath.Dir(spec.binaryPath)) // Go up two levels from bin/caddy to caddy/
	if err := os.RemoveAll(serviceDir); err != nil {
		return logger.Fail(fmt.Sprintf("Remove %s", spec.name), err.Error())
	}

	logger.Success("Removed service", fmt.Sprintf("%s (%s)", spec.name, serviceDir))

	// Remove the service shim
	shimPath := filepath.Join(filepath.Dir(filepath.Dir(spec.binaryPath)), "..", "bin", "shims", spec.name)
	if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove shim", shimPath)
	}

	return nil
}

// handleCaddyRemoval handles caddy removal with dnsmasq validation
func handleCaddyRemoval(spec serviceSpec, force bool, logger *logging.CommandLogger) error {
	if !force {
		fmt.Printf("Remove %s? This will delete %s. Continue? (y/N): ", spec.name, spec.binaryPath)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			logger.Success("Operation cancelled", "")
			return nil
		}
	}

	// Check for dnsmasq and assess removal risk
	hasDnsmasq := system.IsCommandAvailable("dnsmasq")
	removeDNS := false
	
	if hasDnsmasq {
		logger.Info("Checking dnsmasq usage...")
		dnsLogger := logger.NewChildLogger("dns")
		
		dnsLogger.Warn("dnsmasq is installed on the system", "other tools may depend on it")
		
		fmt.Printf("⚠  WARNING: dnsmasq is installed and may be required by other applications.\n")
		fmt.Printf("   Removing Caddy will also offer to remove dnsmasq from the system.\n")
		fmt.Printf("   This could break other tools that rely on local DNS resolution.\n")
		fmt.Printf("\n")
		
		// Second confirmation for dnsmasq removal
		fmt.Printf("Do you also want to remove dnsmasq from the system? (NOT RECOMMENDED) [y/N]: ")
		var removeDNSResponse string
		fmt.Scanln(&removeDNSResponse)
		removeDNS = strings.ToLower(removeDNSResponse) == "y" || strings.ToLower(removeDNSResponse) == "yes"
		
		if removeDNS {
			// Double confirmation for dnsmasq removal
			fmt.Printf("⚠  FINAL WARNING: You are about to remove dnsmasq completely.\n")
			fmt.Printf("   This will affect local DNS resolution for all applications.\n")
			fmt.Printf("   Other tools that depend on dnsmasq may stop working.\n")
			fmt.Printf("\n")
			fmt.Printf("Are you absolutely sure you want to remove dnsmasq? Type 'REMOVE' to confirm: ")
			var finalConfirm string
			fmt.Scanln(&finalConfirm)
			
			if finalConfirm != "REMOVE" {
				dnsLogger.Success("dnsmasq removal cancelled", "keeping system package intact")
				removeDNS = false
			} else {
				dnsLogger.Info("Double confirmation received - proceeding with dnsmasq removal")
			}
		}
		
		if removeDNS {
			// Remove dnsmasq using the system package manager
			pm := system.DetectPackageManager()
			if pm == system.Unknown {
				return dnsLogger.Fail("remove dnsmasq", "unsupported package manager")
			}
			
			dnsLogger.Info(fmt.Sprintf("Removing dnsmasq using %s...", pm))
			
			// Determine the package name based on package manager
			var dnsmasqPackage string
			switch pm {
			case system.Pacman:
				dnsmasqPackage = "dnsmasq"
			case system.Apt:
				dnsmasqPackage = "dnsmasq"
			case system.Yum, system.Dnf:
				dnsmasqPackage = "dnsmasq"
			case system.Zypper:
				dnsmasqPackage = "dnsmasq"
			default:
				dnsmasqPackage = "dnsmasq"
			}
			
			// Check for AUR helpers for Arch Linux
			useSudo := true
			if pm == system.Pacman {
				archPm := system.DetectArchPackageManager()
				if archPm == "yay" || archPm == "paru" {
					useSudo = false // AUR helpers handle sudo internally
				}
			}
			
			// Create the removal command
			var removeCmd *exec.Cmd
			switch pm {
			case system.Pacman:
				if useSudo {
					removeCmd = exec.Command("sudo", "pacman", "-R", "--noconfirm", dnsmasqPackage)
				} else {
					// Use AUR helper (yay/paru) without sudo
					archPm := system.DetectArchPackageManager()
					removeCmd = exec.Command(archPm, "-R", "--noconfirm", dnsmasqPackage)
				}
			case system.Apt:
				removeCmd = exec.Command("sudo", "apt", "remove", "-y", dnsmasqPackage)
			case system.Yum:
				removeCmd = exec.Command("sudo", "yum", "remove", "-y", dnsmasqPackage)
			case system.Dnf:
				removeCmd = exec.Command("sudo", "dnf", "remove", "-y", dnsmasqPackage)
			case system.Zypper:
				removeCmd = exec.Command("sudo", "zypper", "remove", "-y", dnsmasqPackage)
			default:
				return dnsLogger.Fail("remove dnsmasq", fmt.Sprintf("unsupported package manager: %s", pm))
			}
			
			// Execute the removal command
			if err := removeCmd.Run(); err != nil {
				return dnsLogger.Fail("remove dnsmasq", err.Error())
			}
			
			dnsLogger.Success("Removed dnsmasq package", "system DNS resolution may be affected")
		} else {
			dnsLogger.Info("Keeping dnsmasq installed - local DNS resolution will remain functional")
		}
	}

	// Remove caddy directory
	serviceDir := filepath.Dir(filepath.Dir(spec.binaryPath)) // Go up two levels from bin/caddy to caddy/
	if err := os.RemoveAll(serviceDir); err != nil {
		return logger.Fail("Remove Caddy directory", err.Error())
	}

	logger.Success("Removed Caddy service", fmt.Sprintf("%s (%s)", spec.name, serviceDir))

	// Remove the caddy shim
	shimPath := filepath.Join(filepath.Dir(filepath.Dir(spec.binaryPath)), "..", "bin", "shims", spec.name)
	if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove shim", shimPath)
	}

	if hasDnsmasq && !removeDNS {
		logger.Info("Reminder: dnsmasq remains installed for other applications")
	}

	return nil
}

// removePHPSHIMs removes all PHP version shims
func removePHPSHIMs(prefix string) {
	shimsDir := filepath.Join(prefix, "bin", "shims")
	if entries, err := os.ReadDir(shimsDir); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "php-") {
				shimPath := filepath.Join(shimsDir, entry.Name())
				os.Remove(shimPath)
			}
		}
	}
}

// updateDefaultPHPAfterRemoval updates the default PHP if the removed version was the default
func updateDefaultPHPAfterRemoval(prefix, removedVersion string) {
	configPath := filepath.Join(prefix, "config", "chauffeur.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return
	}

	// Check if removed version was the default
	// For now, we'll just remove the default setting if it matches the removed version
	// In a more sophisticated implementation, we'd parse the YAML properly
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	content := string(data)
	if strings.Contains(content, "default: "+removedVersion) {
		// Try to find an alternative PHP version to set as default
		phpDir := filepath.Join(prefix, "php")
		if entries, err := os.ReadDir(phpDir); err == nil && len(entries) > 0 {
			// Set first available version as default
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != removedVersion {
					newDefault := entry.Name()
					// Simple string replacement - in production you'd use a proper YAML parser
					newContent := strings.Replace(content, "default: "+removedVersion, "default: "+newDefault, 1)
					os.WriteFile(configPath, []byte(newContent), 0644)
					fmt.Printf("Switched default PHP to %s\n", newDefault)
					break
				}
			}
		}
	}
}

func printRemoveUsage() {
	fmt.Println(`Usage: chauf remove [--force] <service> [<version>...]

Removes installed Chauffeur-managed services.

Options:
  --force    Remove without confirmation prompts.
  -h, --help Show this message.

Services:
  caddy      Remove installed Caddy web server.
              NOTE: Will check for dnsmasq and offer optional removal with
              double confirmation to prevent breaking other applications.
  nginx      Remove installed Nginx web server.
  php        Remove installed PHP runtime(s).
  
Caddy Removal:
  chauf remove caddy        Remove Caddy with interactive prompts.
                            If dnsmasq is installed, offers optional removal
                            with double confirmation to prevent system damage.
                            Streamlined flow goes directly to dnsmasq prompt after initial caddy confirmation.
  chauf remove caddy --force Remove Caddy without prompts (does not remove dnsmasq).
  
PHP Removal:
  chauf remove php           Remove all installed PHP versions (with confirmation).
  chauf remove php 8.3      Remove specific PHP version 8.3 (with confirmation).
  chauf remove php --force   Remove without confirmation prompts.`)
}
