# Web UI Overhaul Plan

This plan is Phase 4 of the [master overhaul roadmap](./2026-08-19-overhaul-roadmap.md). The UI should consume shared project/service/operation state rather than reimplement runtime logic.

## Goal

Turn the current embedded panel into a reliable operational web UI for Chauffeur. The UI should make it easy to understand project health, PHP/runtime state, services, domains, SSL, DNS, databases, logs, and configuration without requiring users to memorize CLI commands.

The UI is not a separate product. It is another client of the same state and action model used by the CLI.

## Why this change

The current panel proves that Chauffeur can serve an embedded React application and manage Podman database containers, but it is not yet a trustworthy Chauffeur control surface:

- The dashboard displays hard-coded zero/unknown values.
- Navigation advertises unfinished pages with `#` links.
- The backend exposes database-container operations but not projects, PHP, nginx, SSL, DNS, or service health.
- Container detail responses return database passwords.
- Backup requests run synchronously in an HTTP request.
- The SSE logs endpoint sends one snapshot and closes rather than following new logs.
- The current server uses workspace paths inconsistently with `CHAUFFEUR_HOME`.

The Podman runtime migration creates the right moment to define a stable API for all operational state instead of extending the existing database-only API piecemeal.

## Product decision

The panel should become a **full Chauffeur control plane**, but it must be delivered progressively. The first complete slice should be project and runtime visibility, followed by safe actions and database workflows.

The UI must never present a control that has no working backend behavior. If a capability is not implemented, it should be absent or visibly marked as unavailable, not represented by a placeholder route.

## Non-goals

- Do not duplicate business logic in React.
- Do not allow the UI to mutate host/system settings silently.
- Do not expose passwords in normal API responses.
- Do not build a generic remote administration panel in the first version.
- Do not add a new frontend state store before the API/state contract is stable.
- Do not implement every future feature such as tunnels, workers, MCP, monitoring, or worktrees in this plan’s first release.

## Current baseline

### Backend

The current Go server is in `internal/panel/server.go` and provides:

```text
GET    /api/health
GET    /api/containers
GET    /api/containers/{name}
POST   /api/containers/{name}/start
POST   /api/containers/{name}/stop
GET    /api/containers/{name}/logs
GET    /api/containers/{name}/databases
GET    /api/backups
POST   /api/backups
DELETE /api/backups/{name}
POST   /api/backups/{name}/restore
```

### Frontend

The current React/TanStack/Tailwind application has routes for:

- Dashboard.
- Database container list.
- Container detail.
- Container backup.
- Documentation.

The sidebar also advertises unfinished sites, config, DNS, SSL, logs, and settings destinations.

## Information architecture

The navigation should follow user goals rather than implementation packages:

```text
Dashboard
Projects
  Project list
  Project detail
Services
  Runtime / PHP
  Nginx
  Databases
System
  Health / Doctor
  DNS
  SSL certificates
  Configuration
Activity / Logs
Documentation
Settings
```

### Dashboard

The dashboard answers “is my local environment ready?” in one screen:

- Overall health summary.
- Projects requiring attention.
- Active project URLs.
- PHP runtime/image readiness.
- Nginx status.
- Podman/rootless status.
- DNS and SSL state.
- Running database services.
- Recent operations and failures.
- Primary next actions.

Every card must link to evidence or an action. A status without a detail path is not useful.

### Projects

Projects are the primary object because the developer’s goal is a working project, not an abstract container.

The list should show:

- Project name and path.
- Primary URL and aliases.
- Framework/type.
- PHP version and runtime mode.
- SSL state.
- Nginx/FPM health.
- Required services.
- Last diagnostic result.
- Start/stop/restart/open/log actions.

The detail page should group information into:

1. Overview and URLs.
2. Runtime and PHP version.
3. Domains, aliases, SSL, and DNS.
4. Services and databases.
5. Environment/configuration.
6. Logs and recent operations.
7. Diagnostics.

### Services

Services should distinguish:

- Web runtime: nginx.
- PHP runtimes/images.
- Project FPM pools.
- Database containers.

The UI must show whether a service is:

```text
not installed
installed but stopped
starting
healthy
unhealthy
failed
degraded
```

### System

System pages should explain host/runtime integration without pretending the UI can silently fix privileged changes:

- Podman availability and rootless mode.
- Network status.
- DNS resolver mode and `.test` resolution.
- SSL CA/certificate state.
- Port availability/forwarding.
- Autostart configuration.
- Workspace location and disk usage.

