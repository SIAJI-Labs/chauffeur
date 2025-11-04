# Chauffeur TODO Status

This document tracks the current status of features and improvements for the Chauffeur project.

## ✅ Completed Features

### Installation & Setup
- [x] **Smart Installer** with Go requirement checking
- [x] **Multiple Installation Methods**: Repository cloning and curl-based installation
- [x] **Existing Installation Detection**: Detects and guides for existing Chauffeur installations
- [x] **Shell Integration**: Automatic PATH management for Bash and Zsh
- [x] **Clean PATH Management**: No whitespace pollution in shell config files
- [x] **Idempotent Installation**: Safe to run multiple times

### PHP Management
- [x] **PHP Installation**: `chauf php install <version>` for building from source
- [x] **PHP Version Switching**: `chauf php use <version>` for global default switching
- [x] **PHP Version Detection**: Smart detection of installed versions
- [x] **Configuration Management**: Automatic config file creation and updates
- [x] **Legacy PHP Support**: PHP 7.4 with OpenSSL 1.1.1w vendor
- [x] **Per-Project Override**: `chauf php isolate <version>` command for project-local defaults

### CLI Infrastructure
- [x] **Command Structure**: Modular Go packages in `cli/commands/*`
- [x] **Version Command**: `chauf --version` and `chauf php -v`
- [x] **Uninstall Command**: `chauf uninstall [--purge]` with clean PATH removal
- [x] **Error Handling**: Graceful error messages and guidance
- [x] **Logging System**: Structured logging for installer operations
- [x] **Enhanced CLI Logging**: Standardized logging specification with color-coded output, progress bars, spinners, and detailed timing information (based on self-update patterns)
- [x] **Testing Standards**: Comprehensive testing framework with operation-based structure, 80% coverage requirements, and integration test patterns
- [x] **Self-Update**: `chauf self-update` pulls latest git changes and rebuilds the CLI binary  
- [x] **Dev Mode**: `chauf self-update --dev` rebuilds from current directory when it's a valid chauffeur repository

### Project Registration Foundation
- [x] **Complete Project Registration System**: `chauf link`, `chauf links`, and `chauf unlink` commands fully implemented
- [x] **Project Configuration Writer**: `chauf link` generates `.chauffeur/projects/<slug>/project.yaml` with per-project PHP metadata
- [x] **Project Layout Scaffolding**: Runtime socket and log directories created alongside project configuration
- [x] **Test Coverage**: Comprehensive tests for `chauf link`, `chauf links`, and `chauf unlink` commands  
- [x] **Formatted Project Listing**: `chauf links` displays projects in formatted table with domains, SSL status, PHP versions, and creation timestamps
- [x] **Intuitive Unlink**: `chauf unlink` defaults to current directory with confirmation
- [x] **PHP Validation**: Link command validates specified PHP version is installed
- [x] **Project Removal**: `chauf unlink` command with smart defaults (current directory when no flags) and multiple ways to remove projects (by slug, domain, path, or all) with proper confirmation

## 🎯 Current Focus Areas

### Priority 1: Complete Service Orchestration
- [ ] **Service Lifecycle**: Complete `chauf start/stop/restart` commands for service management
- [ ] **Process Management**: Service monitoring and automatic restarts (Nginx, PHP-FPM, Caddy)
- [ ] **Service Templates**: Generate nginx/caddy configs for linked projects
- [ ] **Health Monitoring**: Service status tracking and startup coordination

### Priority 2: Site Accessibility Implementation
- [ ] **Nginx Virtual Hosts**: Automatic configuration for registered project domains
- [ ] **Caddy Integration**: Local domain resolution (no `/etc/hosts` editing required)
- [ ] **SSL Certificate Management**: Local SSL setup for development domains
- [ ] **Domain Routing**: Ensure projects are accessible via configured domains

### Priority 3: Enhanced Logging Implementation
- [ ] **Logging Package**: Implement the `cli/internal/logging` package based on standardized specification
- [ ] **Command Refactoring**: Update existing commands to use the new logging framework
- [ ] **Progress Enhancement**: Implement download progress bars for all remote operations
- [ ] **Terminal Detection**: Ensure proper fallback for non-interactive environments

## 🚧 In Progress (Supporting Work)

### Service Orchestration  
- [ ] **Service Lifecycle**: `chauf start/stop/restart` commands for service management (in progress)
- [ ] **Process Management**: Service monitoring and automatic restarts  
- [ ] **Configuration Templates**: Generate nginx/caddy configs for linked projects
- [ ] **Automated Shims**: Path management for installed services

### Foundation Components
- [x] **Testing Framework**: Comprehensive testing structure with operation-based organization
- [ ] **Error Handling**: Robust error reporting and recovery mechanisms
- [ ] **Test Coverage**: Achieving 80% coverage across all packages as per new standards
- [ ] **Configuration Validation**: Validate project configs and service settings

## 📋 Planned Features

### Advanced PHP Features
- [ ] **Extension Management**: Install PHP extensions during build
- [ ] **PHP Configuration**: Custom php.ini per project
- [ ] **Multiple PHP Versions**: Run multiple versions simultaneously
- [ ] **Performance Tuning**: Optimized PHP-FPM pool settings

