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
	pidFileName = "panel.pid"
	logFileName = "panel.log"
)

func RunServe(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = func() {
		fmt.Println("  chauf serve — start the web admin panel")
		fmt.Println()
		fmt.Println("  Starts a local HTTP server with a web-based admin panel")
		fmt.Println("  for managing database containers.")
		fmt.Println()
		fmt.Printf("  Usage:  %s\n", lib.Bold("chauf serve [flags]"))
		fmt.Println()
		fmt.Println("  Flags:")
		flags.PrintDefaults()
	}

	port := flags.Int("port", 3000, "Port to listen on")
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

	if existingPID, err := getRunningPID(pidPath); err == nil && existingPID > 0 {
		if isProcessRunning(existingPID) {
			lib.Warn(fmt.Sprintf("Panel server is already running at %s", lib.Cyan(url)))
			lib.Info(fmt.Sprintf("  PID: %d", existingPID))
			lib.Info(fmt.Sprintf("  %s", lib.Gray("Run 'chauf serve --stop' to stop")))
			return nil
		}
		os.Remove(pidPath)
	}

	if *foreground {
		return runPanelForeground(*port, *host)
	}

	if *dev {
		return runDevServer(*port)
	}

	return runPanelDaemon(root, pidPath, logPath, *port, *host)
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
	panelAppsDir := filepath.Join("internal", "panel-apps")
	goServerAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	goServer := panel.NewServerWithAddr(goServerAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		goServer.Start(ctx)
	}()

	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = panelAppsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start vite dev server: %w", err)
	}

	lib.Success("Chauffeur Panel dev mode")
	lib.Info(fmt.Sprintf("  Go API server: http://localhost:%d", port))
	lib.Info("  Frontend: http://localhost:5173 (API proxy: /api -> localhost:3000)")
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Press Ctrl+C to stop both servers"))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	lib.Info("Shutting down servers...")
	cmd.Process.Kill()
	cmd.Wait()
	cancel()

	return nil
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
	cmd := exec.Command(binaryPath, "serve", "-f", "--port", strconv.Itoa(port), "--host", host)
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
	fmt.Printf("  %s\n", lib.Gray("Run 'chauf serve --stop' to stop"))

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
