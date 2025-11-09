package installers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/lib"
)

const nginxBinaryName = "nginx"

/**
 * nginxSigningKey represents a trusted nginx signing key and its metadata.
 */
type nginxSigningKey struct {
	Name        string
	URL         string
	Fingerprint string
	UIDs        []string
	Optional    bool
}

var nginxSigningKeys = []nginxSigningKey{
	{
		Name:        "nginx signing key <signing-key@nginx.com>",
		URL:         "https://nginx.org/keys/nginx_signing.key",
		Fingerprint: "ABF5BD826BD0E86F09B160A25EAB6AB0133E7F96",
		UIDs:        []string{"nginx signing key <signing-key@nginx.com>", "nginx signing key"},
		Optional:    true,
	},
	{
		Name:        "Roman Arutyunyan <r.arutyunyan@f5.com>",
		URL:         "https://nginx.org/keys/arut.key",
		Fingerprint: "8540A6F18833A80E9C1653A42FD21310B49F6B46",
		UIDs:        []string{"Roman Arutyunyan <r.arutyunyan@f5.com>", "Roman Arutyunyan", "r.arutyunyan@f5.com"},
		Optional:    false,
	},
	{
		Name:        "Sergey Kandaurov <pluknet@nginx.com>",
		URL:         "https://nginx.org/keys/pluknet.key",
		Fingerprint: "8540A6F18833A80E9C1653A42FD21310B49F6B46",
		UIDs:        []string{"Sergey Kandaurov <pluknet@nginx.com>", "Sergey Kandaurov", "pluknet@nginx.com"},
		Optional:    false,
	},
	{
		Name:        "Serguei Beloussov <sb@openmailbox.org>",
		URL:         "https://nginx.org/keys/sb.key",
		Fingerprint: "8540A6F18833A80E9C1653A42FD21310B49F6B46",
		UIDs:        []string{"Serguei Beloussov", "sb@openmailbox.org"},
		Optional:    false,
	},
	{
		Name:        "Denys Fedoryshchenko <thresh@nginx.com>",
		URL:         "https://nginx.org/keys/thresh.key",
		Fingerprint: "8540A6F18833A80E9C1653A42FD21310B49F6B46",
		UIDs:        []string{"Denys Fedoryshchenko <thresh@nginx.com>", "Denys Fedoryshchenko", "thresh@nginx.com"},
		Optional:    false,
	},
}

/**
 * InstallNginxSource compiles Nginx from source into the Chauffeur workspace.
 *
 * @param opts Installer configuration containing prefix, force flag, and client.
 * @return error when installation cannot complete successfully.
 */
