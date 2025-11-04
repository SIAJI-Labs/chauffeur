package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/workspace"
)

const (
	defaultRepoURL   = "git@github.com:SIAJI-Labs/chauffeur.git"
	defaultGitBranch = "main"
)

const statusPrefix = "[ self-update ]"

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Color functions
func colorize(color, text string) string {
	if !isTerminal(os.Stdout) {
		return text
	}
	return color + text + colorReset
}

func green(text string) string {
	return colorize(colorGreen, text)
}

func red(text string) string {
	return colorize(colorRed, text)
}

func yellow(text string) string {
	return colorize(colorYellow, text)
}

func blue(text string) string {
	return colorize(colorBlue, text)
}

func gray(text string) string {
	return colorize(colorGray, text)
}

func bold(text string) string {
	return colorize(colorBold, text)
}

var (
	runCommand = defaultRunCommand
	goBuild    = defaultGoBuild
	lookPath   = exec.LookPath
)

// RunSelfUpdate handles `chauf self-update`.
func RunSelfUpdate(args []string) error {
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printSelfUpdateUsage()
			return nil
		default:
			return fmt.Errorf("unknown flag for self-update: %s", arg)
		}
	}

	updater, err := newGitSelfUpdater()
	if err != nil {
		return err
	}

	return updater.Execute()
}

type gitSelfUpdater struct {
	repoURL       string
	branch        string
	repoDir       string
	sourceDir     string
	executable    string
	currentSHA    string
	targetVersion string
}

