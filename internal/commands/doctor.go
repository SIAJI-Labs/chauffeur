package commands

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/services"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── types ──────────────────────────────────────────────────────────────────────

type checkResult struct {
	name        string
	ok          bool   // true = pass
	warn        bool   // true = warning (not blocking)
	status      string // one-line status message
	fix         string // command(s) to print when --fix is passed
	skipAutoFix bool   // true = skip this check in auto-fix mode
}

// ── RunDoctor ─────────────────────────────────────────────────────────────────

func RunDoctor(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		return doctorHelp()
	}

	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	doDeps := flags.Bool("check-deps", false, "Check system build dependencies")
	doPHP := flags.Bool("check-php", false, "Check PHP build libraries")
	doSSL := flags.Bool("check-ssl", false, "Check SSL tools")
	doNetwork := flags.Bool("check-network", false, "Check port availability and iptables")
	doDNS := flags.Bool("check-dns", false, "Check DNS resolution for .test domains")
	doFix := flags.Bool("fix", false, "Print fix commands for failed checks")
	doAutoFix := flags.Bool("auto-fix", false, "Execute fix commands for failed checks")

	if err := flags.Parse(args); err != nil {
		return err
	}

	// Default: run all checks
	if !*doDeps && !*doPHP && !*doSSL && !*doNetwork && !*doDNS {
		*doDeps = true
		*doPHP = true
		*doSSL = true
		*doNetwork = true
		*doDNS = true
	}

	root := workspace.Root()
	cfg := workspace.Load()

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Chauffeur Doctor"))
	printRuntime(cfg)
	fmt.Println()

	var all []checkResult
	runtimeChecks := doctorRuntimeOwnership(root, cfg)
	printDoctorSection("Runtime Ownership", runtimeChecks)
	all = append(all, runtimeChecks...)

	if *doDeps {
		r := doctorSystemDeps()
		printDoctorSection("System Dependencies", r)
		all = append(all, r...)
	}
	if *doPHP {
		r := doctorPHPRuntime(cfg)
		printDoctorSection("PHP Runtime", r)
		all = append(all, r...)
	}
	if *doSSL {
		r := doctorSSL(root)
		printDoctorSection("SSL", r)
		all = append(all, r...)
	}
	if *doNetwork {
		r := doctorNetwork(root, cfg)
		printDoctorSection("Network", r)
		all = append(all, r...)
	}
	if *doDNS {
		r := doctorDNS(root)
		printDoctorSection("DNS", r)
		all = append(all, r...)
	}

	// ── Summary ────────────────────────────────────────────────────────────────

	passed, warned, failed := doctorTally(all)
	fmt.Printf("  %s\n", strings.Repeat("─", 60))
	fmt.Printf("  Checks: %d   %s   %s   %s\n",
		len(all),
		lib.Green(fmt.Sprintf("✓ %d passed", passed)),
		lib.Yellow(fmt.Sprintf("⚠ %d warnings", warned)),
		lib.Red(fmt.Sprintf("✗ %d failed", failed)),
	)
	fmt.Println()

	// ── Fix suggestions / auto-fix ─────────────────────────────────────────────

	var fixes []checkResult
	for _, r := range all {
		if !r.ok && r.fix != "" && !r.skipAutoFix {
			fixes = append(fixes, r)
		}
	}

	switch {
	case *doAutoFix:
		if len(fixes) == 0 {
			lib.Success("No fixes needed.")
		} else {
			doctorAutoFix(fixes)
		}
	case *doFix:
		if len(fixes) == 0 {
			lib.Success("No fixes needed.")
		} else {
			fmt.Printf("  %s\n", lib.Bold("Fix Commands"))
			fmt.Println()
			for _, r := range fixes {
				tag := lib.Red("✗")
				if r.warn {
					tag = lib.Yellow("⚠")
				}
				fmt.Printf("  %s  %s\n", tag, lib.Bold(r.name))
				for _, line := range strings.Split(strings.TrimSpace(r.fix), "\n") {
					fmt.Printf("       %s\n", lib.Cyan(line))
				}
				fmt.Println()
			}
		}
	default:
		if failed > 0 || warned > 0 {
			lib.Info(lib.Gray("Run with --fix to see fix commands, or --auto-fix to apply them."))
		}
	}
	fmt.Println()

	if failed > 0 {
		return fmt.Errorf("doctor found %d issue(s)", failed)
	}
	return nil
}

