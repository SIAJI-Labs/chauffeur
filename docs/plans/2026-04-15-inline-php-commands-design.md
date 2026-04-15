# Design: Inline PHP Version Commands

**Date:** 2026-04-15
**Status:** Approved

## Problem

Currently, to run a command with a specific PHP version, users must either:
1. Change the global default: `chauf php use 7.4`
2. Isolate a project: `chauf php isolate <project> --php=7.4`
3. Use env variable: `CHAUFFEUR_PHP_VERSION=7.4 composer install`

There's no inline way to run a single command with a specific PHP version.

## Solution

Add inline PHP version selection to `chauf php`:

```bash
chauf php 7.4 composer install
chauf php 8.3 artisan migrate
chauf php 7.4 phpunit
chauf php 8.1 -v
```

## Implementation

### Command Detection

In `RunPHP()` (`php_cmd.go`), check if `args[0]` is an installed PHP version:

```go
func RunPHP(args []string) error {
    if len(args) == 0 {
        return phpHelp()
    }

    // Check if first arg is an installed PHP version (inline mode)
    if len(args) >= 2 {
        if version := args[0]; isInstalledPHPVersion(version) {
            return runPHPExtended(args[1:], version)
        }
    }

    // ... existing subcommand flow ...
}
```

### `runPHPExtended(version string, args []string) error`

1. Verify PHP version is installed
2. Set `CHAUFFEUR_PHP_VERSION=<version>` in environment
3. Determine if command is "php", "composer", or other
4. Execute via the appropriate shim with `syscall.Exec()` to replace current process

### Command Routing

| Input | Execution |
|-------|-----------|
| `chauf php 7.4 php -v` | `exec $CHAUF_PHP/7.4/bin/php -v` |
| `chauf php 7.4 composer install` | `exec $CHAUF/shims/composer install` (which calls PHP 7.4) |
| `chauf php 7.4 phpunit` | `exec $CHAUF/php/7.4/bin/php phpunit` |
| `chauf php 7.4 artisan migrate` | `exec $CHAUF/php/7.4/bin/php artisan migrate` |

### Error Handling

| Scenario | Error Message |
|----------|---------------|
| No command after version | `Usage: chauf php <version> <command> [args...]` |
| Version not installed | `PHP 7.4 not installed. Run: chauf install php 7.4` |
| Invalid version format | Falls through to subcommand (fails with "unknown php subcommand") |

### `isInstalledPHPVersion(version string) bool`

```go
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

## Files Modified

- `internal/commands/php_cmd.go` - Add inline version detection and execution

## Alternatives Considered

1. **`chauf with 7.4 composer install`** - More explicit but longer
2. **`chauf run --php=7.4 composer install`** - Flag-based, less natural for shell aliases
