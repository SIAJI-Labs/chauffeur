# Backup Diagnostics and Log Follow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add backend diagnostics for database backup enumeration and create a smart auto-follow experience for container logs.

**Architecture:** Keep backup semantics unchanged while adding observability at the server and container layers. In the frontend, replace unconditional log auto-scrolling with explicit follow-mode state that responds to user scrolling and live log updates.

**Tech Stack:** Go, net/http, React, TanStack Router, TanStack Query, TypeScript

---

### Task 1: Add database enumeration diagnostics

**Files:**
- Modify: `internal/podman/container.go`

**Step 1: Add minimal logging around MySQL database enumeration**

Update `listMySQLDatabases` to record:
- container name
- configured username
- raw `SHOW DATABASES;` output after parsing
- filtered system database names
- final returned names

Do not log passwords.

**Step 2: Keep the filtering logic behavior unchanged**

Preserve this exact exclusion behavior:
- `information_schema`
- `performance_schema`
- `mysql`
- `sys`

Only add diagnostics around the existing behavior.

**Step 3: Verify the code still builds**

Run: `go test ./...`

Expected: Tests pass or, if the repo already has unrelated failures, the modified packages compile without new errors.

**Step 4: Commit**

```bash
git add internal/podman/container.go
git commit -m "feat: add mysql backup diagnostics"
```

### Task 2: Add server-side backup request and outcome logging

**Files:**
- Modify: `internal/panel/server.go`

**Step 1: Log database listing requests at the server layer**

In `handleListDatabases`, log:
- container name
- engine if available from config
- number of databases returned
- listing errors when present

**Step 2: Log backup creation request details**

In `handleCreateBackup`, log:
- container name
- requested database names
- per-database backup start
- per-database backup success with filename
- per-database backup failure with error
- final counts for succeeded and failed backups

Keep the current behavior that continues on individual database failures.

**Step 3: Verify the code still builds**

Run: `go test ./...`

Expected: Same as Task 1, with no new failures caused by logging.

**Step 4: Commit**

```bash
git add internal/panel/server.go
git commit -m "feat: log backup request diagnostics"
```

### Task 3: Add smart follow-mode state to the container logs view

**Files:**
- Modify: `internal/panel-apps/src/routes/containers.$name.tsx`

**Step 1: Introduce follow-mode and scroll-container refs**

Add state and refs for:
- whether logs are currently following live output
- the scrollable log container element

Default follow mode to `true`.

**Step 2: Replace unconditional scroll-to-bottom logic**

Update the log update effect so it only scrolls when follow mode is enabled.

**Step 3: Add scroll detection logic**

Detect when the user is near the bottom using a small threshold.
- If near bottom, set follow mode to `true`
- If not near bottom, set follow mode to `false`

This should be driven by the container’s scroll event.

**Step 4: Add a visible resume control**

Render a small button over or near the logs panel when follow mode is paused.
- Label: `Jump to latest` or `Resume follow`
- On click: scroll to bottom and set follow mode to `true`

**Step 5: Keep the current log truncation behavior intact**

Do not change the existing `slice(-100)` log retention behavior in this task.

**Step 6: Verify frontend build/test state**

Run the project’s relevant frontend verification command if available. If no targeted frontend test exists, run:

```bash
go test ./...
```

and verify the TypeScript file remains syntactically correct via the existing build pipeline if available.

**Step 7: Commit**

```bash
git add internal/panel-apps/src/routes/containers.$name.tsx
git commit -m "feat: improve container log auto follow"
```

### Task 4: Manual verification

**Files:**
- No file changes required

**Step 1: Verify backup diagnostics manually**

Run the app and trigger:
- database listing for a MySQL-family container
- a multi-database backup request

Confirm logs show:
- configured username
- raw databases seen
- filtered databases
- returned databases
- selected databases requested for backup
- per-database success/failure outcomes

**Step 2: Verify log follow manually**

Open a running container detail page and confirm:
- logs auto-follow on initial load
- scrolling up pauses follow mode
- new logs do not move the viewport while paused
- clicking resume jumps to the latest logs
- scrolling back to bottom automatically resumes follow mode

**Step 3: Final verification command**

Run: `go test ./...`

Expected: No new regressions introduced by these changes.

**Step 4: Commit verification or fixups if needed**

```bash
git add .
git commit -m "test: verify backup diagnostics and log follow"
```
