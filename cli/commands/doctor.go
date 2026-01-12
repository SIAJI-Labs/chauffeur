package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/services"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// DependencyCheck represents a single dependency check result
type DependencyCheck struct {
	Name        string
	Description string
	Status      string
	Installed   bool
	Version     string
	Required    string
	CanFix      bool
	FixCommand  string
	Errors      []string
	Warnings    []string
}

// DoctorOptions contains configuration for the doctor command
type DoctorOptions struct {
	CheckAll     bool
	CheckDeps    bool
	CheckPHP     bool
	CheckSSL     bool
	CheckNetwork bool
	CheckDNS     bool
	Verbose      bool
	Fix          bool
	AutoFix      bool
	Quiet        bool
}

// FixPlan represents a fix that needs to be executed
type FixPlan struct {
	Name        string
	Description string
	Command     string
	Category    string
	Priority    int // 1=error, 2=warning
}

// RunDoctor performs health checks and dependency validation
func RunDoctor(args []string) error {
	options := parseDoctorArgs(args)

	logger := lib.NewCommandLogger("doctor")

	// Set default behavior
	if options.CheckAll || (!options.CheckDeps && !options.CheckPHP && !options.CheckSSL && !options.CheckNetwork && !options.CheckDNS) {
		options.CheckDeps = true
		options.CheckPHP = true
		options.CheckSSL = true
		options.CheckNetwork = true
		options.CheckDNS = true
	}

	// Check if workspace exists and offer to initialize
	if err := lib.ValidateWorkspace([]string{"--help"}); err != nil && !options.Quiet {
		logger.Info("No Chauffeur workspace found")
		logger.Info("Run 'chauf init' to create a workspace")
		if !options.Fix && !options.AutoFix {
			return nil
		}
	}

	logger.PrintSection("🩺 Chauffeur Doctor")
	logger.Info("Performing health checks...")

	var allChecks []DependencyCheck
	var hasErrors bool
	var hasWarnings bool
	var fixPlans []FixPlan

	// Run dependency checks
	if options.CheckDeps {
		logger.PrintSection("System Dependencies")
		checks, errs, warns := checkSystemDependencies(options, &fixPlans)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
		}
	}

	// Run PHP build dependency checks
	if options.CheckPHP {
		logger.PrintSection("PHP Build Dependencies")
		checks, errs, warns := checkPHPBuildDependencies(options, &fixPlans)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
		}
	}

	// Run SSL dependency checks
	if options.CheckSSL {
		logger.PrintSection("SSL Certificate Dependencies")
		checks, errs, warns := checkSSLDependencies(options, &fixPlans)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
		}
	}

	// Run network dependency checks
	if options.CheckNetwork {
		logger.PrintSection("Network & Port Dependencies")
		checks, errs, warns := checkNetworkDependencies(options, &fixPlans)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
		}
	}

	// Run DNS dependency checks
	if options.CheckDNS {
		logger.PrintSection("DNS Resolution Dependencies")
		checks, errs, warns := checkDNSDependencies(options, &fixPlans)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
		}
	}

	// Handle fix plan display/execution based on flags
	if len(fixPlans) > 0 {
		if options.AutoFix {
			// --auto-fix: Show plan, prompt user, and execute if confirmed
			if executeFixPlan(logger, fixPlans, options.Quiet) {
				logger.Info("Auto-fix completed. Run 'chauf doctor' again to verify all issues were resolved.")
			}
		} else if options.Fix {
			// --fix: Show fix plan only (no execution)
			displayFixPlan(logger, fixPlans)
		}
	}

	// Print summary
	printDoctorSummary(logger, allChecks, hasErrors, hasWarnings)

	// Return appropriate error code
	if hasErrors {
		return fmt.Errorf("doctor found %d error(s) that need to be resolved", countErrors(allChecks))
	}

	if hasWarnings {
		if !options.Quiet {
			logger.Warn("Doctor completed with warnings", "System is functional but some optimizations are recommended")
		}
		return nil
	}

	if !options.Quiet {
		logger.Success("Doctor completed", "All checks passed - system is healthy!")
	}
	return nil
}

