# Chauffeur — Codex Knowledge Base

> Purpose: Give Codex/AI coding tools the context they need to generate correct CLI code, configs, and docs for the **Chauffeur** project.

## Codex Usage Note

This document is the authoritative reference for autonomous agents working on Chauffeur. **Always consult and comply with the rules, command contracts, and filesystem layouts defined here before generating or modifying code.** Keep this file up to date when behavior changes, and avoid duplicating guidance elsewhere.

## Documentation Synchronization Rule

**MAINTAINANCE REQUIREMENT**: When making any code changes, feature additions, or architectural modifications, you **must** also update the project documentation to maintain real-time accuracy:

### Required Documentation Updates on Every Change:

1. **README.md** (Project Overview)
   - Update status indicators (✅/🚧/📋/🎯)
   - Add new commands to usage examples
   - Update installation instructions if needed
   - Adjust roadmap to reflect current priorities
   - Modify Getting Started section for new features

2. **docs/TODO_STATUS.md** (Development Tracking)
   - Mark completed features as ✅ 
   - Move in-progress items to appropriate status
   - Update priority queue for new objectives
   - Add new tasks for implemented functionality
   - Adjust release notes and timelines

3. **AGENTS.md** (Technical Contracts)
   - Add new commands with proper contracts
   - Update filesystem layout rules
   - Modify command flags and behaviors
   - Add new architectural decisions or constraints
   - Update any installation or build requirements

### Documentation Update Process:

1. **Before Code Changes**: Review current documentation to understand existing context
2. **During Code Implementation**: Note documentation changes needed
3. **After Code Completion**: Update all three documentation files immediately
4. **Verification**: Ensure all documentation reflects the actual implemented functionality

### Synchronization Checklist:

- [ ] README.md reflects current feature status and usage
- [ ] docs/TODO_STATUS.md shows accurate development progress  
- [ ] AGENTS.md contains up-to-date technical contracts
- [ ] All examples work with current implementation
- [ ] Status indicators are accurate (completed vs. in-progress)
- [ ] Priority queue reflects current development focus

**PRINCIPLE**: Documentation must be as current as the code. Users should never encounter outdated information or examples that don't work with the current implementation.

### Cross-Document Consistency

Ensure information consistency across all three documentation files:

| Type | README.md | docs/TODO_STATUS.md | AGENTS.md |
|------|-----------|-------------------|------------|
| **Command Status** | ✅/🚧/📋/🎯 indicators | Detailed task breakdown | Implementation contracts |
| **Feature Lists** | User-facing overview | Development progress tracking | Technical specifications |
| **Command Examples** | Working examples | N/A | Contract definitions |
| **Timeline** | Roadmap overview | Release notes & sprints | N/A |

### Content Rules:

1. **No Conflicting Information**: Features marked as ✅ in README must match completed status in TODO_STATUS.md
2. **Synchronized Status**: Commands listed in AGENTS.md must appear in both README and TODO_STATUS.md appropriately
3. **Consistent Terminology**: Use same command names, flag names, and descriptions across all documents
4. **Real-Time Accuracy**: As soon as code is implemented, update documentation - no lag between implementation and documentation

---

## 1) Project Snapshot

- **Name**: Chauffeur (CLI for local PHP dev services)
- **Owner/Brand**: **SIAJI** (logo initials **SIA**, geometric; dark-blue brand)
- **Primary OS**: Linux (Arch/Wayland focus), user TZ **Asia/Jakarta (UTC+7)**
- **Goal**: Simple per-project dev services for PHP using **Nginx**, **PHP & PHP‑FPM**, and **Caddy** (for automatic local domains; avoid editing `/etc/hosts`).
- **Non‑Goals**: No DB/queue/scheduler/mail providers; users bring their own. No integrations with third‑party orchestrators unless explicitly stated.
- **Isolation Strategy**: Project‑scoped services that do **not** conflict with host packages; per‑directory version isolation for PHP‑FPM.
- **Registrations**: Projects are **manually** registered (`chauf link`) rather than auto‑scanned.
- **Tech Options**:
  - CLI written in **Go** (preferred for robust single binary) or **Bash** for helper scripts.
  - Avoid assuming the CLI is already on `$PATH` during first‑install; provide bootstrap guidance.
- **Dependency Management**: **No Devbox or external env managers.** Chauffeur installs and uses its own copies of binaries under `~/.chauffeur/` and never touches `/usr/bin`. All services run directly on the host, with isolated paths under the Chauffeur workspace.

---

## 2) Command Surface (authoritative)

> **Rules for Codex**: Respect names, flags, and behaviors below. Do **not** invent new commands without an explicit ADR entry.

### Global CLI

