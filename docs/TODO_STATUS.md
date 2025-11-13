# Chauffeur Development Tracker

_A living status board for features, debt, and priorities. Keep this in sync with README.md and AGENTS.md._

## 1. Snapshot
- **Maintainer**: @siegg (solo, learning Go; heavy AI assistance)
- **Stability**: Experimental – validated mostly on one Arch-based workstation
- **CI**: `go test ./...` currently red; needs attention before public release
- **Work focus**: Logging overhaul, service orchestration polish, test coverage

## 2. Completed ✅
### Workspace & Bootstrap
- Smart installer (`install.sh`) with Go prerequisite checks
- `chauf init` scaffolds `~/.chauffeur` with default config, templates, PATH guidance
- Workspace layout contracts documented in AGENTS.md
- Debug helper `CHAUFFEUR_KEEP_BUILD_DIR=1` preserves extracted PHP build directories for manual inspection when needed
- Offline-friendly installs (`CHAUFFEUR_PHP_TARBALL`/`CHAUFFEUR_PHP_SIGNATURE`/`CHAUFFEUR_PHP_KEYRING`) to reuse cached PHP sources when mirrors are unreachable
- Port validator recognizes Chauffeur’s own nginx usage and restarts it automatically, so `chauf link` can run while services are active without forcing new ports
- `chauf info` reports GitHub release status plus compare-local-vs-remote commit drift so contributors know when they’re ahead/behind
- Imagick installer fetches the latest stable tarball from PECL (with env overrides) so PHP builds keep working as releases evolve

### Project Registration
- `chauf link` / `links` / `unlink` end-to-end
- Project type detection (Laravel, WordPress, general)
- Nginx template generation + symlinks in `sites-available/enabled`
- Table-formatted `chauf links`
- `chauf unlink` removes nginx configs and restarts nginx when running so sites disappear immediately
- Default nginx catch-all returns 404 for unlinked domains so traffic never falls through to other projects

### PHP & Composer Fundamentals
- `chauf php install/use/isolate`
- Project-aware PHP shim
- Composer installer + shim that reuses Chauffeur PHP
- Laravel-required PHP extensions (gd, zip, exif, freetype, imagick) enabled during builds
- Documented host packages (libzip, libjpeg, libpng, freetype, libxml2, curl, zlib, bzip2, readline, libmagickwand) required before compiling PHP
- PHP installer validates `pkg-config` + all required libraries (libzip/libjpeg/libpng/freetype/libxml2/libcurl/zlib/libxslt/readline/MagickWand) before starting long builds so users get actionable remediation early
- PHP runtimes ship with GNU Readline enabled so PsySH / `php artisan tinker` arrow keys and history behave like system PHP builds
- PHP builds enable mysqlnd-based `mysqli` and `PDO_MySQL` extensions so database-backed apps run immediately after `chauf install php`

### Self-update
- `chauf self-update` fetches latest git release
- `chauf self-update --dev` rebuilds from current repo when run inside it

### CLI Infrastructure
- Structured logging enforced across all commands (`lib.Logger` only; usage/help text excluded)

### Smart Caching System
- Universal intelligent caching system across all services (PHP, Composer, Nginx)
- Auto-caching of successful downloads to `~/.chauffeur/cache/`
- Cache-first installation priority (local config → cache → remote download)
- User-controlled cache management during `chauf remove` operations
- Consistent cache experience across all services with educational messaging
- Smart version detection with API calls + fallback system for all services
- Cache-aware flags: `--no-cache` (skip caching), `--local` (use local tarball), `--force` (override cache)

### GD Extension Infrastructure for Legacy PHP
- Interactive GD extension prompting for PHP 7.4/8.0 with user education about time cost
- Temporary directory preservation during GD builds (`CHAUFFEUR_KEEP_BUILD_DIR` logic)
- Permanent patching system integrated into `cli/installers/php_legacy.go`
- GD compatibility header with wrapper functions for modern libgd integration
- Dramatically reduced patch warnings (from 13+ to 4 warnings)
- Graceful failure handling - PHP installation continues without GD if bundled build fails
- Build infrastructure successfully initiates GD extension compilation for legacy PHP versions

