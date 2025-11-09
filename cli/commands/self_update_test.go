package commands

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type commandCall struct {
	dir    string
	name   string
	args   []string
	stdout string
	err    error
}

func TestSelfUpdateCloneAndBuild(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	binDir := filepath.Join(tmpHome, ".chauffeur", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}

	target := filepath.Join(binDir, "chauf")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	t.Setenv("CHAUF_SELF_UPDATE_TARGET", target)

	SetCLIVersion("0.1.0")
	t.Cleanup(func() { SetCLIVersion("dev") })

	lookPath = func(string) (string, error) { return "/usr/bin/git", nil }
	t.Cleanup(func() { lookPath = exec.LookPath })

	repoDir := filepath.Join(tmpHome, ".chauffeur", "src", "chauffeur")
	cloneArgs := []string{"clone", "--branch", defaultGitBranch, defaultRepoURL, repoDir}

	calls := []commandCall{
		{dir: "", name: "git", args: cloneArgs, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "1111111\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"status", "--porcelain"}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "1111111\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"fetch", "--tags", "--prune"}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"checkout", defaultGitBranch}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"pull", "--ff-only"}, stdout: "Updating\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "2222222\n", err: nil},
	}

	index := 0
	runCommand = func(dir, name string, args ...string) (string, error) {
		if index >= len(calls) {
			t.Fatalf("unexpected command: %s %v (dir %s)", name, args, dir)
		}
		call := calls[index]
		index++
		if dir != call.dir || name != call.name || !equalArgs(args, call.args) {
			t.Fatalf("command mismatch: got %s %v (dir %s), want %s %v (dir %s)", name, args, dir, call.name, call.args, call.dir)
		}
		return call.stdout, call.err
	}
	t.Cleanup(func() { runCommand = defaultRunCommand })

	goBuild = func(repoDir, output, _ string) error {
		return os.WriteFile(output, []byte("new"), 0o755)
	}
	t.Cleanup(func() { goBuild = defaultGoBuild })

	output, err := captureSelfUpdateOutput(func() error {
		return RunSelfUpdate(nil)
	})
	if err != nil {
		t.Fatalf("self update: %v", err)
	}

	if !strings.Contains(output, "Cloning Chauffeur sources ✓ into") {
		t.Fatalf("expected clone summary, got %q", output)
	}
	if !strings.Contains(output, "Updating branch main ✓ updated to 2222222") {
		t.Fatalf("expected update summary, got %q", output)
	}
	if !strings.Contains(output, "Building Chauffeur CLI ✓ binary installed to") {
		t.Fatalf("expected build summary, got %q", output)
	}
	if !strings.Contains(output, "Self-update complete") {
		t.Fatalf("expected completion summary, got %q", output)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("binary not replaced")
	}

	if _, err := os.Stat(target + ".bak"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected backup removed, err=%v", err)
	}
}

func TestSelfUpdateRequiresCleanRepo(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("CHAUF_SELF_UPDATE_TARGET", filepath.Join(tmpHome, "bin", "chauf"))

	repoDir := filepath.Join(tmpHome, ".chauffeur", "src", "chauffeur")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	lookPath = func(string) (string, error) { return "/usr/bin/git", nil }
	t.Cleanup(func() { lookPath = exec.LookPath })

	runCommand = func(dir, name string, args ...string) (string, error) {
		if name == "git" && equalArgs(args, []string{"status", "--porcelain"}) {
			return " M go.mod\n", nil
		}
		return "", nil
	}
	t.Cleanup(func() { runCommand = defaultRunCommand })

	goBuild = func(string, string, string) error { return nil }
	t.Cleanup(func() { goBuild = defaultGoBuild })

	err := RunSelfUpdate(nil)
	if err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("expected dirty repo error, got %v", err)
	}
}

