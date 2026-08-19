# Lerd Comparison Knowledge

This directory records observations from the reference repository at `/home/siegg/Workspaces/Personal/Projects/lerd` and translates them into decisions for Chauffeur’s experience overhaul.

The Lerd repository was scanned read-only. The comparison is not a claim that Chauffeur should become Lerd. It identifies reusable patterns, deliberate differences, and scope boundaries.

## Reading order

1. [Reference overview](./reference-overview.md)
2. [Experience comparison](./experience-comparison.md)
3. [Architecture and state comparison](./architecture-and-state.md)
4. [Feature and interface matrix](./feature-and-interface-matrix.md)
5. [Lessons and adoption plan](./lessons-and-adoption-plan.md)

## Executive conclusion

Lerd’s strongest advantage is not any individual feature. It is the shared operational model across CLI, web UI, TUI, diagnostics, and automation. Chauffeur should adopt that principle selectively: make project state, service state, diagnostics, and actions consistent first. It should not immediately copy Lerd’s full platform scope, multi-platform runtime model, or large service ecosystem.