func doctorRuntimeOwnership(root string, cfg workspace.Config) []checkResult {
	results := []checkResult{}
	nativeEnabled := system.IsUnitEnabled("chauffeur-nginx.service")
	podmanEnabled := system.IsUnitEnabled(system.PodmanNginxUnit())
	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) && nativeEnabled {
		results = append(results, checkResult{name: "runtime ownership", ok: false, status: "Podman selected but native nginx auto-start is enabled", fix: "chauf autostart enable"})
	} else if cfg.Runtime.Engine == string(chauftruntime.EngineNative) && podmanEnabled {
		results = append(results, checkResult{name: "runtime ownership", ok: false, status: "native selected but Podman nginx auto-start is enabled", fix: "chauf autostart disable"})
	} else {
		results = append(results, checkResult{name: "runtime ownership", ok: true, status: cfg.Runtime.Engine + " has no conflicting nginx unit"})
	}
	home, _ := os.UserHomeDir()
	defaultRoot := filepath.Join(home, ".chauffeur")
	if filepath.Clean(root) != filepath.Clean(defaultRoot) && (nativeEnabled || podmanEnabled) {
		results = append(results, checkResult{name: "autostart workspace", ok: false, warn: true, status: fmt.Sprintf("units use %s, but CLI workspace is %s", defaultRoot, root), fix: "disable legacy auto-start units and recreate them for this workspace"})
	}
	return results
}

func doctorPHPRuntime(cfg workspace.Config) []checkResult {
	if cfg.Runtime.Engine != string(chauftruntime.EnginePodman) {
		results := doctorPHPDeps()
		return append(results, doctorNativeImagick(workspace.Root())...)
	}
	results := []checkResult{}
	rt := chauftruntime.Podman{Runner: chauftruntime.ExecRunner{}}
	if err := rt.Preflight(context.Background()); err != nil {
		return []checkResult{{name: "rootless Podman", status: err.Error(), fix: "chauf config runtime native"}}
	}
	for _, version := range chauftruntime.PHPParityTargets() {
		parity, err := chauftruntime.CheckPHPParity(context.Background(), chauftruntime.ExecRunner{}, version)
		if err != nil {
			results = append(results, checkResult{name: "PHP " + version, status: err.Error(), fix: "chauf install php " + version + " --build"})
			continue
		}
		if parity.State == chauftruntime.ParityUnavailable {
			results = append(results, checkResult{name: "PHP " + version, status: parity.Evidence, fix: "chauf install php " + version + " --build"})
			continue
		}
		results = append(results, checkResult{name: "PHP " + version, ok: parity.State == chauftruntime.ParityVerified, status: parity.Evidence, fix: "chauf install php " + version + " --build"})
	}
	return results
}

func doctorNativeImagick(root string) []checkResult {
	var results []checkResult
	for _, version := range installers.ListInstalledPHP(root) {
		etcDir := filepath.Join(root, "php", version, "etc")
		iniPath := filepath.Join(etcDir, "conf.d", "imagick.ini")
		if _, err := os.Stat(iniPath); err != nil {
			continue
		}
		phpBin := filepath.Join(root, "php", version, "bin", "php")
		cmd := exec.Command(phpBin, "-c", etcDir, "-r", "if (!extension_loaded('imagick')) { exit(1); }")
		output, err := cmd.CombinedOutput()
		if err != nil {
			status := strings.TrimSpace(string(output))
			if status == "" {
				status = "module could not be loaded"
			}
			results = append(results, checkResult{
				name:   "PHP " + version + " imagick",
				warn:   true,
				status: status,
				fix:    "chauf install php " + version + " --force",
			})
			continue
		}
		results = append(results, checkResult{name: "PHP " + version + " imagick", ok: true, status: "loadable"})
	}
	return results
}

// ── check sections ─────────────────────────────────────────────────────────────

func doctorSystemDeps() []checkResult {
	type toolDef struct {
		name string
		args []string
		pkgs pkgNames
	}
	tools := []toolDef{
		{"git", []string{"--version"}, pkgNames{"git", "git", "git"}},
		{"curl", []string{"--version"}, pkgNames{"curl", "curl", "curl"}},
		{"tar", []string{"--version"}, pkgNames{"tar", "tar", "tar"}},
		{"gcc", []string{"--version"}, pkgNames{"gcc", "build-essential", "gcc"}},
		{"make", []string{"--version"}, pkgNames{"make", "make", "make"}},
		{"pkg-config", []string{"--version"}, pkgNames{"pkgconf", "pkg-config", "pkgconf"}},
		{"autoconf", []string{"--version"}, pkgNames{"autoconf", "autoconf", "autoconf"}},
		{"bison", []string{"--version"}, pkgNames{"bison", "bison", "bison"}},
	}

	dm := detectDistroType()
	var results []checkResult
	for _, t := range tools {
		path, err := exec.LookPath(t.name)
		if err != nil {
			results = append(results, checkResult{
				name:   t.name,
				ok:     false,
				status: "not found",
				fix:    buildInstallCmd(dm, t.pkgs),
			})
			continue
		}
		ver := firstLine(cmdOutput(path, t.args...))
		results = append(results, checkResult{name: t.name, ok: true, status: lib.Gray(ver)})
	}
	return results
}

