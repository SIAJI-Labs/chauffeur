package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunClean(args []string) error {
	if len(args) == 0 {
		return cleanHelp()
	}

	switch strings.ToLower(args[0]) {
	case "cache":
		return cleanCache(args[1:])
	case "logs":
		return cleanLogs(args[1:])
	case "all":
		return cleanAll(args[1:])
	case "--help", "-h", "help":
		return cleanHelp()
	default:
		return fmt.Errorf("unknown clean subcommand %q — run: chauf clean --help", args[0])
	}
}

// ── shared flags ───────────────────────────────────────────────────────────────

type cleanOpts struct {
	dryRun   bool
	olderThan time.Duration
}

func parseCleanFlags(name string, args []string) (cleanOpts, error) {
	flags := flag.NewFlagSet("clean "+name, flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf clean — remove cached downloads and logs", "chauf clean <cache|logs|all> [--dry-run] [--older-than <duration>]")

	dryRun := flags.Bool("dry-run", false, "Show what would be removed without deleting")
	olderThan := flags.String("older-than", "", "Only remove files older than this duration (e.g. 7d, 24h, 30m)")

	if err := flags.Parse(args); err != nil {
		return cleanOpts{}, err
	}

	opts := cleanOpts{dryRun: *dryRun}
	if *olderThan != "" {
		d, err := parseDuration(*olderThan)
		if err != nil {
			return cleanOpts{}, fmt.Errorf("invalid --older-than value %q: use e.g. 7d, 24h, 30m", *olderThan)
		}
		opts.olderThan = d
	}
	return opts, nil
}

// ── clean cache ────────────────────────────────────────────────────────────────

func cleanCache(args []string) error {
	opts, err := parseCleanFlags("cache", args)
	if err != nil {
		return err
	}

	root := workspace.Root()
	cacheDir := filepath.Join(root, "cache")

	fmt.Println()
	return cleanDir("cache", cacheDir, opts, func(name string) bool {
		// Include all files in cache directory
		return true
	})
}

// ── clean logs ─────────────────────────────────────────────────────────────────

func cleanLogs(args []string) error {
	opts, err := parseCleanFlags("logs", args)
	if err != nil {
		return err
	}

	root := workspace.Root()

	// Nginx logs: ~/.chauffeur/nginx/logs/*.log
	// PHP-FPM logs: ~/.chauffeur/php/<version>/runtime/php-fpm/*.log
	logDirs := []string{
		filepath.Join(root, "nginx", "logs"),
	}

	// Add all PHP version log dirs
	phpDir := filepath.Join(root, "php")
	if entries, err := os.ReadDir(phpDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				logDirs = append(logDirs, filepath.Join(phpDir, e.Name(), "runtime", "php-fpm"))
			}
		}
	}

	fmt.Println()
	total := 0
	totalSize := int64(0)
	for _, dir := range logDirs {
		n, sz, err := cleanDirLogs(dir, opts)
		if err != nil {
			lib.Warn(fmt.Sprintf("Skipping %s: %s", dir, err.Error()))
			continue
		}
		total += n
		totalSize += sz
	}

	if opts.dryRun {
		if total == 0 {
			lib.Info("No log files to remove.")
		} else {
			fmt.Printf("  Would remove %d log file(s) (%s)\n", total, lib.FormatBytes(totalSize))
		}
	} else {
		if total == 0 {
			lib.Info("No log files to remove.")
		} else {
			lib.Success(fmt.Sprintf("Removed %d log file(s) (%s)", total, lib.FormatBytes(totalSize)))
		}
	}
	fmt.Println()
	return nil
}

// ── clean all ──────────────────────────────────────────────────────────────────

func cleanAll(args []string) error {
	opts, err := parseCleanFlags("all", args)
	if err != nil {
		return err
	}

	fmt.Println()
	lib.Info(lib.Bold("Cleaning cache..."))
	root := workspace.Root()

	cacheOpts := opts
	_ = cleanDir("cache", filepath.Join(root, "cache"), cacheOpts, func(string) bool { return true })

	fmt.Println()
	lib.Info(lib.Bold("Cleaning logs..."))
	_ = cleanLogs(buildFlagArgs(opts))

	return nil
}

