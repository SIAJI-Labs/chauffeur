package commands

import (
	"bufio"
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

func RunUninstall(args []string) error {
	flags := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	force := flags.Bool("force", false, "Skip confirmation prompt")
	purge := flags.Bool("purge", false, "Also remove the chauf binary")

	if err := flags.Parse(args); err != nil {
		return err
	}

	wsRoot := workspace.Root()

	if _, err := os.Stat(wsRoot); os.IsNotExist(err) {
		lib.Warn(fmt.Sprintf("Workspace not found: %s", wsRoot))
		lib.Info("Nothing to uninstall.")
		return nil
	}

	info := gatherInfo(wsRoot)

	// ── Show what will be removed ──────────────────────────────────────────────

	fmt.Printf("\nUninstalling Chauffeur...\n\n")
	lib.Pair("Workspace", wsRoot)

	if len(info.services) > 0 {
		lib.Pair("Services", strings.Join(info.services, "  ·  "))
	}

	if info.projectCount > 0 {
		lib.Pair("Projects", fmt.Sprintf("%d registered", info.projectCount))
	}

	if len(info.phpVersions) > 0 {
		lib.Pair("PHP versions", strings.Join(info.phpVersions, "  ·  "))
	}

	if info.cacheSize > 0 {
		lib.Pair("Cache", lib.FormatBytes(info.cacheSize))
	}

	fmt.Printf("\n  This will permanently delete %s and all its contents.\n", wsRoot)

	// ── Confirm ────────────────────────────────────────────────────────────────

	if !*force {
		fmt.Printf("\n  Type 'yes' to confirm: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "yes" {
			fmt.Println("\n  Aborted.")
			return nil
		}
	}

	fmt.Println()

	// ── Stop services ──────────────────────────────────────────────────────────

	stopServices(wsRoot)

	// ── Remove workspace ───────────────────────────────────────────────────────

	if err := os.RemoveAll(wsRoot); err != nil {
		return fmt.Errorf("failed to remove workspace: %w", err)
	}
	lib.Success("Workspace removed")

	// ── Remove shell PATH block ────────────────────────────────────────────────

	removeShellPath()

	// ── Optionally remove binary ───────────────────────────────────────────────

	if *purge {
		bin, err := os.Executable()
		if err == nil {
			if err := os.Remove(bin); err != nil {
				lib.Warn(fmt.Sprintf("Could not remove binary: %s", err))
			} else {
				lib.Success(fmt.Sprintf("Binary removed (%s)", bin))
			}
		}
	}

	// ── Manual cleanup instructions ────────────────────────────────────────────

	printCleanupInstructions()

	return nil
}

// ── workspace info ─────────────────────────────────────────────────────────────

type wsInfo struct {
	services     []string
	projectCount int
	phpVersions  []string
	cacheSize    int64
}

func gatherInfo(root string) wsInfo {
	var info wsInfo

	// nginx
	if pid := readPID(filepath.Join(root, "nginx", "logs", "nginx.pid")); pid > 0 && processRunning(pid) {
		info.services = append(info.services, "nginx")
	}

	// PHP versions + FPM status
	phpDir := filepath.Join(root, "php")
	if entries, err := os.ReadDir(phpDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			ver := e.Name()
			info.phpVersions = append(info.phpVersions, ver)

			pidFile := filepath.Join(phpDir, ver, "runtime", "php-fpm", "php-fpm.pid")
			if pid := readPID(pidFile); pid > 0 && processRunning(pid) {
				info.services = append(info.services, "php-fpm "+ver)
			}
		}
	}

	// Dedicated FPM pools
	projectsDir := filepath.Join(root, "projects")
	if entries, err := os.ReadDir(projectsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			info.projectCount++

			pidFile := filepath.Join(projectsDir, e.Name(), "php-fpm.pid")
			if pid := readPID(pidFile); pid > 0 && processRunning(pid) {
				info.services = append(info.services, "fpm:"+e.Name())
			}
		}
	}

	info.cacheSize = lib.DirSize(filepath.Join(root, "cache"))

	return info
}

// ── service stopping ───────────────────────────────────────────────────────────

