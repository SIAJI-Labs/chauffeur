package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

// CleanOptions defines options for the clean command
type CleanOptions struct {
	What         string // "all", "logs", "cache", "temp", "old-versions", "ssl-certs", "projects"
	DryRun       bool
	Force        bool
	OlderThan    string
	KeepVersions int
	Verbose      bool
	Quiet        bool
}

// Cleaner handles workspace cleanup operations
type Cleaner struct {
	logger       *lib.Logger
	workspaceDir string
	projectsDir  string
}

// NewCleaner creates a new cleaner instance
func NewCleaner() (*Cleaner, error) {
	wsDir, err := workspace.Dir()
	if err != nil {
		return nil, fmt.Errorf("get workspace directory: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		// Use default projects directory if config loading fails
		cfg.ProjectsDir = filepath.Join(wsDir, "projects")
	}

	return &Cleaner{
		logger:       lib.NewCommandLogger("clean"),
		workspaceDir: wsDir,
		projectsDir:  cfg.ProjectsDir,
	}, nil
}

// RunClean implements the clean command
func RunClean(args []string) error {
	// Validate workspace exists, offer to initialize if not (skip for help)
	if err := lib.ValidateWorkspace(args); err != nil {
		return err
	}

	cleaner, err := NewCleaner()
	if err != nil {
		return fmt.Errorf("create cleaner: %w", err)
	}

	options, err := parseCleanArgs(args)
	if err != nil {
		return err
	}

	// If help was requested, options will be nil
	if options == nil {
		return nil
	}

	// Ensure workspace exists
	if err := workspace.Ensure(); err != nil {
		return cleaner.logger.Error("ensure workspace", err.Error())
	}

	return cleaner.Clean(options)
}

// parseCleanArgs parses command line arguments for the clean command
func parseCleanArgs(args []string) (*CleanOptions, error) {
	options := &CleanOptions{
		What:         "all", // Default to cleaning everything
		KeepVersions: 2,     // Keep last 2 versions by default
	}

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printCleanUsage()
			return nil, nil
		case "--dry-run", "-n":
			options.DryRun = true
			i++
		case "--force", "-f":
			options.Force = true
			i++
		case "--verbose", "-v":
			options.Verbose = true
			i++
		case "--older-than":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--older-than requires a duration (e.g., 30d, 7d, 1h)")
			}
			options.OlderThan = args[i+1]
			i += 2
		case "--keep-versions":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--keep-versions requires a number")
			}
			keep, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid keep count: %s", args[i+1])
			}
			options.KeepVersions = keep
			i += 2
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			if options.What != "all" && options.What != "" {
				return nil, fmt.Errorf("multiple clean targets specified")
			}
			options.What = arg
			i++
		}
	}

	// Validate clean target
	validTargets := []string{"all", "logs", "cache", "temp", "old-versions", "ssl-certs", "projects"}
	valid := false
	for _, target := range validTargets {
		if options.What == target {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("invalid clean target: %s", options.What)
	}

	return options, nil
}

// Clean performs the cleaning operation based on options
func (c *Cleaner) Clean(options *CleanOptions) error {
	if options.DryRun {
		c.logger.Info("DRY RUN MODE - No files will be deleted")
	}

	switch options.What {
	case "all":
		return c.cleanAll(options)
	case "logs":
		return c.cleanLogs(options)
	case "cache":
		return c.cleanCache(options)
	case "temp":
		return c.cleanTemp(options)
	case "old-versions":
		return c.cleanOldVersions(options)
	case "ssl-certs":
		return c.cleanSSLCerts(options)
	case "projects":
		return c.cleanProjects(options)
	default:
		return fmt.Errorf("unsupported clean target: %s", options.What)
	}
}

