# Link Wizard TUI Plan

This plan is Phase 2 of the [master overhaul roadmap](./2026-08-19-overhaul-roadmap.md). The wizard consumes shared project detection/setup-plan types and gains service choices after the generic service model is available.

## Goal

Improve `chauf link` with an interactive project setup wizard that lets a developer choose the runtime and supporting services before the project is linked.

The desired flow is inspired by the Lerd experience:

```text
❯ chauf link

┃ PHP version
┃ > 8.3

  Node version
  Clear to follow .nvmrc instead of pinning
  > System

  Database
    SQLite (no service)
  > PostgreSQL (chauf-postgres)
    MySQL (chauf-mysql8)

  Services
  > [•] mailpit
    [ ] meilisearch
    [•] redis

Enter to continue
```

The wizard should make the project’s complete local-development setup visible before mutation, while keeping `chauf link` fully usable in scripts and CI through flags.

## Visual direction

The wizard should use a compact inline form rather than a full-screen bordered selector. The terminal should feel like a calm configuration prompt: the user can see the whole setup at once, the active row is obvious, and selected service state is readable without opening another screen.

Reference composition:

```text
❯ chauf link

┃ PHP version
┃ > 8.5

  Node version
  Clear to follow .nvmrc instead of pinning
  > 18

  Database
    SQLite (no service)
  > MySQL (lerd-mysql)
    PostgreSQL (lerd-postgres)

  Services
  > [•] mailpit
    [ ] meilisearch
    [•] redis
    [•] rustfs

enter next
```

### Visual states

The text color and marker must communicate state consistently:

| Element | Visual treatment | Meaning |
|---|---|---|
| Command prompt | Green command name | The command currently running. |
| Section label | Purple | A setup group: PHP, Node, Database, Services. |
| Active row | `>` cursor plus normal/high-contrast text | The choice controlled by Up/Down. |
| Selected service | Green `[•]` | The service will be included in the setup plan. |
| Unselected service | Gray `[ ]` | The service is available but not selected. |
| Selected database | Green option text plus `>` cursor | The active database choice. |
| Inactive database | Gray option text | A valid alternative that is not active. |
| Help/description | Gray | Secondary explanation, such as `.nvmrc` behavior. |
| Disabled/unavailable option | Dim gray | The choice cannot currently be used; include a reason on demand. |
| Footer action | Gray lowercase text | Contextual control, for example `enter next`. |

The cursor and selection are different concepts:

- `>` identifies the currently focused row.
- `[•]` identifies a selected item in a multi-select group.
- A focused but unselected service should render as `> [ ] service`.
- A selected but unfocused service should render as `  [•] service`.

Do not use color as the only state signal. The markers must remain meaningful when color is disabled or output is captured.

### Spacing and density

- Keep one blank line between setup groups.
- Keep descriptions directly below their parent option.
- Use a two-space content indent for ordinary rows.
- Use a visible left rail such as `┃` only for the currently expanded/focused group if it improves focus; do not draw a full box around the form.
- Keep the footer at the bottom of the rendered form with the smallest useful instruction.
- Avoid decorative borders, large titles, or repeated instructions.

### Interaction copy

Use short lowercase contextual prompts:

```text
enter next
space toggle
esc back
ctrl+c cancel
```

The footer should change based on the active group:

```text
enter next                 # single-select group
space toggle · enter next  # service multi-select
enter apply · esc back     # review screen
```

The wizard should not repeat the full keyboard legend on every frame. Show the compact footer by default and expose the complete legend through `?`.

## Why this change

The current `chauf link` flow supports PHP, SSL, aliases, slug, domain, and dedicated FPM flags, but it does not help users answer the practical setup questions that follow linking:

- Which PHP version should this project use?
- Should the project pin Node or follow `.nvmrc`?
- Does it need SQLite, MySQL, or PostgreSQL?
- Which supporting services should start with the project?
- Which services are already available, installed, or missing?
- What will be changed before the project is linked?

The existing flow writes configuration immediately after parsing flags. The wizard should add progressive disclosure and safer defaults without making the CLI less scriptable.

## Product principles

