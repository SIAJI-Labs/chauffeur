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

// SupportedPHPVersions lists all supported major.minor versions, newest first.
var SupportedPHPVersions = []string{"8.4", "8.3", "8.2", "8.1", "8.0", "7.4"}

// phpFallbackVersions maps major.minor → known stable patch version.
var phpFallbackVersions = map[string]string{
	"8.4": "8.4.5",
	"8.3": "8.3.20",
	"8.2": "8.2.28",
	"8.1": "8.1.31",
	"8.0": "8.0.30",
	"7.4": "7.4.33",
}

const (
	phpDownloadBase     = "https://www.php.net/distributions"
	phpFallbackBase     = "https://secure.php.net/distributions"
	phpReleasesAPI      = "https://www.php.net/releases/active.php?json=1"
	imagickDefaultVer   = "3.8.0"
	imagickVersionURL   = "https://pecl.php.net/rest/r/imagick/latest.txt"
	imagickDownloadBase = "https://pecl.php.net/get"
)

// PHPInstaller manages compiling a specific PHP version from source.
type PHPInstaller struct {
	majorMinor string // e.g. "8.3"
	opts       BuildOpts
	root       string
}

// NewPHPInstaller creates a PHPInstaller for version (e.g. "8.3", "8.3.27").
func NewPHPInstaller(version string, opts BuildOpts) (*PHPInstaller, error) {
	mm := MajorMinor(version)
	if mm == "" {
		return nil, fmt.Errorf("unrecognized PHP version %q", version)
	}
	return &PHPInstaller{majorMinor: mm, opts: opts, root: workspace.Root()}, nil
}

// ResolveVersion returns the latest full version (e.g. "8.3.20") for this major.minor.
func (p *PHPInstaller) ResolveVersion() (string, error) {
	resp, err := http.Get(phpReleasesAPI)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if v := extractPHPVersion(string(body), p.majorMinor); v != "" {
			return v, nil
		}
	}

	if fallback, ok := phpFallbackVersions[p.majorMinor]; ok {
		return fallback, nil
	}
	return "", fmt.Errorf("no known version for PHP %s", p.majorMinor)
}

// extractPHPVersion parses the JSON from active.php to find the version for majorMinor.
func extractPHPVersion(body, mm string) string {
	needle := `"` + mm + `"`
	idx := strings.Index(body, needle)
	if idx < 0 {
		return ""
	}
	chunk := body[idx:]
	vIdx := strings.Index(chunk, `"version"`)
	if vIdx < 0 {
		return ""
	}
	chunk = chunk[vIdx+len(`"version"`):]
	start := strings.IndexByte(chunk, '"')
	if start < 0 {
		return ""
	}
	chunk = chunk[start+1:]
	end := strings.IndexByte(chunk, '"')
	if end < 0 {
		return ""
	}
	return chunk[:end]
}

// Download fetches the PHP source tarball for fullVersion (e.g. "8.3.20").
// Returns (tarPath, cached, err).
func (p *PHPInstaller) Download(fullVersion string) (string, bool, error) {
	filename := fmt.Sprintf("php-%s.tar.gz", fullVersion)
	cacheDir := filepath.Join(p.root, "cache", "php")
	tarPath := filepath.Join(cacheDir, filename)

	cached, err := DownloadWithProgress(
		fmt.Sprintf("%s/%s", phpDownloadBase, filename),
		tarPath,
		p.opts.NoCache,
	)
	if err != nil {
		// Try fallback mirror
		cached, err = DownloadWithProgress(
			fmt.Sprintf("%s/%s", phpFallbackBase, filename),
			tarPath,
			p.opts.NoCache,
		)
	}
	return tarPath, cached, err
}

// InstallDir returns the directory where this PHP version is installed.
func (p *PHPInstaller) InstallDir() string {
	return filepath.Join(p.root, "php", p.majorMinor)
}

// BinPath returns the expected php binary path.
func (p *PHPInstaller) BinPath() string {
	return filepath.Join(p.InstallDir(), "bin", "php")
}

