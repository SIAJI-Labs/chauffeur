# Service Commands

Commands for starting, stopping, and observing running services.

---

## `chauf start`

**Purpose**: Start Chauffeur-managed services (nginx and PHP-FPM pools). By default, starts all services needed by all registered projects. Can be scoped to a single project.

Services start in the correct order: PHP-FPM pools first (shared, then dedicated), nginx last.

**Usage**:
```
chauf start [service] [flags]
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `nginx` | Start only nginx |
| `php` or `php <version>` | Start only PHP-FPM (optionally a specific version's shared pool) |
| *(omitted)* | Start all required services |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | — | Start only the services for a specific project (its PHP-FPM pool and nginx) |
| `--all` | bool | false | Explicit "start everything" (same as no argument) |
| `--dry-run` | bool | false | Show what would be started without starting it |

**Examples**:

```bash
# Start all services
chauf start

# Start only nginx
chauf start nginx

# Start only the PHP 8.3 shared FPM pool
chauf start php 8.3

# Start services for a specific project's dedicated FPM
chauf start --project ~/Projects/my-app

# Preview without starting
chauf start --dry-run
```

**Output**:
```
Starting services...

  ✓ php-fpm 8.3   shared     pid 12345
  ✓ php-fpm 8.1   shared     pid 12346
  ✓ legacy-app    dedicated  pid 12347
  ✓ nginx                    pid 12348

  http://my-app.test:8080
  https://admin-panel.test:8443
  http://legacy-site.test:8080
```

**Already running**: If a service is already running, it is skipped — no error, no restart.

**Port conflict**:
```
  ✗ Port 8080 already in use  (apache2, pid 9999)

    sudo systemctl stop apache2
    — or —
    chauf config set nginx.http_port 8081
```

**DNS check**: On first start (or when `.test` resolution is not detected), prints DNS setup reminder:
```
  ⚠ .test domains not resolving

    Add to /etc/dnsmasq.d/chauffeur.conf:
      address=/.test/127.0.0.1

    sudo systemctl restart dnsmasq
