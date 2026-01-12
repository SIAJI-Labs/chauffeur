# Chauffeur Documentation Index

> **Primary Purpose**: This directory contains comprehensive documentation for the Chauffeur project, serving as the single source of truth for developers, contributors, and AI agents.

## Quick Navigation

| Document | Purpose | Audience |
|----------|---------|----------|
| [README.md](#project-summary) | This file - Project overview and documentation index | Everyone |
| [Architecture.md](Architecture.md) | System architecture, design decisions, and technical patterns | Developers, Architects |
| [Conventions.md](Conventions.md) | Coding standards, development workflow, and required practices | Contributors, AI Agents |
| [TechStack.md](TechStack.md) | Technology stack, dependencies, and version requirements | Developers, DevOps |
| [Plan.md](Plan.md) | Project roadmap, implementation phases, and milestones | Maintainers, Contributors |
| [History.md](History.md) | Chronological log of all significant changes and executions | Everyone, AI Agents |
| [TODO_STATUS.md](TODO_STATUS.md) | Living status board for features, debt, and priorities | Maintainers, Contributors |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines and getting started | New Contributors |
| [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md) | Release process and quality assurance | Maintainers |
| [design/autostart.md](design/autostart.md) | Design document for auto-start service integration | Developers |
| [design/service-updates.md](design/service-updates.md) | Design document for service update management system | Developers |

---

## Project Summary

### What is Chauffeur?

**Chauffeur** is a Linux-first CLI tool that provides per-project PHP development services (nginx, PHP-FPM, dnsmasq-managed `.test` domains) without touching system prefixes. Everything lives under `~/.chauffeur/` so multiple projects can coexist safely.

**Inspired by**: [Laravel Valet](https://laravel.com/docs/valet) & [Beyond Code Herd](https://herd.laravel.com/)

**Key Difference**: Unlike Valet/Herd (macOS-focused), Chauffeur is designed specifically for Linux, respecting Linux packaging norms while providing the same developer experience.

### Background & Purpose

**Why Chauffeur Exists:**

Linux lacks a native equivalent to Laravel Valet or Laravel Herd. Developers migrating from macOS to Linux miss the simplicity of:
- One command to link projects and get a `.test` domain
- Automatic PHP version management per project
- Zero-configuration nginx routing
- Isolated development environments

Chauffeur fills this gap by:
- Bringing Valet/Herd ergonomics to Linux
- Keeping everything in user workspace (`~/.chauffeur/`)
- Providing per-project PHP version isolation
- Enabling automatic `.test` domain routing via dnsmasq

### Core Purpose

Provide Linux developers with:
1. **Instant `.test` domains** for any project directory
2. **Isolated PHP runtimes** - Run PHP 7.4, 8.0, 8.1, 8.2, 8.3, 8.4 simultaneously
3. **Project-aware shims** - `php` and `composer` automatically detect project PHP version
4. **No system mutation** - Everything under `~/.chauffeur/`, no `/usr` or `/etc` pollution
5. **Manual control** - Explicit project registration via `chauf link`, no auto-scanning

---

## For AI Agents & Contributors

### Required Reading Order

**CRITICAL**: All AI agents and contributors MUST read documents in this order:

1. **[Conventions.md](Conventions.md)** (FIRST) - Development conventions including:
   - History logging requirements (every execution must be logged)
   - Code style and patterns
   - Testing standards
   - Documentation synchronization rules

3. **[Architecture.md](Architecture.md)** - Technical architecture:
   - System design and component interaction
   - Data flow and service orchestration
   - Configuration management
   - Multi-domain architecture
   - PHP-FPM strategies (shared vs dedicated)

4. **[TechStack.md](TechStack.md)** - Technology stack:
   - Go version and dependencies
   - Required host tools
   - Service versions (PHP, nginx, Composer)
   - Build and compilation requirements

5. **[Plan.md](Plan.md)** - Implementation roadmap:
   - Current development focus
   - Short-term and long-term goals
   - Feature priorities
   - Implementation phases

6. **[History.md](History.md)** - Change log:
   - Chronological record of all significant changes
   - Execution tracking for reproducibility
   - Decision history and rationale

### Key Policies for AI Agents

**1. Commit Policy (STRICT)**
- **NEVER** make git commits without explicit user instruction
- **NEVER** push changes without explicit user permission
- **ALWAYS** present changes for review before committing
- See [AGENTS.md Section 17](../AGENTS.md#17-commit-policy-for-ai-agents) for details

**2. Development Workspace**
- **ALWAYS** build and test in isolated temporary workspace
- See [AGENTS.md Section 15.1](../AGENTS.md#151-isolated-development-workspace) for workflow
- Example:
  ```bash
  export CHAUFFEUR_TEMP_WS=$(mktemp -d)
  export HOME=$CHAUFFEUR_TEMP_WS
  go build -o $CHAUFFEUR_TEMP_WS/chauf cli/main.go
  $CHAUFFEUR_TEMP_WS/chauf [command]
  ```

**3. Documentation Synchronization**
- **IMMEDIATELY** update all documentation after code changes
- Update checklist:
  - [ ] [AGENTS.md](../AGENTS.md) - Command contracts and architectural changes
  - [ ] [README.md](../README.md) - Feature status, examples, installation
  - [ ] [TODO_STATUS.md](TODO_STATUS.md) - Move items between sections, mark progress
  - [ ] [sites/constants.ts](../sites/constants.ts) - **CRITICAL** - Single source of truth for CLI commands
  - [ ] [History.md](History.md) - Log the changes made

**4. Constants.ts Maintenance**
- When CLI commands change, **MUST** update `sites/constants.ts` immediately
- This file is the **single source of truth** for:
  - All CLI command definitions and examples
  - Command flags, usage patterns, and expected outputs
  - Feature descriptions and navigation constants
  - Documentation site command reference generation
- See [AGENTS.md Section 11.1](../AGENTS.md#11-critical-constants.ts-maintenance-requirement)

---

## Documentation Structure

### User-Facing Documentation

**Located in**: `sites/` directory (Next.js documentation site)

- Complete command reference
- Installation guides
- Troubleshooting guides
- Feature documentation

### Developer Documentation

**Located in**: `docs/` directory (this directory)

- Architecture and design documents
- Contribution guidelines
- Release checklists
- Implementation plans
- Historical logs

### Authoritative Documentation

**Located in**: Root directory

- [AGENTS.md](../AGENTS.md) - **THE** handbook for autonomous contributors
- [README.md](../README.md) - Project overview, quick start, feature status

---

## Core Principles (Quick Reference)

1. **Workspace-first**: All binaries, configs, sockets, and logs live under `~/.chauffeur/`
2. **Minimal host impact**: Prefer workspace changes. Print `sudo` commands for `/etc` edits; never modify directly.
3. **Manual project registration**: Projects are never auto-scanned. `chauf link` is the only way to register.
4. **Idempotent operations**: Re-running commands must be safe. Destructive actions require `--force`.
5. **Linux-focused**: Target Arch/Ubuntu/Debian; no macOS/Windows assumptions.
6. **No external env managers**: No Devbox, nix, Docker. Ships own runtimes in workspace.
7. **Documentation parity**: README.md, docs/TODO_STATUS.md, and AGENTS.md must describe code exactly.
8. **History logging**: Every significant execution/change must be logged in [History.md](History.md).

---

## Quick Links

### For New Users
- [Main README](../README.md) - Installation and quick start
- [Documentation Site](https://chauffeur.siaji.com/docs) - Complete user guides

### For Contributors
- [CONTRIBUTING.md](CONTRIBUTING.md) - Getting started and contribution guidelines
- [AGENTS.md](../AGENTS.md) - **MUST READ** - Development handbook
- [Conventions.md](Conventions.md) - Coding standards and practices

### For Maintainers
- [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md) - Release process
- [TODO_STATUS.md](TODO_STATUS.md) - Feature status and priorities
- [Plan.md](Plan.md) - Roadmap and milestones

### For AI Agents
- [AGENTS.md](../AGENTS.md) - **START HERE** - Authoritative handbook
- [Conventions.md](Conventions.md) - Development conventions and history logging
- [History.md](History.md) - Change log and execution tracking

---

## Document Maintenance

### Synchronization Requirements

All documentation must be kept in sync with the codebase:

1. **Code Changes**: Immediately update all relevant documentation
2. **Command Changes**: Update `sites/constants.ts` immediately
3. **Feature Changes**: Update feature status in `TODO_STATUS.md`
4. **Architecture Changes**: Update `Architecture.md` and `AGENTS.md`
5. **All Changes**: Log in `History.md`

### Documentation Linting

Before submitting changes:
- [ ] All examples tested against actual binary
- [ ] Command references match `sites/constants.ts`
- [ ] Feature status consistent across README, TODO_STATUS, AGENTS
- [ ] Architecture changes reflected in Architecture.md
- [ ] Changes logged in History.md

---

## Getting Help

### Documentation Issues
- If you find inconsistencies, open an issue or PR
- Cite the specific documents that disagree
- Include the expected vs actual behavior

### Code Questions
- Check [AGENTS.md](../AGENTS.md) first for architectural rules
- Review [Architecture.md](Architecture.md) for design patterns
- Look at existing code for implementation examples
- Check [CONTRIBUTING.md](CONTRIBUTING.md) for common patterns

### Feature Requests
- Review [Plan.md](Plan.md) for roadmap alignment
- Check [TODO_STATUS.md](TODO_STATUS.md) for existing items
- Consider the core principles before proposing changes

---

**Remember**: [AGENTS.md](../AGENTS.md) is the single source of truth. When in doubt, check there first.
