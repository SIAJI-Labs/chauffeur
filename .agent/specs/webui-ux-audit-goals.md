# Chauffeur Web UI/UX Audit Goals

This document is the execution plan from the full route audit performed on 2026-08-21.
It supplements `.agent/specs/webui-goals.md`; the existing static UI/UX constraints still
apply and remain authoritative.

## Audit Scope

Reviewed route reachability and source structure for:

- `/`
- `/projects`
- `/projects/commerce-api`
- `/services`
- `/services/PostgreSQL`
- `/system`
- `/system/php`
- `/containers`
- `/containers/postgres`
- `/containers/backup`
- `/docs`

Shared surfaces reviewed:

- `AppNavbar`
- `AppSidebar`
- `NavMain`
- dialog, alert-dialog, tooltip, select, table, and button primitives
- shared stylesheet and responsive media queries

All listed routes returned HTTP 200 from the local development server. Browser-level
visual inspection was not available because no OpenChamber browser client was connected;
the execution pass must re-check rendered desktop, tablet, mobile, keyboard, and dialog
states.

## Constraints

- Keep the static UI/UX-first phase. Do not add Go endpoints, schemas, orchestration,
  WebSockets, SSE, fake API calls, or unsupported mutations.
- Existing live container and backup operations may remain live.
- Planned functionality must remain clearly labeled `Planned`, `Next`, or
  `Requires backend`.
- Preserve the existing design tokens and shadcn/Base UI primitives.
- Prefer shared fixes over route-local patches.
- Do not expose secrets in clear text.

## Priority Goals

### P0 — Correctness and Trust

- [x] Fix the available project browser action in
  `internal/panel-apps/src/routes/projects.$name.tsx:265-284`.
  `Open in browser` currently falls back to `#project-detail-content` instead of
  opening the selected project's URL. Use the project URL, preserve external-link
  behavior, and keep unsupported actions as previews.
- [x] Fix the dashboard service CTA in
  `internal/panel-apps/src/routes/index.tsx:205-209`.
  `View all services` must route to `/services`, not `/containers`.
- [x] Add explicit not-found handling in
  `internal/panel-apps/src/routes/projects.$name.tsx:71-75` and
  `internal/panel-apps/src/routes/services.$name.tsx:64-73`.
  Unknown route params must not silently render the first fixture. Show the requested
  name, explain that the resource was not found, and provide a link back to the list.
- [x] Correct backup size units in
  `internal/panel-apps/src/routes/containers.backup.tsx:148-153`.
  Values in the MB branch must be labeled MB. Add boundary tests for B, KB, MB, and GB.
- [x] Replace placeholder documentation links in
  `internal/panel-apps/src/routes/docs.tsx:55-131` and `:194-207`.
  Use real internal/external destinations where available; otherwise render a clearly
  labeled planned item instead of a clickable `href="#"`.

### P1 — Shared Shell and Navigation

- [x] Create a shared shell composition for the legacy routes currently using ad hoc
  wrappers: `internal/panel-apps/src/routes/containers._index.tsx:71-89`,
  `internal/panel-apps/src/routes/containers.backup.tsx:297-320`, and
  `internal/panel-apps/src/routes/docs.tsx:29-38`.
  All primary pages should share `dashboard-frame`, `dashboard-shell`, the skip link,
  and consistent content scrolling without breaking live container/backup behavior.
- [x] Synchronize sidebar group expansion in
  `internal/panel-apps/src/components/nav-main.tsx:41-58`.
  The active route's group must open after client-side navigation while preserving an
  intentional user collapse where possible.
- [x] Replace the internal full-page brand anchor in
  `internal/panel-apps/src/components/app-sidebar.tsx:313-321` with a TanStack Router
  link so internal navigation preserves SPA state and transitions.
- [x] Make PHP runtime rows discoverable in
  `internal/panel-apps/src/routes/system.tsx:137-175`.
  The visible PHP 8.3/runtime cards should link to the available `/system/php` detail
  route; only unsupported management controls should remain planned/disabled.

### P1 — Live Operation Feedback

- [x] Standardize loading, success, failure, retry, and disconnected states across
  `internal/panel-apps/src/routes/containers._index.tsx:49-115`,
  `internal/panel-apps/src/routes/containers.$name.tsx:90-130`, and
  `internal/panel-apps/src/routes/containers.backup.tsx:176-229`.
  Mutation failures must appear beside the affected action or in a persistent status
  region; log-stream disconnects must say what happened and how to retry; backup query
  failures must not fall through to a misleading empty state.
