# Per-PHP-Version Config Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow configuring PHP runtime settings (upload_max_filesize, post_max_size, memory_limit, etc.) per PHP version in chauffeur.yaml.

**Architecture:** Extend Config struct with PHPVersionConfig, add config parsing/generation, add chauf config php subcommand, write settings to conf.d/limits.ini.

**Tech Stack:** Go (chauffeur CLI), YAML config

---

## Task 1: Read existing workspace/config.go

**Files:**
- Read: `internal/workspace/config.go`

**Step 1: Read current implementation**

Run: `cat internal/workspace/config.go`

Understand:
- Current Config struct
- How Load() parses chauffeur.yaml
- How SetDefaultPHP() works for comparison

This is a reading task - do NOT make changes.

---

## Task 2: Extend Config struct with PHPVersionConfig

**Files:**
- Modify: `internal/workspace/config.go`

**Step 1: Add PHPVersionConfig struct**

Add after the Config struct (around line 33):

```go
// PHPVersionConfig holds runtime settings for a specific PHP version.
type PHPVersionConfig struct {
    UploadMaxFilesize string
    PostMaxSize       string
    MemoryLimit       string
    MaxExecutionTime  int
    MaxInputVars      int
}
```

**Step 2: Extend Config.PHP with Versions map**

Replace the PHP struct in Config (lines 19-21):

```go
PHP struct {
    DefaultVersion string
    Versions       map[string]PHPVersionConfig
}
```

**Step 3: Add DefaultPHPVersionConfig function**

Add after DefaultConfig (around line 48):

```go
// DefaultPHPVersionConfig returns the default PHP runtime settings.
func DefaultPHPVersionConfig() PHPVersionConfig {
    return PHPVersionConfig{
        UploadMaxFilesize: "64M",
        PostMaxSize:       "64M",
        MemoryLimit:       "256M",
        MaxExecutionTime:  300,
        MaxInputVars:      5000,
    }
}
```

**Step 4: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/workspace/`

---

## Task 3: Update Load() to parse PHP versions config

**Files:**
- Modify: `internal/workspace/config.go` - the Load() function

**Step 1: Add parsing for php.versions section**

Find the switch statement in Load() (around lines 81-108) and add:

```go
case "php.default_version":
    c.PHP.DefaultVersion = val
case "php.versions":
    // Next line(s) should have version: key
```

Actually, the current config format uses YAML. We need to decide on format. Let's use a simple key notation:

```yaml
php:
  default_version: "8.3"
  version.8.3.upload_max_filesize: "64M"
  version.8.3.post_max_size: "64M"
  version.8.3.memory_limit: "256M"
  version.8.3.max_execution_time: 300
  version.8.3.max_input_vars: 5000
  version.7.4.upload_max_filesize: "32M"
  version.7.4.post_max_size: "32M"
  version.7.4.memory_limit: "128M"
  version.7.4.max_execution_time: 120
  version.7.4.max_input_vars: 1000
```

Add these cases in the switch statement:

```go
case "php.version":
    // Handled in next iteration via php.version.{ver}.{key}
case "php.version.*":
    // Parse version-specific settings - use strings.HasPrefix
```

Actually, the current simple parser doesn't support nested keys well. Let me use a simpler flat key approach:

```yaml
php:
  default_version: "8.3"
  8.3.upload_max_filesize: "64M"
  8.3.post_max_size: "64M"
  8.3.memory_limit: "256M"
  8.3.max_execution_time: 300
  8.3.max_input_vars: 5000
```

Add this case:

```go
default:
    // Handle php.{version}.{key} = value format
    if strings.HasPrefix(section, "php.") && strings.Contains(line, ".") {
        // Parse php.8.3.upload_max_filesize: "64M"
    }
