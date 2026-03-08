# Install & Remove Commands

Commands for installing and removing managed services (nginx, PHP, Composer).

---

## `chauf install`

**Purpose**: Install one or more services into the workspace. PHP is compiled from source; nginx is compiled from source; Composer is downloaded as a PHAR. All downloads are cached for faster subsequent installs.

**Usage**:
```
chauf install <service> [version] [<service> [version] ...] [flags]
```

**Services**:

| Service | Version arg | Example |
|---------|------------|---------|
| `nginx` | none | `chauf install nginx` |
| `php` | required | `chauf install php 8.3` |
| `composer` | none | `chauf install composer` |
| `all` | — | `chauf install all` (nginx + default PHP + composer) |

Multiple services and versions can be installed in a single command:

```bash
chauf install nginx php 8.3 php 8.1 composer
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Reinstall even if already installed; refreshes cache |
| `--no-cache` | bool | false | Download fresh, skip the download cache |
| `--local` | bool | false | Prompt for a local tarball path instead of downloading |

**Examples**:

```bash
# Install a single service
chauf install nginx
chauf install php 8.3
chauf install composer

# Install everything in one command
chauf install nginx php 8.3 composer

# Install multiple PHP versions at once
chauf install php 8.3 php 8.1 php 7.4

# Install the full stack with multiple PHP versions
chauf install nginx php 8.4 php 8.3 php 8.1 composer

# Force reinstall (re-downloads and recompiles)
chauf install php 8.3 --force

# Install without adding to cache
chauf install nginx --no-cache

# Install from a local tarball (offline/air-gapped)
chauf install php 8.3 --local
# → Prompts: Path to php-8.3.tar.gz:
```

**Output** (multi-service install):
```
Installing nginx...

  Downloading    nginx-1.27.4.tar.gz  (1.1 MB, cached)
  Configuring    --prefix=~/.chauffeur/nginx ...
  Compiling      ████████████████████  100%  (23s)

  ✓ nginx installed

────────────────────────────────────────────────────

Installing PHP 8.3...

  Downloading    php-8.3.14.tar.gz  (24.1 MB) — caching for next time
  Dependencies   ✓ libzip  libjpeg  libpng  freetype2  libxml2  libcurl
                 ✓ zlib  readline  MagickWand  libgmp  libsodium
  Configuring    --with-openssl --enable-gd --with-imagick ...
  Compiling      ████████████████████  100%  (6m 12s)
  Extensions     imagick  ·  openssl

  ✓ PHP 8.3.14 installed

────────────────────────────────────────────────────

Installing Composer...

  Downloading    composer.phar  (3.1 MB, cached)
  ✓ Checksum verified
  ✓ Shim written  ~/.chauffeur/bin/shims/composer

  ✓ Composer installed

────────────────────────────────────────────────────

  3 services installed.

  chauf link
  chauf start
  chauf doctor
```

**Caching behavior**: Successful downloads are stored in `~/.chauffeur/cache/`. On subsequent installs, the cached file is used and verified by checksum. Use `--no-cache` to skip or `--force` to refresh.

**Dependency pre-flight** (PHP only): Before starting compilation, all required libraries are checked via `pkg-config`. If any are missing, the install aborts with specific installation commands for the detected distribution:

```
  ✗ Missing PHP build dependencies

    Not found:
      • libmagickwand (MagickWand / ImageMagick)
      • libzip

    Arch Linux:     sudo pacman -S imagemagick libzip
    Debian/Ubuntu:  sudo apt install libmagickwand-dev libzip-dev
    Fedora:         sudo dnf install ImageMagick-devel libzip-devel

    Run chauf doctor --check-php for full dependency report.
```

**Legacy PHP (7.4, 8.0)**: When installing PHP 7.4 or 8.0, an interactive prompt asks about GD extension support (requires extra patching and adds 2–3 minutes):

```
  ⚠ PHP 7.4 GD requires additional patching
    Adds ~2 minutes. GD enables image uploads, thumbnails, watermarks.

    Enable GD for PHP 7.4?
      1) Yes, enable GD (recommended)
      2) No, skip GD
    Choice [1/2, default 2]:
```

**Environment variable overrides** (for offline/CI use):

| Variable | Purpose |
|----------|---------|
| `CHAUFFEUR_PHP_TARBALL` | Path to local PHP source tarball |
| `CHAUFFEUR_PHP_SIGNATURE` | Path to local PHP `.asc` signature file |
| `CHAUFFEUR_PHP_KEYRING` | Path to local PHP keyring |
| `CHAUFFEUR_IMAGICK_TARBALL` | Path to local imagick tarball |
| `CHAUFFEUR_KEEP_BUILD_DIR` | Set to `1` to keep PHP build directory after install (for debugging) |

---

## `chauf remove`

**Purpose**: Remove an installed service from the workspace. Stops any running processes for the service, removes binaries and configs, and optionally removes cached downloads.

**Usage**:
```
chauf remove <service> [version] [flags]
```

**Services**:

| Service | Version arg | Example |
|---------|------------|---------|
| `nginx` | none | `chauf remove nginx` |
| `php` | required | `chauf remove php 8.1` |
| `composer` | none | `chauf remove composer` |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Skip confirmation prompt |
| `--keep-cache` | bool | false | Keep cached download files (skip the prompt) |
| `--purge-cache` | bool | false | Remove cached download files without prompting |

**Examples**:

```bash
# Remove a PHP version (interactive: confirm + ask about cache)
chauf remove php 7.4

# Remove nginx
chauf remove nginx

# Remove without prompts, keep cache
chauf remove php 8.0 --force --keep-cache
```

**Output**:
```
Removing PHP 7.4...

  Projects using PHP 7.4:
    • legacy-site (legacy-site.test)

  These projects will lose their PHP version — relink them after removal.

  Continue? [y/N]: y

  ✓ PHP-FPM 7.4 stopped
  ✓ ~/.chauffeur/php/7.4/ removed

  Cached: php-7.4.33.tar.gz (21.3 MB)
  Keep cached file? [Y/n]: n
  ✓ Cache removed

  ✓ PHP 7.4 removed

  Relink affected projects:
    cd ~/Projects/legacy-site
    chauf link --php 8.1
```

**Notes**:
- Running services are stopped before removal.
- nginx removal also removes all generated site configs and SSL certificates.
- Composer removal removes the PHAR and shim but not project `vendor/` directories.
- The project source files are never touched.
