package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunRemove(args []string) error {
	flags := flag.NewFlagSet("remove", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf remove — remove an installed service", "chauf remove <nginx|php <version>|composer> [--force]")
	force := flags.Bool("force", false, "Skip confirmation prompt")

	flagArgs, positionals := splitArgs(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}

	if len(positionals) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: chauf remove <nginx|php <version>|composer>\n")
		return fmt.Errorf("no service specified")
	}

	root := workspace.Root()

	switch strings.ToLower(positionals[0]) {
	case "nginx":
		return removeNginx(root, *force)
	case "php":
		version := ""
		if len(positionals) > 1 {
			version = positionals[1]
		}
		if version == "" {
			return fmt.Errorf("usage: chauf remove php <version>  (e.g. 8.3)")
		}
		mm := installers.MajorMinor(version)
		if mm == "" {
			return fmt.Errorf("invalid PHP version: %q", version)
		}
		return removePHP(root, mm, *force)
	case "composer":
		return removeComposer(root, *force)
	default:
		return fmt.Errorf("unknown service %q — available: nginx, php, composer", positionals[0])
	}
}

func removeNginx(root string, force bool) error {
	nginxBin := filepath.Join(root, "nginx", "sbin", "nginx")
	if _, err := os.Stat(nginxBin); os.IsNotExist(err) {
		lib.Info("nginx is not installed.")
		return nil
	}

	fmt.Println()
	lib.Pair("Removing", "nginx binary (config/certs/logs are preserved)")
	lib.Pair("Binary", nginxBin)

	if !force {
		if !confirm("Remove nginx binary?") {
			fmt.Println("\n  Aborted.")
			return nil
		}
	}
	fmt.Println()

	if err := os.RemoveAll(filepath.Join(root, "nginx", "sbin")); err != nil {
		return fmt.Errorf("failed to remove nginx: %w", err)
	}
	// Also remove the modules directory if present
	_ = os.RemoveAll(filepath.Join(root, "nginx", "modules"))

	lib.Success("nginx removed")
	lib.Info(lib.Gray("Config, logs, and certs are preserved in " + filepath.Join(root, "nginx")))
	fmt.Println()
	return nil
}

func removePHP(root, mm string, force bool) error {
	phpDir := filepath.Join(root, "php", mm)
	if _, err := os.Stat(phpDir); os.IsNotExist(err) {
		lib.Info(fmt.Sprintf("PHP %s is not installed.", mm))
		return nil
	}

	fmt.Println()
	lib.Pair("Removing", "PHP "+mm)
	lib.Pair("Directory", phpDir)
	lib.Info(lib.Gray("This will permanently delete " + phpDir + " and all its contents."))

	if !force {
		if !confirm("Remove PHP " + mm + "?") {
			fmt.Println("\n  Aborted.")
			return nil
		}
	}
	fmt.Println()

	// Stop php-fpm if running
	pidFile := filepath.Join(root, "php", mm, "runtime", "php-fpm", "php-fpm.pid")
	if pid := readPID(pidFile); pid > 0 && processRunning(pid) {
		sendSignal(pid, syscall.SIGTERM)
		lib.Success("php-fpm " + mm + " stopped")
	}

	if err := os.RemoveAll(phpDir); err != nil {
		return fmt.Errorf("failed to remove PHP %s: %w", mm, err)
	}
	lib.Success("PHP " + mm + " removed")
	fmt.Println()
	return nil
}

func removeComposer(root string, force bool) error {
	pharPath := filepath.Join(root, "composer", "composer.phar")
	if _, err := os.Stat(pharPath); os.IsNotExist(err) {
		lib.Info("Composer is not installed.")
		return nil
	}

	fmt.Println()
	lib.Pair("Removing", "composer.phar")
	lib.Pair("Path", pharPath)

	if !force {
		if !confirm("Remove Composer?") {
			fmt.Println("\n  Aborted.")
			return nil
		}
	}
	fmt.Println()

	if err := os.Remove(pharPath); err != nil {
		return fmt.Errorf("failed to remove composer.phar: %w", err)
	}
	lib.Success("Composer removed")
	fmt.Println()
	return nil
}

func confirm(prompt string) bool {
	return tui.Confirm(prompt)
}
