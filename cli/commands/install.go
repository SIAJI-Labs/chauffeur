package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/example"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

var phpInstallFunc = defaultPHPInstall

func defaultPHPInstall(version, prefix string, force bool, enableGD bool) error {
	info, err := system.Detect()
	if err != nil {
		return err
	}
	return installers.InstallPHPSource(version, installers.InstallOptions{
		Prefix:   prefix,
		Force:    force,
		Info:     info,
		EnableGD: enableGD,
	})
}

// OverridePHPInstallFunc lets tests inject a fake PHP installer.
func OverridePHPInstallFunc(fn func(version, prefix string, force bool, enableGD bool) error) (reset func()) {
	prev := phpInstallFunc
	if fn != nil {
		phpInstallFunc = fn
	}
	return func() {
		phpInstallFunc = prev
	}
}

/**
 * RunInstall handles `chauf install <service...>` invocations.
 *
 * @param args CLI arguments passed after the install subcommand.
 * @return error when parsing fails or an installation step errors.
 */
func RunInstall(args []string) error {
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

	logger := lib.NewCommandLogger("install")

	if len(args) == 0 {
		printInstallUsage()
		return errors.New("no services specified")
	}

	var (
		force    bool
		local    bool
		noCache  bool
		services []string
		versions map[string]string
	)

	versions = make(map[string]string)

	// Parse arguments to handle service[version] syntax (e.g., "php8.3")
	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "--force":
			force = true
			i++
		case "--local":
			local = true
			i++
		case "--no-cache":
			noCache = true
			i++
		case "--help", "-h":
			printInstallUsage()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for install: %s", arg)
			}

			// Handle service[version] syntax
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

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	if err := workspace.Ensure(); err != nil {
		return err
	}

	info, err := system.Detect()
	if err != nil {
		return err
	}

	for _, name := range services {
		if name == "php" {
			// Handle PHP installation with version support
			version := versions["php"]
			if err := runPHPInstall(version, force, local, noCache); err != nil {
				return err
			}
		} else if name == "composer" {
			// Handle Composer installation
			logger.Info("Installing Composer (PHP dependency manager)...")
			ok, err := handleComposerInstall(prefix, info, force)
			if err != nil {
				return logger.Error("install composer", err.Error())
			}
			if !ok && !force {
				logger.Info("Composer already installed")
			} else {
				logger.Success("Composer installed successfully", "Uses Chauffeur PHP version isolation")
			}
		} else {
			// Handle other services (nginx)
			spec, err := NewServiceSpec(name, prefix, info)
			if err != nil {
				return err
			}

			ok, err := spec.available()
			if err != nil {
				return err
			}
			if ok && !force {
				logger.Info(fmt.Sprintf("%s already installed at %s", spec.Name, spec.BinaryPath))
				continue
			}

			logger.Info(fmt.Sprintf("Installing %s (%s)...", spec.Name, spec.Description))
			if err := spec.install(force); err != nil {
				return logger.Error(fmt.Sprintf("install %s", spec.Name), err.Error())
			}
			logger.Success(fmt.Sprintf("Installed %s successfully", spec.Name), "")
		}
	}

	// Check if we should create/link example project
	if shouldCreateExampleProject(services) {
		logger.Info("Checking example project status...")
		if err := example.LinkExampleProjectIfReady(); err != nil {
			logger.Warn("Failed to link example project", err.Error())
		} else if example.IsExampleProjectLinked() {
			logger.Success("Example project is ready!", "")
			// Load config for port information
			cfg, err := config.Load()
			if err == nil {
				httpURL := fmt.Sprintf("http://%s", example.ExampleProjectDomain)
				if cfg.Nginx.HTTPPort != 80 {
					httpURL = fmt.Sprintf("http://%s:%d", example.ExampleProjectDomain, cfg.Nginx.HTTPPort)
				}
				logger.Info(fmt.Sprintf("Access it at: %s", httpURL))
				logger.Info("The example project helps you test your Chauffeur setup")
				logger.Info("Remove it anytime with: chauf unlink --project example-project")
			}
		}
	}

	return nil
}

/**
 * isValidPHPVersion checks if a string looks like a valid PHP version.
 * Simple heuristic - starts with a digit and contains dots.
 */
func isValidPHPVersion(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] >= '7' && s[0] <= '9' && strings.Contains(s, ".")
}

/**
 * runPHPInstall handles PHP installation with version selection.
 */
