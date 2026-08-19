# Chauffeur Overhaul Roadmap

## Objective

Evolve Chauffeur from a native PHP/nginx process manager with optional database containers into a cohesive local PHP development platform with:

- Reproducible Podman PHP/nginx runtimes.
- Project-first setup and status.
- Managed databases and supporting services.
- A truthful, useful web UI.
- Safe environment/configuration editing.
- Developer observability tools.
- Reliable diagnostics and recovery.
- One operational model across CLI, TUI, and web UI.

The goal is not to reproduce every Lerd feature. The goal is to provide the smallest coherent platform that makes local PHP development predictable and easy to debug.

## Strategy

Copy these Lerd patterns:

1. One shared state/action model across interfaces.
2. Project-first UX.
3. Guided setup with noninteractive parity.
4. Environment doctor plus project doctor.
5. Declarative service presets.
6. Reversible destructive operations.
7. Operation/job progress for long work.
8. Build/runtime fingerprints.
9. Real application-outcome verification.
10. Progressive disclosure instead of an incomplete broad surface.

Do not copy these yet:

- Full service marketplace.
- Public tunnels and remote dashboard control.
- All-platform support.
- Full worker orchestration.
- MCP, tray, and large framework registries.
- Continuous profiling and broad monitoring.
- Multi-version service matrices before data migration is proven.

## Shared architecture decisions

### Resource model

Define these shared concepts before expanding interfaces:

```text
Workspace
Project
Domain
Runtime
PHP image/version
Service
Database
Operation
Diagnostic
Log stream
Observation event
```

Each resource needs:

- Stable identity.
- Human-readable state.
- Machine-readable state.
- Safe actions.
- Evidence/details.
- Error code and next action.

### Action model

CLI, TUI, and web UI must call the same Go operation functions:

```text
CLI command ─┐
TUI action ──┼→ shared operation → state/event/result
Web action ──┘
```

The UI must not implement its own service detection, project recommendations, or mutation logic. The CLI must remain usable without the panel.

### Configuration scopes

Keep scopes distinct:

```text
Workspace/global config  → machine preferences and runtime defaults
Project intent           → PHP, domains, FPM, logical services, Node intent
Runtime state             → containers, images, ports, digests, health
Secret state              → protected credentials, never project intent
Observation state         → bounded logs, requests, dumps, queries, profiles
```

The existing workspace-root abstraction must be used by every subsystem, including Podman and the panel.

### Security baseline

- Rootless Podman by default.
- Loopback-only panel by default.
- Authentication/token protection before exposing captured data or remote controls.
- No passwords in normal API responses.
- No wildcard CORS.
- No arbitrary host paths, privileged containers, or host networking in the first custom-service schema.
- Redaction and bounded retention for env, SQL, dumps, logs, and profiles.
- Explicit confirmation before data loss, service creation, or host/system mutation.

## Phase 0: Foundations and contract

### Purpose

Create the shared architecture and remove ambiguity before runtime/UI expansion.

### Work

- Define resource/state/action/operation types.
- Add a runtime abstraction around current native services.
- Add fakeable command/process/Podman backends.
- Centralize workspace root resolution and fix Podman/panel path inconsistencies.
- Reconcile help, docs, flags, and actual command behavior.
- Define project config migration/versioning.
- Define service preset schema.
- Define API error/operation response shape.
- Add panel loopback/auth/CORS/path-validation baseline.

### Must not do yet

- Do not remove the native runtime.
- Do not add custom service marketplace behavior.
- Do not expose arbitrary env/config editing.

### Gate

- Existing tests remain at the same or better level.
- Native `link`, `start`, `status`, `logs`, and `php` behavior is behind the runtime abstraction.
- Workspace override tests pass for every subsystem.
- API and project config schemas are documented and tested.

## Phase 1: Podman PHP runtime parity

Detailed plan: [Podman Runtime Overhaul](./2026-08-19-podman-runtime-overhaul.md).

### Purpose

Replace host-built PHP/FPM execution with reproducible rootless Podman images while preserving the existing product workflow.

### Work

- Build/pull PHP images.
- Encode PHP 7.4/8.0 compatibility and GD behavior.
- Support local builds as fallback.
- Execute `php`, Composer, and project commands through the selected runtime.
- Implement shared and dedicated FPM containers.
- Preserve project domains, SSL, logs, and service lifecycle.
- Move nginx after PHP/FPM parity is proven.

### Gate

- PHP 7.4, 8.0, 8.3, and newest supported version are usable.
- GD and standard extensions are verified inside the runtime.
- `chauf php`, Composer, and project requests work in shared/dedicated modes.
- A real fixture project returns a successful HTTP response through nginx and PHP-FPM.
- Native state remains recoverable.

## Phase 2: Project setup and link wizard

Detailed plan: [Link Wizard TUI](./2026-08-19-link-wizard-tui.md).

### Purpose

Turn linking into a guided, evidence-based project setup flow.

### Work

- Add read-only project detection and recommendation engine.
- Add compact TUI form with the reference visual language:
  - purple group labels;
  - `>` active cursor;
  - green selected items;
  - gray secondary/disabled items;
  - contextual lowercase footer controls.
- Add PHP/domain/SSL/FPM wizard first.
- Add Node intent and `.nvmrc` handling.
- Add database and service choices after service model exists.
- Add review-before-apply and cancellation without mutation.
- Preserve noninteractive flags for CI.

### Gate

- Existing `chauf link` flags still work.
- Interactive and noninteractive plans produce equivalent project config.
- Cancel/back never mutates project/nginx/SSL state.
- Defaults include evidence and do not trigger hidden builds or container creation.

## Phase 3: Generic managed services

