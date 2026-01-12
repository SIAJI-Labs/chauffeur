# Auto-Start Service Integration Design

## Overview

This design document outlines the implementation of auto-start functionality for Chauffeur services, allowing them to automatically start when the machine boots up.

## Design Goals

1. **Non-intrusive**: Don't require root privileges for basic functionality
2. **User-controlled**: Allow users to enable/disable auto-start per service
3. **System integration**: Work with modern Linux init systems (systemd)
4. **Workspace compliance**: Follow [../README.md](../README.md) rules and workspace-first approach
5. **Graceful handling**: Proper startup and shutdown management

## Implementation Approach

### 1. User-Level Systemd Services

Create user-level systemd services that:
- Run without root privileges
- Start automatically on user login
- Stop gracefully on user logout
- Managed with `systemctl --user`

### 2. Service Templates

Generate systemd service templates for:
- **chauf-nginx.service**: Global nginx service
- **chauf-php-fpm@.service**: Template for PHP-FPM versions
- **chauf-dns.service**: DNS validation service (optional)

### 3. CLI Commands Integration

New commands and flags:
```bash
# Enable auto-start for all services
chauf start --enable-autostart

# Enable auto-start for specific services
chauf start nginx --enable-autostart
chauf start php-fpm --enable-autostart

# Disable auto-start
chauf stop --disable-autostart

# Check auto-start status
chauf status --autostart

# Existing commands with autostart integration
chauf link --autostart  # Auto-start linked project services
```

## File Structure

```
~/.chauffeur/
  systemd/
    user/
      chauf-nginx.service
      chauf-php-fpm@.service
      chauf-dns.service
    enabled/
      chauf-nginx.service -> ~/.config/systemd/user/
      chauf-php-fpm@8.3.service -> ~/.config/systemd/user/
      chauf-php-fpm@7.4.service -> ~/.config/systemd/user/
  config/
    chauffeur.yaml  # Add autostart configuration section
```

## Configuration

### chauffeur.yaml Addition
```yaml
version: 1
autostart:
  enabled: false
  services:
    nginx: false
    php_fpm: []
    dns: false
  start_delay: 5s  # Delay after user login before starting
  shutdown_timeout: 10s
```

## Implementation Details

### 1. Service Template Generation

**chauf-nginx.service**:
```ini
[Unit]
Description=Chauffeur Nginx Service
After=network.target
Wants=network.target

[Service]
Type=forking
User=%i
ExecStart=/home/user/.chauffeur/bin/chauf start nginx
ExecStop=/home/user/.chauffeur/bin/chauf stop nginx
PIDFile=/home/user/.chauffeur/nginx/logs/nginx.pid
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

**chauf-php-fpm@.service** (template):
```ini
[Unit]
Description=Chauffeur PHP-FPM %i Service
After=network.target
Wants=network.target

[Service]
Type=forking
User=%i
ExecStart=/home/user/.chauffeur/bin/chauf start php-fpm-%i
ExecStop=/home/user/.chauffeur/bin/chauf stop php-fpm-%i
PIDFile=/home/user/.chauffeur/php/%i/runtime/php-fpm/php-fpm.pid
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

### 2. CLI Implementation

#### New Functions in cli/commands/autostart.go:

```go
// Enable autostart for services
func EnableAutostart(services []string) error

// Disable autostart for services
func DisableAutostart(services []string) error

// Check autostart status
func GetAutostartStatus() (map[string]bool, error)

// Generate systemd service files
func generateServiceTemplates() error

// Install/enable systemd services
func installSystemdServices(serviceNames []string) error

// Remove/disable systemd services
func removeSystemdServices(serviceNames []string) error
```

#### Integration with existing commands:

**start.go** enhancements:
- Add `--enable-autostart` flag
- Generate and install services if flag is present
- Update configuration file

**stop.go** enhancements:
- Add `--disable-autostart` flag
- Remove/disable services if flag is present

