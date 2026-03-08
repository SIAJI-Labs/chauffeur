package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
)

func RunInstall(args []string) error {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	force := flags.Bool("force", false, "Reinstall even if already present")
	noCache := flags.Bool("no-cache", false, "Skip download cache")
	verbose := flags.Bool("verbose", false, "Stream build output to terminal")
	flags.BoolVar(verbose, "v", false, "Stream build output to terminal")

	// Separate flag args from positional args so flags can appear anywhere,
	// e.g. both "chauf install --verbose php 8.3" and "chauf install php 8.3 --verbose" work.
	flagArgs, positionals := splitArgs(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}

	if len(positionals) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: chauf install <nginx|php <version>|composer>\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  chauf install nginx\n")
		fmt.Fprintf(os.Stderr, "  chauf install php 8.3\n")
		fmt.Fprintf(os.Stderr, "  chauf install composer\n")
		return fmt.Errorf("no service specified")
	}

	opts := installers.BuildOpts{
		Force:   *force,
		NoCache: *noCache,
		Verbose: *verbose,
	}

	switch strings.ToLower(positionals[0]) {
	case "nginx":
		return installNginx(opts)
	case "php":
		version := ""
		if len(positionals) > 1 {
			version = positionals[1]
		}
		if version == "" {
			return fmt.Errorf("usage: chauf install php <version>  (e.g. 8.3, 8.3.20)")
		}
		return installPHP(version, opts)
	case "composer":
		return installComposer(opts)
	default:
		return fmt.Errorf("unknown service %q — available: nginx, php, composer", positionals[0])
	}
}

// splitArgs separates flag arguments (starting with -) from positional arguments.
// This allows flags to appear anywhere in the argument list, not just before positionals.
func splitArgs(args []string) (flagArgs, positionals []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
		} else {
			positionals = append(positionals, arg)
		}
	}
	return
}

// ── step helper ────────────────────────────────────────────────────────────────
//
// In quiet mode (default): animated spinner, build output captured.
// In verbose mode (--verbose/-v): plain section header, build output streams freely.

type buildStep struct {
	label   string
	verbose bool
	spin    *lib.Spinner
}

func startStep(label string, verbose bool) *buildStep {
	s := &buildStep{label: label, verbose: verbose}
	if verbose {
		fmt.Printf("\n  %s  %s\n", lib.Cyan("→"), lib.Bold(label))
	} else {
		s.spin = lib.NewSpinner(label)
	}
	return s
}

func (s *buildStep) success(detail string) {
	if s.verbose {
		suffix := ""
		if detail != "" {
			suffix = "  " + lib.Gray(detail)
		}
		lib.Success(s.label + suffix)
	} else {
		s.spin.Success(detail)
	}
}

func (s *buildStep) fail(detail string) {
	if s.verbose {
		lib.Error(s.label + "  " + detail)
	} else {
		s.spin.Fail(detail)
	}
}

// ── nginx ──────────────────────────────────────────────────────────────────────

func installNginx(opts installers.BuildOpts) error {
	inst := installers.NewNginxInstaller(opts)
	fmt.Println()

	if inst.IsInstalled() && !opts.Force {
		lib.Pair("nginx", lib.Gray("already installed — "+inst.InstalledVersion()))
		lib.Info(lib.Gray("Use --force to reinstall."))
		fmt.Println()
		return nil
	}

	// ── Resolve version ────────────────────────────────────────────────────
	s := startStep("Resolving nginx version", opts.Verbose)
	version := inst.ResolveVersion()
	s.success(version)

	// ── Download ───────────────────────────────────────────────────────────
	tarPath, resolvedVersion, cached, err := inst.Download(version)
	if err != nil {
		return err
	}
	if cached {
		lib.Pair("Download", lib.Gray("cached — nginx-"+resolvedVersion+".tar.gz"))
	} else {
		lib.Success("Downloaded  nginx-" + resolvedVersion + ".tar.gz")
	}

	// ── Build ──────────────────────────────────────────────────────────────
	s = startStep("Building nginx "+resolvedVersion, opts.Verbose)
	if !opts.Verbose {
		lib.Info(lib.Gray("  (this may take a few minutes)"))
	}
	if err := inst.Build(tarPath, resolvedVersion); err != nil {
		s.fail(err.Error())
		return err
	}
	s.success(inst.BinPath())

	fmt.Println()
	lib.Success("nginx " + resolvedVersion + " installed")
	fmt.Println()
	lib.Info(lib.Gray("Start with:  chauf start nginx"))
	fmt.Println()
	return nil
}

