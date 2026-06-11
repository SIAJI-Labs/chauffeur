# Selector Viewport Consistency Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Apply consistent terminal-height-aware viewport behavior to all similar Bubble Tea selectors in the CLI.

**Architecture:** Introduce a small shared viewport helper in the `commands` package, then refactor each selector view to use the same slice/indicator logic while preserving each selector’s existing navigation and selection semantics. Use focused regression tests to prove cursor visibility and bounded rendering in short terminals.

**Tech Stack:** Go, Bubble Tea

---

### Task 1: Add shared viewport tests for remaining selectors

**Files:**
- Modify: `internal/commands/podman_test.go`
- Create or Modify: `internal/commands/install_test.go`

**Step 1: Write a failing test for `singleSelectModel` viewport behavior**

Cover:
- short terminal height
- cursor moved downward
- active item remains visible
- output no longer shows earliest items once scrolled

**Step 2: Run the focused test and verify it fails for the current implementation**

Run: `go test ./internal/commands -run TestSingleSelectModelView_ -v`

Expected: FAIL because the current single selector still renders the full list.

**Step 3: Write a failing test for `containerSelectModel` viewport behavior**

Cover:
- short terminal height
- cursor moved into later container rows
- active container remains visible
- rendered output respects viewport limits

**Step 4: Run the focused test and verify it fails**

Run: `go test ./internal/commands -run TestContainerSelectModelView_ -v`

Expected: FAIL because the current container selector still renders the full list.

**Step 5: Write a failing test for `phpSelectModel` viewport behavior with mixed installed rows**

Cover:
- installed rows included in rendering
- selectable cursor moved downward
- active selectable item remains visible
- output is clipped correctly

**Step 6: Run the focused test and verify it fails**

Run: `go test ./internal/commands -run TestPHPSelectModelView_ -v`

Expected: FAIL because the current PHP selector still renders the full list.

**Step 7: Commit**

```bash
git add internal/commands/podman_test.go internal/commands/install_test.go
git commit -m "test: add selector viewport regressions"
```

### Task 2: Add shared viewport helper

**Files:**
- Create: `internal/commands/selector_viewport.go`
- Test: `internal/commands/podman_test.go` or `internal/commands/selector_viewport_test.go`

**Step 1: Write the failing helper test first**

Test a small helper that computes:
- start/end indices
- whether top indicator is shown
- whether bottom indicator is shown

Use a clear table-driven test with short-height and no-clipping cases.

**Step 2: Run the helper test and verify it fails**

Run: `go test ./internal/commands -run TestComputeSelectorViewport -v`

Expected: FAIL because the helper does not exist yet.

**Step 3: Write the minimal helper implementation**

Implement a helper that accepts:
- total rows
- active row index
- terminal height
- reserved line count

and returns:
- `start`
- `end`
- `showAbove`
- `showBelow`

Keep this helper generic and small.

**Step 4: Run the helper test to verify it passes**

Run: `go test ./internal/commands -run TestComputeSelectorViewport -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/commands/selector_viewport.go internal/commands/selector_viewport_test.go
git commit -m "feat: add selector viewport helper"
```

### Task 3: Refactor `singleSelectModel` and `containerSelectModel`

**Files:**
- Modify: `internal/commands/podman.go`
- Test: `internal/commands/podman_test.go`

**Step 1: Add window size handling to both models**

Store width/height from `tea.WindowSizeMsg` in:
- `singleSelectModel`
- `containerSelectModel`

**Step 2: Refactor `singleSelectModel.View()` to use the shared helper**

Render only the visible slice and add indicators when clipped.

**Step 3: Refactor `containerSelectModel.View()` to use the shared helper**

Render only the visible slice and preserve disabled-row styling and behavior.

**Step 4: Run focused selector tests**

Run: `go test ./internal/commands -run 'Test(Single|Container)SelectModelView_' -v`

Expected: PASS.

**Step 5: Run full commands package tests**

Run: `go test ./internal/commands`

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/commands/podman.go internal/commands/podman_test.go
git commit -m "feat: add viewport scrolling to podman selectors"
```

### Task 4: Refactor `phpSelectModel`

**Files:**
- Modify: `internal/commands/install.go`
- Test: `internal/commands/install_test.go`

**Step 1: Add window size handling to `phpSelectModel`**

Store width/height from `tea.WindowSizeMsg`.

**Step 2: Map selectable cursor index to rendered row index**

Before slicing, compute which displayed row corresponds to the current selectable cursor.

**Step 3: Refactor `phpSelectModel.View()` to use the shared helper**

Render a clipped slice while preserving installed-row display and selectable cursor behavior.

**Step 4: Run focused PHP selector tests**

Run: `go test ./internal/commands -run TestPHPSelectModelView_ -v`

Expected: PASS.

**Step 5: Run full commands package tests**

Run: `go test ./internal/commands`

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/commands/install.go internal/commands/install_test.go
git commit -m "feat: add viewport scrolling to php selector"
```

### Task 5: Final targeted verification

**Files:**
- No file changes required

**Step 1: Run commands package verification**

Run: `go test ./internal/commands`

Expected: PASS.

**Step 2: Run broader targeted verification for previously touched areas**

Run: `go test ./internal/panel ./internal/podman ./internal/commands`

Expected: PASS or `no test files` for `internal/podman`.

**Step 3: Run frontend verification to ensure no unrelated regressions**

Run: `npm run typecheck`

Working directory: `internal/panel-apps`

Expected: PASS.

**Step 4: Commit final verification or fixups if needed**

```bash
git add .
git commit -m "test: verify selector viewport consistency"
```
