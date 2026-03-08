# Configuration Commands *(V2)*

These commands are new in V2. They replace manual YAML editing for workspace and project configuration, and add environment variable and auto-start management.

---

## `chauf config`

**Purpose**: Read and write Chauffeur configuration without editing YAML files directly. Operates on either the global workspace config (`~/.chauffeur/config/chauffeur.yaml`) or a specific project's config.

**Usage**:
```
chauf config <subcommand> [key] [value] [flags]
```

**Subcommands**:

| Subcommand | Description |
|------------|-------------|
| `show` | Display current config as formatted output |
| `set <key> <value>` | Update a single config value |
| `validate` | Validate config against schema; report errors |
| `export` | Print config as YAML (for backup or scripting) |
| `import <file>` | Import config from a YAML file |
| `reset` | Reset config to defaults (prompts for confirmation) |

**Flags** (apply to all subcommands):

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | — | Operate on a project config instead of the workspace config |
| `--slug <slug>` | string | — | Operate on a project config by slug |
| `--format <fmt>` | string | `pretty` | Output format for `show`/`export`: `pretty`, `yaml`, `json` |

---

### `chauf config show`

```bash
# Show global workspace config
chauf config show

# Show a project's config
chauf config show --project ~/Projects/my-app
chauf config show --slug my-app

# Output as raw YAML
chauf config show --format yaml
```

**Output** (global):
```
  Workspace Configuration

  workspace:            ~/.chauffeur
  nginx.http_port:      8080
  nginx.https_port:     8443
  php.default_version:  8.3
  dns.tld:              test
  dns.enabled:          true
  logging.level:        info
  logging.max_size_mb:  10
```

**Output** (project):
```
  Project Configuration: my-app

  slug:           my-app
  path:           /home/user/Projects/my-app
  domain:         my-app.test
  aliases:        admin.my-app.test  ·  api.my-app.test
  php_version:    8.3
  ssl:            true
  fpm.dedicated:  false
  project_type:   laravel
  created_at:     2025-01-01T00:00:00Z
```

---

### `chauf config set`

**Key format**: Dot-notation matches the YAML hierarchy (e.g. `nginx.http_port`, `php.default_version`).

```bash
# Change nginx HTTP port
chauf config set nginx.http_port 8081

# Change global default PHP
chauf config set php.default_version 8.1

# Enable debug logging
chauf config set logging.level debug

# Switch a project to dedicated FPM
chauf config set fpm.dedicated true --project ~/Projects/my-app

# Update a project's PHP version (also regenerates nginx config)
chauf config set php_version 8.2 --slug my-app
```

**Output**:
```
  ✓ nginx.http_port set to 8081  (was 8080)

  nginx must restart for this to take effect:
    chauf restart nginx
```

**Validation**: Invalid values are rejected before writing:
```
  ✗ Invalid value for nginx.http_port: "abc"

  Expected: integer between 1 and 65535
```

---

### `chauf config validate`

```bash
chauf config validate
chauf config validate --slug my-app
```

**Output** (valid):
```
  ✓ Config valid  (~/.chauffeur/config/chauffeur.yaml)
```

**Output** (invalid):
```
  ✗ Config validation failed

  Errors:
    • nginx.http_port: must be an integer (got "abc")
    • php.default_version: "9.0" is not a supported PHP version

  Fix:   ~/.chauffeur/config/chauffeur.yaml
  Reset: chauf config reset
```

---

### `chauf config export`

```bash
# Export workspace config to stdout
chauf config export

# Export project config to a file
chauf config export --slug my-app > my-app-config.yaml

# Export as JSON
chauf config export --format json
```

---

### `chauf config import`

```bash
# Import workspace config from file
chauf config import ./workspace-config.yaml

# Import project config
chauf config import ./my-app-config.yaml --slug my-app
```

Validates the imported file before writing. Does not overwrite if validation fails.

---

### `chauf config reset`

```bash
# Reset workspace config to defaults
chauf config reset

# Reset a project config to defaults (keeps slug, path, domain)
chauf config reset --slug my-app
```

Prompts for confirmation before resetting.

---

## `chauf env`

**Purpose**: Manage per-project environment variables stored in the Chauffeur workspace. These are injected as `fastcgi_param` directives in the project's nginx config, making them available to PHP via `$_SERVER` and `getenv()`.

**Usage**:
```
chauf env <subcommand> [key] [value] [flags]
```

**Subcommands**:

| Subcommand | Description |
|------------|-------------|
| `list` | List all env vars for the project |
| `set <key> <value>` | Set an environment variable |
| `unset <key>` | Remove an environment variable |
| `import <file>` | Import variables from a `.env` file |
| `export` | Print variables in `.env` format |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | CWD | Target a specific project by path |
| `--slug <slug>` | string | — | Target a project by slug |