// ── PHP ────────────────────────────────────────────────────────────────────────

func installPHP(version string, opts installers.BuildOpts) error {
	inst, err := installers.NewPHPInstaller(version, opts)
	if err != nil {
		return err
	}
	fmt.Println()

	if inst.IsInstalled() && !opts.Force {
		lib.Pair("PHP "+inst.MajorMinorStr(), lib.Gray("already installed — "+inst.InstalledVersion()))
		lib.Info(lib.Gray("Use --force to reinstall."))
		fmt.Println()
		return nil
	}

	// ── Resolve full version ───────────────────────────────────────────────
	s := startStep("Resolving PHP "+inst.MajorMinorStr()+" version", opts.Verbose)
	fullVersion, err := inst.ResolveVersion()
	if err != nil {
		s.fail(err.Error())
		return err
	}
	s.success(fullVersion)

	// ── Download (or use local tarball override) ───────────────────────────
	tarPath := os.Getenv("CHAUFFEUR_PHP_TARBALL")
	if tarPath != "" {
		lib.Pair("Source", lib.Gray("local tarball: "+tarPath))
	} else {
		var cached bool
		tarPath, cached, err = inst.Download(fullVersion)
		if err != nil {
			return err
		}
		if cached {
			lib.Pair("Download", lib.Gray("cached — php-"+fullVersion+".tar.gz"))
		} else {
			lib.Success("Downloaded  php-" + fullVersion + ".tar.gz")
		}
	}

	// ── Build ──────────────────────────────────────────────────────────────
	s = startStep("Building PHP "+fullVersion, opts.Verbose)
	if !opts.Verbose {
		lib.Info(lib.Gray("  (this may take 5–15 minutes)"))
	}
	if err := inst.Build(tarPath, fullVersion); err != nil {
		s.fail(err.Error())
		return err
	}
	s.success(inst.BinPath())

	// ── imagick extension ──────────────────────────────────────────────────
	s = startStep("Installing imagick extension", opts.Verbose)
	if err := inst.BuildImagick(); err != nil {
		s.fail(err.Error())
		lib.Warn("imagick failed (non-fatal) — install ImageMagick dev headers and retry")
	} else {
		s.success("extension=imagick")
	}

	fmt.Println()
	lib.Success("PHP " + fullVersion + " installed")
	fmt.Println()
	lib.Info(lib.Gray("Set as default:  chauf php use " + inst.MajorMinorStr()))
	fmt.Println()
	return nil
}

// ── Composer ───────────────────────────────────────────────────────────────────

func installComposer(opts installers.BuildOpts) error {
	inst := installers.NewComposerInstaller(opts)
	fmt.Println()

	if inst.IsInstalled() && !opts.Force {
		lib.Pair("Composer", lib.Gray("already installed — "+inst.InstalledVersion()))
		lib.Info(lib.Gray("Use --force to reinstall."))
		fmt.Println()
		return nil
	}

	// ── Resolve version ────────────────────────────────────────────────────
	s := startStep("Resolving Composer version", opts.Verbose)
	version := inst.ResolveVersion()
	s.success(version)

	// ── Download ───────────────────────────────────────────────────────────
	cached, err := inst.Download(version)
	if err != nil {
		return err
	}
	if cached {
		lib.Pair("composer.phar", lib.Gray("already present"))
	} else {
		lib.Success("Downloaded  composer.phar " + version)
	}

	fmt.Println()
	lib.Success("Composer " + version + " installed  →  " + inst.PharPath())
	fmt.Println()
	lib.Info(lib.Gray("Usage:  composer --version"))
	fmt.Println()
	return nil
}