func doctorPHPDeps() []checkResult {
	type depDef struct {
		pkgConfig string
		display   string
		pkgs      pkgNames
		optional  bool
	}
	deps := []depDef{
		{"libzip", "libzip", pkgNames{"libzip", "libzip-dev", "libzip-devel"}, false},
		{"libjpeg", "libjpeg", pkgNames{"libjpeg-turbo", "libjpeg-dev", "libjpeg-devel"}, false},
		{"libpng", "libpng", pkgNames{"libpng", "libpng-dev", "libpng-devel"}, false},
		{"freetype2", "freetype2", pkgNames{"freetype2", "libfreetype6-dev", "freetype-devel"}, false},
		{"libxml-2.0", "libxml2", pkgNames{"libxml2", "libxml2-dev", "libxml2-devel"}, false},
		{"libcurl", "libcurl", pkgNames{"curl", "libcurl4-openssl-dev", "libcurl-devel"}, false},
		{"zlib", "zlib", pkgNames{"zlib", "zlib1g-dev", "zlib-devel"}, false},
		{"readline", "readline", pkgNames{"readline", "libreadline-dev", "readline-devel"}, false},
		{"libxslt", "libxslt", pkgNames{"libxslt", "libxslt1-dev", "libxslt-devel"}, false},
		{"libpq", "libpq", pkgNames{"postgresql-libs", "libpq-dev", "postgresql-devel"}, false},
		{"gmp", "gmp", pkgNames{"gmp", "libgmp-dev", "gmp-devel"}, false},
		{"openssl", "openssl", pkgNames{"openssl", "libssl-dev", "openssl-devel"}, false},
		{"MagickWand", "ImageMagick", pkgNames{"imagemagick", "libmagickwand-dev", "ImageMagick-devel"}, true},
	}

	dm := detectDistroType()
	var results []checkResult
	for _, d := range deps {
		out, err := exec.Command("pkg-config", "--modversion", d.pkgConfig).Output()
		name := d.display
		if d.optional {
			name += " (optional)"
		}
		if err != nil {
			results = append(results, checkResult{
				name:   name,
				ok:     false,
				warn:   d.optional,
				status: "not found",
				fix:    buildInstallCmd(dm, d.pkgs),
			})
			continue
		}
		ver := strings.TrimSpace(string(out))
		result := checkResult{name: name, ok: true, status: lib.Gray(ver)}
		// Keep remediation available to diagnostics even when the dependency is
		// currently installed; it is unused for healthy command output.
		if d.display == "libpq" {
			result.fix = buildInstallCmd(dm, d.pkgs)
		}
		results = append(results, result)
	}
	return results
}

func doctorSSL(root string) []checkResult {
	var results []checkResult

	// openssl binary
	if path, err := exec.LookPath("openssl"); err != nil {
		results = append(results, checkResult{
			name:   "openssl",
			ok:     false,
			status: "not found",
			fix:    buildInstallCmd(detectDistroType(), pkgNames{"openssl", "openssl", "openssl"}),
		})
	} else {
		ver := firstLine(cmdOutput(path, "version"))
		results = append(results, checkResult{name: "openssl", ok: true, status: lib.Gray(ver)})
	}

	// mkcert
	mkcertPath, err := exec.LookPath("mkcert")
	if err != nil {
		results = append(results, checkResult{
			name:   "mkcert",
			ok:     false,
			status: "not found",
			fix: "# Install from https://github.com/FiloSottile/mkcert\n" +
				buildInstallCmd(detectDistroType(), pkgNames{"mkcert", "mkcert", "mkcert"}),
		})
	} else {
		ver := firstLine(cmdOutput(mkcertPath, "-version"))
		results = append(results, checkResult{name: "mkcert", ok: true, status: lib.Gray(ver)})

		// CA in system trust? Use the improved projects.MkcertCAInstalled()
		// which checks both CAROOT existence AND presence in the system bundle.
		if projects.MkcertCAInstalled() {
			caRoot := strings.TrimSpace(cmdOutput(mkcertPath, "-CAROOT"))
			results = append(results, checkResult{
				name:   "mkcert CA",
				ok:     true,
				status: lib.Gray(shortenHome(caRoot) + " (system trusted)"),
			})
		} else {
			// CA files exist but not trusted, or not installed at all.
			caRoot := strings.TrimSpace(cmdOutput(mkcertPath, "-CAROOT"))
			_, e1 := os.Stat(filepath.Join(caRoot, "rootCA.pem"))
			_, e2 := os.Stat(filepath.Join(caRoot, "rootCA-key.pem"))
			if e1 == nil && e2 == nil {
				results = append(results, checkResult{
					name:   "mkcert CA",
					ok:     false,
					status: "CA files exist but NOT in system trust store",
					fix:    "sudo mkcert -install",
				})
			} else {
				results = append(results, checkResult{
					name:   "mkcert CA",
					ok:     false,
					status: "CA not installed — SSL errors will occur for HTTPS sites",
					fix:    "sudo mkcert -install",
				})
			}
		}
	}

	// PHP CA bundle (curl.cainfo / openssl.cafile) — iterates all installed versions
	for _, res := range doctorPHPCertBundle(root) {
		results = append(results, res)
	}

	// cert directory
	certDir := filepath.Join(root, "nginx", "certs")
	if _, err := os.Stat(certDir); err != nil {
		results = append(results, checkResult{
			name:   "cert directory",
			ok:     false,
			warn:   true,
			status: "missing — created on first: chauf secure",
		})
	} else {
		entries, _ := os.ReadDir(certDir)
		results = append(results, checkResult{
			name:   "cert directory",
			ok:     true,
			status: lib.Gray(fmt.Sprintf("%d file(s)  %s", len(entries), shortenHome(certDir))),
		})
	}

	return results
}