- **Project-first:** choices are presented in the context of the detected project.
- **Preview before mutation:** show the setup summary before writing config, generating certificates, or starting services.
- **Detect, do not assume:** use project files and installed runtimes to suggest defaults.
- **Noninteractive parity:** every wizard decision must have a flag/config equivalent.
- **Safe defaults:** choose the least surprising option and avoid installing or starting hidden services.
- **Reversible:** linking can be repeated; changing choices updates project intent without deleting user data.
- **One model:** the wizard, CLI flags, and web UI use the same project setup model.

## Scope

### Included

- Interactive `chauf link` wizard.
- PHP version selection.
- Node version selection or `.nvmrc` follow mode.
- Database selection.
- Supporting service multi-select.
- SSL/domain/FPM choices within the same setup summary.
- Project configuration persistence.
- Existing-project relink/edit flow.
- Noninteractive flag equivalents.
- Bubble Tea keyboard navigation and terminal-size handling.

### Deferred

- Framework-specific setup commands.
- Queue/scheduler/Horizon/Reverb worker orchestration.
- Public tunnels and LAN sharing.
- Full service marketplace.
- Automatic dependency installation without confirmation.
- Arbitrary custom container definitions in the first wizard release.
- Remote setup or team synchronization.

## Entry points

### Interactive

When all of the following are true:

- `chauf link` is attached to a TTY.
- No setup-affecting flags were supplied, or the user explicitly requests the wizard.
- The target project is valid.

Run the setup wizard.

Explicit form:

```bash
chauf link --interactive
```

### Noninteractive

When stdin/stdout is not a TTY, or `--no-interactive` is passed, use detection and flags without prompting:

```bash
chauf link --no-interactive
```

CI must never hang waiting for a prompt. If a required choice cannot be inferred, return a usage error with the missing flag.

## Wizard flow

### Step 0: Project detection summary

Before choices, detect and show:

```text
Project:     my-shop
Path:        ~/Projects/my-shop
Type:        Laravel
Document:    public/
PHP needs:   8.2–8.3 from composer.json
Node hint:   .nvmrc → 20
Env files:   .env.example, .env
```

Detection must not modify files.

### Step 1: PHP version

Show supported and locally available versions. Mark each state:

```text
> 8.3   installed, recommended
  8.2   installed, compatible
  8.1   available, not installed
  7.4   legacy, not installed
```

Selection rules:

1. Respect an explicit `--php` flag.
2. Respect a valid existing project pin.
3. Detect framework/composer constraints.
4. Prefer the global default if compatible.
5. Prefer an installed compatible version over an unavailable one.
6. Explain when the selected version is outside the detected compatibility range.

For Podman runtime migration, the choice should distinguish:

```text
installed image
available prebuilt image
local build required
unavailable
```

The wizard may offer to build/pull a missing version only after explicit confirmation. It must not silently begin a long PHP build.

### Step 2: Node version

This step is optional and should be shown when the project indicates JavaScript usage, such as `package.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lockb`, or a framework detection rule.

Choices:

```text
> Follow .nvmrc (20)
  System Node (24)
  Pin another version
  No Node runtime
```

If no `.nvmrc` exists:

```text
> Do not pin Node
  Use detected system Node
  Pin a version
```

The initial Chauffeur runtime plan does not currently include Node management. Therefore the first implementation may persist intent and report that Node execution remains host-provided. Do not claim containerized Node support until the runtime supports it.

### Step 3: Database

Show project-detected database requirements and existing managed containers:

```text
Database
  SQLite (no service)
> PostgreSQL (chauf-postgres)   running
  MySQL 8 (chauf-mysql8)        stopped
  MariaDB                        not created
  None
```

Selection rules:

- Detect Laravel `DB_CONNECTION` and existing `.env` values.
- Detect framework defaults and project documentation where possible.
- Prefer an existing compatible Chauffeur database container.
- Do not create a new container without confirmation.
- Explain host port, database name, username, and connection target.
- Store a logical service reference, not a password, in project config.

If the selected engine does not exist, offer:

```text
Create chauf-postgres now
Use an existing external database
Keep database configuration unchanged
```

Database credential generation and persistence remain governed by the Podman runtime/security plan.

### Step 4: Supporting services

Present available service presets and project-detected suggestions as a multi-select:

```text
Services
> [•] mailpit       suggested by mail configuration
  [ ] meilisearch   available
  [•] redis         detected in .env
  [ ] minio/rustfs  not required
```

Each service row should show:

