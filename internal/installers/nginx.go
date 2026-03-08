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
	nginxFallbackVersion = "1.27.3"
	nginxDownloadBase    = "https://nginx.org/download"
)

// NginxInstaller manages downloading and building nginx from source.
type NginxInstaller struct {
	opts BuildOpts
	root string
}

// NewNginxInstaller creates a NginxInstaller using the current workspace root.
func NewNginxInstaller(opts BuildOpts) *NginxInstaller {
	return &NginxInstaller{opts: opts, root: workspace.Root()}
}

// ResolveVersion fetches the latest stable nginx version from the GitHub releases API.
// Falls back to nginxFallbackVersion on any error.
func (n *NginxInstaller) ResolveVersion() string {
	resp, err := http.Get("https://api.github.com/repos/nginx/nginx/releases/latest")
	if err != nil {
		return nginxFallbackVersion
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, `"tag_name"`) {
			start := strings.Index(line, "release-")
			if start < 0 {
				continue
			}
			tail := line[start+len("release-"):]
			end := strings.IndexByte(tail, '"')
			if end < 0 {
				continue
			}
			if v := tail[:end]; v != "" {
				return v
			}
		}
	}
	return nginxFallbackVersion
}

// Download fetches the nginx source tarball.
// version may be "" or "latest" to auto-resolve.
// Returns (tarPath, resolvedVersion, cached, err).
func (n *NginxInstaller) Download(version string) (string, string, bool, error) {
	if version == "" || version == "latest" {
		version = n.ResolveVersion()
	}

	filename := fmt.Sprintf("nginx-%s.tar.gz", version)
	cacheDir := filepath.Join(n.root, "cache", "nginx")
	tarPath := filepath.Join(cacheDir, filename)

	cached, err := DownloadWithProgress(
		fmt.Sprintf("%s/%s", nginxDownloadBase, filename),
		tarPath,
		n.opts.NoCache,
	)
	return tarPath, version, cached, err
}

// Build compiles and installs nginx from the source tarball.
// Skips if already installed and Force is false.
func (n *NginxInstaller) Build(tarPath, version string) error {
	if !n.opts.Force {
		if _, err := os.Stat(n.BinPath()); err == nil {
			return nil
		}
	}

	tmpDir, err := os.MkdirTemp("", "chauf-nginx-*")
	if err != nil {
		return err
	}
	if os.Getenv("CHAUFFEUR_KEEP_BUILD_DIR") != "1" {
		defer os.RemoveAll(tmpDir)
	}

	srcDir, err := ExtractTarGz(tarPath, tmpDir, 0)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	root := n.root
	configureArgs := []string{
		"--prefix=" + filepath.Join(root, "nginx"),
		"--sbin-path=" + n.BinPath(),
		"--conf-path=" + filepath.Join(root, "nginx", "etc", "nginx.conf"),
		"--pid-path=" + filepath.Join(root, "nginx", "logs", "nginx.pid"),
		"--http-log-path=" + filepath.Join(root, "nginx", "logs", "access.log"),
		"--error-log-path=" + filepath.Join(root, "nginx", "logs", "error.log"),
		"--with-pcre-jit",
		"--with-http_gzip_static_module",
		"--with-http_ssl_module",
	}

	if err := RunCmd(srcDir, n.opts.Verbose, "./configure", configureArgs...); err != nil {
		return fmt.Errorf("configure: %w", err)
	}
	if err := RunCmd(srcDir, n.opts.Verbose, "make", "-j"+NumCPUs()); err != nil {
		return fmt.Errorf("make: %w", err)
	}
	if err := RunCmd(srcDir, n.opts.Verbose, "make", "install"); err != nil {
		return fmt.Errorf("make install: %w", err)
	}
	return nil
}

// BinPath returns the expected nginx binary path.
func (n *NginxInstaller) BinPath() string {
	return filepath.Join(n.root, "nginx", "sbin", "nginx")
}

// IsInstalled reports whether nginx is already built and installed.
func (n *NginxInstaller) IsInstalled() bool {
	_, err := os.Stat(n.BinPath())
	return err == nil
}

// InstalledVersion returns the installed nginx version ("1.27.3"), or "" if not present.
func (n *NginxInstaller) InstalledVersion() string {
	cmd := exec.Command(n.BinPath(), "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Output: "nginx version: nginx/1.27.3"
	s := strings.TrimSpace(string(out))
	idx := strings.LastIndexByte(s, '/')
	if idx < 0 {
		return ""
	}
	return s[idx+1:]
}
