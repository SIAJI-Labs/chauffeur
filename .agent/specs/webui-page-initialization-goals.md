# Chauffeur Web UI Page Initialization Goals

This specification supersedes the preview-route approach for the planned sidebar
destinations. The goal is to initialize real product pages and page sections now,
without pretending that unsupported runtime functionality exists.

Read this file together with:

- `.agent/specs/webui-goals.md`
- `.agent/specs/webui-ux-audit-goals.md`

## Product Rule

Every sidebar destination must open a dedicated real page designed for its final
purpose. Do not route a sidebar destination to a generic preview page, a hash-only
section, an unrelated list page, or the documentation route.

Detail tabs are allowed inside a project or service detail page, but they are not
valid sidebar destinations. For example, the Service logs sidebar item must open
`/services/logs`, not `/services/PostgreSQL#logs`. The dedicated page may contain a
service selector and link into a selected service's detail page.

The initialized page may be static and non-functional. It must still have:

- A meaningful page title and breadcrumb.
- Final information architecture and visual hierarchy.
- Centralized fixture-driven sample, empty, loading, error, disconnected, and
  unavailable states where appropriate.
- Real controls in their intended locations.
- Disabled or non-submitting controls when backend support is unavailable.
- Clear `Planned`, `Next`, or `Requires backend` labels beside unsupported work.
- No fake network requests, fake mutations, or simulated success.

## Route Targets

Use real routes for page destinations. Existing working routes and detail tabs must
remain the source of truth where they already provide the intended page.

### Workspace

- `/`: Workspace overview.
- `/activity`: Recent activity page with filters, event groups, status states,
  and a disconnected live-feed message.
- Global command palette: a dialog available from every route with `Ctrl/Cmd+K`.
  It is a real navigation/search surface, not a page preview.

### Projects

- `/projects`: Project inventory and link-project entry point.
- `/projects/link`: Link-project setup page with folder, workspace, runtime, and
  route sections. The submit action is disabled and marked `Requires backend`.
- `/projects/overview`: Project overview page with a project selector and
  workspace-level overview cards.
- `/projects/logs`: Project logs page with project selector, source filters, and
  static application/runtime/worker log states.
- `/projects/environment`: Project environment page with project selector,
  masked variables, source indicators, and unavailable edit controls.
- `/projects/diagnostics`: Project diagnostics page with project selector,
  diagnostic lenses, issue states, and unavailable run controls.
- `/projects/worktrees`: Project worktrees page with project selector, branch
  inventory, and unavailable add/remove/isolation controls.
- `/projects/:name`: Existing project detail page.
- Project detail sections must be stable tabs for Overview, Logs, Environment,
  Tinker, Diagnostics, and Worktrees.

### Services

- `/services`: Installed service inventory.
- `/services/details`: Service details page with a service selector and service
  health/header information.
- `/services/catalog`: Service catalog page with categories, search, installed,
  available, planned, and dependency-warning states.
- `/services/databases`: Databases and entities page with service selector,
  database cards, entity groups, and unavailable lifecycle controls.
- `/services/data-lifecycle`: Import, export, and snapshots page with service and
  database selectors, progress/warning/error states, and no submit path.
- `/services/entities`: Entity actions page with entity groups and final-position
  create/delete controls marked `Requires backend`.
- `/services/logs`: Service logs page with service selector, source filters, and
  loading/empty/disconnected/error states.
- `/services/configuration`: Service configuration page with tuning sections and
  unavailable save/reset controls.
- `/services/dependencies`: Ports and dependencies page with port inventory,
  conflict states, and static dependency visualization.
- `/services/:name`: Existing service detail page.
- Service detail tabs must provide real sections for Admin, Databases, Entities,
  Logs, Environment, Configuration, Tools, and Ports.
- Database/entity lifecycle controls must be located in their final sections and
  remain disabled or non-submitting until backend support exists.

### System

- `/system`: System health overview.
- `/system/runtime`: Chauffeur runtime page.
- `/system/network`: DNS and Nginx page.
- `/system/php`: Existing PHP runtime page.
- `/system/node`: Node.js and Bun page.
- `/system/tools`: Tools and watchers page.
- `/system/debug`: Debug bridge page.
- `/settings`: Settings and updates page, including the initial font-size setting.

