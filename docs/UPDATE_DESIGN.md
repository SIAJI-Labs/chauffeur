# Service Update Management Design

## Overview

This design document outlines a universal service update system for Chauffeur, enabling updates for all managed services (PHP, Composer, nginx, and future services) with proper version management, backup, and rollback capabilities.

## Design Goals

1. **Universal Interface**: Single command for updating all services
2. **Version Intelligence**: Smart version detection and comparison
3. **Safe Updates**: Backup and rollback mechanisms
4. **Service Awareness**: Different update strategies per service type
5. **Minimal Downtime**: Graceful service handling during updates
6. **User Control**: Granular control over update types and timing

## Command Interface

### Primary Update Command
```bash
# Check for available updates
chauf update --check

# Update all services to latest patch versions
chauf update --all

# Update specific service
chauf update php
chauf update composer
chauf update nginx

# Update to specific version
chauf update php 8.3.15
chauf update composer 2.7.1

# Include minor version updates (with confirmation)
chauf update --include-minor

# Dry run to see what would be updated
chauf update --dry-run

# Force update (skip compatibility checks)
chauf update php --force

# Rollback to previous version
chauf update php --rollback
```

### Service-Specific Aliases
```bash
# PHP-specific updates
chauf php update
chauf php update 8.3.15
chauf php update --check

# Composer updates
chauf composer update
chauf composer update --check

# nginx updates
chauf nginx update
chauf nginx update --check
```

## Implementation Architecture

### 1. Core Update Manager

**File**: `cli/internal/updates/manager.go`

```go
type UpdateManager struct {
    services map[string]ServiceUpdater
    config   UpdateConfig
}

type ServiceUpdater interface {
    GetCurrentVersion() (string, error)
    GetLatestVersion() (string, error)
    IsUpdateAvailable() (bool, string, string, error) // (available, current, latest, error)
    Update(targetVersion string, opts UpdateOptions) error
    Backup() error
    Rollback() error
    GetServiceType() ServiceType
}

type UpdateOptions struct {
    Force         bool
    DryRun        bool
    IncludeMinor  bool
    Backup        bool
    ServiceDowntime time.Duration
}
```

### 2. Service-Specific Updaters

#### PHP Updater (`cli/internal/updates/php.go`)
```go
type PHPUpdater struct {
    installPath string
    config     *config.Config
}

// Features:
- Patch version updates (8.3.1 → 8.3.2)
- Minor version updates (8.3 → 8.4) with confirmation
- Extension compatibility checking
- Configuration migration
- PECL extension recompilation
```

#### Composer Updater (`cli/internal/updates/composer.go`)
```go
type ComposerUpdater struct {
    binaryPath string
    config     *config.Config
}

// Features:
- PHAR replacement
- Plugin compatibility checking
- Autoloader regeneration
```

#### Nginx Updater (`cli/internal/updates/nginx.go`)
```go
type NginxUpdater struct {
    binaryPath string
    configPath string
}

// Features:
- Binary replacement
- Configuration validation
- Graceful reload
- SSL certificate migration
```

### 3. Version Detection System

**File**: `cli/internal/versions/detector.go`

```go
type VersionDetector struct {
    client *http.Client
    cache  *VersionCache
}

type VersionInfo struct {
    Current      string
    Latest       string
    LatestPatch  string
    LatestMinor  string
    ReleaseNotes string
    ReleaseDate  time.Time
    EOLDate      *time.Time
}

// Methods:
func (d *VersionDetector) CheckPHP(version string) (*VersionInfo, error)
func (d *VersionDetector) CheckComposer() (*VersionInfo, error)
func (d *VersionDetector) CheckNginx() (*VersionInfo, error)
```

### 4. Backup and Rollback System

**File**: `cli/internal/updates/backup.go`

```go
type BackupManager struct {
    backupDir string
}

type BackupInfo struct {
    Service     string
    Version     string
    BackupPath  string
    Timestamp   time.Time
    Size        int64
    Checksum    string
}

// Methods:
func (bm *BackupManager) CreateBackup(service string) (*BackupInfo, error)
func (bm *BackupManager) RestoreBackup(backupID string) error
func (bm *BackupManager) ListBackups(service string) ([]*BackupInfo, error)
func (bm *BackupManager) CleanupOldBackups(service string, keepCount int) error
```

## User Experience Flow

### 1. Update Check Flow
```bash
$ chauf update --check

Checking for service updates...

Service        Current    Latest     Type      Status
------------- ---------- ---------- --------- -------------
php           8.3.12     8.3.15     patch     ⬆️  Update available
composer      2.6.6      2.7.0      minor     ⬆️  Update available
nginx         1.24.0     1.24.0     -         ✅  Up to date

Run 'chauf update --all' to update all services, or 'chauf update <service>' for specific updates.
```

### 2. Update Execution Flow
```bash
$ chauf update php

Preparing to update php from 8.3.12 to 8.3.15...

📦 Update Summary:
  Service: php
  From: 8.3.12
  To: 8.3.15
  Type: patch update
  Estimated downtime: ~2 minutes

🔄 Steps:
  1. Create backup of current installation
  2. Download PHP 8.3.15
  3. Compile with current extensions
  4. Stop running PHP-FPM services
  5. Replace installation
  6. Restart PHP-FPM services
  7. Verify functionality

Continue? [y/N]: y

✅ Backup created: ~/.chauffeur/backups/php-8.3.12-2024-01-15-10-30
⬇️  Downloading PHP 8.3.15...
🔨 Compiling PHP 8.3.15 with extensions...
⏹️  Stopping PHP-FPM services...
🔄 Installing PHP 8.3.15...
▶️  Starting PHP-FPM services...
✅ Update completed successfully!

💡 To rollback: chauf update php --rollback
```

