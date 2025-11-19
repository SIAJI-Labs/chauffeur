package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/services"
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

	// Run dependency checks
	if options.CheckDeps {
		logger.PrintSection("System Dependencies")
		checks, errs, warns := checkSystemDependencies(options)
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
		checks, errs, warns := checkPHPBuildDependencies(options)
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
		checks, errs, warns := checkSSLDependencies(options)
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
		checks, errs, warns := checkNetworkDependencies(options)
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
		checks, errs, warns := checkDNSDependencies(options)
		allChecks = append(allChecks, checks...)
		if errs > 0 {
			hasErrors = true
		}
		if warns > 0 {
			hasWarnings = true
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
		default:
			fmt.Printf("Unknown flag: %s\n\n", arg)
			printDoctorUsage()
			os.Exit(1)
		}
	}

	return options
}

// printDoctorUsage displays usage information for the doctor command
func printDoctorUsage() {
	fmt.Printf(`Chauffeur Doctor - Health Check and Troubleshooting

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
  --fix, -f            Show fix suggestions for issues found
  --auto-fix           Attempt to automatically fix issues where safe
  --quiet, -q          Suppress non-error output
  --help, -h           Show this help message

Examples:
  chauf doctor                    # Run all health checks
  chauf doctor --check-deps       # Check only system dependencies
  chauf doctor --check-php --fix  # Check PHP dependencies and show fixes
  chauf doctor --auto-fix         # Attempt to fix issues automatically

Description:
  Performs comprehensive health checks on your Chauffeur installation,
  validates system dependencies, and provides guidance for resolving issues.
`)
}