| Command | Flags/Args | Description |
|---|---|---|
| `chauf init` | `--force`, `--quiet` | Initialize Chauffeur workspace in `~/.chauffeur/` (idempotent). Creates default dirs and templates. |
| `chauf start` | `--project <path?>`, `--all`, `--dry-run` | Start services for current project (or all registered with `--all`). |
| `chauf stop` | `--project <path?>`, `--all`, `--dry-run` | Stop services. |
| `chauf uninstall` | `--purge` | Remove Chauffeur workspace. `--purge` also deletes caches and installed runtimes. |
| `chauf link` | `--site <domain>`, `--ssl`, `--php <version>` | Register **PWD** as a project. Optionally map local domain via Caddy, set default PHP for this project. |
| `chauf links` | _none_ | List all registered projects and their metadata. |

### PHP Management

| Command | Flags/Args | Description |
|---|---|---|
| `chauf php install <version>` | `--force`, `--no-ext`, `--from <source>` | Install PHP runtime `<version>` into `~/.chauffeur/php/<version>` (source can be `source`, `tarball`, or `distro-extract`; default `tarball`). |
| `chauf php use <version>` | _none_ | Set global default PHP version used by `chauf php ...` commands. |
| `chauf php isolate <version>` | _none_ | Pin current directory/project to `<version>` (per‑project override). |

> **Version examples**: `8.3`, `8.2`, `7.4`. Keep semantic digits (major.minor), allow patch in metadata but runtime folder stays `major.minor`. Never writes to `/usr/bin`; shims live under `~/.chauffeur/bin/shims`.

> **NOTE**: When implementing new commands or modifying existing ones, update all three documentation files immediately following the Documentation Synchronization Rule (see above).

---

## 3) Filesystem Layout (contract)

```
~/.chauffeur/
  config/
    chauffeur.yaml          # global config (see schema below)
  projects/
    <slug>/
      project.yaml          # per-project config: path, php, domain, ssl, created_at
      runtime/
        php-fpm/            # per-project php-fpm sockets/configs
      logs/                 # nginx, php-fpm, caddy (if not global)
  php/
    8.3/                    # installed runtimes (bin, lib, etc.) — within workspace
    8.2/
    7.4/
  nginx/
    bin/                    # nginx binary if installed by Chauffeur
    etc/                    # nginx.conf, mime.types
    sites-available/
    sites-enabled/
    conf.d/
  caddy/
    bin/                    # caddy binary if installed by Chauffeur
    Caddyfile               # domain routing to per-project services
  bin/
    chauf                   # the installed CLI (if self-managed)
    shims/                  # wrapper scripts exposed on PATH (php-8.3, nginx, caddy)
```

### `chauffeur.yaml` (global)

```yaml
version: 1
telemetry: false
workspace_dir: ~/.chauffeur
caddy:
  enable: true
  http_port: 80
  https_port: 443
nginx:
  enable: true
php:
  default: 8.3
projects_dir: ~/.chauffeur/projects
```

### `project.yaml` (per project)

```yaml
version: 1
path: /absolute/path/to/project
php: 8.3               # overrides global default
site:
  domain: myproj.test  # optional
  ssl: true            # optional
runtime:
  php_fpm_socket: ~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.sock
created_at: 2025-10-30T12:00:00+07:00
```

---

## 4) Behavior & Invariants

- **No external env managers**: Chauffeur runs binaries directly on the host from the workspace prefix.
- **No system prefix writes**: Never touch `/usr/bin`, `/usr/local`, or system service units by default.
- `chauf link` registers **PWD** unless `--project` explicitly provided elsewhere.
- Using **Caddy** avoids editing `/etc/hosts`; Codex should generate Caddyfile entries that route `site.test` → local upstream.
- Services must **not conflict** with host services or ports; prefer per‑project Unix sockets and reverse proxy fan‑out via Caddy/Nginx.
- PHP isolation: `use` sets **global** default; `isolate` writes **project.yaml** override.
- First install must handle not‑on‑PATH scenario: provide shell one‑liner to add `~/.chauffeur/bin` to PATH in `~/.bashrc`/`~/.zshrc`.

---

## 5) Code Generation Guidelines (for Codex)

