# Chauffeur Web UI Goals

This is the authoritative checklist for expanding the Chauffeur Web UI. Read this file before changing the frontend. The reference project is:

`/home/siegg/Workspaces/Personal/Projects/lerd`

The goal is to adapt the reference product shape to Chauffeur, not to copy Lerd branding or implement unrelated backend behavior.

## Phase Rule: Static UI/UX First

The current phase is strictly frontend skeleton work.

- Use centralized fixture data for features without an existing Chauffeur API.
- Do not add Go endpoints, database schema, service orchestration, WebSockets, SSE, or real mutations in this phase.
- Do not create fake API calls that imply unsupported behavior.
- Existing working container and backup operations may remain connected.
- Mark unsupported features as `Planned`, `Next`, or `Requires backend`.
- Use disabled controls, preview dialogs, static logs, and sample operation results to demonstrate future UX.
- Never present planned functionality as completed or live.
- Check this file before each implementation task and update only items verified in the UI.

## Current Baseline

These features already exist and should not be rebuilt unless the task explicitly improves them:

- [x] Shared application navbar with title, breadcrumbs, and route shortcuts.
- [x] Workspace overview dashboard.
- [x] Dashboard signal cards, projects, service pulse, activity, and quick actions.
- [x] Static capability map and operations board on the dashboard.
- [x] Database container list.
- [x] Container detail view.
- [x] Container start and stop operations.
- [x] Container log streaming UI.
- [x] Backup list and backup operations UI.
- [x] Documentation route.
- [x] Responsive sidebar and mobile shortcut behavior.

## Target Navigation

The target product should expose four primary areas, inspired by Lerd:

| Area      | Chauffeur meaning                              | Initial route goal                 |
| --------- | ---------------------------------------------- | ---------------------------------- |
| Dashboard | Workspace command center                       | `/`                                |
| Projects  | Linked local PHP projects                      | `/projects` and `/projects/:name`  |
| Services  | Databases and future local services            | `/services` and `/services/:name`  |
| System    | Chauffeur, Nginx, PHP, DNS, tools, diagnostics | `/system` and `/system/:component` |

Keep current routes working:

- `/containers`
- `/containers/:name`
- `/containers/backup`
- `/docs`

The sidebar may group these routes as Workspace, Projects, Services, System, and Resources. Do not hide the existing container and backup flows while introducing the broader structure.

## Dashboard Checklist

- [x] Health hero showing healthy, attention, failure, and update states.
- [x] Workspace summary for projects, services, databases, backups, and open issues.
- [x] Projects widget with framework, runtime, domain, HTTPS, and status.
- [x] Services widget with installed, active, stopped, and update states.
- [x] Workers widget with active, sleeping, and failing states.
- [x] System health widget for Nginx, PHP, DNS, and Chauffeur runtime.
- [x] Resource widget for CPU, memory, disk, and reclaimable storage.
- [x] Recent activity widget with operation type, subject, status, and time.
- [x] Onboarding panel: link project, choose runtime, provision database, verify domain.
- [x] Capability roadmap with Available, Next, and Planned labels.
- [x] Operations board that links live actions and future plans.
- [x] Command palette preview with pages, projects, services, toggles, and actions.

## Projects Checklist

### Project List

- [x] Project list with loading, empty, error, healthy, paused, and attention states.
- [x] Framework grouping.
- [x] Workspace grouping.
- [x] Project search and sorting controls.
- [x] Project status, local path, domain, PHP version, Node version, and HTTPS state.
- [x] Project linking preview flow.
- [x] Project unlink confirmation flow.
- [x] Workspace creation, rename, collapse, reorder, and delete previews.
- [x] Project pin/unpin and pause/resume previews.
- [x] Worktree summary on project rows.

### Project Detail Header

- [x] Project name, framework, runtime, domain, HTTPS, and status.
- [x] Open project in browser preview.
- [x] Open folder and terminal preview.
- [x] Manage domains preview.
- [x] Nginx configuration preview.
- [x] Restart runtime and development server previews.
- [x] LAN share, tunnel, and public share previews.
- [x] Xdebug toggle and mode selector preview.
- [x] Project doctor/diagnostic entry point.
- [x] Overflow menu for pause, resume, pin, restart, domain, sharing, and unlink.

### Project Detail Tabs

- [x] Overview tab.
- [x] Logs tab.
- [x] Environment tab.
- [x] Tinker tab.
- [x] Diagnostics tab.
- [x] Worktrees tab or worktree selector.

### Project Overview

- [x] Runtime controls for PHP and Node versions.
- [x] Service cards connected to the project.
- [x] Database connection summary.
- [x] Worker toggles for queue, Horizon, scheduler, Reverb, Stripe, Vite, and custom workers.
- [x] Worker failure, sleeping, stopped, and running states.
- [x] Request timing summary.

### Project Logs

- [x] Static log viewer with follow-to-bottom behavior.
- [x] Application logs.
- [x] PHP-FPM or custom container logs.
- [x] Development server logs.
- [x] Queue, Horizon, scheduler, Reverb, Stripe, and framework worker logs.
- [x] Error and warning highlighting.
- [x] Empty, loading, disconnected, and failed log states.
- [x] Copy logs action.

### Project Environment

- [x] Environment variable list with masked secrets.
- [x] Source and inherited-value indicators.
- [x] Edit environment preview.
- [x] Save confirmation preview.
- [x] Restore confirmation preview.
- [x] Duplicate-variable warning.
- [x] Worktree-specific environment state.

