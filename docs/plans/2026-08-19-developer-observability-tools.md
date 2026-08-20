# Developer Observability Tools Plan

This plan covers Phases 5–8 of the [master overhaul roadmap](./2026-08-19-overhaul-roadmap.md). Environment intelligence comes before invasive instrumentation; the debug bridge depends on the Podman runtime transport and panel security foundation.

## Goal

Add a local developer-tools service to Chauffeur’s web UI, inspired by the useful parts of Herd Pro and Lerd, so developers can inspect and debug projects without stitching together terminal commands, log files, framework configuration, and ad-hoc instrumentation.

The service should reduce debugging time while remaining:

- Local-first.
- Opt-in for invasive instrumentation.
- Explicit about overhead and captured data.
- Safe with secrets and user data.
- Compatible with the Podman PHP runtime migration.
- Useful from the web UI and CLI.

## Proposed feature set

### Initial feature groups

1. **Log viewer**
   - Nginx access/error logs.
   - PHP-FPM/container logs.
   - Application logs.
   - Operation/build logs.
   - Search, filters, line limits, pause, follow, copy, and download.

2. **`.env` editor**
   - View and edit a project environment file.
   - Preserve comments and ordering where practical.
   - Validate syntax before writing.
   - Atomic writes, backups, diff preview, and revert.
   - Never render secret values by default.

3. **`.env` configuration comparison**
   - Compare `.env.example`, `.env`, and optionally `.env.testing`.
   - Report keys present in the example but missing from the active env.
   - Report keys present in the active env but absent from the example.
   - Report empty values, duplicate keys, malformed lines, and suspicious production values.
   - Tell the user exactly how many keys/lines differ.

4. **Request logging**
   - Request method, path, status, duration, timestamp, and project.
   - Slow-request threshold and sorting.
   - Query count and memory metadata where available.
   - Exclude static assets and health checks by default.

5. **Performance tracking**
   - Request count and rate.
   - Median/p95 duration.
   - Slowest requests.
   - Status-code distribution.
   - Cold-start and application-boot timing where measurable.
   - Per-project and bounded time-window views.

6. **Dump interception**
   - Capture Laravel `dump()`/`dd()` output into the UI.
   - Preserve request/CLI context, project, timestamp, and source location when available.
   - Make capture toggleable without rebuilding the project.
   - Provide CLI tail/clear/status equivalents.

7. **Eloquent query inspection**
   - Capture query SQL/template, bindings in redacted form, duration, connection, request context, and stack/source location where available.
   - Highlight slow queries and repeated query patterns.
   - Detect likely N+1 behavior heuristically.
   - Provide request-level query totals and query timelines.

8. **Optional profiling/debugging**
   - Xdebug toggle/status.
   - One-request or one-command profiling.
   - Profile report listing and cleanup.
   - Deferred until the lower-risk observability pipeline is stable.

## Product boundary

This is a developer-only local observability service. It is not an application monitoring platform and must not be enabled by default in production-like environments.

The first supported application integration is Laravel. Generic capabilities must remain useful for non-Laravel PHP projects:

| Capability | Generic PHP | Laravel-specific |
|---|---:|---:|
| Nginx/PHP/container logs | Yes | Yes |
| `.env` editor/diff | Yes | Yes |
| Request metadata | Via nginx/PHP bridge | Via framework context enhancement |
| Dump interception | Via explicit bridge/protocol | `dump()`/`dd()` integration |
| Eloquent query capture | No | Yes |
| Query count/N+1 detection | No | Yes |
| Laravel exception/application logs | No | Yes |
| Xdebug/profiling | Runtime-dependent | Runtime-dependent |

Do not pretend that Eloquent query inspection is a generic PHP feature. It requires Laravel hooks or a project package/bootstrap integration.

## Architecture

### Components

```text
PHP project request/CLI
  → optional Chauffeur debug bridge
  → local project debug endpoint/socket
  → Chauf observability collector
  → bounded event store
  → Web UI/API and CLI views
```

The collector should run as part of the local Chauffeur control plane, not as an internet-facing service. With the Podman runtime, the bridge must work across the container boundary.

### Transport

Prefer a Unix socket on Linux when the PHP container can mount it safely:

