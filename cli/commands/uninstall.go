package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/lib"
)

// RunUninstall handles `chauf uninstall` logic.
func RunUninstall(args []string) error {
	logger := lib.NewCommandLogger("uninstall")
	purge := false

	for _, arg := range args {
		switch arg {
		case "--purge":
			purge = true
		case "--help", "-h":
			printUninstallUsage()
			return nil
		default:
			return fmt.Errorf("unknown flag for uninstall: %s", arg)
		}
	}

	workspace, err := defaultWorkspace()
	if err != nil {
		return err
	}

	info, err := os.Stat(workspace)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Info(fmt.Sprintf("No Chauffeur workspace found at %s", workspace))
			reportPathRemoval()
			return nil
		}
		return fmt.Errorf("inspect workspace: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("workspace path %s is not a directory", workspace)
	}

	if purge {
		if err := os.RemoveAll(workspace); err != nil {
			return fmt.Errorf("purge workspace: %w", err)
		}
		logger.Success(fmt.Sprintf("Removed Chauffeur workspace at %s (purged runtimes and caches)", workspace), "")
		reportPathRemoval()
		return nil
	}

	entries, err := os.ReadDir(workspace)
	if err != nil {
		return fmt.Errorf("read workspace entries: %w", err)
	}

	keepDirs := map[string]struct{}{
		"php": {},
		// TODO: remove retained runtimes/caches once separate cleanup is implemented.
	}

	retained := false
	for _, entry := range entries {
		name := entry.Name()
		if _, keep := keepDirs[name]; keep {
			retained = true
			continue
		}

		target := filepath.Join(workspace, name)
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove %s: %w", target, err)
		}
	}

	if !retained {
		if err := os.RemoveAll(workspace); err != nil {
			return fmt.Errorf("remove workspace: %w", err)
		}
		logger.Success(fmt.Sprintf("Removed Chauffeur workspace at %s", workspace), "")
		reportPathRemoval()
		return nil
	}

	logger.Info(fmt.Sprintf("Cleaned Chauffeur workspace at %s (retained runtimes in %s)", workspace, filepath.Join(workspace, "php")))
	reportPathRemoval()
	return nil
}

func printUninstallUsage() {
	fmt.Println(`Usage: chauf uninstall [--purge]

Options:
  --purge    Remove the workspace and delete cached PHP runtimes.
  -h, --help Show this message.`)
}

func defaultWorkspace() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	return filepath.Join(home, ".chauffeur"), nil
}

func reportPathRemoval() {
	logger := lib.NewCommandLogger("uninstall")
	removed, err := removePathExports()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update shell PATH entries: %v\n", err)
		return
	}

	if len(removed) == 0 {
		logger.Info("No Chauffeur PATH entries found in shell profiles.")
		return
	}

	for _, file := range removed {
		logger.Info(fmt.Sprintf("Removed Chauffeur PATH entry from %s", file))
	}
	logger.Info("Reload your shell to drop ~/.chauffeur/bin from PATH.")
}

func removePathExports() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determine home directory: %w", err)
	}

	rcFiles := []string{".bashrc", ".zshrc"}
	targetLine := `export PATH="$HOME/.chauffeur/bin:$PATH"`

	var updated []string

	for _, rc := range rcFiles {
		path := filepath.Join(home, rc)

		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return updated, fmt.Errorf("read %s: %w", path, err)
		}

		info, err := os.Stat(path)
		if err != nil {
			return updated, fmt.Errorf("stat %s: %w", path, err)
		}

		lines := strings.Split(string(data), "\n")
		hasTrailingNewline := len(data) == 0 || data[len(data)-1] == '\n'

		var kept []string
		changed := false

		for _, line := range lines {
			if strings.TrimSpace(line) == targetLine {
				changed = true
				continue
			}
			kept = append(kept, line)
		}

		if !changed {
			continue
		}

		var buf bytes.Buffer
		for i, line := range kept {
			if i > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(line)
		}
		if hasTrailingNewline && buf.Len() > 0 {
			buf.WriteByte('\n')
		}

		if err := os.WriteFile(path, buf.Bytes(), info.Mode().Perm()); err != nil {
			return updated, fmt.Errorf("write %s: %w", path, err)
		}

		updated = append(updated, path)
	}

	return updated, nil
}
