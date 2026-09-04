# Native and Podman Runtime Parity Investigation

Date: 2026-08-30

## Summary

All four reported problems are confirmed.

1. `chauf podman` is a legacy database/cache-container command, not the
   selected PHP/nginx runtime command. Its name became ambiguous after
   `runtime.engine: podman` was added.
2. `start`, `stop`, `status`, and `logs` have separate native and Podman command
   paths and data models. The differences in rows, metadata, failure handling,
   project scoping, and URLs are implementation gaps rather than necessary
   runtime differences.
3. The observed `arvi-ui.test` 502 has a confirmed project-specific cause.
   `arvi-ui` is a reverse-proxy project whose Vite server listens only on IPv6
   loopback (`[::1]:3901`). Native nginx can reach it through `localhost`, but
   containerized nginx tries `host.containers.internal:3901`, which is refused.
4. Auto-start is native-only. Enabled systemd user units directly execute
   native nginx and PHP-FPM without reading `runtime.engine`, so they can start
   after the workspace has switched to Podman and occupy Podman's host ports.

The preferred command rename is **`chauf podman-db`**, with `chauf podman`
retained temporarily as a deprecated compatibility alias. `podmandb` should not
be used because it is less readable and does not follow the CLI's kebab-case
style.

## Scope and evidence levels

This document distinguishes:

- **Source-confirmed**: directly demonstrated by current repository code.
- **Runtime-confirmed**: also reproduced against the current user workspace on
  2026-08-30 with read-only diagnostic commands.
- **Latent risk**: a source-confirmed defect not responsible for this specific
  report.

No application or runtime source was changed during this investigation. The
worktree already contained unrelated uncommitted runtime changes, so the
findings describe the current worktree rather than only the last commit.

## 1. Ambiguous `chauf podman` command

### Finding — confirmed

`cmd/chauf/main.go:123-126` dispatches the top-level `podman` command to
`commands.RunPodman`. `internal/commands/podman.go` implements database and
cache container operations for MySQL, PostgreSQL, MariaDB, MongoDB, and Redis.
It is separate from runtime selection and from top-level service lifecycle
commands.

The two current meanings are:

| Command | Actual responsibility |
| --- | --- |
| `chauf podman ...` | Shared database/cache container create, start, stop, status, console, backup, restore, and related operations |
| `chauf config runtime podman` | Select Podman as the PHP-FPM/nginx runtime used by top-level lifecycle commands |
| `chauf start/stop/status/logs` | Operate the selected PHP-FPM/nginx runtime |

Global help reinforces the ambiguity by placing database operations under the
generic heading “Podman” (`cmd/chauf/main.go:289-298`).

### Recommended fix

1. Make `chauf podman-db` the canonical name for the current database/cache
   command group.
2. Keep `chauf podman` as a compatibility alias for at least one documented
   deprecation window. Print a concise warning that points to
   `chauf podman-db`.
3. Do not rename the existing on-disk `podman/` state, database containers,
   volumes, backups, or internal package as part of the CLI rename.
4. Update global/subcommand help and command docs, including currently omitted
   implemented subcommands such as `import`.
5. Add dispatcher/help tests proving that the new and legacy names operate on
   the same existing state.

The broader service-management plan proposes eventually moving these resources
behind `chauf service ...` while preserving compatibility aliases. That future
work must not block the clarity improvement requested here.

### Workaround

Until the alias exists, interpret `chauf podman ...` strictly as the
database/cache command and use `chauf config runtime ...` plus top-level
lifecycle commands for PHP/nginx.

## 2. Native and Podman lifecycle/output differences

### Finding — confirmed

The repository has a `runtime.Runtime` interface
(`internal/runtime/runtime.go:75-88`), but it is not yet a complete common
lifecycle implementation:

- Podman implements container lifecycle and status.
- The native adapter's `Start`, `Stop`, and `Restart` are no-ops; `Status`
  returns no rows; and `Logs` returns an unsupported error
  (`internal/runtime/runtime.go:742-764`).
- `internal/commands/services.go` therefore branches on the engine and uses
  `services.Manager` directly for native while using `runtime.Podman` for
  Podman.