func runPHPInstall(version string, force bool, local bool, noCache bool) error {
	logger := lib.NewCommandLogger("install")

	var err error

	// If no version specified, show interactive selection
	if version == "" {
		selected, err := selectPHPVersion()
		if err != nil {
			return err
		}
		version = selected
	}

	// Validate PHP version
	if !installers.IsPHPVersionSupported(version) {
		return fmt.Errorf("PHP version %s is not supported. Supported versions: %s", version, installers.GetSupportedVersionsList())
	}

	// Prompt for GD extension for legacy PHP versions
	enableGD, err := installers.PromptGDExtension(version, logger, force)
	if err != nil {
		return logger.Error("gd extension prompt", err.Error())
	}

	prefix, err := workspace.Dir()
	if err != nil {
		return err
	}

	if err := workspace.Ensure(); err != nil {
		return err
	}

	// Handle local installation if requested
	if local {
		localPath, err := handleLocalPHPInstall(version, logger)
		if err != nil {
			return logger.Error("local php install", err.Error())
		}
		if localPath != "" {
			// Set environment variable for the installer to use
			os.Setenv("CHAUFFEUR_PHP_TARBALL", localPath)
			logger.Info(fmt.Sprintf("Using local PHP tarball: %s", localPath))
		} else {
			// User chose to download from internet instead
			logger.Info("Falling back to internet download")
		}
	}

	// Set no-cache flag if requested
	if noCache {
		os.Setenv("CHAUFFEUR_NO_CACHE", "1")
		logger.Info("Skipping download caching")
	}

	logger.Info(fmt.Sprintf("Installing PHP %s...", version))
	if err := phpInstallFunc(version, prefix, force, enableGD); err != nil {
		return logger.Error(fmt.Sprintf("install php %s", version), err.Error())
	}
	logger.Success(fmt.Sprintf("Installed PHP %s successfully", version), "")

	return nil
}

/**
 * selectPHPVersion shows an interactive menu for PHP version selection.
 */
func selectPHPVersion() (string, error) {
	logger := lib.NewCommandLogger("install")
	logger.Info("Select PHP version to install:")

	prefix, err := workspace.Dir()
	if err != nil {
		return "", err
	}

	// Get supported versions
	versions := installers.GetSupportedPHPVersions()

	// Check what versions are already installed
	installed := make(map[string]bool)
	for _, versionMeta := range versions {
		target := fmt.Sprintf("%s/php/%s/bin/php", prefix, versionMeta.Version)
		if _, err := os.Stat(target); err == nil {
			installed[versionMeta.Version] = true
		}
	}

	// Display options
	for i, versionMeta := range versions {
		status := ""
		if installed[versionMeta.Version] {
			status = " (installed)"
		}
		if versionMeta.EndOfLife {
			status += " (EOL)"
		}
		logger.Info(fmt.Sprintf("  %d) PHP %s%s", i+1, versionMeta.Version, status))
	}

	logger.Info(fmt.Sprintf("Enter your choice (1-%d):", len(versions)))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid input: %s", input)
	}

	if choice < 1 || choice > len(versions) {
		return "", fmt.Errorf("invalid choice: %d", choice)
	}

	selected := versions[choice-1]
	logger.Success(fmt.Sprintf("Selected PHP %s", selected.Version), "")
	return selected.Version, nil
}

/**
 * handleComposerInstall manages Composer installation and updates.
 */
func handleComposerInstall(prefix string, info system.Info, force bool) (bool, error) {
	spec, err := NewServiceSpec("composer", prefix, info)
	if err != nil {
		return false, err
	}

	// Check if already available
	available, err := spec.available()
	if err != nil {
		return false, err
	}

	// If available and not forcing, skip installation
	if available && !force {
		return false, nil
	}

	// Install Composer using the service spec
	if err := spec.install(force); err != nil {
		return false, err
	}

	return true, nil
}

/**
 * handleLocalPHPInstall manages the local tarball workflow for PHP installation.
 * @param version PHP version being installed
 * @param logger Logger instance
 * @return Path to local tarball if successful/selected, empty string if user wants internet download
 */