// cleanAll performs comprehensive cleaning
func (c *Cleaner) cleanAll(options *CleanOptions) error {
	c.logger.PrintSection("Comprehensive Workspace Cleanup")

	operations := []struct {
		name string
		fn   func(*CleanOptions) error
	}{
		{"Log Files", c.cleanLogs},
		{"Cache Files", c.cleanCache},
		{"Temporary Files", c.cleanTemp},
		{"Old PHP Versions", c.cleanOldVersions},
		{"Stale SSL Certificates", c.cleanSSLCerts},
	}

	for _, op := range operations {
		if !options.Quiet {
			c.logger.Info(fmt.Sprintf("Cleaning %s", op.name))
		}
		if err := op.fn(options); err != nil {
			c.logger.Warn("Failed to clean", fmt.Sprintf("%s: %v", op.name, err))
		}
	}

	return nil
}

// cleanLogs removes log files
func (c *Cleaner) cleanLogs(options *CleanOptions) error {
	logDirs := []string{
		filepath.Join(c.workspaceDir, "nginx", "logs"),
		filepath.Join(c.workspaceDir, "php"),
		filepath.Join(c.projectsDir),
	}

	var totalSize int64
	var totalFiles int

	for _, dir := range logDirs {
		size, files, err := c.cleanLogsInDirectory(dir, options)
		if err != nil {
			c.logger.Warn("Failed to clean logs in directory", fmt.Sprintf("%s: %v", dir, err))
			continue
		}
		totalSize += size
		totalFiles += files
	}

	if totalFiles > 0 {
		c.logger.Success("Log cleanup complete", fmt.Sprintf("%d files, %s freed", totalFiles, c.formatBytes(totalSize)))
	} else {
		c.logger.Info("No log files found to clean")
	}

	return nil
}

// cleanLogsInDirectory removes log files in a specific directory
func (c *Cleaner) cleanLogsInDirectory(dir string, options *CleanOptions) (int64, int, error) {
	var totalSize int64
	var totalFiles int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a log file
		if strings.HasSuffix(strings.ToLower(info.Name()), ".log") {
			// Check age if specified
			if options.OlderThan != "" {
				olderThan, err := time.ParseDuration(options.OlderThan)
				if err == nil {
					if time.Since(info.ModTime()) < olderThan {
						return nil // Skip recent files
					}
				}
			}

			if options.Verbose {
				fmt.Printf("  Found log: %s (%s)\n", path, c.formatBytes(info.Size()))
			}

			// Only count files that are actually deleted
			if !options.DryRun {
				if options.Force || c.confirmDelete("log file", path) {
					if err := os.Remove(path); err != nil {
						c.logger.Warn("Failed to remove log file", fmt.Sprintf("%s: %v", path, err))
					} else {
						totalSize += info.Size()
						totalFiles++
						if options.Verbose {
							fmt.Printf("  Deleted: %s\n", path)
						}
					}
				}
			} else {
				// In dry-run mode, count what would be deleted
				totalSize += info.Size()
				totalFiles++
			}
		}

		return nil
	})

	return totalSize, totalFiles, err
}

// cleanCache removes cache files
func (c *Cleaner) cleanCache(options *CleanOptions) error {
	cacheDirs := []string{
		filepath.Join(c.workspaceDir, "cache"),
		filepath.Join(os.TempDir(), "chauffeur-cache"),
	}

	// PHP cache directories
	phpDir := filepath.Join(c.workspaceDir, "php")
	if phpInfo, err := os.Stat(phpDir); err == nil && phpInfo.IsDir() {
		if versions, err := os.ReadDir(phpDir); err == nil {
			for _, version := range versions {
				if version.IsDir() {
					cacheDirs = append(cacheDirs,
						filepath.Join(phpDir, version.Name(), "cache"),
						filepath.Join(phpDir, version.Name(), "opcache"),
					)
				}
			}
		}
	}

	var totalSize int64
	var totalFiles int

	for _, dir := range cacheDirs {
		size, files, err := c.cleanDirectory(dir, options, true)
		if err != nil {
			c.logger.Warn("Failed to clean cache directory", fmt.Sprintf("%s: %v", dir, err))
			continue
		}
		totalSize += size
		totalFiles += files
	}

	if totalFiles > 0 {
		c.logger.Success("Cache cleanup complete", fmt.Sprintf("%d files, %s freed", totalFiles, c.formatBytes(totalSize)))
	} else {
		c.logger.Info("No cache files found to clean")
	}

	return nil
}