```

**Implementation approach:**

Replace the Load() switch with a more flexible parser. Here's a simpler approach - add a helper to parse php version keys:

```go
// In Load(), add after the switch:
if strings.HasPrefix(section, "php.") {
    ver := strings.TrimPrefix(section, "php.")
    if _, ok := c.PHP.Versions[ver]; !ok {
        c.PHP.Versions[ver] = DefaultPHPVersionConfig()
    }
    switch key {
    case "upload_max_filesize":
        c.PHP.Versions[ver].UploadMaxFilesize = val
    case "post_max_size":
        c.PHP.Versions[ver].PostMaxSize = val
    case "memory_limit":
        c.PHP.Versions[ver].MemoryLimit = val
    case "max_execution_time":
        if v, err := strconv.Atoi(val); err == nil {
            c.PHP.Versions[ver].MaxExecutionTime = v
        }
    case "max_input_vars":
        if v, err := strconv.Atoi(val); err == nil {
            c.PHP.Versions[ver].MaxInputVars = v
        }
    }
}
```

**Step 2: Verify Go syntax**

Run: `go build ./internal/workspace/`

---

## Task 4: Add SavePHPVersionConfig function

**Files:**
- Modify: `internal/workspace/config.go`

**Step 1: Add SavePHPVersionConfig function**

Add after SetDefaultPHP() (around line 151):

```go
// SavePHPVersionConfig updates the php.ini settings for a specific version in chauffeur.yaml.
func SavePHPVersionConfig(version string, cfg PHPVersionConfig) error {
    root := Root()
    configPath := filepath.Join(root, "config", "chauffeur.yaml")

    data, err := os.ReadFile(configPath)
    if err != nil {
        return fmt.Errorf("read config: %w", err)
    }

    lines := strings.Split(string(data), "\n")
    
    // Find or create the php section
    // Update lines like "  8.3.upload_max_filesize: \"64M\""
    // or insert new ones after "php:" if not found
    
    // For simplicity, rebuild the php section
    var newLines []string
    inPHP := false
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if trimmed == "php:" {
            inPHP = true
            newLines = append(newLines, line)
            // Insert all version configs
            for ver, v := range cfg {
                newLines = append(newLines, fmt.Sprintf("  %s.upload_max_filesize: \"%s\"", ver, v.UploadMaxFilesize))
                newLines = append(newLines, fmt.Sprintf("  %s.post_max_size: \"%s\"", ver, v.PostMaxSize))
                newLines = append(newLines, fmt.Sprintf("  %s.memory_limit: \"%s\"", ver, v.MemoryLimit))
                newLines = append(newLines, fmt.Sprintf("  %s.max_execution_time: %d", ver, v.MaxExecutionTime))
                newLines = append(newLines, fmt.Sprintf("  %s.max_input_vars: %d", ver, v.MaxInputVars))
            }
            continue
        }
        if inPHP {
            // Skip old php version lines
            if strings.HasPrefix(trimmed, "default_version:") {
                newLines = append(newLines, line)
                continue
            }
            if strings.Contains(trimmed, ".upload_max_filesize:") ||
               strings.Contains(trimmed, ".post_max_size:") ||
               strings.Contains(trimmed, ".memory_limit:") ||
               strings.Contains(trimmed, ".max_execution_time:") ||
               strings.Contains(trimmed, ".max_input_vars:") {
                continue // skip old line
            }
            if strings.HasPrefix(trimmed, "dns:") || strings.HasPrefix(trimmed, "logging:") || strings.HasPrefix(trimmed, "nginx:") {
                inPHP = false
            }
        }
        newLines = append(newLines, line)
    }

    return os.WriteFile(configPath, []byte(strings.Join(newLines, "\n")), 0644)
}
```

**Note:** This function signature needs to handle updating a single version, not all versions. Let me revise:

```go
// SavePHPVersionConfig updates settings for a specific PHP version in chauffeur.yaml.
func SavePHPVersionConfig(version string, settings map[string]string) error {
    root := Root()
    configPath := filepath.Join(root, "config", "chauffeur.yaml")

    data, err := os.ReadFile(configPath)
    if err != nil {
        return fmt.Errorf("read config: %w", err)
    }

    lines := strings.Split(string(data), "\n")
    
    // For each setting, find or update the line
    for key, value := range settings {
        keyLine := fmt.Sprintf("  %s.%s:", version, key)
        valueLine := fmt.Sprintf("  %s.%s: \"%s\"", version, key, value)
        found := false
        for i, line := range lines {
            trimmed := strings.TrimSpace(line)
            if strings.HasPrefix(trimmed, keyLine) {
                lines[i] = valueLine
                found = true
                break
            }
        }
        if !found {
            // Insert after php: section
            for i, line := range lines {
                if strings.TrimSpace(line) == "php:" {
                    lines = append(lines[:i+1], append([]string{valueLine}, lines[i+1:]...)...)
                    break
                }
            }
        }
    }

    return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
}
```

**Step 2: Verify Go syntax**

Run: `go build ./internal/workspace/`

---

## Task 5: Add WritePHPConfig to PHP installer

**Files:**
- Modify: `internal/installers/php.go`

**Step 1: Add WritePHPConfig method**

Add after writePhpIni() (around line 288):

```go
// WritePHPConfig writes PHP runtime settings to conf.d/limits.ini.
func (p *PHPInstaller) WritePHPConfig(cfg workspace.PHPVersionConfig) error {
    confdDir := filepath.Join(p.InstallDir(), "etc", "conf.d")
    if err := os.MkdirAll(confdDir, 0755); err != nil {
        return err
    }
    
    iniPath := filepath.Join(confdDir, "limits.ini")
    content := fmt.Sprintf(`; Auto-generated by chauf - edit via: chauf config php <version>
upload_max_filesize = %s
post_max_size = %s
memory_limit = %s
max_execution_time = %d
max_input_vars = %d
`, cfg.UploadMaxFilesize, cfg.PostMaxSize, cfg.MemoryLimit, cfg.MaxExecutionTime, cfg.MaxInputVars)
    
    return os.WriteFile(iniPath, []byte(content), 0644)
}
```

**Step 2: Call WritePHPConfig during install**

In the PHP installer (find where writePhpIni is called), add after:

```go
// In Install() or similar
if err := p.writePhpIni(); err != nil {
    return err
}
// Add this:
if err := p.WritePHPConfig(workspace.DefaultPHPVersionConfig()); err != nil {
    return err
}
```

Find the Install method by searching for writePhpIni call.

**Step 3: Verify Go syntax**

Run: `go build ./internal/installers/`

---

## Task 6: Add config PHP subcommand

**Files:**
- Modify or Create: `internal/commands/config.go`

**Step 1: Read existing config command**

Run: `cat internal/commands/config.go 2>/dev/null || echo "File does not exist"`

If it exists, understand how RunConfig works. If not, we'll create it.

**Step 2: Create config.go with PHP config handling**

If config.go doesn't exist, create it with the full config command structure. If it exists, add PHP handling to RunConfig.

The new structure should support:

```go
func RunConfig(args []string) error {
    if len(args) == 0 {
        return configHelp()
    }
    
    switch args[0] {
    case "php":
        return configPHP(args[1:])
    // ... other config types
    default:
        return fmt.Errorf("unknown config type: %s", args[0])
    }
}

