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
   - Address ignored errors in `cli/internal/services/manager.go`.
   - Ensure explicit logging or failure when config generation fails.
2. **Security Audit: File Permissions**
   - Review and fix `0755` defaults for sensitive directories in `cli/internal/projects` and `cli/lib/ssl.go`.
   - Ensure `0600` for keys and `0644` for configs.
3. **Security Audit: Command Injection**
   - Harden `chauf doctor` internal command parsing to prevent injection risks.
4. **Fix Template Engine Paths**
   - Replace brittle relative paths in `cli/internal/templates/engine.go` with `embed` or strict config paths.

### Phase 2: Refactoring & UX (Medium Priority)
*Focus: Improving code maintainability and user interaction.*

1. **Refactor `RunLink`**
   - Extract `handleNewProject`, `handleAliasAddition`, `handleProjectUpdate` from `cli/commands/link.go`.
   - Simplify main control flow.
2. **Centralize Defaults**
   - Create `cli/internal/config/defaults.go`.
   - Migrate scattered constants (repo URLs, ports) to this single source.
3. **UX: Standardize Flags**
   - Implement global `--verbose`/`--quiet` support across all commands.
4. **UX: Interactive Prompts**
   - Add interactive fallback for missing args in `chauf link`.

### Phase 3: Architecture & Performance (Long Term)
*Focus: Structural improvements and optimization.*

1. **Dependency Injection**
   - Refactor `http.Client` usage in `releases` package.
   - Reduce reliance on `coverage_probe.go` in tests where possible.
2. **Performance Optimization**
   - Implement caching for `detectDistro`.
   - Optimize `TemplateEngine` file checks if metrics indicate slowness.
3. **Testing Improvements**
   - Adopt standard Go interface mocking.
   - Explore `testscript` for robust CLI integration testing.