// ── internal helpers ───────────────────────────────────────────────────────────

// cleanDir removes (or previews) files in dir matching the filter.
func cleanDir(label, dir string, opts cleanOpts, keep func(name string) bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			lib.Info(fmt.Sprintf("%-10s  %s", label, lib.Gray("(empty)")))
			return nil
		}
		return err
	}

	cutoff := time.Time{}
	if opts.olderThan > 0 {
		cutoff = time.Now().Add(-opts.olderThan)
	}

	var removed int
	var removedSize int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !keep(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !cutoff.IsZero() && info.ModTime().After(cutoff) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		sz := info.Size()
		if opts.dryRun {
			fmt.Printf("  %s  %s  %s\n", lib.Gray("(dry-run)"), lib.FormatBytes(sz), lib.Gray(shortenHome(path)))
		} else {
			if err := os.Remove(path); err != nil {
				lib.Warn("Could not remove " + path + ": " + err.Error())
				continue
			}
			fmt.Printf("  %s  %s  %s\n", lib.Red("✗"), lib.FormatBytes(sz), lib.Gray(shortenHome(path)))
		}
		removed++
		removedSize += sz
	}

	if removed == 0 {
		lib.Info(fmt.Sprintf("%-10s  %s", label, lib.Gray("(nothing to clean)")))
		return nil
	}

	verb := "Removed"
	if opts.dryRun {
		verb = "Would remove"
	}
	lib.Success(fmt.Sprintf("%s %d file(s) (%s) from %s", verb, removed, lib.FormatBytes(removedSize), label))
	return nil
}

// cleanDirLogs removes .log files from a directory, returning count and total size.
func cleanDirLogs(dir string, opts cleanOpts) (int, int64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	cutoff := time.Time{}
	if opts.olderThan > 0 {
		cutoff = time.Now().Add(-opts.olderThan)
	}

	var count int
	var totalSize int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !cutoff.IsZero() && info.ModTime().After(cutoff) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		sz := info.Size()
		if opts.dryRun {
			fmt.Printf("  %s  %s  %s\n", lib.Gray("(dry-run)"), lib.FormatBytes(sz), lib.Gray(shortenHome(path)))
		} else {
			if err := os.Remove(path); err != nil {
				lib.Warn("Could not remove " + path + ": " + err.Error())
				continue
			}
			fmt.Printf("  %s  %s  %s\n", lib.Red("✗"), lib.FormatBytes(sz), lib.Gray(shortenHome(path)))
		}
		count++
		totalSize += sz
	}
	return count, totalSize, nil
}

// parseDuration supports "7d" in addition to standard Go duration strings.
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid day count")
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// buildFlagArgs reconstructs flag args from cleanOpts (for cleanAll delegation).
func buildFlagArgs(opts cleanOpts) []string {
	var args []string
	if opts.dryRun {
		args = append(args, "--dry-run")
	}
	if opts.olderThan > 0 {
		args = append(args, "--older-than", fmt.Sprintf("%dh", int(opts.olderThan.Hours())))
	}
	return args
}

// ── help ───────────────────────────────────────────────────────────────────────

func cleanHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf clean — workspace cleanup"))
	fmt.Printf("  %-22s  %s\n", "clean cache", lib.Gray("Remove downloaded source archives"))
	fmt.Printf("  %-22s  %s\n", "clean logs", lib.Gray("Remove nginx and PHP-FPM log files"))
	fmt.Printf("  %-22s  %s\n", "clean all", lib.Gray("Run all clean operations"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Flags:"))
	fmt.Printf("  %-22s  %s\n", "--dry-run", lib.Gray("Show what would be removed"))
	fmt.Printf("  %-22s  %s\n", "--older-than <age>", lib.Gray("Only remove files older than age (e.g. 7d, 24h)"))
	fmt.Println()
	return nil
}
