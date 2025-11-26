package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/siaji/chauffeur/internal/utils"
)

// Version will be fetched dynamically from GitHub
var version = ""

// build metadata overridden via -ldflags "-X main.buildTimestamp=... -X main.buildCommit=..."
var (
	buildTimestamp = "unknown"
	buildCommit    = "unknown"
	buildDir       = "" // Set during self-update to track which repo we're building from
)

func usage() {
	fmt.Print(`Chauffeur CLI

Usage:
  chauf --version        Print the current Chauffeur version.
  chauf version          Same as --version.
  chauf --help           Show this message.
  chauf init             Initialize Chauffeur workspace with default configuration.
  chauf install <service> [version]
                         Install Chauffeur-managed services (nginx, php, composer).
  chauf remove <service> [version]
                         Remove installed Chauffeur-managed services.
  chauf php <command>    Manage PHP runtimes (use <version>, isolate <version>, current, ...).
  chauf self-update      Update the Chauffeur CLI to the latest release.
  chauf nginx [args...]  Run the managed nginx binary with passthrough args.
  chauf start            Start Chauffeur services with chauf- prefix.
  chauf stop             Stop Chauffeur services with chauf- prefix.
  chauf restart          Restart Chauffeur services with chauf- prefix.
  chauf status           Show status of Chauffeur services.
  chauf logs [service]   View and follow logs from Chauffeur services.
  chauf clean [target]   Clean up workspace files and free up disk space.
  chauf migrate <project> <workspace>  Migrate a project to a different workspace.
  chauf hello-world      Print a friendly greeting message.
  chauf link             Register current directory as a project.
  chauf links            List all registered projects.
  chauf unlink           Unlink a registered project (by slug, domain, path, or all).
  chauf secure           Add SSL certificate to current linked project.
  chauf unsecure         Remove SSL certificate from current linked project.
  chauf doctor           Perform health checks and diagnose system issues.
  chauf uninstall        Remove the Chauffeur workspace (keeps runtimes by default).
  chauf uninstall --purge
                         Remove the workspace and delete runtimes/caches.
`)
}

// getVersion fetches the current version, with fallback to build metadata
func getVersion() string {
	// First priority: Try to get version from git repository
	if gitVersion, err := getGitVersion(); err == nil && gitVersion != "" {
		// For production builds from main branch, use latest release version
		if strings.Contains(gitVersion, "main-") {
			if latestVersion, err := utils.GetLatestVersion(); err == nil {
				return strings.TrimPrefix(latestVersion, "v")
			}
			// Fallback to main branch format if API fails
			return gitVersion
		}
		// For other branches (develop, feature branches), use branch-commit format
		return gitVersion
	}

	// Second priority: Try version from build flags (if available)
	if version != "" {
		return version
	}

	// Third priority: try to get latest release from GitHub
	if latestVersion, err := utils.GetLatestVersion(); err == nil {
		return strings.TrimPrefix(latestVersion, "v")
	}

	// Ultimate fallback
	return "unknown"
}

// getGitVersion gets version from git repository, with fallback to buildDir
func getGitVersion() (string, error) {
	// Use buildDir if available (set during self-update)
	checkDir := "."
	if buildDir != "" {
		checkDir = buildDir
	}

	// Check if we're in a git repository
	if _, err := os.Stat(filepath.Join(checkDir, ".git")); os.IsNotExist(err) {
		return "", fmt.Errorf("not a git repository")
	}

	// Get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = checkDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(output))

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = checkDir
	output, err = cmd.Output()
	if err != nil {
		return "", err
	}
	commit := strings.TrimSpace(string(output))

	// Return version based on branch
	if commit != "" {
		if len(commit) >= 7 {
			commit = commit[:7]
		}
		return fmt.Sprintf("%s-%s", branch, commit), nil
	}

	return branch, nil
}

func main() {
	// Initialize version dynamically
	dynamicVersion := getVersion()

	args := os.Args[1:]
	commands.SetCLIVersion(dynamicVersion)
	commands.SetBuildTimestamp(buildTimestamp)
	commands.SetBuildCommit(buildCommit)

	if len(args) == 0 {
		usage()
		return
	}

	switch args[0] {
	case "--version", "-v", "-V", "version":
		fmt.Printf("chauf %s (built %s, commit %s)\n", dynamicVersion, buildTimestamp, buildCommit)
	case "--help", "-h", "help":
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
	case "doctor":
		if err := commands.RunDoctor(args[1:]); err != nil {
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
	case "secure":
		if err := commands.RunSecure(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "unsecure":
		if err := commands.RunUnsecure(args[1:]); err != nil {
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
	case "restart":
		if err := commands.RunRestart(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := commands.RunStatus(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "logs":
		if err := commands.RunLogs(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "clean":
		if err := commands.RunClean(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "migrate":
		if err := commands.RunMigrate(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "hello-world":
		if err := commands.RunHelloWorld(args[1:]); err != nil {
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
