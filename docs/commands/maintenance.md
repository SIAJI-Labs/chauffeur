# Maintenance Commands

Commands for health checking, cleanup, migration, and updating.

> **Implementation status**:
> - `chauf doctor` — Phase 4 (not yet implemented)
> - `chauf clean` — Phase 4 (not yet implemented)
> - `chauf migrate` — Phase 4 (not yet implemented)
> - `chauf self-update` — implemented

---

## `chauf doctor`

**Purpose**: Validate the entire Chauffeur environment — system tools, PHP build dependencies, SSL setup, port availability, and DNS resolution. Reports every issue with specific remediation steps, and can apply fixes automatically.

Running `chauf doctor` before the first PHP install saves time by catching missing libraries early.

**Usage**:
```
chauf doctor [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check-all` | bool | false | Run all check categories (default when no `--check-*` is specified) |
| `--check-system` | bool | false | Check required system tools (git, curl, tar, gcc, make, etc.) |
| `--check-php` | bool | false | Check PHP build dependencies (libzip, libxml2, imagick, etc.) |
| `--check-ssl` | bool | false | Check SSL tools (openssl, mkcert) and existing project certs |
| `--check-network` | bool | false | Check port availability (8080, 8443) and iptables |
| `--check-dns` | bool | false | Check dnsmasq configuration, .test domain resolution, and offline resilience |
| `--fix` | bool | false | Show interactive fix plan; prompt before applying each fix |
| `--auto-fix` | bool | false | Show fix plan and apply all safe fixes after a single confirmation |
| `--verbose` | bool | false | Show passing checks in full detail (not just failures) |
| `--quiet`, `-q` | bool | false | Print only summary; suppress individual check output |

**Examples**:

```bash
# Full health check (default — runs all categories)
chauf doctor

# Check only PHP build dependencies
chauf doctor --check-php

# Check SSL and network only
chauf doctor --check-ssl --check-network

# Show and apply fixes automatically
chauf doctor --auto-fix

# Review fixes interactively before applying
chauf doctor --fix

# Detailed output for all checks (including passing ones)
chauf doctor --verbose
```

**Output** (simplified, default):
```
Doctor

  System Tools       ✓ git  curl  tar  gcc  make  pkg-config
                     ✗ re2c  bison

  PHP Dependencies   ✓ libzip  libjpeg  libpng  freetype2
                     ⚠ libxml2 2.12.3  (incompatible with PHP 7.4/8.0)
                     ✓ libcurl  zlib  readline  MagickWand  libgmp  libsodium

  SSL                ✓ openssl  mkcert  CA installed
                     ✓ my-app.test.crt  admin-panel.test.crt

  Network            ✓ Port 8080  Port 8443
                     ⚠ iptables not configured

  DNS                ✓ dnsmasq  chauffeur.conf
                     ✓ test.test → 127.0.0.1  (dnsmasq)  (system resolver)
                     ⚠ offline resilience not configured

  ────────────────────────────────────────────────────────
  2 errors · 3 warnings · 18 passed

  re2c not found:   sudo pacman -S re2c
  bison not found:  sudo pacman -S bison

  Run chauf doctor --fix to address all issues interactively.
```

**Output** (`--verbose`, shows every check individually):
```
chauf doctor ────────────────────────────────────────

  System Tools
    ✓ git          2.43.0
    ✓ curl         8.5.0
    ✓ tar          1.35
    ✓ gcc          13.2.1
    ✓ make         4.4.1
    ✓ pkg-config   2.1.1
    ✗ re2c         not found
    ✗ bison        not found

  PHP Build Dependencies
    ✓ libzip       1.10.1
    ✓ libjpeg      9f
    ✓ libpng       1.6.43
    ✓ freetype2    2.13.2
    ⚠ libxml2      2.12.3
        PHP 7.4 and 8.0 require ≤ 2.11. Modern PHP unaffected.
    ✓ libcurl      8.5.0
    ✓ zlib         1.3.1
    ✓ readline     8.2
    ✓ MagickWand   7.1.1
    ✓ libgmp       6.3.0
    ✓ libsodium    1.0.19

  SSL
    ✓ openssl      3.1.4
    ✓ mkcert       1.4.4
    ✓ CA installed
    ✓ my-app.test.crt       (valid, expires 2035-01-01)
    ✓ admin-panel.test.crt  (valid, expires 2035-01-01)

  Network
    ✓ Port 8080  available
    ✓ Port 8443  available
    ⚠ iptables not configured
        Port forwarding 80→8080 and 443→8443 not active.
        Domains accessible on :8080/:8443 only.

  DNS
    ✓ dnsmasq    installed and running
    ✓ /etc/dnsmasq.d/chauffeur.conf
    ✓ test.test  → 127.0.0.1  (via dnsmasq directly)
    ✓ test.test  → 127.0.0.1  (via system resolver)
    ⚠ offline resilience not configured
        NetworkManager may drop 127.0.0.1 from /etc/resolv.conf when offline.
        .test domains may stop resolving without internet.

  ────────────────────────────────────────────────────
  2 errors · 3 warnings · 18 passed

  Errors (must fix before PHP install):
    re2c:   sudo pacman -S re2c
    bison:  sudo pacman -S bison

  Warnings (optional):
    libxml2:           affects PHP 7.4/8.0 only
    iptables:          run chauf doctor --fix for commands
    offline DNS:       run chauf doctor --fix for options

  Run chauf doctor --fix to review and apply fixes.
```

