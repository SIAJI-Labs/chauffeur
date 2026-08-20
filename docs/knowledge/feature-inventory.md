# Feature Inventory

Status labels:

- **Implemented:** present in code and reasonably usable.
- **Partial:** present but incomplete, inconsistent, or narrower than docs/specs.
- **Planned/stub:** documented or intended but not currently implemented.

## Workspace and lifecycle

| Area | Status | Current behavior |
|---|---|---|
| `init` | Implemented | Creates workspace/config/shims and diagnoses DNS setup. |
| `info` | Implemented | Shows services, projects, PHP versions, and cache state. |
| `uninstall` | Implemented | Stops services, removes workspace, and prints cleanup hints. |
| `self-update` | Implemented | Supports release and development rebuild flows. |
| `update` | Implemented | Supports nginx, PHP, Composer, all, check, list, rollback flows. |
| `migrate` | Stub | Dispatcher reports not implemented. |

## Installation

| Area | Status | Current behavior |
|---|---|---|
| nginx | Implemented | Downloads and compiles into workspace. |
| PHP 7.4–8.4 | Implemented | Builds from source; legacy versions receive compatibility patches. |
| Composer | Implemented | Downloads Composer PHAR into workspace. |
| Caching/checksums | Implemented | Source/download caching and checksum utilities exist. |
| Host dependency preparation | Partial | `doctor` detects and suggests packages; onboarding is not fully guided. |

## Projects and web serving

| Area | Status | Current behavior |
|---|---|---|
| Project link/list/unlink | Implemented | Explicit registration and config persistence. |
| Project detection | Implemented | Laravel, WordPress, and generic document-root handling. |
| `.test` routing | Implemented | Generated nginx server blocks and optional dnsmasq integration. |
| Multiple aliases | Implemented | Repeatable aliases and SAN certificate support. |
| SSL | Implemented | `secure`/`unsecure` via mkcert and nginx reload. |
| Shared/dedicated FPM | Implemented | Both strategies coexist and are shown in status. |
| Project-scoped service actions | Implemented | Start/stop/restart/status/logs support project selection. |

## Operations and maintenance

| Area | Status | Current behavior |
|---|---|---|
| Start/stop/restart | Implemented | Correct process ordering, readiness checks, graceful reloads. |
| Status | Implemented | Table with PID/uptime/memory and optional paths. |
| Logs/follow | Implemented | CLI follows files by polling; project-specific logs are supported. |
| Doctor | Implemented/partial | Broad dependency, SSL, DNS, network, and port checks; test coverage is incomplete. |
| Clean | Implemented | Cache/log/all cleanup, dry-run, and age filter. |
| Autostart | Implemented | User-level systemd enable/disable/status/list. |

## Configuration and environment

| Area | Status | Current behavior |
|---|---|---|
| PHP config | Partial | Hard-coded keys such as memory and upload limits can be read/set. |
| nginx config | Partial | `upload_max_size` can be read/set. |
| Generic config API | Planned | Docs describe show/set/validate/import/export/reset, but handlers do not. |
| `env` commands | Planned/broken | Main dispatch routes `env` to config handler; no env subcommands are handled. |
| JSON/schema validation | Planned | Not present in current implementation. |

## Databases and panel

| Area | Status | Current behavior |
|---|---|---|
| Podman create/start/stop/list/status/remove | Implemented | Interactive and flag-driven lifecycle for supported engines. |
| Console | Implemented | Database-specific CLI access. |
| Backup/restore | Partial | CLI and panel flows exist; panel backup is synchronous. |
| Panel API | Partial | Container and backup endpoints exist without authentication. |
| Panel frontend | Partial | Container list/detail/backup/docs routes exist; dashboard and navigation contain placeholders. |
| True live logs | Partial | SSE sends one snapshot then exits; it is not continuous follow. |

## Planned future directions

The roadmap names plugins, named workspaces, monitoring, project templates, and remote synchronization. These should remain deferred until the core contract, security model, state model, and integration tests are trustworthy.
