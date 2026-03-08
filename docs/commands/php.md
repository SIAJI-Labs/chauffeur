# PHP Commands

Commands for managing PHP versions. All are subcommands of `chauf php`.

---

## `chauf php list`

**Purpose**: List all PHP versions installed in the workspace, showing their installation path and which one is the global default.

**Usage**:
```
chauf php list [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--detail` | bool | false | Show binary path, extension list, compiled flags |

**Examples**:

```bash
chauf php list
chauf php list --detail
```

**Output**:
```
  PHP     Path                        Status
  ──────  ──────────────────────────  ───────────
  8.4     ~/.chauffeur/php/8.4/       installed
  8.3     ~/.chauffeur/php/8.3/       installed  (default)
  8.1     ~/.chauffeur/php/8.1/       installed
  7.4     ~/.chauffeur/php/7.4/       installed

  8.0, 8.2 — not installed
```

**With `--detail`**:
```
  8.3  (default)
    Binary      ~/.chauffeur/php/8.3/bin/php
    FPM         ~/.chauffeur/php/8.3/bin/php-fpm
    Extensions  mysqli  pdo_mysql  gd  zip  imagick
                curl  openssl  sodium  readline  bcmath  gmp  exif  xsl  bz2
```

---

## `chauf php use`

**Purpose**: Set the global default PHP version. This is the version used by the `php` shim when no project-specific version is configured.

**Usage**:
```
chauf php use <version>
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `<version>` | yes | PHP version to set as default (e.g. `8.3`) |

**Examples**:

```bash
chauf php use 8.3
chauf php use 8.1
```

**Output**:
```
  ✓ Global default PHP set to 8.3  (was 8.1)

  Projects with explicit version pinning are unaffected.
```

**Errors**:
- Not installed: `✗ PHP 8.0 is not installed. Run: chauf install php 8.0`
- Not supported: `✗ Unsupported PHP version: 9.0  (supported: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4)`

---

## `chauf php isolate`

**Purpose**: Pin the current project to a specific PHP version by writing a `.chauffeur-version` file in the project root and updating the project config. The PHP shim reads this file when walking up the directory tree.

**Usage**:
```
chauf php isolate <version> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `<version>` | yes | PHP version to pin to this project (e.g. `8.1`) |

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project <path>` | string | CWD | Isolate a project at a specific path |

**Examples**:

```bash
# Pin current project to PHP 8.1
chauf php isolate 8.1

# Pin a project at a specific path
chauf php isolate 7.4 --project ~/Projects/legacy-app
```

**Output**:
```
  ✓ my-app pinned to PHP 8.1  (was 8.3)

  File     ~/Projects/my-app/.chauffeur-version
  nginx config regenerated  ·  nginx reloaded
```

**How isolation works**: The `.chauffeur-version` file contains only the version string (e.g. `8.1`). The PHP shim walks up the directory tree from the current working directory looking for this file. When found, it uses the specified version's binary.

**Notes**:
- The file is intentionally human-editable — `echo "8.1" > .chauffeur-version` works too.
- To remove isolation (revert to global default), delete `.chauffeur-version` and rerun `chauf link` or `chauf link --php <version>`.
- You may want to add `.chauffeur-version` to `.gitignore` for projects where PHP version is controlled differently.

---

## `chauf php install`

**Purpose**: Alias for `chauf install php <version>`. Compile and install a PHP version from source. See [install.md](./install.md) for full documentation.

**Usage**:
```
chauf php install <version> [flags]
```

**Examples**:

```bash
chauf php install 8.3
chauf php install 7.4 --force
```

---

## `chauf php remove`

**Purpose**: Alias for `chauf remove php <version>`. Remove an installed PHP version from the workspace. See [install.md](./install.md) for full documentation.

**Usage**:
```
chauf php remove <version> [flags]
```

**Examples**:

```bash
chauf php remove 7.4
chauf php remove 8.0 --force
```

**Pre-removal check**: If any projects are using the version being removed (shared or dedicated pool), the command lists them and asks for confirmation before proceeding. Affected projects must be relinked to a different version afterward.
