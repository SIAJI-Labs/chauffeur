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
	"github.com/siaji/chauffeur/cli/lib"
	"github.com/siaji/chauffeur/internal/utils"
)

const (
	defaultRepoURL   = "git@github.com:SIAJI-Labs/chauffeur.git"
	defaultGitBranch = "main"
)

var (
	runCommand = defaultRunCommand
	goBuild    = defaultGoBuild
	lookPath   = exec.LookPath
)

// OverrideSelfUpdateHooks lets tests swap self-update internals.
func OverrideSelfUpdateHooks(commandFn func(string, string, ...string) (string, error), buildFn func(string, string, string) error) (reset func()) {
	prevCommand := runCommand
	prevBuild := goBuild
	if commandFn != nil {
		runCommand = commandFn
	}
	if buildFn != nil {
		goBuild = buildFn
	}
	return func() {
		runCommand = prevCommand
		goBuild = prevBuild
	}
}

// runDevUpdate rebuilds the CLI binary from the current working directory
func runDevUpdate(logger *lib.Logger) error {
	// Verify we're in a valid git repository
	spin := lib.NewSpinner("self-update", "Verifying chauffeur repository")
	repoDir, err := os.Getwd()
	if err != nil {
		spin.Fail("get working directory")
		return fmt.Errorf("get working directory: %w", err)
	}

	// Check if this is a git repository
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		spin.Fail("not a git repository")
		return fmt.Errorf("current directory is not a git repository")
	}

	// Check if it has the required structure for a chauffeur repo
	requiredPaths := []string{
		"cli/main.go",
		"go.mod",
		"AGENTS.md",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(repoDir, path)
		if _, err := os.Stat(fullPath); err != nil {
			spin.Fail("invalid chauf repo structure")
			return fmt.Errorf("missing required file/directory: %s", path)
		}
	}
	spin.Success("validated chauf repo structure")

	// Get the current working directory info
	spin = lib.NewSpinner("self-update", "Preparing rebuild environment")
	start := time.Now()

	// Get current git HEAD for reference
	currentSHA, err := runCommand(repoDir, "git", "rev-parse", "HEAD")
	if err != nil {
		spin.Fail("get current commit")
		return fmt.Errorf("get current commit: %w", err)
	}
	currentSHA = strings.TrimSpace(currentSHA)

	// Get target executable path
	target := os.Getenv("CHAUF_SELF_UPDATE_TARGET")
	if target == "" {
		target, err = os.Executable()
		if err != nil {
			spin.Fail("determine executable path")
			return fmt.Errorf("determine executable path: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(target); err == nil {
			target = resolved
		}
	}

	// Ensure target directory exists
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		spin.Fail("prepare target directory")
		return fmt.Errorf("ensure binary directory: %w", err)
	}

	spin.Success(fmt.Sprintf("ready to rebuild (commit %s, %s)", shortSHA(currentSHA), formatDuration(time.Since(start))))

	// Build the new binary
	if err := buildFromSource(repoDir, target, currentSHA); err != nil {
		return err
	}

	logger.PrintSection("Dev Rebuild Summary")
	logger.Success("Dev rebuild complete", fmt.Sprintf("commit %s", shortSHA(currentSHA)))
	logger.Info(fmt.Sprintf("Built from: %s", repoDir))
	logger.Info(fmt.Sprintf("Installed to: %s", target))

	return nil
}

