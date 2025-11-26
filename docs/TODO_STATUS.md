# Chauffeur Development Tracker

_A living status board for features, debt, and priorities. Keep this in sync with README.md and AGENTS.md._

## 1. Snapshot
- **Maintainer**: @si-aji (solo, learning Go; heavy AI assistance)
- **Stability**: Experimental – validated mostly on one Arch-based workstation
- **CI**: `go test ./...` now green with comprehensive test coverage across all packages
- **Work focus**: Service orchestration polish, documentation sync, release preparation

## 2. Completed ✅
### Workspace & Bootstrap
- Smart installer (`install.sh`) with Go prerequisite checks
- `chauf init` scaffolds `~/.chauffeur` with default config, templates, PATH guidance
- Workspace layout contracts documented in AGENTS.md
- Debug helper `CHAUFFEUR_KEEP_BUILD_DIR=1` preserves extracted PHP build directories for manual inspection when needed
- Offline-friendly installs (`CHAUFFEUR_PHP_TARBALL`/`CHAUFFEUR_PHP_SIGNATURE`/`CHAUFFEUR_PHP_KEYRING`) to reuse cached PHP sources when mirrors are unreachable
- Port validator recognizes Chauffeur's own nginx usage and restarts it automatically, so `chauf link` can run while services are active without forcing new ports
- `chauf info` reports GitHub release status plus compare-local-vs-remote commit drift so contributors know when they're ahead/behind
- Imagick installer fetches the latest stable tarball from PECL (with env overrides) so PHP builds keep working as releases evolve
- **Enhanced DNS resolution error handling** with retry logic, NetworkManager awareness, and graceful failure recovery
- **Improved port forwarding** with better sudo detection, state validation, and fallback when privileges unavailable
- **Robust start/stop orchestration** with comprehensive error handling and integration test coverage
- **`chauf restart` command implementation** with service-specific, project-specific, and all-service restart capabilities
- **Comprehensive test coverage across all packages** including installers, logging, projects, services, system, templates, and nginx templates
- **Fixed PHP-FPM process counting** in `chauf status --detail` to correctly report master and worker processes instead of showing 0
- **Enhanced contributor onboarding documentation** with comprehensive Go basics, AI workflow guidance, and development patterns in `docs/CONTRIBUTING.md`
- **Release checklist documentation** with comprehensive build, test, and deployment procedures in `docs/RELEASE_CHECKLIST.md`
- **CLI ergonomics improvement** with `-v` shorthand for `--version` flag in main command
- **Comprehensive `chauf doctor` command** for dependency validation and health checking across system, PHP build, SSL, network, and DNS dependencies
- **Example project automation** with automatic creation and linking after service installation, providing immediate testing environment for new users

### Project-Level PHP-FPM Architecture
- **Project-specific FPM control**: `--dedicated-fpm` flag for isolated PHP-FPM pools per project
- **Shared FPM by default**: Resource-efficient PHP-FPM pooling per PHP version when no dedicated flag is used
- **Mixed strategy support**: Coexistence of shared and dedicated FPM pools in the same workspace
- **Enhanced project config**: New `fpm:` section with `dedicated: true/false` and socket path tracking
- **Automatic socket management**: Correct socket path resolution based on project FPM settings
- **Service separation**: Clear distinction between global (shared) and project-specific (dedicated) services in status output

