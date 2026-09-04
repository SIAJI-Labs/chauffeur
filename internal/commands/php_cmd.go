package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/workspace"
)

// isInstalledPHPVersion checks if a version string matches an installed PHP.
func isInstalledPHPVersion(version string) bool {
	if mm := installers.MajorMinor(version); mm != "" {
		version = mm
	}
	root := workspace.Root()
	if workspace.Load().Runtime.Engine == string(runtime.EnginePodman) {
		result, err := (runtime.ExecRunner{}).Run(context.Background(), "image", "exists", runtime.PHPImage(version))
		return err == nil && result.ExitCode == 0
	}
	installed := installers.ListInstalledPHP(root)
	for _, v := range installed {
		if v == version {
			return true
		}
	}
	return false
}

// looksLikePHPVersion returns true if s looks like a PHP version (e.g., "7.4", "8.3.0", "8.4")
func looksLikePHPVersion(s string) bool {
	if len(s) == 0 {
		return false
	}
	dots := 0
	for _, c := range s {
		if c == '.' {
			dots++
			if dots > 2 {
				return false
			}
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// runPHPExtended executes a command with a specific PHP version.
// It sets CHAUFFEUR_PHP_VERSION and routes through the composer shim
// (for composer commands) or directly via PHP binary.
func runPHPExtended(version string, args []string) error {
	root := workspace.Root()
	if workspace.Load().Runtime.Engine == string(runtime.EnginePodman) {
		return runPHPInPodman(version, args)
	}

	// Verify PHP is installed
	phpBin := filepath.Join(root, "php", version, "bin", "php")
	if _, err := os.Stat(phpBin); err != nil {
		return fmt.Errorf("PHP %s not installed. Run: chauf install php %s", version, version)
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: chauf php <version> <command> [args...]")
	}

	cmd := args[0]

	// Set env for shim
	env := os.Environ()
	env = append(env, fmt.Sprintf("CHAUFFEUR_PHP_VERSION=%s", version))

	shimPath := filepath.Join(root, "bin", "shims")

	var execPath string
	var execArgs []string

	// Route based on command
	// Note: execArgs[0] must be the program name for syscall.Exec
	switch cmd {
	case "composer":
		// Use composer shim - argv[0] = "composer" is passed to shim
		execPath = filepath.Join(shimPath, "composer")
		execArgs = append([]string{"composer"}, args[1:]...)
	case "php-fpm":
		// Direct php-fpm binary
		execPath = phpBin + "-fpm"
		execArgs = append([]string{"php-fpm"}, args[1:]...)
	case "php":
		// Direct PHP binary - use "php" as argv[0] so flags work correctly
		execPath = phpBin
		execArgs = append([]string{"php"}, args[1:]...)
	default:
		// Treat as PHP script (e.g., artisan, phpunit, wp-cli)
		// argv[0] = "php" so PHP doesn't interpret cmd as a script filename
		execPath = phpBin
		execArgs = append([]string{"php"}, args...)
	}

	// Use syscall.Exec to replace current process
	return syscall.Exec(execPath, execArgs, env)
}

func runPHPInPodman(version string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Usage: chauf php <version> <command> [args...]")
	}
	command := args
	switch args[0] {
	case "composer":
		command = append([]string{"composer"}, args[1:]...)
	case "php":
		command = append([]string{"php"}, args[1:]...)
	case "php-fpm":
		command = append([]string{"php-fpm"}, args[1:]...)
	default:
		command = append([]string{"php"}, args...)
	}
	ctx := context.Background()
	executor, err := runtime.ForWorkspace(workspace.Load())
	if err != nil {
		return err
	}
	cwd := mustWorkingDirectory()
	scope := runtime.Scope{Workspace: workspace.Root(), Version: version}
	workdir := "/workspace"
	if project, findErr := projects.FindByPath(workspace.Root(), cwd); findErr == nil && project != nil {
		scope.Project = project.Path
		scope.Slug = project.Slug
		scope.Dedicated = project.FPM.Dedicated
		workdir = "/workspace/" + project.Slug
	}
	databaseEnv := podmanDatabaseEnv(scope.Project)

	// A workspace start already mounts all linked projects into the shared
	// container. Reuse that container instead of trying to add the CLI cwd as a
	// conflicting single-project mount.
	statuses, statusErr := executor.Status(ctx, scope)
	if statusErr != nil || len(statuses) == 0 || statuses[0].State != "running" {
		if err := executor.EnsureProject(ctx, scope, runtime.PHPImage(version), map[string]string{"/workspace": cwd}); err != nil {
			return err
		}
	}
	result, err := executor.Exec(ctx, scope, command, runtime.ExecOptions{
		Dir:    workdir,
		Env:    databaseEnv,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		if result.ExitCode >= 0 {
			return fmt.Errorf("runtime command exited with status %d: %w", result.ExitCode, err)
		}
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("runtime command exited with status %d", result.ExitCode)
	}
	return nil
}

func podmanDatabaseEnv(projectRoot string) []string {
	if projectRoot == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, ".env"))
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "DB_HOST=") {
			continue
		}
		host := strings.Trim(strings.TrimPrefix(line, "DB_HOST="), `"'`)
		if host == "127.0.0.1" || host == "localhost" {
			return []string{"DB_HOST=host.containers.internal"}
		}
		break
	}
	return nil
}

func mustWorkingDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func RunPHP(args []string) error {
	if len(args) == 0 {
		return phpHelp()
	}

	// Inline version mode: chauf php <version> <command> [args...]
	// Check if first arg is an installed PHP version
	if isInstalledPHPVersion(args[0]) {
		if len(args) == 1 {
			return fmt.Errorf("Usage: chauf php <version> <command> [args...]")
		}
		return runPHPExtended(installers.MajorMinor(args[0]), args[1:])
	}

	// If first arg looks like a PHP version but isn't installed, show specific error
	if looksLikePHPVersion(args[0]) {
		return fmt.Errorf("PHP %s not installed. Run: chauf install php %s", args[0], args[0])
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
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf php list — list installed PHP versions", "chauf php list")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()
	installed := make([]string, 0)
	if cfg.Runtime.Engine == string(runtime.EnginePodman) {
		for _, version := range installers.SupportedPHPVersions {
			if isInstalledPHPVersion(version) {
				installed = append(installed, version)
			}
		}
	} else {
		installed = installers.ListInstalledPHP(root)
	}

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
		ver := "Podman image"
		if cfg.Runtime.Engine != string(runtime.EnginePodman) {
			inst, _ := installers.NewPHPInstaller(mm, installers.BuildOpts{})
			ver = inst.InstalledVersion()
		}
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
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf php use — set the global default PHP version", "chauf php use <version>")
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
	if workspace.Load().Runtime.Engine == string(runtime.EnginePodman) && !isInstalledPHPVersion(mm) {
		return fmt.Errorf("PHP %s Podman image is unavailable — run: chauf install php %s", mm, mm)
	}

	root := workspace.Root()
	cfg := workspace.Load()
	if !isInstalledPHPVersion(mm) {
		lib.Warn(fmt.Sprintf("PHP %s is not installed for the selected runtime.", mm))
		lib.Info(lib.Gray("Install it with:  chauf install php " + mm))
		return nil
	}
	installedVersion := "Podman image"
	if cfg.Runtime.Engine != string(runtime.EnginePodman) {
		inst, _ := installers.NewPHPInstaller(mm, installers.BuildOpts{})
		installedVersion = inst.InstalledVersion()
	}

	if err := workspace.SetDefaultPHP(mm); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Default PHP set to %s  (%s)", mm, installedVersion))
	lib.Info(lib.Gray("The selected runtime will now use PHP " + mm + " for all projects."))

	// If php-fpm for this version is not running, remind the user to start it.
	pidFile := root + "/php/" + mm + "/runtime/php-fpm/php-fpm.pid"
	if cfg.Runtime.Engine != string(runtime.EnginePodman) && (readPID(pidFile) == 0 || !processRunning(readPID(pidFile))) {
		fmt.Println()
		lib.Info(lib.Gray("Start services:  chauf start"))
	}

	fmt.Println()
	return nil
}

// ── php isolate ────────────────────────────────────────────────────────────────

func phpIsolate(args []string) error {
	flags := flag.NewFlagSet("php isolate", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf php isolate — pin a project to a specific PHP version", "chauf php isolate <version> [--project <path>]")
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

	// Update project config and nginx — PHP version is stored in ~/.chauffeur/projects/<slug>/config.yaml.
	// No dotfile is written to the project directory.
	root := workspace.Root()
	cfg := workspace.Load()
	prevVersion := ""
	p, err := projects.FindByPath(root, target)
	if err == nil && p != nil {
		prevVersion = p.PHPVersion
		p.PHPVersion = mm
		if cfg.Runtime.Engine == string(runtime.EnginePodman) {
			if _, applyErr := applyLinkProject(p, root, cfg); applyErr != nil {
				return applyErr
			}
		} else {
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
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Project pinned to PHP %s", mm))
	if p != nil {
		lib.Pair("Config", "~/.chauffeur/projects/"+p.Slug+"/config.yaml")
		if prevVersion != "" && prevVersion != mm {
			lib.Pair("Was", prevVersion)
			lib.Info(lib.Gray("nginx config updated and reloaded."))
		}
	} else {
		lib.Info(lib.Gray("Project not linked — pin will take effect once you run: chauf link"))
	}
	lib.Info(lib.Gray("The PHP shim resolves the version from your project config automatically."))
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
