# Chauffeur V2 — Implementation Plan

## Overview

V2 is a ground-up rewrite of Chauffeur, preserving the proven workspace architecture and core features of V1 while restructuring the codebase for maintainability and adding missing features.

## Current Status

- **Phase**: 3 — Service Orchestration ✅ COMPLETE
- **Stability**: Core workflow fully working — services start/stop/restart, DNS, SSL, port forwarding

## File Structure Summary

```
chauffeur-v2/
├── cmd/chauf/main.go
├── internal/
│   ├── commands/           # CLI command handlers
│   ├── config/             # Config loading and validation
│   ├── installers/         # PHP, nginx, Composer installers
│   ├── projects/           # Project CRUD and config management
│   ├── services/           # Service lifecycle management
│   ├── system/             # Host integration (ports, DNS, systemd)
│   ├── templates/          # Config template rendering
│   ├── workspace/          # Path resolution
│   └── lib/                # Shared utilities
├── tests/
├── .agent/
├── go.mod
├── go.sum
├── .goreleaser.yml
└── install.sh
```

---

## Phase 0: Project Bootstrap ✅ COMPLETE

- [x] Initialize Go module (`github.com/siegg/chauffeur`, Go 1.22)
- [x] Create directory structure (`cmd/`, `internal/`, `tests/`, etc.)
- [x] Set up `cmd/chauf/main.go` with version routing and command dispatch
- [x] Output/color library (`internal/lib/output.go`) — colors, print helpers, DirSize, FormatBytes, `lib.Verbose` flag, `lib.Step()` for verbose-only detail lines
- [x] Port `internal/workspace/` path resolution from V1
- [x] Port `internal/config/` — config struct, YAML load (custom parser, no external deps)
- [x] Write `chauf init` command (workspace initialization, idempotent, --force, --quiet)
- [x] Write `chauf info` command (workspace status — services, projects, PHP, cache)
- [x] Basic `chauf --version` and `chauf help`
- [x] Global `--verbose` / `-v` flag — pre-parsed in `main.go` before dispatch, sets `lib.Verbose`; works as `chauf --verbose <cmd>`, `chauf <cmd> --verbose`, or `chauf <cmd> -v` (note: `chauf -v` alone still prints version)
- [x] Write `chauf uninstall` command (service stop, workspace removal, manual cleanup hints)
- [x] Write `chauf self-update` command (--dev rebuild from repo, --dry-run, --version, release download)
- [x] Create `scripts/install.sh` for curl-based installation (in-repo / release / clone+build)
- [x] Create `Makefile` (build, dev, install, clean targets)
- [x] Create `.gitignore`
- [x] Spinners (`lib.NewSpinner`) + download progress bar (`lib.NewProgressReader`) — indeterminate spinner when `Content-Length` absent, clamps negative total to 0
- [x] Write tests for workspace init and config loading (8 tests, all passing)
- [x] Create `.goreleaser.yml` for release builds (linux amd64/arm64, combined checksums.txt)

---

## Phase 1: Core Services Installation ✅ COMPLETE

- [x] `installers/common.go` — BuildOpts, download+progress, checksum, tar extract, RunCmd
- [x] `installers/nginx.go` — nginx version resolve, download, build from source
- [x] `installers/php_legacy.go` — vendored OpenSSL 1.1.1w + 5 source patches for GCC 14 compat on 7.4/8.0 (libxml, openssl, scanf, gd_ctx, gd) + `-Wno-incompatible-pointer-types` CFLAGS
- [x] `installers/php.go` — PHP 7.4–8.4 compile, FPM config, imagick extension
- [x] `installers/composer.go` — Composer PHAR download
- [x] `chauf install nginx` command
- [x] `chauf install php <version>` command
- [x] `chauf install composer` command
- [x] `chauf remove <service>` command
- [x] `chauf php use <version>` — set global default PHP (updates chauffeur.yaml)
- [x] `chauf php list` — show installed versions with default marker
- [x] `chauf php isolate <version>` — pin project to PHP version (stored in project config, no dotfile in project dir)
- [x] PHP shim v3 at `~/.chauffeur/bin/shims/php` (written by chauf init) — resolves version by scanning `~/.chauffeur/projects/*/config.yaml` for CWD match, no project dotfiles
- [x] Composer shim at `~/.chauffeur/bin/shims/composer` (written by chauf init)
- [x] Tests for installer utilities (download, checksum, extraction, legacy patches) — 29 tests

