package workspace

import (
	"os"
	"path/filepath"
)

// Root returns the Chauffeur workspace directory.
// Respects $CHAUFFEUR_HOME if set; defaults to ~/.chauffeur.
func Root() string {
	if h := os.Getenv("CHAUFFEUR_HOME"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".chauffeur")
}

// Path returns a path inside the workspace root.
func Path(parts ...string) string {
	return filepath.Join(append([]string{Root()}, parts...)...)
}