### 3. Rollback Flow
```bash
$ chauf update php --rollback

Available PHP backups:
  1. php-8.3.12-2024-01-15-10-30 (current version - 2 hours ago)
  2. php-8.3.11-2024-01-10-15-45 (5 days ago)

Select backup to restore [1-2]: 1

Restoring PHP 8.3.12...

⏹️  Stopping PHP-FPM services...
🔄 Restoring from backup...
▶️  Starting PHP-FPM services...
✅ Rollback to PHP 8.3.12 completed successfully!
```

## Configuration Integration

### Config Structure Addition
**File**: `cli/internal/config/config.go`

```go
type UpdateConfig struct {
    AutoCheck       bool          `yaml:"auto_check"`
    CheckInterval   time.Duration `yaml:"check_interval"`
    AutoPatch       bool          `yaml:"auto_patch"`
    BackupRetention int           `yaml:"backup_retention"`
    Notify          bool          `yaml:"notify"`
    Services        ServiceUpdateConfig `yaml:"services"`
}

type ServiceUpdateConfig struct {
    PHP      PHPUpdateConfig      `yaml:"php"`
    Composer ComposerUpdateConfig `yaml:"composer"`
    Nginx    NginxUpdateConfig    `yaml:"nginx"`
}

type PHPUpdateConfig struct {
    AutoMinor       bool     `yaml:"auto_minor"`
    Extensions      []string `yaml:"extensions"`
    RecompilePECL   bool     `yaml:"recompile_pecl"`
}
```

### Configuration File Example
```yaml
# ~/.chauffeur/config/chauffeur.yaml
version: 1
telemetry: false
workspace_dir: ~/.chauffeur

# Update configuration
updates:
  auto_check: true
  check_interval: 24h
  auto_patch: false
  backup_retention: 5
  notify: true
  services:
    php:
      auto_minor: false
      extensions: [gd, zip, exif, freetype, imagick]
      recompile_pecl: true
    composer:
      auto_update: false
    nginx:
      auto_reload: true
```

## Service-Specific Update Logic

### PHP Updates
1. **Pre-update Checks**:
   - Extension compatibility matrix
   - Current configuration validation
   - Disk space requirements

2. **Update Process**:
   - Download and verify new version
   - Compile with existing extensions
   - Migrate configuration files
   - Update symlinks and shims

3. **Post-update Validation**:
   - PHP binary functionality test
   - Extension loading verification
   - PHP-FPM service health check

### Composer Updates
1. **Simple PHAR Replacement**:
   - Download new PHAR
   - Verify checksum
   - Replace binary
   - Update version information

2. **Plugin Compatibility**:
   - Check plugin compatibility matrix
   - Update autoloader if needed
   - Verify composer.json compatibility

### Nginx Updates
1. **Binary Update**:
   - Download new nginx binary
   - Test configuration compatibility
   - Graceful restart (no downtime)

2. **Configuration Migration**:
   - Update configuration syntax if needed
   - Migrate SSL certificates
   - Validate all site configurations

## Error Handling and Recovery

### Update Failure Scenarios
1. **Download Failures**: Retry with alternative mirrors
2. **Compilation Failures**: Fall back to previous version
3. **Configuration Conflicts**: Manual intervention required
4. **Service Start Failures**: Automatic rollback

### Recovery Mechanisms
1. **Automatic Rollback**: If services fail to start
2. **Manual Rollback**: User-initiated version restoration
3. **Partial Recovery**: Restore only failed components
4. **Configuration Repair**: Fix broken configurations

## CLI Enhancement: -v Flag

### Current Commands to Enhance
```bash
# Existing version flags
chauf --version
chauf info --version
chauf php --version

# Enhanced with -v shorthand
chauf -v
chauf info -v
chauf php -v
```

### Implementation
**File**: Update all command parsers to accept `-v` as alias for `--version`:

```go
// Example for main command
case "--version", "-v":
    printVersion()
    return nil

// Example for php command
case "--version", "-v":
    return showPHPVersion()
```

## Implementation Phases

### Phase 1: Core Infrastructure
1. Create update manager framework
2. Implement version detection system
3. Add backup and rollback system
4. Basic PHP updater implementation

### Phase 2: Service Integration
1. Complete PHP updater with all features
2. Implement Composer updater
3. Add nginx updater
4. Add CLI integration

### Phase 3: Advanced Features
1. Configuration management
2. Automatic update checking
3. Update notifications
4. Minor version updates with confirmation

### Phase 4: Polish and Optimization
1. Performance optimization
2. Better error messages
3. Comprehensive testing
4. Documentation updates

## Testing Strategy

### Unit Tests
- Version detection and comparison
- Update logic for each service
- Backup and rollback functionality
- Configuration parsing

### Integration Tests
- End-to-end update workflows
- Service restart validation
- Cross-platform compatibility
- Error recovery scenarios

### Manual Testing
- Real update scenarios
- Performance impact assessment
- User experience validation
- Edge case handling

## Security Considerations

1. **Signature Verification**: Verify checksums and signatures
2. **Safe Downloads**: Use HTTPS and verify certificates
3. **Backup Security**: Encrypt sensitive backup data
4. **Permission Handling**: Maintain proper file permissions
5. **Audit Trail**: Log all update activities

## Backward Compatibility

- Existing commands work unchanged
- Update functionality is additive
- Configuration changes are backward compatible
- No changes to existing installation structure

---

This design follows AGENTS.md principles:
- Workspace-first approach (all updates within ~/.chauffeur)
- Minimal host impact (no system-wide changes)
- Manual control philosophy (updates are opt-in)
- Linux-focused implementation
- Comprehensive error handling and logging