**status.go** enhancements:
- Add `--autostart` flag to show autostart status
- Display which services are configured for autostart

### 3. Configuration Management

**config/chauffeur.yaml updates**:
```go
type AutostartConfig struct {
    Enabled          bool     `yaml:"enabled"`
    Services         Services `yaml:"services"`
    StartDelay       string   `yaml:"start_delay"`
    ShutdownTimeout  string   `yaml:"shutdown_timeout"`
}

type Services struct {
    Nginx   bool     `yaml:"nginx"`
    PhpFpm  []string `yaml:"php_fpm"`
    DNS     bool     `yaml:"dns"`
}
```

## User Experience

### Setup Commands
```bash
# Initial setup - enable autostart for all services
chauf start --enable-autostart

# Enable only specific services
chauf start nginx --enable-autostart
chauf start php-fpm --enable-autostart

# Per-project autostart (when linking)
chauf link my-project --autostart
```

### Management Commands
```bash
# Check autostart status
chauf status --autostart
# Output:
# Auto-start Status:
# nginx: enabled
# php-fpm-8.3: enabled
# php-fpm-7.4: enabled

# Disable autostart
chauf stop --disable-autostart

# Disable specific service autostart
chauf stop nginx --disable-autostart
```

### Integration with Existing Workflow
```bash
# Normal start still works
chauf start

# Autostart services start automatically on login
# User can still manually start/stop services
chauf stop nginx  # Temporarily stop
chauf start nginx  # Manually restart
```

## Implementation Phases

### Phase 1: Core Infrastructure
1. Create `cli/commands/autostart.go` with basic functions
2. Add configuration structure to `config/chauffeur.yaml`
3. Implement systemd service template generation
4. Add basic enable/disable functionality

### Phase 2: CLI Integration
1. Integrate autostart flags into `start`/`stop` commands
2. Add `--autostart` flag to `status` command
3. Update configuration management
4. Add error handling and validation

### Phase 3: Advanced Features
1. Per-project autostart with `chauf link --autostart`
2. Dependency management (nginx depends on DNS)
3. Health monitoring integration
4. Startup delay and timeout configuration

### Phase 4: Polish & Documentation
1. Comprehensive testing
2. Documentation updates
3. Error message improvements
4. Edge case handling

## Error Handling

### Common Scenarios
1. **systemd not available**: Graceful fallback to manual start only
2. **User not logged in**: Services start when user first logs in
3. **Port conflicts**: Service fails to start, logs error, retries
4. **Missing dependencies**: Clear error messages with remediation steps

### Recovery Mechanisms
1. **Service failure**: Automatic restart with backoff
2. **Configuration errors**: Disable autostart, notify user
3. **Workspace corruption**: Fall back to manual start mode

## Security Considerations

1. **User-level services**: No root privileges required
2. **Workspace isolation**: Services run within user's workspace
3. **File permissions**: Maintain existing permission model
4. **Network access**: Same network access as manual start

## Testing Strategy

### Unit Tests
- Service template generation
- Configuration parsing and validation
- Enable/disable functionality
- Error handling scenarios

### Integration Tests
- systemd service installation and removal
- Service startup and shutdown
- Configuration persistence
- Command integration

### Manual Testing
- System reboot scenarios
- User login/logout cycles
- Service failure recovery
- Multiple PHP version handling

## Backward Compatibility

- Existing commands work unchanged
- Autostart is opt-in only
- No changes to workspace layout
- Configuration additions are backward compatible

## Future Enhancements

1. **GUI integration**: Desktop notification integration
2. **Profile-based startup**: Different service sets for different contexts
3. **Health monitoring**: systemd health checks with automatic recovery
4. **Cross-platform**: Support for other init systems (if needed)

---

This design follows AGENTS.md principles:
- Workspace-first approach
- Minimal host impact (user-level services only)
- Manual project registration principles
- Linux-focused implementation
- No external dependencies beyond systemd