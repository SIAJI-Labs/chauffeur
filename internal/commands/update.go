package commands

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunUpdate(args []string) error {
	if len(args) == 0 {
		return updateHelp()
	}
	printRuntime(workspace.Load())
	fmt.Println()
	switch strings.ToLower(args[0]) {
	case "nginx":
		return updateNginx(args[1:])
	case "php":
		return updatePHP(args[1:])
	case "composer":
		return updateComposer(args[1:])
	case "all":
		return updateAll(args[1:])
	case "list", "check":
		return updateList(args[1:])
	case "rollback":
		return updateRollback(args[1:])
	case "--help", "-h", "help":
		return updateHelp()
	default:
		return fmt.Errorf("unknown update subcommand %q — run: chauf update --help", args[0])
	}
}

// ── shared flags ───────────────────────────────────────────────────────────────

type updateOpts struct {
	dryRun   bool
	noBackup bool
	verbose  bool
	force    bool
}

func parseUpdateFlags(name string, args []string) (updateOpts, []string, error) {
	flags := flag.NewFlagSet("update "+name, flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf update — update installed services to latest versions", "chauf update <nginx|php <version>|composer|all> [--dry-run] [--no-backup] [--force]")
	dryRun := flags.Bool("dry-run", false, "Show what would be updated without doing it")
	noBackup := flags.Bool("no-backup", false, "Skip backup before updating")
	verbose := flags.Bool("verbose", false, "Stream build output")
	flags.BoolVar(verbose, "v", false, "Stream build output")
	force := flags.Bool("force", false, "Update even if already on latest")

	flagArgs, positionals := splitArgs(args)
	if err := flags.Parse(flagArgs); err != nil {
		return updateOpts{}, nil, err
	}
	return updateOpts{
		dryRun:   *dryRun,
		noBackup: *noBackup,
		verbose:  *verbose,
		force:    *force,
	}, positionals, nil
}

// ── update nginx ───────────────────────────────────────────────────────────────

func updateNginx(args []string) error {
	if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
		return fmt.Errorf("nginx is managed by the Podman image; pull %s explicitly", chauftruntime.NginxImage)
	}
	opts, _, err := parseUpdateFlags("nginx", args)
	if err != nil {
		return err
	}

	inst := installers.NewNginxInstaller(installers.BuildOpts{})
	if !inst.IsInstalled() {
		return fmt.Errorf("nginx is not installed — run: chauf install nginx")
	}

	current := inst.InstalledVersion()
	fmt.Println()

	s := startStep("Checking latest nginx version", opts.verbose)
	latest := inst.ResolveVersion()
	s.success(latest)

	lib.Pair("Installed", current)
	lib.Pair("Latest   ", latest)
	fmt.Println()

	if !opts.force && !versionNewer(latest, current) {
		lib.Success("nginx is up to date  (" + current + ")")
		fmt.Println()
		lib.Info(lib.Gray("Use --force to reinstall the current version."))
		fmt.Println()
		return nil
	}

	if opts.dryRun {
		lib.Info(lib.Gray("(dry-run) would update nginx  " + current + " → " + latest))
		fmt.Println()
		return nil
	}

	root := workspace.Root()
	if !opts.noBackup {
		if err := backupService(root, "nginx", current, filepath.Join(root, "nginx", "sbin")); err != nil {
			lib.Warn("Backup failed: " + err.Error())
		}
	}

	lib.Info(fmt.Sprintf("Updating nginx  %s → %s", current, lib.Bold(latest)))
	fmt.Println()
	return installNginx(installers.BuildOpts{Force: true, Verbose: opts.verbose})
}

// ── update php ─────────────────────────────────────────────────────────────────