---

## Phase 2: Project Management ✅ COMPLETE

- [x] `internal/projects/config.go` — Project struct + YAML read/write + CRUD (Save/Load/Delete/ListSlugs/ListAll/FindByPath/IsDomainInUse)
- [x] `internal/projects/slug.go` — GenerateSlug, DomainFromSlug, IsValidDomain
- [x] `internal/projects/detect.go` — Detect (laravel/wordpress/generic), DocumentRoot
- [x] `internal/projects/nginx.go` — RenderNginxConfig (HTTP+HTTPS templates), WriteNginxConfig, EnableNginxSite, DisableNginxSite, RemoveNginxConfig
- [x] `internal/projects/services.go` — ReloadNginx, IsNginxRunning, RunMkcert, MkcertInstalled, MkcertCAInstalled
- [x] `chauf link` — register project with nginx config generation
  - [x] `--php <version>` flag
  - [x] `--secure` flag (SSL from the start)
  - [x] `--dedicated-fpm` flag
  - [x] `--alias <domain>` flag (multi-domain, repeatable)
  - [x] Idempotent (updates existing project on re-link)
- [x] `chauf links` — table display of all projects with status + `--detail` flag
- [x] `chauf unlink` — remove project registration + confirmation prompt
  - [x] `--alias <domain>` — remove specific alias
  - [x] `--yes` — skip confirmation
- [x] `chauf php isolate <version>` — set per-project PHP (Phase 1, complete)
- [x] `chauf secure` / `chauf unsecure` — SSL management via mkcert
- [x] nginx template generation (HTTP + HTTPS) with per-type try_files routing
- [x] Multi-domain nginx config (all domains in single server_name directive)
- [x] Updated `chauf info` to use projects package (replaced stub parser)
- [x] `chauf info` projects section uses table format matching `chauf links` (header + separator + rows); `--detail` shows config path as subdued sub-row
- [x] 33 tests for project CRUD, slug, detection, nginx templates, domain conflict detection

---

## Phase 3: Service Orchestration ✅ COMPLETE

- [x] `internal/services/nginx.go` — NginxService: Start (TestConfig first), Stop (SIGQUIT + waitForExit), Reload (SIGHUP), PID, Uptime, MemoryMB, ConfigPath, BinaryPath
- [x] `internal/services/fpm.go` — FPMService: NewSharedFPM, NewDedicatedFPM, Start (--daemonize), Stop (SIGTERM), Reload (SIGUSR2), PID, Uptime, MemoryMB, ConfigPath, BinaryPath, SockPath
- [x] `internal/services/process.go` — readPIDFile, processAlive, waitForPID, waitForExit, waitForSocket, processUptime (/proc/pid/stat), processMemoryMB (/proc/pid/status), FormatUptime
- [x] `internal/services/ports.go` — IsPortAvailable (net.Listen), FindProcessOnPort (/proc/net/tcp inode walk)
- [x] `internal/services/manager.go` — Manager: StartAll (FPM→socket wait→nginx), StopAll (nginx→FPM), AllFPM (shared sorted + dedicated), ReloadAll
- [x] `internal/system/portforward.go` — iptables state tracking, PortForwardingCommands, IsPortForwardingActive
- [x] `chauf start` — start nginx + PHP-FPM in correct order (shared FPM → dedicated FPM → nginx); `--verbose` shows workspace path, ports, start order, per-service binary/config/socket paths
  - [x] Socket readiness wait before nginx starts
  - [x] Port conflict detection with process identification
  - [x] `--project <path>` — start project-specific services only
  - [x] Active domain URLs printed on success
- [x] `chauf stop` — graceful shutdown (nginx SIGQUIT → FPM SIGTERM, 30s timeout)
  - [x] `--project <path>` — stop dedicated FPM for specific project
