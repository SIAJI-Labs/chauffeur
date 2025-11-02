# step-2 — Service Installation Spec (Arch-first)

> **Audience**: Chauffeur developers. We are building the service installers that place binaries under `~/.chauffeur/` (user space), with shims under `~/.chauffeur/bin/`. No system prefix writes. This step focuses on **Arch Linux**; other distros will follow.

---

## Context
Chauffeur manages its own copies of **Caddy**, **Nginx**, and **PHP** under the workspace prefix. Installers must:
- Work without root, write only to `$HOME/.chauffeur/`.
- Verify downloads (checksum/signature).
- Be idempotent and transparent (dry-run supported at CLI level).
- Create/update **shims** (wrappers) in `~/.chauffeur/bin/`.

We will ship Go/Cobra commands that call installer packages.

---

## Goals (this step)
1. Implement distro/arch detection and guard rails (Arch-first).
2. Implement **Caddy** installer via vendor tarball (Arch).
3. Implement **Nginx** installer via source build into workspace (Arch).
4. Scaffold **PHP** installer interface (Arch) — full build in step-3.
5. Generate minimal default configs (Caddyfile, nginx.conf) under workspace.

---

## CLI Surface (user-visible)

- `chauf service install caddy [--from tarball] [--force]`
- `chauf service install nginx [--from source] [--with-mod <module>...] [--force]`
- `chauf php install <version> [--from source|tarball] [--force] [--no-ext]`

Common flags:
- `--dry-run` → print plan only.
- `--prefix <dir>` → install root (default `$HOME/.chauffeur`).
- `--channel <stable|canary>` (for Caddy future use).

---

## Distro & Arch Detection (contract)

Use `/etc/os-release` and `uname -m`.

- **Arch detection**: `ID=arch` or `ID_LIKE=arch`. If ambiguous, treat as "unknown".
- **Architecture**: `uname -m` → map to `{x86_64, aarch64}`.
- If distro is unknown: proceed with **generic** path (tarball/source builds) but print a warning.

**Required data exposed by the detector**:
```json
{
  "distro": "arch" | "unknown",
  "arch": "x86_64" | "aarch64" | "other",
  "pretty": "Arch Linux (x86_64)"
}
```

---

## Installation Prefix & Layout (restate)

```
~/.chauffeur/
  bin/               # shims + chauf
  bin/shims/
  caddy/
    bin/caddy
    Caddyfile
  nginx/
    bin/nginx
    etc/nginx.conf
    conf.d/
    sites-available/
    sites-enabled/
  php/
    8.3/bin/php
    ...
```

---

## Caddy (Arch) — Tarball Installer

**Source**: Official Caddy release tarballs (GitHub Releases).

**Steps**
1. Resolve asset for `linux-<arch>`.
2. Download `.tar.gz` and `.sha256` (or checksum file in manifest).
3. Verify SHA256 → **fail closed** on mismatch.
4. Extract `caddy` binary; place at `~/.chauffeur/caddy/bin/caddy` (0755).
5. Create/refresh shim `~/.chauffeur/bin/caddy` → points to workspace binary.
6. Write default `~/.chauffeur/Caddyfile` if missing:
   ```
   {
     auto_https off
   }
   # Project-specific vhosts written by `chauf link`
   ```

**Acceptance**
- Re-running keeps same version unless `--force` or new version requested.
- `caddy version` returns successfully when invoking the shim.

**Arch-specific notes**
- Do **not** pacman-install system caddy.
- If user already has system caddy, shims must still point to workspace binary.

---

## Nginx (Arch) — Source Build Installer

**Why source**: Upstream tarballs are source; we must build into user prefix to avoid `/usr`.

**Build deps (Arch)** (document only; installer does not install them):
- `base-devel`, `pcre2`, `zlib`, `openssl` (>=1.1 or 3.x as supported).

