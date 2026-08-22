# Project Commands

Commands that register and manage projects.

---

## `chauf link`

**Purpose**: Register the current directory (or a specified path) as a Chauffeur-managed project. Chauffeur detects Laravel, WordPress, PHP, and JavaScript projects, then generates either a PHP/FPM or reverse-proxy nginx site config and sets up `.test` domain routing.

Running `chauf link` on an already-linked project updates its configuration — it does not fail or create a duplicate.

**Usage**:
```
chauf link [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--php <version>` | string | global default | PHP version for this project (e.g. `8.3`, `7.4`) |
| `--type <type>` | string | auto-detected | Explicitly choose `laravel`, `wordpress`, `php`, or `reverse-proxy` when detection is unavailable or should be overridden |
| `--proxy-port <port>` | int | `3000` | Local development-server port for a reverse-proxy project |
| `--secure` | bool | false | Generate a trusted SSL certificate and serve over HTTPS |
| `--dedicated-fpm` | bool | false | Provision a dedicated PHP-FPM pool for this project instead of using the shared pool |
| `--alias <domain>` | string | — | Add an additional `.test` domain pointing to this project. Repeatable. |
| `--project <path>` | string | CWD | Register a project at a specific path instead of the current directory |
| `--force` | bool | false | Overwrite existing nginx config even if project is already linked |

**Examples**:

```bash
# Link current directory with defaults (shared FPM, global PHP version, HTTP)
chauf link

# Link with a specific PHP version
chauf link --php 8.1

# Link and enable HTTPS immediately
chauf link --secure

# Link with a dedicated PHP-FPM pool (isolated from other projects)
chauf link --dedicated-fpm

# Link with multiple domain aliases
chauf link --alias admin.my-app.test --alias api.my-app.test

# Link a JavaScript development server through nginx
chauf link --type reverse-proxy --proxy-port 5173

# Link with all options at once
chauf link --php 8.1 --secure --dedicated-fpm --alias admin.my-app.test

# Add an alias to an already-linked project (re-run link with --alias only)
chauf link --alias reporting.my-app.test

# Link a project at a specific path
chauf link --project ~/Projects/client-site --php 8.2 --secure
```

**Output**:
```
Linking my-app...

  ✓ my-app.test → ~/Projects/my-app
  ✓ PHP 8.3  ·  shared FPM  ·  nginx reloaded

  http://my-app.test:8080

  chauf start
  chauf secure
```

**Slug generation**: The project slug is derived from the directory name — lowercased, spaces and underscores replaced with hyphens, special characters removed (e.g. `My Laravel App` → `my-laravel-app`).

**Project type detection**:

| Type | Detection |
|------|-----------|
| `laravel` | `artisan` file in project root |
| `wordpress` | `wp-config.php` or `wp-login.php` in project root |
| `php` | A PHP entry point such as `index.php`, `public/index.php`, or another root-level `.php` file |
| `reverse-proxy` | `package.json` or a recognized JavaScript framework config such as Vite, Next, Nuxt, Astro, Angular, or Svelte |
| unknown | No supported marker; interactive links ask for a type, while noninteractive links require `--type` |

The detected type selects the appropriate nginx template. Reverse-proxy sites forward HTTP and WebSocket traffic to `127.0.0.1:<proxy-port>` and do not require an installed PHP version.

In interactive mode, the wizard first displays the detection result and offers:

- Continue with the detected setup.
- Change project type and choose Laravel, WordPress, PHP, or reverse proxy.

Reverse-proxy setup also offers detected conventional ports plus `Custom port`.
The custom value must be between `1` and `65535`.

**Offline DNS warning**: When DNS is not protected against network-state changes, `chauf link` prints a warning with the `/etc/hosts` lines needed to guarantee offline access:

```
  ⚠ DNS not protected against offline disconnect.

    Add to /etc/hosts to guarantee offline access:
      127.0.0.1  my-app.test

    Or fix permanently:
      chauf doctor --check-dns --fix
```

The project still links successfully — the warning is informational only.

**Notes**:
- If nginx is not running, the nginx config is written but no reload happens — configs load when services start.
- `--alias` can be passed multiple times in the same command.
- Using `--alias` on an existing project **adds** the alias without relinking the whole project.
- Domain conflicts (alias already used by another project) are detected and reported before any changes are made.

---

## `chauf links`

**Purpose**: List all registered projects in a formatted table, showing domain, PHP version, FPM strategy, and SSL status.

**Usage**:
```
chauf links [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--detail` | bool | false | Show full paths, socket paths, config file locations |

**Examples**:

```bash
# Show all projects
chauf links

# With full paths
chauf links --detail
```

