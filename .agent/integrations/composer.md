# Composer Integration

## Overview

Chauffeur installs Composer as a PHAR file and creates a shim that automatically uses the correct PHP version for the current project.

---

## Installation

```go
// internal/installers/composer.go

const ComposerInstallerURL = "https://getcomposer.org/installer"
const ComposerPHARURL     = "https://getcomposer.org/download/latest-stable/composer.phar"

func Install(ws *workspace.Workspace) error {
    pharPath := ws.ComposerPHAR()

    // Download PHAR
    if err := downloadFile(ComposerPHARURL, pharPath); err != nil {
        return fmt.Errorf("composer: download failed: %w", err)
    }

    // Verify checksum
    checksum, err := fetchComposerChecksum()
    if err != nil {
        return fmt.Errorf("composer: checksum fetch failed: %w", err)
    }
    if err := verifyChecksum(pharPath, checksum); err != nil {
        return fmt.Errorf("composer: checksum mismatch: %w", err)
    }

    // Create shim
    return createComposerShim(ws)
}
```

---

## PHAR Location

```
~/.chauffeur/composer/composer.phar
```

---

## Shim

The shim at `~/.chauffeur/bin/shims/composer` passes all arguments to the Chauffeur PHP binary:

```bash
#!/bin/bash
# ~/.chauffeur/bin/shims/composer

CHAUFFEUR_HOME="${CHAUFFEUR_HOME:-$HOME/.chauffeur}"

# Use the same PHP shim (which resolves version per project)
exec "$CHAUFFEUR_HOME/bin/shims/php" \
    "$CHAUFFEUR_HOME/composer/composer.phar" \
    "$@"
```

This ensures `composer install` in a PHP 8.1 project uses PHP 8.1, not the global default.

---

## Caching

Composer downloads are cached in `~/.chauffeur/cache/composer/`:

```
~/.chauffeur/cache/composer/
└── composer.phar              # Cached PHAR
```

`chauf install composer` uses the cached PHAR if it exists and its checksum matches.

---

## Usage

After installation, `composer` resolves to Chauffeur's shim if `~/.chauffeur/bin/shims/` is on `$PATH`:

```bash
# In a PHP 8.3 project
composer install   # uses ~/.chauffeur/php/8.3/bin/php

# In a PHP 8.1 project (with .chauffeur-version = 8.1)
composer install   # uses ~/.chauffeur/php/8.1/bin/php
```

---

## PATH Setup

`chauf init` outputs PATH setup instructions (never modifies shell config automatically):

```
Add this to your ~/.bashrc or ~/.zshrc:

  export PATH="$HOME/.chauffeur/bin/shims:$PATH"

Then reload your shell:
  source ~/.bashrc
```

---

## Updating Composer

```bash
chauf update composer       # Download latest stable PHAR
```

Or using Composer's self-update:

```bash
composer self-update        # Works via shim
```
