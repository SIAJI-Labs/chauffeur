# Chauffeur — Agent Handbook

This document is the single source of truth for every autonomous contributor working on Chauffeur. Follow it exactly—commands, filesystem structure, and documentation rules here are authoritative.

## 1. Purpose & Inspiration
- **Purpose**: Chauffeur is a Linux CLI that provisions per-project PHP development services (nginx, PHP-FPM, dnsmasq-managed `.test` domains) without touching system prefixes. Everything lives under `~/.chauffeur/` so multiple projects can coexist safely.
- **Why it exists**: Linux lacks a Valet/Herd-equivalent. Chauffeur brings the same experience—instant `.test` domains, isolated PHP runtimes, project-aware shims—while respecting Linux packaging norms.
- **Inspiration**: Laravel Valet and Beyond Code’s Laravel Herd. Borrow their developer ergonomics (one command per action, auto-generated nginx configs, PHP isolation) but keep Chauffeur host-native and workspace-contained.

## 2. Core Principles
1. **Workspace-first**: All binaries, configs, sockets, and logs live under `~/.chauffeur`. Never install or mutate anything under `/usr`, `/etc`, or `/opt` directly.
2. **Minimal host impact**: Prefer workspace changes. If host-level files (e.g., dnsmasq configs in `/etc`) are unavoidable, print exact `sudo` commands, record what changed, and provide cleanup guidance.
3. **Manual project registration**: Projects are never auto-scanned. `chauf link` is the only way to register a directory and must always operate on PWD unless a path flag is provided.
4. **Idempotent operations**: Re-running `init`, `install`, `link`, etc. must be safe. Commands with side effects require `--force` to overwrite.
5. **Linux-focused**: Target Arch/Ubuntu/Debian; no macOS/Windows assumptions.
6. **No external env managers**: No Devbox, nix, Docker, etc. Chauffeur ships its own runtimes in the workspace.
7. **Documentation parity**: README.md, docs/TODO_STATUS.md, and AGENTS.md must describe the codebase exactly as it runs today.
8. **Host edits only as last resort**: Modify `/etc` or system services only when there is no workspace alternative. Such steps must be opt-in (user runs the printed commands) and include instructions for reverting during `chauf remove`/`chauf uninstall`.

## 3. Host Impact Policy
| Area | Allowed | How to handle |
|------|---------|---------------|
| `~/.chauffeur/**` | ✅ Yes | Create/modify freely; this is the managed workspace. |
| Project directories | ✅ Yes | Only for generated configs/metadata tied to that project (e.g., `project.yaml`). |
| `/etc/*`, `/usr/*`, `/var/*` | ⚠️ Only when unavoidable. Print the exact `sudo` commands, require explicit confirmation, and document cleanup steps. |
| Network/dnsmasq/systemd configs | ⚠️ Same as above—emit precise scripts for the user to run and record the changes for later removal. |
| iptables redirect (80→8080, 443→8443) | ✅ Chauffeur manages rules but tracks them under `~/.chauffeur/system/port-forwarding.json` for cleanup. |

## 4. Dependency & Toolchain Policy
- **Primary language**: Go 1.22+. No other languages for CLI commands.
- **Required host tools**: git, curl, tar, build-essential toolchain (gcc/make), pkg-config, openssl headers for PHP builds.
- **No bundled package managers**: Do not integrate Devbox, nix, asdf, etc.
- **Runtime installs**: `chauf install php|nginx|composer` builds or downloads into `~/.chauffeur/{php,nginx,composer}`.
- **Default ports**: nginx HTTP 8080, HTTPS 8443, PHP-FPM fallback 9000. Configurable via `config/chauffeur.yaml`.

## 5. Workspace Layout & Config Contracts
```
~/.chauffeur/
  config/
    chauffeur.yaml          # global config (see schema below)
  projects/
    <slug>/
      project.yaml          # per-project state (path, php, domain, ssl, created_at)
      runtime/php-fpm/
      logs/
  php/<version>/            # 8.3, 8.2, 7.4 … installed runtimes
  nginx/
    bin/
    etc/
    sites-available/
    sites-enabled/
    conf.d/
    certs/
  bin/
    chauf                   # optional self-managed binary
    shims/                  # php, composer, php-<ver>
  cli/templates/nginx/
```

