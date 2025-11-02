# Chauffeur (Work in Progress)

Chauffeur is a host-based CLI for managing per-project PHP development services on Linux. It installs everything into `~/.chauffeur/`, isolates runtimes (PHP-FPM, Nginx, Caddy), and keeps system packages untouched. This repository currently contains the bootstrap scripts and an early Go CLI stub while the full feature set is under active development.

## Status

- 🚧 **Work in progress**: Only installer and `chauf uninstall` command are implemented.
- Linux-focused workflow (Arch/Wayland friendly); other OS targets are not yet supported.
- CLI structure uses modular Go packages (`cli/commands/*`). Additional commands will land incrementally.

## Getting Started

```bash
./install.sh
# Reload your shell or run: source ~/.zshrc
chauf --version
```

If you need to remove the workspace:

```bash
chauf uninstall          # keeps PHP runtimes by default
chauf uninstall --purge  # removes workspace and runtimes/caches
```

## Roadmap

- `chauf init` to scaffold the workspace and global config.
- PHP runtime management (`chauf php install/use/isolate`).
- Project registration (`chauf link`, `chauf links`) with per-project configs.
- Service orchestration (`chauf start`, `chauf stop`) for Nginx, PHP-FPM, and Caddy.
- Automated shim generation and log handling.

## TODO

- Flesh out CLI subcommands beyond `--version` and `uninstall`.
- Implement unit tests for command packages.
- Define binary acquisition strategy (build vs. download) in installer.
- Document contribution guidelines once the CLI stabilizes.

## Development Notes

- Requires Go 1.22+ to build the CLI.
- Scripts are idempotent; rerun `./install.sh` safely.
- Changes should respect the contracts outlined in `CODEX.md`.
