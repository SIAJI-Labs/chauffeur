package installers

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/siegg/chauffeur/internal/lib"
)

// BuildOpts controls installer behavior.
type BuildOpts struct {
	Force   bool // reinstall even if already present
	NoCache bool // skip download cache
	Verbose bool // stream build output to terminal
}

// DownloadWithProgress downloads url to destPath, rendering a progress bar.
// If NoCache is false and destPath already exists, it returns (true, nil) immediately.
func DownloadWithProgress(url, destPath string, noCache bool) (cached bool, err error) {
	if !noCache {
		if _, err := os.Stat(destPath); err == nil {
			return true, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return false, err
	}

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return false, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	tmp := destPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return false, err
	}

	pr := lib.NewProgressReader(resp.Body, resp.ContentLength, filepath.Base(destPath))
	_, copyErr := io.Copy(f, pr)
	pr.Done()
	f.Close()

	if copyErr != nil {
		os.Remove(tmp)
		return false, fmt.Errorf("download %s: %w", url, copyErr)
	}

	if err := os.Rename(tmp, destPath); err != nil {
		os.Remove(tmp)
		return false, err
	}

	return false, nil
}

// VerifySHA256 checks that path's SHA-256 hex digest matches expected.
func VerifySHA256(path, expected string) error {
	return verifyHash(path, expected, sha256.New())
}

// VerifySHA512 checks that path's SHA-512 hex digest matches expected.
func VerifySHA512(path, expected string) error {
	return verifyHash(path, expected, sha512.New())
}

func verifyHash(path, expected string, h hash.Hash) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("checksum mismatch for %s:\n  expected: %s\n  got:      %s",
			filepath.Base(path), expected, got)
	}
	return nil
}

// FetchChecksum downloads checksumURL and extracts the hash for filename.
// Supports "HASH  filename" and "HASH *filename" (sha256sum/sha512sum) format.
func FetchChecksum(checksumURL, filename string) (string, error) {
	resp, err := http.Get(checksumURL) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, checksumURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if strings.TrimLeft(parts[1], "*") == filename {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum for %q not found", filename)
}

// ExtractTarGz extracts a .tar.gz into destDir, stripping strip leading path components.
// Returns the path to the single top-level directory extracted (when strip==0).
func ExtractTarGz(tarPath, destDir string, strip int) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	topDir := ""
	cleanDest := filepath.Clean(destDir) + string(filepath.Separator)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar: %w", err)
		}

		name := filepath.Clean(hdr.Name)
		parts := strings.SplitN(name, string(filepath.Separator), strip+2)

		// Only record topDir from real entries — PAX tarballs (like OpenSSL) emit
		// a pax_global_header as the first entry (TypeXGlobalHeader/TypeXHeader).
		if topDir == "" && strip == 0 && len(parts) >= 1 &&
			hdr.Typeflag != tar.TypeXGlobalHeader && hdr.Typeflag != tar.TypeXHeader {
			topDir = filepath.Join(destDir, parts[0])
		}

		if len(parts) <= strip {
			continue
		}
		rel := strings.Join(parts[strip:], string(filepath.Separator))
		if rel == "." {
			continue
		}

		target := filepath.Join(destDir, rel)
		if !strings.HasPrefix(target, cleanDest) {
			continue // path traversal guard
		}

		switch hdr.Typeflag {
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			continue
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0777)
			if err != nil {
				return "", err
			}
			_, copyErr := io.Copy(out, tr)
			out.Close()
			if copyErr != nil {
				return "", copyErr
			}
		case tar.TypeSymlink:
			_ = os.Remove(target)
			_ = os.Symlink(hdr.Linkname, target)
		}
	}

	return topDir, nil
}

// RunCmd runs name+args in workDir, streaming or capturing output based on verbose.
func RunCmd(workDir string, verbose bool, name string, args ...string) error {
	return RunCmdWithEnv(workDir, verbose, nil, name, args...)
}

// RunCmdWithEnv is RunCmd with extra environment variables appended to os.Environ().
func RunCmdWithEnv(workDir string, verbose bool, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = append(os.Environ(), env...)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// NumCPUs returns the logical CPU count as a string (for make -j<N>).
func NumCPUs() string {
	n := runtime.NumCPU()
	if n < 1 {
		n = 1
	}
	return fmt.Sprintf("%d", n)
}

// MajorMinor extracts "X.Y" from a version string like "8.3", "8.3.27", or "php8.3".
func MajorMinor(v string) string {
	v = strings.TrimPrefix(v, "php")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}
