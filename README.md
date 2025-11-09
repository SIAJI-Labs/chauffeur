# Chauffeur

> Linux-first valet-style CLI for per-project PHP services, inspired by Valet & Herd.

## Project Snapshot
| Item | Detail |
|------|--------|
| Purpose | Host-based CLI that installs nginx, PHP-FPM, and dnsmasq-managed `.test` routing inside `~/.chauffeur` so each project has isolated services. |
| Scope | Linux (Arch/Ubuntu/Debian focus). Manual `chauf link` registration per project; no DB/queue helpers. |
| Inspiration | [Laravel Valet](https://laravel.com/docs/valet) & [Beyond Code Herd](https://herd.laravel.com/). |
| Author | @siegg — single maintainer learning Go. |
| Status | Early adopter preview. Tested mainly on one Linux machine; expect rough edges. |

## Transparency & Expectations
- **AI-assisted codebase**: Most Go code was generated with AI coding agents following the contracts in `AGENTS.md`. Each feature is guided and reviewed manually, but there will be inconsistencies until more eyes are on the project.
- **Learning-in-public**: I am new to Go. Contributions, reviews, and bug reports from experienced Go developers are extremely welcome.
- **Experimental support**: Chauffeur currently "works on my machine". CI and cross-distro testing are still being built out, so please treat releases as experimental until we grow community coverage.
- **Documentation parity**: The AGENT handbook, README, and docs/TODO_STATUS are kept in sync. If you notice drift, open an issue or PR—accuracy matters more than marketing.

## Background & Inspiration
Chauffeur was born after migrating from macOS to Linux and missing the simplicity of Valet/Herd. Rather than containerizing everything, Chauffeur keeps services on the host but isolates them per project inside `~/.chauffeur`. dnsmasq handles `.test` domains, nginx proxies to per-project PHP-FPM pools, and shims ensure `php` understands which version to use based on your current directory.

Design themes borrowed from Valet/Herd:
- One command per action (`chauf link`, `chauf start`, `chauf stop`).
- Automatic nginx template selection (Laravel, WordPress, general).
- DNS-based routing with friendly domains like `myapp.test`.

## Core Principles
1. **Workspace-first**: Everything lives under `~/.chauffeur`—binaries, configs, sockets, logs, templates.
2. **Minimal host impact**: Only print `sudo` steps for `/etc` edits (dnsmasq, NetworkManager). Track change instructions so users can undo them.
3. **Manual project registration**: Chauffeur never auto-scans. `chauf link` registers the current working directory unless flags say otherwise.
4. **Idempotent commands**: Re-running `init`, `install`, `link`, etc. should be safe. Destructive actions require `--force`.
5. **Go-first implementation**: CLI is written in Go 1.22+, with helpers in `cli/lib` and `cli/internal/**`.
6. **Documentation synchronization**: README, docs/TODO_STATUS.md, and AGENTS.md must describe the current code behavior—no aspirational features marked as done.

## Feature Overview
| Area | Status | Notes |
|------|--------|-------|
| Workspace bootstrap (`chauf init`) | ✅ | Creates `~/.chauffeur` directories, default config, PATH guidance. |
| Project linking (`chauf link/links/unlink`) | ✅ | Registers projects, writes `project.yaml`, generates nginx configs. |
| PHP runtimes (`chauf php install/use/isolate`) | 🚧 | Core flows exist but need more testing and logging parity. |
| Service orchestration (`chauf start/stop/status`) | 🚧 | Basic process management exists; dnsmasq integration still evolving. |
| Composer integration | ✅ | Installs Composer PHAR tied to Chauffeur’s PHP shim. |
| Logging revamp | ✅ | All user-facing commands now route output through `lib.Logger`; only help/usage text stays raw. |
| Testing | 🚧 | Unit tests exist for link/links, but `go test ./...` currently fails; more coverage needed. |

## Architecture at a Glance
```
~/.chauffeur/
  bin/                 # shims (php, composer, nginx helpers)
  config/chauffeur.yaml
  projects/<slug>/
    project.yaml
    runtime/php-fpm/
    logs/
  php/<version>/       # compiled runtimes
  nginx/{bin,etc,sites-available,sites-enabled,conf.d,certs}
  logs/
```
- `php` shim: Detects whether you’re inside a linked project and selects the project’s PHP version; otherwise uses the global default (fallback 8.3).
- `chauf link`: Detects Laravel/WordPress/general layout and renders the appropriate nginx template.
- `chauf start`: Validates dnsmasq `.test` routing, offers auto-generated `sudo` commands if the host needs configuration, and manages iptables redirects recorded in `~/.chauffeur/system/port-forwarding.json`.

## Command Reference
| Command | Key Flags | Summary |
|---------|-----------|---------|
| `chauf init` | `--force`, `--quiet` | Bootstrap the workspace under `~/.chauffeur/`. |
| `chauf start` | `--project <path>`, `--all`, `--dry-run` | Start nginx/PHP-FPM (optionally all linked projects). |
| `chauf stop` | `--project <path>`, `--all`, `--dry-run` | Stop services and clean redirects. |
| `chauf status` | `[service-type]`, `--project`, `--detail`, `-v` | Inspect global or per-project services. |
| `chauf link` | `--site`, `--ssl`, `--php`, `--http-port`, `--https-port`, `--force` | Register PWD as a project and generate configs. |
| `chauf links` | — | Table of all registered projects. |
| `chauf unlink` | `--slug`, `--site`, `--project`, `--all`, `--force` | Remove registrations; defaults to current directory. |
| `chauf php install <ver>` | `--force`, `--no-ext`, `--from` | Install PHP runtimes into the workspace. |
| `chauf php use <ver>` | — | Set global default PHP version. |
| `chauf php isolate <ver>` | — | Pin the current linked project to a version. |
| `chauf remove <service> [ver]` | `--force` | Remove installed runtimes (php/nginx/composer). |
| `chauf install composer` | — | Download Composer PHAR + shim. |
| `chauf self-update` | `--dev` | Pull latest release or rebuild from current repo. |
| `chauf uninstall` | `--purge` | Remove workspace (optionally runtimes/caches). |
| `chauf info` | — | Show workspace paths, installed services, and port config. |

## Getting Started
1. **Install prerequisites**: Go 1.22+, git, curl, build tools (gcc/make/pkg-config), openssl headers.
2. **Clone & bootstrap** (installs `chauf` under `~/.chauffeur/bin`):
   ```bash
   git clone https://github.com/siaji/chauffeur.git
   cd chauffeur
   ./install.sh
   ```
3. **Install services**:
   ```bash
   chauf install php 8.3
   chauf install nginx
   chauf install composer
   ```
4. **Link a project**:
   ```bash
   cd /path/to/project
   chauf link --site myapp.test --php 8.3
   ```
5. **Start services & browse**:
   ```bash
   chauf start
   firefox http://myapp.test:8080
   ```

6. **Update the binary from source (optional, run inside repo)**:
   ```bash
   cd /path/to/chauffeur/repo
   chauf self-update --dev
   ```

Refer to `AGENTS.md` for the authoritative workspace layout, dnsmasq instructions, and logging spec.

## Development & Contribution
- **Preferred workflow**: Use `chauf self-update --dev` from the repo root to rebuild binaries; avoid `go build -o chauf` in-tree.
- **Logging**: Every command must use `lib.NewCommandLogger`. No raw `fmt.Printf` for user-facing output. Help converting legacy prints is very welcome.
- **Tests**: Always run `go test ./...` before opening a PR. Tests must isolate `HOME` via `t.TempDir()` to avoid touching real user state.
- **Documentation sync**: If you change behavior, update README, docs/TODO_STATUS.md, and AGENTS.md in the same PR.
- **Issues & PRs**: Please include distro info, Go version, `chauf info` output, and relevant log snippets (paths printed on failure) so we can reproduce problems quickly.

## Roadmap Highlights
Short-term focus (see `docs/TODO_STATUS.md` for the full queue):
1. Replace remaining `fmt.Printf` usage with the structured logger.
2. Stabilize `chauf status`, `start`, and `stop` with better service detection.
3. Expand automated tests to cover PHP installation flows and dnsmasq handling.
4. Document `sudo`-required dnsmasq/NetworkManager steps with reversible scripts.

## Acknowledgements
Thanks to the Valet and Herd teams for inspiring this workflow, to contributors providing feedback, and to the AI tooling that accelerates iteration. If you’d like to help drive Chauffeur toward a community-ready release, please open issues, share ideas, or send PRs.
