package installers

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/lib"
)

const phpBinaryName = "php"
const (
	openssl111wVersion        = "1.1.1w"
	openssl111wTarball        = "openssl-" + openssl111wVersion + ".tar.gz"
	openssl111wPrimaryURL     = "https://www.openssl.org/source/" + openssl111wTarball
	openssl111wFallbackURL    = "https://www.openssl.org/source/old/1.1.1/" + openssl111wTarball
	openssl111wExpectedSHA256 = "cf3098950cb4d853ad95c0841f1f9c6d3dc102dccfcacd521d93925208b76ac8"
)

/**
 * PHPSigningKey represents a trusted PHP signing key and its metadata.
 */
type phpSigningKey struct {
	Name        string
	URL         string
	Fingerprint string
	UIDs        []string
	Optional    bool
}

var phpSigningKeys = []phpSigningKey{
	// PHP 7.4 keys
	{
		Name:        "Derick Rethans (PHP 7.4)",
		URL:         "", // Will use keyring
		Fingerprint: "5A52880781F755608BF815FC910DEB46F53EA312",
		UIDs:        []string{"Derick Rethans <derick@php.net>"},
		Optional:    false,
	},
	{
		Name:        "Peter Kokot (PHP 7.4)",
		URL:         "", // Will use keyring
		Fingerprint: "42670A7FE4D0441C8E4632349E4FDC074A4EF02D",
		UIDs:        []string{"Peter Kokot <petk@php.net>"},
		Optional:    false,
	},
	// PHP 8.0+ current release managers
	{
		Name:        "Pierrick Charron",
		URL:         "", // Will use keyring
		Fingerprint: "1198C0117593497A5EC5C199286AF1F9897469DC",
		UIDs:        []string{"Pierrick Charron <pierrick@php.net>"},
		Optional:    false,
	},
	{
		Name:        "Eric A Mann",
		URL:         "", // Will use keyring
		Fingerprint: "AFD8691FDAEDF03BDF6E460563F15A9B715376CA",
		UIDs:        []string{"Eric A Mann <eric@sixthree.me>", "Eric Mann <ericmann@php.net>"},
		Optional:    false,
	},
	{
		Name:        "Jakub Zelenka",
		URL:         "", // Will use keyring
		Fingerprint: "C28D937575603EB4ABB725861C0779DC5C0A9DE4",
		UIDs:        []string{"Jakub Zelenka <bukka@php.net>"},
		Optional:    false,
	},
}

/**
 * PHPVersion represents supported PHP versions with their metadata.
 */
type PHPVersion struct {
	Version     string
	EndOfLife   bool
	MinimumGCC  string
	RequiredTLS string
}

type pkgRequirement struct {
	Name            string
	Package         string
	MinVersion      string
	BlockedVersions []string
}

// legacyDependency represents version-specific dependency constraints for legacy PHP versions
type legacyDependency struct {
	PHPVersion      string
	PackageName     string
	MinVersion      string
	MaxVersion      string   // Upper bound for legacy compatibility
	BlockedVersions []string // Versions explicitly blocked for this PHP version
}

const (
	defaultImagickVersion = "3.8.0"
	imagickIniContent     = "extension=imagick\n"
)

var phpPkgRequirements = []pkgRequirement{
	{Name: "libzip", Package: "libzip", MinVersion: "0.11", BlockedVersions: []string{"1.3.1", "1.7.0"}},
	{Name: "libjpeg", Package: "libjpeg"},
	{Name: "libpng", Package: "libpng"},
	{Name: "freetype", Package: "freetype2"},
	{Name: "libxml2", Package: "libxml-2.0"},
	{Name: "libcurl", Package: "libcurl"},
	{Name: "zlib", Package: "zlib"},
	{Name: "libxslt", Package: "libxslt"},
	{Name: "readline", Package: "readline"},
	{Name: "ImageMagick (MagickWand)", Package: "MagickWand"},
	{Name: "GMP", Package: "gmp"},
	{Name: "libsodium", Package: "libsodium"},
}

// legacyDependencyMatrix defines version-specific dependency constraints for legacy PHP versions
var legacyDependencyMatrix = []legacyDependency{
	// PHP 7.4 has stricter compatibility requirements due to age
	{
		PHPVersion:  "7.4",
		PackageName: "libxml-2.0",
		MaxVersion:  "2.16.99", // Updated: libxml2 2.15+ works fine with PHP 7.4 with proper compilation
	},
	{
		PHPVersion:  "7.4",
		PackageName: "libcurl",
		MaxVersion:  "9.99.0", // Updated: Newer libcurl versions work fine with PHP 7.4
	},
	{
		PHPVersion:  "7.4",
		PackageName: "MagickWand",
		MinVersion:  "6.9.0",
		MaxVersion:  "7.99.0", // Updated: ImageMagick 7.1+ works fine with PHP 7.4
	},

	// PHP 8.0 has some constraints but is more flexible than 7.4
	{
		PHPVersion:  "8.0",
		PackageName: "libxml-2.0",
		MaxVersion:  "2.16.99", // Updated: PHP 8.0 works well with newer libxml2 versions
	},
	{
		PHPVersion:  "8.0",
		PackageName: "MagickWand",
		MinVersion:  "6.9.0",
	},
}

/**
 * GetSupportedPHPVersions returns the list of supported PHP versions.
 *
 * @return list of supported PHP versions
 */