### Multi-Domain Support for Single Projects ✅
- **Multiple domains per project**: `--alias` flag during linking to add additional domains (e.g., `admin.test`, `api.test`)
- **Dynamic alias management**: `chauf link --alias` command to add domains to existing projects without unlinking
- **Alias removal**: `--alias` flag in unlink command to remove specific domains from projects
- **White-label development**: Support for multiple brands/domains pointing to same project directory
- **Backward compatibility**: Existing single-domain configurations continue to work unchanged
- **Multi-domain SSL certificates**: SAN certificates covering all SSL-enabled domains (primary + aliases)
- **Enhanced SSL support**: Trusted certificates via mkcert, automatic regeneration, per-alias SSL control
- **nginx integration**: Automatic generation of multi-domain server blocks in nginx configuration
- **Enhanced display**: `chauf links` shows primary domain and aliases with SSL indicators `(*)`
- **Improved unlink confirmation**: Shows all domains (primary + aliases) with individual SSL status

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
- Documented host packages (libzip, libjpeg, libpng, freetype, libxml2, curl, zlib, bzip2, readline, libmagickwand, libsodium) required before compiling PHP
- PHP installer validates `pkg-config` + all required libraries (libzip/libjpeg/libpng/freetype/libxml2/libcurl/zlib/libxslt/readline/MagickWand/libsodium) before starting long builds so users get actionable remediation early
- PHP runtimes ship with GNU Readline enabled so PsySH / `php artisan tinker` arrow keys and history behave like system PHP builds
- PHP builds enable mysqlnd-based `mysqli` and `PDO_MySQL` extensions so database-backed apps run immediately after `chauf install php`
- **Enhanced PHP legacy dependency matrix** with version-specific constraints for PHP 7.4/8.0 (libxml, ImageMagick, libcurl compatibility)
- **Legacy dependency validation** that warns about incompatible system package versions for legacy PHP versions

### Self-update
- `chauf self-update` fetches latest git release
- `chauf self-update --dev` rebuilds from current repo when run inside it

### CLI Infrastructure
- Structured logging enforced across all commands (`lib.Logger` only; usage/help text excluded)
- **Enhanced table formatting** with dynamic column widths, alignment options, and color-coded status indicators
- **Comprehensive service health monitoring** with process information, uptime, memory usage, and resource tracking

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
1. **Documentation sync discipline** – ongoing effort to keep README, AGENTS, and this tracker aligned
2. **Command documentation updates** – updating sites/constants.ts and site documentation to match current CLI implementation

## 4. Planned 📋
| Priority | Item | Notes |
|----------|------|-------|
| P2 | Enhanced service monitoring and logging | Add `chauf logs <service>` command and improved service health monitoring with historical data |
| P3 | Project management utilities | Add `chauf clean` for maintenance, and better project migration tools (doctor completed) |
| P4 | Universal service update management | `chauf update <service>` for PHP, Composer, nginx with version checking, backup, and rollback |
| P4 | CLI command expansion and enhancement | Add new commands and flags for improved user experience and functionality |

**CLI Enhancement**: ✅ Added `-v` shorthand for `--version` flag in main command

**Potential Command Enhancements**:
- `chauf logs <service>` - View service logs (nginx, PHP-FPM errors)
- `chauf config <service>` - Edit/view service configuration files
- `chauf clean` - Clean cache, old backups, temporary files
- `chauf export/import` - Export/import configuration and linked projects
- Add `--force` flag to `chauf link` for overwriting existing configurations
- Add `--quiet` flag to reduce verbosity across commands
- Add `--json` output format for scripting/automation

**Recently Completed Commands**:
- ✅ `chauf doctor` - Comprehensive health check and troubleshooting command with:
  - System dependencies validation (git, curl, tar, gcc, make, pkg-config)
  - PHP build dependencies checking (libzip, libjpeg, libpng, freetype, libxml2, libxslt, readline, MagickWand, gmp, libcurl)
  - SSL certificate management verification (openssl, mkcert with trusted certificates)
  - Network and firewall dependencies (iptables, port availability, conflict detection)
  - DNS configuration validation (dnsmasq, .test domain resolution)
  - Smart port detection that ignores Chauffeur-managed services
  - **Enhanced auto-fix workflow** with fix plan collection, user confirmation, and safe execution
  - Distribution-specific package commands (Fedora: dnf, Ubuntu/Debian: apt, Arch: pacman)
  - Priority-based fix execution (errors first, then warnings)
  - Comprehensive flag support (--check-all, --check-*, --fix, --auto-fix, --verbose, --quiet)
  - Cross-platform support with detailed fix suggestions for each Linux distribution

## 5. Future Plans 🔮
*Low priority items that will be addressed in future releases - no ETA.*