func TestSelfUpdateUpToDate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	repoDir := filepath.Join(tmpHome, ".chauffeur", "src", "chauffeur")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	target := filepath.Join(tmpHome, ".chauffeur", "bin", "chauf")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	t.Setenv("CHAUF_SELF_UPDATE_TARGET", target)

	lookPath = func(string) (string, error) { return "/usr/bin/git", nil }
	t.Cleanup(func() { lookPath = exec.LookPath })

	calls := []commandCall{
		{dir: repoDir, name: "git", args: []string{"status", "--porcelain"}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "aaaaaaa\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"fetch", "--tags", "--prune"}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"checkout", defaultGitBranch}, stdout: "", err: nil},
		{dir: repoDir, name: "git", args: []string{"pull", "--ff-only"}, stdout: "Already up to date.\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "aaaaaaa\n", err: nil},
		{dir: repoDir, name: "git", args: []string{"rev-parse", "HEAD"}, stdout: "aaaaaaa\n", err: nil},
	}

	index := 0
	runCommand = func(dir, name string, args ...string) (string, error) {
		if index >= len(calls) {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		call := calls[index]
		index++
		if dir != call.dir || name != call.name || !equalArgs(args, call.args) {
			t.Fatalf("command mismatch: got %s %v (dir %s), want %s %v (dir %s)", name, args, dir, call.name, call.args, call.dir)
		}
		return call.stdout, call.err
	}
	t.Cleanup(func() { runCommand = defaultRunCommand })

	goBuild = func(repoDir, output, _ string) error {
		return os.WriteFile(output, []byte("rebuilt"), fs.FileMode(0o755))
	}
	t.Cleanup(func() { goBuild = defaultGoBuild })

	output, err := captureSelfUpdateOutput(func() error {
		return RunSelfUpdate(nil)
	})
	if err != nil {
		t.Fatalf("self update: %v", err)
	}
	if !strings.Contains(output, "Using existing sources") {
		t.Fatalf("expected reuse message, got %q", output)
	}
	if !strings.Contains(output, "Updating branch main ✓ already up to date") {
		t.Fatalf("expected up-to-date message, got %q", output)
	}
	if !strings.Contains(output, "Building Chauffeur CLI ✓ binary installed to") {
		t.Fatalf("expected rebuild summary, got %q", output)
	}
	if !strings.Contains(output, "Binary rebuilt at commit aaaaaaa") {
		t.Fatalf("expected up-to-date message, got %q", output)
	}
}

func equalArgs(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestSelfUpdateDevMode(t *testing.T) {
	// Setup temporary directory as a valid chauffeur repo
	tmpDir := t.TempDir()

	// Create required structure
	cliDir := filepath.Join(tmpDir, "cli")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir cli: %v", err)
	}

	// Create required files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/siaji/chauffeur\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cliDir, "main.go"), []byte("package main\nfunc main() {}"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Chauffeur Agents\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	// Change to the test directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to test repo: %v", err)
	}

	// Initialize git repo
	if err := os.MkdirAll(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
		t.Fatalf("create .git dir: %v", err)
	}

	// Setup target binary
	target := filepath.Join(tmpDir, "chauf")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("write existing binary: %v", err)
	}
	t.Setenv("CHAUF_SELF_UPDATE_TARGET", target)

	// Mock git command
	lookPath = func(string) (string, error) { return "/usr/bin/git", nil }
	t.Cleanup(func() { lookPath = exec.LookPath })

	runCommand = func(dir, name string, args ...string) (string, error) {
		if name == "git" && equalArgs(args, []string{"rev-parse", "HEAD"}) {
			return "abcdef123456\n", nil
		}
		return "", nil
	}
	t.Cleanup(func() { runCommand = defaultRunCommand })

	// Mock go build
	goBuild = func(repoDir, output, _ string) error {
		return os.WriteFile(output, []byte("new"), 0o755)
	}
	t.Cleanup(func() { goBuild = defaultGoBuild })

	// Test successful dev rebuild
	output, err := captureSelfUpdateOutput(func() error {
		return RunSelfUpdate([]string{"--dev"})
	})
	if err != nil {
		t.Fatalf("dev rebuild failed: %v", err)
	}

	if !strings.Contains(output, "validated chauf repo structure") {
		t.Fatalf("expected repo validation, got %q", output)
	}
	if !strings.Contains(output, "Building from source (@abcdef1)") {
		t.Fatalf("expected build message, got %q", output)
	}
	if !strings.Contains(output, "Dev rebuild complete") {
		t.Fatalf("expected rebuild complete, got %q", output)
	}

	// Verify binary was replaced
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("binary not replaced: got %q", string(data))
	}
}

func TestSelfUpdateDevModeInvalidRepo(t *testing.T) {
	// Test in non-git directory
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to test dir: %v", err)
	}

	output, err := captureSelfUpdateOutput(func() error {
		return RunSelfUpdate([]string{"--dev"})
	})

	if err == nil {
		t.Fatalf("expected error in non-git dir")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("expected git repo error, got %v", err)
	}
	if !strings.Contains(output, "not a git repository") {
		t.Fatalf("expected failure message, got %q", output)
	}

	// Test in git repo without required structure
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}

	output, err = captureSelfUpdateOutput(func() error {
		return RunSelfUpdate([]string{"--dev"})
	})

	if err == nil {
		t.Fatalf("expected error for invalid repo structure")
	}
	if !strings.Contains(err.Error(), "missing required file/directory: cli/main.go") {
		t.Fatalf("expected structure error, got %v", err)
	}
	if !strings.Contains(output, "invalid chauf repo structure") {
		t.Fatalf("expected structure failure message, got %q", output)
	}
}

func captureSelfUpdateOutput(fn func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()

	_ = w.Close()
	os.Stdout = oldStdout

	data, _ := io.ReadAll(r)
	_ = r.Close()

	return string(data), err
}
