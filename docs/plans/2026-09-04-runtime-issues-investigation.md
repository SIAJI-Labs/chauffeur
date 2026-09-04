# Runtime and Podman Issues Investigation

Date: 2026-09-04

Commit inspected: `808a063 feat(runtime): complete Podman service parity`

## Executive summary

All four reported issues are valid, with different implementation states:

| # | Reported issue | Finding | Current state |
|---|---|---|---|
| 1 | Rename the database command and reserve `podman` for Podman operations | Valid product/CLI issue | The database rename is partly implemented; generic `chauf podman` operations do not exist yet |
| 2 | Make auto-start work for either runtime and start Podman first | Valid lifecycle issue | Runtime-aware code exists, but Podman readiness and ownership are incomplete |
| 3 | Keep using `localhost` for database connections under Podman | Valid compatibility issue | Host-side DSNs use `localhost`, but container execution rewrites it to `host.containers.internal` and can still fail |
| 4 | Remove the PHP native `imagick` startup warning | Valid runtime/install issue | `imagick.ini` is enabled without proving that the module and its shared libraries are loadable |

The investigation found additional related issues. They are listed separately so
the implementation can fix the underlying lifecycle and networking contracts
without silently expanding the first patch.

## Evidence and findings

Evidence levels used below:

- **Source-confirmed**: demonstrated directly by repository code.
- **Test-confirmed**: covered by an existing automated test.
- **Runtime hypothesis**: likely explanation that requires reproduction on the affected host.

### 1. Ambiguous `chauf podman` command

**Confirmed valid.** The current command name is overloaded conceptually:

- `cmd/chauf/main.go` dispatches `podman-db` to `RunPodmanDB`.
- The old `podman` name dispatches to `RunPodmanLegacy`, which only warns and
  then runs the same database/cache operations.
- `internal/commands/podman.go` manages MySQL, PostgreSQL, MariaDB, MongoDB,
  and Redis containers; it does not provide a general Podman inspection or
  container-management command.
- `chauf config runtime podman` selects the PHP/nginx runtime and is a
  separate concern.

The requested rename is therefore only partially complete in the inspected
commit. `chauf podman-db` is already the canonical database command, and the
legacy alias is already present. However, `chauf podman ps` cannot yet list
Chauffeur-related runtime containers, so the generic `podman` namespace is not
available for its intended purpose.

#### Fix plan

1. Keep `chauf podman-db` as the database/cache command and preserve the old
   `chauf podman` database behavior for a documented deprecation window.
2. Introduce a separate generic `RunPodman` dispatcher for safe Chauffeur-scoped
   operations, starting with `ps/list`, `inspect`, `logs`, and `exec` as needed.
3. Scope generic results to Chauffeur labels/names and clearly distinguish
   runtime containers from database containers.
4. Preserve the existing `podman/` state directory, names, volumes, backups,
   and database configuration; this is a CLI rename, not a storage migration.
5. Update help and command documentation, including all implemented
   `podman-db` subcommands (`import`, `backup`, and `restore`).
6. Add dispatcher, help, compatibility, and scope tests.

#### Acceptance criteria

- `chauf podman-db ...` continues to manage database/cache containers.
- `chauf podman ...` provides Chauffeur-related Podman operations rather than
  silently meaning database management after the compatibility window.
- Existing scripts using the legacy database command receive a clear migration
  message and continue to work during the compatibility window.

### 2. Runtime-aware auto-start

**Confirmed valid; partially addressed but not complete.** `autostartEnable`
and `autostartDisable` now inspect `workspace.Load().Runtime.Engine`, and
Podman has a generated user unit. This is progress over native-only behavior,
but it does not yet guarantee the requested behavior:

- `enablePodmanAutostart` writes a unit whose `ExecStart` invokes `chauf start`.
  It does not explicitly ensure that Podman itself is ready first.
- `runtime.Podman.Preflight` checks `podman version` and `podman info`, but it
  does not start a Podman machine/service when the host requires one.
- The Podman FPM unit directly runs `podman start`, while the current enable
  path uses one aggregate nginx unit. The ownership model is inconsistent.
- Native/Podman unit migration and stale-unit discovery are not fully
  transactional. Some disable helpers only inspect currently installed local
  PHP versions, so old enabled units can remain behind.
- `config runtime` calls migration helpers that return no error, so a failed
  disable can be reported only as a warning while the runtime selection has
  already changed.

#### Fix plan

1. Define one runtime lifecycle owner for auto-start. It must resolve the
   selected runtime at execution time and never start the other runtime.
2. Add a Podman readiness operation before reconciliation/start. It should
   detect the platform-specific requirement (daemonless rootless Podman,
   Podman Machine, or service) and start/wait for it when supported.
