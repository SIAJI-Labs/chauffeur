package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/example"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunInit handles `chauf init` command invocations
func RunInit(args []string) error {
	var (
		force bool
		quiet bool
	)

	// Parse command line flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printInitUsage()
			return nil
		case "--force":
			force = true
		case "--quiet", "-q":
			quiet = true
		default:
			return fmt.Errorf("unknown flag: %s", arg)
		}
	}

	logger := lib.NewCommandLogger("init")

	if !quiet {
		logger.Info("Initializing Chauffeur workspace...")
	}

	// Get workspace directory
	wsDir, err := workspace.Dir()
	if err != nil {
		return fmt.Errorf("determine workspace directory: %w", err)
	}

	if !quiet {
		logger.Info(fmt.Sprintf("Workspace directory: %s", wsDir))
	}

	// Check if workspace is already initialized
	configPath := filepath.Join(wsDir, "config", "chauffeur.yaml")
	if _, err := os.Stat(configPath); err == nil && !force {
		// Config already exists and --force not specified
		logger.Warn("Workspace already initialized", fmt.Sprintf("Configuration file exists: %s (use --force to overwrite)", configPath))
		return nil
	}

	// Create necessary directories
	dirs := []string{
		wsDir,
		filepath.Join(wsDir, "config"),
		filepath.Join(wsDir, "projects"),
		filepath.Join(wsDir, "logs"),
		filepath.Join(wsDir, "cache"),
		filepath.Join(wsDir, "php"),
		filepath.Join(wsDir, "nginx", "bin"),
		filepath.Join(wsDir, "nginx", "etc"),
		filepath.Join(wsDir, "nginx", "sites-available"),
		filepath.Join(wsDir, "nginx", "sites-enabled"),
		filepath.Join(wsDir, "nginx", "conf.d"),
		filepath.Join(wsDir, "nginx", "logs"),
		filepath.Join(wsDir, "bin"),
		filepath.Join(wsDir, "bin", "shims"),
	}

	for _, dir := range dirs {
		if !quiet {
			logger.Info(fmt.Sprintf("Creating directory: %s", dir))
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Create default configuration file
	if err := createDefaultConfig(wsDir, logger, quiet); err != nil {
		return fmt.Errorf("create default configuration: %w", err)
	}

	// Create .gitignore in workspace directory if it doesn't exist
	gitignorePath := filepath.Join(wsDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := `# Chauffeur workspace
logs/
cache/
projects/*/logs/
projects/*/runtime/
*/bin/
*-build/
tmp/

# Ignore service pid files
*.pid
*.sock

# Ignore old configuration (backups)
*.yaml.bak
*.yaml.old
`
		if !quiet {
			logger.Info("Creating .gitignore file")
		}
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("create .gitignore: %w", err)
		}
	}

	// Ensure PATH includes Chauffeur bin directory
	binDir := filepath.Join(wsDir, "bin")
	if !quiet {
		logger.Info("Checking PATH configuration...")
		logger.Info(fmt.Sprintf("Chauffeur bin directory: %s", binDir))
		logger.Info("Add this to your shell profile (~/.bashrc or ~/.zshrc):")
		logger.Info(fmt.Sprintf("  export PATH=\"%s:$PATH\"", binDir))
		logger.Info("Then restart your shell or run: source ~/.bashrc")
	}

	// Create example project directory immediately
	logger.Info("Creating example project...")
	if err := example.CreateExampleProject(); err != nil {
		logger.Warn("Failed to create example project", err.Error())
	} else {
		logger.Info("Example project created. It will be linked after installing nginx and php.")
	}

	if !quiet {
		logger.Success("Chauffeur workspace initialized successfully", fmt.Sprintf("at %s", wsDir))
		logger.Info("Next steps:")
		logger.Info("  1. Source the shell PATH configuration or restart your shell")
		logger.Info("  2. Install services: chauf install nginx php")
		logger.Info("  3. Create your first project: cd /path/to/project && chauf link")
		logger.Info("  4. Start services: chauf start")
		logger.Info("")
		logger.Info("🎁 Example project created!")
		logger.Info("   Located at: ~/.chauffeur/projects/example-project")
		logger.Info("   Will be automatically linked after installing nginx and php")
	} else {
		logger.Success("Workspace initialized", wsDir)
	}

	return nil
}

// createDefaultConfig creates the default configuration file
func createDefaultConfig(wsDir string, logger *lib.Logger, quiet bool) error {
	configDir := filepath.Join(wsDir, "config")
	configPath := filepath.Join(configDir, "chauffeur.yaml")

	if !quiet {
		logger.Info("Creating default configuration file...")
	}

	// Load default configuration
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("load default config: %w", err)
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}

	if !quiet {
		logger.Success("Configuration file created", configPath)
		logger.Info("Default settings used:")
		logger.Info(fmt.Sprintf("  Nginx HTTP port: %d", cfg.Nginx.HTTPPort))
		logger.Info(fmt.Sprintf("  Nginx HTTPS port: %d", cfg.Nginx.HTTPSPort))
		logger.Info(fmt.Sprintf("  Port range: %d-%d", cfg.Ports.StartRange, cfg.Ports.EndRange))
		logger.Info(fmt.Sprintf("  Conflict resolution: %s", cfg.Ports.ConflictResolution))
	}

	return nil
}

// DefaultConfig provides access to the default config function from internal package
func DefaultConfig() (config.Config, error) {
	return config.DefaultConfig()
}

// printInitUsage prints usage information for the init command
func printInitUsage() {
	fmt.Print(`Chauffeur Workspace Initialization

Usage:
  chauf init [--force] [--quiet]

Flags:
  --force           Overwrite existing configuration files
  --quiet, -q       Suppress verbose output

Description:
  Initializes the Chauffeur workspace with default configuration
  and directory structure. Creates ~/.chauffeur/ with necessary
  subdirectories including config/, projects/, logs/, cache/, and
  service-specific directories.

Default Configuration:
  - Nginx HTTP port: 8080 (user-space to avoid system conflicts)
  - Nginx HTTPS port: 8443
  - Port range for auto-allocation: 8080-8099
  - Conflict resolution: prompt (interactive)

After initialization, you must add the Chauffeur bin directory
to your PATH. The init command will provide the exact command
to run based on your system.

Examples:
  chauf init                    # Initialize with default settings
  chauf init --force             # Overwrite existing configuration
  chauf init --quiet             # Minimal output
`)
}