// parseDoctorArgs parses command line arguments for the doctor command
func parseDoctorArgs(args []string) DoctorOptions {
	options := DoctorOptions{}
	logger := lib.NewCommandLogger("doctor")

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printDoctorUsage()
			os.Exit(0)
		case "--check-all", "-a":
			options.CheckAll = true
			i++
		case "--check-deps", "-d":
			options.CheckDeps = true
			i++
		case "--check-php", "-p":
			options.CheckPHP = true
			i++
		case "--check-ssl", "-s":
			options.CheckSSL = true
			i++
		case "--check-network", "-n":
			options.CheckNetwork = true
			i++
		case "--check-dns":
			options.CheckDNS = true
			i++
		case "--verbose", "-v":
			options.Verbose = true
			i++
		case "--fix", "-f":
			options.Fix = true
			i++
		case "--auto-fix":
			options.AutoFix = true
			options.Fix = true
			i++
		case "--quiet", "-q":
			options.Quiet = true
			i++
		case "--internal-fix-openssl":
			// Handle internal OpenSSL configuration generation
			if i+2 >= len(args) {
				logger.Error("Internal fix-openssl requires --php-version and --workspace-dir flags", "")
				os.Exit(1)
			}
			phpVersion := ""
			workspaceDir := ""
			// Parse the additional flags for internal use
			for j := i + 1; j < len(args) && j < i+3; j++ {
				if strings.HasPrefix(args[j], "--php-version=") {
					phpVersion = strings.TrimPrefix(args[j], "--php-version=")
				} else if strings.HasPrefix(args[j], "--workspace-dir=") {
					workspaceDir = strings.TrimPrefix(args[j], "--workspace-dir=")
				}
			}
			if phpVersion == "" || workspaceDir == "" {
				logger.Error("Internal fix-openssl requires --php-version and --workspace-dir flags", "")
				os.Exit(1)
			}
			// Call the OpenSSL configuration function
			if err := installers.WriteOpenSSLConf(workspaceDir, phpVersion); err != nil {
				logger.Error("Failed to generate OpenSSL configuration", err.Error())
				os.Exit(1)
			}
			logger.Success(fmt.Sprintf("OpenSSL configuration generated for PHP %s", phpVersion), "")
			os.Exit(0)
		default:
			logger.Error(fmt.Sprintf("Unknown flag: %s", arg), "")
			printDoctorUsage()
			os.Exit(1)
		}
	}

	return options
}

// printDoctorUsage displays usage information for the doctor command
func printDoctorUsage() {
	logger := lib.NewCommandLogger("doctor")
	logger.PrintBlock(`Chauffeur Doctor - Health Check and Troubleshooting

Usage:
  chauf doctor [options]

Options:
  --check-all, -a       Run all dependency checks (default behavior)
  --check-deps, -d      Check system dependencies (git, curl, tar, etc.)
  --check-php, -p       Check PHP build dependencies and headers
  --check-ssl, -s       Check SSL certificate dependencies
  --check-network, -n   Check network and port availability
  --check-dns           Check DNS resolution for .test domains
  --verbose, -v         Show detailed diagnostic information
  --fix, -f             Show fix plan for all issues found (no execution)
  --auto-fix            Show fix plan, prompt for confirmation, then execute
  --quiet, -q           Suppress non-error output
  --help, -h            Show this help message

Examples:
  chauf doctor                    # Run all health checks
  chauf doctor --check-deps       # Check only system dependencies
  chauf doctor --fix              # Show fix plan for all issues
  chauf doctor --auto-fix         # Fix issues with confirmation prompt

Description:
  Performs comprehensive health checks on your Chauffeur installation,
  validates system dependencies, and provides guidance for resolving issues.
`)
}

// checkSystemDependencies validates core system dependencies
func checkSystemDependencies(options DoctorOptions, fixPlans *[]FixPlan) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	dependencies := []struct {
		name        string
		binary      string
		versionFlag string
		versionCmd  string
		required    string
		description string
		fixCommands map[string]string // distro -> command
	}{
		{
			name:        "git",
			binary:      "git",
			versionFlag: "--version",
			versionCmd:  "git --version",
			required:    ">= 2.0",
			description: "Version control system",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y git",
				"arch":          "sudo pacman -S git",
				"fedora":        "sudo dnf install -y git",
			},
		},
		{
			name:        "curl",
			binary:      "curl",
			versionFlag: "--version",
			versionCmd:  "curl --version",
			required:    ">= 7.0",
			description: "HTTP client for downloads",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y curl",
				"arch":          "sudo pacman -S curl",
				"fedora":        "sudo dnf install -y curl",
			},
		},
		{
			name:        "tar",
			binary:      "tar",
			versionFlag: "--version",
			versionCmd:  "tar --version",
			required:    ">= 1.0",
			description: "Archive extraction tool",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y tar",
				"arch":          "sudo pacman -S tar",
				"fedora":        "sudo dnf install -y tar",
			},
		},
		{
			name:        "gcc",
			binary:      "gcc",
			versionFlag: "--version",
			versionCmd:  "gcc --version",
			required:    ">= 4.8",
			description: "C compiler for building PHP",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y build-essential",
				"arch":          "sudo pacman -S base-devel",
				"fedora":        "sudo dnf groupinstall -y 'Development Tools'",
			},
		},
		{
			name:        "make",
			binary:      "make",
			versionFlag: "--version",
			versionCmd:  "make --version",
			required:    ">= 3.8",
			description: "Build automation tool",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y make",
				"arch":          "sudo pacman -S make",
				"fedora":        "sudo dnf install -y make",
			},
		},
		{
			name:        "pkg-config",
			binary:      "pkg-config",
			versionFlag: "--version",
			versionCmd:  "pkg-config --version",
			required:    ">= 0.29",
			description: "Package configuration tool for PHP builds",
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt update && sudo apt install -y pkg-config",
				"arch":          "sudo pacman -S pkgconf",
				"fedora":        "sudo dnf install -y pkgconfig",
			},
		},
	}

	var checks []DependencyCheck
	var errorCount int
	var warningCount int

	distro := detectDistribution()

	for _, dep := range dependencies {
		check := DependencyCheck{
			Name:        dep.name,
			Description: dep.description,
			Required:    dep.required,
			CanFix:      true,
		}

		// Check if binary exists
		if _, err := exec.LookPath(dep.binary); err != nil {
			check.Installed = false
			check.Status = "❌ Not installed"
			check.Errors = append(check.Errors, fmt.Sprintf("%s binary not found in PATH", dep.name))

			if fixCmd, exists := dep.fixCommands[distro]; exists {
				check.FixCommand = fixCmd
			} else {
				check.FixCommand = fmt.Sprintf("Install %s using your package manager", dep.name)
			}

			errorCount++
		} else {
			// Get version information
			if output, err := runDoctorCommandOutput(dep.versionCmd); err == nil {
				check.Installed = true
				check.Version = extractVersion(output, dep.name)
				check.Status = fmt.Sprintf("✅ Installed (%s)", check.Version)

				// Version validation would go here if needed
				if check.Version == "" {
					check.Warnings = append(check.Warnings, "Could not determine version")
					warningCount++
				}
			} else {
				check.Installed = true
				check.Status = "⚠️ Installed (version check failed)"
				check.Warnings = append(check.Warnings, fmt.Sprintf("Version check failed: %v", err))
				warningCount++
			}
		}

		checks = append(checks, check)

		// Print check result and collect fix plans
		printCheckResult(logger, check, options, &errorCount, &warningCount, fixPlans, "System Dependencies")
	}

	return checks, errorCount, warningCount
}