func InstallNginxSource(opts InstallOptions) error {
	nginxLogger := lib.NewCommandLogger("nginx")

	if opts.Prefix == "" {
		return nginxLogger.Fail("install prefix is required", "")
	}

	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 60 * time.Second}
	}

	release, err := releases.LatestGitHubRelease(opts.Client, "nginx", "nginx")
	if err != nil {
		return nginxLogger.Fail("resolve latest Nginx release", err.Error())
	}

	tag := release.TagName
	if tag == "" {
		return nginxLogger.Fail("latest Nginx release has empty tag name", "")
	}

	version := strings.TrimPrefix(tag, "release-")
	if version == tag {
		version = strings.TrimPrefix(tag, "v")
	}
	if version == "" {
		return nginxLogger.Fail("cannot derive version from tag", tag)
	}

	nginxLogger.Info("Installing nginx (source build from nginx.org release)...")
	nginxLogger.Info(fmt.Sprintf("Release tag: %s (version %s)", tag, version))
	nginxLogger.Info(fmt.Sprintf("Install prefix: %s", filepath.Join(opts.Prefix, "nginx")))

	binaryPath := filepath.Join(opts.Prefix, "nginx", "sbin", nginxBinaryName)
	if !opts.Force {
		if info, err := os.Stat(binaryPath); err == nil && info.Mode().IsRegular() {
			return nil
		}
	}

	nginxLogger.Info("Download")
	tmpDir, err := os.MkdirTemp("", "chauffeur-nginx-*")
	if err != nil {
		return nginxLogger.Fail("create temp dir", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	tarballName := fmt.Sprintf("nginx-%s.tar.gz", version)
	tarballURL, ok := release.AssetURL(tarballName)
	if !ok {
		tarballURL = fmt.Sprintf("https://github.com/nginx/nginx/releases/download/%s/%s", tag, tarballName)
	}

	tarballPath := filepath.Join(tmpDir, tarballName)
	downloadLogger := nginxLogger.NewChildLogger("download")
	size, err := lib.DownloadToFileWithLogger(opts.Client, tarballURL, tarballPath, fmt.Sprintf("Download %s", tarballName), downloadLogger)
	if err != nil {
		return nginxLogger.Fail("download nginx source", err.Error())
	}
	downloadLogger.Success(fmt.Sprintf("Downloaded %s", tarballName), lib.HumanBytes(size))

	nginxLogger.Info("Verifying")
	verificationLogger := nginxLogger.NewChildLogger("verifying")
	if err := verifyNginxSignature(opts.Client, tarballPath, tarballName, tmpDir, verificationLogger); err != nil {
		return nginxLogger.Fail("verify PGP signature", err.Error())
	}

	sum, algo, err := fetchNginxChecksum(opts.Client, release, tag, version, tarballName)
	if err != nil {
		return nginxLogger.Fail("resolve nginx checksum", err.Error())
	}

	if err := handleChecksum(tarballPath, sum, algo, verificationLogger); err != nil {
		return err
	}

	nginxLogger.Info("Building")
	buildLogger := nginxLogger.NewChildLogger("build")
	extractRoot := filepath.Join(tmpDir, "src")
	sourceDir := filepath.Join(extractRoot, fmt.Sprintf("nginx-%s", version))
	buildLogger.Info(fmt.Sprintf("Extracting sources to %s", extractRoot))
	if err := untar(tarballPath, extractRoot); err != nil {
		return nginxLogger.Fail("extract nginx source", err.Error())
	}
	buildLogger.Success("Sources extracted", "")

	buildLogger.Info("Configuring and compiling nginx")
	if err := buildAndInstallNginx(opts.Prefix, sourceDir); err != nil {
		return err
	}
	buildLogger.Success("nginx built and installed", filepath.Join(opts.Prefix, "nginx"))

	nginxLogger.Info("Finalizing")
	finalizeLogger := nginxLogger.NewChildLogger("finalize")
	finalizeLogger.Info("Ensuring workspace layout")
	if err := ensureNginxLayout(opts.Prefix); err != nil {
		return err
	}
	finalizeLogger.Success("Runtime layout ready", "")

	finalizeLogger.Info("Writing nginx shim")
	if err := writeShim(opts.Prefix, nginxBinaryName, binaryPath); err != nil {
		return err
	}
	finalizeLogger.Success("Shim written", filepath.Join(opts.Prefix, "bin", nginxBinaryName))

	finalizeLogger.Info("Writing default configuration")
	if err := writeDefaultNginxConf(opts.Prefix); err != nil {
		return err
	}
	finalizeLogger.Success("Default nginx.conf ready", "")

	return nil
}

/**
 * fetchNginxChecksum discovers an appropriate checksum for the requested tarball.
 *
 * @param client      HTTP client used for downloading assets.
 * @param release     GitHub release metadata for nginx.
 * @param tag         Original release tag string.
 * @param version     Normalized version string.
 * @param tarballName Target tarball file name.
 * @return Matching checksum string, algorithm name, or empty string when none found.
 */
func fetchNginxChecksum(client *http.Client, release releases.GitHubRelease, tag, version, tarballName string) (string, string, error) {
	// Prefer release-specific checksum assets.
	if url, ok := release.AssetURL(tarballName + ".sha512"); ok {
		content, err := lib.DownloadText(client, url)
		if err == nil {
			if sum, err := lib.ChecksumFromContent(content, tarballName); err == nil {
				sum = strings.ToLower(sum)
				return sum, "sha512", nil
			}
			if sum, err := lib.ChecksumFromContent(content, ""); err == nil {
				sum = strings.ToLower(sum)
				return sum, "sha512", nil
			}
		}
	}
	if url, ok := release.AssetURL(tarballName + ".sha256"); ok {
		content, err := lib.DownloadText(client, url)
		if err == nil {
			if sum, err := lib.ChecksumFromContent(content, tarballName); err == nil {
				sum = strings.ToLower(sum)
				return sum, "sha256", nil
			}
			if sum, err := lib.ChecksumFromContent(content, ""); err == nil {
				sum = strings.ToLower(sum)
				return sum, "sha256", nil
			}
		}
	}

	listCandidates := []string{
		"sha512sums.txt",
		"SHA512SUMS",
		"checksums.txt",
	}
	for _, candidate := range listCandidates {
		if url, ok := release.AssetURL(candidate); ok {
			sum, err := lib.ChecksumFromList(client, url, tarballName)
			if err == nil {
				sum = strings.ToLower(sum)
				return sum, algorithmFromChecksum(sum, "sha512"), nil
			}
		}
	}

	// Fall back to upstream official checksum.
	fallbackURL := fmt.Sprintf("https://nginx.org/download/%s.sha512", tarballName)
	content, err := downloadText(client, fallbackURL)
	if err == nil {
		if sum, err := lib.ChecksumFromContent(content, tarballName); err == nil {
			sum = strings.ToLower(sum)
			return sum, "sha512", nil
		}
		if sum, err := lib.ChecksumFromContent(content, ""); err == nil {
			sum = strings.ToLower(sum)
			return sum, "sha512", nil
		}
	}

	fallbackURL = fmt.Sprintf("https://nginx.org/download/%s.sha256", tarballName)
	content, err = downloadText(client, fallbackURL)
	if err == nil {
		if sum, err := lib.ChecksumFromContent(content, tarballName); err == nil {
			sum = strings.ToLower(sum)
			return sum, "sha256", nil
		}
		if sum, err := lib.ChecksumFromContent(content, ""); err == nil {
			sum = strings.ToLower(sum)
			return sum, "sha256", nil
		}
	}

	return "", "", nil
}

/**
 * verifyNginxSignature validates the tarball using upstream detached signatures and pinned keys.
 *
 * @param client      HTTP client used for downloads.
 * @param tarballPath Local path to the downloaded tarball.
 * @param tarballName Tarball file name (used for forming URLs).
 * @param workDir     Workspace directory for temporary signature/key storage.
 * @return error when signature verification fails.
 */
func verifyNginxSignature(client *http.Client, tarballPath, tarballName, workDir string, parentLogger *lib.Logger) error {
	logger := parentLogger.NewChildLogger("verifying")

	if _, err := exec.LookPath("gpg"); err != nil {
		return logger.Fail("gpg not found in PATH", err.Error())
	}

	sigURL := fmt.Sprintf("https://nginx.org/download/%s.asc", tarballName)
	sigPath := tarballPath + ".asc"
	// Signature download happens silently, progress shown via progress bar
	if _, err := lib.DownloadToFileWithLogger(client, sigURL, sigPath, fmt.Sprintf("Signature %s", tarballName), logger); err != nil {
		return fmt.Errorf("download signature: %w", err)
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

	for idx, key := range nginxSigningKeys {
		keyFile := filepath.Join(keysDir, fmt.Sprintf("nginx-key-%d.asc", idx))
		logger.Info(fmt.Sprintf("Fetching signing key: %s", key.Name))
		if _, err := lib.DownloadToFileWithLogger(client, key.URL, keyFile, fmt.Sprintf("Key %s", key.Name), logger); err != nil {
			return fmt.Errorf("download signing key %s: %w", key.Name, err)
		}
		if err := importNginxKey(gpgHome, keyFile, key.Fingerprint, key.Optional, key.Name); err != nil {
			return err
		}
	}

	if err := verifySignatureWithGPG(gpgHome, tarballPath, sigPath); err != nil {
		return err
	}
	logger.Success("PGP signature verified successfully", "")
	return nil
}

/**
 * importNginxKey imports a signing key after ensuring its fingerprint matches the pinned value.
 *
 * @param gpgHome     Isolated gpg home directory.
 * @param keyPath     Path to the downloaded key file.
 * @param fingerprint Expected fingerprint (no spaces).
 * @param optional    Whether the key is optional (mismatch produces warning instead of error).
 * @param name        Human-readable key label for logging.
 * @return error when fingerprint mismatches for mandatory keys or import fails.
 */
func importNginxKey(gpgHome, keyPath, fingerprint string, optional bool, name string) error {
	logger := lib.NewCommandLogger("install")

	fp, err := readKeyFingerprint(gpgHome, keyPath)
	if err != nil {
		if optional {
			logger.Warn("signing key could not be inspected", fmt.Sprintf("%s (%v); skipping optional key", name, err))
			return nil
		}
		return err
	}

	shouldImport, err := evaluateFingerprint(fp, fingerprint, name, optional)
	if err != nil {
		if optional {
			logger.Warn("signing key skipped", fmt.Sprintf("%s (%v)", name, err))
			return nil
		}
		return err
	}
	if !shouldImport {
		return nil
	}

	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--batch",
		"--import", keyPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gpg import failed: %w\n%s", err, out)
	}
	return nil
}

/**
 * readKeyFingerprint extracts the fingerprint from a key file without importing it.
 *
 * @param gpgHome Isolated gpg home directory.
 * @param keyPath Path to the downloaded key file.
 * @return Fingerprint string with uppercase hex characters.
 */
func readKeyFingerprint(gpgHome, keyPath string) (string, error) {
	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--import-options", "show-only",
		"--dry-run",
		"--with-colons",
		"--fingerprint",
		keyPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fp, fallbackErr := readFingerprintFallback(gpgHome, keyPath)
		if fallbackErr != nil {
			return "", fmt.Errorf("gpg fingerprint check failed: %w\n%s", err, out)
		}
		return fp, nil
	}

	fp := parseFingerprintOutput(string(out))
	if fp == "" {
		return "", errors.New("fingerprint not found in key material")
	}
	return fp, nil
}

