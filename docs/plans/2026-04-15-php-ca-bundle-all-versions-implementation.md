# PHP CA Bundle All Versions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Modify `doctorPHPCertBundle()` to check ALL installed PHP versions for missing `openssl.cafile`/`curl.cainfo` settings, and provide an aggregated fix command.

**Architecture:** Split `doctorPHPCertBundle()` into a helper that checks a single PHP version, then iterate over all installed versions. Return `[]checkResult` per version instead of a single struct.

**Tech Stack:** Go (chauffeur CLI), PHP shim management

---

## Task 1: Read existing `doctorPHPCertBundle()` function

**Files:**
- Modify: `internal/commands/doctor.go:341-488`

**Step 1: Read current implementation**

Run: `cat internal/commands/doctor.go | sed -n '341,488p'`

Read and understand:
- How `phpPath` is determined via `LookPath`
- How `phpIniFile` is extracted from `php --ini`
- How `curl.cainfo` and `openssl.cafile` are parsed from `php -i`
- How the fix command is generated
- How results are returned

---

## Task 2: Create `checkSinglePHPVersionCerts` helper function

**Files:**
- Modify: `internal/commands/doctor.go` (add after line 340)

**Step 1: Add helper function structure**

Add this function before `doctorPHPCertBundle()`:

```go
// checkSinglePHPVersionCerts checks curl.cainfo and openssl.cafile for a specific PHP binary.
func checkSinglePHPVersionCerts(phpBin string) struct {
    ok         bool
    warn       bool
    status     string
    fix        string
    needsFix   bool
    phpIniFile string
    version    string // PHP version string like "8.3"
} {
    ret := struct {
        ok         bool
        warn       bool
        status     string
        fix        string
        needsFix   bool
        phpIniFile string
        version    string
    }{warn: true, needsFix: true}

    // Get php -i output for version and settings
    infoOut := cmdOutput(phpBin, "-i")
    
    // Parse PHP version from "PHP Version => 8.3.30"
    for _, line := range strings.Split(infoOut, "\n") {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "PHP Version =>") {
            parts := strings.Split(trimmed, "=>")
            if len(parts) >= 2 {
                ret.version = strings.TrimSpace(parts[1])
            }
            break
        }
    }

    // Get loaded php.ini
    iniOut := cmdOutput(phpBin, "--ini")
    for _, line := range strings.Split(iniOut, "\n") {
        if strings.Contains(line, "Loaded Configuration") {
            parts := strings.Split(line, ":")
            if len(parts) >= 2 {
                ret.phpIniFile = strings.TrimSpace(parts[len(parts)-1])
            }
            break
        }
    }

    // Parse curl.cainfo and openssl.cafile
    parsePHPValue := func(raw string) string {
        val := strings.TrimSpace(raw)
        if val == "" || strings.EqualFold(val, "no value") {
            return ""
        }
        return val
    }

    var curlCA, opensslCA string
    for _, line := range strings.Split(infoOut, "\n") {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "curl.cainfo =>") {
            parts := strings.Split(trimmed, "=>")
            if len(parts) >= 2 {
                curlCA = parsePHPValue(parts[1])
            }
        }
        if strings.HasPrefix(trimmed, "openssl.cafile =>") {
            parts := strings.Split(trimmed, "=>")
            if len(parts) >= 2 {
                opensslCA = parsePHPValue(parts[1])
            }
        }
    }

    // Determine effective CA bundle
    caBundle := curlCA
    if caBundle == "" {
        caBundle = opensslCA
    }

    // Build status
    var details []string
    if curlCA != "" {
        details = append(details, "curl.cainfo="+shortenHome(curlCA))
    }
    if opensslCA != "" && opensslCA != curlCA {
        details = append(details, "openssl.cafile="+shortenHome(opensslCA))
    }

    if caBundle == "" {
        ret.status = "not configured"
        if ret.phpIniFile != "" && ret.phpIniFile != "none" {
            dist := detectDistroType()
            var caPath string
            switch dist {
            case distroArch, distroDebian:
                caPath = "/etc/ssl/certs/ca-certificates.crt"
            case distroFedora:
                caPath = "/etc/pki/tls/certs/ca-bundle.crt"
            default:
                caPath = "/etc/ssl/certs/ca-certificates.crt"
            }
            ret.fix = fmt.Sprintf("printf 'curl.cainfo=%s\\nopenssl.cafile=%s\\n' >> %s", caPath, caPath, ret.phpIniFile)
        } else {
            ret.fix = "Configure curl.cainfo and openssl.cafile in php.ini"
        }
        return ret
    }

    // Check if bundle exists
    info, err := os.Stat(caBundle)
    if err != nil {
        ret.status = fmt.Sprintf("%s (file missing)", shortenHome(caBundle))
        return ret
    }
    if info.IsDir() {
        ret.status = fmt.Sprintf("%s (is a directory)", shortenHome(caBundle))
        return ret
    }

    // Check mkcert CA
    mkcertInstalled := projects.MkcertCAInstalled()
    if !mkcertInstalled {
        ret.ok = true
        ret.warn = true
        ret.needsFix = false
        ret.status = strings.Join(details, "  ") + fmt.Sprintf("  %s", shortenHome(caBundle))
        return ret
    }

    data, _ := os.ReadFile(caBundle)
    if strings.Contains(string(data), "mkcert") {
        ret.ok = true
        ret.needsFix = false
        ret.status = strings.Join(details, "  ") + fmt.Sprintf("  %s (mkcert OK)", shortenHome(caBundle))
        return ret
    }

    ret.status = fmt.Sprintf("mkcert CA not in bundle (%s)", shortenHome(caBundle))
    dist := detectDistroType()
    var suggested string
    switch dist {
    case distroArch, distroDebian:
        suggested = "/etc/ssl/certs/ca-certificates.crt"
    case distroFedora:
        suggested = "/etc/pki/tls/certs/ca-bundle.crt"
    default:
        suggested = caBundle
    }
    if suggested != caBundle && ret.phpIniFile != "" {
        ret.fix = fmt.Sprintf("printf 'curl.cainfo=%s\\nopenssl.cafile=%s\\n' | tee -a %s", suggested, suggested, ret.phpIniFile)
    }
    return ret
}
```

