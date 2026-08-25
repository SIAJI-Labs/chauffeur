package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/siegg/chauffeur/internal/commands"
	"github.com/siegg/chauffeur/internal/lib"
)

// Injected at build time via -ldflags "-X main.version=x.y.z"
var version = "dev"

func main() {
	args := os.Args[1:]

	// Strip --verbose / -v (except when -v is the leading "version" command)
	// and enable verbose output globally before dispatch.
	args = extractVerboseFlag(args)

	if len(args) == 0 {
		printHelp()
		return
	}

	var err error

	switch args[0] {
	case "--help", "-h", "help":
		printHelp()

	case "--version", "-v", "version":
		fmt.Printf("chauf %s\n", version)

	case "commands":
		printCommands()

	// ── Workspace ──────────────────────────────────────────────────────────────

	case "init":
		err = commands.RunInit(args[1:])

	case "info":
		err = commands.RunInfo(args[1:])

	case "uninstall":
		err = commands.RunUninstall(args[1:])

	// ── Install & Remove ───────────────────────────────────────────────────────

	case "install":
		err = commands.RunInstall(args[1:])

	case "remove":
		err = commands.RunRemove(args[1:])

	// ── Projects ───────────────────────────────────────────────────────────────

	case "link":
		err = commands.RunLink(args[1:])

	case "links":
		err = commands.RunLinks(args[1:])

	case "unlink":
		err = commands.RunUnlink(args[1:])

	case "secure":
		err = commands.RunSecure(args[1:])

	case "unsecure":
		err = commands.RunUnsecure(args[1:])

	// ── PHP ────────────────────────────────────────────────────────────────────

	case "php":
		err = commands.RunPHP(args[1:])

	// ── Services ───────────────────────────────────────────────────────────────

	case "start":
		err = commands.RunStart(args[1:])

	case "stop":
		err = commands.RunStop(args[1:])

	case "restart":
		err = commands.RunRestart(args[1:])

	case "status":
		err = commands.RunStatus(args[1:])

	case "logs":
		err = commands.RunLogs(args[1:])

	// ── Config ─────────────────────────────────────────────────────────────────

	case "autostart":
		err = commands.RunAutostart(args[1:])

	case "config", "env":
		err = commands.RunConfig(args[1:])

	// ── Maintenance ────────────────────────────────────────────────────────────

	case "self-update":
		err = commands.RunSelfUpdate(args[1:], version)

	case "doctor":
		err = commands.RunDoctor(args[1:])

	case "update":
		err = commands.RunUpdate(args[1:])

	case "clean":
		err = commands.RunClean(args[1:])

	case "migrate":
		notImplemented(args[0])

	// ── Podman ────────────────────────────────────────────────────────────────

	case "podman":
		err = commands.RunPodman(args[1:])

	// ── Web UI ───────────────────────────────────────────────────────────────

	case "webui":
		err = commands.RunWebUI(args[1:])

	default:
		lib.Error(fmt.Sprintf("unknown command %q", args[0]))
		fmt.Println()
		printHelp()
		os.Exit(1)
	}

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0) // -h / --help: usage already printed, exit cleanly
		}
		lib.Error(err.Error())
		var exitCoder interface{ ExitCode() int }
		if errors.As(err, &exitCoder) && exitCoder.ExitCode() > 0 {
			os.Exit(exitCoder.ExitCode())
		}
		os.Exit(1)
	}
}