- [x] Replace the native `alert()` backup failure path at
  `internal/panel-apps/src/routes/containers.backup.tsx:214-218` with the shared dialog/
  status pattern and an actionable recovery message.

### P1 — Accessibility and Secure Data

- [x] Add accessible names and pressed/loading states to icon-only container and backup
  controls in `internal/panel-apps/src/routes/containers._index.tsx:155-184`,
  `internal/panel-apps/src/routes/containers.$name.tsx:287-404`, and
  `internal/panel-apps/src/routes/containers.backup.tsx:656-701`.
  Tooltips are supplementary and must not be the only accessible name.
- [x] Mask database passwords by default in
  `internal/panel-apps/src/routes/containers.$name.tsx:375-404`.
  Add an explicit reveal/hide control, accessible copy feedback, and avoid exposing
  secret values in rendered markup until reveal is requested.
- [x] Finish the shared tab pattern in
  `internal/panel-apps/src/routes/projects.$name.tsx:132-157`,
  `internal/panel-apps/src/routes/services.$name.tsx:173-198`, and
  `internal/panel-apps/src/routes/system.php.tsx:131-157`.
  Add stable tab IDs, `aria-labelledby` on panels, roving keyboard focus, and deep-link
  state where practical without replacing the existing visual language.

### P2 — Inventory Clarity

- [x] Make project filtering and sorting truthful in
  `internal/panel-apps/src/routes/projects.tsx:67-71` and `:143-203`.
  Show filtered versus total counts in the heading and replace the hard-coded “recent”
  comparator with fixture metadata that represents recency.

## Acceptance Criteria

- Every P0/P1 item has a visible, testable result and remains consistent with the static
  phase rule.
- Invalid project/service URLs render a not-found state rather than another entity.
- No clickable placeholder links remain in the docs route.
- All live operations expose loading and failure feedback without browser alerts.
- Every icon-only control has an accessible name and a visible focus state.
- Password values are masked until explicitly revealed.
- Tabs work with keyboard input and expose correct tab/tabpanel relationships.
- The project list count and sort labels match the displayed data.
- Desktop, tablet, and mobile layouts are checked at representative widths; no status
  label, table action, dialog footer, or log viewer is clipped or overlapped.
- Existing container start/stop, log streaming, backup create/restore/delete, and docs
  route reachability remain working.

## Verification Plan

- Run `npm run typecheck` from `internal/panel-apps`.
- Run targeted ESLint on changed source files.
- Run `npm run build` from `internal/panel-apps`.
- Run `git diff --check`.
- Visit every route listed in Audit Scope at desktop, tablet, and mobile widths.
- Keyboard-test sidebar expansion, route links, tabs, dialogs, copy/reveal controls,
  start/stop actions, and backup actions.
- Verify 404/not-found route params for unknown project and service names.
- Verify docs links, dashboard service CTA, project browser action, and PHP runtime
  row destinations.
- Verify live error, retry, disconnected, and success feedback with the existing API.
- Run available component tests and add focused tests for new shared primitives or
  stateful interaction fixes.

## Files To Set As Goals

- `.agent/specs/webui-ux-audit-goals.md`
- `.agent/specs/webui-goals.md`
- `internal/panel-apps/src/components/app-sidebar.tsx`
- `internal/panel-apps/src/components/nav-main.tsx`
- `internal/panel-apps/src/components/app-navbar.tsx`
- `internal/panel-apps/src/routes/index.tsx`
- `internal/panel-apps/src/routes/projects.tsx`
- `internal/panel-apps/src/routes/projects.$name.tsx`
- `internal/panel-apps/src/routes/services.$name.tsx`
- `internal/panel-apps/src/routes/system.tsx`
- `internal/panel-apps/src/routes/system.php.tsx`
- `internal/panel-apps/src/routes/containers._index.tsx`
- `internal/panel-apps/src/routes/containers.$name.tsx`
- `internal/panel-apps/src/routes/containers.backup.tsx`
- `internal/panel-apps/src/routes/docs.tsx`
- `internal/panel-apps/src/styles.css`

## Ready-To-Use Execution Prompt

Read `.agent/specs/webui-ux-audit-goals.md` and `.agent/specs/webui-goals.md` first.
Execute the unchecked P0 and P1 goals in order, then handle P2. Keep the work frontend-only
and static unless an existing container/backup operation is being repaired. Do not add fake
API calls, backend endpoints, or unsupported mutations. Preserve the existing visual system
and shared shadcn/Base UI primitives. After each implementation pass, update only verified
checkboxes in the audit goal file, run the listed validation commands, and visit every route
at desktop, tablet, and mobile widths. Report any browser-visual limitation instead of
assuming the layout is correct.
