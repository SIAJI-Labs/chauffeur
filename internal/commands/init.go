package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf init — initialize the Chauffeur workspace", "chauf init [--force] [--quiet]")
	force := flags.Bool("force", false, "Overwrite existing config with defaults")
	quiet := flags.Bool("quiet", false, "Suppress output; only print errors")
	flags.BoolVar(quiet, "q", false, "Suppress output; only print errors")

	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()

	if !*quiet {
		fmt.Println("Initializing workspace...")
		fmt.Println()
	}

	// ── Create directories ─────────────────────────────────────────────────────

	dirs := []string{
		"bin/shims",
		"config",
		"projects",
		"php",
		"nginx/bin",
		"nginx/etc/sites-available",
		"nginx/etc/sites-enabled",
		"nginx/certs",
		"nginx/logs",
		"composer",
		"cache/php",
		"cache/nginx",
		"cache/composer",
		"logs/commands",
		"system",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	if !*quiet {
		lib.Pair("Directories", "bin/shims/  config/  projects/  php/")
		fmt.Printf("  %-14s  %s\n", "", "nginx/etc/sites-available/  nginx/etc/sites-enabled/")
		fmt.Printf("  %-14s  %s\n", "", "nginx/certs/  cache/  logs/  system/")
	}

	// ── Write config ───────────────────────────────────────────────────────────

	configPath := filepath.Join(root, "config", "chauffeur.yaml")
	configCreated, err := ensureFile(configPath, workspace.DefaultConfigYAML(root), 0644, *force)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if !*quiet {
		printRow("Config", configPath, configCreated)
	}

	// ── Write nginx files ──────────────────────────────────────────────────────

	nginxFiles := []struct {
		rel     string
		content string
	}{
		{"nginx/etc/nginx.conf", nginxConfContent(root)},
		{"nginx/etc/fastcgi_params", fastcgiParamsContent},
		{"nginx/etc/mime.types", mimeTypesContent},
		{"nginx/etc/sites-available/default.conf", defaultSiteContent()},
	}

	for _, f := range nginxFiles {
		path := filepath.Join(root, f.rel)
		created, err := ensureFile(path, f.content, 0644, *force)
		if err != nil {
			return fmt.Errorf("%s: %w", f.rel, err)
		}
		if !*quiet {
			printRow("nginx/"+filepath.Base(f.rel), path, created)
		}
	}

	// ── Write shims ────────────────────────────────────────────────────────────
	// Shims are always updated if their content differs — this ensures a
	// running `chauf init` picks up shim fixes without requiring --force.

	phpShimPath := filepath.Join(root, "bin/shims/php")
	phpCreated, err := ensureShim(phpShimPath, phpShimContent, *force)
	if err != nil {
		return fmt.Errorf("php shim: %w", err)
	}
	if !*quiet {
		printRow("PHP shim", phpShimPath, phpCreated)
	}

	composerShimPath := filepath.Join(root, "bin/shims/composer")
	compCreated, err := ensureShim(composerShimPath, composerShimContent, *force)
	if err != nil {
		return fmt.Errorf("composer shim: %w", err)
	}
	if !*quiet {
		printRow("Composer shim", composerShimPath, compCreated)
	}

	// ── Shell config ───────────────────────────────────────────────────────────

	shellResult := injectShellPath(*quiet)

	// ── DNS & port forwarding checks ───────────────────────────────────────────

	cfg := workspace.Load()
	dnsStatus := checkDNSStatus()
	pfActive := system.IsPortForwardingActive(root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort)

	// ── Summary ────────────────────────────────────────────────────────────────

	if !*quiet {
		fmt.Println()
		lib.Success("Workspace ready")
		fmt.Println()

		switch shellResult.status {
		case shellInjected:
			lib.Pair("Shell config", shellResult.file+"  "+lib.Gray("(PATH added — run: source "+shellResult.file+")"))
		case shellAlready:
			lib.Pair("Shell config", lib.Gray(shellResult.file+"  (already configured)"))
		case shellSkipped:
			fmt.Printf("  Add to %s or %s:\n", lib.Cyan("~/.bashrc"), lib.Cyan("~/.zshrc"))
			fmt.Printf("    %s\n", lib.Bold(`export PATH="$HOME/.chauffeur/bin:$HOME/.chauffeur/bin/shims:$PATH"`))
			fmt.Println()
		}

		printDNSStatus(dnsStatus)
		printPortForwardingStatus(pfActive, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort)

		fmt.Println()
		fmt.Printf("  Next steps:\n")
		fmt.Printf("    %s\n", lib.Gray("chauf install nginx php 8.3 composer"))
		fmt.Printf("    %s\n", lib.Gray("chauf doctor"))
		fmt.Println()
	}

	return nil
}

