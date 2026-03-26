package commands

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/podman"
	"github.com/siegg/chauffeur/internal/services"
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

	if err := flags.Parse(args); err != nil {
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
			if !interactiveConfirm("Recreate it?") {
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
		// List all and let user pick
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold("Select container to start:"))
		fmt.Println()
		for i, e := range engines {
			cfg, _ := podman.Load(e)
			status := lib.Gray("unknown")
			if cfg != nil {
				container := podman.NewContainer(client, cfg)
				running, _ := container.IsRunning(ctx)
				if running {
					status = lib.Green("running")
				} else {
					status = lib.Gray("stopped")
				}
			}
			fmt.Printf("    %d) %-20s  %s (%s)\n", i+1, cfg.ContainerName, string(cfg.Engine), status)
		}
		fmt.Println()
		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-" + fmt.Sprintf("%d", len(engines)) + " or container name]: "))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		// Try parsing as number first
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
		// List running containers first
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold("Select container to stop:"))
		fmt.Println()
		var runningList []string
		for _, e := range engines {
			cfg, _ := podman.Load(e)
			if cfg != nil {
				container := podman.NewContainer(client, cfg)
				running, _ := container.IsRunning(ctx)
				if running {
					runningList = append(runningList, e)
					fmt.Printf("    %d) %-20s  %s\n", len(runningList), cfg.ContainerName, lib.Green("running"))
				}
			}
		}
		if len(runningList) == 0 {
			lib.Info("No running containers.")
			return nil
		}
		fmt.Println()
		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-" + fmt.Sprintf("%d", len(runningList)) + " or container name]: "))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(runningList) {
			target = runningList[idx-1]
		} else {
			target = strings.ToLower(input)
		}
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

			fmt.Printf("  %s  %s  :%d\n", lib.Bold(cfg.ContainerName), statusStr, cfg.Port)
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
				cfg.ContainerName, string(cfg.Engine), statusStr, cfg.Port)
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
		if err == nil && status != nil {
			if status.Running {
				statusStr = lib.Green("● running")
			} else {
				statusStr = lib.Gray("○ stopped")
			}
		}

		fmt.Printf(" %-20s  %-10s  %-8s  :%d\n",
			cfg.ContainerName, string(cfg.Engine), statusStr, cfg.Port)
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
		fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-" + fmt.Sprintf("%d", len(engines)) + " or container name]: "))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

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

		// Confirm removal
		if !*yesFlag {
			fmt.Println()
			lib.Info(fmt.Sprintf("Remove container %s?", lib.Bold(c.ContainerName)))
			lib.Pair("  Engine", string(c.Engine))
			lib.Pair("  Container", c.ContainerName)
			lib.Pair("  Volume", c.VolumePath)
			fmt.Println()
			fmt.Print("  Confirm [y/N]: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			resp := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if resp != "y" && resp != "yes" {
				lib.Info("Cancelled.")
				return nil
			}
		}

		if verbose {
			fmt.Printf("  Removing container %s...\n", c.ContainerName)
		}
		if err := container.Remove(ctx, *forceFlag); err != nil {
			return fmt.Errorf("remove container: %w", err)
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
				podman.EngineMySQL57: 3307,
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
			if !interactiveConfirm("Overwrite it?") {
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
// Target can be "all", an engine name (mysql8, postgres, etc), or a container name.
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

	// Try as engine name first
	if podman.IsValidEngine(target) {
		cfg, err := podman.Load(target)
		if err == nil && cfg != nil {
			return []*podman.DatabaseConfig{cfg}, nil
		}
	}

	// Try as container name
	cfg, err := podman.Load(target)
	if err == nil && cfg != nil {
		return []*podman.DatabaseConfig{cfg}, nil
	}

	return nil, nil
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
func interactiveConfirm(prompt string) bool {
	fmt.Print("  " + prompt + " " + lib.Gray("[y/N]") + ": ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	resp := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return resp == "y" || resp == "yes"
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
	flags := flag.NewFlagSet("podman backup", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman backup — backup a container to file",
		"chauf podman backup <container-name> [--output <path>]")

	outputFlag := flags.String("output", "", "Output file path (default: <container-name>-<timestamp>.tar.gz in current dir)")

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
		target = flags.Arg(0)
	}

	containers, err := resolveContainers(target)
	if err != nil || len(containers) == 0 {
		lib.Warn("No containers found matching " + lib.Bold(target))
		return nil
	}

	if len(containers) > 1 {
		lib.Warn("Backup requires exactly one container. Use container name.")
		return nil
	}

	cfg := containers[0]
	container := podman.NewContainer(client, cfg)
	container.SetLogger(&verboseLogger{verbose: verbose})

	// Check if running
	running, _ := container.IsRunning(ctx)
	if !running {
		lib.Warn(fmt.Sprintf("Container %s is not running. Start it first.", lib.Bold(cfg.ContainerName)))
		return nil
	}

	// Determine output path
	outputPath := *outputFlag
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		outputPath = fmt.Sprintf("%s-%s.tar.gz", cfg.ContainerName, timestamp)
	}

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Backing up container:"))
	fmt.Printf("    Container:  %s\n", cfg.ContainerName)
	fmt.Printf("    Engine:     %s\n", cfg.Engine)
	fmt.Printf("    Database:   app\n", cfg.Username)
	fmt.Printf("    Output:     %s\n", outputPath)
	fmt.Println()

	backupData, err := container.Backup(ctx, outputPath)
	if err != nil {
		lib.Error("Backup failed: " + err.Error())
		return err
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Backup created: %s (%d bytes)", outputPath, len(backupData)))

	// Show DSN for reference
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Connection string:"))
	fmt.Printf("    %s\n", podman.DSN(cfg))
	fmt.Println()

	return nil
}

// ── chauf podman restore ──────────────────────────────────────────────────────

func runPodmanRestore(args []string) error {
	flags := flag.NewFlagSet("podman restore", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf podman restore — restore a container from backup",
		"chauf podman restore <container-name> [--input <path>]")

	inputFlag := flags.String("input", "", "Input backup file path (required)")
	yesFlag := flags.Bool("yes", false, "Skip confirmation if container is running")

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
		target = flags.Arg(0)
	}

	if *inputFlag == "" {
		lib.Error("Missing --input flag. Usage: chauf podman restore <container> --input <backup-file>")
		return nil
	}

	// Read backup file
	backupData, err := os.ReadFile(*inputFlag)
	if err != nil {
		return fmt.Errorf("read backup file: %w", err)
	}

	containers, err := resolveContainers(target)
	if err != nil || len(containers) == 0 {
		lib.Warn("No containers found matching " + lib.Bold(target))
		return nil
	}

	if len(containers) > 1 {
		lib.Warn("Restore requires exactly one container. Use container name.")
		return nil
	}

	cfg := containers[0]
	container := podman.NewContainer(client, cfg)
	container.SetLogger(&verboseLogger{verbose: verbose})

	// Check if running - need to stop for restore
	running, _ := container.IsRunning(ctx)
	if running && !*yesFlag {
		fmt.Println()
		lib.Info(fmt.Sprintf("Container %s is running. Stop it first?", lib.Bold(cfg.ContainerName)))
		if !interactiveConfirm("Stop container") {
			return nil
		}
		if err := container.Stop(ctx, 10*time.Second); err != nil {
			return fmt.Errorf("stop container: %w", err)
		}
	}

	// Ensure container exists and is started
	exists, _ := client.ContainerExists(ctx, cfg.ContainerName)
	if !exists {
		lib.Warn(fmt.Sprintf("Container %s does not exist. Create it first with 'chauf podman create'.", lib.Bold(cfg.ContainerName)))
		return nil
	}

	if !running {
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold("Starting container for restore..."))
		if err := container.Start(ctx); err != nil {
			return fmt.Errorf("start container: %w", err)
		}
		// Wait for database to be ready
		time.Sleep(2 * time.Second)
	}

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Restoring container:"))
	fmt.Printf("    Container:  %s\n", cfg.ContainerName)
	fmt.Printf("    Engine:     %s\n", cfg.Engine)
	fmt.Printf("    Backup:     %s\n", *inputFlag)
	fmt.Println()

	if err := container.Restore(ctx, backupData); err != nil {
		lib.Error("Restore failed: " + err.Error())
		return err
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Container %s restored successfully", lib.Bold(cfg.ContainerName)))

	// Show DSN
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Connection string:"))
	fmt.Printf("    %s\n", podman.DSN(cfg))
	fmt.Println()

	return nil
}
