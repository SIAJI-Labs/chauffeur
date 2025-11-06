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
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/releases"
	"github.com/siaji/chauffeur/cli/internal/system"
	"github.com/siaji/chauffeur/cli/lib"
)

const (
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
)

// Helper function for colored output
func colorize(color, text string) string {
	return color + text + colorReset
}

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
	caddyLogger := lib.NewCommandLogger("caddy")
	
	if opts.Prefix == "" {
		return caddyLogger.Fail("install prefix is required", "")
	}

	// Check for DNS resolution dependencies
	if err := checkDNSResolution(caddyLogger); err != nil {
		return err
	}

	// Check dnsmasq configuration for local domains
	if err := checkDnsmasqConfiguration(caddyLogger); err != nil {
		return err
	}

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	caddyLogger.Info("Preparing for caddy installation")
	caddyLogger.Info(fmt.Sprintf("Detected architecture: %s", opts.Info.Arch))
	caddyLogger.Info("Fetching release metadata from GitHub…")
	release, err := releases.LatestGitHubRelease(client, "caddyserver", "caddy")
	if err != nil {
		return caddyLogger.Fail("resolve latest Caddy release", err.Error())
	}
	versionTag := release.TagName
	if versionTag == "" {
		return caddyLogger.Fail("latest Caddy release has empty tag name", "")
	}
	version := strings.TrimPrefix(versionTag, "v")
	caddyLogger.Success("Latest release identified", fmt.Sprintf("%s (tag %s)", version, versionTag))

	assetName, tarballURL, err := selectCaddyAsset(release, version, opts.Info.Arch)
	if err != nil {
		return err
	}

	// Create download section
	caddyLogger.Info("Download")
	downloadLogger := caddyLogger.NewChildLogger("download")
	downloadLogger.Info(fmt.Sprintf("Resolved asset: %s", assetName))
	downloadLogger.Info(fmt.Sprintf("Source URL: %s", tarballURL))

	checksumURL, checksumIsList, err := locateCaddyChecksum(release, assetName, version)
	if err != nil {
		return err
	}
	downloadLogger.Info(fmt.Sprintf("Checksum source: %s", checksumURL))

	destBinDir := filepath.Join(opts.Prefix, "caddy", "bin")
	if err := os.MkdirAll(destBinDir, 0o755); err != nil {
		return caddyLogger.Fail("ensure caddy bin dir", err.Error())
	}

	tmpDir, err := os.MkdirTemp("", "chauffeur-caddy-*")
	if err != nil {
		return caddyLogger.Fail("create temp dir", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, assetName)
	size, err := lib.DownloadToFileWithLogger(client, tarballURL, tarballPath, fmt.Sprintf("Download %s", assetName), downloadLogger)
	if err != nil {
		return caddyLogger.Fail("download caddy tarball", err.Error())
	}
	downloadLogger.Success(fmt.Sprintf("Downloaded %s", assetName), fmt.Sprintf("%d bytes", size))

	// Verification section
	caddyLogger.Info("Verifying")
	verifyLogger := caddyLogger.NewChildLogger("verify")
	verifyLogger.Info("Verifying checksum…")
	expectedSum, err := fetchCaddyChecksum(client, checksumURL, assetName, checksumIsList)
	if err != nil {
		return caddyLogger.Fail("resolve checksum", err.Error())
	}
	if err := validateChecksum(tarballPath, expectedSum); err != nil {
		return caddyLogger.Fail("validate caddy tarball", err.Error())
	}
	verifyLogger.Success("Checksum verification passed", "")

	targetBinary := filepath.Join(destBinDir, caddyBinaryName)
	if !opts.Force {
		if info, err := os.Stat(targetBinary); err == nil && info.Mode().IsRegular() {
			downloadLogger.Info("Existing binary detected; skipping extraction (use --force to overwrite)")
			return nil
		}
	}

	// Installation section
	caddyLogger.Info("Installing")
	installLogger := caddyLogger.NewChildLogger("install")
	if err := extractBinary(tarballPath, targetBinary); err != nil {
		return caddyLogger.Fail("extract caddy binary", err.Error())
	}
	installLogger.Success("Installed binary", targetBinary)

	if err := writeShim(opts.Prefix, caddyBinaryName, targetBinary); err != nil {
		return caddyLogger.Fail("write shim", err.Error())
	}
	installLogger.Success("Updated shim", filepath.Join(opts.Prefix, "bin", caddyBinaryName))

	if err := writeDefaultCaddyfile(opts.Prefix); err != nil {
		return caddyLogger.Fail("write default caddyfile", err.Error())
	}
	installLogger.Success("Workspace Caddyfile ready", "")

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

	// Read global config to get correct port settings
	configPath := filepath.Join(prefix, "config", "chauffeur.yaml")
	httpPort := "8080"
	httpsPort := "8443"
	
	if configData, err := os.ReadFile(configPath); err == nil {
		configContent := string(configData)
		// Simple parsing for port settings - in production you'd use a proper YAML parser
		lines := strings.Split(configContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http_port:") {
				httpPort = strings.TrimSpace(strings.TrimPrefix(line, "http_port:"))
			}
			if strings.HasPrefix(line, "https_port:") {
				httpsPort = strings.TrimSpace(strings.TrimPrefix(line, "https_port:"))
			}
		}
	}

	content := fmt.Sprintf(`{
	auto_https off
	http_port %s
	https_port %s
}
# Project sites are appended by chauf link
`, httpPort, httpsPort)

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
		return lib.ChecksumFromList(client, url, assetName)
	}

	content, err := lib.DownloadText(client, url)
	if err != nil {
		return "", err
	}
	return lib.ChecksumFromContent(content, assetName)
}