func GetSupportedPHPVersions() []PHPVersion {
	return []PHPVersion{
		{Version: "8.4", EndOfLife: false, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
		{Version: "8.3", EndOfLife: false, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
		{Version: "8.2", EndOfLife: false, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
		{Version: "8.1", EndOfLife: false, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
		{Version: "8.0", EndOfLife: true, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
		{Version: "7.4", EndOfLife: true, MinimumGCC: "4.8", RequiredTLS: "1.1.1"},
	}
}

/**
 * IsPHPVersionSupported checks if the requested PHP version is supported.
 *
 * @param version PHP version string (major.minor)
 * @return true if supported, false otherwise
 */
func IsPHPVersionSupported(version string) bool {
	for _, v := range GetSupportedPHPVersions() {
		if v.Version == version {
			return true
		}
	}
	return false
}

// isLegacyPHPVersion checks if a PHP version requires legacy dependency handling
func isLegacyPHPVersion(version string) bool {
	legacyVersions := []string{"7.4", "8.0"}
	for _, v := range legacyVersions {
		if version == v {
			return true
		}
	}
	return false
}

// getLegacyDependencyRequirement returns legacy-specific constraints for a PHP version and package
func getLegacyDependencyRequirement(phpVersion, packageName string) (legacyDependency, bool) {
	for _, dep := range legacyDependencyMatrix {
		if dep.PHPVersion == phpVersion && dep.PackageName == packageName {
			return dep, true
		}
	}
	return legacyDependency{}, false
}

/**
 * GetSupportedVersionsList returns a comma-separated list of supported PHP versions.
 *
 * @return string of supported versions
 */
func GetSupportedVersionsList() string {
	versions := GetSupportedPHPVersions()
	var versionStrings []string
	for _, v := range versions {
		if v.EndOfLife {
			versionStrings = append(versionStrings, fmt.Sprintf("%s (EOL)", v.Version))
		} else {
			versionStrings = append(versionStrings, v.Version)
		}
	}
	return strings.Join(versionStrings, ", ")
}

/**
 * InstallPHPSource compiles PHP from source into the Chauffeur workspace.
 *
 * @param version Target PHP version (e.g., "8.3", "8.2", "7.4")
 * @param opts    Installer configuration containing prefix, force flag, and client.
 * @return error when installation cannot complete successfully.
 */
func InstallPHPSource(version string, opts InstallOptions) (err error) {
	phpLogger := lib.NewCommandLogger("php")

	if opts.Prefix == "" {
		return phpLogger.Fail("install prefix is required", "")
	}

	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 60 * time.Second}
	}

	// Validate PHP version
	if !IsPHPVersionSupported(version) {
		return phpLogger.Fail(fmt.Sprintf("PHP version %s is not supported", version), GetSupportedVersionsList())
	}

	phpLogger.Info("Checking host dependencies")
	depsLogger := phpLogger.NewChildLogger("deps")
	if err := ensurePHPBuildDependenciesForVersion(depsLogger, version); err != nil {
		return err
	}

	defer func() {
		if err != nil && opts.Prefix != "" {
			if logPath, logErr := logToolFailure(opts.Prefix, "php", "install", version, err); logErr != nil {
				phpLogger.Warn("failed to write php installer log", logErr.Error())
			} else {
				phpLogger.Info(fmt.Sprintf("Installation failed. See %s for full log.", logPath))
			}
		}
	}()

	phpLogger.Info("Preparing")
	prepareLogger := phpLogger.NewChildLogger("prepare")
	prepareLogger.Info(fmt.Sprintf("Target PHP version: %s", version))
	prepareLogger.Info(fmt.Sprintf("Architecture: %s", opts.Info.Arch))
	prepareLogger.Info(fmt.Sprintf("Install prefix: %s", filepath.Join(opts.Prefix, "php", version)))

	installDir := filepath.Join(opts.Prefix, "php", version)
	binaryPath := filepath.Join(installDir, "bin", phpBinaryName)
	if !opts.Force {
		if info, err := os.Stat(binaryPath); err == nil && info.Mode().IsRegular() {
			prepareLogger.Info(fmt.Sprintf("PHP %s is already installed", version))
			return nil
		}
	}

	tmpDir, err := os.MkdirTemp("", "chauffeur-php-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	// Determine if we need to preserve build directory for GD extension build
	needsGDExtension := (version == "7.4" || version == "8.0") && opts.EnableGD
	keepBuildDir := os.Getenv("CHAUFFEUR_KEEP_BUILD_DIR") == "1" || needsGDExtension

	if !keepBuildDir {
		defer os.RemoveAll(tmpDir)
	} else {
		if needsGDExtension {
			phpLogger.Info("preserving PHP build directory for GD extension build")
		} else {
			phpLogger.Warn("preserving PHP build directory for debugging", "")
		}
	}

	// Get latest patch version for the requested major.minor
	patchVersion, err := getLatestPHPPatchVersion(opts.Client, version)
	if err != nil {
		return fmt.Errorf("resolve latest PHP %s patch version: %w", version, err)
	}

	prepareLogger.Info(fmt.Sprintf("Latest patch version: %s", patchVersion))

	tarballName := fmt.Sprintf("php-%s.tar.gz", patchVersion)

	// Try official mirrors
	tarballURLs := []string{
		fmt.Sprintf("https://www.php.net/distributions/%s", tarballName),
		fmt.Sprintf("https://secure.php.net/distributions/%s", tarballName),
	}

	var tarballURL string
	var downloadErr error

	phpLogger.Info("Downloading")
	downloadLogger := phpLogger.NewChildLogger("download")
	tarballPath := filepath.Join(tmpDir, tarballName)

	// Check environment variable first (for --local flag)
	if localTarball := os.Getenv("CHAUFFEUR_PHP_TARBALL"); localTarball != "" {
		if err := copyFile(localTarball, tarballPath); err != nil {
			downloadLogger.Warn("Failed to copy local tarball", err.Error())
		} else {
			tarballURL = "local"
			downloadLogger.Success(fmt.Sprintf("Reused local %s", tarballName), localTarball)
		}
	}

	// Check config for stored local tarball paths if not already found
	if tarballURL == "" {
		if configPath, err := checkConfigLocalTarball(version, downloadLogger); err == nil && configPath != "" {
			if err := copyFile(configPath, tarballPath); err != nil {
				downloadLogger.Warn("Failed to copy config tarball", err.Error())
			} else {
				tarballURL = "config-local"
				downloadLogger.Success(fmt.Sprintf("Reused configured local %s", tarballName), configPath)
			}
		}
	}

	// Check universal cache for downloaded files if not already found
	if tarballURL == "" {
		cachedPath := checkForCachedFile("php", tarballName)
		if cachedPath != "" {
			if err := copyFile(cachedPath, tarballPath); err != nil {
				downloadLogger.Warn("Failed to copy cached PHP file", err.Error())
			} else {
				tarballURL = "cache"
				downloadLogger.Success(fmt.Sprintf("Reused cached %s", tarballName), cachedPath)
			}
		}
	}

	if tarballURL == "" {
		downloadLogger.Info(fmt.Sprintf("Attempting download of %s", tarballName))
		for i, url := range tarballURLs {
			downloadLogger.Info(fmt.Sprintf("Trying mirror %d/%d: %s", i+1, len(tarballURLs), url))
			size, err := lib.DownloadToFileWithLogger(opts.Client, url, tarballPath, fmt.Sprintf("Download %s", tarballName), downloadLogger)
			if err == nil {
				tarballURL = url
				downloadLogger.Success(fmt.Sprintf("Downloaded %s from %s", tarballName, url), lib.HumanBytes(size))

				// Auto-cache successful downloads (unless explicitly disabled)
				if os.Getenv("CHAUFFEUR_NO_CACHE") != "1" {
					if err := cacheDownloadedTarball(tarballPath, version, tarballName, downloadLogger); err != nil {
						downloadLogger.Warn("Failed to cache downloaded file", err.Error())
					} else {
						downloadLogger.Success("Download cached for future use", "")
					}
				}
				break
			}
			downloadLogger.Warn(fmt.Sprintf("Mirror %d failed: %s", i+1, url), err.Error())
			downloadErr = err
		}
	}

	if tarballURL == "" {
		downloadLogger.Warn("All download attempts failed", downloadErr.Error())
		return fmt.Errorf("all download attempts failed: %w", downloadErr)
	}

	phpLogger.Info("Verifying")
	verificationLogger := phpLogger.NewChildLogger("verifying")
	// Add spinner for GPG verification
	verificationSpin := lib.NewSpinner("verify", "Validating GPG signature")
	if err := verifyPHPSignature(verificationLogger, opts.Client, tarballPath, tarballName, tmpDir); err != nil {
		verificationSpin.Fail("GPG signature verification failed")
		return phpLogger.Fail("verify GPG signature", err.Error())
	}
	verificationSpin.Success("GPG validation complete")

	phpLogger.Info("Building")
	buildLogger := phpLogger.NewChildLogger("build")
	extractRoot := filepath.Join(tmpDir, "src")
	sourceDir := filepath.Join(extractRoot, fmt.Sprintf("php-%s", patchVersion))
	buildLogger.Info(fmt.Sprintf("Extracting sources to %s", extractRoot))
	if err := untarPHP(tarballPath, extractRoot); err != nil {
		return phpLogger.Fail("extract php source", err.Error())
	}
	buildLogger.Success("Sources extracted", "")

	if err := applyLegacyPHPSourcePatches(version, sourceDir, buildLogger); err != nil {
		return phpLogger.Fail("apply compatibility patches", err.Error())
	}

	spin := lib.NewSpinner("install", "Configuring and compiling PHP")
	if err := buildAndInstallPHP(opts, version, sourceDir, buildLogger); err != nil {
		spin.Fail("compilation failed")
		return phpLogger.Fail("configure and compile PHP", err.Error())
	}
	spin.Success("PHP compiled and installed")
	buildLogger.Success(fmt.Sprintf("PHP %s built and installed to %s", version, filepath.Join(opts.Prefix, "php", version)), "")

	// Build bundled GD extension for legacy PHP versions if requested
	if (version == "7.4" || version == "8.0") && opts.EnableGD {
		gdLogger := phpLogger.NewChildLogger("gd")
		gdLogger.Info("Building bundled GD extension...")
		if err := buildBundledGDExtension(sourceDir, installDir, version, gdLogger); err != nil {
			gdLogger.Warn("Failed to build bundled GD extension", err.Error())
			gdLogger.Info("PHP installation completed without GD support")
		} else {
			gdLogger.Success("Bundled GD extension built and enabled", "")
		}

		// Clean up temporary build directory after GD extension build is complete
		// Check if we should preserve directory (from CHAUFFEUR_KEEP_BUILD_DIR or GD requirements)
		shouldPreserve := os.Getenv("CHAUFFEUR_KEEP_BUILD_DIR") == "1" || ((version == "7.4" || version == "8.0") && opts.EnableGD)
		if !shouldPreserve {
			sourceDirParent := filepath.Dir(sourceDir) // Get the temp directory path
			if err := os.RemoveAll(sourceDirParent); err != nil {
				phpLogger.Warn("Failed to clean up temporary build directory", err.Error())
			} else {
				gdLogger.Info("Cleaned up temporary build directory")
			}
		}
	}

	extensionsLogger := phpLogger.NewChildLogger("extensions")
	if err := installImagickExtension(opts, version, installDir, extensionsLogger); err != nil {
		return err
	}

	phpLogger.Info("Finalizing")
	finalizeLogger := phpLogger.NewChildLogger("finalize")
	finalizeLogger.Info("Ensuring workspace layout")
	if err := ensurePHPlayout(opts.Prefix, version); err != nil {
		return err
	}
	finalizeLogger.Success("Runtime layout ready", "")

	shimName := fmt.Sprintf("php-%s", version)
	finalizeLogger.Info("Writing PHP shim")
	if err := writeShim(opts.Prefix, shimName, binaryPath); err != nil {
		return err
	}
	finalizeLogger.Success(fmt.Sprintf("Shim written to %s", filepath.Join(opts.Prefix, "bin", shimName)), "")

	defaultShim := filepath.Join(opts.Prefix, "bin", "php")
	if _, err := os.Stat(defaultShim); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			finalizeLogger.Info(fmt.Sprintf("Setting default PHP shim to version %s", version))
			if err := UpdateDefaultPHPShim(opts.Prefix, version); err != nil {
				return err
			}
			finalizeLogger.Success(fmt.Sprintf("Default PHP shim now targets %s", version), "")
		} else {
			return fmt.Errorf("stat default php shim: %w", err)
		}
	}

	if cfg, err := config.Load(); err != nil {
		finalizeLogger.Warn("unable to load config", err.Error())
	} else if cfg.PHP.Default == "" {
		if err := config.SetDefaultPHPVersion(version); err != nil {
			finalizeLogger.Warn("failed to set default PHP version", err.Error())
		} else {
			finalizeLogger.Success(fmt.Sprintf("Default PHP version set to %s", version), "")
		}
	}

	finalizeLogger.Info("Writing PHP-FPM configuration")
	if err := writeDefaultPHPFPMConf(opts.Prefix, version); err != nil {
		return err
	}
	finalizeLogger.Success("PHP-FPM configuration ready", "")

	finalizeLogger.Info("Setting up PHP upload limits")
	if err := writeUploadLimitsConf(opts.Prefix, version); err != nil {
		return err
	}
	finalizeLogger.Success("PHP upload limits configured (256MB)", "")

	finalizeLogger.Info("Setting up OpenSSL certificate configuration")
	if err := WriteOpenSSLConf(opts.Prefix, version); err != nil {
		return err
	}
	finalizeLogger.Success("OpenSSL certificate verification configured", "")

	return nil
}

/**
 * PHPVersionInfo represents the structure of PHP version information from the API.
 */
type PHPVersionInfo struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Source  []struct {
		Filename string `json:"filename"`
		Name     string `json:"name"`
		SHA256   string `json:"sha256"`
		Date     string `json:"date"`
	} `json:"source"`
}

/**
 * checkConfigLocalTarball checks the config for a local tarball path and validates it.
 *
 * @param version PHP version to check for
 * @param logger Logger instance for reporting
 * @return Path to valid local tarball or empty string if not found/invalid
 */
func checkConfigLocalTarball(version string, logger *lib.Logger) (string, error) {
	configPath, err := getLocalTarballFromConfig(version)
	if err != nil {
		return "", err
	}

	if configPath == "" {
		// Don't log anything here - the universal cache check happens next
		return "", nil
	}

	// Validate the configured path exists and is accessible
	if _, err := os.Stat(configPath); err != nil {
		logger.Warn("Configured tarball not accessible", err.Error())
		logger.Info("Removing invalid path from config")
		// Remove invalid path from config
		if removeErr := removeLocalTarballFromConfig(version); removeErr != nil {
			logger.Warn("Failed to remove invalid path from config", removeErr.Error())
		}
		return "", nil
	}

	// Validate tarball version matches
	if err := validatePHPTarball(configPath, version, logger); err != nil {
		logger.Warn("Configured tarball validation failed", err.Error())
		logger.Info("Removing invalid path from config")
		// Remove mismatched path from config
		if removeErr := removeLocalTarballFromConfig(version); removeErr != nil {
			logger.Warn("Failed to remove mismatched path from config", removeErr.Error())
		}
		return "", nil
	}

	logger.Success("Found valid local tarball in config", configPath)
	return configPath, nil
}

/**
 * hasInternetConnection checks if there's internet connectivity by testing a reliable endpoint.
 *
 * @param client HTTP client for requests
 * @return true if connection is available, false otherwise
 */
func hasInternetConnection(client *http.Client) bool {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	// Test connectivity to a reliable endpoint with a quick HEAD request
	req, err := http.NewRequest("HEAD", "https://www.php.net", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

/**
 * fetchLatestPHPVersionFromAPI attempts to fetch the latest PHP version information from PHP.net API.
 *
 * @param client  HTTP client for requests
 * @param version Major.minor version (e.g., "8.3")
 * @return Latest full version string (e.g., "8.3.14") and error
 */
func fetchLatestPHPVersionFromAPI(client *http.Client, version string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	// Use the active releases API which provides latest versions for each supported series
	apiURL := "https://www.php.net/releases/active.php?json=1"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "chauffeur-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch API data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API returned status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read API response: %w", err)
	}

	// Parse the JSON response
	var apiData map[string]map[string]PHPVersionInfo
	if err := json.Unmarshal(body, &apiData); err != nil {
		return "", fmt.Errorf("parse API response: %w", err)
	}

	// Look for the major version in the API response
	for majorVersion, versions := range apiData {
		if !strings.HasPrefix(majorVersion, version) {
			continue
		}

		// Find the exact version match
		if versionInfo, exists := versions[version]; exists {
			return versionInfo.Version, nil
		}
	}

	return "", fmt.Errorf("PHP %s not found in API response", version)
}

/**
 * GetLatestPHPPatchVersionForTesting is an exported version for testing purposes.
 * This function is only used in tests and should not be called from production code.
 *
 * @param client  HTTP client for requests
 * @param version Major.minor version (e.g., "8.3")
 * @return Latest full version string (e.g., "8.3.14") and error
 */
func GetLatestPHPPatchVersionForTesting(client *http.Client, version string) (string, error) {
	return getLatestPHPPatchVersion(client, version)
}

/**
 * validatePHPTarball checks if the tarball matches the expected PHP version.
 * @param tarballPath Path to the tarball file
 * @param expectedVersion Expected PHP version (e.g., "8.3")
 * @param logger Logger instance
 * @return error if validation fails
 */
func validatePHPTarball(tarballPath, expectedVersion string, logger *lib.Logger) error {
	// Extract filename and check version
	filename := filepath.Base(tarballPath)

	// Check file extension first
	lowerFilename := strings.ToLower(filename)
	if !strings.HasSuffix(lowerFilename, ".tar.gz") &&
		!strings.HasSuffix(lowerFilename, ".tgz") &&
		!strings.HasSuffix(lowerFilename, ".tar.bz2") &&
		!strings.HasSuffix(lowerFilename, ".tar.xz") {
		return fmt.Errorf("invalid file extension, expected .tar.gz, .tgz, .tar.bz2, or .tar.xz")
	}

	// Try to extract version from filename
	// Common patterns: php-8.3.27.tar.gz, php-8.3.27.tgz, etc.
	var extractedVersion string

	// Remove file extensions
	baseName := strings.TrimSuffix(filename, ".tar.gz")
	baseName = strings.TrimSuffix(baseName, ".tgz")
	baseName = strings.TrimSuffix(baseName, ".tar.bz2")
	baseName = strings.TrimSuffix(baseName, ".tar.xz")

	// Remove php- prefix
	if strings.HasPrefix(baseName, "php-") {
		extractedVersion = strings.TrimPrefix(baseName, "php-")
	} else if strings.HasPrefix(baseName, "php") {
		extractedVersion = strings.TrimPrefix(baseName, "php")
	}

	// Validate extracted version matches expected major.minor
	if extractedVersion == "" {
		return fmt.Errorf("could not extract version from filename: %s", filename)
	}

	// Extract major.minor from extracted version
	parts := strings.Split(extractedVersion, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid version format in filename: %s", extractedVersion)
	}

	majorMinor := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if majorMinor != expectedVersion {
		return fmt.Errorf("version mismatch: filename contains PHP %s but requested PHP %s", majorMinor, expectedVersion)
	}

	return nil
}

/**
 * ValidatePHPTarballForTesting is an exported version for testing purposes.
 * This function is only used in tests and should not be called from production code.
 *
 * @param tarballPath Path to the tarball file
 * @param expectedVersion Expected PHP version (e.g., "8.3")
 * @param logger Logger instance (can be nil for testing)
 * @return error if validation fails
 */
func ValidatePHPTarballForTesting(tarballPath, expectedVersion string, logger interface{}) error {
	// Extract filename and check version
	filename := filepath.Base(tarballPath)

	// Check file extension first
	lowerFilename := strings.ToLower(filename)
	if !strings.HasSuffix(lowerFilename, ".tar.gz") &&
		!strings.HasSuffix(lowerFilename, ".tgz") &&
		!strings.HasSuffix(lowerFilename, ".tar.bz2") &&
		!strings.HasSuffix(lowerFilename, ".tar.xz") {
		return fmt.Errorf("invalid file extension, expected .tar.gz, .tgz, .tar.bz2, or .tar.xz")
	}

	// Try to extract version from filename
	// Common patterns: php-8.3.27.tar.gz, php-8.3.27.tgz, etc.
	var extractedVersion string

	// Remove file extensions
	baseName := strings.TrimSuffix(filename, ".tar.gz")
	baseName = strings.TrimSuffix(baseName, ".tgz")
	baseName = strings.TrimSuffix(baseName, ".tar.bz2")
	baseName = strings.TrimSuffix(baseName, ".tar.xz")

	// Remove php- prefix
	if strings.HasPrefix(baseName, "php-") {
		extractedVersion = strings.TrimPrefix(baseName, "php-")
	} else if strings.HasPrefix(baseName, "php") {
		extractedVersion = strings.TrimPrefix(baseName, "php")
	}

	// Validate extracted version matches expected major.minor
	if extractedVersion == "" {
		return fmt.Errorf("could not extract version from filename: %s", filename)
	}

	// Extract major.minor from extracted version
	parts := strings.Split(extractedVersion, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid version format in filename: %s", extractedVersion)
	}

	majorMinor := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if majorMinor != expectedVersion {
		return fmt.Errorf("version mismatch: filename contains PHP %s but requested PHP %s", majorMinor, expectedVersion)
	}

	return nil
}

// Known stable versions for all services (used when offline)
var knownServiceVersions = map[string]map[string]string{
	"php": {
		"8.4": "8.4.14",
		"8.3": "8.3.27",
		"8.2": "8.2.29",
		"8.1": "8.1.33",
		"8.0": "8.0.30",
		"7.4": "7.4.33",
	},
	"composer": {
		"2.8": "2.8.4",
		"2.7": "2.7.9",
		"2.6": "2.6.6",
		"2.5": "2.5.8",
		"2.4": "2.4.4",
		"2.3": "2.3.17",
		"2.2": "2.2.22",
	},
	"nginx": {
		"1.29": "1.29.3",
		"1.28": "1.28.2",
		"1.27": "1.27.3",
		"1.26": "1.26.2",
		"1.25": "1.25.5",
		"1.24": "1.24.0",
		"1.23": "1.23.4",
		"1.22": "1.22.1",
	},
}

/**
 * getLatestServiceVersion gets the latest version for any service.
 * It first attempts API fetching when available, falling back to known versions.
 *
 * @param client HTTP client for requests
 * @param service Service name (e.g., "php", "composer", "nginx")
 * @param version Optional version constraint (e.g., "2.7" for Composer)
 * @return Latest version string and error
 */
func getLatestServiceVersion(client *http.Client, service string, version string) (string, error) {
	// Check if we have known versions for this service
	if serviceVersions, exists := knownServiceVersions[service]; exists {
		// If no specific version requested, get the latest known version
		if version == "" {
			// Try to fetch from API first for the absolute latest
			if hasInternetConnection(client) {
				if apiVersion, err := fetchLatestVersionFromAPI(client, service, ""); err == nil {
					return apiVersion, nil
				}
			}
			// Fallback: return the first (latest) version in the map
			for latestVersion := range serviceVersions {
				return serviceVersions[latestVersion], nil
			}
		}

		// Try to get specific version if available
		if latestVersion, exists := serviceVersions[version]; exists {
			return latestVersion, nil
		}
	}

	// Fallback: try to fetch from API if we have internet
	if hasInternetConnection(client) {
		if apiVersion, err := fetchLatestVersionFromAPI(client, service, version); err == nil {
			return apiVersion, nil
		}
	}

	return "", fmt.Errorf("no version information available for %s %s", service, version)
}

/**
 * fetchLatestVersionFromAPI attempts to fetch the latest version from service APIs.
 *
 * @param client HTTP client for requests
 * @param service Service name
 * @param version Optional version constraint
 * @return Latest version string and error
 */
func fetchLatestVersionFromAPI(client *http.Client, service, version string) (string, error) {
	switch service {
	case "php":
		return fetchLatestPHPVersionFromAPI(client, version)
	case "composer":
		return fetchLatestComposerVersion(client)
	case "nginx":
		return fetchLatestNginxVersion(client)
	default:
		return "", fmt.Errorf("unsupported service for API version checking: %s", service)
	}
}

/**
 * fetchLatestComposerVersion gets the latest Composer version from GitHub API.
 *
 * @param client HTTP client for requests
 * @return Latest version string and error
 */
func fetchLatestComposerVersion(client *http.Client) (string, error) {
	release, err := releases.LatestGitHubRelease(client, "composer", "composer")
	if err != nil {
		return "", fmt.Errorf("fetch Composer releases: %w", err)
	}

	// Extract version from tag (e.g., "v2.8.4" -> "2.8.4")
	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		return "", fmt.Errorf("invalid Composer release tag: %s", release.TagName)
	}

	return version, nil
}

