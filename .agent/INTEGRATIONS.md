# Chauffeur V2 — Integrations

This document indexes all service integration guides.

## Integrations Overview

| Document | Description |
|----------|-------------|
| [integrations/nginx.md](./integrations/nginx.md) | Nginx build from source, config templates, management |
| [integrations/php.md](./integrations/php.md) | PHP compilation, version management, FPM config |
| [integrations/dnsmasq.md](./integrations/dnsmasq.md) | DNS routing for `.test` domains |
| [integrations/mkcert.md](./integrations/mkcert.md) | Local trusted SSL certificates |
| [integrations/composer.md](./integrations/composer.md) | Composer PHAR download and shim setup |
| [integrations/systemd.md](./integrations/systemd.md) | Systemd user services for auto-start |
| [specs/podman.md](./specs/podman.md) | Podman database containers and backup/restore |

## Service Architecture Overview

```
User Request (browser)
    ↓ port 80/443 (via iptables redirect, optional)
    ↓ port 8080/8443 (default)
nginx (~/.chauffeur/nginx/bin/nginx)
    ↓ FastCGI via Unix socket
PHP-FPM (~/.chauffeur/php/<version>/bin/php-fpm)
    ↓
Project Files (/home/user/Projects/<project>/public/)
```

## DNS Architecture

```
Browser → <project>.test
    ↓ dnsmasq (system)
    ↓ resolves *.test → 127.0.0.1
    ↓ nginx listens on 127.0.0.1:8080
    ↓ routes by Host header to project
```

## PHP Version Routing

```
Project directory lookup
    ↓ Check ~/.chauffeur/projects/<slug>/config.yaml
    ↓ Get php_version field
    ↓ Use ~/.chauffeur/php/<version>/bin/php
```

## Service Interactions

| Service | Interacts With | Via |
|---------|---------------|-----|
| nginx | PHP-FPM | Unix socket (`/tmp/chauffeur-<version>.sock` or `/tmp/chauffeur-<slug>.sock`) |
| nginx | mkcert | Certificate files in `~/.chauffeur/nginx/certs/` |
| nginx | dnsmasq | dnsmasq routes domain → nginx listens |
| PHP-FPM | Composer | Composer shim uses PHP-FPM's PHP binary |
| CLI | All services | `os/exec` for process management |

## Quick Reference: Service Paths

| Service | Binary | Config | Logs | Sockets/PIDs |
|---------|--------|--------|------|-------------|
| nginx | `~/.chauffeur/nginx/bin/nginx` | `~/.chauffeur/nginx/etc/nginx.conf` | `~/.chauffeur/nginx/logs/` | `~/.chauffeur/nginx/logs/nginx.pid` |
| PHP-FPM (shared) | `~/.chauffeur/php/<ver>/bin/php-fpm` | `~/.chauffeur/php/<ver>/etc/php-fpm.conf` | `~/.chauffeur/php/<ver>/var/log/` | `/tmp/chauffeur-<ver>.sock` |
| PHP-FPM (dedicated) | `~/.chauffeur/php/<ver>/bin/php-fpm` | `~/.chauffeur/projects/<slug>/php-fpm.conf` | `~/.chauffeur/projects/<slug>/logs/` | `~/.chauffeur/projects/<slug>/php-fpm.sock` |
| Composer | `~/.chauffeur/composer/composer.phar` | N/A | N/A | N/A |
