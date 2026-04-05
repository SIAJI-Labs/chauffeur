# Tech Stack

## Overview

Chauffeur V2 is built in Go with zero external dependencies, managing multiple runtime services (PHP, nginx, Composer) within an isolated user workspace.

### Philosophy

1. **Minimal Dependencies** — Use only Go standard library. No external packages.
2. **Static Binaries** — Produce statically-linked binaries for easy distribution.
3. **Workspace Isolation** — All runtimes live under `~/.chauffeur/`.
4. **System Compatibility** — Work with standard Linux packaging, not against it.

---

## Primary Language

### Go 1.22+

**Purpose**: All CLI logic — command dispatch, service orchestration, config management, process control, template rendering.

**Why Go**:
- Single binary distribution
- Excellent concurrency for parallel operations
- Cross-compilation (amd64, arm64)
- Strong standard library (no deps needed)
- Fast startup time (< 100ms)

**Key standard library packages used**:

| Package | Purpose |
|---------|---------|
| `encoding/json` | JSON config serialization |
| `gopkg.in/yaml.v3` | YAML config parsing (only allowed external dep) |
| `text/template` | Nginx/PHP-FPM config generation |
| `os/exec` | Spawning nginx, PHP-FPM, mkcert processes |
| `net/http` | Downloading PHP, nginx, Composer tarballs |
| `io` | File I/O, stream handling |
| `path/filepath` | Cross-platform path manipulation |
| `syscall` | Process management, Unix signals |
| `time` | Timestamps, timeouts, duration formatting |
| `testing` | Unit and integration tests |

> **Exception**: `gopkg.in/yaml.v3` is the single allowed external dependency for YAML parsing. The standard library has no YAML support.

### Shell Scripts

**Purpose**: Bootstrap installation only.

**Files**:
- `install.sh` — curl-based installer, sets up workspace and PATH
- `scripts/` — Dev helper scripts (build, test, release)

---

## Managed Services

Chauffeur compiles or downloads these services into the workspace — they are not system packages.

### PHP 7.4 — 8.4

**Source**: Official tarballs from `php.net`
**Compile target**: `~/.chauffeur/php/<version>/`

**Extensions compiled in**:
- Database: `mysqli`, `pdo_mysql`, `mysqlnd`
- Images: `gd`, `freetype`, `jpeg`, `png`
- Archives: `zip`, `bz2`, `zlib`
- Web: `curl`, `openssl`, `libxml`, `libxslt`
- Math: `gmp`, `bcmath`
- Crypto: `sodium`
- CLI: `readline`
- PECL: `imagick` (from PECL, latest stable tarball)

**Legacy notes**:
- PHP 7.4, 8.0: Require additional patching for GD extension (see `installers/php_legacy.go`)
- PHP 7.4, 8.0: LibXML ≤ 2.11 required (2.12+ has breaking changes)
- PHP 7.4, 8.0: ImageMagick 6.9–7.0 (7.1+ may have issues)

### Nginx

**Source**: Official tarballs from `nginx.org`
**Compile target**: `~/.chauffeur/nginx/`

**Modules enabled**:
- `http_ssl_module` — SSL/TLS
- `http_gzip_static_module` — Gzip compression
- `http_v2_module` — HTTP/2
- `http_fastcgi_module` — PHP-FPM integration

**Default ports**: HTTP 8080, HTTPS 8443

### Composer

**Source**: PHAR from `getcomposer.org`
**Install target**: `~/.chauffeur/composer/composer.phar`
**Shim**: `~/.chauffeur/bin/shims/composer`

The shim passes execution to the Chauffeur PHP binary for the current project version, ensuring Composer uses project-matched PHP.

### mkcert

**Source**: GitHub releases binary
**Purpose**: Generate locally-trusted SSL certificates for `.test` domains
**Usage**: `chauf secure` invokes mkcert automatically

### dnsmasq

**Source**: Host system package manager
**Purpose**: Route `*.test` → `127.0.0.1`
**Configuration**: Chauffeur prints the config snippet; the user applies it

---

## Host Requirements

### Required (Build Tools for PHP)

| Tool | Purpose | Min Version |
|------|---------|------------|
| `git` | Source control | 2.0+ |
| `curl` | Downloads | 7.0+ |
| `tar` | Extraction | 1.0+ |
| `gcc` | C compiler | 8.0+ |
| `make` | Build automation | 4.0+ |
| `pkg-config` | Library detection | 1.0+ |
| `autoconf` | PHP build config | 2.69+ |
| `bison` | PHP parser generator | 3.0+ |
| `re2c` | PHP lexer generator | 1.0+ |

### Required Libraries (PHP Build Dependencies)

```
libzip, libjpeg, libpng, libfreetype, libxml2, libcurl,
zlib, libbz2, libxslt, libreadline, libmagickwand (ImageMagick),
libgmp, libsodium, openssl
```

