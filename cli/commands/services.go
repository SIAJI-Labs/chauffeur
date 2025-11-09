package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/installers"
	"github.com/siaji/chauffeur/cli/internal/system"
)

var serviceNames = []string{"caddy", "nginx", "php", "composer"}

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

type ServiceSpec struct {
	Name        string
	Description string
	BinaryPath  string
	installFunc func(force bool) error
}

// NewServiceSpec returns metadata and installer wiring for a named service.
func NewServiceSpec(name, prefix string, info system.Info) (ServiceSpec, error) {
	switch name {
	case "caddy":
		target := filepath.Join(prefix, "caddy", "bin", "caddy")
		return ServiceSpec{
			Name:        "caddy",
			Description: "verified tarball from GitHub releases",
			BinaryPath:  target,
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
		return ServiceSpec{
			Name:        "nginx",
			Description: "source build from nginx.org release",
			BinaryPath:  target,
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
		return ServiceSpec{
			Name:        "php",
			Description: "default PHP CLI managed by Chauffeur",
			BinaryPath:  target,
		}, nil
	case "composer":
		target := filepath.Join(prefix, "bin", "composer")
		return ServiceSpec{
			Name:        "composer",
			Description: "PHP dependency manager with Chauffeur PHP version isolation",
			BinaryPath:  target,
			installFunc: func(force bool) error {
				return installers.InstallComposer(installers.InstallOptions{
					Prefix: prefix,
					Force:  force,
					Info:   info,
				})
			},
		}, nil
	default:
		return ServiceSpec{}, fmt.Errorf("unknown service %q", name)
	}
}

// available reports whether the service binary already exists on disk.
func (s ServiceSpec) available() (bool, error) {
	info, err := os.Stat(s.BinaryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat %s: %w", s.BinaryPath, err)
	}
	return info.Mode().IsRegular(), nil
}

// install executes the underlying installer, optionally forcing a reinstall.
func (s ServiceSpec) install(force bool) error {
	if s.installFunc == nil {
		return fmt.Errorf("installer for %s is not defined", s.Name)
	}
	return s.installFunc(force)
}