// ── DNS detection ──────────────────────────────────────────────────────────────

type initDNSStatus struct {
	installed             bool // dnsmasq binary exists
	configured            bool // dnsmasq chauffeur.conf exists
	nmManaged             bool // NetworkManager manages dnsmasq
	resolvedActive        bool // systemd-resolved is running
	resolvedInNSS         bool // nsswitch.conf routes hostname lookups through systemd-resolved
	resolvedDropIn        bool // /etc/systemd/resolved.conf.d/chauffeur.conf exists
	nssResolveFallthrough bool // nsswitch.conf allows TRYAGAIN to fall through to dns module
}

func checkDNSStatus() initDNSStatus {
	r := initDNSStatus{}
	_, err := exec.LookPath("dnsmasq")
	r.installed = err == nil
	// NetworkManager dnsmasq plugin.
	entries, _ := os.ReadDir("/etc/NetworkManager/conf.d")
	for _, e := range entries {
		data, _ := os.ReadFile("/etc/NetworkManager/conf.d/" + e.Name())
		if strings.Contains(string(data), "dns=dnsmasq") {
			r.nmManaged = true
			break
		}
	}
	// Check the correct conf path based on who manages dnsmasq.
	if r.nmManaged {
		_, err = os.Stat("/etc/NetworkManager/dnsmasq.d/chauffeur.conf")
	} else {
		_, err = os.Stat("/etc/dnsmasq.d/chauffeur.conf")
	}
	r.configured = err == nil
	// systemd-resolved detection (reuse helpers defined in doctor.go, same package).
	r.resolvedActive = isResolvedActive()
	r.resolvedInNSS = isResolvedInNSS()
	_, err = os.Stat(resolvedDropIn)
	r.resolvedDropIn = err == nil
	r.nssResolveFallthrough = isNSSResolveFallthrough()
	return r
}

func printDNSStatus(r initDNSStatus) {
	if r.configured {
		confPath := "/etc/dnsmasq.d/chauffeur.conf"
		if r.nmManaged {
			confPath = "/etc/NetworkManager/dnsmasq.d/chauffeur.conf"
		}
		lib.Pair("DNS", lib.Green("✓")+"  .test domains → dnsmasq  ("+confPath+")")

		// Even if dnsmasq is configured, systemd-resolved may bypass it when offline.
		if r.resolvedActive && r.resolvedInNSS {
			if r.resolvedDropIn && r.nssResolveFallthrough {
				lib.Pair("   ", lib.Green("✓")+"  systemd-resolved .test route  ("+resolvedDropIn+")")
				lib.Pair("   ", lib.Green("✓")+"  NSS offline fallback  (resolve [NOTFOUND=return])")
			} else {
				fmt.Println()
				lib.Warn(".test sites unreachable without internet")
				fmt.Println()
				fmt.Printf("  systemd-resolved goes offline when your WAN interface disconnects, returning\n")
				fmt.Printf("  TRYAGAIN which NSS treats as fatal. The fix allows TRYAGAIN to fall through\n")
				fmt.Printf("  to the dns module, which uses resolv.conf → dnsmasq → .test works offline.\n")
				fmt.Println()
				fmt.Printf("  Run once to fix:\n")
				fmt.Println()
				if !r.resolvedDropIn {
					fmt.Printf("    %s\n", lib.Bold("sudo mkdir -p /etc/systemd/resolved.conf.d"))
					fmt.Printf("    %s\n", lib.Bold("printf '[Resolve]\\nDNS=127.0.0.1\\nDomains=~test\\n' | sudo tee "+resolvedDropIn))
					fmt.Printf("    %s\n", lib.Bold("sudo systemctl restart systemd-resolved"))
					fmt.Println()
				}
				if !r.nssResolveFallthrough {
					fmt.Printf("    %s\n", lib.Bold("sudo sed -i 's/resolve \\[!UNAVAIL=return\\]/resolve [NOTFOUND=return]/' /etc/nsswitch.conf"))
				}
				fmt.Println()
			}
		}
		return
	}

	fmt.Println()
	lib.Warn("DNS not configured — .test domains won't resolve")
	fmt.Println()

	if !r.installed {
		fmt.Printf("  Install dnsmasq first:\n")
		fmt.Println()
		// Detect distro from /etc/os-release
		osRelease, _ := os.ReadFile("/etc/os-release")
		switch {
		case strings.Contains(string(osRelease), "ID=arch") || strings.Contains(string(osRelease), "ID_LIKE=arch"):
			fmt.Printf("    %s\n", lib.Bold("sudo pacman -S dnsmasq"))
		case strings.Contains(string(osRelease), "ID=fedora") || strings.Contains(string(osRelease), "ID_LIKE=fedora"):
			fmt.Printf("    %s\n", lib.Bold("sudo dnf install dnsmasq"))
		default:
			fmt.Printf("    %s\n", lib.Bold("sudo apt-get install dnsmasq"))
		}
		fmt.Println()
		fmt.Printf("  Then run the DNS setup below.\n")
		fmt.Println()
	}

	fmt.Printf("  Run these commands to configure DNS:\n")
	fmt.Println()

	if r.nmManaged {
		fmt.Printf("    %s\n", lib.Bold("sudo mkdir -p /etc/NetworkManager/dnsmasq.d"))
		fmt.Printf("    %s\n", lib.Bold("echo 'address=/.test/127.0.0.1' | sudo tee /etc/NetworkManager/dnsmasq.d/chauffeur.conf"))
		fmt.Println()
		fmt.Printf("    %s\n", lib.Bold("sudo systemctl restart NetworkManager"))
		fmt.Printf("    %s\n", lib.Gray("(NetworkManager manages dnsmasq on your system)"))
	} else {
		fmt.Printf("    %s\n", lib.Bold("sudo mkdir -p /etc/dnsmasq.d"))
		fmt.Printf("    %s\n", lib.Bold("echo 'address=/.test/127.0.0.1' | sudo tee /etc/dnsmasq.d/chauffeur.conf"))
		fmt.Println()
		fmt.Printf("    %s\n", lib.Bold("sudo systemctl enable --now dnsmasq"))
		fmt.Printf("    %s\n", lib.Bold("sudo systemctl restart dnsmasq"))
	}
	fmt.Println()
	fmt.Printf("  Or add per-project lines to %s:\n", lib.Cyan("/etc/hosts"))
	fmt.Printf("    %s\n", lib.Gray("127.0.0.1  my-project.test"))
}

