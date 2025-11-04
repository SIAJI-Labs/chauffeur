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
| `chauf link` | `--site <domain>`, `--ssl`, `--php <version>`, `--force` | Register **PWD** as a project. Creates `project.yaml`, prepares runtime/log dirs, optionally map local domain via Caddy, set default PHP for this project. Validates specified PHP version is installed. |
| `chauf links` | _none_ | List all registered projects and their metadata. |
| `chauf unlink` | `--force`, `--slug <slug>`, `--site <domain>`, `--project <path>`, `--all` | Remove a registered project. Can unlink by slug, domain, path, all projects with proper confirmation unless --force is used. |
| `chauf self-update` | `--dev` | Pull latest Chauffeur git changes via SSH and rebuild the CLI binary in-place (services unaffected; requires git & go). With --dev, rebuild from current directory if it's a valid chauffeur repository. |

### PHP Management

| Command | Flags/Args | Description |
|---|---|---|
| `chauf php install <version>` | `--force`, `--no-ext`, `--from <source>` | Install PHP runtime `<version>` into `~/.chauffeur/php/<version>` (source can be `source`, `tarball`, or `distro-extract`; default `tarball`). |
| `chauf php use <version>` | _none_ | Set global default PHP version used by `chauf php ...` commands. |
| `chauf php isolate <version>` | _none_ | Pin current directory/project to `<version>` (requires linked project and installed runtime). |

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
- `chauf link` generates `~/.chauffeur/projects/<slug>/project.yaml` (slug from directory name) and prepares runtime/log directories; `--force` must be supplied to overwrite an existing registration.
- `chauf unlink` removes project registrations and their directories; when run without flags and inside a registered project directory, it unlinks that project by default; explicit flags include `--slug <slug>` to unlink by project slug, `--site <domain>` to unlink by site domain, `--project <path>` to unlink by absolute path, and `--all` to unlink all projects; requires confirmation unless `--force` is provided.
- `chauf php isolate <version>` validates that the requested runtime is installed and the current directory is linked before updating the project configuration.
- `chauf self-update` ensures a clean git workspace, fast-forwards the Chauffeur repo under `~/.chauffeur/src/chauffeur` using the SSH remote `git@github.com:SIAJI-Labs/chauffeur.git` by default (override via `CHAUF_REPO_URL`), then rebuilds the CLI binary; service runtimes remain untouched (use `chauf install <service> --force` to refresh runtimes).
- With `--dev` flag, `chauf self-update --dev` rebuilds the CLI binary from the current working directory if it's a valid Chauffeur repository (must be a git repo containing `cli/main.go`, `go.mod`, and `AGENTS.md`).
- First install must handle not‑on‑PATH scenario: provide shell one‑liner to add `~/.chauffeur/bin` to PATH in `~/.bashrc`/`~/.zshrc`.

---

## 5) Code Generation Guidelines (for Codex)

- Prefer **Go** for the main CLI (single static binary). Use Bash only for thin wrappers.
- Support Linux first; avoid macOS/Windows assumptions.
- **Idempotency**: Re-running `init`, `install`, `link` should never corrupt state.
- **Install prefix**: All binaries/configs live under `~/.chauffeur/` only.
- **PATH shims**: Create wrappers in `~/.chauffeur/bin/shims` for each binary.
- **Dry‑runs**: When `--dry-run` is present, print planned actions without side-effects.
- **Logging**: Follow the standardized CLI logging specification below.
- **Failure logs**: Any command failure must append a detailed log entry under `<workspace>/logs/<component>/`. Log filenames follow `<action>[-<version>]-<YYYYMMDDTHHMMSSZ>.log`, and CLI output must surface the exact path.
- **Errors**: Clear actionable messages with suggested fix.
- **Permissions**: Do not require root; if privileged steps are unavoidable, print the exact `sudo` command for the user to run.
- **CLI Modularity**: Keep `main.go` limited to dispatch; implement each command in its own Go file/package (e.g. `cli/commands/<command>.go`) with focused helpers.

---

## 5.1) CLI Logging Specification

All Chauffeur CLI commands must follow a consistent, human-readable logging pattern that balances visual appeal with informative output. The logging system is inspired by the `self-update` command's style and incorporates structured progress indicators.

### 5.1.1) Color and Formatting Standards

**ANSI Color Palette (from self-update command):**
```go
const (
    colorReset  = "\033[0m"
    colorRed    = "\033[31m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorBlue   = "\033[34m"
    colorGray   = "\033[90m"
    colorBold   = "\033[1m"
)
```

