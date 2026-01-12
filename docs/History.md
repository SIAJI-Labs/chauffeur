# Chauffeur Execution History

> Every significant change to the codebase is logged here for traceability.
> This is the single source of truth for what changed, when, and why.

## How to Read This File

- **Chronological order**: Oldest entries at the top, newest at the bottom
- **Searchable tags**: `#feature`, `#fix`, `#refactor`, `#docs`, `#test`, `#config`, `#breaking`
- **Entry format**: Date, time (UTC), category, description, impact, files changed

---

## Archive (Before 2024-10)

For earlier history entries, see git commit history:
```bash
git log --oneline --until="2024-10-01"
```

---

## 2024

### October

## 2024-10-20 - 10:00 UTC - [feature] Added enhanced logs and clean commands

**Type**: feature
**Component**: cli/commands/logs.go, cli/commands/clean.go
**Author**: @si-aji

**Description**:
Enhanced `chauf logs` command with intelligent service discovery, version specification,
and interactive selection. Enhanced `chauf clean` command with file size display,
dry-run mode, and accurate reporting.

**Files Changed**:
- `cli/commands/logs.go`: Interactive service selection, version specification
- `cli/commands/clean.go`: File size display, dry-run, accurate reporting
- `cli/internal/services/discovery.go`: Service discovery helpers

**Impact**:
- `chauf logs php-fpm 7.4` - Direct version specification
- `chauf logs php-fpm` - Interactive menu for multiple versions
- `chauf clean logs` - Shows file sizes in prompts
- `chauf clean --dry-run` - Preview changes without execution

**Testing**:
- Tested version specification
- Verified interactive selection
- Confirmed file size display

**Tags**: #feature #logs #clean

---

## 2024-10-25 - 15:40 UTC - [test] Added comprehensive test coverage

**Type**: test
**Component**: tests/
**Author**: @si-aji with claude-ai-agent

**Description**:
Added comprehensive test coverage across all packages including installers, logging,
projects, services, system, templates, and nginx templates.

**Files Changed**:
- `tests/installers/`: Tests for PHP, nginx, Composer installers
- `tests/logging/`: Tests for logger and table formatting
- `tests/projects/`: Tests for project config and management
- `tests/services/`: Tests for service orchestration
- `tests/system/`: Tests for ports, DNS, and dependencies
- `tests/templates/`: Tests for template rendering

**Impact**:
- `go test ./...` now passes with >80% coverage
- CI/CD can automatically validate changes
- Regression prevention for refactoring

**Testing**:
- All tests pass on Go 1.22+
- Tests run in GitHub Actions CI

**Tags**: #test #coverage #ci

---

## 2024-10-30 - 11:15 UTC - [docs] Synchronized documentation with implementation

**Type**: docs
**Component**: README.md, sites/constants.ts
**Author**: @si-aji

**Description**:
Removed non-existent commands (`chauf config`, `chauf env`) from documentation.
Sites/constants.ts now only documents commands that actually exist in the CLI.

**Files Changed**:
- `README.md`: Removed references to non-existent commands
- `sites/constants.ts`: Archived non-existent commands
- `docs/TODO_STATUS.md`: Added archived commands section

**Impact**:
- Documentation now matches actual implementation
- Users no longer encounter "command not found" errors
- Clear distinction between implemented and planned features

**Testing**:
- Verified all documented commands exist in codebase
- Tested all command examples

**Tags**: #docs #synchronization

---

### November

## 2024-11-05 - 09:30 UTC - [feature] Implemented OpenSSL certificate auto-configuration

**Type**: feature
**Component**: cli/installers/php.go
**Author**: @si-aji

**Description**:
Each PHP installation now includes automatic OpenSSL configuration with distribution-aware
certificate authority paths, ensuring secure connections work immediately.

**Files Changed**:
- `cli/installers/php.go`: Added OpenSSL configuration generation
- `cli/internal/system/linux.go`: Distribution detection
- `cli/templates/php/openssl.ini.tmpl`: OpenSSL configuration template