- Selection state.
- Availability: running, stopped, not created, unavailable.
- Why it is suggested.
- Port/URL if selected.
- Whether it will be created, started, or only recorded.

The first implementation should use the existing Podman database/container abstraction where possible. A generic service preset system is a later extension and must not be invented as hidden ad-hoc YAML.

### Step 5: Web/SSL options

Keep existing link decisions in the wizard:

```text
Primary domain:  my-shop.test
Aliases:         none
HTTPS:           enabled
FPM mode:        shared
```

Choices:

- Edit project name/slug.
- Edit primary domain.
- Add aliases.
- Enable/disable HTTPS.
- Shared/dedicated FPM/runtime mode.

The default should preserve current flags and global settings. If mkcert is unavailable, clearly explain the HTTP-only fallback and remediation rather than failing after the wizard summary.

### Step 6: `.env` configuration check

Before final confirmation, show a read-only environment summary if `.env.example` exists:

```text
Environment
  .env.example keys: 42
  .env keys:         45
  Missing:           3
  Extra:             6
  Empty:             2

Review differences in the web UI after linking.
```

Do not automatically copy values from `.env.example` during link. That is an explicit later action through the environment editor.

### Step 7: Review and apply

Show a complete mutation plan:

```text
Ready to link my-shop

  URL             https://my-shop.test:8443
  PHP             8.3 (Podman image ready)
  Node            follow .nvmrc (20), host runtime
  Database        PostgreSQL → chauf-postgres:5432
  Services        mailpit, redis
  FPM             shared
  SSL             generate trusted certificate
  Files           project config + nginx route

Enter apply   Esc back   Ctrl+C cancel
```

After confirmation:

1. Save project intent/config.
2. Create or prepare selected runtime/database/service resources only with confirmed choices.
3. Generate nginx config.
4. Generate SSL if selected and available.
5. Enable the site.
6. Start/reload required services according to the selected behavior.
7. Run readiness/health checks.
8. Print URL and next actions.

If a later step fails, preserve the project config and provide a retry/remediation path. Do not silently roll back successful user choices unless the state is unsafe or impossible to represent.

## Project configuration model

The current `projects.Project` contains PHP, SSL, FPM, domain, and project type but not Node, database, or service intent. Extend it in a backward-compatible way:

```yaml
slug: my-shop
path: /home/user/Projects/my-shop
aliases: []
php_version: "8.3"
runtime:
  engine: podman
  fpm: shared
node:
  mode: nvmrc
  version: "20"
database:
  engine: postgres
  container: chauf-postgres
  database: my_shop
  external: false
services:
  - mailpit
  - redis
ssl: true
fpm:
  dedicated: false
```

Rules:

- Missing new fields load as unset/default, preserving old projects.
- Passwords and secret values never belong in project config.
- `node.mode=none` and `database.engine=none` must be representable.
- Logical service names are portable; machine-specific image/port/container data stays in workspace/runtime state.
- Save atomically and preserve a backup before schema migration.
- The schema must be shared by CLI, web UI, and future project manifests.

## Noninteractive flag contract

Every wizard choice needs a scriptable equivalent:

```bash
chauf link \
  --php 8.3 \
  --node nvmrc \
  --database postgres \
  --service mailpit \
  --service redis \
  --secure \
  --fpm shared \
  --no-interactive
```

Proposed flags:

| Flag | Meaning |
|---|---|
| `--interactive` | Force wizard even when flags are present. |
| `--no-interactive` | Never prompt; infer or fail clearly. |
| `--php VERSION` | Select PHP version. |
| `--node VERSION` | Pin Node version. |
| `--node nvmrc` | Follow `.nvmrc`. |
| `--no-node` | Do not record Node runtime intent. |
| `--database ENGINE` | Select SQLite, MySQL, PostgreSQL, MariaDB, or none. |
| `--database-container NAME` | Use an existing managed container. |
| `--service NAME` | Select a repeatable supporting service. |
| `--no-services` | Do not select supporting services. |
| `--fpm shared|dedicated` | Select FPM strategy. |
| `--secure` / `--insecure` | Select HTTPS state explicitly. |
| `--site DOMAIN` | Set primary domain. |
| `--alias DOMAIN` | Add repeatable alias. |
| `--yes` | Confirm the final plan without prompting. |

