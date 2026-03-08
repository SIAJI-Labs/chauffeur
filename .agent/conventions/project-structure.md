# Project Structure

## Repository Layout

```
chauffeur-v2/
├── cmd/
│   └── chauf/
│       └── main.go             # CLI entry point, version, command routing
├── internal/
│   ├── commands/               # CLI command implementations
│   │   ├── init.go
│   │   ├── info.go
│   │   ├── link.go
│   │   ├── links.go
│   │   ├── unlink.go
│   │   ├── install.go
│   │   ├── remove.go
│   │   ├── start.go
│   │   ├── stop.go
│   │   ├── restart.go
│   │   ├── status.go
│   │   ├── logs.go
│   │   ├── secure.go
│   │   ├── unsecure.go
│   │   ├── doctor.go
│   │   ├── clean.go
│   │   ├── migrate.go
│   │   ├── selfupdate.go
│   │   ├── config.go           # V2 new
│   │   ├── env.go              # V2 new
│   │   ├── autostart.go        # V2 new
│   │   └── update.go           # V2 new
│   ├── config/
│   │   ├── workspace.go        # Global config schema and loading
│   │   ├── project.go          # Project config schema and loading
│   │   └── validate.go         # Config validation
│   ├── installers/
│   │   ├── php.go              # PHP build from source
│   │   ├── php_legacy.go       # PHP 7.4/8.0 specific patches
│   │   ├── nginx.go            # Nginx build from source
│   │   ├── composer.go         # Composer PHAR download
│   │   └── common.go           # Shared download, checksum, extract utilities
│   ├── projects/
│   │   ├── manager.go          # Project CRUD operations
│   │   ├── slug.go             # Slug generation from directory name
│   │   └── detect.go           # Project type detection (Laravel, WordPress, generic)
│   ├── services/
│   │   ├── nginx.go            # Nginx process management
│   │   ├── php_fpm.go          # PHP-FPM process management
│   │   └── orchestrator.go     # Coordinate multiple services
│   ├── system/
│   │   ├── ports.go            # Port conflict detection and resolution
│   │   ├── dns.go              # dnsmasq config management
│   │   ├── systemd.go          # Systemd user service management (V2)
│   │   └── portforward.go      # iptables port forwarding state
│   ├── templates/
│   │   ├── nginx_site.go       # Nginx site config template rendering
│   │   ├── php_fpm.go          # PHP-FPM pool config template rendering
│   │   └── files/
│   │       ├── nginx-site.conf.tmpl
│   │       ├── nginx-site-ssl.conf.tmpl
│   │       ├── php-fpm-shared.conf.tmpl
│   │       └── php-fpm-dedicated.conf.tmpl
│   ├── workspace/
│   │   ├── paths.go            # All workspace path resolution
│   │   └── init.go             # Workspace directory scaffolding
│   └── lib/
│       ├── logging.go          # Structured logger with colors, spinners, progress
│       ├── download.go         # HTTP download with progress
│       ├── checksum.go         # SHA256/GPG verification
│       ├── ssl.go              # mkcert invocation helpers
│       ├── ports.go            # Port validation utilities
│       ├── input.go            # User input validation (domains, paths, versions)
│       ├── flags.go            # Common flag parsing helpers
│       └── table.go            # Table formatting for output
├── tests/
│   ├── integration/            # Integration tests (real filesystem)
│   │   ├── link_test.go
│   │   ├── install_test.go
│   │   └── helpers.go
│   └── fixtures/               # Test data and config fixtures
├── scripts/
│   ├── build.sh               # Build script
│   ├── test.sh                # Test script with coverage
│   └── release.sh             # Release preparation
├── .agent/                    # AI documentation (this directory)
├── .github/
│   └── workflows/
│       ├── ci.yml             # Test on PR
│       └── release.yml        # Release on tag
├── .goreleaser.yml            # Release build config
├── install.sh                 # Curl-based installer
├── go.mod
├── go.sum
├── README.md
└── CHANGELOG.md
```

---

## Package Responsibilities

### `cmd/chauf/main.go`

Entry point only. Responsibilities:
- Parse global flags (`--version`, `--help`)
- Route to command handler in `internal/commands/`
- Print usage for unknown commands
- Exit with appropriate code

**Never** put business logic here. Keep it thin.

### `internal/commands/`

One file per top-level command. Responsibilities:
- Parse command-specific flags
- Validate input (delegate to `lib/input.go`)
- Call internal packages to do work
- Output results through `lib.Logger`

**Pattern**:
```go
// commands/link.go
func RunLink(args []string) error {
    logger := lib.NewCommandLogger("link")
    // 1. Parse flags
    // 2. Validate input
    // 3. Call internal packages
    // 4. Log result
    return nil
}
```

