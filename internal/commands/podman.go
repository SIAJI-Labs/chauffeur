package commands

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/podman"
	"github.com/siegg/chauffeur/internal/services"
	"github.com/siegg/chauffeur/internal/workspace"
)

// RunPodman dispatches to the appropriate podman subcommand.
func RunPodman(args []string) error {
	if len(args) == 0 {
		return runPodmanHelp(args)
	}

	switch args[0] {
	case "create":
		return runPodmanCreate(args[1:])
	case "start":
		return runPodmanStart(args[1:])
	case "stop":
		return runPodmanStop(args[1:])
	case "status":
		return runPodmanStatus(args[1:])
	case "list", "ls":
		return runPodmanList(args[1:])
	case "remove", "rm":
		return runPodmanRemove(args[1:])
	case "console":
		return runPodmanConsole(args[1:])
	case "import":
		return runPodmanImport(args[1:])
	case "backup":
		return runPodmanBackup(args[1:])
	case "restore":
		return runPodmanRestore(args[1:])
	case "help", "--help", "-h":
		return runPodmanHelp(args[1:])
	default:
		lib.Error(fmt.Sprintf("unknown subcommand %q", args[0]))
		fmt.Println()
		return runPodmanHelp(args[1:])
	}
}

// ── chauf podman help ──────────────────────────────────────────────────────────

func runPodmanHelp(args []string) error {
	flags := flag.NewFlagSet("podman help", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman — manage shared database containers via Podman",
		"chauf podman <create|start|stop|status|list|remove|console|backup|restore> [args]")

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Podman database containers"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Manage shared database containers (MySQL, PostgreSQL, MariaDB, MongoDB, Redis)"))
	fmt.Printf("  %s\n", lib.Gray("using Podman. Containers are independent of PHP services."))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Commands:"))
	fmt.Printf("    %-16s  %s\n", "create", lib.Gray("Create a new database container"))
	fmt.Printf("    %-16s  %s\n", "start", lib.Gray("Start a container (or 'all')"))
	fmt.Printf("    %-16s  %s\n", "stop", lib.Gray("Stop a container (or 'all')"))
	fmt.Printf("    %-16s  %s\n", "status", lib.Gray("Show container status"))
	fmt.Printf("    %-16s  %s\n", "list", lib.Gray("List all managed containers"))
	fmt.Printf("    %-16s  %s\n", "remove", lib.Gray("Remove a container"))
	fmt.Printf("    %-16s  %s\n", "console", lib.Gray("Attach to container for CLI access"))
	fmt.Printf("    %-16s  %s\n", "backup", lib.Gray("Backup a container to file"))
	fmt.Printf("    %-16s  %s\n", "restore", lib.Gray("Restore a container from backup"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray(`Run "chauf podman <command> --help" for detailed usage.`))
	fmt.Println()
	return nil
}

// ── chauf podman create ───────────────────────────────────────────────────────

func runPodmanCreate(args []string) error {
	flags := flag.NewFlagSet("podman create", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman create — create a database container",
		"chauf podman create [mysql|postgres|maria|mongo|redis] [--name <container-name>] [--user <user>] [--pass <pass>] [--port <port>] [--volume <path>]")

	nameFlag := flags.String("name", "", "Container name (default: chauf-<engine>)")
	userFlag := flags.String("user", "", "Database username (auto-generated if not set)")
	passFlag := flags.String("pass", "", "Database password (auto-generated if not set)")
	portFlag := flags.Int("port", 0, "Host port to expose (default varies by engine)")
	volumeFlag := flags.String("volume", "", "Volume path for data persistence")
	yesFlag := flags.Bool("yes", false, "Skip confirmation if container exists")

	// Reorder args: move all flags (with their values) before positional args
	// This allows: chauf podman create mysql57 --name foo --port 3307
	// We need to keep flag-value pairs together
	var flagArgs, posArgs []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// This is a flag
			flagArgs = append(flagArgs, arg)
			// Check if next arg is a value (not a flag)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagArgs = append(flagArgs, args[i+1])
				i++
			}
		} else {
			posArgs = append(posArgs, arg)
		}
		i++
	}
	reorderedArgs := append(flagArgs, posArgs...)
	if err := flags.Parse(reorderedArgs); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()

	// Check podman availability
	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			lib.Info("Install it from: https://podman.io/getting-started/installation")
			return nil
		}
		return err
	}

	if verbose {
		lib.Info("Podman available, proceeding with container creation...")
	}

	// Determine if we're in full interactive mode (no engine arg, no yes flag)
	// Note: --verbose is allowed in interactive mode
	flagSet := make(map[string]bool)
	flags.Visit(func(f *flag.Flag) {
		flagSet[f.Name] = true
	})
	// Interactive if no engine arg AND (no flags OR only --verbose)
	// Note: --verbose is consumed globally by main.go, so we check lib.Verbose
	isInteractive := flags.NArg() == 0 && len(flagSet) == 0

	var engineStr string
	var engine podman.EngineType
	cfg := &podman.DatabaseConfig{}

	if isInteractive {
		// Full interactive wizard
		engineStr = interactiveSelectEngine()
		if engineStr == "" {
			return nil // user cancelled
		}
		engine = podman.EngineType(engineStr)
		*cfg = *podman.DefaultConfig(engine)

		// Step 1: container name
		promptField("Container name", &cfg.ContainerName, cfg.ContainerName, false)

		// Update volume path default to use container name
		cfg.VolumePath = filepath.Join(podman.Root(), "volumes", cfg.ContainerName)

		// Step 2: username and password
		promptField("Username", &cfg.Username, "chauf", false)
		promptField("Password", &cfg.Password, cfg.Password, false)

		// Step 3: port — loop until available
		for {
			promptFieldInt("Port", &cfg.Port, cfg.Port, false)
			if services.IsPortAvailable(cfg.Port) {
				break
			}
			pid, name, _ := services.FindProcessOnPort(cfg.Port)
			if pid > 0 {
				lib.Warn(fmt.Sprintf("Port %d is already in use by %s (pid %d)", cfg.Port, name, pid))
			} else {
				lib.Warn(fmt.Sprintf("Port %d is not available", cfg.Port))
			}
			fmt.Println()
		}

		// Step 4: volume path
		promptField("Volume path", &cfg.VolumePath, cfg.VolumePath, false)
	} else {
		// Direct or partial mode: engine required
		if flags.NArg() != 1 {
			flags.Usage()
			return flag.ErrHelp
		}
		engineStr = strings.ToLower(flags.Arg(0))
		if !podman.IsValidEngine(engineStr) {
			lib.Error(fmt.Sprintf("unsupported engine %q", flags.Arg(0)))
			lib.Info(lib.Gray("Supported: mysql8, mysql57, postgres, maria, mongo, redis"))
			return nil
		}

		engine = podman.EngineType(engineStr)
		*cfg = *podman.DefaultConfig(engine)

		// Apply flag overrides only if explicitly provided
		if flagSet["name"] {
			cfg.ContainerName = *nameFlag
		}
		if flagSet["user"] {
			cfg.Username = *userFlag
		}
		if flagSet["pass"] {
			cfg.Password = *passFlag
		}
		if flagSet["port"] {
			cfg.Port = *portFlag
		}
		if flagSet["volume"] {
			cfg.VolumePath = *volumeFlag
		}
	}

	// Validate port is available
	if !services.IsPortAvailable(cfg.Port) {
		pid, name, _ := services.FindProcessOnPort(cfg.Port)
		if pid > 0 {
			lib.Error(fmt.Sprintf("port %d is already in use by process %d (%s)", cfg.Port, pid, name))
		} else {
			lib.Error(fmt.Sprintf("port %d is not available", cfg.Port))
		}
		return nil
	}

	// Validate volume path is writable
	if err := os.MkdirAll(cfg.VolumePath, 0755); err != nil {
		lib.Error(fmt.Sprintf("cannot create volume path %s: %v", cfg.VolumePath, err))
		return nil
	}
	if !isDirWritable(cfg.VolumePath) {
		lib.Error(fmt.Sprintf("volume path %s is not writable", cfg.VolumePath))
		return nil
	}

	// Check if already exists
	existingCfg, err := podman.Load(engineStr)
	if err == nil && existingCfg != nil {
		if !*yesFlag && !isInteractive {
			fmt.Println()
			lib.Info(fmt.Sprintf("Container %s already exists.", lib.Bold(engineStr)))
			fmt.Print("  Recreate it [y/N]? ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			resp := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if resp != "y" && resp != "yes" {
				lib.Info("Cancelled.")
				return nil
			}
		} else if *yesFlag {
			// Just remove and recreate
		} else {
			// Interactive mode: confirm interactively
			fmt.Println()
			lib.Info(fmt.Sprintf("Container %s already exists.", lib.Bold(engineStr)))
			if !interactiveConfirm("Recreate it?", nil) {
				lib.Info("Cancelled.")
				return nil
			}
		}
		// Remove existing container
		container := podman.NewContainer(client, existingCfg)
		if err := container.Remove(ctx, true); err != nil {
			lib.Warn("Could not remove existing container: " + err.Error())
		}
	}

	// Save config first
	if err := podman.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Create container
	container := podman.NewContainer(client, cfg)

	if verbose {
		// Verbose mode: show detailed progress
		fmt.Println()
		fmt.Printf("  %s %s\n", lib.Bold("Engine:"), cfg.Engine)
		fmt.Printf("  %s %s\n", lib.Bold("Container:"), cfg.ContainerName)
		fmt.Printf("  %s %s\n", lib.Bold("Image:"), cfg.Image)
		fmt.Printf("  %s %d\n", lib.Bold("Port:"), cfg.Port)
		fmt.Printf("  %s %s\n", lib.Bold("Volume:"), cfg.VolumePath)
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold("Creating container..."))
		os.Stdout.Sync() // flush before podman pull starts
		fmt.Println()

		// Set logger for verbose output during creation
		container.SetLogger(&verboseLogger{verbose: true})
		if err := container.Create(ctx); err != nil {
			return fmt.Errorf("create container: %w", err)
		}
		fmt.Println()
	} else {
		// Normal mode: show spinner
		spinner := lib.NewSpinner("Creating container...")
		if err := container.Create(ctx); err != nil {
			spinner.Fail("Failed to create container")
			return fmt.Errorf("create container: %w", err)
		}
		spinner.Success("Container created")
	}

	fmt.Println()
	lib.Pair("Engine", string(cfg.Engine))
	lib.Pair("Image", cfg.Image)
	lib.Pair("Container", cfg.ContainerName)
	lib.Pair("Username", cfg.Username)
	lib.Pair("Password", cfg.Password)
	lib.Pair("Port", fmt.Sprintf("%d", cfg.Port))
	lib.Pair("Volume", cfg.VolumePath)
	fmt.Println()
	lib.Info("DSN:")
	fmt.Printf("  %s\n", lib.Cyan(podman.DSN(cfg)))
	fmt.Println()

	return nil
}