```

---

## `chauf stop`

**Purpose**: Stop Chauffeur-managed services. Sends graceful shutdown signals — nginx receives `SIGQUIT` (waits for active requests), PHP-FPM receives `SIGTERM` (waits for workers). Falls back to `SIGKILL` after 30 seconds.

**Usage**:
```
chauf stop [service] [flags]
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `nginx` | Stop only nginx |
| `php` or `php <version>` | Stop only PHP-FPM (optionally a specific version's pool) |
| *(omitted)* | Stop all services |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | — | Stop only the dedicated FPM pool for a specific project |
| `--all` | bool | false | Stop all services (same as no argument) |
| `--dry-run` | bool | false | Show what would be stopped without stopping it |

**Examples**:

```bash
# Stop all services
chauf stop

# Stop only nginx (PHP-FPM stays running)
chauf stop nginx

# Stop only the PHP 8.1 shared pool
chauf stop php 8.1

# Stop a project's dedicated FPM pool
chauf stop --project ~/Projects/isolated-app
```

**Output**:
```
Stopping services...

  ✓ nginx stopped
  ✓ php-fpm 8.3  stopped
  ✓ php-fpm 8.1  stopped
  ✓ legacy-app   stopped
```

**Already stopped**: If a service is already stopped, it is skipped — no error.

---

## `chauf restart`

**Purpose**: Restart services, using graceful reload where possible. nginx supports zero-downtime config reload via `SIGHUP` (no connection drops). PHP-FPM supports graceful worker reload via `SIGUSR2`.

**Usage**:
```
chauf restart [service] [flags]
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `nginx` | Reload nginx config (SIGHUP — zero connection drops) |
| `php` | Gracefully reload all PHP-FPM pools (SIGUSR2) |
| `php <version>` | Gracefully reload one PHP version's shared pool |
| `fpm` | Alias for `php` |
| *(omitted)* | Restart all services |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | — | Restart only the dedicated FPM pool for a specific project |
| `--all` | bool | false | Restart all services |
| `--dry-run` | bool | false | Show what would be restarted |

**Examples**:

```bash
# Restart all services
chauf restart

# Reload nginx config without dropping connections
chauf restart nginx

# Gracefully reload all PHP-FPM workers
chauf restart php

# Reload only the PHP 8.3 shared pool
chauf restart php 8.3

# Restart a project's dedicated FPM
chauf restart --project ~/Projects/my-app
```

**Output**:
```
Restarting services...

  ✓ nginx reloaded  (config valid, 0 connection drops)
  ✓ php-fpm 8.3 reloaded  (workers replaced gracefully)
  ✓ php-fpm 8.1 reloaded
```

**Config validation before nginx reload**: Before sending `SIGHUP`, the nginx config is validated with `nginx -t`. If validation fails, the reload is aborted:
```
  ✗ nginx config invalid — not reloaded

    nginx: [emerg] unknown directive "invalid_value" in .../nginx.conf:15

    Previous config still active. Fix and retry: chauf restart nginx
```

**Not running**: If a service is not running, `restart` starts it instead of failing.

---

## `chauf status`

**Purpose**: Show the health and state of all running services. Displays status, PID, uptime, and memory for each service. `--detail` adds per-service verbose output with socket paths, config locations, and worker counts.

**Usage**:
```
chauf status [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--detail` | bool | false | Verbose per-service output with paths, process counts, socket info |
| `--project <path>` | string | — | Show status for one project's services only |

**Examples**:

```bash
# Overview of all services
chauf status

# Detailed view with paths and process counts
chauf status --detail

# Status for a specific project
chauf status --project ~/Projects/my-app
```

**Output** (default):
```
  Service            Status     PID      Uptime    Memory
  ─────────────────  ─────────  ───────  ────────  ──────
  nginx              ● running   12348    2h 34m    45 MB
  php-fpm 8.3        ● running   12345    2h 34m   128 MB
  php-fpm 8.1        ○ stopped
  fpm: isolated-app  ● running   12347    1h 02m    64 MB
```

**Status indicators**:

| Symbol | Color | Meaning |
|--------|-------|---------|
| `●` | green | Running, healthy |
| `●` | yellow | Running, degraded (e.g. high slow-request rate) |
| `●` | red | Running, error state (nginx config error, etc.) |
| `○` | gray | Stopped |
| `?` | gray | Unknown (stale PID file, missing socket) |

**With `--detail`**:
```
  nginx
    Status   ● running
    PID      12348
    Uptime   2h 34m
    Memory   45 MB
    Workers  1 master + 4 workers
    Config   ~/.chauffeur/nginx/etc/nginx.conf
    Logs     ~/.chauffeur/nginx/logs/error.log

  php-fpm 8.3  (shared)
    Status   ● running
    PID      12345
    Uptime   2h 34m
    Memory   128 MB
    Workers  2 active / 5 idle (7 total)
    Socket   /tmp/chauffeur-8.3.sock
    Slow     0 slow requests
```

---

## `chauf logs`

**Purpose**: View log output from nginx or PHP-FPM. Supports filtering, real-time tailing, and targeting specific services or projects.

**Usage**:
```
chauf logs [service] [version] [flags]
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| *(omitted)* | nginx error log (default) |
| `nginx` | nginx error log |
| `access` | nginx access log |
| `php` or `php-fpm` | All PHP-FPM logs (interactive selector if multiple versions) |
| `php <version>` | Specific PHP version's FPM log (e.g. `chauf logs php 8.3`) |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--follow`, `-f` | bool | false | Tail the log in real time (like `tail -f`) |
| `--lines <n>` | int | 50 | Number of lines to show |
| `--level <level>` | string | — | Filter by log level: `error`, `warn`, `info`, `debug` |
| `--context` | bool | false | Show the log file path above the output |
| `--project <path>` | string | — | Show logs for a project's dedicated FPM pool |

**Examples**:

```bash
# Show last 50 lines of nginx error log
chauf logs

# Show nginx access log, last 100 lines
chauf logs access --lines 100

# Tail nginx error log in real time
chauf logs nginx --follow

# Show PHP 8.3 FPM log
chauf logs php 8.3

# Tail PHP logs filtered to errors only
chauf logs php --follow --level error

# Show logs for a dedicated FPM project
chauf logs --project ~/Projects/isolated-app

# Show which file is being read
chauf logs nginx --context
```

**Output** (`--context`):
```
  ~/.chauffeur/nginx/logs/error.log

2025/01/15 09:23:01 [error] 12349#0: *1 connect() failed (111: Connection refused)...
2025/01/15 09:23:14 [notice] 12348#0: signal process started
```

**Interactive selector** (when `chauf logs php` matches multiple versions):
```
  Multiple PHP-FPM services. Select one:

    1) php-fpm 8.3  ● running
    2) php-fpm 8.1  ○ stopped
    3) fpm: isolated-app  ● running

  Choice [1-3]:
```

**Log file locations**:

| Service | Log file |
|---------|---------|
| nginx error | `~/.chauffeur/nginx/logs/error.log` |
| nginx access | `~/.chauffeur/nginx/logs/access.log` |
| PHP-FPM shared | `~/.chauffeur/php/<version>/var/log/php-fpm.log` |
| PHP-FPM dedicated | `~/.chauffeur/projects/<slug>/logs/php-fpm.log` |