// checkSinglePHPVersionCerts checks curl.cainfo and openssl.cafile for a specific PHP binary.
func checkSinglePHPVersionCerts(phpBin string) struct {
	ok         bool
	warn       bool
	status     string
	fix        string
	needsFix   bool
	phpIniFile string
	version    string // PHP version string like "8.3"
} {
	ret := struct {
		ok         bool
		warn       bool
		status     string
		fix        string
		needsFix   bool
		phpIniFile string
		version    string
	}{warn: true, needsFix: true}

	// Get php -i output for version and settings
	infoOut := cmdOutput(phpBin, "-i")

	// Parse PHP version from "PHP Version => 8.3.30"
	for _, line := range strings.Split(infoOut, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "PHP Version =>") {
			parts := strings.Split(trimmed, "=>")
			if len(parts) >= 2 {
				ret.version = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	// Get loaded php.ini
	iniOut := cmdOutput(phpBin, "--ini")
	for _, line := range strings.Split(iniOut, "\n") {
		if strings.Contains(line, "Loaded Configuration") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ret.phpIniFile = strings.TrimSpace(parts[len(parts)-1])
			}
			break
		}
	}

	// Parse curl.cainfo and openssl.cafile
	parsePHPValue := func(raw string) string {
		val := strings.TrimSpace(raw)
		if val == "" || strings.EqualFold(val, "no value") {
			return ""
		}
		return val
	}

	var curlCA, opensslCA string
	for _, line := range strings.Split(infoOut, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "curl.cainfo =>") {
			parts := strings.Split(trimmed, "=>")
			if len(parts) >= 2 {
				curlCA = parsePHPValue(parts[1])
			}
		}
		if strings.HasPrefix(trimmed, "openssl.cafile =>") {
			parts := strings.Split(trimmed, "=>")
			if len(parts) >= 2 {
				opensslCA = parsePHPValue(parts[1])
			}
		}
	}

	// Determine effective CA bundle
	caBundle := curlCA
	if caBundle == "" {
		caBundle = opensslCA
	}

	// Build status
	var details []string
	if curlCA != "" {
		details = append(details, "curl.cainfo="+shortenHome(curlCA))
	}
	if opensslCA != "" && opensslCA != curlCA {
		details = append(details, "openssl.cafile="+shortenHome(opensslCA))
	}

	if caBundle == "" {
		ret.status = "not configured"
		if ret.phpIniFile != "" && ret.phpIniFile != "none" {
			dist := detectDistroType()
			var caPath string
			switch dist {
			case distroArch, distroDebian:
				caPath = "/etc/ssl/certs/ca-certificates.crt"
			case distroFedora:
				caPath = "/etc/pki/tls/certs/ca-bundle.crt"
			default:
				caPath = "/etc/ssl/certs/ca-certificates.crt"
			}
			ret.fix = fmt.Sprintf("printf 'curl.cainfo=%s\\nopenssl.cafile=%s\\n' >> %s", caPath, caPath, ret.phpIniFile)
		} else {
			ret.fix = "Configure curl.cainfo and openssl.cafile in php.ini"
		}
		return ret
	}

	// Check if bundle exists
	info, err := os.Stat(caBundle)
	if err != nil {
		ret.status = fmt.Sprintf("%s (file missing)", shortenHome(caBundle))
		return ret
	}
	if info.IsDir() {
		ret.status = fmt.Sprintf("%s (is a directory)", shortenHome(caBundle))
		return ret
	}

	// Check mkcert CA
	mkcertInstalled := projects.MkcertCAInstalled()
	if !mkcertInstalled {
		ret.ok = true
		ret.warn = true
		ret.needsFix = false
		ret.status = strings.Join(details, "  ") + fmt.Sprintf("  %s", shortenHome(caBundle))
		return ret
	}

	data, _ := os.ReadFile(caBundle)
	if strings.Contains(string(data), "mkcert") {
		ret.ok = true
		ret.needsFix = false
		ret.status = strings.Join(details, "  ") + fmt.Sprintf("  %s (mkcert OK)", shortenHome(caBundle))
		return ret
	}

	ret.status = fmt.Sprintf("mkcert CA not in bundle (%s)", shortenHome(caBundle))
	dist := detectDistroType()
	var suggested string
	switch dist {
	case distroArch, distroDebian:
		suggested = "/etc/ssl/certs/ca-certificates.crt"
	case distroFedora:
		suggested = "/etc/pki/tls/certs/ca-bundle.crt"
	default:
		suggested = caBundle
	}
	if suggested != caBundle && ret.phpIniFile != "" {
		ret.fix = fmt.Sprintf("printf 'curl.cainfo=%s\\nopenssl.cafile=%s\\n' | tee -a %s", suggested, suggested, ret.phpIniFile)
	}
	return ret
}