// checkPHPBuildDependencies validates PHP-specific build dependencies
func checkPHPBuildDependencies(options DoctorOptions, fixPlans *[]FixPlan) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	// PHP build dependencies based on AGENTS.md requirements
	dependencies := []struct {
		name        string
		pkgConfig   string
		description string
		required    bool
		fixCommands map[string]string
	}{
		{
			name:        "libzip",
			pkgConfig:   "libzip",
			description: "ZIP archive support",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libzip-dev",
				"arch":          "sudo pacman -S libzip",
				"fedora":        "sudo dnf install -y libzip-devel",
			},
		},
		{
			name:        "libjpeg",
			pkgConfig:   "libjpeg",
			description: "JPEG image processing",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libjpeg-dev",
				"arch":          "sudo pacman -S libjpeg-turbo",
				"fedora":        "sudo dnf install -y libjpeg-turbo-devel",
			},
		},
		{
			name:        "libpng",
			pkgConfig:   "libpng",
			description: "PNG image processing",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libpng-dev",
				"arch":          "sudo pacman -S libpng",
				"fedora":        "sudo dnf install -y libpng-devel",
			},
		},
		{
			name:        "freetype",
			pkgConfig:   "freetype2",
			description: "TrueType font rendering",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libfreetype6-dev",
				"arch":          "sudo pacman -S freetype2",
				"fedora":        "sudo dnf install -y freetype-devel",
			},
		},
		{
			name:        "libxml2",
			pkgConfig:   "libxml-2.0",
			description: "XML processing",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libxml2-dev",
				"arch":          "sudo pacman -S libxml2",
				"fedora":        "sudo dnf install -y libxml2-devel",
			},
		},
		{
			name:        "libcurl",
			pkgConfig:   "libcurl",
			description: "HTTP client support",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libcurl4-openssl-dev",
				"arch":          "sudo pacman -S curl",
				"fedora":        "sudo dnf install -y libcurl-devel",
			},
		},
		{
			name:        "zlib",
			pkgConfig:   "zlib",
			description: "Compression library",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y zlib1g-dev",
				"arch":          "sudo pacman -S zlib",
				"fedora":        "sudo dnf install -y zlib-devel",
			},
		},
		{
			name:        "readline",
			pkgConfig:   "readline",
			description: "Command line editing",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libreadline-dev",
				"arch":          "sudo pacman -S readline",
				"fedora":        "sudo dnf install -y readline-devel",
			},
		},
		{
			name:        "libxslt",
			pkgConfig:   "libxslt",
			description: "XSLT processing",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libxslt1-dev",
				"arch":          "sudo pacman -S libxslt",
				"fedora":        "sudo dnf install -y libxslt-devel",
			},
		},
		{
			name:        "MagickWand",
			pkgConfig:   "MagickWand",
			description: "ImageMagick for imagick extension",
			required:    false, // Optional but recommended
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libmagickwand-dev",
				"arch":          "sudo pacman -S imagemagick",
				"fedora":        "sudo dnf install -y ImageMagick-devel",
			},
		},
		{
			name:        "gmp",
			pkgConfig:   "gmp",
			description: "Arbitrary precision math",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libgmp-dev",
				"arch":          "sudo pacman -S gmp",
				"fedora":        "sudo dnf install -y gmp-devel",
			},
		},
	}

	var checks []DependencyCheck
	var errorCount int
	var warningCount int

	distro := detectDistribution()

	// First check if pkg-config is available
	if _, err := exec.LookPath("pkg-config"); err != nil {
		check := DependencyCheck{
			Name:        "pkg-config",
			Description: "Package configuration tool",
			Status:      "❌ Not available - cannot check PHP dependencies",
			Installed:   false,
			Errors:      []string{"pkg-config is required to check PHP build dependencies"},
			CanFix:      true,
			FixCommand:  getDistroCommand(distro, "pkg-config"),
		}

		checks = append(checks, check)
		errorCount++

		if !options.Quiet {
			logger.Error(check.Status, check.Name)
			if options.Fix {
				logger.Info(fmt.Sprintf("Fix: %s", check.FixCommand))
			}
		}

		return checks, errorCount, warningCount
	}

	for _, dep := range dependencies {
		check := DependencyCheck{
			Name:        dep.name,
			Description: dep.description,
			CanFix:      true,
		}

		// Check using pkg-config
		if output, err := runDoctorCommandOutput(fmt.Sprintf("pkg-config --modversion %s 2>/dev/null", dep.pkgConfig)); err == nil {
			check.Installed = true
			check.Version = strings.TrimSpace(output)
			check.Status = fmt.Sprintf("✅ Available (%s)", check.Version)
		} else {
			if dep.required {
				check.Installed = false
				check.Status = "❌ Missing"
				check.Errors = append(check.Errors, fmt.Sprintf("%s development headers not found", dep.name))
				errorCount++
			} else {
				check.Installed = false
				check.Status = "⚠️ Missing (optional)"
				check.Warnings = append(check.Warnings, fmt.Sprintf("%s is recommended but not required", dep.name))
				warningCount++
			}

			if fixCmd, exists := dep.fixCommands[distro]; exists {
				check.FixCommand = fixCmd
			}
		}

		checks = append(checks, check)

		// Print check result and collect fix plans
		printCheckResult(logger, check, options, &errorCount, &warningCount, fixPlans, "PHP Build Dependencies")
	}

	return checks, errorCount, warningCount
}

