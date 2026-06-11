# Backup Diagnostics and Log Follow Design

## Summary

This change adds two focused improvements:

1. Backend diagnostic logging for database enumeration and backup creation so missing MySQL databases can be explained with evidence.
2. Smarter log-panel auto-follow behavior that follows live output by default, pauses when the user scrolls away, and resumes with an explicit control or when the user returns to the bottom.

The goal is to diagnose why backups appear to omit some tables without changing backup semantics, while also improving the container log viewing experience.

## Problem

The current backup flow enumerates databases using the configured container credentials and silently filters system databases. When the operator compares that behavior to a manual `mysql -u root ... SHOW DATABASES;` check, differences are hard to explain because the application does not log what it saw, what it filtered, or which selected backups succeeded or failed.

Separately, the container detail page always scrolls the log view to the bottom whenever logs update. This keeps live output visible, but it prevents users from inspecting older logs without being pulled back to the bottom.

## Goals

- Add low-risk diagnostics to explain which databases the backup system sees.
- Preserve the existing API contract for backup creation and database listing.
- Keep backup behavior unchanged except for better observability.
- Make the log panel follow live output by default.
- Pause auto-follow when the user manually scrolls upward.
- Provide an obvious way to resume follow mode.

## Non-Goals

- No new backup API endpoints.
- No changes to which databases are filtered.
- No changes to how mysqldump selects tables.
- No streaming terminal emulation or xterm integration.

## Design

### 1. Backup diagnostics

Diagnostics will be added in the existing server and container backup path.

For database enumeration:
- Log container name and engine.
- Log the configured database username used for enumeration.
- For MySQL-family engines, log the raw database names returned by `SHOW DATABASES;`.
- Log which databases are filtered as system databases.
- Log the final list returned to the UI.

For backup creation:
- Log the incoming container name and selected database list from the request.
- Log per-database backup start events.
- Log per-database backup failures with the database name and error.
- Log per-database backup success including output filename.
- Log a final summary including created and failed backup counts.

These logs should use the project’s existing logging style where possible and avoid exposing raw passwords. Logging the configured username is acceptable because it is already a non-secret operational identifier and is necessary for diagnosing visibility mismatches.

### 2. Log panel smart follow mode

The container detail page will shift from unconditional bottom scrolling to a follow-mode model.

Behavior:
- Follow mode starts enabled.
- When new log data arrives and follow mode is enabled, the panel scrolls to the bottom.
- If the user scrolls upward beyond a small bottom threshold, follow mode turns off.
- While follow mode is off, new log data does not change the current scroll position.
- A small “Jump to latest” or “Resume follow” control appears while paused.
- Clicking the control scrolls to the bottom and re-enables follow mode.
- If the user manually scrolls back near the bottom, follow mode re-enables automatically.

Implementation detail:
- The scroll listener should be attached to the actual scrollable logs container, not only to the `<pre>` content node.
- Bottom detection should use a small threshold to prevent flicker from sub-pixel or font rendering differences.

## Files Expected To Change

- `internal/podman/container.go`
- `internal/panel/server.go`
- `internal/panel-apps/src/routes/containers.$name.tsx`

## Testing Strategy

### Backup diagnostics
- Verify database listing still returns the same JSON structure.
- Verify backup creation still succeeds for selected databases.
- Manually inspect server logs during listing and backup creation to confirm the new diagnostics appear and include raw, filtered, and final database lists.

### Log follow behavior
- Open a running container detail page.
- Confirm logs follow live output initially.
- Scroll upward and confirm follow mode pauses.
- Confirm new logs do not yank the viewport.
- Click the resume control and confirm the panel jumps to the latest output.
- Scroll back to the bottom manually and confirm follow mode auto-resumes.

## Risks and Mitigations

- **Risk:** Excessive logging noise.
  - **Mitigation:** Keep logs concise and only emit at high-value state transitions.

- **Risk:** Incorrect scroll target leads to flaky follow behavior.
  - **Mitigation:** Attach the ref and scroll calculations to the actual scrolling container and use a bottom threshold.

- **Risk:** Sensitive data exposure.
  - **Mitigation:** Never log passwords or full command arguments that contain passwords.

## Acceptance Criteria

- When listing databases for a MySQL-family container, logs show raw, filtered, and returned database names.
- When creating backups, logs show which databases were requested and which succeeded or failed.
- The container log panel follows live output by default.
- Scrolling up pauses follow mode.
- A visible resume control restores follow mode and scroll position.
- Returning to the bottom re-enables follow mode automatically.
