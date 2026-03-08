# Project Linking Specification

## Overview

Project linking registers a local directory as a Chauffeur-managed project. It generates nginx configs, assigns a PHP version, and sets up the `.test` domain routing.

---

## Project Config Schema

File: `~/.chauffeur/projects/<slug>/config.yaml`

```yaml
slug: my-project                  # URL-safe identifier derived from directory name
path: /home/user/Projects/my-app  # Absolute path to project directory
domain: my-project.test           # Primary domain
aliases:                          # Additional domains (optional)
  - admin.my-project.test
  - api.my-project.test
php_version: "8.3"                # PHP version for this project
ssl: false                        # SSL enabled?
fpm:
  dedicated: false                # Use dedicated FPM pool?
  socket: ""                      # Auto-populated if dedicated: true
project_type: laravel             # Detected: laravel, wordpress, generic
created_at: "2025-01-01T00:00:00Z"
updated_at: "2025-01-01T00:00:00Z"
```

---

## `chauf link`

Registers the current working directory as a project.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--php <version>` | string | global default | PHP version for this project |
| `--secure` | bool | false | Enable SSL from the start |
| `--dedicated-fpm` | bool | false | Use dedicated PHP-FPM pool |
| `--alias <domain>` | string | none | Add alias domain (repeatable) |

### Flow

1. **Validate** current directory exists and is readable
2. **Generate slug** from directory name (lowercase, hyphens)
3. **Check for conflicts**: is this path already linked?
4. **Detect project type**: Laravel (check `artisan`), WordPress (check `wp-config.php`), generic
5. **Validate PHP version** (if specified): must be installed
6. **Create project config** at `~/.chauffeur/projects/<slug>/config.yaml`
7. **Generate nginx site config** from template
8. **Symlink** `sites-enabled/<slug>.conf` → `../sites-available/<slug>.conf`
9. **If SSL**: call mkcert to generate certificate (see [ssl.md](./ssl.md))
10. **If dedicated FPM**: generate dedicated PHP-FPM pool config
11. **Reload nginx** if running (send SIGHUP)
12. **Output** summary with domain, PHP version, FPM strategy

### Idempotency

Running `chauf link` on an already-linked path:
- Updates the nginx config with current flags
- Does NOT overwrite if no flags change
- Reloads nginx if config changed

### Output Example

```
✓ Project linked

  Domain:     http://my-project.test:8080
  Path:       /home/user/Projects/my-app
  PHP:        8.3 (shared FPM)
  Type:       Laravel

  To start services: chauf start
  To enable SSL:     chauf secure
```

---

## `chauf links`

Lists all registered projects.

### Output Format

```
 Project           Domain                    PHP    FPM        SSL
─────────────────────────────────────────────────────────────────────
 my-app            my-app.test               8.3    shared     HTTP
 admin-panel       admin-panel.test          8.1    dedicated  HTTPS
                   admin.admin-panel.test    8.1    dedicated  HTTPS (*)
 legacy-site       legacy-site.test          7.4    shared     HTTP

(*) alias domain
```

### Flags

| Flag | Description |
|------|-------------|
| `--detail` | Show full paths, socket paths, config paths |

---

## `chauf unlink`

Unregisters a project.

### Flags

| Flag | Description |
|------|-------------|
| `--alias <domain>` | Remove a specific alias domain only |
| `--all` | Remove all aliases, then unlink the project |

### Default Flow (full unlink)

1. Show confirmation prompt with all domains and SSL status
2. Remove symlink `sites-enabled/<slug>.conf`
3. Remove config `sites-available/<slug>.conf`
4. If dedicated FPM: stop dedicated PHP-FPM, remove pool config
5. If SSL: optionally remove cert files (ask user)
6. Remove project dir `~/.chauffeur/projects/<slug>/`
7. Reload nginx if running

### Alias-Only Flow (`--alias`)

1. Validate the alias exists on the project
2. Remove from `aliases` in project config
3. Regenerate nginx site config (without the alias)
4. If SSL: regenerate SAN cert without the alias
5. Reload nginx

---

## `chauf php isolate <version>`

Sets a per-project PHP version override.

1. Validate version is installed in workspace
2. Write `.chauffeur-version` to project root with the version string
3. Update project config `php_version`
4. Regenerate nginx site config (FPM socket path may change for shared pools)
5. Reload nginx if running

```
✓ PHP version set

  Project:  my-app
  PHP:      8.1 (was 8.3)
  File:     /home/user/Projects/my-app/.chauffeur-version
```

---

## Project Type Detection

Detection logic in `internal/projects/detect.go`:

| Type | Detection | Document Root |
|------|-----------|--------------|
| `laravel` | `artisan` file exists in root | `public/` |
| `wordpress` | `wp-config.php` or `wp-login.php` exists | root `/` |
| `generic` | Fallback | root `/` or `public/` if exists |

The detected type affects the nginx config template used:
- Laravel: `try_files` routes to `public/index.php`
- WordPress: `try_files` includes permalink rewriting
- Generic: Simple PHP routing

---

## Slug Generation

```go
// internal/projects/slug.go

func GenerateSlug(dirPath string) string {
    name := filepath.Base(dirPath)
    // Convert to lowercase
    slug := strings.ToLower(name)
    // Replace spaces and underscores with hyphens
    slug = strings.NewReplacer(" ", "-", "_", "-").Replace(slug)
    // Remove non-alphanumeric characters (except hyphens)
    slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
    // Collapse multiple hyphens
    slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
    // Trim leading/trailing hyphens
    slug = strings.Trim(slug, "-")
    // Limit length
    if len(slug) > 50 {
        slug = slug[:50]
    }
    return slug
}
```

---

## Migration: `chauf migrate`

Moves a project's Chauffeur config to a different workspace.

```bash
chauf migrate my-app ~/.chauffeur-work
```

1. Copy `~/.chauffeur/projects/my-app/` to `~/.chauffeur-work/projects/my-app/`
2. Update `path` in the new workspace's project config if path changed
3. Remove project from source workspace
4. Reload both workspaces' nginx (if running)
