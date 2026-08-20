# Podman Service Management Plan

This plan is Phase 3 of the [master overhaul roadmap](./2026-08-19-overhaul-roadmap.md). It depends on the runtime abstraction and provides the service registry consumed by the link wizard and web UI.

## Goal

Extend Chauffeur's existing Podman database support into a project-aware managed service system for databases, caches, mail testing, search, object storage, queues, and other local development dependencies.

The intended user experience is:

```bash
chauf service list
chauf service add redis
chauf service start redis
chauf service status
```

and from a project:

```bash
chauf link
# select PostgreSQL, Redis, Mailpit, and RustFS in the setup wizard
```

The web UI should expose the same services, status, logs, configuration, connection information, and lifecycle actions.

## Current state

Chauffeur already has a Podman package that supports database-style containers:

- MySQL 5.7 and 8.
- PostgreSQL.
- MariaDB.
- MongoDB.
- Redis.
- Persistent volumes.
- Host port mapping.
- Container lifecycle.
- Database console.
- Backup and restore.

The current model is database-specific: `DatabaseConfig`, `EngineType`, and `Container` contain engine-specific image, environment, port, readiness, and privilege behavior. Redis exists as an engine, but it is not yet modeled as a general service dependency.

The service overhaul should generalize the lifecycle model without breaking existing `chauf podman` commands or existing database config files.

## Why this change

Modern PHP applications commonly depend on more than PHP and one database:

- Redis for cache, queue, sessions, or locks.
- Mailpit for local mail capture.
- Meilisearch, Typesense, or OpenSearch for search.
- RustFS or MinIO for S3-compatible storage.
- RabbitMQ, Beanstalkd, or a queue broker.
- Stripe Mock for payment integration.
- Selenium/Chromium for browser testing.
- Admin UIs such as phpMyAdmin or pgAdmin.

Without a managed service layer, users must manually create containers, remember networks and ports, configure environment variables, and track which service belongs to which project. Lerd's service preset model demonstrates the value of making those dependencies declarative and visible.

## Product boundary

Chauffeur should support a curated service registry first, not an unlimited marketplace.

### Initial built-in services

| Family | Service | Purpose | Initial priority |
|---|---|---|---|
| Database | MySQL 8 | Relational database | Existing |
| Database | MySQL 5.7 | Legacy compatibility | Existing |
| Database | PostgreSQL | Relational database | Existing |
| Database | MariaDB | Relational database | Existing |
| Database | MongoDB | Document database | Existing |
| Cache | Redis | Cache, queue, session, pub/sub | Existing engine, generalize |
| Mail | Mailpit | SMTP capture and mail UI | First new service |
| Search | Meilisearch | Local search engine | First new service |
| Object storage | RustFS | S3-compatible local storage | First new service |
| Queue | RabbitMQ | Message broker | Later |
| Queue | Beanstalkd | Queue backend | Later |
| Search | Typesense/OpenSearch | Search alternatives | Later |
| Testing | Selenium/Chromium | Browser testing | Later |
| Admin | phpMyAdmin/pgAdmin | Database browser | Later |

Do not add a service merely because an image exists. Every built-in service needs a tested image, data policy, health check, port policy, logs path, environment contract, and removal/recovery behavior.

## Service definition model

Replace the database-only configuration with a generic service definition while preserving a compatibility loader for existing files.

Suggested workspace config:

```yaml
name: redis
family: cache
image: docker.io/library/redis:7-alpine
container_name: chauf-redis
description: Redis cache and queue service
ports:
  - host: 6379
    container: 6379
data:
  enabled: true
  host_path: ~/.chauffeur/services/data/redis
  container_path: /data
environment:
  REDIS_ARGS: ""
healthcheck:
  command: ["redis-cli", "ping"]
  interval: 2s
  timeout: 2s
  retries: 30
dashboard:
  url: http://127.0.0.1:8025
  external: false
connection:
  scheme: redis
  host: chauf-redis
  port: 6379
```

### Fields

- `name`: stable logical service name.
- `family`: database, cache, mail, search, storage, queue, admin, testing.
- `image`: validated OCI image reference.
- `container_name`: generated runtime identity.
- `description`: user-facing explanation.
- `ports`: host/container mappings; internal ports are service-defined.
- `data`: persistence policy and mount target.
- `environment`: non-secret runtime configuration.
- `secrets`: references to protected workspace secret storage, not plaintext UI data.
- `healthcheck`: readiness command and timing.
- `dependencies`: services that must be available first.
- `dashboard`: optional local UI URL.
- `connection`: generated connection metadata/DSN template.
- `tuning`: optional config override support.
- `project_bindings`: projects that declare or use the service.

