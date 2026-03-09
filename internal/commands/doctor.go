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

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/services"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── types ──────────────────────────────────────────────────────────────────────

type checkResult struct {
	name   string
	ok     bool   // true = pass
	warn   bool   // true = warning (not blocking)
	status string // one-line status message
	fix    string // command(s) to print when --fix is passed
}

// ── RunDoctor ─────────────────────────────────────────────────────────────────

func RunDoctor(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		return doctorHelp()
	}

	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	doDeps    := flags.Bool("check-deps", false, "Check system build dependencies")
	doPHP     := flags.Bool("check-php", false, "Check PHP build libraries")
	doSSL     := flags.Bool("check-ssl", false, "Check SSL tools")
	doNetwork := flags.Bool("check-network", false, "Check port availability and iptables")
	doDNS     := flags.Bool("check-dns", false, "Check DNS resolution for .test domains")
	doFix     := flags.Bool("fix", false, "Print fix commands for failed checks")
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
	fmt.Println()

	var all []checkResult

	if *doDeps {
		r := doctorSystemDeps()
		printDoctorSection("System Dependencies", r)
		all = append(all, r...)
	}
	if *doPHP {
		r := doctorPHPDeps()
		printDoctorSection("PHP Build Libraries", r)
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
		if !r.ok && r.fix != "" {
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

// ── check sections ─────────────────────────────────────────────────────────────

func doctorSystemDeps() []checkResult {
	type toolDef struct {
		name string
		args []string
		pkgs pkgNames
	}
	tools := []toolDef{
		{"git",        []string{"--version"}, pkgNames{"git", "git", "git"}},
		{"curl",       []string{"--version"}, pkgNames{"curl", "curl", "curl"}},
		{"tar",        []string{"--version"}, pkgNames{"tar", "tar", "tar"}},
		{"gcc",        []string{"--version"}, pkgNames{"gcc", "build-essential", "gcc"}},
		{"make",       []string{"--version"}, pkgNames{"make", "make", "make"}},
		{"pkg-config", []string{"--version"}, pkgNames{"pkgconf", "pkg-config", "pkgconf"}},
		{"autoconf",   []string{"--version"}, pkgNames{"autoconf", "autoconf", "autoconf"}},
		{"bison",      []string{"--version"}, pkgNames{"bison", "bison", "bison"}},
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
		{"libzip",     "libzip",      pkgNames{"libzip", "libzip-dev", "libzip-devel"}, false},
		{"libjpeg",    "libjpeg",     pkgNames{"libjpeg-turbo", "libjpeg-dev", "libjpeg-devel"}, false},
		{"libpng",     "libpng",      pkgNames{"libpng", "libpng-dev", "libpng-devel"}, false},
		{"freetype2",  "freetype2",   pkgNames{"freetype2", "libfreetype6-dev", "freetype-devel"}, false},
		{"libxml-2.0", "libxml2",     pkgNames{"libxml2", "libxml2-dev", "libxml2-devel"}, false},
		{"libcurl",    "libcurl",     pkgNames{"curl", "libcurl4-openssl-dev", "libcurl-devel"}, false},
		{"zlib",       "zlib",        pkgNames{"zlib", "zlib1g-dev", "zlib-devel"}, false},
		{"readline",   "readline",    pkgNames{"readline", "libreadline-dev", "readline-devel"}, false},
		{"libxslt",    "libxslt",     pkgNames{"libxslt", "libxslt1-dev", "libxslt-devel"}, false},
		{"gmp",        "gmp",         pkgNames{"gmp", "libgmp-dev", "gmp-devel"}, false},
		{"openssl",    "openssl",     pkgNames{"openssl", "libssl-dev", "openssl-devel"}, false},
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
		results = append(results, checkResult{name: name, ok: true, status: lib.Gray(ver)})
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

		// CA installed?
		caRoot := strings.TrimSpace(cmdOutput(mkcertPath, "-CAROOT"))
		_, e1 := os.Stat(filepath.Join(caRoot, "rootCA.pem"))
		_, e2 := os.Stat(filepath.Join(caRoot, "rootCA-key.pem"))
		if e1 == nil && e2 == nil {
			results = append(results, checkResult{
				name:   "mkcert CA",
				ok:     true,
				status: lib.Gray(shortenHome(caRoot)),
			})
		} else {
			results = append(results, checkResult{
				name:   "mkcert CA",
				ok:     false,
				status: "CA not installed — browsers will show SSL warnings",
				fix:    "mkcert -install",
			})
		}
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
		cmds := system.PortForwardingCommands(cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort)
		results = append(results, checkResult{
			name:   "port forwarding",
			ok:     false,
			warn:   true,
			status: fmt.Sprintf("not configured — access needs :%d / :%d", cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort),
			fix:    strings.Join(cmds, "\n"),
		})
	}

	// nginx + FPM ports
	nginxSvc := services.NewNginxService(root)
	for _, port := range []int{cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort} {
		label := fmt.Sprintf("port %d", port)
		if nginxSvc.IsRunning() || services.IsPortAvailable(port) {
			extra := "available"
			if nginxSvc.IsRunning() {
				extra = fmt.Sprintf("nginx pid %d", nginxSvc.PID())
			}
			results = append(results, checkResult{name: label, ok: true, status: lib.Gray(extra)})
		} else {
			pid, name, _ := services.FindProcessOnPort(port)
			msg := "in use by another process"
			fix := ""
			if pid > 0 {
				msg = fmt.Sprintf("in use by %s (pid %d)", name, pid)
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
	distroArch    distroType = iota
	distroDebian             // Ubuntu / Debian
	distroFedora             // Fedora / RHEL / Rocky
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
			fmt.Printf("       %s %s\n", lib.Gray("$"), lib.Cyan(cmd))
			out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			output := strings.TrimSpace(string(out))
			if err != nil {
				fmt.Printf("       %s %s\n", lib.Red("✗"), lib.Red(err.Error()))
				if output != "" {
					for _, line := range strings.Split(output, "\n") {
						fmt.Printf("         %s\n", lib.Gray(line))
					}
				}
				anyFailed = true
				break // stop on first failure within this fix block
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