func printPortForwardingStatus(active bool, httpPort, httpsPort int) {
	if active {
		lib.Pair("Port forwarding", lib.Green("✓")+"  80→"+fmt.Sprintf("%d", httpPort)+"  443→"+fmt.Sprintf("%d", httpsPort)+"  (access without port)")
		return
	}

	fmt.Println()
	lib.Warn(fmt.Sprintf("Port forwarding not configured — access requires :%d / :%d", httpPort, httpsPort))
	fmt.Println()
	fmt.Printf("  Run once to access sites without a port number:\n")
	fmt.Println()
	for _, cmd := range system.PortForwardingCommands(httpPort, httpsPort) {
		fmt.Printf("    %s\n", lib.Bold(cmd))
	}
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Note: iptables rules reset on reboot. To make permanent:"))
	fmt.Printf("  %s\n", lib.Gray("  sudo iptables-save | sudo tee /etc/iptables/rules.v4"))
}

// ── shell config injection ─────────────────────────────────────────────────────

const (
	shellBlockOpen  = "# >>> chauffeur <<<"
	shellBlockClose = "# <<< chauffeur <<<"
	shellPathExport = `export PATH="$HOME/.chauffeur/bin:$HOME/.chauffeur/bin/shims:$PATH"`
)

var shellBlock = shellBlockOpen + "\n" + shellPathExport + "\n" + shellBlockClose + "\n"

type shellInjectStatus int

const (
	shellInjected shellInjectStatus = iota // newly written
	shellAlready                           // block already present
	shellSkipped                           // could not determine config file
)

type shellInjectResult struct {
	status shellInjectStatus
	file   string
}