```text
PHP container → mounted Unix socket → chauf panel/collector
```

Use a loopback/container-reachable TCP fallback where Unix sockets are unavailable or impractical. The transport must support:

- Length-delimited or newline-delimited messages.
- Maximum frame/event size.
- Timeouts and idle connection limits.
- Authentication or an unguessable local token.
- Backpressure behavior.
- Dropping old events instead of blocking application requests.

The debug bridge must never make a project request hang because the UI is closed or the collector is unavailable.

### Runtime injection

For Podman PHP images, mount a small, version-compatible bridge into PHP containers and enable it through configuration. Possible mechanisms:

- PHP `auto_prepend_file` for generic request/CLI hooks.
- A runtime sentinel/config file checked on each request.
- Laravel service provider or bootstrap integration for framework hooks.
- Container-mounted project-specific configuration.

The bridge must be version-aware and tested on PHP 7.4, 8.0, current supported versions, CLI SAPI, and FPM SAPI.

Do not bake project-specific secrets or absolute host paths into published PHP images.

## Event model

All captured data should use typed events with a common envelope:

```json
{
  "id": "evt_01J...",
  "type": "request",
  "project": "my-app",
  "context": "fpm",
  "occurredAt": "2026-08-19T12:00:00Z",
  "durationMs": 182,
  "payload": {}
}
```

Initial event types:

```text
request
query
dump
log
operation
exception
profile
```

### Request event

```json
{
  "type": "request",
  "method": "GET",
  "path": "/orders",
  "status": 200,
  "durationMs": 182,
  "queryCount": 14,
  "project": "shop",
  "requestId": "req_..."
}
```

### Query event

```json
{
  "type": "query",
  "requestId": "req_...",
  "connection": "mysql",
  "sql": "select * from `orders` where `id` = ?",
  "bindings": ["[redacted]"],
  "durationMs": 42,
  "source": "app/Http/Controllers/OrderController.php:31"
}
```

SQL and bindings are sensitive. The default representation should redact values while preserving useful query shape. Full bindings require an explicit unsafe/debug setting and a visible warning.

### Dump event

Dump payloads may contain credentials, tokens, customer data, and full model objects. Apply size limits, type-aware redaction, and retention limits before storing or broadcasting.

## Log viewer

### Sources

The viewer should normalize these sources:

- Nginx access log.
- Nginx error log.
- PHP-FPM/container logs.
- Per-project application log.
- Podman database/container logs.
- PHP image build/rebuild logs.
- Chauffeur operation logs.

Each source needs a stable label, health state, and availability reason.

### Interaction

- Select project/source.
- Choose historical snapshot or live follow.
- Filter by severity/status/text.
- Set line count.
- Pause follow without disconnecting.
- Resume from the last cursor.
- Copy selected lines.
- Download a bounded log excerpt.
- Redact secrets before display/download.

The first release may poll bounded snapshots. If using SSE/WebSocket, implement cursor/reconnect behavior and clearly label the stream as live.

## `.env` editor and configuration intelligence

### Safety rules

- Never edit a file without showing its path and project.
- Preserve permissions where possible.
- Write through a temporary file and atomic rename.
- Create a timestamped backup before mutation.
- Show a diff before applying nontrivial changes.
- Validate syntax and duplicate keys.
- Do not interpolate values into shell commands.
- Redact values in logs, operation events, and error messages.
- Require explicit reveal for sensitive values.

### File selection

The UI should show:

```text
.env
.env.example
.env.testing
.env.local
```

The active file must be identified using framework/project rules. The UI must never silently edit `.env.example` when the user intended `.env`.

### Difference report

The comparison view should report at least:

```text
Example keys: 42
Active keys: 45
Missing from .env: 3
Only in .env: 6
Empty in .env: 2
Duplicate keys: 1
Malformed lines: 0
```

Categories:

- **Missing from active env:** exists in `.env.example`, not in `.env`.
- **Only in active env:** exists in `.env`, not in `.env.example`.
- **Empty:** key exists but has no usable value.
- **Different default:** optional advisory when the example value differs.
- **Duplicate:** repeated key definitions with effective-value explanation.
- **Malformed:** lines that cannot be parsed safely.

