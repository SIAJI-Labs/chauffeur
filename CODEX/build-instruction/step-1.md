# step-1 — Installer Script Spec (`install.sh`)

> **Audience**: Chauffeur developers (we are building the tool). This is a **specification** for the installer script we will ship as `install.sh`. No code here—only requirements and behavior.

## Context
We ship a single-file installer (`install.sh`) that, when executed by end users, installs the Chauffeur CLI and prepares the workspace in user space only. It must be safe, verifiable, and idempotent.

## Objectives
- Provide a frictionless install experience for Linux users.
- Keep all files under `~/.chauffeur/`; never touch `/usr`, `/usr/local`, or system service units by default.
- Verify downloaded artifacts (checksums/signatures) before install.
- Be **idempotent** (re-running causes no harm) and **transparent** (clear logs & prompts).

## Deliverable
- A POSIX-compatible shell script named **`install.sh`** (bash preferred, dash-compatible where possible).
- Hosted alongside releases (e.g., GitHub Releases) and downloadable via curl/wget piping.

## Invocation & UX
- Common usage: `curl -fsSL https://…/install.sh | bash`
- Flags (script-level):
  - `--prefix <dir>`: default `$HOME/.chauffeur`
  - `--channel <stable|canary>`: default `stable`
  - `--version <x.y.z|latest>`: default `latest` within channel
  - `--no-path`: do not modify shell rc files
  - `--dry-run`: print planned actions, no changes
  - `--verbose`: extra logging
- Non-interactive by default; print next steps on success.

## Environment Assumptions
- Linux x86_64 primary; plan for arm64 later (detect via `uname -m`).
- Requires: `curl` or `wget`, `tar`, `sha256sum` (or `shasum -a 256`), `install`.
- Shell: `bash` (gracefully degrade for `sh` where possible).

## Install Prefix & Layout (contract)
- **Prefix**: `$HOME/.chauffeur`
- Directories created:
  - `bin/` — `chauf` binary and shims
  - `bin/shims/` — wrappers (empty initially)
  - `config/` — created empty (init fills later)
  - `projects/`, `php/`
  - `nginx/{bin,etc,sites-available,sites-enabled,conf.d}`
  - `caddy/{bin}`

## Artifact Acquisition & Verification
- Source of truth: release manifest JSON per version & arch:
  - Contains: artifact URL, SHA256, optional minisig/ed25519 signature URL.
- Script flow:
  1. Resolve version/channel → choose artifact for `linux-<arch>`.
  2. Download to temp file.
  3. Verify checksum (and signature when available). **Fail closed** on mismatch.
  4. Place to `${PREFIX}/bin/chauf` with mode `0755`.

## PATH Management Policy
- Add `export PATH="$HOME/.chauffeur/bin:$PATH"` to `~/.bashrc` or `~/.zshrc` **only if missing** (exact string check).
- Respect `--no-path` to skip.
- For other shells, print a manual instruction line.

## Idempotency & Rollback
- If `${PREFIX}` exists, do not clobber files unless updating `chauf` atomically via rename.
- If verification fails, **do not** modify the existing install.
- Temporary files cleaned on exit (trap `INT`, `TERM`, `EXIT`).

## Logging & Telemetry
- Human-friendly logs to stdout; errors to stderr.
- **No telemetry** during install.

## Error Handling (must-have cases)
- Missing tools (`curl`/`wget`, `tar`, checksum tool) → actionable error with package names to install.
- Permission denied writing under `$HOME` → clear message and exit non-zero.
- Unsupported arch/OS → print matrix and stop.
- Network failure → retry suggestion (no partial state left behind).

## Security Considerations
- HTTPS-only downloads; optional signature verification when release provides `.sig` + public key embedded in script.
- Avoid executing remote content beyond the verified binary.
- Do not `sudo` automatically. If user chooses system-wide install later, provide explicit commands separately.

## Acceptance Criteria
1. Running `install.sh` twice leaves a valid, single copy of `chauf` and no duplicate PATH lines.
2. Checksum verification prevents tampered or incomplete downloads.
3. Fresh shell has `~/.chauffeur/bin` on `PATH` unless `--no-path` is used.
4. No writes outside `${HOME}/.chauffeur`.

## Test Matrix (minimum)
- Distros: Arch, Ubuntu LTS, Fedora
- Shells: bash, zsh
- Arches: x86_64 (arm64 later)
- Scenarios: fresh install, re-install, offline (failure), missing curl (with wget fallback), checksum mismatch.

## Open Questions
- Signature scheme: minisign vs ed25519 with `age-keygen`? (decide in ADR)
- Where to host release manifest? (GitHub Releases vs CDN)
- Self-update command (`chauf self-update`) vs re-running installer?

## Next step
Implement **step-2 — Initialize services** specs (nginx, caddy, php installers) with similar rigor: sources, checksums, shims, and minimal configs.