func parseFingerprintOutput(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "fpr:") {
			parts := strings.Split(line, ":")
			if len(parts) > 9 {
				return strings.ToUpper(parts[9])
			}
		}
	}
	return ""
}

func readFingerprintFallback(gpgHome, keyPath string) (string, error) {
	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--batch",
		"--import", keyPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("fallback import failed: %w\n%s", err, out)
	}

	cmd = exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--with-colons",
		"--fingerprint",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("fallback fingerprint read failed: %w\n%s", err, out)
	}
	return parseFingerprintOutput(string(out)), nil
}

/**
 * evaluateFingerprint decides whether to import a key based on expected vs actual fingerprint.
 *
 * @param actual    Fingerprint reported by gpg for the key file.
 * @param expected  Pinned fingerprint from nginx.org.
 * @param name      Human-readable key label for logging.
 * @param optional  Whether the key is optional.
 * @return Tuple (shouldImport, error). When optional mismatches occur, returns false and nil.
 */
func evaluateFingerprint(actual, expected, name string, optional bool) (bool, error) {
	logger := lib.NewCommandLogger("install")

	if strings.EqualFold(actual, expected) {
		return true, nil
	}

	if optional {
		logger.Warn("signing key fingerprint mismatch", fmt.Sprintf("%s (expected %s, got %s); skipping optional key", name, expected, actual))
		return false, nil
	}

	return false, fmt.Errorf("unexpected key fingerprint for %s: got %s expected %s", name, actual, expected)
}