// doctorPHPCertBundle checks ALL installed PHP versions for curl.cainfo and
// openssl.cafile configuration. Returns one checkResult per PHP version.
func doctorPHPCertBundle(root string) []checkResult {
	var results []checkResult

	versions := installedPHPVersions(root)
	if len(versions) == 0 {
		return []checkResult{{
			name:   "PHP CA bundle",
			ok:     false,
			warn:   true,
			status: "no PHP versions installed",
			fix:    "Install PHP: chauf install php 8.3",
		}}
	}

	var configured, total int
	var allFixes []string

	for _, ver := range versions {
		phpBin := filepath.Join(root, "php", ver, "bin", "php")

		// Check if binary exists
		if _, err := os.Stat(phpBin); err != nil {
			results = append(results, checkResult{
				name:   fmt.Sprintf("PHP %s CA bundle", ver),
				ok:     false,
				warn:   true,
				status: "not installed",
			})
			continue
		}

		total++
		check := checkSinglePHPVersionCerts(phpBin)

		if check.ok && !check.needsFix {
			configured++
		}

		// Build fix command with full path
		if check.needsFix && check.fix != "" {
			// The fix references phpIniFile which checkSinglePHPVersionCerts already computed
			// We need to pass the full phpIniFile path
			allFixes = append(allFixes, fmt.Sprintf("printf 'curl.cainfo=/etc/ssl/certs/ca-certificates.crt\\nopenssl.cafile=/etc/ssl/certs/ca-certificates.crt\\n' >> %s", filepath.Join(root, "php", ver, "etc", "php.ini")))
		}

		results = append(results, checkResult{
			name: fmt.Sprintf("PHP %s CA bundle", ver),
			ok:   check.ok,
			warn: check.warn,
			status: func() string {
				if check.phpIniFile != "" {
					return check.status + " (" + shortenHome(check.phpIniFile) + ")"
				}
				return check.status
			}(),
			fix:         check.fix,
			skipAutoFix: check.needsFix, // skip individual fix - let aggregated fix handle it
		})
	}

	// Add summary result
	summaryStatus := fmt.Sprintf("%d of %d configured", configured, total)
	if configured == total && total > 0 {
		summaryStatus += " — all OK"
	} else if configured < total && configured > 0 {
		summaryStatus += fmt.Sprintf(" — %d need fixing", total-configured)
	}

	// Build aggregated fix command
	var aggregatedFix string
	if len(allFixes) > 0 {
		aggregatedFix = "# Fix all PHP versions:\n" + strings.Join(allFixes, "\n")
	}

	results = append([]checkResult{{
		name:        "PHP CA bundle",
		ok:          configured == total && total > 0,
		warn:        configured < total,
		status:      summaryStatus,
		fix:         aggregatedFix,
		skipAutoFix: len(allFixes) == 0,
	}}, results...)

	return results
}