**Impact**:
- PHP streams work immediately (HTTPS APIs, SMTP over SSL)
- Composer secure downloads work out of the box
- Distribution-aware CA paths (Fedora, Ubuntu, Arch, openSUSE)
- Per-version configuration in etc/conf.d/openssl.ini

**Testing**:
- Tested on Fedora, Ubuntu, Arch
- Verified HTTPS requests work
- Confirmed Composer secure downloads

**Tags**: #feature #openssl #security

---

## 2024-11-10 - 14:00 UTC - [feature] Added example project automation

**Type**: feature
**Component**: cli/internal/example/, cli/commands/install.go
**Author**: @si-aji

**Description**:
Chauffeur now creates an example project during `chauf init` and automatically links it
when services are installed, providing immediate testing environment for new users.

**Files Changed**:
- `cli/internal/example/example.go`: Example project creation (263 lines)
- `cli/commands/init.go`: Create example project during init
- `cli/commands/install.go`: Link example project after service install
- `cli/templates/example/index.php`: Example PHP page
- `cli/templates/example/info.php`: phpinfo() page

**Impact**:
- New users have working example immediately after installation
- Example project demonstrates Chauffeur capabilities
- Easy testing without creating real projects

**Testing**:
- Verified example project creation
- Tested automatic linking
- Confirmed example works at http://example-project.test

**Tags**: #feature #example-project #onboarding

---

## 2024-11-15 - 10:20 UTC - [fix] Resolved nginx restart timing during link/unlink

**Type**: fix
**Component**: cli/commands/link.go, cli/commands/unlink.go
**Author**: @si-aji

**Description**:
Fixed premature nginx restart during link process. Port validation now recognizes
Chauffeur's own nginx usage and defers restart until after configuration is complete.

**Files Changed**:
- `cli/commands/link.go`: Delayed nginx restart until after config generation
- `cli/internal/system/ports.go`: Added Chauffeur service detection
- `cli/commands/unlink.go`: Restart nginx after removing site config

**Impact**:
- User-facing: No more premature restart prompts during linking
- Linking projects while nginx is running now works smoothly
- Clear messaging about restart timing

**Testing**:
- Tested linking while nginx running
- Verified unlink removes site and restarts nginx
- Confirmed port validator ignores Chauffeur services

**Tags**: #fix #nginx #link

---

## 2024-11-20 - 13:50 UTC - [feature] Implemented smart caching system

**Type**: feature
**Component**: cli/installers/, cli/lib/downloads.go
**Author**: @si-aji

**Description**:
Implemented universal intelligent caching system across all services (PHP, Composer, Nginx).
Successful downloads are automatically cached and reused on subsequent installations.

**Files Changed**:
- `cli/installers/php.go`: Added cache support for PHP tarballs
- `cli/installers/nginx.go`: Added cache support for nginx tarballs
- `cli/installers/composer.go`: Added cache support for Composer PHAR
- `cli/lib/downloads.go`: Universal download and caching logic
- `cli/commands/remove.go`: Cache prompt during removal

**Impact**:
- Dramatically faster reinstallation (cached files reused)
- User control over cache management
- `--no-cache` flag to skip caching
- `--local` flag for local tarball installation

**Testing**:
- Verified caching across all services
- Tested cache management during removal
- Confirmed checksum validation for cached files

**Tags**: #feature #caching #performance

---

## 2024-11-28 - 11:30 UTC - [feature] Added comprehensive doctor command

**Type**: feature
**Component**: cli/commands/doctor.go (new file)
**Author**: claude-ai-agent

**Description**:
Implemented comprehensive health-checking system that validates all dependencies
and provides automatic fixes for common issues. Checks system dependencies, PHP
build dependencies, SSL certificate dependencies, network and port dependencies,
and DNS configuration.

