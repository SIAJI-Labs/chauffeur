# Design: Fix `openssl.cafile` Across All PHP Versions

**Date:** 2026-04-15
**Status:** Approved

## Problem

The `doctorPHPCertBundle()` function in `internal/commands/doctor.go` only checks the currently active PHP version (via `exec.LookPath("php")`). When multiple PHP versions are installed via chauffeur shims, some versions may be missing the `openssl.cafile` and `curl.cainfo` settings in their `php.ini` files.

This causes issues when switching between PHP versions - for example, PHP 8.3 has the CA bundle configured correctly, but PHP 7.4 does not.

## Solution

Modify `doctorPHPCertBundle()` to check **all installed PHP versions** and provide a unified fix that updates all php.ini files at once.

## Implementation Details

### 1. New Helper Function

```go
// checkSinglePHPVersionCerts checks curl.cainfo and openssl.cafile for a specific PHP binary.
// Returns a struct with ok, status, and per-version details.
func checkSinglePHPVersionCerts(phpBin string) struct {
    ok         bool
    status     string
    curlCA     string
    opensslCA  string
    needsFix   bool
    phpIniFile string
}
```

This function:
- Takes an explicit PHP binary path
- Runs `phpBin --ini` to find the loaded php.ini
- Runs `phpBin -i` to parse `curl.cainfo` and `openssl.cafile`
- Returns whether the settings are properly configured

### 2. Modify `doctorPHPCertBundle()`

**Current signature:**
```go
func doctorPHPCertBundle() struct { ... }
```

**New signature:**
```go
func doctorPHPCertBundle(root string) []checkResult
```

Changes:
- Accept `root` parameter to locate PHP installations
- Use `installedPHPVersions(root)` to get all PHP versions
- For each version, call `checkSinglePHPVersionCerts()`
- Aggregate results into a `[]checkResult`

### 3. Update `doctorSSL()` Call Site

**Current (line 304):**
```go
if phpini := doctorPHPCertBundle(); phpini.ok {
```

**New:**
```go
if phpini := doctorPHPCertBundle(root); len(phpini) == 0 || phpiniAllOk(phpini) {
```

The function now returns multiple `checkResult` entries - one per PHP version.

### 4. Fix Command Format

When multiple versions need fixing, the fix command aggregates all:

```bash
# Single version
printf 'curl.cainfo=/etc/ssl/certs/ca-certificates.crt\nopenssl.cafile=/etc/ssl/certs/ca-certificates.crt\n' >> ~/.chauffeur/php/7.4/etc/php.ini

# All versions at once
for ver in 7.4 8.1 8.3; do
  printf 'curl.cainfo=/etc/ssl/certs/ca-certificates.crt\nopenssl.cafile=/etc/ssl/certs/ca-certificates.crt\n' >> ~/.chauffeur/php/$ver/etc/php.ini
done
```

The `auto-fix` logic will iterate each result with `needsFix=true` and apply fixes.

### 5. Result Display

```
SSL:
  ✓ openssl 3.x.x
  ✓ mkcert v1.1.x
  ⚠ PHP CA bundle (2 of 3 configured)
    - 8.3: curl.cainfo=✓  openssl.cafile=✓
    - 8.1: curl.cainfo=✓  openssl.cafile=✓
    - 7.4: curl.cainfo=✗  openssl.cafile=✗  ← MISSING
  fix: for ver in 7.4; do printf 'curl.cainfo=...\nopenssl.cafile=...\n' >> $CHAUF_HOME/php/$ver/etc/php.ini; done
```

## Files Affected

- `internal/commands/doctor.go` - Main changes to `doctorPHPCertBundle()` and helper function

## Testing

1. Run `chauf doctor --check-ssl` with multiple PHP versions
2. Verify output shows all versions
3. Test `--fix` flag shows correct command
4. Test `--auto-fix` applies to all versions

## Distro Detection

The fix command uses distro-specific CA bundle paths (already implemented in existing code):
- Arch/Debian: `/etc/ssl/certs/ca-certificates.crt`
- Fedora: `/etc/pki/tls/certs/ca-bundle.crt`
