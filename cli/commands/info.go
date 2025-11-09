package commands

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/internal/workspace"
	"github.com/siaji/chauffeur/cli/lib"
)

var fetchLatestCLIVersion = defaultFetchLatestCLIVersion

// RunInfo reports Chauffeur environment details.
func RunInfo(args []string) error {
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printInfoUsage()
			return nil
		default:
			return fmt.Errorf("unknown flag for info: %s", arg)
		}
	}

	logger := lib.NewCommandLogger("info")

	wsDir, err := workspace.Dir()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		execPath = "unknown"
	} else if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolved
	}

	logger.PrintSection("Environment")
	logger.Info(fmt.Sprintf("Workspace: %s", wsDir))
	logger.Info(fmt.Sprintf("Binary: %s", execPath))
	logger.Info(fmt.Sprintf("Projects dir: %s", cfg.ProjectsDir))
	logger.Info(fmt.Sprintf("Config file: %s", filepath.Join(wsDir, "config", "chauffeur.yaml")))

	logger.PrintSection("Versions")
	currentVersion := getCLIVersion()
	logger.Info(fmt.Sprintf("Current CLI: %s", currentVersion))

	if version, err := fetchLatestCLIVersion(); err == nil && version != "" {
		status := "up to date"
		if !strings.EqualFold(strings.TrimPrefix(version, "v"), strings.TrimPrefix(currentVersion, "v")) {
			status = "update available"
		}
		logger.Info(fmt.Sprintf("Latest release: %s (%s)", version, status))
	} else if err != nil {
		logger.Warn("Failed to fetch remote version", err.Error())
	}

	logger.PrintSection("Managed Services")
	prefix := wsDir

	for _, svc := range gatherBinaryServices(prefix) {
		logger.Info(fmt.Sprintf("%s: %s", svc.Name, svc.Status))
	}

	if phpStatus := describePHP(prefix, cfg.PHP.Default); phpStatus != "" {
		logger.Info(phpStatus)
	}

	logger.PrintSection("Port Configuration")
	logger.Info(fmt.Sprintf("Caddy HTTP: %d", cfg.Caddy.HTTPPort))
	logger.Info(fmt.Sprintf("Caddy HTTPS: %d", cfg.Caddy.HTTPSPort))
	logger.Info(fmt.Sprintf("Nginx HTTP: %d", cfg.Nginx.HTTPPort))
	logger.Info(fmt.Sprintf("Nginx HTTPS: %d", cfg.Nginx.HTTPSPort))
	logger.Info(fmt.Sprintf("Port range: %d-%d (%s)", cfg.Ports.StartRange, cfg.Ports.EndRange, strings.ToUpper(cfg.Ports.ConflictResolution)))
	logger.Info(fmt.Sprintf("PHP-FPM fallback: %d", cfg.Ports.PHPFPMFallback))

	return nil
}

func printInfoUsage() {
	fmt.Print(`Chauffeur Environment Information

Usage:
  chauf info

Description:
  Displays the current Chauffeur installation details, CLI version, managed services,
  and configured ports/directories.
`)
}

type serviceStatus struct {
	Name   string
	Status string
}

func gatherBinaryServices(prefix string) []serviceStatus {
	services := []struct {
		name string
		path string
		args []string
	}{
		{"Caddy", filepath.Join(prefix, "caddy", "bin", "caddy"), []string{"version"}},
		{"Nginx", filepath.Join(prefix, "nginx", "sbin", "nginx"), []string{"-v"}},
		{"Composer", filepath.Join(prefix, "bin", "composer"), []string{"--version"}},
	}

	var statuses []serviceStatus
	for _, svc := range services {
		statuses = append(statuses, describeBinaryService(svc.name, svc.path, svc.args...))
	}

	return statuses
}

func describeBinaryService(name, path string, args ...string) serviceStatus {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return serviceStatus{Name: name, Status: "not installed"}
		}
		return serviceStatus{Name: name, Status: fmt.Sprintf("error checking binary: %v", err)}
	}

	output, err := runCommandOutput(path, args...)
	if err != nil {
		return serviceStatus{Name: name, Status: fmt.Sprintf("installed (version check failed: %v)", err)}
	}

	lines := strings.Split(output, "\n")
	version := strings.TrimSpace(lines[0])
	if version == "" {
		version = "installed"
	}
	return serviceStatus{Name: name, Status: fmt.Sprintf("%s (%s)", version, path)}
}

func describePHP(prefix, defaultVersion string) string {
	phpDir := filepath.Join(prefix, "php")
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "PHP: no runtimes installed"
		}
		return fmt.Sprintf("PHP: error reading runtimes (%v)", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	if len(versions) == 0 {
		return "PHP: no runtimes installed"
	}

	sort.Strings(versions)
	for i, v := range versions {
		if v == defaultVersion {
			versions[i] = fmt.Sprintf("%s (default)", v)
		}
	}

	return fmt.Sprintf("PHP: %s", strings.Join(versions, ", "))
}

func runCommandOutput(binary string, args ...string) (string, error) {
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w (%s)", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func defaultFetchLatestCLIVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	release, err := releases.LatestGitHubRelease(client, "SIAJI-Labs", "chauffeur")
	if err != nil {
		return "", err
	}
	return release.TagName, nil
}

// OverrideLatestVersionFetcher allows tests to inject a deterministic version fetcher.
func OverrideLatestVersionFetcher(fn func() (string, error)) {
	if fn == nil {
		fetchLatestCLIVersion = defaultFetchLatestCLIVersion
		return
	}
	fetchLatestCLIVersion = fn
}