/**
 * checkDnsmasqConfiguration validates if dnsmasq is configured for local domain resolution.
 *
 * @param logger Command logger for status reporting.
 * @return error if dnsmasq is not configured and user declines to configure it.
 */
func checkDnsmasqConfiguration(logger *lib.Logger) error {
	logger.Info("Checking dnsmasq configuration for local domain resolution...")
	
	dnsLogger := logger.NewChildLogger("dns")
	
	// Check if dnsmasq is available (NetworkManager or standalone)
	if !system.IsDnsmasqAvailable() {
		return dnsLogger.Fail("dnsmasq not available", " Install dnsmasq first")
	}
	
	// Check if chauffeur.conf exists in either location
	configPaths := []string{
		"/etc/dnsmasq.d/chauffeur.conf",
		"/etc/NetworkManager/dnsmasq.d/chauffeur.conf",
	}
	
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			if strings.Contains(configPath, "NetworkManager") {
				dnsLogger.Success("dnsmasq configuration found", configPath+" (NetworkManager)")
			} else {
				dnsLogger.Success("dnsmasq configuration found", configPath+" (standalone)")
			}
			return nil
		}
	}
	
	dnsLogger.Warn("dnsmasq configuration not found", "Local .test domains won't resolve")
	dnsLogger.Info("Chauffeur requires dnsmasq configuration to resolve .test domains")
	dnsLogger.Info("Add this configuration to make .test domains resolve to localhost:")
	
	fmt.Printf("\n%s\n", colorize(colorYellow, "Required dnsmasq configuration:"))
	fmt.Printf("sudo install -d -m 755 /etc/dnsmasq.d\n")
	fmt.Printf("sudo tee /etc/dnsmasq.d/chauffeur.conf >/dev/null <<'EOF'\n")
	fmt.Printf("# Chauffeur local development resolver\n")
	fmt.Printf("# Redirect all *.test domains to localhost\n")
	fmt.Printf("address=/.test/127.0.0.1\n")
	fmt.Printf("# Only listen locally\n")
	fmt.Printf("listen-address=127.0.0.1\n")
	fmt.Printf("bind-interfaces\n")
	fmt.Printf("EOF\n")
	
	fmt.Printf("\n%s", colorize(colorYellow, "Do you want to add this configuration now? [y/N]: "))
	var response string
	fmt.Scanln(&response)
	
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		return dnsLogger.Fail("configuration declined", "Local .test domains will not work without this configuration")
	}
	
	// Use the new unified setup function
	if err := system.SetupLocalDNSResolution(); err != nil {
		if system.IsNetworkManagerDnsmasqRunning() {
			// Don't fail hard for NetworkManager conflicts
			dnsLogger.Warn("dnsmasq setup completed with warnings", "NetworkManager is managing DNS resolution")
			dnsLogger.Info("Local .test domains should now resolve to localhost (Configuration updated via NetworkManager)")
		} else {
			return dnsLogger.Fail("setup dnsmasq configuration", err.Error())
		}
	} else {
		dnsLogger.Success("dnsmasq configuration completed", "Local .test domains should now resolve to localhost")
	}
	
	return nil
}



/**
 * checkDNSResolution checks for available DNS resolution packages and provides guidance.
 *
 * @param logger Command logger for status reporting.
 * @return error for critical issues, but continues installation if packages are missing.
 */