// ── chauf podman start ────────────────────────────────────────────────────────

func runPodmanStart(args []string) error {
	flags := flag.NewFlagSet("podman start", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman start — start a container",
		"chauf podman start [container-name|engine|all]")

	if err := flags.Parse(args); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	target := ""
	if flags.NArg() > 0 {
		target = strings.ToLower(flags.Arg(0))
	}

	// Interactive mode if no target specified
	if target == "" {
		engines, _ := podman.ListEngines()
		if len(engines) == 0 {
			lib.Info("No containers configured.")
			return nil
		}

		var containerList []containerInfo
		for _, e := range engines {
			cfg, _ := podman.Load(e)
			if cfg == nil {
				continue
			}
			container := podman.NewContainer(client, cfg)
			running, _ := container.IsRunning(ctx)
			status := lib.Gray("stopped")
			disabled := false
			if running {
				status = lib.Green("running")
				disabled = true
			}
			containerList = append(containerList, containerInfo{
				Name:     e,
				Engine:   string(cfg.Engine),
				Status:   status,
				Disabled: disabled,
			})
		}

		if len(containerList) == 0 {
			lib.Info("No containers configured.")
			return nil
		}

		selected := interactiveSelectContainers(containerList, "Select containers to start:")
		if len(selected) == 0 {
			return nil
		}
		target = strings.Join(selected, ",")
	}

	containers, err := resolveContainers(target)
	if err != nil || len(containers) == 0 {
		lib.Warn("No containers found matching " + lib.Bold(target))
		return nil
	}

	for _, cfg := range containers {
		container := podman.NewContainer(client, cfg)
		running, _ := container.IsRunning(ctx)
		if running {
			lib.Warn(fmt.Sprintf("Container %s is already running", lib.Bold(cfg.ContainerName)))
			continue
		}

		if verbose {
			fmt.Printf("  Starting %s...\n", cfg.ContainerName)
		}
		if err := container.Start(ctx); err != nil {
			return fmt.Errorf("start %s: %w", cfg.ContainerName, err)
		}
		lib.Success(fmt.Sprintf("Started %s", lib.Bold(cfg.ContainerName)))
	}

	return nil
}

// ── chauf podman stop ─────────────────────────────────────────────────────────

func runPodmanStop(args []string) error {
	flags := flag.NewFlagSet("podman stop", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman stop — stop a container",
		"chauf podman stop [container-name|engine|all] [--time <seconds>]")

	timeout := flags.Int("time", 10, "Seconds to wait before killing the container")

	// Reorder args: move all flags (with their values) before positional args
	// This allows: chauf podman stop chauf-mysql57 --time 20
	var flagArgs, posArgs []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// This is a flag
			flagArgs = append(flagArgs, arg)
			// Check if next arg is a value (not a flag)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagArgs = append(flagArgs, args[i+1])
				i++
			}
		} else {
			posArgs = append(posArgs, arg)
		}
		i++
	}
	reorderedArgs := append(flagArgs, posArgs...)
	if err := flags.Parse(reorderedArgs); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	target := ""
	if flags.NArg() > 0 {
		target = strings.ToLower(flags.Arg(0))
	}

	// Interactive mode if no target specified
	if target == "" {
		engines, _ := podman.ListEngines()
		if len(engines) == 0 {
			lib.Info("No containers configured.")
			return nil
		}

		var containerList []containerInfo
		for _, e := range engines {
			cfg, _ := podman.Load(e)
			if cfg == nil {
				continue
			}
			container := podman.NewContainer(client, cfg)
			running, _ := container.IsRunning(ctx)
			status := lib.Green("running")
			disabled := false
			if !running {
				status = lib.Gray("stopped")
				disabled = true
			}
			containerList = append(containerList, containerInfo{
				Name:     e,
				Engine:   string(cfg.Engine),
				Status:   status,
				Disabled: disabled,
			})
		}

		if len(containerList) == 0 {
			lib.Info("No running containers.")
			return nil
		}

		selected := interactiveSelectContainers(containerList, "Select containers to stop:")
		if len(selected) == 0 {
			return nil
		}
		target = strings.Join(selected, ",")
	}

	containers, err := resolveContainers(target)
	if err != nil || len(containers) == 0 {
		lib.Warn("No containers found matching " + lib.Bold(target))
		return nil
	}

	for _, cfg := range containers {
		container := podman.NewContainer(client, cfg)
		running, _ := container.IsRunning(ctx)
		if !running {
			lib.Warn(fmt.Sprintf("Container %s is not running", lib.Bold(cfg.ContainerName)))
			continue
		}

		if verbose {
			fmt.Printf("  Stopping %s (timeout=%ds)...\n", cfg.ContainerName, *timeout)
		}
		if err := container.Stop(ctx, time.Duration(*timeout)*time.Second); err != nil {
			return fmt.Errorf("stop %s: %w", cfg.ContainerName, err)
		}
		lib.Success(fmt.Sprintf("Stopped %s", lib.Bold(cfg.ContainerName)))
	}

	return nil
}

// ── chauf podman status ────────────────────────────────────────────────────────

func runPodmanStatus(args []string) error {
	flags := flag.NewFlagSet("podman status", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman status — show container status",
		"chauf podman status [container-name|engine|all]")

	if err := flags.Parse(args); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	target := ""
	if flags.NArg() > 0 {
		target = strings.ToLower(flags.Arg(0))
	}

	var containers []*podman.DatabaseConfig

	if target == "" {
		// No target: show all
		containers, _ = resolveContainers("all")
	} else {
		containers, _ = resolveContainers(target)
	}

	if len(containers) == 0 {
		lib.Warn("No containers found")
		return nil
	}

	fmt.Println()

	if verbose {
		// Detailed view
		for _, cfg := range containers {
			container := podman.NewContainer(client, cfg)
			status, err := container.Status(ctx)
			if err != nil {
				fmt.Printf("  %s  %s\n", lib.Bold(cfg.ContainerName), lib.Red("error"))
				continue
			}

			statusStr := lib.Gray("○ stopped")
			if status.Running {
				statusStr = lib.Green("● running")
			}

			fmt.Printf("  %s  %s  :%d\n", lib.Bold(cfg.ContainerName), statusStr, status.HostPort)
			fmt.Printf("    Engine:  %s\n", cfg.Engine)
			fmt.Printf("    Image:   %s\n", cfg.Image)
			fmt.Printf("    Volume:  %s\n", cfg.VolumePath)
			if status.Running {
				fmt.Printf("    DSN:     %s\n", podman.DSN(cfg))
			}
		}
	} else {
		// Compact table view
		header := fmt.Sprintf(" %-20s  %-10s  %-8s  %s",
			"Container", "Engine", "Status", "Port")
		sep := strings.Repeat("─", len(header))
		fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

		for _, cfg := range containers {
			container := podman.NewContainer(client, cfg)
			status, err := container.Status(ctx)
			if err != nil {
				fmt.Printf(" %-20s  %-10s  %s\n", cfg.ContainerName, string(cfg.Engine), lib.Red("error"))
				continue
			}

			statusStr := lib.Gray("○ stopped")
			if status.Running {
				statusStr = lib.Green("● running")
			}

			fmt.Printf(" %-20s  %-10s  %-8s  :%d\n",
				cfg.ContainerName, string(cfg.Engine), statusStr, status.HostPort)
		}
	}

	fmt.Println()
	return nil
}

