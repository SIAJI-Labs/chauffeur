package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunInstall handles `chauf install <service...>` invocations.
 *
 * @param args CLI arguments passed after the install subcommand.
 * @return error when parsing fails or an installation step errors.
 */
func RunInstall(args []string) error {
	if len(args) == 0 {
		printInstallUsage()
		return errors.New("no services specified")
	}

	var (
		force    bool
		services []string
	)

	for _, arg := range args {
		switch arg {
		case "--force":
			force = true
		case "--help", "-h":
			printInstallUsage()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag for install: %s", arg)
			}
			services = append(services, arg)
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
		spec, err := newServiceSpec(name, prefix, info)
		if err != nil {
			return err
		}

		ok, err := spec.available()
		if err != nil {
			return err
		}
		if ok && !force {
			fmt.Printf("%s already installed at %s\n", spec.name, spec.binaryPath)
			continue
		}

		fmt.Printf("Installing %s (%s)...\n", spec.name, spec.description)
		if err := spec.install(force); err != nil {
			return fmt.Errorf("install %s: %w", spec.name, err)
		}
		fmt.Printf("Installed %s successfully.\n", spec.name)
	}

	return nil
}

/**
 * printInstallUsage renders CLI help for the install command.
 */
func printInstallUsage() {
	fmt.Println(`Usage: chauf install [--force] <service> [<service>...]

Installs one or more Chauffeur-managed services.

Options:
  --force    Reinstall even if the service is already present.
  -h, --help Show this message.

Services:
  caddy      Verified tarball from GitHub releases.
  nginx      Source build from the latest GitHub release.`)
}
