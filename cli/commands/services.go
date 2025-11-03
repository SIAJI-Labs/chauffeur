package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/system"
)

var serviceNames = []string{"caddy", "nginx", "php"}

// KnownServices returns a copy of registered service names.
func KnownServices() []string {
	names := make([]string, len(serviceNames))
	copy(names, serviceNames)
	return names
}

// IsKnownService reports whether name is a supported Chauffeur service.
func IsKnownService(name string) bool {
	for _, svc := range serviceNames {
		if svc == name {
			return true
		}
	}
	return false
}

type serviceSpec struct {
	name        string
	description string
	binaryPath  string
	installFunc func(force bool) error
}

// newServiceSpec returns metadata and installer wiring for a named service.
func newServiceSpec(name, prefix string, info system.Info) (serviceSpec, error) {
	switch name {
	case "caddy":
		target := filepath.Join(prefix, "caddy", "bin", "caddy")
		return serviceSpec{
			name:        "caddy",
			description: "verified tarball from GitHub releases",
			binaryPath:  target,
			installFunc: func(force bool) error {
				return installers.InstallCaddyTarball(installers.InstallOptions{
					Prefix: prefix,
					Force:  force,
					Info:   info,
				})
			},
		}, nil
	case "nginx":
		target := filepath.Join(prefix, "nginx", "sbin", "nginx")
		return serviceSpec{
			name:        "nginx",
			description: "source build from nginx.org release",
			binaryPath:  target,
			installFunc: func(force bool) error {
				return installers.InstallNginxSource(installers.InstallOptions{
					Prefix: prefix,
					Force:  force,
					Info:   info,
				})
			},
		}, nil
	case "php":
		target := filepath.Join(prefix, "bin", "php")
		return serviceSpec{
			name:        "php",
			description: "default PHP CLI managed by Chauffeur",
			binaryPath:  target,
		}, nil
	default:
		return serviceSpec{}, fmt.Errorf("unknown service %q", name)
	}
}

// available reports whether the service binary already exists on disk.
func (s serviceSpec) available() (bool, error) {
	info, err := os.Stat(s.binaryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat %s: %w", s.binaryPath, err)
	}
	return info.Mode().IsRegular(), nil
}

// install executes the underlying installer, optionally forcing a reinstall.
func (s serviceSpec) install(force bool) error {
	if s.installFunc == nil {
		return fmt.Errorf("installer for %s is not defined", s.name)
	}
	return s.installFunc(force)
}