// ── chauf podman list ─────────────────────────────────────────────────────────

func runPodmanList(args []string) error {
	flags := flag.NewFlagSet("podman list", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman list — list all managed containers",
		"chauf podman list")

	if err := flags.Parse(args); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	engines, err := podman.ListEngines()
	if err != nil {
		return fmt.Errorf("list engines: %w", err)
	}

	if len(engines) == 0 {
		fmt.Println()
		lib.Info("No database containers configured.")
		fmt.Println()
		lib.Info(lib.Gray("Create one with:  chauf podman create <engine>"))
		fmt.Println()
		return nil
	}

	fmt.Println()

	header := fmt.Sprintf(" %-20s  %-10s  %-8s  %s",
		"Container", "Engine", "Status", "Port")
	sep := strings.Repeat("─", len(header))
	fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

	for _, engine := range engines {
		cfg, err := podman.Load(engine)
		if err != nil {
			continue
		}

		container := podman.NewContainer(client, cfg)
		status, err := container.Status(ctx)
		statusStr := lib.Gray("○ unknown")
		if err != nil {
			if err == podman.ErrContainerNotFound {
				statusStr = lib.Red("⚠ missing")
			}
		} else if status != nil {
			if status.Running {
				statusStr = lib.Green("● running")
			} else {
				statusStr = lib.Gray("○ stopped")
			}
		}

		port := cfg.Port
		if status != nil && status.HostPort > 0 {
			port = status.HostPort
		}
		fmt.Printf(" %-20s  %-10s  %-8s  :%d\n",
			cfg.ContainerName, string(cfg.Engine), statusStr, port)
	}

	if verbose {
		fmt.Println()
		fmt.Printf("  Config files: %s\n", lib.Gray(podman.Root()))
		for _, engine := range engines {
			cfg, _ := podman.Load(engine)
			if cfg != nil {
				fmt.Printf("    %-10s  →  %s\n", cfg.ContainerName, podman.ConfigPath(engine))
			}
		}
	}

	fmt.Println()
	return nil
}

// ── chauf podman remove ───────────────────────────────────────────────────────

func runPodmanRemove(args []string) error {
	flags := flag.NewFlagSet("podman remove", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman remove — remove a container",
		"chauf podman remove [container-name|engine] [--force] [--yes]")

	forceFlag := flags.Bool("force", false, "Force remove running container")
	yesFlag := flags.Bool("yes", false, "Skip confirmation prompt")

	// Reorder args: move all flags (starting with -) before positional args
	// This allows: chauf podman remove <container> --force --yes
	var flagArgs, posArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
		} else {
			posArgs = append(posArgs, arg)
		}
	}
	reorderedArgs := append(flagArgs, posArgs...)
	if err := flags.Parse(reorderedArgs); err != nil {
		return err
	}

	verbose := lib.Verbose
	ctx := context.Background()
	client := podman.NewPodmanClient()
	reader := bufio.NewReader(os.Stdin)

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	engines, err := podman.ListEngines()
	if err != nil {
		return fmt.Errorf("list engines: %w", err)
	}

	if len(engines) == 0 {
		fmt.Println()
		lib.Info("No containers configured.")
		fmt.Println()
		return nil
	}

	target := ""
	if flags.NArg() > 0 {
		target = strings.ToLower(flags.Arg(0))
	}

	// Interactive mode if no target
	if target == "" {
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold("Select container to remove:"))
		fmt.Println()
		for i, e := range engines {
			c, _ := podman.Load(e)
			status := lib.Gray("unknown")
			if c != nil {
				container := podman.NewContainer(client, c)
				running, _ := container.IsRunning(ctx)
				if running {
					status = lib.Red("running")
				} else {
					status = lib.Gray("stopped")
				}
			}
			fmt.Printf("    %d) %-20s  %s (%s)\n", i+1, c.ContainerName, string(c.Engine), status)
		}
		fmt.Println()
		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(engines))+" or container name]: "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ReplaceAll(input, "\r", "")

		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(engines) {
			target = engines[idx-1]
		} else {
			target = strings.ToLower(input)
		}
	}

	containers, err := resolveContainers(target)
	if err != nil || len(containers) == 0 {
		lib.Warn("No containers found matching " + lib.Bold(target))
		return nil
	}

	for _, c := range containers {
		container := podman.NewContainer(client, c)

		// Check if container exists in podman
		exists, _ := client.ContainerExists(ctx, c.ContainerName)
		if !exists {
			// Orphaned config - container doesn't exist, just clean up config
			fmt.Println()
			lib.Warn(fmt.Sprintf("Container %s not found in podman ( orphaned config)", lib.Bold(c.ContainerName)))
			fmt.Println()
			if *yesFlag || interactiveConfirm("Remove orphaned config?", reader) {
				if err := podman.Delete(c.ContainerName); err != nil {
					lib.Error("Failed to remove config: " + err.Error())
				} else {
					fmt.Printf("  %s Config removed for %s\n", lib.Green("✓"), c.ContainerName)
				}
			} else {
				lib.Info("Cancelled.")
			}
			continue
		}

		// Check if running
		running, _ := container.IsRunning(ctx)

		// Confirm removal
		if !*yesFlag {
			fmt.Println()
			lib.Info(fmt.Sprintf("Remove container %s?", lib.Bold(c.ContainerName)))
			lib.Pair("  Engine", string(c.Engine))
			lib.Pair("  Container", c.ContainerName)
			lib.Pair("  Volume", c.VolumePath)
			fmt.Println()
			fmt.Print("  Confirm [y/N]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(resp)), "\r", "")
			if resp != "y" && resp != "yes" {
				lib.Info("Cancelled.")
				return nil
			}
		}

		// If running and not force, ask to stop first
		if running && !*forceFlag {
			fmt.Println()
			lib.Info(fmt.Sprintf("Container %s is running.", lib.Bold(c.ContainerName)))
			fmt.Print("  Stop it first? [y/N]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(resp)), "\r", "")
			if resp != "y" && resp != "yes" {
				lib.Info("Cancelled.")
				return nil
			}
			if err := container.Stop(ctx, 10*time.Second); err != nil {
				lib.Error("Failed to stop container: " + err.Error())
				return err
			}
			fmt.Printf("  %s Stopped %s\n", lib.Green("✓"), c.ContainerName)
		}

		if verbose {
			fmt.Printf("  Removing container %s...\n", c.ContainerName)
		}
		// If force flag was used, remove even if running
		removeForce := *forceFlag
		if err := container.Remove(ctx, removeForce); err != nil {
			return fmt.Errorf("remove container: %w", err)
		}

		// Clean up config after successful removal
		if err := podman.Delete(c.ContainerName); err != nil {
			lib.Warn("Container removed but failed to clean up config: " + err.Error())
		}

		// Remove config file
		configPath := podman.ConfigPath(c.ContainerName)
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			lib.Warn("Could not remove config: " + err.Error())
		} else if verbose {
			fmt.Printf("  Removed config: %s\n", configPath)
		}

		// Remove volume directory
		if c.VolumePath != "" {
			if verbose {
				fmt.Printf("  Removing volume: %s...\n", c.VolumePath)
			}
			if err := os.RemoveAll(c.VolumePath); err != nil {
				lib.Warn("Could not remove volume: " + err.Error())
			}
		}

		fmt.Println()
		lib.Success(fmt.Sprintf("Container %s removed", lib.Bold(c.ContainerName)))
	}

	fmt.Println()
	return nil
}

// ── chauf podman console ──────────────────────────────────────────────────────