// checkSSLDependencies validates SSL certificate dependencies
func checkSSLDependencies(options DoctorOptions, fixPlans *[]FixPlan) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	var checks []DependencyCheck
	var errorCount int
	var warningCount int

	// Check OpenSSL
	opensslCheck := DependencyCheck{
		Name:        "openssl",
		Description: "SSL/TLS toolkit",
		CanFix:      true,
	}

	if output, err := runDoctorCommandOutput("openssl version"); err == nil {
		opensslCheck.Installed = true
		opensslCheck.Version = extractVersion(output, "openssl")
		opensslCheck.Status = fmt.Sprintf("✅ Available (%s)", opensslCheck.Version)
	} else {
		opensslCheck.Installed = false
		opensslCheck.Status = "❌ Not available"
		opensslCheck.Errors = append(opensslCheck.Errors, "OpenSSL is required for SSL certificate generation")
		opensslCheck.FixCommand = getDistroCommand(detectDistribution(), "openssl")
		errorCount++
	}

	checks = append(checks, opensslCheck)

	// Check mkcert for trusted certificates
	mkcertCheck := DependencyCheck{
		Name:        "mkcert",
		Description: "Local trusted certificate authority",
		CanFix:      true,
	}

	if output, err := runDoctorCommandOutput("mkcert -version 2>/dev/null"); err == nil {
		mkcertCheck.Installed = true
		mkcertCheck.Version = extractVersion(output, "mkcert")
		mkcertCheck.Status = fmt.Sprintf("✅ Available (%s)", mkcertCheck.Version)
	} else {
		mkcertCheck.Installed = false
		mkcertCheck.Status = "⚠️ Not available (will use self-signed certificates)"
		mkcertCheck.Warnings = append(mkcertCheck.Warnings, "mkcert recommended for trusted SSL certificates")
		mkcertCheck.FixCommand = getDistroCommand(detectDistribution(), "mkcert")
		warningCount++
	}

	checks = append(checks, mkcertCheck)

	// Check OpenSSL Configuration for PHP
	opensslConfigCheck := DependencyCheck{
		Name:        "php-openssl-config",
		Description: "PHP OpenSSL configuration with CA certificates",
		CanFix:      true,
	}

	// Check if any PHP installations exist and validate OpenSSL configuration
	workspaceDir, err := workspace.Dir()
	if err == nil {
		phpDir := filepath.Join(workspaceDir, "php")
		if phpInstallations, err := os.ReadDir(phpDir); err == nil && len(phpInstallations) > 0 {
			// Found PHP installations, check OpenSSL configuration
			configIssues := []string{}
			configuredVersions := []string{}

			for _, phpVersionDir := range phpInstallations {
				if !phpVersionDir.IsDir() {
					continue
				}

				version := phpVersionDir.Name()
				opensslConfPath := filepath.Join(workspaceDir, "php", version, "etc", "conf.d", "openssl.ini")

				if _, err := os.Stat(opensslConfPath); err != nil {
					configIssues = append(configIssues, fmt.Sprintf("PHP %s: openssl.ini missing", version))
				} else {
					// Validate configuration content
					if content, err := os.ReadFile(opensslConfPath); err == nil {
						configStr := string(content)
						if strings.Contains(configStr, "openssl.cafile") && strings.Contains(configStr, "openssl.capath") {
							configuredVersions = append(configuredVersions, version)
						} else {
							configIssues = append(configIssues, fmt.Sprintf("PHP %s: openssl.ini incomplete", version))
						}
					}
				}
			}

			if len(configIssues) > 0 {
				opensslConfigCheck.Installed = false
				opensslConfigCheck.Status = "❌ OpenSSL configuration issues found"
				opensslConfigCheck.Errors = append(opensslConfigCheck.Errors, configIssues...)
				// Add fix plan for each PHP version that needs OpenSSL configuration
				for _, issue := range configIssues {
					if strings.Contains(issue, "missing") || strings.Contains(issue, "incomplete") {
						version := strings.TrimPrefix(issue, "PHP ")
						version = strings.TrimSuffix(version, ": openssl.ini missing")
						version = strings.TrimSuffix(version, ": openssl.ini incomplete")
						*fixPlans = append(*fixPlans, FixPlan{
							Name:        fmt.Sprintf("Generate OpenSSL config for PHP %s", version),
							Description: "Create openssl.ini with proper CA certificate paths",
							Command:     fmt.Sprintf("chauf doctor --fix-openssl --php-version %s", version),
							Category:    "SSL Certificate Dependencies",
							Priority:    2, // Warning level
						})
					}
				}
				errorCount++
			} else if len(configuredVersions) > 0 {
				opensslConfigCheck.Installed = true
				opensslConfigCheck.Version = fmt.Sprintf("configured for PHP %s", strings.Join(configuredVersions, ", "))
				opensslConfigCheck.Status = fmt.Sprintf("✅ Available (%s)", opensslConfigCheck.Version)
			} else {
				opensslConfigCheck.Installed = false
				opensslConfigCheck.Status = "⚠️ No OpenSSL configuration found"
				opensslConfigCheck.Warnings = append(opensslConfigCheck.Warnings, "PHP OpenSSL configuration recommended for secure connections")
				// Add fix plan to generate OpenSSL config for all PHP versions
				for _, phpVersionDir := range phpInstallations {
					if !phpVersionDir.IsDir() {
						continue
					}
					version := phpVersionDir.Name()
					*fixPlans = append(*fixPlans, FixPlan{
						Name:        fmt.Sprintf("Generate OpenSSL config for PHP %s", version),
						Description: "Create openssl.ini with proper CA certificate paths",
						Command:     fmt.Sprintf("chauf doctor --internal-fix-openssl --php-version %s --workspace-dir %s", version, workspaceDir),
						Category:    "SSL Certificate Dependencies",
						Priority:    2, // Warning level
					})
				}
				warningCount++
			}
		} else {
			opensslConfigCheck.Status = "ℹ️ No PHP installations found"
		}
	} else {
		opensslConfigCheck.Status = "⚠️ Could not access Chauffeur workspace"
		opensslConfigCheck.Warnings = append(opensslConfigCheck.Warnings, "Unable to validate OpenSSL configuration")
	}

	checks = append(checks, opensslConfigCheck)

	// Print results and collect fix plans
	for _, check := range checks {
		printCheckResult(logger, check, options, &errorCount, &warningCount, fixPlans, "SSL Certificate Dependencies")
	}

	return checks, errorCount, warningCount
}

