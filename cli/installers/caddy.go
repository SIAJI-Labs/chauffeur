package installers

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/internal/system"
)

const caddyBinaryName = "caddy"

// InstallOptions drives installer behavior.
type InstallOptions struct {
	Prefix string
	Force  bool
	Info   system.Info
	Client *http.Client
}

/**
 * InstallCaddyTarball downloads and places the Caddy binary inside the workspace.
 *
 * @param opts Installer configuration such as prefix, force flag, and host info.
 * @return error when the installation fails at any step.
 */
func InstallCaddyTarball(opts InstallOptions) error {
	if opts.Prefix == "" {
		return errors.New("install prefix is required")
	}

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	startCaddyLogSection("Preparing")
	logCaddyInfo("Detected architecture: %s", opts.Info.Arch)
	logCaddyInfo("Fetching release metadata from GitHub…")
	release, err := releases.LatestGitHubRelease(client, "caddyserver", "caddy")
	if err != nil {
		return fmt.Errorf("resolve latest Caddy release: %w", err)
	}
	versionTag := release.TagName
	if versionTag == "" {
		return errors.New("latest Caddy release has empty tag name")
	}
	version := strings.TrimPrefix(versionTag, "v")
	logCaddySuccess("Latest release: %s (tag %s)", version, versionTag)

	assetName, tarballURL, err := selectCaddyAsset(release, version, opts.Info.Arch)
	if err != nil {
		return err
	}
	startCaddyLogSection("Download")
	logCaddyInfo("Resolved asset: %s", assetName)
	logCaddyInfo("Source URL: %s", tarballURL)

	checksumURL, checksumIsList, err := locateCaddyChecksum(release, assetName, version)
	if err != nil {
		return err
	}
	logCaddyInfo("Checksum source: %s", checksumURL)

	destBinDir := filepath.Join(opts.Prefix, "caddy", "bin")
	if err := os.MkdirAll(destBinDir, 0o755); err != nil {
		return fmt.Errorf("ensure caddy bin dir: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "chauffeur-caddy-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, assetName)
	logCaddyInfo("Downloading tarball into %s", tmpDir)
	size, err := downloadToFile(client, tarballURL, tarballPath, fmt.Sprintf("Download %s", assetName))
	if err != nil {
		return fmt.Errorf("download caddy tarball: %w", err)
	}
	logCaddySuccess("Downloaded %s (%d bytes)", assetName, size)

	startCaddyLogSection("Verification")
	logCaddyInfo("Verifying checksum…")
	expectedSum, err := fetchCaddyChecksum(client, checksumURL, assetName, checksumIsList)
	if err != nil {
		return fmt.Errorf("resolve checksum: %w", err)
	}
	if err := validateChecksum(tarballPath, expectedSum); err != nil {
		return fmt.Errorf("validate caddy tarball: %w", err)
	}
	logCaddySuccess("Checksum verification passed")

	targetBinary := filepath.Join(destBinDir, caddyBinaryName)
	if !opts.Force {
		if info, err := os.Stat(targetBinary); err == nil && info.Mode().IsRegular() {
			logCaddyInfo("Existing binary detected; skipping extraction (use --force to overwrite)")
			return nil
		}
	}

	startCaddyLogSection("Installation")
	if err := extractBinary(tarballPath, targetBinary); err != nil {
		return fmt.Errorf("extract caddy binary: %w", err)
	}
	logCaddySuccess("Installed binary at %s", targetBinary)

	if err := writeShim(opts.Prefix, caddyBinaryName, targetBinary); err != nil {
		return err
	}
	logCaddySuccess("Updated shim %s", filepath.Join(opts.Prefix, "bin", caddyBinaryName))

	if err := writeDefaultCaddyfile(opts.Prefix); err != nil {
		return err
	}
	logCaddySuccess("Workspace Caddyfile ready")

	return nil
}

/**
 * caddyArchSuffix maps the detected architecture to a Caddy release suffix.
 *
 * @param arch Normalized architecture string from system detection.
 * @return Caddy release suffix or an error when unsupported.
 */
func caddyArchSuffix(arch string) (string, error) {
	switch arch {
	case "x86_64", "amd64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	default:
		if strings.HasPrefix(arch, "armv7") {
			return "", fmt.Errorf("caddy installer does not support architecture %s yet", arch)
		}
		return "", fmt.Errorf("unsupported architecture for caddy tarball: %s", arch)
	}
}

/**
 * extractBinary pulls the Caddy binary out of the tar archive into dest.
 *
 * @param tarballPath Path to the downloaded tarball file.
 * @param dest        Destination path for the extracted binary.
 * @return error when extraction fails or the binary cannot be found.
 */
func extractBinary(tarballPath, dest string) error {
	file, err := os.Open(tarballPath)
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
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		base := filepath.Base(header.Name)
		if base != caddyBinaryName {
			continue
		}

		tmp := dest + ".tmp"
		out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
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

		return os.Rename(tmp, dest)
	}

	return fmt.Errorf("no %s binary found in tarball", caddyBinaryName)
}