func doctorNetwork(root string, cfg workspace.Config) []checkResult {
	var results []checkResult

	// iptables / port forwarding
	if _, err := exec.LookPath("iptables"); err != nil {
		results = append(results, checkResult{
			name:   "iptables",
			ok:     false,
			warn:   true,
			status: "not found — port forwarding unavailable",
			fix:    buildInstallCmd(detectDistroType(), pkgNames{"iptables", "iptables", "iptables"}),
		})
	} else if system.IsPortForwardingActive(root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort) {
		results = append(results, checkResult{
			name:   "port forwarding",
			ok:     true,
			status: lib.Gray(fmt.Sprintf("80→%d  443→%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort)),
		})
	} else {
		cmds := system.PortForwardingSystemdCommands(cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort)
		fix := "# Installs a systemd service — runs at boot, requires sudo once:\n" +
			strings.Join(cmds, "\n")
		results = append(results, checkResult{
			name:   "port forwarding",
			ok:     false,
			warn:   true,
			status: fmt.Sprintf("not configured — access needs :%d / :%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort),
			fix:    fix,
		})
	}

	// nginx + FPM ports
	nginxSvc := services.NewNginxService(root)
	for _, port := range []int{cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort} {
		label := fmt.Sprintf("port %d", port)

		if cfg.Runtime.Engine != string(chauftruntime.EnginePodman) && nginxSvc.IsRunning() {
			// PID file confirms our nginx owns this port.
			results = append(results, checkResult{
				name: label, ok: true,
				status: lib.Gray(fmt.Sprintf("nginx pid %d", nginxSvc.PID())),
			})
			continue
		}

		pid, procName, _ := services.FindProcessOnPort(port)

		switch {
		case services.IsPortAvailable(port):
			// Port is free — nginx can start.
			results = append(results, checkResult{name: label, ok: true, status: lib.Gray("available")})

		case pid > 0 && strings.Contains(procName, "nginx"):
			// Port is held by an nginx process — that's our nginx even if the
			// PID file is stale (e.g. after a failed systemd handoff).
			results = append(results, checkResult{
				name: label, ok: true,
				status: lib.Gray(fmt.Sprintf("nginx pid %d", pid)),
			})

		default:
			// Genuinely occupied by a foreign process.
			msg := "in use by another process"
			fix := ""
			if pid > 0 {
				msg = fmt.Sprintf("in use by %s (pid %d)", procName, pid)
				fix = fmt.Sprintf("kill %d", pid)
			}
			results = append(results, checkResult{name: label, ok: false, status: msg, fix: fix})
		}
	}

	return results
}

func doctorDNS(root string) []checkResult {
	var results []checkResult

	// dnsmasq binary
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		results = append(results, checkResult{
			name:   "dnsmasq",
			ok:     false,
			status: "not installed",
			fix:    buildInstallCmd(detectDistroType(), pkgNames{"dnsmasq", "dnsmasq", "dnsmasq"}),
		})
	} else {
		results = append(results, checkResult{name: "dnsmasq", ok: true, status: lib.Gray("installed")})
	}

	// chauffeur.conf
	nmManaged := isNMManagingDnsmasq()
	confPath := "/etc/dnsmasq.d/chauffeur.conf"
	if nmManaged {
		confPath = "/etc/NetworkManager/dnsmasq.d/chauffeur.conf"
	}
	if _, err := os.Stat(confPath); err != nil {
		restart := "sudo systemctl enable --now dnsmasq"
		if nmManaged {
			restart = "sudo systemctl restart NetworkManager"
		}
		results = append(results, checkResult{
			name:   "dnsmasq config",
			ok:     false,
			status: confPath + " missing",
			fix: "sudo mkdir -p " + filepath.Dir(confPath) + "\n" +
				"echo 'address=/.test/127.0.0.1' | sudo tee " + confPath + "\n" +
				restart,
		})
	} else {
		results = append(results, checkResult{
			name:   "dnsmasq config",
			ok:     true,
			status: lib.Gray(confPath),
		})
	}

	// systemd-resolved: when in the NSS chain, check two things:
	//   1. Global routing: resolved.conf.d drop-in routes .test to dnsmasq when online.
	//   2. Offline fallback: nsswitch.conf must allow TRYAGAIN to fall through to `dns`
	//      module. When resolved goes offline it returns TRYAGAIN (not UNAVAIL).
	//      The default [!UNAVAIL=return] blocks TRYAGAIN, so curl fails offline.
	//      Changing to [NOTFOUND=return] lets TRYAGAIN fall through → dns module
	//      reads resolv.conf (nameserver 127.0.0.1) → dnsmasq → .test works offline.
	if isResolvedActive() && isResolvedInNSS() {
		// Check 1: online .test routing via resolved global config.
		if _, err := os.Stat(resolvedDropIn); err == nil {
			results = append(results, checkResult{
				name:   "resolved .test route",
				ok:     true,
				status: lib.Gray(resolvedDropIn),
			})
		} else {
			results = append(results, checkResult{
				name:   "resolved .test route",
				ok:     false,
				warn:   true,
				status: ".test may not resolve online — global routing config missing",
				fix: "sudo mkdir -p /etc/systemd/resolved.conf.d\n" +
					"printf '[Resolve]\\nDNS=127.0.0.1\\nDomains=~test\\n' | sudo tee " + resolvedDropIn + "\n" +
					"sudo systemctl restart systemd-resolved",
			})
		}

		// Check 2: offline fallback via NSS TRYAGAIN passthrough.
		if isNSSResolveFallthrough() {
			results = append(results, checkResult{
				name:   "NSS offline fallback",
				ok:     true,
				status: lib.Gray("resolve [NOTFOUND=return] → dns → dnsmasq"),
			})
		} else {
			results = append(results, checkResult{
				name:   "NSS offline fallback",
				ok:     false,
				status: ".test unreachable offline — NSS stops on TRYAGAIN before reaching dns module",
				fix:    "sudo sed -i 's/resolve \\[!UNAVAIL=return\\]/resolve [NOTFOUND=return]/' /etc/nsswitch.conf",
			})
		}
	}

	// .test resolution (offline-safe: query 127.0.0.1:53 directly)
	if resolves := doctorTestResolution(); resolves {
		results = append(results, checkResult{
			name:   ".test resolution",
			ok:     true,
			status: lib.Gray("*.test → 127.0.0.1"),
		})
	} else {
		results = append(results, checkResult{
			name:   ".test resolution",
			ok:     false,
			warn:   true,
			status: "*.test not resolving to 127.0.0.1",
			fix:    "dig @127.0.0.1 test.test +short  # should return 127.0.0.1",
		})
	}

	return results
}

// ── output helpers ─────────────────────────────────────────────────────────────

func printDoctorSection(title string, results []checkResult) {
	const nameW = 24
	fmt.Printf("  %s\n", lib.Bold(title))
	for _, r := range results {
		var icon string
		switch {
		case r.ok:
			icon = lib.Green("✓")
		case r.warn:
			icon = lib.Yellow("⚠")
		default:
			icon = lib.Red("✗")
		}
		fmt.Printf("  %s  %-*s  %s\n", icon, nameW, r.name, r.status)
	}
	fmt.Println()
}

func doctorTally(results []checkResult) (passed, warned, failed int) {
	for _, r := range results {
		switch {
		case r.ok:
			passed++
		case r.warn:
			warned++
		default:
			failed++
		}
	}
	return
}

// ── distro / package helpers ───────────────────────────────────────────────────

// pkgNames holds [arch, debian, fedora] package names.
type pkgNames [3]string

type distroType int

const (
	distroArch   distroType = iota
	distroDebian            // Ubuntu / Debian
	distroFedora            // Fedora / RHEL / Rocky
	distroUnknown
)

func detectDistroType() distroType {
	data, _ := os.ReadFile("/etc/os-release")
	s := string(data)
	switch {
	case strings.Contains(s, "ID=arch") || strings.Contains(s, "ID_LIKE=arch"):
		return distroArch
	case strings.Contains(s, "ID=fedora") || strings.Contains(s, "ID_LIKE=fedora") ||
		strings.Contains(s, "ID=rhel") || strings.Contains(s, "ID_LIKE=rhel"):
		return distroFedora
	case strings.Contains(s, "ID=ubuntu") || strings.Contains(s, "ID=debian") ||
		strings.Contains(s, "ID_LIKE=debian"):
		return distroDebian
	}
	return distroUnknown
}

func buildInstallCmd(dm distroType, pkgs pkgNames) string {
	switch dm {
	case distroArch:
		if pkgs[0] != "" {
			return "sudo pacman -S " + pkgs[0]
		}
	case distroDebian:
		if pkgs[1] != "" {
			return "sudo apt install " + pkgs[1]
		}
	case distroFedora:
		if pkgs[2] != "" {
			return "sudo dnf install " + pkgs[2]
		}
	}
	// Unknown: show all options
	var lines []string
	if pkgs[0] != "" {
		lines = append(lines, "# Arch:   sudo pacman -S "+pkgs[0])
	}
	if pkgs[1] != "" {
		lines = append(lines, "# Debian: sudo apt install "+pkgs[1])
	}
	if pkgs[2] != "" {
		lines = append(lines, "# Fedora: sudo dnf install "+pkgs[2])
	}
	return strings.Join(lines, "\n")
}

func cmdOutput(bin string, args ...string) string {
	out, _ := exec.Command(bin, args...).CombinedOutput()
	return strings.TrimSpace(string(out))
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx != -1 {
		return s[:idx]
	}
	return s
}

// resolvedDropIn is kept for backward-compat detection only.
const resolvedDropIn = "/etc/systemd/resolved.conf.d/chauffeur.conf"

// resolvedNetworkFile is a systemd-networkd .network file for the loopback
// interface. Networkd registers DNS config for lo with systemd-resolved; since
// lo is always up, resolved never considers the .test scope "offline".
// Note: resolvectl dns lo requires networkd (org.freedesktop.network1 D-Bus).
const resolvedNetworkFile = "/etc/systemd/network/99-chauffeur-lo.network"

// isResolvedActive returns true when systemd-resolved is running.
func isResolvedActive() bool {
	out, err := exec.Command("systemctl", "is-active", "systemd-resolved").Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

// isResolvedInNSS returns true when glibc will route hostname lookups through
// systemd-resolved. This happens when nsswitch.conf has "resolve" in the hosts
// line OR when /etc/resolv.conf points at the systemd-resolved stub (127.0.0.53).
// In either case curl/getaddrinfo bypasses dnsmasq and uses resolved's own upstream.
func isResolvedInNSS() bool {
	// Check nsswitch.conf hosts line for "resolve" NSS module.
	data, err := os.ReadFile("/etc/nsswitch.conf")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "hosts:") &&
				strings.Contains(line, "resolve") {
				return true
			}
		}
	}
	// Fallback: resolv.conf stub listener.
	target, err2 := os.Readlink("/etc/resolv.conf")
	if err2 == nil && (strings.Contains(target, "systemd") || strings.Contains(target, "stub")) {
		return true
	}
	rc, _ := os.ReadFile("/etc/resolv.conf")
	return strings.Contains(string(rc), "127.0.0.53")
}

