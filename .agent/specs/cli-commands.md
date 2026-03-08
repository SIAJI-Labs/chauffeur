# CLI Command Reference

## Overview

The `chauf` binary provides all Chauffeur operations. Commands follow a `chauf <verb> [<object>] [flags]` pattern.

---

## Workspace Commands

### `chauf init`

Initialize the Chauffeur workspace.

```bash
chauf init
```

Creates `~/.chauffeur/` with all subdirectories, default configs, shim scripts.
Outputs PATH setup instructions.

---

### `chauf info`

Show workspace status and summary.

```bash
chauf info
chauf info --detail          # Include full paths
```

---

### `chauf uninstall`

Remove the entire workspace. Prompts for confirmation.

```bash
chauf uninstall
chauf uninstall --force      # Skip confirmation
```

---

## Installation Commands

### `chauf install`

Install a service into the workspace.

```bash
chauf install nginx
chauf install php 8.3        # Install specific PHP version
chauf install php             # Install default PHP version
chauf install composer
chauf install all             # Install nginx + default PHP + composer
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--force` | Reinstall even if already installed |
| `--no-cache` | Download fresh, skip cache |

---

### `chauf remove`

Remove a service from the workspace.

```bash
chauf remove php 8.1         # Remove specific PHP version
chauf remove nginx
chauf remove composer
```

---

## Project Commands

### `chauf link`

Register the current directory as a Chauffeur project.

```bash
chauf link
chauf link --php 8.1
chauf link --secure
chauf link --dedicated-fpm
chauf link --alias admin.my-project.test
chauf link --alias admin.my-project.test --alias api.my-project.test
chauf link --php 8.1 --secure --alias admin.my-project.test
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--php <version>` | string | global default | PHP version |
| `--secure` | bool | false | Enable SSL |
| `--dedicated-fpm` | bool | false | Dedicated FPM pool |
| `--alias <domain>` | string | — | Add alias domain (repeatable) |
| `--project <path>` | string | CWD | Target project path |

---

### `chauf links`

List all registered projects.

```bash
chauf links
chauf links --detail         # Show full paths, socket info
```

---

### `chauf unlink`

Unregister a project.

```bash
chauf unlink                             # Unlink project in CWD
chauf unlink --project /path/to/project  # Unlink specific project
chauf unlink --alias admin.project.test  # Remove only this alias
chauf unlink --all                       # Remove all aliases then unlink
```

---

### `chauf secure`

Enable SSL for a project.

```bash
chauf secure
chauf secure --project /path/to/project
```

---

### `chauf unsecure`

Disable SSL for a project.

```bash
chauf unsecure
chauf unsecure --project /path/to/project
```

---

## PHP Commands

### `chauf php`

PHP version management.

```bash
chauf php list               # List installed PHP versions
chauf php use 8.3            # Set global default PHP version
chauf php isolate 8.1        # Set PHP version for current project
chauf php install 8.4        # Alias for chauf install php 8.4
chauf php remove 8.0         # Alias for chauf remove php 8.0
```

---

## Service Commands

### `chauf start`

Start services.

```bash
chauf start                           # Start all services
chauf start nginx                     # Start only nginx
chauf start php 8.3                   # Start shared PHP 8.3 FPM
chauf start --project ./my-app        # Start services for one project
chauf start --all                     # Explicit all
```

---

### `chauf stop`

Stop services.

```bash
chauf stop                            # Stop all services
chauf stop nginx                      # Stop only nginx
chauf stop php 8.3                    # Stop shared PHP 8.3 FPM
chauf stop --project ./my-app         # Stop dedicated FPM for project
chauf stop --all                      # Stop all
```

---

### `chauf restart`

Restart services (graceful where possible).

```bash
chauf restart                         # Restart all services
chauf restart nginx                   # Reload nginx config (SIGHUP, no drops)
chauf restart php                     # Reload all PHP-FPM pools (SIGUSR2)
chauf restart php 8.3                 # Reload PHP 8.3 FPM only
chauf restart --project ./my-app      # Restart dedicated FPM for project
chauf restart --all                   # Restart everything
```

---

### `chauf status`

Show service status.

```bash
chauf status
chauf status --detail                 # Verbose with process counts, memory
chauf status --project ./my-app       # Status for one project's services
```

---

### `chauf logs`

View service logs.

