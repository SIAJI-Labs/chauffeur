# Chauffeur Scripts

This directory contains utility scripts for Chauffeur development and release management.

## tag-creator

A Go-based interactive tool for creating Git tags with proper versioning. This is the Go equivalent of the original JavaScript version.

### Features

- **Interactive version selection** - Choose between major, minor, patch, or custom versions
- **Git safety checks** - Validates clean working directory before proceeding
- **Branch management** - Automatically switches to main branch and back
- **Dry-run mode** - Safe testing without actual Git operations
- **Validation** - Ensures version tags follow semantic versioning format

### Usage

```bash
# Interactive mode (creates actual tags)
./tag-creator

# Dry-run mode (safe preview)
./tag-creator --dry-run

# Show help
./tag-creator --help

# Show version
./tag-creator --version

# Build from source
go build -o tag-creator tag-creator.go
```

### Safety Features

- **Uncommitted changes detection** - Refuses to proceed with dirty working directory
- **Branch validation** - Ensures operations happen on main branch
- **Remote sync** - Fetches latest tags and pulls changes
- **Version format validation** - Enforces vX.Y.Z format
- **Dry-run mode** - Safe preview without actual Git operations

### Command Line Flags

- `--dry-run` - Run in dry-run mode (no actual Git operations)
- `--help` - Show help message
- `--version` - Show version information

### Dependencies

- Go 1.22+
- `github.com/AlecAivazis/survey/v2` for interactive prompts

### Example Output

**Dry-run mode:**
```
🏷️  Chauffeur Tag Creator (DRY RUN)
⚠️  No actual Git operations will be performed

Currently on develop, switching to main...
🌿 DRY RUN: Would switch to branch: main
Fetching latest changes...
🔍 DRY RUN: Would execute: git fetch --tags
📥 DRY RUN: Would pull current branch
Latest tag found: v0.0.0
? Select version bump type: [Use arrows to move, type to filter]
  > major (v1.0.0)
    minor (v0.1.0)
    patch (v0.0.1)
    other (enter manually)

🏷️  DRY RUN: Would create tag: v0.0.1
🚀 DRY RUN: Would push tag to remote: v0.0.1

✅ Dry run completed - no actual changes made
```

**Live mode:**
```
🏷️  Chauffeur Tag Creator

Currently on develop, switching to main...
Switched to branch 'main'
Fetching latest changes...
Latest tag found: v0.0.0
? Select version bump type: [Use arrows to move, type to filter]
  > major (v1.0.0)
    minor (v0.1.0)
    patch (v0.0.1)
    other (enter manually)

🏷️  Tag created: v0.0.1
🚀 Tag pushed to remote: v0.0.1

✅ Tag creation completed!
```