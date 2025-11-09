# Chauffeur

Chauffeur is a host-based CLI for managing per-project PHP development services on Linux. It installs everything into `~/.chauffeur/`, isolates runtimes (PHP-FPM, Nginx, Caddy), and keeps system packages untouched.

## Status

- ✅ **Installer & Configuration**: Smart installer with Go requirement checking, existing installation detection, and PATH management  
- ✅ **PHP Management**: `chauf php install/use/isolate` with version switching and workspace setup
- ✅ **Project-Aware PHP Shims**: Automatic PHP version detection based on project context  
- ✅ **CLI Bootstrap**: Both repository cloning and curl-based installation methods
- ✅ **Safe Caddy Removal**: dnsmasq validation with double confirmation to prevent system damage  
- ✅ **Self-Update**: `chauf self-update` pulls latest git changes and rebuilds the CLI binary (services untouched)  
- ✅ **Dev Mode**: `chauf self-update --dev` rebuilds CLI from current directory for development  
- ✅ **Shell Integration**: Clean PATH management with no whitespace pollution  
- ✅ **Enhanced Logging**: Structured CLI output with color-coded status, progress indicators, and detailed timing information - fully implemented and tested  
- ✅ **Project Registration**: Complete `chauf link`, `chauf links`, and `chauf unlink` commands with comprehensive project management
- ✅ **Composer Integration**: `chauf install composer` installs a verified Composer PHAR with project-aware PHP shims
- ✅ **Port Management**: Workspace defaults live on user-space ports with automatic conflict detection and per-project overrides
- ✅ **Smart PHP Defaults**: When the config points to a missing PHP runtime, Chauffeur detects installed versions and offers to switch the default automatically
- ✅ **Nginx Template System**: Automatic nginx configuration generation with Laravel, WordPress, and general templates  
- ✅ **Template Detection**: Smart project type detection for optimal nginx configuration  
- ✅ **Template Updates**: Automatic nginx config updates on PHP version changes  
- ✅ **DNS Management**: dnsmasq integration for local .test domain resolution with configuration validation
- ✅ **Safe Removal**: Comprehensive service removal with dnsmasq safety checks and user confirmation  
- ✅ **Comprehensive Testing**: Full test suite with operation-based structure and 80% coverage standards  
- ✅ **Service Orchestration**: `chauf start/stop` with dnsmasq validation  
- ✅ **Environment Insights**: `chauf info` surfaces workspace paths, versions, installed services, and port configuration at a glance  
- ✅ **Site Accessibility**: Caddy integration for local domain routing with port 80/8080 access
- 🚧 **Port Forwarding Automation**: Port 80 to 8080 redirection setup (manual implementation complete)
- Linux-focused workflow (Arch/Ubuntu/Debian friendly); other OS targets are not yet supported

## Background & Inspiration

Chauffeur emerged from a practical need for reliable development environment management on Linux. The project's journey began when I switched from macOS to Linux and found myself missing the elegant development environment tools that made other platforms so productive.

### Previous Experience
- **macOS with Valet**: Valet's seamless PHP version management and virtual host configuration made local development effortless
- **Windows & macOS with Herd**: Herd provided similar convenience for both Windows and macOS development environments
- **Linux Gap**: Upon moving to Linux, I discovered there was no equivalent tool with the same simplicity and reliability

### The Linux Challenge
While Linux offers powerful development capabilities, the ecosystem lacked a Valet/Herd equivalent:
- Existing Linux solutions were often complex or required manual configuration
- Most tools focused on Docker containers or virtualization, adding unnecessary overhead
- Forked ports of Valet existed but weren't actively maintained, creating reliability concerns

### Project Philosophy
Chauffeur was created to fill this gap with a principled approach:
- **Host-based**: Runs services directly on the host system without virtualization overhead
- **Project-isolated**: Each project gets its own isolated environment without conflicts
- **Simple yet Powerful**: Complex functionality hidden behind straightforward commands
- **Reliable**: Built with robust error handling and clear failure messages
- **Host Configuration Minimization**: Isolated environments minimize impact on host system configuration - configs and services are focused in project-specific directories to prevent conflicts with other packages and system-wide installations

## Development Approach

This project is unique in its development methodology. The majority of the codebase was written with assistance from AI tools, guided by architectural direction and project management. This approach allows for rapid development while maintaining high code quality through clear specifications and requirements.