/**
 * fetchLatestNginxVersion gets the latest Nginx version from GitHub API.
 *
 * @param client HTTP client for requests
 * @return Latest version string and error
 */
func fetchLatestNginxVersion(client *http.Client) (string, error) {
	release, err := releases.LatestGitHubRelease(client, "nginx", "nginx")
	if err != nil {
		return "", fmt.Errorf("fetch Nginx releases: %w", err)
	}

	// Extract version from tag (e.g., "release-1.27.3" -> "1.27.3")
	version := strings.TrimPrefix(release.TagName, "release-")
	if version == "" {
		version = strings.TrimPrefix(release.TagName, "v")
	}
	if version == "" {
		return "", fmt.Errorf("invalid Nginx release tag: %s", release.TagName)
	}

	return version, nil
}

/**
 * getLatestPHPPatchVersion discovers the latest patch version for a given major.minor.
 * It first attempts to fetch from PHP.net API, falling back to hardcoded versions if offline.
 *
 * @param client  HTTP client for requests
 * @param version Major.minor version (e.g., "8.3")
 * @return Latest full version string (e.g., "8.3.14") and error
 */
func getLatestPHPPatchVersion(client *http.Client, version string) (string, error) {
	// Use the universal version detection
	return getLatestServiceVersion(client, "php", version)
}

/**
 * verifyPHPSignature validates the PHP tarball using detached GPG signatures and trusted keys.
 *
 * @param client      HTTP client used for downloads.
 * @param tarballPath Local path to the downloaded tarball.
 * @param tarballName Tarball file name (used for forming URLs).
 * @param workDir     Workspace directory for temporary signature/key storage.
 * @return error when signature verification fails.
 */
func verifyPHPSignature(logger *lib.Logger, client *http.Client, tarballPath, tarballName, workDir string) error {
	if _, err := exec.LookPath("gpg"); err != nil {
		return fmt.Errorf("gpg not found in PATH: %w", err)
	}

	sigURL := fmt.Sprintf("https://www.php.net/distributions/%s.asc", tarballName)
	sigPath := tarballPath + ".asc"
	if sigSource := os.Getenv("CHAUFFEUR_PHP_SIGNATURE"); sigSource != "" {
		if err := copyFile(sigSource, sigPath); err != nil {
			return fmt.Errorf("copy signature: %w", err)
		}
	} else {
		// Signature download happens silently, progress shown via progress bar
		if _, err := lib.DownloadToFileWithLogger(client, sigURL, sigPath, fmt.Sprintf("Signature %s", tarballName), logger); err != nil {
			return fmt.Errorf("download signature: %w", err)
		}
	}

	gpgHome, err := os.MkdirTemp(workDir, "gpg-")
	if err != nil {
		return fmt.Errorf("create gpg homedir: %w", err)
	}
	if err := os.Chmod(gpgHome, 0o700); err != nil {
		return fmt.Errorf("chmod gpg homedir: %w", err)
	}
	defer os.RemoveAll(gpgHome)

	keysDir := filepath.Join(workDir, "keys")
	if err := os.MkdirAll(keysDir, 0o755); err != nil {
		return fmt.Errorf("create keys dir: %w", err)
	}

	// Download and import the complete PHP keyring
	keyringURL := "https://www.php.net/distributions/php-keyring.gpg"
	keyringPath := filepath.Join(keysDir, "php-keyring.gpg")
	if keyringSource := os.Getenv("CHAUFFEUR_PHP_KEYRING"); keyringSource != "" {
		if err := copyFile(keyringSource, keyringPath); err != nil {
			return fmt.Errorf("copy PHP keyring: %w", err)
		}
	} else {
		// Keyring download happens silently, progress shown via progress bar
		if _, err := lib.DownloadToFileWithLogger(client, keyringURL, keyringPath, "PHP Keyring", logger); err != nil {
			return fmt.Errorf("download PHP keyring: %w", err)
		}
	}

	// Import the entire keyring
	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--no-autostart",
		"--pinentry-mode", "loopback",
		"--batch",
		"--import", keyringPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("import PHP keyring: %w\n%s", err, out)
	}
	logPHPSuccess(logger, "PHP keyring imported successfully")

	if err := verifyPHPSignatureWithGPG(gpgHome, tarballPath, sigPath, logger); err != nil {
		return err
	}
	logPHPSuccess(logger, "GPG signature verified successfully")
	return nil
}

/**
 * verifyPHPSignatureWithGPG validates the PHP tarball signature using the imported keyring.
 *
 * @param gpgHome     Isolated gpg home directory populated with PHP keyring.
 * @param tarballPath Path to the downloaded tarball.
 * @param sigPath     Path to the detached signature file.
 * @return error when signature validation fails.
 */
func verifyPHPSignatureWithGPG(gpgHome, tarballPath, sigPath string, logger *lib.Logger) error {
	if logger == nil {
		logger = lib.NewCommandLogger("install")
	}
	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--no-autostart",
		"--pinentry-mode", "loopback",
		"--batch",
		"--verify", sigPath, tarballPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpg signature verification failed: %w\n%s", err, out)
	}

	// Since we're using the official PHP keyring, any valid signature is acceptable
	logger.Success("GPG signature verified successfully", "")
	return nil
}