### 5.1 `chauffeur.yaml`
```yaml
version: 1
telemetry: false
workspace_dir: ~/.chauffeur
nginx:
  enable: true
  http_port: 8080
  https_port: 8443
php:
  default: 8.3
ports:
  start_range: 8080
  end_range: 8099
  conflict_resolution: prompt   # prompt|auto|fail
  nginx_http_fallback: 8080
  nginx_https_fallback: 8443
  php_fpm_fallback: 9000
projects_dir: ~/.chauffeur/projects
```

### 5.2 `project.yaml`
```yaml
version: 1
path: /absolute/path/to/project
php: 8.3
site:
  domain: slug.test
  ssl: false
domains:
  aliases:
    - domain: admin.test
      ssl: true
    - domain: api.test
      ssl: false
runtime:
  php_fpm_socket: ~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.sock
created_at: 2025-10-30T12:00:00+07:00
```

## 6. Command Surface (authoritative)
| Command | Key Flags | Summary |
|---------|-----------|---------|
| `chauf init` | `--force`, `--quiet` | Bootstrap workspace under `~/.chauffeur/`. Idempotent. |
| `chauf start` | `--project <path>`, `--all`, `--dry-run` | Start nginx/PHP-FPM plus dnsmasq validation. |
| `chauf stop` | same flags as start | Stop services and clean port-forward rules. |
| `chauf restart` | `--project <slug>`, `--all`, `--dry-run` | Restart services (equivalent to stop then start, preserves configuration). |
| `chauf status` | `[service-type]`, `--project`, `--detail`, `-v` | Show status for global or per-project services. |
| `chauf link` | `--site`, `--secure`, `--php`, `--http-port`, `--https-port`, `--alias`, `--add-alias`, `--force` | Register current directory, detect template (Laravel/WordPress/general), generate configs with multi-domain support. |
| `chauf links` | — | List all registered projects in a formatted table. |
| `chauf unlink` | `--slug`, `--site`, `--project`, `--alias`, `--all`, `--force` | Remove registrations or specific aliases. Defaults to current dir. |
| `chauf secure` | — | Add SSL certificate to current linked project. Must be run from linked project directory. |
| `chauf unsecure` | — | Remove SSL certificate from current linked project. Must be run from linked project directory. |
| `chauf php install <ver>` | `--force`, `--no-ext`, `--from` | Build/install PHP runtime under workspace. |
| `chauf php use <ver>` | — | Set global default PHP. |
| `chauf php isolate <ver>` | — | Pin current linked project to a version. |
| `chauf php current` | — | Show current PHP version for directory or global default. |
| `chauf remove <service> [version]` | `--force` | Remove installed runtimes (php, nginx, composer). |
| `chauf uninstall` | `--purge` | Remove workspace (and runtimes with `--purge`). |
| `chauf self-update` | `--dev` | Update binary from git or rebuild from current repo. |
| `chauf install composer` | — | Fetch verified composer PHAR and shims. |
| `chauf logs [service] [version]` | `--follow`, `-f`, `--lines`, `-n`, `--level`, `--context`, `-c`, `--verbose`, `-v`, `--quiet`, `-q` | View and follow service logs with interactive version selection. |
| `chauf clean [target]` | `--dry-run`, `--force`, `--older-than`, `--keep-versions`, `--what` | Clean workspace files with file size display and accurate reporting. |
| `chauf info` | — | Show workspace paths, installed services, versions, port config. |

