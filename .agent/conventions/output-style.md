# Output Style

Defines visual and structural standards for all `chauf` CLI output.

---

## Core Principle

**No repeating prefix.** V1 stamped `[ command ]` on every line — noisy, redundant. V2 sets context once (in the header) and uses clean indented lines for all subsequent output.

Two modes:
- **Simplified** (default): outcome-focused, minimal noise
- **Verbose** (`--verbose` flag or `chauf config set logging.verbose true`): every phase named, full paths shown, timing included

Both modes use the same symbol set and color palette. Verbose adds section headers, sub-action lines, and timing.

---

## Color Palette

| Color | ANSI | Used For |
|-------|------|---------|
| Green | `\033[32m` | `✓` success, `●` running |
| Red | `\033[31m` | `✗` error, `●` error state |
| Yellow | `\033[33m` | `⚠` warning, `●` degraded |
| Gray | `\033[90m` | Secondary info, paths, timing, separators |
| Cyan | `\033[36m` | URLs, prompts |
| Bold | `\033[1m` | Section headers, command header |
| Reset | `\033[0m` | Always appended after colored text |

Colors auto-disable when stdout is not a TTY or when `NO_COLOR` / `--no-color` is set.

---

## Symbol Set

| Symbol | Color | Meaning |
|--------|-------|---------|
| `✓` | green | Success |
| `✗` | red | Failure |
| `⚠` | yellow | Warning |
| `●` | green / yellow / red | Service running / degraded / error |
| `○` | gray | Service stopped |
| `•` | plain | Bullet item |
| `·` | gray | Inline separator |
| `─` | gray | Table divider, separator line |

---

## Modes

### Simplified (default)

**Header**: Present-continuous action on the first line (`Linking my-app...`). Indented body. Final URL or key result on a blank-separated line at the end.

```
Linking my-app...

  ✓ my-app.test → ~/Projects/my-app
  ✓ PHP 8.3  ·  SSL enabled  ·  nginx reloaded

  http://my-app.test:8080
  https://my-app.test:8443
```

Rules:
- One blank line before the body, one before the footer
- Indent body with 2 spaces
- No per-line prefix
- Show paths only on failure, not on success (unless the path is the outcome)
- Combine related success items onto one line with ` · ` when they're all short
- Next-step hints in gray below the result, when genuinely useful

### Verbose (`--verbose`)

**Header**: `chauf <command> ───────────────────` (bold, fills terminal width). Blank line. Sections labeled with bold nouns. Every sub-action shown.

```
chauf link ──────────────────────────────────────────

  Resolving
    Path    /home/user/Projects/my-app
    Slug    my-app
    Domain  my-app.test
    Type    laravel  (auto-detected)
    PHP     8.3  (global default)

  Generating nginx config
    ✓ Written  ~/.chauffeur/nginx/sites/my-app.conf

  Issuing SSL certificate
    ✓ mkcert my-app.test
    ✓ ~/.chauffeur/ssl/my-app.test.crt  (expires 2035-01-01)

  Reloading nginx
    ✓ Config valid  (nginx -t)
    ✓ Reloaded  (SIGHUP — 0 connection drops)

  ────────────────────────────────────────────────────
  ✓ my-app linked

    http://my-app.test:8080
    https://my-app.test:8443

    Config  ~/.chauffeur/projects/my-app/config.yaml
    Nginx   ~/.chauffeur/nginx/sites/my-app.conf
```

Rules:
- Horizontal rule header: `chauf <command> ` + `─` filling to 52 chars
- Section labels: bold, 2-space indent, no colon
- Sub-actions: 4-space indent, `key  value` alignment
- Timing shown for operations > 1s: `(6m 12s)`
- Horizontal `────` separator before the final result block
- Full paths always shown

---

## Command Output Patterns

### Action Commands

Commands performing multi-step operations: `link`, `install`, `start`, `stop`, `secure`, `unlink`.

