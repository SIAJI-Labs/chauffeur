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

## Quick Start

```bash
# Install Chauffeur
curl -sSL https://chauffeur.siaji.com/install | bash

# Initialize workspace
chauf init

# Install services
chauf install nginx php 8.3 composer

# Link and serve your project
cd my-project
chauf link --secure

# Visit https://my-project.test
```

## Core Features

| Feature | Description |
|---------|-------------|
| **Workspace Isolation** | Per-project PHP-FPM pools and nginx configurations in `~/.chauffeur` |
| **Multi-Version PHP** | Run PHP 7.4-8.4 simultaneously with project-specific isolation |
| **Zero-Config Routing** | Automatic `.test` domain resolution with dnsmasq |
| **One-Click SSL** | Generate trusted certificates with `chauf secure` |
| **Smart Service Management** | Start/stop/restart services per-project or globally |
| **Health & Diagnostics** | Comprehensive system checks with `chauf doctor` |

## Documentation

📖 **Complete Command Reference**: [chauffeur.siaji.com/docs/reference/commands](https://chauffeur.siaji.com/docs/reference/commands)

**Essential Commands**:
- `chauf init` - Initialize workspace
- `chauf link` - Register current directory as project
- `chauf start/stop` - Manage services
- `chauf doctor` - System health check
- `chauf logs` - View service logs

## Status & Development

- **Current State**: Early adopter preview, works on major Linux distributions
- **Development**: Go 1.22+ with comprehensive test coverage
- **Philosophy**: Manual project registration, minimal host impact, idempotent operations
- **AI-Assisted**: Codebase developed with AI agents following `AGENTS.md` contracts

See the [documentation site](https://chauffeur.siaji.com/docs) for complete guides, troubleshooting, and API reference.

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
| `chauf link` | `--site`, `--secure`, `--php`, `--dedicated-fpm`, `--http-port`, `--https-port`, `--alias`, `--force` | Register project with multi-domain support (shared FPM by default). |
| `chauf links` | `--slug`, `--site` | Table of all registered projects with SSL indicators for aliases; supports detailed view. |
| `chauf unlink` | `--alias`, `--slug`, `--site`, `--project`, `--all`, `--force` | Remove registrations or specific aliases; defaults to current directory. |
| `chauf secure` | — | Add SSL certificate to current linked project. |
| `chauf unsecure` | — | Remove SSL certificate from current linked project. |
| `chauf doctor` | `--check-all`, `--check-deps`, `--check-php`, `--check-ssl`, `--check-network`, `--check-dns`, `--fix`, `--auto-fix`, `--verbose`, `--quiet` | Perform comprehensive health checks with auto-fix suggestions and execution. |
| `chauf install <service> [ver...]` | `--force`, `--local`, `--no-cache` | Install services with visual separators; supports multiple PHP versions. |
| `chauf php install <ver>` | `--force`, `--no-ext`, `--from` | Install PHP runtimes into the workspace. |
| `chauf php use <ver>` | — | Set global default PHP version. |
| `chauf logs [service]` | `--follow`, `--lines`, `--level`, `--context`, `--verbose` | View and follow logs from nginx, PHP-FPM, and other services. |
| `chauf clean [target]` | `--dry-run`, `--force`, `--older-than`, `--keep-versions` | Clean workspace files (logs, cache, temp, old versions, SSL certs, projects). |
| `chauf migrate <project> <workspace>` | `--backup`, `--no-backup`, `--dry-run`, `--force` | Migrate projects between workspaces with safety validation. |
| `chauf doctor` | `--check-*`, `--fix`, `--auto-fix`, `--verbose` | System health checks with fix plan review and automatic execution. |
| `chauf php isolate <ver>` | — | Pin the current linked project to a version. |
| `chauf remove <service> [ver]` | `--force` | Remove installed runtimes with cache management (php/nginx/composer). |
| `chauf install composer` | — | Download Composer PHAR + shim. |
| `chauf self-update` | `--dev` | Pull latest release or rebuild from current repo. |
| `chauf uninstall` | `--purge` | Remove workspace (optionally runtimes/caches). |
| `chauf info` | — | Show workspace paths, installed services, port config, plus GitHub release/commit drift checks. |

## System Health & Dependencies

### 🔧 `chauf doctor` - Comprehensive Health Checking

Chauffeur includes a powerful health-checking system that validates all dependencies and provides automatic fixes:

```bash
# Quick health check
chauf doctor

# Show fix suggestions
chauf doctor --fix

# Review fix plan before execution (recommended)
chauf doctor --auto-fix

# Target specific areas
chauf doctor --check-ssl --check-php
chauf doctor --check-network --check-dns

# Detailed diagnostics
chauf doctor --verbose
```

#### What `chauf doctor` checks:

**System Dependencies:**
- `git`, `curl`, `tar` - Core tools for installation
- `gcc`, `make`, `pkg-config` - Build tools for PHP compilation
- Provides distro-specific install commands

**PHP Build Dependencies:**
- `libzip`, `libjpeg`, `libpng`, `freetype` - Image and ZIP support
- `libxml2`, `libcurl`, `zlib` - XML and HTTP support
- `readline`, `libxslt`, `MagickWand`, `gmp` - Extended functionality
- Validates via `pkg-config --modversion` before PHP builds

**SSL Certificate Dependencies:**
- `openssl` - SSL/TLS toolkit availability
- `mkcert` - Local trusted certificates (optional but recommended)
- Shows appropriate package commands for each distribution

**Network & Port Dependencies:**
- `iptables` - Port forwarding for privileged ports (80→8080, 443→8443)
- Port availability - Default Chauffeur ports (8080, 8443, 9000)
- Conflict detection with running services

**DNS Resolution Dependencies:**
- `dnsmasq` - Local DNS server for `.test` domains
- `.test` domain resolution - Validates DNS configuration

#### Auto-Fix Workflow:

```bash
chauf doctor --auto-fix
```

The auto-fix process:

1. **Collects all issues** across dependency categories
2. **Shows fix plan** with detailed breakdown:
   ```
   🔧 Fix Plan
   Found 1 fixable issue(s): 0 errors, 1 warnings

   SSL Certificate Dependencies:
     ⚠️ mkcert: Local trusted certificate authority
       Command: sudo dnf install -y mkcert
   ```
3. **User confirmation** - No surprise package installations
4. **Safe execution** - Only runs fixes after approval
5. **Progress feedback** - Shows execution results and verification

#### Distribution-Specific Fixes:

The doctor provides package commands tailored to your distribution:
- **Fedora**: `sudo dnf install -y mkcert`
- **Ubuntu/Debian**: `sudo apt update && sudo apt install -y mkcert`
- **Arch**: `sudo pacman -S mkcert`

## Getting Started
1. **Install prerequisites**: Run `chauf doctor` to check what's needed, or install Go 1.22+, git, curl, build tools (gcc/make/pkg-config), openssl headers.

2. **Install Chauffeur** (installs `chauf` under `~/.chauffeur/bin`):

   **Option A: Quick Install (Recommended)**
   ```bash
   curl -sSL https://chauffeur.siaji.com/install | bash
   ```

   **Option B: Manual Clone**
   ```bash
   git clone https://github.com/SIAJI-Labs/chauffeur.git
   cd chauffeur
   ./install.sh
   ```

   *Both methods automatically detect if you're in a Chauffeur git repo and build locally, or clone and build from source if needed.*

3. **Initialize workspace**:
   ```bash
   chauf init                    # Creates ~/.chauffeur structure and config
   ```

4. **Install services** (with intelligent caching):
   ```bash
   chauf install php 8.3        # First download - auto-cached for future
   chauf install nginx           # Instant if cached, downloads if not
   chauf install composer        # Reuses cached PHAR when available

   # Install multiple PHP versions in one command:
   chauf install php 8.3 php 7.4 composer
   ```
5. **Link projects** (shared FPM by default, dedicated when needed):
   ```bash
   cd ~/simple-project          # Shared FPM (resource efficient)
   chauf link

   cd ~/production-app          # Dedicated FPM (isolated)
   chauf link --dedicated-fpm --php 8.1

   cd ~/legacy-project         # Dedicated FPM (custom config needs)
   chauf link --dedicated-fpm --php 7.4
   ```
6. **Start services & browse**:
   ```bash
   chauf start
   firefox http://project-name.test:8080
   ```

`chauf link` automatically detects when Chauffeur's nginx instance already owns the configured HTTP/HTTPS ports and simply restarts it so the new site configuration is loaded—no more prompts to pick new ports while services are running. Likewise, `chauf unlink` removes the generated nginx config, restarts nginx when it's active, and the built-in catch‑all server returns a 404 for any unlinked domain so the site disappears immediately.

## Multi-Domain Support

Chauffeur supports multiple domains pointing to the same project directory, perfect for white-label development or multi-tenant applications.

### Multi-Domain Project Setup

**Link with multiple domains:**
```bash
# Link project with primary domain and aliases
chauf link --site myapp.test --alias admin.myapp.test --alias api.myapp.test --secure

# Add aliases to existing projects
chauf link --alias www.myapp.test --secure
chauf link --alias staging.myapp.test    # HTTP only
```

**View all projects with domains:**
```bash
chauf links
```
Output:
```
[ links ] Linked Projects (2)
[ links ] SLUG    PATH                    DOMAIN           ALIAS                              SSL  PHP   CREATED
[ links ] ------  ----------------------  ---------------  ---------------------------------  ---  ----  -------------------
[ links ] myapp   /home/user/myapp        myapp.test       admin.myapp.test (*), api.myapp.test (*)  *    8.3   2025-11-19 15:00
[ links ] cms     /home/user/cms          cms.test         -                                 *    8.3   2025-11-19 14:30
```

**Key Features:**
- `(*)` indicates aliases with SSL enabled
- Automatic multi-domain SSL certificates with SAN (Subject Alternative Names)
- Per-alias SSL configuration (HTTP/HTTPS)
- Backward compatible with existing single-domain projects

### SSL Certificate Management

**Multi-Domain SSL Certificates:**
- Single certificate covers all SSL-enabled domains
- Trusted certificates via mkcert (no browser warnings)
- Automatic regeneration when adding/removing SSL aliases
- Support for both primary and alias domains

**SSL Usage Examples:**
```bash
# All domains work with HTTPS
curl -k https://myapp.test:8443
curl -k https://admin.myapp.test:8443
curl -k https://api.myapp.test:8443

# HTTP access for non-SSL aliases
curl http://staging.myapp.test:8080
```

### Alias Management

**Add aliases to existing projects:**
```bash
# Add SSL-enabled alias
chauf link --alias new.domain.test --secure

# Add multiple aliases
chauf link --alias www.domain.test --alias api.domain.test --secure
```

**Remove specific aliases:**
```bash
chauf unlink --alias unwanted.domain.test
```

**Remove entire project (all domains):**
```bash
chauf unlink  # Shows confirmation with all domains listed
```

The enhanced unlink command displays all domains before confirmation:
```
[ unlink ] Project to unlink
[ unlink ]   Slug: myapp
[ unlink ]   Primary Domain: myapp.test (ssl=true)
[ unlink ]   Alias Domains:
[ unlink ]     - admin.myapp.test (HTTPS)
[ unlink ]     - api.myapp.test (HTTPS)
[ unlink ]     - staging.myapp.test (HTTP)
```

### Configuration Schema

Multi-domain configuration is stored in `~/.chauffeur/projects/{slug}/project.yaml`:

```yaml
version: 1
path: /home/user/myapp
php: "8.3"
site:
  domain: myapp.test
  ssl: true
domains:
  aliases:
    - domain: admin.myapp.test
      ssl: true
    - domain: api.myapp.test
      ssl: true
    - domain: staging.myapp.test
      ssl: false
runtime:
  php_fpm_socket: /home/user/.chauffeur/projects/myapp/runtime/php-fpm/php-fpm.sock
created_at: "2025-11-19T15:00:00Z"
```

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

# Install multiple PHP versions in one command
chauf install php 8.3 php 7.4  # Downloads and caches both versions

# Skip caching (useful for testing)
chauf install --no-cache nginx   # Download without caching

# Use local tarball (advanced usage)
chauf install php 8.3 --local    # Prompt for local tarball path

# Force reinstall (ignores cache for download)
chauf install composer --force   # Fresh download, updates cache
```

### Enhanced Installation Experience

When installing multiple services or PHP versions, Chauffeur provides **visual separators** to clearly distinguish between installation phases:

```bash
# Multiple services installation output example
chauf install nginx php 8.3 php 7.4 composer

[ install ] Installing nginx (source build from nginx.org release)...
[ nginx ] ✓ nginx built and installed (/home/siegg/.chauffeur/nginx)
[ install ] ✓ Installed nginx successfully

[ install ] ────────────────────────────────────────────────────────────
[ install ] Installing PHP 8.3...
[ php ] ✓ PHP 8.3 built and installed to /home/siegg/.chauffeur/php/8.3
[ install ] ✓ Installed PHP 8.3 successfully

[ install ] ────────────────────────────────────────────────────────────
[ install ] Installing PHP 7.4...
[ php ] ✓ PHP 7.4 built and installed to /home/siegg/.chauffeur/php/7.4
[ install ] ✓ Installed PHP 7.4 successfully

[ install ] ────────────────────────────────────────────────────────────
[ install ] Installing Composer (PHP dependency manager)...
[ composer ] ✓ Composer installed successfully (Uses Chauffeur PHP version isolation)
```

**Visual separators provide:**
- **Clear boundaries** between different service installations
- **Better progress tracking** during multi-service operations
- **Reduced confusion** when installing multiple PHP versions
- **Professional output** that's easy to scan and understand

### Example Project Feature

Chauffeur creates an example project during `chauf init` and automatically links it when you install services, helping you get started immediately:

```bash
# 1. Initialize workspace (creates example project)
chauf init
# Example project created at: ~/.chauffeur/projects/example-project

# 2. Install services (links example project)
chauf install nginx php

# Or install multiple PHP versions with all services:
chauf install nginx php 8.3 php 7.4 composer
# Example project linked successfully at: example-project.test

# 3. Start services
chauf start

# 4. Access example project
# Available at http://example-project.test (or your configured port)
# The example includes:
#   - Welcome page with Chauffeur information
#   - phpinfo() output to verify PHP setup
#   - Links to documentation

# 5. Remove example project when ready
chauf unlink --project example-project
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

### Enhanced Logging & Workspace Management

#### **Improved Logs Command (`chauf logs`)**

The logs command now supports intelligent service discovery and interactive selection:

```bash
# Direct version specification
chauf logs php-fpm 7.4          # Shows PHP 7.4 FPM logs directly
chauf logs nginx                 # Shows nginx logs

# Interactive service selection (when multiple versions exist)
chauf logs php-fpm               # Interactive menu: [1] php-fpm-7.4, [2] php-fpm-8.3

# Follow logs in real-time
chauf logs php-fpm --follow

# Filter by log level and limit lines
chauf logs nginx --level error --lines 50

# Show file context
chauf logs php-fpm --context
```

**Features:**
- **Version Specification**: `chauf logs php 7.4` automatically targets `php-7.4` services
- **Interactive Selection**: Shows menu when multiple services match (e.g., multiple PHP versions)
- **Service Status**: Shows running/stopped status with 🟢/🔴 indicators
- **Deduplication**: Removes duplicate service entries from global and project sources
- **Real-time Following**: `--follow` flag for tailing logs live

#### **Enhanced Clean Command (`chauf clean`)**

The clean command now provides detailed file information and accurate reporting:

```bash
# Interactive cleanup with file sizes
chauf clean logs                 # Shows: Delete log file: access.log (178 B)? [y/N]
chauf clean cache                # Shows: Delete file: composer.phar (3.1 MB)? [y/N]

# Dry-run mode to preview what would be cleaned
chauf clean --dry-run           # Shows potential deletions without actually deleting

# Target specific cleanup areas
chauf clean temp                 # Clean temporary files only
chauf clean ssl-certs           # Clean stale SSL certificates only
chauf clean old-versions        # Remove old PHP versions
chauf clean projects            # Remove unlinked project directories

# Force cleanup without prompts
chauf clean --force             # Skip all confirmation prompts

# Clean files older than specified time
chauf clean --older-than 7d      # Clean files older than 7 days
```

**Features:**
- **File Size Display**: Shows human-readable sizes in prompts (B, KB, MB)
- **Accurate Reporting**: Only counts actually deleted files in summaries
- **Clear Feedback**: "No files found to clean" messages for empty categories
- **Selective Cleanup**: Target specific types of files or services
- **Dry-run Mode**: Preview changes without executing
- **Age-based Filtering**: Clean files based on modification time

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
| Debian / Ubuntu | `sudo apt-get install build-essential pkg-config autoconf bison re2c libzip-dev libjpeg-dev libpng-dev libfreetype6-dev libxml2-dev libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev libreadline-dev libmagickwand-dev libgmp-dev libsodium-dev` |
| Arch Linux | `sudo pacman -S base-devel pkgconf libzip libjpeg-turbo libpng freetype2 libxml2 curl bzip2 zlib libxslt readline imagemagick gmp libsodium` |

If your distribution splits libraries differently, install the equivalent dev packages providing `libzip`, `libjpeg`, `libpng`, `freetype`, `zlib`, `curl`, `libxml2`, `libxslt`, `readline`, `MagickWand` (ImageMagick), `gmp`, and `libsodium`. Once they're in place, `chauf install php <version>` will produce runtimes with GD, ZIP, Exif, freetype, jpeg, readline, imagick, GMP, BCMath, sodium, and mysqlnd-backed database extensions for Laravel apps (so `php -a`/`artisan tinker` keep arrow keys and history while `php artisan migrate`, image manipulation, mathematics-intensive code, and modern cryptography work without extra modules). Chauffeur automatically fetches the latest stable Imagick release from PECL for every build (override via `CHAUFFEUR_IMAGICK_VERSION` or `CHAUFFEUR_IMAGICK_TARBALL` when you need a specific tarball).

**Enhanced Modern System Compatibility**: Chauffeur now supports the latest library versions including libxml2 2.15+, libcurl 8.0+, and ImageMagick 7.1+, ensuring compatibility with modern Linux distributions while maintaining support for legacy PHP versions.

`chauf install php` now preflights `pkg-config` plus all of the required libraries (`libzip`, `libjpeg`, `libpng`, `freetype`, `libxml2`, `libcurl`, `zlib`, `libxslt`, `readline`, `MagickWand`, `gmp`, `libsodium`) via `pkg-config --modversion …` before downloading or compiling so you get actionable guidance instead of waiting for `./configure` to fail.

All compiled runtimes include GNU Readline via `--with-readline`, which fixes cursor navigation inside PsySH/`php artisan tinker` and provides persistent line editing in `php -a`. Chauffeur also enables `mysqli`, `PDO_MySQL`, `mysqlnd`, the PECL `imagick` extension, and math extensions `gmp` and `bcmath` by default so database-heavy, image-processing, and mathematics-intensive apps work immediately after `chauf install php`.

## OpenSSL Certificate Configuration

**Automatic OpenSSL Configuration**: Each PHP installation now includes automatic OpenSSL configuration with distribution-aware certificate authority paths, ensuring secure connections work immediately.

**Features**:
- **Distribution Detection**: Automatically detects Linux distribution (Fedora, Ubuntu, Arch, openSUSE, etc.)
- **CA Path Management**: Sets appropriate certificate bundle paths for each distribution
- **SSL Verification**: Enables PHP streams (HTTPS APIs, SMTP over SSL, Composer secure downloads)
- **Per-Version Configuration**: Each PHP version gets its own `openssl.ini` in `etc/conf.d/`

**Certificate Paths by Distribution**:
- **RHEL/Family** (Fedora, CentOS, Rocky): `/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem`
- **Debian/Ubuntu**: `/etc/ssl/certs/ca-certificates.crt`
- **Arch Linux**: `/etc/ssl/certs/ca-certificates.crt`
- **openSUSE**: `/etc/ssl/ca-bundle.pem`

**Doctor Integration**: `chauf doctor` now validates OpenSSL configuration and can auto-generate missing configuration files with `chauf doctor --auto-fix`.

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
