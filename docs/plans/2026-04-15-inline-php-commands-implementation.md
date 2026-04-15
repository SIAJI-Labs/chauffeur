# Inline PHP Commands Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `chauf php <version> <command> [args...]` to run commands with a specific PHP version inline.

**Architecture:** Detect if first arg is an installed PHP version, then execute the remaining args via the PHP shim with `CHAUFFEUR_PHP_VERSION` set.

**Tech Stack:** Go (chauffeur CLI)

---

## Task 1: Read existing `RunPHP` function

**Files:**
- Modify: `internal/commands/php_cmd.go`

**Step 1: Read current implementation**

Run: `cat internal/commands/php_cmd.go`

Understand:
- How `RunPHP` parses subcommands (list, use, isolate, install, remove)
- How it handles the help case
- The overall flow of argument parsing

---

## Task 2: Add `isInstalledPHPVersion` helper

**Files:**
- Modify: `internal/commands/php_cmd.go` (add after imports)

**Step 1: Add helper function**

Add this function after the imports:

```go
// isInstalledPHPVersion checks if a version string matches an installed PHP.
func isInstalledPHPVersion(version string) bool {
    root := workspace.Root()
    installed := installers.ListInstalledPHP(root)
    for _, v := range installed {
        if v == version {
            return true
        }
    }
    return false
}
```

**Step 2: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

---

## Task 3: Add `runPHPExtended` function

**Files:**
- Modify: `internal/commands/php_cmd.go` (add after `isInstalledPHPVersion`)

**Step 1: Add `runPHPExtended` function**

Add this function after `isInstalledPHPVersion`:

```go
// runPHPExtended executes a command with a specific PHP version.
// It sets CHAUFFEUR_PHP_VERSION and routes through the composer shim
// (for composer commands) or directly via PHP binary.
func runPHPExtended(version string, args []string) error {
    root := workspace.Root()
    
    // Verify PHP is installed
    phpBin := filepath.Join(root, "php", version, "bin", "php")
    if _, err := os.Stat(phpBin); err != nil {
        return fmt.Errorf("PHP %s not installed. Run: chauf install php %s", version, version)
    }

    if len(args) == 0 {
        return fmt.Errorf("Usage: chauf php <version> <command> [args...]")
    }

    cmd := args[0]
    
    // Set env for shim
    env := os.Environ()
    env = append(env, fmt.Sprintf("CHAUFFEUR_PHP_VERSION=%s", version))
    
    shimPath := filepath.Join(root, "bin", "shims")
    
    var execPath string
    var execArgs []string
    
    // Route based on command
    switch cmd {
    case "composer":
        // Use composer shim
        execPath = filepath.Join(shimPath, "composer")
        execArgs = args
    case "php", "php-fpm":
        // Direct PHP binary
        execPath = phpBin
        execArgs = args
    default:
        // Treat as PHP script (e.g., artisan, phpunit, wp-cli)
        execPath = phpBin
        execArgs = args
    }

    // Use syscall.Exec to replace current process
    return syscall.Exec(execPath, execArgs, env)
}
```

**Step 2: Add required imports**

The function uses `syscall`. Make sure it's in the imports:

```go
import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "syscall"  // ADD THIS

    "github.com/siegg/chauffeur/internal/installers"
    "github.com/siegg/chauffeur/internal/lib"
    "github.com/siegg/chauffeur/internal/projects"
    "github.com/siegg/chauffeur/internal/workspace"
)
```

**Step 3: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

---

## Task 4: Modify `RunPHP` to detect inline version

**Files:**
- Modify: `internal/commands/php_cmd.go` - the `RunPHP` function

**Step 1: Modify `RunPHP`**

Replace the beginning of `RunPHP` (lines 16-39):

```go
func RunPHP(args []string) error {
    if len(args) == 0 {
        return phpHelp()
    }

    // Inline version mode: chauf php <version> <command> [args...]
    // Check if first arg is an installed PHP version
    if len(args) >= 2 && isInstalledPHPVersion(args[0]) {
        return runPHPExtended(args[0], args[1:])
    }

    switch strings.ToLower(args[0]) {
    case "list", "ls":
        return phpList(args[1:])
    case "use":
        return phpUse(args[1:])
    case "install":
        // Alias: chauf php install 8.3 → chauf install php 8.3
        return RunInstall(append([]string{"php"}, args[1:]...))
    case "remove":
        // Alias: chauf php remove 8.3 → chauf remove php 8.3
        return RunRemove(append([]string{"php"}, args[1:]...))
    case "isolate":
        return phpIsolate(args[1:])
    case "--help", "-h", "help":
        return phpHelp()
    default:
        return fmt.Errorf("unknown php subcommand %q — run: chauf php --help", args[0])
    }
}
```

**Step 2: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

---

## Task 5: Build and test

**Step 1: Build entire project**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build -o chauf ./cmd/chauf/`

**Step 2: Test version detection (should show help/error)**

Run: `./chauf php 7.4`
Expected: Should show usage error or pass through to error

**Step 3: Test `chauf php 7.4 -v`**

Run: `./chauf php 7.4 -v`
Expected: Should output PHP 7.4 version

**Step 4: Test `chauf php 8.3 -v`**

Run: `./chauf php 8.3 -v`
Expected: Should output PHP 8.3 version

**Step 5: Test `chauf php 7.4 composer --version`**

Run: `./chauf php 7.4 composer --version`
Expected: Should show Composer version via PHP 7.4

**Step 6: Test `chauf php 8.3 composer --version`**

Run: `./chauf php 8.3 composer --version`
Expected: Should show Composer version via PHP 8.3

**Step 7: Test non-existent version**

Run: `./chauf php 9.9 composer --version`
Expected: Error "PHP 9.9 not installed"

**Step 8: Test existing subcommands still work**

Run: `./chauf php list`
Expected: Lists installed PHP versions

---

## Task 6: Commit

**Step 1: Stage and commit**

```bash
git add internal/commands/php_cmd.go
git commit -m "feat(php): add inline version selection with 'chauf php <version> <command>'

Allows running commands with a specific PHP version without changing
global default or project isolation:

  chauf php 7.4 composer install
  chauf php 8.3 artisan migrate
  chauf php 7.4 phpunit
"
```