// checkNetworkDependencies validates network and port dependencies
func checkNetworkDependencies(options DoctorOptions, fixPlans *[]FixPlan) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	var checks []DependencyCheck
	var errorCount int
	var warningCount int

	// Check iptables for port forwarding
	iptablesCheck := DependencyCheck{
		Name:        "iptables",
		Description: "Port forwarding for privileged ports",
		CanFix:      false, // System-level configuration
	}

	if _, err := exec.LookPath("iptables"); err == nil {
		iptablesCheck.Installed = true
		iptablesCheck.Status = "✅ Available"
	} else {
		iptablesCheck.Installed = false
		iptablesCheck.Status = "⚠️ Not available"
		iptablesCheck.Warnings = append(iptablesCheck.Warnings, "iptables required for port forwarding (80→8080, 443→8443)")
		iptablesCheck.FixCommand = "Install iptables using your system package manager"
		warningCount++
	}

	checks = append(checks, iptablesCheck)

	// Check available ports
	portCheck := DependencyCheck{
		Name:        "port-availability",
		Description: "Default Chauffeur ports",
		CanFix:      true,
	}

	// Check if default ports are available
	defaultPorts := []int{8080, 8443, 9000}
	var occupiedPorts []string

	for _, port := range defaultPorts {
		if isPortOccupiedByNonChauffeur(port) {
			occupiedPorts = append(occupiedPorts, strconv.Itoa(port))
		}
	}

	if len(occupiedPorts) == 0 {
		portCheck.Status = "✅ Default ports available"
	} else {
		portCheck.Status = fmt.Sprintf("⚠️ Ports occupied: %s", strings.Join(occupiedPorts, ", "))
		portCheck.Warnings = append(portCheck.Warnings, "Some default ports are in use")
		portCheck.FixCommand = "Configure alternative ports in ~/.chauffeur/config/chauffeur.yaml"
		warningCount++
	}

	checks = append(checks, portCheck)

	// Print results and collect fix plans
	for _, check := range checks {
		printCheckResult(logger, check, options, &errorCount, &warningCount, fixPlans, "Network & Port Dependencies")
	}

	return checks, errorCount, warningCount
}