func updatePHP(args []string) error {
	opts, positionals, err := parseUpdateFlags("php", args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		return fmt.Errorf("usage: chauf update php <version>  (e.g. 8.3)")
	}

	mm := installers.MajorMinor(positionals[0])
	if mm == "" {
		return fmt.Errorf("invalid PHP version: %q", positionals[0])
	}
	if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
		return updatePodmanPHP(mm, opts)
	}

	inst, err := installers.NewPHPInstaller(mm, installers.BuildOpts{})
	if err != nil {
		return err
	}
	if !inst.IsInstalled() {
		return fmt.Errorf("PHP %s is not installed — run: chauf install php %s", mm, mm)
	}

	current := inst.InstalledVersion()
	fmt.Println()

	s := startStep("Checking latest PHP "+mm+" version", opts.verbose)
	latest, err := inst.ResolveVersion()
	if err != nil {
		s.fail(err.Error())
		return err
	}
	s.success(latest)

	lib.Pair("Installed", current)
	lib.Pair("Latest   ", latest)
	fmt.Println()

	if !opts.force && !versionNewer(latest, current) {
		lib.Success("PHP " + mm + " is up to date  (" + current + ")")
		fmt.Println()
		lib.Info(lib.Gray("Use --force to reinstall the current version."))
		fmt.Println()
		return nil
	}

	if opts.dryRun {
		lib.Info(lib.Gray("(dry-run) would update PHP " + mm + "  " + current + " → " + latest))
		fmt.Println()
		return nil
	}

	root := workspace.Root()
	if !opts.noBackup {
		phpBinDir := filepath.Join(root, "php", mm, "bin")
		phpSbinDir := filepath.Join(root, "php", mm, "sbin")
		if err := backupServiceDirs(root, "php-"+mm, current, phpBinDir, phpSbinDir); err != nil {
			lib.Warn("Backup failed: " + err.Error())
		}
	}

	lib.Info(fmt.Sprintf("Updating PHP %s  %s → %s", mm, current, lib.Bold(latest)))
	fmt.Println()
	return installPHP(mm, installers.BuildOpts{Force: true, Verbose: opts.verbose})
}

func updatePodmanPHP(version string, opts updateOpts) error {
	if !isInstalledPHPVersion(version) {
		return fmt.Errorf("PHP %s Podman image is not installed — run: chauf install php %s", version, version)
	}
	if opts.dryRun {
		lib.Info(lib.Gray("(dry-run) would rebuild Podman PHP image " + version))
		return nil
	}
	lib.Info(fmt.Sprintf("Rebuilding Podman PHP %s image", version))
	return installPodmanPHP(version, true, opts.verbose)
}

func podmanPHPImages() []string {
	versions := make([]string, 0)
	for _, version := range installers.SupportedPHPVersions {
		if isInstalledPHPVersion(version) {
			versions = append(versions, version)
		}
	}
	return versions
}

// ── update composer ────────────────────────────────────────────────────────────

func updateComposer(args []string) error {
	if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
		return fmt.Errorf("Composer is managed inside Podman PHP images; rebuild the selected PHP image instead")
	}
	opts, _, err := parseUpdateFlags("composer", args)
	if err != nil {
		return err
	}

	inst := installers.NewComposerInstaller(installers.BuildOpts{})
	if !inst.IsInstalled() {
		return fmt.Errorf("Composer is not installed — run: chauf install composer")
	}

	current := inst.InstalledVersion()
	fmt.Println()

	s := startStep("Checking latest Composer version", opts.verbose)
	latest := inst.ResolveVersion()
	s.success(latest)

	lib.Pair("Installed", current)
	lib.Pair("Latest   ", latest)
	fmt.Println()

	if !opts.force && !versionNewer(latest, current) {
		lib.Success("Composer is up to date  (" + current + ")")
		fmt.Println()
		lib.Info(lib.Gray("Use --force to reinstall the current version."))
		fmt.Println()
		return nil
	}

	if opts.dryRun {
		lib.Info(lib.Gray("(dry-run) would update Composer  " + current + " → " + latest))
		fmt.Println()
		return nil
	}

	root := workspace.Root()
	if !opts.noBackup {
		if err := backupFile(root, "composer", current, inst.PharPath()); err != nil {
			lib.Warn("Backup failed: " + err.Error())
		}
	}

	lib.Info(fmt.Sprintf("Updating Composer  %s → %s", current, lib.Bold(latest)))
	fmt.Println()
	return installComposer(installers.BuildOpts{Force: true, Verbose: opts.verbose})
}

