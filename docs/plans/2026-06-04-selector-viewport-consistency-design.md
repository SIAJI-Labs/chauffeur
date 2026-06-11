# Selector Viewport Consistency Design

## Summary

The Bubble Tea selectors in the CLI currently render full lists without respecting terminal height. This causes inconsistent visible output, where shorter terminals hide earlier or later items and make it appear that choices are missing. The database selector has already been fixed once; this design extends the same viewport-aware behavior to all similar selectors for consistent terminal UX.

## Problem

Several selector views in the CLI render all rows into a single view string and rely on the terminal to clip what does not fit. As a result:

- the visible item set changes with terminal height,
- the active cursor can move outside the visible portion,
- users may think items are missing,
- behavior differs between selectors.

Affected selectors include:

- database multi-select in `internal/commands/podman.go`
- single-item selector in `internal/commands/podman.go`
- container selector in `internal/commands/podman.go`
- PHP version selector in `internal/commands/install.go`

## Goals

- Make all similar Bubble Tea selectors respect terminal height.
- Keep the cursor’s active item visible.
- Use consistent `↑ more` / `↓ more` indicators where content is clipped.
- Preserve existing selection behavior and keybindings.
- Reduce duplication by centralizing viewport calculations.

## Non-Goals

- No redesign of keybindings.
- No migration to a larger viewport framework.
- No changes to non-TUI fallback selectors.
- No changes to business logic for selection, filtering, or disabled items.

## Design

### Shared viewport helper

Introduce a small shared helper in the `commands` package that computes:

- visible item count based on terminal height,
- start and end indices for the rendered slice,
- whether `↑ more` and `↓ more` indicators are needed.

The helper should accept:

- total row count,
- active row index,
- terminal height,
- reserved line count for title/help/footer,
- optional indicator line allowance.

It should return:

- `start`
- `end`
- `showAbove`
- `showBelow`

### Selector model updates

Each affected selector model will store terminal dimensions received through `tea.WindowSizeMsg`.

Each `View()` implementation will:

1. compute its reserved non-item rows,
2. compute the visible slice using the shared helper,
3. render only the visible rows,
4. insert `↑ more` above the list when clipped at the top,
5. insert `↓ more` below the list when clipped at the bottom.

### Selector-specific considerations

#### Database multi-select
- Already has partial viewport support.
- Will be refactored to use the shared helper instead of local logic.

#### Single-item selector
- Simplest case: cursor maps directly to item index.
- Viewport should scroll as the cursor moves.

#### Container selector
- Must keep disabled entries rendered and visible.
- Cursor navigation already skips disabled entries; viewport logic only affects rendering.

#### PHP version selector
- Hardest case because displayed rows include both installed and selectable items.
- Cursor maps only across selectable rows, but viewport must be computed over rendered rows.
- The design should derive the rendered active row index from the current selectable cursor before slicing.

## Testing Strategy

Add focused tests that render selectors with a short `tea.WindowSizeMsg` and assert:

- the active cursor row is visible,
- earlier rows scroll out when the cursor moves down,
- later rows are clipped in short terminals,
- rendered output does not exceed the intended height budget,
- indicators appear when clipping occurs.

Specific tests:

- multi-select viewport regression test
- single-select viewport test
- container-select viewport test
- PHP selector viewport test with installed/non-installed mixed rows

## Risks and Mitigations

- **Risk:** Different selectors need different reserved row counts.
  - **Mitigation:** Keep reserved line count explicit per view while sharing only the index/window math.

- **Risk:** PHP selector cursor mapping becomes incorrect.
  - **Mitigation:** Test with mixed installed/selectable rows and compute active rendered row explicitly.

- **Risk:** Off-by-one line count bugs with indicator rows.
  - **Mitigation:** Include line-budget assertions in tests.

## Acceptance Criteria

- All similar Bubble Tea selectors respect terminal height.
- The active cursor item remains visible while navigating.
- Clipped selectors show `↑ more` and/or `↓ more` as appropriate.
- Existing selection semantics remain unchanged.
- Added regression tests pass.