// checkDNSDependencies validates DNS resolution for .test domains
func checkDNSDependencies(options DoctorOptions, fixPlans *[]FixPlan) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	var checks []DependencyCheck
	var errorCount int
	var warningCount int

	// Check dnsmasq installation
	dnsmasqCheck := DependencyCheck{
		Name:        "dnsmasq",
		Description: "Local DNS server for .test domains",
		CanFix:      true,
	}

	if _, err := exec.LookPath("dnsmasq"); err == nil {
		dnsmasqCheck.Installed = true
		dnsmasqCheck.Status = "✅ Available"
	} else {
		dnsmasqCheck.Installed = false
		dnsmasqCheck.Status = "❌ Not available"
		dnsmasqCheck.Errors = append(dnsmasqCheck.Errors, "dnsmasq is recommended for .test domain resolution")
		dnsmasqCheck.FixCommand = getDistroCommand(detectDistribution(), "dnsmasq")
		errorCount++
	}

	checks = append(checks, dnsmasqCheck)

	// Check if .test domains resolve
	dnsCheck := DependencyCheck{
		Name:        "dns-resolution",
		Description: ".test domain resolution",
		CanFix:      false,
	}

	// Test DNS resolution for a random .test domain
	testDomain := "chauffeur-dns-probe.test"
	if output, err := runDoctorCommandOutput(fmt.Sprintf("dig +short %s 2>/dev/null", testDomain)); err == nil && strings.TrimSpace(output) == "127.0.0.1" {
		dnsCheck.Status = "✅ .test domains resolve correctly"
	} else {
		dnsCheck.Status = "⚠️ .test domains not resolving"
		dnsCheck.Warnings = append(dnsCheck.Warnings, "Configure dnsmasq for .test domain resolution")
		dnsCheck.FixCommand = "Configure /etc/dnsmasq.d/chauffeur.conf and restart dnsmasq"
		warningCount++
	}

	checks = append(checks, dnsCheck)

	// Print results and collect fix plans
	for _, check := range checks {
		printCheckResult(logger, check, options, &errorCount, &warningCount, fixPlans, "DNS Resolution Dependencies")
	}

	return checks, errorCount, warningCount
}

// displayFixPlan shows the collected fix plan without executing (for --fix flag)
func displayFixPlan(logger *lib.Logger, fixPlans []FixPlan) {
	if len(fixPlans) == 0 {
		logger.Info("No fixes needed - all checks passed!")
		return
	}

	// Sort fixes by priority (errors first)
	for i := 0; i < len(fixPlans); i++ {
		for j := i + 1; j < len(fixPlans); j++ {
			if fixPlans[i].Priority > fixPlans[j].Priority {
				fixPlans[i], fixPlans[j] = fixPlans[j], fixPlans[i]
			}
		}
	}

	logger.PrintSection("🔧 Fix Plan")

	// Count errors vs warnings
	errorCount := 0
	warningCount := 0
	for _, plan := range fixPlans {
		if plan.Priority == 1 {
			errorCount++
		} else {
			warningCount++
		}
	}

	logger.Info(fmt.Sprintf("Found %d fixable issue(s): %d errors, %d warnings", len(fixPlans), errorCount, warningCount))
	logger.Info("")

	// Group fixes by category
	categories := make(map[string][]FixPlan)
	for _, plan := range fixPlans {
		categories[plan.Category] = append(categories[plan.Category], plan)
	}

	// Show fixes by category
	for category, plans := range categories {
		logger.Info(fmt.Sprintf("%s:", category))
		for _, plan := range plans {
			priorityIcon := "❌"
			if plan.Priority == 2 {
				priorityIcon = "⚠️"
			}
			logger.Info(fmt.Sprintf("  %s %s: %s", priorityIcon, plan.Name, plan.Description))
			logger.Info(fmt.Sprintf("    Command: %s", plan.Command))
		}
		logger.Info("")
	}

	// Show hint to use --auto-fix
	logger.Info("To execute these fixes, run: chauf doctor --auto-fix")
}

// executeFixPlan shows the fix plan and asks for user confirmation before executing