3. Keep the required order: Podman/network readiness, PHP-FPM readiness, then
   nginx. Do not report auto-start success until required services are ready.
4. Discover and disable stale native and Podman units independently of whether
   their current installation directories still exist.
5. Make runtime switching report and handle migration failures, preserving the
   previous unit state when the migration cannot complete.
6. Ensure generated units use the resolved `CHAUFFEUR_HOME` and an explicit,
   verified `chauf` executable path.
7. Add tests for native and Podman unit generation, Podman-not-ready startup,
   FPM-before-nginx ordering, runtime switching, stale units, and custom home.

#### Acceptance criteria

- Native selection starts native nginx/PHP-FPM only.
- Podman selection makes Podman ready first, then starts/reconciles PHP-FPM and
  nginx without native port owners.
- Repeating auto-start is idempotent and reports actual state.
- Switching runtimes cannot leave both runtimes enabled for the same ports.

### 3. Database connections while using Podman

**Confirmed valid as a compatibility requirement.** The repository contains
two different connection contexts:

- `internal/podman/dsn.go` intentionally presents published database ports as
  `localhost:<port>`, which is correct for applications running on the host.
- PHP commands and generated Podman nginx configuration detect
  `DB_HOST=localhost`/`127.0.0.1` and rewrite it to
  `host.containers.internal` (`internal/commands/php_cmd.go` and
  `internal/runtime/nginx_config.go`). Inside a container, `localhost` means
  that PHP container, not the host where the published database port exists.

This translation explains why the current code does not literally preserve
`localhost` inside the application runtime. It may also fail when the Podman
host gateway is unavailable, when the database is bound only to an unexpected
interface, or when the environment is injected for CLI commands but not for
web/FPM requests. The repository has unit coverage for the rewrite, but no
complete application-level test proving a real database connection through
both CLI and nginx/FPM.

#### Fix plan

1. Define the compatibility contract explicitly: project `.env` and displayed
   DSNs remain `localhost`; any container-only translation must be internal,
   documented, and must not rewrite the project file.
2. Centralize environment translation instead of maintaining separate logic in
   the PHP shim, Composer shim, PHP command, and nginx renderer.
3. Ensure the selected database container publishes its configured port on a
   host interface reachable through the chosen Podman network.
4. Probe the actual path from the PHP container and from the FPM request path,
   including DNS/gateway availability and TCP readiness. Return an actionable
   diagnostic when the path is unavailable.
5. Test the supported design against MySQL/PostgreSQL fixtures for both
   `chauf php ...` and a real web request. Decide explicitly whether a
   localhost-compatible network mode/proxy is required; do not switch to host
   networking without checking FPM/nginx port and isolation consequences.

#### Acceptance criteria

- Users can keep `DB_HOST=localhost` in project configuration and continue to
  use the same host-facing DSN.
- CLI and web requests use the same effective database route.
- A stopped database, unavailable gateway, and wrong port produce a useful
  failure rather than a generic connection error.

### 4. Native PHP `imagick` startup warning

**Confirmed valid from source and the supplied warning.** PHP installation
always writes `etc/conf.d/imagick.ini` containing `extension=imagick` after
`make install` succeeds for the extension. There is no subsequent load test.
The warning shows that `imagick.so` exists, but its dependency
`libMagickWand-7.Q16HDRI.so.11` cannot be found by the dynamic linker.

The installer treats the extension build as successful if `make install`
returns successfully; that does not prove a new PHP process can load it. The
result is a warning on every native PHP invocation, even though the extension
is unusable. The existing `LD_LIBRARY_PATH` handling in the legacy OpenSSL
build path is build-time-oriented and does not establish a durable runtime
library path for ImageMagick.

#### Fix plan

1. After installing imagick, run the exact installed PHP binary with the
   version's configuration and verify `extension_loaded('imagick')` (or an
   equivalent minimal load check).
2. Inspect the module with `ldd`/platform-equivalent diagnostics and identify
   whether ImageMagick libraries are missing, outside the loader path, or
   incompatible.
3. Make the runtime library resolution durable using the project-supported
   mechanism (for example an appropriate rpath or generated runtime
   environment), without affecting unrelated PHP versions.
4. If the module cannot be made loadable, do not leave `imagick.ini` enabled.
   Report the missing dependency and provide a repair command; PHP itself must
   remain usable without a recurring startup warning.
5. Add installer and doctor checks for module loadability, dependency errors,
   repair/reinstall behavior, and a clean `chauf php <version> -v` output.

#### Acceptance criteria

- A native PHP command emits no imagick startup warning when imagick is
  installed correctly.
- A missing optional dependency produces one actionable install/doctor result,
  not a warning on every PHP invocation.
- Other PHP versions and extensions are not affected by the fix.

