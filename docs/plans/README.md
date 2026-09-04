# Chauffeur Overhaul Plans

This directory contains implementation plans for the Chauffeur overhaul. The plans are intentionally separated by capability, while the master roadmap defines dependency order and release gates.

## Master roadmap

- [2026-08-19-overhaul-roadmap.md](./2026-08-19-overhaul-roadmap.md) — sequencing, shared architecture, priorities, and release gates.

## Capability plans

- [2026-08-30-native-podman-runtime-parity-investigation.md](./2026-08-30-native-podman-runtime-parity-investigation.md) — confirmed native/Podman command, lifecycle output, reverse-proxy, and auto-start parity issues with fixes and acceptance criteria.
- [2026-08-19-podman-runtime-overhaul.md](./2026-08-19-podman-runtime-overhaul.md) — migrate native nginx/PHP runtime to rootless Podman.
- [2026-08-19-podman-service-management.md](./2026-08-19-podman-service-management.md) — generalize database containers into managed service presets.
- [2026-08-19-link-wizard-tui.md](./2026-08-19-link-wizard-tui.md) — interactive PHP/Node/database/service setup flow.
- [2026-08-19-web-ui-overhaul.md](./2026-08-19-web-ui-overhaul.md) — project-centric web control plane.
- [2026-08-19-developer-observability-tools.md](./2026-08-19-developer-observability-tools.md) — logs, env intelligence, requests, dumps, queries, and profiling.

## How to use these plans

1. Start with the master roadmap.
2. Implement one phase and its gate before opening dependent feature work.
3. Keep CLI and web UI mutations on shared Go operations.
4. Preserve compatibility aliases and existing state until migration gates pass.
5. Update the relevant plan when implementation decisions change.

## Cross-cutting rule

The overhaul copies Lerd’s cohesion, progressive disclosure, safety, and verification discipline. It does not copy Lerd’s entire feature count. New features must fit the shared workspace, project, service, operation, diagnostic, and event models before implementation.
