# Podman Database Containers Specification

## Overview

Chauffeur V2 includes integrated Podman support for managing database containers (MySQL, PostgreSQL, MariaDB, MongoDB, Redis). Containers are managed via `chauf podman` subcommands with the same UX philosophy as other Chauffeur services.

---

## Supported Database Engines

| Engine | Internal Port | Image | Env Var for Password |
|--------|--------------|-------|---------------------|
| MySQL 5.7 | 3306 | `docker.io/library/mysql:5.7` | `MYSQL_ROOT_PASSWORD` |
| MySQL 8 | 3306 | `docker.io/library/mysql:8.0` | `MYSQL_ROOT_PASSWORD` |
| MariaDB 11 | 3306 | `docker.io/library/mariadb:11` | `MARIADB_ROOT_PASSWORD` |
| PostgreSQL 16 | 5432 | `docker.io/library/postgres:16` | `POSTGRES_PASSWORD` |
| MongoDB 7 | 27017 | `docker.io/library/mongo:7` | `MONGO_INITDB_ROOT_USERNAME/PASSWORD` |
| Redis 7 | 6379 | `docker.io/library/redis:7-alpine` | (none, unauthenticated) |

---

## Workspace Layout

```
~/.chauffeur/
├── podman/
│   ├── <container-name>.yaml   # Container configuration (e.g., chauf-mysql57.yaml)
│   ├── volumes/
│   │   └── <container-name>/   # Bind mount volumes for data persistence
│   └── backups/
│       └── <container>-<database>-<timestamp>.tar.gz
```

Note: Config files are stored directly in `~/.chauffeur/podman/` with the container name as filename.

---

## Container Config Schema

File: `~/.chauffeur/podman/<container-name>.yaml`

```yaml
name: mysql57
engine: mysql57
image: docker.io/library/mysql:5.7
container_name: chauf-mysql57
username: root
password: mysecretpassword
port: 3306
volume_path: /home/user/.chauffeur/podman/volumes/chauf-mysql57
env:
  - key: MYSQL_ROOT_PASSWORD
    value: mysecretpassword
created_at: "2025-01-01T00:00:00Z"
```

---

## CLI Commands

### `chauf podman create`

Interactive creation of a database container.

```bash
chauf podman create [mysql57|mysql8|postgres|maria|mongo|redis] [--verbose]
```

**Flow**:
1. Select engine type
2. Enter container name (optional, default: `chauf-<engine>`)
3. Enter database username (optional, default: `chauf`)
4. Enter database password (optional, auto-generated)
5. Enter host port (optional, engine-specific default)
6. Enter volume path (optional, default: `~/.chauffeur/podman/volumes/<container-name>`)
7. Confirm and create

**Post-Creation**:
- Container is started automatically
- Privileges are granted to the database user
- If privilege granting fails, container is rolled back (removed)

**Flags**:
- `--verbose`: Show detailed progress during creation

---

### `chauf podman start`

Start a stopped container.

```bash
chauf podman start [<container-name>]
```

If no container specified, shows interactive picker listing all containers (running and stopped).

---

### `chauf podman stop`

Stop a running container.

```bash
chauf podman stop [<container-name>] [--time <seconds>]
```

**Flags**:
- `--time <seconds>`: Seconds to wait before SIGKILL (default: 10)

---

### `chauf podman remove`

Remove a container and its configuration.

```bash
chauf podman remove [<container-name|engine>] [--force] [--yes]
```

**Flow**:
1. If container is running and not `--force`: ask to stop first
2. Confirm removal
3. Remove container from podman
4. Clean up configuration file

**Flags**:
- `--force`: Force remove even if running
- `--yes`: Skip all confirmations

**Orphaned Config Handling**:
- If container doesn't exist but config does (orphaned), offers to clean up config only

---

### `chauf podman list`

List all managed containers.

```bash
chauf podman list [--verbose]
```

**Status Values**:
- `● running` - Container is running
- `○ stopped` - Container exists but is not running
- `⚠ missing` - Config exists but container not found in podman (orphaned)

**Flags**:
- `--verbose`: Show config file paths

---

### `chauf podman status`

Show detailed status of a container.

```bash
chauf podman status [<container-name>]
```

---

### `chauf podman console`

Attach to container's database CLI.

```bash
chauf podman console <container>
```

For MySQL/MariaDB: `mysql -u <user> -p`
For PostgreSQL: `psql -U <user>`
For MongoDB: `mongosh`
For Redis: `redis-cli`

---

### `chauf podman backup`

Interactive backup of databases from a running container.

```bash
chauf podman backup
```

**Flow**:
1. Select container (running containers only)
2. Test connection, prompt for credentials if needed
3. List databases in container
4. Interactive database selection (TUI with bubbletea)
5. Ask if user wants to add descriptions (default: N)
6. If yes, prompt for description per database
7. Show backup summary with descriptions
8. Confirm and execute backups

