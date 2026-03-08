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
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunInfo(args []string) error {
	flags := flag.NewFlagSet("info", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
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
	fmt.Println()

	// ── Services ───────────────────────────────────────────────────────────────

	fmt.Printf("  %s\n", lib.Bold("Services"))

	// nginx
	nginxPID := filepath.Join(root, "nginx", "logs", "nginx.pid")
	nginxRunning := pidFileRunning(nginxPID)
	nginxStatus := serviceStatus(nginxRunning)
	if *detail {
		lib.Pair("  nginx", fmt.Sprintf("%s    %d / %d    %s",
			nginxStatus, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort,
			lib.Gray(filepath.Join(root, "nginx", "bin", "nginx"))))
	} else {
		lib.Pair("  nginx", fmt.Sprintf("%s    %d / %d",
			nginxStatus, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort))
	}

	// PHP-FPM per installed version
	phpVersions := installedPHPVersions(root)
	for _, ver := range phpVersions {
		fpmPID := filepath.Join(root, "php", ver, "runtime", "php-fpm", "php-fpm.pid")
		running := pidFileRunning(fpmPID)
		status := serviceStatus(running)
		label := fmt.Sprintf("  php-fpm %s", ver)
		suffix := ""
		if ver == cfg.PHP.DefaultVersion {
			suffix = lib.Gray("  (default)")
		}
		if *detail {
			lib.Pair(label, fmt.Sprintf("%s    shared pool%s    %s",
				status, suffix,
				lib.Gray(filepath.Join(root, "php", ver, "bin", "php-fpm"))))
		} else {
			lib.Pair(label, fmt.Sprintf("%s    shared pool%s", status, suffix))
		}
	}
	if len(phpVersions) == 0 {
		lib.Pair("  php-fpm", lib.Gray("none installed"))
	}
	fmt.Println()

	// ── Projects ───────────────────────────────────────────────────────────────

	projects := listProjects(root)
	fmt.Printf("  %s  %s\n", lib.Bold("Projects"), lib.Gray(fmt.Sprintf("(%d)", len(projects))))

	if len(projects) == 0 {
		fmt.Printf("  %s\n", lib.Gray("No projects registered. Run: chauf link <path>"))
	} else {
		for _, p := range projects {
			if *detail {
				lib.Pair("  "+p.Slug, fmt.Sprintf("%-28s PHP %-4s %-10s %s",
					p.Domain, p.PHPVersion, p.Mode, lib.Gray(p.ConfigPath)))
			} else {
				lib.Pair("  "+p.Slug, fmt.Sprintf("%-28s PHP %-4s %-10s %s",
					p.Domain, p.PHPVersion, p.Mode, schemeLabel(p.Secure)))
			}
		}
	}
	fmt.Println()

	// ── PHP Versions ───────────────────────────────────────────────────────────

	fmt.Printf("  %s\n", lib.Bold("PHP versions"))

	if len(phpVersions) == 0 {
		fmt.Printf("  %s\n", lib.Gray("None installed. Run: chauf install php 8.3"))
	} else {
		for _, ver := range phpVersions {
			label := "  " + ver
			path := filepath.Join(root, "php", ver)
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

type projectInfo struct {
	Slug       string
	Domain     string
	PHPVersion string
	Mode       string // "shared" or "dedicated"
	Secure     bool
	ConfigPath string
}

func listProjects(root string) []projectInfo {
	projectsDir := filepath.Join(root, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil
	}

	var projects []projectInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slug := e.Name()
		cfgPath := filepath.Join(projectsDir, slug, "config.yaml")
		p := projectInfo{
			Slug:       slug,
			Domain:     slug + ".test",
			PHPVersion: "8.3",
			Mode:       "shared",
			Secure:     false,
			ConfigPath: cfgPath,
		}
		if data, err := os.ReadFile(cfgPath); err == nil {
			parseProjectConfig(data, &p)
		}
		projects = append(projects, p)
	}
	return projects
}

func parseProjectConfig(data []byte, p *projectInfo) {
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(strings.Trim(parts[1], `"`))
		switch key {
		case "domain":
			p.Domain = val
		case "php_version":
			p.PHPVersion = val
		case "fpm_mode":
			p.Mode = val
		case "secure":
			p.Secure = val == "true"
		}
	}
}