- [x] `chauf restart [nginx|php|fpm [version]]` — zero-downtime reload
  - [x] Per-service progress output (label + PID after reload)
  - [x] `--project <path>` — reload project-specific FPM pool
- [x] `chauf status` — service health table (status, PID, uptime, memory)
  - [x] `--detail` — verbose output with config/socket paths
  - [x] `--project <path>` — project-specific status with FPM pool sharing info
  - [x] Installed-but-unused PHP versions shown as stopped
- [x] `chauf logs [nginx|access|php [version]]` — log viewing
  - [x] `--follow` — tail mode (200ms poll, Ctrl+C to stop)
  - [x] `--lines <n>` — number of lines to show (default 50)
  - [x] `--project <path>` — project-specific nginx logs
- [x] `chauf info` — updated with `(no projects)` tag for unused PHP-FPM pools
- [x] `chauf init` — DNS setup detection (NM-managed vs standalone dnsmasq) + port forwarding instructions
  - [x] Detects correct dnsmasq config path (`/etc/NetworkManager/dnsmasq.d/` vs `/etc/dnsmasq.d/`)
  - [x] Detects systemd-resolved stub conflict
  - [x] iptables port forwarding commands (80→8080, 443→8443)
- [x] Binary path fixes: nginx `sbin/nginx`, php-fpm `sbin/php-fpm` (not `bin/`)
- [x] `http2 on;` removed from nginx HTTPS template (not built into installed nginx)
- [x] `chauf self-update --dev` — timestamp suffix on dev builds (`<commit>-YYYYMMDD-HHMM`)

---

## Phase 4: Health and Maintenance ✅ PARTIAL

- [x] `commands/doctor.go` — comprehensive health checking
  - [x] System tool checks (git, curl, tar, gcc, make, pkg-config, autoconf, bison)
  - [x] PHP build dependency checks (libzip, libjpeg, libpng, freetype2, libxml2, libcurl, zlib, readline, libxslt, gmp, openssl, ImageMagick)
  - [x] SSL checks (mkcert, openssl, mkcert CA, cert directory)
  - [x] Network checks (iptables port forwarding, port availability)
  - [x] DNS checks (dnsmasq binary, chauffeur.conf, .test resolution)
  - [x] `--fix` flag — prints commands, never runs them
  - [x] `--auto-fix` flag — executes fix commands via `sh -c`, stops on first failure per block, skips comments
  - [x] `--check-deps / --check-php / --check-ssl / --check-network / --check-dns` — individual section flags
  - [x] Distribution-aware package install suggestions (arch/debian/fedora)
  - [x] NM-managed dnsmasq detection (reads `/etc/NetworkManager/conf.d/`)
  - [x] DNS resolution test via direct UDP dial to `127.0.0.1:53`
  - [x] systemd-resolved offline detection: `isResolvedInNSS()` checks `nsswitch.conf` for `resolve` module
  - [x] NSS offline fallback check: detects `[!UNAVAIL=return]` blocking TRYAGAIN passthrough
  - [x] Fix: `sed` changes nsswitch.conf `resolve [!UNAVAIL=return]` → `[NOTFOUND=return]` so offline TRYAGAIN falls through to `dns` → resolv.conf → dnsmasq → `.test` works offline
  - [x] Global resolved routing: `resolved.conf.d/chauffeur.conf` routes `.test` to dnsmasq when online
- [x] `commands/clean.go` — workspace cleanup
  - [x] `chauf clean cache` — remove download cache
  - [x] `chauf clean logs` — remove nginx and PHP-FPM log files
  - [x] `chauf clean all` — run all clean operations
  - [x] `--dry-run` — show what would be cleaned
  - [x] `--older-than <age>` — age-based cleanup (supports `7d`, `24h`, `30m`)
- [ ] Port `commands/migrate.go` — project migration
- [ ] Tests for doctor checks (with mocked system commands)

---

## Phase 5: V2 New Features

### Configuration Management