### Resources

- `/containers`: Existing database container inventory.
- `/containers/backup`: Existing backup operations.
- `/resources/workers`: Worker inventory and execution-state page.
- `/resources/usage`: Resource usage page.
- `/resources/telemetry`: CPU and memory telemetry page.
- `/resources/issues`: Issues and diagnostics page.

## Page Surface Requirements

### Recent Activity

- Full-width activity timeline using existing activity fixtures.
- Group by time or operation category.
- Show healthy, warning, failed, and disconnected states.
- Provide filter controls with truthful fixture-driven results.
- Live updates are not connected; state this without showing a fake stream.

### Sidebar Destination Contract

The sidebar route model must expose a dedicated destination for every item below:

| Sidebar item | Dedicated destination |
| --- | --- |
| Recent activity | `/activity` |
| Command palette | Global dialog opened from any route |
| Link a project | `/projects/link` |
| Project overview | `/projects/overview` |
| Project logs | `/projects/logs` |
| Project environment | `/projects/environment` |
| Project diagnostics | `/projects/diagnostics` |
| Project worktrees | `/projects/worktrees` |
| Service details | `/services/details` |
| Service catalog | `/services/catalog` |
| Databases and entities | `/services/databases` |
| Import, export, snapshots | `/services/data-lifecycle` |
| Entity actions | `/services/entities` |
| Service logs | `/services/logs` |
| Configuration | `/services/configuration` |
| Ports and dependencies | `/services/dependencies` |
| Chauffeur runtime | `/system/runtime` |
| DNS and Nginx | `/system/network` |
| Node.js and Bun | `/system/node` |
| Tools and watchers | `/system/tools` |
| Debug bridge | `/system/debug` |
| Settings and updates | `/settings` |
| Workers | `/resources/workers` |
| Resource usage | `/resources/usage` |
| CPU and memory | `/resources/telemetry` |
| Issues and diagnostics | `/resources/issues` |

The sidebar must not contain `#logs`, `#environment`, `#configuration`, or any
other hash-only destination. A route may use hash/query state only after the user
has entered a dedicated page or a selected project/service detail route.

### Command Palette

- Global dialog with search input, keyboard focus, result groups, shortcuts, and
  empty state.
- Navigate to implemented pages using real router links.
- Show unsupported actions as disabled rows with a `Requires backend` label.
- Support Escape to close, arrow/tab navigation, and visible focus.

### Link Project

- Multi-section setup layout: local folder, workspace identity, runtime profile,
  domain/HTTPS review, and final summary.
- Use static example values only as clearly labeled sample data.
- No folder picker, filesystem mutation, route creation, or submit request.
- Include a disabled primary action with an explanation.

### Project Detail

- Preserve the existing working project detail route and all current fixture data.
- Sidebar destinations for project sections must open the corresponding dedicated
  page. The detail route may expose the same content as internal tabs.
- Keep project actions as final-position controls with preview/unavailable
  messaging; never make planned actions look successful.
- Preserve not-found behavior for unknown project names.

### Service Catalog and Detail

- Catalog must be a first-class page, not an embedded preview card.
- Detail sections must use shared tab and card patterns.
- Keep connection strings and credentials masked; never add clear-text secrets.
- Database import/export/snapshot, entity, configuration, and dependency actions
  must have their final dialogs or panels, but no live submit path.
- Preserve not-found behavior for unknown service names.

### System Pages

- Each system destination gets a dedicated page with a component-specific header,
  status summary, details, logs/configuration sections where relevant, and an
  explicit unavailable state for management controls.
- Runtime, DNS, Nginx, Node/Bun, tools/watchers, and debug bridge pages must use
  centralized fixtures rather than duplicated route-local data.
- `/settings` must include font-size options `xs`, `sm`, `m` (default), `lg`, and
  `xl`. The setting may persist locally in the browser, but must not call an API.
- Other settings and update actions remain visibly planned.

### Resource Pages

