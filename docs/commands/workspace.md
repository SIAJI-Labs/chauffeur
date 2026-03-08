# Workspace Commands

Commands that manage the Chauffeur workspace itself.

---

## `chauf init`

**Purpose**: Initialize the Chauffeur workspace. Creates `~/.chauffeur/` with all required subdirectories, writes the default global config, generates the PHP and Composer shim scripts, and prints PATH setup instructions.

Safe to run on an existing workspace — it updates missing pieces without overwriting existing config.

**Usage**:
```
chauf init [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Overwrite existing config with defaults |
| `--quiet`, `-q` | bool | false | Suppress output; only print errors |

**Examples**:

```bash
# First-time setup
chauf init

# Reset workspace config to defaults
chauf init --force
```

**Output**:
```
Initializing workspace...

  Directories     ~/.chauffeur/bin/shims/  config/  projects/  php/
                  nginx/etc/sites-available/  nginx/etc/sites-enabled/
                  nginx/certs/  cache/  logs/  system/
  Config          ~/.chauffeur/config/chauffeur.yaml
  PHP shim        ~/.chauffeur/bin/shims/php
  Composer shim   ~/.chauffeur/bin/shims/composer

  ✓ Workspace ready

  Add to ~/.bashrc or ~/.zshrc:
    export PATH="$HOME/.chauffeur/bin/shims:$PATH"

  chauf install nginx php 8.3 composer
  chauf doctor
```

**Notes**:
- Does NOT install nginx, PHP, or Composer. Use `chauf install` for that.
- Does NOT modify shell config files. PATH setup is always manual.
- Does NOT start any services.
- Does NOT require sudo.

---

## `chauf info`

**Purpose**: Display workspace status — installed services, registered projects, PHP versions, and configuration. Does not change any state.

**Usage**:
```
chauf info [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--detail` | bool | false | Show full paths, config files, socket paths |

**Examples**:

```bash
# Workspace overview
chauf info

# With full paths
chauf info --detail
```

**Output**:
```
Chauffeur 2.0.0

  Workspace   ~/.chauffeur
  Config      ~/.chauffeur/config/chauffeur.yaml

  Services
    nginx         ● running    8080 / 8443
    php-fpm 8.3   ● running    shared pool
    php-fpm 8.1   ○ stopped

  Projects (3)
    my-app        my-app.test          PHP 8.3   shared     HTTP
    admin-panel   admin-panel.test     PHP 8.1   dedicated  HTTPS
    legacy-site   legacy-site.test     PHP 7.4   shared     HTTP

  PHP versions
    8.4   ~/.chauffeur/php/8.4/
    8.3   ~/.chauffeur/php/8.3/   (default)
    8.1   ~/.chauffeur/php/8.1/
    7.4   ~/.chauffeur/php/7.4/

  Cache   245 MB  (~/.chauffeur/cache/)
```

**With `--detail`**, each service and project entry expands to show binary path, config file, log file, socket path, and PID file.

**Related**: `chauf status` for service health detail, `chauf links` for project list detail.

---

## `chauf uninstall`

**Purpose**: Remove the entire Chauffeur workspace. Prompts for confirmation before deleting anything. After removal the `chauf` binary itself remains (if installed separately) but the workspace is gone.

Also prints the manual steps needed to clean up dnsmasq config, iptables rules, and systemd units — Chauffeur never removes those silently.

**Usage**:
```
chauf uninstall [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Skip the confirmation prompt |
| `--purge` | bool | false | Also remove the `chauf` binary from `~/.chauffeur/bin/` |

**Examples**:

```bash
# Interactive removal
chauf uninstall

# Skip confirmation (use in scripts)
chauf uninstall --force
```

**Output** (before confirmation):
```
Uninstalling Chauffeur...

  Workspace     ~/.chauffeur
  Services      nginx  ·  php-fpm 8.3  ·  php-fpm 8.1
  Projects      3 registered
  PHP versions  8.3  ·  8.1  ·  7.4
  Cache         245 MB

  This will permanently delete ~/.chauffeur/ and all its contents.

  Type 'yes' to confirm: _
```

**After removal**:
```
  ✓ Workspace removed

  Manual cleanup (if needed):

  dnsmasq:
    sudo rm /etc/dnsmasq.d/chauffeur.conf
    sudo systemctl restart dnsmasq

  iptables (if port forwarding was enabled):
    sudo iptables -t nat -D OUTPUT -p tcp --dport 80 -j REDIRECT --to-port 8080
    sudo iptables -t nat -D OUTPUT -p tcp --dport 443 -j REDIRECT --to-port 8443

  systemd (if auto-start was enabled):
    systemctl --user disable --now chauffeur-nginx.service
    rm ~/.config/systemd/user/chauffeur-*.service

  PATH (remove from ~/.bashrc or ~/.zshrc):
    export PATH="$HOME/.chauffeur/bin/shims:$PATH"
```

**Notes**:
- Running services are stopped first (nginx, PHP-FPM).
- dnsmasq, iptables, and systemd changes are **never** reversed automatically — commands are printed for the user.
- The `chauf` binary itself is not removed unless `--purge` is passed.