```bash
chauf logs                            # nginx error log (last 50 lines)
chauf logs nginx                      # nginx error log
chauf logs access                     # nginx access log
chauf logs php                        # All PHP-FPM logs
chauf logs php 8.3                    # PHP 8.3 FPM log
chauf logs --project ./my-app         # Project-specific logs
chauf logs --follow                   # Tail mode
chauf logs --lines 100                # Show N lines
chauf logs --level error              # Filter by level
```

---

## Configuration Commands (V2)

### `chauf config`

Manage workspace or project configuration.

```bash
chauf config show                              # Show workspace config
chauf config show --project ./my-app          # Show project config
chauf config set nginx.http_port 8081          # Set workspace config
chauf config set php.default_version 8.3       # Set global default PHP
chauf config set fpm.dedicated true --project ./my-app  # Set project config
chauf config validate                          # Validate workspace config
chauf config validate --project ./my-app      # Validate project config
chauf config export                            # Export to JSON
chauf config export --project ./my-app
chauf config import config.yaml                # Import config
chauf config reset                             # Reset to defaults (confirm first)
chauf config reset --project ./my-app
```

---

### `chauf env` (V2)

Manage project environment variables.

```bash
chauf env list                                 # List all env vars
chauf env list --project ./my-app
chauf env set DATABASE_URL "mysql://..."       # Set env var
chauf env set --project ./my-app KEY=VALUE
chauf env unset DATABASE_URL                   # Remove env var
chauf env import .env                          # Import from .env file
chauf env import .env --project ./my-app
chauf env export                               # Export as .env format
chauf env export --project ./my-app
```

---

## Auto-Start Commands (V2)

### `chauf autostart`

Manage systemd auto-start.

```bash
chauf autostart enable                         # Enable all services
chauf autostart enable nginx                   # Enable nginx only
chauf autostart enable php 8.3                 # Enable PHP 8.3 FPM
chauf autostart disable                        # Disable all
chauf autostart disable nginx
chauf autostart disable php 8.3
chauf autostart status                         # Show systemd unit status
chauf autostart list                           # List all Chauffeur systemd units
```

---

## Update Commands (V2)

### `chauf update`

Update Chauffeur-managed services.

```bash
chauf update nginx                             # Update nginx to latest
chauf update php 8.3                           # Update PHP 8.3 patch version
chauf update composer                          # Update Composer PHAR
chauf update all                               # Update all services
chauf update --dry-run nginx                   # Show what would update
chauf update --backup nginx                    # Backup before updating
chauf update rollback nginx                    # Rollback to previous version
chauf update list-available                    # Check for available updates
chauf update list-available --service nginx
```

---

## Maintenance Commands

### `chauf doctor`

Run health checks.

```bash
chauf doctor                          # Full health check
chauf doctor --check-system           # System tool checks only
chauf doctor --check-php              # PHP build dependency checks
chauf doctor --check-ssl              # SSL checks
chauf doctor --check-network          # Port and DNS checks
chauf doctor --fix                    # Interactive fix mode
chauf doctor --auto-fix               # Automatic fix (non-interactive)
```

---

### `chauf clean`

Clean workspace artifacts.

```bash
chauf clean cache                     # Remove download cache
chauf clean logs                      # Remove old log files
chauf clean all                       # Remove cache + logs

chauf clean --dry-run                 # Show what would be removed
chauf clean --older-than 30d         # Only items older than 30 days
```

---

### `chauf migrate`

Move a project's config to a different workspace.

```bash
chauf migrate my-app ~/.chauffeur-work
chauf migrate my-app ~/.chauffeur-work --update-path /new/path
```

---

### `chauf self-update`

Update the `chauf` CLI binary.

```bash
chauf self-update                     # Update to latest release
chauf self-update --dev               # Rebuild from local repo (dev mode)
```

---

## Global Flags

Available on all commands:

| Flag | Description |
|------|-------------|
| `--version`, `-v` | Print version and exit |
| `--help`, `-h` | Show command help |
| `--workspace <path>` | Override workspace root (default: `~/.chauffeur`) |
| `--log-level <level>` | Set log verbosity: debug, info, warn, error |
| `--no-color` | Disable colored output |
| `--quiet`, `-q` | Suppress non-essential output |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (invalid flags, missing args) |
| 3 | Service error (FPM/nginx failed to start/stop) |
| 4 | Configuration error (invalid config, schema error) |
| 5 | Dependency error (missing host tool or library) |
