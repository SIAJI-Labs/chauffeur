package lib

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Verbose enables step-level output. Set by the global --verbose / -v flag.
var Verbose bool

var IsTTY = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}()

func ansi(code, s string) string {
	if !IsTTY {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

// Color primitives
func Bold(s string) string   { return ansi("1", s) }
func Gray(s string) string   { return ansi("90", s) }
func Cyan(s string) string   { return ansi("36", s) }
func Red(s string) string    { return ansi("31", s) }
func Green(s string) string  { return ansi("32", s) }
func Yellow(s string) string { return ansi("33", s) }

// Print helpers
func Success(msg string) { fmt.Printf("  %s  %s\n", Green("✓"), msg) }
func Warn(msg string)    { fmt.Printf("  %s  %s\n", Yellow("⚠"), msg) }
func Error(msg string)   { fmt.Fprintf(os.Stderr, "  %s  %s\n", Red("✗"), msg) }
func Info(msg string)    { fmt.Printf("  %s\n", msg) }

// Pair prints an indented label-value line with consistent column alignment.
func Pair(key, val string) { fmt.Printf("  %-14s  %s\n", key, val) }

// SetFlagUsage sets a styled usage function on fs. Commands should call this
// immediately after flag.NewFlagSet so that -h / --help prints a proper header
// instead of the raw "Usage of <name>:" default.
func SetFlagUsage(fs *flag.FlagSet, title, synopsis string) {
	fs.Usage = func() {
		w := fs.Output()
		fmt.Fprintf(w, "\n%s\n\n", Bold(title))
		fmt.Fprintf(w, "  %s\n\n", Gray("Usage: "+synopsis))
		fs.PrintDefaults()
		fmt.Fprintln(w)
	}
}

// Step prints a detail line visible only when Verbose is true.
func Step(msg string) {
	if !Verbose {
		return
	}
	fmt.Printf("  %s  %s\n", Gray("→"), Gray(msg))
}

// DirSize returns the total size in bytes of all files under path.
func DirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// FormatBytes formats a byte count as a human-readable string.
func FormatBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