// ── update all ─────────────────────────────────────────────────────────────────

func updateAll(args []string) error {
	opts, _, err := parseUpdateFlags("all", args)
	if err != nil {
		return err
	}
	root := workspace.Root()
	allArgs := buildFlagArgs(cleanOpts{dryRun: opts.dryRun})
	if opts.noBackup {
		allArgs = append(allArgs, "--no-backup")
	}
	if opts.verbose {
		allArgs = append(allArgs, "--verbose")
	}
	if opts.force {
		allArgs = append(allArgs, "--force")
	}

	var anyErr error

	nginxInst := installers.NewNginxInstaller(installers.BuildOpts{})
	if workspace.Load().Runtime.Engine != string(chauftruntime.EnginePodman) && nginxInst.IsInstalled() {
		if err := updateNginx(allArgs); err != nil {
			lib.Warn("nginx update failed: " + err.Error())
			anyErr = err
		}
	}

	phpVersions := installers.ListInstalledPHP(root)
	if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
		phpVersions = podmanPHPImages()
	}
	for _, mm := range phpVersions {
		phpArgs := append([]string{mm}, allArgs...)
		if err := updatePHP(phpArgs); err != nil {
			lib.Warn("PHP " + mm + " update failed: " + err.Error())
			anyErr = err
		}
	}

	compInst := installers.NewComposerInstaller(installers.BuildOpts{})
	if workspace.Load().Runtime.Engine != string(chauftruntime.EnginePodman) && compInst.IsInstalled() {
		if err := updateComposer(allArgs); err != nil {
			lib.Warn("Composer update failed: " + err.Error())
			anyErr = err
		}
	}

	return anyErr
}

// ── update list ────────────────────────────────────────────────────────────────

func updateList(args []string) error {
	flags := flag.NewFlagSet("update list", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf update list — show installed vs latest versions", "chauf update list")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	fmt.Println()
	fmt.Printf("  %s\n\n", lib.Bold("Available Updates"))

	type row struct {
		name      string
		installed string
		latest    string
		hasUpdate bool
	}
	var rows []row

	// nginx
	nginxInst := installers.NewNginxInstaller(installers.BuildOpts{})
	if workspace.Load().Runtime.Engine != string(chauftruntime.EnginePodman) && nginxInst.IsInstalled() {
		current := nginxInst.InstalledVersion()
		latest := nginxInst.ResolveVersion()
		rows = append(rows, row{"nginx", current, latest, versionNewer(latest, current)})
	}

	// PHP versions
	phpVersions := installers.ListInstalledPHP(root)
	if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
		phpVersions = podmanPHPImages()
	}
	for _, mm := range phpVersions {
		if workspace.Load().Runtime.Engine == string(chauftruntime.EnginePodman) {
			rows = append(rows, row{"php " + mm, "Podman image", "rebuild locally", false})
			continue
		}
		phpInst, err := installers.NewPHPInstaller(mm, installers.BuildOpts{})
		if err != nil {
			continue
		}
		current := phpInst.InstalledVersion()
		latest, err := phpInst.ResolveVersion()
		if err != nil {
			latest = lib.Gray("(unavailable)")
		}
		rows = append(rows, row{"php " + mm, current, latest, err == nil && versionNewer(latest, current)})
	}

	// composer
	compInst := installers.NewComposerInstaller(installers.BuildOpts{})
	if workspace.Load().Runtime.Engine != string(chauftruntime.EnginePodman) && compInst.IsInstalled() {
		current := compInst.InstalledVersion()
		latest := compInst.ResolveVersion()
		rows = append(rows, row{"composer", current, latest, versionNewer(latest, current)})
	}

	if len(rows) == 0 {
		lib.Info("No services installed.")
		fmt.Println()
		return nil
	}

	for _, r := range rows {
		icon := lib.Green("✓")
		status := lib.Gray("up to date")
		if r.hasUpdate {
			icon = lib.Yellow("↑")
			status = lib.Yellow(r.latest) + lib.Gray("  (installed: "+r.installed+")")
		} else {
			status = lib.Gray(r.installed)
		}
		fmt.Printf("  %s  %-12s  %s\n", icon, r.name, status)
	}

	fmt.Println()
	updates := 0
	for _, r := range rows {
		if r.hasUpdate {
			updates++
		}
	}
	if updates > 0 {
		lib.Info(lib.Gray(fmt.Sprintf("%d update(s) available — run: chauf update all", updates)))
	} else {
		lib.Info(lib.Gray("All services are up to date."))
	}
	fmt.Println()
	return nil
}

