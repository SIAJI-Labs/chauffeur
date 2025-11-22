package lib

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrHelpRequested is returned when --help or -h flag is encountered.
var ErrHelpRequested = errors.New("help requested")

// FlagSet wraps standard library flag.FlagSet with Chauffeur-specific enhancements.
type FlagSet struct {
	*flag.FlagSet
	logger *Logger
	output io.Writer // allows redirecting output for testing
}

// NewFlagSet creates a new FlagSet with the given name and logger.
func NewFlagSet(name string, logger *Logger) *FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() {} // Suppress default usage message
	return &FlagSet{
		FlagSet: fs,
		logger:  logger,
		output:  os.Stderr, // Default to stderr for errors/usage
	}
}

// Parse parses command-line flags from args.
// It specifically handles --help/-h and returns ErrHelpRequested.
func (fs *FlagSet) Parse(args []string) error {
	// Custom help handling
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return ErrHelpRequested
		}
	}

	// Redirect FlagSet output to a buffer to prevent it from printing directly
	// unless there's an actual error.
	var buf bytes.Buffer
	fs.SetOutput(&buf)

	err := fs.FlagSet.Parse(args)
	if err != nil {
		// If there was an error, print the usage and the error message
		fs.printUsage()
		return fmt.Errorf("%s: %w", buf.String(), err)
	}

	// If there's content in the buffer (e.g., error messages from flag.FlagSet),
	// it means something went wrong, and we should display it.
	if buf.Len() > 0 {
		fs.logger.Error("flag parsing error", buf.String())
	}

	return nil
}

// printUsage prints the usage message for the FlagSet.
func (fs *FlagSet) printUsage() {
	if fs.logger != nil {
		var buf bytes.Buffer
		fs.SetOutput(&buf) // Capture FlagSet's default usage output
		fs.FlagSet.PrintDefaults()
		fs.SetOutput(fs.output) // Restore output

		// Print a more Chauffeur-like usage if a logger is available
		fs.logger.Info(fmt.Sprintf("Usage for %s:", fs.Name()))
		scanner := bufio.NewScanner(&buf)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				fs.logger.Info(fmt.Sprintf("  %s", line))
			}
		}
		if err := scanner.Err(); err != nil {
			fs.logger.Error("error scanning flag usage", err.Error())
		}

	} else {
		fs.FlagSet.PrintDefaults() // Fallback to default if no logger
	}
}

// SetOutput sets the output writer for the FlagSet.
func (fs *FlagSet) SetOutput(w io.Writer) {
	fs.output = w
	fs.FlagSet.SetOutput(w)
}