- Workers page must show active, sleeping, failed, and stopped fixtures.
- Usage page must show CPU, memory, disk, and reclaimable storage summaries.
- Telemetry page must show CPU/memory cards and a static chart/table treatment.
- Issues page must show warning, failure, resolved, and empty states.
- Worker controls and diagnostic runs remain disabled or marked `Requires backend`.

## Shared Implementation

- Keep sidebar destinations in a typed shared navigation model.
- Use shared page shell, header, status badge, card, empty state, tab, dialog, and
  unavailable-action components.
- Use TanStack Router links for implemented pages.
- Use stable hash/query state only for tabs inside a selected project/service
  detail page. Never use it as the sidebar destination itself.
- Add `aria-current` for active navigation and tabs.
- Add correct `aria-expanded`, `aria-controls`, `aria-selected`, `aria-labelledby`,
  and `role="tabpanel"` relationships.
- Keep all controls keyboard reachable with visible focus states.
- Use `aria-live` only for real local state changes such as font-size selection or
  static status announcements.
- Keep mobile layouts usable: no clipped labels, tables, dialog footers, badges,
  tab lists, or log viewers.
- Respect reduced-motion preferences and existing design tokens.

## Fixture Requirements

Extend `internal/panel-apps/src/data/webui-fixtures.ts` or a shared fixture module
with data for:

- Recent activity events.
- Link-project setup steps and sample review state.
- Service catalog categories and catalog rows.
- Runtime, network, tool, watcher, and debug bridge states.
- Workers, resource usage, telemetry, and issue states.
- Font-size options and local preference metadata.

Fixtures must cover healthy, loading, empty, warning, failed, disconnected,
available, planned, next, and requires-backend states without implying live data.

## Explicit Non-Goals

- Do not keep or add a generic `/preview/:slug` destination for these sidebar items.
- Do not add Go endpoints, schemas, orchestration, WebSockets, SSE, or fake API
  calls.
- Do not implement filesystem selection, project linking, service installation,
  database import/export, entity mutation, runtime management, worker control,
  telemetry collection, or diagnostic execution.
- Do not replace existing working project, service, container, backup, log-stream,
  or documentation routes with placeholders.

## Execution Passes

### Pass 1: Navigation, Workspace, and Projects

- Add the shared destination model.
- Add `/activity`, `/projects/link`, `/projects/overview`, `/projects/logs`,
  `/projects/environment`, `/projects/diagnostics`, and `/projects/worktrees`.
- Remove generic preview destinations from the sidebar.
- Verify every Project sidebar item opens its dedicated page. Verify project detail
  tabs separately and verify invalid project routes.

### Pass 2: Services

- Add `/services/catalog`.
- Add `/services/details`, `/services/databases`, `/services/data-lifecycle`,
  `/services/entities`, `/services/logs`, `/services/configuration`, and
  `/services/dependencies`.
- Complete service detail tabs and final-position static dialogs/panels.
- Verify Service logs opens `/services/logs`, not a `#logs` section.
- Verify invalid service routes and masked connection data.

### Pass 3: System and Settings

- Add dedicated runtime, network, Node/Bun, tools, and debug pages.
- Add `/settings` and functional local font-size selection.
- Keep management and update operations unavailable.

### Pass 4: Resources

- Add workers, usage, telemetry, and issues pages.
- Add fixture-driven state variations and unavailable controls.

### Pass 5: Accessibility and Responsive Polish

- Test sidebar expansion, route links, tabs, dialogs, command palette, settings
  radio options, copy/reveal controls, and disabled actions.
- Check desktop, tablet, and mobile layouts.
- Verify no generic preview pages, hash-only sidebar destinations, or placeholder
  links remain.

## Verification After Every Pass

Run from `internal/panel-apps`:

```sh
npm run typecheck
npx eslint <changed-files>
npm test
npm run build
```

Run from the repository root:

```sh
git diff --check
```

Visit every current and newly added route at desktop, tablet, and mobile widths.
Verify keyboard behavior, sidebar expansion, tabs, dialogs, command palette,
font-size selection, copy/reveal controls, and unavailable actions. Report any
browser-visual limitation instead of assuming the layout is correct.

At the end of each pass, report:

- Completed pages and interactions.
- Audit/spec checks visibly verified.
- Unresolved issues and backend-dependent work.
- The exact next pass.