**Usage Patterns:**
- `blue("[ command-name ]")` - Command prefix for status messages
- `bold(text)` - Important information, headers, summaries
- `green("✓")` - Success indicators
- `red("✗")` - Failure indicators  
- `yellow(text)` - Warnings, EOL notices, cautions
- `gray(text)` - Secondary information, URLs, SHAs, file paths
- `colorize(color, text)` - Helper function with terminal detection

### 5.1.2) Progress Indicators

**Spinners (for indeterminate operations):**
```go
type progressSpinner struct {
    message    string
    enabled    bool
    stop       chan struct{}
    done       chan struct{}
    startTime  time.Time
    dotCounter int
}

// Usage patterns from self-update:
spin := newSpinner("Cloning Chauffeur sources")  // Active operation
spin.Success("into ~/.chauffeur/src/chauffeur") // Success with context
spin.Fail("clone failed")                        // Failure with context
```

**Progress Bars (for downloads/long-running transfers):**
```go
type progressPrinter struct {
    label     string
    total     int64
    current   int64
    width     int
    lastPrint time.Time
}

// Usage from installers:
progress := newProgressPrinter("Download nginx.tar.gz", totalBytes)
writer = io.MultiWriter(file, progress)
// Renders: "    - Download nginx.tar.gz [##########........] 65%"
```

### 5.1.3) Command Output Structure

**Multi-step Operation Format (like self-update):**
```
[ self-update ] Starting self-update process...
[ self-update ] Cloning Chauffeur sources ✓ into ~/.chauffeur/src/chauffeur (42.1s)
[ self-update ] Updating branch main ✓ updated to a1b2c3d (was f4e5d6c, 12.3s)
[ self-update ] Building Chauffeur CLI ✓ binary installed to /usr/local/bin/chauf (8.7s)

[ self-update ] Summary:
  └── Duration: 63.2s
  └── Previous: f4e5d6c
  └── Current:  a1b2c3d
  └── Changes:  updated

[ self-update ] Self-update complete (commit a1b2c3d).
```

**Installation Command Format (like service install):**
```
[ install ] Installing PHP 8.3...
[ install ] Downloading php-8.3.12.tar.gz... [####################] 100% (15.2 MiB)
[ install ] Extracting source archive... ✓ (2.1s)
[ install ] Configuring build environment... ✓ (1.8s)
[ install ] Compiling PHP 8.3... ✓ (3m 42.1s)
[ install ] Installing to ~/.chauffeur/php/8.3... ✓ (0.8s)

[ install ] PHP 8.3 installation complete.
  └── Location: ~/.chauffeur/php/8.3
  └── PHP CLI: ~/.chauffeur/bin/shims/php-8.3
  └── Extensions: gd, pdo, curl, json, mbstring
```

**Error Handling Format:**
```
[ install ] Downloading nginx.tar.gz ✗ failed
  └── Error: HTTP 404: https://nginx.org/download/nginx-1.25.3.tar.gz
  └── Suggestion: Check if the version exists or try --force to reinstall
```

**Warning/Informational Format:**
```
[ install ] Installing PHP 7.4...
  ⚠ Warning: PHP 7.4 has reached End of Life (EOL)
  └── Consider using PHP 8.2+ for production deployments
[ install ] Continuing with PHP 7.4 installation...
```

### 5.1.4) Status Message Patterns

**Command Prefixes (consistent across all commands):**
```go
const statusPrefix = "[ command-name ]"  // Replace with actual command
```

**Operation Progress Messages:**
- Present continuous: `"Downloading..."`, `"Configuring..."`, `"Building..."`
- Active with context: `"Downloading to ~/.chauffeur/cache..."`

**Status Indicators:**
- `✓` or `green("✓")` - Success completion
- `✗` or `red("✗")` - Failed completion  
- `⚠` or `yellow("⚠")` - Warning/caution
- `└──` - Hierarchical context information (gray)

### 5.1.5) Timing Information

**Duration Formatting:**
```go
func formatDuration(d time.Duration) string {
    // Handles appropriate human-readable formatting
    // Examples: "42.1s", "3m 42.1s", "1h 23m 45s"
}
```

**Always show timing:**
- Operation completion: `✓ (2.3s)`
- Download progress: `[####...] 67% (12.3 MiB/s)`
- Summary sections: Include overall duration

### 5.1.6) Terminal Detection

**Always implement terminal-aware output:**
```go
func isTerminal(f *os.File) bool {
    info, err := f.Stat()
    if err != nil {
        return false
    }
    return (info.Mode() & os.ModeCharDevice) != 0
}

func colorize(color, text string) string {
    if !isTerminal(os.Stdout) {
        return text  // Strip colors when not a terminal
    }
    return color + text + colorReset
}
```