// ── update rollback ────────────────────────────────────────────────────────────

func updateRollback(args []string) error {
	flags := flag.NewFlagSet("update rollback", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf update rollback — restore a previous version from backup", "chauf update rollback <nginx|php <version>|composer>")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	backupsDir := filepath.Join(root, "backups")

	service := flags.Arg(0)
	if service == "" {
		return listBackups(backupsDir)
	}

	// Specific backup to restore — second arg is the backup filename (without .tar.gz)
	backupName := flags.Arg(1)
	if backupName == "" {
		return listBackupsFor(backupsDir, service)
	}

	backupPath := filepath.Join(backupsDir, backupName)
	if !strings.HasSuffix(backupPath, ".tar.gz") {
		backupPath += ".tar.gz"
	}

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not found: %s", backupPath)
	}

	// Determine restore target dir
	restoreDir, err := restoreDirForService(root, service)
	if err != nil {
		return err
	}

	fmt.Println()
	lib.Info(fmt.Sprintf("Restoring %s from %s", service, lib.Gray(shortenHome(backupPath))))
	if err := extractTarGzTo(backupPath, restoreDir); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	lib.Success("Restored — restart services: chauf restart")
	fmt.Println()
	return nil
}

func listBackups(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println()
			lib.Info("No backups found.")
			fmt.Println()
			return nil
		}
		return err
	}
	fmt.Println()
	fmt.Printf("  %s\n\n", lib.Bold("Available Backups"))
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".tar.gz") {
			continue
		}
		info, _ := e.Info()
		size := ""
		if info != nil {
			size = lib.Gray("  " + lib.FormatBytes(info.Size()))
		}
		name := strings.TrimSuffix(e.Name(), ".tar.gz")
		fmt.Printf("  %s%s\n", name, size)
	}
	fmt.Println()
	lib.Info(lib.Gray("Restore with: chauf update rollback <service> <backup-name>"))
	fmt.Println()
	return nil
}

func listBackupsFor(dir, service string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		fmt.Println()
		lib.Info("No backups for " + service)
		fmt.Println()
		return nil
	}
	fmt.Println()
	fmt.Printf("  %s\n\n", lib.Bold("Backups for "+service))
	found := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".tar.gz") {
			continue
		}
		if !strings.HasPrefix(e.Name(), service+"-") {
			continue
		}
		info, _ := e.Info()
		size := ""
		if info != nil {
			size = lib.Gray("  " + lib.FormatBytes(info.Size()))
		}
		name := strings.TrimSuffix(e.Name(), ".tar.gz")
		fmt.Printf("  %s%s\n", name, size)
		found++
	}
	if found == 0 {
		lib.Info("No backups for " + service)
	}
	fmt.Println()
	return nil
}

func restoreDirForService(root, service string) (string, error) {
	if service == "nginx" {
		return filepath.Join(root, "nginx", "sbin"), nil
	}
	if service == "composer" {
		return filepath.Join(root, "composer"), nil
	}
	if strings.HasPrefix(service, "php-") {
		mm := strings.TrimPrefix(service, "php-")
		if installers.MajorMinor(mm) == "" {
			return "", fmt.Errorf("invalid PHP version in service name %q", service)
		}
		return filepath.Join(root, "php", mm, "bin"), nil
	}
	return "", fmt.Errorf("unknown service %q — use: nginx, php-<version>, composer", service)
}

