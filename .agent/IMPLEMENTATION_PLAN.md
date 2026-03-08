# Chauffeur V2 — Implementation Plan

## Overview

V2 is a ground-up rewrite of Chauffeur, preserving the proven workspace architecture and core features of V1 while restructuring the codebase for maintainability and adding missing features.

## Current Status

- **Phase**: 2 — Project Management
- **Stability**: Early development — Phase 1 complete, no project linking yet

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
- [x] Output/color library (`internal/lib/output.go`) — colors, print helpers, DirSize, FormatBytes
- [x] Port `internal/workspace/` path resolution from V1
- [x] Port `internal/config/` — config struct, YAML load (custom parser, no external deps)
- [x] Write `chauf init` command (workspace initialization, idempotent, --force, --quiet)
- [x] Write `chauf info` command (workspace status — services, projects, PHP, cache)
- [x] Basic `chauf --version` and `chauf help`
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
- [x] `chauf php isolate <version>` — pin project to PHP version (.chauffeur-php)
- [x] PHP shim at `~/.chauffeur/bin/shims/php` (written by chauf init)
- [x] Composer shim at `~/.chauffeur/bin/shims/composer` (written by chauf init)
- [x] Tests for installer utilities (download, checksum, extraction, legacy patches) — 29 tests

---

## Phase 2: Project Management

- [ ] Port `internal/projects/` — project config CRUD
- [ ] `chauf link` — register project with nginx config generation
  - [ ] `--php <version>` flag
  - [ ] `--secure` flag (SSL from the start)
  - [ ] `--dedicated-fpm` flag
  - [ ] `--alias <domain>` flag (multi-domain)
- [ ] `chauf links` — table display of all projects with status
- [ ] `chauf unlink` — remove project registration
  - [ ] `--alias <domain>` — remove specific alias
  - [ ] `--all` — remove all aliases then unlink
- [ ] `chauf php isolate <version>` — set per-project PHP
- [ ] `chauf secure` / `chauf unsecure` — SSL management
- [ ] Port nginx template generation from V1
- [ ] Port multi-domain nginx config from V1
- [ ] Tests for project linking, config management, nginx templates

---

## Phase 3: Service Orchestration

- [ ] Port `internal/services/` — service start/stop/restart
- [ ] `chauf start` — start nginx + PHP-FPM
  - [ ] `--project <path>` — start project-specific services only
  - [ ] `--all` — start all services
- [ ] `chauf stop` — stop services
  - [ ] `--project <path>` flag
  - [ ] `--all` flag
- [ ] `chauf restart [nginx|php|fpm]` — restart specific service
  - [ ] `--project <path>` flag
  - [ ] `--all` flag
- [ ] `chauf status` — service health display
  - [ ] `--detail` — verbose output with process counts, memory
  - [ ] `--project <path>` — project-specific status
- [ ] `chauf logs [nginx|php|access|error]` — log viewing
  - [ ] `--follow` — tail mode
  - [ ] `--level <level>` — filter by log level
- [ ] Tests for service orchestration with mock executors

---

## Phase 4: Health and Maintenance

- [ ] Port `commands/doctor.go` — comprehensive health checking
  - [ ] System tool checks (git, curl, tar, gcc, make, pkg-config)
  - [ ] PHP build dependency checks (libzip, libjpeg, etc.)
  - [ ] SSL checks (mkcert, openssl)
  - [ ] Network checks (port availability, dnsmasq)
  - [ ] `--fix` / `--auto-fix` flags
  - [ ] Distribution-aware package install suggestions
- [ ] Port `commands/clean.go` — workspace cleanup
  - [ ] `chauf clean cache` — remove download cache
  - [ ] `chauf clean logs` — remove old log files
  - [ ] `--dry-run` — show what would be cleaned
  - [ ] `--older-than <age>` — age-based cleanup
- [ ] Port `commands/migrate.go` — project migration
- [ ] Port `chauf self-update` — binary update
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

### Phase 2: Project Management
- [ ] Project CRUD
- [ ] `link/links/unlink` commands
- [ ] Nginx template generation
- [ ] Multi-domain support
- [ ] SSL support

### Phase 3: Service Orchestration
- [ ] Service lifecycle
- [ ] `start/stop/restart` commands
- [ ] `status/logs` commands

### Phase 4: Health & Maintenance
- [ ] `doctor` command
- [ ] `clean` command
- [ ] `migrate/self-update` commands

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
