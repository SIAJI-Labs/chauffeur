# Reference Overview

## Product positioning

Lerd positions itself as an open-source Herd-like local PHP development environment for Linux and macOS, with Windows support through WSL2 in beta. It is Podman-native, rootless, and includes a built-in web UI.

Primary references:

- `/home/siegg/Workspaces/Personal/Projects/lerd/README.md:1-5,17-25`
- `/home/siegg/Workspaces/Personal/Projects/lerd/docs/reference/architecture.md:1-87`

Lerd’s product boundary is broader than Chauffeur’s current boundary. It serves as a local-development operating layer rather than only a PHP/nginx process manager.

## Supported users and workflows

Lerd supports:

- Linux and macOS PHP developers.
- Windows developers through WSL2 beta.
- Laravel, Symfony, WordPress, Drupal, Magento, CakePHP, CodeIgniter, Statamic, Tempest, and custom PHP projects.
- Non-PHP projects through custom containers or host-proxy development servers.
- Developers using SSH/tmux through a terminal dashboard.
- Teams sharing project setup through committed `.lerd.yaml` files.
- AI-assisted workflows through MCP configuration and tools.

Chauffeur’s current intended audience is narrower: Linux PHP developers who want Valet/Herd-like domains and per-project PHP without Docker for web serving.

## Lerd’s core promise

The short path is:

```bash
lerd install
cd my-project
lerd link
lerd open
```

The result is a working local domain, normally with HTTPS. Lerd also provides `.localhost` mode when users do not want managed DNS or certificate setup.

## Runtime model

Lerd uses rootless Podman containers for nginx, PHP-FPM, DNS, databases, and supporting services. Podman Quadlets and systemd user services supervise the runtime on Linux. macOS uses Podman Machine and launchd.

Chauffeur instead compiles and runs nginx and PHP-FPM directly under `~/.chauffeur`; Podman is optional and focused on databases.

## Major Lerd subsystems

- Cobra CLI command tree.
- Go configuration and project model.
- Podman/Quadlet service orchestration.
- nginx and DNS integration.
- Svelte web UI with HTTP/WebSocket APIs.
- Bubble Tea TUI.
- System tray binary.
- Filesystem watcher and event bus.
- Framework and service registries.
- Site/environment doctors.
- Logs, request statistics, profiling, debug bridge, and notifications.
- MCP server.
- Database snapshots, service rollback, and recovery workflows.

## Repository evidence

The reference repository has a large package and test surface, including roughly 1,421 Go files, 794 Go test files, 254 Svelte files, and 171 frontend test files according to the scan inventory. The exact counts can change with repository revisions; the important observation is that Lerd is a mature platform with extensive verification, not a small utility.

## What Chauffeur should and should not infer

Lerd proves that a local PHP tool can become a cohesive development platform. It does not prove that every feature belongs in Chauffeur. Chauffeur should copy patterns that improve trust and time-to-success, while retaining its simpler direct-process architecture unless container isolation becomes an explicit product decision.
