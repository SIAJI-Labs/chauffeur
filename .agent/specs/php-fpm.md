# PHP-FPM Specification

## Overview

Chauffeur supports two PHP-FPM strategies that can coexist in the same workspace:

| Strategy | Default | Pool | Socket Path | Best For |
|----------|---------|------|-------------|---------|
| **Shared** | ✅ | One pool per PHP version, shared across projects | `/tmp/chauffeur-<version>.sock` | Most projects — resource efficient |
| **Dedicated** | via `--dedicated-fpm` | One pool per project | `~/.chauffeur/projects/<slug>/php-fpm.sock` | Isolation, different FPM settings per project |

---

## Shared FPM Strategy

A single PHP-FPM pool per PHP version serves all projects using that version.

**Socket**: `/tmp/chauffeur-<version>.sock` (e.g., `/tmp/chauffeur-8.3.sock`)

**Benefits**:
- One PHP-FPM process for multiple projects
- Lower memory overhead
- Faster start times

**Limitations**:
- All projects share the same FPM settings (pm.max_children, etc.)
- One project's FPM crash affects others on same version

**Config**: `~/.chauffeur/php/<version>/etc/php-fpm.conf`

```ini
[global]
pid = {{ .PHPDir }}/runtime/php-fpm/php-fpm.pid
error_log = {{ .PHPDir }}/var/log/php-fpm.log

[chauffeur-{{ .Version }}]
listen = /tmp/chauffeur-{{ .Version }}.sock
listen.owner = {{ .User }}
listen.group = {{ .User }}
listen.mode = 0660

user = {{ .User }}
group = {{ .User }}

pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 5
pm.max_requests = 500

access.log = {{ .PHPDir }}/var/log/php-fpm-access.log
slowlog = {{ .PHPDir }}/var/log/php-fpm-slow.log
request_slowlog_timeout = 5s

php_admin_value[error_log] = {{ .PHPDir }}/var/log/php-fpm.log
php_admin_flag[log_errors] = on
```

---

## Dedicated FPM Strategy

Each project gets its own PHP-FPM pool, running with the project's PHP version.

**Flag**: `chauf link --dedicated-fpm`

**Socket**: `~/.chauffeur/projects/<slug>/php-fpm.sock`

**Benefits**:
- Project isolation — one project's FPM can't affect another
- Per-project FPM tuning possible
- Cleaner log separation

**Limitations**:
- More memory per project
- More processes to manage

**Config**: `~/.chauffeur/projects/<slug>/php-fpm.conf`

```ini
[global]
pid = {{ .ProjectDir }}/php-fpm.pid
error_log = {{ .ProjectDir }}/logs/php-fpm.log

[{{ .Slug }}]
listen = {{ .ProjectDir }}/php-fpm.sock
listen.owner = {{ .User }}
listen.group = {{ .User }}
listen.mode = 0660

user = {{ .User }}
group = {{ .User }}

pm = dynamic
pm.max_children = 5
pm.start_servers = 1
pm.min_spare_servers = 1
pm.max_spare_servers = 3
pm.max_requests = 200

access.log = {{ .ProjectDir }}/logs/php-fpm-access.log
slowlog = {{ .ProjectDir }}/logs/php-fpm-slow.log
request_slowlog_timeout = 5s
```

**Binary used**: `~/.chauffeur/php/<project.php_version>/bin/php-fpm`

---

## Mixed Strategy

A workspace can have both shared and dedicated projects simultaneously:

```
my-app           → shared PHP 8.3 FPM (socket: /tmp/chauffeur-8.3.sock)
another-app      → shared PHP 8.3 FPM (same socket as my-app)
isolated-app     → dedicated FPM     (socket: ~/.chauffeur/projects/isolated-app/php-fpm.sock)
legacy-app       → shared PHP 7.4 FPM (socket: /tmp/chauffeur-7.4.sock)
```

nginx routes each project to its correct socket:

```nginx
# my-app (shared 8.3)
fastcgi_pass unix:/tmp/chauffeur-8.3.sock;

# isolated-app (dedicated)
fastcgi_pass unix:/home/user/.chauffeur/projects/isolated-app/php-fpm.sock;
```

---

## Process Lifecycle

### Start

```
chauf start
  → Start nginx
  → For each unique PHP version in shared projects: start shared FPM pool
  → For each dedicated project: start dedicated FPM pool
```

Order:
1. Start shared PHP-FPM pools (alphabetical by version)
2. Start dedicated PHP-FPM pools
3. Start nginx (last — depends on FPM pools being ready)

### Stop

```
chauf stop
  → Stop nginx first
  → Stop all FPM pools (shared and dedicated)
```

### Restart PHP-FPM

```bash
chauf restart php         # Restart all FPM pools (graceful: SIGUSR2)
chauf restart fpm 8.3     # Restart shared PHP 8.3 pool only
chauf restart --project ./my-app  # Restart dedicated FPM for my-app
```

---

## Status Display

`chauf status`:

```
Service             Status      Uptime     Memory
──────────────────────────────────────────────────
nginx               ● running   2h 34m     45 MB
php-fpm 8.3         ● running   2h 34m     128 MB    [shared, 3 projects]
php-fpm 8.1         ● running   2h 34m     96 MB     [shared, 1 project]
php-fpm 7.4         ○ stopped
fpm: isolated-app   ● running   2h 34m     64 MB     [dedicated]
```

`chauf status --detail` adds:
- Worker process count (master + workers)
- Active connections
- Slow request count
- FPM socket path

---

## Switching Strategy

To switch a project from shared to dedicated FPM (or vice versa):

```bash
chauf unlink             # Unlink project
chauf link --dedicated-fpm  # Relink with new strategy
```

Or in V2, via config:

```bash
chauf config set fpm.dedicated true    # Will regenerate FPM config and nginx config
```

---

## Health Checks

`chauf doctor` FPM checks:

1. PHP-FPM binary exists for each installed version
2. PHP-FPM config is valid: `php-fpm --fpm-config <conf> --test`
3. FPM socket exists and is writable (for running pools)
4. PID file exists and process is alive (for running pools)
5. No zombie FPM processes
6. Log file growth check (warn if > 100MB)
