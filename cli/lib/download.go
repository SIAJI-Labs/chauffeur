package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/logging"
)

// DownloadToFile streams a remote file down to dest and returns the byte count.
// When label is provided, a progress bar is rendered to stdout.
//
// @param client HTTP client used for the request.
// @param url    Remote resource to download.
// @param dest   Local path to persist the file.
// @param label  Optional label for progress rendering; leave empty to disable.
// @return Number of bytes written and an error, if any.
func DownloadToFile(client *http.Client, url, dest, label string) (int64, error) {
	return DownloadToFileWithLogger(client, url, dest, label, nil)
}

// DownloadToFileWithLogger streams a remote file down to dest and returns the byte count.
// When label is provided, a progress bar is rendered to stdout.
//
// @param client HTTP client used for the request.
// @param url    Remote resource to download.
// @param dest   Local path to persist the file.
// @param label  Optional label for progress rendering; leave empty to disable.
// @param logger Optional logger for progress bar prefix; if nil, uses default "install" logger.
// @return Number of bytes written and an error, if any.
func DownloadToFileWithLogger(client *http.Client, url, dest, label string, logger *logging.CommandLogger) (int64, error) {
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
		if logger != nil {
			progress = NewProgressPrinterWithLogger(label, resp.ContentLength, logger)
		} else {
			progress = NewProgressPrinter(label, resp.ContentLength)
		}
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

// DownloadText fetches the resource at url and returns it as trimmed text.
//
// @param client HTTP client used for the request.
// @param url    Remote resource to download.
// @return Trimmed textual contents of the response body.
func DownloadText(client *http.Client, url string) (string, error) {
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
