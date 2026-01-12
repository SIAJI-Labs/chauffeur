# Chauffeur Codebase Analysis Report

This document outlines areas for improvement, potential issues, and inconsistencies identified in the Chauffeur CLI codebase.

## 1. User Experience (UX) Improvements

### Flag Consistency
- **Issue**: Inconsistent verbosity flags across commands.
  - `chauf init` uses `--quiet`.
  - `chauf doctor` uses `--verbose` and `--quiet`.
  - `chauf install` currently lacks `--quiet`.
- **Recommendation**: Standardize `--verbose` (`-v`) and `--quiet` (`-q`) flags globally across all commands.

### Interactive Prompts
- **Issue**: `chauf link` fails immediately if required arguments (like `--php`) are missing.
- **Comparison**: `chauf install php` offers a nice interactive menu for version selection.
- **Recommendation**: Implement interactive prompts for missing required arguments in `link` and `status --project` to improve "happy path" usage.

### Help & Documentation
- **Issue**: `chauf link` usage output is dense and mixes basic and advanced usage.
- **Recommendation**: Split help output into "Basic Usage" and "Advanced/Multi-domain" sections.

## 2. Code Quality & Maintenance

### Refactoring `RunLink`
- **Observation**: `cli/commands/link.go` contains the `RunLink` function which has grown significantly. It handles:
  - New project creation
  - Alias additions (recent feature)
  - Updates
  - SSL generation
  - Nginx template generation
  - Service restarting
- **Recommendation**: Refactor `RunLink` by extracting logic into dedicated handlers:
  - `handleNewProject(...)`
  - `handleAliasAddition(...)`
  - `handleProjectUpdate(...)`

### Centralized Defaults
- **Observation**: Some default values (e.g., git repo URLs in `doctor.go`, port defaults) are scattered.
- **Recommendation**: Move all constants and default values to `cli/internal/config/defaults.go` to maintain a single source of truth.

## 3. Potential Issues & Robustness

### Error Handling in Service Manager
- **Issue**: In `cli/internal/services/manager.go`, errors during `generateNginxConfig` are explicitly ignored in some paths (`// Continue even if config generation fails`).
- **Risk**: This could leave the system in an inconsistent state where the service starts with invalid or missing configuration.
- **Recommendation**: Log warnings explicitly or halt operation if critical configuration cannot be generated.

### Input Validation
- **Observation**: `safeDomainPattern` in `link.go` might be too restrictive for some valid domain names (e.g., international domains), though `.test` TLD restriction is good for security.
- **Recommendation**: Review domain validation regex against broader use cases while maintaining security.

## 4. Security Considerations

### File Permissions
- **Observation**: Critical files (SSL keys) are generally handled correctly, but directory creation often defaults to `0755`.
- **Recommendation**: Audit all file creation calls to ensure strict permissions (`0600` for keys, `0644` for configs) are consistently applied, especially in `cli/internal/projects` and `cli/lib/ssl.go`.

### Internal Command Injection
- **Observation**: `chauf doctor` calls itself via `exec` for OpenSSL fixes.
- **Recommendation**: Ensure strict validation of arguments passed to internal subcommands to prevent any potential argument injection, even if they currently originate from CLI args.

## 5. Inconsistencies

### Logging
- **Observation**: Mixed use of `logger.Prompt` and raw `fmt.Scan` functions.
- **Recommendation**: encapsulate input reading in `cli/lib` to handle consistent prompting styles and potentially support non-interactive modes (auto-accept/fail).

## 6. Architecture & Dependency Injection

### Internal Packages Structure
- **Observation**: `cli/internal/templates/engine.go` has hardcoded fallbacks to parent directories (`filepath.Join(filepath.Dir(exePath), "..", "..", "templates", "nginx")`) which makes it brittle if the binary location changes.
- **Recommendation**: Use a proper resource embedding strategy (Go `embed`) or a strictly defined configuration path for templates to avoid relative path guessing.

### GitHub Release Handling
- **Observation**: `cli/internal/releases/github.go` creates a new `http.Client` inside `LatestGitHubRelease` if nil is passed.
- **Recommendation**: Inject a configured `http.Client` with timeouts from `main.go` or a factory to ensure consistent network behavior (timeouts, proxies, etc.) across the application.

## 7. Testing Strategy

### Integration Tests
- **Observation**: `tests/` directory is rich with integration tests (e.g., `tests/doctor/integration/integration_test.go`), but many tests seem to rely on `coverage_probe.go` files which suggests a custom coverage mechanism.
- **Recommendation**: Ensure standard Go testing patterns are used alongside custom probes. Consider using `testscript` for CLI interaction testing which is more robust for shell-like interactions.

