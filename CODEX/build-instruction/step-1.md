# step-1 — Bootstrap `chauf` installation (instructions only)

## Context
Chauffeur is a host-based CLI that installs and manages its own binaries under `~/.chauffeur/` without touching system prefixes. Users need a simple, **idempotent** way to set up the workspace and expose `~/.chauffeur/bin` on their shell `PATH`.

## Goals
- Create the initial workspace structure in the user’s home directory.
- Ensure `~/.chauffeur/bin` is on `PATH` for new shells (bash & zsh) **exactly once**.
- Place (or fetch) the `chauf` binary into `~/.chauffeur/bin/chauf`.
- Keep everything **user-space only**; do not modify `/usr` or system service units.

## Constraints / Policies
- Install prefix is **`~/.chauffeur/`**. Never write to `/usr`, `/usr/local`, or `/etc`.
- No external environment managers (no Devbox, nix, etc.).
- Re-running bootstrap must be safe (no duplicate PATH lines, no errors if dirs exist).
- Telemetry defaults to **off** unless explicitly enabled later.
- Support Linux (Arch/Wayland friendly); other OS targets are out of scope for this step.

## What this step must create
- Directories:
  - `~/.chauffeur/bin/` (for `chauf` and shims)
  - `~/.chauffeur/bin/shims/` (wrappers like `php-8.3`, `nginx`, `caddy` — created empty here)
  - `~/.chauffeur/config/`
  - `~/.chauffeur/projects/`
  - `~/.chauffeur/php/`
  - `~/.chauffeur/nginx/{bin,etc,sites-available,sites-enabled,conf.d}`
  - `~/.chauffeur/caddy/{bin}`
- File(s):
  - `~/.chauffeur/bin/chauf` (binary placed or downloaded by URL, executable bit set)
  - `~/.chauffeur/config/chauffeur.yaml` (created later by `chauf init`; not required in this step)

## Required UX messages (examples)
- Success: "Workspace created at `~/.chauffeur`. Add `~/.chauffeur/bin` to PATH (bash/zsh)"
- If PATH already configured: "PATH already contains `~/.chauffeur/bin` — skipping"
- If `chauf` binary missing and no URL provided: "Place your built `chauf` binary at `~/.chauffeur/bin/chauf` or set `CHAUF_RELEASE_URL` to download"

## PATH handling policy
- Add `export PATH="$HOME/.chauffeur/bin:$PATH"` to `~/.bashrc` (on bash) or `~/.zshrc` (on zsh) **only if** the exact line is not already present.
- Do not modify other shells; print a notice for manual action if the shell is unknown.

## Error handling guidelines
- Missing `curl` (when downloading) → clear error with remediation.
- Permission issues writing to home → actionable message and exit non‑zero.
- Partial setups (e.g., dirs exist) are fine; continue idempotently.

## Acceptance criteria
1. Running bootstrap twice is a no‑op (no duplicate lines, no failures).
2. `~/.chauffeur/bin` exists and is on PATH for new interactive shells.
3. `~/.chauffeur/bin/chauf` exists or the user receives a clear message how to provide it.
4. No writes occurred outside `~/.chauffeur/`.

## Out of scope (deferred)
- Generating `chauffeur.yaml` or per‑project configs.
- Installing Nginx/Caddy/PHP (covered in step-2).
- Creating service shims (will be created as installers land).

## Next step
Proceed to **step-2 — Initialize services** to install user‑space copies of Nginx, Caddy, and PHP under `~/.chauffeur/` and prepare minimal configs and shims.