### AI-Assisted Development
- **AI Coding**: AI tools generate the bulk of the implementation based on detailed specifications
- **Human Direction**: I manage the architecture, design decisions, and overall project vision
- **Iterative Refinement**: Code is reviewed, tested, and refined through collaboration

### Why This Approach?
The AI-assisted methodology enables:
- **Rapid Prototyping**: Quickly test architectural ideas and implementation approaches
- **Consistent Code Quality**: AI tools generate code that follows established patterns and requirements
- **Focus on Architecture**: More time spent on design decisions and less on boilerplate implementation
- **Documentation-First**: Comprehensive documentation written alongside code development

### AI Collaboration Model
The AI-assisted development approach is open to all contributors, not just for CLI development beginners. This methodology allows developers with varying experience levels to contribute meaningfully:

- **For All Contributors**: Anyone can use AI tools to help generate code while following the `AGENTS.md` specifications and contracts
- **Structured Process**: AI generates implementation based on detailed requirements and architectural direction
- **Quality Assurance**: All AI-generated code should be reviewed, tested, and refined before integration
- **Learning Opportunity**: Contributors gain experience by reviewing AI output and understanding system architecture

### Contribution Guidelines
Chauffeur is open source and welcomes diverse contribution approaches:

- **AI-Assisted Development**: Follow the same methodology used for the core codebase - generate with AI while managing architecture
- **Traditional Development**: Experienced CLI developers are highly valued for architectural insights and code reviews
- **Mixed Approach**: Use AI where helpful, but focus on producing high-quality, well-tested code
- **AGENTS.md Compliance**: All contributions must respect the contracts and specifications in the project knowledge base

### Project Structure
Given the unique development approach, the project is structured to succeed with various collaboration models:
- **Solid Foundation**: AI-generated core functionality provides working CLI tools and installation system
- **Extensible Architecture**: Well-defined contracts and modular structure support future enhancements
- **Community Integration**: Open source model allows collective improvement and innovation
- **Forking Option**: Community members can propose alternative approaches through forks if desired

**Note**: While the AI-assisted nature means the project may not be perfect, it provides a solid starting point that the community can collectively improve. The combination of AI generation and human management creates a unique opportunity for rapid development while maintaining architectural integrity.

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
- **Enhanced Logging**: Structured output with progress indicators, timing, and visual feedback

### Service Management

Once installed, you can manage services:

```bash
# Install services
chauf install php 8.3
chauf install nginx
chauf install caddy
chauf install composer

# Remove services
chauf remove php 8.3        # Remove specific PHP version
chauf remove php           # Remove all PHP versions
chauf remove nginx         # Remove nginx
chauf remove caddy         # Remove caddy (with dnsmasq validation)
chauf remove composer      # Remove Composer shim + PHAR

# Remove without confirmation
chauf remove nginx --force

# Inspect current installation
chauf info
```

#### dnsmasq Integration for Local DNS Resolution

Chauffeur requires dnsmasq configuration to resolve local `.test` domains:

```bash
# Manual dnsmasq setup (if not using Chauffeur's automated prompts)
sudo install -d -m 755 /etc/dnsmasq.d
sudo tee /etc/dnsmasq.d/chauffeur.conf >/dev/null <<'EOF'
# Chauffeur local development resolver
# Redirect all *.test domains to localhost
address=/.test/127.0.0.1

# Only listen locally
listen-address=127.0.0.1
bind-interfaces
EOF

sudo systemctl restart dnsmasq
```

**Automated Setup Features:**
- **Configuration Validation**: Checks for dnsmasq setup during `chauf start` and `chauf install caddy`
- **Interactive Prompts**: Offers to automatically install dnsmasq configuration when missing
- **Service Restart**: Automatically restarts dnsmasq to apply configuration changes
- **Domain Resolution**: Ensures `.test` domains resolve to localhost for local development

### Site Accessibility & Domain Routing

Once you have services installed and dnsmasq configured, you can access your projects:

```bash
# Link your project (automatically assigns <project-name>.test domain)
cd /path/to/my-laravel-project
chauf link

# With custom domain
chauf link --site myapp.test --ssl

# Override user-space Caddy ports if needed
chauf link --caddy-http-port 8090 --caddy-https-port 8543

# Start services
chauf start

# Access your site
curl http://my-project.test  # Works via port 80
curl http://my-project.test:8080  # Direct access to Caddy
```