func newGitSelfUpdater() (*gitSelfUpdater, error) {
	if _, err := lookPath("git"); err != nil {
		return nil, fmt.Errorf("git is required by self-update: %w", err)
	}
	if _, err := lookPath("go"); err != nil {
		return nil, fmt.Errorf("go is required by self-update: %w", err)
	}

	repoURL := strings.TrimSpace(os.Getenv("CHAUF_REPO_URL"))
	if repoURL == "" {
		repoURL = defaultRepoURL
	}

	branch := strings.TrimSpace(os.Getenv("CHAUF_UPDATE_BRANCH"))
	if branch == "" {
		branch = defaultGitBranch
	}

	ws, err := workspace.Dir()
	if err != nil {
		return nil, err
	}

	sourceDir := filepath.Join(ws, "src")
	if err := workspace.Ensure("src"); err != nil {
		return nil, err
	}

	repoDir := filepath.Join(sourceDir, "chauffeur")

	target := os.Getenv("CHAUF_SELF_UPDATE_TARGET")
	if target == "" {
		target, err = os.Executable()
		if err != nil {
			return nil, fmt.Errorf("determine executable path: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(target); err == nil {
			target = resolved
		}
	}

	return &gitSelfUpdater{
		repoURL:       repoURL,
		branch:        branch,
		repoDir:       repoDir,
		sourceDir:     sourceDir,
		executable:    target,
		targetVersion: getCLIVersion(),
	}, nil
}

func (su *gitSelfUpdater) Execute() error {
	overallStart := time.Now()
	fmt.Printf("%s %s...\n", blue(statusPrefix), bold("Starting self-update process"))

	created, err := su.ensureRepository()
	if err != nil {
		return err
	}

	if err := su.ensureCleanWorkingTree(); err != nil {
		return err
	}

	beforeSHA, err := su.currentRevision()
	if err != nil && !created {
		return err
	}

	changed, afterSHA, err := su.updateRepository(beforeSHA)
	if err != nil {
		return err
	}

	if afterSHA == "" {
		afterSHA, err = su.currentRevision()
		if err != nil {
			return err
		}
	}

	if err := su.buildAndInstall(); err != nil {
		return err
	}

	// Print final summary
	overallDuration := time.Since(overallStart)
	fmt.Printf("\n%s %s:\n", blue(statusPrefix), bold("Summary"))
	fmt.Printf("  └── %s: %s\n", yellow("Duration"), formatDuration(overallDuration))
	fmt.Printf("  └── %s: %s\n", gray("Previous"), func() string {
		if beforeSHA != "" {
			return gray(shortSHA(beforeSHA))
		}
		return gray("fresh clone")
	}())
	fmt.Printf("  └── %s:  %s\n", gray("Current"), gray(shortSHA(afterSHA)))
	fmt.Printf("  └── %s:   %s\n", gray("Changes"), func() string {
		if created {
			return green("fresh installation")
		} else if changed {
			return yellow("updated")
		} else {
			return gray("rebuilt only")
		}
	}())
	
	switch {
	case created:
		fmt.Printf("\n%s %s from fresh clone (commit %s).\n", blue(statusPrefix), bold(green("Self-update complete")), gray(shortSHA(afterSHA)))
	case changed:
		fmt.Printf("\n%s %s (commit %s).\n", blue(statusPrefix), bold(green("Self-update complete")), gray(shortSHA(afterSHA)))
	default:
		fmt.Printf("\n%s %s at commit %s (no source changes).\n", blue(statusPrefix), bold(blue("Binary rebuilt")), gray(shortSHA(afterSHA)))
	}
	return nil
}

func (su *gitSelfUpdater) ensureRepository() (bool, error) {
	if _, err := os.Stat(su.repoDir); errors.Is(err, os.ErrNotExist) {
		spin := newSpinner("Cloning Chauffeur sources")
		start := time.Now()
		if _, err := runCommand("", "git", "clone", "--branch", su.branch, su.repoURL, su.repoDir); err != nil {
			spin.Fail("clone failed")
			return false, fmt.Errorf("git clone: %w", err)
		}
		sha, err := su.currentRevision()
		if err != nil {
			spin.Fail("revision lookup failed")
			return false, err
		}
		spin.Success(fmt.Sprintf("into %s (branch %s @ %s, %s)", su.repoDir, su.branch, shortSHA(sha), formatDuration(time.Since(start))))
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("check repository: %w", err)
	}
	fmt.Printf("%s Using existing sources at %s\n", blue(statusPrefix), gray(su.repoDir))
	return false, nil
}

func (su *gitSelfUpdater) ensureCleanWorkingTree() error {
	out, err := runCommand(su.repoDir, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("repository has uncommitted changes; commit or stash them before running self-update")
	}
	return nil
}

func (su *gitSelfUpdater) currentRevision() (string, error) {
	out, err := runCommand(su.repoDir, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (su *gitSelfUpdater) updateRepository(before string) (bool, string, error) {
	spin := newSpinner(fmt.Sprintf("Updating branch %s", su.branch))
	start := time.Now()
	if _, err := runCommand(su.repoDir, "git", "fetch", "--tags", "--prune"); err != nil {
		spin.Fail("fetch failed")
		return false, "", fmt.Errorf("git fetch: %w", err)
	}

	if _, err := runCommand(su.repoDir, "git", "checkout", su.branch); err != nil {
		spin.Fail("checkout failed")
		return false, "", fmt.Errorf("git checkout %s: %w", su.branch, err)
	}

	output, err := runCommand(su.repoDir, "git", "pull", "--ff-only")
	if err != nil {
		spin.Fail("pull failed")
		return false, "", fmt.Errorf("git pull: %w", err)
	}

	after, err := su.currentRevision()
	if err != nil {
		spin.Fail("revision lookup failed")
		return false, "", err
	}

	if before == "" {
		before = after
	}

	if before == after || strings.Contains(output, "Already up to date") {
		spin.Success(fmt.Sprintf("already up to date (commit %s, %s)", shortSHA(after), formatDuration(time.Since(start))))
		return false, after, nil
	}
	spin.Success(fmt.Sprintf("updated to %s (was %s, %s)", shortSHA(after), shortSHA(before), formatDuration(time.Since(start))))
	return true, after, nil
}

func (su *gitSelfUpdater) buildAndInstall() error {
	dir := filepath.Dir(su.executable)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure binary directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "chauf-build-*")
	if err != nil {
		return fmt.Errorf("prepare temporary binary: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	spin := newSpinner("Building Chauffeur CLI")
	start := time.Now()

	if err := goBuild(su.repoDir, tmpPath); err != nil {
		spin.Fail("build failed")
		return fmt.Errorf("go build: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		spin.Fail("set permissions failed")
		return fmt.Errorf("set binary permissions: %w", err)
	}

	backupPath := su.executable + ".bak"
	if err := os.Rename(su.executable, backupPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		spin.Fail("backup failed")
		return fmt.Errorf("backup existing binary: %w", err)
	}

	if err := os.Rename(tmpPath, su.executable); err != nil {
		_ = os.Rename(backupPath, su.executable)
		spin.Fail("install failed")
		return fmt.Errorf("install new binary: %w", err)
	}

	_ = os.Remove(backupPath)
	spin.Success(fmt.Sprintf("binary installed to %s (%s)", su.executable, formatDuration(time.Since(start))))
	return nil
}

func defaultRunCommand(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			err = fmt.Errorf("%w: %s", err, msg)
		}
		return stdout.String(), err
	}

	return stdout.String(), nil
}

func defaultGoBuild(repoDir, output string) error {
	cmd := exec.Command("go", "build", "-o", output, "./cli")
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "GOMODCACHE="+filepath.Join(repoDir, ".gomodcache"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func shortSHA(sha string) string {
	if len(sha) >= 7 {
		return sha[:7]
	}
	return sha
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	return d.Round(time.Millisecond).String()
}

type progressSpinner struct {
	message    string
	enabled    bool
	stop       chan struct{}
	done       chan struct{}
	startTime  time.Time
	dotCounter int
}

func newSpinner(message string) *progressSpinner {
	s := &progressSpinner{
		message:   message,
		enabled:   isTerminal(os.Stdout),
		startTime: time.Now(),
	}
	if s.enabled {
		s.stop = make(chan struct{})
		s.done = make(chan struct{})
		fmt.Printf("\r%s %s %s%s", blue(statusPrefix), s.message, blue(string(spinnerFrames[0])), gray(""))
		go s.loop(1)
	} else {
		fmt.Printf("%s %s...\n", statusPrefix, s.message)
	}
	return s
}

func (s *progressSpinner) loop(startIndex int) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	i := startIndex
	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(s.startTime)
			dots := s.getEllipsis()
			fmt.Printf("\r%s %s %s %s", blue(statusPrefix), s.message, blue(string(spinnerFrames[i%len(spinnerFrames)])), gray(formatDuration(elapsed)+dots))
			i++
			s.dotCounter = (s.dotCounter + 1) % 4
		case <-s.stop:
			fmt.Print("\r")
			close(s.done)
			return
		}
	}
}

func (s *progressSpinner) getEllipsis() string {
	dots := ""
	for i := 0; i < s.dotCounter; i++ {
		dots += "."
	}
	for len(dots) < 3 {
		dots += " "
	}
	return dots
}

func (s *progressSpinner) stopSpinner() {
	if !s.enabled {
		return
	}
	select {
	case <-s.done:
		return
	default:
	}
	close(s.stop)
	<-s.done
}

func (s *progressSpinner) Success(summary string) {
	if s == nil {
		return
	}
	elapsed := time.Since(s.startTime)
	s.stopSpinner()
	if s.enabled {
		if summary != "" {
			fmt.Printf("\r%s %s %s %s (%s)\n", blue(statusPrefix), s.message, green("✓"), summary, gray(formatDuration(elapsed)))
		} else {
			fmt.Printf("\r%s %s %s (%s)\n", blue(statusPrefix), s.message, green("✓"), gray(formatDuration(elapsed)))
		}
	} else {
		if summary != "" {
			fmt.Printf("%s %s ✓ %s (%s)\n", statusPrefix, s.message, summary, formatDuration(elapsed))
		} else {
			fmt.Printf("%s %s ✓ (%s)\n", statusPrefix, s.message, formatDuration(elapsed))
		}
	}
}

func (s *progressSpinner) Fail(summary string) {
	if s == nil {
		return
	}
	elapsed := time.Since(s.startTime)
	s.stopSpinner()
	if s.enabled {
		if summary != "" {
			fmt.Printf("\r%s %s %s %s (%s)\n", blue(statusPrefix), s.message, red("✗"), summary, gray(formatDuration(elapsed)))
		} else {
			fmt.Printf("\r%s %s %s (%s)\n", blue(statusPrefix), s.message, red("✗"), gray(formatDuration(elapsed)))
		}
	} else {
		if summary != "" {
			fmt.Printf("%s %s ✗ %s (%s)\n", statusPrefix, s.message, summary, formatDuration(elapsed))
		} else {
			fmt.Printf("%s %s ✗ (%s)\n", statusPrefix, s.message, formatDuration(elapsed))
		}
	}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

func printSelfUpdateUsage() {
	fmt.Print(`Chauffeur Self-Update

Usage:
  chauf self-update   Update chauffeur by pulling the latest git changes and rebuilding the CLI binary.
`)
}
