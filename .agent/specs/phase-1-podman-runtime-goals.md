# Phase 1 Podman Runtime Goals

## Purpose

Move PHP execution from host-built binaries and native PHP-FPM processes to a reproducible, rootless Podman runtime without changing the user-facing Chauffeur workflow.

The existing native runtime remains available as a migration fallback until the Podman path passes all release gates.

## Source Specifications

- `docs/plans/2026-08-19-overhaul-roadmap.md`
- `docs/plans/2026-08-19-podman-runtime-overhaul.md`
- `.agent/specs/podman.md`
- `.agent/specs/php-fpm.md`
- `.agent/specs/service-orchestration.md`
- `.agent/specs/workspace.md`

## Required Outcomes

1. A runtime abstraction owns PHP, FPM, nginx, lifecycle, status, logs, and command execution decisions.
2. The current native implementation is behind that abstraction and remains recoverable.
3. Podman commands are argument-safe, fakeable, and always scoped to the configured workspace.
4. Rootless Podman prerequisites and failures produce actionable diagnostics.
5. A minimal PHP 8.3 image can be pulled or built locally and used to execute PHP.
6. PHP and Composer commands execute in the project-selected runtime with correct exit codes.
7. Shared and dedicated FPM semantics remain distinguishable and project-aware.
8. A real fixture project returns a successful HTTP response through nginx and the selected PHP-FPM runtime.

## Non-Goals

- Do not remove the native runtime in this phase.
- Do not redesign the CLI command vocabulary.
- Do not add arbitrary custom service definitions or an extension marketplace.
- Do not migrate databases, DNS, or SSL ownership unless required by the PHP/FPM request path.
- Do not silently build images, create containers, or change the default runtime mode.

## Runtime Contract

Create `internal/runtime` with intent-oriented operations. The exact types may follow existing repository conventions, but the contract must cover:

```go
type Runtime interface {
    Ensure(ctx context.Context, version string) error
    Start(ctx context.Context, scope Scope) error
    Stop(ctx context.Context, scope Scope) error
    Restart(ctx context.Context, scope Scope) error
    Status(ctx context.Context, scope Scope) ([]ServiceStatus, error)
    Logs(ctx context.Context, target LogTarget, opts LogOptions) error
    Exec(ctx context.Context, scope Scope, command []string, opts ExecOptions) error
}
```

The contract must represent:

- Runtime engine: `native` or `podman`.
- Shared versus dedicated FPM scope.
- Stable service identity.
- Human-readable and machine-readable state.
- Image reference, digest, and build fingerprint where applicable.
- Readiness and health evidence.
- Error code, impact, and next action.

Command handlers must express intent and formatting only. They must not contain Podman-specific orchestration or duplicate native/Podman logic.

## Podman Command Runner

Add a fakeable command backend that:

- Executes Podman through argument arrays, never shell strings.
- Captures stdout, stderr, exit status, and duration.
- Accepts context cancellation and timeouts.
- Does not inherit stdin unless explicitly requested.
- Distinguishes executable-not-found, prerequisite, command, timeout, and readiness failures.
- Allows unit tests to assert exact arguments without requiring Podman.

The backend must support at least:

- `podman version` and `podman info`.
- `podman network exists/create`.
- `podman image exists/inspect/pull/build`.
- `podman container exists/create/start/stop/restart/rm/inspect/logs/exec`.

## Workspace and Naming

All runtime state must resolve through the existing workspace-root abstraction. No runtime or Podman package may independently read `HOME` or construct a second root.

Use stable, validated names and labels:

- Network: `chauf-net`.
- Shared FPM: `chauf-php<major><minor>-fpm`.
- Dedicated FPM: `chauf-cfpm-<slug>`.
- Workspace nginx: `chauf-nginx`.

Container labels must identify Chauffeur ownership, workspace, runtime role, PHP version, and project slug where relevant.

Reject invalid slugs, versions, image references, paths, ports, and names before invoking Podman.

## Preflight and Diagnostics

Before any mutating Podman operation, validate:

- Podman is installed.
- Rootless mode is active or explicitly explain why it is not.
- User namespaces and storage are usable.
- `podman info` succeeds.
- `chauf-net` is available or can be created.
- Required ports and mounts are usable.
- The selected image is available or a local build is possible.

