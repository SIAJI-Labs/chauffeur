# Chauffeur System Architecture

> **Last Updated**: 2025-01-12
> **Maintainer**: @si-aji
> **Status**: Active - Reflects current implementation

## Table of Contents

1. [Overview](#overview)
2. [Design Philosophy](#design-philosophy)
3. [Workspace Architecture](#workspace-architecture)
4. [Component Architecture](#component-architecture)
5. [PHP-FPM Strategies](#php-fpm-strategies)
6. [Multi-Domain Architecture](#multi-domain-architecture)
7. [Service Orchestration](#service-orchestration)
8. [Data Flow](#data-flow)
9. [Configuration Management](#configuration-management)
10. [Security Model](#security-model)

---

## Overview

Chauffeur is a Linux-first development environment manager that provides per-project PHP services through a unified CLI interface. The architecture prioritizes workspace isolation, minimal host impact, and manual control.

### Core Architectural Goals

1. **Workspace Isolation**: All services, configs, and runtime artifacts live under `~/.chauffeur/`
2. **Zero Host Mutation**: No installation to `/usr`, `/etc`, or `/opt` except opt-in user-level services
3. **Explicit Control**: Manual project registration; no auto-scanning or automatic configuration
4. **Multi-Version Support**: Simultaneous PHP versions with project-specific isolation
5. **Service Flexibility**: Shared and dedicated PHP-FPM strategies for different use cases

---

## Design Philosophy

### 1. Workspace-First Principle

**Definition**: All binaries, configurations, sockets, logs, and runtime state are contained within `~/.chauffeur/`.

**Benefits**:
- Multiple isolated workspaces possible
- Clean uninstallation (remove directory)
- No permission issues (no sudo required for basic operations)
- Reproducible environments across machines

**Implementation**:
```bash
~/.chauffeur/
  ├── bin/           # Shims and helper binaries
  ├── config/        # Global configuration
  ├── projects/      # Per-project state
  ├── php/           # Compiled PHP runtimes
  ├── nginx/         # Nginx installation and configs
  ├── cache/         # Download cache
  └── logs/          # Service and command logs
```

### 2. Minimal Host Impact

**Definition**: Avoid modifying the host system. When necessary, print exact commands for user to run.

**Examples**:
- dnsmasq configuration: Print `sudo tee /etc/dnsmasq.d/chauffeur.conf` commands
- Port forwarding: Track iptables rules in `~/.chauffeur/system/port-forwarding.json`
- systemd services: Use user-level services (`systemctl --user`), no root required

**Fallback Strategy**: When host changes are unavoidable:
1. Print exact commands with explanations
2. Record what changed in workspace metadata
3. Provide cleanup instructions for `chauf remove`/`chauf uninstall`

### 3. Explicit Project Registration

**Definition**: Projects are never auto-scanned. `chauf link` is the only registration mechanism.

**Rationale**:
- Predictable behavior
- No surprises from directory traversal
- Clear project lifecycle management
- User has full control

**Implications**:
- `chauf link` operates on PWD unless `--project` flag provided
- `chauf unlink` requires explicit confirmation
- No background watching or automatic reconfiguration

---

## Workspace Architecture

### Directory Structure

```
~/.chauffeur/
├── bin/                        # User-facing binaries and shims
│   ├── chauf                   # Main CLI binary (optional self-managed)
│   ├── shims/                  # Version-aware shims
│   │   ├── php                 # PHP version detector shim
│   │   ├── composer            # Composer shim using project PHP
│   │   └── php-<version>       # Version-specific PHP shims
│   └── helpers/                # Helper scripts
│
├── config/                     # Global configuration
│   ├── chauffeur.yaml          # Main config file
│   └── ports.yaml              # Port allocation state
│
├── projects/                   # Linked projects
│   └── <slug>/                 # Project directory (slugified name)
│       ├── project.yaml        # Project configuration
│       ├── runtime/            # Runtime-specific files
│       │   └── php-fpm/        # Dedicated FPM socket (if --dedicated-fpm)
│       │       ├── php-fpm.sock
│       │       ├── php-fpm.conf
│       │       └── php-fpm.pid
│       └── logs/               # Project-specific logs
│
├── php/                        # Compiled PHP runtimes
│   └── <version>/              # e.g., 8.3, 8.2, 7.4
│       ├── bin/                # PHP binaries
│       │   ├── php             # Main PHP binary
│       │   ├── php-cgi         # CGI binary
│       │   ├── php-fpm         # FPM binary
│       │   └── php-config      # Config helper
│       ├── etc/                # PHP configuration
│       │   ├── php.ini         # Main PHP config
│       │   ├── php-fpm.conf    # FPM master config
│       │   └── conf.d/         # Additional configs
│       │       ├── openssl.ini # OpenSSL certificate configuration
│       │       └── imagick.ini # Imagick extension config
│       ├── lib/                # PHP extensions
│       ├── include/            # Header files
│       ├── runtime/            # Runtime files (for shared FPM)
│       │   ├── php-fpm/        # Shared FPM socket and PID
│       │   │   ├── php-fpm.sock
│       │   │   ├── php-fpm.pid
│       │   │   └── php-fpm.conf
│       │   └── ...             # Other runtime state
│       └── var/                # Variable data
│
├── nginx/                      # Nginx installation
│   ├── bin/                    # Nginx binary
│   ├── etc/                    # Nginx configuration
│   │   ├── nginx.conf          # Main nginx config
│   │   ├── conf.d/             # Additional configs
│   │   ├── mime.types          # MIME type definitions
│   │   └── fastcgi_params      # FastCGI parameters
│   ├── sites-available/        # Available site configs
│   │   ├── <slug>.conf         # Per-project nginx config
│   │   └── catch-all.conf      # Default 404 handler
│   ├── sites-enabled/          # Symlinks to active sites
│   │   ├── <slug>.conf -> ../sites-available/<slug>.conf
│   │   └── catch-all.conf
│   ├── certs/                  # SSL certificates
│   │   ├── <domain>.crt        # Certificate file
│   │   └── <domain>.key        # Private key file
│   └── logs/                   # Nginx logs
│       ├── access.log
│       └── error.log
│
├── cache/                      # Download cache
│   ├── php/                    # PHP source tarballs
│   │   ├── php-8.3.27.tar.gz
│   │   └── php-8.4.14.tar.gz
│   ├── composer/               # Composer PHAR files
│   │   └── composer.phar
│   └── nginx/                  # Nginx source tarballs
│       └── nginx-1.29.3.tar.gz
│
├── logs/                       # Command and service logs
│   ├── commands/               # Per-command execution logs
│   │   ├── link/
│   │   ├── install/
│   │   └── ...
│   └── services/               # Service logs (symlinks to actual logs)
│
├── system/                     # System integration state
│   ├── port-forwarding.json    # Tracked iptables port redirects
│   └── dnsmasq-config.json     # DNS configuration metadata
│
├── cli/                        # CLI templates (from repo)
│   └── templates/
│       └── nginx/              # Nginx configuration templates
│           ├── laravel.conf.tmpl
│           ├── wordpress.conf.tmpl
│           ├── proxy.conf.tmpl
│           └── static.conf.tmpl
│
└── composer/                   # Composer installation
    ├── composer.phar           # Composer PHAR binary
    └── composer.json           # Composer autoloader config
```

### Configuration Files

#### `~/.chauffeur/config/chauffeur.yaml`

Global configuration file controlling Chauffeur behavior:

```yaml
version: 1                      # Config version
telemetry: false                # Disable telemetry
workspace_dir: ~/.chauffeur     # Workspace location

nginx:
  enable: true                  # Enable nginx service
  http_port: 8080               # Default HTTP port
  https_port: 8443              # Default HTTPS port

php:
  default: 8.3                  # Default PHP version

ports:
  start_range: 8080            # Port allocation start
  end_range: 8099              # Port allocation end
  conflict_resolution: prompt   # prompt|auto|fail
  nginx_http_fallback: 8080     # Fallback HTTP port
  nginx_https_fallback: 8443    # Fallback HTTPS port
  php_fpm_fallback: 9000        # Fallback FPM port

projects_dir: ~/.chauffeur/projects  # Projects directory
```

#### `~/.chauffeur/projects/<slug>/project.yaml`

Per-project configuration file:

```yaml
version: 1
path: /absolute/path/to/project
php: 8.3

site:
  domain: slug.test
  ssl: true

domains:
  aliases:
    - domain: admin.test
      ssl: true
    - domain: api.test
      ssl: false

runtime:
  fpm:
    dedicated: false            # true for dedicated FPM
    socket: /path/to/php-fpm.sock

created_at: 2025-10-30T12:00:00+07:00
```

---

## Component Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        User / Developer                          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Interface                            │
│                    (cli/commands/*.go)                           │
│                                                                   │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐   │
│  │  link   │ │  start  │ │  stop   │ │ install │ │  doctor  │   │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬─────┘   │
│       │            │            │            │            │       │
└───────┼────────────┼────────────┼────────────┼────────────┼───────┘
        │            │            │            │            │
        ▼            ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Library Layer                               │
│                      (cli/lib/*.go)                              │
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐            │
│  │   Logger    │  │  Templates  │  │   Helpers    │            │
│  │   (output)  │  │ (nginx/PHP) │  │ (ports, etc) │            │
│  └─────────────┘  └─────────────┘  └──────────────┘            │
└─────────────────────────────────────────────────────────────────┘
        │            │            │            │            │
        ▼            ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Internal Services Layer                       │
│                 (cli/internal/*.go)                              │
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐            │
│  │  Projects   │  │  Services   │  │   System     │            │
│  │  (link/un)  │  │ (start/stop)│  │  (ports/DNS) │            │
│  └─────────────┘  └─────────────┘  └──────────────┘            │
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐            │
│  │ Templates   │  │  Workspace  │  │   Config     │            │
│  │  (render)   │  │  (paths)    │  │  (YAML I/O)  │            │
│  └─────────────┘  └─────────────┘  └──────────────┘            │
└─────────────────────────────────────────────────────────────────┘
        │            │            │            │            │
        ▼            ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Installer Layer                                │
│                 (cli/installers/*.go)                            │
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐            │
│  │     PHP     │  │   Nginx     │  │   Composer   │            │
│  │  (compile)  │  │  (build)    │  │   (PHAR)     │            │
│  └─────────────┘  └─────────────┘  └──────────────┘            │
└─────────────────────────────────────────────────────────────────┘
        │            │            │            │            │
        ▼            ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Workspace Filesystem                          │
│                      (~/.chauffeur/)                             │
│                                                                   │
│  Projects │ PHP Runtimes │ Nginx │ Cache │ Logs │ Config       │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### 1. CLI Commands Layer (`cli/commands/`)

**Responsibility**: User-facing command implementations

**Key Components**:
- `link.go`: Project registration and nginx configuration
- `start.go`: Service activation orchestration
- `stop.go`: Service deactivation and cleanup
- `restart.go`: Service restart with health checks
- `install.go`: Service installation coordination
- `doctor.go`: System health validation and fixing
- `logs.go`: Log viewing and following
- `clean.go`: Workspace maintenance

**Patterns**:
- Each command creates a logger: `logger := lib.NewCommandLogger("command")`
- Validate workspace state before operations
- Provide dry-run mode for safe testing
- Print exact commands for host-level changes

#### 2. Library Layer (`cli/lib/`)

**Responsibility**: Shared utilities and helper functions

**Key Components**:
- `logger.go`: Structured logging with colors, spinners, progress bars
- `templates.go`: Template rendering and file generation
- `helpers.go`: Common helper functions
- `ports.go`: Port allocation and conflict detection
- `downloads.go`: HTTP downloads with checksum verification
- `checksum.go`: File integrity verification

**Patterns**:
- Reusable across all commands
- No command-specific logic
- Pure functions when possible
- Clear error messages with context

#### 3. Internal Services Layer (`cli/internal/`)

**Responsibility**: Core business logic and state management

**Key Components**:
- `projects/`: Project registration and configuration management
- `services/`: Service lifecycle management (start/stop/restart)
- `system/`: Host integration (ports, DNS, networking)
- `templates/`: Template engine and config generation
- `workspace/`: Workspace path resolution and validation
- `config/`: Configuration file parsing and validation

**Patterns**:
- Export interfaces for testing
- Use dependency injection for external services
- Maintain separation of concerns
- Provide clear error types

#### 4. Installer Layer (`cli/installers/`)

**Responsibility**: Service installation and compilation

**Key Components**:
- `php.go`: PHP runtime compilation from source
- `php_legacy.go`: Legacy PHP (7.4, 8.0) specific handling
- `nginx.go`: Nginx binary building
- `composer.go`: Composer PHAR download and installation

**Patterns**:
- Dependency validation before starting
- Progress reporting for long operations
- Automatic caching of downloads
- Graceful failure handling

---

## PHP-FPM Strategies

Chauffeur provides two PHP-FPM strategies to balance resource efficiency and isolation needs.

### Shared FPM (Default)

**Architecture**: Multiple projects share the same PHP-FPM pool per PHP version.

```
PHP 8.3 Runtime
  └── Shared FPM Pool
      ├── php-fpm.sock (shared socket)
      └── Projects:
          ├── project-a.test (→ shared socket)
          ├── project-b.test (→ shared socket)
          └── project-c.test (→ shared socket)
```

**Configuration**:
```yaml
# project.yaml
runtime:
  fpm:
    dedicated: false
    socket: ~/.chauffeur/php/8.3/runtime/php-fpm/php-fpm.sock
```

**Benefits**:
- Resource efficient (one FPM process per PHP version)
- Lower memory footprint
- Suitable for typical development workflows

**Use Cases**:
- Standard development projects
- Resource-constrained environments
- Multiple projects with same PHP version

### Dedicated FPM (Optional)

**Architecture**: Each project gets its own isolated PHP-FPM pool.

```
Project A (PHP 8.3)
  └── Dedicated FPM Pool
      └── ~/.chauffeur/projects/project-a/runtime/php-fpm/php-fpm.sock

Project B (PHP 8.3)
  └── Dedicated FPM Pool
      └── ~/.chauffeur/projects/project-b/runtime/php-fpm/php-fpm.sock
```

**Configuration**:
```yaml
# project.yaml
runtime:
  fpm:
    dedicated: true
    socket: ~/.chauffeur/projects/myproject/runtime/php-fpm/php-fpm.sock
```

**Benefits**:
- Maximum isolation
- Custom php.ini per project
- Independent restart
- Different pool configurations

**Use Cases**:
- Production-mirroring environments
- Projects requiring custom PHP settings
- Security-sensitive applications
- Performance tuning needs

### Service Orchestration

**Service Manager** (`cli/internal/services/manager.go`) handles mixed strategies:

```go
type Service struct {
    Type     ServiceType    // php-fpm, nginx
    Name     string         // php-fpm-8.3, myproject-dedicated
    Scope    Scope          // global, project
    Status   Status         // running, stopped
    PidFile  string
}

func (m *Manager) StartServices(services []Service) error
func (m *Manager) StopServices(services []Service) error
func (m *Manager) RestartServices(services []Service) error
func (m *Manager) GetServiceStatus() ([]Service, error)
```

**Status Display**:
```
chauf status --detail

Global Services:
  nginx                    running  pid: 1234  uptime: 2h  mem: 45M
  php-fpm-8.3 (shared)    running  pid: 5678  workers: 5

Project Services (myproject):
  php-fpm-8.3 (dedicated) running  pid: 9012  workers: 2
```

---

## Multi-Domain Architecture

Chauffeur supports multiple domains per project with isolated SSL certificate management.

### Data Structures

```go
type Config struct {
    Version   int
    Path      string
    PHP       string
    Site      *Site           // Primary domain
    Domains   *Domains        // Alias domains
    Runtime   Runtime
    CreatedAt time.Time
}

type Site struct {
    Domain string
    SSL    bool
}

type DomainAlias struct {
    Domain string `yaml:"domain"`
    SSL    bool   `yaml:"ssl"`
}

type Domains struct {
    Aliases []DomainAlias `yaml:"aliases,omitempty"`
}
```

### SSL Certificate Management

**Multi-Domain SAN Certificates**:

Single certificate covers all SSL-enabled domains (primary + aliases):

```bash
# Certificate generation
mkcert -install
mkcert -cert-file hja-cms.test.crt -key-file hja-cms.test.key \
  hja-cms.test admin.hja-cms.test api.hja-cms.test
```

**File Naming**: Uses base domain as certificate filename:
- `hja-cms.test.crt` (covers all domains)
- `hja-cms.test.key`

**Automatic Regeneration**: Certificates regenerated when SSL aliases are added/removed.

### Domain Resolution

**Helper Methods**:

```go
// GetAllDomains returns primary + all aliases
func (p *Project) GetAllDomains() []DomainAlias

// GetServerNames returns nginx server_name directives
func (p *Project) GetServerNames() []string

// GetSSLDomains returns only SSL-enabled domains
func (p *Project) GetSSLDomains() []DomainAlias

// HasSSLEnabled checks if any domain has SSL
func (p *Project) HasSSLEnabled() bool
```

**nginx Integration**:

```nginx
# Generated server block
server {
    listen 8443 ssl;
    server_name myapp.test admin.myapp.test api.myapp.test;

    ssl_certificate /path/to/myapp.test.crt;
    ssl_certificate_key /path/to/myapp.test.key;
    # ...
}
```

### User Experience

**Links Display**:
```
chauf links

SLUG    PATH                    DOMAIN           ALIAS                              SSL
------  ----------------------  ---------------  ---------------------------------  ---
myapp   /home/user/myapp        myapp.test       admin.myapp.test (*), api.myapp (*)  *
```

`(*)` indicates SSL-enabled aliases.

**Link Command**:
```bash
# Add aliases to existing project
chauf link --alias www.myapp.test --secure
chauf link --alias api.myapp.test    # HTTP only
```

---

## Service Orchestration

### Service Types

| Service | Description | Binary Location | Config Location |
|---------|-------------|-----------------|-----------------|
| `nginx` | Web server | `~/.chauffeur/nginx/bin/nginx` | `~/.chauffeur/nginx/etc/nginx.conf` |
| `php-fpm-<version>` | PHP FastCGI | `~/.chauffeur/php/<version>/bin/php-fpm` | `~/.chauffeur/php/<version>/etc/php-fpm.conf` |
| `php-fpm-<project>` | Dedicated FPM | `~/.chauffeur/php/<version>/bin/php-fpm` | `~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.conf` |

### Service Lifecycle

**Start**:
1. Validate dependencies (ports, configs)
2. Check for conflicts (already running services)
3. Create runtime directories if missing
4. Start service with appropriate config
5. Verify startup (PID file, socket creation)
6. Log startup with duration

**Stop**:
1. Find service PID from PID file or process name
2. Gracefully terminate (SIGTERM)
3. Wait for shutdown (timeout: 10s)
4. Force kill if necessary (SIGKILL)
5. Clean up runtime artifacts (sockets, temp files)
6. Log shutdown with duration

**Restart**:
1. Stop service (if running)
2. Wait for shutdown
3. Start service
4. Verify startup
5. Log restart with total duration

### Health Monitoring

**Health Checks**:
- PID file exists and contains valid PID
- Process is running (PID exists in process table)
- Socket file exists (for PHP-FPM)
- Listening port is bound (for nginx)
- Log file is writable

**Status Output**:
```
chauf status --detail

Service              Status  PID    Uptime    Memory    Port
--------------------  ------  -----  --------  --------  -----
nginx                ●       1234   2h 15m    45.2 MB   8080, 8443
php-fpm-8.3 (shared) ●       5678   2h 14m    128.5 MB  -
myproject (dedicated) ●      9012   1h 30m    64.1 MB   -

Legend: ● running, ○ stopped, ⚠ warning, ✗ error
```

---

## Data Flow

### Project Link Flow

```
User runs: chauf link --secure --php 8.3
                │
                ▼
    ┌───────────────────────┐
    │  Validate Command     │
    │  - Check PHP version  │
    │  - Check path exists  │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Detect Project Type  │
    │  - Laravel            │
    │  - WordPress          │
    │  - General            │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Generate Config      │
    │  - project.yaml       │
    │  - nginx server block │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  SSL Certificate      │
    │  - Generate SAN cert  │
    │  - Save to nginx/certs│
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  PHP-FPM Setup        │
    │  - Shared (default)   │
    │  - Dedicated (flag)   │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Nginx Configuration  │
    │  - Render template    │
    │  - Enable site        │
    │  - Test configuration │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Nginx Reload         │
    │  - Check if running   │
    │  - Reload or start    │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Complete             │
    │  - Print success msg  │
    │  - Show URL           │
    └───────────────────────┘
```

### Service Start Flow

```
User runs: chauf start
                │
                ▼
    ┌───────────────────────┐
    │  Load Config          │
    │  - chauffeur.yaml     │
    │  - Enabled services   │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Check Ports          │
    │  - Default ports      │
    │  - Resolve conflicts  │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Port Forwarding      │
    │  - 80 → 8080          │
    │  - 443 → 8443         │
    │  - Track rules        │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  DNS Validation       │
    │  - Check dnsmasq      │
    │  - Print setup cmd    │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Start Nginx          │
    │  - Validate config    │
    │  - Start process      │
    │  - Verify startup     │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Start PHP-FPM        │
    │  - All installed ver  │
    │  - Shared pools       │
    │  - Dedicated pools    │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Complete             │
    │  - Show status        │
    │  - Log duration       │
    └───────────────────────┘
```

---

## Configuration Management

### Global Configuration (`~/.chauffeur/config/chauffeur.yaml`)

**Structure**:
```go
type Config struct {
    Version   int              `yaml:"version"`
    Telemetry bool             `yaml:"telemetry"`
    WorkspaceDir string        `yaml:"workspace_dir"`
    Nginx     NginxConfig      `yaml:"nginx"`
    PHP       PHPConfig        `yaml:"php"`
    Ports     PortsConfig      `yaml:"ports"`
    ProjectsDir string         `yaml:"projects_dir"`
}
```

**Loading**:
```go
func LoadConfig(path string) (*Config, error)
func (c *Config) Save(path string) error
func (c *Config) Validate() error
```

### Project Configuration (`~/.chauffeur/projects/<slug>/project.yaml`)

**Structure**:
```go
type ProjectConfig struct {
    Version   int              `yaml:"version"`
    Path      string           `yaml:"path"`
    PHP       string           `yaml:"php"`
    Site      *Site            `yaml:"site"`
    Domains   *Domains         `yaml:"domains"`
    Runtime   Runtime          `yaml:"runtime"`
    CreatedAt time.Time        `yaml:"created_at"`
}
```

**Operations**:
```go
func LoadProject(slug string) (*Project, error)
func (p *Project) Save() error
func (p *Project) Validate() error
func (p *Project) GetPHPVersion() string
func (p *Project) IsSecure() bool
```

---

## Security Model

### Isolation

**Workspace Isolation**:
- All services run under user's UID/GID
- No privilege escalation required
- No setuid/setgid binaries

**Process Isolation**:
- PHP-FPM pools: Separate per version and project
- Nginx: Single instance, separate worker processes
- No containerization (by design)

### File Permissions

**Directories**:
- `~/.chauffeur/`: `700` (user only)
- `logs/`: `755` (readable by user)
- `bin/`: `755` (executable by user)

**Configuration Files**:
- `*.yaml`: `600` (user read/write only)
- `*.conf`: `644` (nginx can read)
- SSL keys: `600` (user only)
- SSL certs: `644` (nginx can read)

### Network Security

**Port Binding**:
- Default ports: 8080 (HTTP), 8443 (HTTPS)
- Optional port forwarding: 80→8080, 443→8443
- PHP-FPM: Unix sockets (not TCP)

**SSL/TLS**:
- Local development certificates (mkcert)
- Trusted by local browsers
- Not suitable for production exposure

### Secrets Management

**No Secrets Storage**:
- Chauffeur does not store passwords or API keys
- Environment variables handled by host shell
- Project `.env` files not managed by Chauffeur

**Future: `chauf env` command**:
Will allow project environment variable management (see [Plan.md](Plan.md)).

---

## Extension Points

### Custom Nginx Templates

Users can provide custom nginx templates:

```
~/.chauffeur/cli/templates/nginx/
  ├── custom-app.conf.tmpl    # User-provided template
  └── ...
```

**Template Variables**:
- `{{.ServerName}}` - Domain name
- `{{.DocumentRoot}}` - Project path
- `{{.PHPSocket}}` - PHP-FPM socket path
- `{{.SSLCert}}` - SSL certificate path
- `{{.SSLKey}}` - SSL private key path
- `{{.HTTPPort}}` - HTTP port
- `{{.HTTPSPort}}` - HTTPS port

### Custom PHP Extensions

Users can compile custom PHP extensions:

```bash
# Build extension with phpize
cd ~/.chauffeur/php/8.3/bin
./phpize
./configure --with-php-config=/path/to/php-config
make

# Add to php.ini or conf.d/
echo "extension=custom.so" >> ~/.chauffeur/php/8.3/etc/conf.d/custom.ini
```

### Service Hooks (Future)

Planned hooks for service lifecycle events:

```yaml
# project.yaml
hooks:
  pre-start: ["echo 'Starting...'"]
  post-start: ["echo 'Started!'"]
  pre-stop: ["echo 'Stopping...'"]
  post-stop: ["echo 'Stopped!'"]
```

---

## Performance Considerations

### Resource Limits

**PHP-FPM Pools**:
- Shared: 5 workers (default)
- Dedicated: 2 workers (default, per project)
- Configurable in `php-fpm.conf`

**Nginx**:
- `worker_processes: auto`
- `worker_connections: 1024`
- `keepalive_timeout: 65`

### Startup Time

**Service Startup**:
- Nginx: ~100ms
- PHP-FPM: ~200ms per pool
- Total (3 projects): ~700ms

**Project Link**:
- Config generation: ~50ms
- Nginx reload: ~100ms
- SSL generation: ~500ms (mkcert)
- Total: ~650ms

### Memory Usage

**Typical Memory Footprint**:
- Nginx: ~45 MB
- PHP-FPM (shared, 5 workers): ~128 MB
- PHP-FPM (dedicated, 2 workers): ~64 MB
- Per project (dedicated FPM): +64 MB

**Example**: 10 projects with shared PHP 8.3:
- Nginx: 45 MB
- PHP-FPM (shared): 128 MB
- **Total**: ~173 MB

---

## Monitoring and Observability

### Logging

**Command Logs**:
- Location: `~/.chauffeur/logs/commands/<command>/`
- Format: `<timestamp>.log`
- Retention: 30 days (configurable)

**Service Logs**:
- Nginx access: `~/.chauffeur/nginx/logs/access.log`
- Nginx error: `~/.chauffeur/nginx/logs/error.log`
- PHP-FPM: `~/.chauffeur/php/<version>/var/log/php-fpm.log`
- Project FPM: `~/.chauffeur/projects/<slug>/logs/php-fpm.log`

### Health Checks

**`chauf doctor`** validates:
- System dependencies (git, curl, tar, gcc, make, pkg-config)
- PHP build dependencies (libzip, libjpeg, libpng, freetype, etc.)
- SSL certificate dependencies (openssl, mkcert)
- Network and firewall (iptables, port availability)
- DNS configuration (dnsmasq, .test domain resolution)

**`chauf status --detail`** shows:
- Service status (running/stopped)
- Process information (PID, uptime, memory)
- Port bindings
- Socket files
- Recent log entries

---

## Error Handling

### Error Categories

**User Errors**:
- Invalid command syntax
- Missing dependencies
- Permission issues
- Port conflicts

**System Errors**:
- Service startup failures
- Configuration errors
- Network issues
- Filesystem problems

**Development Errors**:
- Bugs in code
- Incorrect assumptions
- Race conditions

### Error Recovery

**Automatic Recovery**:
- Port conflicts: Auto-select alternative port (if configured)
- Service crashes: Restart with backoff
- Network failures: Retry with exponential backoff

**Manual Recovery**:
- `chauf doctor --auto-fix` - Fix common issues
- `chauf stop && chauf start` - Restart services
- `chauf clean logs` - Clear log files
- `chauf reinstall <service>` - Reinstall service

### Error Messages

**Structure**:
```
✗ Error category: Specific error message

Context:
  - Detail 1
  - Detail 2

Fix:
  1. Step one
  2. Step two

Log: ~/.chauffeur/logs/commands/<command>/<timestamp>.log
```

**Example**:
```
✗ Port conflict: Port 8080 is already in use

Context:
  - Port 8080 is bound by process nginx (PID 1234)
  - This port is configured as the HTTP port in chauffeur.yaml

Fix:
  1. Stop the conflicting service: sudo systemctl stop nginx
  2. Or configure a different port: chauf config set nginx.http_port 8081
  3. Or enable auto port resolution: chauf config set ports.conflict_resolution auto

Log: ~/.chauffeur/logs/commands/start/2025-01-12_10-30-45.log
```

---

## Future Architecture Enhancements

See [Plan.md](Plan.md) for planned architectural improvements.

### Short-Term (3 months)
- Enhanced monitoring and metrics collection
- Configuration management system (`chauf config`)
- Environment variable management (`chauf env`)

### Medium-Term (6-12 months)
- Service auto-start integration (systemd)
- Universal update management system
- Advanced debugging and profiling tools

### Long-Term (12+ months)
- Plugin system for custom services
- Multi-workspace support
- Distributed configuration sync

---

**See Also**:
- [AGENTS.md](../AGENTS.md) - Agent handbook with implementation rules
- [Conventions.md](Conventions.md) - Development conventions and practices
- [TechStack.md](TechStack.md) - Technology stack and dependencies
- [Plan.md](Plan.md) - Project roadmap and milestones
