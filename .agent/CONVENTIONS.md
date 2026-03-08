# Chauffeur V2 — Conventions

This directory contains project conventions and guidelines for the Chauffeur V2 CLI tool.

## Conventions Overview

| Document | Description |
|----------|-------------|
| [conventions/tech-stack.md](./conventions/tech-stack.md) | Technology choices and versions |
| [conventions/project-structure.md](./conventions/project-structure.md) | Directory structure and organization |
| [conventions/code-style.md](./conventions/code-style.md) | Go coding standards and best practices |
| [conventions/naming-conventions.md](./conventions/naming-conventions.md) | Naming rules for all identifiers |
| [conventions/output-style.md](./conventions/output-style.md) | CLI output format, colors, symbols, command patterns |

## Integrations

| Document | Description |
|----------|-------------|
| [integrations/nginx.md](./integrations/nginx.md) | Nginx build and config patterns |
| [integrations/php.md](./integrations/php.md) | PHP-FPM management |
| [integrations/dnsmasq.md](./integrations/dnsmasq.md) | DNS for .test domains |
| [integrations/mkcert.md](./integrations/mkcert.md) | SSL certificate generation |
| [integrations/composer.md](./integrations/composer.md) | Composer PHAR + shim |
| [integrations/systemd.md](./integrations/systemd.md) | Systemd user services |

## Specifications

See [SPECS.md](./SPECS.md) for feature specifications.

## Quick Reference

### Code Style

- Go 1.22+, zero external dependencies
- All output through `lib.Logger` — never raw `fmt.Printf`
- Commands follow `RunXXX(args []string) error` pattern
- All operations are idempotent
- Table-driven tests with `t.TempDir()` isolation

### Tech Stack

- Go 1.22+ (standard library only)
- YAML for configuration files
- Go `text/template` for config generation
- Shell scripts for bootstrap only

### File Naming

- Commands: `commands/<verb>.go` (e.g., `commands/link.go`)
- Internal packages: `internal/<domain>/` (e.g., `internal/projects/`)
- Installers: `installers/<service>.go` (e.g., `installers/php.go`)
- Library: `lib/<concern>.go` (e.g., `lib/logging.go`)

### Project Structure

```
chauffeur-v2/
├── cmd/
│   └── chauf/
│       └── main.go         # Entry point
├── internal/
│   ├── commands/           # CLI command implementations
│   ├── config/             # Config loading and validation
│   ├── installers/         # Service installation (php, nginx, composer)
│   ├── projects/           # Project registration and management
│   ├── services/           # Service lifecycle (start/stop/restart)
│   ├── system/             # Host integration (ports, DNS, systemd)
│   ├── templates/          # Config file templates
│   ├── workspace/          # Path resolution and workspace state
│   └── lib/                # Shared utilities (logging, downloads, ssl)
├── tests/                  # Integration tests
├── scripts/                # Build and dev scripts
├── .agent/                 # AI documentation
└── go.mod
```

### Output Style

- **No repeating prefix** — context set once in the header, clean indented body
- **Simplified** (default): outcome-focused, combines passing checks, minimal lines
- **Verbose** (`--verbose`): every phase labeled, full paths, timing shown
- Errors to stderr; colors auto-disabled when not a TTY or `--no-color` set
- ✓/✗/⚠/●/○ symbols — consistent across all commands
- Tables for list commands; label-value blocks for detail/verbose output

### Key Rules for AI Agents

1. **Never** use raw `fmt.Printf` for user output — always use `lib.Logger`
2. **Never** auto-register projects — only via explicit `chauf link`
3. **Always** make operations idempotent — check before creating/modifying
4. **Always** validate input at command boundaries — domains, paths, ports, PHP versions
5. **Always** use `t.TempDir()` in tests — never touch real `~/.chauffeur`
6. **Always** log significant changes to `CHANGELOG.md`
7. **Never** execute system-level mutations silently — print commands for user approval

---

See individual files for detailed guidelines.
