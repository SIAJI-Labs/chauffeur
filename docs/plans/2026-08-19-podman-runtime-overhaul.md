# Podman Runtime Overhaul Plan

This plan is Phase 1 of the [master overhaul roadmap](./2026-08-19-overhaul-roadmap.md). It depends on the shared runtime/state contracts from Phase 0 and provides the runtime required by the service, link wizard, web UI, and observability plans.

## Goal

Replace Chauffeur's native nginx/PHP runtime with a rootless Podman-based runtime while preserving the existing user-facing functionality and project workflow.

The target experience remains:

```bash
chauf init
chauf install nginx php 8.3 composer
chauf link --secure
chauf start
```

The implementation behind those commands changes from host processes and host-built PHP binaries to managed containers and reproducible PHP images.

## Why this change

The current native model is lightweight and transparent, but PHP compilation and extension compatibility are coupled to the host toolchain. PHP 7.4 and 8.0 already require vendored OpenSSL and source patches, including GD/compiler compatibility fixes. Every supported distribution, compiler, and system-library combination can create another build failure.

Podman provides a controlled build/runtime boundary:

- PHP build dependencies are defined once in a Containerfile.
- Common PHP versions can be distributed as prebuilt images.
- Legacy PHP compatibility is fixed centrally instead of on every developer machine.
- PHP extensions and system packages can become declarative, persistent configuration.
- Project environments become more reproducible across machines.

## Non-goals

- Do not redesign the CLI command vocabulary in the first migration.
- Do not remove project linking, aliases, SSL, DNS, Composer, FPM strategies, logs, status, doctor, autostart, or database workflows.
- Do not add a broad service marketplace, worker platform, public tunneling, MCP, or multi-platform support as part of the runtime conversion.
- Do not require Docker compatibility; Podman remains the supported container engine.
- Do not silently delete or migrate existing native installations.

## Target architecture

```text
Browser
  → *.test / *.localhost
  → host port 8080 / 8443
  → chauf-nginx container
  → Podman network: chauf-net
  → PHP-FPM container for the selected PHP version
  → bind-mounted project directory
```

### Core containers

| Container | Responsibility | Lifetime |
|---|---|---|
| `chauf-nginx` | HTTP/HTTPS termination and virtual-host routing | Workspace-wide |
| `chauf-php<version>-fpm` | Shared PHP-FPM runtime for a PHP version | One per installed/active version |
| `chauf-cfpm-<slug>` | Optional dedicated project PHP-FPM runtime | One per dedicated project |
| `chauf-<engine>` | Database containers | Existing optional workflow |
| `chauf-dns` | Optional containerized DNS integration, only if selected later | Deferred decision |

The first implementation should keep host dnsmasq and mkcert integration where that reduces migration risk. DNS and certificate ownership are separate concerns from PHP runtime conversion.

### Runtime modes during migration

The runtime needs an explicit mode while the migration is underway:

```yaml
runtime:
  engine: podman # native | podman during transition
```

The default should switch to Podman only after the Podman path passes the release gates. Native mode is a temporary migration fallback, not a permanent second product unless maintaining both is explicitly accepted.

## Functionality preservation map

| Existing functionality | Podman implementation |
|---|---|
| `chauf install php <version>` | Pull/register a prebuilt image; optionally build locally with `--local`. |
| PHP 7.4–8.5 selection | Normalize version and select a tagged image/Containerfile recipe. |
| `chauf php use` | Update global default and ensure the image exists. |
| `chauf php isolate` | Update project/worktree runtime intent and relink service state. |
| `chauf php <version> ...` | Execute `podman exec` in the selected FPM/container runtime. |
| PHP shim | Resolve project version and execute through the Podman runtime. |
| Composer shim | Execute Composer in the same PHP container/version as the project. |
| Shared FPM | One FPM container per PHP version, shared by linked projects. |
| Dedicated FPM | One project-specific FPM container or image, with the same project mounts. |
| `chauf start/stop/restart` | Idempotent Podman lifecycle operations; nginx starts after FPM readiness. |
| `chauf status` | Container health, image, PID/process state, uptime, ports, sockets, and project usage. |
| `chauf logs` | `podman logs` for containers plus mounted nginx/application log files. |
| Project nginx config | Generate config for host-mounted or container-local nginx config, then reload container. |
| SSL | Keep mkcert certificates in workspace and mount them read-only into nginx. |
| DNS | Keep existing dnsmasq/host integration initially. |
| `chauf link --secure` | Generate project config, certificate, nginx route, and ensure required runtime. |
| `chauf podman ...` | Reuse existing database abstraction and network where possible. |
| Admin panel | Extend the existing API to represent PHP/nginx containers only after core CLI parity. |
| Autostart | Generate systemd user units for Podman/Quadlet services rather than native binaries. |
| `chauf doctor` | Add Podman, rootless, network, image, mount, health, and permission checks. |

