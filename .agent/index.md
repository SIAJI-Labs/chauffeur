# Chauffeur V2 Documentation Index

Welcome to the Chauffeur V2 project documentation. This is the authoritative handbook for all contributors and AI agents working on this project.

## Documentation Structure

```
.agent/
├── index.md                    # This file
├── PRD.md                      # Product Requirements
├── SPECS.md                    # Feature Specifications Index
├── CONVENTIONS.md              # Code Conventions Index
├── INTEGRATIONS.md             # Tech Integrations Index
└── IMPLEMENTATION_PLAN.md      # Implementation Roadmap
```

## Guides

| Document | Description |
|----------|-------------|
| [PRD.md](./PRD.md) | Product requirements and project overview |
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | Phase-by-phase implementation roadmap |
| [SPECS.md](./SPECS.md) | Feature specifications index |
| [CONVENTIONS.md](./CONVENTIONS.md) | Code style, naming, project structure |
| [INTEGRATIONS.md](./INTEGRATIONS.md) | Service and library integration guides |

## Conventions

| Document | Description |
|----------|-------------|
| [conventions/tech-stack.md](./conventions/tech-stack.md) | Technology choices and versions |
| [conventions/project-structure.md](./conventions/project-structure.md) | Directory organization |
| [conventions/code-style.md](./conventions/code-style.md) | Go coding standards and best practices |
| [conventions/naming-conventions.md](./conventions/naming-conventions.md) | Naming rules for files, variables, commands |
| [conventions/output-style.md](./conventions/output-style.md) | CLI output format, colors, symbols, and command patterns |

## Integrations

| Document | Description |
|----------|-------------|
| [integrations/nginx.md](./integrations/nginx.md) | Nginx build, config, and management |
| [integrations/php.md](./integrations/php.md) | PHP-FPM compilation and version management |
| [integrations/dnsmasq.md](./integrations/dnsmasq.md) | DNS resolution for .test domains |
| [integrations/mkcert.md](./integrations/mkcert.md) | Local SSL certificate management |
| [integrations/composer.md](./integrations/composer.md) | Composer installation and shim |
| [integrations/systemd.md](./integrations/systemd.md) | Systemd user service integration |

## Feature Specifications

| Document | Description |
|----------|-------------|
| [specs/workspace.md](./specs/workspace.md) | Workspace architecture and layout |
| [specs/project-linking.md](./specs/project-linking.md) | Project registration and management |
| [specs/php-fpm.md](./specs/php-fpm.md) | PHP-FPM strategies (shared vs dedicated) |
| [specs/ssl.md](./specs/ssl.md) | SSL/TLS certificate management |
| [specs/multi-domain.md](./specs/multi-domain.md) | Multi-domain and alias support |
| [specs/service-orchestration.md](./specs/service-orchestration.md) | Start, stop, restart, status |
| [specs/cli-commands.md](./specs/cli-commands.md) | Complete CLI command reference |
| [specs/future-plans.md](./specs/future-plans.md) | Planned features and roadmap |

## Getting Started

1. Read [PRD.md](./PRD.md) for project vision and goals
2. Review [SPECS.md](./SPECS.md) for feature details
3. Check [CONVENTIONS.md](./CONVENTIONS.md) for coding standards
4. Set up understanding with [INTEGRATIONS.md](./INTEGRATIONS.md)
5. Follow [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) for task priorities

## Key Principles for AI Agents

- **Never auto-scan projects** — registration is always explicit via `chauf link`
- **Workspace-first** — everything lives under `~/.chauffeur/`, zero host mutation
- **Idempotent operations** — all commands are safe to re-run
- **Structured logging** — all output goes through `lib.Logger`, never raw `fmt.Printf`
- **History logging** — every significant change must be logged in `CHANGELOG.md`