## 7. PHP & Composer Behavior
- PHP shims always prefer the project’s isolated version (from `project.yaml`). Outside linked projects they use the global default (`chauf php use`) and fall back to 8.3 if nothing is configured.
- `chauf install php` must verify `pkg-config` is available and that all required dev headers (`libzip`, `libjpeg`, `libpng`, `freetype`, `libxml2`, `libcurl`, `zlib`, `libxslt`, `readline`, `MagickWand`, `gmp`) are discoverable via `pkg-config --modversion …`, failing fast with remediation guidance when anything is missing/outdated.
- PHP needs readline-style line editing for PsySH/Laravel Tinker parity, so every build must pass `--with-readline` (or `--with-libedit` if ever required) and ensure the dependency is documented for users.
- Every PHP runtime must compile mysqlnd-based `mysqli` and `pdo_mysql` extensions so Laravel migrations and other PDO clients have a working driver immediately after installation.
- Every PHP runtime must include `gmp` (for arbitrary-precision mathematics) and `bcmath` (for precision decimal math) extensions, which are compiled directly into PHP rather than as separate extensions.
- `chauf install php` must build and enable the PECL `imagick` extension (requiring MagickWand/ImageMagick dev headers) so image-heavy apps work without follow-up steps. Drop `imagick.ini` into `etc/conf.d/` for each runtime.
- Imagick builds should default to the latest stable release reported by PECL’s REST API while allowing overrides via `CHAUFFEUR_IMAGICK_VERSION`/`CHAUFFEUR_IMAGICK_TARBALL`. Fall back to the pinned default when the API is unreachable.
- `chauf link` validates `--php` versions are installed before writing configs.
- Composer shim (`~/.chauffeur/bin/composer`) always uses Chauffeur’s PHP shim so `composer install` respects isolation.
- Removing PHP versions via `chauf remove php <ver>` must update shims and reassign defaults when necessary.

## 8. Multi-Domain Architecture

Chauffeur supports multiple domains per project with isolated SSL certificate management and alias-specific configuration.

### Data Structures
```go
type Config struct {
    Version   int
    Path      string
    PHP       string
    Site      *Site           // Primary domain
    Domains   *Domains        // Alias domains
    Runtime   Runtime
    CreatedAt time.Time
}

type Site struct {
    Domain string
    SSL    bool
}

type DomainAlias struct {
    Domain string `yaml:"domain"`
    SSL    bool   `yaml:"ssl"`
}

type Domains struct {
    Aliases []DomainAlias `yaml:"aliases,omitempty"`
}
```

### SSL Certificate Management
- **Multi-domain SAN certificates**: Single certificate covers all SSL-enabled domains (primary + aliases)
- **Certificate generation**: Supports both mkcert (trusted) and self-signed certificates
- **Automatic regeneration**: Certificates regenerated when SSL aliases are added/removed
- **File naming**: Uses base domain as certificate filename (e.g., `hja-cms.test.crt`)

### Certificate Implementation
```go
// generateMultiDomainSSLCertificate creates SAN certificates
func generateMultiDomainSSLCertificate(logger, certPath, keyPath, certBase string, domains []DomainAlias)

// OpenSSL SAN configuration generation
func createSANConfig(domains []string) string
```

### Domain Resolution
- **Primary domain**: From `proj.Site.Domain` (legacy support)
- **Alias domains**: From `proj.Domains.Aliases` array
- **Helper methods**: `GetAllDomains()`, `GetServerNames()`, `HasSSLEnabled()`
- **nginx integration**: Template engine uses `GetServerNames()` for multi-domain server blocks

### User Experience
- **Links display**: `(*)` indicators show SSL-enabled aliases
- **Unlink confirmation**: Shows all domains (primary + aliases) with SSL status
- **Command interface**: `--alias` flag for adding domains, `--secure` for per-alias SSL control

### Backward Compatibility
- Existing single-domain configurations unchanged
- Legacy `site.domain` and `site.ssl` fields maintained
- Migration path: single domains automatically work with new multi-domain system

## 9. Logging & Output Standards
- **Single logger**: Each command must create a `logger := lib.NewCommandLogger("<command>")`. No raw `fmt.Printf` for user-facing output.
- **Helpers only**: Use `logger.Info/Success/Warn/Error`, `logger.PrintSection`, `logger.PrintSummary`, `lib.NewSpinner`, and `lib.NewProgressPrinter`. Prevent duplicate spinner/progress implementations.
- **TTY detection**: Colors, spinners, and progress bars must disable themselves when stdout isn’t a terminal.
- **Structured failures**: Errors show `✗` marker, the human-readable fix, and the path to detailed log under `~/.chauffeur/logs/<command>/...log`.
- **Duration reporting**: Long operations log start/end with human-readable durations using `lib.formatDuration` helpers.