/**
 * buildAndInstallPHP configures, builds, and installs PHP into the workspace.
 *
 * @param prefix    Chauffeur workspace root.
 * @param version   PHP version string (major.minor).
 * @param sourceDir Directory containing the extracted PHP sources.
 * @return error when configure/build/install steps fail.
 */
func buildAndInstallPHP(opts InstallOptions, version, sourceDir string, logger *lib.Logger) error {
	if logger == nil {
		logger = lib.NewCommandLogger("install")
	}
	prefix := opts.Prefix
	installDir := filepath.Join(prefix, "php", version)

	confArgs := []string{
		fmt.Sprintf("--prefix=%s", installDir),
		"--enable-debug",
		"--enable-cli",
		"--enable-fpm",
		"--enable-mysqlnd",
		"--with-fpm-user=" + getPHPUser(),
		"--with-fpm-group=" + getPHPUser(),
		"--with-config-file-path=" + filepath.Join(installDir, "etc"),
		"--with-config-file-scan-dir=" + filepath.Join(installDir, "etc", "conf.d"),
		"--enable-mbstring",
		"--with-openssl",
		"--enable-json",
		"--enable-tokenizer",
		"--enable-xml",
		"--with-libxml",
		"--enable-simplexml",
		"--enable-dom",
		"--enable-filter",
		"--with-curl",
		"--with-zlib",
		"--with-bz2",
		"--enable-ctype",
		"--with-zip",
		"--enable-gd",
		"--with-jpeg=/usr",
		"--with-freetype=/usr",
		"--enable-exif",
		"--enable-posix",
		"--with-pear",
		"--with-readline",
		"--with-mysqli=mysqlnd",
		"--with-pdo-mysql=mysqlnd",
		"--enable-gmp",
		"--enable-bcmath",
	}

	// Add sodium support for legacy PHP versions (7.4, 8.0)
	// PHP 8.1+ has native sodium support
	if version == "7.4" || version == "8.0" {
		confArgs = append(confArgs, "--with-sodium")
	}

	// Handle GD-related options for legacy versions
	switch version {
	case "8.0":
		if opts.EnableGD {
			// User wants GD - we'll build it as bundled extension later
			var filteredArgs []string
			gdOptions := map[string]bool{
				"--enable-gd":          true,
				"--with-jpeg=/usr":     true,
				"--with-freetype=/usr": true,
			}
			for _, arg := range confArgs {
				if !gdOptions[arg] {
					filteredArgs = append(filteredArgs, arg)
				}
			}
			confArgs = filteredArgs
			logger.Info("GD extension will be built as bundled extension after PHP compilation")
		} else {
			// Remove GD-related options for PHP 8.0 when user doesn't want it
			var filteredArgs []string
			gdOptions := map[string]bool{
				"--enable-gd":          true,
				"--with-jpeg=/usr":     true,
				"--with-freetype=/usr": true,
			}
			for _, arg := range confArgs {
				if !gdOptions[arg] {
					filteredArgs = append(filteredArgs, arg)
				}
			}
			confArgs = filteredArgs
			logger.Info("GD extension disabled as requested")
		}
	case "7.4":
		if opts.EnableGD {
			// User wants GD - we'll build it as bundled extension later
			var filteredArgs []string
			gdOptions := map[string]bool{
				"--enable-gd":          true,
				"--with-jpeg=/usr":     true,
				"--with-freetype=/usr": true,
			}
			for _, arg := range confArgs {
				if !gdOptions[arg] {
					filteredArgs = append(filteredArgs, arg)
				}
			}
			confArgs = filteredArgs
			logger.Info("GD extension will be built as bundled extension after PHP compilation")
		} else {
			// Remove GD-related options for PHP 7.4 when user doesn't want it
			var filteredArgs []string
			gdOptions := map[string]bool{
				"--enable-gd":          true,
				"--with-jpeg=/usr":     true,
				"--with-freetype=/usr": true,
			}
			for _, arg := range confArgs {
				if !gdOptions[arg] {
					filteredArgs = append(filteredArgs, arg)
				}
			}
			confArgs = filteredArgs
			logger.Info("GD extension disabled as requested")
		}
	default:
		confArgs = append(confArgs, "--enable-mbstring")
	}

	if err := runCommandForPHP(sourceDir, nil, "./buildconf", "--force"); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("buildconf failed: %w", err)
	}

	legacy := version == "7.4" || version == "8.0"
	var (
		buildEnv       []string
		pkgConfigValue string
		ldLibraryValue string
	)

	if legacy {
		buildEnv = append(buildEnv,
			"CFLAGS=-Wno-deprecated-declarations -Wno-discarded-qualifiers",
			"CPPFLAGS=-DOPENSSL_API_COMPAT=0x10100000L",
		)

		vendorPrefix := filepath.Join(prefix, "vendors", "openssl-1.1.1w")
		if err := ensureOpenSSL111(prefix, vendorPrefix, opts.Client, nil, logger); err != nil {
			return logger.Fail("vendor OpenSSL 1.1.1w", err.Error())
		}
		logger.Info(fmt.Sprintf("Using vendored OpenSSL 1.1.1w at %s", vendorPrefix))

		confArgs = rewriteWithOpenSSL(confArgs, vendorPrefix)
		pkgConfigValue = prependEnvPath(filepath.Join(vendorPrefix, "lib", "pkgconfig"), os.Getenv("PKG_CONFIG_PATH"))
		ldLibraryValue = prependEnvPath(filepath.Join(vendorPrefix, "lib"), os.Getenv("LD_LIBRARY_PATH"))

		buildEnv = append(buildEnv,
			"PKG_CONFIG_PATH="+pkgConfigValue,
			"LD_LIBRARY_PATH="+ldLibraryValue,
		)
	}

	logger.Info(fmt.Sprintf("Source directory: %s", sourceDir))
	if legacy {
		logger.Info(fmt.Sprintf("PKG_CONFIG_PATH=%s", pkgConfigValue))
		logger.Info(fmt.Sprintf("LD_LIBRARY_PATH=%s", ldLibraryValue))
	}
	logger.Info(fmt.Sprintf("Configure args: ./configure %s", strings.Join(confArgs, " ")))

	if err := runCommandForPHP(sourceDir, buildEnv, "./configure", confArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("configure php: %w", err)
	}

	makeArgs := []string{"-j"}
	if n := runtime.NumCPU(); n > 0 {
		makeArgs = append(makeArgs, fmt.Sprintf("%d", n))
	}
	if err := runCommandForPHP(sourceDir, buildEnv, "make", makeArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("make php: %w", err)
	}

	if err := runCommandForPHP(sourceDir, buildEnv, "make", "install"); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("make install php: %w", err)
	}

	return nil
}

func installImagickExtension(opts InstallOptions, version, installDir string, logger *lib.Logger) error {
	if logger == nil {
		logger = lib.NewCommandLogger("imagick")
	}
	logger.Info(fmt.Sprintf("Ensuring imagick extension for PHP %s", version))

	phpizePath := filepath.Join(installDir, "bin", "phpize")
	phpConfigPath := filepath.Join(installDir, "bin", "php-config")
	if !fileExists(phpizePath) || !fileExists(phpConfigPath) {
		return logger.Fail(
			"phpize/php-config not found",
			fmt.Sprintf("Expected PHP binaries under %s/bin. Reinstall PHP %s with 'chauf install php %s --force'.", installDir, version, version),
		)
	}

	extDir, err := phpExtensionDir(phpConfigPath)
	if err != nil {
		return logger.Fail("detect PHP extension directory", err.Error())
	}

	modulePath := filepath.Join(extDir, "imagick.so")
	iniPath := filepath.Join(installDir, "etc", "conf.d", "imagick.ini")
	if fileExists(modulePath) && fileExists(iniPath) {
		logger.Success("imagick already installed", modulePath)
		return nil
	}

	imagickVersion := determineImagickVersion(opts.Client, logger)
	tarballName := fmt.Sprintf("imagick-%s.tgz", imagickVersion)
	sourceFolder := fmt.Sprintf("imagick-%s", imagickVersion)
	downloadURL := fmt.Sprintf("https://pecl.php.net/get/%s", tarballName)
	logger.Info(fmt.Sprintf("Imagick version: %s", imagickVersion))

	tmpDir, err := os.MkdirTemp("", "chauffeur-imagick-*")
	if err != nil {
		return fmt.Errorf("create imagick temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, tarballName)
	usedLocal := false
	if localTarball := strings.TrimSpace(os.Getenv("CHAUFFEUR_IMAGICK_TARBALL")); localTarball != "" {
		logger.Info(fmt.Sprintf("Using local imagick tarball %s", localTarball))
		if err := copyFile(localTarball, tarballPath); err != nil {
			logger.Warn("Failed to copy local imagick tarball", err.Error())
		} else {
			usedLocal = true
		}
	}

	if !usedLocal {
		logger.Info(fmt.Sprintf("Downloading imagick %s", imagickVersion))
		if _, err := lib.DownloadToFileWithLogger(opts.Client, downloadURL, tarballPath, fmt.Sprintf("imagick %s", imagickVersion), logger); err != nil {
			return logger.Fail("download imagick extension", err.Error())
		}
	}

	logger.Info("Extracting imagick sources")
	if err := untarPHP(tarballPath, tmpDir); err != nil {
		return logger.Fail("extract imagick extension", err.Error())
	}

	sourceDir := filepath.Join(tmpDir, sourceFolder)
	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("imagick source directory missing: %w", err)
	}

	if err := patchImagickStubConditionals(sourceDir, logger); err != nil {
		logger.Warn("Failed to normalize Imagick stub conditionals", err.Error())
	}

	logger.Info("phpize")
	if err := runCommandForPHP(sourceDir, nil, phpizePath); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("phpize imagick: %w", err)
	}

	logger.Info("configure")
	confArgs := []string{fmt.Sprintf("--with-php-config=%s", phpConfigPath)}
	if err := runCommandForPHP(sourceDir, nil, "./configure", confArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("configure imagick: %w", err)
	}

	logger.Info("make")
	makeArgs := []string{"-j"}
	if n := runtime.NumCPU(); n > 0 {
		makeArgs = append(makeArgs, fmt.Sprintf("%d", n))
	}
	if err := runCommandForPHP(sourceDir, nil, "make", makeArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("make imagick: %w", err)
	}

	logger.Info("make install")
	if err := runCommandForPHP(sourceDir, nil, "make", "install"); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("make install imagick: %w", err)
	}

	if !fileExists(modulePath) {
		return fmt.Errorf("imagick module missing at %s after installation", modulePath)
	}

	if err := os.MkdirAll(filepath.Dir(iniPath), 0o755); err != nil {
		return fmt.Errorf("ensure imagick ini directory: %w", err)
	}
	if err := os.WriteFile(iniPath, []byte(imagickIniContent), 0o644); err != nil {
		return fmt.Errorf("write imagick ini: %w", err)
	}

	logger.Success("imagick extension enabled", iniPath)
	return nil
}

func determineImagickVersion(client *http.Client, logger *lib.Logger) string {
	if logger == nil {
		logger = lib.NewCommandLogger("imagick")
	}

	if override := strings.TrimSpace(os.Getenv("CHAUFFEUR_IMAGICK_VERSION")); override != "" {
		logger.Info(fmt.Sprintf("Using imagick version override %s", override))
		return override
	}

	version, err := fetchLatestImagickVersion(client)
	if err != nil {
		logger.Warn("Failed to detect latest imagick version", err.Error())
		return defaultImagickVersion
	}
	return version
}

func fetchLatestImagickVersion(client *http.Client) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequest("GET", "https://pecl.php.net/rest/r/imagick/latest.txt", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "chauffeur-cli")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(body))
	if version == "" {
		return "", fmt.Errorf("imagick version response empty")
	}
	return version, nil
}