---

### `chauf env list`

```bash
chauf env list
chauf env list --slug my-app
```

**Output**:
```
  Environment Variables: my-app

  APP_ENV      local
  APP_DEBUG    true
  DB_HOST      127.0.0.1
  DB_PORT      3306
  DB_DATABASE  my_app_db

  5 variables set
```

---

### `chauf env set`

```bash
chauf env set APP_ENV local
chauf env set DB_HOST 127.0.0.1 --slug my-app
chauf env set APP_KEY "base64:abc...xyz"
```

**Output**:
```
  ✓ APP_ENV set to "local"

  nginx config regenerated. Reload to apply:
    chauf restart nginx
```

Key naming: uppercase letters, digits, underscores only. Validated before writing.

---

### `chauf env unset`

```bash
chauf env unset APP_DEBUG
chauf env unset DB_PASSWORD --slug my-app
```

---

### `chauf env import`

Import from a standard `.env` file. Lines starting with `#` are comments. `KEY=VALUE`, `KEY="VALUE"`, and `KEY='VALUE'` formats are all supported.

```bash
# Import from .env in current directory
chauf env import .env

# Import from a specific file
chauf env import .env.local --slug my-app
```

**Output**:
```
Importing from .env...

  ✓ APP_ENV
  ✓ APP_DEBUG
  ✓ APP_KEY
  ✓ DB_HOST
  ✓ DB_DATABASE
  − DB_PASSWORD  skipped (empty value)

  5 variables imported. Reload to apply:
    chauf restart nginx
```

---

### `chauf env export`

Print current env vars in `.env` format — suitable for piping into a file or sharing.

```bash
chauf env export
chauf env export --slug my-app > .env.backup
```

**Output**:
```
APP_ENV=local
APP_DEBUG=true
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=my_app_db
```

**Notes**:
- Variables are stored in the project config, not in the project's source directory.
- The project's own `.env` file is never read or modified — `chauf env` is a separate management layer.
- To apply changes to a running project, reload nginx: `chauf restart nginx`.

---

## `chauf autostart`

**Purpose**: Manage systemd user services for automatic service startup on login. Uses `systemctl --user` — no root required.

**Usage**:
```
chauf autostart <subcommand> [service] [flags]
```

**Subcommands**:

| Subcommand | Description |
|------------|-------------|
| `enable [service]` | Create unit file, enable, and start the service |
| `disable [service]` | Stop, disable, and optionally remove the unit file |
| `status` | Show status of all Chauffeur systemd units |
| `list` | List all Chauffeur unit files and their enabled state |

**Services** (for `enable`/`disable`):

| Service | Unit name |
|---------|-----------|
| *(omitted)* | Enable/disable nginx + all installed PHP-FPM pools |
| `nginx` | `chauffeur-nginx.service` |
| `php <version>` | `chauffeur-php-fpm@<version>.service` |

---

### `chauf autostart enable`

```bash
# Enable all Chauffeur services
chauf autostart enable

# Enable only nginx
chauf autostart enable nginx

# Enable PHP 8.3 FPM only
chauf autostart enable php 8.3
```

**Output**:
```
Enabling Chauffeur services...

  ✓ chauffeur-nginx.service          enabled and started
  ✓ chauffeur-php-fpm@8.3.service    enabled and started

  Services will start automatically on next login.

  ⚠ Services start on login, not on boot.
    To start on boot: loginctl enable-linger $USER
```

---

### `chauf autostart disable`

```bash
chauf autostart disable
chauf autostart disable nginx
chauf autostart disable php 8.3
```

---

### `chauf autostart status`

```bash
chauf autostart status
```

**Output**:
```
  Chauffeur Auto-Start

  chauffeur-nginx.service        ● active   since 09:00:01
  chauffeur-php-fpm@8.3          ● active   since 09:00:02
  chauffeur-php-fpm@8.1          ○ inactive  — not enabled

  Lingering: disabled
  To start on boot: loginctl enable-linger $USER
```

---

### `chauf autostart list`

```bash
chauf autostart list
```

**Output**:
```
  Unit                              File                                           Enabled
  ────────────────────────────────  ─────────────────────────────────────────────  ───────
  chauffeur-nginx.service           ~/.config/systemd/user/chauffeur-nginx.service  yes
  chauffeur-php-fpm@8.3.service     ~/.config/systemd/user/chauffeur-php-fpm@...    yes
  chauffeur-php-fpm@8.1.service     (no unit file)                                  no
```