**Step 2: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

Expected: No errors (or only pre-existing errors)

---

## Task 3: Modify `doctorPHPCertBundle()` to iterate all versions

**Files:**
- Modify: `internal/commands/doctor.go:341-488` (replace existing function)

**Step 1: Replace the function**

Replace the existing `doctorPHPCertBundle()` function with:

```go
// doctorPHPCertBundle checks ALL installed PHP versions for curl.cainfo and
// openssl.cafile configuration. Returns one checkResult per PHP version.
func doctorPHPCertBundle(root string) []checkResult {
    var results []checkResult

    versions := installedPHPVersions(root)
    if len(versions) == 0 {
        return []checkResult{{
            name:   "PHP CA bundle",
            ok:     false,
            warn:   true,
            status: "no PHP versions installed",
            fix:    "Install PHP: chauf install php 8.3",
        }}
    }

    var configured, total int
    var allFixes []string

    for _, ver := range versions {
        phpBin := filepath.Join(root, "php", ver, "bin", "php")
        
        // Check if binary exists
        if _, err := os.Stat(phpBin); err != nil {
            results = append(results, checkResult{
                name:   fmt.Sprintf("PHP %s CA bundle", ver),
                ok:     false,
                warn:   true,
                status: "not installed",
            })
            continue
        }

        total++
        check := checkSinglePHPVersionCerts(phpBin)

        if check.ok && !check.needsFix {
            configured++
        }

        // Build fix command with full path
        if check.needsFix && check.fix != "" {
            // The fix references phpIniFile which checkSinglePHPVersionCerts already computed
            // We need to pass the full phpIniFile path
            allFixes = append(allFixes, fmt.Sprintf("printf 'curl.cainfo=/etc/ssl/certs/ca-certificates.crt\\nopenssl.cafile=/etc/ssl/certs/ca-certificates.crt\\n' >> %s", filepath.Join(root, "php", ver, "etc", "php.ini")))
        }

        results = append(results, checkResult{
            name: fmt.Sprintf("PHP %s CA bundle", ver),
            ok:   check.ok,
            warn: check.warn,
            status: func() string {
                if check.phpIniFile != "" {
                    return check.status + " (" + shortenHome(check.phpIniFile) + ")"
                }
                return check.status
            }(),
            fix:        check.fix,
            skipAutoFix: !check.needsFix,
        })
    }

    // Add summary result
    summaryStatus := fmt.Sprintf("%d of %d configured", configured, total)
    if configured == total && total > 0 {
        summaryStatus += " — all OK"
    } else if configured < total && configured > 0 {
        summaryStatus += fmt.Sprintf(" — %d need fixing", total-configured)
    }

    // Build aggregated fix command
    var aggregatedFix string
    if len(allFixes) > 0 {
        aggregatedFix = "# Fix all PHP versions:\n" + strings.Join(allFixes, "\n")
    }

    results = append([]checkResult{{
        name:        "PHP CA bundle",
        ok:          configured == total && total > 0,
        warn:        configured < total,
        status:      summaryStatus,
        fix:         aggregatedFix,
        skipAutoFix: len(allFixes) == 0,
    }}, results...)

    return results
}
```

**Step 2: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

---

## Task 4: Update call site in `doctorSSL()`

**Files:**
- Modify: `internal/commands/doctor.go:303-318`

**Step 1: Update the call**

Replace lines 303-318:
```go
// PHP CA bundle (curl.cainfo / openssl.cafile)
if phpini := doctorPHPCertBundle(); phpini.ok {
    results = append(results, checkResult{
        name:   "PHP CA bundle",
        ok:     true,
        status: lib.Gray(phpini.status),
    })
} else {
    results = append(results, checkResult{
        name:   "PHP CA bundle",
        ok:     false,
        warn:   phpini.warn,
        status: phpini.status,
        fix:    phpini.fix,
    })
}
```

With:
```go
// PHP CA bundle (curl.cainfo / openssl.cafile) - checks ALL versions
phpResults := doctorPHPCertBundle(root)
results = append(results, phpResults...)
```

**Step 2: Verify Go syntax**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./internal/commands/`

---

## Task 5: Build and test

**Step 1: Build entire project**

Run: `cd /home/siegg/Workspaces/Personal/Projects/chauffeur-v2 && go build ./...`

Expected: No errors

**Step 2: Run doctor with SSL check**

Run: `./chauf doctor --check-ssl`

Expected: Output shows per-PHP-version CA bundle status

**Step 3: Run with --fix flag**

Run: `./chauf doctor --check-ssl --fix`

Expected: Shows aggregated fix command for all unconfigured versions

---

## Task 6: Commit

**Step 1: Stage and commit**

```bash
git add internal/commands/doctor.go
git commit -m "feat(doctor): check all PHP versions for CA bundle config

- Split doctorPHPCertBundle into checkSinglePHPVersionCerts helper
- Iterate all installed PHP versions instead of just LookPath('php')
- Return aggregated fix command for all unconfigured versions
- Fixes openssl.cafile detection when switching between PHP shims"
```
