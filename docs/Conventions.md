# Chauffeur Development Conventions

> **Last Updated**: 2025-01-12
> **Maintainer**: @si-aji
> **Status**: Active - All contributors and AI agents MUST follow

## Table of Contents

1. [Overview](#overview)
2. [Critical Convention: History Logging](#critical-convention-history-logging)
3. [Git Workflow](#git-workflow)
4. [Code Style](#code-style)
5. [Logging Standards](#logging-standards)
6. [Testing Conventions](#testing-conventions)
7. [Documentation Conventions](#documentation-conventions)
8. [AI Agent Conventions](#ai-agent-conventions)
9. [Development Workflow](#development-workflow)
10. [Code Review Checklist](#code-review-checklist)

---

## Overview

This document defines the conventions and standards that all contributors (human and AI) must follow when working on the Chauffeur project. These conventions ensure consistency, maintainability, and traceability across the codebase.

### Guiding Principles

1. **Explicit is better than implicit** - Make decisions visible and documented
2. **Everything is tracked** - Every significant change must be logged in History
3. **Consistency over cleverness** - Follow established patterns
4. **Test first, test often** - Maintain comprehensive test coverage
5. **Document as you code** - Keep documentation synchronized

---

## Critical Convention: History Logging

### The Golden Rule

> **EVERY execution that changes the state of the codebase, workspace, or configuration MUST be logged in `docs/History.md`**

### What Must Be Logged

**All of the following MUST be logged in [History.md](History.md):**

1. **Code Changes**
   - New features implemented
   - Bug fixes applied
   - Refactoring work
   - Performance improvements
   - Breaking changes

2. **Configuration Changes**
   - Global config changes
   - Project config changes
   - Template modifications
   - Schema changes

3. **Installation/Uninstallation**
   - Services installed (nginx, PHP, Composer)
   - PHP versions added/removed
   - Workspace initialization
   - Project linking/unlinking

4. **System Changes**
   - Port forwarding rules
   - DNS configuration
   - Service restarts
   - Process management

5. **Documentation Changes**
   - Documentation updates
   - README changes
   - AGENTS.md updates
   - New documentation files

6. **AI Agent Executions**
   - AI agent operations
   - Automated refactoring
   - Test generation
   - Documentation generation

### What Does NOT Need to Be Logged

**Minor operations that don't affect state:**
- Reading files
- Running `chauf status`
- Viewing logs (`chauf logs`)
- Dry-run operations
- Help commands

### History Entry Format

**Required Format for History.md:**

```markdown
## YYYY-MM-DD - HH:MM UTC - [Category] Title

**Type**: [feature/fix/refactor/docs/test/config/chore]
**Component**: [affected component]
**Author**: [author name/AI agent]

**Description**:
[Brief description of what was done and why]

**Files Changed**:
- `path/to/file1.ts`: [change description]
- `path/to/file2.ts`: [change description]

**Impact**:
- [User-facing impact]
- [Technical impact]
- [Breaking changes?]

**Testing**:
- [Testing performed]
- [Tests added/modified]

**Documentation Updated**:
- [x] AGENTS.md
- [x] README.md
- [x] TODO_STATUS.md
- [x] sites/constants.ts
- [x] docs/History.md

**Follow-up**:
- [Any follow-up work needed]
```

**Example Entry:**

```markdown
## 2025-01-12 - 10:30 UTC - [feature] Add multi-domain SSL support

**Type**: feature
**Component**: cli/commands/link.go, cli/internal/ssl
**Author**: claude-ai-agent

**Description**:
Added support for multiple SSL-enabled domains per project with SAN certificates.
Projects can now have multiple domains (primary + aliases) with individual SSL
control per alias. Single SAN certificate covers all SSL-enabled domains.

**Files Changed**:
- `cli/commands/link.go`: Added --alias and --secure flags for alias management
- `cli/internal/ssl/generator.go`: Implemented SAN certificate generation
- `cli/internal/projects/config.go`: Extended config with Domains struct
- `cli/templates/nginx/multi-domain.conf.tmpl`: New template for multi-domain sites

**Impact**:
- User can now link multiple domains to single project
- SSL can be enabled/disabled per alias
- Backward compatible with existing single-domain projects
- No breaking changes

**Testing**:
- Added tests for multi-domain config parsing
- Added tests for SAN certificate generation
- Manual testing with 3 domains per project
- Verified backward compatibility with existing projects

**Documentation Updated**:
- [x] AGENTS.md (multi-domain architecture section)
- [x] README.md (multi-domain usage examples)
- [x] TODO_STATUS.md (marked multi-domain as completed)
- [x] sites/constants.ts (updated link command examples)
- [x] docs/History.md (this entry)

**Follow-up**:
- Consider adding domain management commands (add/remove)
- Monitor for certificate generation edge cases
```

### Logging Enforcement

**For AI Agents:**
- AI agents MUST automatically update History.md after any state-changing operation
- History updates are part of the completion criteria for tasks
- AI agents should fail if unable to write to History.md

**For Humans:**
- Developers must update History.md as part of every commit
- PRs must include History.md updates to be merged
- CI should check that History.md is updated for significant changes

### History Organization

**Structure of History.md:**

```markdown
# Chauffeur Execution History

> Every significant change to the codebase is logged here for traceability.

## [Current Year]

### [Month]

## [Previous Years]
```

**Searchable Tags:**
- `#feature` - New features
- `#fix` - Bug fixes
- `#refactor` - Code refactoring
- `#docs` - Documentation changes
- `#test` - Test additions/changes
- `#config` - Configuration changes
- `#breaking` - Breaking changes

---

## Git Workflow

### Commit Message Format

**Follow [Conventional Commits](https://www.conventionalcommits.org/):**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples:**

```
feat(link): add multi-domain support with SSL aliases

Allow projects to have multiple domains with per-alias SSL control.
Implements SAN certificate generation for all SSL-enabled domains.

Closes #123
```

```
fix(nginx): resolve duplicate service listings in status output

Prevent shared PHP-FPM services from appearing multiple times.
Deduplication based on service name and socket path.

Fixes #456
```

### Branch Strategy

**Main Branch:** `main` - Production-ready code
**Development Branch:** `refactor` - Active development
**Feature Branches:** `feature/<name>` - Individual features

**Workflow:**
```bash
# Start feature branch
git checkout -b feature/multi-domain-support

# Make changes
# ... (update History.md for each change)

# Commit with conventional format
git add -A
git commit -m "feat(link): add multi-domain support"

# Push and create PR
git push origin feature/multi-domain-support
```

### Merge Policy

**Allowed Merge Methods:**
- **Squash and merge** - Preferred for feature branches
- **Merge commit** - For feature branches with meaningful history

**Forbidden:**
- **Rebase merge** - Breaks git blame

### Commit Policy for AI Agents

**CRITICAL** - From [AGENTS.md Section 17](../AGENTS.md#17-commit-policy-for-ai-agents):

> AI agents MUST NOT make git commits without explicit user instruction.

**AI Agent Workflow:**
1. Make changes and stage them
2. Present changes for user review
3. Ask for commit approval
4. Commit with user-approved message

**Exception:** Only for critical hotfixes with immediate user notification.

---

## Code Style

### Go Conventions

**Formatting:**
- Use `gofmt` for all Go code
- Run `go fmt ./...` before committing
- Maximum line length: 120 characters (soft limit)

**Naming:**
- Use `MixedCaps` for exported names
- Use `mixedCaps` for private names
- Use `PascalCase` for types
- Use `camelCase` for variables and functions

**Examples:**

```go
// Good
type ProjectConfig struct {
    ProjectPath string
    PHPVersion  string
}

func LoadProjectConfig(path string) (*ProjectConfig, error) {
    // ...
}

// Bad
type project_config struct {
    project_path string
    php_version  string
}

func load_project_config(path string) (*project_config, error) {
    // ...
}
```

### File Organization

**Package Structure:**

```
cli/
├── commands/          # Command implementations
│   ├── link.go
│   ├── start.go
│   └── ...
├── lib/              # Shared utilities
│   ├── logger.go
│   └── helpers.go
└── internal/         # Internal packages
    ├── projects/
    ├── services/
    └── system/
```

**File Naming:**
- Use `snake_case` for file names
- One major type per file when possible
- Test files: `*_test.go`

### Comments

**When to Comment:**
- Exported functions must have comments
- Non-obvious logic must be explained
- Complex algorithms need explanations
- TODOs and FIXMEs must be documented

**Comment Style:**

```go
// LoadProjectConfig reads and parses a project configuration file.
// It validates the configuration and returns an error if invalid.
func LoadProjectConfig(path string) (*ProjectConfig, error) {
    // ...
}

// HACK: Temporary workaround for PHP 7.4 GD extension compilation
// Remove when upstream PHP is fixed.
func patchGDCompatibility(buildDir string) error {
    // ...
}
```

### Error Handling

**Principles:**
- Never ignore errors
- Wrap errors with context
- Use descriptive error messages
- Return early on errors

**Example:**

```go
// Good
func StartService(name string) error {
    config, err := LoadServiceConfig(name)
    if err != nil {
        return fmt.Errorf("failed to load service config: %w", err)
    }

    if err := validateConfig(config); err != nil {
        return fmt.Errorf("invalid config for %s: %w", name, err)
    }

    // ... rest of function
}

// Bad
func StartService(name string) error {
    config, _ := LoadServiceConfig(name) // NEVER ignore errors!

    if err := validateConfig(config); err != nil {
        return err // No context!
    }

    // ... rest of function
}
```

---

## Logging Standards

### Command Logger

**Required:** Each command must create a logger:

```go
logger := lib.NewCommandLogger("command-name")
```

**Logger Methods:**
- `logger.Info(message, details)` - Informational messages
- `logger.Success(message, details)` - Success messages
- `logger.Warn(message, details)` - Warning messages
- `logger.Error(message, details)` - Error messages (and returns error)
- `logger.PrintSection(title)` - Section headers
- `logger.PrintSummary(items)` - Summary tables/lists

**Example:**

```go
func RunLink(args []string) error {
    logger := lib.NewCommandLogger("link")

    logger.Info("Linking project", "Reading project configuration...")

    project, err := loadProjectConfig()
    if err != nil {
        return logger.Error("Failed to load project", err.Error())
    }

    logger.Success("Project linked", "Now available at http://"+project.Domain)
    return nil
}
```

### Prohibited Patterns

**NEVER use raw `fmt.Printf` for user-facing output:**

```go
// Bad - Don't do this!
func RunStart(args []string) error {
    fmt.Printf("Starting services...\n") // ❌
    // ...
}

// Good - Use logger instead
func RunStart(args []string) error {
    logger := lib.NewCommandLogger("start")
    logger.Info("Starting services", "Initializing...")
    // ...
}
```

### Log Files

**Command Logs:**
- Location: `~/.chauffeur/logs/commands/<command>/`
- Format: `<timestamp>.log`
- Includes: Full command output with timestamps

**Service Logs:**
- Nginx: `~/.chauffeur/nginx/logs/access.log`, `error.log`
- PHP-FPM: `~/.chauffeur/php/<version>/var/log/php-fpm.log`
- Projects: `~/.chauffeur/projects/<slug>/logs/`

---

## Testing Conventions

### Test Organization

**Unit Tests:**
- Place alongside source: `filename_test.go`
- Test package: `package commands` (not `package commands_test`)
- Use table-driven tests for multiple scenarios

**Integration Tests:**
- Place in `tests/` directory
- Test multiple components together
- Use `t.TempDir()` for workspace isolation

**Example:**

```go
// commands/link_test.go
func TestRunLink(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        setup      func(*testing.T) string
        wantErr    bool
        errMessage string
    }{
        {
            name: "successful link",
            args: []string{"--site", "test.test"},
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                // Setup test project
                return dir
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            projectDir := tt.setup(t)
            err := RunLink(tt.args)
            // ... assertions
        })
    }
}
```

### Test Isolation

**Workspace Isolation:**
```go
func TestSomething(t *testing.T) {
    // Create temporary workspace
    tempDir := t.TempDir()

    // Set HOME to temp directory
    t.Setenv("HOME", tempDir)

    // Now test can't touch real user state
    // ...
}
```

### Test Coverage

**Minimum Coverage:** 80% for new code

**Check Coverage:**
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Documentation Conventions

### Synchronization Requirement

**When code changes, ALL relevant documentation must be updated immediately:**

- [ ] [AGENTS.md](../AGENTS.md) - Command contracts, architectural changes
- [ ] [README.md](../README.md) - Feature status, examples, installation
- [ ] [docs/TODO_STATUS.md](TODO_STATUS.md) - Move items, mark progress
- [ ] [sites/constants.ts](../sites/constants.ts) - **CRITICAL** - CLI command definitions
- [ ] [docs/History.md](History.md) - Log all changes (MANDATORY)

### Documentation Style

**Principles:**
- Clear and concise
- Examples first
- Show, don't tell
- Keep it accurate

**Format:**
- Use Markdown for all docs
- Code blocks with syntax highlighting
- Tables for structured data
- Links to related documentation

### Code Comments

**Exported Functions:**
```go
// LoadProjectConfig reads and parses a project configuration file from
// the given path. It validates the structure and returns an error if
// the configuration is invalid or missing required fields.
//
// The configuration file must be in YAML format with the following
// required fields: version, path, php, site.domain.
//
// Returns an error if the file doesn't exist, cannot be parsed,
// or fails validation.
func LoadProjectConfig(path string) (*ProjectConfig, error) {
    // ...
}
```

---

## AI Agent Conventions

### Required Reading Order

**ALL AI agents MUST read documents in this order:**

1. **[AGENTS.md](../AGENTS.md)** - Single source of truth, all policies
2. **[docs/Conventions.md](Conventions.md)** - This file, conventions and history logging
3. **[docs/Architecture.md](Architecture.md)** - System architecture and design
4. **[docs/TechStack.md](TechStack.md)** - Technology stack and dependencies
5. **[docs/Plan.md](Plan.md)** - Roadmap and priorities
6. **[docs/History.md](History.md)** - Past changes and context

### Development Workspace

**CRITICAL:** Always build and test in isolated temporary workspace:

```bash
# Create isolated development environment
export CHAUFFEUR_TEMP_WS=$(mktemp -d)
export HOME=$CHAUFFEUR_TEMP_WS

# Build development binary in temp workspace
go build -o $CHAUFFEUR_TEMP_WS/chauf cli/main.go

# Test your changes safely
$CHAUFFEUR_TEMP_WS/chauf init
$CHAUFFEUR_TEMP_WS/chauf link --category "Test Category"
$CHAUFFEUR_TEMP_WS/chauf category list

# Cleanup when done
rm -rf $CHAUFFEUR_TEMP_WS
unset CHAUFFEUR_TEMP_WS
```

**Why This Matters:**
- Prevents accidental changes to production workspace
- Clean testing environment for each session
- Reproducible bugs and testing
- Isolated development won't interfere with installed Chauffeur

### AI Agent Workflow

**1. Understand the Task:**
- Read relevant documentation
- Clarify requirements with user
- Plan the approach

**2. Create Task List:**
```go
// Use TodoWrite tool for multi-step tasks
TodoWrite{
    todos: [
        {content: "Read existing code", status: "in_progress"},
        {content: "Implement feature", status: "pending"},
        {content: "Write tests", status: "pending"},
        {content: "Update documentation", status: "pending"},
        {content: "Log in History.md", status: "pending"},
    ]
}
```

**3. Work in Isolated Workspace:**
- Set up temporary workspace
- Build development binary
- Test changes in isolation

**4. Update Documentation:**
- Update ALL relevant docs
- Update sites/constants.ts
- Log in History.md

**5. Present for Review:**
- Show changes made
- Explain what was done
- Ask for commit approval

### AI Agent Restrictions

**STRICTLY FORBIDDEN:**
- Making git commits without explicit instruction
- Pushing changes without permission
- Modifying system files outside `~/.chauffeur/`
- Installing packages to system directories
- Skipping history logging

**REQUIRED:**
- Update History.md for all state changes
- Work in isolated temporary workspace
- Present changes for user review
- Follow all conventions in this document

---

## Development Workflow

### Before Writing Code

1. **Read Documentation:**
   - Read [AGENTS.md](../AGENTS.md) for architectural rules
   - Read [Architecture.md](Architecture.md) for design patterns
   - Read [Conventions.md](Conventions.md) for coding standards

2. **Understand Context:**
   - Read existing code for similar functionality
   - Check [TODO_STATUS.md](TODO_STATUS.md) for related items
   - Review [History.md](History.md) for recent changes

3. **Plan Approach:**
   - Create task list for complex features
   - Identify files that need changes
   - Plan testing approach

### During Development

1. **Use Isolated Workspace:**
   ```bash
   export CHAUFFEUR_TEMP_WS=$(mktemp -d)
   export HOME=$CHAUFFEUR_TEMP_WS
   go build -o $CHAUFFEUR_TEMP_WS/chauf cli/main.go
   ```

2. **Follow Conventions:**
   - Use structured logging
   - Write tests as you go
   - Keep functions focused
   - Document non-obvious code

3. **Test Frequently:**
   ```bash
   go test ./...
   $CHAUFFEUR_TEMP_WS/chauf <command>
   ```

4. **Update Documentation:**
   - Update docs as you make changes
   - Don't batch documentation updates
   - Keep documentation in sync with code

### After Development

1. **Run All Tests:**
   ```bash
   go test ./...
   go vet ./...
   go fmt ./...
   ```

2. **Verify Documentation:**
   - [ ] AGENTS.md updated
   - [ ] README.md updated
   - [ ] TODO_STATUS.md updated
   - [ ] sites/constants.ts updated
   - [ ] History.md updated

3. **Present Changes:**
   - Show what was changed
   - Explain why changes were made
   - Ask for commit approval

---

## Code Review Checklist

### For All Changes

- [ ] Code follows Go conventions
- [ ] Structured logging used (no `fmt.Printf`)
- [ ] Errors are handled properly
- [ ] Tests added/updated
- [ ] Test coverage ≥80%
- [ ] Comments for non-obvious code
- [ ] No hardcoded paths (use workspace helpers)

### For Features

- [ ] Feature is documented in README
- [ ] AGENTS.md updated with new contracts
- [ ] sites/constants.ts updated
- [ ] Examples tested and working
- [ ] Backward compatibility maintained
- [ ] Breaking changes documented

### For Bug Fixes

- [ ] Root cause identified
- [ ] Fix tested with regression test
- [ ] Related issues referenced
- [ ] History.md updated with fix details

### For Documentation

- [ ] All examples tested
- [ ] Command references match binary
- [ ] Links are valid
- [ ] Feature status accurate

### For AI Agents

- [ ] Work done in isolated workspace
- [ ] History.md updated with all changes
- [ ] No unauthorized commits made
- [ ] Changes presented for review
- [ ] All conventions followed

---

## Enforcing Conventions

### Automated Checks

**Pre-commit Hooks (Recommended):**
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Format code
go fmt ./...

# Run tests
go test ./...

# Check for History.md updates
# (Implementation depends on your workflow)

# Check for fmt.Printf usage
if grep -r "fmt.Printf" cli/commands/*.go; then
    echo "ERROR: Use logger instead of fmt.Printf in commands"
    exit 1
fi
```

### Manual Reviews

**Before Merging:**
1. Review all documentation changes
2. Verify History.md was updated
3. Test all examples in README
4. Run full test suite
5. Check AGENTS.md for compliance

### Consequences of Non-Compliance

**For Humans:**
- PRs rejected without proper documentation
- Requests for changes until conventions followed
- Loss of merge privileges for repeat offenses

**For AI Agents:**
- Tasks marked incomplete if History not updated
- Corrections requested for convention violations
- Reprimand for unauthorized commits

---

## Resources

### Related Documentation

- [AGENTS.md](../AGENTS.md) - Authoritative development handbook
- [Architecture.md](Architecture.md) - System architecture
- [TechStack.md](TechStack.md) - Technology stack
- [Plan.md](Plan.md) - Project roadmap
- [History.md](History.md) - Change log
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guide

### External Resources

- [Effective Go](https://go.dev/doc/effective_go) - Go style guide
- [Conventional Commits](https://www.conventionalcommits.org/) - Commit message format
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - Go code review

---

**Remember: The most critical convention is to log EVERY significant execution in [History.md](History.md). This ensures traceability and helps both humans and AI agents understand what changed and why.**
