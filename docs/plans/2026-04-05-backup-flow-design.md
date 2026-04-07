# Backup Flow Redesign

**Date:** 2026-04-05

## Overview

Redesign the web panel backup flow to match the TUI behavior:
1. Select container
2. Show available databases
3. Multi-select databases (with checkboxes)
4. Optional descriptions per database
5. Create backups

## TUI Flow Reference

From `internal/commands/podman.go:1809-1937`:
```go
// 1. Select container
target := interactiveSelectContainer(runningContainers, ctx, client, "Select container to backup:")

// 2. List databases
databases, err := container.ListDatabases(ctx)

// 3. Multi-select databases
selected := interactiveSelectDatabases(databases)

// 4. Optional descriptions per database
addDescriptions := promptYesNo("Add descriptions?")
if addDescriptions {
    for _, db := range selected {
        descriptions[db] = promptDescription(db)
    }
}

// 5. Execute backups
for _, db := range selected {
    container.BackupDatabaseWithDescription(ctx, db, descriptions[db])
}
```

## API Changes

### GET /api/containers/{name}/databases

Returns list of database names for a container.

**Response:**
```json
{
  "databases": ["app", "api_prod", "logs"]
}
```

### POST /api/backups

Accepts multiple databases with optional descriptions.

**Request:**
```json
{
  "container": "chauf-mysql57",
  "databases": [
    { "name": "app", "description": "before migration" },
    { "name": "api_prod", "description": "" }
  ]
}
```

**Response:**
```json
{
  "message": "Backed up 2 database(s)",
  "backups": ["chauf-mysql57-app-20260405-143022.tar.gz", "chauf-mysql57-api_prod-20260405-143022.tar.gz"]
}
```

## UI Changes

### Create Backup Modal

**Step 1: Container Selection**
- Dropdown showing running containers (name + engine)
- Default state: placeholder "Select a container..."

**Step 2: Database Selection (appears after container selected)**
- List of databases with checkboxes
- System databases filtered out (information_schema, performance_schema, mysql, sys)
- "Select all" / "Deselect all" toggle

**Step 3: Descriptions**
- When a database is checked, show optional description input below it
- Input labeled "Description for [database]"

**Step 4: Actions**
- Cancel button (closes modal, resets state)
- Create Backup button (disabled if no databases selected)
- Shows count: "Backup 3 database(s)"

### Error Handling

- If container not running: show error, disable database selection
- If no databases found: show "No databases found" message
- On backup failure: show error toast, keep modal open

## Components

- shadcn `Checkbox` for database selection
- shadcn `Input` for descriptions
- shadcn `Select` for container dropdown
- shadcn `Dialog` for modal

## Status

- [x] Design approved
- [ ] Implementation pending
