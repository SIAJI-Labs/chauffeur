package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── chauf autostart ───────────────────────────────────────────────────────────

func RunAutostart(args []string) error {
	if len(args) == 0 {
		return autostartStatus()
	}

	switch args[0] {
	case "-h", "--help", "help":
		return autostartHelp()
	case "enable":
		return autostartEnable(args[1:])
	case "disable":
		return autostartDisable(args[1:])
	case "status", "list":
		return autostartStatus()
	default:
		return fmt.Errorf("unknown autostart subcommand %q — use: enable, disable, status", args[0])
	}
}

func autostartHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf autostart — manage systemd user services for auto-start on login"))
	fmt.Printf("  %s\n\n", lib.Gray("Usage: chauf autostart <subcommand> [target]"))
	fmt.Printf("  %-34s  %s\n", "enable", lib.Gray("Enable all Chauffeur services"))
	fmt.Printf("  %-34s  %s\n", "enable nginx", lib.Gray("Enable only nginx"))
	fmt.Printf("  %-34s  %s\n", "enable php <version>", lib.Gray("Enable a specific PHP-FPM version"))
	fmt.Printf("  %-34s  %s\n", "disable", lib.Gray("Disable all Chauffeur services"))
	fmt.Printf("  %-34s  %s\n", "disable nginx", lib.Gray("Disable only nginx"))
	fmt.Printf("  %-34s  %s\n", "disable php <version>", lib.Gray("Disable a specific PHP-FPM version"))
	fmt.Printf("  %-34s  %s\n", "status", lib.Gray("Show enabled/running state for all services"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Services start on user login. For boot-time start (before login):"))
	fmt.Printf("  %s\n", lib.Gray("  loginctl enable-linger $USER"))
	fmt.Println()
	return nil
}

// ── enable ────────────────────────────────────────────────────────────────────

func autostartEnable(args []string) error {
	fmt.Println()

	switch {
	case len(args) == 0:
		// Enable nginx + all installed PHP versions
		if err := enableNginx(); err != nil {
			return err
		}
		root := autostartRoot()
		for _, ver := range installedPHPVersions(root) {
			if err := enableFPM(ver); err != nil {
				lib.Warn(fmt.Sprintf("php-fpm %s: %s", ver, err))
			}
		}

	case args[0] == "nginx":
		if err := enableNginx(); err != nil {
			return err
		}

	case args[0] == "php" || args[0] == "fpm":
		if len(args) < 2 {
			return fmt.Errorf("usage: chauf autostart enable php <version>")
		}
		if err := enableFPM(args[1]); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown target %q — use: nginx, php <version>", args[0])
	}

	fmt.Println()
	if !system.IsLingeringEnabled() {
		lib.Info(lib.Gray("Services start on login. To also start on boot (before login):"))
		lib.Info(lib.Gray(fmt.Sprintf("  loginctl enable-linger %s", os.Getenv("USER"))))
		fmt.Println()
	}

	return nil
}

func enableNginx() error {
	lib.Step("writing chauffeur-nginx.service")
	if err := system.WriteNginxUnit(); err != nil {
		return fmt.Errorf("write nginx unit: %w", err)
	}
	if err := system.UserDaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	if err := system.EnableUnit("chauffeur-nginx.service"); err != nil {
		return fmt.Errorf("enable nginx: %w", err)
	}
	lib.Success("nginx  auto-start enabled")

	// Start under systemd now. If nginx is already running outside systemd
	// (e.g. via chauf start), this will fail — tell the user how to hand off.
	if !system.IsUnitActive("chauffeur-nginx.service") {
		if err := system.StartUnit("chauffeur-nginx.service"); err != nil {
			lib.Warn("nginx could not start — it may already be running outside systemd.")
			lib.Info(lib.Gray("  To hand off management:  chauf stop && systemctl --user start chauffeur-nginx.service"))
		}
	}
	return nil
}

func enableFPM(version string) error {
	lib.Step("writing chauffeur-php-fpm@.service")
	if err := system.WriteFPMTemplateUnit(); err != nil {
		return fmt.Errorf("write php-fpm template unit: %w", err)
	}
	unit := system.FPMInstanceUnit(version)
	lib.Step("enabling " + unit)
	if err := system.UserDaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	if err := system.EnableUnit(unit); err != nil {
		return fmt.Errorf("enable %s: %w", unit, err)
	}
	lib.Success(fmt.Sprintf("php-fpm %s  auto-start enabled", version))

	if !system.IsUnitActive(unit) {
		if err := system.StartUnit(unit); err != nil {
			lib.Warn(fmt.Sprintf("php-fpm %s could not start — it may already be running outside systemd.", version))
			lib.Info(lib.Gray(fmt.Sprintf("  To hand off:  chauf stop && systemctl --user start %s", unit)))
		}
	}
	return nil
}

// ── disable ───────────────────────────────────────────────────────────────────

func autostartDisable(args []string) error {
	fmt.Println()

	switch {
	case len(args) == 0:
		// Disable everything
		root := autostartRoot()
		for _, ver := range installedPHPVersions(root) {
			disableFPM(ver)
		}
		disableNginx()

	case args[0] == "nginx":
		disableNginx()

	case args[0] == "php" || args[0] == "fpm":
		if len(args) < 2 {
			return fmt.Errorf("usage: chauf autostart disable php <version>")
		}
		disableFPM(args[1])

	default:
		return fmt.Errorf("unknown target %q — use: nginx, php <version>", args[0])
	}

	fmt.Println()
	return nil
}

func disableNginx() {
	if err := system.DisableUnit("chauffeur-nginx.service"); err != nil {
		lib.Warn("nginx: " + err.Error())
		return
	}
	lib.Success("nginx  auto-start disabled")
}

func disableFPM(version string) {
	unit := system.FPMInstanceUnit(version)
	if err := system.DisableUnit(unit); err != nil {
		lib.Warn(fmt.Sprintf("php-fpm %s: %s", version, err))
		return
	}
	lib.Success(fmt.Sprintf("php-fpm %s  auto-start disabled", version))
}

// ── status ────────────────────────────────────────────────────────────────────

func autostartStatus() error {
	fmt.Println()
	fmt.Printf("  %s\n\n", lib.Bold("Chauffeur Auto-Start"))

	const lblW = 20
	header := fmt.Sprintf(" %-*s  %-10s  %s", lblW, "Service", "Enabled", "Running")
	sep := strings.Repeat("─", len(header))
	fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

	printAutostartRow(lblW, "nginx", "chauffeur-nginx.service")

	root := autostartRoot()
	for _, ver := range installedPHPVersions(root) {
		printAutostartRow(lblW, "php-fpm "+ver, system.FPMInstanceUnit(ver))
	}

	fmt.Println()

	if !system.IsLingeringEnabled() {
		lib.Info(lib.Gray("Lingering disabled — services start on login only, not on boot."))
		lib.Info(lib.Gray(fmt.Sprintf("  Enable boot start:  loginctl enable-linger %s", os.Getenv("USER"))))
		fmt.Println()
	}

	return nil
}

func printAutostartRow(lblW int, label, unit string) {
	enabled := lib.Gray("○ disabled")
	if system.IsUnitEnabled(unit) {
		enabled = lib.Green("● enabled ")
	}
	running := lib.Gray("○ stopped")
	if system.IsUnitActive(unit) {
		running = lib.Green("● running")
	}
	fmt.Printf(" %-*s  %s  %s\n", lblW, label, enabled, running)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func autostartRoot() string {
	return workspace.Root()
}