Detailed plan: [Podman Service Management](./2026-08-19-podman-service-management.md).

### Purpose

Generalize existing database containers into a stable service layer.

### Work

- Add generic service identity and health/status types.
- Add `chauf service` commands while preserving `chauf podman` aliases.
- Promote Redis to a cache service.
- Add Mailpit, Meilisearch, and RustFS presets.
- Add stable `chauf-net` networking and logical container DNS.
- Add project service bindings and dependency graph.
- Add safe persistence/removal/recovery policy.
- Add host/container connection target distinction.

### Gate

- Existing database backup/restore still works.
- A Laravel fixture can use PostgreSQL, Redis, and Mailpit through container DNS.
- Service removal preserves data by default.
- Dependent projects are visible before destructive actions.
- Health checks and logs explain failed service startup.

## Phase 4: Web control plane

Detailed plan: [Web UI Overhaul](./2026-08-19-web-ui-overhaul.md).

### Purpose

Make project and service state visible and adjustable without requiring CLI knowledge.

### Work

- Replace hard-coded dashboard metrics with live API data.
- Remove placeholder navigation.
- Add project list/detail pages.
- Add runtime, service, database, SSL, DNS, and diagnostics views.
- Add safe actions backed by shared operations.
- Add operation progress and durable results.
- Add responsive/accessibility behavior.

### Gate

- CLI and UI show the same final state after either performs a mutation.
- Every visible action works or is omitted.
- No secrets are exposed.
- Project URL, PHP/runtime, FPM, nginx, SSL, service, and health state are visible.
- Web UI remains useful when Podman or a service is unavailable.

## Phase 5: Environment intelligence and safe configuration

This phase starts the lower-risk portion of [Developer Observability Tools](./2026-08-19-developer-observability-tools.md).

### Work

- Implement safe `.env` parser.
- Compare `.env.example` and `.env`.
- Report missing, extra, empty, duplicate, and malformed keys.
- Add read-only comparison UI and CLI.
- Add reviewed editor with diff, atomic write, backup, and revert.
- Add generated service env diff without copying secrets automatically.
- Add project doctor checks.

### Gate

- Parser preserves comments/order where supported.
- Secret values are redacted in UI/logs/errors.
- Editing is atomic and reversible.
- Exact difference counts are tested against fixtures.
- Project doctor gives evidence and next actions.

## Phase 6: Logs and request observability

### Purpose

Reduce time spent locating the evidence for a failure or slow request.

### Work

- Add bounded observation collector and local transport.
- Normalize nginx, PHP-FPM, container, app, and operation logs.
- Add project log viewer.
- Add request IDs, method/path/status/duration, and bounded request history.
- Add median/p95, error rate, slow requests, and status distribution.
- Add honest snapshot first; add reconnectable follow streams afterward.

### Gate

- Collector failure never blocks a project request.
- Memory and retention are bounded.
- Logs and request data redact sensitive headers/values.
- UI distinguishes snapshot, following, paused, reconnecting, and disconnected.
- A real request appears with correct project/status/duration.

## Phase 7: Laravel debug bridge

### Purpose

Provide the Herd Pro/Lerd-style high-value debugging experience without making it mandatory for generic PHP.

### Work

- Add Laravel bridge/provider.
- Add dump/`dd()` interception with bounded serialization and redaction.
- Add Eloquent query capture with redacted bindings.
- Add slow query and repeated query evidence.
- Add cautious N+1 warnings.
- Add UI/CLI toggles, tail, clear, and status.

### Gate

- Capture is disabled by default.
- Laravel requests remain functional with collector unavailable.
- SQL bindings and dumps are redacted by default.
- Query count/time and dump context are correct in end-to-end fixtures.
- Clear/disable removes buffered sensitive data.

## Phase 8: Optional profiling and framework expansion

### Work

- Xdebug toggle/status/on-demand mode.
- One-shot profiling and bounded report storage.
- Performance profile cleanup.
- Symfony/Doctrine and WordPress adapters where demand supports them.
- Framework-specific doctor checks.

### Gate

- Instrumentation overhead is visible.
- Profiling is opt-in and reversible.
- Reports do not expose data outside the local control boundary.
- Framework adapters do not leak into the generic runtime core.

## Release gates across all phases

### Contract gate

Help, CLI flags, docs, web UI labels, API responses, and project configuration agree with implemented behavior.

### Safety gate

No secret exposure, path traversal, silent host mutation, unsafe service privilege, or irreversible data deletion remains in the phase scope.

### Runtime gate

The target operation works in a clean test workspace with the intended runtime and dependencies.

### Outcome gate

A real linked fixture project answers a real HTTP request through nginx and the selected PHP runtime.

### Interface gate

CLI, TUI, and web UI use the same action/state behavior where the feature is exposed.

### Recovery gate

Failed or destructive operations have a clear retry, rollback, backup, or preserved-data path.

## Recommended implementation order

```text
shared contracts
  → Podman PHP runtime parity
  → link wizard PHP/domain flow
  → generic services
  → full web control plane
  → .env intelligence/project doctor
  → logs/request tracking
  → Laravel dumps/queries
  → profiling/framework adapters
```

The web UI can begin with read-only dashboard/project state earlier, but write actions should follow the shared operation model and security foundation.

## Immediate next slice

The first implementation unit should be Phase 0’s smallest vertical slice:

1. Define a `ProjectStatus`/`ServiceStatus`/`OperationResult` model.
2. Put current native service status behind a runtime adapter.
3. Add a fake-backed `GET /api/dashboard`.
4. Render real project/service status in the dashboard.
5. Add one contract test proving a CLI mutation and API read see the same state.

This creates the seam required by every later plan without prematurely starting the full Podman migration or debug bridge.