func runPodmanConsole(args []string) error {
	flags := flag.NewFlagSet("podman console", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman console — attach to container for CLI access",
		"chauf podman console <name>")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() != 1 {
		flags.Usage()
		return flag.ErrHelp
	}

	ctx := context.Background()
	client := podman.NewPodmanClient()

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	engine := strings.ToLower(flags.Arg(0))

	cfg, err := podman.Load(engine)
	if err != nil {
		if err == podman.ErrConfigNotFound {
			lib.Error(fmt.Sprintf("No container found for %q", engine))
			return nil
		}
		return err
	}

	container := podman.NewContainer(client, cfg)

	// Check if running
	running, err := container.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		lib.Error(fmt.Sprintf("Container %s is not running", engine))
		lib.Info("Start it with:  chauf podman start " + engine)
		return nil
	}

	// Build the exec command based on engine
	var execArgs []string
	switch cfg.Engine {
	case podman.EngineMySQL8, podman.EngineMySQL57, podman.EngineMaria:
		execArgs = []string{"exec", "-it", cfg.ContainerName, "mysql", "-u", cfg.Username, "-p" + cfg.Password}
	case podman.EnginePostgres:
		execArgs = []string{"exec", "-it", cfg.ContainerName, "psql", "-U", cfg.Username, "-d", "app"}
	case podman.EngineMongo:
		execArgs = []string{"exec", "-it", cfg.ContainerName, "mongosh", "-u", cfg.Username, "-p", cfg.Password}
	case podman.EngineRedis:
		execArgs = []string{"exec", "-it", cfg.ContainerName, "redis-cli"}
	default:
		execArgs = []string{"exec", "-it", cfg.ContainerName, "/bin/sh"}
	}

	cmd := exec.CommandContext(ctx, "podman", execArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("console: %w", err)
	}

	return nil
}

// ── chauf podman import ──────────────────────────────────────────────────────

func runPodmanImport(args []string) error {
	flags := flag.NewFlagSet("podman import", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman import — adopt an existing container",
		"chauf podman import <engine> --name <container-name> [--user <user>] [--pass <pass>] [--port <port>]")

	nameFlag := flags.String("name", "", "Existing container name (required)")
	userFlag := flags.String("user", "", "Database username (auto-detected if not set)")
	passFlag := flags.String("pass", "", "Database password (auto-detected if not set)")
	portFlag := flags.Int("port", 0, "Host port (auto-detected from container if not set)")
	yesFlag := flags.Bool("yes", false, "Skip confirmation if config already exists")

	if err := flags.Parse(args); err != nil {
		return err
	}

	// Engine argument required
	if flags.NArg() != 1 {
		flags.Usage()
		return flag.ErrHelp
	}
	engineStr := strings.ToLower(flags.Arg(0))
	if !podman.IsValidEngine(engineStr) {
		lib.Error(fmt.Sprintf("unsupported engine %q", flags.Arg(0)))
		lib.Info(lib.Gray("Supported: mysql8, mysql57, postgres, maria, mongo, redis"))
		return nil
	}

	// Container name required
	if *nameFlag == "" {
		lib.Error("--name <container-name> is required")
		return nil
	}

	ctx := context.Background()
	client := podman.NewPodmanClient()

	// Check podman availability
	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	// Check container exists in podman
	exists, err := client.ContainerExists(ctx, *nameFlag)
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !exists {
		lib.Error(fmt.Sprintf("Container %q does not exist in podman", *nameFlag))
		return nil
	}

	// Get container info via podman inspect
	inspectOut, err := client.Run(ctx, "container", "inspect", "--format",
		"{{.Config.Image}}|{{.HostConfig.PortBindings}}", *nameFlag)
	if err != nil {
		return fmt.Errorf("inspect container: %w", err)
	}

	// Parse port from inspect output (format: "0.0.0.0:3306->3306/tcp" or similar)
	port := *portFlag
	if port == 0 {
		// Try to extract port from port bindings
		parts := strings.Split(inspectOut, "|")
		if len(parts) >= 2 {
			portBindings := parts[1]
			// Look for host port in format like "0.0.0.0:3306->3306/tcp"
			if strings.Contains(portBindings, "0.0.0.0:") {
				idx := strings.Index(portBindings, "0.0.0.0:")
				rest := portBindings[idx+len("0.0.0.0:"):]
				for i, c := range rest {
					if c < '0' || c > '9' {
						portStr := rest[:i]
						fmt.Sscanf(portStr, "%d", &port)
						break
					}
				}
			}
		}
		// Fall back to engine default
		if port == 0 {
			engine := podman.EngineType(engineStr)
			defaults := map[podman.EngineType]int{
				podman.EngineMySQL8:   3306,
				podman.EngineMySQL57:  3307,
				podman.EnginePostgres: 5432,
				podman.EngineMaria:    3306,
				podman.EngineMongo:    27017,
				podman.EngineRedis:    6379,
			}
			port = defaults[engine]
		}
	}

	// Create config
	cfg := &podman.DatabaseConfig{
		Name:          engineStr,
		Engine:        podman.EngineType(engineStr),
		ContainerName: *nameFlag,
		Port:          port,
		VolumePath:    filepath.Join(podman.Root(), "volumes", *nameFlag),
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	// Auto-detect image
	if inspectOut != "" {
		parts := strings.Split(inspectOut, "|")
		if len(parts) >= 1 && parts[0] != "" {
			cfg.Image = parts[0]
		}
	}

	// Set defaults for credentials if not provided
	if *userFlag != "" {
		cfg.Username = *userFlag
	} else {
		cfg.Username = "chauf"
	}
	if *passFlag != "" {
		cfg.Password = *passFlag
	} else {
		cfg.Password = podman.GeneratePassword()
	}

	// Check if config already exists
	existingCfg, err := podman.Load(engineStr)
	if err == nil && existingCfg != nil {
		if !*yesFlag {
			fmt.Println()
			lib.Info(fmt.Sprintf("Config for %s already exists (%s).", lib.Bold(engineStr), podman.ConfigPath(existingCfg.ContainerName)))
			if !interactiveConfirm("Overwrite it?", nil) {
				lib.Info("Cancelled.")
				return nil
			}
		}
	}

	// Save config
	if err := podman.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Imported container %s as %s", lib.Bold(*nameFlag), lib.Bold(engineStr)))
	fmt.Println()
	lib.Pair("  Engine", string(cfg.Engine))
	lib.Pair("  Container", cfg.ContainerName)
	lib.Pair("  Image", cfg.Image)
	lib.Pair("  Port", fmt.Sprintf("%d", cfg.Port))
	fmt.Println()
	lib.Info("Note: Update username/password if they differ from the defaults.")
	fmt.Println()

	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// resolveContainers returns configs for containers matching the target.
// Target can be "all", an engine name (mysql8, postgres, etc), a container name,
// or a comma-separated list of engine names or container names.
func resolveContainers(target string) ([]*podman.DatabaseConfig, error) {
	if target == "all" {
		// Return all configs
		engines, err := podman.ListEngines()
		if err != nil {
			return nil, err
		}
		var configs []*podman.DatabaseConfig
		for _, e := range engines {
			cfg, err := podman.Load(e)
			if err != nil {
				continue
			}
			configs = append(configs, cfg)
		}
		return configs, nil
	}

	// Handle comma-separated targets (from multi-select)
	parts := strings.Split(target, ",")
	var configs []*podman.DatabaseConfig
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try as engine name first
		if podman.IsValidEngine(part) {
			cfg, err := podman.Load(part)
			if err == nil && cfg != nil {
				configs = append(configs, cfg)
				continue
			}
		}

		// Try as container name
		cfg, err := podman.Load(part)
		if err == nil && cfg != nil {
			configs = append(configs, cfg)
			continue
		}
	}

	if len(configs) == 0 {
		return nil, nil
	}
	return configs, nil
}

// isDirWritable checks if a directory is writable by attempting to create a temp file.
func isDirWritable(dir string) bool {
	testFile := dir + "/.chauf-write-test"
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

// interactiveSelectEngine prompts the user to select a database engine.
func interactiveSelectEngine() string {
	engines := []struct {
		key  string
		name string
		desc string
	}{
		{"1", "mysql8", "MySQL 8.0 — port 3306"},
		{"2", "mysql57", "MySQL 5.7 — port 3307"},
		{"3", "postgres", "PostgreSQL 16 — port 5432"},
		{"4", "maria", "MariaDB 11 — port 3306"},
		{"5", "mongo", "MongoDB 7 — port 27017"},
		{"6", "redis", "Redis 7 (Alpine) — port 6379"},
	}

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Create a database container"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Select engine:"))
	fmt.Println()
	for _, e := range engines {
		fmt.Printf("    %s  %-10s  %s\n", lib.Gray(e.key+")"), e.name, lib.Gray(e.desc))
	}
	fmt.Println()

	for {
		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-6]") + ": ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		for _, e := range engines {
			if input == e.key {
				return e.name
			}
		}

		// Try by name
		if podman.IsValidEngine(input) {
			return input
		}

		lib.Warn(fmt.Sprintf("Invalid choice %q — enter 1-%d or an engine name", input, len(engines)))
	}
}

