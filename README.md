# Chauffeur

Chauffeur is a host-based CLI for managing per-project PHP development services on Linux. It installs everything into `~/.chauffeur/`, isolates runtimes (PHP-FPM, Nginx, Caddy), and keeps system packages untouched.

## Status

- ✅ **Installer & Configuration**: Smart installer with Go requirement checking, existing installation detection, and PATH management  
- ✅ **PHP Management**: `chauf php install/use/isolate` with version switching and workspace setup  
- ✅ **CLI Bootstrap**: Both repository cloning and curl-based installation methods  
- ✅ **Self-Update**: `chauf self-update` pulls latest git changes and rebuilds the CLI binary (services untouched)  
- ✅ **Dev Mode**: `chauf self-update --dev` rebuilds CLI from current directory for development  
- ✅ **Shell Integration**: Clean PATH management with no whitespace pollution  
- 🚧 **Service Orchestration**: `chauf start/stop` (in progress)  
- 🚧 **Project Registration**: Basic project config creation via `chauf link`; listing & service wiring in progress  
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

### Leadership & Project Direction
While I don't have strong background in CLI tool development, I focus on:
- **Architecture & Design**: Managing the overall system design, command contracts, and project vision
- **Requirements Management**: Defining what tools should do and how they should work
- **Integration & Validation**: Testing AI-generated code and ensuring it meets project goals
- **Community Coordination**: Facilitating collaboration and maintaining project direction

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

# Per-project isolation
chauf php isolate 8.2
```

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

### Current Focus 🎯
**Priority 1: Per-Project PHP Isolation**
- Automatic detection of project-specific PHP requirements

**Priority 2: Project Linking & Service Registration**
- `chauf link --site <domain> [--ssl] [--php <version>] [--force]` to register projects
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

## Project Registration Flow

- Run `chauf link --site myproj.test --ssl` inside a project directory to create `~/.chauffeur/projects/<slug>/project.yaml`.
- The command records the absolute project path, PHP version (defaults to global), optional domain metadata, and prepares runtime/log directories.
- Use `--php <version>` to pin a per-project PHP version without touching global defaults.
- Run `chauf php isolate <version>` in the project directory to switch the linked PHP runtime (requires the version to be installed).
- Re-run with `--force` when intentionally overwriting an existing project registration.
- Upcoming work: emit Nginx/Caddy templates and expose `chauf links` for listing registrations.

## Development Notes

- Requires Go 1.22+ to build the CLI (enforced by installer with helpful error messages).
- Installation scripts are idempotent; safe to run multiple times.
- Clean PATH management prevents shell config pollution.
- Changes should respect the contracts outlined in `AGENTS.md` (project knowledge base).
- Supports both development (repo clone) and production (curl) installation workflows.