## API and state contract

### Principles

- The API exposes facts and operations, not frontend-specific shapes.
- CLI and UI actions call the same internal service/runtime functions.
- Mutations return structured operation results.
- Every resource has a stable identity and state.
- Errors include code, message, impact, and suggested action.
- Secrets are omitted by default.
- Long-running work returns an operation/job identifier.

### Resource model

The first API contract should cover:

```text
Workspace
Project
Domain
Runtime
Service
DatabaseContainer
Backup
DiagnosticReport
Operation
LogStream
```

Example project response:

```json
{
  "slug": "my-app",
  "path": "/home/user/Projects/my-app",
  "framework": "laravel",
  "domains": ["my-app.test"],
  "url": "https://my-app.test:8443",
  "php": {
    "version": "8.3",
    "runtime": "podman",
    "fpm": "shared",
    "ready": true
  },
  "nginx": { "status": "healthy" },
  "ssl": { "enabled": true, "trusted": true },
  "health": { "status": "healthy", "checkedAt": "2026-08-19T00:00:00Z" }
}
```

### Endpoint groups

The exact URL naming can be finalized during implementation, but the contract should cover:

```text
GET  /api/health
GET  /api/dashboard

GET  /api/projects
GET  /api/projects/{slug}
POST /api/projects/link
POST /api/projects/{slug}/unlink
POST /api/projects/{slug}/start
POST /api/projects/{slug}/stop
POST /api/projects/{slug}/restart
POST /api/projects/{slug}/secure
POST /api/projects/{slug}/unsecure
GET  /api/projects/{slug}/diagnostics
GET  /api/projects/{slug}/logs

GET  /api/runtimes/php
POST /api/runtimes/php/install
POST /api/runtimes/php/{version}/rebuild
POST /api/runtimes/php/{version}/use
GET  /api/runtimes/php/{version}/config
PUT  /api/runtimes/php/{version}/config

GET  /api/services
POST /api/services/{name}/start
POST /api/services/{name}/stop
POST /api/services/{name}/restart

GET  /api/databases
POST /api/databases
GET  /api/databases/{name}
POST /api/databases/{name}/start
POST /api/databases/{name}/stop
DELETE /api/databases/{name}
GET  /api/databases/{name}/logs
GET  /api/databases/{name}/databases

GET  /api/backups
POST /api/backups
GET  /api/backups/{name}
DELETE /api/backups/{name}
POST /api/backups/{name}/restore

GET  /api/operations/{id}
GET  /api/operations/{id}/events
GET  /api/events
```

Do not implement the entire endpoint list before shipping. Define the contract, then deliver endpoint groups alongside their UI slices.

### Operation responses

Start, stop, install, rebuild, backup, restore, and migration actions may take longer than an HTTP request. They should return:

```json
{
  "operationId": "op_01J...",
  "status": "queued",
  "resource": "project:my-app",
  "message": "Starting PHP 8.3 runtime"
}
```

The UI then observes operation progress and renders:

- Current step.
- Completed steps.
- Logs/evidence.
- Failure reason.
- Suggested retry or rollback.

This is required for PHP image builds and database backup/restore. Do not block the browser request for the whole operation.

## UI state and interaction rules

### Loading

- Use shape-matching skeletons for lists/cards.
- Preserve the previous successful state during background refresh.
- Do not replace the entire page with a spinner for a local action.

### Empty states

Every empty state needs a useful next action:

- No projects: link a project.
- No PHP versions: install a PHP version.
- No databases: create a database.
- No backups: create a backup.
- No diagnostics: run a check.

### Errors

Errors must state:

```text
What failed
What it affects
What the user can do next
```

Example:

```text
PHP 7.4 image is unavailable.
The legacy-app project cannot start with PHP 7.4.
Build locally, pull the compatible image, or switch the project to PHP 8.3.
```

### Destructive actions

Require confirmation for:

- Unlinking a project.
- Removing a database/container.
- Deleting a backup.
- Restoring over an existing database.
- Resetting configuration.
- Removing PHP versions or extensions.

The confirmation must identify the exact target and consequence. For restore, show backup timestamp, size, database, and whether a pre-restore snapshot will be made.

### Optimistic updates

Do not optimistically mark services as running before the backend confirms readiness. Show `starting`/`stopping` operation states and invalidate resource data when the operation completes.

## Configuration editing

The UI should not initially expose arbitrary YAML editing. Start with typed forms for safe, validated settings:

- Default PHP version.
- PHP per-version limits.
- Project PHP version/FPM strategy.
- Domain and aliases.
- SSL enablement.
- Nginx upload limit.
- Runtime image/extension declarations after the Podman migration.

For advanced configuration, provide a reviewed text editor only when the backend supports:

- Syntax/schema validation.
- Atomic writes.
- Timestamped backups.
- Diff preview.
- Reset/revert.
- Permission checks.

Never let a browser form write arbitrary host paths or execute arbitrary shell commands.

## Logs and diagnostics

### Logs

Log views should support:

- Project/application logs.
- Nginx access/error logs.
- PHP-FPM logs.
- Container/service logs.
- Operation/build logs.
- Search/filter.
- Line limits.
- Pause/resume follow.
- Copy/download.
- Clear distinction between historical snapshot and live stream.

The first implementation can use polling if it is honest. True SSE/WebSocket streaming should include:

- Reconnect behavior.
- Event IDs or cursor support.
- Backpressure/line limits.
- Authorization.
- Clean cancellation.

### Diagnostics

Display diagnostics as structured checks:

```text
check name
severity: pass | info | warning | error
evidence
impact
next action
```

The UI should expose both:

- Environment doctor: Podman, ports, DNS, SSL, workspace, dependencies.
- Project doctor: project path, framework, PHP compatibility, nginx route, FPM readiness, env/database/service checks.

## Security model

The web UI controls local development processes and can access credentials, files, and databases. Treat it as a privileged local application even when it binds to loopback.

### Required baseline

- Bind to loopback by default.
- Require an explicit opt-in for remote binding.
- Use a generated local bearer token or equivalent authentication.
- Restrict CORS to the panel origin; do not use wildcard CORS.
- Validate `Origin`/Host where browser control operations are exposed.
- Never return database passwords in normal resource responses.
- Provide explicit reveal/copy actions only if required, with audit/log redaction.
- Validate project slugs, container names, backup names, image names, ports, and paths.
- Add CSRF protection if cookie-based sessions are introduced.
- Redact secrets from operation logs and diagnostics.
- Require confirmation for destructive operations.

### Remote access

Remote access is not part of the first release. If added later, it requires:

- Explicit remote bind flag.
- Strong authentication.
- TLS or an authenticated tunnel.
- Permission model.
- Audit trail.
- Remote-control allowlist.

## Responsive and accessibility requirements

- All core workflows work on desktop and narrow mobile widths.
- Keyboard navigation reaches every action and modal control.
- Focus moves into dialogs and returns to the triggering control.
- Status is not conveyed by color alone.
- Buttons have action-specific accessible names.
- Tables have meaningful headers and responsive alternatives.
- Live operation/log updates use appropriate `aria-live` regions without overwhelming screen readers.
- Reduced-motion preferences are honored.
- Destructive actions cannot be triggered accidentally through keyboard focus.
- Touch targets are sufficiently large for mobile use.

## Design direction

The visual language should communicate a calm operations console rather than a generic admin template:

- Clear health hierarchy.
- Strong distinction between healthy, degraded, blocked, and destructive states.
- Project URLs as primary interactive objects.
- Compact service/runtime status with expandable evidence.
- Dense information where operational, generous spacing where decision-making.
- Consistent action placement across project, service, and database detail pages.
- No placeholder cards or decorative metrics without live data.

Use the existing React/TanStack/Tailwind stack unless implementation evidence shows it is blocking the product. Do not add a second frontend framework.

## CLI and UI convergence

The UI and CLI must share implementation functions for mutations:

```text
CLI command ─┐
             ├─ runtime/project/service operation ─ state + operation event
Web UI ------┘
```

The panel must not shell out to a user-facing CLI binary as its primary integration. Both surfaces should call shared Go packages so behavior, validation, and rollback rules remain identical.

The UI may invoke the same operation functions with a different renderer. The CLI remains fully functional without the panel.

## Phased implementation

### Phase 0: Contract and security foundation

- Decide panel scope and route names.
- Centralize workspace-root resolution.
- Define API resource/state/operation types.
- Remove password fields from container responses.
- Validate all path/name parameters.
- Bind loopback by default and add local authentication/token handling.
- Remove wildcard CORS.
- Add API error codes and structured messages.

### Phase 1: Data-driven shell

- Replace hard-coded dashboard values with `/api/dashboard`.
- Remove all placeholder navigation links.
- Build shared query/fetch/error/operation utilities.
- Add global loading, offline, stale, and retry states.
- Add responsive navigation and accessible focus behavior.
- Add frontend API contract types generated or checked against Go types.