func configPHP(args []string) error {
    // chauf config php - list all versions
    // chauf config php 8.3 - show 8.3 config
    // chauf config php 8.3 upload_max_filesize 128M - set value
}
```

**Step 3: Verify Go syntax**

Run: `go build ./internal/commands/`

---

## Task 7: Build and test

**Step 1: Build entire project**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build -o chauf ./cmd/chauf/`

**Step 2: Test `chauf config php` (list all)**

Run: `./chauf config php`

Expected: Shows all PHP versions with their configs

**Step 3: Test `chauf config php 8.3` (show one)**

Run: `./chauf config php 8.3`

Expected: Shows 8.3 current config

**Step 4: Test `chauf config php 8.3 upload_max_filesize 128M`**

Run: `./chauf config php 8.3 upload_max_filesize 128M`

Expected: Updates config and confirms

**Step 5: Verify limits.ini**

Run: `cat ~/.chauffeur/php/8.3/etc/conf.d/limits.ini`

Expected: Contains new upload_max_filesize value

**Step 6: Verify PHP sees new value**

Run: `./chauf php 8.3 -i | grep upload_max_filesize`

Expected: Shows 128M

---

## Task 8: Commit

**Step 1: Stage and commit**

```bash
git add internal/workspace/config.go internal/installers/php.go internal/commands/config.go
git commit -m "feat(config): add per-PHP-version configuration support

Allows configuring PHP runtime settings (upload_max_filesize,
post_max_size, memory_limit, etc.) per PHP version in chauffeur.yaml:

  chauf config php 8.3 upload_max_filesize 128M
  chauf config php 8.3

Settings written to conf.d/limits.ini per PHP version."
```