### Service Configuration
- [ ] **Custom Nginx Configs**: User-provided nginx.conf templates
- [ ] **SSL Certificate Management**: Automatic local SSL setup
- [ ] **Reverse Proxy Configuration**: Advanced routing rules
- [ ] **Service Dependencies**: Start services in correct order

### Advanced PHP Features
- [ ] **Extension Management**: Install PHP extensions during build
- [ ] **PHP Configuration**: Custom php.ini per project
- [ ] **Multiple PHP Versions**: Run multiple versions simultaneously
- [ ] **Performance Tuning**: Optimized PHP-FPM pool settings

### Developer Experience
- [ ] **Health Checks**: Service status monitoring
- [ ] **Log Management**: Centralized log viewing and rotation
- [ ] **Performance Metrics**: Resource usage monitoring
- [ ] **Debug Tools**: Development utilities and diagnostics

### Distribution & Packaging
- [ ] **Pre-built Binaries**: Reduce build requirements for end users
- [ ] **Package Managers**: Support for system package managers
- [ ] **CI/CD Pipeline**: Automated testing and releases
- [ ] **Documentation**: Comprehensive user and developer guides

## Technical Debt & Improvements

### Code Quality
- [ ] **Unit Tests**: Comprehensive test suite for all commands
- [ ] **Integration Tests**: End-to-end testing of installation flows
- [ ] **Code Coverage**: Aim for high test coverage
- [ ] **Linting**: Consistent code style and formatting

### Architecture
- [ ] **Plugin System**: Extensible architecture for future features
- [ ] **Configuration Schema**: Formal configuration validation
- [ ] **Error Handling**: More granular error types and recovery
- [ ] **Performance**: Optimized startup and operation times

### Documentation
- [ ] **API Documentation**: Complete API reference
- [ ] **User Guides**: Step-by-step tutorials
- [ ] **Contributing Guide**: Development setup and contribution process
- [ ] **Architecture Docs**: System design and decision records

## Priority Queue

### High Priority (Next Sprint - Project Focus)
1. **Priority 1**: Auto-detect project PHP requirements and suggest isolation versions
2. **Priority 1**: Validate isolated PHP versions against installed runtimes and report gaps
3. **Priority 2**: Extend `chauf link` to generate PHP-FPM/Nginx/Caddy assets
4. **Priority 2**: Add `chauf links` command to list registered projects
5. **Logging Enhancement**: Implement the standardized logging package and refactor key commands
6. **Testing Framework**: Reorganize test structure and achieve 80% coverage per new standards

### Medium Priority (Following Sprint - Access Focus)
1. **Priority 3**: Add Nginx virtual host generation for projects
2. **Priority 3**: Implement Caddy integration for local domain resolution
3. **Priority 3**: Create SSL certificate management for local domains
4. Add service health checks and startup coordination

### Low Priority (Infrastructure & Polish)
1. Complete service orchestration framework (`chauf start/stop`)
2. Create comprehensive test suite for new features
3. Add service health monitoring and log management
4. Implement automated shims for all managed binaries
5. Add PHP extension management capabilities
6. Pre-built binary distribution strategy
7. Update `chauf self-update` defaults to public HTTPS remote and release branch once repository is published

## blockers & Dependencies

### External Dependencies
- **Go 1.22+**: Required for CLI compilation
- **Git**: Required for curl installation method
- **System build tools**: gcc, make, etc. for PHP compilation

### Technical blockers
- **Service Integration Complexity**: Coordinating multiple services (Nginx, PHP-FPM, Caddy)
- **Permission Management**: Non-root user constraints for system services
- **Port Conflicts**: Avoiding conflicts with existing services
- **Configuration Drift**: Managing configuration updates over time

## Release Notes

### v0.1.0 (Current)
- ✅ Installer with Go requirement checking
- ✅ Installation detection and guidance
- ✅ Clean PATH management
- ✅ PHP version management
- ✅ Multiple installation methods
- ✅ Enhanced CLI logging specification with visual feedback standards
- ✅ Comprehensive testing framework standards with 80% coverage requirements

### v0.1.1 (Planned - Project Focus)
- ✅ **Priority 1**: Per-project PHP isolation (`chauf php isolate`)
- ✅ **Priority 1**: CLI self-update command (`chauf self-update`)
- ✅ **Priority 2**: Complete project linking system (`chauf link`/`chauf links`/`chauf unlink`)
- ✅ **Testing Standards**: Comprehensive testing framework with operation-based structure and 80% coverage requirements
- 🎯 **Priority 1**: Service orchestration implementation (`chauf start`/`chauf stop`)
- 🎯 **Priority 3**: Enhanced logging framework implementation
- 🚧 Service orchestration foundation (in progress)

### v0.1.2 (Planned - Access Focus)
- 🎯 **Priority 2**: Site accessibility with Nginx & Caddy
- 🔗 Automatic domain resolution and SSL management
- 🚧 Service orchestration completion (start/stop/health)

### v0.2.0 (Future)
- 📋 Complete service management and monitoring
- 📋 Advanced project features and multi-project support
- 📋 Performance monitoring and optimization tools

---

*Last Updated: 2025-11-04*
*Status reflects current development progress and priorities with completed project registration system and testing framework*