## Presets and custom services

### Built-in presets

Built-in presets should be embedded or shipped with the binary for offline use. A preset is metadata, not Go code, wherever possible:

```yaml
name: mailpit
family: mail
image: docker.io/axllent/mailpit:latest
ports:
  - host: 1025
    container: 1025
  - host: 8025
    container: 8025
dashboard:
  url: http://127.0.0.1:8025
connection:
  scheme: smtp
  host: 127.0.0.1
  port: 1025
healthcheck:
  command: ["wget", "-q", "-O", "/dev/null", "http://127.0.0.1:8025/api/v1/info"]
```

### Custom services

Support user-defined OCI services after built-ins are stable:

```bash
chauf service add ./mailpit.yaml
chauf service add \
  --name my-cache \
  --image docker.io/library/redis:7-alpine \
  --port 6380:6379 \
  --data /data
```

Custom service definitions should live under a user config directory or workspace registry. They must be schema-validated and cannot declare arbitrary host mounts, privileged mode, or host networking by default.

The initial custom schema should support:

- Image.
- Ports.
- Environment variables.
- Persistent data mount.
- Health check.
- Dependencies.
- Dashboard URL.
- Connection metadata.

Defer arbitrary commands, host mounts, and privileged capabilities until there is a clear security model.

## CLI contract

### Canonical commands

```bash
chauf service list
chauf service info <name>
chauf service add <preset|file>
chauf service remove <name>
chauf service start [name|all]
chauf service stop [name|all]
chauf service restart <name>
chauf service status [name]
chauf service logs <name> [--follow]
chauf service exec <name> <command...>
chauf service port <name> <host-port>
chauf service config <name>
chauf service update <name>
```

### Compatibility aliases

Keep the current database commands during migration:

```bash
chauf podman create
chauf podman start
chauf podman stop
chauf podman list
chauf podman status
chauf podman remove
chauf podman console
chauf podman backup
chauf podman restore
```

They should delegate to the generic service layer, with deprecation guidance pointing users toward `chauf service`.

### Service creation flow

Interactive creation should use the same compact TUI language as the link wizard:

```text
❯ chauf service add

  Service preset
  > Redis       cache and queues
    Mailpit     local SMTP capture
    Meilisearch search engine
    RustFS      S3-compatible storage

  Host port
  > 6379

  Persistent data
  > enabled  ~/.chauffeur/services/data/redis

  Review
  Redis 7 on chauf-net
  127.0.0.1:6379 → redis:6379

enter apply
```

Creation must show image, ports, persistence, dependencies, and dashboard URL before applying.

## Project service bindings

Services must be usable without forcing every project to own every container.

Project config should store logical requirements:

```yaml
services:
  - redis
  - mailpit
  - postgres
```

The workspace owns the actual service instance. A project binding can resolve to an existing instance:

```yaml
service_bindings:
  redis: chauf-redis
  postgres: chauf-postgres
```

Rules:

- A project may reference a service that is not installed; status should show the missing dependency.
- Linking should detect service hints from `.env` and framework configuration but must ask before creating anything.
- Starting a project may offer to start bound services, with a visible confirmation/policy.
- Removing a project must not remove shared services or their data.
- Removing a service must show all projects that depend on it.
- The service layer must distinguish “selected by project” from “running globally.”

## Environment integration

The service system should generate connection information without blindly editing `.env`.

Provide:

```bash
chauf env service redis --project ./my-app
chauf env service redis --project ./my-app --apply
```

The web UI can offer a reviewed env diff. Generated values should use the correct hostname based on execution location:

```text
Host-side PHP:       127.0.0.1:<mapped-port>
Container-side PHP:  chauf-redis:6379
```

This distinction matters during the native-to-Podman PHP migration. Project runtime containers should use the Podman network name; host tools should use the mapped loopback port.

Never store passwords in project service requirements. Credentials belong to protected workspace/service secret state and are revealed only through explicit UI/CLI actions.

## Networking

All managed services should join one workspace network:

```text
chauf-net
├── chauf-nginx
├── chauf-php83-fpm
├── chauf-postgres
├── chauf-redis
├── chauf-mailpit
└── chauf-meilisearch
```

Requirements:

- Create network idempotently.
- Use stable container DNS names.
- Avoid host port exposure unless requested or required for host tools.
- Detect host port conflicts before creation.
- Support IPv4/IPv6 behavior explicitly where Podman differs.
- Do not use host networking by default.
- Keep database ports private to the network when project containers are the only consumers.