**Simplified**:
```
Installing PHP 8.3...

  Downloading    php-8.3.14.tar.gz  (24.1 MB, cached)
  Configuring    --with-openssl --enable-gd --with-imagick ...
  Compiling      ████████████████████  100%  (6m 12s)
  Extensions     imagick  ·  openssl

  ✓ PHP 8.3 installed
```

Label-value format for phase lines: `  <label padded to 13>  <value>`. Inline progress bar for long compilations.

**Verbose**:
```
chauf install php 8.3 ───────────────────────────────

  Downloading
    https://www.php.net/distributions/php-8.3.14.tar.gz
    ✓ php-8.3.14.tar.gz  (24.1 MB)  — cached

  Validating dependencies
    ✓ libzip      1.10.1    ✓ libjpeg    9f
    ✓ libpng      1.6.43   ✓ freetype2  2.13.2
    ✓ libxml2     2.12.3   ✓ libcurl    8.5.0
    ✓ zlib        1.3.1    ✓ readline   8.2
    ✓ MagickWand  7.1.1    ✓ libgmp     6.3.0
    ✓ libsodium   1.0.19

  Configuring
    --prefix=~/.chauffeur/php/8.3  --with-openssl
    --enable-gd  --with-jpeg  --with-freetype
    --enable-fpm  --with-zip  --with-curl
    ✓ ./configure  (12s)

  Compiling  (8 cores)
    ████████████████████  100%  (6m 12s)
    ✓ make complete

  Installing
    ✓ ~/.chauffeur/php/8.3/bin/php
    ✓ ~/.chauffeur/php/8.3/bin/php-fpm
    ✓ ~/.chauffeur/php/8.3/etc/php.ini

  Extensions
    ✓ imagick  (PECL, compiled from source)
    ✓ openssl.ini → /etc/ssl/certs/ca-certificates.crt

  ────────────────────────────────────────────────────
  ✓ PHP 8.3.14 installed  (6m 47s)

    Binary  ~/.chauffeur/php/8.3/bin/php
    Shim    ~/.chauffeur/bin/shims/php  →  php 8.3
    Next    chauf start php 8.3
```

### List Commands

Commands displaying tabular data: `links`, `php list`, `autostart list`.

**Simplified** (same in both modes — table is already the right density):
```
  Project       Domain               PHP    SSL  FPM
  ────────────  ───────────────────  ─────  ───  ─────────
  my-app        my-app.test          8.3    ✓    shared
  admin-panel   admin-panel.test     8.3    ✓    shared
  legacy-site   legacy-site.test     7.4    ✗    dedicated
```

**Verbose** (verbose expands each project into a detail block):
```
chauf links ─────────────────────────────────────────

  3 projects registered

  my-app
    Domain   my-app.test
    Path     ~/Projects/my-app
    PHP      8.3  (shared FPM)
    SSL      ✓  ~/.chauffeur/ssl/my-app.test.crt
    Status   ● running

  admin-panel
    Domain   admin-panel.test
    Aliases  admin.my-app.test  ·  api.my-app.test
    Path     ~/Projects/admin-panel
    PHP      8.3  (shared FPM)
    SSL      ✓  ~/.chauffeur/ssl/admin-panel.test.crt
    Status   ● running

  legacy-site
    Domain   legacy-site.test
    Path     ~/Projects/legacy-site
    PHP      7.4  (dedicated FPM)
    SSL      ✗  not configured
    Status   ● running
```

Empty state (both modes):
```
  No projects linked.

  To link your first project:
    cd ~/Projects/my-app
    chauf link
```

### Status Commands

Commands displaying service health: `status`.

**Simplified** (same in both modes at default):
```
  Service            Status     PID      Uptime    Memory
  ─────────────────  ─────────  ───────  ────────  ──────
  nginx              ● running   12348    2h 34m    45 MB
  php-fpm 8.3        ● running   12345    2h 34m   128 MB
  php-fpm 8.1        ○ stopped
  fpm: isolated-app  ● running   12347    1h 02m    64 MB
```