**Fix plan** (`--fix` / `--auto-fix`):
```
Fix Plan

  Errors (2):
    re2c:   sudo pacman -S re2c
    bison:  sudo pacman -S bison

  Warnings (2 fixable):

    iptables 80→8080 redirect:
      sudo iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-port 8080
      sudo iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-port 8443

    DNS offline resilience (detected: Arch Linux + NetworkManager):

      Option A — NetworkManager dnsmasq plugin (recommended):
        sudo tee /etc/NetworkManager/conf.d/chauffeur-dns.conf << 'EOF'
        [main]
        dns=dnsmasq
        EOF
        sudo systemctl restart NetworkManager

      Option B — systemd-resolved split DNS:
        sudo mkdir -p /etc/systemd/resolved.conf.d/
        sudo tee /etc/systemd/resolved.conf.d/chauffeur.conf << 'EOF'
        [Resolve]
        DNS=127.0.0.1
        Domains=~test
        EOF
        sudo systemctl restart systemd-resolved

      Option C — /etc/hosts entries (no dnsmasq required):
        sudo tee -a /etc/hosts << 'EOF'
        127.0.0.1  my-app.test
        127.0.0.1  admin-panel.test
        EOF

  Apply all fixes? [y/N]:
```

**Distribution-specific commands**: Doctor detects the running distribution and tailors package install commands:

| Distribution | Package manager |
|-------------|----------------|
| Arch | `sudo pacman -S` |
| Debian/Ubuntu | `sudo apt install` |
| Fedora/RHEL | `sudo dnf install` |
| openSUSE | `sudo zypper install` |

---

## `chauf clean`

**Purpose**: Remove workspace artifacts — cached downloads, log files, stale SSL certificates, orphaned project directories, and temporary build files. Interactive by default, with dry-run and age-filtering support.

**Usage**:
```
chauf clean [target] [flags]
```

**Targets** (if omitted, all targets are shown interactively):

| Target | Description |
|--------|-------------|
| `cache` | Download cache (`~/.chauffeur/cache/`) |
| `logs` | Log files in `~/.chauffeur/logs/` and service log files |
| `temp` | Temporary build directories (from failed PHP installs, etc.) |
| `ssl-certs` | SSL certificates for projects that are no longer linked |
| `old-versions` | Old PHP versions not used by any project |
| `projects` | Orphaned project directories (project was unlinked but directory remains) |
| `all` | All of the above |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Show what would be removed without removing anything |
| `--force` | bool | false | Skip per-file confirmation prompts |
| `--older-than <age>` | string | — | Only remove files older than this age (e.g. `7d`, `30d`, `1h`) |
| `--keep-versions <n>` | int | — | Keep the N most recent PHP versions when cleaning old-versions |

**Examples**:

```bash
# Interactive cleanup (asks per target category)
chauf clean

# Clean only the download cache
chauf clean cache

# Preview what would be removed
chauf clean all --dry-run

# Clean logs older than 30 days without confirmation
chauf clean logs --older-than 30d --force

# Remove all cache without prompts
chauf clean cache --force

# Clean old PHP versions, keeping 3 most recent
chauf clean old-versions --keep-versions 3
```

