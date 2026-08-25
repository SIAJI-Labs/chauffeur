# Phase 2 Link Wizard Goals

## Purpose

Turn `chauf link` into a guided, evidence-based project setup flow while preserving all existing noninteractive behavior and ensuring cancellation never mutates project, nginx, or SSL state.

The first delivery is intentionally limited to PHP, domain, SSL, and FPM. Node, database, and supporting service choices must use the same setup-plan model and may be added only after their underlying runtime/service contracts exist.

## Source Specifications

- `docs/plans/2026-08-19-overhaul-roadmap.md`
- `docs/plans/2026-08-19-link-wizard-tui.md`
- `.agent/specs/project-linking.md`
- `.agent/specs/php-fpm.md`
- `.agent/specs/ssl.md`
- `.agent/specs/cli-commands.md`
- `.agent/specs/phase-1-podman-runtime-goals.md`

## Required Outcomes

1. Project detection is read-only and separated from rendering and mutation.
2. Detection produces a setup plan with recommendations and evidence.
3. Interactive and noninteractive link flows produce equivalent project intent.
4. The wizard shows a complete mutation preview before applying changes.
5. Apply is explicit, ordered, retryable, and safe for existing projects.
6. Cancel and back navigation do not write config, nginx, certificates, or runtime state.
7. Existing `chauf link` flags continue to work in scripts and CI.
8. Non-TTY execution never waits for input.

## Non-Goals

- Do not add hidden service creation or automatic dependency installation.
- Do not copy `.env.example` values into `.env`.
- Do not execute project-provided scripts during detection.
- Do not claim Node container support before the runtime provides it.
- Do not add arbitrary custom containers or a service marketplace.
- Do not implement database/service selections before their shared models and lifecycle policies exist.

## Setup Model

Introduce shared types for detection, choices, warnings, plan changes, and apply results. Names may follow repository conventions, but the model must distinguish facts from intent and runtime state.

The setup plan must include:

- Project facts: path, slug, type, document root, framework evidence.
- PHP choices: version, availability, compatibility, recommendation evidence.
- Domain and aliases.
- SSL choice and readiness warnings.
- Shared or dedicated FPM choice.
- Existing project values and unspecified values.
- Planned file/config/nginx/SSL/runtime changes.
- Warnings, blockers, and required confirmations.

Each recommendation must include evidence, for example:

```text
PHP 8.3 recommended because composer.json allows 8.2-8.3 and 8.3 is available.
```

Suggested interface shape:

```go
type SetupPlan struct {
    Facts       ProjectFacts
    Choices     SetupChoices
    Changes     []PlannedChange
    Warnings    []SetupWarning
    Blockers    []SetupBlocker
}

func DetectProjectSetup(ctx context.Context, input DetectionInput) (SetupPlan, error)
func ApplyProjectSetup(ctx context.Context, plan SetupPlan, deps ApplyDependencies) (ApplyResult, error)
```

Detection must not write files, generate certificates, build images, create containers, reload nginx, or run project commands.

## Detection Goals

Implement pure/read-only detection for:

- Laravel, WordPress, generic PHP, and reverse-proxy project type.
- Document root.
- Existing project config and project version selection.
- PHP constraints from `composer.json` without executing Composer.
- Installed/native and Podman runtime availability.
- `.nvmrc` and package-manager hints, persisted only as intent for now.
- Existing domain, aliases, SSL, and FPM settings.
- DNS, mkcert, port, and path readiness.

Conflicting evidence must produce a warning or require explicit choice. Never silently select an incompatible PHP version.

PHP recommendation precedence:

1. Explicit `--php`.
2. Existing valid project pin.
3. Compatible installed/available runtime based on project constraints.
4. Workspace default if compatible.
5. Clear blocker if no valid choice exists.

## Apply Boundary

Move current link mutation logic out of argument parsing and wizard rendering into one apply function.

Apply must:

1. Validate the plan has no unresolved blockers.
2. Re-check critical paths and conflicts.
3. Snapshot existing project config before changes.
4. Save project intent atomically.
5. Prepare only explicitly selected runtime resources.
6. Generate nginx configuration.
7. Generate SSL only when selected and available.
8. Enable the project route.
9. Reload nginx only after configuration validation succeeds.
10. Run readiness checks.
11. Return linked, ready, degraded, or failed status with evidence and remediation.

If a later step fails, preserve valid user choices and provide retry information. Do not silently delete or revert successful work unless the resulting state is unsafe or impossible to represent.

Config writes must use a temporary file, atomic rename, and a backup for existing configuration. The backup location and retention behavior must be testable.