/**
 * writeDefaultCaddyfile seeds the workspace with a minimal Caddyfile when absent.
 *
 * @param prefix Workspace root for the Chauffeur installation.
 * @return error when the Caddyfile cannot be written.
 */
func writeDefaultCaddyfile(prefix string) error {
	dest := filepath.Join(prefix, "caddy", "Caddyfile")
	if _, err := os.Stat(dest); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat Caddyfile: %w", err)
	}

	content := `{
	auto_https off
}
# Project sites are appended by chauf link
`

	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write default Caddyfile: %w", err)
	}
	return nil
}

/**
 * selectCaddyAsset locates the best matching Caddy artifact for the requested architecture.
 *
 * @param release GitHub release metadata.
 * @param version Resolved Caddy version string.
 * @param arch    Normalized architecture string.
 * @return Artifact name, download URL, and an error if not found.
 */
func selectCaddyAsset(release releases.GitHubRelease, version, arch string) (string, string, error) {
	archSuffix, err := caddyArchSuffix(arch)
	if err != nil {
		return "", "", err
	}

	archAsset := fmt.Sprintf("caddy_%s_linux_%s.tar.gz", version, archSuffix)
	if url, ok := release.AssetURL(archAsset); ok {
		fmt.Printf("Selected Caddy asset: %s\n", archAsset)
		return archAsset, url, nil
	}

	// Fallback to buildable artifact if architecture-specific tarball is missing.
	buildable := fmt.Sprintf("caddy_%s_buildable-artifact.tar.gz", version)
	if url, ok := release.AssetURL(buildable); ok {
		fmt.Printf("Selected Caddy asset: %s\n", buildable)
		return buildable, url, nil
	}

	// Final fallback: construct direct download URL for the arch asset.
	fallbackURL := fmt.Sprintf("https://github.com/caddyserver/caddy/releases/download/%s/%s", release.TagName, archAsset)
	fmt.Printf("Falling back to constructed URL: %s\n", fallbackURL)
	return archAsset, fallbackURL, nil
}

/**
 * locateCaddyChecksum finds a checksum resource for the chosen Caddy asset.
 *
 * @param release   GitHub release metadata.
 * @param assetName Selected artifact name.
 * @param version   Resolved Caddy version string.
 * @return URL to checksum resource, whether it's a list, and error if missing.
 */
func locateCaddyChecksum(release releases.GitHubRelease, assetName, version string) (url string, fromList bool, err error) {
	directCandidates := []string{
		assetName + ".sha256",
		assetName + ".sha512",
	}
	for _, candidate := range directCandidates {
		if url, ok := release.AssetURL(candidate); ok {
			return url, false, nil
		}
	}

	listCandidates := []string{
		fmt.Sprintf("caddy_%s_checksums.txt", version),
		fmt.Sprintf("caddy_%s_sha256sums.txt", version),
	}
	for _, candidate := range listCandidates {
		if url, ok := release.AssetURL(candidate); ok {
			return url, true, nil
		}
	}

	return "", false, fmt.Errorf("no checksum asset found for Caddy %s", version)
}

/**
 * fetchCaddyChecksum retrieves the checksum value either from a list or direct file.
 *
 * @param client    HTTP client for downloading.
 * @param url       Checksum resource location.
 * @param assetName Artifact file name to match.
 * @param fromList  Whether the resource is a manifest of multiple checksums.
 * @return Matching checksum string or error.
 */
func fetchCaddyChecksum(client *http.Client, url, assetName string, fromList bool) (string, error) {
	if fromList {
		return checksumFromList(client, url, assetName)
	}

	content, err := downloadText(client, url)
	if err != nil {
		return "", err
	}
	return checksumFromContent(content, assetName)
}

func startCaddyLogSection(title string) {
	fmt.Printf("\n[ CADDY ] %s\n", strings.ToUpper(title))
}

func logCaddyInfo(format string, args ...interface{}) {
	fmt.Printf("    - %s\n", fmt.Sprintf(format, args...))
}

func logCaddySuccess(format string, args ...interface{}) {
	fmt.Printf("    [OK] %s\n", fmt.Sprintf(format, args...))
}
