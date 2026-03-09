package commands

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/services"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── chauf start ───────────────────────────────────────────────────────────────

func RunStart(args []string) error {
	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	projectPath := flags.String("project", "", "Start only services for a specific project path")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	mgr := services.NewManager(root)

	fmt.Println()
	lib.Info("Starting services...")
	fmt.Println()

	if *projectPath != "" {
		return startForProject(mgr, root, *projectPath)
	}

	results, err := mgr.StartAll()
	for _, r := range results {
		printStartResult(r)
	}
	fmt.Println()
	if err != nil {
		return err
	}

	anyFailed := false
	for _, r := range results {
		if r.Err != nil {
			anyFailed = true
			break
		}
	}
	if anyFailed {
		return fmt.Errorf("one or more services failed to start")
	}

	lib.Success("All services running.")

	// Print active domains
	all, _ := projects.ListAll(root)
	cfg := workspace.Load()
	if len(all) > 0 {
		fmt.Println()
		for _, p := range all {
			scheme := "http"
			port := cfg.Nginx.HTTPPort
			if p.SSL {
				scheme = "https"
				port = cfg.Nginx.HTTPSPort
			}
			lib.Info(fmt.Sprintf("  %s://%s:%d", scheme, p.Domain, port))
		}
	}
	fmt.Println()
	return nil
}

func startForProject(mgr *services.Manager, root, path string) error {
	p, err := projects.FindByPath(root, path)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("no project registered at %s", path)
	}

	// Start the FPM pool for this project
	var fpm *services.FPMService
	if p.FPM.Dedicated {
		fpm = services.NewDedicatedFPM(root, p.Slug, p.PHPVersion, p.FPM.Socket)
	} else {
		fpm = services.NewSharedFPM(root, p.PHPVersion)
	}

	r := services.StartResult{Label: "php-fpm " + fpm.Label()}
	if fpm.IsRunning() {
		r.AlreadyRunning = true
		r.PID = fpm.PID()
	} else {
		if err := fpm.Start(); err != nil {
			r.Err = err
		} else {
			r.PID = fpm.PID()
		}
	}
	printStartResult(r)

	// Then nginx
	nginx := mgr.Nginx()
	nr := services.StartResult{Label: "nginx"}
	if nginx.IsRunning() {
		nr.AlreadyRunning = true
		nr.PID = nginx.PID()
	} else {
		if err := nginx.Start(); err != nil {
			nr.Err = err
		} else {
			nr.PID = nginx.PID()
		}
	}
	printStartResult(nr)
	fmt.Println()
	return nil
}

func printStartResult(r services.StartResult) {
	if r.Err != nil {
		lib.Error(fmt.Sprintf("%-22s  %s", r.Label, r.Err.Error()))
		return
	}
	if r.AlreadyRunning {
		lib.Info(fmt.Sprintf("  %s  %-22s  pid %d  %s",
			lib.Gray("·"), r.Label, r.PID, lib.Gray("(already running)")))
		return
	}
	lib.Info(fmt.Sprintf("  %s  %-22s  pid %d",
		lib.Green("✓"), r.Label, r.PID))
}

// ── chauf stop ────────────────────────────────────────────────────────────────

func RunStop(args []string) error {
	flags := flag.NewFlagSet("stop", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	projectPath := flags.String("project", "", "Stop only dedicated FPM for a specific project")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	mgr := services.NewManager(root)

	fmt.Println()
	lib.Info("Stopping services...")
	fmt.Println()

	if *projectPath != "" {
		return stopForProject(mgr, root, *projectPath)
	}

	results, err := mgr.StopAll()
	for _, r := range results {
		printStopResult(r)
	}
	fmt.Println()
	if err != nil {
		return err
	}
	lib.Success("All services stopped.")
	fmt.Println()
	return nil
}

func stopForProject(mgr *services.Manager, root, path string) error {
	p, err := projects.FindByPath(root, path)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("no project registered at %s", path)
	}
	if !p.FPM.Dedicated {
		lib.Info(fmt.Sprintf("Project %s uses shared FPM — stopping it would affect other projects.", p.Slug))
		lib.Info(lib.Gray("Use: chauf stop  to stop all services."))
		return nil
	}
	fpm := services.NewDedicatedFPM(root, p.Slug, p.PHPVersion, p.FPM.Socket)
	r := services.StopResult{Label: "php-fpm " + fpm.Label()}
	if !fpm.IsRunning() {
		r.AlreadyStopped = true
	} else {
		r.Err = fpm.Stop(30 * time.Second)
	}
	printStopResult(r)
	fmt.Println()
	return nil
}

