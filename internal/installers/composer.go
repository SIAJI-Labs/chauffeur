package installers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siegg/chauffeur/internal/workspace"
)

const (
	composerFallbackVersion = "2.8.4"
	composerDownloadBase    = "https://getcomposer.org/download"
)

// ComposerInstaller manages downloading and installing Composer PHAR.
type ComposerInstaller struct {
	opts BuildOpts
	root string
}

// NewComposerInstaller creates a ComposerInstaller using the current workspace root.
func NewComposerInstaller(opts BuildOpts) *ComposerInstaller {
	return &ComposerInstaller{opts: opts, root: workspace.Root()}
}

// ResolveVersion fetches the latest Composer release version from GitHub.
// Falls back to composerFallbackVersion on any error.
func (c *ComposerInstaller) ResolveVersion() string {
	resp, err := http.Get("https://api.github.com/repos/composer/composer/releases/latest")
	if err != nil {
		return composerFallbackVersion
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.Contains(line, `"tag_name"`) {
			continue
		}
		// "tag_name": "2.8.4"
		parts := strings.SplitN(line, `"tag_name"`, 2)
		if len(parts) < 2 {
			continue
		}
		rest := parts[1]
		start := strings.IndexByte(rest, '"')
		if start < 0 {
			continue
		}
		rest = rest[start+1:]
		end := strings.IndexByte(rest, '"')
		if end < 0 {
			continue
		}
		v := strings.TrimPrefix(rest[:end], "v")
		if v != "" {
			return v
		}
	}
	return composerFallbackVersion
}

// PharPath returns the path where composer.phar is stored.
func (c *ComposerInstaller) PharPath() string {
	return filepath.Join(c.root, "composer", "composer.phar")
}

// IsInstalled reports whether composer.phar is present.
func (c *ComposerInstaller) IsInstalled() bool {
	_, err := os.Stat(c.PharPath())
	return err == nil
}

// Download fetches composer.phar for the given version and marks it executable.
// Returns (cached, err). cached=true means the file was already present.
func (c *ComposerInstaller) Download(version string) (bool, error) {
	pharPath := c.PharPath()

	if !c.opts.Force {
		if _, err := os.Stat(pharPath); err == nil {
			return true, nil
		}
	}

	url := fmt.Sprintf("%s/%s/composer.phar", composerDownloadBase, version)
	cached, err := DownloadWithProgress(url, pharPath, c.opts.NoCache)
	if err != nil {
		return false, err
	}
	_ = os.Chmod(pharPath, 0755)
	return cached, nil
}

// InstalledVersion returns the installed Composer version, or "" if not present.
func (c *ComposerInstaller) InstalledVersion() string {
	pharPath := c.PharPath()
	if _, err := os.Stat(pharPath); err != nil {
		return ""
	}

	// Prefer running via the workspace PHP shim.
	phpBin := filepath.Join(c.root, "bin", "shims", "php")
	if _, err := os.Stat(phpBin); err != nil {
		phpBin = "php"
	}

	cmd := exec.Command(phpBin, pharPath, "--version", "--no-ansi")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// "Composer version 2.8.4 2024-12-11 ..."
	parts := strings.Fields(strings.TrimSpace(string(out)))
	for i, p := range parts {
		if p == "version" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