// IsInstalled reports whether this PHP version is already compiled and installed.
func (p *PHPInstaller) IsInstalled() bool {
	_, err := os.Stat(p.BinPath())
	return err == nil
}

// MajorMinorStr returns the "X.Y" string for this installer.
func (p *PHPInstaller) MajorMinorStr() string {
	return p.majorMinor
}

// Build compiles PHP from the source tarball and installs it into the workspace.
// Skips if already installed and Force is false.
func (p *PHPInstaller) Build(tarPath, fullVersion string) error {
	if !p.opts.Force && p.IsInstalled() {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", "chauf-php-*")
	if err != nil {
		return err
	}
	if os.Getenv("CHAUFFEUR_KEEP_BUILD_DIR") != "1" {
		defer os.RemoveAll(tmpDir)
	}

	srcDir, err := ExtractTarGz(tarPath, tmpDir, 0)
	if err != nil {
		return fmt.Errorf("extract PHP: %w", err)
	}

	// ── Legacy PHP: vendor OpenSSL + apply source patches ───────────────────
	var buildEnv []string
	if IsLegacy(p.majorMinor) {
		vendorDir := filepath.Join(p.root, "cache", "vendor-openssl-"+p.majorMinor)
		// Build vendored OpenSSL if not already done.
		if _, err := os.Stat(filepath.Join(vendorDir, "lib", "pkgconfig")); err != nil {
			if err := BuildVendoredOpenSSL(filepath.Join(p.root, "cache"), vendorDir, p.opts); err != nil {
				return fmt.Errorf("vendored OpenSSL: %w", err)
			}
		}
		buildEnv = VendoredOpenSSLEnv(vendorDir)

		if err := ApplyLegacyPatches(srcDir); err != nil {
			return err
		}
	}

	// ── buildconf ────────────────────────────────────────────────────────────
	if err := RunCmd(srcDir, p.opts.Verbose, "./buildconf", "--force"); err != nil {
		return fmt.Errorf("buildconf: %w", err)
	}

	// ── configure ────────────────────────────────────────────────────────────
	if err := RunCmdWithEnv(srcDir, p.opts.Verbose, buildEnv, "./configure", p.configureArgs()...); err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	// ── make ─────────────────────────────────────────────────────────────────
	if err := RunCmdWithEnv(srcDir, p.opts.Verbose, buildEnv, "make", "-j"+NumCPUs()); err != nil {
		return fmt.Errorf("make: %w", err)
	}

	// ── make install ─────────────────────────────────────────────────────────
	if err := RunCmdWithEnv(srcDir, p.opts.Verbose, buildEnv, "make", "install"); err != nil {
		return fmt.Errorf("make install: %w", err)
	}

	// ── Post-install config ───────────────────────────────────────────────────
	if err := p.writePhpIni(); err != nil {
		return err
	}
	if err := p.WritePHPConfig(workspace.DefaultPHPVersionConfig()); err != nil {
		return err
	}
	if err := p.writeFPMConfig(); err != nil {
		return err
	}

	return nil
}

// configureArgs returns the ./configure arguments for this PHP version.
func (p *PHPInstaller) configureArgs() []string {
	root := p.root
	mm := p.majorMinor
	installDir := p.InstallDir()
	major := strings.SplitN(mm, ".", 2)[0]

	user := os.Getenv("USER")
	if user == "" {
		user = "www-data"
	}

	args := []string{
		"--prefix=" + installDir,
		"--enable-cli",
		"--enable-fpm",
		"--enable-mysqlnd",
		"--with-fpm-user=" + user,
		"--with-fpm-group=" + user,
		"--with-config-file-path=" + filepath.Join(root, "php", mm, "etc"),
		"--with-config-file-scan-dir=" + filepath.Join(root, "php", mm, "etc", "conf.d"),
		"--enable-mbstring",
		"--with-openssl",
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
		"--enable-exif",
		"--enable-posix",
		"--with-pear",
		"--with-readline",
		"--with-mysqli=mysqlnd",
		"--with-pdo-mysql=mysqlnd",
		"--enable-gmp",
		"--enable-bcmath",
	}

	// GD: 8.0 builds GD separately; 7.4 and 8.1+ use the unified flag.
	if mm == "8.0" {
		// GD for 8.0 is handled by BuildGD80() post-install.
	} else {
		args = append(args, "--enable-gd", "--with-jpeg=/usr", "--with-freetype=/usr")
	}

	// sodium: 7.4 and 8.0 need the explicit flag; 8.1+ includes it by default.
	if major == "7" || mm == "8.0" {
		args = append(args, "--with-sodium")
	}

	return args
}

// writePhpIni writes a minimal php.ini into the install etc dir.
func (p *PHPInstaller) writePhpIni() error {
	etcDir := filepath.Join(p.InstallDir(), "etc")
	if err := os.MkdirAll(filepath.Join(etcDir, "conf.d"), 0755); err != nil {
		return err
	}

	iniPath := filepath.Join(etcDir, "php.ini")
	if _, err := os.Stat(iniPath); err == nil {
		return nil // already exists
	}

	logsDir := filepath.Join(p.root, "php", p.majorMinor, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(
		"date.timezone = UTC\nerror_reporting = E_ALL\ndisplay_errors = On\nlog_errors = On\nerror_log = %s/php-error.log\n",
		logsDir,
	)
	return os.WriteFile(iniPath, []byte(content), 0644)
}

// WritePHPConfig writes PHP runtime settings to conf.d/limits.ini.
func (p *PHPInstaller) WritePHPConfig(cfg workspace.PHPVersionConfig) error {
	confdDir := filepath.Join(p.InstallDir(), "etc", "conf.d")
	if err := os.MkdirAll(confdDir, 0755); err != nil {
		return err
	}

	iniPath := filepath.Join(confdDir, "limits.ini")
	content := fmt.Sprintf(`; Auto-generated by chauf - edit via: chauf config php <version>
upload_max_filesize = %s
post_max_size = %s
memory_limit = %s
max_execution_time = %d
max_input_vars = %d
`, cfg.UploadMaxFilesize, cfg.PostMaxSize, cfg.MemoryLimit, cfg.MaxExecutionTime, cfg.MaxInputVars)

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// writeFPMConfig writes php-fpm.conf and the default pool config.
func (p *PHPInstaller) writeFPMConfig() error {
	mm := p.majorMinor
	root := p.root
	installDir := p.InstallDir()
	etcDir := filepath.Join(installDir, "etc")
	runtimeDir := filepath.Join(root, "php", mm, "runtime", "php-fpm")
	logsDir := filepath.Join(root, "php", mm, "logs")

	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	poolDir := filepath.Join(etcDir, "php-fpm.d")
	if err := os.MkdirAll(poolDir, 0755); err != nil {
		return err
	}

	// php-fpm.conf (global)
	fpmConf := filepath.Join(etcDir, "php-fpm.conf")
	if _, err := os.Stat(fpmConf); err != nil {
		content := fmt.Sprintf("[global]\npid = %s/php-fpm.pid\nerror_log = %s/php-fpm.log\ninclude = %s/*.conf\n",
			runtimeDir, logsDir, poolDir)
		if err := os.WriteFile(fpmConf, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Pool config
	user := os.Getenv("USER")
	if user == "" {
		user = "www-data"
	}
	sock := filepath.Join(runtimeDir, "php-fpm.sock")

	poolConf := filepath.Join(poolDir, "chauf-php.conf")
	if _, err := os.Stat(poolConf); err != nil {
		pool := fmt.Sprintf(`[chauf-php-%s]
user = %s
group = %s
listen = %s
listen.owner = %s
listen.group = %s
listen.mode = 0660
pm = ondemand
pm.max_children = 20
pm.start_servers = 4
pm.min_spare_servers = 2
pm.max_spare_servers = 6
php_admin_value[error_log] = %s/php-fpm-error.log
php_admin_flag[log_errors] = on
`, mm, user, user, sock, user, user, logsDir)
		if err := os.WriteFile(poolConf, []byte(pool), 0644); err != nil {
			return err
		}
	}

	return nil
}

// BuildImagick compiles and installs the imagick PECL extension for this PHP version.
func (p *PHPInstaller) BuildImagick() error {
	version, err := resolveImagickVersion()
	if err != nil {
		version = imagickDefaultVer
	}

	// Check for local tarball override
	tarPath := os.Getenv("CHAUFFEUR_IMAGICK_TARBALL")
	if tarPath == "" {
		filename := fmt.Sprintf("imagick-%s.tgz", version)
		cacheDir := filepath.Join(p.root, "cache", "php")
		tarPath = filepath.Join(cacheDir, filename)

		_, err = DownloadWithProgress(
			fmt.Sprintf("%s/imagick-%s.tgz", imagickDownloadBase, version),
			tarPath,
			p.opts.NoCache,
		)
		if err != nil {
			return fmt.Errorf("download imagick: %w", err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "chauf-imagick-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Construct the source directory explicitly from the version name.
	// PECL tarballs contain package.xml as their first entry, so we cannot
	// rely on ExtractTarGz's topDir return value (it would point to a file).
	if _, err := ExtractTarGz(tarPath, tmpDir, 0); err != nil {
		return fmt.Errorf("extract imagick: %w", err)
	}
	srcDir := filepath.Join(tmpDir, "imagick-"+version)
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf("imagick source dir not found (expected %s): %w", srcDir, err)
	}

	phpize := filepath.Join(p.InstallDir(), "bin", "phpize")
	phpConfig := filepath.Join(p.InstallDir(), "bin", "php-config")

	if fi, err := os.Stat(phpize); err != nil || fi.IsDir() {
		return fmt.Errorf("phpize not found at %s", phpize)
	}

	if err := RunCmd(srcDir, p.opts.Verbose, phpize); err != nil {
		return fmt.Errorf("phpize: %w", err)
	}
	if err := RunCmd(srcDir, p.opts.Verbose, "./configure", "--with-php-config="+phpConfig); err != nil {
		return fmt.Errorf("configure imagick: %w", err)
	}
	if err := RunCmd(srcDir, p.opts.Verbose, "make", "-j"+NumCPUs()); err != nil {
		return fmt.Errorf("make imagick: %w", err)
	}
	if err := RunCmd(srcDir, p.opts.Verbose, "make", "install"); err != nil {
		return fmt.Errorf("install imagick: %w", err)
	}

	iniPath := filepath.Join(p.InstallDir(), "etc", "conf.d", "imagick.ini")
	return os.WriteFile(iniPath, []byte("extension=imagick\n"), 0644)
}

// InstalledVersion returns the installed PHP version string, or "" if not present.
func (p *PHPInstaller) InstalledVersion() string {
	cmd := exec.Command(p.BinPath(), "--version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// "PHP 8.3.20 (cli) ..."
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 && parts[0] == "PHP" {
		return parts[1]
	}
	return ""
}

// ListInstalledPHP returns all PHP versions installed in the workspace (e.g. ["8.3", "8.1"]).
func ListInstalledPHP(root string) []string {
	phpDir := filepath.Join(root, "php")
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bin := filepath.Join(phpDir, e.Name(), "bin", "php")
		if _, err := os.Stat(bin); err == nil {
			versions = append(versions, e.Name())
		}
	}
	return versions
}

func resolveImagickVersion() (string, error) {
	if v := os.Getenv("CHAUFFEUR_IMAGICK_VERSION"); v != "" {
		return v, nil
	}
	resp, err := http.Get(imagickVersionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	v := strings.TrimSpace(string(body))
	if v == "" {
		return "", fmt.Errorf("empty response from PECL")
	}
	return v, nil
}