### Auto-Start Service Integration
- **Systemd service integration** for automatic Chauffeur startup on machine boot
- **User-level systemd services** that don't require root privileges (`systemctl --user`)
- **CLI command extensions**: `chauf start --enable-autostart`, `chauf status --autostart`
- **Configuration integration**: Auto-start settings in `config/chauffeur.yaml`
- **Service template generation**: Dynamic systemd service file creation in `~/.chauffeur/systemd/`
- **Per-service control**: Enable/disable autostart for nginx, specific PHP-FPM versions, projects
- **Graceful shutdown handling** during user logout/system reboot
- **Status monitoring** with systemd integration for service health checking
- **Project-specific autostart**: `chauf link --autostart` for individual project services

**Implementation Status**: Design complete in `docs/AUTOSTART_DESIGN.md`
**Key Files**: `cli/commands/autostart.go` (new), `cli/internal/config/config.go` (extend), systemd templates

### GD Extension for Legacy PHP (Future Enhancement)
- **Complete bundled GD extension support for PHP 7.4/8.0**
- **Validate GD extension functionality across all PHP versions** (7.4, 8.0, 8.1, 8.2, 8.3, 8.4)
- **Add GD extension validation tests** to verify common PHP GD functions work correctly

**Current Status**: Infrastructure is in place with interactive prompting, patching system, and graceful failure handling. Modern PHP (8.1+) works perfectly. Legacy PHP compilation fails during `make` phase due to function pointer compatibility issues between old PHP GD source and modern libgd libraries.

**Technical Details**: Function signature mismatches between PHP 7.4/8.0 GD source and modern libgd libraries require additional compiler flags or libgd version constraints. The build process successfully initiates GD compilation but fails during final linking.

**Future Approach**: Will investigate alternative compilation strategies, potentially targeting specific libgd versions or implementing additional compatibility layers for legacy PHP versions.

## 6. GD Extension Technical Status
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

## 6. Archived Documentation Features

### Commands Removed from Documentation (Never Implemented)
The following commands were documented but **never implemented** in the actual CLI:

#### `chauf config` - ARCHIVED
- **Status**: ❌ **Never implemented** - No traces found in codebase
- **Documentation References**: Was referenced in sites/constants.ts and site documentation
- **Intended Purpose**: Unknown - documentation was vague about configuration management
- **Analysis**: No `RunConfig()` function exists in cli/commands/
- **Recommendation**: Remove from all documentation, consider as future enhancement

#### `chauf env` - ARCHIVED
- **Status**: ❌ **Never implemented** - No traces found in codebase
- **Documentation References**: Was referenced in sites/constants.ts and site documentation
- **Intended Purpose**: Unknown - documentation was vague about environment management
- **Analysis**: No `RunEnv()` function exists in cli/commands/
- **Recommendation**: Remove from all documentation, consider as future enhancement

### Analysis Summary
- **Root Cause**: Documentation was written based on planned features that were never implemented
- **Impact**: Users encountering "command not found" errors when following documentation
- **Action Taken**: Commands archived in this section with recommendations for future consideration

## 7. Known Issues / Tech Debt
- Repo currently contains built binaries (`chauf`, `main`) checked in by mistake – remove and add to `.gitignore`
- DNS setup path requires clearer rollback instructions when NetworkManager/systemd-resolved changes are applied manually
- **RESOLVED**: Documentation synchronization completed - non-existent commands removed from docs/constants.ts

## 8. Testing & QA
- ✅ **`go test ./...` now green on Go 1.22+** with comprehensive unit test coverage across all packages
- Integration tests should stub HOME to temp directories, avoiding host mutation
- CI (`.github/workflows/go-tests.yml`) runs the suite on every pull request to `release`
- **New test coverage areas**: installers, logging, projects, services, system, templates, nginx templates with proper error handling and edge case validation

## 9. Release Readiness Checklist
- [x] `go test ./...` passes locally with comprehensive test coverage
- [ ] README.md / docs/TODO_STATUS.md / AGENTS.md agree on feature status
- [ ] No compiled binaries or caches tracked in git
- [ ] Logging contract enforced across commands
- [ ] dnsmasq/NetworkManager instructions reviewed and verified
- [x] GD extension infrastructure complete (legacy PHP compilation is future enhancement)