// promptField shows a prompt with a default value and lets user override.
func promptField(label string, value *string, defaultVal string, required bool) {
	*value = defaultVal
	fmt.Println()
	for {
		fmt.Printf("  %s%s %s(%s)%s: ",
			lib.Bold(label+":"),
			requiredText(required),
			lib.Gray("[default: "),
			*value,
			lib.Gray("]"))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			return // keep default
		}
		if required && input == "" {
			lib.Warn(fmt.Sprintf("%s is required", label))
			continue
		}
		*value = input
		return
	}
}

// promptFieldInt shows a prompt with an integer default and lets user override.
func promptFieldInt(label string, value *int, defaultVal int, required bool) {
	*value = defaultVal
	fmt.Println()
	for {
		fmt.Printf("  %s%s %s(%d)%s: ",
			lib.Bold(label+":"),
			requiredText(required),
			lib.Gray("[default: "),
			*value,
			lib.Gray("]"))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			return // keep default
		}
		if _, err := fmt.Sscanf(input, "%d", value); err != nil {
			lib.Warn(fmt.Sprintf("Invalid number: %q", input))
			continue
		}
		return
	}
}

func requiredText(required bool) string {
	if required {
		return ""
	}
	return lib.Gray(" (optional)")
}

// interactiveConfirm asks the user a yes/no question. Returns true on y/yes.
// If reader is nil, creates its own reader.
func interactiveConfirm(prompt string, reader *bufio.Reader) bool {
	fmt.Print("  " + prompt + " " + lib.Gray("[y/N]") + ": ")
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	resp, _ := reader.ReadString('\n')
	resp = strings.ReplaceAll(strings.TrimSpace(resp), "\r", "")
	return strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes"
}

// interactiveSelectDatabases shows an interactive list with tick marks for selection.
// Uses bubbletea for proper TUI rendering.
func interactiveSelectDatabases(databases []string) []string {
	model := multiSelectModel{
		choices:  databases,
		selected: make(map[int]bool),
		cursor:   0,
		title:    "Select databases to backup",
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		// Fall back to simple text input if TTY is not available
		return interactiveSelectDatabasesSimple(databases)
	}

	if finalModel, ok := result.(multiSelectModel); ok {
		if finalModel.aborted {
			return []string{}
		}

		var selected []string
		for i, sel := range finalModel.selected {
			if sel {
				selected = append(selected, finalModel.choices[i])
			}
		}
		return selected
	}

	return []string{}
}

// multiSelectModel is the bubbletea model for database selection
type multiSelectModel struct {
	choices  []string
	selected map[int]bool
	cursor   int
	width    int
	height   int
	done     bool
	aborted  bool
	title    string
}

// Init initializes the model
func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case tea.KeySpace:
			m.selected[m.cursor] = !m.selected[m.cursor]
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m multiSelectModel) View() string {
	const reservedLines = 8
	visibleChoices := len(m.choices)
	if m.height > reservedLines {
		visibleChoices = m.height - reservedLines
	}
	if visibleChoices < 1 {
		visibleChoices = 1
	}

	start := 0
	if len(m.choices) > visibleChoices {
		if m.cursor >= visibleChoices {
			start = m.cursor - visibleChoices + 1
		}
		maxStart := len(m.choices) - visibleChoices
		if start > maxStart {
			start = maxStart
		}
	}
	end := start + visibleChoices
	if end > len(m.choices) {
		end = len(m.choices)
	}

	// Count selections
	selectedCount := 0
	for _, sel := range m.selected {
		if sel {
			selectedCount++
		}
	}

	var lines []string

	// Title
	lines = append(lines, fmt.Sprintf("  %s", lib.Bold(m.title)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s", lib.Gray("[↑/↓] Move  [space] Select  [enter] Confirm")))
	lines = append(lines, "")
	if start > 0 {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↑ more")))
	}

	// Choices
	for i := start; i < end; i++ {
		choice := m.choices[i]
		cursor := "  "
		if m.cursor == i {
			cursor = lib.Green(">")
		}

		selected := "[ ]"
		if m.selected[i] {
			selected = lib.Green("[x]")
		}

		// Truncate long names
		displayName := choice
		if len(displayName) > 50 {
			displayName = displayName[:47] + "..."
		}

		lines = append(lines, fmt.Sprintf("  %s %s %s", cursor, selected, displayName))
	}
	if end < len(m.choices) {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↓ more")))
	}

	lines = append(lines, "")

	// Footer with selection count
	if selectedCount == 0 {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("No databases selected")))
	} else if selectedCount == len(m.choices) {
		lines = append(lines, fmt.Sprintf("  %s %s", lib.Green("✓"), lib.Gray(fmt.Sprintf("All %d selected", selectedCount))))
	} else {
		lines = append(lines, fmt.Sprintf("  %s %d of %d selected", lib.Green("✓"), selectedCount, len(m.choices)))
	}

	return strings.Join(lines, "\n")
}

// interactiveSelectSingle shows an interactive single-item selector with arrow keys.
// Uses bubbletea for proper TUI rendering.
const (
	singleSelectReservedLines    = 4
	containerSelectReservedLines = 6
)

func interactiveSelectSingle(items []string, title string) string {
	if len(items) == 0 {
		return ""
	}
	model := singleSelectModel{
		items:  items,
		cursor: 0,
		title:  title,
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return interactiveSelectSingleSimple(items, title)
	}

	if finalModel, ok := result.(singleSelectModel); ok {
		if finalModel.aborted {
			return ""
		}
		return finalModel.items[finalModel.cursor]
	}

	return ""
}

// singleSelectModel is the bubbletea model for single item selection
type singleSelectModel struct {
	items   []string
	cursor  int
	done    bool
	aborted bool
	title   string
	height  int
}

func (m singleSelectModel) Init() tea.Cmd {
	return nil
}

func (m singleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case tea.KeySpace, tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m singleSelectModel) View() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("  %s", lib.Bold(m.title)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s", lib.Gray("[↑/↓] Move  [space/enter] Select")))
	lines = append(lines, "")

	start, end, showAbove, showBelow := selectorViewportWithIndicators(len(m.items), m.cursor, m.height, singleSelectReservedLines)
	if showAbove {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↑ more")))
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		cursor := "  "
		if m.cursor == i {
			cursor = lib.Green(">")
		}

		lines = append(lines, fmt.Sprintf("  %s %s", cursor, item))
	}

	if showBelow {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↓ more")))
	}

	return strings.Join(lines, "\n")
}

// interactiveSelectSingleSimple is the fallback text-based single selector when TTY is not available.
func interactiveSelectSingleSimple(items []string, title string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold(title))
	fmt.Println()

	for i, item := range items {
		fmt.Printf("    %d) %s\n", i+1, item)
	}
	fmt.Println()

	fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(items))+" or name]: "))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\r", "")

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(items) {
		return items[idx-1]
	}

	inputLower := strings.ToLower(input)
	for _, item := range items {
		if strings.ToLower(item) == inputLower {
			return item
		}
	}

	return inputLower
}

// interactiveSelectDatabasesSimple is the fallback text-based selector when TTY is not available.
func interactiveSelectDatabasesSimple(databases []string) []string {
	reader := bufio.NewReader(os.Stdin)

	selected := make(map[int]bool)

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Select databases to backup:"))
	fmt.Println()
	fmt.Printf("    %s  %s\n", lib.Gray("[a]"), lib.Gray("Toggle all"))
	fmt.Printf("    %s  %s\n", lib.Gray("[space]"), lib.Gray("Toggle selected"))
	fmt.Printf("    %s  %s\n", lib.Gray("[enter]"), lib.Gray("Confirm selection"))
	fmt.Println()

	for {
		// Display current selection state
		fmt.Print("  ")
		for i, db := range databases {
			tick := " "
			if selected[i] {
				tick = "✓"
			}
			fmt.Printf("[%s] %d) %s  ", tick, i+1, db)
			if (i+1)%3 == 0 {
				fmt.Println()
				fmt.Print("  ")
			}
		}
		if len(databases)%3 != 0 {
			fmt.Println()
		}
		fmt.Println()

		// Count selected
		selCount := 0
		for _, v := range selected {
			if v {
				selCount++
			}
		}

		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(databases))+", a=all, enter=done]: "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		if input == "d" || input == "done" || input == "" {
			if selCount > 0 {
				break
			}
			continue
		}

		if input == "a" || input == "all" {
			// Toggle all
			if selCount == len(databases) {
				for i := range selected {
					selected[i] = false
				}
			} else {
				for i := range selected {
					selected[i] = true
				}
			}
			continue
		}

		// Try to parse as number
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(databases) {
			selected[idx-1] = !selected[idx-1]
		}
	}

	// Collect selected databases
	var result []string
	for i, db := range databases {
		if selected[i] {
			result = append(result, db)
		}
	}

	return result
}

