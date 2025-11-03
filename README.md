# Chauffeur

Chauffeur is a host-based CLI for managing per-project PHP development services on Linux. It installs everything into `~/.chauffeur/`, isolates runtimes (PHP-FPM, Nginx, Caddy), and keeps system packages untouched.

## Status

- ✅ **Installer & Configuration**: Smart installer with Go requirement checking, existing installation detection, and PATH management  
- ✅ **PHP Management**: `chauf php install/use/isolate` with version switching and workspace setup  
- ✅ **CLI Bootstrap**: Both repository cloning and curl-based installation methods  
- ✅ **Shell Integration**: Clean PATH management with no whitespace pollution  
- 🚧 **Service Orchestration**: `chauf start/stop` (in progress)  
- 🚧 **Project Registration**: `chauf link/links` (in progress)  
- Linux-focused workflow (Arch/Ubuntu/Debian friendly); other OS targets are not yet supported

## Requirements

- **Go 1.22 or newer** - Chauffeur CLI is built with Go and requires a compatible Go installation
- **Git** - Required for cloning the repository during curl installation
- **Linux** - Currently Linux-focused (Arch/Ubuntu/Debian friendly)

## Getting Started

### Method 1: Clone Repository (Recommended for Development)

```bash
# First, ensure you have Go 1.22+ installed
go version

git clone https://github.com/SIAJI-Labs/chauffeur.git
cd chauffeur
./install.sh
# Reload your shell or run: source ~/.zshrc
chauf --version
```

### Method 2: Direct Install (Public Repository)

For quick installation without cloning the repository:

```bash
# First, ensure you have Go 1.22+ and Git installed
go version
git --version

curl -fsSL https://raw.githubusercontent.com/SIAJI-Labs/chauffeur/refs/heads/main/install.sh | bash
# Reload your shell or run: source ~/.zshrc
chauf --version
```

### Installer Features

- **Smart Installation Detection**: Detects existing Chauffeur installations and provides guidance
- **Go Requirement Checking**: Validates Go 1.22+ availability with clear installation instructions
- **Clean PATH Management**: Automatically manages shell PATH without creating whitespace pollution
- **Idempotent Installation**: Safe to run multiple times; handles upgrades gracefully
- **Multiple Shell Support**: Works with Bash and Zsh with proper rc file handling

### PHP Management

Once installed, you can manage PHP versions:

```bash
# Install PHP versions
chauf php install 8.3
chauf php install 8.2
chauf php install 7.4

# Switch between versions
chauf php use 8.3
chauf php use 7.4

# Check current version
chauf php -v

# Per-project isolation (planned)
chauf php isolate 8.2
```

### Uninstallation

If you need to remove the workspace:

```bash
chauf uninstall          # keeps PHP runtimes by default
chauf uninstall --purge  # removes workspace and runtimes/caches
```

The uninstaller cleanly removes PATH entries without leaving whitespace pollution in your shell config files.

## Roadmap

### Completed ✅
- Smart installer with Go requirement checking
- Existing installation detection and guidance
- Clean PATH management without whitespace pollution  
- Support for both curl and repository cloning installations
- PHP runtime management (`chauf php install/use/isolate`)
- Config management with automatic file creation
- Shell integration (Bash/Zsh) with clean PATH handling

### Current Focus 🎯
**Priority 1: Per-Project PHP Isolation**
- `chauf php isolate <version>` to pin specific PHP version to project directory
- Project configuration (`.chauffeur/project.yaml`) with per-project PHP settings
- Automatic detection of project-specific PHP requirements

**Priority 2: Project Linking & Service Registration**
- `chauf link --site <domain> --ssl --php <version>` to register projects
- `chauf links` to list all registered projects and their configurations
- Integration with Nginx and Caddy for automatic service discovery

**Priority 3: Site Accessibility**
- Automatic Nginx virtual host configuration for registered domains
- Caddy integration for local domain resolution (no `/etc/hosts` editing)
- SSL certificate management for local development domains
- Service health checks and startup coordination

### In Progress 🚧
- Service orchestration (`chauf start`, `chauf stop`) for Nginx, PHP-FPM, and Caddy
- Automated shim generation and log handling
- Project registration system foundation

### Planned 📋
- `chauf init` explicit workspace initialization
- Enhanced service configuration management
- Log rotation and management utilities
- Performance monitoring and health checks

See [TODO_STATUS.md](docs/TODO_STATUS.md) for comprehensive project status and roadmap details.

## Development Notes

- Requires Go 1.22+ to build the CLI (enforced by installer with helpful error messages).
- Installation scripts are idempotent; safe to run multiple times.
- Clean PATH management prevents shell config pollution.
- Changes should respect the contracts outlined in `AGENTS.md` (project knowledge base).
- Supports both development (repo clone) and production (curl) installation workflows.