func executeFixPlan(logger *lib.Logger, fixPlans []FixPlan, quiet bool) bool {
	if len(fixPlans) == 0 {
		return true
	}

	// Sort fixes by priority (errors first)
	for i := 0; i < len(fixPlans); i++ {
		for j := i + 1; j < len(fixPlans); j++ {
			if fixPlans[i].Priority > fixPlans[j].Priority {
				fixPlans[i], fixPlans[j] = fixPlans[j], fixPlans[i]
			}
		}
	}

	logger.PrintSection("🔧 Fix Plan")

	// Count errors vs warnings
	errorCount := 0
	warningCount := 0
	for _, plan := range fixPlans {
		if plan.Priority == 1 {
			errorCount++
		} else {
			warningCount++
		}
	}

	logger.Info(fmt.Sprintf("Found %d fixable issue(s): %d errors, %d warnings", len(fixPlans), errorCount, warningCount))
	logger.Info("")

	// Group fixes by category
	categories := make(map[string][]FixPlan)
	for _, plan := range fixPlans {
		categories[plan.Category] = append(categories[plan.Category], plan)
	}

	// Show fixes by category
	for category, plans := range categories {
		logger.Info(fmt.Sprintf("%s:", category))
		for _, plan := range plans {
			priorityIcon := "❌"
			if plan.Priority == 2 {
				priorityIcon = "⚠️"
			}
			logger.Info(fmt.Sprintf("  %s %s: %s", priorityIcon, plan.Name, plan.Description))
			logger.Info(fmt.Sprintf("    Command: %s", plan.Command))
		}
		logger.Info("")
	}

	// Ask for confirmation
	logger.Info("This will execute the above commands to fix the identified issues.")
	logger.Warn("⚠️ Some commands may require sudo privileges.", "")

	logger.Prompt("Do you want to proceed with these fixes? [y/N]", "")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		logger.Info("Auto-fix cancelled by user.")
		return false
	}

	// Execute fixes
	logger.Info("")
	logger.Info("Executing fixes...")

	successCount := 0
	for i, plan := range fixPlans {
		logger.Info(fmt.Sprintf("[%d/%d] Fixing %s...", i+1, len(fixPlans), plan.Name))

		cmd := exec.Command("sh", "-c", plan.Command)
		output, err := cmd.CombinedOutput()

		if err != nil {
			logger.Error(fmt.Sprintf("  ✗ Failed: %v", err), "")
			if len(output) > 0 {
				logger.Error(fmt.Sprintf("    Output: %s", strings.TrimSpace(string(output))), "")
			}
		} else {
			logger.Success("  ✓ Fixed", "")
			successCount++
			if len(output) > 0 && !quiet {
				logger.Info(fmt.Sprintf("    Output: %s", strings.TrimSpace(string(output))))
			}
		}
	}

	// Show summary
	logger.Info("")
	logger.Info(fmt.Sprintf("Fix summary: %d/%d successful", successCount, len(fixPlans)))

	if successCount == len(fixPlans) {
		logger.Success("All fixes completed successfully!", "")
	} else {
		logger.Warn(fmt.Sprintf("%d fix(es) failed. You may need to run them manually.", len(fixPlans)-successCount), "")
	}

	return true
}

// Helper functions