### Tinker and Diagnostics

- [x] Tinker editor placeholder with code, run, output, SQL, dump, and error regions.
- [x] Saved snippet list and save/load/delete previews.
- [x] Draft persistence preview.
- [x] Fullscreen and split-direction controls.
- [x] Diagnostics lenses for dumps, queries, jobs, views, mail, cache, events, and HTTP.
- [x] Count badges, filters, sample records, empty states, and planned status.

### Worktrees

- [x] Main branch and worktree selector.
- [x] Add worktree preview.
- [x] Remove worktree confirmation preview.
- [x] Worktree-specific domain and runtime.
- [x] Worktree-specific database isolation toggle.
- [x] Worktree database drop confirmation.
- [x] Worktree-specific workers, logs, environment, and sharing.

## Services Checklist

### Services Overview

- [x] Installed services dashboard.
- [x] Discoverable service preset catalog.
- [x] Service categories: databases, cache, queues, search, mail, storage, and workers.
- [x] Service status, version, project count, ports, dependencies, and update state.
- [x] Port conflict and dependency warnings.
- [x] Add/install service preview.
- [x] Empty and all-installed states.

### Service Detail

- [x] Service header with status, version, category, and project usage.
- [x] Start, stop, restart, pin, update, upgrade, migrate, rollback, reinstall, and remove previews.
- [x] Admin dashboard and connection URL previews.
- [x] Databases tab.
- [x] Entities tab.
- [x] Logs tab.
- [x] Environment tab.
- [x] Configuration/tuning tab.
- [x] Tools/client shims tab.
- [x] Ports tab.
- [x] Dependency visualization preview.

### Databases and Entities

- [x] Database cards with name, size, owner project, and status.
- [x] Create database preview.
- [x] Drop database confirmation preview.
- [x] Copy DSN action.
- [x] Export database preview.
- [x] Import database preview with progress, warning, error, and issue states.
- [x] Snapshot list and restore preview.
- [x] Testing database pairing.
- [x] Entity list grouped by entity kind.
- [x] Create and delete entity previews.

## System Checklist

- [x] System component list with status indicators.
- [x] Chauffeur runtime detail.
- [x] DNS detail with status, logs, and configuration preview.
- [x] Nginx detail with logs and configuration editor preview.
- [x] PHP runtime list and version cards.
- [x] PHP-FPM start/stop, default version, update, rebuild, remove, and Xdebug previews.
- [x] PHP tabs for logs, INI/configuration, ports, and extensions.
- [x] Node.js/Bun runtime list.
- [x] Node version install, remove, default, manager, and system-managed states.
- [x] Tool availability and update check page.
- [x] Watcher status and start preview.
- [x] Worker execution mode preview: host versus container.
- [x] Notification settings preview.
- [x] Debug bridge enable/disable and lens preview.
- [x] Autostart and idle-suspension settings preview.
- [x] LAN exposure and remote dashboard setup preview.
- [x] Version, update, changelog, and terminal-update preview.

## Global UX Checklist

- [x] Desktop navigation rail/sidebar.
- [x] Mobile bottom navigation.
- [x] Mobile list-to-detail flow.
- [x] Dynamic dashboard/app launcher preview.
- [x] Shared detail header and tab pattern.
- [x] Command palette preview.
- [x] Toast and activity feedback.
- [x] Modal host and confirmation pattern.
- [x] Permission-aware local versus remote controls.
- [x] Consistent status labels and status colors.
- [x] Keyboard focus states and semantic landmarks.
- [x] Accessible labels for icon-only actions.
- [x] `aria-current`, `aria-expanded`, `aria-live`, and dialog semantics where needed.
- [x] Reduced-motion support.
- [x] Responsive behavior at desktop, tablet, and mobile widths.

## Fixture Data Requirements

Create one centralized fixture module for the static phase. It should provide:

- [x] Centralized fixture module used by the dashboard.

- Projects
- Worktrees
- Services
- Workers
- Databases
- Backups
- System components
- Activity events
- Operations
- Roadmap features

Fixtures must demonstrate:

- [x] Healthy
- [x] Running
- [x] Stopped
- [x] Paused
- [x] Attention
- [x] Failed
- [x] Loading
- [x] Empty
- [x] Update available
- [x] Available
- [x] Next
- [x] Planned
- [x] Requires backend

## Implementation Constraints

- Use React, TanStack Router, existing shadcn components, and the current design tokens.
- Preserve the shared `AppNavbar` pattern.
- Do not manually recreate shadcn primitives.
- Keep current container and backup functionality working.
- Keep route-specific content outside the shared navbar.
- Do not add backend code during the static phase.
- Do not add fake network requests.
- Prefer fixture-driven reusable components over route-local duplicated data.
- Use real links for implemented screens and disabled/planned affordances for future screens.
- Add tests for new interactive components when practical.

## Completion Rules

For each implementation pass:

1. Read this checklist.
2. Inspect the existing UI before adding a feature.
3. Mark a checklist item complete only when it is visibly implemented and usable in the UI.
4. Keep unsupported operations clearly marked.
5. Run typecheck and production build.
6. Check desktop and mobile layouts.
7. Report completed checklist items, unchanged items, and verification results.

The static UI/UX phase is complete when the navigation, dashboard, Projects, Services, System, operation states, fixture data, responsive behavior, and accessibility foundations are represented without requiring new backend functionality.