## Persistence and data safety

Every service preset must declare whether it is:

- Ephemeral.
- Persistent by default.
- Persistent but user-selectable.

Removal behavior:

```text
chauf service remove redis
  → stop service
  → remove container
  → preserve data directory
  → remove runtime metadata
```

Explicit purge should be separate:

```bash
chauf service remove redis --purge
```

Before purge:

- Show exact data path and size.
- Show projects using the service.
- Require confirmation.
- Create a backup/snapshot where the service supports it.
- Rename data aside rather than immediately deleting it when practical.

Existing database backup/restore behavior must be retained and generalized only after the database path remains compatible.

## Service lifecycle and dependencies

Use a dependency graph:

```text
phpmyadmin → mysql
project → postgres, redis, mailpit
```

Start behavior:

1. Resolve dependencies.
2. Check image/config/network.
3. Start dependencies in order.
4. Wait for health checks.
5. Start the requested service.
6. Return a structured operation result.

Stop behavior:

- Stop dependents first when the user explicitly stops a dependency.
- Never stop a shared service merely because one project stopped.
- Explain which projects remain dependent on a service.

Restart behavior:

- Restart only the selected service by default.
- Offer dependent restart only when required by configuration changes.
- Preserve data and report readiness.

## Health and status

Service status must go beyond “container exists”:

```text
name:       chauf-redis
family:     cache
image:      redis:7-alpine
status:     healthy
container:  running
network:    chauf-net
host port:  6379 → 6379
projects:   shop, blog
data:       persistent, 12 MB
dashboard:  unavailable
```

States:

```text
not installed
image pulling
created
starting
healthy
unhealthy
stopped
missing/orphaned
failed
removing
```

Health checks should be service-defined and bounded. A failed health check should include captured evidence and a next action.

## Configuration and tuning

Provide a typed configuration path first:

```bash
chauf service config redis
```

The first version can show the generated config and supported settings. Later it may open a validated tuning override:

```yaml
tuning:
  target: /usr/local/etc/redis/redis.conf
  template: |
    # Chauffeur service tuning
```

Requirements for editable tuning:

- Never overwrite user changes.
- Validate before restart.
- Show diff.
- Backup before apply.
- Report whether restart is required.
- Keep the override separate from the preset.

## Web UI integration

The Services section should include:

- Installed services.
- Available presets.
- Add service flow.
- Per-service status/detail.
- Start/stop/restart.
- Logs.
- Ports and connection info.
- Project usage.
- Data size and backup state.
- Dashboard/open link.
- Config/tuning.
- Remove/reinstall/update.

The project detail page should show:

```text
Services
  PostgreSQL   healthy   chauf-postgres
  Redis        stopped   chauf-redis       Start
  Mailpit      missing   Add service
```

The link wizard should use the same preset registry and service plan API.

## Service API

Proposed endpoints:

```text
GET    /api/services
GET    /api/services/presets
GET    /api/services/{name}
POST   /api/services
POST   /api/services/{name}/start
POST   /api/services/{name}/stop
POST   /api/services/{name}/restart
DELETE /api/services/{name}
GET    /api/services/{name}/logs
GET    /api/services/{name}/logs/stream
GET    /api/services/{name}/connection
GET    /api/services/{name}/config
PUT    /api/services/{name}/config
GET    /api/services/{name}/projects
GET    /api/operations/{id}
```

Mutating calls return an operation ID for image pulls, creation, health waits, backups, restores, and migrations.

Connection responses must redact credentials by default:

```json
{
  "service": "redis",
  "host": "127.0.0.1",
  "port": 6379,
  "scheme": "redis",
  "url": "redis://127.0.0.1:6379",
  "password": null,
  "passwordAvailable": false
}
```

## Registry and update policy

Start with built-in YAML-like definitions embedded in the binary. Each preset should include:

- Version/image.
- Update strategy.
- Supported architectures.
- Health check.
- Port defaults.
- Data policy.
- Dependencies.
- Connection template.
- Dashboard URL.
- Backup support.

An external preset registry is a later feature. If introduced, it must support:

- Signature/checksum or trusted source policy.
- Schema validation.
- Offline cache.
- Version compatibility.
- Rollback to the last valid preset.
- No arbitrary code execution from preset files.

Never allow a remote preset to silently request privileged mode, host networking, arbitrary host mounts, or shell scripts.

## Security requirements