func handleLocalPHPInstall(version string, logger *lib.Logger) (string, error) {
	// 1. Check config for stored tarball path
	configPath, err := config.GetLocalTarballPath(version)
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	if configPath != "" {
		// Validate the configured path exists and is accessible
		if _, err := os.Stat(configPath); err != nil {
			logger.Warn("Configured tarball not accessible", err.Error())
			logger.Info("Configured path will be removed from config")
			// Remove invalid path from config
			if err := config.SetLocalTarballPath(version, ""); err != nil {
				logger.Warn("Failed to remove invalid path from config", err.Error())
			}
		} else {
			// Validate tarball version matches
			if err := validatePHPTarball(configPath, version, logger); err != nil {
				logger.Warn("Configured tarball validation failed", err.Error())
				logger.Info("Configured path will be removed from config")
				// Remove mismatched path from config
				if err := config.SetLocalTarballPath(version, ""); err != nil {
					logger.Warn("Failed to remove mismatched path from config", err.Error())
				}
			} else {
				// Configured path is valid
				logger.Success("Using configured local tarball", configPath)
				return configPath, nil
			}
		}
	}

	// 2. Prompt user for tarball location
	logger.Info("No valid local tarball found in config")
	return promptForLocalTarball(version, logger)
}

/**
 * validatePHPTarball checks if the tarball matches the expected PHP version.
 * @param tarballPath Path to the tarball file
 * @param expectedVersion Expected PHP version (e.g., "8.3")
 * @param logger Logger instance
 * @return error if validation fails
 */
func validatePHPTarball(tarballPath, expectedVersion string, logger *lib.Logger) error {
	// Check file exists and is readable
	info, err := os.Stat(tarballPath)
	if err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(tarballPath), ".tar.gz") &&
		!strings.HasSuffix(strings.ToLower(tarballPath), ".tgz") &&
		!strings.HasSuffix(strings.ToLower(tarballPath), ".tar.bz2") &&
		!strings.HasSuffix(strings.ToLower(tarballPath), ".tar.xz") {
		return fmt.Errorf("invalid file extension, expected .tar.gz, .tgz, .tar.bz2, or .tar.xz")
	}

	// Extract filename and check version
	filename := filepath.Base(tarballPath)

	// Try to extract version from filename
	// Common patterns: php-8.3.27.tar.gz, php-8.3.27.tgz, etc.
	var extractedVersion string

	// Remove file extensions
	baseName := strings.TrimSuffix(filename, ".tar.gz")
	baseName = strings.TrimSuffix(baseName, ".tgz")
	baseName = strings.TrimSuffix(baseName, ".tar.bz2")
	baseName = strings.TrimSuffix(baseName, ".tar.xz")

	// Remove php- prefix
	if strings.HasPrefix(baseName, "php-") {
		extractedVersion = strings.TrimPrefix(baseName, "php-")
	} else if strings.HasPrefix(baseName, "php") {
		extractedVersion = strings.TrimPrefix(baseName, "php")
	}

	// Validate extracted version matches expected major.minor
	if extractedVersion == "" {
		return fmt.Errorf("could not extract version from filename: %s", filename)
	}

	// Extract major.minor from extracted version
	parts := strings.Split(extractedVersion, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid version format in filename: %s", extractedVersion)
	}

	majorMinor := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if majorMinor != expectedVersion {
		return fmt.Errorf("version mismatch: filename contains PHP %s but requested PHP %s", majorMinor, expectedVersion)
	}

	logger.Info(fmt.Sprintf("Validated tarball version: %s", extractedVersion))
	return nil
}

/**
 * promptForLocalTarball prompts the user to provide a tarball path or choose internet download.
 * @param version PHP version being installed
 * @param logger Logger instance
 * @return Path to local tarball if provided, empty string if user chooses internet
 */
