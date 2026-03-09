package commands

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunPHP(args []string) error {
	if len(args) == 0 {
		return phpHelp()
	}

	switch strings.ToLower(args[0]) {
	case "list", "ls":
		return phpList(args[1:])
	case "use":
		return phpUse(args[1:])
	case "install":
		// Alias: chauf php install 8.3 → chauf install php 8.3
		return RunInstall(append([]string{"php"}, args[1:]...))
	case "remove":
		// Alias: chauf php remove 8.3 → chauf remove php 8.3
		return RunRemove(append([]string{"php"}, args[1:]...))
	case "isolate":
		return phpIsolate(args[1:])
	case "--help", "-h", "help":
		return phpHelp()
	default:
		return fmt.Errorf("unknown php subcommand %q — run: chauf php --help", args[0])
	}
}

// ── php list ───────────────────────────────────────────────────────────────────

func phpList(args []string) error {
	flags := flag.NewFlagSet("php list", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()
	installed := installers.ListInstalledPHP(root)

	fmt.Println()
	if len(installed) == 0 {
		lib.Info("No PHP versions installed.")
		fmt.Println()
		lib.Info(lib.Gray("Install one with:  chauf install php 8.3"))
		fmt.Println()
		return nil
	}

	// Sort versions descending
	sort.Slice(installed, func(i, j int) bool { return installed[i] > installed[j] })

	lib.Pair("Default", cfg.PHP.DefaultVersion)
	fmt.Println()

	for _, mm := range installed {
		inst, _ := installers.NewPHPInstaller(mm, installers.BuildOpts{})
		ver := inst.InstalledVersion()
		marker := "  "
		label := mm
		if mm == cfg.PHP.DefaultVersion {
			marker = lib.Green("✓") + " "
			label = lib.Bold(mm)
		}
		detail := lib.Gray(ver)
		fmt.Printf("  %s %-6s  %s\n", marker, label, detail)
	}

	fmt.Println()
	lib.Info(lib.Gray("Switch default:  chauf php use <version>"))
	fmt.Println()
	return nil
}

// ── php use ────────────────────────────────────────────────────────────────────

func phpUse(args []string) error {
	flags := flag.NewFlagSet("php use", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}

	version := flags.Arg(0)
	if version == "" {
		return fmt.Errorf("usage: chauf php use <version>  (e.g. 8.3)")
	}

	mm := installers.MajorMinor(version)
	if mm == "" {
		return fmt.Errorf("invalid PHP version: %q", version)
	}

	root := workspace.Root()
	inst, _ := installers.NewPHPInstaller(mm, installers.BuildOpts{})
	if !inst.IsInstalled() {
		lib.Warn(fmt.Sprintf("PHP %s is not installed.", mm))
		lib.Info(lib.Gray("Install it with:  chauf install php " + mm))
		return nil
	}

	if err := workspace.SetDefaultPHP(mm); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Default PHP set to %s  (%s)", mm, inst.InstalledVersion()))
	lib.Info(lib.Gray("The PHP shim will now use PHP " + mm + " for all projects."))

	// If php-fpm for this version is not running, remind the user to start it.
	pidFile := root + "/php/" + mm + "/runtime/php-fpm/php-fpm.pid"
	if pid := readPID(pidFile); pid == 0 || !processRunning(pid) {
		fmt.Println()
		lib.Info(lib.Gray("Start services:  chauf start"))
	}

	fmt.Println()
	return nil
}

// ── php isolate ────────────────────────────────────────────────────────────────

func phpIsolate(args []string) error {
	flags := flag.NewFlagSet("php isolate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	projectPath := flags.String("project", "", "Path to the project (defaults to current directory)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	version := flags.Arg(0)
	if version == "" {
		return fmt.Errorf("usage: chauf php isolate <version>  (e.g. 8.1)")
	}

	mm := installers.MajorMinor(version)
	if mm == "" {
		return fmt.Errorf("invalid PHP version: %q", version)
	}

	target := *projectPath
	if target == "" {
		var err error
		target, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Write .chauffeur-php file in the project dir (for the PHP shim)
	phpFile := target + "/.chauffeur-php"
	if err := os.WriteFile(phpFile, []byte(mm+"\n"), 0644); err != nil {
		return fmt.Errorf("write .chauffeur-php: %w", err)
	}

	// If this directory is a registered project, update its config and nginx.
	root := workspace.Root()
	cfg := workspace.Load()
	prevVersion := ""
	p, err := projects.FindByPath(root, target)
	if err == nil && p != nil {
		prevVersion = p.PHPVersion
		p.PHPVersion = mm
		if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
			lib.Warn("Could not regenerate nginx config: " + err.Error())
		}
		if err := projects.Save(p, root); err != nil {
			lib.Warn("Could not update project config: " + err.Error())
		}
		if projects.IsNginxRunning(root) {
			_ = projects.ReloadNginx(root)
		}
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Project pinned to PHP %s", mm))
	lib.Pair("File", phpFile)
	if p != nil && prevVersion != "" && prevVersion != mm {
		lib.Pair("Was", prevVersion)
		lib.Info(lib.Gray("nginx config and project config updated."))
	}
	lib.Info(lib.Gray("The PHP shim reads CHAUFFEUR_PHP_VERSION or .chauffeur-php when present."))
	fmt.Println()

	// Remind user to add to .gitignore if not already there
	gitignore := target + "/.gitignore"
	if data, err := os.ReadFile(gitignore); err == nil {
		if !strings.Contains(string(data), ".chauffeur-php") {
			lib.Info(lib.Gray("Add to .gitignore:  .chauffeur-php"))
		}
	}

	fmt.Println()
	return nil
}

// ── help ───────────────────────────────────────────────────────────────────────

func phpHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf php — PHP version management"))
	fmt.Printf("  %-22s  %s\n", "php list", lib.Gray("List installed PHP versions"))
	fmt.Printf("  %-22s  %s\n", "php use <version>", lib.Gray("Set the global default PHP version"))
	fmt.Printf("  %-22s  %s\n", "php isolate <version>", lib.Gray("Pin the current project to a PHP version"))
	fmt.Printf("  %-22s  %s\n", "php install <version>", lib.Gray("Alias for: chauf install php <version>"))
	fmt.Printf("  %-22s  %s\n", "php remove <version>", lib.Gray("Alias for: chauf remove php <version>"))
	fmt.Println()
	return nil
}