// buildFromSource builds the binary from the specified directory
func buildFromSource(repoDir, target, commitSHA string) error {
	// Create temporary file for new binary
	dir := filepath.Dir(target)
	tmpFile, err := os.CreateTemp(dir, "chauf-dev-build-*")
	if err != nil {
		return fmt.Errorf("prepare temporary binary: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	spin := lib.NewSpinner("self-update", fmt.Sprintf("Building from source (@%s)", shortSHA(commitSHA)))
	start := time.Now()

	buildTS := time.Now().UTC().Format(time.RFC3339)

	// Build the binary
	if err := goBuild(repoDir, tmpPath, buildTS); err != nil {
		spin.Fail("build failed")
		return fmt.Errorf("go build: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		spin.Fail("set permissions failed")
		return fmt.Errorf("set binary permissions: %w", err)
	}

	// Backup existing binary
	backupPath := target + ".bak"
	if err := os.Rename(target, backupPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		spin.Fail("backup failed")
		return fmt.Errorf("backup existing binary: %w", err)
	}

	// Install new binary
	if err := os.Rename(tmpPath, target); err != nil {
		_ = os.Rename(backupPath, target)
		spin.Fail("install failed")
		return fmt.Errorf("install new binary: %w", err)
	}

	// Clean up backup
	_ = os.Remove(backupPath)
	spin.Success(fmt.Sprintf("installed to %s (%s)", target, formatDuration(time.Since(start))))

	return nil
}

// RunSelfUpdate handles `chauf self-update`.
func RunSelfUpdate(args []string) error {
	logger := lib.NewCommandLogger("self-update")
	var isDev bool
	remainingArgs := []string{}

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printSelfUpdateUsage()
			return nil
		case "--dev":
			isDev = true
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	if isDev {
		return runDevUpdate(logger)
	}

	// Check if there's a newer version available
	currentVersion := getCLIVersion()
	if isUpdate, latestVersion, err := utils.IsUpdateAvailable(currentVersion); err == nil {
		if !isUpdate {
			logger.Info(fmt.Sprintf("You're already using the latest version (chauf %s)", strings.TrimPrefix(latestVersion, "v")))
			logger.Info("Run with --dev to rebuild from current source")
			return nil
		}
		logger.Info(fmt.Sprintf("Update available: chauffeur %s → %s", strings.TrimPrefix(currentVersion, "v"), strings.TrimPrefix(latestVersion, "v")))
	} else if err != nil {
		// If we can't check for updates, just continue with the update process
		logger.Warn("Unable to check for updates", err.Error())
	}

	updater, err := newGitSelfUpdater(logger)
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
	logger        *lib.Logger
	buildFromRepo string
}

func newGitSelfUpdater(logger *lib.Logger) (*gitSelfUpdater, error) {
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
		logger:        logger,
		buildFromRepo: repoDir, // Add this to track which repo we're building from
	}, nil
}

func (su *gitSelfUpdater) Execute() error {
	overallStart := time.Now()
	su.logger.Info("Starting self-update process...")

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
	su.logger.PrintSection("Summary")
	su.logger.Info(fmt.Sprintf("  Duration: %s", formatDuration(overallDuration)))
	su.logger.Info(fmt.Sprintf("  Previous: %s", func() string {
		if beforeSHA != "" {
			return shortSHA(beforeSHA)
		}
		return "fresh clone"
	}()))
	su.logger.Info(fmt.Sprintf("  Current: %s", shortSHA(afterSHA)))
	su.logger.Info(fmt.Sprintf("  Changes: %s", func() string {
		if created {
			return "fresh installation"
		}
		if changed {
			return "updated"
		}
		return "rebuilt only"
	}()))

	switch {
	case created:
		su.logger.Success("Self-update complete", fmt.Sprintf("fresh clone @ %s", shortSHA(afterSHA)))
	case changed:
		su.logger.Success("Self-update complete", fmt.Sprintf("commit %s", shortSHA(afterSHA)))
	default:
		su.logger.Success("Binary rebuilt", fmt.Sprintf("commit %s (no source changes)", shortSHA(afterSHA)))
	}
	return nil
}

func (su *gitSelfUpdater) ensureRepository() (bool, error) {
	if _, err := os.Stat(su.repoDir); errors.Is(err, os.ErrNotExist) {
		spin := lib.NewSpinner("self-update", "Cloning Chauffeur sources")
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
	su.logger.Info(fmt.Sprintf("Using existing sources at %s", su.repoDir))
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
	spin := lib.NewSpinner("self-update", fmt.Sprintf("Updating branch %s", su.branch))
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

	spin := lib.NewSpinner("self-update", "Building Chauffeur CLI")
	start := time.Now()

	buildTS := time.Now().UTC().Format(time.RFC3339)

	if err := goBuild(su.repoDir, tmpPath, buildTS); err != nil {
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

func defaultGoBuild(repoDir, output, buildTimestamp string) error {
	args := []string{"build"}
	var ldflags []string
	if ts := strings.TrimSpace(buildTimestamp); ts != "" {
		ldflags = append(ldflags, fmt.Sprintf("-X main.buildTimestamp=%s", ts))
	}
	if commit, err := revParseHEAD(repoDir); err == nil && commit != "" {
		ldflags = append(ldflags, fmt.Sprintf("-X main.buildCommit=%s", commit))
	}
	// Pass the build directory for version detection
	ldflags = append(ldflags, fmt.Sprintf("-X main.buildDir=%s", repoDir))
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}
	args = append(args, "-o", output, "./cli")

	// Set up Go module cache with proper permissions
	gomodCacheDir := filepath.Join(repoDir, ".gomodcache")
	if err := os.MkdirAll(gomodCacheDir, 0o755); err != nil {
		return fmt.Errorf("create Go module cache directory: %w", err)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "GOMODCACHE="+gomodCacheDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	// Fix permissions on the Go module cache directory and its contents
	if err := fixGomodCachePermissions(gomodCacheDir); err != nil {
		// Log warning but don't fail the build
		fmt.Fprintf(os.Stderr, "Warning: Failed to fix Go cache permissions: %v\n", err)
	}

	return nil
}

// fixGomodCachePermissions ensures the Go module cache directory and its contents
// have proper permissions for the user to clean up
func fixGomodCachePermissions(gomodCacheDir string) error {
	// Set permissions on the main cache directory
	if err := os.Chmod(gomodCacheDir, 0o755); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("chmod gomodcache directory: %w", err)
	}

	// Recursively set permissions on all files and subdirectories
	return filepath.Walk(gomodCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil // Skip if file doesn't exist (race condition)
			}
			return err
		}

		// Set appropriate permissions based on file type
		if info.IsDir() {
			if err := os.Chmod(path, 0o755); err != nil {
				return fmt.Errorf("chmod directory %s: %w", path, err)
			}
		} else {
			if err := os.Chmod(path, 0o644); err != nil {
				return fmt.Errorf("chmod file %s: %w", path, err)
			}
		}

		return nil
	})
}

func shortSHA(sha string) string {
	if len(sha) >= 7 {
		return sha[:7]
	}
	return sha
}

func revParseHEAD(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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

func printSelfUpdateUsage() {
	logger := lib.NewCommandLogger("self-update")
	logger.PrintBlock(`Chauffeur Self-Update

Usage:
  chauf self-update           Update chauffeur by pulling the latest git changes and rebuilding the CLI binary.
  chauf self-update --dev     Rebuild the CLI binary from the current working directory (must be a chauffeur repo).

Flags:
  --dev                       Rebuild from current directory if it's a valid chauffeur repository.
`)
}