func stopServices(root string) {
	// nginx — graceful shutdown (SIGQUIT waits for active connections)
	if pid := readPID(filepath.Join(root, "nginx", "logs", "nginx.pid")); pid > 0 && processRunning(pid) {
		sendSignal(pid, syscall.SIGQUIT)
		lib.Success("nginx stopped")
	}

	// Shared PHP-FPM pools
	phpDir := filepath.Join(root, "php")
	if entries, err := os.ReadDir(phpDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			pidFile := filepath.Join(phpDir, e.Name(), "runtime", "php-fpm", "php-fpm.pid")
			if pid := readPID(pidFile); pid > 0 && processRunning(pid) {
				sendSignal(pid, syscall.SIGTERM)
				lib.Success(fmt.Sprintf("php-fpm %s stopped", e.Name()))
			}
		}
	}

	// Dedicated FPM pools
	projectsDir := filepath.Join(root, "projects")
	if entries, err := os.ReadDir(projectsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			pidFile := filepath.Join(projectsDir, e.Name(), "php-fpm.pid")
			if pid := readPID(pidFile); pid > 0 && processRunning(pid) {
				sendSignal(pid, syscall.SIGTERM)
				lib.Success(fmt.Sprintf("fpm: %s stopped", e.Name()))
			}
		}
	}
}

// ── cleanup instructions ───────────────────────────────────────────────────────

func printCleanupInstructions() {
	fmt.Printf("\n  %s\n", lib.Gray("Manual cleanup (if needed):"))

	sections := []struct {
		title string
		lines []string
	}{
		{
			"dnsmasq:",
			[]string{
				"sudo rm /etc/dnsmasq.d/chauffeur.conf",
				"sudo systemctl restart dnsmasq",
			},
		},
		{
			"iptables (if port forwarding was enabled):",
			[]string{
				"sudo iptables -t nat -D OUTPUT -p tcp --dport 80 -j REDIRECT --to-port 8080",
				"sudo iptables -t nat -D OUTPUT -p tcp --dport 443 -j REDIRECT --to-port 8443",
			},
		},
		{
			"systemd (if auto-start was enabled):",
			[]string{
				"systemctl --user disable --now chauffeur-nginx.service",
				"rm ~/.config/systemd/user/chauffeur-*.service",
			},
		},
	}

	for _, s := range sections {
		fmt.Printf("\n  %s\n", s.title)
		for _, l := range s.lines {
			fmt.Printf("    %s\n", lib.Gray(l))
		}
	}
	fmt.Println()
}

// removeShellPath strips the chauffeur PATH block from the user's shell config.
// It removes everything between shellBlockOpen and shellBlockClose (inclusive).
func removeShellPath() {
	cfgFile := detectShellConfig()
	if cfgFile == "" {
		return
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return
	}
	content := string(data)

	if !strings.Contains(content, shellBlockOpen) {
		return // nothing to remove
	}

	// Strip the block including surrounding blank lines where possible.
	var out strings.Builder
	lines := strings.Split(content, "\n")
	skip := false
	for i, line := range lines {
		if strings.TrimSpace(line) == shellBlockOpen {
			skip = true
			// Also remove a preceding blank line if present.
			s := out.String()
			if strings.HasSuffix(s, "\n\n") {
				out.Reset()
				out.WriteString(strings.TrimSuffix(s, "\n"))
			}
			continue
		}
		if skip {
			if strings.TrimSpace(line) == shellBlockClose {
				skip = false
				// Skip the blank line immediately after the closing marker.
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
					// handled by advancing i — we just don't write the close marker
				}
			}
			continue
		}
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}

	if err := os.WriteFile(cfgFile, []byte(out.String()), 0644); err != nil {
		lib.Warn("Could not update " + cfgFile + ": " + err.Error())
		return
	}

	home, _ := os.UserHomeDir()
	display := strings.Replace(cfgFile, home, "~", 1)
	lib.Success("PATH removed from " + display)
}

// ── helpers ────────────────────────────────────────────────────────────────────

func readPID(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return pid
}

func processRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Linux, FindProcess always succeeds; signal 0 checks if process exists.
	return proc.Signal(syscall.Signal(0)) == nil
}

func sendSignal(pid int, sig syscall.Signal) {
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Signal(sig)
	}
}