func patchImagickStubConditionals(sourceDir string, logger *lib.Logger) error {
	stubPath := filepath.Join(sourceDir, "Imagick.stub.php")
	data, err := os.ReadFile(stubPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var depth int
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "#if "):
			depth++
		case strings.HasPrefix(trim, "#ifdef"):
			depth++
		case strings.HasPrefix(trim, "#ifndef"):
			depth++
		case strings.HasPrefix(trim, "#endif"):
			if depth > 0 {
				depth--
			}
		}
	}

	if depth <= 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(string(data))
	builder.WriteString("\n")
	for i := 0; i < depth; i++ {
		builder.WriteString("#endif /* auto-closed by Chauffeur */\n")
	}

	if err := os.WriteFile(stubPath, []byte(builder.String()), 0o644); err != nil {
		return err
	}
	if logger != nil {
		logger.Info(fmt.Sprintf("Patched Imagick stub with %d auto-closed #endif directives", depth))
	}
	return nil
}

func ensurePHPBuildDependencies(logger *lib.Logger) error {
	return ensurePHPBuildDependenciesForVersion(logger, "")
}

// ensurePHPBuildDependenciesForVersion validates dependencies with legacy version constraints
func ensurePHPBuildDependenciesForVersion(logger *lib.Logger, phpVersion string) error {
	if logger == nil {
		logger = lib.NewCommandLogger("deps")
	}

	logger.Info("Verifying pkg-config availability")
	if err := ensurePkgConfigAvailable(logger); err != nil {
		return err
	}
	logger.Success("pkg-config found", "")

	// Add legacy version context to logging if applicable
	if phpVersion != "" && isLegacyPHPVersion(phpVersion) {
		logger.Info(fmt.Sprintf("Validating dependencies for legacy PHP %s", phpVersion))
	}

	for _, req := range phpPkgRequirements {
		// Skip sodium check for PHP 8.1+ (native sodium support)
		if req.Name == "libsodium" && phpVersion != "" && !isLegacyPHPVersion(phpVersion) {
			continue
		}

		logger.Info(fmt.Sprintf("Checking %s development headers", req.Name))
		version, err := ensurePkgRequirementWithLegacy(logger, req, phpVersion)
		if err != nil {
			return err
		}
		if version == "" {
			version = "detected"
		}
		logger.Success(fmt.Sprintf("%s headers detected", req.Name), version)
	}

	logger.Success("Build prerequisites satisfied", "")
	return nil
}

func ensurePkgConfigAvailable(logger *lib.Logger) error {
	if _, err := exec.LookPath("pkg-config"); err != nil {
		return logger.Fail(
			"pkg-config not found in PATH",
			"Install pkg-config/pkgconf (e.g., 'sudo apt-get install pkg-config' or 'sudo pacman -S pkgconf') before running 'chauf install php'.",
		)
	}
	return nil
}

func ensurePkgRequirement(logger *lib.Logger, req pkgRequirement) (string, error) {
	const remediation = "Ensure required development headers are installed. See README \"System Dependencies for PHP Builds\"."

	cmd := exec.Command("pkg-config", "--modversion", req.Package)
	output, err := cmd.Output()
	if err != nil {
		var detail string
		if exitErr, ok := err.(*exec.ExitError); ok {
			detail = strings.TrimSpace(string(exitErr.Stderr))
		}
		if detail == "" {
			detail = fmt.Sprintf("pkg-config could not locate %s on this system.", req.Name)
		}
		context := fmt.Sprintf("%s\n%s", detail, remediation)
		return "", logger.Fail(fmt.Sprintf("missing %s development headers", req.Name), context)
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", logger.Fail(fmt.Sprintf("unable to detect %s version", req.Name), remediation)
	}

	if req.MinVersion != "" && compareSemver(version, req.MinVersion) < 0 {
		context := fmt.Sprintf("Detected %s; require >= %s.\n%s", version, req.MinVersion, remediation)
		return "", logger.Fail(fmt.Sprintf("%s version too old for PHP", req.Name), context)
	}

	for _, blockedVersion := range req.BlockedVersions {
		if compareSemver(version, blockedVersion) == 0 {
			context := fmt.Sprintf("Detected %s which PHP explicitly blocks.\n%s", version, remediation)
			return "", logger.Fail(fmt.Sprintf("unsupported %s release detected", req.Name), context)
		}
	}

	return version, nil
}

// ensurePkgRequirementWithLegacy validates package requirements with legacy-specific constraints
func ensurePkgRequirementWithLegacy(logger *lib.Logger, req pkgRequirement, phpVersion string) (string, error) {
	// First perform standard validation
	version, err := ensurePkgRequirement(logger, req)
	if err != nil {
		return "", err
	}

	// If no PHP version specified or not a legacy version, return standard result
	if phpVersion == "" || !isLegacyPHPVersion(phpVersion) {
		return version, nil
	}

	// Apply legacy-specific constraints
	legacyReq, hasLegacyReq := getLegacyDependencyRequirement(phpVersion, req.Package)
	if !hasLegacyReq {
		return version, nil
	}

	const remediation = "Ensure required development headers are installed. See README \"System Dependencies for PHP Builds\"."

	// Check maximum version constraint for legacy compatibility
	if legacyReq.MaxVersion != "" && compareSemver(version, legacyReq.MaxVersion) > 0 {
		context := fmt.Sprintf("Detected %s which may be incompatible with legacy PHP %s (max: %s).\n%s",
			version, phpVersion, legacyReq.MaxVersion, remediation)
		return "", logger.Fail(fmt.Sprintf("%s version too new for legacy PHP %s", req.Name, phpVersion), context)
	}

	// Check minimum version constraint for legacy requirements
	if legacyReq.MinVersion != "" && compareSemver(version, legacyReq.MinVersion) < 0 {
		context := fmt.Sprintf("Detected %s which is below minimum required for PHP %s (min: %s).\n%s",
			version, phpVersion, legacyReq.MinVersion, remediation)
		return "", logger.Fail(fmt.Sprintf("%s version too old for legacy PHP %s", req.Name, phpVersion), context)
	}

	// Check legacy-specific blocked versions
	for _, blockedVersion := range legacyReq.BlockedVersions {
		if compareSemver(version, blockedVersion) == 0 {
			context := fmt.Sprintf("Detected %s which PHP %s explicitly blocks.\n%s", version, phpVersion, remediation)
			return "", logger.Fail(fmt.Sprintf("unsupported %s release for PHP %s", req.Name, phpVersion), context)
		}
	}

	return version, nil
}