This produces the reported differences.

### Start comparison

Native `RunStart` receives a `[]services.StartResult` and renders one result per
FPM pool and nginx, including PID and already-running state
(`internal/commands/services.go:73-108`, `:174-186`). It also prints linked
project URLs.

Podman `startPodmanServices` prints planned PHP container names before calling
`EnsureWorkspace`, prints no nginx row, receives only an aggregate error, and
prints no linked URLs (`internal/commands/services.go:424-441`).

Consequences:

- nginx is absent from Podman's start result even though it is reconciled;
- bullets represent planned names rather than confirmed service outcomes;
- Podman cannot report created, started, already running, restarted, or failed
  per service;
- native and Podman URL footers differ;
- `--project` is bypassed in Podman mode because the runtime branch occurs
  before project handling.

### Stop comparison

Native stop returns and prints a result for nginx and every FPM pool. Podman
stop emits no success row per service, converts individual failures to
warnings, and always prints “All services stopped”
(`internal/commands/services.go:459-491`). This can report aggregate success
after a partial failure.

Podman `--project` is also bypassed. A dedicated-container naming risk exists
because some stop/status scopes omit the registered project slug and may fall
back to the project directory basename (`internal/runtime/runtime.go:824-835`).

### Status comparison

Native status directly reads process metadata and passes PID, uptime, and
memory into `printStatusRow` (`internal/commands/services.go:722-791`).

Podman status inspects only `State.Status` and defines healthy as
`state == "running"` (`internal/runtime/runtime.go:650-690`). The Podman row
renderer then hard-codes `-` for PID, uptime, and memory
(`internal/commands/services.go:368-377`). Podman can provide this data through
container inspect/stats; it is simply not collected by the current status
model.

Podman detail has useful container-specific fields (`Container`, `Image`, and
`Evidence`), while native detail has process-specific paths (`Config`, error
log, and socket). These should be additional details beneath one common base
status rather than separate output formats.

### Logs comparison

The same split affects logs (`internal/commands/services.go:869-947`):

- Podman ignores `--project` because it branches before project resolution.
- Podman `access` and `error` both select the same container output.
- Podman `php` without a version chooses only the default PHP version, while
  native aggregates installed PHP logs.
- The current command runner buffers `podman logs --follow`; it does not provide
  true incremental streaming.

### Recommended shared model

Commands should resolve resources once, invoke one runtime contract, and render
one result shape. Runtime-specific data should extend the shared output rather
than replace it.

At minimum, expand the shared service model to support:

- stable service ID, label, role, and project/scope association;
- lifecycle state and health/readiness state;
- PID, start time, uptime, and memory when available;
- runtime-neutral evidence and warnings;
- optional native details: config, socket, PID/log paths;
- optional Podman details: container, image/digest, network, and published
  ports.

Add a shared operation result for start/stop/restart/reconcile with one entry
per service:

- before and after state;
- changed/already-running/already-stopped state;
- service metadata;
- warning/error evidence.

Then:

1. Implement the native runtime adapter over the existing `services.Manager`,
   `NginxService`, and `FPMService`.
2. Make Podman reconciliation return structured per-service outcomes rather
   than only `error`.
3. Use one project/service inventory resolver so slugs and shared/dedicated
   ownership remain stable across every action.
4. Use one start/stop/status renderer for both engines.
5. Preserve FPM-before-nginx startup and nginx-before-FPM shutdown for both
   engines.
6. Print the same linked URLs after a successful start.
7. Never print aggregate success if any required service failed.
8. Change the log contract to stream to writers (or an iterator) so Podman
   follow mode is actually interactive.

### Expected output contract

The default table should use the same columns and ordering for both runtimes:

```text
 Runtime          podman
 Service                 Status        PID       Uptime      Memory
 ───────────────────────────────────────────────────────────────────
 nginx                   ● running    12345     12s         8 MB
 php-fpm 7.4 (shared)    ● running    12346     12s         18 MB
 php-fpm 8.3 (shared)    ● running    12347     11s         20 MB
```

