# Chauffeur

> Linux-first valet-style CLI for per-project PHP services, inspired by Valet & Herd.

## Project Snapshot
| Item | Detail |
|------|--------|
| Purpose | Host-based CLI that installs nginx, PHP-FPM, and dnsmasq-managed `.test` routing inside `~/.chauffeur` so each project has isolated services. |
| Scope | Linux (Arch/Ubuntu/Debian focus). Manual `chauf link` registration per project; no DB/queue helpers. |
| Inspiration | [Laravel Valet](https://laravel.com/docs/valet) & [Beyond Code Herd](https://herd.laravel.com/). |
| Author | @si-aji — single maintainer learning Go. |
| Status | Early adopter preview. Tested mainly on one Linux machine; expect rough edges. |

## Transparency & Expectations
- **AI-assisted codebase**: Most Go code was generated with AI coding agents following the contracts in `AGENTS.md`. Each feature is guided and reviewed manually, but there will be inconsistencies until more eyes are on the project.
- **Learning-in-public**: I am new to Go. Contributions, reviews, and bug reports from experienced Go developers are extremely welcome.
- **Experimental support**: Chauffeur currently "works on my machine". CI and cross-distro testing are still being built out, so please treat releases as experimental until we grow community coverage.
- **Documentation parity**: The AGENT handbook, README, and docs/TODO_STATUS are kept in sync. If you notice drift, open an issue or PR—accuracy matters more than marketing.

## Background & Inspiration
Chauffeur was born after migrating from macOS to Linux and missing the simplicity of Valet/Herd. Rather than containerizing everything, Chauffeur keeps services on the host but isolates them per project inside `~/.chauffeur`. dnsmasq handles `.test` domains, nginx proxies to per-project PHP-FPM pools, and shims ensure `php` understands which version to use based on your current directory.

Design themes borrowed from Valet/Herd:
- One command per action (`chauf link`, `chauf start`, `chauf stop`, `chauf restart`).
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
| PHP runtimes (`chauf php install/use/isolate`) | ✅ | Full PHP 7.4-8.4 support with smart caching; GD extension infrastructure in place (modern PHP ✅, legacy PHP 🚧); additional harness tests coming soon. |
| Service orchestration (`chauf start/stop/status/restart`) | ✅ | Full process management with service-specific and project-specific restarts; dnsmasq integration working. |
| Composer integration | ✅ | Installs Composer PHAR tied to Chauffeur's PHP shim. |
| Logging revamp | ✅ | All user-facing commands now route output through `lib.Logger`; only help/usage text stays raw. |
| Smart caching system | ✅ | Universal download cache with auto-detection and user control. |
| Testing | ✅ | Comprehensive test coverage across all packages including unit tests for installers, logging, projects, services, system, templates, and nginx templates; integration tests cover CLI workflows; CI enforces `go test ./...` on PRs to `main`. |

## Architecture at a Glance
```
~/.chauffeur/
  bin/                 # shims (php, composer, nginx helpers)
  config/chauffeur.yaml
  cache/               # download cache for faster reinstallation
  projects/<slug>/
    project.yaml
    runtime/php-fpm/
    logs/
  php/<version>/       # compiled runtimes
  nginx/{bin,etc,sites-available,sites-enabled,conf.d,certs}
  logs/
```
- `php` shim: Detects whether you're inside a linked project and selects the project's PHP version; otherwise uses the global default (fallback 8.3).
- `chauf link`: Detects Laravel/WordPress/general layout and renders the appropriate nginx template with project-level FPM control (shared by default, `--dedicated-fpm` for isolation).
- `chauf start`: Validates dnsmasq `.test` routing, offers auto-generated `sudo` commands if the host needs configuration, and manages iptables redirects recorded in `~/.chauffeur/system/port-forwarding.json`.

## PHP-FPM Architecture

Chauffeur provides **project-level PHP-FPM control** to balance resource efficiency and isolation:

### Shared FPM (Default)
- **Resource efficient**: Multiple projects share the same PHP-FPM pool per PHP version
- **Default behavior**: `chauf link` creates shared FPM unless `--dedicated-fpm` is specified
- **Example**: 10 projects using PHP 8.3 = 1 shared PHP-FPM process
- **Socket path**: `~/.chauffeur/php/8.3/runtime/php-fpm/php-fpm.sock`

### Dedicated FPM (Optional)
- **Maximum isolation**: Each project gets its own PHP-FPM pool
- **Usage**: `chauf link --dedicated-fpm` for critical projects needing custom configuration
- **Example**: 1 project = 1 dedicated PHP-FPM process
- **Socket path**: `~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.sock`

### Mixed Strategy Support
- **Flexible workspace**: Mix shared and dedicated FPM in the same environment
- **Automatic routing**: nginx automatically routes to the correct socket based on project configuration
- **Clear status**: `chauf status` shows global (shared) and project-specific (dedicated) services separately

## Command Reference
| Command | Key Flags | Summary |
|---------|-----------|---------|
| `chauf init` | `--force`, `--quiet` | Bootstrap the workspace under `~/.chauffeur/`. |
| `chauf start` | `--project <path>`, `--all`, `--dry-run` | Start nginx/PHP-FPM (optionally all linked projects). |
| `chauf stop` | `--project <path>`, `--all`, `--dry-run` | Stop services and clean redirects. |
| `chauf restart` | `[service-type]`, `--project <path>`, `--all`, `--dry-run` | Restart specific services, projects, or all services. |
| `chauf status` | `[service-type]`, `--project`, `--detail`, `-v` | Inspect global or per-project services. |
| `chauf link` | `--site`, `--ssl`, `--php`, `--dedicated-fpm`, `--http-port`, `--https-port`, `--force` | Register PWD as a project and generate configs (shared FPM by default). |
| `chauf links` | — | Table of all registered projects. |
| `chauf unlink` | `--slug`, `--site`, `--project`, `--all`, `--force` | Remove registrations; defaults to current directory. |
| `chauf install <service> [ver]` | `--force`, `--local`, `--no-cache` | Install services with intelligent caching (php, composer, nginx). |
| `chauf php install <ver>` | `--force`, `--no-ext`, `--from` | Install PHP runtimes into the workspace. |
| `chauf php use <ver>` | — | Set global default PHP version. |
| `chauf php isolate <ver>` | — | Pin the current linked project to a version. |
| `chauf remove <service> [ver]` | `--force` | Remove installed runtimes with cache management (php/nginx/composer). |
| `chauf install composer` | — | Download Composer PHAR + shim. |
| `chauf self-update` | `--dev` | Pull latest release or rebuild from current repo. |
| `chauf uninstall` | `--purge` | Remove workspace (optionally runtimes/caches). |
| `chauf info` | — | Show workspace paths, installed services, port config, plus GitHub release/commit drift checks. |

## Getting Started
1. **Install prerequisites**: Go 1.22+, git, curl, build tools (gcc/make/pkg-config), openssl headers.
2. **Clone & bootstrap** (installs `chauf` under `~/.chauffeur/bin`):
   ```bash
   git clone https://github.com/siaji/chauffeur.git
   cd chauffeur
   ./install.sh
   ```
3. **Install services** (with intelligent caching):
   ```bash
   chauf install php 8.3        # First download - auto-cached for future
   chauf install nginx           # Instant if cached, downloads if not
   chauf install composer        # Reuses cached PHAR when available
   ```
4. **Link projects** (shared FPM by default, dedicated when needed):
   ```bash
   cd ~/simple-project          # Shared FPM (resource efficient)
   chauf link

   cd ~/production-app          # Dedicated FPM (isolated)
   chauf link --dedicated-fpm --php 8.1

   cd ~/legacy-project         # Dedicated FPM (custom config needs)
   chauf link --dedicated-fpm --php 7.4
   ```
5. **Start services & browse**:
   ```bash
   chauf start
   firefox http://project-name.test:8080
   ```

`chauf link` automatically detects when Chauffeur’s nginx instance already owns the configured HTTP/HTTPS ports and simply restarts it so the new site configuration is loaded—no more prompts to pick new ports while services are running. Likewise, `chauf unlink` removes the generated nginx config, restarts nginx when it’s active, and the built-in catch‑all server returns a 404 for any unlinked domain so the site disappears immediately.

6. **Update the binary from source (optional, run inside repo)**:
   ```bash
   cd /path/to/chauffeur/repo
   chauf self-update --dev
   ```

## Smart Caching System

Chauffeur includes a universal intelligent caching system that dramatically speeds up service installations and gives users control over cached downloads.

### How It Works

**Installation Priority:**
1. **Local config paths** (for `--local` flag usage)
2. **Universal cache** (`~/.chauffeur/cache/`)
3. **Download from remote** (fallback when no cache available)

**Cache Management:**
- **Auto-caching**: Successful downloads are automatically cached unless `--no-cache` is used
- **Smart reuse**: Cached files are detected and reused instantly on subsequent installations
- **User control**: `chauf remove` prompts users to keep or remove cached files
- **Storage**: All cached files stored in `~/.chauffeur/cache/`

### Cache Files by Service

| Service | Cache Files | Example Names |
|---------|-------------|---------------|
| **PHP** | Source tarballs | `php-8.3.27.tar.gz`, `php-8.4.14.tar.gz` |
| **Composer** | PHAR binaries & checksums | `composer.phar`, `composer-2.8.4.phar`, `.sha256` files |
| **Nginx** | Source tarballs | `nginx-1.29.3.tar.gz`, `nginx-1.28.2.tar.gz` |

### Installation Options

```bash
# Standard installation with auto-caching
chauf install php 8.3          # Downloads and caches for next time
chauf install php 8.3          # Instant - reuses cached file

# Skip caching (useful for testing)
chauf install --no-cache nginx   # Download without caching

# Use local tarball (advanced usage)
chauf install php 8.3 --local    # Prompt for local tarball path

# Force reinstall (ignores cache for download)
chauf install composer --force   # Fresh download, updates cache
```

### Cache Management

```bash
# Remove with cache prompt
chauf remove php 8.3           # Asks: keep or remove cached php-8.3.27.tar.gz?
chauf remove composer          # Asks: keep or remove cached composer.phar?
chauf remove nginx             # Asks: keep or remove cached nginx-1.29.3.tar.gz?

# Force remove without prompts (keeps cache)
chauf remove --force nginx     # Removes nginx installation, preserves cache

# Manual cache inspection
ls -la ~/.chauffeur/cache/    # View all cached downloads
du -sh ~/.chauffeur/cache/    # See cache size usage
```

### Cache Behavior Explained

**Keep Cached Files (Default):**
- ✅ Faster future installations (no re-downloads)
- ✅ Saves bandwidth and time
- ⚠️ Uses disk space (typically 20-40MB per service, varies by versions cached)

**Remove Cached Files:**
- ✅ Frees up disk space
- ✅ Fresh downloads ensure latest versions
- ⚠️ Slower reinstallation (downloads again)

**Universal Intelligence:**
- Same caching logic works across all services (PHP, Composer, Nginx)
- Automatic version detection with API + fallback system
- Consistent user experience for cache management

### GD Extension Support for Legacy PHP

**Modern PHP (8.1+):** GD extension works out of the box ✅

**Legacy PHP (7.4, 8.0):** Interactive GD extension support with user education ⚠️

When installing legacy PHP versions, Chauffeur provides an interactive prompt:

```bash
⚠ Warning: PHP 7.4 requires additional compilation for GD support
[ install ] GD extension enables image processing (uploads, thumbnails, watermarks)
[ install ] This adds 2-3 minutes to installation time

Would you like to enable GD image processing support?
  1) Enable GD (recommended for image processing)
  2) Skip GD (faster installation)
Enter your choice (1-2, default=2):
```

**Implementation Status:**
- ✅ Interactive user prompting with time cost education
- ✅ Temporary directory preservation for bundled extension builds
- ✅ Permanent patching system integrated into installer
- ✅ GD compatibility header with wrapper functions
- ✅ Graceful failure handling (PHP installation continues without GD)
- 🚧 GD extension compilation in progress for PHP 7.4/8.0

**Technical Details:**
- Build infrastructure successfully initiates GD compilation
- Dramatically reduced patch warnings (13+ → 4)
- If GD compilation fails, installation continues gracefully
- Use `CHAUFFEUR_KEEP_BUILD_DIR=1` to preserve build directories for debugging

### System Dependencies for PHP Builds
Chauffeur compiles PHP from source and expects the common image/zip libraries to be available on the host. Install these once before running `chauf install php …`.

| Distro | Command |
|--------|---------|
| Debian / Ubuntu | `sudo apt-get install build-essential pkg-config autoconf bison re2c libzip-dev libjpeg-dev libpng-dev libfreetype6-dev libxml2-dev libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev libreadline-dev libmagickwand-dev libgmp-dev` |
| Arch Linux | `sudo pacman -S base-devel pkgconf libzip libjpeg-turbo libpng freetype2 libxml2 curl bzip2 zlib libxslt readline imagemagick gmp` |

If your distribution splits libraries differently, install the equivalent dev packages providing `libzip`, `libjpeg`, `libpng`, `freetype`, `zlib`, `curl`, `libxml2`, `libxslt`, `readline`, `MagickWand` (ImageMagick), and `gmp`. Once they're in place, `chauf install php <version>` will produce runtimes with GD, ZIP, Exif, freetype, jpeg, readline, imagick, GMP, BCMath, and mysqlnd-backed database extensions for Laravel apps (so `php -a`/`artisan tinker` keep arrow keys and history while `php artisan migrate`, image manipulation, and mathematics-intensive code work without extra modules). Chauffeur automatically fetches the latest stable Imagick release from PECL for every build (override via `CHAUFFEUR_IMAGICK_VERSION` or `CHAUFFEUR_IMAGICK_TARBALL` when you need a specific tarball).

`chauf install php` now preflights `pkg-config` plus all of the required libraries (`libzip`, `libjpeg`, `libpng`, `freetype`, `libxml2`, `libcurl`, `zlib`, `libxslt`, `readline`, `MagickWand`, `gmp`) via `pkg-config --modversion …` before downloading or compiling so you get actionable guidance instead of waiting for `./configure` to fail.

All compiled runtimes include GNU Readline via `--with-readline`, which fixes cursor navigation inside PsySH/`php artisan tinker` and provides persistent line editing in `php -a`. Chauffeur also enables `mysqli`, `PDO_MySQL`, `mysqlnd`, the PECL `imagick` extension, and math extensions `gmp` and `bcmath` by default so database-heavy, image-processing, and mathematics-intensive apps work immediately after `chauf install php`.

Refer to `AGENTS.md` for the authoritative workspace layout, dnsmasq instructions, and logging spec.

## Development & Contribution
- **Preferred workflow**: Use `chauf self-update --dev` from the repo root to rebuild binaries; avoid `go build -o chauf` in-tree.
- **Debug builds**: Set `CHAUFFEUR_KEEP_BUILD_DIR=1` while running `chauf install php …` to preserve the extracted PHP sources under `/tmp` when you need to inspect or patch them manually.
- **Offline sources**: When mirrors are unavailable, point `CHAUFFEUR_PHP_TARBALL`, `CHAUFFEUR_PHP_SIGNATURE`, and `CHAUFFEUR_PHP_KEYRING` at local files to skip tarball/signature downloads during `chauf install php …`.
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
