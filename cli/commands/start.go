package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/internal/workspace"
)

/**
 * RunStart checks prerequisites and (eventually) starts Chauffeur services.
 *
 * @param args CLI arguments passed after the start subcommand.
 * @return error when prerequisite checks or prompted installs fail.
 */
func RunStart(args []string) error {
	var dryRun bool

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--help", "-h":
			printStartUsage()
			return nil
		default:
			return fmt.Errorf("unknown flag for start: %s", arg)
		}
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

	var missing []serviceSpec
	for _, name := range serviceNames {
		spec, err := newServiceSpec(name, prefix, info)
		if err != nil {
			return err
		}

		ok, err := spec.available()
		if err != nil {
			return err
		}
		if !ok {
			missing = append(missing, spec)
		}
	}

	if len(missing) == 0 {
		fmt.Println("All Chauffeur services are installed. Service bootstrap coming soon.")
		return nil
	}

	if dryRun {
		fmt.Println("Missing services:")
		for _, svc := range missing {
			fmt.Printf("  - %s (%s)\n", svc.name, svc.description)
		}
		fmt.Println("Run `chauf install <service>` to install the required components.")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	for _, svc := range missing {
		fmt.Printf("%s is not installed (%s).\n", svc.name, svc.description)
		fmt.Printf("Run `chauf install %s` now? [y/N]: ", svc.name)
		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read response: %w", err)
		}

		response := strings.ToLower(strings.TrimSpace(input))
		if response != "y" && response != "yes" {
			fmt.Printf("Skipping %s installation. Run `chauf install %s` later.\n", svc.name, svc.name)
			continue
		}

		fmt.Printf("Invoking installer for %s...\n", svc.name)
		if err := svc.install(false); err != nil {
			return fmt.Errorf("install %s: %w", svc.name, err)
		}
		fmt.Printf("Installed %s successfully.\n", svc.name)
	}

	fmt.Println("All required services installed. Service bootstrap coming soon.")
	return nil
}

/**
 * printStartUsage renders CLI help for the start command.
 */
func printStartUsage() {
	fmt.Println(`Usage: chauf start [--dry-run]

Ensures Chauffeur-managed services (caddy, nginx) are available.

Options:
  --dry-run  Show missing services without taking action.
  -h, --help Show this message.`)
}