// injectShellPath finds the user's shell config and appends the chauffeur PATH
// block if it isn't already present. It inserts right after any existing
// "# Chauffeur" comment line, or appends at the end if not found.
func injectShellPath(quiet bool) shellInjectResult {
	cfgFile := detectShellConfig()
	if cfgFile == "" {
		return shellInjectResult{status: shellSkipped}
	}

	home, _ := os.UserHomeDir()
	display := strings.Replace(cfgFile, home, "~", 1)

	data, err := os.ReadFile(cfgFile)
	if err != nil && !os.IsNotExist(err) {
		return shellInjectResult{status: shellSkipped}
	}
	content := string(data)

	// Already configured — nothing to do.
	if strings.Contains(content, shellBlockOpen) {
		return shellInjectResult{status: shellAlready, file: display}
	}

	// Try to insert right after a bare "# Chauffeur" section comment.
	const sectionComment = "# Chauffeur"
	var newContent string

	if idx := strings.Index(content, "\n"+sectionComment+"\n"); idx != -1 {
		insertAt := idx + len("\n"+sectionComment+"\n")
		newContent = content[:insertAt] + shellBlock + content[insertAt:]
	} else if strings.HasSuffix(strings.TrimRight(content, "\n"), sectionComment) {
		newContent = strings.TrimRight(content, "\n") + "\n" + shellBlock
	} else {
		// Append with a fresh section header.
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		newContent = content + "\n" + sectionComment + "\n" + shellBlock
	}

	if err := os.WriteFile(cfgFile, []byte(newContent), 0644); err != nil {
		return shellInjectResult{status: shellSkipped}
	}

	return shellInjectResult{status: shellInjected, file: display}
}

// detectShellConfig returns the path to the user's shell rc file.
func detectShellConfig() string {
	home, _ := os.UserHomeDir()
	shell := os.Getenv("SHELL")

	switch {
	case strings.HasSuffix(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.HasSuffix(shell, "bash"):
		return filepath.Join(home, ".bashrc")
	}

	// Fallback: use whichever rc file exists.
	for _, name := range []string{".zshrc", ".bashrc", ".profile"} {
		p := filepath.Join(home, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ensureFile writes content to path only if the file doesn't exist (or force is true).
// Returns true if the file was newly written or updated.
func ensureFile(path, content string, perm os.FileMode, force bool) (bool, error) {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return false, nil // already exists
		}
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return false, err
	}
	return true, nil
}

// ensureShim writes the shim if it doesn't exist OR if the content differs.
// This allows `chauf init` (without --force) to self-update stale shims.
// Returns true if the file was written.
func ensureShim(path, content string, force bool) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == content && !force {
		return false, nil // already up to date
	}
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return false, err
	}
	return true, nil
}

// printRow prints a label-value row, dimming the value when the file already existed.
func printRow(label, path string, created bool) {
	home, _ := os.UserHomeDir()
	display := strings.Replace(path, home, "~", 1)
	if created {
		lib.Pair(label, display)
	} else {
		lib.Pair(label, lib.Gray(display+"  (exists)"))
	}
}

// ── nginx config templates ─────────────────────────────────────────────────────

func nginxConfContent(root string) string {
	return fmt.Sprintf(`worker_processes auto;

error_log  %s/nginx/logs/error.log warn;
pid        %s/nginx/logs/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include       %s/nginx/etc/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent"';

    access_log  %s/nginx/logs/access.log  main;

    sendfile        on;
    tcp_nopush      on;
    keepalive_timeout  65;

    gzip  on;
    gzip_types text/plain text/css application/json application/javascript
               text/xml application/xml text/javascript;

    include %s/nginx/etc/sites-enabled/*.conf;
}
`, root, root, root, root, root)
}

func defaultSiteContent() string {
	return `server {
    listen 8080 default_server;
    server_name _;

    location / {
        default_type text/plain;
        return 503 "Chauffeur: no project configured for this domain.\n";
    }
}
`
}

const fastcgiParamsContent = `fastcgi_param  SCRIPT_FILENAME    $document_root$fastcgi_script_name;
fastcgi_param  QUERY_STRING       $query_string;
fastcgi_param  REQUEST_METHOD     $request_method;
fastcgi_param  CONTENT_TYPE       $content_type;
fastcgi_param  CONTENT_LENGTH     $content_length;

fastcgi_param  SCRIPT_NAME        $fastcgi_script_name;
fastcgi_param  REQUEST_URI        $request_uri;
fastcgi_param  DOCUMENT_URI       $document_uri;
fastcgi_param  DOCUMENT_ROOT      $document_root;
fastcgi_param  SERVER_PROTOCOL    $server_protocol;
fastcgi_param  REQUEST_SCHEME     $scheme;
fastcgi_param  HTTPS              $https if_not_empty;

fastcgi_param  GATEWAY_INTERFACE  CGI/1.1;
fastcgi_param  SERVER_SOFTWARE    nginx/$nginx_version;

fastcgi_param  REMOTE_ADDR        $remote_addr;
fastcgi_param  REMOTE_PORT        $remote_port;
fastcgi_param  SERVER_ADDR        $server_addr;
fastcgi_param  SERVER_PORT        $server_port;
fastcgi_param  SERVER_NAME        $server_name;

# PHP only, required if PHP was built with --enable-force-cgi-redirect
fastcgi_param  REDIRECT_STATUS    200;
`

