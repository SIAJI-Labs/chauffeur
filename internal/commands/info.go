package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunInfo(args []string) error {
	flags := flag.NewFlagSet("info", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf info — show workspace status", "chauf info [--detail]")
	detail := flags.Bool("detail", false, "Show full paths, config files, socket paths")

	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()

	home, _ := os.UserHomeDir()
	short := func(p string) string {
		return strings.Replace(p, home, "~", 1)
	}

	// ── Header ─────────────────────────────────────────────────────────────────

	fmt.Printf("  %s  %s\n", lib.Bold("Chauffeur"), lib.Gray("v"+cfg.Version))
	fmt.Println()
	lib.Pair("Workspace", short(root))
	lib.Pair("Config", short(filepath.Join(root, "config", "chauffeur.yaml")))
	printRuntime(cfg)
	fmt.Println()

	// ── Services ───────────────────────────────────────────────────────────────

	fmt.Printf("  %s\n", lib.Bold("Services"))
	phpVersions := installedPHPVersions(root)
	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		if err := showPodmanStatus(root, *detail, ""); err != nil {
			return err
		}
	} else {

		// nginx
		nginxPID := filepath.Join(root, "nginx", "logs", "nginx.pid")
		nginxRunning := pidFileRunning(nginxPID)
		nginxStatus := serviceStatus(nginxRunning)
		if *detail {
			lib.Pair("  nginx", fmt.Sprintf("%s    %d / %d    %s",
				nginxStatus, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort,
				lib.Gray(filepath.Join(root, "nginx", "sbin", "nginx"))))
		} else {
			lib.Pair("  nginx", fmt.Sprintf("%s    %d / %d",
				nginxStatus, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort))
		}

		// PHP-FPM per installed version
		usedPHPVersions := phpVersionsInUse(root)
		for _, ver := range phpVersions {
			fpmPID := filepath.Join(root, "php", ver, "runtime", "php-fpm", "php-fpm.pid")
			running := pidFileRunning(fpmPID)
			status := serviceStatus(running)
			label := fmt.Sprintf("  php-fpm %s", ver)

			var tags []string
			if !usedPHPVersions[ver] {
				tags = append(tags, lib.Gray("(no projects)"))
			}
			if ver == cfg.PHP.DefaultVersion {
				tags = append(tags, lib.Gray("(default)"))
			}
			suffix := ""
			if len(tags) > 0 {
				suffix = "  " + strings.Join(tags, " ")
			}

			if *detail {
				lib.Pair(label, fmt.Sprintf("%s    shared pool%s    %s",
					status, suffix,
					lib.Gray(filepath.Join(root, "php", ver, "sbin", "php-fpm"))))
			} else {
				lib.Pair(label, fmt.Sprintf("%s    shared pool%s", status, suffix))
			}
		}
		if len(phpVersions) == 0 {
			lib.Pair("  php-fpm", lib.Gray("none installed"))
		}
	}
	fmt.Println()

	// ── Projects ───────────────────────────────────────────────────────────────

	allProjects, _ := projects.ListAll(root)
	fmt.Printf("  %s  %s\n", lib.Bold("Projects"), lib.Gray(fmt.Sprintf("(%d)", len(allProjects))))
	fmt.Println()

	if len(allProjects) == 0 {
		fmt.Printf("  %s\n", lib.Gray("No projects registered. Run: chauf link"))
	} else {
		const domW = 32
		header := fmt.Sprintf(" %-20s  %-*s  %-5s  %-9s  %s",
			"Project", domW, "Domain", "PHP", "FPM", "SSL")
		sep := strings.Repeat("─", len(header))
		fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

		for _, p := range allProjects {
			fpmMode := "shared"
			if p.FPM.Dedicated {
				fpmMode = "dedicated"
			}
			fmt.Printf(" %-20s  %-*s  %-5s  %-9s  %s\n",
				p.Slug, domW, p.Domain, p.PHPVersion, fpmMode, schemeLabel(p.SSL))
			if *detail {
				fmt.Printf(" %-20s  %-*s\n", "", domW, lib.Gray(p.ConfigPath(root)))
			}
			for _, a := range p.Aliases {
				fmt.Printf(" %-20s  %-*s\n", "", domW, lib.Gray("↳ "+a))
			}
		}
	}
	fmt.Println()

	// ── PHP Versions ───────────────────────────────────────────────────────────

	fmt.Printf("  %s\n", lib.Bold("PHP versions"))
	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		phpVersions = nil
		for _, ver := range chauftruntime.PHPParityTargets() {
			if isInstalledPHPVersion(ver) {
				phpVersions = append(phpVersions, ver)
			}
		}
	}

	if len(phpVersions) == 0 {
		fmt.Printf("  %s\n", lib.Gray("None installed. Run: chauf install php 8.3"))
	} else {
		for _, ver := range phpVersions {
			label := "  " + ver
			path := filepath.Join(root, "php", ver)
			if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
				path = chauftruntime.PHPImage(ver)
			}
			if ver == cfg.PHP.DefaultVersion {
				lib.Pair(label, fmt.Sprintf("%s  %s", short(path), lib.Cyan("(default)")))
			} else {
				lib.Pair(label, short(path))
			}
		}
	}
	fmt.Println()

	// ── Cache ──────────────────────────────────────────────────────────────────

	cacheDir := filepath.Join(root, "cache")
	lib.Pair("Cache", fmt.Sprintf("%s  %s", lib.FormatBytes(lib.DirSize(cacheDir)), lib.Gray(short(cacheDir))))

	return nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func serviceStatus(running bool) string {
	if running {
		return lib.Green("● running")
	}
	return lib.Gray("○ stopped")
}

func schemeLabel(secure bool) string {
	if secure {
		return lib.Cyan("HTTPS")
	}
	return "HTTP"
}

func pidFileRunning(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func phpVersionsInUse(root string) map[string]bool {
	all, _ := projects.ListAll(root)
	used := map[string]bool{}
	for _, p := range all {
		used[p.PHPVersion] = true
	}
	return used
}

func installedPHPVersions(root string) []string {
	phpDir := filepath.Join(root, "php")
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions
}
