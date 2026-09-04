package commands

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunConfig(args []string) error {
	if len(args) == 0 {
		configHelp()
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "runtime":
		return configRuntime(args[1:])
	case "php":
		return configPHP(args[1:])
	case "nginx":
		return configNginx(args[1:])
	case "--help", "-h", "help":
		configHelp()
		return nil
	default:
		return fmt.Errorf("unknown config type %q — run: chauf config --help", args[0])
	}
}

func configHelp() {
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("chauf config"))
	fmt.Println()
	fmt.Println("  Read and write chauffeur configuration.")
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Usage:"))
	fmt.Println("    chauf config php [version] [key] [value]")
	fmt.Println("    chauf config runtime [native|podman]")
	fmt.Println("    chauf config nginx [key] [value]")
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Examples:"))
	fmt.Println("    chauf config php              # List all PHP version configs")
	fmt.Println("    chauf config php 8.3         # Show PHP 8.3 config")
	fmt.Println("    chauf config php 8.3 upload_max_filesize 128M")
	fmt.Println("    chauf config nginx           # Show nginx config")
	fmt.Println("    chauf config nginx upload_max_size 256M")
	fmt.Println("    chauf config nginx upload_max_size \"\"  # reset: follow PHP post_max_size")
	fmt.Println()
}

func configRuntime(args []string) error {
	cfg := workspace.Load()
	if len(args) == 0 {
		fmt.Println()
		lib.Pair("Runtime engine", cfg.Runtime.Engine)
		fmt.Println()
		return nil
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: chauf config runtime [native|podman]")
	}
	engine := strings.ToLower(args[0])
	previousEngine := cfg.Runtime.Engine
	if previousEngine != engine {
		if err := stopRunningRuntimeServices(cfg); err != nil {
			return fmt.Errorf("stop %s services before changing runtime: %w", previousEngine, err)
		}
	}
	if err := workspace.SetRuntimeEngine(engine); err != nil {
		return err
	}
	fmt.Printf("  %s runtime.engine = %s\n", lib.Green("✓"), engine)
	if previousEngine != engine {
		// Runtime selection is an ownership boundary. Remove enabled units for
		// the old engine now, while preserving its installed artifacts and
		// containers so the user can switch back safely.
		if previousEngine == string(chauftruntime.EngineNative) && engine == string(chauftruntime.EnginePodman) {
			if err := migrateNativeAutostartToPodman(); err != nil {
				_ = workspace.SetRuntimeEngine(previousEngine)
				return fmt.Errorf("migrate auto-start to Podman: %w", err)
			}
		} else if previousEngine == string(chauftruntime.EnginePodman) && engine == string(chauftruntime.EngineNative) {
			if err := disablePodmanAutostart(); err != nil {
				_ = workspace.SetRuntimeEngine(previousEngine)
				return fmt.Errorf("disable Podman auto-start: %w", err)
			}
		}
	}
	if engine == string(chauftruntime.EnginePodman) && system.IsUnitEnabled("chauffeur-nginx.service") {
		lib.Warn("native nginx auto-start is still enabled; run `chauf autostart enable` to migrate it to Podman or `chauf autostart disable`")
	}
	if engine == string(chauftruntime.EngineNative) && system.IsUnitEnabled(system.PodmanNginxUnit()) {
		lib.Warn("Podman auto-start is still enabled; run `chauf autostart disable` before starting native services")
	}
	return nil
}

// stopRunningRuntimeServices makes runtime changes an explicit ownership
// handoff. Inspect status first so stopped or absent services are left alone;
// only running services are stopped before the config is changed.
func stopRunningRuntimeServices(cfg workspace.Config) error {
	rt, err := chauftruntime.ForWorkspace(cfg)
	if err != nil {
		return err
	}
	scopes, err := runtimeServiceScopes(workspace.Root())
	if err != nil {
		return err
	}
	scopes = append(scopes, chauftruntime.Scope{Service: "nginx"})
	ctx := context.Background()
	for _, scope := range scopes {
		statuses, statusErr := rt.Status(ctx, scope)
		if statusErr != nil {
			return fmt.Errorf("check %s status: %w", scopeLabel(scope), statusErr)
		}
		running := false
		for _, status := range statuses {
			if status.Healthy || status.State == "running" {
				running = true
				break
			}
		}
		if !running {
			continue
		}
		if err := rt.Stop(ctx, scope); err != nil {
			return fmt.Errorf("stop %s: %w", scopeLabel(scope), err)
		}
		lib.Info(lib.Gray("stopped " + scopeLabel(scope)))
	}
	return nil
}

func scopeLabel(scope chauftruntime.Scope) string {
	if scope.Service != "" {
		return scope.Service
	}
	if scope.Version != "" {
		return "php-fpm " + scope.Version
	}
	return "runtime service"
}