Chauffeur validates every configured port during linking. The default `~/.chauffeur/config/chauffeur.yaml` created by `chauf init` pins user-space ports (Caddy 8080/8443, Nginx 8081/8444, PHP-FPM 9000) and defines a conflict-resolution strategy (`prompt`, `auto`, or `fail`). You can override ports per project with `--caddy-http-port` / `--caddy-https-port`, and the CLI will either prompt for alternatives, auto-select the next free port inside the configured range, or stop with an actionable error based on that strategy.

#### Site Access Features

- **Standard Port Access**: `http://project-name.test` (port 80) - works automatically
- **Direct Port Access**: `http://project-name.test:8080` (port 8080) - for debugging
- **DNS Resolution**: All `.test` domains resolve to `127.0.0.1` via NetworkManager dnsmasq
- **Port Forwarding**: Port 80 traffic is transparently redirected to port 8080
- **PHP Processing**: Full PHP-FPM integration with per-project isolation
- **Static Files**: Automatic static file serving with proper headers

#### Port Forwarding Setup

For the best user experience, Chauffeur includes port forwarding from port 80 to 8080:

```bash
# Manual setup (if not automated in your installation)
sudo iptables -t nat -A OUTPUT -p tcp --dport 80 -d 127.0.0.1 -j REDIRECT --to-port 8080
```

This allows you to access sites using natural URLs without port specifications while keeping Caddy running on the non-privileged port 8080 for security and compatibility.

#### Safe Caddy Removal with dnsmasq Validation

When removing Caddy, Chauffeur includes safety validation for `dnsmasq`:

```bash
chauf remove caddy           # Interactive removal with dnsmasq warnings
chauf remove caddy --force    # Remove caddy only (without touching dnsmasq)
```

**Safety Features:**
- **dnsmasq Detection**: Automatically checks if dnsmasq is installed on the system
- **Risk Warning**: Warns users that removing dnsmasq may break other applications
- **Double Confirmation**: Requires typing "REMOVE" to confirm dnsmasq deletion
- **Streamlined Flow**: After initial caddy confirmation, goes directly to dnsmasq prompt without redundant intermediate step
- **Safe Default**: `--force` flag only removes Caddy, never touches system packages
- **User Choice**: Users can keep dnsmasq while removing Caddy
- **Configuration Cleanup**: Offers to remove chauffeur dnsmasq configuration file

### Inspecting Your Chauffeur Environment

Use `chauf info` whenever you need a quick snapshot of your setup:

```bash
chauf info
```

The command reports:

- Workspace location, projects directory, and config file path
- Current CLI version alongside the latest GitHub release (with update hints)
- Installed services (Caddy, Nginx, Composer) and their reported versions/paths
- Installed PHP runtimes and the active default version
- Port configuration (Caddy/Nginx bindings, port range, PHP-FPM fallback)

This makes it easy to confirm whether a machine is up to date before sharing instructions or reproducing an issue.
```

### PHP Management

PHP versions have additional management options:

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
php -v  # Now also project-aware!

# Per-project isolation
cd my-project
chauf php isolate 8.2
php -v  # Uses PHP 8.2 for this project only
cd ~
php -v  # Uses globalPHP version
```

#### Project-Aware PHP Shims

Chauffeur's PHP shims now automatically detect project context:

- **Inside Project**: Uses the PHP version specified in the project's isolation setting
- **Outside Project**: Uses the global default PHP version set via `chauf php use <version>`
- **Seamless Integration**: Both `php` and `chauf php` commands now behave consistently
- **Automatic Detection**: Projects detected via the projects registry directory structure

### Updating Chauffeur

Refresh the CLI in-place by pulling the latest git changes and rebuilding (managed services stay intact):

```bash
# Update from remote repository
chauf self-update

# Or rebuild from current directory (development mode)
chauf self-update --dev
```

The command clones (or updates) the Chauffeur repository under `~/.chauffeur/src/chauffeur`, verifies the tree is clean, fast-forwards to the latest `main` commit, and runs `go build` to replace `~/.chauffeur/bin/chauf`. It uses the SSH remote `git@github.com:SIAJI-Labs/chauffeur.git` by default—make sure your SSH key has access (override via `CHAUF_REPO_URL` when the project becomes public). You’ll need both `git` and `go` available in your PATH.

With the `--dev` flag, the command rebuilds the CLI binary from the current working directory if it's a valid Chauffeur repository. The directory must be a git repository containing the required files (`cli/main.go`, `go.mod`, and `AGENTS.md`), providing a convenient way to test local changes without committing them to the repository.

