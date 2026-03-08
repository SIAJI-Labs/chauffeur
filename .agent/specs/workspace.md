# Workspace Specification

## Overview

The Chauffeur workspace is a self-contained directory at `~/.chauffeur/` (or `$CHAUFFEUR_HOME`) that holds all compiled services, project configs, logs, and runtime state. Nothing exists outside this directory (except optional dnsmasq config which the user applies manually).

---

## Workspace Layout

```
~/.chauffeur/
├── bin/
│   ├── chauf                       # CLI binary (self-managed, optional)
│   └── shims/
│       ├── php                     # PHP version-aware shim
│       └── composer                # Composer shim using PHP shim
├── config/
│   └── chauffeur.yaml              # Global workspace configuration
├── projects/
│   └── <slug>/
│       ├── config.yaml             # Project config
│       ├── php-fpm.conf            # FPM config (dedicated strategy only)
│       ├── php-fpm.sock            # FPM socket (dedicated strategy only)
│       ├── php-fpm.pid             # FPM PID (dedicated strategy only)
│       └── logs/
│           ├── php-fpm.log
│           └── php-fpm-slow.log
├── php/
│   └── <version>/                  # Compiled PHP runtime (one per version)
│       ├── bin/php, php-cgi, php-fpm
│       ├── etc/php.ini, php-fpm.conf, conf.d/openssl.ini
│       ├── lib/php/extensions/
│       ├── runtime/php-fpm/        # Shared FPM pool socket/PID
│       └── var/log/
├── nginx/
│   ├── bin/nginx
│   ├── etc/
│   │   ├── nginx.conf
│   │   ├── mime.types
│   │   ├── fastcgi_params
│   │   ├── sites-available/
│   │   │   ├── default.conf
│   │   │   └── <slug>.conf
│   │   └── sites-enabled/
│   │       └── <slug>.conf -> ../sites-available/<slug>.conf
│   ├── certs/
│   │   ├── <domain>.crt
│   │   └── <domain>.key
│   └── logs/
│       ├── access.log
│       ├── error.log
│       └── nginx.pid
├── composer/
│   └── composer.phar
├── cache/
│   ├── php/                        # PHP source tarballs + signatures
│   ├── nginx/                      # Nginx source tarballs
│   └── composer/                   # Composer PHAR cache
├── logs/
│   ├── commands/                   # Per-command execution logs (timestamped)
│   └── chauffeur.log               # Main application log
└── system/
    ├── port-forwarding.json        # iptables rules state
    └── dns.json                    # dnsmasq config state
```

---

## Global Config Schema

File: `~/.chauffeur/config/chauffeur.yaml`

```yaml
workspace: ~/.chauffeur       # Workspace root path
nginx:
  http_port: 8080             # HTTP listen port
  https_port: 8443            # HTTPS listen port
php:
  default_version: "8.3"     # Default PHP version
dns:
  tld: test                   # Local TLD for .test domains
  enabled: true               # Whether dnsmasq is configured
logging:
  level: info                 # debug, info, warn, error
  max_size_mb: 10             # Max log file size before rotation
version: "2"                  # Config schema version
created_at: "2025-01-01T00:00:00Z"
```

---

## Initialization: `chauf init`

Creates the workspace directory structure from scratch. Idempotent — safe to run on existing workspace.

**Steps**:

1. Create `~/.chauffeur/` and all subdirectories
2. Write default `config/chauffeur.yaml` if not present
3. Create nginx directory structure with default configs
4. Write `nginx/etc/fastcgi_params` (standard PHP-FPM params)
5. Write `nginx/etc/mime.types` (standard MIME types)
6. Create default catch-all nginx site config
7. Write PHP and Composer shim scripts to `bin/shims/`
8. Output PATH setup instructions

**Does NOT**:
- Install PHP, nginx, or Composer (use `chauf install` for that)
- Modify shell config files
- Start any services
- Require sudo

---

## `chauf info` Output

```
Chauffeur Workspace

  Root:          ~/.chauffeur
  Version:       2.0.0
  Config:        ~/.chauffeur/config/chauffeur.yaml

Services
  nginx          ● running    (8080/8443)
  php-fpm 8.3    ● running    (shared pool)
  php-fpm 8.1    ○ stopped

Projects (3)
  my-app         my-app.test            PHP 8.3  HTTP
  admin-panel    admin-panel.test       PHP 8.1  HTTPS  (dedicated FPM)
  legacy-site    legacy-site.test       PHP 7.4  HTTP

PHP Versions Installed
  8.4   ~/.chauffeur/php/8.4/
  8.3   ~/.chauffeur/php/8.3/   [default]
  8.1   ~/.chauffeur/php/8.1/
  7.4   ~/.chauffeur/php/7.4/

Cache
  ~/.chauffeur/cache/   (245 MB)
```

---

## Environment Variable Overrides

| Variable | Default | Purpose |
|----------|---------|---------|
| `CHAUFFEUR_HOME` | `~/.chauffeur` | Override workspace root |
| `CHAUFFEUR_PHP_TARBALL` | auto | Offline PHP source path |
| `CHAUFFEUR_PHP_SIGNATURE` | auto | Offline PHP signature path |
| `CHAUFFEUR_PHP_KEYRING` | auto | Offline PHP keyring path |
| `CHAUFFEUR_IMAGICK_TARBALL` | auto | Offline imagick tarball path |
| `CHAUFFEUR_KEEP_BUILD_DIR` | `0` | Set `1` to keep PHP build dirs |
| `CHAUFFEUR_LOG_LEVEL` | `info` | Override log verbosity |

---

## Uninstallation: `chauf uninstall`

Removes the entire workspace. Prompts for confirmation.

```
This will remove:
  ~/.chauffeur/ (including all compiled services, certs, and project configs)

Type 'yes' to confirm:
```

**Does NOT automatically**:
- Remove dnsmasq config (prints commands to remove it)
- Remove iptables rules (prints commands to remove them)
- Remove entries from /etc/hosts
- Remove systemd user services (prints commands to remove them)

---

## `chauf clean`

Removes workspace artifacts without full uninstall.

```bash
chauf clean cache               # Remove download cache
chauf clean logs                # Remove old log files
chauf clean all                 # Remove cache + logs

chauf clean --dry-run           # Show what would be removed
chauf clean --older-than 30d   # Only remove items older than 30 days
```
