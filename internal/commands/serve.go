package commands

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/panel"
	"github.com/siegg/chauffeur/internal/workspace"
)

const (
	panelDomain = "panel.test"
	panelPort   = 3083
	pidFileName = "panel.pid"
	logFileName = "panel.log"
)

func runWebUIServer(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = func() {
		fmt.Println("  chauf webui start — start the web admin panel")
		fmt.Println()
		fmt.Println("  Starts a local HTTP server with a web-based admin panel")
		fmt.Println("  for managing database containers.")
		fmt.Println()
		fmt.Printf("  Usage:  %s\n", lib.Bold("chauf webui start [flags]"))
		fmt.Println()
		fmt.Println("  Flags:")
		flags.PrintDefaults()
	}

	port := flags.Int("port", panelPort, "Port to listen on")
	host := flags.String("host", panelDomain, "Hostname for the panel")
	stop := flags.Bool("stop", false, "Stop a running panel server")
	foreground := flags.Bool("f", false, "Run in foreground instead of background")
	dev := flags.Bool("dev", false, "Run Vite dev server for frontend development")

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	root := workspace.Root()
	pidPath := filepath.Join(root, pidFileName)
	logPath := filepath.Join(root, logFileName)
	url := fmt.Sprintf("http://%s", net.JoinHostPort(*host, strconv.Itoa(*port)))

	if *stop {
		return stopPanelServer(pidPath, url)
	}
	if *dev {
		return runDevServer(*port)
	}

	if existingPID, err := getRunningPID(pidPath); err == nil && existingPID > 0 {
		if isProcessRunning(existingPID) {
			lib.Warn(fmt.Sprintf("Panel server is already running at %s", lib.Cyan(url)))
			lib.Info(fmt.Sprintf("  PID: %d", existingPID))
			lib.Info(fmt.Sprintf("  %s", lib.Gray("Run 'chauf webui stop' to stop")))
			return nil
		}
		os.Remove(pidPath)
	}

	if *foreground {
		return runPanelForeground(*port, *host)
	}

	return runPanelDaemon(root, pidPath, logPath, *port, *host)
}

// RunWebUI provides the canonical lifecycle command for the local web UI.
// The existing serve command remains the compatibility entry point.
func RunWebUI(args []string) error {
	if len(args) == 0 {
		return webUIHelp()
	}

	switch args[0] {
	case "start":
		startArgs := withoutArg(args[1:], "--fresh")
		if err := rebuildWebUIIfChanged(startArgs, containsArg(args[1:], "--fresh")); err != nil {
			return err
		}
		return runWebUIServer(startArgs)
	case "stop":
		return runWebUIServer(append([]string{"--stop"}, args[1:]...))
	case "status":
		return webUIStatus(args[1:])
	case "help", "--help", "-h":
		return webUIHelp()
	default:
		return fmt.Errorf("unknown webui subcommand %q — use: start, status, stop", args[0])
	}
}

// rebuildWebUIIfChanged keeps source-checkout development honest. The panel is
// embedded in the Go binary, so refreshing frontend assets without replacing
// the executable would still serve the old UI.
func rebuildWebUIIfChanged(args []string, force bool) error {
	if os.Getenv("CHAUFFEUR_WEBUI_REBUILT") == "1" || containsArg(args, "--dev") {
		return nil
	}

	repoRoot, ok := sourceRepoRoot()
	if !ok {
		return nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil
	}
	executableInfo, err := os.Stat(executable)
	if err != nil {
		return nil
	}

	changed, err := sourceIsNewer(repoRoot, executableInfo.ModTime())
	if err != nil {
		return err
	}
	if !changed && !force {
		return nil
	}

	if force {
		lib.Info("Fresh build requested, rebuilding Web UI...")
	} else {
		lib.Info("Changes detected, rebuilding Web UI...")
	}
	build := exec.Command("make", "build")
	build.Dir = repoRoot
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("rebuild Web UI: %w", err)
	}

	binary := filepath.Join(repoRoot, "build", "chauf")
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("rebuilt Web UI binary not found: %w", err)
	}

	env := append(os.Environ(), "CHAUFFEUR_WEBUI_REBUILT=1")
	return syscall.Exec(binary, append([]string{binary, "webui", "start"}, args...), env)
}

func containsArg(args []string, wanted string) bool {
	for _, arg := range args {
		if arg == wanted {
			return true
		}
	}
	return false
}

func withoutArg(args []string, unwanted string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg != unwanted {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

func sourceRepoRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, "internal", "panel-apps", "package.json")) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sourceIsNewer(root string, since time.Time) (bool, error) {
	paths := []string{
		filepath.Join(root, "internal", "panel-apps", "index.html"),
		filepath.Join(root, "internal", "panel-apps", "src"),
		filepath.Join(root, "internal", "panel-apps", "public"),
		filepath.Join(root, "internal", "panel-apps", "package.json"),
		filepath.Join(root, "internal", "panel-apps", "package-lock.json"),
		filepath.Join(root, "internal", "panel-apps", "vite.config.ts"),
		filepath.Join(root, "Makefile"),
		filepath.Join(root, "internal", "panel", "static"),
	}
	for _, path := range paths {
		changed, err := pathIsNewer(path, since)
		if err != nil {
			return false, err
		}
		if changed {
			return true, nil
		}
	}
	return false, nil
}

func pathIsNewer(path string, since time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return info.ModTime().After(since), nil
	}
	var newer bool
	err = filepath.WalkDir(path, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if newer {
			return filepath.SkipAll
		}
		if entry.IsDir() {
			return nil
		}
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		newer = fileInfo.ModTime().After(since)
		return nil
	})
	return newer, err
}

func webUIHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf webui — manage the local web UI"))
	fmt.Printf("  %s\n\n", lib.Gray("Usage: chauf webui <start|status|stop> [flags]"))
	fmt.Printf("  %-18s  %s\n", "start", lib.Gray("Start the web UI in the background"))
	fmt.Printf("  %-18s  %s\n", "status", lib.Gray("Show whether the web UI is running"))
	fmt.Printf("  %-18s  %s\n", "stop", lib.Gray("Stop the web UI"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Use chauf webui start [--port PORT] [--host HOST] [--fresh|--dev]"))
	fmt.Println()
	return nil
}

func webUIStatus(args []string) error {
	flags := flag.NewFlagSet("webui status", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	port := flags.Int("port", panelPort, "Web UI port")
	host := flags.String("host", panelDomain, "Web UI hostname")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	pidPath := filepath.Join(root, pidFileName)
	url := fmt.Sprintf("http://%s", net.JoinHostPort(*host, strconv.Itoa(*port)))
	pid, err := getRunningPID(pidPath)
	if err != nil || pid <= 0 || !isProcessRunning(pid) {
		if err == nil {
			_ = os.Remove(pidPath)
		}
		lib.Info("Web UI is stopped")
		lib.Pair("URL", url)
		return nil
	}

	lib.Success("Web UI is running")
	lib.Pair("URL", lib.Cyan(url))
	lib.Pair("PID", strconv.Itoa(pid))
	lib.Pair("Log", filepath.Join(root, logFileName))
	return nil
}

func runPanelForeground(port int, host string) error {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	server := panel.NewServerWithAddr(addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		lib.Info("Shutting down server...")
		cancel()
	}()

	url := fmt.Sprintf("http://%s", net.JoinHostPort(host, strconv.Itoa(port)))
	lib.Success(fmt.Sprintf("Chauffeur Panel running at %s", lib.Cyan(url)))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Press Ctrl+C to stop"))

	return server.Start(ctx)
}

func runDevServer(port int) error {
	repoRoot, ok := sourceRepoRoot()
	if !ok {
		return fmt.Errorf("web UI development mode requires a Chauffeur source checkout")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("web UI development mode requires npm in PATH")
	}

	panelAppsDir := filepath.Join(repoRoot, "internal", "panel-apps")
	goServerAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	goServer := panel.NewDevServerWithAddr(goServerAddr, "http://localhost:5173")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- goServer.Start(ctx)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("start web UI API server: %w", err)
		}
		return fmt.Errorf("web UI API server stopped before Vite started")
	case <-time.After(300 * time.Millisecond):
	}

	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = panelAppsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("CHAUFFEUR_WEBUI_API_PORT=%d", port))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start vite dev server: %w", err)
	}
	viteErr := make(chan error, 1)
	go func() {
		viteErr <- cmd.Wait()
	}()

	lib.Success("Chauffeur Panel dev mode")
	lib.Info(fmt.Sprintf("  Go API server: http://localhost:%d", port))
	lib.Info(fmt.Sprintf("  Frontend with HMR: http://localhost:5173"))
	lib.Info(fmt.Sprintf("  API proxy: /api -> localhost:%d", port))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Press Ctrl+C to stop both servers"))

	select {
	case <-ctx.Done():
		lib.Info("Shutting down development servers...")
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-viteErr:
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
		}
		return nil
	case err := <-serverErr:
		_ = cmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("web UI API server: %w", err)
		}
		return fmt.Errorf("web UI API server stopped unexpectedly")
	case err := <-viteErr:
		if err != nil {
			return fmt.Errorf("vite dev server: %w", err)
		}
		return nil
	}
}

func getBinaryPath() string {
	path, err := os.Executable()
	if err != nil {
		return os.Args[0]
	}
	return path
}

func runPanelDaemon(root, pidPath, logPath string, port int, host string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	logFile.Close()

	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "webui", "start", "-f", "--port", strconv.Itoa(port), "--host", host)
	cmd.Stdout, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	cmd.Stderr, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	cmd.Dir = root
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	pid := cmd.Process.Pid

	time.Sleep(200 * time.Millisecond)

	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return fmt.Errorf("process exited immediately with code %d", cmd.ProcessState.ExitCode())
	}

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("write pid file: %w", err)
	}

	url := fmt.Sprintf("http://%s", net.JoinHostPort(host, strconv.Itoa(port)))
	lib.Success(fmt.Sprintf("Chauffeur Panel started in background"))
	lib.Info(fmt.Sprintf("  PID: %d", pid))
	lib.Info(fmt.Sprintf("  URL: %s", lib.Cyan(url)))
	lib.Info(fmt.Sprintf("  Log: %s", logPath))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Run 'chauf webui stop' to stop"))

	return nil
}

func stopPanelServer(pidPath, url string) error {
	pid, err := getRunningPID(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			lib.Info("Panel server is not running")
			return nil
		}
		return fmt.Errorf("read pid file: %w", err)
	}

	if !isProcessRunning(pid) {
		os.Remove(pidPath)
		lib.Info("Panel server was not running (stale PID file)")
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if proc.Kill() == nil {
			os.Remove(pidPath)
			lib.Success("Panel server stopped (force killed)")
			return nil
		}
		return fmt.Errorf("send signal: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	if isProcessRunning(pid) {
		proc.Kill()
		os.Remove(pidPath)
		lib.Success("Panel server stopped (force killed)")
		return nil
	}

	os.Remove(pidPath)
	lib.Success("Panel server stopped")

	return nil
}

func getRunningPID(pidPath string) (int, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func isProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
