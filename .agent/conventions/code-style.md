# Code Style

## Go Conventions

### Package Organization

Each package has a single, clear responsibility. Avoid "util" packages that accumulate unrelated code. If a utility is used by only one package, put it there.

### File Naming

- All lowercase, underscore-separated: `php_installer.go`, `port_validation.go`
- Test files: `<filename>_test.go`
- Command files named after the verb: `link.go`, `start.go`, `doctor.go`

### Function Signatures

**Command handlers** — always this signature:
```go
func RunLink(args []string) error
func RunStart(args []string) error
func RunDoctor(args []string) error
```

**Internal functions** — return error as last value:
```go
func (p *ProjectManager) Create(config ProjectConfig) error
func (i *PHPInstaller) Install(version string) error
func downloadFile(url, dest string) error
```

### Error Handling

Return errors up the call stack. Wrap with context at package boundaries:

```go
// CORRECT: wrap error at boundary
func (i *PHPInstaller) Install(version string) error {
    if err := downloadSource(version); err != nil {
        return fmt.Errorf("php install: download failed: %w", err)
    }
    return nil
}

// WRONG: ignore error
downloadSource(version)

// WRONG: log and return (double logging)
if err := downloadSource(version); err != nil {
    log.Printf("error: %v", err)
    return err
}
```

Command handlers log the final error, not internal functions:

```go
// commands/link.go
func RunLink(args []string) error {
    logger := lib.NewCommandLogger("link")
    if err := projects.Create(config); err != nil {
        logger.Error("Failed to link project", err)
        return err
    }
    logger.Success("Project linked: %s", config.Domain)
    return nil
}
```

### Logging

**NEVER** use `fmt.Printf`, `fmt.Println`, `log.Printf` for user-facing output. Always use `lib.Logger`.

```go
// CORRECT
logger := lib.NewCommandLogger("link")
logger.Info("Generating nginx config...")
logger.Success("Project linked successfully")
logger.Warning("Port 8080 already in use, using 8081")
logger.Error("Failed to link project", err)

// WRONG
fmt.Printf("Generating nginx config...\n")
log.Printf("error: %v", err)
```

**Usage/help text** is the only exception — `fmt.Println` is acceptable for static help output.

### Configuration

Always use the `internal/workspace` package for paths — never hardcode:

```go
// CORRECT
ws := workspace.New()
phpDir := ws.PHPDir("8.3")

// WRONG
phpDir := filepath.Join(os.Getenv("HOME"), ".chauffeur", "php", "8.3")
```

### Input Validation

Validate at command boundaries. Use `lib/input.go` helpers:

```go
// Validate PHP version
if !lib.IsValidPHPVersion(phpVersion) {
    return fmt.Errorf("invalid PHP version: %s (supported: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4)", phpVersion)
}

// Validate domain
if !lib.IsValidDomain(domain) {
    return fmt.Errorf("invalid domain: %s (must be alphanumeric with hyphens)", domain)
}

// Validate port
if !lib.IsValidPort(port) {
    return fmt.Errorf("invalid port: %d (must be 1-65535)", port)
}
```

### Process Execution

Never use `os/exec` directly in command files. Wrap in internal packages:

```go
// CORRECT — in internal/services/nginx.go
func (n *NginxService) Start() error {
    cmd := exec.Command(n.binaryPath, "-c", n.configPath)
    return cmd.Run()
}

// WRONG — in commands/start.go
cmd := exec.Command(os.Getenv("HOME") + "/.chauffeur/nginx/bin/nginx", ...)
```

### Idempotency

Every operation must be safe to re-run:

```go
// CORRECT: check before creating
func (w *Workspace) Init() error {
    if _, err := os.Stat(w.Root); err == nil {
        // Already exists, update as needed
        return w.ensureDefaults()
    }
    return w.createFresh()
}

// WRONG: always create (will fail on second run)
func (w *Workspace) Init() error {
    return os.MkdirAll(w.Root, 0755)  // fine, but config overwrite is not
}
```

---

## Testing Conventions

### Test File Organization

- Unit tests alongside source: `internal/projects/manager_test.go`
- Integration tests in `tests/integration/`
- Test helpers in `tests/integration/helpers.go`

### Workspace Isolation in Tests

**ALWAYS** isolate tests from the real `~/.chauffeur`:

```go
// CORRECT
func TestLinkProject(t *testing.T) {
    tmpDir := t.TempDir()
    t.Setenv("HOME", tmpDir)
    // Now ~/.chauffeur points to tmpDir/.chauffeur
    // ...
}

// WRONG
func TestLinkProject(t *testing.T) {
    // touches real ~/.chauffeur — never do this
}
```

### Table-Driven Tests

Use table-driven tests for anything with multiple input cases:

```go
func TestParsePHPVersion(t *testing.T) {
    tests := []struct {
        input   string
        want    string
        wantErr bool
    }{
        {"8.3", "8.3", false},
        {"8.3.12", "8.3", false},
        {"latest", "", true},
        {"", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got, err := ParsePHPVersion(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("ParsePHPVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("ParsePHPVersion(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}
```

### Mocking External Processes

Use an `Executor` interface for `os/exec` calls so tests don't spawn real processes:

```go
// internal/lib/executor.go
type Executor interface {
    Run(name string, args ...string) error
    Output(name string, args ...string) ([]byte, error)
}

type RealExecutor struct{}

func (e *RealExecutor) Run(name string, args ...string) error {
    return exec.Command(name, args...).Run()
}

// In tests:
type MockExecutor struct {
    Commands []string
}

func (m *MockExecutor) Run(name string, args ...string) error {
    m.Commands = append(m.Commands, name+" "+strings.Join(args, " "))
    return nil
}
```

### Coverage Target

Aim for ≥ 80% coverage on new packages. Run with:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Import Organization

Organize imports in three groups, separated by blank lines:

```go
import (
    // 1. Standard library
    "fmt"
    "os"
    "path/filepath"

    // 2. External packages (only gopkg.in/yaml.v3 allowed)
    "gopkg.in/yaml.v3"

    // 3. Internal packages
    "github.com/SIAJI-Labs/chauffeur/internal/lib"
    "github.com/SIAJI-Labs/chauffeur/internal/workspace"
)
```

---

## Struct Design

Prefer constructor functions over bare struct literals:

```go
// CORRECT
type ProjectManager struct {
    workspace *workspace.Workspace
    logger    *lib.Logger
}

func NewProjectManager(ws *workspace.Workspace) *ProjectManager {
    return &ProjectManager{
        workspace: ws,
        logger:    lib.NewLogger("projects"),
    }
}

// WRONG — direct struct literal in calling code
mgr := &ProjectManager{workspace: ws, logger: lib.NewLogger("projects")}
```

---

## Changelog Convention

Every significant code change must be recorded in `CHANGELOG.md` using Keep a Changelog format:

```markdown
## [Unreleased]

### Added
- `chauf config show` command for displaying current workspace config

### Fixed
- PHP-FPM process count showing 0 in `chauf status --detail`

### Changed
- `chauf link` now validates domain before nginx config generation
```

AI agents must append to `[Unreleased]` after every feature or fix.