Do not overload existing flags with a new meaning. For example, `--dedicated-fpm` may remain as an alias for `--fpm dedicated` during migration.

## TUI interaction design

### Controls

- Up/down: move within the current group.
- Left/right: move between groups or tabs where applicable.
- Space: toggle a multi-select item.
- Enter: accept the current group or apply the final plan.
- Esc: go back one group.
- `?`: show controls/help.
- Ctrl+C: cancel without mutation.
- Home/End: jump within long lists.

The footer must always show context-appropriate controls:

```text
↑/↓ move   Space toggle   Enter continue   Esc back   Ctrl+C cancel
```

### Layout

Use a stable vertical wizard layout:

```text
Header: project name and step counter
Body: current group and choices
Summary: current selections/warnings
Footer: keyboard controls
```

Long lists must use the existing viewport logic and keep the cursor visible. The selector viewport tests already cover this behavior for PHP and Podman selectors; link wizard controls must reuse the same rules rather than introducing another scrolling implementation.

### Terminal fallback

When Bubble Tea cannot start or output is not interactive:

- Use deterministic numbered prompts only when interactive stdin is available.
- Otherwise require flags or return a clear error.
- Never print a prompt and wait forever in a pipe/CI environment.

## Detection and recommendation engine

Separate detection from rendering:

```go
type ProjectSetup struct {
    Project       ProjectFacts
    PHPChoices    []RuntimeChoice
    NodeChoices   []RuntimeChoice
    DatabaseChoices []DatabaseChoice
    Services      []ServiceChoice
    Warnings      []SetupWarning
}
```

Detectors should be pure/read-only where possible:

- Framework/project type.
- PHP constraints from `composer.json`.
- Node hints from `.nvmrc`, package manager files, and `package.json`.
- Database hints from `.env`/framework config.
- Service hints from `.env`, framework config, or explicit project config.
- Installed/runtime availability.
- Domain/port/SSL readiness.

Recommendations must include evidence:

```text
PHP 8.3 recommended because composer.json allows 8.2–8.3 and 8.3 is installed.
Redis suggested because CACHE_STORE=redis and QUEUE_CONNECTION=redis.
```

If evidence conflicts, show the conflict and ask the user instead of silently choosing.

## Web UI relationship

The wizard and web UI should share the same setup plan API:

```text
GET  /api/projects/setup?path=...
POST /api/projects/setup/preview
POST /api/projects/setup/apply
GET  /api/operations/{id}
```

The web UI should render the same groups:

- PHP.
- Node.
- Database.
- Supporting services.
- Domain/SSL/FPM.
- Environment summary.
- Review/apply.

The TUI must remain fully functional if the panel is not running. The web UI must not implement separate detection or recommendation rules.

## Service lifecycle behavior

The wizard needs a clear policy for selected services:

| Existing state | User selection | Action |
|---|---|---|
| Running | Selected | Keep running and record dependency. |
| Stopped | Selected | Start after confirmation or record “selected, not started” based on preference. |
| Missing | Selected | Offer creation; never create silently. |
| Running | Not selected | Do not stop; do not claim project owns it. |
| Missing | Not selected | No action. |
| Unavailable | Selected | Show blocker and alternative. |

The initial default should be conservative: selected services are recorded, and only resources explicitly confirmed in the final plan are created or started.

## Error and recovery behavior

Each failed setup step must identify:

```text
Step: Create PostgreSQL service
Problem: host port 5432 is already in use
Impact: the selected project database is not available
Options: choose another host port, use chauf-postgres, or keep external database
Retry: chauf link --database postgres --database-container chauf-postgres
```

The wizard must support:

- Backing up config before changing it.
- Retrying the failed step.
- Returning to choices when a dependency is unavailable.
- Continuing with a degraded setup only when the user accepts it.
- Final status that distinguishes linked, ready, degraded, and failed.

## Security and privacy

- Never display database passwords in the wizard summary.
- Mask secrets if environment detection finds them.
- Do not copy `.env` values into project config.
- Do not execute project-provided commands during detection.
- Treat `composer.json`, package manifests, and project setup files as untrusted input for parsing.
- Validate domains, versions, service names, ports, paths, and image references.
- Require confirmation before creating containers, writing env files, enabling SSL, or changing FPM mode.
- Ensure service selections cannot inject shell arguments.

## Phased implementation