**Visual Feedback**: All commands now feature enhanced logging with color-coded output, progress bars for downloads, spinners for long operations, and structured summaries with timing information. The logging system provides clear visual feedback about operation progress and completion.

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
- Project configuration writer via `chauf link` (slug creation, per-project PHP metadata)
- Per-project PHP overrides via `chauf php isolate`
- CLI binary refresh via `chauf self-update`  
- Development mode rebuild via `chauf self-update --dev` for testing local changes
- Complete project registration system with `chauf links` and `chauf unlink` commands
- Project-aware PHP shims that automatically detect and use appropriate PHP versions
- **Nginx Template System**: Automatic configuration generation for Laravel, WordPress, and general PHP applications
- **Template Detection**: Smart project type identification for optimal nginx configurations
- **Template Updates**: Automatic nginx config regeneration on PHP version changes
- Comprehensive testing framework with operation-based structure
- Enhanced CLI logging specification with visual feedback standards

### Current Focus 🎯
**Priority 1: Complete Service Orchestration**
- ✅ `chauf start` and `chauf stop` commands for managing services
- ✅ Integration with Nginx, PHP-FPM, and Caddy processes
- ✅ Service process monitoring and health checks

**Priority 2: Site Accessibility Implementation**
- ✅ Caddy integration for local domain resolution with port 80/8080 access
- ✅ Nginx virtual host configuration for registered domains
- ✅ DNS resolution via NetworkManager dnsmasq for `.test` domains
- 🚧 Port forwarding automation (port 80 to 8080 redirection)

**Priority 4: Log File Management**
- Implement detailed failure logging with structured log files
- Add log rotation and maintenance utilities
- Enhanced error reporting with log file paths on failures

### In Progress 🚧
- Service orchestration (`chauf start`, `chauf stop`) for Nginx, PHP-FPM, and Caddy
- Caddy integration for local domain resolution (no `/etc/hosts` editing)
- Service process monitoring and health checks

### Planned 📋
- Enhanced service configuration management
- Log rotation and management utilities
- Performance monitoring and health checks

See [TODO_STATUS.md](docs/TODO_STATUS.md) for comprehensive project status and roadmap details.

## Project Registration Flow

- Run `chauf link` inside a project directory to automatically assign `<project-name>.test` as domain, or use `chauf link --site myproj.test --ssl` for custom domains. Both create `~/.chauffeur/projects/<slug>/project.yaml`.
- The command records the absolute project path, PHP version (defaults to global), optional domain metadata, and prepares runtime/log directories.
- **Automatic Nginx Templates**: Chauffeur automatically generates nginx configurations based on project type:
  - **Laravel**: Detects `artisan` and `composer.json` with Laravel structure for optimized routing
  - **WordPress**: Detects `wp-config.php`, `wp-admin`, and `wp-includes` for WordPress-specific settings
  - **General**: Fallback for standard PHP applications with security headers and proper PHP-FPM integration
- **SSL Support**: When `--ssl` is provided with `--site`, templates include HTTPS configuration with internal TLS on port 8443
- **User-Space Ports**: All configurations use non-privileged ports (HTTP: 8080, HTTPS: 8443) to avoid conflicts with system services
- Use `--php <version>` to pin a per-project PHP version without touching global defaults.
- Use `--caddy-http-port` and `--caddy-https-port` to override Caddy listener ports per project; Chauffeur validates ports according to the configured conflict-resolution strategy before writing project metadata.
- Run `chauf php isolate <version>` in the project directory to switch the linked PHP runtime (requires the version to be installed). This automatically updates the nginx configuration to use the new PHP-FPM socket.
- Re-run with `--force` when intentionally overwriting an existing project registration.
- Run `chauf unlink` to remove a project registration (defaults to current directory when no flags provided). This also removes the associated nginx configuration.
- Run `chauf links` to list all registered projects and their configurations in a formatted table with domains, SSL status, PHP versions, and creation timestamps.
- Nginx configurations are stored in `~/.chauffeur/nginx/sites-available/` and `~/.chauffeur/nginx/sites-enabled/` with proper symlinks for easy management.

## Development Notes

- Requires Go 1.22+ to build the CLI (enforced by installer with helpful error messages).
- Installation scripts are idempotent; safe to run multiple times.
- Clean PATH management prevents shell config pollution.
- Changes should respect the contracts outlined in `AGENTS.md` (project knowledge base).
- Supports both development (repo clone) and production (curl) installation workflows.
- **Testing Standards**: Comprehensive test suite with operation-based structure, 80% coverage requirement, and integration testing for complete workflows.