**Output** (interactive):
```
Cleaning cache...

  Delete php-7.4.33.tar.gz  (21.3 MB)?  [y/N]: y
  ✓ Deleted  php-7.4.33.tar.gz  (21.3 MB)

  Delete php-8.0.30.tar.gz  (23.1 MB)?  [y/N]: n
  − Skipped  php-8.0.30.tar.gz

  Delete nginx-1.24.0.tar.gz  (1.0 MB)?  [y/N]: y
  ✓ Deleted  nginx-1.24.0.tar.gz  (1.0 MB)

  2 deleted (22.3 MB freed)  ·  1 skipped
```

**Dry-run output**:
```
  Dry run — nothing will be deleted

  cache/
    php-7.4.33.tar.gz       21.3 MB
    nginx-1.24.0.tar.gz      1.0 MB

  logs/
    access.log              45.2 MB
    error.log                2.1 MB

  Total: 69.6 MB would be freed
```

---

## `chauf migrate`

**Purpose**: Move a project's Chauffeur configuration from one workspace to another. Copies the project config and nginx site config, updates paths as needed, and removes the project from the source workspace.

Useful when managing multiple workspaces (e.g. work vs personal) or moving to a new machine configuration.

**Usage**:
```
chauf migrate <project> <target-workspace> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `<project>` | yes | Project slug or path to migrate |
| `<target-workspace>` | yes | Absolute path to the target Chauffeur workspace |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--update-path <path>` | string | — | Update the project's source path in the new workspace (if the project files moved) |
| `--backup` | bool | true | Create a backup of the project config before migrating |
| `--no-backup` | bool | false | Skip backup |
| `--dry-run` | bool | false | Show what would be migrated without making changes |
| `--force` | bool | false | Overwrite if project already exists in target workspace |

**Examples**:

```bash
# Move a project to a second workspace
chauf migrate my-app ~/.chauffeur-work

# Move with a different source path (project files also moved)
chauf migrate my-app ~/.chauffeur-work --update-path /new/location/my-app

# Preview the migration
chauf migrate my-app ~/.chauffeur-work --dry-run
```

**Output**:
```
Migrating my-app → ~/.chauffeur-work...

  Source   ~/.chauffeur/projects/my-app/
  Target   ~/.chauffeur-work/projects/my-app/

  ✓ Config backed up  (~/.chauffeur/projects/my-app/config.yaml.bak)
  ✓ Project config copied
  ✓ Nginx config copied
  ✓ Removed from source workspace
  ✓ Source nginx reloaded
  ✓ Target nginx reloaded

  ✓ my-app migrated

  Now managed by ~/.chauffeur-work
  http://my-app.test:8080  (target workspace ports)
```

---

## `chauf self-update`

**Purpose**: Update the `chauf` binary to the latest release. Fetches the latest release from GitHub, verifies the checksum, and replaces the current binary.

When run from inside the Chauffeur source repository, `--dev` rebuilds the binary from the local source instead of downloading.

**Usage**:
```
chauf self-update [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dev` | bool | false | Rebuild binary from the local source repository (for contributors) |
| `--dry-run` | bool | false | Show what version would be installed without installing it |
| `--version <ver>` | string | latest | Install a specific version instead of latest |

**Examples**:

```bash
# Update to latest release
chauf self-update

# Preview without updating
chauf self-update --dry-run

# Rebuild from current repo source (contributor workflow)
chauf self-update --dev

# Install a specific version
chauf self-update --version 2.1.0
```

**Output**:
```
Updating chauf...

  Current   2.0.0
  Latest    2.1.0

  Downloading    chauf 2.1.0  (6.2 MB)
  ✓ Checksum verified
  ✓ Binary replaced

  ✓ Updated to 2.1.0
```

**Already up to date**:
```
  Already on latest version (2.1.0)
```

**`--dev` mode** (inside repo):
```
Building from source...

  Repository  /home/user/Projects/chauffeur-v2
  Commit      abc1234  (main, 2 commits ahead of origin)

  ✓ go build succeeded
  ✓ Binary replaced

  ✓ Binary updated from source  (2.0.0-dev+abc1234)
```

**Notes**:
- The binary is replaced atomically (write to temp file, then rename) to avoid partial writes.
- The running binary cannot update itself on Linux — the replacement takes effect on the next invocation.
- `--dev` requires being in the repository root with Go 1.22+ available.
