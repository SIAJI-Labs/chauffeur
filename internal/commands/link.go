package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

// ── chauf link ────────────────────────────────────────────────────────────────

func RunLink(args []string) error {
	flags := flag.NewFlagSet("link", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf link — configure and register a project directory", "chauf link [path] [--interactive|--no-interactive] [--type <type>] [--proxy-port <port>] [--php <version>] [--site <domain>] [--secure|--insecure] [--fpm shared|dedicated] [--dedicated-fpm] [--name <slug>] [--alias <domain>]")

	phpFlag := flags.String("php", "", "PHP version for this project")
	typeFlag := flags.String("type", "", "Project type: laravel, wordpress, php, or reverse-proxy")
	proxyPortFlag := flags.Int("proxy-port", 0, "Local port for a reverse-proxy project (default 3000)")
	interactiveFlag := flags.Bool("interactive", false, "Configure project options interactively")
	noInteractiveFlag := flags.Bool("no-interactive", false, "Never prompt for project options")
	secureFlag := flags.Bool("secure", false, "Enable SSL from the start")
	insecureFlag := flags.Bool("insecure", false, "Disable SSL explicitly")
	dedicatedFPM := flags.Bool("dedicated-fpm", false, "Use a dedicated PHP-FPM pool")
	fpmFlag := flags.String("fpm", "", "FPM mode: shared or dedicated")
	yesFlag := flags.Bool("yes", false, "Confirm the setup plan without prompting")
	nameFlag := flags.String("name", "", "Custom slug (required when two directories share the same name)")
	siteFlag := flags.String("site", "", "Custom primary domain (e.g. myapp.test)")
	var aliases []string
	flags.Func("alias", "Add alias domain (repeatable)", func(v string) error {
		aliases = append(aliases, v)
		return nil
	})

	if err := flags.Parse(normalizeLinkArgs(args)); err != nil {
		return err
	}
	if *secureFlag && *insecureFlag {
		return fmt.Errorf("--secure and --insecure cannot be used together")
	}
	if *fpmFlag != "" {
		switch strings.ToLower(*fpmFlag) {
		case "shared":
			*dedicatedFPM = false
		case "dedicated":
			*dedicatedFPM = true
		default:
			return fmt.Errorf("invalid --fpm %q; choose shared or dedicated", *fpmFlag)
		}
	}
	phpExplicit := *phpFlag != ""
	secureExplicit := false
	dedicatedExplicit := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "secure" {
			secureExplicit = true
		}
		if f.Name == "dedicated-fpm" {
			dedicatedExplicit = true
		}
	})
	dedicatedExplicit = dedicatedExplicit || *fpmFlag != ""

	root := workspace.Root()
	cfg := workspace.Load()
	printRuntime(cfg)
	fmt.Println()

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

	detectedType := projects.Detect(dirPath)
	projectType := detectedType
	if *typeFlag != "" {
		parsed, err := parseProjectType(*typeFlag)
		if err != nil {
			return err
		}
		projectType = parsed
	}
	wantsWizard := (*interactiveFlag || (!*noInteractiveFlag && tui.Interactive() && flags.NFlag() == 0)) && tui.Interactive() && !*yesFlag

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
	if existing != nil && *typeFlag == "" {
		projectType = existing.ProjectType
	}

	if projectType == projects.TypeUnknown {
		if !wantsWizard {
			return fmt.Errorf("could not detect project type; choose one with --type laravel|wordpress|php|reverse-proxy")
		}
	}
	if !wantsWizard || *typeFlag != "" {
		printProjectDetection(detectedType, projectType)
	}

	proxyPort := *proxyPortFlag
	proxyPortExplicit := *proxyPortFlag != 0
	if existing != nil && proxyPort == 0 {
		proxyPort = existing.ProxyPort
	}

	// An unqualified interactive link is the project setup wizard. Explicit
	// flags remain the scriptable path and are never overridden by prompts.
	if wantsWizard {
		if *typeFlag == "" {
			selected, err := runProjectTypeWizard(detectedType)
			if err != nil {
				return err
			}
			if selected == projects.TypeUnknown {
				return nil
			}
			projectType = selected
		}
		setup, err := runLinkWizard(dirPath, cfg, projectType, existing)
		if err != nil {
			return err
		}
		if setup.cancelled {
			return nil
		}
		if setup.php != "" {
			if !phpExplicit {
				*phpFlag = setup.php
			}
		}
		if setup.domain != "" && *siteFlag == "" {
			domain = setup.domain
		}
		if len(setup.aliases) > 0 && len(aliases) == 0 {
			aliases = append([]string(nil), setup.aliases...)
		}
		if setup.proxyPort > 0 && !proxyPortExplicit {
			proxyPort = setup.proxyPort
		}
		if !secureExplicit && !*insecureFlag {
			*secureFlag = setup.secure
		}
		if !dedicatedExplicit && *fpmFlag == "" {
			*dedicatedFPM = setup.dedicated
		}
		secureExplicit = secureExplicit || *insecureFlag
		dedicatedExplicit = dedicatedExplicit || *fpmFlag != ""
	}
	if *insecureFlag {
		*secureFlag = false
		secureExplicit = true
	}
	selectedSSL := *secureFlag
	selectedDedicated := *dedicatedFPM
	if existing != nil {
		if !secureExplicit {
			selectedSSL = existing.SSL
		}
		if !dedicatedExplicit {
			selectedDedicated = existing.FPM.Dedicated
		}
	}
	if projectType == projects.TypeReverseProxy {
		if proxyPort == 0 {
			proxyPort = projects.DefaultProxyPortFor(dirPath)
		}
		if proxyPort < 1 || proxyPort > 65535 {
			return fmt.Errorf("invalid --proxy-port %d; must be between 1 and 65535", proxyPort)
		}
	} else {
		if *proxyPortFlag != 0 {
			return fmt.Errorf("--proxy-port requires --type reverse-proxy or a detected JavaScript project")
		}
		proxyPort = 0
	}
	if projectType == projects.TypeReverseProxy {
		if *phpFlag != "" {
			return fmt.Errorf("--php is not used by reverse-proxy projects")
		}
		if *dedicatedFPM {
			return fmt.Errorf("--dedicated-fpm is not used by reverse-proxy projects")
		}
	}

	phpVersion := ""
	if projectType != projects.TypeReverseProxy {
		phpVersion = cfg.PHP.DefaultVersion
		if existing != nil && *phpFlag == "" {
			phpVersion = existing.PHPVersion
		}
		if *phpFlag != "" {
			mm := installers.MajorMinor(*phpFlag)
			if mm == "" {
				return fmt.Errorf("invalid PHP version: %q", *phpFlag)
			}
			phpVersion = mm
		}
		constraint := projects.PHPConstraint(dirPath)
		if constraint != "" && !projects.PHPVersionSatisfies(phpVersion, constraint) {
			return fmt.Errorf("PHP %s does not satisfy composer.json constraint %q", phpVersion, constraint)
		}

		// Validate the PHP version is installed for PHP-backed projects only.
		if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
			result, runErr := (chauftruntime.ExecRunner{}).Run(context.Background(), "image", "exists", chauftruntime.PHPImage(phpVersion))
			if runErr != nil || result.ExitCode != 0 {
				return fmt.Errorf("PHP %s Podman image is unavailable: %s", phpVersion, chauftruntime.PHPImage(phpVersion))
			}
		} else {
			inst, _ := installers.NewPHPInstaller(phpVersion, installers.BuildOpts{})
			if !inst.IsInstalled() {
				lib.Warn(fmt.Sprintf("PHP %s is not installed.", phpVersion))
				lib.Info(lib.Gray("Install it first:  chauf install php " + phpVersion))
				return nil
			}
		}
	}

	phpChoices := make([]projects.RuntimeChoice, 0)
	if projectType != projects.TypeReverseProxy {
		available := installedPHPForRuntime(installers.SupportedPHPVersions, cfg.Runtime.Engine)
		for _, version := range installers.SupportedPHPVersions {
			if !available[version] {
				continue
			}
			phpChoices = append(phpChoices, projects.RuntimeChoice{
				Version:  version,
				State:    "installed",
				Evidence: fmt.Sprintf("PHP %s is available in the selected %s runtime", version, cfg.Runtime.Engine),
			})
		}
	}
	plan := projects.BuildSetupPlan(
		projects.ProjectFacts{
			Path:         dirPath,
			Slug:         slug,
			Type:         projectType,
			DocumentRoot: projects.DocumentRoot(dirPath, projectType),
			Existing:     existing,
		},
		projects.SetupChoices{
			PHPVersion: phpVersion,
			Domain:     domain,
			Aliases:    append([]string(nil), aliases...),
			SSL:        selectedSSL,
			Dedicated:  selectedDedicated,
		},
		phpChoices,
	)
	if err := plan.Validate(); err != nil {
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
		p.ProjectType = projectType
		p.ProxyPort = proxyPort
		p.FPM.Dedicated = selectedDedicated
		p.SSL = selectedSSL
		if *siteFlag != "" {
			p.Domain = domain
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
		if selectedDedicated {
			sockPath = root + "/projects/" + slug + "/php-fpm.sock"
		}

		p = &projects.Project{
			Slug:        slug,
			Path:        dirPath,
			Domain:      domain,
			Aliases:     aliases,
			PHPVersion:  phpVersion,
			SSL:         selectedSSL,
			FPM:         projects.FPMConfig{Dedicated: selectedDedicated, Socket: sockPath},
			ProjectType: projectType,
			ProxyPort:   proxyPort,
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

	applyResult, err := applyLinkProject(p, root, cfg)
	if err != nil {
		lib.Error(fmt.Sprintf("Link failed [%s]", projects.ApplyFailed))
		return err
	}

	// ── Output ──────────────────────────────────────────────────────────────────

	fmt.Println()
	if isUpdate {
		lib.Success(fmt.Sprintf("Project updated [%s]", applyResult.Status))
	} else {
		lib.Success(fmt.Sprintf("Project linked [%s]", applyResult.Status))
	}
	for _, evidence := range applyResult.Evidence {
		lib.Info(lib.Gray(evidence))
	}
	if applyResult.Remediation != "" {
		lib.Info(lib.Gray("Next: " + applyResult.Remediation))
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
	if p.ProjectType == projects.TypeReverseProxy {
		lib.Pair("Proxy", fmt.Sprintf("http://127.0.0.1:%d", p.ProxyPort))
	} else {
		lib.Pair("PHP", phpFPMLabel(p))
	}
	lib.Pair("Type", titleCase(string(p.ProjectType)))

	if len(p.Aliases) > 0 {
		lib.Pair("Aliases", strings.Join(p.Aliases, ", "))
	}
	fmt.Println()

	if !nginxRuntimeRunning(root, cfg) {
		lib.Info(lib.Gray("Start services:    chauf start"))
	}
	if !p.SSL {
		lib.Info(lib.Gray("Enable SSL:        chauf secure"))
	}
	fmt.Println()

	return nil
}

func nginxRuntimeRunning(root string, cfg workspace.Config) bool {
	if cfg.Runtime.Engine != string(chauftruntime.EnginePodman) {
		return projects.IsNginxRunning(root)
	}
	rt, err := chauftruntime.ForWorkspace(cfg)
	if err != nil {
		return false
	}
	statuses, err := rt.Status(context.Background(), chauftruntime.Scope{Service: "nginx"})
	if err != nil {
		return false
	}
	for _, status := range statuses {
		if status.Healthy {
			return true
		}
	}
	return false
}

// flag.FlagSet stops parsing at the first positional argument, while the
// documented link syntax allows the path before its flags. Move that leading
// path behind the flags without changing the existing flags-first form.
func normalizeLinkArgs(args []string) []string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return args
	}
	return append(append([]string(nil), args[1:]...), args[0])
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
	printRuntime(cfg)
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
		"Project", domW, "Domain", "Type", "FPM", "SSL")
	sep := strings.Repeat("─", len(header))
	fmt.Printf(" %s\n%s\n", lib.Bold(header[1:]), sep)

	for _, p := range all {
		scheme := schemeLabel(p.SSL)
		fpmMode := "shared"
		if p.FPM.Dedicated {
			fpmMode = "dedicated"
		}
		if p.ProjectType == projects.TypeReverseProxy {
			fpmMode = "-"
		}
		projectType := string(p.ProjectType)
		if p.ProjectType == projects.TypeReverseProxy {
			projectType = fmt.Sprintf("proxy:%d", p.ProxyPort)
		}
		fmt.Printf(" %-20s  %-*s  %-5s  %-9s  %s\n",
			p.Slug, domW, p.Domain, projectType, fpmMode, scheme)
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
	runtimeLabel, runtimeValue := projectDetailRuntime(p)
	lib.Pair("  "+runtimeLabel, runtimeValue)
	lib.Pair("  Type", titleCase(string(p.ProjectType)))
	lib.Pair("  SSL", schemeLabel(p.SSL))
	if p.SSL {
		lib.Pair("  Cert", short(root+"/nginx/certs/"+p.Domain+".crt"))
	}
	if p.ProjectType != projects.TypeReverseProxy {
		if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
			lib.Pair("  Socket", lib.Gray("container-managed"))
		} else {
			lib.Pair("  Socket", short(p.FPMSocketPath(root)))
		}
	}
	lib.Pair("  Nginx", short(root+"/nginx/etc/sites-available/"+p.Slug+".conf"))
	lib.Pair("  Config", short(p.ConfigPath(root)))

	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		rt, err := chauftruntime.ForWorkspace(cfg)
		if err != nil {
			lib.Pair("  Services", lib.Gray("unavailable: "+err.Error()))
			return
		}
		nginxStatus := runtimeContainerStatus(rt, chauftruntime.Scope{Service: "nginx"})
		if p.ProjectType == projects.TypeReverseProxy {
			lib.Pair("  Services", fmt.Sprintf("nginx %s  /  proxy target localhost:%d", nginxStatus, p.ProxyPort))
			return
		}
		fpmStatus := runtimeContainerStatus(rt, chauftruntime.Scope{Version: p.PHPVersion, Project: p.Path, Dedicated: p.FPM.Dedicated})
		lib.Pair("  Services", fmt.Sprintf("nginx %s  /  php-fpm %s %s", nginxStatus, p.PHPVersion, fpmStatus))
		return
	}

	nginxRunning := pidFileRunning(root + "/nginx/logs/nginx.pid")
	if p.ProjectType == projects.TypeReverseProxy {
		lib.Pair("  Services", fmt.Sprintf("nginx %s  /  proxy target localhost:%d",
			serviceStatus(nginxRunning), p.ProxyPort))
		return
	}
	fpmRunning := pidFileRunning(root + "/php/" + p.PHPVersion + "/runtime/php-fpm/php-fpm.pid")
	lib.Pair("  Services", fmt.Sprintf("nginx %s  /  php-fpm %s %s",
		serviceStatus(nginxRunning), p.PHPVersion, serviceStatus(fpmRunning)))
}