- Rootless Podman only by default.
- Validate service slugs and image references.
- Use argument arrays, never shell interpolation.
- Restrict mounts to generated workspace data paths unless explicitly approved.
- Reject privileged, host-network, and device options in the initial custom schema.
- Protect credentials at rest and in API/UI output.
- Redact environment values and logs.
- Require confirmation for removal, purge, restore, and port changes.
- Show dependent projects before destructive actions.
- Bind admin dashboards to loopback by default.
- Do not expose service dashboards or ports to LAN without an explicit opt-in.

## Phased implementation

### Phase 0: Generalize the existing database model

- Introduce generic service identity/status types.
- Keep `DatabaseConfig` loader compatibility.
- Fix all service paths to honor the shared workspace root.
- Add stable `chauf-net` and container labels.
- Add fakeable Podman lifecycle operations.

### Phase 1: Service command aliases

- Add `chauf service list/info/start/stop/restart/status/logs`.
- Delegate existing `chauf podman` lifecycle commands to the service layer.
- Preserve database backup/restore behavior.
- Add structured status and health results.

### Phase 2: First non-database presets

- Add Redis as a generic cache service.
- Add Mailpit.
- Add Meilisearch.
- Add RustFS.
- Define image, port, data, health, connection, and dashboard metadata.
- Add unit and real-container tests per preset.

### Phase 3: Project bindings and link wizard

- Add logical service requirements to project config.
- Detect project service hints.
- Integrate service multi-select into `chauf link`.
- Show create/start actions only after confirmation.
- Add service dependency status to `chauf status` and project detail.

### Phase 4: Web UI service management

- Add services list/detail pages.
- Add preset picker and operation progress.
- Add logs, connection metadata, project usage, and dashboard links.
- Add safe remove/reinstall flows.
- Add project service cards.

### Phase 5: Configuration and recovery

- Add validated tuning overrides.
- Add service update/rollback.
- Add service-specific snapshots/migrations.
- Add data directory rename-aside recovery.
- Add dependency-aware destructive-action warnings.

### Phase 6: Custom services

- Add validated custom OCI service YAML.
- Add safe flags-based creation.
- Add schema/version validation.
- Add import/export for shareable definitions without secrets.

### Phase 7: Add-on presets

- Add RabbitMQ, Beanstalkd, Typesense/OpenSearch, admin UIs, and browser testing services based on validated demand.
- Add multi-version database preset support only after data-version and port policies are proven.

## Verification strategy

### Unit tests

- Preset schema parsing and validation.
- Service name/image/port validation.
- Legacy database config compatibility.
- Service dependency graph ordering/cycles.
- Project binding resolution.
- Environment/connection template rendering.
- Secret redaction.
- Port conflict detection.
- Data path containment.
- Operation state transitions.

### Container tests

For every built-in preset:

- Pull/create container.
- Join `chauf-net`.
- Mount data correctly.
- Apply environment safely.
- Map expected ports.
- Pass health check.
- Produce usable logs.
- Stop/restart without data loss.
- Remove while preserving data.
- Recreate from config.

### Project integration tests

- Link a Laravel fixture with PostgreSQL, Redis, and Mailpit.
- Verify project runtime reaches services by container DNS names.
- Verify host tools reach only intentionally published ports.
- Stop/start a shared service while multiple projects are bound.
- Remove a project without removing shared services.
- Refuse destructive service removal with dependent projects unless confirmed.
- Verify connection info is correct for host and container contexts.

### Web UI tests

- Service list states.
- Add preset flow and operation progress.
- Project binding display.
- Logs and connection info.
- Redacted secrets.
- Remove/purge confirmation.
- Offline/Podman-unavailable states.
- Keyboard and responsive behavior.

## Success criteria

The service system succeeds when:

- Users can create and manage databases, Redis, Mailpit, search, and object-storage services from CLI and web UI.
- Existing `chauf podman` database workflows continue to work.
- Projects can declare logical service requirements without owning shared service data.
- Link wizard service choices use real availability/status information.
- Services share a stable rootless Podman network.
- Dependencies start in the correct order and health is visible.
- Removing a project never deletes shared service data.
- Removing a service is explicit, dependency-aware, and recoverable.
- Host/container connection targets are not confused.
- No credentials are exposed by default.

## Immediate next implementation slice

Start with the smallest useful generalization:

1. Add generic service identity/status types around the current database container code.
2. Add `chauf service list`, `info`, `start`, `stop`, `status`, and `logs`.
3. Keep `chauf podman` as compatibility aliases.
4. Promote Redis from an engine-only path to a first-class cache service.
5. Add one Mailpit preset with health check, ports, persistence policy, and dashboard URL.
6. Test creation/status/removal with fake Podman commands before adding more presets.