## PHP image strategy

### Image sources

Use a two-tier strategy:

1. **Prebuilt images:** default for supported versions and architectures.
2. **Local builds:** explicit fallback for unavailable versions, custom extensions, development, or offline/internal registries.

Suggested tags:

```text
ghcr.io/<owner>/chauffeur-php:7.4-fpm
ghcr.io/<owner>/chauffeur-php:8.0-fpm
ghcr.io/<owner>/chauffeur-php:8.3-fpm
ghcr.io/<owner>/chauffeur-php:8.5-fpm
```

The exact registry is a release decision. The image must be pinned by digest in production-like workflows, while a friendly version tag can remain the user-facing default.

### Containerfile requirements

Each image recipe must define:

- PHP version and patch version.
- Base distribution/version.
- Compiler and system package dependencies.
- Standard extensions.
- FPM configuration.
- Non-root runtime user/group behavior.
- CA certificate injection strategy.
- Build metadata and recipe fingerprint.
- Architecture support.

The standard extension set should preserve the current native build’s baseline, including MySQL, PostgreSQL, GD, mbstring, OpenSSL, XML, curl, zip, GMP, bcmath, sodium, and related extensions already expected by Chauffeur.

### Legacy PHP policy

PHP 7.4 and 8.0 require an explicit compatibility policy. The image build must preserve the current native workarounds where needed:

- OpenSSL 1.1 compatibility.
- PHP source patches for modern compiler behavior.
- GD compatibility and supported image libraries.
- libxml compatibility.
- scanf/function-pointer compatibility.
- Legacy extension version pins.

Do not assume that moving into a container makes legacy builds automatically safe. It makes the build environment controllable; the compatibility work still needs to be encoded and tested.

### Extension model

Add persistent, declarative extension/package configuration:

```yaml
php:
  extensions:
    - redis
    - imagick
  packages:
    - imagemagick
```

Proposed commands:

```bash
chauf php ext list
chauf php ext add redis
chauf php ext remove redis
chauf php pkg list
chauf php pkg add imagemagick
chauf php rebuild 8.3
```

These are new commands and should not block the runtime migration. The migration must first preserve the existing fixed extension set. Extension management becomes the first post-parity capability.

Every build must report four states separately:

```text
declared → build attempted → loaded → available to project
```

An extension failure must not be silently reported as success.

## Nginx container strategy

Nginx should move into a workspace-wide container so the full web request path is consistent:

```text
host port 8080/8443
  → chauf-nginx
  → chauf-net
  → chauf-php83-fpm:9000
```

Project files and generated nginx config need mounts with clear ownership:

- Project roots: read-only by default where possible, writable when the app requires it.
- Nginx configs: generated workspace state, mounted read-only or copied through a controlled reload operation.
- Certificates: read-only.
- Logs: named volume or workspace bind mount.

The first migration should retain generated per-project nginx configuration semantics and domain/alias behavior. Changing template behavior at the same time makes failures difficult to attribute.

## Shared and dedicated FPM semantics

### Shared mode

The shared mode becomes one container per PHP major/minor version. Each project’s nginx route uses the same FPM container, with the project path mounted at a stable location.

The container must support multiple project roots without leaking one project’s configuration into another. Project-specific runtime values should be passed through generated pool/config fragments or request parameters only where safe.

### Dedicated mode

Dedicated mode becomes a project-specific FPM container. It may initially use the same base PHP image as shared mode and a project-specific config/mount set. A custom project image is a later capability.

Dedicated mode must preserve:

- Project-specific socket/endpoint identity.
- `chauf start --project` and related scoped operations.
- Project-specific logs.
- Removal/restart isolation.
- SSL/domain behavior.

Do not map every current native FPM socket one-to-one into the host. Prefer stable Podman network names and internal ports, with host sockets only if a compatibility constraint proves necessary.

## CLI contract during migration

The first implementation should preserve the current short commands:

```bash
chauf start
chauf stop
chauf restart
chauf status
chauf logs
chauf php list
chauf php use 8.3
chauf php isolate 8.2
chauf link --secure
```

Improve execution ergonomics separately:

```bash
chauf php artisan migrate
chauf composer install
```

Project-aware version resolution should be the default. An explicit override remains available:

```bash
chauf php --version 8.3 artisan migrate
```

The existing `chauf php <version> <command>` form should remain during migration if compatibility requires it, but should be documented as an explicit version override rather than the primary everyday workflow.

