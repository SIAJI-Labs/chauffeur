# Architecture And State

## Direct process versus container runtime

| Concern | Chauffeur | Lerd |
|---|---|---|
| PHP/nginx runtime | Workspace-local compiled binaries and direct processes | Rootless Podman containers and Quadlets |
| PHP-FPM | Shared or dedicated Unix-socket pools | Shared PHP-FPM containers per version, with broader runtime options |
| Databases | Optional Podman containers | First-class service presets and database lifecycle |
| Supervision | Custom process manager plus optional systemd autostart | Podman Quadlets, systemd user services, launchd on macOS |
| Platforms | Linux-first | Linux/macOS, WSL2 beta |
| Host mutation | Explicit instructions for system integrations | Consolidated install/bootstrap with controlled privileged setup |

Neither model is universally better. Chauffeur’s direct-process model is smaller and avoids container overhead for PHP. Lerd’s container model gives stronger isolation, broader service portability, and a common runtime primitive.

## Shared operational state

Lerd’s most important architectural lesson is to make CLI, web UI, TUI, watcher, diagnostics, and MCP consume the same state and action model. The TUI often shells out to public CLI commands for mutations, while the UI uses shared event/state mechanisms and periodic cache refresh.

Chauffeur currently has separate command handlers and a narrower panel API. Before adding more interfaces, define shared concepts:

```text
Workspace
Project
Domain
Runtime
FPM pool
Service
Container
Backup
Diagnostic check
Operation/job
```

Each concept should have a stable state representation and action result that can be rendered in CLI and panel surfaces.

## Configuration scopes

Lerd separates:

- Global config under XDG config.
- Project intent in `.lerd.yaml`.
- Site registry under XDG data.
- Runtime/service data under XDG data.
- Workspaces as display-only groups.

Chauffeur primarily stores global and project config under the workspace. This is coherent with workspace isolation but weaker for committed project setup and cross-interface discovery.

Potential Chauffeur split:

```text
project/.chauf.yaml       optional portable project intent
~/.chauffeur/config/      machine/global preferences
~/.chauffeur/projects/    registered/generated runtime state
~/.chauffeur/services/    process/runtime state
```

Do not introduce this split without first defining precedence and migration behavior. The existing workspace model should remain valid for users who do not want project files changed.

## Eventing and live state

Lerd has an event publisher/coalescer, snapshot invalidation, WebSocket updates, and polling safety nets. This supports live dashboard updates and cross-interface consistency while avoiding expensive continuous polling.

Chauffeur’s CLI status is computed on demand. The panel polls containers and its logs SSE currently sends one snapshot rather than a true stream. If a panel remains strategic, Chauffeur needs either:

- A small shared event/state layer with WebSocket/SSE subscriptions, or
- A deliberately polling-based UI with honest snapshot semantics.

Do not label snapshots as live streams.

## Service and framework registries

Lerd decouples framework definitions and service presets from binary releases. Logical service references resolve to current versions/images/configuration. This allows capability growth without shipping a new binary for every preset.

Chauffeur currently hard-codes its core PHP/nginx/Composer installers and Podman engines. A registry could eventually help with database engines or PHP build metadata, but it adds update, trust, compatibility, and offline-cache concerns. Defer until the core model needs it.

## Worktrees and portability

Lerd treats Git worktrees as first-class sites with optional PHP/Node/database/environment overrides and branch domains. Chauffeur does not currently model worktrees explicitly.

This is a valuable future feature for project-centric development, but it depends on a portable project manifest and a reliable project identity model. It should follow, not precede, state-model cleanup.

## Security architecture lessons

Lerd contains explicit tests for WebSocket origin checks, frame limits/deadlines, remote access gates, sudo refusal, environment-file safety, nginx escaping, and auth/session behavior. Chauffeur’s panel currently has no authentication, returns passwords in detail responses, and uses wildcard CORS for SSE.

The immediate lesson is not to copy Lerd’s full remote-access feature set. It is to make security behavior explicit and testable before expanding local control surfaces.
