package installers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/lib"
)

const (
	composerBinaryName = "composer"
	composerPharName   = "composer.phar"
)

// InstallComposer downloads Composer, verifies integrity, and installs a project-aware shim.
func InstallComposer(opts InstallOptions) error {
	logger := lib.NewCommandLogger("composer")

	if opts.Prefix == "" {
		return logger.Fail("install prefix is required", "")
	}

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	logger.Info("Resolving Composer release metadata")
	release, err := releases.LatestGitHubRelease(client, "composer", "composer")
	if err != nil {
		logger.Warn("Failed to query GitHub releases", err.Error())
	}

	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		version = "latest-stable"
	}
	logger.Info(fmt.Sprintf("Selected Composer release: %s", version))

	downloadURL, checksumURLs := resolveComposerAssets(release, version)
	if downloadURL == "" {
		return logger.Fail("resolve composer release", "no download URL available")
	}
	if len(checksumURLs) == 0 {
		return logger.Fail("resolve composer release", "no checksum URLs available")
	}
	logger.Info(fmt.Sprintf("Download URL: %s", downloadURL))
	logger.Info(fmt.Sprintf("Checksum sources: %s", strings.Join(checksumURLs, ", ")))

	composerDir := filepath.Join(opts.Prefix, "composer")
	targetPhar := filepath.Join(composerDir, composerPharName)
	shimPath := filepath.Join(opts.Prefix, "bin", "shims", composerBinaryName)
	entryPoint := filepath.Join(opts.Prefix, "bin", composerBinaryName)

	if !opts.Force {
		if info, err := os.Stat(entryPoint); err == nil && info.Mode().IsRegular() {
			logger.Info("Composer already installed (use --force to reinstall)")
			return nil
		}
	}

	if err := os.MkdirAll(composerDir, 0o755); err != nil {
		return logger.Fail("create composer directory", err.Error())
	}

	tmpDir, err := os.MkdirTemp("", "chauffeur-composer-*")
	if err != nil {
		return logger.Fail("create temp dir", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	downloadPath := filepath.Join(tmpDir, composerPharName)
	checksumPath := filepath.Join(tmpDir, composerPharName+".sha256")

	downloadLogger := logger.NewChildLogger("download")
	size, err := lib.DownloadToFileWithLogger(client, downloadURL, downloadPath, fmt.Sprintf("Download %s", composerPharName), downloadLogger)
	if err != nil {
		return logger.Fail("download composer", err.Error())
	}
	downloadLogger.Success(fmt.Sprintf("Downloaded %s", composerPharName), fmt.Sprintf("%d bytes", size))

	if err := downloadComposerChecksum(client, checksumURLs, checksumPath, downloadLogger); err != nil {
		return logger.Fail("download checksum", err.Error())
	}

	if err := verifyComposerChecksum(downloadPath, checksumPath); err != nil {
		return logger.Fail("verify composer checksum", err.Error())
	}
	downloadLogger.Success("Checksum verification passed", "")

	if err := copyFile(downloadPath, targetPhar); err != nil {
		return logger.Fail("install composer phar", err.Error())
	}

	if err := writeComposerShim(opts.Prefix); err != nil {
		return logger.Fail("write composer shim", err.Error())
	}

	if err := ensureComposerEntryPoint(shimPath, entryPoint); err != nil {
		return logger.Fail("link composer shim", err.Error())
	}

	logger.Success("Composer installation complete", fmt.Sprintf("version %s", version))
	logger.Info("Use 'chauf install php <version>' to control the PHP runtime Composer executes")
	return nil
}

func resolveComposerAssets(release releases.GitHubRelease, version string) (string, []string) {
	var downloadURL string
	var checksumURLs []string

	tag := strings.TrimPrefix(release.TagName, "v")
	if tag == "" {
		tag = version
	}

	pharCandidates := []string{"composer.phar"}
	if tag != "" && tag != "latest-stable" {
		pharCandidates = append([]string{fmt.Sprintf("composer-%s.phar", tag)}, pharCandidates...)
	}

	checksumCandidates := []string{
		"composer.phar.sha256",
		"composer.phar.sha256sum",
		"composer.phar.sha256sum.txt",
		"sha256sum.txt",
	}

	for _, candidate := range pharCandidates {
		if url, ok := release.AssetURL(candidate); ok {
			downloadURL = url
			break
		}
	}

	for _, candidate := range checksumCandidates {
		if url, ok := release.AssetURL(candidate); ok {
			checksumURLs = append(checksumURLs, url)
		}
	}

	var base string
	switch {
	case tag == "" || tag == "latest-stable":
		base = "https://getcomposer.org/download/latest-stable"
	default:
		base = fmt.Sprintf("https://getcomposer.org/download/%s", tag)
	}

	if downloadURL == "" {
		downloadURL = fmt.Sprintf("%s/%s", base, composerPharName)
	}

	if len(checksumURLs) == 0 {
		checksumURLs = append(checksumURLs,
			fmt.Sprintf("%s/%s", base, "composer.phar.sha256sum"),
			fmt.Sprintf("%s/%s", base, "composer.phar.sha256sum.txt"),
			fmt.Sprintf("%s/%s", base, "composer.phar.sha256"),
		)
	}

	return downloadURL, checksumURLs
}

func verifyComposerChecksum(pharPath, checksumPath string) error {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("read checksum file: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return fmt.Errorf("checksum file is empty")
	}
	expected := strings.TrimSpace(fields[0])

	file, err := os.Open(pharPath)
	if err != nil {
		return fmt.Errorf("open phar: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("hash phar: %w", err)
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func downloadComposerChecksum(client *http.Client, candidates []string, dest string, logger *lib.Logger) error {
	var lastErr error
	for _, url := range candidates {
		logger.Info(fmt.Sprintf("Attempting checksum download: %s", url))

		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("unexpected status %s from %s", resp.Status, url)
			resp.Body.Close()
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if err := os.WriteFile(dest, data, 0o644); err != nil {
			lastErr = err
			continue
		}
		logger.Success("Downloaded checksum", url)
		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no checksum URLs available")
	}
	return lastErr
}

func writeComposerShim(prefix string) error {
	shimsDir := filepath.Join(prefix, "bin", "shims")
	if err := os.MkdirAll(shimsDir, 0o755); err != nil {
		return fmt.Errorf("create shims directory: %w", err)
	}

	composerShim := filepath.Join(shimsDir, composerBinaryName)
	chaufHome := prefix

	content := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

CHAUF_HOME="${CHAUF_HOME:-%s}"
COMPOSER_PHAR="$CHAUF_HOME/composer/%s"
PHP_SHIM="$CHAUF_HOME/bin/php"

if [[ -x "$PHP_SHIM" ]]; then
	exec "$PHP_SHIM" "$COMPOSER_PHAR" "$@"
fi

if command -v php >/dev/null 2>&1; then
	exec "$(command -v php)" "$COMPOSER_PHAR" "$@"
fi

echo "Unable to locate a PHP binary for Composer" >&2
echo "Install PHP via 'chauf install php <version>' or expose php on PATH" >&2
exit 1
`, chaufHome, composerPharName)

	if err := os.WriteFile(composerShim, []byte(content), 0o755); err != nil {
		return fmt.Errorf("write composer shim: %w", err)
	}
	return nil
}

func ensureComposerEntryPoint(shimPath, entryPoint string) error {
	if err := os.MkdirAll(filepath.Dir(entryPoint), 0o755); err != nil {
		return fmt.Errorf("create bin directory: %w", err)
	}

	if err := os.Remove(entryPoint); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous composer binary: %w", err)
	}

	if err := os.Symlink(shimPath, entryPoint); err == nil {
		return nil
	}

	// Fall back to copying when symlinks are not available
	content, err := os.ReadFile(shimPath)
	if err != nil {
		return fmt.Errorf("read shim content: %w", err)
	}
	if err := os.WriteFile(entryPoint, content, 0o755); err != nil {
		return fmt.Errorf("write composer entry point: %w", err)
	}
	return nil
}

// GetSupportedComposerVersions returns the Composer major versions supported by Chauffeur.
func GetSupportedComposerVersions() []string {
	return []string{"2.7", "2.6", "2.5", "2.4", "2.3", "2.2"}
}

// IsComposerVersionSupported checks if the provided Composer version is part of the supported list.
func IsComposerVersionSupported(version string) bool {
	for _, supported := range GetSupportedComposerVersions() {
		if supported == version {
			return true
		}
	}
	return false
}

// GetSupportedComposerVersionsList returns the supported Composer versions as a comma-separated list.
func GetSupportedComposerVersionsList() string {
	return strings.Join(GetSupportedComposerVersions(), ", ")
}