Comparison must operate on parsed keys, not raw line count alone. The UI can show both key count and physical line count so comments and blank lines are not confused with configuration entries.

## Eloquent query capture

### Integration approach

For Laravel projects, provide a local integration package or generated bootstrap hook that registers `DB::listen`. The integration sends structured query events to the collector without changing application behavior when disabled.

Possible enablement modes:

```bash
chauf debug queries on --project ./shop
chauf debug queries off --project ./shop
chauf debug queries status --project ./shop
```

The web UI can expose the same toggle per project.

### Capture fields

- Request/CLI correlation ID.
- Connection name.
- SQL with placeholders.
- Redacted bindings.
- Duration.
- Transaction state if available.
- Route/command context.
- Source file/line where available.
- Timestamp.

### Analysis

Initial analysis should remain conservative:

- Total query count per request.
- Total query time.
- Slow query threshold.
- Repeated identical query shape.
- Query count grouped by route.
- Suspected N+1 only as a warning with evidence, never as a definitive claim.

Do not store unlimited SQL history. Use a bounded ring buffer and opt-in persistence for a selected time window.

## Dump interception

### Behavior

When enabled, Laravel `dump()` and `dd()` output should appear in the UI and CLI stream with:

- Project.
- Request/CLI context.
- Timestamp.
- Value type and bounded representation.
- Source location if available.
- Request path/route where available.

`dd()` behavior requires an explicit decision. Recommended default:

- Capture the dump.
- Preserve `dd()` termination semantics for the application request.
- Show the captured event in the UI.

Do not silently change application control flow just to make the UI convenient.

### Redaction and limits

- Maximum serialized payload size.
- Maximum nesting depth.
- Circular-reference handling.
- Password/token/key-name redaction.
- Binary and resource handling.
- No arbitrary object serialization across the transport.

## Request logging and performance

### Collection

Request instrumentation should record only what can be collected safely and cheaply. At minimum:

- Method/path.
- Status.
- Duration.
- Project/site.
- PHP version/runtime.
- Request ID.
- Query count/time when query capture is enabled.

Nginx access logs can provide an independent request record. Application instrumentation can enrich it. The collector must deduplicate or clearly distinguish the two sources.

### Metrics

Per project and time window:

- Request count.
- Requests per minute.
- Median duration.
- p95 duration.
- Slow request count.
- Error rate.
- Status distribution.
- Query count/time distribution when enabled.
- Recent request list.

Default retention should be short and bounded, such as the last 15 minutes in memory and an optional small persisted window. The UI must show the retention window.

### Performance overhead

The UI should show instrumentation state and expected overhead:

```text
Request logging: on
Query capture: off
Dump capture: off
Estimated overhead: low
```

Query bindings, dumps, and profiling can have materially higher overhead and must be explicitly enabled.

## Profiling and Xdebug

These are later features, not prerequisites for the first observability release.

### Xdebug

Provide:

- Per-version/per-project enabled state.
- Mode selection: debug, coverage, profile, trace.
- On-demand/trigger mode.
- Clear restart/reload impact.
- IDE connection guidance.

Xdebug should default off because it changes request performance and may expose debugging ports.

### Profiling

Support:

- One-request or one-command profile.
- Explicit start/stop.
- Bounded report storage.
- Browser/report opening.
- Clear cleanup.

Do not run continuous profiling by default. Treat profile files as sensitive because they may contain paths, SQL, arguments, and application data.

## API surface

The developer-tools API should be project-scoped:

```text
GET  /api/projects/{slug}/logs
GET  /api/projects/{slug}/logs/stream
GET  /api/projects/{slug}/env/files
GET  /api/projects/{slug}/env/compare
GET  /api/projects/{slug}/env/{file}
PUT  /api/projects/{slug}/env/{file}
POST /api/projects/{slug}/env/{file}/revert

GET  /api/projects/{slug}/requests
GET  /api/projects/{slug}/requests/{id}
GET  /api/projects/{slug}/performance

GET  /api/projects/{slug}/queries
GET  /api/projects/{slug}/queries/summary
POST /api/projects/{slug}/queries/enable
POST /api/projects/{slug}/queries/disable

GET  /api/projects/{slug}/dumps
GET  /api/projects/{slug}/dumps/stream
POST /api/projects/{slug}/dumps/enable
POST /api/projects/{slug}/dumps/disable
DELETE /api/projects/{slug}/dumps

GET  /api/projects/{slug}/debug/status
POST /api/projects/{slug}/debug/reload
GET  /api/projects/{slug}/profiles
POST /api/projects/{slug}/profiles
DELETE /api/projects/{slug}/profiles/{id}
```