func compareSemver(a, b string) int {
	parse := func(v string) []int {
		parts := strings.Split(v, ".")
		result := make([]int, len(parts))
		for i, part := range parts {
			val, err := strconv.Atoi(part)
			if err != nil {
				val = 0
			}
			result[i] = val
		}
		return result
	}

	aParts := parse(a)
	bParts := parse(b)
	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for len(aParts) < maxLen {
		aParts = append(aParts, 0)
	}
	for len(bParts) < maxLen {
		bParts = append(bParts, 0)
	}

	for i := 0; i < maxLen; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

func ensureOpenSSL111(workspacePrefix, vendorPrefix string, client *http.Client, logf func(string, ...interface{}), logger *lib.Logger) error {
	if logger == nil {
		logger = lib.NewCommandLogger("install")
	}

	// Safe log function that handles nil logf
	safeLogf := func(format string, args ...interface{}) {
		if logf != nil {
			logf(format, args...)
		}
	}

	expected := strings.ToLower(openssl111wExpectedSHA256)
	if len(expected) != 64 {
		return fmt.Errorf("invalid expected SHA256 length: %d", len(expected))
	}

	cryptoLib := filepath.Join(vendorPrefix, "lib", "libcrypto.so.1.1")
	sslLib := filepath.Join(vendorPrefix, "lib", "libssl.so.1.1")
	if fileExists(cryptoLib) && fileExists(sslLib) {
		safeLogf("Using vendored OpenSSL %s at %s", openssl111wVersion, vendorPrefix)
		return nil
	}

	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	cacheDir := filepath.Join(workspacePrefix, "cache", "openssl")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("ensure openssl cache dir: %w", err)
	}
	cachePath := filepath.Join(cacheDir, openssl111wTarball)

	var (
		tarballPath string
		computed    string
		usedURL     string
	)

	if fileExists(cachePath) {
		sum, err := fileSHA256(cachePath)
		if err == nil && strings.EqualFold(sum, expected) {
			logger.Info(fmt.Sprintf("Reusing cached %s (computed=%s expected=%s)", cachePath, strings.ToLower(sum), expected))
			tarballPath = cachePath
			computed = strings.ToLower(sum)
			usedURL = "cache"
		} else {
			if err != nil {
				logger.Warn("failed to hash cached OpenSSL tarball", err.Error())
			} else {
				logger.Warn("Cached OpenSSL tarball checksum mismatch", fmt.Sprintf("(computed=%s expected=%s); re-downloading", strings.ToLower(sum), expected))
			}
			_ = os.Remove(cachePath)
		}
	}

	logger.Info(fmt.Sprintf("Vendoring OpenSSL %s into %s", openssl111wVersion, vendorPrefix))
	if err := os.MkdirAll(vendorPrefix, 0o755); err != nil {
		return fmt.Errorf("ensure openssl prefix: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "chauffeur-openssl-*")
	if err != nil {
		return fmt.Errorf("create temp dir for openssl: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var lastErr error
	if tarballPath == "" {
		urls := []string{openssl111wPrimaryURL, openssl111wFallbackURL}
		for _, url := range urls {
			downloadDest := filepath.Join(tmpDir, openssl111wTarball)
			safeLogf("Downloading %s", url)
			if _, err := lib.DownloadToFileWithLogger(client, url, downloadDest, "OpenSSL 1.1.1w", logger); err != nil {
				logger.Warn("Download from "+url+" failed", err.Error())
				lastErr = err
				continue
			}

			sum, err := fileSHA256(downloadDest)
			if err != nil {
				lastErr = fmt.Errorf("hash openssl tarball: %w", err)
				logger.Warn("Failed to compute checksum for "+url, err.Error())
				continue
			}
			lower := strings.ToLower(sum)
			safeLogf("Downloaded %s (computed=%s expected=%s)", url, lower, expected)
			if !strings.EqualFold(sum, expected) {
				logger.Warn("OpenSSL "+openssl111wVersion+": checksum mismatch", fmt.Sprintf("expected %s, got %s from %s", expected, lower, url))
				logger.Warn("If mirrors moved, try the fallback URL "+openssl111wFallbackURL+" (already auto-handled)", "")
				_ = os.Remove(downloadDest)
				lastErr = fmt.Errorf("OpenSSL %s: checksum mismatch (expected %s, got %s) from %s", openssl111wVersion, expected, lower, url)
				continue
			}

			if err := copyFile(downloadDest, cachePath); err != nil {
				lastErr = fmt.Errorf("cache openssl tarball: %w", err)
				logger.Warn("Failed to cache OpenSSL tarball", err.Error())
				_ = os.Remove(downloadDest)
				continue
			}
			tarballPath = cachePath
			computed = lower
			usedURL = url
			safeLogf("Cached OpenSSL tarball at %s", cachePath)
			_ = os.Remove(downloadDest)
			break
		}
		if tarballPath == "" {
			if lastErr != nil {
				return lastErr
			}
			return fmt.Errorf("failed to download OpenSSL %s tarball", openssl111wVersion)
		}
	}

	if usedURL == "" {
		usedURL = tarballPath
	}
	if computed == "" {
		sum, err := fileSHA256(tarballPath)
		if err == nil {
			computed = strings.ToLower(sum)
		}
	}
	safeLogf(fmt.Sprintf("Using OpenSSL archive from %s (computed=%s expected=%s)", usedURL, computed, expected))

	sourceRoot := filepath.Join(tmpDir, "src")
	if err := untarPHP(tarballPath, sourceRoot); err != nil {
		return fmt.Errorf("extract openssl %s: %w", openssl111wVersion, err)
	}
	opensslSource := filepath.Join(sourceRoot, "openssl-"+openssl111wVersion)

	configArgs := []string{
		fmt.Sprintf("--prefix=%s", vendorPrefix),
		fmt.Sprintf("--openssldir=%s", filepath.Join(vendorPrefix, "ssl")),
		"shared",
	}
	logger.Info(fmt.Sprintf("Running: ./config %s", strings.Join(configArgs, " ")))
	if err := runCommandForPHP(opensslSource, nil, "./config", configArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("openssl config: %w", err)
	}

	makeArgs := []string{"-j"}
	if n := runtime.NumCPU(); n > 0 {
		makeArgs = append(makeArgs, fmt.Sprintf("%d", n))
	}
	logger.Info(fmt.Sprintf("Running: make %s", strings.Join(makeArgs, " ")))
	if err := runCommandForPHP(opensslSource, nil, "make", makeArgs...); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("openssl make: %w", err)
	}

	logger.Info("Running: make install_sw")
	if err := runCommandForPHP(opensslSource, nil, "make", "install_sw"); err != nil {
		logCommandFailure(err, logger)
		return fmt.Errorf("openssl make install_sw: %w", err)
	}

	if !fileExists(cryptoLib) || !fileExists(sslLib) {
		return fmt.Errorf("vendored OpenSSL seems incomplete (missing libcrypto.so.1.1/libssl.so.1.1)")
	}

	logger.Success(fmt.Sprintf("OpenSSL %s installed at %s", openssl111wVersion, vendorPrefix), "")
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func rewriteWithOpenSSL(args []string, prefix string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "--with-openssl") {
			continue
		}
		filtered = append(filtered, arg)
	}
	return append(filtered, fmt.Sprintf("--with-openssl=%s", prefix))
}

func prependEnvPath(newValue, existing string) string {
	if newValue == "" {
		return existing
	}
	if existing == "" {
		return newValue
	}
	return newValue + ":" + existing
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

/**
 * ensurePHPlayout creates the directory structure PHP expects inside the workspace.
 *
 * @param prefix Chauffeur workspace root.
 * @param version PHP version string.
 * @return error if any directory cannot be created.
 */
func ensurePHPlayout(prefix, version string) error {
	phpDir := filepath.Join(prefix, "php", version)
	paths := []string{
		filepath.Join(phpDir, "etc"),
		filepath.Join(phpDir, "etc", "conf.d"),
		filepath.Join(phpDir, "var", "log"),
		filepath.Join(phpDir, "var", "run"),
		filepath.Join(phpDir, "lib", "php", "extensions"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("ensure php path %s: %w", path, err)
		}
	}

	return nil
}

// UpdateDefaultPHPShim repoints the global php shim to the specified version.
func UpdateDefaultPHPShim(prefix, version string) error {
	binary := filepath.Join(prefix, "php", version, "bin", phpBinaryName)
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("php binary for %s: %w", version, err)
	}
	if err := writeShim(prefix, phpBinaryName, binary); err != nil {
		return fmt.Errorf("write default php shim: %w", err)
	}
	return nil
}

/**
 * writeDefaultPHPFPMConf seeds the workspace with a minimal php-fpm.conf and pool configuration.
 *
 * @param prefix Chauffeur workspace root.
 * @param version PHP version string.
 * @return error when configuration files cannot be written.
 */
func writeDefaultPHPFPMConf(prefix, version string) error {
	phpDir := filepath.Join(prefix, "php", version)

	// Main php-fpm.conf
	fpmConf := fmt.Sprintf(`[global]
pid = %s
error_log = %s
log_level = notice
daemonize = no
include=%s/etc/php-fpm.d/*.conf
`,
		filepath.Join(phpDir, "var", "run", "php-fpm.pid"),
		filepath.Join(phpDir, "var", "log", "php-fpm.log"),
		phpDir,
	)

	fpmConfPath := filepath.Join(phpDir, "etc", "php-fpm.conf")
	if err := os.WriteFile(fpmConfPath, []byte(fpmConf), 0o644); err != nil {
		return fmt.Errorf("write php-fpm.conf: %w", err)
	}

	// Pool configuration
	poolDir := filepath.Join(phpDir, "etc", "php-fpm.d")
	if err := os.MkdirAll(poolDir, 0o755); err != nil {
		return fmt.Errorf("create php-fpm pool dir: %w", err)
	}

	user := getPHPUser()
	poolConfPath := filepath.Join(poolDir, "default.conf")
	socketPath := filepath.Join(phpDir, "var", "run", "php-fpm.sock")

	poolConf := fmt.Sprintf(`[default]
user = %s
group = %s
listen = %s
listen.owner = %s
listen.group = %s
listen.mode = 0666
pm = ondemand
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
pm.process_idle_timeout = 10s
php_admin_value[error_log] = %s
php_admin_flag[log_errors] = on
`, user, user, socketPath, user, user, filepath.Join(phpDir, "var", "log", "php-fpm-error.log"))

	if err := os.WriteFile(poolConfPath, []byte(poolConf), 0o644); err != nil {
		return fmt.Errorf("write php-fpm pool conf: %w", err)
	}

	return nil
}

/**
 * writeUploadLimitsConf creates the default PHP upload limits configuration.
 *
 * @param prefix Installation prefix directory
 * @param version PHP version
 * @return error if configuration creation fails
 */
func writeUploadLimitsConf(prefix, version string) error {
	phpDir := filepath.Join(prefix, "php", version)

	// Create conf.d directory if it doesn't exist
	confDDir := filepath.Join(phpDir, "etc", "conf.d")
	if err := os.MkdirAll(confDDir, 0o755); err != nil {
		return fmt.Errorf("failed to create extension config directory: %w", err)
	}

	// Create upload-limits.ini with generous default limits
	uploadLimitsContent := `; Chauffeur PHP Upload Limits Configuration
; Default values optimized for modern web applications
; Compatible with nginx client_max_body_size 256M

; Maximum size of an uploaded file
upload_max_filesize = 256M

; Maximum size of POST data that PHP will accept
post_max_size = 256M

; Maximum time in seconds a script is allowed to run before it is terminated by the parser
max_execution_time = 300

; Maximum amount of memory a script may consume
memory_limit = 512M

; Maximum time in seconds a script is allowed to parse input data
max_input_time = 300

; Maximum number of input variables allowed per request
max_input_vars = 3000

; Enable file uploads
file_uploads = On

; Temporary directory for HTTP uploaded files
upload_tmp_dir = /tmp
`

	uploadLimitsPath := filepath.Join(confDDir, "upload-limits.ini")
	if err := os.WriteFile(uploadLimitsPath, []byte(uploadLimitsContent), 0o644); err != nil {
		return fmt.Errorf("failed to create upload-limits.ini: %w", err)
	}

	return nil
}

/**
 * getPHPUser returns the current user for PHP-FPM configuration.
 *
 * @return current username
 */
func getPHPUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("LOGNAME"); user != "" {
		return user
	}
	return "www-data" // fallback
}

func startPHPLogSection(logger *lib.Logger, title string) {
	logger.Info(strings.ToUpper(title))
}

func logPHPInfo(logger *lib.Logger, format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

func logPHPSuccess(logger *lib.Logger, format string, args ...interface{}) {
	logger.Success(fmt.Sprintf(format, args...), "")
}

func logPHPWarn(logger *lib.Logger, format string, args ...interface{}) {
	logger.Warn(fmt.Sprintf(format, args...), "")
}

func logPHPError(logger *lib.Logger, format string, args ...interface{}) {
	logger.Info(fmt.Sprintf("ERROR: %s", fmt.Sprintf(format, args...)))
}

func logCommandFailure(err error, logger *lib.Logger) {
	if logger == nil {
		logger = lib.NewCommandLogger("install")
	}
	var detail detailedError
	if errors.As(err, &detail) {
		logPHPError(logger, detail.Detail())
	}
}

/**
 * untarPHP extracts the PHP tarball contents into dest while preserving permissions.
 *
 * @param tarball Path to the downloaded PHP source tarball.
 * @param dest    Destination directory for extraction.
 * @return error when extraction fails.
 */
func untarPHP(tarball, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	file, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

/**
 * runCommandForPHP executes a command inside dir and returns stderr/stdout on failure.
 *
 * @param dir  Working directory for execution.
 * @param env  Additional environment variables for the command.
 * @param name Command binary to run.
 * @param args Additional command arguments.
 * @return error when the command exits non-zero.
 */
func runCommandForPHP(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return commandError{
			Name:   name,
			Args:   args,
			Err:    err,
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
	}
	return nil
}

func phpExtensionDir(phpConfigPath string) (string, error) {
	cmd := exec.Command(phpConfigPath, "--extension-dir")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("php-config --extension-dir: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	dir := strings.TrimSpace(stdout.String())
	if dir == "" {
		return "", errors.New("php-config returned empty extension dir")
	}
	return dir, nil
}

/**
 * getLocalTarballFromConfig retrieves a local tarball path from the config.
 * This is a wrapper around the config package to avoid circular imports.
 *
 * @param version PHP version
 * @return Path to local tarball or empty string if not found
 */
func getLocalTarballFromConfig(version string) (string, error) {
	// Create a simple YAML parser for our specific use case
	// This avoids circular imports with the config package
	workspaceDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(workspaceDir, ".chauffeur", "config", "chauffeur.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", nil // No config file is fine
	}

	// Simple YAML parser for our specific structure
	lines := strings.Split(string(data), "\n")
	inPHPSection := false
	inLocalTarballs := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for php section
		if trimmed == "php:" {
			inPHPSection = true
			continue
		}

		// Check for other sections (exit php section)
		if inPHPSection && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, " ") {
			inPHPSection = false
			inLocalTarballs = false
			continue
		}

		// Check for local_tarballs subsection
		if inPHPSection && trimmed == "local_tarballs:" {
			inLocalTarballs = true
			continue
		}

		// Look for version entry in local_tarballs
		if inLocalTarballs && strings.HasPrefix(trimmed, "    ") && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				entryVersion := strings.TrimSpace(parts[0])
				entryPath := strings.TrimSpace(parts[1])

				// Remove quotes if present
				entryPath = strings.Trim(entryPath, "\"'")

				if entryVersion == version && entryPath != "" {
					return entryPath, nil
				}
			}
		}
	}

	return "", nil
}

/**
 * removeLocalTarballFromConfig removes a local tarball path from the config.
 *
 * @param version PHP version to remove
 * @return error if config operation fails
 */
func removeLocalTarballFromConfig(version string) error {
	workspaceDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(workspaceDir, ".chauffeur", "config", "chauffeur.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil // No config file is fine
	}

	// Parse and modify YAML
	lines := strings.Split(string(data), "\n")
	var result []string
	inPHPSection := false
	inLocalTarballs := false
	versionRemoved := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for php section
		if trimmed == "php:" {
			inPHPSection = true
			result = append(result, line)
			continue
		}

		// Check for other sections (exit php section)
		if inPHPSection && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, " ") {
			inPHPSection = false
			inLocalTarballs = false
			result = append(result, line)
			continue
		}

		// Check for local_tarballs subsection
		if inPHPSection && trimmed == "local_tarballs:" {
			inLocalTarballs = true
			result = append(result, line)
			continue
		}

		// Look for version entry to remove
		if inLocalTarballs && strings.HasPrefix(trimmed, "    ") && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				entryVersion := strings.TrimSpace(parts[0])
				if entryVersion == version {
					// Skip this line (remove the entry)
					versionRemoved = true
					continue
				}
			}
		}

		result = append(result, line)
	}

	// Clean up empty local_tarballs section if no entries remain
	if versionRemoved {
		result = cleanupEmptyLocalTarballsSection(result)
	}

	// Write back modified config
	return os.WriteFile(configPath, []byte(strings.Join(result, "\n")), 0644)
}