/**
 * verifySignatureWithGPG validates the tarball signature and enforces trusted fingerprints.
 *
 * @param gpgHome     Isolated gpg home directory populated with trusted keys.
 * @param tarballPath Path to the downloaded tarball.
 * @param sigPath     Path to the detached signature file.
 * @return error when signature validation fails or signer is untrusted.
 */
func verifySignatureWithGPG(gpgHome, tarballPath, sigPath string) error {
	cmd := exec.Command("gpg",
		"--homedir", gpgHome,
		"--no-default-keyring",
		"--trust-model", "always",
		"--batch",
		"--status-fd=1",
		"--verify", sigPath, tarballPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpg signature verification failed: %w\n%s", err, out)
	}

	if !parseValidSignatures(string(out)) {
		return errors.New("signature verified but signer fingerprint is not in the trusted set")
	}

	return nil
}

/**
 * parseValidSignatures scans GPG status output for trusted signer fingerprints.
 *
 * @param status Raw status output from gpg --status-fd.
 * @return true when a trusted fingerprint is observed in VALIDSIG records.
 */
func parseValidSignatures(status string) bool {
	fpTrusted := map[string]struct{}{}
	var uidTrusted []string
	for _, k := range nginxSigningKeys {
		if k.Fingerprint != "" {
			fpTrusted[strings.ToUpper(strings.TrimSpace(k.Fingerprint))] = struct{}{}
		}
		for _, u := range k.UIDs {
			uidTrusted = append(uidTrusted, strings.ToLower(strings.TrimSpace(u)))
		}
	}

	var signerFPs []string
	var signerUIDs []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "[GNUPG:] VALIDSIG "):
			parts := strings.Fields(line)
			for _, token := range parts[1:] {
				value := strings.ToUpper(strings.TrimSpace(token))
				if len(value) == 40 && isHexString(value) {
					signerFPs = append(signerFPs, value)
				}
			}
		case strings.HasPrefix(line, "[GNUPG:] GOODSIG "):
			idx := strings.Index(line, "GOODSIG ")
			if idx >= 0 {
				rest := strings.TrimSpace(line[idx+8:])
				sp := strings.IndexByte(rest, ' ')
				if sp > 0 && sp+1 < len(rest) {
					uid := strings.ToLower(strings.TrimSpace(rest[sp+1:]))
					if uid != "" {
						signerUIDs = append(signerUIDs, uid)
					}
				}
			}
		}
	}
	for _, fp := range signerFPs {
		if _, ok := fpTrusted[fp]; ok {
			return true
		}
	}
	for _, seen := range signerUIDs {
		for _, allow := range uidTrusted {
			if seen == allow || strings.Contains(seen, allow) {
				return true
			}
		}
	}
	return false
}

