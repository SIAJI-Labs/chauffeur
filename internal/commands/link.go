package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── chauf link ────────────────────────────────────────────────────────────────

func RunLink(args []string) error {
	flags := flag.NewFlagSet("link", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf link — register a project directory", "chauf link [path] [--php <version>] [--site <domain>] [--secure] [--dedicated-fpm] [--name <slug>] [--alias <domain>]")

	phpFlag := flags.String("php", "", "PHP version for this project")
	secureFlag := flags.Bool("secure", false, "Enable SSL from the start")
	dedicatedFPM := flags.Bool("dedicated-fpm", false, "Use a dedicated PHP-FPM pool")
	nameFlag := flags.String("name", "", "Custom slug (required when two directories share the same name)")
	siteFlag := flags.String("site", "", "Custom primary domain (e.g. myapp.test)")
	var aliases []string
	flags.Func("alias", "Add alias domain (repeatable)", func(v string) error {
		aliases = append(aliases, v)
		return nil
	})

	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()

	// Resolve the target directory (CWD by default)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	dirPath := cwd
	if flags.NArg() > 0 {
		dirPath = flags.Arg(0)
	}

	// Validate directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("directory not found: %s", dirPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dirPath)
	}

	// Resolve PHP version
	phpVersion := cfg.PHP.DefaultVersion
	if *phpFlag != "" {
		mm := installers.MajorMinor(*phpFlag)
		if mm == "" {
			return fmt.Errorf("invalid PHP version: %q", *phpFlag)
		}
		phpVersion = mm
	}

	// Validate the PHP version is installed
	inst, _ := installers.NewPHPInstaller(phpVersion, installers.BuildOpts{})
	if !inst.IsInstalled() {
		lib.Warn(fmt.Sprintf("PHP %s is not installed.", phpVersion))
		lib.Info(lib.Gray("Install it first:  chauf install php " + phpVersion))
		return nil
	}

	// Generate slug and primary domain
	slug := projects.GenerateSlug(dirPath)
	if slug == "" {
		return fmt.Errorf("could not derive a slug from directory name: %s", dirPath)
	}
	if *nameFlag != "" {
		custom := projects.GenerateSlug(*nameFlag)
		if custom == "" {
			return fmt.Errorf("invalid --name %q", *nameFlag)
		}
		slug = custom
	}
	domain := projects.DomainFromSlug(slug, cfg.DNS.TLD)

	// --site overrides the primary domain; if --name was not also given,
	// derive the slug from the site name so the project identifier matches.
	if *siteFlag != "" {
		site := *siteFlag
		if !strings.HasSuffix(site, "."+cfg.DNS.TLD) {
			site = site + "." + cfg.DNS.TLD
		}
		if !projects.IsValidDomain(site) {
			return fmt.Errorf("invalid --site domain %q — must be a valid *.%s domain", *siteFlag, cfg.DNS.TLD)
		}
		domain = site
		if *nameFlag == "" {
			base := strings.TrimSuffix(site, "."+cfg.DNS.TLD)
			if derived := projects.GenerateSlug(base); derived != "" {
				slug = derived
			}
		}
	}

	// Check if this path is already linked
	existing, err := projects.FindByPath(root, dirPath)
	if err != nil {
		return err
	}

	// Detect slug collision: same slug, different path (two dirs with the same name)
	if existing == nil {
		if collision, err := projects.Load(root, slug); err == nil && collision.Path != dirPath {
			return fmt.Errorf(
				"slug %q is already used by a different project at %s\n\n  Use a custom name:\n    chauf link --name <unique-name>",
				slug, collision.Path,
			)
		}
	}

	var p *projects.Project
	isUpdate := existing != nil

	if isUpdate {
		p = existing
		// Apply flag overrides to existing project
		p.PHPVersion = phpVersion
		p.FPM.Dedicated = *dedicatedFPM
		if *secureFlag {
			p.SSL = true
		}
		// Append any new aliases (skip duplicates)
		for _, a := range aliases {
			if !sliceContains(p.Aliases, a) {
				p.Aliases = append(p.Aliases, a)
			}
		}
	} else {
		// New project: validate aliases and domain conflicts
		if err := validateDomainUnused(root, domain, ""); err != nil {
			return err
		}
		for _, a := range aliases {
			if !projects.IsValidDomain(a) {
				return fmt.Errorf("invalid alias domain %q — must match *.test", a)
			}
			if err := validateDomainUnused(root, a, ""); err != nil {
				return err
			}
		}

		sockPath := ""
		if *dedicatedFPM {
			sockPath = root + "/projects/" + slug + "/php-fpm.sock"
		}

		p = &projects.Project{
			Slug:        slug,
			Path:        dirPath,
			Domain:      domain,
			Aliases:     aliases,
			PHPVersion:  phpVersion,
			SSL:         *secureFlag,
			FPM:         projects.FPMConfig{Dedicated: *dedicatedFPM, Socket: sockPath},
			ProjectType: projects.Detect(dirPath),
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Validate any newly added aliases
	if isUpdate {
		for _, a := range aliases {
			if !projects.IsValidDomain(a) {
				return fmt.Errorf("invalid alias domain %q — must match *.test", a)
			}
			if err := validateDomainUnused(root, a, p.Slug); err != nil {
				return err
			}
		}
	}

	// Generate and write nginx config
	if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
		return fmt.Errorf("write nginx config: %w", err)
	}

	// Symlink into sites-enabled
	if err := projects.EnableNginxSite(p, root); err != nil {
		return fmt.Errorf("enable nginx site: %w", err)
	}

	// Handle SSL
	if p.SSL {
		if err := ensureSSL(p, root); err != nil {
			return err
		}
	}

	// Dedicated FPM pool config
	if p.FPM.Dedicated {
		if err := writeDedicatedFPMConfig(p, root); err != nil {
			return fmt.Errorf("write dedicated FPM config: %w", err)
		}
	}

	// Save project config
	if err := projects.Save(p, root); err != nil {
		return fmt.Errorf("save project config: %w", err)
	}

	// Reload nginx if running
	if projects.IsNginxRunning(root) {
		_ = projects.ReloadNginx(root)
	}

	// ── Output ──────────────────────────────────────────────────────────────────

	fmt.Println()
	if isUpdate {
		lib.Success("Project updated")
	} else {
		lib.Success("Project linked")
	}
	fmt.Println()

	scheme := "http"
	port := cfg.Nginx.HTTPPort
	if p.SSL {
		scheme = "https"
		port = cfg.Nginx.HTTPSPort
	}
	lib.Pair("Domain", fmt.Sprintf("%s://%s:%d", scheme, p.Domain, port))
	lib.Pair("Path", p.Path)
	lib.Pair("PHP", phpFPMLabel(p))
	lib.Pair("Type", titleCase(string(p.ProjectType)))

	if len(p.Aliases) > 0 {
		lib.Pair("Aliases", strings.Join(p.Aliases, ", "))
	}
	fmt.Println()

	if !projects.IsNginxRunning(root) {
		lib.Info(lib.Gray("Start services:    chauf start"))
	}
	if !p.SSL {
		lib.Info(lib.Gray("Enable SSL:        chauf secure"))
	}
	fmt.Println()

	return nil
}

// ── chauf links ───────────────────────────────────────────────────────────────

func RunLinks(args []string) error {
	flags := flag.NewFlagSet("links", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf links — list registered projects", "chauf links [--detail] [--project <name>]")
	detail := flags.Bool("detail", false, "Show paths, socket paths, config paths")
	projectFlag := flags.String("project", "", "Show detail for a specific project by name")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()

	fmt.Println()

	// ── Single project detail ─────────────────────────────────────────────────
	if *projectFlag != "" {
		return showProjectDetail(root, cfg, *projectFlag)
	}

	all, err := projects.ListAll(root)
	if err != nil {
		return err
	}

	if len(all) == 0 {
		lib.Info("No projects registered.")
		fmt.Println()
		lib.Info(lib.Gray("Register one with:  chauf link"))
		fmt.Println()
		return nil
	}

	if *detail {
		for _, p := range all {
			printProjectDetail(p, root, cfg)
			fmt.Println()
		}
		return nil
	}

	// ── Tabular output ────────────────────────────────────────────────────────
	const domW = 32
	header := fmt.Sprintf(" %-20s  %-*s  %-5s  %-9s  %s",
		"Project", domW, "Domain", "PHP", "FPM", "SSL")
	sep := strings.Repeat("─", len(header))
	fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

	for _, p := range all {
		scheme := schemeLabel(p.SSL)
		fpmMode := "shared"
		if p.FPM.Dedicated {
			fpmMode = "dedicated"
		}
		fmt.Printf(" %-20s  %-*s  %-5s  %-9s  %s\n",
			p.Slug, domW, p.Domain, p.PHPVersion, fpmMode, scheme)
		for _, a := range p.Aliases {
			fmt.Printf(" %-20s  %-*s  %-5s  %-9s  %s\n",
				"", domW, "↳ "+a, "", "", lib.Gray("alias"))
		}
	}

	fmt.Println()
	return nil
}

func showProjectDetail(root string, cfg workspace.Config, name string) error {
	// Exact match first
	p, err := projects.Load(root, name)
	if err != nil {
		// Partial match: collect candidates
		all, lerr := projects.ListAll(root)
		if lerr != nil {
			return lerr
		}
		var matches []*projects.Project
		for _, candidate := range all {
			if strings.Contains(candidate.Slug, name) {
				matches = append(matches, candidate)
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("no project found matching %q", name)
		}
		if len(matches) == 1 {
			p = matches[0]
		} else {
			lib.Warn(fmt.Sprintf("Multiple projects match %q — be more specific:", name))
			fmt.Println()
			for _, m := range matches {
				lib.Info(fmt.Sprintf("  %s  %s", lib.Bold(m.Slug), lib.Gray(m.Path)))
			}
			fmt.Println()
			return nil
		}
	}
	printProjectDetail(p, root, cfg)
	fmt.Println()
	return nil
}

func printProjectDetail(p *projects.Project, root string, cfg workspace.Config) {
	home, _ := os.UserHomeDir()
	short := func(s string) string { return strings.Replace(s, home, "~", 1) }

	scheme := "http"
	port := cfg.Nginx.HTTPPort
	if p.SSL {
		scheme = "https"
		port = cfg.Nginx.HTTPSPort
	}

	sep := strings.Repeat("─", 46)
	fmt.Printf("  %s\n  %s\n", lib.Bold(p.Slug), sep)
	lib.Pair("  Domain", fmt.Sprintf("%s://%s:%d", scheme, p.Domain, port))
	for _, a := range p.Aliases {
		lib.Pair("", fmt.Sprintf("  ↳ %s", lib.Gray(a)))
	}
	lib.Pair("  Path", short(p.Path))
	lib.Pair("  PHP", phpFPMLabel(p))
	lib.Pair("  Type", titleCase(string(p.ProjectType)))
	lib.Pair("  SSL", schemeLabel(p.SSL))
	if p.SSL {
		lib.Pair("  Cert", short(root+"/nginx/certs/"+p.Domain+".crt"))
	}
	lib.Pair("  Socket", short(p.FPMSocketPath(root)))
	lib.Pair("  Nginx", short(root+"/nginx/etc/sites-available/"+p.Slug+".conf"))
	lib.Pair("  Config", short(p.ConfigPath(root)))

	nginxRunning := pidFileRunning(root + "/nginx/logs/nginx.pid")
	fpmRunning := pidFileRunning(root + "/php/" + p.PHPVersion + "/runtime/php-fpm/php-fpm.pid")
	lib.Pair("  Services", fmt.Sprintf("nginx %s  /  php-fpm %s %s",
		serviceStatus(nginxRunning), p.PHPVersion, serviceStatus(fpmRunning)))
}

// ── chauf unlink ──────────────────────────────────────────────────────────────

func RunUnlink(args []string) error {
	flags := flag.NewFlagSet("unlink", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf unlink — unregister a project", "chauf unlink [path] [--alias <domain>] [--all] [--yes]")
	aliasFlag := flags.String("alias", "", "Remove a specific alias domain only")
	allFlag := flags.Bool("all", false, "Remove all aliases then unlink the project")
	yesFlag := flags.Bool("yes", false, "Skip confirmation prompt")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()

	// Resolve target directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	dirPath := cwd
	if flags.NArg() > 0 {
		dirPath = flags.Arg(0)
	}

	p, err := projects.FindByPath(root, dirPath)
	if err != nil {
		return err
	}
	if p == nil {
		lib.Warn("No project registered for this directory.")
		lib.Info(lib.Gray("Register with:  chauf link"))
		return nil
	}

	// ── Alias-only removal ────────────────────────────────────────────────────
	if *aliasFlag != "" {
		return removeAlias(p, root, *aliasFlag)
	}

	// ── Full unlink (optionally remove all aliases first) ──────────────────────
	if !*yesFlag {
		fmt.Println()
		lib.Info(fmt.Sprintf("Unlink project %s?", lib.Bold(p.Slug)))
		lib.Pair("  Domain", p.Domain)
		if len(p.Aliases) > 0 {
			lib.Pair("  Aliases", strings.Join(p.Aliases, ", "))
		}
		lib.Pair("  SSL", fmt.Sprintf("%v", p.SSL))
		fmt.Println()
		fmt.Print("  Confirm [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		resp := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if resp != "y" && resp != "yes" {
			lib.Info("Cancelled.")
			return nil
		}
	}

	// Remove nginx configs
	if err := projects.DisableNginxSite(p, root); err != nil {
		lib.Warn("Could not remove sites-enabled symlink: " + err.Error())
	}
	if err := projects.RemoveNginxConfig(p, root); err != nil {
		lib.Warn("Could not remove sites-available config: " + err.Error())
	}

	// Remove project dir
	if err := projects.Delete(root, p.Slug); err != nil {
		return fmt.Errorf("remove project config: %w", err)
	}

	// Reload nginx
	if projects.IsNginxRunning(root) {
		_ = projects.ReloadNginx(root)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Project %s unlinked", lib.Bold(p.Slug)))
	fmt.Println()

	_ = *allFlag // --all handled implicitly (full unlink removes everything)
	return nil
}

// ── chauf secure ─────────────────────────────────────────────────────────────

func RunSecure(args []string) error {
	flags := flag.NewFlagSet("secure", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf secure — enable HTTPS for a project", "chauf secure [--project <path>]")
	projectPath := flags.String("project", "", "Target project by path")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()

	dirPath := *projectPath
	if dirPath == "" {
		var err error
		dirPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	p, err := projects.FindByPath(root, dirPath)
	if err != nil {
		return err
	}
	if p == nil {
		lib.Warn("No project registered for this directory.")
		lib.Info(lib.Gray("Register with:  chauf link"))
		return nil
	}

	if !projects.MkcertInstalled() {
		lib.Error("mkcert is not installed.")
		lib.Info("Install it from: https://github.com/FiloSottile/mkcert")
		lib.Info("Then run:  mkcert -install")
		return nil
	}

	if !projects.MkcertCAInstalled() {
		lib.Warn("mkcert CA is not installed in your system trust store.")
		lib.Info(lib.Gray("Run:  mkcert -install"))
		return nil
	}

	fmt.Println()
	lib.Info(fmt.Sprintf("Generating certificate for %s...", strings.Join(p.AllDomains(), ", ")))

	if err := projects.RunMkcert(root, p.Domain, p.AllDomains()); err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}

	p.SSL = true
	if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
		return fmt.Errorf("write nginx config: %w", err)
	}
	if err := projects.Save(p, root); err != nil {
		return fmt.Errorf("save project config: %w", err)
	}
	if projects.IsNginxRunning(root) {
		_ = projects.ReloadNginx(root)
	}

	fmt.Println()
	lib.Success("SSL enabled")
	fmt.Println()
	lib.Pair("Certificate", root+"/nginx/certs/"+p.Domain+".crt")
	lib.Pair("Domains", strings.Join(p.AllDomains(), ", "))
	fmt.Println()
	lib.Info(fmt.Sprintf("Visit: https://%s:%d", p.Domain, cfg.Nginx.HTTPSPort))
	fmt.Println()

	return nil
}

// ── chauf unsecure ────────────────────────────────────────────────────────────

func RunUnsecure(args []string) error {
	flags := flag.NewFlagSet("unsecure", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf unsecure — disable HTTPS for a project", "chauf unsecure [--project <path>]")
	projectPath := flags.String("project", "", "Target project by path")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()

	dirPath := *projectPath
	if dirPath == "" {
		var err error
		dirPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	p, err := projects.FindByPath(root, dirPath)
	if err != nil {
		return err
	}
	if p == nil {
		lib.Warn("No project registered for this directory.")
		return nil
	}

	if !p.SSL {
		lib.Info("SSL is not enabled for this project.")
		return nil
	}

	p.SSL = false
	if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
		return fmt.Errorf("write nginx config: %w", err)
	}
	if err := projects.Save(p, root); err != nil {
		return fmt.Errorf("save project config: %w", err)
	}
	if projects.IsNginxRunning(root) {
		_ = projects.ReloadNginx(root)
	}

	fmt.Println()
	lib.Success("SSL disabled")
	lib.Info(lib.Gray(fmt.Sprintf("Certificate files at %s/nginx/certs/%s.{crt,key} were not removed.", root, p.Domain)))
	fmt.Println()

	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func phpFPMLabel(p *projects.Project) string {
	mode := "shared FPM"
	if p.FPM.Dedicated {
		mode = "dedicated FPM"
	}
	return fmt.Sprintf("%s (%s)", p.PHPVersion, mode)
}

func validateDomainUnused(workspaceRoot, domain, skipSlug string) error {
	owner, err := projects.IsDomainInUse(workspaceRoot, domain, skipSlug)
	if err != nil {
		return err
	}
	if owner != "" {
		return fmt.Errorf(
			"domain %q is already used by project %q\n\n  Remove it first:\n    chauf unlink --alias %s  (in the %s directory)",
			domain, owner, domain, owner,
		)
	}
	return nil
}

func removeAlias(p *projects.Project, root string, alias string) error {
	wCfg := workspace.Load()

	idx := -1
	for i, a := range p.Aliases {
		if a == alias {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("alias %q is not registered on project %q", alias, p.Slug)
	}

	p.Aliases = append(p.Aliases[:idx], p.Aliases[idx+1:]...)

	if err := projects.WriteNginxConfig(p, root, wCfg.Nginx.HTTPPort, wCfg.Nginx.HTTPSPort); err != nil {
		return err
	}
	if p.SSL {
		if err := projects.RunMkcert(root, p.Domain, p.AllDomains()); err != nil {
			lib.Warn("Could not regenerate SSL certificate: " + err.Error())
		}
	}
	if err := projects.Save(p, root); err != nil {
		return err
	}
	if projects.IsNginxRunning(root) {
		_ = projects.ReloadNginx(root)
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Alias %s removed from %s", lib.Bold(alias), lib.Bold(p.Slug)))
	fmt.Println()
	return nil
}

func ensureSSL(p *projects.Project, root string) error {
	if !projects.MkcertInstalled() {
		lib.Warn("mkcert not found — skipping SSL setup.")
		lib.Info(lib.Gray("Install mkcert and run:  chauf secure"))
		p.SSL = false
		return nil
	}
	if !projects.MkcertCAInstalled() {
		lib.Warn("mkcert CA not installed — skipping SSL setup.")
		lib.Info(lib.Gray("Run:  mkcert -install"))
		p.SSL = false
		return nil
	}
	return projects.RunMkcert(root, p.Domain, p.AllDomains())
}

func writeDedicatedFPMConfig(p *projects.Project, root string) error {
	user := os.Getenv("USER")
	if user == "" {
		user = "nobody"
	}

	tpl := fmt.Sprintf(`[global]
pid = %s/projects/%s/php-fpm.pid
error_log = %s/projects/%s/logs/php-fpm.log

[%s]
listen = %s
listen.owner = %s
listen.group = %s
listen.mode = 0660

user = %s
group = %s

pm = dynamic
pm.max_children = 5
pm.start_servers = 1
pm.min_spare_servers = 1
pm.max_spare_servers = 3
`,
		root, p.Slug,
		root, p.Slug,
		p.Slug,
		p.FPM.Socket,
		user, user,
		user, user,
	)

	dir := root + "/projects/" + p.Slug
	if err := os.MkdirAll(dir+"/logs", 0755); err != nil {
		return err
	}
	return os.WriteFile(dir+"/php-fpm.conf", []byte(tpl), 0644)
}

func sliceContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