- [ ] `chauf config show [--project <slug>]` — display current config
- [ ] `chauf config set <key> <value> [--project <slug>]` — update config value
- [ ] `chauf config validate [--project <slug>]` — validate config schema
- [ ] `chauf config export [--project <slug>]` — export to JSON/YAML
- [ ] `chauf config import <file> [--project <slug>]` — import config
- [ ] `chauf config reset [--project <slug>]` — reset to defaults
- [ ] JSON Schema validation for workspace and project configs

### Environment Management

- [ ] `chauf env list [--project <slug>]` — list env vars
- [ ] `chauf env set <key> <value> [--project <slug>]` — set env var
- [ ] `chauf env unset <key> [--project <slug>]` — remove env var
- [ ] `chauf env import <file>` — import from `.env` file
- [ ] `chauf env export [--project <slug>]` — export to `.env` format
- [ ] Integration with nginx `fastcgi_param` injection

### Auto-Start via Systemd

- [ ] `chauf autostart enable [--service nginx|php]` — enable systemd service
- [ ] `chauf autostart disable [--service]` — disable systemd service
- [ ] `chauf autostart status` — show systemd service state
- [ ] User-level systemd unit file generation (`~/.config/systemd/user/`)
- [ ] No root required (systemctl --user)
- [ ] Tests for systemd unit file generation

### Service Update Management

- [ ] `chauf update <service> [--dry-run] [--backup]` — update specific service
- [ ] `chauf update all [--dry-run] [--backup]` — update all services
- [ ] `chauf update rollback <service> <version>` — rollback to previous version
- [ ] `chauf update list-available [--service]` — check for available updates
- [ ] Backup before update (tar archive of service dir)

---

## Phase 6: Documentation Site

- [ ] Port `sites/` Next.js documentation site from V1
- [ ] Update all command references to match V2 CLI
- [ ] Add V2 new features documentation
- [ ] Deploy to `https://chauffeur.siaji.com/docs`

---

## Implementation Checklist Summary

### Phase 0: Bootstrap ✅
- [x] Go module init
- [x] Directory structure
- [x] Output/color library (spinners deferred to Phase 1)
- [x] Workspace paths
- [x] Config loading
- [x] `chauf init` + `chauf info`
- [x] `chauf uninstall` + `chauf self-update`
- [x] `scripts/install.sh` + `Makefile`
- [x] Tests (8 passing)
- [x] `.goreleaser.yml`

### Phase 1: Installation ✅
- [x] Nginx installer
- [x] PHP installer (7.4–8.4)
- [x] PHP legacy patches (5 patches) + vendored OpenSSL + GCC 14 CFLAGS
- [x] Composer installer
- [x] PHP + Composer shims (written by chauf init)
- [x] `install/remove/php` commands

### Phase 2: Project Management ✅
- [x] Project CRUD
- [x] `link/links/unlink` commands
- [x] Nginx template generation (HTTP + HTTPS)
- [x] Multi-domain support (aliases)
- [x] SSL support (mkcert)
- [x] Tests (33 passing)

### Phase 3: Service Orchestration ✅
- [x] Service lifecycle (nginx + PHP-FPM)
- [x] `start/stop/restart` commands
- [x] `status/logs` commands
- [x] DNS setup detection in `chauf init`
- [x] iptables port forwarding (80→8080, 443→8443)
- [x] PHP shim v3 — project config lookup, no dotfiles

### Phase 4: Health & Maintenance
- [x] `doctor` command
- [x] `clean` command
- [ ] `migrate` command

### Phase 5: V2 New Features
- [ ] `config` command
- [ ] `env` command
- [ ] `autostart` command
- [ ] `update` command

### Phase 6: Docs Site
- [ ] Documentation site update

---

## Priority Order

1. Phase 0 → Get a working `chauf init` binary
2. Phase 1 → Install nginx and PHP (core dependency)
3. Phase 2 → Link projects (core value proposition)
4. Phase 3 → Start/stop/status (make it usable)
5. Phase 4 → Doctor + clean (stability)
6. Phase 5 → V2 new features (differentiators)
7. Phase 6 → Documentation site (polish)