## 11. Future Feature Roadmap

### High Priority (P1) - Essential for Production Readiness

#### Enhanced Service Monitoring & Logging
- **Description**: Add comprehensive service monitoring with historical data, metrics, and log aggregation
- **Commands to Implement**:
  - `chauf logs <service>` - Enhanced with filtering, historical analysis
  - `chauf monitor` - Real-time service metrics dashboard
  - `chauf metrics` - Performance statistics and resource usage
- **Complexity**: Medium
- **User Value**: High - Critical for production debugging and performance tuning
- **Implementation Recommendation**: Build on existing logs infrastructure, add metrics collection

#### Configuration Management System
- **Description**: Implement `chauf config` command for workspace and project configuration management
- **Commands to Implement**:
  - `chauf config show [--project <slug>]` - Display current configuration
  - `chauf config set <key> <value> [--project <slug>]` - Set configuration values
  - `chauf config validate [--project <slug>]` - Validate configuration files
  - `chauf config export [--project <slug>]` - Export configuration to file
  - `chauf config import <file> [--project <slug>]` - Import configuration from file
  - `chauf config reset [--project <slug>]` - Reset to defaults
- **Complexity**: Medium
- **User Value**: High - Essential for advanced configuration and automation
- **Implementation Recommendation**: Create config management library, integrate with existing YAML structures

#### Environment Management (`chauf env`)
- **Description**: Implement environment variable management for projects
- **Commands to Implement**:
  - `chauf env list [--project <slug>]` - List environment variables
  - `chauf env set <key> <value> [--project <slug>]` - Set environment variable
  - `chauf env unset <key> [--project <slug>]` - Remove environment variable
  - `chauf env import <file> [--project <slug>]` - Import from .env file
  - `chauf env export [--project <slug>]` - Export to .env file
- **Complexity**: Low-Medium
- **User Value**: High - Critical for modern development workflows
- **Implementation Recommendation**: Build on existing project configuration system

### Medium Priority (P2) - Quality of Life Improvements

#### Advanced Service Management
- **Description**: Enhanced service control with better process management and dependencies
- **Commands to Implement**:
  - `chauf services status` - Detailed service health with dependency graph
  - `chauf services dependencies` - Show service dependency relationships
  - `chauf services heal` - Auto-repair common service issues
  - `chauf services benchmark` - Performance benchmarking tools
- **Complexity**: Medium
- **User Value**: Medium-High
- **Implementation Recommendation**: Extend existing service manager

#### Project Management Utilities
- **Description**: Advanced project lifecycle management and utilities
- **Commands to Implement**:
  - `chauf project create <name> [--template]` - Create new project from template
  - `chauf project clone <source> <destination>` - Clone project configuration
  - `chauf project backup <slug> [--include-data]` - Create project backups
  - `chauf project restore <backup-file>` - Restore from backup
  - `chauf project archive <slug>` - Archive inactive projects
- **Complexity**: Medium
- **User Value**: Medium
- **Implementation Recommendation**: Build on existing linking/unlinking infrastructure

#### Enhanced Security & SSL Management
- **Description**: Advanced SSL certificate management with automation
- **Commands to Implement**:
  - `chauf ssl status [--project <slug>]` - Detailed SSL certificate status
  - `chauf ssl renew [--project <slug>] [--force]` - Manual certificate renewal
  - `chauf ssl import <cert-file> <key-file> [--project <slug>]` - Import custom certificates
  - `chauf ssl generate [--wildcard] [--project <slug>]` - Advanced certificate generation
- **Complexity**: Medium
- **User Value**: Medium
- **Implementation Recommendation**: Extend existing secure/unsecure commands

### Low Priority (P3) - Advanced Features

#### Universal Service Update Management
- **Description**: Unified update system for all Chauffeur-managed services
- **Commands to Implement**:
  - `chauf update <service> [--dry-run] [--backup]` - Update specific service
  - `chauf update all [--dry-run] [--backup]` - Update all services
  - `chauf update rollback <service> <version>` - Rollback to previous version
  - `chauf update list-available [--service]` - List available updates
- **Complexity**: High
- **User Value**: Medium
- **Implementation Recommendation**: Create unified update framework with backup/rollback