## Data and configuration migration

### Existing state to preserve

- `~/.chauffeur/config/chauffeur.yaml`.
- `~/.chauffeur/projects/<slug>/config.yaml`.
- Domains and aliases.
- SSL certificate files.
- PHP default/project versions.
- Nginx project intent.
- Podman database config, volumes, and backups.
- Composer installation/cache.

### New state

```text
~/.chauffeur/
├── images/                  # image metadata, digests, build fingerprints
├── podman/
│   ├── network metadata
│   ├── quadlets/             # generated service definitions
│   ├── containers/           # runtime metadata
│   └── existing databases
├── nginx/
│   ├── configs and certs
│   └── logs
└── projects/<slug>/
    ├── config.yaml
    ├── runtime.yaml          # optional generated runtime state
    └── logs/
```

All path resolution must go through one workspace-root abstraction. The existing Podman package currently risks bypassing `CHAUFFEUR_HOME`; fix that before migration.

### Migration behavior

Migration must be explicit and reversible:

```bash
chauf runtime status
chauf runtime migrate --to podman
chauf runtime migrate --dry-run --to podman
chauf runtime rollback --to native
```

If a command is not implemented yet, the plan must not expose it as shipped functionality. During development, a configuration flag or test-only mode is sufficient.

The migration must:

1. Validate Podman and rootless prerequisites.
2. Check image availability/build requirements.
3. Check host ports and network names.
4. Stop or isolate native services without deleting them.
5. Generate Podman service definitions.
6. Start FPM and wait for health/readiness.
7. Start nginx and test configuration.
8. Test a real linked project request.
9. Record the active runtime mode and image fingerprints.
10. Leave native state recoverable until explicitly cleaned.

## Service orchestration changes

Replace direct process assumptions in `internal/services` with a runtime abstraction. The abstraction should expose operations such as:

```go
type Runtime interface {
    Ensure(version string) error
    Start(scope Scope) error
    Stop(scope Scope) error
    Restart(scope Scope) error
    Status(scope Scope) ([]ServiceStatus, error)
    Logs(target LogTarget, opts LogOptions) error
    Exec(scope Scope, command []string, opts ExecOptions) error
}
```

Do not create parallel native and Podman implementations with duplicated command logic. Keep command handlers responsible for intent and output; keep runtime-specific behavior inside the runtime package.

The existing native implementation should remain behind the abstraction until Podman parity is verified. Then decide whether it is retained as a supported fallback or removed in a separate change.

## Doctor and error model

Add Podman-specific checks:

- Podman binary and minimum version.
- Rootless mode and user namespace support.
- `podman info` health.
- Network existence and subnet conflicts.
- Port mappings and conflicts.
- Image/tag/digest availability.
- Mount visibility and permissions.
- Container health/readiness.
- systemd/Quadlet capability for autostart.
- macOS/WSL/remote Podman constraints only if those platforms are later supported.

Errors should use the pattern:

```text
Problem: PHP 7.4 image is unavailable and local build prerequisites are missing.
Impact: project example.test cannot start.
Next: chauf php rebuild 7.4 --local, or choose an installed PHP version with chauf php isolate 8.3.
```

## Security requirements

- Use rootless Podman by default.
- Never run the PHP/nginx containers as privileged unless an explicit, documented exception exists.
- Mount certificates read-only.
- Validate all generated container names, image references, paths, and ports.
- Do not interpolate user input into shell command strings.
- Use argument arrays for Podman execution.
- Keep project mounts scoped to the project and required shared paths.
- Do not expose database passwords through the panel by default.
- Add panel authentication/token protection before expanding panel control to PHP/nginx.
- Treat project `Containerfile` or extension build instructions as executable code and require explicit consent where appropriate.

## Phased implementation

### Phase 0: Baseline and decision lock

- Record current native behavior and supported commands.
- Define the Podman image naming/tagging policy.
- Define supported Podman versions and rootless requirements.
- Decide whether nginx moves in the same phase as PHP or immediately after PHP parity.
- Add the runtime mode abstraction without changing behavior.

### Phase 1: Podman runtime foundation

- Add a dedicated runtime package for network/container/image operations.
- Centralize workspace-root resolution.
- Create `chauf-net` idempotently.
- Add container naming and label conventions.
- Add health/readiness checks.
- Add argument-safe Podman command execution.
- Add test fakes for Podman commands.

### Phase 2: PHP image parity

- Encode the current standard PHP extension set in a Containerfile.
- Encode PHP 7.4/8.0 compatibility behavior.
- Support prebuilt image pull and local build fallback.
- Register installed versions by image digest/fingerprint.
- Implement `php list/use/isolate` against image/runtime state.
- Implement `chauf php` and Composer execution through `podman exec`.
- Verify CLI output and exit-code propagation.