### Mocks & Stubs
- **Observation**: Some tests use `coverage_probe.go` to expose internals for testing.
- **Recommendation**: Prefer dependency injection (interfaces) over exposing private members via probes where possible to keep production code clean.

## 8. Performance

### File I/O
- **Observation**: `detectDistro` in `cli/internal/system/detect.go` opens `/etc/os-release` every time.
- **Recommendation**: Cache system detection results (singleton or memoization) as distro/arch won't change during command execution.

### Template Processing
- **Observation**: `TemplateEngine.DetectTemplateType` checks file existence sequentially.
- **Recommendation**: For deep directory trees or slow filesystems, parallelize checks or optimize file walking if this becomes a bottleneck (currently unlikely for typical project roots).

---

## Plan Priority & Phasing

This section outlines the execution plan for addressing the identified issues, prioritized by impact and stability.

### Phase 1: Stability & Security (High Priority)
*Focus: Fixing potential bugs, security risks, and ensuring robust configuration.*

1. **Fix Service Manager Error Handling**
   - **Context**: In `cli/internal/services/manager.go`, `generateNginxConfig` failures are currently ignored.
   - **Example**:
     ```go
     // Before
     _ = generateNginxConfig(config) // Continue even if config generation fails

     // After
     if err := generateNginxConfig(config); err != nil {
         return fmt.Errorf("generate nginx config: %w", err)
     }
     ```
2. **Security Audit: File Permissions**
   - **Context**: `cli/internal/projects` and `cli/lib/ssl.go` often use `0755` for directories.
   - **Example**:
     ```go
     // Before
     os.MkdirAll(certDir, 0o755)

     // After
     os.MkdirAll(certDir, 0o700) // Restricted access for sensitive dirs
     ```
3. **Security Audit: Command Injection**
   - **Context**: `chauf doctor` executes internal commands with flags.
   - **Example**: Ensure `phpVersion` and `workspaceDir` are strictly validated against allowlists or regex (e.g., `^[0-9]+\.[0-9]+$`) before passing to `exec`.
4. **Fix Template Engine Paths**
   - **Context**: `cli/internal/templates/engine.go` uses `filepath.Join(exe, "../..")`.
   - **Example**: Use `//go:embed templates/*` to bundle templates directly into the binary, removing runtime filesystem dependency.

### Phase 2: Refactoring & UX (Medium Priority)
*Focus: Improving code maintainability and user interaction.*

1. **Refactor `RunLink`**
   - **Context**: `cli/commands/link.go` has a massive `RunLink` function.
   - **Example**:
     ```go
     func RunLink(args []string) error {
         // ... parsing logic ...
         if opts.IsNew {
             return handleNewProject(opts)
         } else if opts.AddAlias {
             return handleAliasAddition(opts)
         }
         return handleProjectUpdate(opts)
     }
     ```
2. **Centralize Defaults**
   - **Context**: Defaults like git URLs are hardcoded in `doctor.go`.
   - **Example**:
     ```go
     // cli/internal/config/defaults.go
     const (
         DefaultRepoURL = "https://github.com/SIAJI-Labs/chauffeur.git"
         DefaultHTTPPort = 8080
     )
     ```
3. **UX: Standardize Flags**
   - **Context**: `install` lacks `--quiet`.
   - **Example**: Create a shared `cli/internal/flags` package or struct mixin that defines `Verbose`, `Quiet`, `Help` and handles them consistently.
4. **UX: Interactive Prompts**
   - **Context**: `chauf link` fails if `--php` is missing.
   - **Example**:
     ```go
     if version == "" {
         version = lib.PromptSelect("Select PHP version", []string{"8.1", "8.2", "8.3"})
     }
     ```

### Phase 3: Architecture & Performance (Long Term)
*Focus: Structural improvements and optimization.*

1. **Dependency Injection**
   - **Context**: `releases` package creates its own `http.Client`.
   - **Example**: Pass `*http.Client` as a dependency to `LatestGitHubRelease` to allow injecting a mock client in tests.
2. **Performance Optimization**
   - **Context**: `detectDistro` reads files on every call.
   - **Example**:
     ```go
     var cachedInfo *Info
     func Detect() (Info, error) {
         if cachedInfo != nil { return *cachedInfo, nil }
         // ... detection logic ...
     }
     ```
3. **Testing Improvements**
   - **Context**: `coverage_probe.go` pattern usage.
   - **Example**: Replace probes with `internal` packages for white-box testing or use strictly public interfaces for black-box testing.
