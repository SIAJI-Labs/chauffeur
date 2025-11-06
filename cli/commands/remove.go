package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
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
	logger := lib.NewCommandLogger("remove")

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
func runRemovePHP(version string, force bool, logger *lib.Logger) error {
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
			// SENSITIVE: Destructive operation - user confirmation for removing all PHP installations
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
		// SENSITIVE: Destructive operation - user confirmation for removing specific PHP version
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
func runRemoveService(spec serviceSpec, force bool, logger *lib.Logger) error {
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
		// SENSITIVE: Destructive operation - user confirmation for removing service (nginx/caddy)
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
func handleCaddyRemoval(spec serviceSpec, force bool, logger *lib.Logger) error {
	if !force {
		// SENSITIVE: Destructive operation - user confirmation for removing caddy service
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
		// SENSITIVE: Destructive operation - system package removal confirmation for dnsmasq
		fmt.Printf("Do you also want to remove dnsmasq from the system? (NOT RECOMMENDED) [y/N]: ")
		var removeDNSResponse string
		fmt.Scanln(&removeDNSResponse)
		removeDNS = strings.ToLower(removeDNSResponse) == "y" || strings.ToLower(removeDNSResponse) == "yes"
		
		if removeDNS {
			// Double confirmation for dnsmasq removal
			// SENSITIVE: Destructive operation - double confirmation for critical system package removal
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
			// Before removing dnsmasq, ask about chauffeur config
			if err := removeDnsmasqConfigurationBeforeRemoval(logger); err != nil {
				return err
			}
			
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
			// SENSITIVE: Process execution - running system package manager with elevated privileges
			if err := removeCmd.Run(); err != nil {
				return dnsLogger.Fail("remove dnsmasq", err.Error())
			}
			
			dnsLogger.Success("Removed dnsmasq package", "system DNS resolution may be affected")
		} else {
			dnsLogger.Info("Keeping dnsmasq installed - local DNS resolution will remain functional")
			// Still offer to remove configuration even if keeping dnsmasq
			if err := removeDnsmasqConfiguration(logger); err != nil {
				return err
			}
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

// removeDnsmasqConfigurationBeforeRemoval removes the chauffeur dnsmasq configuration before package removal.
func removeDnsmasqConfigurationBeforeRemoval(logger *lib.Logger) error {
	logger.Info("Checking for Chauffeur dnsmasq configuration...")
	
	dnsLogger := logger.NewChildLogger("dns")
	
	// Check if configuration exists
	if _, err := os.Stat("/etc/dnsmasq.d/chauffeur.conf"); os.IsNotExist(err) {
		dnsLogger.Info("No Chauffeur dnsmasq configuration found - no cleanup needed")
		return nil
	}
	
	dnsLogger.Info("Found Chauffeur dnsmasq configuration at /etc/dnsmasq.d/chauffeur.conf")
	
	fmt.Printf("\n%s", "Do you want to remove the Chauffeur dnsmasq configuration before removing dnsmasq? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		dnsLogger.Info("Keeping dnsmasq configuration - file will be removed with package")
		return nil
	}
	
	dnsLogger.Info("Removing dnsmasq configuration...")
	
	// SENSITIVE: System file modification - removing system configuration file with elevated privileges
	if err := exec.Command("sudo", "rm", "-f", "/etc/dnsmasq.d/chauffeur.conf").Run(); err != nil {
		return dnsLogger.Fail("remove dnsmasq configuration", err.Error())
	}
	
	dnsLogger.Success("dnsmasq configuration removed", "Configuration file deleted")
	
	// Offer to restart dnsmasq to apply changes before package removal
	fmt.Printf("\n%s", "Do you want to restart dnsmasq to apply configuration changes before removing it? [y/N]: ")
	var restartResponse string
	fmt.Scanln(&restartResponse)
	
	if strings.ToLower(restartResponse) == "y" || strings.ToLower(restartResponse) == "yes" {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Info("NetworkManager is managing dnsmasq - reloading NetworkManager to apply changes...")
			// SENSITIVE: Process execution - restarting system service with elevated privileges
			if err := exec.Command("sudo", "systemctl", "reload", "NetworkManager").Run(); err != nil {
				return dnsLogger.Fail("reload NetworkManager service", err.Error())
			}
			dnsLogger.Success("NetworkManager reloaded", "Configuration changes applied via NetworkManager")
		} else {
			dnsLogger.Info("Restarting dnsmasq service to apply configuration changes...")
			// SENSITIVE: Process execution - restarting system service with elevated privileges
			if err := exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run(); err != nil {
				return dnsLogger.Fail("restart dnsmasq service", err.Error())
			}
			dnsLogger.Success("dnsmasq restarted", "Configuration changes applied")
		}
	} else {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Info("NetworkManager not reloaded - configuration changes will apply when NetworkManager restarts")
		} else {
			dnsLogger.Info("dnsmasq not restarted - continuing with package removal")
		}
	}
	
	return nil
}

// removeDnsmasqConfiguration removes the chauffeur dnsmasq configuration.
func removeDnsmasqConfiguration(logger *lib.Logger) error {
	logger.Info("Checking for Chauffeur dnsmasq configuration...")
	
	dnsLogger := logger.NewChildLogger("dns")
	
	// Check if configuration exists
	if _, err := os.Stat("/etc/dnsmasq.d/chauffeur.conf"); os.IsNotExist(err) {
		dnsLogger.Info("No Chauffeur dnsmasq configuration found - no cleanup needed")
		return nil
	}
	
	dnsLogger.Info("Found Chauffeur dnsmasq configuration at /etc/dnsmasq.d/chauffeur.conf")
	
	fmt.Printf("\n%s", "Do you want to remove the Chauffeur dnsmasq configuration? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		dnsLogger.Info("Keeping dnsmasq configuration - file remains")
		return nil
	}
	
	dnsLogger.Info("Removing dnsmasq configuration...")
	
	// SENSITIVE: System file modification - removing system configuration file with elevated privileges
	if err := exec.Command("sudo", "rm", "-f", "/etc/dnsmasq.d/chauffeur.conf").Run(); err != nil {
		return dnsLogger.Fail("remove dnsmasq configuration", err.Error())
	}
	
	dnsLogger.Success("dnsmasq configuration removed", "Configuration file deleted")
	
	// Optionally restart dnsmasq (only if dnsmasq is still installed)
	fmt.Printf("\n%s", "Do you want to restart dnsmasq to apply changes? [y/N]: ")
	var restartResponse string
	fmt.Scanln(&restartResponse)
	
	if strings.ToLower(restartResponse) == "y" || strings.ToLower(restartResponse) == "yes" {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Info("NetworkManager is managing dnsmasq - reloading NetworkManager to apply changes...")
			if err := exec.Command("sudo", "systemctl", "reload", "NetworkManager").Run(); err != nil {
				dnsLogger.Fail("reload NetworkManager service", err.Error())
			} else {
				dnsLogger.Success("NetworkManager reloaded", "Changes applied via NetworkManager")
			}
		} else {
			// Check if dnsmasq is still installed before trying to restart
			if system.IsCommandAvailable("dnsmasq") {
				dnsLogger.Info("Restarting dnsmasq...")
				if err := exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run(); err != nil {
					return dnsLogger.Fail("restart dnsmasq service", err.Error())
				}
				dnsLogger.Success("dnsmasq restarted", "Changes applied")
			} else {
				dnsLogger.Info("dnsmasq package not found - skipping restart (Changes will apply when dnsmasq is reinstalled)")
			}
		}
	} else {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Info("NetworkManager not reloaded - changes will apply when NetworkManager restarts")
		} else {
			dnsLogger.Info("dnsmasq not restarted - changes will apply on next service restart")
		}
	}
	
	return nil
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