const mimeTypesContent = `types {
    text/html                             html htm shtml;
    text/css                              css;
    text/xml                              xml;
    text/plain                            txt;
    text/mathml                           mml;
    text/x-component                      htc;

    image/gif                             gif;
    image/jpeg                            jpeg jpg;
    image/png                             png;
    image/svg+xml                         svg svgz;
    image/tiff                            tif tiff;
    image/webp                            webp;
    image/avif                            avif;
    image/x-icon                          ico;
    image/x-ms-bmp                        bmp;

    font/woff                             woff;
    font/woff2                            woff2;

    application/javascript                js;
    application/json                      json;
    application/pdf                       pdf;
    application/wasm                      wasm;
    application/xhtml+xml                 xhtml;
    application/atom+xml                  atom;
    application/rss+xml                   rss;
    application/zip                       zip;
    application/x-7z-compressed           7z;
    application/x-rar-compressed          rar;
    application/x-tar                     tar;
    application/x-x509-ca-cert            der pem crt;
    application/vnd.ms-excel              xls;
    application/vnd.ms-powerpoint        ppt;
    application/vnd.ms-fontobject         eot;
    application/vnd.openxmlformats-officedocument.spreadsheetml.sheet          xlsx;
    application/vnd.openxmlformats-officedocument.presentationml.presentation  pptx;
    application/vnd.openxmlformats-officedocument.wordprocessingml.document    docx;
    application/octet-stream              bin exe dll deb dmg iso img;

    audio/mpeg                            mp3;
    audio/ogg                             ogg;
    audio/x-m4a                           m4a;

    video/mp4                             mp4;
    video/mpeg                            mpeg mpg;
    video/webm                            webm;
    video/quicktime                       mov;
    video/x-flv                           flv;
    video/x-msvideo                       avi;
    video/mp2t                            ts;
}
`

// ── shim scripts ───────────────────────────────────────────────────────────────

const phpShimContent = `#!/usr/bin/env bash
# Chauffeur PHP shim — routes to the correct PHP binary based on active version.
# shim-version: 4

CHAUF_HOME="${CHAUFFEUR_HOME:-$HOME/.chauffeur}"
CONFIG="$CHAUF_HOME/config/chauffeur.yaml"
PROJECTS_DIR="$CHAUF_HOME/projects"

# 1. Explicit env override (e.g. CHAUFFEUR_PHP_VERSION=8.3 php artisan ...)
PHP_VER="${CHAUFFEUR_PHP_VERSION:-}"
PROJECT_WORKDIR=""

# 2. Scan project configs for one whose path matches CWD or a parent.
#    PHP version is stored in ~/.chauffeur/projects/<slug>/config.yaml — no dotfile in project.
if [[ -z "$PHP_VER" && -d "$PROJECTS_DIR" ]]; then
    CWD="$PWD"
    for cfg in "$PROJECTS_DIR"/*/config.yaml; do
        [[ -f "$cfg" ]] || continue
        proj_path=$(grep '^path:' "$cfg" 2>/dev/null | head -1 | sed 's/^path:[[:space:]]*//' | tr -d '"')
        [[ -n "$proj_path" ]] || continue
        # Match if CWD is the project dir or any subdirectory of it
        if [[ "$CWD" == "$proj_path" || "$CWD" == "$proj_path/"* ]]; then
            PHP_VER=$(grep '^php_version:' "$cfg" 2>/dev/null | head -1 | sed 's/^php_version:[[:space:]]*//' | tr -d '"')
            PROJECT_WORKDIR="/workspace/$(basename "$(dirname "$cfg")")"
            break
        fi
    done
fi

# 3. Fall back to global default from workspace config
if [[ -z "$PHP_VER" && -f "$CONFIG" ]]; then
    PHP_VER=$(grep 'default_version' "$CONFIG" 2>/dev/null | head -1 | sed 's/.*"\(.*\)".*/\1/')
fi

if [[ -z "$PHP_VER" ]]; then
    echo "chauf: no PHP version configured. Run: chauf php use <version>" >&2
    exit 1
fi

# Podman owns PHP execution when selected. Keep the shim transparent while
# routing the command through the matching PHP-FPM container.
RUNTIME=$(grep 'engine:' "$CONFIG" 2>/dev/null | head -1 | sed 's/.*engine:[[:space:]]*//' | tr -d '"')
if [[ "$RUNTIME" == "podman" ]]; then
    CONTAINER="chauf-php${PHP_VER//./}-fpm"
    IMAGE="ghcr.io/siegg/chauffeur-php:${PHP_VER}-fpm"
    HOST_USER="$(id -u):$(id -g)"
    if [[ -z "$PROJECT_WORKDIR" ]]; then
        exec podman run --rm --userns keep-id --user "$HOST_USER" --volume "$PWD:/workspace:Z" --workdir /workspace "$IMAGE" php "$@"
    fi
    EXEC_ARGS=(exec)
    EXEC_ARGS+=(--user "$HOST_USER")
    if [[ -n "$PROJECT_WORKDIR" ]]; then
        EXEC_ARGS+=(--workdir "$PROJECT_WORKDIR")
    fi
    exec podman "${EXEC_ARGS[@]}" "$CONTAINER" php "$@"
fi

PHP_BIN="$CHAUF_HOME/php/$PHP_VER/bin/php"

if [[ ! -x "$PHP_BIN" ]]; then
    echo "chauf: PHP $PHP_VER not installed. Run: chauf install php $PHP_VER" >&2
    exit 1
fi

exec "$PHP_BIN" "$@"
`

