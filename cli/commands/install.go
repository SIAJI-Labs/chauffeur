package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * RunInstall handles `chauf install <service...>` invocations.
 *
 * @param args CLI arguments passed after the install subcommand.
 * @return error when parsing fails or an installation step errors.
 */
func RunInstall(args []string) error {
	logger := lib.NewCommandLogger("install")

	if len(args) == 0 {
		printInstallUsage()
		return errors.New("no services specified")
	}

	var (
		force    bool
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
			if err := runPHPInstall(version, force); err != nil {
				return err
			}
		} else {
			// Handle other services (nginx, caddy)
			spec, err := newServiceSpec(name, prefix, info)
			if err != nil {
				return err
			}

			ok, err := spec.available()
			if err != nil {
				return err
			}
			if ok && !force {
				logger.Info(fmt.Sprintf("%s already installed at %s", spec.name, spec.binaryPath))
				continue
			}

			logger.Info(fmt.Sprintf("Installing %s (%s)...", spec.name, spec.description))
			if err := spec.install(force); err != nil {
				return logger.Error(fmt.Sprintf("install %s", spec.name), err.Error())
			}
			logger.Success(fmt.Sprintf("Installed %s successfully", spec.name), "")
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
func runPHPInstall(version string, force bool) error {
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

	logger.Info(fmt.Sprintf("Installing PHP %s...", version))
	if err := installers.InstallPHPSource(version, installers.InstallOptions{
		Prefix: prefix,
		Force:  force,
		Info:   info,
	}); err != nil {
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
		fmt.Printf("  %d) PHP %s%s\n", i+1, versionMeta.Version, status)
	}

	fmt.Printf("Enter your choice (1-%d): ", len(versions))

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
 * printInstallUsage renders CLI help for the install command.
 */
func printInstallUsage() {
	fmt.Println(`Usage: chauf install [--force] <service> [<version>...] [<service>...]

Installs one or more Chauffeur-managed services.

Options:
  --force    Reinstall even if the service is already present.
  -h, --help Show this message.

Services:
  caddy      Verified tarball from GitHub releases.
  nginx      Source build from the latest GitHub release.
  php        Source build with development extensions.
  
PHP Installation:
  chauf install php           Interactive version selection.
  chauf install php 8.3      Install specific version.
  chauf install php --force   Reinstall selected version.`)
}
