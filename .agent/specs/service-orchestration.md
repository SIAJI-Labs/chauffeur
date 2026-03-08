# Service Orchestration Specification

## Overview

Chauffeur manages three types of processes: nginx, shared PHP-FPM pools (per version), and dedicated PHP-FPM pools (per project). The orchestrator coordinates their lifecycle.

---

## `chauf start`

Starts all Chauffeur services.

### Flags

| Flag | Description |
|------|-------------|
| `--project <path>` | Start only the project-specific services |
| `--all` | Start all services (default) |

### Start Order

For full workspace start:

1. **Shared PHP-FPM pools** — one per installed PHP version that has projects using it
2. **Dedicated PHP-FPM pools** — one per project with `fpm.dedicated: true`
3. **nginx** — last, after all FPM pools are ready

Rationale: nginx routes to FPM sockets. If nginx starts before FPM, requests will fail with 502 until FPM is up. Starting FPM first avoids this.

### Socket Readiness

Before starting nginx, verify each expected FPM socket exists and is writable:

```go
func waitForSocket(sockPath string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if _, err := os.Stat(sockPath); err == nil {
            return nil
        }
        time.Sleep(100 * time.Millisecond)
    }
    return fmt.Errorf("socket not ready: %s", sockPath)
}
```

### Already Running

If a service is already running, skip it (idempotent). Do not return an error.

### Output

```
Starting services...
  ✓ php-fpm 8.3 (shared)    pid 12345
  ✓ php-fpm 8.1 (shared)    pid 12346
  ✓ php-fpm 7.4 (isolated)  pid 12347   [dedicated: legacy-app]
  ✓ nginx                   pid 12348

All services running.
  http://my-app.test:8080
  https://secure-app.test:8443
```

---

## `chauf stop`

Stops all Chauffeur services.

### Flags

| Flag | Description |
|------|-------------|
| `--project <path>` | Stop only project-specific dedicated FPM |
| `--all` | Stop all services (default) |

### Stop Order

1. **nginx** — first, stop accepting new requests
2. **Shared PHP-FPM pools** — after nginx stops
3. **Dedicated PHP-FPM pools** — after nginx stops

Signals:
- nginx: `SIGQUIT` (graceful — waits for active requests to complete)
- PHP-FPM: `SIGTERM` (graceful — waits for active workers)

### Timeout

Wait up to 30 seconds for graceful shutdown. If timeout exceeded, send `SIGKILL`.

---

## `chauf restart`

Restarts services without dropping active connections where possible.

### Forms

```bash
chauf restart                    # Restart all services
chauf restart nginx              # Restart nginx only (reload config via SIGHUP)
chauf restart php                # Restart all PHP-FPM pools
chauf restart php 8.3            # Restart shared PHP 8.3 pool only
chauf restart fpm                # Alias for restart php
chauf restart --project <path>   # Restart dedicated FPM for a project
chauf restart --all              # Restart everything
```

### nginx Restart

nginx supports graceful config reload via `SIGHUP` — no connection drops:

```go
func (n *NginxService) Reload() error {
    pid, err := n.readPID()
    if err != nil {
        return n.Start()  // not running, just start it
    }
    return syscall.Kill(pid, syscall.SIGHUP)
}
```

### PHP-FPM Restart

PHP-FPM supports graceful worker reload via `SIGUSR2`:

```go
func (p *PHPFPMService) Reload() error {
    pid, err := p.readPID()
    if err != nil {
        return p.Start()
    }
    return syscall.Kill(pid, syscall.SIGUSR2)
}
```

---

## `chauf status`

Shows service health.

### Default Output

```
Service               Status      PID      Uptime
───────────────────────────────────────────────────
nginx                 ● running   12348    2h 34m
php-fpm 8.3           ● running   12345    2h 34m
php-fpm 8.1           ● running   12346    2h 34m
php-fpm 7.4           ○ stopped
fpm: legacy-app       ● running   12347    2h 34m
```

### `--detail` Output

Adds process counts, memory, socket paths:

```
nginx
  Status:     ● running
  PID:        12348
  Uptime:     2h 34m
  Memory:     45 MB
  Config:     ~/.chauffeur/nginx/etc/nginx.conf
  Error log:  ~/.chauffeur/nginx/logs/error.log

php-fpm 8.3 (shared)
  Status:     ● running
  PID:        12345
  Workers:    2 active / 5 idle (7 total)
  Uptime:     2h 34m
  Memory:     128 MB
  Socket:     /tmp/chauffeur-8.3.sock
  Slow reqs:  0
```

### `--project` Output

Shows status for one project's services:

```bash
chauf status --project ./my-app
```

```
my-app (PHP 8.3, shared FPM)

  nginx       ● running   (handles my-app.test, admin.my-app.test)
  php-fpm 8.3 ● running   (shared with 2 other projects)
```

---

## `chauf logs`

Views service log output.

```bash
chauf logs                         # Show nginx error.log
chauf logs nginx                   # nginx error log
chauf logs access                  # nginx access log
chauf logs php                     # All PHP-FPM logs
chauf logs php 8.3                 # PHP 8.3 FPM log
chauf logs --project ./my-app      # Logs for a specific project
chauf logs --follow                # Tail mode (like tail -f)
chauf logs --level error           # Filter by level
chauf logs --lines 100             # Show last N lines (default: 50)
```

---

## Process State Tracking

Chauffeur reads PID files to determine if processes are running:

```go
// internal/services/nginx.go

func (n *NginxService) IsRunning() bool {
    pid, err := n.readPID()
    if err != nil {
        return false
    }
    // Signal 0: check if process exists without sending real signal
    return syscall.Kill(pid, 0) == nil
}

func (n *NginxService) readPID() (int, error) {
    data, err := os.ReadFile(n.pidPath)
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(strings.TrimSpace(string(data)))
}
```

---

## Port Conflict Detection

Before starting nginx, check if configured ports (8080, 8443) are available:

```go
// internal/lib/ports.go

func IsPortAvailable(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

func FindProcessOnPort(port int) (int, string, error) {
    // Parse /proc/net/tcp or use lsof/ss
    // Returns: pid, process name, error
}
```

If port is in use by Chauffeur's own nginx (e.g., restart scenario), send SIGHUP instead of reporting conflict.

If port is in use by another process:
```
Error: port 8080 is already in use by nginx (pid 99999)

Kill it:
  kill 99999

Or change Chauffeur's port:
  chauf config set nginx.http_port 8081
```

---

## Error Recovery

### FPM Crash Recovery

If PHP-FPM crashes while nginx is running, requests to PHP routes will return 502. Chauffeur does not auto-restart on crash (that's systemd's job when autostart is configured).

Manual recovery:
```bash
chauf restart php 8.3
```

### nginx Config Error

If nginx config test fails on restart/reload:
1. Do NOT apply the bad config
2. Show the nginx error output
3. nginx continues running with the previous config

```
Error: nginx config invalid

  nginx: [emerg] invalid value "invalid_value" in ~/.chauffeur/nginx/etc/nginx.conf:15
  nginx: configuration file ... test failed

Config NOT reloaded. Previous config still active.
```