func configPHP(args []string) error {
	cfg := workspace.Load()
	printRuntime(cfg)
	fmt.Println()

	if len(args) == 0 {
		// List all PHP versions with configs
		return configPHPList(cfg)
	}

	version := args[0]
	if _, ok := cfg.PHP.Versions[version]; !ok {
		// Check if PHP version is installed
		var installed []string
		if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
			for _, supported := range installers.SupportedPHPVersions {
				if isInstalledPHPVersion(supported) {
					installed = append(installed, supported)
				}
			}
		} else {
			installed = installers.ListInstalledPHP(workspace.Root())
		}
		found := false
		for _, v := range installed {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("PHP %s is not installed", version)
		}
		// Initialize with defaults
		cfg.PHP.Versions[version] = workspace.DefaultPHPVersionConfig()
	}

	if len(args) == 1 {
		// Show specific version config
		return configPHPShow(version, cfg.PHP.Versions[version])
	}

	if len(args) < 3 {
		return fmt.Errorf("Usage: chauf config php <version> <key> <value>")
	}

	key := args[1]
	value := args[2]

	// Validate key
	validKeys := []string{"upload_max_filesize", "post_max_size", "memory_limit", "max_execution_time", "max_input_vars"}
	valid := false
	for _, k := range validKeys {
		if k == key {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid key %q — valid keys: %s", key, strings.Join(validKeys, ", "))
	}

	// Update config in memory
	phpCfg := cfg.PHP.Versions[version]
	switch key {
	case "upload_max_filesize":
		phpCfg.UploadMaxFilesize = value
	case "post_max_size":
		phpCfg.PostMaxSize = value
	case "memory_limit":
		phpCfg.MemoryLimit = value
	case "max_execution_time", "max_input_vars":
		// These should be integers - validate
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
			return fmt.Errorf("value for %s must be an integer", key)
		}
		if key == "max_execution_time" {
			phpCfg.MaxExecutionTime = intVal
		} else {
			phpCfg.MaxInputVars = intVal
		}
	}

	// Store updated config back to cfg
	cfg.PHP.Versions[version] = phpCfg

	// Save to chauffeur.yaml
	if err := workspace.SavePHPVersionSetting(version, key, value); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Regenerate limits.ini
	if cfg.Runtime.Engine != string(chauftruntime.EnginePodman) {
		inst, err := installers.NewPHPInstaller(version, installers.BuildOpts{})
		if err == nil {
			inst.WritePHPConfig(phpCfg)
		}
	} else {
		lib.Info(lib.Gray("PHP settings were saved for the Podman runtime; rebuild the image to bake them into a new image."))
	}

	fmt.Printf("  %s %s.%s = %s\n", lib.Green("✓"), version, key, value)
	fmt.Println()

	return nil
}

func configPHPList(cfg workspace.Config) error {
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("PHP Version Configurations"))
	fmt.Println()

	versions := make([]string, 0, len(cfg.PHP.Versions))
	for v := range cfg.PHP.Versions {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	if len(versions) == 0 {
		fmt.Println("  No PHP version configs found.")
		fmt.Println("  Run 'chauf config php <version> <key> <value>' to add one.")
		fmt.Println()
		return nil
	}

	for _, ver := range versions {
		c := cfg.PHP.Versions[ver]
		fmt.Printf("  %s\n", lib.Bold(ver))
		lib.Pair("upload_max_filesize", c.UploadMaxFilesize)
		lib.Pair("post_max_size", c.PostMaxSize)
		lib.Pair("memory_limit", c.MemoryLimit)
		fmt.Printf("  %-20s %d\n", "max_execution_time:", c.MaxExecutionTime)
		fmt.Printf("  %-20s %d\n", "max_input_vars:", c.MaxInputVars)
		fmt.Println()
	}

	return nil
}

func configPHPShow(version string, cfg workspace.PHPVersionConfig) error {
	fmt.Println()
	fmt.Printf("  %s %s\n", lib.Bold("PHP"), version)
	fmt.Println()
	lib.Pair("upload_max_filesize", cfg.UploadMaxFilesize)
	lib.Pair("post_max_size", cfg.PostMaxSize)
	lib.Pair("memory_limit", cfg.MemoryLimit)
	fmt.Printf("  %-20s %d\n", "max_execution_time:", cfg.MaxExecutionTime)
	fmt.Printf("  %-20s %d\n", "max_input_vars:", cfg.MaxInputVars)
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Update: chauf config php "+version+" <key> <value>"))
	fmt.Println()
	return nil
}

// ── nginx config ───────────────────────────────────────────────────────────────

func configNginx(args []string) error {
	cfg := workspace.Load()
	printRuntime(cfg)
	fmt.Println()

	if len(args) == 0 {
		return configNginxShow(cfg)
	}

	if len(args) < 2 {
		return fmt.Errorf("Usage: chauf config nginx <key> <value>")
	}

	key := args[0]
	value := args[1]

	// Validate key
	validKeys := []string{"http_port", "https_port", "upload_max_size"}
	valid := false
	for _, k := range validKeys {
		if k == key {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid key %q — valid keys: %s", key, strings.Join(validKeys, ", "))
	}
	if key == "http_port" || key == "https_port" {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("%s must be a port between 1 and 65535", key)
		}
	}

	// Save to chauffeur.yaml
	if err := workspace.SaveNginxSetting(key, value); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if value == "" {
		fmt.Printf("  %s nginx.%s reset (will follow PHP post_max_size)\n", lib.Green("✓"), key)
	} else {
		fmt.Printf("  %s nginx.%s = %s\n", lib.Green("✓"), key, value)
	}
	fmt.Println()

	return nil
}

func configNginxShow(cfg workspace.Config) error {
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("nginx"))
	fmt.Println()
	lib.Pair("http_port", strconv.Itoa(cfg.Nginx.HTTPPort))
	lib.Pair("https_port", strconv.Itoa(cfg.Nginx.HTTPSPort))
	if cfg.Nginx.UploadMaxSize != "" {
		lib.Pair("upload_max_size", cfg.Nginx.UploadMaxSize)
		fmt.Printf("  %s\n", lib.Gray("(override — empty value to reset and follow PHP)"))
	} else {
		lib.Pair("upload_max_size", "(follows PHP post_max_size)")
	}
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Update: chauf config nginx <key> <value>"))
	fmt.Println()
	return nil
}