**Configure & install (into workspace)**
- Download stable nginx source tarball.
- Verify SHA256.
- Configure with:
  ```
  ./configure \
    --prefix=$HOME/.chauffeur/nginx \
    --conf-path=$HOME/.chauffeur/nginx/etc/nginx.conf \
    --pid-path=$HOME/.chauffeur/nginx/nginx.pid \
    --http-log-path=$HOME/.chauffeur/nginx/nginx.access.log \
    --error-log-path=$HOME/.chauffeur/nginx/nginx.error.log \
    --with-pcre-jit \
    --with-http_gzip_static_module \
    --with-http_ssl_module
  ```
- `make -j$(nproc)` then `make install` (no sudo).
- Place shim `~/.chauffeur/bin/nginx` → `~/.chauffeur/nginx/sbin/nginx` or `bin/nginx` depending on the build output.
- Write default `nginx.conf` under `~/.chauffeur/nginx/etc/nginx.conf` if missing (minimal, includes `conf.d/*.conf`).

**Acceptance**
- `nginx -v` works via shim.
- No writes outside workspace.

**Arch-specific notes**
- On Arch, openssl headers may be `openssl` (3.x). Ensure compatibility with chosen nginx release.
- We may add optional modules later (`--with-stream`, `--with-http_v2_module`).

---

## PHP (Arch) — Installer Scaffold (build in step-3)

**Scope now**: Define interface and prerequisites; full build will be covered in **step-3** due to complexity.

**Interface**
- `chauf php install 8.3 [--from source] [--force] [--no-ext]`

**Pre-reqs (Arch)** (document only):
- `base-devel`, `libxml2`, `openssl`, `zlib`, `curl`, `libjpeg-turbo`, `libpng`, `oniguruma`, `icu`, `sqlite`, `readline` (and others as chosen).

**Layout target**
- `~/.chauffeur/php/8.3/bin/php`
- Per-project PHP-FPM pool config handled by later steps.

**Decision**
- We will compile from source with `--prefix=$HOME/.chauffeur/php/8.3` and selective extensions.

---

## Shims (contract)

Create/update wrappers in `~/.chauffeur/bin/`:
- `caddy` → `~/.chauffeur/caddy/bin/caddy`
- `nginx` → `~/.chauffeur/nginx/sbin/nginx` (or `bin/nginx`)
- `php-<major.minor>` → `~/.chauffeur/php/<ver>/bin/php`

Shim properties:
- Executable (0755), no hardcoded absolute `$HOME` in content when possible; expand at runtime.
- Print helpful error if target binary missing.

---

## Minimal Default Configs

- **Caddyfile** at workspace root with global options; project vhosts created by `chauf link`.
- **nginx.conf** with `worker_processes auto;` and `include conf.d/*.conf;`. Provide a default `conf.d/chauffeur-status.conf` disabled by default.

---

## Error Handling & Logs

- On checksum/signature failure → abort with clear message; do not modify existing install.
- If required build deps missing → print distro-specific package names for Arch; exit non-zero.
- All installer logs stored under `~/.chauffeur/logs/installers/*.log` (optional enhancement).

---

## Acceptance Criteria
1. `chauf service install caddy` installs a verified caddy binary into workspace and creates a working shim.
2. `chauf service install nginx` builds and installs nginx into workspace and creates a working shim.
3. Re-running both with no flags makes no changes; with `--force` re-installs/overwrites.
4. Minimal config files are present if previously missing.
5. Distro detector reports `Arch Linux (x86_64)` correctly on Arch.

---

## Out of Scope (this step)
- Building PHP and PHP-FPM (covered in step-3).
- System service units (systemd) — we will run services in foreground or via `chauf start` later.
- Other distros (Ubuntu/Fedora) — add in future steps; keep detector extensible.

---

## Next Step
Proceed to **step-3 — PHP build & PHP-FPM pools (Arch)**, implementing `chauf php install <version>` and generating per-project pools and shims.

