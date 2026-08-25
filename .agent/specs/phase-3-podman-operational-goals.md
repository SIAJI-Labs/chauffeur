# Phase 3 Podman Operational Goals

## Purpose

Make the Podman PHP runtime operationally complete after image installation.
The user must be able to distinguish image state, container state, project
configuration, and readiness without stale native state or opaque Podman
errors.

This phase builds on:

- `.agent/specs/phase-1-podman-runtime-goals.md`
- `.agent/specs/phase-2-link-wizard-goals.md`
- `.agent/specs/podman.md`
- `.agent/specs/php-fpm.md`
- `.agent/specs/service-orchestration.md`

The native runtime remains supported and must not be changed when
`runtime.engine: native` is selected.

## Observed Problems

The current workflow has demonstrated these failures:

- `chauf install php <version> --build` creates an image but not a running
  container.
- `chauf status` can print repeated FPM entries for the same stable container
  because it reports once per linked project instead of once per runtime
  resource.
- Stale linked projects can request PHP images that are no longer installed,
  such as PHP 7.4 or 8.4, causing `chauf start` to fail.
- Podman status 125 errors currently omit the underlying Podman stderr and the
  next corrective action.
- `podman ps -a` does not show PHP containers after image installation, which
  is technically correct but not sufficiently clear in the CLI workflow.

## Required Behavior

### Image Installation

- `chauf install php <version>` must use the selected runtime.
- Native mode installs host PHP binaries as before.
- Podman mode explicitly pulls or locally builds the PHP image.
- Registry failures must include Podman stderr and recommend local build or
  registry authentication.
- Installation must report image reference, digest, PHP version, and parity
  state.
- Image installation must not silently create a project-scoped FPM container
  without a confirmed project mount.

### Container Preparation

- `chauf start` must prepare all linked PHP runtime resources exactly once.
- Shared FPM must have one stable container per PHP version.
- Dedicated FPM must have one stable container per project slug.
- Nginx must have one stable container per workspace.
- Repeated starts must be idempotent and must reconcile stale stopped
  containers safely.
- Container mounts, labels, network, ports, and selected image must be
  validated before reporting readiness.
- A missing image must stop preparation before partial nginx startup and must
  identify the affected project, version, image, and recovery command.

### Status and Monitoring

- `chauf status` must deduplicate resources by stable container name.
- Status must distinguish:
  - image available or missing;
  - container absent, stopped, starting, running, or unhealthy;
  - PHP readiness passed or failed;
  - linked projects using the resource.
- Native status must continue to report native services only.
- Podman status must never infer state from native PID files.
- `podman ps -a` and `chauf status` should agree about container existence and
  lifecycle state.
- Empty or missing project registrations must produce a clear no-project
  result instead of repeated empty resources.

### Stale Project Configuration

- Detection must identify linked projects whose selected PHP image is missing.
- `chauf status` must report stale runtime configuration without mutating it.
- `chauf start` must fail before creating unrelated resources when a required
  image is missing.
- The error must recommend one of:
  - `chauf install php <version>`;
  - `chauf install php <version> --build`;
  - `chauf php isolate <available-version>`;
  - `chauf unlink --site <domain> --yes` for registrations whose path was deleted;
  - relinking with an available version.
- Changing the runtime from native to Podman must not delete native PHP
  installations or silently reinterpret them as Podman images.

### Command Consistency

The selected runtime must be checked before execution or discovery in:

- `chauf install php|nginx|composer`;
- `chauf remove php|nginx|composer`;
- `chauf update php|nginx|composer|all|list`;
- `chauf php <version>`, `php list`, `php use`, and `php isolate`;
- `chauf config php`;
- `chauf doctor`;
- `chauf link`, `secure`, `unsecure`, `unlink`, and alias operations;
- `chauf start`, `stop`, `restart`, `status`, and `logs`.

No Podman path may execute a native PHP binary, native Composer shim, native
PHP-FPM service, or host nginx reload.

## Tests

Add or improve tests for:

- Podman image discovery independent of native PHP directories.
- Native image and binary state remaining separate after runtime switching.
- Stable-resource status deduplication across multiple linked projects.
- Missing-image diagnostics with captured Podman stderr and recovery commands.
- Shared FPM idempotent start and mount reconciliation.
- Dedicated FPM idempotent start and mount reconciliation.
- Nginx preparation after PHP preparation succeeds.
- Start failure cleanup and prevention of unrelated partial resources.
- Stale project PHP version detection and non-mutating diagnostics.
- PHP and Composer execution through Podman after image build.
- Native command behavior remaining unchanged.

Add an integration fixture that:

1. Starts with no Chauffeur PHP images or PHP containers.
2. Builds one PHP image explicitly.
3. Confirms the image is visible through `chauf php list`.
4. Links a real PHP fixture with confirmation.
5. Starts shared FPM and nginx.
6. Confirms exactly one shared FPM container and one nginx container.
7. Confirms `chauf status` has no duplicate resource rows.
8. Serves a real HTTP request through nginx and PHP-FPM.
9. Removes or makes unavailable the selected image and verifies the stale
   configuration error is actionable.
10. Cleans temporary containers, network resources, and test state.

## Completion Gate

Phase 3 is complete only when:

- A fresh Podman workspace can build a PHP image, link a project, start the
  correct containers, and serve a real HTTP request.
- `chauf php list` reports Podman images and never host PHP binaries when
  Podman is selected.
- `chauf status` reports each stable Podman resource once with useful state,
  evidence, and linked-project context.
- Stale PHP project configuration fails safely with an exact recovery command.
- Podman status 125 and readiness failures include captured stderr and a next
  action.
- Shared and dedicated FPM lifecycle is idempotent and correctly mounted.
- Native runtime tests and behavior remain unchanged.
- No command creates, pulls, builds, or reloads resources during detection or
  before explicit user confirmation, except explicit install/update commands.
- Full unit tests, Podman integration tests, formatting, and diff checks pass.
