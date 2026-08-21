package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

func RunInstall(args []string) error {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	lib.SetFlagUsage(flags, "chauf install — install nginx, PHP, or Composer from source", "chauf install <nginx|php <version>|composer> [--force] [--no-cache]")
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
			version = interactiveSelectPHPVersion(installers.SupportedPHPVersions, "Select PHP version to install")
			if version == "" {
				return nil // user cancelled or all installed
			}
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

// ── PHP version interactive selector ──────────────────────────────────────────

// phpSelectModel is a bubbletea model for selecting a PHP version with installed versions marked/disabled.
type phpSelectModel struct {
	items     []string // display items (may include " (installed)" suffix)
	cursor    int      // cursor index into selectable (non-installed) items
	done      bool
	aborted   bool
	title     string
	installed map[string]bool // set of installed version strings
	width     int
	height    int
}

func (m phpSelectModel) Init() tea.Cmd {
	return nil
}

func (m phpSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.cursor = tui.Move(m.cursor, len(m.selectableItems()), -1)
		case tea.KeyDown:
			m.cursor = tui.Move(m.cursor, len(m.selectableItems()), 1)
		case tea.KeyHome:
			m.cursor = 0
		case tea.KeyEnd:
			m.cursor = tui.Move(len(m.selectableItems())-1, len(m.selectableItems()), 0)
		case tea.KeySpace, tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// selectableItems returns items that are not installed (in display form).
func (m phpSelectModel) selectableItems() []string {
	var selectable []string
	for _, item := range m.items {
		if !m.installed[item] {
			selectable = append(selectable, item)
		}
	}
	return selectable
}

func (m phpSelectModel) activeRenderedRow() int {
	if m.cursor < 0 {
		return -1
	}

	selectableIdx := 0

	for rowIdx, item := range m.items {
		if m.installed[item] {
			continue
		}

		if selectableIdx == m.cursor {
			return rowIdx
		}

		selectableIdx++
	}

	return -1
}

func (m phpSelectModel) nearestInstalledRowBefore(activeRow int) int {
	if activeRow <= 0 {
		return -1
	}

	for rowIdx := activeRow - 1; rowIdx >= 0; rowIdx-- {
		if m.installed[m.items[rowIdx]] {
			return rowIdx
		}
	}

	return -1
}

func (m phpSelectModel) View() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("  %s", lib.Bold(m.title)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s", tui.Footer("↑/↓ move · enter select · esc cancel · ? help")))

	// Build display rows and track which selectable index maps to which item.
	selectable := m.selectableItems()
	activeRow := m.activeRenderedRow()
	if activeRow < 0 {
		activeRow = 0
	}
	start, end, showAbove, showBelow := selectorViewportWithIndicators(len(m.items), activeRow, m.height, 4)
	visibleRows := end - start
	if installedRow := m.nearestInstalledRowBefore(activeRow); installedRow >= 0 && installedRow < start {
		candidateStart := max(0, start-1)
		if candidateStart > installedRow {
			candidateStart = installedRow
		}
		candidateEnd := candidateStart + visibleRows
		if candidateEnd > len(m.items) {
			candidateEnd = len(m.items)
			candidateStart = max(0, candidateEnd-visibleRows)
		}
		if candidateStart <= activeRow && activeRow < candidateEnd {
			start = candidateStart
			end = candidateEnd
			showAbove = start > 0
			showBelow = end < len(m.items)
		}
	}
	selectableIdx := 0

	if showAbove {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↑ more")))
	}

	for rowIdx, item := range m.items {
		if rowIdx < start || rowIdx >= end {
			if !m.installed[item] {
				selectableIdx++
			}
			continue
		}

		cursor := "  "
		installed := m.installed[item]

		if installed {
			lines = append(lines, fmt.Sprintf("  %s %s", lib.Gray("  "), lib.Gray(item+" (installed)")))
		} else {
			if selectableIdx == m.cursor {
				cursor = tui.Cursor(true)
			}
			lines = append(lines, fmt.Sprintf("  %s %s", cursor, item))
			selectableIdx++
		}
	}

	if showBelow {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("↓ more")))
	}

	// Hint
	if len(selectable) == 0 {
		lines = append(lines, fmt.Sprintf("  %s", lib.Gray("All PHP versions are already installed")))
	} else {
		lines = append(lines, fmt.Sprintf("  %s %d available", lib.Gray("✓"), len(selectable)))
	}

	return strings.Join(lines, "\n")
}

// interactiveSelectPHPVersion shows an interactive selector for PHP versions,
// marking installed ones as already-installed and only allowing selection of non-installed versions.
func interactiveSelectPHPVersion(versions []string, title string) string {
	if len(versions) == 0 {
		return ""
	}
	if !tui.Interactive() {
		return ""
	}

	installedSet := make(map[string]bool)
	for _, v := range installers.ListInstalledPHP(workspace.Root()) {
		installedSet[v] = true
	}

	// If all are installed, fall back to simple list
	selectableCount := 0
	for _, v := range versions {
		if !installedSet[v] {
			selectableCount++
		}
	}
	if selectableCount == 0 {
		fmt.Println()
		fmt.Printf("  %s\n", lib.Bold(title))
		fmt.Println()
		for _, v := range versions {
			fmt.Printf("  %s %s\n", lib.Gray("  "), lib.Gray(v+" (installed)"))
		}
		fmt.Println()
		lib.Warn("All PHP versions are already installed")
		return ""
	}

	model := phpSelectModel{
		items:     versions,
		cursor:    0,
		title:     title,
		installed: installedSet,
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return interactiveSelectPHPSimple(versions, installedSet, title)
	}

	if finalModel, ok := result.(phpSelectModel); ok {
		if finalModel.aborted {
			return ""
		}
		// Return the selectable item at cursor position
		selectable := finalModel.selectableItems()
		if finalModel.cursor < len(selectable) {
			return selectable[finalModel.cursor]
		}
	}

	return ""
}

// interactiveSelectPHPSimple is the fallback text-based PHP version selector.
func interactiveSelectPHPSimple(versions []string, installed map[string]bool, title string) string {
	if !tui.Interactive() {
		return ""
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("  %s\n", lib.Bold(title))
	fmt.Println()

	selectable := []string{}
	for _, v := range versions {
		suffix := ""
		if installed[v] {
			suffix = " (installed)"
		} else {
			selectable = append(selectable, v)
		}
		if installed[v] {
			fmt.Printf("  %s %s%s\n", lib.Gray("  "), v, lib.Gray(suffix))
		}
	}
	fmt.Println()

	if len(selectable) == 0 {
		lib.Warn("All PHP versions are already installed")
		return ""
	}

	fmt.Printf("  %s %d available\n", lib.Gray("✓"), len(selectable))
	fmt.Println()
	fmt.Print("  " + lib.Bold("Choice") + " " + lib.Gray("[1-"+fmt.Sprintf("%d", len(selectable))+" or name]: "))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\r", "")

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(selectable) {
		return selectable[idx-1]
	}

	inputLower := strings.ToLower(input)
	for _, item := range selectable {
		if strings.ToLower(item) == inputLower {
			return item
		}
	}

	return ""
}