**Arch Linux**:
```bash
sudo pacman -S base-devel pkgconf libzip libjpeg-turbo libpng freetype2 \
  libxml2 curl bzip2 zlib libxslt readline imagemagick gmp libsodium openssl
```

**Debian/Ubuntu**:
```bash
sudo apt-get install build-essential pkg-config autoconf bison re2c \
  libzip-dev libjpeg-dev libpng-dev libfreetype6-dev libxml2-dev \
  libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev libreadline-dev \
  libmagickwand-dev libgmp-dev libsodium-dev openssl libssl-dev
```

**Fedora/RHEL**:
```bash
sudo dnf install gcc make pkg-config autoconf bison re2c \
  libzip-devel libjpeg-devel libpng-devel freetype-devel libxml2-devel \
  libcurl-devel bzip2-devel zlib-devel libxslt-devel readline-devel \
  ImageMagick-devel gmp-devel libsodium-devel openssl-devel
```

### Optional

- `dnsmasq` — DNS for `.test` domains
- `mkcert` — Local trusted SSL
- `iptables` — Port 80/443 redirect to 8080/8443
- `systemctl` — Systemd auto-start (Phase 5)

---

## Build Tools

### Build CLI Binary

```bash
# Development build
go build -o chauf ./cmd/chauf/

# Optimized production build
go build -ldflags="-s -w" -o chauf ./cmd/chauf/

# With version info
go build -ldflags="-s -w -X main.Version=2.0.0" -o chauf ./cmd/chauf/

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o chauf-linux-amd64 ./cmd/chauf/
GOOS=linux GOARCH=arm64 go build -o chauf-linux-arm64 ./cmd/chauf/
```

### Run Tests

```bash
go test ./...                  # All tests
go test ./... -cover           # With coverage
go test ./... -race            # Race detector
go test ./internal/commands/   # Specific package
go test -run TestLink ./...    # Specific test
```

### Release

```bash
goreleaser release --clean     # Full release (requires git tag)
goreleaser build --single-target --snapshot  # Local snapshot
```

---

## Platform Support

| Platform | Support Level |
|----------|--------------|
| Arch Linux (rolling) | Primary ✅ |
| Ubuntu 22.04+ | Primary ✅ |
| Debian 12+ | Primary ✅ |
| Fedora 39+ | Secondary ✅ |
| Rocky Linux 9+ | Secondary ✅ |
| openSUSE | Future 🚧 |

| Architecture | Support Level |
|-------------|--------------|
| x86_64 (amd64) | Primary ✅ |
| ARM64 (aarch64) | Secondary ✅ |
| ARMv7 | Future 🚧 |

---

## Admin Panel (Web UI)

### Architecture

The admin panel is a React SPA served by the Go binary via `chauf serve`.

```
chauffeur-v2/
├── internal/
│   ├── panel/              # Go backend
│   │   ├── server.go       # HTTP server, routes, API handlers
│   │   ├── types.go        # Request/response types
│   │   ├── embed.go        # Static file embedding
│   │   └── static/         # Built frontend assets
│   └── panel-apps/         # React frontend source
│       └── src/
│           ├── components/ui/  # shadcn components
│           ├── lib/             # Utils, theme
│           ├── App.tsx          # Dashboard
│           └── main.tsx         # Entry point
```

### Running the Panel

```bash
chauf serve              # Start in background (default)
chauf serve -f          # Run in foreground
chauf serve --stop      # Stop running server
chauf serve --port 8080 # Custom port (default: 3000)
chauf serve --host app.test # Custom hostname (default: panel.test)
```

### Frontend Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Framework | React 19 + TypeScript | UI components |
| Routing | TanStack Router | Type-safe routing |
| Data | TanStack Query | Server state, caching |
| Styling | Tailwind CSS v4 | Utility-first CSS |
| Components | shadcn/ui (manual) | Button, Card, Skeleton |
| Theme | CSS Variables + React Context | Dark/light mode |

### Backend API

REST API at `/api/*`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/containers` | List containers |
| GET | `/api/containers/:name` | Container details |
| POST | `/api/containers/:name/start` | Start container |
| POST | `/api/containers/:name/stop` | Stop container |
| GET | `/api/containers/:name/logs` | SSE log stream |
| GET | `/api/backups` | List backups |

### Building Frontend

```bash
cd internal/panel-apps
npm install
npm run build
# Output copied to internal/panel/static/
```

---

## Versioning Policy

Chauffeur V2 follows **Semantic Versioning 2.0.0**:
- **MAJOR**: Breaking CLI changes or workspace format changes
- **MINOR**: New commands or features (backward compatible)
- **PATCH**: Bug fixes

**PHP version policy**: Add new PHP versions within 30 days of stable release. Maintain until official EOL.
