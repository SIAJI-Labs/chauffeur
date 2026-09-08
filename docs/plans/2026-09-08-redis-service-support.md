# Redis Service Support Plan

Date: 2026-09-08

## Goal

Promote Redis from an engine-specific option in `chauf podman-db` to a
first-class managed cache service, while preserving the existing Redis
container configuration and compatibility commands.

The initial user experience should support:

```bash
chauf service add redis
chauf service start redis
chauf service status redis
chauf service logs redis
chauf service exec redis redis-cli PING
chauf service remove redis
```

The existing workflow must continue to work during migration:

```bash
chauf podman-db create redis
chauf podman-db start redis
chauf podman-db stop redis
chauf podman-db status redis
chauf podman-db console redis
```

## Current state and evidence

Redis is already present, but only as a database-style engine:

- `internal/podman/config.go` defines `EngineRedis`, the
  `docker.io/library/redis:7-alpine` image, port `6379`, and the generated
  `chauf-redis` identity.
- `internal/commands/podman.go` includes Redis in interactive creation,
  engine validation, lifecycle resolution, console selection, and help text.
- `internal/podman/container.go` maps Redis to container port `6379`, joins
  `chauf-net`, mounts persistent data, exposes `redis-cli`, and includes
  Redis-specific backup/list behavior.
- `internal/podman/paths.go` stores configuration below the shared workspace
  root at `podman/` and data below `podman/volumes/`.
- `docs/plans/2026-08-19-podman-service-management.md` explicitly identifies
  Redis as an existing engine that must be generalized into a cache service.

The missing work is the generic service identity/preset/binding layer, not a
new container image. The current Redis backup implementation also needs
verification before it is exposed as a supported service: `redisDump` does
not reliably wait for `BGSAVE` completion, and restore currently expects a
dump file to already exist in the volume rather than explicitly copying the
backup payload into it.

There is also a current UX inconsistency: the generic `podman-db create`
prompt asks for username and password for Redis, and the command output shows
the generated password, even though Redis is created without authentication.
This is misleading and should be corrected as part of the service migration.

## Scope

### In scope

- A built-in Redis cache preset with stable metadata.
- Generic service lifecycle operations backed by the existing Podman client.
- Redis connection information for host-side and container-side consumers.
- Persistent data, health/readiness, logs, console/exec, and safe removal.
- Project-level logical service requirements for Redis, without project-owned
  Redis data.
- Compatibility loading of existing Redis `DatabaseConfig` files.
- Unit tests using fake Podman command execution and opt-in real-container
  integration tests.

### Out of scope for the first slice

- Redis Cluster, Sentinel, replication, or multiple Redis versions.
- Remote or LAN exposure by default.
- Arbitrary Redis configuration files or custom service marketplace support.
- Password protection in the default preset.
- Destructive data purge as part of ordinary removal.
- Replacing all database backup/restore code in one change.

## Proposed Redis preset

The preset should be embedded with the service registry so it works offline:

```yaml
name: redis
family: cache
image: docker.io/library/redis:7-alpine
container_name: chauf-redis
description: Redis cache, queue, session, lock, and pub/sub service
ports:
  - host: 6379
    container: 6379
data:
  enabled: true
  host_path: <workspace>/podman/volumes/chauf-redis
  container_path: /data
healthcheck:
  command: ["redis-cli", "PING"]
  timeout: 2s
  retries: 30
connection:
  scheme: redis
  host: chauf-redis
  port: 6379
```

Host tools should receive `127.0.0.1:6379` (or the configured mapped port),
while PHP/nginx containers on `chauf-net` should receive `chauf-redis:6379`.
Connection output must not include a password unless explicitly requested.

## Implementation sequence

### 1. Establish the service model and compatibility boundary

- Introduce or complete generic service identity, family, state, health,
  connection, and operation types around the existing `podman.DatabaseConfig`.
- Add a Redis preset registry entry and map legacy `EngineRedis` configs to it.
- Keep the existing config filenames, `chauf-redis` container name, volume
  path, `chauf-net` network, and `chauf podman-db` dispatch intact.
- Ensure all paths use `workspace.Root()` and add migration/compatibility tests
  for an existing Redis YAML file.

### 2. Reuse Podman lifecycle primitives through shared operations

- Extract the common create/start/stop/status/logs/exec/remove behavior from
  the database-specific command path into service operations.
- Make operations idempotent and return structured before/after states,
  warnings, and errors rather than relying only on printed output.
- Ensure Redis creation validates host-port conflicts, creates the network and
  persistent volume, pulls the pinned image, and starts the container.
- Add stable labels/metadata so Redis can be discovered without treating every
  `chauf-*` container as a database.

### 3. Implement Redis health and connection behavior

- Use `redis-cli PING` as the bounded readiness check and distinguish missing,
  stopped, starting, healthy, unhealthy, and orphaned states.
- Parse inspect/port output into the shared status model, including container,
  image, network, published port, and evidence.
