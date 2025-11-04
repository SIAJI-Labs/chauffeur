package installers

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/logging"
)

/**
 * writeShim ensures the shim executable points to the provided target binary.
 *
 * @param prefix Workspace root where shims live.
 * @param name   Shim filename to create or update.
 * @param target Absolute path to the managed binary.
 * @return error when the shim cannot be written.
 */
func writeShim(prefix, name, target string) error {
	binDir := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("ensure bin dir: %w", err)
	}

	scriptPath := filepath.Join(binDir, name)
	if err := os.WriteFile(scriptPath, []byte(shimContent(name, target)), 0o755); err != nil {
		return fmt.Errorf("write shim %s: %w", scriptPath, err)
	}
	return nil
}

/**
 * shimContent constructs the shell script content for a shim.
 *
 * @param name   Shim executable name.
 * @param target Absolute path to the real binary.
 * @return Shell script body that delegates to target.
 */
func shimContent(name, target string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
TARGET="%s"
if [[ ! -x "$TARGET" ]]; then
  echo "%s binary is missing at $TARGET" >&2
  exit 1
fi
exec "$TARGET" "$@"
`, target, name)
}

/**
 * downloadToFile streams a remote file down to dest and returns the byte count.
 * When label is provided, a simple progress bar is rendered to stdout.
 *
 * @param client HTTP client used for the request.
 * @param url    Remote resource to download.
 * @param dest   Local path to persist the file.
 * @param label  Optional label for progress rendering; leave empty to disable.
 * @return Number of bytes written and an error, if any.
 */
func downloadToFile(client *http.Client, url, dest, label string) (int64, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("unexpected status %s from %s", resp.Status, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	var writer io.Writer = out
	var progress *progressPrinter
	if label != "" {
		progress = newProgressPrinter(label, resp.ContentLength)
		defer progress.Finish()
		writer = io.MultiWriter(out, progress)
	}

	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return written, err
	}

	if err := out.Sync(); err != nil {
		return written, err
	}

	return written, nil
}

/**
 * downloadText fetches the resource at url and returns it as trimmed text.
 *
 * @param client HTTP client used for the request.
 * @param url    Remote resource to download.
 * @return Trimmed textual contents of the response body.
 */
func downloadText(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %s from %s", resp.Status, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

/**
 * checksumFromList downloads a checksum list and extracts the value for assetName.
 *
 * @param client HTTP client used to fetch the list.
 * @param url    Location of the checksum manifest.
 * @param assetName File name whose checksum is required.
 * @return Matching checksum string for the asset.
 */
func checksumFromList(client *http.Client, url, assetName string) (string, error) {
	content, err := downloadText(client, url)
	if err != nil {
		return "", err
	}
	return checksumFromContent(content, assetName)
}

/**
 * checksumFromContent parses checksum text and returns the hash for assetName.
 *
 * @param content  Raw checksum list contents.
 * @param assetName File name to match against the list.
 * @return Checksum string matching assetName.
 */
func checksumFromContent(content, assetName string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		fieldCount := len(fields)

		var sum, name string
		if fieldCount == 1 {
			sum = fields[0]
			name = ""
		} else {
			sum = fields[0]
			name = strings.TrimPrefix(fields[len(fields)-1], "*")
		}

		if assetName == "" || name == assetName || fieldCount == 1 {
			return sum, nil
		}
	}

	if assetName == "" {
		return "", fmt.Errorf("checksum not found in provided content")
	}
	return "", fmt.Errorf("checksum for %s not found in provided content", assetName)
}

/**
 * validateChecksum reads the file at path and verifies the digest matches expected.
 *
 * @param path     Local file path to hash.
 * @param expected Expected checksum string (sha256 or sha512).
 * @return error when the calculated digest differs from expected.
 */
func validateChecksum(path, expected string) error {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fmt.Errorf("empty checksum provided")
	}

	// Split away filenames or prefixes like "sha256:"
	if strings.Contains(expected, " ") {
		expected = strings.Fields(expected)[0]
	}
	if strings.Contains(expected, ":") {
		parts := strings.Split(expected, ":")
		expected = parts[len(parts)-1]
	}

	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fmt.Errorf("empty checksum after normalization")
	}

	var (
		hasher hash.Hash
		err    error
	)

	switch len(expected) {
	case 64:
		hasher = sha256.New()
	case 128:
		hasher = sha512.New()
	default:
		return fmt.Errorf("unsupported checksum length %d", len(expected))
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: expected %s got %s", expected, actual)
	}

	return nil
}

func fileSHA256(path string) (string, error) {
	return computeFileHash(path, sha256.New())
}

func fileSHA512(path string) (string, error) {
	return computeFileHash(path, sha512.New())
}

func computeFileHash(path string, hasher hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type progressPrinter struct {
	logger     *logging.CommandLogger
	label      string
	total      int64
	current    int64
	width      int
	lastPrint  time.Time
	done       bool
}

func newProgressPrinter(label string, total int64) *progressPrinter {
	logger := logging.NewCommandLogger("install")
	return &progressPrinter{
		logger:    logger,
		label:     label,
		total:     total,
		width:     32,
		lastPrint: time.Time{},
	}
}

func (p *progressPrinter) Write(b []byte) (int, error) {
	n := len(b)
	p.current += int64(n)
	if time.Since(p.lastPrint) >= 120*time.Millisecond || p.current == p.total {
		p.render()
	}
	return n, nil
}

func (p *progressPrinter) Finish() {
	if p == nil || p.done {
		return
	}
	p.renderFinal()
	p.done = true
}

func (p *progressPrinter) render() {
	if p.total > 0 {
		ratio := float64(p.current) / float64(p.total)
		if ratio > 1 {
			ratio = 1
		}
		filled := int(ratio * float64(p.width))
		if filled > p.width {
			filled = p.width
		}
		bar := strings.Repeat("#", filled) + strings.Repeat(".", p.width-filled)
		// Use command logger prefix format and clear line
		fmt.Printf("\r\033[K%s %s... [%s] %3.0f%%", p.logger.Prefix(), p.label, bar, ratio*100)
	} else {
		fmt.Printf("\r\033[K%s %s... %s", p.logger.Prefix(), p.label, humanBytes(p.current))
	}
	// Ensure the output is flushed immediately
	fmt.Print("")
	p.lastPrint = time.Now()
}

func (p *progressPrinter) renderFinal() {
	if p.total > 0 {
		fmt.Printf("\r\033[K%s %s... [%s] 100%% (%s)\n", p.logger.Prefix(), p.label, strings.Repeat("#", p.width), humanBytes(p.current))
	} else {
		fmt.Printf("\r\033[K%s %s... %s\n", p.logger.Prefix(), p.label, humanBytes(p.current))
	}
}

func humanBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	value := float64(n)
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", n, units[idx])
	}
	return fmt.Sprintf("%.1f %s", value, units[idx])
}