func promptForLocalTarball(version string, logger *lib.Logger) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\nLocal PHP %s tarball not configured.\n", version)
		fmt.Println("Options:")
		fmt.Println("  1) Enter path to local PHP tarball")
		fmt.Println("  2) Download from internet (default)")
		fmt.Print("Enter your choice (1-2) or press Enter for internet download: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" || input == "2" {
			// User wants internet download
			logger.Info("User selected internet download")
			return "", nil
		}

		if input == "1" {
			// User wants to provide local path
			fmt.Print("Enter path to PHP tarball: ")
			pathInput, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("failed to read path: %w", err)
			}

			tarballPath := strings.TrimSpace(pathInput)
			if tarballPath == "" {
				logger.Warn("Empty path provided", "Please enter a valid file path")
				continue
			}

			// Expand ~ to home directory
			if strings.HasPrefix(tarballPath, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					logger.Warn("Could not expand home directory", err.Error())
					continue
				}
				if tarballPath == "~" {
					tarballPath = home
				} else {
					tarballPath = filepath.Join(home, tarballPath[2:])
				}
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(tarballPath)
			if err != nil {
				logger.Warn("Could not resolve absolute path", err.Error())
				continue
			}

			// Validate the tarball
			if err := validatePHPTarball(absPath, version, logger); err != nil {
				logger.Warn("Invalid tarball", err.Error())
				fmt.Println("Would you like to:")
				fmt.Println("  1) Try a different path")
				fmt.Println("  2) Download from internet")
				fmt.Print("Enter your choice (1-2): ")

				choiceInput, err := reader.ReadString('\n')
				if err != nil {
					return "", fmt.Errorf("failed to read choice: %w", err)
				}

				choice := strings.TrimSpace(choiceInput)
				if choice == "2" {
					logger.Info("User selected internet download after validation failure")
					return "", nil
				}
				continue
			}

			// Ask if user wants to save this path for future use
			fmt.Printf("Save this path for future PHP %s installations? (y/N): ", version)
			saveInput, err := reader.ReadString('\n')
			if err != nil {
				logger.Warn("Could not read save preference", err.Error())
			} else if strings.ToLower(strings.TrimSpace(saveInput)) == "y" {
				if err := config.SetLocalTarballPath(version, absPath); err != nil {
					logger.Warn("Failed to save path to config", err.Error())
				} else {
					logger.Success("Saved path to config", absPath)
				}
			}

			return absPath, nil
		} else {
			logger.Warn("Invalid choice", "Please enter 1 or 2")
		}
	}
}

/**
 * printInstallUsage renders CLI help for the install command.
 */
// shouldCreateExampleProject determines if example project should be created based on installed services
func shouldCreateExampleProject(services []string) bool {
	// Only check if nginx or php was just installed
	nginxInstalled := false
	phpInstalled := false
	
	for _, service := range services {
		if service == "nginx" {
			nginxInstalled = true
		}
		if service == "php" {
			phpInstalled = true
		}
	}
	
	// Trigger example project if either nginx or php was installed
	return nginxInstalled || phpInstalled
}

// printInstallUsage renders CLI help for the install command.
func printInstallUsage() {
	fmt.Println(`Usage: chauf install [--force] [--local] [--no-cache] <service> [<version>...] [<service>...]

Installs one or more Chauffeur-managed services.

Options:
  --force     Reinstall even if the service is already present.
  --local     Use local PHP tarball from config or prompt for location (PHP only).
  --no-cache  Skip automatic caching of downloaded files (all services).
  -h, --help  Show this message.

Services:
  composer   PHP dependency manager with Chauffeur PHP version isolation.
  nginx      Source build from the latest GitHub release.
  php        Source build with development extensions.

Service Installation:
  chauf install php              Interactive PHP version selection.
  chauf install nginx            Install latest nginx (auto-cached).
  chauf install composer          Install Composer (auto-cached).
  chauf install php 8.3          Install specific PHP version (auto-cached).
  chauf install php 8.3 --local   Use local tarball for PHP 8.3.
  chauf install composer --force  Reinstall selected service.
  chauf install php --no-cache   Install without caching downloads.

Smart Caching System (enabled by default):
  1. Check local config paths first (PHP tarballs)
  2. Use cached files if available
  3. Download from remote sources as fallback
  4. Auto-cache successful downloads for future reuse

Example Workflows:
  chauf install php 8.4        # First download, auto-cached
  chauf install php 8.4        # Instant - uses cached file
  chauf install nginx            # First download, auto-cached
  chauf install nginx            # Instant - uses cached file
  chauf install composer         # First download, auto-cached
  chauf install composer         # Instant - uses cached file
  chauf install --no-cache php   # Download without caching

Local PHP Installation:
  --local flag workflow:
  1. Check config for stored tarball path
  2. Prompt for tarball location if not configured
  3. Validate tarball matches requested version
  4. Fall back to internet download if desired

Cache Storage:
  Downloaded files are stored in ~/.chauffeur/cache/
  - PHP tarballs: php-8.4.14.tar.gz, php-8.3.27.tar.gz, etc.
  - Composer: composer.phar (versioned files)
  - Nginx: nginx-1.25.3.tar.gz, nginx-1.24.0.tar.gz, etc.

  Config is automatically updated with PHP tarball paths
  Future installations reuse cached files without downloading
  Use --no-cache to disable automatic caching

Composer Installation:
  chauf install composer          Download and install Composer with PHP isolation.
  chauf install composer --force   Reinstall Composer.`)
}
