# Chauffeur V2 — Feature Specifications

This document indexes all feature specifications. Each spec file contains detailed behavior, data structures, and implementation notes.

## Specifications Overview

| Document | Description |
|----------|-------------|
| [specs/workspace.md](./specs/workspace.md) | Workspace layout, initialization, and paths |
| [specs/project-linking.md](./specs/project-linking.md) | Project registration, linking, and management |
| [specs/php-fpm.md](./specs/php-fpm.md) | PHP-FPM shared and dedicated strategies |
| [specs/ssl.md](./specs/ssl.md) | SSL certificate generation and management |
| [specs/multi-domain.md](./specs/multi-domain.md) | Multiple domains and alias support |
| [specs/service-orchestration.md](./specs/service-orchestration.md) | Service lifecycle management |
| [specs/podman.md](./specs/podman.md) | Podman database containers and backup/restore |
| [specs/cli-commands.md](./specs/cli-commands.md) | Complete CLI command reference |
| [specs/future-plans.md](./specs/future-plans.md) | Planned V2 features |

## Quick Reference

### Workspace Layout

```
~/.chauffeur/
├── bin/
│   ├── chauf               # CLI binary
│   └── shims/
│       ├── php             # PHP version-aware shim
│       └── composer        # Composer shim using Chauffeur PHP
├── config/
│   └── chauffeur.yaml      # Global workspace config
├── projects/
│   └── <slug>/
│       ├── config.yaml     # Project config
│       └── php-fpm/        # Dedicated FPM socket/PID (if dedicated)
├── php/
│   └── <version>/          # Compiled PHP runtime
├── nginx/
│   ├── bin/nginx
│   ├── etc/
│   │   ├── nginx.conf
│   │   ├── sites-available/
│   │   └── sites-enabled/
│   ├── certs/              # SSL certificates
│   └── logs/
├── composer/
│   └── composer.phar
├── podman/
│   ├── <container-name>.yaml  # Container config (e.g., chauf-mysql57.yaml)
│   ├── volumes/
│   │   └── <container-name>/  # Bind mount volumes for data persistence
│   └── backups/              # Database backup files (*.tar.gz)
├── cache/                  # PHP tarballs, Composer PHAR
├── logs/                   # Command and service logs
└── system/                 # Port forwarding state, DNS metadata
```

### Project Config Schema

```yaml
slug: my-project
path: /home/user/Projects/my-project
domain: my-project.test
aliases:
  - admin.my-project.test
php_version: "8.3"
ssl: false
fpm:
  dedicated: false
  socket: ""   # auto-populated if dedicated
created_at: "2025-01-01T00:00:00Z"
updated_at: "2025-01-01T00:00:00Z"
```

### Global Config Schema

```yaml
workspace: ~/.chauffeur
nginx:
  http_port: 8080
  https_port: 8443
php:
  default_version: "8.3"
dns:
  tld: test
  enabled: true
logging:
  level: info
  max_size_mb: 10
```

### CLI Command Overview

| Command | Description |
|---------|-------------|
| `chauf init` | Initialize workspace |
| `chauf info` | Show workspace info |
| `chauf link` | Register current directory as project |
| `chauf links` | List all projects |
| `chauf unlink` | Unregister project |
| `chauf install <service>` | Install nginx, PHP, Composer |
| `chauf remove <service>` | Remove a service |
| `chauf php use/isolate <version>` | Manage PHP versions |
| `chauf start` | Start services |
| `chauf stop` | Stop services |
| `chauf restart` | Restart services |
| `chauf status` | Show service status |
| `chauf logs` | View logs |
| `chauf secure` | Enable SSL for project |
| `chauf unsecure` | Disable SSL for project |
| `chauf doctor` | Health check |
| `chauf config` | Manage workspace/project config (V2) |
| `chauf env` | Manage project environment variables (V2) |
| `chauf autostart` | Manage systemd auto-start (V2) |
| `chauf clean` | Clean workspace artifacts |
| `chauf migrate` | Migrate project to new workspace |
| `chauf self-update` | Update CLI binary |
| `chauf uninstall` | Remove workspace |

### Podman Commands

| Command | Description |
|---------|-------------|
| `chauf podman create` | Create a database container |
| `chauf podman start` | Start a container |
| `chauf podman stop` | Stop a container |
| `chauf podman remove` | Remove a container |
| `chauf podman list` | List all containers |
| `chauf podman status` | Show container status |
| `chauf podman console` | Attach to container's database CLI |
| `chauf podman backup` | Interactive database backup |
| `chauf podman restore` | Interactive database restore |

### PHP Version Support

| Version | Status | Notes |
|---------|--------|-------|
| PHP 8.4 | ✅ Full | Latest stable |
| PHP 8.3 | ✅ Full | Default version |
| PHP 8.2 | ✅ Full | LTS |
| PHP 8.1 | ✅ Full | |
| PHP 8.0 | ⚠️ Legacy | GD needs patching |
| PHP 7.4 | ⚠️ Legacy | EOL, best-effort |