If a field is genuinely unavailable, `-` remains valid, but the adapter must
attempt to collect equivalent data first. `--detail` may add engine-specific
sections without changing the base model.

### Workaround

For missing Podman metadata:

```bash
podman container inspect chauf-nginx chauf-php83-fpm
podman stats --no-stream
```

For live Podman logs:

```bash
podman logs --follow chauf-nginx
podman logs --follow chauf-php83-fpm
```

Do not rely on Podman `start --project`, `stop --project`, or project-specific
log selection until their target-resolution paths are unified.

## 3. `arvi-ui.test` returns 502 only on Podman

### Finding — runtime-confirmed root cause

This incident is a reverse-proxy reachability problem, not a PHP-FPM problem.

Evidence:

1. `~/.chauffeur/projects/arvi-ui/config.yaml` registers `arvi-ui` as
   `project_type: reverse-proxy` on port `3901`.
2. Native generated config proxies to `localhost:3901` at
   `~/.chauffeur/nginx/etc/sites-available/arvi-ui.conf`.
3. Podman generated config proxies to
   `host.containers.internal:3901` at
   `~/.chauffeur/nginx/container.conf`.
4. The host listener was reproduced as `[::1]:3901`. A host request to
   `http://[::1]:3901/` returned 200, while `127.0.0.1:3901` was refused.
5. Inside `chauf-nginx`, `host.containers.internal` resolved to `169.254.1.2`,
   and connecting to port 3901 returned “Connection refused.”
6. nginx logged:

   ```text
   connect() failed (111: Connection refused) while connecting to upstream
   upstream: "http://169.254.1.2:3901/"
   server: arvi-ui.test
   ```

Native succeeds because its nginx process shares the host loopback namespace.
Podman nginx has a separate network namespace and cannot reach a service bound
only to host loopback through the host gateway.

The route selection is implemented in
`internal/runtime/nginx_config.go:50-59`; workspace startup does not probe the
reverse-proxy target from the container network before claiming success.

### Immediate workaround

Start the Vite server on an interface reachable from the Podman host gateway:

```bash
pnpm dev -- --host 0.0.0.0
# or
pnpm exec vite dev --host 0.0.0.0 --port 3901
```

Then verify both the listener and the container path:

```bash
ss -ltnp 'sport = :3901'
podman exec chauf-nginx \
  wget -S -O /dev/null -T 3 http://host.containers.internal:3901/
curl -k --resolve arvi-ui.test:8443:127.0.0.1 \
  https://arvi-ui.test:8443/
```

Binding to `0.0.0.0` can expose the development server beyond loopback. Use an
appropriate host firewall and document the exposure. Switching back to native
is the fallback workaround.

### Recommended fix

1. Do not reconcile or validate linked projects as part of `chauf start` or
   `chauf stop`. Those commands operate existing service identities only;
   project/container reconciliation belongs to link/apply (or an explicit
   repair operation).
2. If reverse-proxy reachability is requested explicitly (for example through
   `doctor`, a diagnostic command, or detailed status), probe the host endpoint
   from the nginx container/network context rather than only from the host.
3. If the host process is loopback-only, return a degraded/failed result with
   the project, target, and a framework-neutral explanation. An example command
   such as `--host 0.0.0.0` may be shown as a workaround, not silently applied.
4. Include optional reverse-proxy diagnostics in `status --detail` and `doctor`,
   but do not make ordinary lifecycle commands fail because of them.
5. Add a Podman integration test for an unreachable loopback-only host fixture
   and a reachable host-gateway fixture.
6. Do not mark a linked route ready merely because the Podman runtime was
   selected or the nginx container is running.

### Related latent Podman defects discovered

These did not cause the `arvi-ui` 502 but should be addressed while hardening
runtime parity:

- Multiple dedicated projects all use `/workspace`, so aggregate nginx mounts
  can overwrite each other (`internal/runtime/nginx_config.go:61-77`). Use a
  unique path such as `/workspace/projects/<slug>` for every project.
- Existing container reconciliation validates mounts/selected labels but not
  all required image, network, published-port, and workspace identity. A stale
  container can survive with an invalid topology.