func isHexString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

func algorithmFromChecksum(sum, defaultAlgo string) string {
	switch len(sum) {
	case 128:
		return "sha512"
	case 64:
		return "sha256"
	default:
		return defaultAlgo
	}
}

func handleChecksum(tarballPath, sum, algo string, logger *lib.Logger) error {
	sum = strings.ToLower(strings.TrimSpace(sum))
	algo = strings.ToLower(strings.TrimSpace(algo))

	if sum == "" {
		if os.Getenv("CHAUF_REQUIRE_CHECKSUM") == "1" {
			return logger.Fail("ERROR: nginx checksum required but no upstream checksum asset found", "")
		}
		logger.Info("No upstream checksum asset found; relying on PGP signature.")
		if sha256Hex, err := fileSHA256(tarballPath); err == nil {
			logger.Info(fmt.Sprintf("Local SHA256: %s", sha256Hex))
		} else {
			logger.Warn("failed to compute local SHA256", err.Error())
		}
		if sha512Hex, err := fileSHA512(tarballPath); err == nil {
			logger.Info(fmt.Sprintf("Local SHA512: %s", sha512Hex))
		} else {
			logger.Warn("failed to compute local SHA512", err.Error())
		}
		return nil
	}

	if err := validateChecksum(tarballPath, sum); err != nil {
		actual := "unknown"
		switch algo {
		case "sha512":
			if digest, hashErr := fileSHA512(tarballPath); hashErr == nil {
				actual = digest
			}
		case "sha256":
			if digest, hashErr := fileSHA256(tarballPath); hashErr == nil {
				actual = digest
			}
		default:
			if digest, hashErr := fileSHA512(tarballPath); hashErr == nil {
				actual = digest
			}
		}
		return fmt.Errorf("ERROR: nginx tarball checksum mismatch (expected %s got %s)", sum, actual)
	}

	if algo == "" {
		algo = algorithmFromChecksum(sum, "sha512")
	}
	installLogger := lib.NewCommandLogger("install")
	installLogger.Info(fmt.Sprintf("Upstream checksum verified (%s).", algo))
	return nil
}

/**
 * untar extracts the tarball contents into dest while preserving permissions.
 *
 * @param tarball Path to the downloaded nginx source tarball.
 * @param dest    Destination directory for extraction.
 * @return error when extraction fails.
 */