func printStopResult(r services.StopResult) {
	if r.Err != nil {
		lib.Error(fmt.Sprintf("%-22s  %s", r.Label, r.Err.Error()))
		return
	}
	if r.AlreadyStopped {
		lib.Info(fmt.Sprintf("  %s  %-22s  %s",
			lib.Gray("·"), r.Label, lib.Gray("(not running)")))
		return
	}
	lib.Info(fmt.Sprintf("  %s  %-22s  stopped", lib.Green("✓"), r.Label))
}

// ── chauf restart ─────────────────────────────────────────────────────────────

func RunRestart(args []string) error {
	flags := flag.NewFlagSet("restart", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	projectPath := flags.String("project", "", "Restart dedicated FPM for a specific project")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	mgr := services.NewManager(root)

	// chauf restart [nginx|php|fpm [version]]
	target := flags.Arg(0)

	fmt.Println()

	switch strings.ToLower(target) {
	case "nginx":
		nginx := mgr.Nginx()
		if err := nginx.Reload(); err != nil {
			return err
		}
		lib.Success(fmt.Sprintf("nginx reloaded  (pid %d)", nginx.PID()))

	case "php", "fpm":
		version := flags.Arg(1)
		if version != "" {
			fpm := mgr.SharedFPM(version)
			if err := fpm.Reload(); err != nil {
				return err
			}
			lib.Success(fmt.Sprintf("php-fpm %s reloaded  (pid %d)", version, fpm.PID()))
		} else {
			// Reload all FPM pools
			fpms, err := mgr.AllFPM()
			if err != nil {
				return err
			}
			for _, fpm := range fpms {
				if err := fpm.Reload(); err != nil {
					lib.Warn(fmt.Sprintf("php-fpm %s: %s", fpm.Label(), err))
				} else {
					lib.Success(fmt.Sprintf("php-fpm %s reloaded  (pid %d)", fpm.Label(), fpm.PID()))
				}
			}
		}

	case "":
		if *projectPath != "" {
			return restartForProject(mgr, root, *projectPath)
		}
		// Restart everything with per-service progress
		lib.Info("Reloading services...")
		fmt.Println()

		fpms, err := mgr.AllFPM()
		if err != nil {
			return err
		}
		anyErr := false
		for _, fpm := range fpms {
			if fpm.IsRunning() {
				if err := fpm.Reload(); err != nil {
					lib.Error(fmt.Sprintf("  %-24s  %s", "php-fpm "+fpm.Label(), err))
					anyErr = true
				} else {
					lib.Info(fmt.Sprintf("  %s  %-24s  pid %d", lib.Green("✓"), "php-fpm "+fpm.Label(), fpm.PID()))
				}
			} else {
				lib.Info(fmt.Sprintf("  %s  %-24s  %s", lib.Gray("·"), "php-fpm "+fpm.Label(), lib.Gray("(not running — skipped)")))
			}
		}
		nginx := mgr.Nginx()
		if nginx.IsRunning() {
			if err := nginx.Reload(); err != nil {
				lib.Error(fmt.Sprintf("  %-24s  %s", "nginx", err))
				anyErr = true
			} else {
				lib.Info(fmt.Sprintf("  %s  %-24s  pid %d", lib.Green("✓"), "nginx", nginx.PID()))
			}
		} else {
			lib.Info(fmt.Sprintf("  %s  %-24s  %s", lib.Gray("·"), "nginx", lib.Gray("(not running — skipped)")))
		}
		fmt.Println()
		if anyErr {
			return fmt.Errorf("one or more services failed to reload")
		}
		lib.Success("All services reloaded.")

	default:
		return fmt.Errorf("unknown restart target %q — use: nginx, php, fpm, or --project", target)
	}

	fmt.Println()
	return nil
}

func restartForProject(mgr *services.Manager, root, path string) error {
	p, err := projects.FindByPath(root, path)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("no project registered at %s", path)
	}
	if p.FPM.Dedicated {
		fpm := services.NewDedicatedFPM(root, p.Slug, p.PHPVersion, p.FPM.Socket)
		if err := fpm.Reload(); err != nil {
			return err
		}
		lib.Success(fmt.Sprintf("php-fpm %s reloaded", fpm.Label()))
	} else {
		fpm := mgr.SharedFPM(p.PHPVersion)
		if err := fpm.Reload(); err != nil {
			return err
		}
		lib.Success(fmt.Sprintf("php-fpm %s reloaded  (shared pool — affects all %s projects)", fpm.Label(), p.PHPVersion))
	}
	return nil
}

