package installers

import (
	"archive/tar"
	"compress/gzip"
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

const (
	imagickVersion    = "3.7.0"
	imagickTarball    = "imagick-" + imagickVersion + ".tgz"
	imagickSourceDir  = "imagick-" + imagickVersion
	imagickSourceURL  = "https://pecl.php.net/get/" + imagickTarball
	imagickIniContent = "extension=imagick\n"
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
}

/**
 * GetSupportedPHPVersions returns the list of supported PHP versions.
 *
 * @return list of supported PHP versions
 */
func GetSupportedPHPVersions() []PHPVersion {
	return []PHPVersion{
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
	if err := ensurePHPBuildDependencies(depsLogger); err != nil {
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

	keepBuildDir := os.Getenv("CHAUFFEUR_KEEP_BUILD_DIR") == "1"
	if !keepBuildDir {
		defer os.RemoveAll(tmpDir)
	} else {
		phpLogger.Warn("preserving PHP build directory for debugging", tmpDir)
	}

	// Get latest patch version for the requested major.minor
	patchVersion, err := getLatestPHPPatchVersion(opts.Client, version)
	if err != nil {
		return fmt.Errorf("resolve latest PHP %s patch version: %w", version, err)
	}

	prepareLogger.Info(fmt.Sprintf("Latest patch version: %s", patchVersion))

	tarballName := fmt.Sprintf("php-%s.tar.gz", patchVersion)

	// Try official mirrors first, then build servers
	tarballURLs := []string{
		fmt.Sprintf("https://www.php.net/distributions/%s", tarballName),
		fmt.Sprintf("https://secure.php.net/distributions/%s", tarballName),
		fmt.Sprintf("https://downloads.php.net/~php/%s", tarballName),
	}

	var tarballURL string
	var downloadErr error

	phpLogger.Info("Downloading")
	downloadLogger := phpLogger.NewChildLogger("download")
	tarballPath := filepath.Join(tmpDir, tarballName)

	if localTarball := os.Getenv("CHAUFFEUR_PHP_TARBALL"); localTarball != "" {
		if err := copyFile(localTarball, tarballPath); err != nil {
			downloadLogger.Warn("Failed to copy local tarball", err.Error())
		} else {
			tarballURL = "local"
			downloadLogger.Success(fmt.Sprintf("Reused local %s", tarballName), localTarball)
		}
	}

	if tarballURL == "" {
		downloadLogger.Info(fmt.Sprintf("Attempting download of %s", tarballName))
		for _, url := range tarballURLs {
			size, err := lib.DownloadToFileWithLogger(opts.Client, url, tarballPath, fmt.Sprintf("Download %s", tarballName), downloadLogger)
			if err == nil {
				tarballURL = url
				downloadLogger.Success(fmt.Sprintf("Downloaded %s", tarballName), lib.HumanBytes(size))
				break
			}
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

	return nil
}

/**
 * getLatestPHPPatchVersion discovers the latest patch version for a given major.minor.
 *
 * @param client  HTTP client for requests
 * @param version Major.minor version (e.g., "8.3")
 * @return Latest full version string (e.g., "8.3.14") and error
 */
func getLatestPHPPatchVersion(client *http.Client, version string) (string, error) {
	// Fallback to pinned patch versions to avoid relying on network connectivity
	knownVersions := map[string]string{
		"8.3": "8.3.14",
		"8.2": "8.2.26",
		"8.1": "8.1.31",
		"8.0": "8.0.30",
		"7.4": "7.4.33",
	}

	if patch, exists := knownVersions[version]; exists {
		return patch, nil
	}

	return "", fmt.Errorf("no patch version mapping for PHP %s", version)
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
	}

	switch version {
	case "7.4":
		// placeholder for future 7.4-specific options
	case "8.0":
		// placeholder for future 8.0-specific options
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

	tmpDir, err := os.MkdirTemp("", "chauffeur-imagick-*")
	if err != nil {
		return fmt.Errorf("create imagick temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, imagickTarball)
	logger.Info(fmt.Sprintf("Downloading imagick %s", imagickVersion))
	if _, err := lib.DownloadToFileWithLogger(opts.Client, imagickSourceURL, tarballPath, fmt.Sprintf("imagick %s", imagickVersion), logger); err != nil {
		return logger.Fail("download imagick extension", err.Error())
	}

	logger.Info("Extracting imagick sources")
	if err := untarPHP(tarballPath, tmpDir); err != nil {
		return logger.Fail("extract imagick extension", err.Error())
	}

	sourceDir := filepath.Join(tmpDir, imagickSourceDir)
	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("imagick source directory missing: %w", err)
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

func ensurePHPBuildDependencies(logger *lib.Logger) error {
	if logger == nil {
		logger = lib.NewCommandLogger("deps")
	}

	logger.Info("Verifying pkg-config availability")
	if err := ensurePkgConfigAvailable(logger); err != nil {
		return err
	}
	logger.Success("pkg-config found", "")

	for _, req := range phpPkgRequirements {
		logger.Info(fmt.Sprintf("Checking %s development headers", req.Name))
		version, err := ensurePkgRequirement(logger, req)
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