### Phase 2: Project control plane

- Add project list and project detail pages.
- Expose project URL, PHP version, FPM mode, SSL, nginx, and health state.
- Add link/unlink/secure/unsecure/start/stop/restart actions.
- Add project diagnostics and project logs.
- Add open URL and copy URL actions.
- Add empty/error states with CLI fallback commands.

### Phase 3: Podman runtime management

- Add PHP runtime/image list and detail.
- Show installed, building, ready, stale, failed, and unavailable states.
- Stream or poll rebuild operation progress.
- Add PHP version selection for projects.
- Show image digest/recipe fingerprint and standard/custom extensions.
- Add container/network/runtime health evidence.

### Phase 4: Configuration and SSL/DNS

- Add typed workspace and project settings forms.
- Add PHP limits and nginx upload configuration.
- Add domains and aliases management.
- Add SSL certificate status and secure/unsecure actions.
- Add DNS/port diagnostic view with exact user-run remediation commands.
- Show configuration diff before applying complex changes.

### Phase 5: Database and backup operations

- Upgrade container list to full database management.
- Add create flow with engine, name, credentials, port, and volume choices.
- Never display passwords by default.
- Add database selection and connection information with explicit reveal/copy controls.
- Convert backup/restore to operation jobs.
- Add progress, cancellation where safe, failure details, and pre-restore safety checks.
- Add backup list/search/filter and retention information.

### Phase 6: Logs and activity

- Add unified log sources and filters.
- Implement honest snapshot mode first.
- Add real event streaming only after operation/event storage is stable.
- Add recent activity timeline from operation results.
- Add copy/download and bounded retention behavior.

### Phase 7: Documentation and polish

- Make docs searchable and link contextual commands from failures.
- Add keyboard shortcuts only where discoverable and accessible.
- Add responsive/mobile layouts.
- Add visual regression checks for primary states.
- Add onboarding walkthrough for first project.
- Remove obsolete database-only terminology where the panel is now full-runtime.

## Verification strategy

### Backend tests

- API response schemas and error codes.
- Authentication and origin checks.
- Secret omission/redaction.
- Path/name validation.
- Idempotent start/stop/restart operations.
- Operation/job lifecycle.
- Workspace-root override behavior.
- Project/service/runtime handlers with fake backends.
- Backup/restore safety and traversal rejection.

### Frontend tests

- Dashboard renders live API state.
- Loading, stale, offline, empty, error, and success states.
- Project filters and actions.
- Operation progress and retry.
- Destructive confirmation dialogs.
- Password omission/reveal behavior.
- Keyboard navigation and focus return.
- Responsive navigation.
- Log snapshot/follow/reconnect behavior.

### End-to-end tests

Use a test workspace and fake or real Podman backend as appropriate:

1. Start panel.
2. Load dashboard.
3. Link a fixture project.
4. Install/select a PHP runtime.
5. Start project services.
6. Verify the project URL and health state.
7. Change a safe setting.
8. Run a diagnostic.
9. View logs.
10. Create/list/restore a database backup.
11. Verify the UI reflects final state after a CLI mutation.

The final step is mandatory: CLI and UI must converge on the same state.

## Success criteria

The web UI overhaul succeeds when:

- The dashboard contains no hard-coded operational metrics.
- Every visible navigation destination works or is intentionally omitted.
- A user can understand and operate a linked project without switching to the CLI for ordinary tasks.
- PHP/runtime, nginx, SSL, DNS, databases, backups, logs, and diagnostics have clear states.
- Long-running actions provide progress and durable results.
- Secrets are not exposed by default.
- The panel is safe on loopback and explicitly protected before any remote exposure.
- The UI remains useful when Podman or a service is unavailable and explains the next action.
- CLI and UI mutations use the same validation and operation behavior.
- Primary workflows pass on desktop, mobile-width, keyboard, and screen-reader-oriented checks.

## Immediate next implementation slice

Start with the smallest useful vertical slice:

1. Add a safe `/api/dashboard` endpoint backed by real workspace, project, service, and Podman status.
2. Remove hard-coded dashboard values.
3. Replace placeholder sidebar links with only working routes.
4. Add project list data and a project detail read-only page.
5. Add API contract tests and frontend loading/error/empty tests.
6. Do not add configuration editing or destructive actions until authentication, validation, and operation responses exist.