All endpoints must enforce project identity and path containment. File names must be allow-listed; the API must not accept arbitrary filesystem paths.

## Web UI information architecture

Add developer tools inside the project detail page first:

```text
Project
├── Overview
├── Logs
├── Requests
├── Queries
├── Dumps
├── Performance
├── Environment
│   ├── .env editor
│   └── .env comparison
├── Services
├── Runtime
└── Diagnose
```

The project detail header should always show:

- Project name.
- URL/open action.
- Runtime/PHP version.
- Instrumentation badges.
- Health status.
- Quick actions for logs, diagnose, and environment comparison.

### Environment comparison UX

Show a summary card first, then categorized differences:

```text
3 variables missing
6 variables only in .env
2 variables empty
```

Each category should support:

- Key name.
- Example/active presence.
- Redacted value state.
- Copy key name.
- Add/update action where safe.
- Explanation of why the variable matters when framework metadata provides it.

### Query UX

Show:

- Query count and total time.
- Slowest queries.
- Repeated query shapes.
- Request/route association.
- SQL with bindings redacted by default.
- Source location when known.

Use a query detail drawer rather than displaying full SQL in a dense table.

### Dump UX

Show a chronological stream with:

- Value preview.
- Expand/collapse.
- Request/CLI badge.
- Project/route context.
- Source location.
- Clear/pause controls.

Make the capture state obvious at all times.

## CLI parity

Every operational toggle should have a CLI equivalent for SSH/headless workflows:

```bash
chauf logs --project ./shop --follow
chauf env diff --project ./shop
chauf env edit --project ./shop
chauf debug queries on --project ./shop
chauf debug queries tail --project ./shop
chauf debug dumps on --project ./shop
chauf debug dumps tail --project ./shop
chauf debug performance --project ./shop
chauf debug status --project ./shop
```

The UI and CLI must call the same collector/configuration functions. The web UI is not allowed to create a feature that cannot be diagnosed or disabled from the CLI.

## Security and privacy

These features capture application data and are higher risk than ordinary service status.

- Bind the collector to local IPC by default.
- Require panel authentication/token protection.
- Never send secrets to the browser unless explicitly revealed.
- Redact keys matching password/token/secret/private/key patterns.
- Redact SQL bindings by default.
- Redact authorization/cookie headers from request records.
- Never capture request bodies by default.
- Never capture full environment values in telemetry or bug reports.
- Bound all payload sizes and retention.
- Provide clear “clear captured data” actions.
- Keep capture disabled by default for new projects.
- Log only metadata about toggles, never captured values.
- Do not expose debug endpoints on the public project domain.
- Treat uploaded/downloaded profile and log files as sensitive.

## Performance and reliability constraints

- Capture must be non-blocking or bounded with a fast failure path.
- The collector must tolerate the UI being closed.
- Ring buffers must cap memory usage.
- Slow queries should not cause query capture itself to become a performance incident.
- Sampling must be available for high-volume projects.
- Request correlation IDs must not be guessable or reused across projects.
- A crashed collector must not crash PHP-FPM or nginx.
- Instrumentation changes must be reversible without rebuilding the project.

## Phased implementation

### Phase 0: Observability foundation

- Define event envelope and bounded storage.
- Add collector process/endpoint.
- Add local transport from Podman PHP containers.
- Add redaction, size limits, retention, and clear operations.
- Add panel authentication before exposing captured data.
- Add CLI `debug status` and `debug clear`.

### Phase 1: Logs

- Normalize nginx, PHP-FPM, container, application, and operation logs.
- Add project-scoped API and UI log viewer.
- Add bounded snapshot first.
- Add follow/stream with reconnect after cursor support is implemented.

### Phase 2: `.env` intelligence

