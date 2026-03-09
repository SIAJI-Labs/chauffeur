package commands

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/lib"
)

const githubRepo = "siegg/chauffeur"

func RunSelfUpdate(args []string, currentVersion string) error {
	flags := flag.NewFlagSet("self-update", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	dev        := flags.Bool("dev", false, "Build from local source repository")
	dryRun     := flags.Bool("dry-run", false, "Show what would happen without replacing the binary")
	reqVersion := flags.String("version", "", "Install a specific release version (e.g. v2.1.0)")

	if err := flags.Parse(args); err != nil {
		return err
	}

	currentBin, err := resolveExecutable()
	if err != nil {
		return err
	}

	fmt.Println("Updating chauf...")
	fmt.Println()

	if *dev {
		return devUpdate(currentBin, *dryRun)
	}
	return releaseUpdate(currentBin, currentVersion, *reqVersion, *dryRun)
}

// ── dev mode: build from local source ─────────────────────────────────────────

func devUpdate(currentBin string, dryRun bool) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	version := gitOutput(repoRoot, "describe", "--tags", "--always", "--dirty") +
		"-" + time.Now().Format("20060102-1504")
	branch := gitOutput(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")

	lib.Pair("Repository", repoRoot)
	lib.Pair("Commit", fmt.Sprintf("%s  (%s)", version, branch))
	fmt.Println()

	if dryRun {
		lib.Info(lib.Gray("Would build and replace: " + currentBin))
		fmt.Println()
		lib.Info(lib.Gray("Run without --dry-run to apply."))
		fmt.Println()
		return nil
	}

	lib.Info("Building...")

	tmpBin := currentBin + ".tmp"

	ldflags := fmt.Sprintf(`-X main.version=%s`, version)
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", tmpBin, "./cmd/chauf")
	cmd.Dir = repoRoot

	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		os.Remove(tmpBin)
		return fmt.Errorf("go build failed:\n%s", strings.TrimSpace(errBuf.String()))
	}
	lib.Success("go build succeeded")

	if err := os.Rename(tmpBin, currentBin); err != nil {
		os.Remove(tmpBin)
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	lib.Success(fmt.Sprintf("Binary replaced  (%s)", currentBin))

	fmt.Println()
	lib.Success(fmt.Sprintf("Updated to %s", lib.Bold(version)))
	fmt.Println()

	return nil
}

// findRepoRoot walks up from CWD looking for the chauffeur go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(data), "github.com/siegg/chauffeur") {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(
		"chauffeur repository not found\n\n"+
			"  Run %s from inside the repository,\n"+
			"  or use %s to download a release.",
		lib.Bold("chauf self-update --dev"),
		lib.Bold("chauf self-update"),
	)
}

func gitOutput(dir string, gitArgs ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, gitArgs...)...).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// ── release mode: download from GitHub ────────────────────────────────────────

func releaseUpdate(currentBin, currentVersion, reqVersion string, dryRun bool) error {
	var targetVersion string

	if reqVersion != "" {
		targetVersion = reqVersion
	} else {
		lib.Info("Checking for updates...")
		v, err := fetchLatestVersion()
		if err != nil {
			return err
		}
		targetVersion = v
	}

	lib.Pair("Current", currentVersion)
	lib.Pair("Latest", targetVersion)
	fmt.Println()

	if currentVersion == targetVersion {
		lib.Success(fmt.Sprintf("Already on latest version (%s)", targetVersion))
		fmt.Println()
		return nil
	}

	plat         := currentPlatform()
	filename     := fmt.Sprintf("chauf_%s_%s.tar.gz", targetVersion, plat)
	baseURL      := fmt.Sprintf("https://github.com/%s/releases/download/%s", githubRepo, targetVersion)
	downloadURL  := baseURL + "/" + filename
	checksumURL  := fmt.Sprintf("%s/chauf_%s_checksums.txt", baseURL, targetVersion)

	if dryRun {
		lib.Info(fmt.Sprintf("Would download  %s  (%s)", filename, plat))
		lib.Info(fmt.Sprintf("Would replace   %s", currentBin))
		fmt.Println()
		lib.Info(lib.Gray("Run without --dry-run to apply."))
		fmt.Println()
		return nil
	}

	lib.Info(fmt.Sprintf("Downloading  %s  (%s)", lib.Bold("chauf "+targetVersion), plat))

	tmpDir, err := os.MkdirTemp("", "chauf-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, filename)
	if err := downloadFile(downloadURL, tarPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum if available (parses goreleaser's combined checksums.txt)
	if body, err := httpGetString(checksumURL); err == nil {
		if expected := extractChecksumForFile(body, filename); expected != "" {
			if err := verifyChecksum(tarPath, expected); err != nil {
				return err
			}
			lib.Success("Checksum verified")
		}
	}

	// Extract binary from tarball
	newBin := filepath.Join(tmpDir, "chauf")
	if err := extractBinary(tarPath, newBin); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Atomic replace
	tmpBin := currentBin + ".tmp"
	if err := copyFile(newBin, tmpBin); err != nil {
		return fmt.Errorf("failed to stage binary: %w", err)
	}
	if err := os.Chmod(tmpBin, 0755); err != nil {
		os.Remove(tmpBin)
		return err
	}
	if err := os.Rename(tmpBin, currentBin); err != nil {
		os.Remove(tmpBin)
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	lib.Success("Binary replaced")

	fmt.Println()
	lib.Success(fmt.Sprintf("Updated to %s", lib.Bold(targetVersion)))
	fmt.Println()

	return nil
}

func fetchLatestVersion() (string, error) {
	url  := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	body, err := httpGetString(url)
	if err != nil {
		return "", fmt.Errorf("could not reach GitHub: %w", err)
	}

	var release struct {
		Tag string `json:"tag_name"`
	}
	if err := json.Unmarshal([]byte(body), &release); err != nil || release.Tag == "" {
		return "", fmt.Errorf("no releases found — use --dev to build from source")
	}
	return release.Tag, nil
}

// extractChecksumForFile parses a goreleaser checksums.txt (format: "HASH  filename")
// and returns the hash for the given filename, or "" if not found.
func extractChecksumForFile(body, filename string) string {
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := strings.TrimPrefix(fields[1], "*")
			if name == filename {
				return fields[0]
			}
		}
	}
	return ""
}

// ── helpers ────────────────────────────────────────────────────────────────────

func resolveExecutable() (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine binary path: %w", err)
	}
	return filepath.EvalSymlinks(bin)
}

func currentPlatform() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

func httpGetString(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if got != expected {
		return fmt.Errorf("checksum mismatch\n  expected: %s\n  got:      %s", expected, got)
	}
	return nil
}

func extractBinary(tarPath, destBin string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) == "chauf" {
			out, err := os.Create(destBin)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, tr)
			return err
		}
	}
	return fmt.Errorf("chauf binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