// ── chauf status ──────────────────────────────────────────────────────────────

func RunStatus(args []string) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	detail := flags.Bool("detail", false, "Show memory, socket paths, config paths")
	projectPath := flags.String("project", "", "Show status for a specific project")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	mgr := services.NewManager(root)

	fmt.Println()

	if *projectPath != "" {
		return statusForProject(mgr, root, *projectPath)
	}

	nginx := mgr.Nginx()
	fpms, err := mgr.AllFPM()
	if err != nil {
		return err
	}

	if *detail {
		printServiceDetail("nginx", nginx.IsRunning(), nginx.PID(), nginx.Uptime(), nginx.MemoryMB(),
			root+"/nginx/etc/nginx.conf", root+"/nginx/logs/error.log", "")
		for _, fpm := range fpms {
			printServiceDetail("php-fpm "+fpm.Label(), fpm.IsRunning(), fpm.PID(), fpm.Uptime(), fpm.MemoryMB(),
				"", "", fpm.SockPath())
		}
		fmt.Println()
		return nil
	}

	// Tabular output
	const lblW = 22
	header := fmt.Sprintf(" %-*s  %-12s  %-8s  %-10s  %s",
		lblW, "Service", "Status", "PID", "Uptime", "Memory")
	sep := strings.Repeat("─", len(header))
	fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

	printStatusRow(lblW, "nginx", nginx.IsRunning(), nginx.PID(), nginx.Uptime(), nginx.MemoryMB())
	for _, fpm := range fpms {
		printStatusRow(lblW, "php-fpm "+fpm.Label(), fpm.IsRunning(), fpm.PID(), fpm.Uptime(), fpm.MemoryMB())
	}

	// Also show installed-but-unused PHP versions as stopped
	installed := installedPHPVersions(root)
	usedVersions := map[string]bool{}
	for _, fpm := range fpms {
		// Extract version from label "8.3 (shared)"
		parts := strings.Fields(fpm.Label())
		if len(parts) > 0 {
			usedVersions[parts[0]] = true
		}
	}
	for _, ver := range installed {
		if !usedVersions[ver] {
			printStatusRow(lblW, "php-fpm "+ver+" (unused)", false, 0, 0, 0)
		}
	}

	fmt.Println()
	return nil
}

func printStatusRow(lblW int, label string, running bool, pid int, uptime time.Duration, memMB int) {
	status := lib.Gray("○ stopped")
	pidStr := "-"
	uptimeStr := "-"
	memStr := "-"
	if running {
		status = lib.Green("● running")
		pidStr = fmt.Sprintf("%d", pid)
		uptimeStr = services.FormatUptime(uptime)
		if memMB > 0 {
			memStr = fmt.Sprintf("%d MB", memMB)
		}
	}
	fmt.Printf(" %-*s  %-20s  %-8s  %-10s  %s\n",
		lblW, label, status, pidStr, uptimeStr, memStr)
}

func printServiceDetail(label string, running bool, pid int, uptime time.Duration, memMB int, config, errorLog, socket string) {
	sep := strings.Repeat("─", 44)
	fmt.Printf("  %s\n  %s\n", lib.Bold(label), sep)
	if running {
		lib.Pair("  Status", lib.Green("● running"))
		lib.Pair("  PID", fmt.Sprintf("%d", pid))
		lib.Pair("  Uptime", services.FormatUptime(uptime))
		if memMB > 0 {
			lib.Pair("  Memory", fmt.Sprintf("%d MB", memMB))
		}
	} else {
		lib.Pair("  Status", lib.Gray("○ stopped"))
	}
	if config != "" {
		lib.Pair("  Config", shortenHome(config))
	}
	if errorLog != "" {
		lib.Pair("  Error log", shortenHome(errorLog))
	}
	if socket != "" {
		lib.Pair("  Socket", shortenHome(socket))
	}
	fmt.Println()
}