**Fallback for non-terminals:**
- Remove spinner animations, show static messages
- Remove color formatting
- Keep progress bars functional (text-based)

### 5.1.7) Logging Implementation Requirements

**New Helper Functions (create in cli/internal/logging):**
```go
// Command logger with consistent formatting
type CommandLogger struct {
    command string
    colors  bool
}

// Methods to implement:
- NewCommandLogger(command string) *CommandLogger
- Info(message string)
- Success(message, context string)
- Fail(message, error string) 
- Warn(message, context string)
- StartSpinner(message string) *Spinner
- StartProgress(label string, total int64) *ProgressPrinter
- PrintSummary(items []SummaryItem)
```

**Integration Points:**
- All command handlers should create a logger instance
- Replace `fmt.Printf` calls with logger methods
- Ensure consistent prefix usage: `[ command-name ]`
- Add appropriate color and formatting based on output type

### 5.1.8) Log File Structure

**Detailed Failure Logs:**
```
~/.chauffeur/logs/
  install/
    php-8.3-install-20250104T143022Z.log  // Detailed error log
    nginx-install-20250104T142155Z.log
  self-update/
    update-20250104T141533Z.log
```

**Log File Format:**
```
2025-01-04T14:30:22Z [INFO] Starting PHP 8.3 installation
2025-01-04T14:30:22Z [DEBUG] Workspace: /home/user/.chauffeur
2025-01-04T14:30:23Z [INFO] Downloading from https://www.php.net/distributions/php-8.3.12.tar.gz
2025-01-04T14:30:38Z [ERROR] Download failed: HTTP 404: Not Found
2025-01-04T14:30:38Z [ERROR] Stack trace: ...
```

**Always show log file path on failures:**
```
[ install ] Download failed ✗
  └── Error: HTTP 404 from download URL
  └── Detailed log: ~/.chauffeur/logs/install/php-8.3-install-20250104T143022Z.log
```

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
```

## 6.4) Testing Standards

**Directory Structure**: All tests must be organized under `tests/` following operation-based structure:
```
tests/
├── install/
│   ├── php_test.go          # PHP installation tests
│   ├── nginx_test.go        # Nginx installation tests
│   └── caddy_test.go        # Caddy installation tests
├── link/
│   ├── link_project_test.go  # Project linking tests
│   ├── unlink_project_test.go # Project unlinking tests
│   └── list_projects_test.go # Project listing tests
├── php/
│   ├── php_use_test.go      # PHP global version switching
│   ├── php_isolate_test.go  # Project PHP isolation
│   └── php_binary_test.go   # PHP binary execution
├── self_update/
│   ├── update_test.go       # Self-update functionality
│   └── dev_update_test.go   # Development mode updates
├── start_stop/
│   ├── start_test.go        # Service startup tests
│   └── stop_test.go         # Service shutdown tests
└── integration/
    ├── end_to_end_test.go   # Full workflow tests
    └── real_world_test.go   # Real-world scenario tests
```

**Test Function Naming**: Use descriptive names following the pattern:
```go
func TestCommandComponentAction(t *testing.T)
func TestLinkProjectCreatesConfig(t *testing.T)
func TestPhpUseSetsDefaultVersion(t *testing.T)
func TestSelfUpdateDevModeRebuildsFromRepo(t *testing.T)
```

**Test Structure and Coverage**:

### 6.4.1) Standard Test Template

All tests should follow this structure:
```go
func TestCommandSpecificBehavior(t *testing.T) {
    // 1. Setup: Create temporary environment
    tmpHome := t.TempDir()
    t.Setenv("HOME", tmpHome)
    
    // 2. Arrange: Set up the test state
    // - Create necessary directories
    // - Set up mock files
    // - Configure environment variables
    
    // 3. Act: Execute the function/behavior being tested
    output := captureOutput(func() error {
        return commands.RunCommand([]string{"arg1", "arg2"})
    })
    
    // 4. Assert: Verify the results
    assert.Contains(t, output, "expected message")
    assert.FileExists(t, expectedFilePath)
    
    // 5. Cleanup: Use t.Cleanup() for any manual cleanup
    t.Cleanup(func() {
        // Clean up resources
    })
}
```

### 6.4.2) Coverage Requirements

**Minimum Coverage Standards**:
- **Unit Tests**: 80% line coverage minimum for all packages
- **Integration Tests**: Cover all command entry points and success/failure paths
- **Error Paths**: Test all error conditions and edge cases
- **File Operations**: Test file creation, modification, deletion
- **Environment Variables**: Test with different HOME, PATH configurations

**Required Test Categories**:

#### 6.4.2.1) Command Tests
- **Success Cases**: Normal operation flows
- **Error Cases**: Invalid arguments, missing files, permission issues
- **Edge Cases**: Empty inputs, corrupted files, network failures
- **Flag Combinations**: Test all flag combinations and interactions

#### 6.4.2.2) Installation Tests  
- **Clean Install**: First-time installation scenarios
- **Reinstall**: Installing over existing installations
- **Force Install**: `--force` flag behavior
- **Version Validation**: Supported/unsupported version handling
- **Network Errors**: Timeout, connection failure scenarios

#### 6.4.2.3) Configuration Tests
- **Default Config**: Creating and reading default configurations
- **Custom Config**: Handling user modifications
- **Migration**: Config version upgrades and backwards compatibility
- **Validation**: Invalid configuration detection and error reporting

#### 6.4.2.4) File System Tests
- **Permission Handling**: Different permission scenarios
- **Path Resolution**: Relative/absolute path handling
- **Race Conditions**: Concurrent access scenarios
- **Symlinks**: Symlink creation, following, and validation

### 6.4.3) Test Utilities and Helpers

**Standard Helper Functions** (in `tests/helpers_test.go`):
```go
// Already exists: captureOutput(), captureError()
// Add these standard helpers:

// createTempFile creates a temporary file with given content
func createTempFile(t *testing.T, content string) string

// createMockWorkspace sets up a complete mock chauffeur workspace
func createMockWorkspace(t *testing.T, versions ...string) string

// mockPHPInstallation creates a fake PHP installation
func mockPHPInstallation(t *testing.T, tmpHome, version string) string

// assertProjectConfig validates project.yaml contents
func assertProjectConfig(t *testing.T, configPath string, expected interface{})

// assertLogEntry checks if specific log entry exists
func assertLogEntry(t *testing.T, logPath, expectedMessage string)
```

### 6.4.4) Integration Test Standards

**End-to-End Tests**: Should cover complete user workflows:
```go
func TestCompleteWorkflow_PHPProjectSetup(t *testing.T) {
    // 1. Install Chauffeur
    // 2. Install PHP 8.3
    // 3. Create test project
    // 4. Link project with custom domain
    // 5. Isolate project to specific PHP version
    // 6. Start services
    // 7. Verify accessibility
    // 8. Stop and cleanup
}
```

### 6.4.5) Mock and Test Structure Guidelines

**Test Isolation**: Each test should:
- Use `t.TempDir()` for temporary directories
- Use `t.Setenv()` for environment variables
- Use `t.Cleanup()` for cleanup operations
- Never modify the actual user's `~/.chauffeur/` directory

**Dependency Injection**:
- Use interface-based design for easier mocking
- Create testable functions that accept dependencies
- Avoid static global state in production code

### 6.4.6) Performance and Reliability Tests

**Benchmark Tests**:
```go
func BenchmarkPHPInstallation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark installation performance
    }
}
```

**Parallel Testing**:
```go
func TestConcurrentOperations(t *testing.T) {
    t.Parallel()
    // Test concurrent command execution
}
```

### 6.4.7) Test Execution and CI Integration

**Test Standards for CI/CD**:
- All tests must pass on clean environments
- Use deterministic versions and checksums in tests
- Mock external dependencies (HTTP calls, filesystem operations)
- Include timeout handling for long-running operations

**Local Development**:
```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./tests/...

# Run specific test category  
go test -v ./tests/install/...

# Run with coverage requirements
go test -v -race -cover ./tests/ && \
  go tool cover -func=coverage.out | \
  awk '/total:/{print $3}' | \
  sed 's/%//' | \
  awk '{if($1 < 80) {exit 1}}'
```

### 6.4.8) Test Documentation

**Test Documentation Requirements**:
- Each test package should have a package comment explaining its purpose
- Complex tests should have inline comments explaining the test scenario
- Integration tests should document the user workflow being tested
- Performance-sensitive tests should include timing expectations

**Example Package Comment**:
```go
// Package install tests the chauffeur installation commands.
// These tests verify:
// - Service installation (php, nginx, caddy)
// - Configuration file generation
// - Error handling for various failure scenarios
// - File system interaction and permission handling
package install
```

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
4. `chauf link --site myproj.test --ssl --php 8.2` creates project folder structure and `project.yaml` metadata (Caddy/Nginx assets pending).
5. `chauf php isolate 7.4` updates the linked project's `project.yaml` with the per-project PHP override.
6. `chauf self-update` fast-forwards the workspace git clone and rebuilds the CLI binary in-place.
7. `chauf start` in project directory boots required services; `stop` halts them cleanly.

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
