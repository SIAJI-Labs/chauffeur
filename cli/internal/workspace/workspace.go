package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

const workspaceDirName = ".chauffeur"

// Dir returns the root Chauffeur workspace path (defaults to ~/.chauffeur).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, workspaceDirName), nil
}

// Path joins sub paths under the workspace root.
func Path(parts ...string) (string, error) {
	root, err := Dir()
	if err != nil {
		return "", err
	}
	all := append([]string{root}, parts...)
	return filepath.Join(all...), nil
}

// Ensure sets up required directories if they do not exist.
func Ensure(paths ...string) error {
	root, err := Dir()
	if err != nil {
		return err
	}

	toEnsure := []string{root}
	toEnsure = append(toEnsure, paths...)

	for _, rel := range toEnsure {
		var target string
		if filepath.IsAbs(rel) {
			target = rel
		} else {
			target = filepath.Join(root, rel)
		}

		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("ensure directory %s: %w", target, err)
		}
	}
	return nil
}