Errors must identify problem, impact, and next action. Example:

```text
Problem: rootless Podman is unavailable.
Impact: PHP 8.3 cannot start in the selected runtime.
Next: enable rootless user namespaces or use the native runtime while troubleshooting.
```

## PHP Image Goals

Start with a minimal PHP 8.3 FPM image and preserve the standard extension baseline required by current Chauffeur projects. The image metadata must record:

- PHP version and patch version.
- Base image/distribution.
- Extension set.
- Containerfile or recipe fingerprint.
- Architecture.
- Image digest.

Support two explicit paths:

1. Pull a supported prebuilt image.
2. Build locally when explicitly requested or when no prebuilt image exists.

Do not report an image as ready until the intended state is verified:

```text
declared -> build/pull attempted -> loaded -> available to project
```

The first parity fixture must verify:

- `php -v`.
- A PHP script execution.
- Standard extensions, including GD, OpenSSL, XML, mbstring, curl, zip, bcmath, sodium, MySQL, and PostgreSQL support where available.
- Exit-code propagation.

PHP 7.4 and 8.0 compatibility must be encoded as a follow-up parity target, not assumed to work because the runtime is containerized.

## PHP and Composer Execution

Project-aware resolution must select the project version by this precedence:

1. Explicit command override.
2. Project runtime intent.
3. Project version file if supported by current behavior.
4. Workspace default.

`chauf php`, project PHP commands, and Composer must execute in the same selected runtime. Preserve current explicit version syntax during migration.

The implementation must preserve:

- Working directory.
- Project mounts.
- Environment behavior without exposing secrets in output.
- User-visible stdout/stderr.
- Process exit code.

## FPM Goals

Implement Podman equivalents of the current FPM strategies:

- Shared: one container per PHP major/minor version.
- Dedicated: one project-specific container per dedicated project.

Shared and dedicated containers must use stable network endpoints rather than host Unix sockets unless a proven compatibility requirement exists.

Project routing must preserve PHP version, slug, domains, aliases, and dedicated isolation. A shared container must safely serve multiple mounted project roots without leaking project configuration.

Readiness must be checked before nginx is started or reloaded.

## Nginx Migration Boundary

Keep host/native nginx temporarily if needed while PHP and FPM parity is proven. Move nginx into `chauf-nginx` only when:

- FPM containers are healthy.
- Generated routes work over the Podman network.
- Existing domains and aliases remain functional.
- Certificates remain read-only mounts.
- Existing HTTP/HTTPS ports and reload behavior remain compatible.

Do not combine nginx template redesign with the initial PHP runtime migration.

## Tests and Verification

Add tests for:

- Runtime interface behavior through a native fake.
- Exact Podman argument construction.
- Rootless preflight success and failure.
- Workspace override resolution.
- Version normalization and image selection.
- Container names, labels, mounts, and ports.
- Image digest/fingerprint registration.
- PHP and Composer exit-code propagation.
- Readiness timeout and actionable failure output.
- Shared/dedicated scope resolution.

Add an integration test, skipped with a clear reason when Podman is unavailable, that:

1. Creates or reuses an isolated test workspace.
2. Creates `chauf-net`.
3. Starts PHP 8.3 FPM.
4. Executes `php -v` and a fixture script.
5. Validates the fixture through the web request path when nginx support is enabled.
6. Cleans up containers, network, and temporary state.

## Completion Gate

Phase 1 is complete only when:

- Existing native tests still pass.
- Native state remains recoverable.
- `chauf php -v` and Composer work through the selected Podman runtime.
- PHP 8.3 FPM works in shared and dedicated modes.
- PHP 7.4, 8.0, 8.3, and the newest supported version have an explicit verified status.
- A real linked fixture answers a real HTTP request through nginx and PHP-FPM.
- No default workflow requires privileged containers.
- Failures include evidence and a next action.

## Recommended Implementation Order

1. Add runtime types and native adapter.
2. Add command runner interface and Podman fake.
3. Centralize workspace paths.
4. Add rootless preflight and network management.
5. Add PHP 8.3 image metadata and execution.
6. Add shared FPM.
7. Add dedicated FPM.
8. Integrate CLI PHP and Composer execution.
9. Add request-path fixture verification.
10. Move nginx only after the preceding gates pass.
