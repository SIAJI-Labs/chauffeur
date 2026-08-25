package podman

import (
	"os"
	"path/filepath"

	"github.com/siegg/chauffeur/internal/workspace"
)

const (
	// PodmanRoot is the directory under ~/.chauffeur where all podman data lives.
	PodmanRoot = ".chauffeur/podman"

	// ConfigDir is the directory containing per-engine config files.
	ConfigDir = ""

	// VolumesDir is the directory containing volume data for each engine.
	VolumesDir = "volumes"

	// NetworkName is the shared Podman network name.
	NetworkName = "chauf-net"
)

// Root returns the podman root directory path (~/.chauffeur/podman).
func Root() string {
	return filepath.Join(workspace.Root(), "podman")
}

// ConfigPath returns the path to the config file for a given container name.
// Configs are keyed by container name for intuitive file matching.
func ConfigPath(containerName string) string {
	return filepath.Join(Root(), containerName+".yaml")
}

// GlobalConfigPath returns the path to the global podman config.
func GlobalConfigPath() string {
	return filepath.Join(Root(), "config.yaml")
}

// VolumesPath returns the volumes directory path for a given container.
func VolumesPath(containerName string) string {
	return filepath.Join(Root(), VolumesDir, containerName)
}

// EnsureRoot creates the podman root directory and volumes subdirectories if they don't exist.
func EnsureRoot() error {
	if err := os.MkdirAll(Root(), 0755); err != nil {
		return err
	}
	return nil
}

// EnsureVolumePath creates the volume directory for a given container.
func EnsureVolumePath(containerName string) error {
	return os.MkdirAll(VolumesPath(containerName), 0755)
}