/**
 * cleanupEmptyLocalTarballsSection removes the local_tarballs section if it's empty.
 */
func cleanupEmptyLocalTarballsSection(lines []string) []string {
	var result []string
	inLocalTarballs := false
	hasEntries := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "local_tarballs:" {
			inLocalTarballs = true
			result = append(result, line)
			continue
		}

		if inLocalTarballs {
			if strings.HasPrefix(trimmed, "    ") && strings.Contains(trimmed, ":") {
				hasEntries = true
			} else if !strings.HasPrefix(trimmed, " ") {
				// End of local_tarballs section
				inLocalTarballs = false
				if !hasEntries {
					// Remove the local_tarballs: line from result
					result = result[:len(result)-1]
				}
				result = append(result, line)
				continue
			}
		}

		result = append(result, line)
	}

	return result
}

/**
 * cacheDownloadedFile copies a downloaded file to the cache directory
 * and optionally updates the config to point to it.
 *
 * @param tarballPath Path to the temporary downloaded file
 * @param service Service name (e.g., "php", "nginx", "composer")
 * @param version Service version (optional, for service-specific caching)
 * @param tarballName Name of the downloaded file
 * @param logger Logger instance
 * @return error if caching fails
 */
func cacheDownloadedFile(tarballPath, service, version, tarballName string, logger *lib.Logger) error {
	// Create cache directory
	workspaceDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	cacheDir := filepath.Join(workspaceDir, ".chauffeur", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	// Destination path in cache
	cachePath := filepath.Join(cacheDir, tarballName)

	// Copy file to cache
	if err := copyFile(tarballPath, cachePath); err != nil {
		return fmt.Errorf("copy to cache: %w", err)
	}

	// Update config to point to cached file for PHP services
	if service == "php" && version != "" {
		if err := addLocalTarballToConfig(version, cachePath); err != nil {
			logger.Warn("Failed to update config with cached path", err.Error())
			// Don't fail the operation - the file is cached, just not in config
		} else {
			logger.Info(fmt.Sprintf("Updated config with cached path: %s", cachePath))
		}
	}

	return nil
}

/**
 * cacheDownloadedTarball copies the downloaded tarball to the cache directory
 * and updates the config to point to it. (Legacy function for PHP compatibility)
 *
 * @param tarballPath Path to the temporary downloaded tarball
 * @param version PHP version
 * @param tarballName Name of the tarball file
 * @param logger Logger instance
 * @return error if caching fails
 */
func cacheDownloadedTarball(tarballPath, version, tarballName string, logger *lib.Logger) error {
	return cacheDownloadedFile(tarballPath, "php", version, tarballName, logger)
}

/**
 * checkForCachedFile checks if a downloaded file exists in the cache directory.
 *
 * @param service Service name
 * @param fileName Name of the file to check
 * @return Path to cached file if found, empty string otherwise
 */
func checkForCachedFile(service, fileName string) string {
	workspaceDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	cachePath := filepath.Join(workspaceDir, ".chauffeur", "cache", fileName)
	if info, err := os.Stat(cachePath); err == nil && info.Mode().IsRegular() {
		return cachePath
	}

	return ""
}

/**
 * CheckForServiceCache checks for any cached files related to a specific service.
 *
 * @param service Service name (e.g., "php", "composer", "nginx")
 * @param version Optional version (for PHP version-specific cache)
 * @return List of cached file paths found, empty list if none
 */
func CheckForServiceCache(service, version string) []string {
	var cachedFiles []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cachedFiles
	}

	cacheDir := filepath.Join(homeDir, ".chauffeur", "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return cachedFiles
	}

	// Read cache directory
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return cachedFiles
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		filePath := filepath.Join(cacheDir, fileName)

		// Check if file belongs to the service
		if isServiceCacheFile(service, version, fileName) {
			cachedFiles = append(cachedFiles, filePath)
		}
	}

	return cachedFiles
}

/**
 * isServiceCacheFile determines if a cached file belongs to a specific service.
 *
 * @param service Service name
 * @param version Optional version for version-specific matching
 * @param fileName Cached file name
 * @return true if file belongs to the service
 */
func isServiceCacheFile(service, version, fileName string) bool {
	fileName = strings.ToLower(fileName)

	switch service {
	case "php":
		if version != "" {
			// Specific PHP version: php-8.3.27.tar.gz, php-8.4.14.tar.gz, etc.
			return strings.Contains(fileName, "php-") && strings.Contains(fileName, version)
		}
		// Any PHP version
		return strings.HasPrefix(fileName, "php-") && strings.HasSuffix(fileName, ".tar.gz")

	case "composer":
		// Composer cache files: composer.phar, composer-2.8.4.phar, etc.
		return strings.HasPrefix(fileName, "composer") && (strings.HasSuffix(fileName, ".phar") || strings.Contains(fileName, "sha256"))

	case "nginx":
		// Nginx cache files: nginx-1.29.3.tar.gz, etc.
		return strings.HasPrefix(fileName, "nginx-") && strings.HasSuffix(fileName, ".tar.gz")

	default:
		return false
	}
}

/**
 * GetCacheFileInfo provides human-readable information about cached files.
 *
 * @param cachedFiles List of cached file paths
 * @return Formatted string with file information
 */
func GetCacheFileInfo(cachedFiles []string) string {
	if len(cachedFiles) == 0 {
		return "No cached files found"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d cached file(s):\n", len(cachedFiles)))

	for _, filePath := range cachedFiles {
		if fileInfo, err := os.Stat(filePath); err == nil {
			fileName := filepath.Base(filePath)
			size := lib.HumanBytes(fileInfo.Size())
			result.WriteString(fmt.Sprintf("  - %s (%s)\n", fileName, size))
		}
	}

	return result.String()
}

/**
 * RemoveServiceCache removes cached files for a specific service.
 *
 * @param service Service name
 * @param version Optional version for version-specific removal
 * @return Number of files removed and any error
 */
func RemoveServiceCache(service, version string) (int, error) {
	cachedFiles := CheckForServiceCache(service, version)
	removed := 0

	for _, filePath := range cachedFiles {
		if err := os.Remove(filePath); err == nil {
			removed++
		}
	}

	return removed, nil
}

/**
 * addLocalTarballToConfig adds or updates a local tarball path in the config.
 *
 * @param version PHP version
 * @param path Path to the tarball
 * @return error if config update fails
 */
func addLocalTarballToConfig(version, path string) error {
	workspaceDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(workspaceDir, ".chauffeur", "config", "chauffeur.yaml")

	// Read existing config or create new one
	var data []byte
	if _, err := os.Stat(configPath); err == nil {
		data, err = os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}
	} else {
		// Create default config structure if it doesn't exist
		data = []byte("version: 1\ntelemetry: false\nphp:\n  default: \"8.3\"\n")
	}

	// Parse and modify YAML
	lines := strings.Split(string(data), "\n")
	var result []string
	inPHPSection := false
	inLocalTarballs := false
	localTarballsAdded := false
	versionEntryAdded := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for php section
		if trimmed == "php:" {
			inPHPSection = true
			result = append(result, line)
			continue
		}

		// Check for other sections (exit php section)
		if inPHPSection && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, " ") {
			// Add local_tarballs section before exiting php section if not already added
			if inPHPSection && !localTarballsAdded && !versionEntryAdded {
				result = append(result, "  local_tarballs:")
				result = append(result, fmt.Sprintf("    %s: \"%s\"", version, path))
				versionEntryAdded = true
				localTarballsAdded = true
			}
			inPHPSection = false
			inLocalTarballs = false
			result = append(result, line)
			continue
		}

		// Check for local_tarballs subsection
		if inPHPSection && trimmed == "local_tarballs:" {
			inLocalTarballs = true
			localTarballsAdded = true
			result = append(result, line)
			continue
		}

		// Look for version entry to update
		if inLocalTarballs && strings.HasPrefix(trimmed, "    ") && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				entryVersion := strings.TrimSpace(parts[0])
				if entryVersion == version {
					// Update existing entry
					result = append(result, fmt.Sprintf("    %s: \"%s\"", version, path))
					versionEntryAdded = true
					continue
				}
			}
		}

		result = append(result, line)
	}

	// Add local_tarballs section and version entry if they don't exist
	if !versionEntryAdded {
		if !inPHPSection {
			// This shouldn't happen in a valid config, but handle it gracefully
			result = append(result, "php:")
			result = append(result, "  default: \"8.3\"")
		}

		if !localTarballsAdded {
			// Find where to insert local_tarballs (after php section header)
			insertIndex := -1
			for i, line := range result {
				if strings.TrimSpace(line) == "php:" {
					insertIndex = i + 1
					break
				}
			}

			if insertIndex >= 0 {
				// Insert local_tarballs section
				newResult := make([]string, 0, len(result)+2)
				newResult = append(newResult, result[:insertIndex]...)
				newResult = append(newResult, "  local_tarballs:")
				newResult = append(newResult, fmt.Sprintf("    %s: \"%s\"", version, path))
				newResult = append(newResult, result[insertIndex:]...)
				result = newResult
			}
		} else {
			// Just add the version entry to existing local_tarballs section
			result = append(result, fmt.Sprintf("    %s: \"%s\"", version, path))
		}
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Write back modified config
	return os.WriteFile(configPath, []byte(strings.Join(result, "\n")), 0644)
}

/**
 * patchGDExtension applies function pointer type patches to GD extension source code.
 * This fixes compilation errors with modern GCC/Clang compilers for PHP 7.4.
 *
 * @param gdSourceDir Path to GD extension source directory
 * @param logger Logger instance
 * @return error if patching fails
 */
func patchGDExtension(gdSourceDir string, logger *lib.Logger) error {
	gdFile := filepath.Join(gdSourceDir, "gd.c")
	ctxFile := filepath.Join(gdSourceDir, "gd_ctx.c")

	// Check if files exist
	for _, f := range []string{gdFile, ctxFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("GD source file not found: %s", f)
		}
	}

	// Define patches for gd.c - add casts to generic void(*)() function pointers
	gdPatches := []struct {
		pattern string
		replacement string
	}{
		{
			", gdImageGd);",
			", (void (*)())gdImageGd);",
		},
		{
			", gdImageGd2);",
			", (void (*)())gdImageGd2);",
		},
		{
			", gdImageWbmp);",
			", (void (*)())gdImageWbmp);",
		},
		{
			", gdImageJpeg);",
			", (void (*)())gdImageJpeg);",
		},
		{
			", gdImagePng);",
			", (void (*)())gdImagePng);",
		},
		{
			", gdImageGif);",
			", (void (*)())gdImageGif);",
		},
	}

	// Apply patches to gd.c
	for _, patch := range gdPatches {
		content, err := os.ReadFile(gdFile)
		if err != nil {
			return fmt.Errorf("failed to read gd.c: %w", err)
		}
		strContent := string(content)
		if !strings.Contains(strContent, patch.replacement) {
			strContent = strings.ReplaceAll(strContent, patch.pattern, patch.replacement)
			if err := os.WriteFile(gdFile, []byte(strContent), 0644); err != nil {
				return fmt.Errorf("failed to patch gd.c: %w", err)
			}
		}
	}

	// Define patches for ctx functions (these are called from gd.c, so patch gd.c too)
	ctxPatches := []struct {
		pattern string
		replacement string
	}{
		{
			", gdImageXbmCtx);",
			", (void (*)())gdImageXbmCtx);",
		},
		{
			", gdImageGifCtx);",
			", (void (*)())gdImageGifCtx);",
		},
		{
			", gdImagePngCtxEx);",
			", (void (*)())gdImagePngCtxEx);",
		},
		{
			", gdImageWebpCtx);",
			", (void (*)())gdImageWebpCtx);",
		},
		{
			", gdImageJpegCtx);",
			", (void (*)())gdImageJpegCtx);",
		},
		{
			", gdImageWBMPCtx);",
			", (void (*)())gdImageWBMPCtx);",
		},
		{
			", gdImageBmpCtx);",
			", (void (*)())gdImageBmpCtx);",
		},
		{
			", gdImagePngCtx);",
			", (void (*)())gdImagePngCtx);",
		},
		{
			", gdImageGd2Ctx);",
			", (void (*)())gdImageGd2Ctx);",
		},
	}

	// Apply ctx patches to gd.c (where the function calls are)
	for _, patch := range ctxPatches {
		content, err := os.ReadFile(gdFile)
		if err != nil {
			return fmt.Errorf("failed to read gd.c: %w", err)
		}
		strContent := string(content)
		if !strings.Contains(strContent, patch.replacement) {
			strContent = strings.ReplaceAll(strContent, patch.pattern, patch.replacement)
			if err := os.WriteFile(gdFile, []byte(strContent), 0644); err != nil {
				return fmt.Errorf("failed to patch gd.c with ctx patches: %w", err)
			}
		}
	}

	return nil
}