// checkSystemDependencies validates core system dependencies
func checkSystemDependencies(options DoctorOptions) ([]DependencyCheck, int, int) {
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
				"centos/rhel":   "sudo yum install -y git",
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
				"centos/rhel":   "sudo yum install -y curl",
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
				"centos/rhel":   "sudo yum install -y tar",
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
				"centos/rhel":   "sudo yum groupinstall -y 'Development Tools'",
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
				"centos/rhel":   "sudo yum install -y make",
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
				"centos/rhel":   "sudo yum install -y pkgconfig",
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

		// Print check result
		if !options.Quiet {
			if len(check.Errors) > 0 {
				logger.Error(check.Status, check.Name)
				if options.Verbose {
					for _, err := range check.Errors {
						logger.Error("  ├─ Error", err)
					}
				}
				if options.Fix && check.FixCommand != "" {
					logger.Info(fmt.Sprintf("  └─ Fix: %s", check.FixCommand))
				}
			} else if len(check.Warnings) > 0 {
				logger.Warn(check.Status, check.Name)
				if options.Verbose {
					for _, warn := range check.Warnings {
						logger.Warn("  ├─ Warning", warn)
					}
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

	return checks, errorCount, warningCount
}

// checkPHPBuildDependencies validates PHP-specific build dependencies
func checkPHPBuildDependencies(options DoctorOptions) ([]DependencyCheck, int, int) {
	logger := lib.NewCommandLogger("doctor")

	// PHP build dependencies based on AGENTS.md requirements
	dependencies := []struct {
		name         string
		pkgConfig    string
		description  string
		required     bool
		fixCommands  map[string]string
	}{
		{
			name:        "libzip",
			pkgConfig:   "libzip",
			description: "ZIP archive support",
			required:    true,
			fixCommands: map[string]string{
				"ubuntu/debian": "sudo apt install -y libzip-dev",
				"centos/rhel":   "sudo yum install -y libzip-devel",
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
				"centos/rhel":   "sudo yum install -y libjpeg-turbo-devel",
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
				"centos/rhel":   "sudo yum install -y libpng-devel",
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
				"centos/rhel":   "sudo yum install -y freetype-devel",
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
				"centos/rhel":   "sudo yum install -y libxml2-devel",
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
				"centos/rhel":   "sudo yum install -y libcurl-devel",
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
				"centos/rhel":   "sudo yum install -y zlib-devel",
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
				"centos/rhel":   "sudo yum install -y readline-devel",
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
				"centos/rhel":   "sudo yum install -y libxslt-devel",
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
				"centos/rhel":   "sudo yum install -y ImageMagick-devel",
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
				"centos/rhel":   "sudo yum install -y gmp-devel",
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

		// Print check result
		if !options.Quiet {
			if len(check.Errors) > 0 {
				logger.Error(check.Status, check.Name)
				if options.Verbose {
					for _, err := range check.Errors {
						logger.Error("  ├─ Error", err)
					}
				}
				if options.Fix && check.FixCommand != "" {
					logger.Info(fmt.Sprintf("  └─ Fix: %s", check.FixCommand))
				}
			} else if len(check.Warnings) > 0 {
				logger.Warn(check.Status, check.Name)
				if options.Verbose {
					for _, warn := range check.Warnings {
						logger.Warn("  ├─ Warning", warn)
					}
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

	return checks, errorCount, warningCount
}

// checkSSLDependencies validates SSL certificate dependencies
func checkSSLDependencies(options DoctorOptions) ([]DependencyCheck, int, int) {
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
		mkcertCheck.FixCommand = "go install -r filippo.io/mkcert@latest && mkcert -install"
		warningCount++
	}

	checks = append(checks, mkcertCheck)

	// Print results
	for _, check := range checks {
		if !options.Quiet {
			if len(check.Errors) > 0 {
				logger.Error(check.Status, check.Name)
				if options.Verbose {
					for _, err := range check.Errors {
						logger.Error("  ├─ Error", err)
					}
				}
				if options.Fix && check.FixCommand != "" {
					logger.Info(fmt.Sprintf("  └─ Fix: %s", check.FixCommand))
				}
			} else if len(check.Warnings) > 0 {
				logger.Warn(check.Status, check.Name)
				if options.Verbose {
					for _, warn := range check.Warnings {
						logger.Warn("  ├─ Warning", warn)
					}
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

	return checks, errorCount, warningCount
}

// checkNetworkDependencies validates network and port dependencies
func checkNetworkDependencies(options DoctorOptions) ([]DependencyCheck, int, int) {
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

	// Print results
	for _, check := range checks {
		if !options.Quiet {
			if len(check.Errors) > 0 {
				logger.Error(check.Status, check.Name)
				if options.Verbose {
					for _, err := range check.Errors {
						logger.Error("  ├─ Error", err)
					}
				}
				if options.Fix && check.FixCommand != "" {
					logger.Info(fmt.Sprintf("  └─ Fix: %s", check.FixCommand))
				}
			} else if len(check.Warnings) > 0 {
				logger.Warn(check.Status, check.Name)
				if options.Verbose {
					for _, warn := range check.Warnings {
						logger.Warn("  ├─ Warning", warn)
					}
				}
			} else {
				if !options.Quiet {
					logger.Success(check.Status, check.Name)
				}
			}
		}
	}

	return checks, errorCount, warningCount
}

// checkDNSDependencies validates DNS resolution for .test domains
func checkDNSDependencies(options DoctorOptions) ([]DependencyCheck, int, int) {
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

	// Print results
	for _, check := range checks {
		if !options.Quiet {
			if len(check.Errors) > 0 {
				logger.Error(check.Status, check.Name)
				if options.Verbose {
					for _, err := range check.Errors {
						logger.Error("  ├─ Error", err)
					}
				}
				if options.Fix && check.FixCommand != "" {
					logger.Info(fmt.Sprintf("  └─ Fix: %s", check.FixCommand))
				}
			} else if len(check.Warnings) > 0 {
				logger.Warn(check.Status, check.Name)
				if options.Verbose {
					for _, warn := range check.Warnings {
						logger.Warn("  ├─ Warning", warn)
					}
				}
			} else {
				if !options.Quiet {
					logger.Success(check.Status, check.Name)
				}
			}
		}
	}

	return checks, errorCount, warningCount
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
	if _, err := os.Stat("/etc/centos-release"); err == nil {
		return "centos/rhel"
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
			"centos/rhel":   "sudo yum install -y pkgconfig",
			"arch":          "sudo pacman -S pkgconf",
			"fedora":        "sudo dnf install -y pkgconfig",
		},
		"openssl": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y openssl",
			"centos/rhel":   "sudo yum install -y openssl-devel",
			"arch":          "sudo pacman -S openssl",
			"fedora":        "sudo dnf install -y openssl-devel",
		},
		"dnsmasq": {
			"ubuntu/debian": "sudo apt update && sudo apt install -y dnsmasq",
			"centos/rhel":   "sudo yum install -y dnsmasq",
			"arch":          "sudo pacman -S dnsmasq",
			"fedora":        "sudo dnf install -y dnsmasq",
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
		cfg.Nginx.HTTPPort:     true,
		cfg.Nginx.HTTPSPort:    true,
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