**Output**:
```
  Project       Domain                 PHP    SSL  FPM
  ────────────  ─────────────────────  ─────  ───  ─────────
  my-app        my-app.test            8.3    ✗    shared
  admin-panel   admin-panel.test       8.1    ✓    dedicated
                admin.admin-panel.test              (alias)
  legacy-site   legacy-site.test       7.4    ✗    shared
```

Alias domains are indented under their parent project.

**Empty state**:
```
  No projects linked.

  To link your first project:
    cd ~/Projects/my-app
    chauf link
```

**With `--detail`**, each project entry adds: `Path`, `Socket`, `Config`, `Created`.

---

## `chauf unlink`

**Purpose**: Unregister a project. Removes the nginx site config and symlink, stops any dedicated PHP-FPM pool, and optionally removes SSL certificates.

**Usage**:
```
chauf unlink [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--alias <domain>` | string | — | Remove only this alias domain from the project, leaving the project linked |
| `--all` | bool | false | Remove all alias domains first, then unlink the project |
| `--project <path>` | string | CWD | Unlink a project at a specific path |
| `--slug <slug>` | string | — | Unlink by project slug instead of path |
| `--force` | bool | false | Skip the confirmation prompt |

**Examples**:

```bash
# Unlink the project in the current directory
chauf unlink

# Remove a single alias without unlinking the project
chauf unlink --alias admin.my-app.test

# Unlink by slug (from outside the project directory)
chauf unlink --slug my-app

# Unlink a project at a specific path
chauf unlink --project ~/Projects/client-site

# Skip confirmation (for scripts)
chauf unlink --force
```

**Output** (confirmation):
```
Unlinking my-app...

  Domains    my-app.test  ·  admin.my-app.test (alias)  ·  api.my-app.test (alias)

  Removes:
    ~/.chauffeur/nginx/etc/sites-available/my-app.conf
    ~/.chauffeur/nginx/etc/sites-enabled/my-app.conf
    ~/.chauffeur/projects/my-app/

  Type 'my-app' to confirm: _
```

**After confirmation**:
```
  ✓ my-app unlinked
```

**Alias-only unlink** (`--alias`):
```
Removing alias admin.my-app.test...

  ✓ Alias removed  ·  nginx config regenerated  ·  SSL cert updated  ·  nginx reloaded
```

**Notes**:
- SSL certificate files are **not** automatically removed on unlink. Use `chauf clean ssl-certs` to remove stale certs.
- If nginx is running, it is reloaded after config removal — the site disappears immediately.
- The project's source files are never touched.

---

## `chauf secure`

**Purpose**: Enable HTTPS for the project in the current directory. Generates a locally-trusted SSL certificate via `mkcert` covering all project domains (primary + aliases), regenerates the nginx site config with an HTTPS server block and HTTP→HTTPS redirect, and reloads nginx.

**Usage**:
```
chauf secure [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | CWD | Secure a project at a specific path |
| `--slug <slug>` | string | — | Secure by project slug |

**Examples**:

```bash
# Enable SSL for the current project
chauf secure

# Enable SSL for a project at a specific path
chauf secure --project ~/Projects/client-site
```

**Output**:
```
Enabling SSL for my-app...

  ✓ my-app.test.crt  (my-app.test  ·  admin.my-app.test, expires 2035-01-01)
  ✓ nginx reloaded

  https://my-app.test:8443
```

**If mkcert is not installed**:
```
  ✗ mkcert not found

    Arch:          sudo pacman -S mkcert
    Debian/Ubuntu: sudo apt install mkcert
    Manual:        go install filippo.io/mkcert@latest
```

**If mkcert CA is not installed**:
```
  ⚠ mkcert CA not installed. Run once:
    mkcert -install

  Then: chauf secure
```

**Notes**:
- Certificate covers all current aliases via SAN. Adding an alias later (`chauf link --alias`) automatically regenerates the cert.
- Certificate is valid for 10 years (mkcert default).
- Files: `~/.chauffeur/nginx/certs/<domain>.crt` (0644) and `.key` (0600).

---

## `chauf unsecure`

**Purpose**: Disable HTTPS for the project. Regenerates the nginx site config as HTTP-only and removes the HTTP→HTTPS redirect. Does not delete the certificate files.

**Usage**:
```
chauf unsecure [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | CWD | Unsecure a project at a specific path |
| `--slug <slug>` | string | — | Unsecure by project slug |

**Examples**:

```bash
# Disable SSL for the current project
chauf unsecure

# Disable SSL for a specific project
chauf unsecure --slug legacy-site
```

**Output**:
```
Disabling SSL for my-app...

  ✓ nginx config updated (HTTP only)  ·  nginx reloaded

  http://my-app.test:8080

  Certificate files kept. To remove: chauf clean ssl-certs --project my-app
```

**Notes**:
- Certificate files are intentionally preserved — re-enabling SSL with `chauf secure` is instant if domains haven't changed.
- The project's SSL status in `config.yaml` is updated to `ssl: false`.