### Phase 0: Setup model and detection

- Extend project config with optional Node/database/service/runtime intent.
- Add backward-compatible serialization and tests.
- Implement read-only project fact detection.
- Implement PHP recommendation logic using installed versions and Composer constraints.
- Define database/service choice types.

### Phase 1: PHP-only interactive wizard

- Extract current link mutation logic into a setup plan/apply flow.
- Add Bubble Tea PHP selection using existing viewport utilities.
- Add review screen and final confirmation.
- Preserve all current link flags and noninteractive behavior.
- Add cancellation/no-mutation tests.

### Phase 2: Web/domain/FPM options

- Add domain, aliases, SSL, and shared/dedicated FPM groups.
- Show DNS/mkcert/port warnings before apply.
- Add existing-project relink/edit behavior.
- Add snapshot/backup before changing project config.

### Phase 3: Node intent

- Detect `.nvmrc` and package manager files.
- Add follow/pin/none choices.
- Persist Node intent without claiming runtime support that does not exist.
- Add clear host-runtime/container-runtime status.

### Phase 4: Database selection

- Detect database requirements from env/framework files.
- List existing Chauffeur Podman containers and status.
- Add SQLite/external/managed-container choices.
- Add create-container preview and explicit confirmation.
- Persist logical database reference without credentials.

### Phase 5: Supporting services

- Define a small built-in service registry.
- Add multi-select service TUI with availability and reason labels.
- Integrate selected services with project status and web UI.
- Add conservative start/create policy.

### Phase 6: Web setup wizard

- Expose setup preview/apply APIs.
- Reuse the same detection and plan model in React.
- Add operation progress and failure recovery.
- Add CLI/UI convergence tests.

### Phase 7: Framework-specific recommendations

- Add Laravel `.env`/service recommendations.
- Add WordPress database recommendations.
- Add generic PHP fallback.
- Keep framework adapters optional and independently testable.

## Verification strategy

### Unit tests

- Project fact detection with fixtures.
- Composer PHP constraint selection.
- `.nvmrc`/package-manager detection.
- Database/env hint detection.
- Service recommendation evidence.
- Backward-compatible project config loading.
- Setup plan diff and mutation preview.
- Flag-to-plan mapping.
- Domain/port/service validation.

### TUI tests

- Cursor remains visible in clipped lists.
- Group navigation and back/cancel behavior.
- Space toggles only multi-select items.
- Enter advances and final confirmation applies once.
- Esc applies no mutations when cancelled.
- Existing project values are preselected.
- Missing runtime/service states are visually distinct.
- Non-TTY mode never starts an interactive program.

### Integration tests

- New project link with PHP only.
- Laravel project with detected PHP/database/services.
- Existing project relink preserves unspecified settings.
- Selecting an existing database does not recreate it.
- Selecting a missing database requires explicit creation confirmation.
- SSL unavailable produces a usable HTTP/degraded result.
- Podman runtime/image readiness is reflected in the final status.
- A real linked project request succeeds after apply.

### Web UI tests

- Setup preview matches TUI preview for the same fixture.
- Apply returns an operation and final state.
- Selection changes invalidate project/dashboard queries.
- Secrets remain redacted.
- Keyboard and responsive setup flows work.

## Success criteria

The link wizard succeeds when:

- A new user can configure the project’s PHP, Node intent, database, services, SSL, domain, and FPM choices in one flow.
- The wizard explains why defaults were selected.
- No resource is created or started without visible confirmation.
- Existing projects can be relinked without losing unspecified configuration.
- All choices have noninteractive CLI equivalents.
- TUI, web UI, and CLI use the same setup plan and detection rules.
- The flow works on short terminals and never traps CI/non-TTY execution.
- The final screen distinguishes linked, ready, degraded, and failed states.
- The project can answer a real HTTP request after a successful apply.

## Immediate next implementation slice

Start with a narrow PHP/domain/SSL wizard:

1. Introduce a read-only setup-plan model.
2. Move current `chauf link` mutation into an apply function.
3. Add PHP version selection with current installed/default recommendations.
4. Add review, confirm, and cancel behavior.
5. Preserve existing flags and add `--no-interactive`.
6. Add tests proving cancellation causes no config/nginx/SSL mutation.
7. Add Node, database, and services only after the PHP-only wizard is stable.