func statusForProject(mgr *services.Manager, root, path string) error {
	p, err := projects.FindByPath(root, path)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("no project registered at %s", path)
	}

	var fpm *services.FPMService
	if p.FPM.Dedicated {
		fpm = services.NewDedicatedFPM(root, p.Slug, p.PHPVersion, p.FPM.Socket)
	} else {
		fpm = services.NewSharedFPM(root, p.PHPVersion)
	}
	nginx := mgr.Nginx()

	lib.Info(fmt.Sprintf("  %s  (PHP %s, %s FPM)", lib.Bold(p.Slug), p.PHPVersion, fpmModeStr(p)))
	fmt.Println()

	domainsStr := strings.Join(p.AllDomains(), ", ")
	fmt.Printf("  %-12s  %s  %s\n", "nginx", serviceStatus(nginx.IsRunning()), lib.Gray("(handles "+domainsStr+")"))

	if p.FPM.Dedicated {
		fmt.Printf("  %-12s  %s  %s\n", "php-fpm "+p.PHPVersion, serviceStatus(fpm.IsRunning()), lib.Gray("(dedicated)"))
	} else {
		// Count how many projects share this pool
		all, _ := projects.ListAll(root)
		count := 0
		for _, other := range all {
			if !other.FPM.Dedicated && other.PHPVersion == p.PHPVersion {
				count++
			}
		}
		shared := fmt.Sprintf("(shared with %d project(s))", count-1)
		fmt.Printf("  %-12s  %s  %s\n", "php-fpm "+p.PHPVersion, serviceStatus(fpm.IsRunning()), lib.Gray(shared))
	}

	fmt.Println()
	return nil
}

func fpmModeStr(p *projects.Project) string {
	if p.FPM.Dedicated {
		return "dedicated"
	}
	return "shared"
}

// ── chauf logs ────────────────────────────────────────────────────────────────

func RunLogs(args []string) error {
	flags := flag.NewFlagSet("logs", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	follow := flags.Bool("follow", false, "Tail mode — stream new lines as they arrive")
	lines := flags.Int("lines", 50, "Number of lines to show")
	projectPath := flags.String("project", "", "Show logs for a specific project")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	target := flags.Arg(0)  // nginx | access | php | php <version>
	version := flags.Arg(1) // only used when target == "php"

	var logPath string

	if *projectPath != "" {
		p, err := projects.FindByPath(root, *projectPath)
		if err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("no project registered at %s", *projectPath)
		}
		logPath = filepath.Join(root, "nginx", "logs", p.Slug+"-error.log")
		if target == "access" {
			logPath = filepath.Join(root, "nginx", "logs", p.Slug+"-access.log")
		}
	} else {
		switch strings.ToLower(target) {
		case "", "nginx", "error":
			logPath = filepath.Join(root, "nginx", "logs", "error.log")
		case "access":
			logPath = filepath.Join(root, "nginx", "logs", "access.log")
		case "php", "fpm":
			if version != "" {
				logPath = filepath.Join(root, "php", version, "var", "log", "php-fpm.log")
			} else {
				// Show all PHP-FPM logs concatenated
				return showAllPHPLogs(root, *lines, *follow)
			}
		default:
			return fmt.Errorf("unknown log target %q — use: nginx, access, php [version]", target)
		}
	}

	return showLog(logPath, *lines, *follow)
}

func showLog(path string, lines int, follow bool) error {
	if _, err := os.Stat(path); err != nil {
		lib.Warn(fmt.Sprintf("Log file not found: %s", shortenHome(path)))
		return nil
	}

	home, _ := os.UserHomeDir()
	fmt.Printf("  %s  %s\n\n", lib.Gray("→"), strings.Replace(path, home, "~", 1))

	if !follow {
		return printLastLines(path, lines)
	}

	return tailFollow(path)
}

func showAllPHPLogs(root string, lines int, follow bool) error {
	versions := installedPHPVersions(root)
	if len(versions) == 0 {
		lib.Info("No PHP versions installed.")
		return nil
	}
	for _, ver := range versions {
		path := filepath.Join(root, "php", ver, "var", "log", "php-fpm.log")
		if _, err := os.Stat(path); err == nil {
			if err := showLog(path, lines, false); err != nil {
				lib.Warn(fmt.Sprintf("php %s: %s", ver, err))
			}
		}
	}
	if follow {
		lib.Info(lib.Gray("--follow not supported with multiple log files. Specify a version: chauf logs php 8.3 --follow"))
	}
	return nil
}

// printLastLines reads the last n lines of a file efficiently.
func printLastLines(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Collect all lines (simple approach — logs are typically not huge)
	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := 0
	if len(allLines) > n {
		start = len(allLines) - n
	}
	for _, line := range allLines[start:] {
		fmt.Println(line)
	}
	return scanner.Err()
}

// tailFollow streams new content appended to path until Ctrl+C.
func tailFollow(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Seek to end so we only show new lines
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	lib.Info(lib.Gray("Streaming — press Ctrl+C to stop"))
	fmt.Println()

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			fmt.Println()
			return nil
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if len(line) > 0 {
					fmt.Print(line)
				}
				if err != nil {
					break // no more data yet
				}
			}
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func shortenHome(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, home, "~", 1)
}