## Additional issues found

These findings are related to the reported problems but are not all required
for the first implementation patch.

1. **Reverse-proxy reachability differs by runtime.** Podman nginx routes host
   development servers through `host.containers.internal`, while native nginx
   uses `localhost`. A host service bound only to IPv6 loopback or IPv4
   loopback can therefore produce a Podman-only 502. See
   `docs/plans/2026-08-30-native-podman-runtime-parity-investigation.md` for
   the previously recorded `arvi-ui` evidence and integration-test plan.
2. **Runtime lifecycle output is not fully uniform.** Native and Podman start,
   stop, status, and logs still have different resource resolution and output
   behavior. Podman status cannot always provide native PID/uptime/memory
   equivalents, and project-scoped paths need shared resolution.
3. **Podman project topology has latent risks.** Multiple dedicated projects
   can use the same `/workspace` path, and stale containers require complete
   image, network, mount, port, and workspace-identity validation.
4. **Generated shim logic is duplicated.** `php` and `composer` shims each
   parse YAML and `.env` independently. This increases the chance that CLI,
   Composer, and web/FPM database behavior diverge.
5. **Podman follow logs need true streaming.** The production command runner
   buffers ordinary command output; follow-mode logging should stream and
   honor cancellation.
6. **Non-TUI commands emit terminal query responses.** Importing Bubble Tea
   initializes Lip Gloss, whose dependency `termenv` probes the terminal with
   OSC 11 and cursor-position queries. On terminals that expose those responses
   as text, commands such as `chauf self-update --dev` show sequences such as
   `ESC]11;rgb:...` and `ESC[2;1R` around otherwise normal output. This was
   reproduced under a pseudo-terminal and originates in the Bubble Tea
   dependency initialization, before `RunSelfUpdate` executes. The CLI should
   avoid loading/querying TUI dependencies for non-interactive commands, or
   move to a dependency/version that does not perform global terminal probes.

## Recommended implementation sequence

1. Add tests and finalize the command namespace (`podman-db` plus the generic
   `podman` operations and compatibility policy).
2. Centralize runtime readiness, service inventory, and environment resolution.
3. Fix database connectivity and add real CLI/web connection tests.
4. Make auto-start use the readiness/lifecycle contract and migrate stale units.
5. Fix imagick installation validation and runtime library resolution.
6. Address reverse-proxy diagnostics, topology validation, lifecycle output, and
   log streaming as separate reviewable changes.

## Implementation notes

The first implementation pass now includes:

- a generic, Chauffeur-scoped `chauf podman` dispatcher for `ps`, `inspect`,
  `logs`, and `exec`, while database subcommands remain available through the
  deprecated compatibility path;
- a shared Go database-host translation helper used by CLI and nginx config;
- a Podman readiness gate that attempts `podman machine start` when the initial
  Podman info check fails and polls for readiness with cancellation/timeout;
- stale Chauffeur systemd unit discovery based on unit files rather than only
  currently installed PHP versions or registered projects;
- systemd units now resolve the installed/running `chauf` executable instead of
  assuming a workspace-local binary;
- imagick builds now add pkg-config-discovered ImageMagick library directories
  as version-local ELF rpaths before validating module loading;
- native `doctor --check-php` now reports installed imagick modules that fail
  the same loadability check used by the installer;
- an opt-in Podman integration fixture verifies a PHP container can reach a
  published database port while the project `.env` remains `DB_HOST=localhost`;
- generated FPM parameters now carry the configured database port alongside
  the container-only host translation, enabling the same route for web/FPM;
- post-install imagick load validation that removes the extension ini when the
  module cannot be loaded.
- Podman `start` now reconciles the declared workspace before starting service
  identities, so login auto-start can recreate missing PHP-FPM/nginx
  containers instead of issuing `podman container start` against absent names.
- Runtime changes now inspect each current service first and stop only running
  native or Podman services before changing the engine, preventing the old
  runtime from retaining shared ports during the handoff.

The generated PHP and Composer shims now share one generated database
environment override helper, including a configured `DB_PORT`; the Go PHP
command uses the same override contract. Broader database engine/application
integration and durable ImageMagick loader-path repair remain environment-
dependent follow-up work.

## Verification performed during investigation

- `go test ./internal/commands ./internal/runtime ./internal/system` — passed.
- `go test ./...` — passed.
- `CHAUFFEUR_PODMAN_INTEGRATION=1 go test ./...` — passed; the opt-in
  Podman fixtures completed successfully in the available rootless environment.
- `go run ./cmd/chauf self-update --dev --dry-run` — passed without terminal
  query escape sequences in the current environment.
- `git diff --check` — passed before the investigation began.

This document records the findings, implementation decisions, and remaining
verification work; it is updated alongside the source changes.
