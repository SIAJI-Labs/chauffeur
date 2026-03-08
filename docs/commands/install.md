# Install & Remove Commands

Commands for installing and removing managed services (nginx, PHP, Composer).

---

## `chauf install`

**Purpose**: Install a service into the workspace. PHP is compiled from source; nginx is compiled from source; Composer is downloaded as a PHAR. All downloads are cached for faster subsequent installs.

**Usage**:
```
chauf install <service> [version] [flags]
```

**Services**:

| Service | Version arg | Example |
|---------|------------|---------|
| `nginx` | none | `chauf install nginx` |
| `php` | required | `chauf install php 8.3` |
| `composer` | none | `chauf install composer` |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | | bool | false | Reinstall even if already installed |
| `--no-cache` | | bool | false | Download fresh, skip the download cache |
| `--verbose` | `-v` | bool | false | Stream build output to terminal (default: spinner) |

Flags may appear before or after positional arguments:

```bash
chauf install php 8.3 --verbose
chauf install --verbose php 8.3   # equivalent
```

**Examples**:

```bash
# Install a service
chauf install nginx
chauf install php 8.3
chauf install composer

# Force reinstall (re-downloads and recompiles)
chauf install php 8.3 --force

# Install without using cache
chauf install nginx --no-cache

# Stream full build output
chauf install php 8.3 --verbose

# Use a local tarball (offline/CI) via environment variable
CHAUFFEUR_PHP_TARBALL=/path/to/php-8.3.20.tar.gz chauf install php 8.3
```

**Output** (default quiet mode — spinner per step):
```

  ⠋ Resolving PHP 8.3 version
  ✓ Resolving PHP 8.3 version   8.3.20
  ✓ Downloaded  php-8.3.20.tar.gz
  ⠋ Building PHP 8.3.20  (this may take 5–15 minutes)
  ✓ Building PHP 8.3.20   ~/.chauffeur/php/8.3/bin/php
  ✓ Installing imagick extension   extension=imagick

  ✓ PHP 8.3.20 installed

  Set as default:  chauf php use 8.3
```

**Output** (`--verbose` — streams make output):
```

  → Resolving PHP 8.3 version
  ✓ Resolving PHP 8.3 version   8.3.20

  ✓ Downloaded  php-8.3.20.tar.gz

  → Building PHP 8.3.20
  [build output streams here...]
  ✓ Building PHP 8.3.20   ~/.chauffeur/php/8.3/bin/php

  → Installing imagick extension
  [build output streams here...]
  ✓ Installing imagick extension   extension=imagick

  ✓ PHP 8.3.20 installed
```

**Caching behavior**: Successful downloads are stored in `~/.chauffeur/cache/`. On subsequent installs, the cached file is used and verified by checksum. Use `--no-cache` to skip or `--force` to refresh.

**Environment variable overrides** (for offline/CI use):

| Variable | Purpose |
|----------|---------|
| `CHAUFFEUR_PHP_TARBALL` | Path to local PHP source tarball (skips download) |

---

## `chauf remove`

**Purpose**: Remove an installed service from the workspace. Stops any running processes for the service before removal.

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

**Examples**:

```bash
# Remove a PHP version (prompts for confirmation)
chauf remove php 7.4

# Remove nginx
chauf remove nginx

# Remove without prompting
chauf remove php 8.0 --force
```

**Output**:
```

  Removing  PHP 7.4
  Directory  ~/.chauffeur/php/7.4
  This will permanently delete ~/.chauffeur/php/7.4 and all its contents.

  Remove PHP 7.4? [y/N]: y

  ✓ php-fpm 7.4 stopped
  ✓ PHP 7.4 removed
```

**Notes**:
- nginx removal only removes the `sbin/` binary and `modules/` directory. Config, logs, and certs are preserved.
- PHP removal permanently deletes the entire version directory including binaries, extensions, and FPM config.
- Composer removal removes only `composer.phar`. Project `vendor/` directories are never touched.
- The project source files are never touched.