**`--detail`** (available in both modes, always verbose format):
```
  nginx
    Status   ● running
    PID      12348
    Uptime   2h 34m
    Memory   45 MB
    Workers  1 master + 4 workers
    Config   ~/.chauffeur/nginx/etc/nginx.conf
    Logs     ~/.chauffeur/nginx/logs/error.log
```

### Health Commands

`chauf doctor`.

**Simplified**:
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

  DNS                ✓ dnsmasq  chauffeur.conf  test.test → 127.0.0.1

  ────────────────────────────────────────────────────────
  2 errors · 2 warnings · 18 passed

  re2c not found:   sudo pacman -S re2c
  bison not found:  sudo pacman -S bison

  Run chauf doctor --fix to address all issues interactively.
```

Simplified packs passing checks inline per category. Only failures and warnings get their own line.

**Verbose**:
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
    ✓ test.test  → 127.0.0.1

  ────────────────────────────────────────────────────
  2 errors · 2 warnings · 18 passed

  Errors (must fix before PHP install):
    re2c:   sudo pacman -S re2c
    bison:  sudo pacman -S bison

  Warnings (optional):
    libxml2:   affects PHP 7.4/8.0 only
    iptables:  run chauf doctor --fix for commands

  Run chauf doctor --fix to review and apply fixes.
```

### Interactive Commands

Commands with confirmation prompts: `unlink`, `uninstall`, `clean`.

**Both modes** (interactive is always visible):
```
Unlinking my-app...

  Removes:
    • ~/.chauffeur/nginx/sites/my-app.conf
    • ~/.chauffeur/projects/my-app/

  Confirm project name: _

  ✓ my-app unlinked
```

### Error Output

Errors always go to stderr. Format is consistent in both modes.

```
  ✗ Port 8080 already in use  (apache2, pid 9999)

    sudo systemctl stop apache2
    — or —
    chauf config set nginx.http_port 8081
```

For errors with multiple options, use labeled lines:

```
  ✗ nginx failed to start

    Error    Port 8080 already in use
    Process  apache2  (pid 9999)

    Options
      Stop apache2    sudo systemctl stop apache2
      Change port     chauf config set nginx.http_port 8081
      Kill process    kill 9999
```

---

## Next-Step Hints

Show one or two hints after success when the next action is non-obvious. Gray text, 2-space indent.

```
  ✓ PHP 8.3 installed

    chauf start php 8.3
    chauf link --php 8.3
```

No "Next steps:" label needed. The commands are self-evident.

---

## Global Flags Affecting Output

| Flag | Effect |
|------|--------|
| `--verbose` | Verbose mode output |
| `--quiet` / `-q` | Show only final result and errors |
| `--no-color` | Disable ANSI codes |

---

## Non-TTY Behavior

When stdout is piped or redirected:
- Colors disabled automatically
- Progress bars replaced with static percentage: `Compiling... 100%`
- Tables retain spacing (column alignment preserved)

Test with:
```bash
chauf links | cat
chauf status > /tmp/out.txt
```

---

## Dos and Don'ts

| Do | Don't |
|----|-------|
| Set context once in the header, then clean indented lines | Repeat a `[ command ]` prefix on every line |
| Show the final URL or key result in the footer | End with a generic "Done!" |
| Print paths only when they're the outcome or on failure | Show full paths for every intermediate step in simplified mode |
| Use `·` to combine short related facts on one line | Use separate lines for every minor detail in simplified mode |
| Show errors with clear options for remediation | Print a raw error string and exit |
| Use `--verbose` for implementation details | Force verbose output on users who just want the result |
| Show empty-state message with a next step | Print nothing for an empty list |
| Write warnings to stderr | Mix warnings into stdout |