// extractVerboseFlag strips --verbose and -v (when not at position 0, where -v
// means the "version" command) from args, sets lib.Verbose, and returns the
// filtered slice. It preserves -v when it appears after what looks like a
// PHP version (e.g., "chauf php 7.4 -v") so the inline version feature works.
func extractVerboseFlag(args []string) []string {
	out := args[:0:0] // same backing array, zero length
	for i, a := range args {
		if a == "--verbose" || (a == "-v" && i > 0) {
			// Preserve -v if previous arg looks like a PHP version (e.g., "7.4", "8.3.0")
			// This allows "chauf php 7.4 -v" to pass -v to PHP
			if a == "-v" && i > 0 && looksLikePHPVersion(args[i-1]) {
				out = append(out, a)
				continue
			}
			lib.Verbose = true
		} else {
			out = append(out, a)
		}
	}
	return out
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

func notImplemented(cmd string) {
	lib.Info(fmt.Sprintf("%s  chauf %s is not yet implemented", lib.Gray("·"), cmd))
	fmt.Println()
	lib.Info(lib.Gray(`Run "chauf --help" to see all available commands.`))
}

func printHelp() {
	fmt.Printf("%s  %s\n", lib.Bold("chauf "+version), lib.Gray("Local PHP development environment for Linux"))
	fmt.Println()
	fmt.Printf("  Manages nginx, PHP-FPM (7.4–8.4), Composer, SSL, and DNS.\n")
	fmt.Printf("  All services run in %s — no root required.\n", lib.Cyan("~/.chauffeur/"))
	fmt.Println()
	fmt.Printf("%s\n", lib.Bold("Usage:"))
	fmt.Printf("  chauf <command> [flags]\n")
	fmt.Printf("  chauf <command> --help\n")
	fmt.Println()
	fmt.Printf("%s\n", lib.Bold("Commands:"))
	printCommands()
	fmt.Println()
	fmt.Printf("%s\n", lib.Gray(`Run "chauf <command> --help" for detailed usage.`))
}

func printCommands() {
	type entry struct {
		cmd  string
		desc string
	}

	sections := []struct {
		title   string
		entries []entry
	}{
		{
			"Workspace",
			[]entry{
				{"init", "Initialize the Chauffeur workspace"},
				{"info", "Show workspace status and installed services"},
				{"uninstall", "Remove the workspace"},
			},
		},
		{
			"Install & Remove",
			[]entry{
				{"install", "Install nginx, PHP, or Composer from source"},
				{"remove", "Remove an installed service"},
			},
		},
		{
			"Projects",
			[]entry{
				{"link", "Register a project and generate nginx config"},
				{"links", "List registered projects"},
				{"unlink", "Unregister a project"},
				{"secure", "Enable HTTPS (mkcert certificate)"},
				{"unsecure", "Disable HTTPS"},
			},
		},
		{
			"PHP",
			[]entry{
				{"php list", "List installed PHP versions"},
				{"php use", "Set the global default PHP version"},
				{"php isolate", "Pin a project to a specific PHP version"},
				{"php install", "Alias for: chauf install php <version>"},
				{"php remove", "Alias for: chauf remove php <version>"},
			},
		},
		{
			"Services",
			[]entry{
				{"start", "Start nginx and PHP-FPM"},
				{"stop", "Stop services"},
				{"restart", "Reload services (zero-downtime for nginx)"},
				{"status", "Show service health, PID, uptime, memory"},
				{"logs", "View nginx or PHP-FPM logs"},
			},
		},
		{
			"Configuration",
			[]entry{
				{"autostart", "Manage systemd user services for auto-start on login"},
				{"config", "Read and write workspace or project config"},
				{"env", "Manage per-project environment variables"},
			},
		},
		{
			"Maintenance",
			[]entry{
				{"doctor", "Validate environment and check dependencies"},
				{"clean", "Remove cached downloads, logs, stale certs"},
				{"migrate", "Move a project between workspaces"},
				{"self-update", "Update the chauf binary"},
			},
		},
		{
			"Podman",
			[]entry{
				{"podman create", "Create a database container (mysql|postgres|maria|mongo|redis)"},
				{"podman start", "Start a container (or 'all')"},
				{"podman stop", "Stop a container (or 'all')"},
				{"podman status", "Show container status"},
				{"podman list", "List all managed containers"},
				{"podman remove", "Remove a container"},
				{"podman console", "Attach to container for CLI access"},
			},
		},
		{
			"Web UI",
			[]entry{
				{"webui start", "Start the local web UI"},
				{"webui status", "Show web UI status"},
				{"webui stop", "Stop the local web UI"},
			},
		},
	}

	for _, section := range sections {
		fmt.Printf("\n  %s\n", lib.Bold(section.title))
		for _, e := range section.entries {
			fmt.Printf("    %-16s  %s\n", e.cmd, lib.Gray(e.desc))
		}
	}
}
