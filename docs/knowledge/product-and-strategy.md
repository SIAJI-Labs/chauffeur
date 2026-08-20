# Product And Strategy

## Product definition

Chauffeur is a Linux-native local development environment for PHP projects. It provides the convenience of Valet/Herd without requiring Docker for web serving and without installing managed services into system directories.

The product manages:

- nginx as the local HTTP/HTTPS front end.
- Multiple PHP versions, currently intended to cover 7.4 through 8.4.
- PHP-FPM pools, shared by PHP version or dedicated to a project.
- Composer through a workspace-local PHAR and shim.
- Project registration and `.test` domain routing.
- Local trusted certificates through mkcert.
- Optional DNS and low-port forwarding integrations.
- Optional Podman database containers and backup/restore.
- An embedded local web panel for the Podman workflow.

## Mission

Provide Linux PHP developers with a predictable, project-oriented local environment where services are isolated in `~/.chauffeur`, operations are explicit, and common project setup takes a few commands.

## Target users

Primary users are individual Linux PHP developers, freelancers, and developers moving from macOS Valet/Herd. They likely work on several client or Laravel projects and need different PHP versions without system conflicts.

Chauffeur is not positioned as a production deployment system, a team orchestration platform, or a replacement for Docker/Kubernetes in large collaborative environments.

## Product principles

- **Workspace-first:** binaries, configuration, runtime state, logs, sockets, and caches live under one workspace.
- **Zero silent host mutation:** root/system changes are shown as explicit commands or instructions.
- **Explicit registration:** projects are linked intentionally; there is no background project scanner.
- **Idempotence:** repeating an operation should update or skip safely rather than corrupt state.
- **Linux-native:** behavior should follow Linux process, DNS, systemd, and filesystem conventions.
- **Project-centric value:** users care primarily about a working project URL and runtime, not individual daemons.
- **Actionable failure:** errors should explain the problem, its impact, and the next command.

## Core value proposition

The shortest successful path is:

```text
prepare host → initialize workspace → install runtime → link project → start → open URL
```

The experience should optimize for successful first request, reliable switching between projects, and fast diagnosis when DNS, SSL, PHP, or nginx is unhealthy.

## Strategic direction for the overhaul

The highest-leverage improvement is trust, not more features. Users must be able to believe the command help, docs, status output, and panel. The recommended strategic sequence is:

1. Reconcile implementation, help, docs, and roadmap.
2. Make workspace state and project health observable.
3. Fix security boundaries around secrets, panel access, and filesystem paths.
4. Make onboarding diagnostic-first and guided without hiding explicit changes.
5. Make the panel either a complete Chauffeur control plane or explicitly narrow it to databases.
6. Add integration coverage for the real `link → start → request` journey.
7. Only then consider plugins, monitoring, templates, or multi-workspace expansion.

## Success measures

Existing product goals include CLI startup under 100 ms, fast warm service startup, at least 80% test coverage, zero critical data-loss bugs, and idempotent commands. For the experience overhaul, add outcome measures:

- A new user can identify missing host prerequisites before compilation.
- A linked project exposes a clear URL, PHP version, SSL state, and service health in one place.
- Every documented command and flag is either supported or explicitly marked planned.
- A failed start or request produces a direct remediation path.
- No panel response exposes credentials by default.
- Backup and restore operations preserve progress, failure state, and destructive-action confirmation.

## Non-goals and constraints

- Linux remains the primary platform.
- PHP web serving remains process-based rather than Docker-based.
- The default workspace remains user-local and requires no root for ordinary service operations.
- Host integrations such as dnsmasq, systemd, and iptables must remain explicit.
- First PHP compilation will remain materially slower than a package install; the product should set expectations and improve visibility rather than pretend otherwise.