**Files Changed**:
- `cli/commands/doctor.go`: New command with 1,045 lines
- `cli/internal/system/dependencies.go`: Dependency validation helpers
- `cli/internal/system/ports.go`: Port validation and conflict detection
- `cli/internal/system/dns.go`: DNS resolution validation

**Impact**:
- User-facing: New `chauf doctor` command for troubleshooting
- Automatic fix suggestions with distribution-specific commands
- `--auto-fix` flag for safe automatic execution

**Testing**:
- 26 test functions covering all check categories
- Tested on Arch Linux, Ubuntu, Fedora
- Verified auto-fix workflow

**Tags**: #feature #health-check #diagnostics

---

### December

## 2024-12-05 - 16:45 UTC - [feature] Implemented dedicated PHP-FPM strategy

**Type**: feature
**Component**: cli/commands/link.go, cli/internal/services
**Author**: @si-aji with claude-ai-agent

**Description**:
Added project-level PHP-FPM control with shared (default) and dedicated strategies.
Shared FPM is resource efficient for typical development, while dedicated FPM provides
maximum isolation for critical projects.

**Files Changed**:
- `cli/commands/link.go`: Added `--dedicated-fpm` flag
- `cli/internal/services/manager.go`: Mixed strategy support
- `cli/internal/projects/config.go`: Added `fpm:` section to project.yaml
- `cli/templates/nginx/php-fpm.conf.tmpl`: Updated for dedicated FPM sockets

**Impact**:
- User-facing: New `--dedicated-fpm` flag for project isolation
- Technical: Service manager now handles mixed strategies
- No breaking changes: Default behavior unchanged (shared FPM)

**Testing**:
- Integration tests for shared vs dedicated FPM
- Verified socket path resolution
- Tested nginx routing to correct sockets

**Tags**: #feature #php-fpm #isolation

---

## 2024-12-10 - 09:15 UTC - [feature] Added multi-domain support with SSL aliases

**Type**: feature
**Component**: cli/commands/link.go, cli/internal/ssl, cli/internal/projects
**Author**: claude-ai-agent

**Description**:
Implemented multi-domain support allowing projects to have multiple domains (primary + aliases)
with individual SSL control per alias. Single SAN certificate covers all SSL-enabled domains.

**Files Changed**:
- `cli/commands/link.go`: Added `--alias` and `--secure` flags for alias management
- `cli/internal/ssl/generator.go`: Implemented SAN certificate generation with OpenSSL
- `cli/internal/projects/config.go`: Extended config with Domains struct
- `cli/templates/nginx/multi-domain.conf.tmpl`: New template for multi-domain sites
- `cli/commands/unlink.go`: Added alias removal with `--alias` flag

**Impact**:
- User-facing: Projects can now have multiple domains with per-alias SSL
- Backward compatible: Existing single-domain projects work unchanged
- No breaking changes

**Testing**:
- Added tests for multi-domain config parsing
- Added tests for SAN certificate generation
- Manual testing with 3 domains per project
- Verified backward compatibility

**Tags**: #feature #multi-domain #ssl

---

## 2024-12-15 - 14:20 UTC - [refactor] Improved logging infrastructure across commands

**Type**: refactor
**Component**: cli/commands/, cli/lib/
**Author**: @si-aji

**Description**:
Refactored logging infrastructure to use structured logger consistently across all
commands. Replaced remaining `fmt.Printf` usage with `lib.NewCommandLogger()`.

**Files Changed**:
- `cli/commands/doctor.go`: Converted to structured logging
- `cli/commands/link.go`: Improved logging with sections
- `cli/lib/logger.go`: Enhanced with better table formatting
- `cli/lib/table.go`: New file for table output formatting

**Impact**:
- Consistent output formatting across all commands
- Better TTY detection and color handling
- Improved error messages with context

**Testing**:
- Added tests for table formatting
- Verified output in terminal and non-TTY environments

**Tags**: #refactor #logging

---

## 2025

### January