- nginx config rendering passes the configured host HTTPS port as the internal
  listen port, while the container always publishes to internal port 8443
  (`internal/runtime/runtime.go:266`, `:533-536`). Defaults hide the mismatch.
- PHP readiness runs `php -v`, which does not prove that FPM port 9000 accepts
  FastCGI requests (`internal/runtime/runtime.go:456-465`).

These should have explicit tests, but fixes must remain separate enough that
the confirmed reverse-proxy issue can be reviewed and verified independently.

## 4. Auto-start starts native services after selecting Podman

### Finding — source-confirmed; current unit state inspected

`chauf autostart enable` never loads the selected runtime. It always writes and
enables native nginx/PHP units (`internal/commands/autostart.go:53-141`).

The generated units directly execute:

- `%h/.chauffeur/nginx/sbin/nginx`
- `%h/.chauffeur/php/%i/sbin/php-fpm`

See `internal/system/systemd.go:22-63`. They do not invoke `chauf`, read
`chauffeur.yaml`, or contain a Podman/Quadlet definition.

Changing runtime through `chauf config runtime podman` only updates workspace
configuration. It does not stop/disable old native units or migrate auto-start.
Once runtime is Podman, `chauf stop` also returns through the Podman branch and
does not stop a native service left running by systemd.

Current user-state inspection on 2026-08-30 found:

- `chauffeur-nginx.service`: enabled, currently inactive;
- systemd lingering: enabled;
- `chauf-nginx`: running in Podman.

This means the legacy native nginx unit is still eligible to start at boot via
the user manager even though Podman is selected. Whichever nginx starts first
can occupy the configured HTTP/HTTPS host ports and prevent the other from
binding. Podman correctly reports an unavailable host port, but does not
identify or migrate the legacy unit.

The units are attached to the user `default.target`: normally they start on
login; with lingering enabled they can start at boot before login.

There is also a workspace-path defect: CLI commands honor `CHAUFFEUR_HOME`, but
the unit content hard-codes `%h/.chauffeur`.

### Immediate workaround

Disable the native units before using Podman:

```bash
systemctl --user disable --now chauffeur-nginx.service
systemctl --user list-unit-files 'chauffeur-php-fpm@*.service'
# Repeat for every enabled instance, for example:
systemctl --user disable --now chauffeur-php-fpm@8.3.service
```

For ordinary installations, this shortcut may be sufficient:

```bash
chauf autostart disable
```

Direct `systemctl` commands are safer for stale PHP instances because
`chauf autostart disable` only discovers versions still installed under the
current workspace's native `php/` directory.

Verify before starting Podman:

```bash
systemctl --user is-active chauffeur-nginx.service
systemctl --user list-units 'chauffeur-php-fpm@*.service'
chauf config runtime
chauf start
```

Using alternate ports is only a temporary workaround; it leaves both runtimes
active and hides the lifecycle ownership problem.

### Recommended fix

1. Make `autostart enable/disable/status` load the workspace and selected
   runtime.
2. Under native, preserve native units but generate paths from the resolved
   workspace instead of hard-coding `~/.chauffeur`.
3. Under Podman, generate runtime-appropriate user services/Quadlets for the
   network, PHP-FPM containers, and nginx with correct dependency/readiness
   order. Never write or enable native service units in Podman mode.
4. Treat runtime switching as an explicit, reversible migration:
   - detect active/enabled old-runtime units;
   - show what will stop/disable and what will be enabled;
   - stop/disable old runtime ownership before starting the new runtime;
   - preserve native installations and Podman images for rollback;
   - restore previous unit state if migration fails.
5. Add `doctor` diagnostics for selected runtime versus enabled-unit mismatch,
   port owner, lingering, legacy unit content, and custom workspace mismatch.
6. Add upgrade handling for already-written legacy units and stale FPM
   instances, including instances whose native installation directory no
   longer exists.
7. Fix dependency order: current FPM unit declares
   `After=chauffeur-nginx.service`, contrary to the required FPM-before-nginx
   startup order.

## Implementation sequence

To limit risk and preserve compatibility:

