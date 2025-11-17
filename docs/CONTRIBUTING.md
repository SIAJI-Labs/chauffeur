# Contributing to Chauffeur

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification for our commit messages. This leads to more readable messages that are easy to follow when looking through the project history.

### Commit Message Format
Each commit message consists of a **header**, a **body** and a **footer**. The header has a special format that includes a **type**, a **scope** and a **subject**:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

### Types
Must be one of the following:

* **feat**: A new feature
* **fix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **perf**: A code change that improves performance
* **test**: Adding missing tests or correcting existing tests
* **chore**: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Scope
The scope should be the name of the module affected (as perceived by the person reading the changelog generated from commit messages).

### Subject
The subject contains a succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize the first letter
* no dot (.) at the end

### Examples

```
feat(cli): add service status command

Add new command to check the status of installed services.
The command provides detailed information about service state and uptime.

Closes #123
```

```
fix(installer): resolve nginx config template issue

Fixed incorrect template variable substitution that caused nginx 
configuration generation to fail on certain systems.

Fixes #456
```

```
docs(readme): update installation instructions

Update the README with more detailed installation steps and
troubleshooting information.
```

## Getting Started as a Contributor

### Prerequisites
- **Go 1.22+**: Primary language for the CLI
- **Linux development environment**: Chauffeur is Linux-focused
- **Basic shell knowledge**: For testing and debugging
- **Git**: For version control

### Quick Start
1. **Fork and clone** the repository
2. **Build locally**: `chauf self-update --dev` (from repo root)
3. **Run tests**: `go test ./...`
4. **Read AGENTS.md**: The authoritative handbook for development patterns

### Go Basics for Chauffeur Development

#### Project Structure
```
cli/
├── commands/          # CLI command implementations
├── lib/              # Shared utilities and helpers
├── internal/         # Internal packages (services, projects, etc.)
└── installers/       # Service installation logic

tests/                # Integration and unit tests
docs/                 # Documentation
templates/            # Config templates (nginx, PHP-FPM, etc.)
```

#### Key Go Concepts Used
- **Packages**: Each directory is a package (`package commands`, `package lib`)
- **Interfaces**: Service abstractions in `internal/services/`
- **Error handling**: Explicit error returns, no exceptions
- **Structs**: Service definitions and configuration
- **Slices**: Dynamic arrays for service lists and configurations

#### Common Patterns
```go
// Logger initialization (required for commands)
logger := lib.NewCommandLogger("command-name")

// Error handling pattern
if err != nil {
    return logger.Error("operation description", err.Error())
}

// Success logging
logger.Success("operation completed", "additional info")

// Reading service configuration
manager, err := services.NewServiceManager()
if err != nil {
    return logger.Error("create service manager", err.Error())
}
```

### AI-Assisted Development Workflow

This project uses AI assistance extensively. Here's how to work effectively with AI tools:

#### Before Coding
1. **Read AGENTS.md first**: Contains architectural rules and contracts
2. **Check TODO_STATUS.md**: Understand current project state
3. **Review existing patterns**: Look at similar commands for structure

#### During Development
1. **Use TodoWrite tool**: Track multi-step tasks systematically
2. **Follow logging contract**: Use `lib.NewCommandLogger()`, never raw `fmt.Printf`
3. **Test frequently**: Run `go test ./...` before pushing changes
4. **Update docs synchronously**: Keep README.md, TODO_STATUS.md, and AGENTS.md aligned

#### Code Review Checklist
- [ ] Does code follow AGENTS.md workspace rules?
- [ ] Is logging structured and consistent?
- [ ] Are error messages actionable?
- [ ] Do docs match implementation?
- [ ] Tests pass: `go test ./...`

### Development Workflow

#### 1. Setup Development Environment
```bash
# Clone your fork
git clone https://github.com/yourusername/chauffeur.git
cd chauffeur

# Install dependencies (Go modules)
go mod tidy

# Build CLI locally
chauf self-update --dev
```

#### 2. Make Changes
```bash
# Create a feature branch
git checkout -b feature-name

# Work on your changes
# Use chauf self-update --dev to test changes
```

#### 3. Test Your Changes
```bash
# Run all tests
go test ./...

# Test specific functionality
~/.chauffeur/bin/chauf status --detail
```

#### 4. Keep Documentation in Sync
After any code change, update:
- **README.md**: Feature status, examples, installation changes
- **docs/TODO_STATUS.md**: Move items between sections, mark progress
- **AGENTS.md**: Update command contracts or architectural changes

#### 5. Submit Changes
```bash
# Commit with conventional format
git commit -m "feat(commands): add new functionality"

# Push to your fork
git push origin feature-name

# Create pull request
```

### Testing Guidelines

#### Unit Tests
- Place tests alongside source files: `filename_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (filesystem, network)

#### Integration Tests
- Use `t.TempDir()` to avoid touching real user state
- Set `t.Setenv("HOME", tempDir)` for workspace isolation
- Test CLI commands by capturing output

#### Key Testing Areas
- Service installation and management
- Template generation
- Configuration validation
- CLI command parsing and execution

### Common Gotchas

#### Workspace Management
- Never touch `/usr`, `/etc`, or `/opt` directly
- Always work under `~/.chauffeur/`
- Use workspace helpers from `cli/internal/workspace`

#### Logging
- Always create a logger: `logger := lib.NewCommandLogger("command")`
- Use structured logging: `logger.Info()`, `logger.Success()`, `logger.Error()`
- Never use `fmt.Printf` for user-facing output

#### Service Management
- Use service manager from `cli/internal/services`
- Handle both global and project-specific services
- Check service status before operations

### Getting Help

1. **Issues**: Report bugs or request features on GitHub
2. **Documentation**: Check AGENTS.md for architectural guidance
3. **Existing code**: Look at similar commands for patterns
4. **Tests**: Review test files for usage examples

### Code Style

- Follow Go conventions (gofmt)
- Use clear, descriptive variable names
- Add comments for non-obvious logic
- Keep functions focused and small
- Prefer existing helpers over reinventing

Remember: **Read AGENTS.md before writing any code!** It contains the authoritative rules for Chauffeur development.