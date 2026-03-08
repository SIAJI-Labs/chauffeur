# Chauffeur V2 — Product Requirements Document

## Project Overview

**Chauffeur** is a Linux-first CLI tool that provides per-project PHP development services — a Valet/Herd alternative designed for Linux developers. V2 is a ground-up rewrite that retains the proven architecture of V1 while fixing structural issues, adding missing features, and improving developer ergonomics.

### Name

**Chauffeur** — because it drives your local PHP projects, so you don't have to.

### Documentation

See [index.md](./index.md) for the complete documentation index.

| Guide | Description |
|-------|-------------|
| [SPECS.md](./SPECS.md) | Feature specifications |
| [CONVENTIONS.md](./CONVENTIONS.md) | Code conventions |
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | Implementation roadmap |
| [INTEGRATIONS.md](./INTEGRATIONS.md) | Service integrations |

---

## Mission Statement

> Provide Linux PHP developers with a Valet-like experience for per-project services, with workspace isolation, zero host mutation, and explicit control over every operation.

---

## Core Problems Chauffeur Solves

| Problem | Chauffeur Solution |
|---------|-------------------|
| Laravel Valet is macOS-only | Native Linux design, no macOS compatibility shims |
| Docker overhead for simple local dev | Lightweight, no containers — direct process management |
| System PHP version conflicts | Per-project PHP version isolation via shims |
| nginx/PHP config scattered everywhere | Everything in `~/.chauffeur/`, clean uninstall |
| Port 80/443 requires root | Ports 8080/8443 by default, optional iptables redirect |
| SSL certs are complicated locally | One command SSL via mkcert + SAN cert support |

---

## Core Features

### 1. Workspace Management

- Single user-space directory `~/.chauffeur/` contains all services
- `chauf init` bootstraps the full workspace
- `chauf info` shows workspace state and service health
- Clean uninstall with `chauf uninstall`

### 2. Project Linking

- `chauf link` registers the current directory as a project
- Generates nginx config with `.test` domain routing
- Assigns PHP version and FPM strategy at link time
- `chauf links` lists all registered projects
- `chauf unlink` removes project registration

### 3. Multi-PHP Version Support

- Supports PHP 7.4, 8.0, 8.1, 8.2, 8.3, 8.4 simultaneously
- PHP compiled from source into `~/.chauffeur/php/<version>/`
- Per-project PHP version override via `chauf php isolate <version>`
- PHP shim at `~/.chauffeur/bin/shims/php` auto-routes to correct version
- Composer shim uses Chauffeur PHP for version consistency

### 4. PHP-FPM Strategies

- **Shared FPM** (default): One PHP-FPM pool per PHP version, shared across projects
- **Dedicated FPM**: Isolated pool per project via `--dedicated-fpm` flag
- Strategies coexist in the same workspace
- Unix sockets for all PHP-FPM pools (no TCP overhead)

### 5. Multi-Domain Support

- Primary domain per project (`<slug>.test`)
- Additional alias domains via `--alias` flag
- Dynamic alias add/remove without relinking
- SAN SSL certificates covering all domains

### 6. SSL Certificates

- Local trusted SSL via `mkcert`
- `chauf secure` / `chauf unsecure` per project
- Auto-regeneration when aliases change
- Certificates stored in `~/.chauffeur/nginx/certs/`

### 7. Service Orchestration

- `chauf start` / `stop` / `restart` for nginx, PHP-FPM
- Project-scoped or workspace-wide service control
- `chauf status` with optional detail view
- `chauf logs` with filtering and follow mode

### 8. Configuration Management (V2 New)

- `chauf config show/set/validate/export` commands
- Eliminates manual YAML file editing
- JSON Schema validation for config files

### 9. Environment Management (V2 New)

- `chauf env list/set/unset/import/export` commands
- Per-project `.env` management integrated into workspace
- Secure storage of project-specific environment variables

### 10. Auto-Start via Systemd (V2 New)

- `chauf autostart enable/disable` commands
- User-level systemd services (no root required)
- Services restart on login automatically

### 11. Health Checking

- `chauf doctor` validates all dependencies and services
- Checks: system tools, PHP build deps, SSL, DNS, network, port conflicts
- `--fix` / `--auto-fix` flags for automated remediation
- Distribution-aware package install commands

### 12. Smart Caching

- PHP source tarballs cached in `~/.chauffeur/cache/`
- Composer PHAR cached after first download
- Checksum verification for all cached downloads
- `chauf clean cache` for cache management

---

## Design Philosophy

### 1. Workspace-First

All binaries, configs, sockets, logs, and runtime state live under `~/.chauffeur/`. No installation to `/usr`, `/etc`, or `/opt`. Clean uninstallation = `rm -rf ~/.chauffeur`.

### 2. Zero Host Mutation

When system changes are necessary (dnsmasq config, iptables rules), Chauffeur prints the exact command for the user to run — never executes system-level mutations silently.

### 3. Explicit Control

Projects are never auto-scanned. `chauf link` is the only registration mechanism. No background watchers. No automatic reconfiguration.

### 4. Idempotent Operations

All commands are safe to re-run. `chauf link` on an already-linked project updates rather than fails. `chauf install` skips already-installed services.

### 5. Minimal Dependencies

Zero external Go dependencies. CLI core uses only the standard library. This reduces attack surface, speeds compilation, and simplifies maintenance.

---

## Technical Architecture

### CLI Tool

- **Language**: Go 1.22+
- **Distribution**: Single compiled binary
- **External dependencies**: None (standard library only)
- **Config format**: YAML

### Managed Services

| Service | Source | Location |
|---------|--------|----------|
| PHP 7.4-8.4 | Compiled from source (php.net) | `~/.chauffeur/php/<version>/` |
| nginx | Compiled from source (nginx.org) | `~/.chauffeur/nginx/` |
| Composer | PHAR download (getcomposer.org) | `~/.chauffeur/composer/` |
| mkcert | Binary download (GitHub) | Host-provided or `~/.chauffeur/bin/` |
| dnsmasq | System package | Host-managed |

### Platform Support

- **Primary**: Arch Linux (rolling), Ubuntu 22.04+, Debian 12+
- **Secondary**: Fedora 39+, Rocky Linux 9+, AlmaLinux 9+
- **Architecture**: x86_64 (primary), ARM64 (secondary)

---

## Target Users

**Primary**:
- Individual PHP developers on Linux
- Developers migrating from macOS (Valet/Herd)
- Freelancers managing multiple client projects

**Non-Target**:
- Large teams (use Docker/Kubernetes instead)
- Production deployments
- Windows/macOS users

---

## V2 vs V1 Differences

| Area | V1 | V2 |
|------|----|----|
| Code organization | Commands in flat `commands/` dir | Clean architecture with internal packages |
| Config management | Manual YAML editing | `chauf config` command |
| Env management | Not implemented | `chauf env` command |
| Auto-start | Not implemented | Systemd user services |
| Update management | `chauf self-update` only | Full service update management |
| Testing | Unit tests per package | Integration test suite + unit tests |
| Error messages | Inconsistent | Structured, actionable error messages |
| Documentation | Scattered across `docs/` | Unified `.agent/` AI-first documentation |

---

## Success Metrics

- CLI startup time: < 100ms
- Service startup time: < 500ms (cold), < 100ms (warm)
- Test coverage: ≥ 80% across all packages
- Zero critical data loss bugs
- All commands work idempotently
