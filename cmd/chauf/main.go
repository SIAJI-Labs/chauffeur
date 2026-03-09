package main

import (
	"fmt"
	"os"

	"github.com/siegg/chauffeur/internal/commands"
	"github.com/siegg/chauffeur/internal/lib"
)

// Injected at build time via -ldflags "-X main.version=x.y.z"
var version = "dev"

func main() {
	args := os.Args[1:]

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

	case "start", "stop", "restart", "status", "logs":
		notImplemented(args[0])

	// ── Config ─────────────────────────────────────────────────────────────────

	case "config", "env", "autostart":
		notImplemented(args[0])

	// ── Maintenance ────────────────────────────────────────────────────────────

	case "self-update":
		err = commands.RunSelfUpdate(args[1:], version)

	case "doctor", "clean", "migrate":
		notImplemented(args[0])

	default:
		lib.Error(fmt.Sprintf("unknown command %q", args[0]))
		fmt.Println()
		printHelp()
		os.Exit(1)
	}

	if err != nil {
		lib.Error(err.Error())
		os.Exit(1)
	}
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
				{"config", "Read and write workspace or project config"},
				{"env", "Manage per-project environment variables"},
				{"autostart", "Manage systemd auto-start services"},
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
	}

	for _, section := range sections {
		fmt.Printf("\n  %s\n", lib.Bold(section.title))
		for _, e := range section.entries {
			fmt.Printf("    %-16s  %s\n", e.cmd, lib.Gray(e.desc))
		}
	}
}
