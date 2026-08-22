# Architecture And Runtime

## Top-level structure

```text
cmd/chauf/main.go          manual CLI dispatcher and global behavior
internal/commands/         command handlers and user workflows
internal/workspace/        workspace root, paths, and global config
internal/projects/         project config, detection, slugs, nginx files
internal/installers/       source downloads and builds for nginx/PHP/Composer
internal/services/         nginx/FPM lifecycle, processes, ports, sockets
internal/system/           systemd and port-forwarding integration
internal/podman/           database container abstraction and persistence
internal/panel/             embedded Go HTTP API and static frontend serving
internal/lib/               output, progress, and spinner helpers
internal/panel-apps/        React/TanStack/Tailwind panel source
```

The implementation plan mentions `internal/config`, `internal/templates`, and `tests`, but those directories are not present in the current repository. Configuration is currently in `internal/workspace`, and nginx rendering is in `internal/projects`.

## Workspace model

The default root is `~/.chauffeur`; `CHAUFFEUR_HOME` overrides it for most workspace operations. The intended layout is:

```text
~/.chauffeur/
├── bin/chauf and bin/shims/php,composer
├── config/chauffeur.yaml
├── projects/<slug>/config.yaml and optional php-fpm state
├── php/<version>/ compiled runtimes
├── nginx/ binary, etc, sites, certs, logs
├── composer/composer.phar
├── podman/ container configs, volumes, backups
├── cache/ source archives and downloads
├── logs/ command/service logs
├── system/ integration metadata
└── panel.pid and panel.log
```

Important current inconsistency: `internal/podman.Root()` derives `~/.chauffeur/podman` directly instead of using the shared workspace root. A custom `CHAUFFEUR_HOME` can therefore split state across two locations.

## Global configuration

The runtime configuration defaults to:

```yaml
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

The loader is a small custom line-oriented YAML parser. Invalid or missing configuration falls back to defaults. This is simple and dependency-light but cannot provide robust YAML semantics, schema validation, or precise parse diagnostics.

Project config stores slug, filesystem path, primary domain, aliases, PHP version, SSL state, FPM strategy/socket, and timestamps. The project config is deliberately outside the project directory.

## Request flow

```text
Browser
  → *.test DNS or hosts resolution
  → localhost:8080 / localhost:8443
  → workspace nginx
  → project server block
  → Unix PHP-FPM socket
  → project document root/application
```

The default service startup order is:

1. Start shared FPM pools.
2. Start dedicated FPM pools.
3. Wait for FPM sockets.
4. Start nginx after configuration testing and port checks.

Shutdown reverses the dependency direction: nginx stops before FPM. Restart uses graceful reloads where possible.

## Project linking flow

`chauf link` resolves the path, detects Laravel/WordPress/PHP/JavaScript project type, asks for an explicit type when no supported marker is found, selects a PHP version for PHP-backed projects or a local proxy port for JavaScript projects, generates a slug and `.test` domain, checks domain conflicts, renders either PHP-FPM or reverse-proxy nginx configuration, enables the site, optionally generates mkcert certificates, optionally writes dedicated FPM state, persists project config, and reloads nginx when active.

## PHP and Composer resolution

PHP versions are installed under `php/<version>`. The PHP shim scans project configuration based on the current working directory and falls back to the global default. `chauf php isolate` changes project config without adding a dotfile to the project. Composer uses the Chauffeur PHP runtime through its shim.

## FPM strategies

- **Shared:** one pool per PHP version, efficient for ordinary projects.
- **Dedicated:** one pool per project, useful for isolation or custom FPM settings.

Both strategies can coexist. Status should make the distinction visible because it affects resource usage, restart scope, and debugging.

## Host integrations

- `dnsmasq` can route the configured TLD to localhost.
- systemd-resolved behavior and NSS ordering are diagnosed, including offline fallback cases.
- iptables commands can forward ports 80/443 to 8080/8443, but host-level mutations are intended to remain explicit.
- user-level systemd units support autostart.
- mkcert provides a local CA and SAN certificates.
- Podman is an optional external service for database containers.