## 10. Writing Standards
1. **Language**: Go only. Commands live in `cli/commands`, helpers in `cli/lib` or `cli/internal/*` packages.
2. **Re-use helpers**: Before writing new utilities, look for existing helpers (ports, downloads, logging, checksum). Extend them instead of duplicating logic.
3. **Readable & visual**: Output should be human-friendly (clear sections, tables, color cues) yet informative enough for debugging.
4. **Consistency**: Uniform flag parsing, error messages, confirmation prompts, and path handling. Prefer `cobra`-style minimalism even without the library.
5. **Dry-run support**: When a command accepts `--dry-run`, execute zero side effects and print the plan via logger.
6. **Comments**: Only for non-obvious logic; keep concise.
7. **No secrets**: Never log tokens or credentials. Mask paths only when necessary.

## 11. Documentation Synchronization
Every code change requires immediate updates to:
1. **README.md** – feature status (✅/🚧/📋/🎯), installation changes, new commands/examples, roadmap tweaks.
2. **docs/TODO_STATUS.md** – mark tasks as ✅, move items between sections, update release notes/priority queue.
3. **sites/app/docs/**** – update Next.js site documentation to match current CLI behavior and new features.
4. **sites/constants.ts** – update command examples, feature descriptions, and navigation constants.
5. **sites/public/install.sh** – verify symlink to latest install script if CLI behavior changed.
6. **AGENTS.md** – update command contracts, filesystem rules, or architectural guidance if behavior changed.

### 11.1 Critical: Constants.ts Maintenance Requirement

**When CLI commands change, AI models MUST update `sites/constants.ts` immediately.**

The `sites/constants.ts` file serves as the **single source of truth** for:
- All CLI command definitions and examples
- Command flags, usage patterns, and expected outputs
- Feature descriptions and navigation constants
- Documentation site command reference generation

#### **Required Workflow for Command Changes:**
1. **CLI Implementation**: Modify Go commands in `cli/commands/`
2. **Immediate Constants Update**: Update corresponding command definitions in `sites/constants.ts`
3. **Example Verification**: Test all command examples against the actual binary
4. **Documentation Sync**: Ensure site documentation reflects changes

#### **Constants.ts Structure:**
- `CLI_COMMANDS[]`: Array of all command definitions with examples
- `COMMAND_CATEGORIES[]`: Category definitions with icons and descriptions
- `findCommand()`: Helper to locate command definitions
- Feature constants and navigation data

#### **Verification Checklist:**
- [ ] All new commands added to `CLI_COMMANDS[]`
- [ ] Command examples match actual binary output
- [ ] Flag descriptions match Go implementation
- [ ] Usage patterns are accurate
- [ ] Categories assigned correctly
- [ ] Examples tested with real binary

**Failure to update `sites/constants.ts` when commands change creates documentation divergence and user confusion.**

### 11.2 Site Documentation Specifics

The Next.js documentation site (`sites/`) must be kept in sync with all changes to Chauffeur's implementation:

#### **Command Accuracy Requirements**
- **CRITICAL**: All CLI commands and flags in site documentation MUST match the actual Go implementation in `cmd/`
- Before documenting any command, run `chauf --help` and specific command help to verify exact syntax
- Test all command examples to ensure they work with the current binary version
- Pay special attention to flag names (e.g., `--secure` vs `--secure`) and command structure (e.g., `chauf php isolate` vs `chauf isolate`)

#### **Binary Implementation Verification**
- All documented commands should reflect the actual behavior of the `chauf` binary
- When documenting features, first verify the implementation in the Go source code
- Check that command outputs, error messages, and status indicators match what users will see
- Ensure all command examples in `sites/constants.ts` are tested against the real binary

#### **UI Consistency Standards**
- Use the `CodeBlock` component for all command examples and outputs
- Maintain consistent styling for commands, flags, and file paths across all pages
- Follow established patterns for feature descriptions and badges
- Use responsive design patterns established in existing components

#### **URL and Script Management**
- All installation URLs must point to `https://chauffeur.siaji.com/install`
- Verify the `/install` route serves the install script correctly for both browser viewing and curl execution
- Ensure `sites/public/install.sh` symlink points to `/install.sh`
- Test installation process from the documented URLs

#### **Feature Status Reflection**
- Keep "Updated", "New", "Deprecated" badges current in `sites/constants.ts`
- Remove or mark deprecated features appropriately
- Add new features to the site as soon as they're implemented in the CLI

Checklist (do not skip):
- [ ] README reflects implementation.
- [ ] docs/TODO_STATUS.md matches current progress.
- [ ] sites/app/docs/** documentation matches CLI behavior.
- [ ] sites/constants.ts command examples are accurate.
- [ ] AGENTS.md matches actual behavior.
- [ ] Examples and commands have been run/tested.

## 13. Testing Standards
- **Location**: Place tests in `tests/` or alongside packages as `*_test.go`, but external tests may not import `cli/internal/**`. Use exported seams instead.
- **Execution**: `go test ./...` must pass before you push. Tests must isolate HOME/workspace via `t.TempDir()` and `t.Setenv` so they never touch the real user state.
- **Coverage**: Aim ≥80% on new packages. Use table-driven tests for commands, and capture output via helpers (see existing tests) instead of relying on global state.
- **Integration hooks**: Provide fakes/mocks for network and filesystem interactions so tests stay deterministic.

## 14. Dependency & Release Hygiene
- Do not build binaries directly inside the repo (e.g., `go build -o chauf`). Use `chauf self-update --dev` from the repo root to rebuild the CLI for debugging, which places artifacts under the workspace.
- Do not commit compiled binaries, build artifacts, or caches. Keep the repo source-only.
- Generated files (templates, configs) must be reproducible. If a command generates files, include instructions/tests to verify them.
- Releases should be performed by building from clean git state and tagging. Document the steps in README when they change.

## 15. Implementation Workflow for Agents
1. **Read docs first**: Before coding, skim README and docs/TODO_STATUS to understand current state.
2. **Plan**: Break work into verifiable steps. Call out doc updates early.
3. **Work inside workspace**: Respect Host Impact Policy and workspace layout.
4. **Keep build dirs when debugging**: Export `CHAUFFEUR_KEEP_BUILD_DIR=1` before running `chauf install php …` if you need the extracted PHP sources to stick around under `/tmp` for manual inspection; unset it to restore automatic cleanup.
5. **Offline tarballs**: When network access is blocked, set `CHAUFFEUR_PHP_TARBALL`, `CHAUFFEUR_PHP_SIGNATURE`, and `CHAUFFEUR_PHP_KEYRING` to point at local files so the installer can reuse cached PHP artifacts while still verifying signatures.
6. **Use helpers**: Logging, downloads, checksum, port allocation all have established helpers—use them.
7. **Test & lint**: Run `go test ./...` and relevant linters before marking tasks complete.
8. **Update docs**: Apply the synchronization checklist.
9. **Review output**: Ensure logs look clean, colors degrade gracefully, and errors cite log file locations.
10. **Final verification**: Re-run key commands (e.g., `chauf link --dry-run`) to validate the new behavior shown in docs.

## 16. PHP-FPM Architecture

### Project-Level PHP-FPM Control
Chauffeur provides **project-level PHP-FPM control** to balance resource efficiency and isolation needs:

### Shared FPM (Default)
- **Resource efficient**: Multiple projects share the same PHP-FPM pool per PHP version
- **Default behavior**: `chauf link` creates shared FPM unless `--dedicated-fpm` is specified
- **Example**: 10 projects using PHP 8.3 = 1 shared PHP-FPM process
- **Socket path**: `~/.chauffeur/php/8.3/runtime/php-fpm/php-fpm.sock`

### Dedicated FPM (Optional)
- **Maximum isolation**: Each project gets its own PHP-FPM pool
- **Usage**: `chauf link --dedicated-fpm` for critical projects needing custom configuration
- **Example**: 1 project = 1 dedicated PHP-FPM process
- **Socket path**: `~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.sock`

### Mixed Strategy Support
- **Flexible workspace**: Mix shared and dedicated FPM in the same environment
- **Automatic routing**: nginx automatically routes to the correct socket based on project configuration
- **Clear status**: `chauf status` shows global (shared) and project-specific (dedicated) services separately

### Implementation Details
- **Project configuration**: New `fpm:` section in `project.yaml`:
  ```yaml
  runtime:
    fpm:
      dedicated: false    # or true for dedicated FPM
      socket: /path/to/socket
  ```
- **Link command**: `chauf link --dedicated-fpm --php 8.1` creates dedicated FPM
- **Service management**: `cli/internal/services/manager.go` handles mixed strategies
- **nginx templates**: Automatically route to version-specific (shared) or project-specific (dedicated) sockets

### Benefits
- ✅ **Default efficiency**: Shared FPM conserves memory for typical development workflows
- ✅ **Selective isolation**: Dedicated FPM available when needed for specific projects
- ✅ **Backward compatibility**: Existing projects continue using shared FPM
- ✅ **Simple management**: Clear project-level control via flag during linking

## 17. Commit Policy for AI Agents

### 🚫 **NO UNAUTHORIZED COMMITS OR PUSHES** - STRICT PROHIBITION

**AI agents MUST NOT**:
- Make git commits without explicit user instruction
- Push changes to remote repositories without explicit user permission
- Use `git add`, `git commit`, or `git push` commands autonomously
- Run any git operations that modify repository state without user consent

### ⚠️ **IMMEDIATE POLICY REMINDER**
- **CRITICAL**: AI agents MUST ask for explicit permission before any git operations
- **NEVER** assume git operations are allowed, even for "obvious" changes
- **ALWAYS** present changes for user review before committing/pushing
- **STOP** immediately if unsure about git operation permissions

### ✅ **APPROVED WORKFLOW**

**AI agents SHOULD**:
1. **Stage changes locally** with `git add` for user review
2. **Present commit messages** for user approval
3. **Wait for explicit instruction** before committing
4. **Consolidate related changes** into single commits when requested
5. **Follow conventional commit format** when user approves

### 📋 **Example Correct Approach**

```
User: "Let's implement the doctor command"
AI: *implements functionality, stages changes*
AI: "I've implemented the doctor command with comprehensive dependency validation.
     Changes are staged and ready for your review:
     - cli/commands/doctor.go (1,045+ lines)
     - cli/main.go (CLI integration)
     - tests/doctor/ (26 test functions)
     Would you like me to commit this with the message 'feat(doctor): ...'?"
User: "Yes, please commit and push"
AI: *commits with approved message and pushes*
```

### ❌ **Example Forbidden Behavior**

```
AI: *automatically runs `git add . && git commit -m "..." && git push`*
```

### ⚡ **EXCEPTIONS**

The only cases where AI agents may make commits without explicit instruction:
1. **Critical hotfixes** for production systems (with immediate user notification)
2. **User-configured automation** where this policy is explicitly overridden

### 🔍 **ENFORCEMENT**

This policy applies to:
- All AI assistants and agents
- All automation tools that act on behalf of AI
- All development workflows involving autonomous code changes

**VIOLATIONS** of this policy must be immediately corrected by:
1. **Reverting unauthorized commits** if pushed
2. **Explaining the violation** to the user
3. **Following proper workflow** going forward

---

## 18. DNS & Port Automation Notes
- `chauf start` must verify dnsmasq configuration for `.test`. If missing, print the exact `sudo tee /etc/dnsmasq.d/chauffeur.conf` block plus restart commands; never edit system files directly.
- Port conflicts are resolved per the config's `ports.conflict_resolution`:
  - `prompt`: ask user via logger prompts.
  - `auto`: pick first free port within `start_range`/`end_range`.
  - `fail`: abort with actionable guidance.
- Port forwarding (80→HTTP port, 443→HTTPS port) is optional and tracked in `~/.chauffeur/system/port-forwarding.json`. Cleanup happens during `chauf stop` and `chauf remove nginx`.
- Port validators must detect when a conflict comes from Chauffeur-managed services (e.g., `chauf-nginx`). In that case, automatically restart the service to pick up new configs instead of forcing the user to pick a new port.
- `chauf link`/`chauf unlink` must restart `chauf-nginx` whenever they change site configs so new domains go live (or disappear) immediately.
- Nginx must always ship with a default catch-all server (`server_name _; return 404;`) on the configured HTTP port so unlinked domains don't bleed into other projects.

By following this handbook, every new change stays consistent with Chauffeur's goals: Valet-like ergonomics, Linux-friendly isolation, and crystal-clear documentation.