// ── backup helpers ─────────────────────────────────────────────────────────────

func backupService(root, service, version string, dirs ...string) error {
	return backupServiceDirs(root, service, version, dirs...)
}

func backupServiceDirs(root, service, version string, dirs ...string) error {
	backupsDir := filepath.Join(root, "backups")
	if err := os.MkdirAll(backupsDir, 0o755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102-1504")
	name := fmt.Sprintf("%s-%s-%s.tar.gz", service, version, ts)
	dest := filepath.Join(backupsDir, name)

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if err := tarDir(tw, dir, filepath.Base(dir)); err != nil {
			return err
		}
	}

	info, _ := os.Stat(dest)
	sz := ""
	if info != nil {
		sz = "  (" + lib.FormatBytes(info.Size()) + ")"
	}
	lib.Pair("Backup", lib.Gray(shortenHome(dest)+sz))
	return nil
}

func backupFile(root, service, version, srcPath string) error {
	backupsDir := filepath.Join(root, "backups")
	if err := os.MkdirAll(backupsDir, 0o755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102-1504")
	ext := filepath.Ext(srcPath)
	name := fmt.Sprintf("%s-%s-%s%s", service, version, ts, ext)
	dest := filepath.Join(backupsDir, name)

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	lib.Pair("Backup", lib.Gray(shortenHome(dest)))
	return nil
}

func tarDir(tw *tar.Writer, dir, prefix string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join(prefix, rel)
		if info.IsDir() {
			header.Name += "/"
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func extractTarGzTo(tarPath, destDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Strip the first path component (the prefix used in tarDir)
		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		target := filepath.Join(destDir, parts[1])
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

// ── version comparison ─────────────────────────────────────────────────────────

// versionNewer returns true if candidate is strictly newer than current.
// Compares dot-separated integer segments (e.g. "1.27.4" > "1.27.3").
func versionNewer(candidate, current string) bool {
	ca := splitVersion(candidate)
	cb := splitVersion(current)
	n := len(ca)
	if len(cb) > n {
		n = len(cb)
	}
	for i := range n {
		a, b := 0, 0
		if i < len(ca) {
			a = ca[i]
		}
		if i < len(cb) {
			b = cb[i]
		}
		if a > b {
			return true
		}
		if a < b {
			return false
		}
	}
	return false
}

func splitVersion(v string) []int {
	var parts []int
	for _, s := range strings.Split(strings.TrimPrefix(v, "v"), ".") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			break
		}
		parts = append(parts, n)
	}
	return parts
}

// ── help ───────────────────────────────────────────────────────────────────────

func updateHelp() error {
	fmt.Printf("\n%s\n\n", lib.Bold("chauf update — update installed services"))
	fmt.Printf("  %-30s  %s\n", "update list", lib.Gray("Check for available updates"))
	fmt.Printf("  %-30s  %s\n", "update nginx", lib.Gray("Update nginx to latest version"))
	fmt.Printf("  %-30s  %s\n", "update php <version>", lib.Gray("Update PHP major.minor to latest patch"))
	fmt.Printf("  %-30s  %s\n", "update composer", lib.Gray("Update Composer to latest version"))
	fmt.Printf("  %-30s  %s\n", "update all", lib.Gray("Update all installed services"))
	fmt.Printf("  %-30s  %s\n", "update rollback [service] [backup]", lib.Gray("List or restore a backup"))
	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold("Flags:"))
	fmt.Printf("  %-30s  %s\n", "--dry-run", lib.Gray("Show what would be updated"))
	fmt.Printf("  %-30s  %s\n", "--no-backup", lib.Gray("Skip backup before updating"))
	fmt.Printf("  %-30s  %s\n", "--force", lib.Gray("Update even if already on latest"))
	fmt.Printf("  %-30s  %s\n", "--verbose / -v", lib.Gray("Stream build output"))
	fmt.Println()
	return nil
}