## 3. In Progress 🚧
1. **GD Extension Compilation for Legacy PHP** – Complete bundled GD extension support for PHP 7.4/8.0
2. **Service orchestration stability** – refine `chauf start/stop/status` to handle mixed project states, better error surfacing
3. **dnsmasq/NetworkManager automation** – ensure instructions are clear, reversible, and logged
4. **Documentation sync discipline** – ongoing effort to keep README, AGENTS, and this tracker aligned

## 4. Planned 📋
| Priority | Item | Notes |
|----------|------|-------|
| P1 | Complete GD extension compilation for PHP 7.4/8.0 | Fix remaining `make` phase errors during bundled GD build |
| P1 | Rework tests to avoid importing internal packages | Required for Go module hygiene |
| P1 | Stabilize `chauf start/stop` interaction with dnsmasq & port forwarding | Needs integration tests |
| P1 | Validate GD extension functionality across all PHP versions (7.4, 8.0, 8.1, 8.2, 8.3, 8.4) | Test image processing, thumbnails, watermarks |
| P2 | Expand PHP installer matrix (8.2/8.1/7.4 legacy deps) | Ensure workspace fallback works |
| P2 | Improve `chauf status --detail` output (tables, health info) | Align with logging spec |
| P2 | Add GD extension validation tests | Verify common PHP GD functions work correctly |
| P3 | Add onboarding docs for contributors (Go basics + AI workflow) | Help new maintainers |
| P3 | Publish release checklist (binary build, docs sync, testing) | Needed for first public release |

## 5. GD Extension Technical Status
### Current Implementation Status
- **PHP 8.1+**: GD extension works ✅ (modern PHP versions have compatible libgd integration)
- **PHP 8.0/7.4**: GD infrastructure in place, compilation failing during `make` phase ⚠️

### What's Working
- ✅ Interactive user prompting for GD extension choice with time cost education
- ✅ Temporary directory preservation for bundled extension builds
- ✅ Permanent patching system in `cli/installers/php_legacy.go`
- ✅ GD compatibility header with wrapper functions
- ✅ Dramatically reduced patch warnings (13+ → 4)
- ✅ Graceful failure handling (PHP install continues without GD)
- ✅ Build process successfully initiates GD compilation

### Remaining Technical Challenges
- **GD Make Phase Errors**: The bundled GD extension compilation fails during the `make` step with "exit status 2"
- **Function Pointer Compatibility**: Need to resolve remaining function signature mismatches between PHP 7.4/8.0 GD source and modern libgd libraries
- **Build Environment**: May need additional compiler flags or libgd version constraints

### Files Involved
- `cli/installers/php.go`: GD prompting and build orchestration
- `cli/installers/php_legacy.go`: GD compatibility patches and wrapper functions
- Build process preserves `/tmp/chauffeur-php-*/src/php-*/ext/gd/` for debugging

## 6. Known Issues / Tech Debt
- `go test ./...` fails due to tests importing `cli/internal/config` (Go disallows external packages from accessing internal)
- Repo currently contains built binaries (`chauf`, `main`) checked in by mistake – remove and add to `.gitignore`
- DNS setup path requires clearer rollback instructions when NetworkManager/systemd-resolved changes are applied manually
- GD extension compilation for PHP 7.4/8.0 fails during make phase (infrastructure works, compilation errors remain)

## 7. Testing & QA
- Target: `go test ./...` green on Go 1.22+
- Integration tests should stub HOME to temp directories, avoiding host mutation
- CI (`.github/workflows/go-tests.yml`) runs the suite on every pull request to `main`

## 8. Release Readiness Checklist
- [ ] `go test ./...` passes locally
- [ ] README.md / docs/TODO_STATUS.md / AGENTS.md agree on feature status
- [ ] No compiled binaries or caches tracked in git
- [ ] Logging contract enforced across commands
- [ ] dnsmasq/NetworkManager instructions reviewed and verified
- [ ] GD extension compilation verified for PHP 7.4/8.0 (optional for release)

## 9. How to Help
- Help expand service orchestration tests/logging and harden dnsmasq/NetworkManager flows
- Improve tests around `chauf link`/`links` to avoid double-run conflicts
- Document real-world setups (distro, Go version, dnsmasq config) in issues to broaden coverage
- Review AGENTS.md and propose clarifications before building new features

_Last updated: 2025-11-13T02:20:00Z_