- Add safe parser preserving comments/order where possible.
- Add `.env`/`.env.example` comparison.
- Add counts for missing/extra/empty/duplicate/malformed keys.
- Add read-only UI first.
- Add diff-preview/atomic-write editor.
- Add backups/revert and CLI parity.

### Phase 3: Request tracking

- Add request IDs and basic request events.
- Ingest nginx/application metadata.
- Add project request list and performance summary.
- Add slow-request threshold and bounded retention.
- Exclude static/health traffic by default.

### Phase 4: Laravel dumps and Eloquent queries

- Add Laravel bridge/provider integration.
- Add dump capture with bounded serialization/redaction.
- Add query listener with redacted bindings.
- Add query count/time and slow query views.
- Add cautious repeated-query/N+1 warnings.
- Add UI and CLI toggles/tails.

### Phase 5: Performance and debugging controls

- Add Xdebug status/toggle.
- Add one-shot profiling.
- Add profile list/view/cleanup.
- Add optional sampling and per-project instrumentation presets.
- Add overhead/status indicators.

### Phase 6: Framework expansion

- Add framework adapters behind a stable bridge interface.
- Support Symfony/Doctrine query capture where useful.
- Support WordPress application log conventions.
- Keep generic PHP behavior independent of framework packages.

## Verification strategy

### Unit tests

- `.env` parsing, duplicate handling, comments, quoting, multiline values, and diff counts.
- Redaction rules and secret-key matching.
- Event serialization and size limits.
- Ring-buffer retention and clear behavior.
- Request/query/dump correlation.
- SQL binding redaction.
- Path containment and project identity.
- Operation toggle state transitions.

### Runtime tests

- PHP 7.4, 8.0, current PHP versions.
- CLI and FPM SAPI.
- Podman transport unavailable/slow/full scenarios.
- Collector restart while an application request is running.
- Large dumps and circular objects.
- High query volume and sampling.
- Laravel `DB::listen`, `dump()`, and `dd()` behavior.
- A project request remains successful with the bridge disabled or collector offline.

### UI tests

- Log source/filter/follow behavior.
- `.env` counts and categorized diff states.
- Secret redaction and explicit reveal.
- Query list/detail and redacted bindings.
- Dump expand/collapse/clear.
- Request timing and p95 summaries.
- Instrumentation toggle confirmation and status.
- Loading, offline, empty, stale, and error states.
- Keyboard and screen-reader operation for every debug surface.

### End-to-end acceptance

For a Laravel fixture project:

1. Link and serve the project through the Podman runtime.
2. Open the project in the web UI.
3. Compare `.env.example` and `.env` and verify exact counts.
4. Edit a non-secret env value, preview the diff, apply, and revert.
5. Enable request logging.
6. Make a request and verify its status/duration appears.
7. Enable Eloquent query capture.
8. Make a database-backed request and verify query count/time appears with redacted bindings.
9. Trigger `dump()` and verify it appears without breaking the request.
10. View logs and follow new entries.
11. Disable every instrumentation feature and confirm normal behavior returns.

The application must remain usable when the collector is stopped, the UI is closed, or capture is disabled.

## Success criteria

The developer observability service succeeds when:

- A developer can find relevant logs without leaving the project page.
- `.env` drift is explained with exact missing/extra/empty counts.
- Environment edits are safe, reviewable, atomic, and reversible.
- Request performance can be inspected over a bounded time window.
- Laravel query capture shows useful SQL shape, timing, and context without leaking bindings by default.
- Dump output can be inspected without relying on terminal output.
- Instrumentation is visibly enabled, bounded, reversible, and low-risk.
- The collector never blocks or crashes a project request.
- PHP 7.4 and newer supported runtimes behave consistently.
- CLI and web UI expose the same toggles and evidence.

## Immediate next implementation slice

Start with the lowest-risk vertical slice:

1. Implement a safe `.env` parser and comparison engine.
2. Add read-only project `.env` comparison to the project page.
3. Show exact missing/extra/empty/duplicate/malformed counts.
4. Add unit tests for parser/diff/redaction behavior.
5. Add project-scoped log snapshot viewing using existing logs.
6. Do not enable query capture or dump interception until the collector transport and panel security foundation are implemented.