**Output**:
- Backups saved to `~/.chauffeur/podman/backups/`
- Filename: `<container>-<database>-<timestamp>.tar.gz`
- Format: tar.gz containing `meta.json` and `<database>.sql|dump`

**Backup Metadata** (`meta.json`):
```json
{
  "engine": "mysql57",
  "container_name": "chauf-mysql57",
  "timestamp": "2026-03-28T09:39:49Z",
  "username": "chauf",
  "database": "hja-cms",
  "description": "Sample backup"
}
```

---

### `chauf podman restore`

Interactive restore of databases from backup.

```bash
chauf podman restore
```

**Flow**:
1. Select container (running containers only)
2. Test connection, prompt for credentials if needed
3. List available backups from `~/.chauffeur/podman/backups/`
4. Group backups by database name, sorted by newest first
5. Select database to restore
6. If multiple backups, select which backup to use with timestamp/description
7. Show restore summary with timestamp, description, size
8. Confirm and execute restore

**Backup Selection UI**:
```
  Select backup to restore for personal-cms:

   Timestamp           Description
  --- ---------        -----------
  1) 2026-03-28 09:51:24  -
  2) 2026-03-28 09:43:41  Another sample
  3) 2026-03-28 09:42:59  Sample description

  Choice [1-3]:
```

**Restore Summary**:
```
  Restore summary:
    Container:  chauf-mysql57
    Database:  personal-cms
    Backup:    chauf-mysql57-personal-cms-20260328-094258.tar.gz (650 B)
    Created:   2026-03-28 09:42:59
    Description: Sample description

  Proceed? [y/N]:
```

---

## Internal Port Mapping

Container ports are fixed by database engine, not user-configurable:

| Engine | Container Port |
|--------|---------------|
| MySQL 5.7/8, MariaDB | 3306 |
| PostgreSQL | 5432 |
| MongoDB | 27017 |
| Redis | 6379 |

User selects the **host port** during creation, which maps to the fixed container port.

---

## Privilege Granting

After container creation, Chauffeur automatically grants database privileges:

- **MySQL 5.7/8**: Grants ALL PRIVILEGES ON *.* TO user
- **MariaDB**: Uses `mariadb` client (not `mysql`) to grant privileges
- **PostgreSQL**: Default postgres user has superuser privileges via POSTGRES_PASSWORD
- **MongoDB**: User created via MONGO_INITDB_ROOT_USERNAME/PASSWORD
- **Redis**: No authentication needed

### MySQL/MariaDB Startup Detection

- MySQL: Uses `mysqladmin ping` to detect readiness
- MariaDB: Uses `mariadb-admin ping` to detect readiness (MariaDB uses different commands)
- Timeout: 30-60 seconds for database to become ready
- If startup fails within timeout, container is rolled back and removed

---

## Connection String (DSN)

```bash
mysql://<user>:<pass>@127.0.0.1:<port>/<database>
postgres://<user>:<pass>@127.0.0.1:<port>/<database>
```

---

## Environment Variable Defaults

| Engine | Password Env Var | Additional Env Vars |
|--------|-----------------|-------------------|
| MySQL 5.7/8 | `MYSQL_ROOT_PASSWORD` | `MYSQL_DATABASE=app` |
| MariaDB | `MARIADB_ROOT_PASSWORD` | `MARIADB_DATABASE=app` |
| PostgreSQL | `POSTGRES_PASSWORD` | `POSTGRES_USER`, `POSTGRES_DB=app` |
| MongoDB | `MONGO_INITDB_ROOT_USERNAME`, `MONGO_INITDB_ROOT_PASSWORD` | |
| Redis | (none) | |

**Note**: For MySQL/MariaDB, `MYSQL_USER` / `MARIADB_USER` is NOT set when username is `root` (the Docker image rejects this configuration).

---

## Dependencies

- `podman` binary must be installed and in PATH
- Uses `podman exec` for commands inside containers
- Uses `podman cp` for backup/restore operations
- Network `chauf-net` is created if not exists

---

## Error Handling

| Error | Message |
|-------|---------|
| Podman not found | "Podman is not installed or not in PATH." |
| Container not found | "No containers found matching <name>" |
| Container not running | "Container <name> is not running. Start it first." |
| Orphaned config | "Container <name> not found in podman (orphaned config)" |
| Connection failed | Prompts for credentials with retry loop |
| MariaDB startup timeout | Rolls back container automatically |
| Backup failed | "Backup failed for <db>: <error>" |
| Restore failed | "Restore failed: <error>" |
| Port not available | "Port <port> is already in use by process <pid> (<name>)" |

---

## File Operations

- **Config files**: `~/.chauffeur/podman/<container-name>.yaml`
- **Volume directories**: Created at user-specified path or `~/.chauffeur/podman/volumes/<container-name>/`
- **Backup files**: `~/.chauffeur/podman/backups/<container>-<database>-<timestamp>.tar.gz`
- **Podman volumes**: Named volumes created for each container (managed by podman)