- Add a connection-info operation that renders host and container targets.
- Keep the default Redis preset unauthenticated and make future password
  support an explicit secret-state change rather than a plaintext config field.
- Make interactive creation conditional on service capabilities: Redis should
  not prompt for database username/password while unauthenticated, and its
  summary should not print a generated password that the container does not
  use.

### 4. Preserve and harden Redis-specific data operations

- Treat Redis persistence as an RDB snapshot under `/data/dump.rdb`.
- Implement a correct snapshot flow: trigger `BGSAVE`, poll `LASTSAVE` or
  `Persistence.rdb_bgsave_in_progress`, enforce a timeout, then copy the
  completed dump bytes.
- Define backup metadata as Redis/cache data rather than a SQL “database”.
- Restore into a temporary file, stop Redis, atomically replace `dump.rdb`,
  start Redis, and wait for `PING`; leave the original data recoverable if any
  step fails.
- Decide and document whether `list` reports Redis logical DBs (`db0`, etc.) or
  keyspace statistics; do not present keyspaces as SQL databases.

### 5. Add CLI and project integration

- Add `chauf service list|info|add|start|stop|status|logs|exec|remove` for the
  Redis preset using the shared service operations.
- Preserve `chauf podman-db ... redis` as a compatibility alias and retain
  `console redis` as a convenient `redis-cli` entry point.
- Add Redis to the link wizard's service selection and project config as a
  logical requirement (`services: [redis]`) only after the service registry
  exists.
- Show missing/stopped Redis dependencies and require confirmation before
  creating or starting a shared service.
- Update command help and `docs/commands/services.md` with examples,
  connection targets, data policy, and removal semantics.

### 6. Add panel support after the CLI contract is stable

- Expose Redis through the service/preset API rather than adding Redis-only
  handlers to the existing database panel.
- Display status, port, connection information, logs, project usage, start /
  stop / restart, console/exec, and safe remove actions.
- Redact secrets and make destructive removal visibly distinct from purge.

## Data and removal policy

Normal removal should:

1. Show the data path, current status, and projects using Redis.
2. Require confirmation.
3. Stop Redis if necessary.
4. Remove the container and service metadata.
5. Preserve `podman/volumes/chauf-redis`.

`--purge` must be a separate, explicitly confirmed operation that reports the
exact path and size before deletion. Removing a project must never remove this
shared Redis data.

## Verification plan

### Unit tests

- Preset defaults: image, family, port, container name, data mount, and health
  command.
- Legacy Redis config load/save compatibility.
- Service-to-Podman argument construction with no shell interpolation.
- Port conflict and workspace path handling.
- Health/status parsing for running, stopped, missing, and unhealthy states.
- Host/container connection target rendering.
- Lifecycle idempotence, partial failure, and safe removal behavior.
- RDB backup polling timeout and restore rollback behavior.
- Secret redaction in config, status, logs, and API responses.
- Interactive prompt/output snapshots proving Redis does not request or display
  unused credentials, while database engines retain their credential prompts.

### Opt-in Podman integration test

Run with `CHAUFFEUR_PODMAN_INTEGRATION=1` on rootless Podman:

1. Create Redis with a temporary workspace and `chauf-net`.
2. Verify `redis-cli PING` from inside the container.
3. Verify a host client reaches the intentionally published port.
4. Verify a second `chauf-*` container reaches Redis by DNS name.
5. Write/read a key, restart Redis, and verify persistence.
6. Create and restore an RDB backup, then verify the key/value.
7. Verify logs, status, stop/start, and removal preserve the volume.
8. Verify an unrelated project removal leaves Redis and its data intact.

## Acceptance criteria

- Redis appears as a cache service, not only as a database engine.
- `chauf service ... redis` supports the documented lifecycle and observation
  commands.
- Existing `chauf podman-db ... redis` workflows continue to operate on the
  same container, config, volume, and network.
- Redis is healthy only after `redis-cli PING` succeeds.
- Redis creation does not ask users for, generate, or display unused
  credentials when authentication is disabled.
- Host and container connection targets are correct and clearly distinguished.
- Creation is rootless, networked through `chauf-net`, and persistent by
  default without exposing ports beyond the requested loopback mapping.
- Backup/restore is verified against real Redis data or explicitly marked
  unsupported until the snapshot implementation is correct.
- Normal removal preserves data; purge is separate, confirmed, and visible.
- CLI, link wizard, and panel use shared service operations rather than
  Redis-specific mutation logic.

## Dependencies and rollout

This plan depends on the Phase 0 generic service model and the Phase 1/2
runtime and link-wizard contracts described in
`docs/plans/2026-08-19-overhaul-roadmap.md` and
`docs/plans/2026-08-19-podman-service-management.md`.

Roll out in this order: service model and compatibility, CLI lifecycle,
health/connection, data safety, project bindings, then panel integration.
Each stage must pass existing Podman database tests before the next stage is
started.