// cleanTemp removes temporary files
func (c *Cleaner) cleanTemp(options *CleanOptions) error {
	tempDirs := []string{
		filepath.Join(c.workspaceDir, "tmp"),
		filepath.Join(c.workspaceDir, "temp"),
		filepath.Join(os.TempDir(), "chauffeur-*"),
	}

	var totalSize int64
	var totalFiles int

	for _, dir := range tempDirs {
		size, files, err := c.cleanDirectory(dir, options, true)
		if err != nil {
			c.logger.Warn("Failed to clean temp directory", fmt.Sprintf("%s: %v", dir, err))
			continue
		}
		totalSize += size
		totalFiles += files
	}

	if totalFiles > 0 {
		c.logger.Success("Temp cleanup complete", fmt.Sprintf("%d files, %s freed", totalFiles, c.formatBytes(totalSize)))
	} else {
		c.logger.Info("No temporary files found to clean")
	}

	return nil
}

// cleanOldVersions removes old PHP versions
func (c *Cleaner) cleanOldVersions(options *CleanOptions) error {
	phpDir := filepath.Join(c.workspaceDir, "php")
	phpInfo, err := os.Stat(phpDir)
	if err != nil || !phpInfo.IsDir() {
		c.logger.Info("No PHP installations found")
		return nil
	}

	versions, err := os.ReadDir(phpDir)
	if err != nil {
		return fmt.Errorf("read PHP directory: %w", err)
	}

	var versionDirs []string
	for _, version := range versions {
		if version.IsDir() {
			versionDirs = append(versionDirs, filepath.Join(phpDir, version.Name()))
		}
	}

	if len(versionDirs) <= options.KeepVersions {
		c.logger.Info(fmt.Sprintf("No old PHP versions to remove (found %d versions, keeping %d)", len(versionDirs), options.KeepVersions))
		return nil
	}

	// Sort by modification time and keep the most recent
	type versionInfo struct {
		path    string
		modTime time.Time
	}

	var versionInfos []versionInfo
	for _, versionDir := range versionDirs {
		if info, err := os.Stat(versionDir); err == nil {
			versionInfos = append(versionInfos, versionInfo{
				path:    versionDir,
				modTime: info.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(versionInfos)-1; i++ {
		for j := i + 1; j < len(versionInfos); j++ {
			if versionInfos[i].modTime.Before(versionInfos[j].modTime) {
				versionInfos[i], versionInfos[j] = versionInfos[j], versionInfos[i]
			}
		}
	}

	// Remove old versions
	var totalSize int64
	var totalRemoved int

	for i := options.KeepVersions; i < len(versionInfos); i++ {
		versionDir := versionInfos[i].path
		size, err := c.getDirectorySize(versionDir)
		if err != nil {
			c.logger.Warn("Failed to calculate size", fmt.Sprintf("%s: %v", versionDir, err))
			continue
		}

		if options.Verbose {
			fmt.Printf("  Old version: %s (%s)\n", filepath.Base(versionDir), c.formatBytes(size))
		}

		if !options.DryRun {
			if options.Force || c.confirmDelete("PHP version", versionDir) {
				if err := os.RemoveAll(versionDir); err != nil {
					c.logger.Warn("Failed to remove PHP version", fmt.Sprintf("%s: %v", versionDir, err))
				} else {
					totalSize += size
					totalRemoved++
					if options.Verbose {
						fmt.Printf("  Deleted: %s\n", versionDir)
					}
				}
			}
		} else {
			// In dry-run mode, count what would be deleted
			totalSize += size
			totalRemoved++
		}
	}

	if totalRemoved > 0 {
		c.logger.Success("Old versions cleanup complete", fmt.Sprintf("%d versions, %s freed", totalRemoved, c.formatBytes(totalSize)))
	}

	return nil
}

// cleanSSLCerts removes stale SSL certificates
func (c *Cleaner) cleanSSLCerts(options *CleanOptions) error {
	// Find all linked projects to know which certs are active
	activeCerts := make(map[string]bool)

	linkedProjects, err := findAllLinkedProjects()
	if err == nil {
		for _, projectSlug := range linkedProjects {
			configPath := filepath.Join(c.projectsDir, projectSlug, "project.yaml")
			if projectConfig, err := projects.LoadConfig(configPath); err == nil {
				if projectConfig.Domains != nil {
					for _, domain := range projectConfig.Domains.Aliases {
						if domain.SSL {
							activeCerts[domain.Domain] = true
						}
					}
				}
				if projectConfig.Site != nil && projectConfig.Site.SSL {
					activeCerts[projectConfig.Site.Domain] = true
				}
			}
		}
	}

	sslDir := filepath.Join(c.workspaceDir, "ssl")
	sslInfo, err := os.Stat(sslDir)
	if err != nil || !sslInfo.IsDir() {
		c.logger.Info("No SSL certificates found")
		return nil
	}

	var totalSize int64
	var totalRemoved int

	err = filepath.Walk(sslDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a certificate file
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".crt" || ext == ".key" || ext == ".pem" {
			// Extract domain name from filename
			domain := strings.TrimSuffix(filepath.Base(path), ext)
			domain = strings.TrimSuffix(domain, ".key")
			domain = strings.TrimSuffix(domain, ".crt")

			// Check if this certificate is still active
			if !activeCerts[domain] {
				if options.Verbose {
					fmt.Printf("  Stale certificate: %s (%s)\n", filepath.Base(path), c.formatBytes(info.Size()))
				}

				if !options.DryRun {
					if options.Force || c.confirmDelete("SSL certificate", path) {
						if err := os.Remove(path); err != nil {
							c.logger.Warn("Failed to remove SSL certificate", fmt.Sprintf("%s: %v", path, err))
						} else {
							totalSize += info.Size()
							totalRemoved++
							if options.Verbose {
								fmt.Printf("  Deleted: %s\n", path)
							}
						}
					}
				} else {
					// In dry-run mode, count what would be deleted
					totalSize += info.Size()
					totalRemoved++
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk SSL directory: %w", err)
	}

	if totalRemoved > 0 {
		c.logger.Success("SSL certificate cleanup complete", fmt.Sprintf("%d files, %s freed", totalRemoved, c.formatBytes(totalSize)))
	} else {
		c.logger.Info("No stale SSL certificates found to clean")
	}

	return nil
}

// cleanProjects removes unlinked project directories
func (c *Cleaner) cleanProjects(options *CleanOptions) error {
	// Get all linked projects
	linkedProjects, err := findAllLinkedProjects()
	if err != nil {
		return fmt.Errorf("find linked projects: %w", err)
	}

	linkedSet := make(map[string]bool)
	for _, project := range linkedProjects {
		linkedSet[project] = true
	}

	projectsInfo, err := os.Stat(c.projectsDir)
	if err != nil || !projectsInfo.IsDir() {
		c.logger.Info("No projects directory found")
		return nil
	}

	projectDirs, err := os.ReadDir(c.projectsDir)
	if err != nil {
		return fmt.Errorf("read projects directory: %w", err)
	}

	var totalSize int64
	var totalRemoved int

	for _, projectDir := range projectDirs {
		if !projectDir.IsDir() {
			continue
		}

		projectSlug := projectDir.Name()
		if !linkedSet[projectSlug] {
			projectPath := filepath.Join(c.projectsDir, projectSlug)
			size, err := c.getDirectorySize(projectPath)
			if err != nil {
				c.logger.Warn("Failed to calculate size", fmt.Sprintf("%s: %v", projectPath, err))
				continue
			}

			if options.Verbose {
				fmt.Printf("  Unlinked project: %s (%s)\n", projectSlug, c.formatBytes(size))
			}

			if !options.DryRun {
				if options.Force || c.confirmDelete("unlinked project", projectPath) {
					if err := os.RemoveAll(projectPath); err != nil {
						c.logger.Warn("Failed to remove project", fmt.Sprintf("%s: %v", projectPath, err))
					} else {
						totalSize += size
						totalRemoved++
						if options.Verbose {
							fmt.Printf("  Deleted: %s\n", projectPath)
						}
					}
				}
			} else {
				// In dry-run mode, count what would be deleted
				totalSize += size
				totalRemoved++
			}
		}
	}

	if totalRemoved > 0 {
		c.logger.Success("Projects cleanup complete", fmt.Sprintf("%d projects, %s freed", totalRemoved, c.formatBytes(totalSize)))
	} else {
		c.logger.Info("No unlinked projects found to clean")
	}

	return nil
}

// cleanDirectory removes all files in a directory
func (c *Cleaner) cleanDirectory(dir string, options *CleanOptions, recursive bool) (int64, int, error) {
	var totalSize int64
	var totalFiles int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Check age if specified
		if options.OlderThan != "" {
			olderThan, err := time.ParseDuration(options.OlderThan)
			if err == nil {
				if time.Since(info.ModTime()) < olderThan {
					return nil // Skip recent files
				}
			}
		}

		// Only count files that are actually deleted
		if !options.DryRun {
			if options.Force || c.confirmDelete("file", path) {
				if err := os.Remove(path); err != nil {
					c.logger.Warn("Failed to remove file", fmt.Sprintf("%s: %v", path, err))
				} else {
					totalSize += info.Size()
					totalFiles++
					if options.Verbose {
						fmt.Printf("  Deleted: %s\n", path)
					}
				}
			}
		} else {
			// In dry-run mode, count what would be deleted
			totalSize += info.Size()
			totalFiles++
		}

		return nil
	})

	return totalSize, totalFiles, err
}

// getDirectorySize calculates the total size of a directory
func (c *Cleaner) getDirectorySize(dir string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
		}

		return nil
	})

	return totalSize, err
}

// confirmDelete asks for confirmation before deleting
func (c *Cleaner) confirmDelete(itemType, path string) bool {
	// Get file size
	var sizeStr string
	if info, err := os.Stat(path); err == nil {
		sizeStr = fmt.Sprintf(" (%s)", c.formatBytes(info.Size()))
	}

	fmt.Printf("Delete %s: %s%s? [y/N] ", itemType, path, sizeStr)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// formatBytes formats bytes in human readable format
func (c *Cleaner) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// printCleanUsage renders CLI help for the clean command
func printCleanUsage() {
	fmt.Println(`Usage: chauf clean [target] [options]

Clean up workspace files and free up disk space.

Targets:
  all             Clean everything (logs, cache, temp, old versions, stale certs)
  logs            Remove log files from nginx, php-fpm, and projects
  cache           Remove cache files (PHP opcache, composer cache, etc.)
  temp            Remove temporary files and directories
  old-versions    Remove old PHP versions (keep recent versions)
  ssl-certs       Remove SSL certificates for unlinked projects
  projects        Remove directories of unlinked projects

Options:
  --dry-run, -n           Show what would be deleted without actually deleting
  --force, -f             Skip confirmation prompts
  --verbose, -v           Show detailed output
  --older-than <duration> Only delete files older than specified duration
                          (e.g., 30d, 7d, 1h, 30m)
  --keep-versions <num>   Number of PHP versions to keep (default: 2)
  -h, --help             Show this message

Examples:
  chauf clean                    # Clean everything with confirmation
  chauf clean logs --force       # Remove all log files without confirmation
  chauf clean cache --dry-run    # Show what cache files would be deleted
  chauf clean old-versions --keep-versions 1    # Keep only latest PHP version
  chauf clean all --older-than 30d --force     # Clean files older than 30 days

Safety:
  - Always runs in dry-run mode first unless --force is used
  - Never removes active SSL certificates or linked projects
  - Keeps a specified number of recent PHP versions
  - Shows size of files to be deleted before confirmation`)
}