/**
 * buildBundledGDExtension builds the GD extension as a bundled extension from PHP source.
 * This is used for legacy PHP versions (7.4, 8.0) that have GD compatibility issues.
 *
 * @param sourceDir Path to extracted PHP source directory
 * @param installDir Path to PHP installation directory
 * @param version PHP version
 * @param logger Logger instance
 * @return error if GD extension build fails
 */
func buildBundledGDExtension(sourceDir, installDir, version string, logger *lib.Logger) error {
	// Construct path to GD extension source
	gdSourceDir := filepath.Join(sourceDir, "ext", "gd")

	// Verify GD source directory exists
	if _, err := os.Stat(gdSourceDir); os.IsNotExist(err) {
		return fmt.Errorf("GD extension source not found at %s", gdSourceDir)
	}

	logger.Info("Preparing GD extension build")

	// Use phpize from the newly installed PHP
	phpizePath := filepath.Join(installDir, "bin", "phpize")

	// Run phpize in GD directory
	if err := runCommandForPHP(gdSourceDir, nil, phpizePath); err != nil {
		return fmt.Errorf("phpize failed: %w", err)
	}

	logger.Info("Configuring GD extension")

	// Configure GD extension with proper image format support
	phpConfigPath := filepath.Join(installDir, "bin", "php-config")
	gdConfigureArgs := []string{
		fmt.Sprintf("--with-php-config=%s", phpConfigPath),
		"--with-gd=shared",
		"--with-freetype=/usr",
		"--with-jpeg=/usr",
		"--with-png=/usr",
		"--with-webp=/usr",
	}

	if err := runCommandForPHP(gdSourceDir, nil, "./configure", gdConfigureArgs...); err != nil {
		return fmt.Errorf("GD configure failed: %w", err)
	}

	// Modify Makefile to add compiler flags that suppress strict prototype checking
	// Use C89 standard where void() means "unspecified parameters" instead of "no parameters"
	makefilePath := filepath.Join(gdSourceDir, "Makefile")
	makefileContent, err := os.ReadFile(makefilePath)
	if err == nil {
		// Use sed to add -std=gnu89 to CFLAGS line
		lines := strings.Split(string(makefileContent), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "CFLAGS = ") {
				lines[i] = line + " -std=gnu89"
				break
			}
		}
		modifiedMakefile := strings.Join(lines, "\n")
		os.WriteFile(makefilePath, []byte(modifiedMakefile), 0644)
	}

	logger.Info("Compiling GD extension")

	// Compile GD extension
	if err := runCommandForPHP(gdSourceDir, nil, "make"); err != nil {
		return fmt.Errorf("GD make failed: %w", err)
	}

	logger.Info("Installing GD extension")

	// Install GD extension
	if err := runCommandForPHP(gdSourceDir, nil, "make", "install"); err != nil {
		return fmt.Errorf("GD make install failed: %w", err)
	}

	// Create extension configuration directory
	confDDir := filepath.Join(installDir, "etc", "conf.d")
	if err := os.MkdirAll(confDDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension config directory: %w", err)
	}

	// Create gd.ini extension configuration
	gdIniPath := filepath.Join(confDDir, "gd.ini")
	gdIniContent := "extension=gd.so\n"
	if err := os.WriteFile(gdIniPath, []byte(gdIniContent), 0644); err != nil {
		return fmt.Errorf("failed to create gd.ini: %w", err)
	}

	logger.Info("Created GD extension configuration")

	// Test that GD extension loads properly
	phpBinary := filepath.Join(installDir, "bin", "php")
	testCmd := exec.Command(phpBinary, "-m")
	testCmd.Env = append(os.Environ(), fmt.Sprintf("PHPRC=%s", filepath.Join(installDir, "etc")))
	output, err := testCmd.Output()
	if err != nil {
		logger.Warn("Failed to test GD extension", err.Error())
		return nil // Don't fail installation, just warn
	}

	if strings.Contains(string(output), "gd") {
		logger.Success("GD extension loaded successfully", "")
	} else {
		logger.Warn("GD extension not found in module list", "Extension may not be properly configured")
	}

	return nil
}

/**
 * PromptGDExtension prompts user for GD extension preference for legacy PHP versions.
 * Modern PHP versions (8.1+) always return true as GD is enabled by default.
 *
 * @param version PHP version to install
 * @param logger Logger instance
 * @param force Whether to skip prompts (return default behavior)
 * @return true if GD should be enabled, false otherwise
 */
func PromptGDExtension(version string, logger *lib.Logger, force bool) (bool, error) {
	// Modern PHP versions always get GD
	if version != "7.4" && version != "8.0" {
		return true, nil
	}

	// Skip prompting in force mode - enable GD by default for PHP 7.4/8.0
	if force {
		logger.Info("Force mode: enabling GD extension by default")
		return true, nil
	}

	logger.Warn(fmt.Sprintf("PHP %s requires additional compilation for GD support", version), "")
	logger.Info("GD extension enables image processing (uploads, thumbnails, watermarks)")
	logger.Info("This adds 2-3 minutes to installation time")

	logger.PrintBlock("\nWould you like to enable GD image processing support?\n  1) Enable GD (recommended for image processing)\n  2) Skip GD (faster installation)")
	logger.Prompt("Enter your choice (1-2, default=2):", "")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("Could not read GD choice", "Skipping GD extension")
		return false, nil
	}

	choice := strings.TrimSpace(input)
	switch choice {
	case "1":
		logger.Info("GD extension will be enabled")
		return true, nil
	case "2", "":
		logger.Info("GD extension will be skipped")
		return false, nil
	default:
		logger.Warn("Invalid choice", "Skipping GD extension")
		return false, nil
	}
}

/**
 * writeOpenSSLConf creates OpenSSL configuration with proper CA bundle paths.
 * This automatically detects the Linux distribution and sets appropriate certificate paths.
 *
 * @param prefix Installation prefix directory
 * @param version PHP version
 * @return error if configuration creation fails
 */
func WriteOpenSSLConf(prefix, version string) error {
	phpDir := filepath.Join(prefix, "php", version)

	// Create conf.d directory if it doesn't exist
	confDDir := filepath.Join(phpDir, "etc", "conf.d")
	if err := os.MkdirAll(confDDir, 0o755); err != nil {
		return fmt.Errorf("failed to create extension config directory: %w", err)
	}

	// Detect Linux distribution and get appropriate CA bundle paths
	caBundlePath, caDirPath := getCAPaths()

	// Create OpenSSL configuration with proper certificate paths
	opensslConfContent := fmt.Sprintf(`; Chauffeur OpenSSL Configuration for PHP %s
; Auto-generated for %s distribution
;
; Certificate Authority settings for SSL/TLS verification
; This configuration enables secure connections to remote services
; including SMTP servers, HTTPS APIs, and other TLS-enabled services
;

; Path to the CA bundle file containing trusted root certificates
openssl.cafile = %s

; Path to the directory containing CA certificates
openssl.capath = %s

; Additional SSL/TLS security settings (optional but recommended)
; These settings enhance security for SSL/TLS connections

; Minimum TLS version (modern secure setting)
; Uncomment the line below to require TLS 1.2 or higher
; openssl.ciphers = HIGH:!aNULL:!MD5

; Enable strict certificate verification by default
; This is now the default in modern PHP versions
; openssl.verify_peer = 1
; openssl.verify_peer_name = 1
`, version, detectLinuxDistribution(), caBundlePath, caDirPath)

	opensslConfPath := filepath.Join(confDDir, "openssl.ini")
	if err := os.WriteFile(opensslConfPath, []byte(opensslConfContent), 0o644); err != nil {
		return fmt.Errorf("failed to create openssl.ini: %w", err)
	}

	return nil
}

/**
 * detectLinuxDistribution detects the current Linux distribution.
 *
 * @return distribution name as string
 */
func detectLinuxDistribution() string {
	if _, err := os.Stat("/etc/os-release"); err == nil {
		data, readErr := os.ReadFile("/etc/os-release")
		if readErr == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "ID=") {
					return strings.TrimPrefix(line, "ID=")
				}
			}
		}
	}

	// Fallback detection methods
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return "rhel"
	}
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}

	return "unknown"
}

/**
 * getCAPaths returns the appropriate CA bundle and CA directory paths
 * based on the detected Linux distribution.
 *
 * @return caBundlePath, caDirPath
 */
func getCAPaths() (string, string) {
	distro := detectLinuxDistribution()

	switch distro {
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		return "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", "/etc/pki/tls/certs"
	case "ubuntu", "debian":
		return "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs"
	case "arch":
		return "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs"
	case "opensuse", "sles":
		return "/etc/ssl/ca-bundle.pem", "/etc/ssl/certs"
	default:
		// Fallback: try common locations
		for _, path := range []string{
			"/etc/pki/tls/cert.pem",
			"/etc/ssl/certs/ca-certificates.crt",
			"/etc/ssl/ca-bundle.pem",
		} {
			if _, err := os.Stat(path); err == nil {
				return path, "/etc/ssl/certs"
			}
		}
		// Default fallback
		return "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs"
	}
}

// RecreateExistingPHPShims recreates shims for any existing PHP installations found in the workspace.
// This is useful after uninstall/reinstall to restore shim functionality.
func RecreateExistingPHPShims(prefix string) (int, error) {
	phpDir := filepath.Join(prefix, "php")

	// Check if php directory exists
	if _, err := os.Stat(phpDir); os.IsNotExist(err) {
		return 0, nil // No PHP installations, nothing to do
	}

	// Scan for PHP version directories
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return 0, fmt.Errorf("read php directory: %w", err)
	}

	supportedVersions := GetSupportedPHPVersions()
	supportedMap := make(map[string]bool)
	for _, v := range supportedVersions {
		supportedMap[v.Version] = true
	}

	var recreated []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version := entry.Name()

		// Skip if this is not a supported version
		if !supportedMap[version] {
			continue
		}

		// Check if PHP binary exists
		binaryPath := filepath.Join(phpDir, version, "bin", phpBinaryName)
		if _, err := os.Stat(binaryPath); err != nil {
			continue // Binary doesn't exist, skip
		}

		// Check if versioned shim already exists
		shimName := fmt.Sprintf("php-%s", version)
		shimPath := filepath.Join(prefix, "bin", shimName)
		if _, err := os.Stat(shimPath); err == nil {
			continue // Shim already exists, skip
		}

		// Create the versioned shim
		if err := writeShim(prefix, shimName, binaryPath); err != nil {
			return 0, fmt.Errorf("write shim for php %s: %w", version, err)
		}

		recreated = append(recreated, version)

		// If this is the default version, also create the default shim
		defaultVersion, err := config.GetDefaultPHPVersion()
		if err == nil && defaultVersion == version {
			if err := writeShim(prefix, phpBinaryName, binaryPath); err != nil {
				return 0, fmt.Errorf("write default php shim: %w", err)
			}
		}
	}

	return len(recreated), nil
}