#### Advanced Debugging & Profiling
- **Description**: Professional debugging and profiling tools
- **Commands to Implement**:
  - `chauf debug start <service>` - Enable debugging mode for service
  - `chauf debug profile <service> [--duration]` - Performance profiling
  - `chauf debug trace <service> [--filter]` - Request tracing and debugging
  - `chauf debug dump <service>` - Service state dump for analysis
- **Complexity**: High
- **User Value**: Low-Medium (niche, but valuable for advanced users)
- **Implementation Recommendation**: Build on existing service management and logging

#### Auto-Start Service Integration (Design Complete)
- **Description**: Systemd integration for automatic startup on system boot
- **Status**: Design documented in `docs/AUTOSTART_DESIGN.md`
- **Commands to Implement**:
  - `chauf autostart enable [--service]` - Enable auto-start
  - `chauf autostart disable [--service]` - Disable auto-start
  - `chauf autostart status` - Show auto-start status
  - `chauf autostart list` - List all auto-start services
- **Complexity**: Medium-High (system integration)
- **User Value**: Medium
- **Implementation Recommendation**: Follow existing design documents

### Implementation Complexity Analysis

#### Low Complexity (2-4 weeks each)
- **Environment Management**: Builds on existing project config system
- **Enhanced SSL Management**: Extension of existing secure/unsecure commands
- **Project Management Utilities**: Leverages existing linking infrastructure

#### Medium Complexity (4-8 weeks each)
- **Service Monitoring**: Requires metrics collection and storage system
- **Configuration Management**: New system but well-defined requirements
- **Advanced Service Management**: Extensions to existing service manager
- **Auto-Start Integration**: Systemd integration with security considerations

#### High Complexity (8-16 weeks each)
- **Universal Update System**: Complex backup/rollback and dependency management
- **Advanced Debugging**: Requires deep service integration and profiling tools

### Recommended Implementation Order

1. **Phase 1** (Next 2-3 months): P1 - Configuration Management and Environment Management
   - Foundation for other features
   - High user impact
   - Enables automation and advanced workflows

2. **Phase 2** (3-6 months): P1 - Enhanced Service Monitoring and P2 - Project Management Utilities
   - Production readiness features
   - Improves debugging and maintenance workflows

3. **Phase 3** (6-12 months): P2 - Advanced features and P3 - Universal updates
   - Advanced user features
   - Professional development workflows

4. **Phase 4** (12+ months): P3 - Debugging and profiling tools
   - Specialized features for power users
   - Advanced production troubleshooting

### Resource Requirements

#### For P1 Features (Configuration & Environment Management):
- **Development**: 1-2 developers
- **Testing**: Comprehensive unit and integration tests
- **Documentation**: Updated CLI reference and user guides
- **Estimated effort**: 6-10 developer-weeks

#### For Full Roadmap Completion:
- **Development**: 2-3 developers over 12-18 months
- **Infrastructure**: Monitoring backend, metrics storage
- **Testing**: End-to-end testing suite expansion
- **Documentation**: Complete documentation overhaul
- **Estimated total effort**: 40-60 developer-weeks

### Success Metrics

#### User Experience Metrics:
- **Command discovery time**: Reduce from current to <30 seconds for any command
- **Configuration changes**: Reduce from manual file editing to single commands
- **Troubleshooting time**: Reduce from hours to minutes with enhanced monitoring
- **Onboarding time**: New users productive in <15 minutes

#### Technical Metrics:
- **Service reliability**: 99.9% uptime with auto-healing
- **Performance monitoring**: Real-time metrics with 1-minute resolution
- **Backup/restore**: <5 minutes for complete project restore
- **Update safety**: 100% rollback success rate

## 12. How to Help
- Help expand service orchestration tests/logging and harden dnsmasq/NetworkManager flows
- Improve tests around `chauf link`/`links` to avoid double-run conflicts
- Document real-world setups (distro, Go version, dnsmasq config) in issues to broaden coverage
- Review AGENTS.md and propose clarifications before building new features

_Last updated: 2025-11-19T21:58:00Z_