func runtimeContainerStatus(rt chauftruntime.Runtime, scope chauftruntime.Scope) string {
	statuses, err := rt.Status(context.Background(), scope)
	if err != nil || len(statuses) == 0 {
		return lib.Gray("○ unavailable")
	}
	if statuses[0].Healthy {
		return lib.Green("● running")
	}
	if statuses[0].State == "image-missing" {
		return lib.Red("✗ image missing")
	}
	return lib.Gray("○ stopped")
}

func projectDetailRuntime(p *projects.Project) (string, string) {
	if p.ProjectType == projects.TypeReverseProxy {
		proxyPort := p.ProxyPort
		if proxyPort == 0 {
			proxyPort = projects.DefaultProxyPortFor(p.Path)
		}
		return "Proxy", fmt.Sprintf("http://localhost:%d", proxyPort)
	}
	return "PHP", phpFPMLabel(p)
}

// ── chauf unlink ──────────────────────────────────────────────────────────────

func RunUnlink(args []string) error {
	flags := flag.NewFlagSet("unlink", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf unlink — unregister a project", "chauf unlink [path] [--site <domain>] [--alias <domain>] [--all] [--yes]")
	aliasFlag := flags.String("alias", "", "Remove a specific alias domain only")
	siteFlag := flags.String("site", "", "Find the project by primary domain or alias")
	allFlag := flags.Bool("all", false, "Remove all aliases then unlink the project")
	yesFlag := flags.Bool("yes", false, "Skip confirmation prompt")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root := workspace.Root()
	cfg := workspace.Load()
	printRuntime(cfg)
	fmt.Println()

	var p *projects.Project
	var err error
	if *siteFlag != "" {
		if flags.NArg() > 0 {
			return fmt.Errorf("use either a path or --site, not both")
		}
		p, err = projects.FindByDomain(root, *siteFlag)
	} else {
		// Resolve target directory
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return fmt.Errorf("get current directory: %w", cwdErr)
		}
		dirPath := cwd
		if flags.NArg() > 0 {
			dirPath = flags.Arg(0)
		}
		p, err = projects.FindByPath(root, dirPath)
	}
	if err != nil {
		return err
	}
	if p == nil {
		if *siteFlag != "" {
			lib.Warn("No project registered for site " + *siteFlag + ".")
		} else {
			lib.Warn("No project registered for this directory.")
			lib.Info(lib.Gray("Register with:  chauf link"))
		}
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
		for _, item := range unlinkSummary(p, workspace.Load()) {
			lib.Pair(item.label, item.value)
		}
		fmt.Println()
		if !tui.Confirm("Confirm") {
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
	if cfg.Runtime.Engine != string(chauftruntime.EnginePodman) && projects.IsNginxRunning(root) {
		if err := projects.ReloadNginx(root); err != nil {
			return fmt.Errorf("reload nginx: %w", err)
		}
	}

	fmt.Println()
	lib.Success(fmt.Sprintf("Project %s unlinked", lib.Bold(p.Slug)))
	fmt.Println()

	_ = *allFlag // --all handled implicitly (full unlink removes everything)
	return nil
}

type unlinkSummaryItem struct {
	label string
	value string
}

func unlinkSummary(p *projects.Project, cfg workspace.Config) []unlinkSummaryItem {
	scheme := "http"
	port := cfg.Nginx.HTTPPort
	sslStatus := "Disabled (HTTP only)"
	if p.SSL {
		scheme = "https"
		port = cfg.Nginx.HTTPSPort
		sslStatus = fmt.Sprintf("Enabled at %s://%s:%d (certificate retained)", scheme, p.Domain, port)
	}

	items := []unlinkSummaryItem{
		{label: "Path", value: p.Path},
		{label: "Domain", value: fmt.Sprintf("%s://%s:%d", scheme, p.Domain, port)},
		{label: "SSL", value: sslStatus},
		{label: "Type", value: projectTypeSetupLabel(p.ProjectType)},
	}
	if len(p.Aliases) > 0 {
		items = append(items, unlinkSummaryItem{label: "Aliases", value: strings.Join(p.Aliases, ", ")})
	}
	if p.ProjectType == projects.TypeReverseProxy {
		proxyPort := p.ProxyPort
		if proxyPort == 0 {
			proxyPort = projects.DefaultProxyPortFor(p.Path)
		}
		items = append(items,
			unlinkSummaryItem{label: "Proxy target", value: fmt.Sprintf("http://localhost:%d", proxyPort)},
			unlinkSummaryItem{label: "Runtime", value: "No PHP-FPM; external development server"},
		)
	} else {
		fpmMode := "shared FPM"
		if p.FPM.Dedicated {
			fpmMode = "dedicated FPM"
		}
		items = append(items, unlinkSummaryItem{
			label: "Runtime",
			value: fmt.Sprintf("PHP %s · %s", p.PHPVersion, fpmMode),
		})
	}
	items = append(items, unlinkSummaryItem{
		label: "Nginx",
		value: "Route and enabled-site link will be removed",
	})
	return items
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
	printRuntime(cfg)
	fmt.Println()

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
	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		if _, err := applyLinkProject(p, root, cfg); err != nil {
			return err
		}
	} else {
		if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
			return fmt.Errorf("write nginx config: %w", err)
		}
		if err := projects.Save(p, root); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
		if projects.IsNginxRunning(root) {
			if err := projects.ReloadNginx(root); err != nil {
				return fmt.Errorf("reload nginx: %w", err)
			}
		}
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
	printRuntime(cfg)
	fmt.Println()

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
	if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		if _, err := applyLinkProject(p, root, cfg); err != nil {
			return err
		}
	} else {
		if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
			return fmt.Errorf("write nginx config: %w", err)
		}
		if err := projects.Save(p, root); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
		if projects.IsNginxRunning(root) {
			if err := projects.ReloadNginx(root); err != nil {
				return fmt.Errorf("reload nginx: %w", err)
			}
		}
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

func parseProjectType(value string) (projects.ProjectType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "laravel":
		return projects.TypeLaravel, nil
	case "wordpress", "wp":
		return projects.TypeWordPress, nil
	case "php":
		return projects.TypePHP, nil
	case "reverse-proxy", "reverse_proxy", "proxy":
		return projects.TypeReverseProxy, nil
	default:
		return projects.TypeUnknown, fmt.Errorf("invalid project type %q; choose laravel, wordpress, php, or reverse-proxy", value)
	}
}