const composerShimContent = `#!/usr/bin/env bash
# Chauffeur Composer shim — routes Composer through the selected runtime.

CHAUF_HOME="${CHAUFFEUR_HOME:-$HOME/.chauffeur}"
CONFIG="$CHAUF_HOME/config/chauffeur.yaml"
PROJECTS_DIR="$CHAUF_HOME/projects"
COMPOSER_PHAR="$CHAUF_HOME/composer/composer.phar"

PHP_VER="${CHAUFFEUR_PHP_VERSION:-}"
PROJECT_WORKDIR=""
if [[ -z "$PHP_VER" && -d "$PROJECTS_DIR" ]]; then
    CWD="$PWD"
    for cfg in "$PROJECTS_DIR"/*/config.yaml; do
        [[ -f "$cfg" ]] || continue
        proj_path=$(grep '^path:' "$cfg" 2>/dev/null | head -1 | sed 's/^path:[[:space:]]*//' | tr -d '"')
        [[ -n "$proj_path" ]] || continue
        if [[ "$CWD" == "$proj_path" || "$CWD" == "$proj_path/"* ]]; then
            PHP_VER=$(grep '^php_version:' "$cfg" 2>/dev/null | head -1 | sed 's/^php_version:[[:space:]]*//' | tr -d '"')
            PROJECT_WORKDIR="/workspace/$(basename "$(dirname "$cfg")")"
            break
        fi
    done
fi
if [[ -z "$PHP_VER" && -f "$CONFIG" ]]; then
    PHP_VER=$(grep 'default_version' "$CONFIG" 2>/dev/null | head -1 | sed 's/.*"\(.*\)".*/\1/')
fi

RUNTIME=$(grep 'engine:' "$CONFIG" 2>/dev/null | head -1 | sed 's/.*engine:[[:space:]]*//' | tr -d '"')
if [[ "$RUNTIME" == "podman" ]]; then
    if [[ -z "$PHP_VER" ]]; then
        echo "chauf: no PHP version configured. Run: chauf php use <version>" >&2
        exit 1
    fi
    CONTAINER="chauf-php${PHP_VER//./}-fpm"
    IMAGE="ghcr.io/siegg/chauffeur-php:${PHP_VER}-fpm"
    HOST_USER="$(id -u):$(id -g)"
    if [[ -z "$PROJECT_WORKDIR" ]]; then
        exec podman run --rm --userns keep-id --user "$HOST_USER" --volume "$PWD:/workspace:Z" --workdir /workspace "$IMAGE" composer "$@"
    fi
    EXEC_ARGS=(exec)
    EXEC_ARGS+=(--user "$HOST_USER")
    if [[ -n "$PROJECT_WORKDIR" ]]; then
        EXEC_ARGS+=(--workdir "$PROJECT_WORKDIR")
    fi
    exec podman "${EXEC_ARGS[@]}" "$CONTAINER" composer "$@"
fi

if [[ ! -f "$COMPOSER_PHAR" ]]; then
    echo "chauf: Composer not installed. Run: chauf install composer" >&2
    exit 1
fi

exec "$(dirname "$0")/php" "$COMPOSER_PHAR" "$@"
`