// containerSelectModel is the bubbletea model for container selection
type containerSelectModel struct {
	containers []containerInfo
	selected   map[int]bool
	cursor     int
	done       bool
	aborted    bool
	title      string
	height     int
}

// containerInfo holds container display information
type containerInfo struct {
	Name     string
	Engine   string
	Status   string
	Disabled bool
}

// Init initializes the model
func (m containerSelectModel) Init() tea.Cmd {
	return nil
}

func (m containerSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
			for m.cursor > 0 && m.containers[m.cursor].Disabled {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
			for m.cursor < len(m.containers)-1 && m.containers[m.cursor].Disabled {
				m.cursor++
			}
		case tea.KeySpace:
			if !m.containers[m.cursor].Disabled {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case tea.KeyEnter:
			if len(m.selected) == 0 && len(m.containers) > 0 && !m.containers[m.cursor].Disabled {
				m.selected[m.cursor] = true
			}
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m containerSelectModel) View() string {
	selectedCount := 0
	for _, sel := range m.selected {
		if sel {
			selectedCount++
		}
	}

	var lines []string

	lines = append(lines, fmt.Sprintf("  %s", lib.Bold(m.title)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s", lib.Gray("[↑/↓] Move  [space] Select  [enter] Confirm")))

	if len(m.containers) == 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("No containers available")))
		lines = append(lines, "")
		lines = append(lines, "  Press Ctrl+C to cancel")
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")

	start, end, showAbove, showBelow := selectorViewportWithIndicators(len(m.containers), m.cursor, m.height, containerSelectReservedLines)
	if showAbove {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↑ more")))
	}

	for i := start; i < end; i++ {
		c := m.containers[i]
		cursor := "  "
		if m.cursor == i {
			cursor = lib.Green(">")
		}

		sel := "[ ]"
		if m.selected[i] {
			sel = lib.Green("[x]")
		}

		prefix := "  "
		if c.Disabled {
			prefix = lib.Gray("- ")
			sel = lib.Gray("[ ]")
			lines = append(lines, fmt.Sprintf("  %s%s%-20s  %s (%s)", prefix, sel, c.Name, lib.Gray(c.Engine), lib.Gray(c.Status)))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s %-20s  %s (%s)", cursor, sel, c.Name, c.Engine, c.Status))
		}
	}

	if showBelow {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↓ more")))
	}

	lines = append(lines, "")

	if selectedCount == 0 {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("No containers selected")))
	} else if selectedCount == len(m.containers) {
		lines = append(lines, fmt.Sprintf("  %s %s", lib.Green("✓"), lib.Gray(fmt.Sprintf("All %d selected", selectedCount))))
	} else {
		lines = append(lines, fmt.Sprintf("  %s %d of %d selected", lib.Green("✓"), selectedCount, len(m.containers)))
	}

	return strings.Join(lines, "\n")
}

func selectorViewportWithIndicators(totalRows, activeRow, terminalHeight, baseReservedLines int) (start, end int, showAbove, showBelow bool) {
	start, end, showAbove, showBelow = computeSelectorViewport(totalRows, activeRow, terminalHeight, baseReservedLines)

	indicatorLines := 0
	if showAbove {
		indicatorLines++
	}
	if showBelow {
		indicatorLines++
	}

	if indicatorLines == 0 {
		return start, end, showAbove, showBelow
	}

	return computeSelectorViewport(totalRows, activeRow, terminalHeight, baseReservedLines+indicatorLines)
}

// interactiveSelectContainers shows an interactive list of containers with selection.
// Returns list of selected container names (engine keys).
func interactiveSelectContainers(containers []containerInfo, title string) []string {
	if len(containers) == 0 {
		return []string{}
	}

	model := containerSelectModel{
		containers: containers,
		selected:   make(map[int]bool),
		cursor:     0,
		title:      title,
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return interactiveSelectContainersSimple(containers, title)
	}

	if finalModel, ok := result.(containerSelectModel); ok {
		if finalModel.aborted {
			return []string{}
		}

		var selected []string
		for i, sel := range finalModel.selected {
			if sel {
				selected = append(selected, finalModel.containers[i].Name)
			}
		}
		return selected
	}

	return []string{}
}

// interactiveSelectContainersSimple is the fallback text-based selector when TTY is not available.
func interactiveSelectContainersSimple(containers []containerInfo, title string) []string {
	reader := bufio.NewReader(os.Stdin)

	selected := make(map[int]bool)

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold(title))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("[a] Toggle all  [1-N] Toggle item"))
	fmt.Println()

	for {
		fmt.Print("  ")
		for i, c := range containers {
			tick := " "
			if selected[i] {
				tick = "✓"
			}
			if c.Disabled {
				fmt.Printf("%s %d) %s (%s)  ", lib.Gray("x"), i+1, c.Name, c.Status)
			} else {
				fmt.Printf("[%s] %d) %s (%s)  ", tick, i+1, c.Name, c.Status)
			}
			if (i+1)%2 == 0 {
				fmt.Println()
				fmt.Print("  ")
			}
		}
		if len(containers)%2 != 0 {
			fmt.Println()
		}
		fmt.Println()

		selCount := 0
		for _, v := range selected {
			if v {
				selCount++
			}
		}

		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(containers))+", a=all, enter=done]: "))
		input, _ := reader.ReadString('\n')
		input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")

		if input == "d" || input == "done" || input == "" {
			if selCount > 0 {
				break
			}
			continue
		}

		if input == "a" || input == "all" {
			if selCount == len(containers) {
				for i := range selected {
					selected[i] = false
				}
			} else {
				for i := range selected {
					if !containers[i].Disabled {
						selected[i] = true
					}
				}
			}
			continue
		}

		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(containers) {
			if !containers[idx-1].Disabled {
				selected[idx-1] = !selected[idx-1]
			}
		}
	}

	var result []string
	for i, c := range containers {
		if selected[i] && !c.Disabled {
			result = append(result, c.Name)
		}
	}

	return result
}

// verboseLogger implements podman.Logger for verbose output.
type verboseLogger struct {
	verbose bool
}

func (l *verboseLogger) Print(args ...interface{}) {
	if l.verbose {
		fmt.Println(args...)
	}
}

// ── chauf podman backup ───────────────────────────────────────────────────────

func runPodmanBackup(args []string) error {
	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Println()
			fmt.Printf("  %s\n", lib.Bold("chauf podman backup"))
			fmt.Println()
			fmt.Printf("  %s\n", lib.Gray("Interactively backup databases from a running container."))
			fmt.Printf("  %s\n", lib.Gray("Shows running containers, lets you select databases,"))
			fmt.Printf("  %s\n", lib.Gray("optionally add descriptions, and creates backup files."))
			fmt.Println()
			return nil
		}
	}

	// 1. Setup (client, ctx, verbose)
	ctx := context.Background()
	client := podman.NewPodmanClient()
	reader := bufio.NewReader(os.Stdin)

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	// 2. Container selection (show running containers only)
	engines, _ := podman.ListEngines()
	runningContainers := filterRunningContainers(client, engines, ctx)
	if len(runningContainers) == 0 {
		lib.Info("No running containers. Start a container first.")
		return nil
	}

	target := interactiveSelectContainer(runningContainers, ctx, client, "Select container to backup:")
	if target == "" {
		return nil // User cancelled
	}

	// 3. Load container config
	cfg, err := podman.Load(target)
	if err != nil {
		lib.Error("Could not load config: " + err.Error())
		return nil
	}
	container := podman.NewContainer(client, cfg)

	// 4. Test connection, prompt for credentials if needed
	databases, err := container.ListDatabases(ctx)
	if err != nil {
		databases, err = promptCredentialsLoop(container, ctx)
		if err != nil {
			lib.Error("Could not connect: " + err.Error())
			return nil
		}
	}

	if len(databases) == 0 {
		lib.Warn("No databases found in container.")
		return nil
	}

	// 5. Database selection (TTY interactive)
	selected := interactiveSelectDatabases(databases)
	if len(selected) == 0 {
		lib.Info("No databases selected.")
		return nil
	}

	// 6. Ask if user wants to add descriptions
	descriptions := make(map[string]string)
	fmt.Println()
	fmt.Print("  " + lib.Bold("Add descriptions?") + " " + lib.Gray("[y/N]: "))
	input, _ := reader.ReadString('\n')
	input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
	addDescriptions := strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"

	if addDescriptions {
		for _, db := range selected {
			desc := promptDescription(db)
			descriptions[db] = desc
		}
	}

	// 7. Backup summary & confirmation
	if !confirmBackup(cfg.ContainerName, selected, descriptions) {
		return nil
	}

	// 8. Execute backups
	timestamp := time.Now().Format("20060102-150405")
	backupDir := workspace.Path("podman", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		lib.Error("Could not create backup directory: " + err.Error())
		return err
	}

	successCount := 0
	for _, db := range selected {
		backupData, err := container.BackupDatabaseWithDescription(ctx, db, descriptions[db])
		if err != nil {
			lib.Warn(fmt.Sprintf("Backup failed for %s: %v", db, err))
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar.gz", cfg.ContainerName, db, timestamp)
		outputPath := filepath.Join(backupDir, filename)
		if err := os.WriteFile(outputPath, backupData, 0644); err != nil {
			lib.Warn(fmt.Sprintf("Failed to save %s: %v", outputPath, err))
			continue
		}

		fmt.Printf("  %s %s (%d bytes)\n", lib.Green("✓"), filename, len(backupData))
		successCount++
	}

	fmt.Println()
	if successCount == len(selected) {
		lib.Success(fmt.Sprintf("Backed up %d database(s) to %s", successCount, backupDir))
	} else {
		lib.Warn(fmt.Sprintf("Backed up %d/%d database(s)", successCount, len(selected)))
	}

	return nil
}