- Prefer **Go** for the main CLI (single static binary). Use Bash only for thin wrappers.
- Support Linux first; avoid macOS/Windows assumptions.
- **Idempotency**: Re-running `init`, `install`, `link` should never corrupt state.
- **Install prefix**: All binaries/configs live under `~/.chauffeur/` only.
- **PATH shims**: Create wrappers in `~/.chauffeur/bin/shims` for each binary.
- **Dry‑runs**: When `--dry-run` is present, print planned actions without side-effects.
- **Logging**: Human-readable logs to STDOUT; structured logs to `~/.chauffeur/<area>/logs`.
- **Failure logs**: Any command failure must append a detailed log entry under `<workspace>/logs/<component>/`. Log filenames follow `<action>[-<version>]-<YYYYMMDDTHHMMSSZ>.log`, and CLI output must surface the exact path.
- **Errors**: Clear actionable messages with suggested fix.
- **Permissions**: Do not require root; if privileged steps are unavoidable, print the exact `sudo` command for the user to run.
- **CLI Modularity**: Keep `main.go` limited to dispatch; implement each command in its own Go file/package (e.g. `cli/commands/<command>.go`) with focused helpers.

---

## 6) Example Snippets

### 6.1 Generate per‑project PHP‑FPM pool

**Prompt to Codex**: "Create a php-fpm pool config that listens on `<socket>` for user `<user>`, sets `pm = ondemand`, and logs to `<logdir>`."

**Target** (`.conf`):
```
[project]
user = <user>
group = <user>
listen = <socket>
listen.owner = <user>
listen.group = <user>
pm = ondemand
pm.max_children = 10
php_admin_value[error_log] = <logdir>/php-fpm-error.log
php_admin_flag[log_errors] = on
```

### 6.2 Caddy v2 site block for project domain

**Prompt**: "Route `myproj.test` to Unix socket `<socket>` via FastCGI, enable TLS internal if `ssl: true`."

**Target** (`Caddyfile`):
```
myproj.test {
	encode gzip
	tls internal
	php_fastcgi unix/<socket>
	file_server
	root * /absolute/path/to/project/public
}
```

### 6.3 `chauf link` writes project.yaml

**Prompt**: "If `--php` given, write per‑project override; compute `<slug>` from basename; ensure dirs created; avoid overwrite unless `--force`."

**Target** (Go pseudo‑code):
```
slug := slugify(filepath.Base(pwd))
proj := Project{ Path: pwd, PHP: optPHP, Site: {...} }
WriteYAML("~/.chauffeur/projects/"+slug+"/project.yaml", proj)

---

## 7) ADRs (Architecture Decision Records)

1. **ADR‑001: Manual Registration**  – Accepted 2025‑10‑30  
2. **ADR‑002: Caddy for Local Domains** – Accepted 2025‑10‑30  
3. **ADR‑003: PHP Isolation Model** – Accepted 2025‑10‑30  
4. **ADR‑004: No DB/Queues** – Accepted 2025‑10‑30  
5. **ADR‑005: Go as Primary Language** – Accepted 2025‑10‑30  
6. **ADR‑006: No Devbox; Host-Scoped Install Prefix** – Accepted 2025‑11‑02  
   - **Context**: Avoid external dependencies and system conflicts.  
   - **Decision**: Manage all tool binaries (PHP, Nginx, Caddy, etc.) inside `~/.chauffeur/` with shims; do not modify `/usr/bin`.  
   - **Consequences**: Reproducible, user-space installs; zero collision with distro package managers.

---

## 8) Prompts & Guardrails for Codex

- Always assume **UTC** timestamps in generated files unless otherwise specified.
- When creating files in `$HOME`, expand to absolute paths.
- Do not write outside `~/.chauffeur/` unless the user explicitly asks.
- Prefer Unix sockets over TCP for local FastCGI.
- For PHP versions, normalize to `major.minor` for folder names.
- Provide migration messages when a schema changes (`version:` key).

**Style**
- Write clean, commented Go code; small focused packages; avoid global state.
- Configuration access via a single `config` package with typed structs and defaults.
- Command parsing via `spf13/cobra` (or stdlib `flag` if minimal). Provide `--help` autogen.

---

## 9) Acceptance Tests (high‑level)

1. `chauf init` creates workspace; re‑running is no‑op.
2. `chauf php install 8.3` creates `~/.chauffeur/php/8.3/bin/php`.
3. `chauf php use 8.3` updates `chauffeur.yaml` default.
4. `chauf link --site myproj.test --ssl --php 8.2` creates project folder, Caddy block, and PHP‑FPM pool.
5. `chauf start` in project directory boots required services; `stop` halts them cleanly.

---

## 10) Glossary

- **Workspace**: `~/.chauffeur/` root directory.  
- **Isolation**: Per‑project PHP‑FPM and config; no host conflicts.  
- **Registration**: Tracking a project via `project.yaml` under `projects/`.

---

## 11) Open Questions / TODOs

- Decide acquisition method per binary: build-from-source vs vendor tarballs (with checksum/signature verification).  
- Define Windows/macOS support stance.  
- Choose logging format and rotation policy.  
- SSL internal CA persistence & trust bootstrapping UX.

---