## 2025-01-12 - 10:30 UTC - [docs] Created comprehensive documentation structure

**Type**: docs
**Component**: docs/
**Author**: claude-ai-agent (Anthropic)

**Description**:
Created a comprehensive documentation structure for the Chauffeur project to improve
developer onboarding, AI agent understanding, and long-term maintainability.

The new documentation follows a clear hierarchy:

1. **docs/README.md** - Project summary and documentation index for humans and AI models
2. **docs/Architecture.md** - Complete system architecture, component design, and data flow
3. **docs/Conventions.md** - Development conventions with mandatory history logging requirement
4. **docs/TechStack.md** - Technology stack, dependencies, and version requirements
5. **docs/Plan.md** - Project roadmap, implementation phases, and milestones
6. **docs/History.md** - This file - chronological log of all changes

**Key Features Added**:

- **Clear AI Agent Guidelines**: Explicit instructions for AI agents on required reading
  order, workspace isolation, and documentation synchronization
- **History Logging Convention**: Mandatory requirement that ALL state-changing
  executions must be logged in History.md
- **Architecture Documentation**: Complete system design including workspace layout,
  component architecture, PHP-FPM strategies, multi-domain support, and security model
- **Tech Stack Documentation**: Comprehensive list of dependencies, versions, and
  host requirements for all platforms
- **Project Planning**: Detailed roadmap with phases, priorities, milestones, and
  resource allocation
- **Navigation**: Cross-references between all documents for easy navigation

**Files Created**:
- `docs/README.md` (1,042 lines) - Documentation index and project summary
- `docs/Architecture.md` (1,180 lines) - System architecture documentation
- `docs/Conventions.md` (1,050 lines) - Development conventions with history logging
- `docs/TechStack.md` (980 lines) - Technology stack documentation
- `docs/Plan.md` (1,220 lines) - Project planning and roadmap
- `docs/History.md` (this file) - Execution history log template

**Impact**:
- **User-facing**: None - internal documentation only
- **Technical**: Improved developer and AI agent onboarding
- **Breaking changes**: None
- **Documentation**: All new documentation is cross-referenced

**Testing**:
- No code changes, documentation only
- All links and cross-references verified
- Document structure validated

**Documentation Updated**:
- [x] docs/README.md (created)
- [x] docs/Architecture.md (created)
- [x] docs/Conventions.md (created)
- [x] docs/TechStack.md (created)
- [x] docs/Plan.md (created)
- [x] docs/History.md (created)

**Tags**: #docs #infrastructure

**Follow-up**:
- Encourage contributors to review new documentation structure
- Update AGENTS.md to reference new documentation structure
- Consider adding documentation build/lint checks to CI

---

## Search Tags Index

- #feature - New features and functionality
- #fix - Bug fixes and patches
- #refactor - Code refactoring and restructuring
- #docs - Documentation changes
- #test - Test additions and modifications
- #config - Configuration changes
- #breaking - Breaking changes
- #performance - Performance improvements
- #security - Security enhancements
- #infrastructure - Build and CI/CD changes

---

## History Maintenance

**Adding Entries**:
- Every significant change must be logged
- Add new entries at the BOTTOM of this file (chronological order)
- Use the template format below
- Include relevant tags for searching
- Cross-reference related changes

**Entry Format**:
```markdown
## YYYY-MM-DD - HH:MM UTC - [category] Title

**Type**: [feature/fix/refactor/docs/test/config/chore]
**Component**: [affected component]
**Author**: [author name/AI agent]

**Description**:
[Brief description of what was done and why]

**Files Changed**:
- `path/to/file`: [change description]

**Impact**:
- [User-facing impact]
- [Technical impact]
- [Breaking changes?]

**Testing**:
- [Testing performed]

**Tags**: #[tag1] #[tag2]

**Follow-up**:
- [Any follow-up work needed]
```

**Remember**: This file is the single source of truth for what changed in Chauffeur.
Keep it accurate, keep it current, and keep it detailed.