// resolvedLoFix returns commands to bind .test DNS to the loopback interface via
// systemd-networkd. Networkd registers DNS config for lo with systemd-resolved;
// since lo is always up, the .test scope is never considered "offline".
// Also cleans up the failed chauffeur-dns-route.service if present.
func resolvedLoFix() string {
	netFile := resolvedNetworkFile
	content := "[Match]\\nName=lo\\n\\n[Network]\\nDNS=127.0.0.1\\nDomains=~test\\n"
	return "# Clean up previous (failed) approach if present\n" +
		"sudo systemctl disable --now chauffeur-dns-route.service 2>/dev/null; sudo rm -f /etc/systemd/system/chauffeur-dns-route.service; sudo systemctl daemon-reload\n" +
		"# Configure systemd-networkd to register .test DNS for loopback\n" +
		"sudo mkdir -p /etc/systemd/network\n" +
		"printf '" + content + "' | sudo tee " + netFile + "\n" +
		"sudo systemctl enable --now systemd-networkd\n" +
		"sudo systemctl restart systemd-resolved"
}

// isNSSResolveFallthrough returns true when the nsswitch.conf hosts line allows
// NSS TRYAGAIN (systemd-resolved offline) to fall through to the `dns` module.
// With [!UNAVAIL=return], TRYAGAIN stops the lookup. With [NOTFOUND=return],
// TRYAGAIN falls through to `dns` which uses resolv.conf → dnsmasq → .test works.
func isNSSResolveFallthrough() bool {
	data, _ := os.ReadFile("/etc/nsswitch.conf")
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "hosts:") && strings.Contains(line, "resolve") {
			// Bad: [!UNAVAIL=return] blocks TRYAGAIN
			if strings.Contains(line, "resolve [!UNAVAIL=return]") {
				return false
			}
			// Good: [NOTFOUND=return] allows TRYAGAIN to fall through
			return true
		}
	}
	return false
}