func runDoctorCommandOutput(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w (%s)", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func extractVersion(output, tool string) string {
	// Extract version from various command outputs
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return ""
	}

	versionLine := strings.TrimSpace(lines[0])

	// Common version patterns
	patterns := []string{
		`v?(\d+\.\d+(?:\.\d+)?)`,
		`(\d+\.\d+(?:\.\d+)?-\S+)`,
		`(\d+\.\d+(?:\.\d+)?)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(versionLine); len(matches) > 1 {
			return matches[1]
		}
	}

	return versionLine
}

func detectDistribution() string {
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "ubuntu/debian"
	}
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}
	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return "fedora"
	}
	return "unknown"
}

func getDistroCommand(distro, package_ string) string {
	commands := map[string]map[string]string{
		"pkg-config": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y pkg-config",
			"arch":          "sudo pacman -S pkgconf",
			"fedora":        "sudo dnf install -y pkgconfig",
		},
		"openssl": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y openssl",
			"arch":          "sudo pacman -S openssl",
			"fedora":        "sudo dnf install -y openssl-devel",
		},
		"dnsmasq": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y dnsmasq",
			"arch":          "sudo pacman -S dnsmasq",
			"fedora":        "sudo dnf install -y dnsmasq",
		},
		"mkcert": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y mkcert",
			"arch":          "sudo pacman -S mkcert",
			"fedora":        "sudo dnf install -y mkcert",
		},
	}

	if pkgCommands, exists := commands[package_]; exists {
		if cmd, exists := pkgCommands[distro]; exists {
			return cmd
		}
	}

	return fmt.Sprintf("Install %s using your system package manager", package_)
}

func isPortOccupied(port int) bool {
	// Try to bind to the port to see if it's occupied
	addr := fmt.Sprintf(":%d", port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return true // Port is occupied
	}
	listener.Close()
	return false
}

// isPortOccupiedByNonChauffeur checks if a port is occupied by a non-Chauffeur service
func isPortOccupiedByNonChauffeur(port int) bool {
	// First check if port is occupied at all
	if !isPortOccupied(port) {
		return false // Port is free
	}

	// Load configuration to get expected Chauffeur ports
	cfg, err := config.Load()
	if err != nil {
		// If we can't load config, assume it's a conflict (better safe than sorry)
		return true
	}

	// Check if this is one of Chauffeur's expected ports
	chauffeurPorts := map[int]bool{
		cfg.Nginx.HTTPPort:       true,
		cfg.Nginx.HTTPSPort:      true,
		cfg.Ports.PHPFPMFallback: true,
	}

	// Add ports from the dynamic range
	for p := cfg.Ports.StartRange; p <= cfg.Ports.EndRange; p++ {
		chauffeurPorts[p] = true
	}

	// If this isn't a Chauffeur port, it's definitely a conflict
	if !chauffeurPorts[port] {
		return true
	}

	// Check if any Chauffeur service is running that would use this port
	serviceManager, err := services.NewServiceManager()
	if err != nil {
		// If we can't get service manager, assume it's a conflict
		return true
	}

	// Get all services and check if any running service would use this port
	allServices := serviceManager.ListGlobalServices()

	// Try to get project services (this might fail if no projects exist)
	if projectServices, err := serviceManager.ListProjectServices(""); err == nil {
		allServices = append(allServices, projectServices...)
	}

	for _, svc := range allServices {
		running, err := serviceManager.IsRunning(svc)
		if err == nil && running {
			// Check if this service would use the port we're checking
			if serviceUsesPort(svc, port, cfg) {
				return false // Port is used by running Chauffeur service, not a conflict
			}
		}
	}

	// Port is occupied but not by a running Chauffeur service
	return true
}

// serviceUsesPort checks if a service is expected to use a specific port
func serviceUsesPort(svc services.Service, port int, cfg config.Config) bool {
	switch svc.Name {
	case "chauf-nginx":
		return port == cfg.Nginx.HTTPPort || port == cfg.Nginx.HTTPSPort
	default:
		// For PHP-FPM services, check if they use a port that matches our check
		if strings.HasPrefix(svc.Name, "chauf-php-fpm-") {
			return port == cfg.Ports.PHPFPMFallback
		}
	}
	return false
}

func printDoctorSummary(logger *lib.Logger, checks []DependencyCheck, hasErrors, hasWarnings bool) {
	logger.PrintSection("📊 Summary")

	totalChecks := len(checks)
	errorChecks := 0
	warningChecks := 0
	passedChecks := 0

	for _, check := range checks {
		if len(check.Errors) > 0 {
			errorChecks++
		} else if len(check.Warnings) > 0 {
			warningChecks++
		} else {
			passedChecks++
		}
	}

	logger.Info(fmt.Sprintf("Total checks: %d", totalChecks))
	logger.Success("Checks passed", fmt.Sprintf("%d", passedChecks))

	if warningChecks > 0 {
		logger.Warn("Warnings found", fmt.Sprintf("%d", warningChecks))
	}

	if errorChecks > 0 {
		logger.Error("Errors found", fmt.Sprintf("%d", errorChecks))
	}

	if hasErrors {
		logger.Error("Overall Status", "❌ Issues found that need attention")
	} else if hasWarnings {
		logger.Warn("Overall Status", "⚠️ System functional but has warnings")
	} else {
		logger.Success("Overall Status", "✅ All systems healthy")
	}

	// Provide next steps
	if hasErrors {
		logger.PrintSection("🔧 Recommended Actions")
		logger.Info("Run 'chauf doctor --fix' to see suggested fixes")
		logger.Info("Run 'chauf doctor --auto-fix' to attempt automatic fixes")
	}
}

func countErrors(checks []DependencyCheck) int {
	count := 0
	for _, check := range checks {
		if len(check.Errors) > 0 {
			count++
		}
	}
	return count
}

// executeFix attempts to run a fix command
func executeFix(logger *lib.Logger, command, packageName string, quiet bool) error {
	logger.Info(fmt.Sprintf("    Running fix for %s...", packageName))

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}

	if len(output) > 0 && !quiet {
		logger.Info(fmt.Sprintf("    Output: %s", strings.TrimSpace(string(output))))
	}

	return nil
}

// validateFix re-checks a dependency after attempting to fix it
func validateFix(check DependencyCheck) DependencyCheck {
	// Re-run the version check for the package
	var cmd string
	switch check.Name {
	case "git", "curl", "tar", "gcc", "make", "pkg-config":
		cmd = fmt.Sprintf("%s --version", check.Name)
	case "mkcert":
		cmd = "mkcert -version"
	case "openssl":
		cmd = "openssl version"
	case "dnsmasq":
		cmd = "dnsmasq --version"
	default:
		// For pkg-config based dependencies, try to check with pkg-config
		cmd = fmt.Sprintf("pkg-config --modversion %s 2>/dev/null", check.Name)
	}

	if output, err := runDoctorCommandOutput(cmd); err == nil {
		check.Installed = true
		check.Version = extractVersion(output, check.Name)
	}

	return check
}

// printCheckResult handles the unified printing of dependency checks and collects fix plans
func printCheckResult(logger *lib.Logger, check DependencyCheck, options DoctorOptions, errorCount, warningCount *int, fixPlans *[]FixPlan, category string) {
	if !options.Quiet {
		if len(check.Errors) > 0 {
			logger.Error(check.Status, check.Name)
			if options.Verbose {
				for _, err := range check.Errors {
					logger.Error("  ├─ Error", err)
				}
			}
			// Collect fix plan when --fix is set (for both --fix and --auto-fix modes)
			if options.Fix && check.FixCommand != "" && check.CanFix {
				fixPlan := FixPlan{
					Name:        check.Name,
					Description: check.Description,
					Command:     check.FixCommand,
					Category:    category,
					Priority:    1, // Error priority
				}
				*fixPlans = append(*fixPlans, fixPlan)
			}
		} else if len(check.Warnings) > 0 {
			logger.Warn(check.Status, check.Name)
			if options.Verbose {
				for _, warn := range check.Warnings {
					logger.Warn("  ├─ Warning", warn)
				}
			}
			// Collect fix plan for warnings when --fix is set
			if options.Fix && check.CanFix && check.FixCommand != "" && !check.Installed {
				fixPlan := FixPlan{
					Name:        check.Name,
					Description: check.Description,
					Command:     check.FixCommand,
					Category:    category,
					Priority:    2, // Warning priority
				}
				*fixPlans = append(*fixPlans, fixPlan)
			}
		} else {
			if !options.Quiet {
				logger.Success(check.Status, check.Name)
				if options.Verbose && check.Version != "" {
					logger.Info(fmt.Sprintf("  └─ Version: %s", check.Version))
				}
			}
		}
	}
}
