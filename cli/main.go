package main

import (
	"fmt"
	"os"

	"github.com/siaji/chauffeur/cli/commands"
)

const version = "0.1.0"

func usage() {
	fmt.Print(`Chauffeur CLI

Usage:
  chauf --version        Print the current Chauffeur version.
  chauf version          Same as --version.
  chauf --help           Show this message.
  chauf init             Initialize Chauffeur workspace with default configuration.
  chauf install <service> [version]
                         Install Chauffeur-managed services (caddy, nginx, php).
  chauf remove <service> [version]
                         Remove installed Chauffeur-managed services.
  chauf php <command>    Manage PHP runtimes (use <version>, ...).
  chauf self-update      Update the Chauffeur CLI to the latest release.
  chauf nginx [args...]  Run the managed nginx binary with passthrough args.
  chauf caddy [args...]  Run the managed caddy binary with passthrough args.
  chauf start            Start Chauffeur services with chauf- prefix.
  chauf stop             Stop Chauffeur services with chauf- prefix.
  chauf status           Show status of Chauffeur services.
  chauf link             Register current directory as a project.
  chauf links            List all registered projects.
  chauf unlink           Unlink a registered project (by slug, domain, path, or all).
  chauf uninstall        Remove the Chauffeur workspace (keeps runtimes by default).
  chauf uninstall --purge
                         Remove the workspace and delete runtimes/caches.
`)
}

func main() {
	args := os.Args[1:]
	commands.SetCLIVersion(version)

	if len(args) == 0 {
		usage()
		return
	}

	switch args[0] {
	case "--version", "-V", "version":
		fmt.Printf("chauf %s\n", version)
	case "--help", "-h":
		usage()
	case "init":
		if err := commands.RunInit(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "install":
		if err := commands.RunInstall(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if err := commands.RunRemove(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "php":
		if err := commands.RunPHP(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "self-update":
		if err := commands.RunSelfUpdate(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "link":
		if err := commands.RunLink(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "info":
		if err := commands.RunInfo(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "links":
		if err := commands.RunLinks(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "unlink":
		if err := commands.RunUnlink(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "start":
		if err := commands.RunStart(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "stop":
		if err := commands.RunStop(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := commands.RunStatus(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "uninstall":
		if err := commands.RunUninstall(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		if commands.IsKnownService(args[0]) {
			if err := commands.RunServiceCommand(args[0], args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "Unsupported command: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Run 'chauf --help' for available commands.")
		os.Exit(1)
	}
}
