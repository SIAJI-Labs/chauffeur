# PHP Commands

Commands for managing PHP versions. All are subcommands of `chauf php`.

---

## `chauf php list`

**Purpose**: List all PHP versions installed in the workspace, showing their installation path and which one is the global default.

**Usage**:
```
chauf php list
```

No flags.

**Examples**:

```bash
chauf php list
```

**Output**:
```
  PHP     Path                        Status
  ──────  ──────────────────────────  ───────────
  8.4     ~/.chauffeur/php/8.4/       installed
  8.3     ~/.chauffeur/php/8.3/       installed  (default)
  8.1     ~/.chauffeur/php/8.1/       installed
  7.4     ~/.chauffeur/php/7.4/       installed
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

**Purpose**: Pin the current directory to a specific PHP version by writing a `.chauffeur-php` file. The PHP shim reads this file when walking up the directory tree.

**Usage**:
```
chauf php isolate <version>
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `<version>` | yes | PHP version to pin (e.g. `8.1`) |

No flags. Run from the project root directory to isolate that project.

**Examples**:

```bash
# Pin current directory to PHP 8.1
cd ~/Projects/legacy-app
chauf php isolate 8.1
```

**Output**:
```
  ✓ Pinned to PHP 8.1  (.chauffeur-php)
```

**How isolation works**: A `.chauffeur-php` file containing only the version string (e.g. `8.1`) is written to the current directory. The PHP shim walks up the directory tree from `$PWD` looking for this file. When found, it invokes that version's binary.

**Notes**:
- The file is intentionally human-editable — `echo "8.1" > .chauffeur-php` works too.
- To remove isolation (revert to global default), delete `.chauffeur-php`.
- Consider adding `.chauffeur-php` to `.gitignore` if PHP version is controlled differently per-developer.

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