// runningContainer holds a container config with its lookup key
type runningContainer struct {
	Key    string
	Config *podman.DatabaseConfig
}

// filterRunningContainers returns only running containers with their config keys
func filterRunningContainers(client *podman.PodmanClient, engines []string, ctx context.Context) []runningContainer {
	var running []runningContainer
	for _, e := range engines {
		cfg, err := podman.Load(e)
		if err != nil || cfg == nil {
			continue
		}
		container := podman.NewContainer(client, cfg)
		runningNow, _ := container.IsRunning(ctx)
		if runningNow {
			running = append(running, runningContainer{Key: e, Config: cfg})
		}
	}
	return running
}

// interactiveSelectContainer shows running containers and returns selected container name.
// Uses bubbletea TUI for arrow key selection.
func interactiveSelectContainer(containers []runningContainer, ctx context.Context, client *podman.PodmanClient, title string) string {
	if len(containers) == 0 {
		return ""
	}

	items := make([]string, len(containers))
	for i, c := range containers {
		items[i] = fmt.Sprintf("%-20s  %s (%s)", c.Config.ContainerName, string(c.Config.Engine), lib.Green("running"))
	}

	p := tea.NewProgram(singleSelectModel{
		items:  items,
		cursor: 0,
		title:  title,
	})
	result, err := p.Run()
	if err != nil {
		return interactiveSelectContainerSimple(containers, title)
	}

	if finalModel, ok := result.(singleSelectModel); ok {
		if finalModel.aborted {
			return ""
		}
		return containers[finalModel.cursor].Key
	}

	return ""
}

// interactiveSelectContainerSimple is the fallback text-based container selector when TTY is not available.
func interactiveSelectContainerSimple(containers []runningContainer, title string) string {
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold(title))
	fmt.Println()

	for i, c := range containers {
		status := lib.Green("running")
		fmt.Printf("    %d) %-20s  %s (%s)\n", i+1, c.Config.ContainerName, string(c.Config.Engine), status)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(containers))+" or container name]: "))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\r", "")

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(containers) {
		return containers[idx-1].Key
	}

	inputLower := strings.ToLower(input)
	for _, c := range containers {
		if strings.ToLower(c.Config.ContainerName) == inputLower || strings.ToLower(string(c.Config.Engine)) == inputLower {
			return c.Key
		}
	}

	return inputLower
}

// promptCredentialsLoop asks for user/pass and retries connection
func promptCredentialsLoop(container *podman.Container, ctx context.Context) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get current config for defaults
	cfg := container.Config()
	username := cfg.Username
	password := cfg.Password

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Connection failed. Enter credentials:"))
	fmt.Println()
	fmt.Printf("  %s %s%s%s\n", lib.Gray("Username"), lib.Gray("["), lib.Gray(username), lib.Gray("]"))
	fmt.Printf("  %s\n", lib.Gray("Password [********]"))
	fmt.Println()

	for {
		fmt.Print("  " + lib.Bold("Username") + " " + lib.Gray("[default: "+username+"]: "))
		input, _ := reader.ReadString('\n')
		input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
		if input != "" {
			username = input
		}

		fmt.Print("  " + lib.Bold("Password") + " " + lib.Gray("[********]: "))
		input, _ = reader.ReadString('\n')
		input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
		if input != "" {
			password = input
		}

		// Update container config with new credentials
		cfg.Username = username
		cfg.Password = password

		// Retry connection
		databases, err := container.ListDatabases(ctx)
		if err == nil {
			return databases, nil
		}

		fmt.Println()
		lib.Warn(fmt.Sprintf("Connection failed: %v", err))
		fmt.Println()
		fmt.Print("  " + lib.Bold("Try again? [Y/n]: "))
		input, _ = reader.ReadString('\n')
		input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
		if input == "n" || input == "N" {
			return nil, fmt.Errorf("cancelled")
		}
	}
}

// promptDescription asks for a description for a database backup
func promptDescription(database string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Printf("  %s %s %s\n", lib.Bold("Description for"), lib.Green(`"`+database+`"`), lib.Bold(" (optional, press enter to skip):"))
	fmt.Print("  > ")
	input, _ := reader.ReadString('\n')
	input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
	return input
}

// confirmBackup shows summary and asks for confirmation
func confirmBackup(containerName string, databases []string, descriptions map[string]string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Backup summary:"))
	fmt.Printf("    Container: %s\n", containerName)
	fmt.Printf("    Databases: %s\n", strings.Join(databases, ", "))

	// Show descriptions if any
	hasDescriptions := false
	for _, db := range databases {
		if descriptions[db] != "" {
			hasDescriptions = true
			break
		}
	}
	if hasDescriptions {
		fmt.Println()
		fmt.Printf("    %s\n", lib.Bold("Descriptions:"))
		for _, db := range databases {
			if descriptions[db] != "" {
				fmt.Printf("      %s: %s\n", db, descriptions[db])
			}
		}
	}
	fmt.Println()

	fmt.Print("  " + lib.Bold("Proceed? [Y/n]: "))
	input, _ := reader.ReadString('\n')
	input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
	if input == "n" || input == "N" {
		return false
	}
	return true
}

// ── chauf podman restore ──────────────────────────────────────────────────────