### `internal/config/`

YAML configuration loading and validation. No side effects — pure read/write.

### `internal/installers/`

Service installation from source or download. Each installer is self-contained. Responsibilities:
- Download source/binary
- Verify checksum
- Compile or extract
- Configure and install to workspace path

### `internal/projects/`

Project registration CRUD. Manages `~/.chauffeur/projects/<slug>/config.yaml`.

### `internal/services/`

Process management for running services. Start, stop, restart, status check. Uses PID files and Unix signals.

### `internal/system/`

Host system integration that requires care:
- Port detection (avoid conflicts)
- DNS config (print commands, never execute)
- iptables (print commands, never execute)
- Systemd user services (safe — no root needed)

### `internal/templates/`

Config file generation using `text/template`. Templates in `files/*.tmpl`. No file I/O here — return rendered strings.

### `internal/workspace/`

Centralized path resolution. All workspace paths come from here. Never hardcode paths elsewhere.

```go
// workspace/paths.go
type Workspace struct {
    Root string  // ~/.chauffeur
}

func (w *Workspace) PHPDir(version string) string {
    return filepath.Join(w.Root, "php", version)
}

func (w *Workspace) ProjectDir(slug string) string {
    return filepath.Join(w.Root, "projects", slug)
}
// ... etc
```

### `internal/lib/`

Shared utilities used across all packages. No business logic. Pure utility functions.

---

## Dependency Rules

```
cmd/chauf/main.go
    → internal/commands/       (only package cmd imports directly)

internal/commands/
    → internal/config/
    → internal/projects/
    → internal/services/
    → internal/installers/
    → internal/system/
    → internal/workspace/
    → internal/lib/

internal/projects/
    → internal/config/
    → internal/workspace/
    → internal/lib/

internal/services/
    → internal/workspace/
    → internal/lib/

internal/installers/
    → internal/workspace/
    → internal/lib/

internal/system/
    → internal/workspace/
    → internal/lib/

internal/templates/
    → (no internal deps — pure template rendering)

internal/workspace/
    → (no internal deps — foundation package)

internal/lib/
    → (no internal deps — pure utilities)
```

**Rule**: No circular dependencies. Lower-level packages (`lib`, `workspace`) never import higher-level packages.

---

## Workspace Directory (Runtime)

The runtime workspace lives at `~/.chauffeur/` by default, overridable via `CHAUFFEUR_HOME` env var.

```
~/.chauffeur/
├── bin/
│   ├── chauf                       # CLI binary (optional, if self-managed)
│   └── shims/
│       ├── php                     # PHP version-aware shim
│       └── composer                # Composer shim
├── config/
│   └── chauffeur.yaml              # Global workspace config
├── projects/
│   └── <slug>/
│       ├── config.yaml             # Project config
│       ├── php-fpm.conf            # FPM config (dedicated only)
│       ├── php-fpm.sock            # FPM socket (dedicated only)
│       ├── php-fpm.pid             # FPM PID (dedicated only)
│       └── logs/
│           ├── php-fpm.log
│           └── php-fpm-slow.log
├── php/
│   └── <version>/
│       ├── bin/
│       │   ├── php
│       │   ├── php-cgi
│       │   └── php-fpm
│       ├── etc/
│       │   ├── php.ini
│       │   ├── php-fpm.conf        # Shared pool config
│       │   └── conf.d/
│       │       └── openssl.ini
│       ├── lib/php/extensions/
│       ├── runtime/php-fpm/        # Shared pool socket/PID
│       └── var/log/
├── nginx/
│   ├── bin/nginx
│   ├── etc/
│   │   ├── nginx.conf
│   │   ├── mime.types
│   │   ├── fastcgi_params
│   │   ├── sites-available/
│   │   │   ├── default.conf        # Catch-all 404
│   │   │   └── <slug>.conf
│   │   └── sites-enabled/
│   │       └── <slug>.conf -> ../sites-available/<slug>.conf
│   ├── certs/
│   │   ├── <domain>.crt
│   │   └── <domain>.key
│   └── logs/
│       ├── access.log
│       ├── error.log
│       └── nginx.pid
├── composer/
│   └── composer.phar
├── cache/
│   ├── php/                        # PHP source tarballs
│   ├── nginx/                      # Nginx source tarballs
│   └── composer/                   # Composer PHAR
├── logs/
│   ├── commands/                   # Per-command execution logs
│   └── chauffeur.log               # Main application log
└── system/
    ├── port-forwarding.json        # iptables rules state
    └── dns.json                    # DNS config state
```