// isResolvedDropInConfigured returns true when the systemd-networkd .network file
// for loopback exists and networkd is active (so resolved has a permanent lo-link
// DNS scope that is never considered offline).
func isResolvedDropInConfigured() bool {
	if _, err := os.Stat(resolvedNetworkFile); err != nil {
		return false
	}
	out, err := exec.Command("systemctl", "is-active", "systemd-networkd").Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

func isNMManagingDnsmasq() bool {
	entries, err := os.ReadDir("/etc/NetworkManager/conf.d")
	if err != nil {
		return false
	}
	for _, e := range entries {
		data, _ := os.ReadFile("/etc/NetworkManager/conf.d/" + e.Name())
		if strings.Contains(string(data), "dns=dnsmasq") {
			return true
		}
	}
	return false
}

func doctorAutoFix(fixes []checkResult) {
	fmt.Printf("  %s\n", lib.Bold("Auto-Fix"))
	fmt.Println()

	for _, r := range fixes {
		tag := lib.Red("✗")
		if r.warn {
			tag = lib.Yellow("⚠")
		}
		fmt.Printf("  %s  %s\n", tag, lib.Bold(r.name))

		cmds := doctorFixCmds(r.fix)
		anyFailed := false
		for _, cmd := range cmds {
			autoCmd := doctorAutoFixCmd(cmd)
			fmt.Printf("       %s %s\n", lib.Gray("$"), lib.Cyan(autoCmd))
			out, err := exec.Command("sh", "-c", autoCmd).CombinedOutput()
			output := strings.TrimSpace(string(out))
			if err != nil {
				fmt.Printf("       %s %s\n", lib.Red("✗"), lib.Red(err.Error()))
				if output != "" {
					for _, line := range strings.Split(output, "\n") {
						fmt.Printf("         %s\n", lib.Gray(line))
					}
				}
				anyFailed = true
				continue // continue with other commands
			}
			if output != "" {
				for _, line := range strings.Split(output, "\n") {
					fmt.Printf("         %s\n", lib.Gray(line))
				}
			}
		}
		if !anyFailed {
			fmt.Printf("       %s done\n", lib.Green("✓"))
		}
		fmt.Println()
	}
}

func doctorAutoFixCmd(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "sudo pacman -S ") && !strings.Contains(cmd, "--noconfirm") {
		return strings.Replace(cmd, "sudo pacman -S ", "sudo pacman -S --noconfirm ", 1)
	}
	return cmd
}

// doctorFixCmds splits a fix string into runnable commands, skipping comments and blanks.
func doctorFixCmds(fix string) []string {
	var cmds []string
	for _, line := range strings.Split(strings.TrimSpace(fix), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cmds = append(cmds, line)
	}
	return cmds
}

func doctorHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf doctor — environment health check"))
	fmt.Printf("  %s\n\n", lib.Gray("Validates system dependencies, PHP build libraries, SSL, DNS, and network."))
	fmt.Printf("  %-26s  %s\n", "--check-deps", lib.Gray("Check system build tools (gcc, make, curl, …)"))
	fmt.Printf("  %-26s  %s\n", "--check-php", lib.Gray("Check PHP build libraries (libzip, libjpeg, …)"))
	fmt.Printf("  %-26s  %s\n", "--check-ssl", lib.Gray("Check openssl and mkcert"))
	fmt.Printf("  %-26s  %s\n", "--check-network", lib.Gray("Check port availability and iptables forwarding"))
	fmt.Printf("  %-26s  %s\n", "--check-dns", lib.Gray("Check dnsmasq config and .test resolution"))
	fmt.Printf("  %-26s  %s\n", "--fix", lib.Gray("Print commands to fix failed checks"))
	fmt.Printf("  %-26s  %s\n", "--auto-fix", lib.Gray("Execute fix commands (may prompt for sudo)"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Gray("Without flags, all checks are run."))
	fmt.Println()
	return nil
}

func doctorTestResolution() bool {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	addrs, err := resolver.LookupHost(ctx, "chauffeur-probe.test")
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if a == "127.0.0.1" {
			return true
		}
	}
	return false
}
