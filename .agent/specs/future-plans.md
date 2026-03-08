# Future Plans

Features planned beyond the V2 core implementation.

---

## V2.1 — Plugin System Foundation

**Goal**: Allow extending Chauffeur with custom service types (Redis, MySQL, Node proxies, etc.)

**API Design**:
- Plugins are Go plugins (`.so`) or external executables with a defined interface
- Each plugin can register: install, start, stop, status, doctor check handlers
- Plugin config stored in `~/.chauffeur/plugins/<name>/`

**Commands**:
```bash
chauf plugin install <name>      # Install a plugin
chauf plugin remove <name>       # Remove a plugin
chauf plugin list                # List installed plugins
chauf plugin update <name>       # Update a plugin
```

**Example Plugins** (community-built):
- `chauf-redis` — local Redis instance per workspace
- `chauf-mysql` — local MySQL 8.x instance
- `chauf-mailhog` — local SMTP for email testing
- `chauf-minio` — local S3-compatible storage

---

## V2.2 — Multi-Workspace Support

**Goal**: Allow multiple named workspaces (e.g., work, personal, client-a).

**Concept**:
```bash
chauf workspace create work
chauf workspace use work
chauf workspace list
chauf workspace switch personal
```

Each workspace is a separate `~/.chauffeur-<name>/` directory with its own services.

**Use case**: Separate nginx ports and PHP versions per client/context without service conflicts.

---

## V2.3 — Monitoring Dashboard

**Goal**: Real-time monitoring of all Chauffeur services.

**Features**:
- Live TUI dashboard (`chauf monitor`) with service status, request rates, memory
- Historical uptime tracking
- Slow request alerts
- Per-project request metrics
- Log aggregation with search

**Technical approach**:
- Embedded HTTP server in `chauf monitor` serving a local web dashboard
- Collect metrics from nginx access logs and PHP-FPM status endpoints
- Store in SQLite under `~/.chauffeur/metrics.db`

---

## V2.4 — Project Templates

**Goal**: Bootstrap new projects from templates.

```bash
chauf new laravel my-project        # Create + link a new Laravel project
chauf new wordpress my-blog         # Create + link a new WordPress install
chauf new generic my-app            # Create + link a blank PHP project
```

**Flow**:
1. Run framework installer (composer create-project, etc.)
2. Change into new directory
3. Run `chauf link` with sensible defaults for the template type

---

## V2.5 — Remote Workspace Sync

**Goal**: Sync workspace configuration (not services, just configs) across machines.

**Use case**: Developer works on laptop and desktop — projects and configs stay in sync.

**Technical approach**:
- Export/import workspace config to a git repo
- `chauf sync push` / `chauf sync pull`
- Configs only — compiled services are rebuilt locally

---

## Known Limitations (V2)

These are V2 limitations acknowledged for V3 or later:

1. **No Windows/macOS support** — Linux-only by design. macOS users should use Laravel Herd.
2. **No Docker integration** — Chauffeur is the alternative, not a Docker wrapper.
3. **No distributed config** — Single-machine only.
4. **PHP compilation time** — 5-10 minutes for first PHP version install. Acceptable for development tools.
5. **No PHP extension manager** — Extensions must be compiled in at install time. Adding a new extension requires reinstalling that PHP version.

---

## Deferred Decisions

**Plugin system architecture**:
- Options: Go plugins (`.so`), Lua scripts, WASM, external executables
- Consideration: Security, portability, ease of plugin development
- Status: Deferred to V2.1 — need V2 core stable first

**Multi-workspace strategy**:
- Options: Active workspace switching, simultaneous workspaces, workspace inheritance
- Status: Deferred to V2.2

**Metrics storage**:
- Options: SQLite, flat files, Prometheus, in-memory only
- Status: Deferred to V2.3
