# Chauffeur

> Laravel Valet / Herd for Linux — per-project PHP services without containers.

Chauffeur manages nginx, PHP-FPM (7.4–8.4), and `.test` domain routing in a single user-space directory (`~/.chauffeur`). No Docker. No system pollution. One command per action.

## Quick Start

```bash
# Install
curl -sSL https://chauffeur.siaji.com/install | bash

# Set up workspace and services
chauf init
chauf install nginx php 8.3 composer

# Link your first project
cd ~/Projects/my-laravel-app
chauf link --secure

# Start everything
chauf start

# Visit https://my-laravel-app.test:8443
```

## Core Features

| Feature | Description |
|---------|-------------|
| **Workspace isolation** | All services and configs live in `~/.chauffeur/` — clean `rm -rf` uninstall |
| **Multi-version PHP** | PHP 7.4–8.4 simultaneously, per-project version pinning |
| **Zero-config domains** | `*.test` routing via dnsmasq — no `/etc/hosts` editing |
| **One-click SSL** | Locally trusted certificates via `mkcert` with SAN support |
| **Shared or dedicated FPM** | Resource-efficient shared pools or per-project isolation |
| **Multi-domain projects** | Multiple `.test` domains per project for multi-tenant dev |
| **Health diagnostics** | `chauf doctor` validates all dependencies with fix suggestions |
| **Smart caching** | PHP/nginx source tarballs cached — reinstalls are instant |

## Philosophy

- **Workspace-first** — everything in `~/.chauffeur/`, nothing touches system directories
- **Explicit control** — `chauf link` is the only registration path, no auto-scanning
- **Zero silent mutation** — commands that need `sudo` print the exact commands to run
- **Idempotent** — safe to re-run any command
- **Linux-native** — designed for Linux, not ported from macOS

## Architecture

```
~/.chauffeur/
├── bin/shims/          # php, composer — version-aware shims
├── config/             # Global configuration (chauffeur.yaml)
├── projects/<slug>/    # Per-project config, FPM pool, logs
├── php/<version>/      # Compiled PHP runtimes (one per version)
├── nginx/              # nginx binary, configs, certs, logs
├── composer/           # Composer PHAR
└── cache/              # Source tarballs — for fast reinstalls
```

Request flow:
```
Browser → *.test → dnsmasq → 127.0.0.1:8080 → nginx → Unix socket → PHP-FPM
```

## PHP-FPM Strategies

**Shared** (default) — one pool per PHP version, all projects using that version share it. Efficient for typical dev workloads.

**Dedicated** (`--dedicated-fpm`) — isolated pool per project. Better for projects needing custom FPM settings or strict isolation.

Both strategies coexist. `chauf status` shows them separately.

## Documentation

| Guide | Description |
|-------|-------------|
| [docs/commands/](./docs/commands/) | Complete CLI command reference |
| [docs/getting-started.md](./docs/getting-started.md) | Installation and first project walkthrough |
| [docs/php-fpm.md](./docs/php-fpm.md) | PHP-FPM strategies in depth |
| [docs/ssl.md](./docs/ssl.md) | SSL certificate setup |
| [docs/multi-domain.md](./docs/multi-domain.md) | Multi-domain project setup |
| [docs/troubleshooting.md](./docs/troubleshooting.md) | Common issues and fixes |

AI agent documentation lives in [.agent/](./.agent/index.md).

## Platform Support

| Distribution | Status |
|-------------|--------|
| Arch Linux (rolling) | Primary ✅ |
| Ubuntu 22.04+ | Primary ✅ |
| Debian 12+ | Primary ✅ |
| Fedora 39+ | Secondary ✅ |
| Rocky / AlmaLinux 9+ | Secondary ✅ |

**Architecture**: x86_64 (primary), ARM64 (secondary)

## Development

```bash
# Build
go build -o chauf ./cmd/chauf/

# Test
go test ./...

# Rebuild from repo
chauf self-update --dev
```

See [.agent/CONVENTIONS.md](./.agent/CONVENTIONS.md) for code standards and [.agent/IMPLEMENTATION_PLAN.md](./.agent/IMPLEMENTATION_PLAN.md) for the build roadmap.

## Status

V2 is a ground-up rewrite — organized codebase, new features (`chauf config`, `chauf env`, `chauf autostart`), and better test coverage. See the implementation plan for current progress.

---

*Inspired by [Laravel Valet](https://laravel.com/docs/valet) and [Beyond Code Herd](https://herd.laravel.com/).*