func runPodmanRestore(args []string) error {
	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Println()
			fmt.Printf("  %s\n", lib.Bold("chauf podman restore"))
			fmt.Println()
			fmt.Printf("  %s\n", lib.Gray("Interactively restore a database from a backup file."))
			fmt.Printf("  %s\n", lib.Gray("Shows running containers, lists available backups,"))
			fmt.Printf("  %s\n", lib.Gray("lets you select the database and backup to restore."))
			fmt.Println()
			return nil
		}
	}

	// 1. Setup (client, ctx, reader)
	ctx := context.Background()
	client := podman.NewPodmanClient()
	reader := bufio.NewReader(os.Stdin)

	if err := client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			lib.Error("Podman is not installed or not in PATH.")
			return nil
		}
		return err
	}

	// 2. Container selection (show running containers only)
	engines, _ := podman.ListEngines()
	runningContainers := filterRunningContainers(client, engines, ctx)
	if len(runningContainers) == 0 {
		lib.Info("No running containers. Start a container first.")
		return nil
	}

	target := interactiveSelectContainer(runningContainers, ctx, client, "Select container to restore:")
	if target == "" {
		return nil // User cancelled
	}

	// 3. Load container config
	cfg, err := podman.Load(target)
	if err != nil {
		lib.Error("Could not load config: " + err.Error())
		return nil
	}
	container := podman.NewContainer(client, cfg)

	// 4. Test connection, prompt for credentials if needed
	_, err = container.ListDatabases(ctx)
	if err != nil {
		_, err = promptCredentialsLoop(container, ctx)
		if err != nil {
			lib.Error("Could not connect: " + err.Error())
			return nil
		}
	}

	// 5. List available backups for this container
	backups, err := listBackupsForContainer(cfg.ContainerName)
	if err != nil {
		lib.Warn("Could not list backups: " + err.Error())
		return nil
	}

	if len(backups) == 0 {
		lib.Info("No backup files found for " + lib.Bold(cfg.ContainerName))
		return nil
	}

	// Group backups by database
	backupsByDB := groupBackupsByDatabase(backups)

	// Show database and backup count
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Available backups for "+cfg.ContainerName+":"))
	fmt.Println()
	for db, files := range backupsByDB {
		fmt.Printf("    %-20s  %d backup(s)\n", lib.Green(db), len(files))
	}
	fmt.Println()

	// 6. Select databases to restore (multi-select)
	dbNames := make([]string, 0, len(backupsByDB))
	for db := range backupsByDB {
		dbNames = append(dbNames, db)
	}
	sort.Strings(dbNames)

	selectedDBs := interactiveSelectDatabases(dbNames)
	if len(selectedDBs) == 0 {
		lib.Info("No databases selected.")
		return nil
	}

	// 7. For each selected database, select which backup to restore
	type restoreSelection struct {
		database string
		backup   BackupFile
	}
	var selections []restoreSelection
	for _, db := range selectedDBs {
		availableBackups := backupsByDB[db]
		backupItems := make([]string, len(availableBackups))
		backupIndex := make(map[int]BackupFile, len(availableBackups))
		for i, b := range availableBackups {
			desc := b.Description
			if desc == "" {
				desc = "-"
			}
			backupItems[i] = fmt.Sprintf("%s  %s", b.FormattedTime(), desc)
			backupIndex[i] = b
		}

		selectedIdxStr := interactiveSelectSingle(backupItems, "Select backup for "+lib.Green(db)+":")
		selectedIdx := 0
		for i, item := range backupItems {
			if item == selectedIdxStr {
				selectedIdx = i
				break
			}
		}
		selections = append(selections, restoreSelection{
			database: db,
			backup:   backupIndex[selectedIdx],
		})
	}

	// 8. Summary
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Restore summary:"))
	fmt.Printf("    Container:  %s\n", cfg.ContainerName)
	fmt.Printf("    Databases:  %d\n", len(selections))
	for _, sel := range selections {
		fmt.Printf("      - %s: %s (%s)\n", sel.database, sel.backup.Filename(), sel.backup.FormattedSize())
	}
	fmt.Println()

	fmt.Print("  " + lib.Bold("Proceed? [y/N]: "))
	input, _ := reader.ReadString('\n')
	input = strings.ReplaceAll(strings.TrimSpace(input), "\r", "")
	if input != "y" && input != "Y" && input != "yes" {
		lib.Info("Cancelled.")
		return nil
	}

	// 9. Execute restores with temp backup safety
	fmt.Println()
	for _, sel := range selections {
		fmt.Printf("  %s Restoring %s...\n", lib.Bold("→"), sel.database)

		// Step 1: Create temp backup of current database
		fmt.Printf("    %s Creating temp backup...\n", lib.Gray("○"))
		tempBackup, err := container.BackupDatabase(ctx, sel.database)
		if err != nil {
			fmt.Printf("    %s Could not create temp backup: %v (proceeding anyway)\n", lib.Gray("⚠"), err)
			tempBackup = nil
		}

		// Save temp backup to a file in case we need it
		var tempBackupPath string
		if tempBackup != nil {
			tempBackupPath = filepath.Join(workspace.Path("podman", "backups"), fmt.Sprintf("%s-%s-temp-%d.tar.gz",
				cfg.ContainerName, sel.database, time.Now().Unix()))
			if err := os.WriteFile(tempBackupPath, tempBackup, 0644); err != nil {
				fmt.Printf("    %s Could not save temp backup: %v\n", lib.Gray("⚠"), err)
				tempBackup = nil
				tempBackupPath = ""
			} else {
				fmt.Printf("    %s Temp backup saved\n", lib.Gray("○"))
			}
		}

		// Step 2: Drop the database
		if err := container.DropDatabase(ctx, sel.database); err != nil {
			fmt.Printf("    %s Could not drop database: %v\n", lib.Gray("⚠"), err)
		}

		// Step 3: Create the database fresh
		if err := container.CreateDatabase(ctx, sel.database); err != nil {
			fmt.Printf("    %s Could not create database: %v\n", lib.Gray("⚠"), err)
		}

		// Step 4: Restore from selected backup
		backupData, err := os.ReadFile(sel.backup.Path)
		if err != nil {
			fmt.Printf("  %s Failed to read backup: %v\n", lib.Red("✗"), err)
			if tempBackup != nil && tempBackupPath != "" {
				restoreFromTemp(ctx, container, cfg, sel.database, tempBackup, tempBackupPath)
			}
			continue
		}

		if err := container.Restore(ctx, backupData); err != nil {
			fmt.Printf("  %s Restore failed: %v\n", lib.Red("✗"), err)
			if tempBackup != nil && tempBackupPath != "" {
				restoreFromTemp(ctx, container, cfg, sel.database, tempBackup, tempBackupPath)
			}
			continue
		}

		// Step 5: Success - clean up temp backup
		if tempBackupPath != "" {
			os.Remove(tempBackupPath)
		}

		fmt.Printf("  %s %s restored successfully\n", lib.Green("✓"), sel.database)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Restored %d database(s)", len(selections)))

	// Show DSN
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Connection string:"))
	fmt.Printf("    %s\n", podman.DSN(cfg))
	fmt.Println()

	return nil
}

// BackupFile represents a backup file with its metadata
type BackupFile struct {
	Path        string
	Meta        podman.BackupMeta
	Description string
	Size        int64
}

func (b BackupFile) Filename() string {
	parts := strings.Split(b.Path, "/")
	return parts[len(parts)-1]
}

func (b BackupFile) FormattedTime() string {
	return b.Meta.Timestamp.Format("2006-01-02 15:04:05")
}

func (b BackupFile) FormattedSize() string {
	const unit = 1024
	if b.Size < unit {
		return fmt.Sprintf("%d B", b.Size)
	}
	if b.Size < unit*unit {
		return fmt.Sprintf("%.1f KB", float64(b.Size)/unit)
	}
	if b.Size < unit*unit*unit {
		return fmt.Sprintf("%.1f MB", float64(b.Size)/(unit*unit))
	}
	return fmt.Sprintf("%.1f GB", float64(b.Size)/(unit*unit*unit))
}

// listBackupsForContainer finds all backup files for a given container
func listBackupsForContainer(containerName string) ([]BackupFile, error) {
	var backups []BackupFile

	backupDir := workspace.Path("podman", "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	prefix := containerName + "-"
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		path := filepath.Join(backupDir, entry.Name())
		meta, err := podman.ReadBackupMeta(path)
		if err != nil {
			continue
		}

		info, _ := os.Stat(path)
		size := int64(0)
		if info != nil {
			size = info.Size()
		}

		backups = append(backups, BackupFile{
			Path:        path,
			Meta:        meta,
			Description: meta.Description,
			Size:        size,
		})
	}

	return backups, nil
}

// groupBackupsByDatabase groups backup files by database name
func groupBackupsByDatabase(backups []BackupFile) map[string][]BackupFile {
	result := make(map[string][]BackupFile)
	for _, b := range backups {
		db := b.Meta.Database
		result[db] = append(result[db], b)
	}

	// Sort each group by timestamp (newest first)
	for db := range result {
		sort.Slice(result[db], func(i, j int) bool {
			return result[db][i].Meta.Timestamp.After(result[db][j].Meta.Timestamp)
		})
	}

	return result
}

// restoreFromTemp restores a database from a temp backup file and cleans up.
func restoreFromTemp(ctx context.Context, container *podman.Container, cfg *podman.DatabaseConfig, database string, tempBackup []byte, tempBackupPath string) {
	fmt.Printf("    %s Restoring from temp backup...\n", lib.Gray("○"))

	// Drop the database again
	if err := container.DropDatabase(ctx, database); err != nil {
		fmt.Printf("    %s Could not drop database: %v\n", lib.Red("✗"), err)
		return
	}

	// Create fresh
	if err := container.CreateDatabase(ctx, database); err != nil {
		fmt.Printf("    %s Could not create database: %v\n", lib.Red("✗"), err)
		return
	}

	// Restore from temp backup
	if err := container.Restore(ctx, tempBackup); err != nil {
		fmt.Printf("    %s Temp restore failed: %v (temp backup preserved at %s)\n", lib.Red("✗"), err, tempBackupPath)
		return
	}

	// Success - remove temp backup
	os.Remove(tempBackupPath)
	fmt.Printf("    %s Database restored from temp backup and temp backup removed\n", lib.Green("✓"))
}