func printProjectDetection(detected, selected projects.ProjectType) {
	if detected == projects.TypeUnknown {
		lib.Info(fmt.Sprintf("Project type was not detected automatically; using %s setup.", projectTypeSetupLabel(selected)))
		return
	}
	lib.Info(fmt.Sprintf("Project detected as %s, using %s setup.", projectTypeDetectionLabel(detected), projectTypeSetupLabel(selected)))
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

	if wCfg.Runtime.Engine == string(chauftruntime.EnginePodman) {
		if _, err := applyLinkProject(p, root, wCfg); err != nil {
			return err
		}
	} else {
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
			if err := projects.ReloadNginx(root); err != nil {
				return fmt.Errorf("reload nginx: %w", err)
			}
		}
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

// applyLinkProject is the single mutation boundary for link. Wizard and
// noninteractive callers both construct intent before entering this function.
func applyLinkProject(p *projects.Project, root string, cfg workspace.Config) (projects.ApplyResult, error) {
	plan := projects.SetupPlan{
		Facts:   projects.ProjectFacts{Path: p.Path, Slug: p.Slug, Type: p.ProjectType, DocumentRoot: projects.DocumentRoot(p.Path, p.ProjectType), Existing: p},
		Choices: projects.SetupChoices{PHPVersion: p.PHPVersion, Domain: p.Domain, Aliases: p.Aliases, SSL: p.SSL, Dedicated: p.FPM.Dedicated},
	}
	return projects.ApplyProjectSetup(context.Background(), plan, projects.ApplyDependencies{
		Save: func() error { return projects.Save(p, root) },
		GenerateSSL: func() error {
			if !p.SSL {
				return nil
			}
			requested := p.SSL
			if err := ensureSSL(p, root); err != nil {
				return err
			}
			// mkcert may downgrade SSL to HTTP when trust prerequisites are
			// unavailable; persist that explicit degraded intent.
			if requested && !p.SSL {
				return projects.Save(p, root)
			}
			return nil
		},
		PrepareRuntime: func() error {
			if p.FPM.Dedicated {
				if err := writeDedicatedFPMConfig(p, root); err != nil {
					return fmt.Errorf("write dedicated FPM config: %w", err)
				}
			}
			if p.ProjectType == projects.TypeReverseProxy {
				return nil
			}
			rt, err := chauftruntime.ForWorkspace(cfg)
			if err != nil {
				return err
			}
			return rt.EnsureLinkedProject(context.Background(), root, filepath.Join(root, "nginx", "container.conf"), filepath.Join(root, "nginx", "certs"), cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort, chauftruntime.ProjectSpec{
				Slug: p.Slug, Path: p.Path, Version: p.PHPVersion, Domains: p.AllDomains(), Dedicated: p.FPM.Dedicated, SSL: p.SSL, CertName: p.Domain,
			})
		},
		GenerateNginx: func() error {
			if err := projects.WriteNginxConfig(p, root, cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort); err != nil {
				return fmt.Errorf("write nginx config: %w", err)
			}
			return nil
		},
		EnableRoute: func() error {
			if err := projects.EnableNginxSite(p, root); err != nil {
				return fmt.Errorf("enable nginx site: %w", err)
			}
			return nil
		},
		Reload: func() error {
			if !projects.IsNginxRunning(root) {
				return nil
			}
			if err := projects.ReloadNginx(root); err != nil {
				return fmt.Errorf("reload nginx: %w", err)
			}
			return nil
		},
		Readiness: func() (string, error) {
			if cfg.Runtime.Engine == string(chauftruntime.EnginePodman) || projects.IsNginxRunning(root) {
				return "selected runtime and nginx route are ready", nil
			}
			return "project intent and nginx route were saved; services are not running", fmt.Errorf("services are not running")
		},
	})
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