## Existing CLI Contract

Preserve current flags and behavior:

```bash
chauf link [path]
chauf link --php 8.3
chauf link --secure
chauf link --dedicated-fpm
chauf link --site my-shop.test
chauf link --alias admin.my-shop.test
```

Support or add:

```bash
chauf link --interactive
chauf link --no-interactive
chauf link --yes
```

Rules:

- TTY plus no setup-affecting flags may enter the wizard.
- `--no-interactive` must never prompt.
- Non-TTY execution must never prompt or hang.
- Explicit flags override recommendations.
- `--yes` confirms only the displayed plan; it must not bypass validation.
- Existing aliases and unspecified project settings must be preserved during relink.

## TUI Goals

Use the existing Bubble Tea and selector viewport utilities. Do not create a second scrolling implementation.

Initial groups:

1. PHP version.
2. Primary domain and aliases.
3. SSL.
4. FPM mode.
5. Review and apply.

Visual language:

- Green command name.
- Purple group labels.
- `>` for focused rows.
- Green selected values.
- Gray secondary or disabled values.
- Lowercase contextual footer controls.
- Markers must communicate state without color.

Required controls:

- Up/down: move.
- Enter: continue or apply.
- Esc: go back; on the first step, cancel.
- Ctrl+C: cancel without mutation.
- `?`: show the complete keyboard legend.
- Home/End: move within long lists.

The UI must remain usable on short terminals. Keep the focused row visible and make unavailable choices explainable.

## Review Screen

The review screen must show:

- Project path, slug, type, and document root.
- Primary URL and aliases.
- PHP version and runtime readiness.
- SSL state and certificate action.
- Shared or dedicated FPM mode.
- Files and resources that will change.
- Warnings and degraded outcomes.
- Explicit apply and cancel controls.

It must clearly distinguish:

- No mutation yet.
- Changes to be written.
- Resources that will be created or started.
- Resources that will remain untouched.

No hidden image pull, build, container creation, certificate generation, or nginx reload may happen before confirmation.

## Project Configuration

Maintain backward-compatible loading for existing flat project configs. New intent fields must:

- Have safe defaults when absent.
- Exclude passwords and secret values.
- Represent no Node runtime and no database.
- Store logical service references rather than machine-specific runtime details.
- Be serialized atomically with a schema/version migration path.

If the existing model remains flat during the first slice, document the boundary and avoid introducing partially duplicated nested and flat representations.

## Security and Validation

Validate before apply:

- Project path is an allowed directory.
- Slug and domains are safe and collision-free.
- PHP versions are normalized.
- Aliases are valid and unique.
- Ports are within range and available where required.
- Runtime/image references cannot inject arguments.

Never display passwords or secret values. Parse manifests and environment files as untrusted data. Detection must not execute arbitrary project code.

## Tests and Verification

Add unit tests for:

- Project facts and document-root detection.
- PHP constraint parsing and recommendation precedence.
- Existing config defaults and relink preservation.
- Setup-plan diff generation.
- Flag-to-plan mapping.
- Domain, alias, path, port, and SSL validation.
- Atomic save and backup behavior.

Add TUI tests for:

- Initial selection and existing-value preselection.
- Cursor visibility in clipped lists.
- Enter advances exactly once.
- Esc back behavior.
- Ctrl+C cancellation.
- Cancel produces no filesystem mutation.
- Final confirmation applies once.
- Non-TTY mode never starts an interactive program.

Add integration tests for:

- PHP-only new project link.
- Existing project relink preserving unspecified settings.
- HTTPS fallback when mkcert is unavailable.
- Invalid or unavailable PHP choice.
- Nginx generation and reload after apply.
- Real fixture HTTP request after successful apply.

## Completion Gate

Phase 2 is complete only when:

- Existing `chauf link` flags work unchanged.
- Interactive and noninteractive flows produce equivalent intent.
- Detection is read-only.
- Review is shown before mutation.
- Cancel/back never changes project, nginx, or SSL state.
- Relink preserves unspecified values.
- CI and pipes cannot hang.
- Final output distinguishes linked, ready, degraded, and failed.
- A real linked fixture answers an HTTP request using the selected runtime.

## Deferred Expansion

After the PHP/domain/SSL/FPM slice is stable, add in this order:

1. Node intent and `.nvmrc` handling.
2. Database detection and existing managed-container selection.
3. Supporting service multi-select.
4. Explicit create/start confirmation for missing resources.
5. Shared setup-plan APIs for the web UI.