func checkDNSResolution(logger *lib.Logger) error {
	logger.Info("Checking DNS resolution dependencies...")
	
	dnsLogger := logger.NewChildLogger("dns")
	
	// Check if required commands are available
	hasDnsmasq := system.IsDnsmasqAvailable()
	hasResolvectl := system.IsCommandAvailable("resolvectl")
	pm := system.DetectPackageManager()
	
	if hasDnsmasq && hasResolvectl {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Success("DNS resolution dependencies are satisfied", "NetworkManager dnsmasq and resolvectl available")
		} else {
			dnsLogger.Success("DNS resolution dependencies are satisfied", "standalone dnsmasq and resolvectl available")
		}
		return nil
	}
	
	if !hasResolvectl {
		dnsLogger.Warn("resolvectl not found", "systemd-resolved may not be available")
	}
	
	if !hasDnsmasq {
		if system.IsNetworkManagerDnsmasqRunning() {
			dnsLogger.Warn("NetworkManager dnsmasq not running", "local .test domains may not resolve")
		} else {
			dnsLogger.Warn("dnsmasq not available", "local .test domains may not resolve")
		}
		
		// Get missing packages list
		missing := system.GetMissingPackages()
		
		if len(missing) > 0 {
			// Handle Arch Linux specifically
			if pm == system.Pacman {
				archPm := system.DetectArchPackageManager()
				if archPm == "" {
					return dnsLogger.Fail("no arch package manager found", "Neither pacman, yay, nor paru found")
				}
				
				for _, pkg := range missing {
					dnsLogger.Info(fmt.Sprintf("Package %s (%s) is required for local domain resolution", pkg.Name, pkg.PackageName))
					dnsLogger.Info(fmt.Sprintf("Description: %s", pkg.Description))
					
					var response string
					if archPm == "pacman" {
						dnsLogger.Info(fmt.Sprintf("Do you want to install it via %s? [y/N]:", archPm))
					} else {
						dnsLogger.Info(fmt.Sprintf("Do you want to install it via %s? [y/N]:", archPm))
					}
					
					fmt.Scanln(&response)
					response = strings.ToLower(strings.TrimSpace(response))
					
					if response == "y" || response == "yes" {
						// Install the package
						dnsLogger.Info(fmt.Sprintf("Installing %s via %s...", pkg.PackageName, archPm))
						
						var installCmd string
						if archPm == "pacman" {
							installCmd = fmt.Sprintf("sudo pacman -S --noconfirm %s", pkg.PackageName)
						} else {
							// yay/paru - these AUR helpers handle sudo internally
							installCmd = fmt.Sprintf("%s -S --noconfirm %s", archPm, pkg.PackageName)
						}
						
						cmd := exec.Command("sh", "-c", installCmd)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						
						if err := cmd.Run(); err != nil {
							return dnsLogger.Fail(fmt.Sprintf("install %s via %s", pkg.PackageName, archPm), err.Error())
						}
						
						dnsLogger.Success(fmt.Sprintf("Installed %s", pkg.PackageName), "")
					} else {
						return dnsLogger.Fail("package installation declined", fmt.Sprintf("Caddy installation cancelled - %s is required for local domain resolution", pkg.Name))
					}
				}
			} else {
				// Non-Arch distributions
				dnsLogger.Info("For proper local domain resolution, you may need:")
				for _, pkg := range missing {
					dnsLogger.Info(fmt.Sprintf("  - %s (%s): %s", pkg.Name, pkg.PackageName, pkg.Description))
				}
				
				if pm == system.Unknown {
					dnsLogger.Info("Work in progress: Automatic installation is not supported for this distribution")
					dnsLogger.Info("Work in progress: Please install dnsmasq manually and try again in the next version")
				} else {
					dnsLogger.Info(fmt.Sprintf("Work in progress: Automatic installation is not yet supported for %s", pm))
					dnsLogger.Info("Work in progress: Please install dnsmasq manually and try again in the next version")
				}
				
				// For non-Arch, continue with installation but provide guidance
				dnsLogger.Info("Current Caddy installation will continue for basic functionality")
			}
		}
	} else {
		dnsLogger.Success("DNS resolution partially available", "resolvectl found")
	}
	
	return nil
}

/**
 * getInstallationCommand returns the appropriate installation command for a package.
 *
 * @param pm Package manager to use
 * @param pkg Package to install
 * @return Command string for manual installation
 */
func getInstallationCommand(pm system.PackageManager, pkg system.Package) string {
	switch pm {
	case system.Pacman:
		return fmt.Sprintf("sudo pacman -S %s", pkg.PackageName)
	case system.Apt:
		return fmt.Sprintf("sudo apt install %s", pkg.PackageName)
	case system.Yum:
		return fmt.Sprintf("sudo yum install %s", pkg.PackageName)
	case system.Dnf:
		return fmt.Sprintf("sudo dnf install %s", pkg.PackageName)
	case system.Zypper:
		return fmt.Sprintf("sudo zypper install %s", pkg.PackageName)
	default:
		return fmt.Sprintf("# Install %s using your system package manager", pkg.PackageName)
	}
}




