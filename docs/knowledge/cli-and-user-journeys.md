# CLI And User Journeys

## Command dispatcher

The entry point is `cmd/chauf/main.go`. It manually dispatches commands and pre-processes `--verbose`/`-v`. Supported top-level commands are:

```text
init info uninstall
install remove
link links unlink secure unsecure
php
start stop restart status logs
autostart config env
self-update doctor update clean migrate
podman serve
```

`migrate` is an explicit stub. `env` is dispatched to `RunConfig` and currently does not provide the documented environment subcommands.

## Canonical first-run journey

```bash
curl -sSL https://chauffeur.siaji.com/install | bash
chauf doctor
chauf init
chauf install nginx php 8.3 composer
cd ~/Projects/my-laravel-app
chauf link --secure
chauf start
```

The expected URL is `https://my-laravel-app.test:8443`, subject to DNS and port-forwarding setup.

## Project journey

```bash
chauf link [path] [--php VERSION] [--secure] [--dedicated-fpm]
chauf links [--detail]
chauf php isolate VERSION
chauf secure [--project PATH]
chauf unsecure [--project PATH]
chauf unlink [--alias DOMAIN] [--all] [--yes]
```

Linking is the explicit registration boundary. Relinking is intended to update an existing project. A project can have a primary domain and repeatable aliases.

## PHP journey

```bash
chauf php list
chauf php use VERSION
chauf php isolate VERSION
chauf php install VERSION
chauf php remove VERSION
```

The global default is stored in workspace config; project isolation is stored in project config. Installed versions are compiled under the workspace.

## Service journey

```bash
chauf start [--project PATH]
chauf stop [--project PATH]
chauf restart [nginx|php|fpm [VERSION]] [--project PATH]
chauf status [--detail] [--project PATH]
chauf logs [nginx|access|php [VERSION]] [--follow] [--lines N] [--project PATH]
```

The verified implementation centers on these flags and targets. Some docs/specs describe additional positional forms such as `start nginx`, `start php`, `--all`, or `--dry-run`; those forms need reconciliation before being treated as public API.

## Install and maintenance journey

```bash
chauf install nginx|php [VERSION]|composer [--force] [--no-cache]
chauf remove nginx|php VERSION|composer [--force]
chauf doctor [--check-deps|--check-php|--check-ssl|--check-network|--check-dns] [--fix|--auto-fix]
chauf clean cache|logs|all [--dry-run] [--older-than AGE]
chauf autostart enable|disable|status|list [nginx|php VERSION]
chauf update nginx|php VERSION|composer|all|list|check|rollback
```

`doctor --fix` prints remediation commands. `--auto-fix` executes generated shell commands and is therefore a sensitive trust boundary.

## Podman journey

```bash
chauf podman create [ENGINE] [--name NAME] [--user USER] [--pass PASS] [--port PORT]
chauf podman start [NAME|ENGINE|all]
chauf podman stop [NAME|ENGINE|all] [--time SECONDS]
chauf podman list
chauf podman status
chauf podman remove [NAME]
chauf podman console [NAME]
chauf podman backup
chauf podman restore
```

Supported engines are MySQL 8, MySQL 5.7, PostgreSQL, MariaDB, MongoDB, and Redis. Config and persistent data are intended to stay within the Chauffeur workspace.

## Current contract problems

- Docs advertise global flags such as `--workspace`, `--quiet`, `--no-color`, and `--log-level` that are not parsed globally in `main.go`.
- Documentation uses `unlink --force` in places; code uses `--yes`.
- The docs describe generic config and env management; code only handles narrow PHP/nginx settings.
- README links to guide files absent from the current docs tree.
- `doctor` flag naming differs between docs and implementation.

## Recommended UX grammar

Keep current short commands for compatibility, but choose a canonical project-centric vocabulary for new work:

```text
chauf project link/list/unlink/secure/status/logs
chauf service start/stop/restart/status/logs
chauf db create/list/start/stop/backup/restore
```

The current aliases can remain, but help output and new docs should show one primary grammar.