func untar(tarball, dest string) error {
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
 * buildAndInstallNginx configures, builds, and installs nginx into the workspace.
 *
 * @param prefix    Chauffeur workspace root.
 * @param sourceDir Directory containing the extracted nginx sources.
 * @return error when configure/build/install steps fail.
 */
func buildAndInstallNginx(prefix, sourceDir string) error {
	// Start spinner for nginx configuration and compilation
	spin := lib.NewSpinner("nginx", "Configuring and compiling nginx")

	confArgs := []string{
		"--prefix=" + filepath.Join(prefix, "nginx"),
		"--conf-path=" + filepath.Join(prefix, "nginx", "etc", "nginx.conf"),
		"--pid-path=" + filepath.Join(prefix, "nginx", "nginx.pid"),
		"--http-log-path=" + filepath.Join(prefix, "nginx", "logs", "access.log"),
		"--error-log-path=" + filepath.Join(prefix, "nginx", "logs", "error.log"),
		"--with-pcre-jit",
		"--with-http_gzip_static_module",
		"--with-http_ssl_module",
	}

	if err := runCommand(sourceDir, "./configure", confArgs...); err != nil {
		spin.Fail("configuration failed")
		return fmt.Errorf("configure nginx: %w", err)
	}

	makeArgs := []string{"-j"}
	if n := runtime.NumCPU(); n > 0 {
		makeArgs = append(makeArgs, fmt.Sprintf("%d", n))
	}
	if err := runCommand(sourceDir, "make", makeArgs...); err != nil {
		spin.Fail("compilation failed")
		return fmt.Errorf("make nginx: %w", err)
	}

	if err := runCommand(sourceDir, "make", "install"); err != nil {
		spin.Fail("installation failed")
		return fmt.Errorf("make install nginx: %w", err)
	}

	spin.Success(filepath.Join(prefix, "nginx", "sbin", "nginx"))
	return nil
}

/**
 * runCommand executes a command inside dir and returns stderr/stdout on failure.
 *
 * @param dir  Working directory for execution.
 * @param name Command binary to run.
 * @param args Additional command arguments.
 * @return error when the command exits non-zero.
 */
func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return nil
}

/**
 * ensureNginxLayout creates the directory structure nginx expects inside the workspace.
 *
 * @param prefix Chauffeur workspace root.
 * @return error if any directory cannot be created.
 */
func ensureNginxLayout(prefix string) error {
	paths := []string{
		filepath.Join(prefix, "nginx", "etc"),
		filepath.Join(prefix, "nginx", "conf.d"),
		filepath.Join(prefix, "nginx", "sites-available"),
		filepath.Join(prefix, "nginx", "sites-enabled"),
		filepath.Join(prefix, "nginx", "logs"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("ensure nginx path %s: %w", path, err)
		}
	}

	return nil
}

/**
 * writeDefaultNginxConf seeds the workspace with a minimal nginx.conf and helper file.
 *
 * @param prefix Chauffeur workspace root.
 * @return error when configuration files cannot be written.
 */
func writeDefaultNginxConf(prefix string) error {
	confPath := filepath.Join(prefix, "nginx", "etc", "nginx.conf")
	confDir := filepath.Dir(confPath)
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		return fmt.Errorf("ensure nginx etc: %w", err)
	}

	// Backup vendor-provided config if it exists
	if _, err := os.Stat(confPath); err == nil {
		if err := os.Rename(confPath, confPath+".dist"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("backup nginx.conf: %w", err)
		}
	}

	nginxDir := filepath.Join(prefix, "nginx")
	logsDir := filepath.Join(nginxDir, "logs")
	sitesAvailable := filepath.Join(nginxDir, "sites-available")
	sitesEnabled := filepath.Join(nginxDir, "sites-enabled")
	confD := filepath.Join(nginxDir, "conf.d")
	mimeTarget := filepath.Join(confDir, "mime.types")

	for _, dir := range []string{logsDir, sitesAvailable, sitesEnabled, confD} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ensure nginx directory %s: %w", dir, err)
		}
	}

	// Ensure mime.types exists under etc/
	if _, err := os.Stat(mimeTarget); errors.Is(err, os.ErrNotExist) {
		sourceCandidates := []string{
			filepath.Join(nginxDir, "conf", "mime.types"),
			filepath.Join(nginxDir, "mime.types"),
		}
		var copied bool
		for _, candidate := range sourceCandidates {
			if data, readErr := os.ReadFile(candidate); readErr == nil {
				if err := os.WriteFile(mimeTarget, data, 0o644); err != nil {
					return fmt.Errorf("write mime.types: %w", err)
				}
				copied = true
				break
			}
		}
		if !copied {
			return fmt.Errorf("mime.types not found in nginx installation")
		}
	}

	content := fmt.Sprintf(`# Chauffeur Nginx Configuration
# Custom ports to avoid conflicts with system services

worker_processes auto;
error_log %s/error.log notice;
pid %s/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include       %s;
    default_type  application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log %s/access.log main;

    sendfile        on;
    tcp_nopush      on;
    tcp_nodelay     on;
    keepalive_timeout  65;
    types_hash_max_size 2048;

    include %s/conf.d/*.conf;
    include %s/*.conf;
}
`,
		logsDir,
		logsDir,
		mimeTarget,
		logsDir,
		nginxDir,
		sitesEnabled,
	)

	if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write nginx.conf: %w", err)
	}

	return nil
}