1. Add `podman-db` plus the deprecated `podman` alias and tests.
2. Introduce shared service status and operation-result models without changing
   output yet.
3. Implement the native runtime adapter and shared resource inventory.
4. Convert `status`, then start/stop/restart, then logs to shared command paths
   and renderers.
5. Add Podman inspect/stats metadata and true log streaming.
6. Add reverse-proxy reachability checks and the dedicated-path/port/readiness
   hardening identified above.
7. Make auto-start runtime-aware and add legacy-unit migration.
8. Run unit tests and opt-in real Podman integration tests, including real HTTP
   requests through nginx.

Each step should preserve existing native installations, database state,
volumes, backups, and old command compatibility unless a separately documented
migration has completed successfully.

## Acceptance criteria

### Command naming

- `chauf podman-db ...` provides all current database/cache operations.
- `chauf podman ...` still works against the same state and prints a deprecation
  hint.
- Help and docs clearly distinguish database/cache management from
  `runtime.engine: podman`.

### Lifecycle and output parity

- Native and Podman use one service inventory, operation result, and renderer.
- Start and stop print truthful per-service rows for nginx and every selected
  FPM service, including already-running/stopped and failures.
- The same URL list is printed after successful workspace startup.
- Status uses the same base columns/order and reports PID, uptime, and memory
  for Podman when inspect/stats provide them.
- Runtime-specific metadata appears only as additional detail.
- Project-scoped lifecycle and logs select the same logical resources under
  both engines and preserve registered slugs.
- Aggregate success is impossible after a required service failure.
- Follow-mode logs stream incrementally and stop on cancellation.

### Routing/readiness

- A reachable host reverse proxy returns a real successful request through
  Podman nginx.
- A loopback-only host reverse proxy fails readiness with an actionable message
  instead of reporting all services running.
- Shared and multiple dedicated PHP fixtures serve real HTTP responses through
  nginx, with unique mount paths and verified FastCGI readiness.
- Reconciliation replaces resources with stale network, image, mount, port, or
  workspace identity.

### Auto-start and migration

- `runtime.engine: podman` cannot enable or start native nginx/PHP units.
- Switching engines detects and safely migrates old-runtime auto-start state.
- Podman auto-start respects FPM readiness before nginx.
- Custom `CHAUFFEUR_HOME` is preserved in generated service ownership.
- Existing legacy units are detected and migrated during upgrade.
- Native artifacts remain available for rollback.

## Verification matrix for implementation

Required automated coverage:

- command dispatcher/help/legacy alias tests;
- equivalent renderer snapshots for native and Podman fixtures;
- lifecycle order, idempotence, partial failure, and project-scope tests;
- Podman inspect/stats parsing and stopped/absent/unhealthy states;
- streaming and cancellation tests for logs;
- reverse-proxy host-gateway integration tests;
- shared and multiple-dedicated PHP HTTP integration tests;
- auto-start generation, engine-switch migration, legacy-unit upgrade, and
  custom-workspace tests;
- occupied-port diagnostics that identify old-runtime ownership.

Before completion, run the repository's full Go test suite and the opt-in
Podman integration suite (`CHAUFFEUR_PODMAN_INTEGRATION=1`) on a rootless Podman
host. Report skipped environment-dependent checks explicitly.

## Primary implementation files

Expected files include, but are not limited to:

- `cmd/chauf/main.go`
- `internal/commands/podman.go`
- `internal/commands/services.go`
- `internal/commands/autostart.go`
- `internal/commands/doctor.go`
- `internal/runtime/runtime.go`
- `internal/runtime/factory.go`
- `internal/runtime/nginx_config.go`
- `internal/runtime/runner.go`
- `internal/services/manager.go`
- `internal/system/systemd.go`
- related unit and integration tests under `internal/**`
- command documentation under `docs/commands/`

Architecture and compatibility constraints are also documented in:

- `docs/plans/2026-08-19-overhaul-roadmap.md`
- `docs/plans/2026-08-19-podman-runtime-overhaul.md`
- `docs/plans/2026-08-19-podman-service-management.md`