### Phase 3: FPM and project routing

- Implement shared PHP-FPM containers.
- Implement dedicated project FPM containers.
- Preserve project PHP selection and domain/alias routing.
- Keep nginx host/native temporarily if needed to reduce migration surface.
- Add project request integration tests.

### Phase 4: Nginx and full web path

- Build/pull the nginx image or define a minimal supported nginx Containerfile.
- Mount generated site configs and certificates.
- Route nginx to FPM containers on `chauf-net`.
- Preserve HTTP/HTTPS ports and reload semantics.
- Move logs/status/diagnostics to container-aware implementations.

### Phase 5: Lifecycle, autostart, and operations

- Convert `start/stop/restart/status/logs` fully to the runtime abstraction.
- Generate Quadlet/systemd user units where supported.
- Preserve project-scoped operations.
- Add migration status and rollback behavior.
- Update `doctor`, `info`, and `uninstall`.

### Phase 6: Extension flexibility

- Add persistent extension/package declarations.
- Add `php ext` and `php pkg` commands.
- Add rebuild progress and build fingerprinting.
- Verify extensions with `php -m`/version-specific checks.
- Add GD/Imagick/Redis/Xdebug compatibility tests.

### Phase 7: Panel parity and cleanup

- Decide whether the panel is database-focused or full-runtime.
- Replace hard-coded dashboard values.
- Add containerized PHP/nginx status only after CLI data is reliable.
- Implement real log streaming or label logs as snapshots.
- Remove native runtime code only after the deprecation window and migration tests.

## Verification gates

The migration cannot be considered complete because containers start successfully. Each phase needs these gates:

### Unit and contract tests

- Runtime command argument construction.
- Image/version normalization.
- Container names and labels.
- Mount and port generation.
- Config migration and rollback.
- Extension declarations and fingerprints.
- Exit-code/error propagation.
- CLI help/output contract.

### Integration tests

- Create rootless network.
- Start PHP-FPM container.
- Execute `php -v` and a project script.
- Execute Composer in the selected project runtime.
- Start nginx and validate generated configuration.
- Link a Laravel/generic fixture.
- Request the fixture through HTTP.
- Request it through HTTPS when mkcert is available.
- Switch project PHP version and repeat the request.
- Test shared and dedicated FPM behavior.
- Test logs and scoped restart.
- Test Podman database connectivity from the project runtime.

### Legacy compatibility tests

At minimum test PHP 7.4, 8.0, 8.3, and the newest supported version across supported architectures. For PHP 7.4 verify:

- Image/build completes.
- GD loads.
- JPEG and FreeType support work.
- OpenSSL loads.
- XML loads.
- Composer can run.
- A real application request returns successfully.

### Release gate

Use the outcome rule:

> A runtime phase is incomplete until a real linked fixture site answers a real HTTP request successfully through nginx and the selected PHP-FPM container.

## Rollback criteria

Stop the migration and retain native runtime if any of these remain unresolved:

- PHP 7.4/GD cannot be built or pulled reliably.
- Project files require unsafe or undocumented mounts.
- Composer or common framework commands behave differently without a clear migration path.
- Shared/dedicated FPM semantics cannot be preserved.
- SSL/DNS behavior regresses.
- `chauf start/status/logs/doctor` cannot explain container failures.
- Podman requires privileged operation for the default workflow.
- Existing database containers or backups are endangered.

## Success criteria

The conversion succeeds when:

- Existing documented core workflows work without a new mental model.
- PHP versions are reproducible through image metadata/digests.
- PHP 7.4 and GD are more reliable than the native build path.
- Adding an extension/package does not require changing Chauffeur source code.
- CLI commands consistently execute inside the project’s selected runtime.
- Nginx, FPM, SSL, DNS, Composer, logs, status, and autostart retain equivalent behavior.
- A real project request passes in shared and dedicated modes.
- Native runtime state is recoverable during migration and removed only through an explicit cleanup step.

## Immediate next implementation slice

The first coding slice should be intentionally small:

1. Add `internal/runtime` interfaces and a native adapter around current service behavior.
2. Add a fakeable Podman command runner and rootless preflight checks.
3. Generate one `chauf-php83-fpm` container from a minimal Containerfile.
4. Execute `chauf php -v` and one PHP fixture script inside it.
5. Add one integration test that starts the container, executes PHP, and cleans up.
6. Do not change nginx, project linking, or default runtime selection until this slice is stable.

This keeps the first change reversible and proves the core runtime boundary before migrating the entire product.
