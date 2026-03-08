# Chauffeur V2 Documentation

## Command Reference

| File | Commands Covered |
|------|----------------|
| [commands/workspace.md](./commands/workspace.md) | `init`, `info`, `uninstall` |
| [commands/project.md](./commands/project.md) | `link`, `links`, `unlink`, `secure`, `unsecure` |
| [commands/php.md](./commands/php.md) | `php list`, `php use`, `php isolate`, `php install`, `php remove` |
| [commands/services.md](./commands/services.md) | `start`, `stop`, `restart`, `status`, `logs` |
| [commands/install.md](./commands/install.md) | `install`, `remove` |
| [commands/config.md](./commands/config.md) | `config`, `env`, `autostart` *(V2)* |
| [commands/maintenance.md](./commands/maintenance.md) | `doctor`, `clean`, `migrate`, `self-update` |

## Guides

> These files will be created as the project grows.

- `getting-started.md` — Installation and first project
- `php-fpm.md` — Shared vs dedicated FPM strategies
- `ssl.md` — SSL certificate setup
- `multi-domain.md` — Multiple domains per project
- `troubleshooting.md` — Common issues

## Global Flags

Available on every command:

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show command help |
| `--version`, `-v` | Print version and exit |
| `--workspace <path>` | Override workspace root (default: `~/.chauffeur`) |
| `--log-level <level>` | Set verbosity: `debug`, `info`, `warn`, `error` |
| `--no-color` | Disable ANSI color output |
| `--quiet`, `-q` | Suppress info output; show only warnings and errors |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Usage error (invalid flags, missing arguments) |
| `3` | Service error (nginx/PHP-FPM failed to start or stop) |
| `4` | Configuration error (invalid config, schema validation failure) |
| `5` | Dependency error (missing host tool or library) |
