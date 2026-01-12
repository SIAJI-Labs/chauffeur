# Chauffeur Technology Stack

> **Last Updated**: 2025-01-12
> **Maintainer**: @si-aji
> **Status**: Active - Reflects current dependencies and versions

## Table of Contents

1. [Overview](#overview)
2. [Primary Languages](#primary-languages)
3. [Core Dependencies](#core-dependencies)
4. [Service Versions](#service-versions)
5. [Host Requirements](#host-requirements)
6. [Build Tools](#build-tools)
7. [Testing Framework](#testing-framework)
8. [Development Tools](#development-tools)
9. [Platform Support](#platform-support)
10. [Version Policy](#version-policy)

---

## Overview

Chauffeur is built primarily in Go 1.22+ and manages multiple runtime services (PHP, nginx, Composer) within an isolated workspace. This document details the complete technology stack, version requirements, and dependencies.

### Technology Philosophy

1. **Minimal Dependencies**: Prefer Go standard library over external packages
2. **Static Linking**: Produce static binaries where possible for portability
3. **Workspace Isolation**: All runtimes live under `~/.chauffeur/`
4. **System Compatibility**: Work with standard Linux packaging, not against it

---

## Primary Languages

### Go (Golang)

**Version**: Go 1.22 or later

**Purpose**: CLI implementation, service orchestration, configuration management

**Usage**:
- All CLI commands in `cli/commands/`
- Shared utilities in `cli/lib/`
- Internal services in `cli/internal/`
- Service installers in `cli/installers/`

**Why Go**:
- Compiled to single binary
- Excellent concurrency support
- Cross-platform compilation
- Strong standard library
- Fast compilation

### Shell Scripts

**Purpose**: Installation and bootstrapping

**Files**:
- `install.sh` - Main installation script
- Various helper scripts in `cli/bin/helpers/`

**Usage**:
- Bootstrap initial installation
- Set up PATH environment
- Create workspace directories
- Generate systemd service files (future)

### Template Languages

**Purpose**: Configuration file generation

**Types**:
- **Go templates** (`text/template`) - Nginx, PHP-FPM configs
- **YAML** - Project and global configuration files

---

## Core Dependencies

### Go Modules

**Primary Go dependencies** (from `go.mod`):

```go
module github.com/SIAJI-Labs/chauffeur

go 1.22

require (
    // Standard library only (no external dependencies for CLI core)
)
```

**Rationale**:
- Zero external dependencies reduces attack surface
- Standard library provides all necessary functionality
- Easier maintenance and security updates
- Faster compilation

### Standard Library Packages Used

| Package | Purpose |
|---------|---------|
| `encoding/yaml` | Configuration file parsing |
| `text/template` | Config file generation |
| `os/exec` | Process spawning |
| `net/http` | HTTP downloads, API calls |
| `io` | File I/O, stream handling |
| `path/filepath` | Path manipulation |
| `syscall` | Process management, signals |
| `time` | Timestamps, duration formatting |
| `fmt` | String formatting (limited use) |
| `log` | Internal logging infrastructure |

---

## Service Versions

### PHP Runtimes

**Supported Versions**: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4

**Source**: Official PHP releases from `php.net`

**Compilation**:
- Built from source tarballs
- Compiled extensions: mysqli, PDO_MySQL, mysqlnd, gd, zip, exif, freetype, imagick, gmp, bcmath, sodium, readline, libxml, curl, openssl
- Configured via `php.ini` and `php-fpm.conf`

**PHP Installation Paths**:
```
~/.chauffeur/php/<version>/
  ├── bin/php, php-cgi, php-fpm
  ├── etc/php.ini, php-fpm.conf, conf.d/
  ├── lib/php/extensions/
  ├── include/php/
  ├── runtime/php-fpm/ (shared pool)
  └── var/log/
```

**Extensions by Default**:
- Database: mysqli, pdo_mysql, mysqlnd
- Images: gd, freetype, jpeg, png, imagick (PECL)
- Archives: zip, bz2, zlib
- Web: curl, openssl, libxml, libxslt
- Math: gmp, bcmath
- Crypto: sodium
- CLI: readline

**Legacy PHP Notes**:
- PHP 7.4, 8.0 require additional compatibility patches
- GD extension compilation in progress (infrastructure ready)
- See [docs/TODO_STATUS.md](TODO_STATUS.md) for current status

### Nginx

**Version**: Latest stable from `nginx.org`

**Source**: Official nginx releases

**Configuration**:
- Single instance for all projects
- Worker processes: `auto` (CPU count)
- Worker connections: 1024
- HTTP port: 8080 (default)
- HTTPS port: 8443 (default)

**Nginx Installation Path**:
```
~/.chauffeur/nginx/
  ├── bin/nginx
  ├── etc/nginx.conf, mime.types, fastcgi_params
  ├── sites-available/<project>.conf
  ├── sites-enabled/<project>.conf -> ../sites-available/
  ├── certs/<domain>.crt, <domain>.key
  └── logs/access.log, error.log
```

**Modules Enabled**:
- `http_ssl_module` - SSL/TLS support
- `http_gzip_static_module` - Gzip compression
- `http_proxy_module` - Proxy support (future)
- `http_fastcgi_module` - PHP-FPM integration

### Composer

**Version**: Latest stable (2.x) from `getcomposer.org`

**Source**: Composer PHAR binary

**Installation**:
- Downloaded as PHAR file
- Installed to `~/.chauffeur/composer/composer.phar`
- Shim created at `~/.chauffeur/bin/composer`

**Composer Integration**:
- Uses Chauffeur's PHP shim for version isolation
- Respects project PHP version automatically
- Downloads cached in `~/.chauffeur/cache/composer/`

### mkcert (Optional)

**Purpose**: Local trusted SSL certificates

**Version**: Latest stable from GitHub releases

**Installation**:
- Host-provided package or manual install
- Used by `chauf secure` command
- Generates certificates trusted by local browsers

**Certificate Storage**:
- Location: `~/.chauffeur/nginx/certs/`
- Naming: `<base-domain>.crt`, `<base-domain>.key`
- Multi-domain SAN certificates supported

---

## Host Requirements

### Required Host Tools

**Core Tools** (must be installed on host):

| Tool | Purpose | Version |
|------|---------|---------|
| `git` | Version control, source downloads | 2.0+ |
| `curl` | HTTP downloads, API calls | 7.0+ |
| `tar` | Archive extraction | 1.0+ |

**Build Tools** (for PHP compilation):

| Tool | Purpose | Version |
|------|---------|---------|
| `gcc` | C compiler | 8.0+ |
| `make` | Build automation | 4.0+ |
| `pkg-config` | Library detection | 1.0+ |
| `autoconf` | Build configuration (PHP) | 2.69+ |
| `bison` | Parser generator (PHP) | 3.0+ |
| `re2c` | Lexer generator (PHP) | 1.0+ |

### Required Development Libraries

**PHP Build Dependencies** (must be installed on host):

**Common Libraries**:
```
libzip        - ZIP archive support
libjpeg       - JPEG image support
libpng        - PNG image support
libfreetype   - Font rendering
libxml2       - XML parsing
libcurl       - HTTP client
zlib          - Compression
libbz2        - Bzip2 compression
libxslt       - XSLT transformations
libreadline   - Command line editing
libmagickwand - ImageMagick support (imagick extension)
libgmp        - Arbitrary precision math
libsodium     - Modern cryptography
```

**Version Compatibility**:

| Library | Modern PHP (8.1+) | Legacy PHP (7.4, 8.0) |
|---------|-------------------|----------------------|
| libxml2 | 2.9+ (2.12+ tested) | 2.9-2.11 (2.12+ has breaking changes) |
| libcurl | 7.0+ (8.0+ tested) | 7.0+ (8.0+ may have issues) |
| ImageMagick | 7.1+ | 6.9-7.0 (7.1+ may have compatibility issues) |

**Distribution Installation Commands**:

**Arch Linux**:
```bash
sudo pacman -S base-devel pkgconf libzip libjpeg-turbo libpng freetype2 \
  libxml2 curl bzip2 zlib libxslt readline imagemagick gmp libsodium
```

**Debian/Ubuntu**:
```bash
sudo apt-get install build-essential pkg-config autoconf bison re2c \
  libzip-dev libjpeg-dev libpng-dev libfreetype6-dev libxml2-dev \
  libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev libreadline-dev \
  libmagickwand-dev libgmp-dev libsodium-dev
```

**Fedora/RHEL**:
```bash
sudo dnf install gcc make pkg-config autoconf bison re2c \
  libzip-devel libjpeg-devel libpng-devel freetype-devel libxml2-devel \
  libcurl-devel bzip2-devel zlib-devel libxslt-devel readline-devel \
  ImageMagick-devel gmp-devel libsodium-devel
```

### Optional Host Tools

**Network and DNS**:
- `dnsmasq` - Local DNS server for `.test` domains
- `iptables` - Port forwarding (80→8080, 443→8443)
- `NetworkManager` - DNS integration (optional)

**SSL Certificates**:
- `mkcert` - Local trusted certificate authority
- `openssl` - SSL/TLS toolkit (required for PHP)

**Monitoring**:
- `systemctl` - Service management (for systemd integration)

---

## Build Tools

### Go Build Tools

**Primary Build**:
```bash
# Build development binary
go build -o chauf cli/main.go

# Build with optimizations
go build -ldflags="-s -w" -o chauf cli/main.go

# Cross-compile for different architectures
GOOS=linux GOARCH=amd64 go build -o chauf cli/main.go
GOOS=linux GOARCH=arm64 go build -o chauf cli/main.go
```

**Version Information**:
```bash
# Build with version info (future)
go build -ldflags="-X main.Version=1.0.0" -o chauf cli/main.go
```

### PHP Build Tools

**Configuration**:
```bash
# PHP configure script
./configure \
  --prefix=$HOME/.chauffeur/php/8.3 \
  --enable-fpm \
  --with-fpm-user=$USER \
  --with-fpm-group=$USER \
  --enable-mysqlnd \
  --with-mysqli=mysqlnd \
  --with-pdo-mysql=mysqlnd \
  --with-openssl \
  --with-zlib \
  --with-curl \
  --with-gd \
  --with-jpeg \
  --with-freetype \
  --with-zip \
  --enable-bcmath \
  --enable-gmp \
  --with-readline \
  --with-libxml \
  --with-xsl \
  --with-sodium \
  --enable-opcache
```

**Compilation**:
```bash
# Build PHP (uses all available cores)
make -j$(nproc)

# Install to workspace
make install
```

### Nginx Build Tools

**Configuration**:
```bash
# Nginx configure script
./configure \
  --prefix=$HOME/.chauffeur/nginx \
  --sbin-path=$HOME/.chauffeur/nginx/bin/nginx \
  --conf-path=$HOME/.chauffeur/nginx/etc/nginx.conf \
  --http-log-path=$HOME/.chauffeur/nginx/logs/access.log \
  --error-log-path=$HOME/.chauffeur/nginx/logs/error.log \
  --pid-path=$HOME/.chauffeur/nginx/logs/nginx.pid \
  --with-http_ssl_module \
  --with-http_gzip_static_module \
  --with-http_v2_module
```

**Compilation**:
```bash
# Build nginx
make

# Install to workspace
make install
```

---

## Testing Framework

### Go Testing

**Framework**: Built-in `testing` package

**Test Organization**:
- Unit tests: `*_test.go` alongside source files
- Integration tests: `tests/` directory
- Benchmark tests: `*_benchmark_test.go` (future)

**Running Tests**:
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with coverage profile
go test ./... -coverprofile=coverage.out

# View coverage as HTML
go tool cover -html=coverage.out

# Run specific test
go test ./cli/commands -run TestLink

# Run tests with verbose output
go test ./... -v

# Run tests with race detection
go test ./... -race
```

**Coverage Target**: ≥80% for new code

### Test Utilities

**Workspace Isolation**:
```go
func TestSomething(t *testing.T) {
    tempDir := t.TempDir()  // Creates temp dir, auto-cleanup
    t.Setenv("HOME", tempDir)  // Isolate from real workspace
    // ... test code
}
```

**Table-Driven Tests**:
```go
func TestParseVersion(t *testing.T) {
    tests := []struct {
        input    string
        expected string
        wantErr  bool
    }{
        {"8.3", "8.3", false},
        {"8.3.12", "8.3", false},
        {"invalid", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result, err := ParseVersion(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("ParseVersion() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

---

## Development Tools

### IDE Support

**Recommended IDEs**:
- **VS Code** with Go extension
- **GoLand** (JetBrains)
- **Vim/Neovim** with vim-go
- **Emacs** with go-mode

### Essential Go Tools

**From `golang.org/x/tools`**:

```bash
# Install essential tools
go install golang.org/x/tools/gopls@latest         # LSP server
go install golang.org/x/tools/cmd/goimports@latest # Import management
go install github.com/fatih/gomodifytags@latest   # Struct tag manipulation
go install github.com/cweill/gotests/...@latest   # Test generation
```

**Usage**:
- `gopls` - Language Server Protocol for Go
- `goimports` - Manage imports automatically
- `go vet` - Static analysis
- `golint` - Style checker (optional)

### Code Quality Tools

**Static Analysis**:
```bash
# Run vet (built-in)
go vet ./...

# Run staticcheck (optional)
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Run nilaway (nil pointer analysis, optional)
go install go.uber.org/nilaway/cmd/nilaway@latest
nilaway ./...
```

**Formatting**:
```bash
# Format all Go code
go fmt ./...

# Or use goimports (imports + format)
goimports -w .
```

---

## Platform Support

### Target Platforms

**Primary Support**:
- ✅ Arch Linux (rolling)
- ✅ Ubuntu 22.04+ (LTS)
- ✅ Debian 12+ (stable)

**Secondary Support** (tested but not primary):
- ✅ Fedora 39+
- ✅ Rocky Linux 9+
- ✅ AlmaLinux 9+

**Future Support** (not tested):
- 🚧 openSUSE Tumbleweed
- 🚧 Gentoo Linux

### Architecture Support

**Current**:
- ✅ x86_64 (amd64) - Primary
- ✅ ARM64 (aarch64) - Tested on Raspberry Pi 4+

**Future**:
- 🚧 ARMv7 (arm) - Raspberry Pi 3 and older
- 🚧 RISC-V - Community interest

### Platform Differences

**Package Manager Differences**:

| Platform | Package Manager | PHP Build Deps Command |
|----------|----------------|------------------------|
| Arch | `pacman` | `sudo pacman -S base-devel pkgconf libzip...` |
| Ubuntu/Debian | `apt` | `sudo apt-get install build-essential libzip-dev...` |
| Fedora/RHEL | `dnf` | `sudo dnf install gcc make libzip-devel...` |
| openSUSE | `zypper` | `sudo zypper install gcc make libzip-devel...` |

**Library Path Differences**:

| Platform | Library Location | pkg-config |
|----------|-----------------|------------|
| Arch | `/usr/lib/`, `/usr/lib64/` | Standard |
| Debian/Ubuntu | `/usr/lib/x86_64-linux-gnu/` | Multiarch |
| Fedora/RHEL | `/usr/lib64/` | Standard |

---

## Version Policy

### Semantic Versioning

**Chauffeur follows Semantic Versioning 2.0.0**:

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality
- **PATCH**: Backward-compatible bug fixes

**Example**: `v1.2.3`
- MAJOR: 1
- MINOR: 2
- PATCH: 3

### Dependency Version Policy

**Go**: Minimum version 1.22 (no maximum yet)
- Follow Go 1.x compatibility guarantee
- Test with latest Go release

**PHP**: Support multiple versions simultaneously
- Security updates: All supported versions
- New features: Latest stable (8.3+)
- EOL versions: Best effort, no guarantees

**Nginx**: Latest stable from nginx.org
- Security updates: As released
- New features: When stable

**Composer**: Latest stable 2.x
- Security updates: As released
- New features: When stable

### Update Strategy

**Go Toolchain**:
- Update to latest stable Go 1.x when available
- Test compatibility before requiring new version

**PHP Versions**:
- Add new PHP versions within 1 month of stable release
- Maintain existing versions until EOL
- Default version tracks latest stable

**Service Versions**:
- nginx: Update to latest stable monthly
- Composer: Update to latest stable monthly

---

## Security Considerations

### Vulnerability Management

**PHP Vulnerabilities**:
- Monitor PHP security announcements
- Update affected versions promptly
- Notify users of critical vulnerabilities

**Nginx Vulnerabilities**:
- Monitor nginx security advisories
- Update on critical issues
- Provide migration instructions

**Dependency Vulnerabilities**:
- Use Go standard library (minimal attack surface)
- Monitor CVE databases
- Prompt updates for critical issues

### Build Security

**Source Verification**:
- PHP: Verify GPG signatures from php.net
- Nginx: Verify SHA256 checksums from nginx.org
- Composer: Verify SHA256 checksums from getcomposer.org

**Download Security**:
- Use HTTPS only
- Verify checksums after download
- Cache verified downloads in `~/.chauffeur/cache/`

### Runtime Security

**Process Isolation**:
- All services run as unprivileged user
- No setuid/setgid binaries
- Unix sockets for PHP-FPM (no TCP)

**File Permissions**:
- Config files: 600 (user only)
- SSL keys: 600 (user only)
- SSL certs: 644 (readable by nginx)
- Binaries: 755 (executable by user)

---

## Performance Considerations

### Build Performance

**PHP Compilation**:
- Time: 5-10 minutes on modern hardware
- Parallel: `make -j$(nproc)` uses all cores
- Cache: Source tarballs cached for rebuilds

**Nginx Compilation**:
- Time: 1-2 minutes on modern hardware
- Minimal dependencies (fast build)

### Runtime Performance

**Memory Usage**:
- nginx: ~45 MB base
- PHP-FPM (shared): ~128 MB (5 workers)
- PHP-FPM (dedicated): ~64 MB (2 workers)

**Startup Time**:
- nginx: ~100ms
- PHP-FPM pool: ~200ms
- Full stack (10 projects): ~700ms

### Disk Usage

**Per PHP Version**: ~200-300 MB (including extensions)
**Nginx**: ~20 MB
**Composer**: ~3 MB
**Typical Workspace** (3 PHP versions, nginx, composer): ~1 GB

---

## Compatibility Matrix

### PHP Extension Support

| Extension | PHP 7.4 | PHP 8.0 | PHP 8.1+ | Notes |
|-----------|---------|---------|----------|-------|
| mysqli | ✅ | ✅ | ✅ | mysqlnd-based |
| pdo_mysql | ✅ | ✅ | ✅ | mysqlnd-based |
| gd | ⚠️ | ⚠️ | ✅ | Legacy PHP in progress |
| zip | ✅ | ✅ | ✅ | libzip-based |
| curl | ✅ | ✅ | ✅ | libcurl-based |
| openssl | ✅ | ✅ | ✅ | Auto-configuration |
| imagick | ✅ | ✅ | ✅ | PECL, latest stable |
| gmp | ✅ | ✅ | ✅ | Compiled in |
| bcmath | ✅ | ✅ | ✅ | Compiled in |
| sodium | ✅ | ✅ | ✅ | Modern crypto |
| readline | ✅ | ✅ | ✅ | CLI history |

### Distribution Compatibility

| Distribution | Go 1.22 | PHP 8.3 | Nginx | mkcert | dnsmasq |
|--------------|---------|---------|-------|--------|---------|
| Arch | ✅ | ✅ | ✅ | ✅ | ✅ |
| Ubuntu 22.04+ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Debian 12+ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Fedora 39+ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Rocky 9+ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## Future Technology Additions

### Planned Additions

See [Plan.md](Plan.md) for detailed roadmap:

**Short Term** (3 months):
- Enhanced monitoring infrastructure
- Configuration management system
- Environment variable management

**Medium Term** (6-12 months):
- Systemd auto-start integration
- Universal update management
- Advanced debugging tools

**Long Term** (12+ months):
- Plugin system for custom services
- Multi-workspace support
- Distributed configuration

### Technology Under Consideration

**Monitoring**:
- Prometheus metrics export
- OpenTelemetry tracing
- Grafana dashboards

**Configuration**:
- JSON Schema validation
- Configuration versioning
- Distributed sync (etcd, Consul)

**Security**:
- Certificate automation (Let's Encrypt for staging)
- Secret management integration
- Audit logging

---

## Resources

### Official Documentation

- [Go Documentation](https://go.dev/doc/)
- [PHP Documentation](https://www.php.net/docs.php)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Composer Documentation](https://getcomposer.org/doc/)

### External Libraries

- [PECL :: PHP Extension Community Library](https://pecl.php.net/)
- [ImageMagick](https://imagemagick.org/)
- [mkcert - Local HTTPS](https://github.com/FiloSottile/mkcert)

### Build Documentation

- [PHP Installation on Linux](https://www.php.net/manual/en/install.unix.php)
- [Building Nginx from Source](https://nginx.org/en/docs/configure.html)
- [Compiling PHP Extensions](https://www.php.net/manual/en/install.pecl.phpize.php)

---

**See Also**:
- [Architecture.md](Architecture.md) - System architecture and design
- [Conventions.md](Conventions.md) - Development conventions
- [Plan.md](Plan.md) - Project roadmap
