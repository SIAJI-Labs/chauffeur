# dnsmasq Integration

## Overview

Chauffeur uses dnsmasq to route `*.test` domains to `127.0.0.1`. dnsmasq is a system package — Chauffeur does NOT install it. Instead, it prints the configuration for the user to apply.

**Zero silent host mutation**: Chauffeur never writes to `/etc/dnsmasq.d/` directly. It always prints the commands for user approval.

---

## Setup Flow

When `chauf init` runs, it checks if dnsmasq is configured and prints instructions if not:

```
DNS configuration required.

Add the following to your dnsmasq config:

  sudo tee /etc/dnsmasq.d/chauffeur.conf << 'EOF'
  address=/.test/127.0.0.1
  EOF

Then restart dnsmasq:
  sudo systemctl restart dnsmasq

Or for NetworkManager-based systems:
  sudo systemctl restart NetworkManager
```

---

## dnsmasq Config

Content of `/etc/dnsmasq.d/chauffeur.conf`:

```
address=/.test/127.0.0.1
```

This routes all `*.test` domains to localhost. nginx then routes by `Host` header to the correct project.

---

## Offline Resilience

### The Problem

When you go offline, **NetworkManager (or DHCP) rewrites `/etc/resolv.conf`**. If it removes `127.0.0.1` as a nameserver, the OS stops querying dnsmasq entirely — `.test` domains stop resolving even though dnsmasq is still running and `address=/.test/127.0.0.1` is still configured.

The dnsmasq `address=` directive is purely local — it never needs internet to resolve `.test`. The problem is the OS never reaches dnsmasq to ask.

**Diagnosis**: Test by disabling Wi-Fi/Ethernet, then running `cat /etc/resolv.conf`. If `127.0.0.1` is missing, that's the problem.

### Fixes (choose one)

All three approaches print commands for the user to run — Chauffeur never executes them.

---

**Option A: NetworkManager dnsmasq plugin** *(recommended for NM-based desktops)*

Makes NM manage dnsmasq directly. NM always sets `127.0.0.1` as the nameserver regardless of network state.

```bash
# Create file (requires sudo once)
sudo tee /etc/NetworkManager/conf.d/chauffeur-dns.conf << 'EOF'
[main]
dns=dnsmasq
EOF

sudo systemctl restart NetworkManager
```

After this, NM starts its own dnsmasq instance. The Chauffeur config file `/etc/dnsmasq.d/chauffeur.conf` is still read by it. `/etc/resolv.conf` will contain `127.0.0.1` permanently.

---

**Option B: systemd-resolved split DNS** *(recommended for systems using systemd-resolved)*

Routes `.test` queries through `127.0.0.1` at the systemd-resolved level, before `/etc/resolv.conf` is involved. Survives all network state changes.

```bash
sudo mkdir -p /etc/systemd/resolved.conf.d/

sudo tee /etc/systemd/resolved.conf.d/chauffeur.conf << 'EOF'
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF

sudo systemctl restart systemd-resolved
```

The `~test` prefix is a routing-only domain: systemd-resolved sends `.test` queries to `127.0.0.1` (dnsmasq) and nothing else changes. Works alongside NetworkManager without conflict.

---

**Option C: Per-project `/etc/hosts` entries** *(most reliable, no dependency on dnsmasq/NM)*

Completely offline-proof. No dnsmasq required.

```
# /etc/hosts
127.0.0.1  my-app.test
127.0.0.1  admin.my-app.test
```

When `chauf link` runs and DNS is not protected, Chauffeur prints the exact lines to add to `/etc/hosts`. `chauf doctor --fix` also shows them per-project.

---

### Which Option to Recommend

| Scenario | Recommended fix |
|----------|----------------|
| Desktop Arch/Fedora with NetworkManager | Option A (NM dnsmasq plugin) |
| Ubuntu/Debian with systemd-resolved | Option B (split DNS) |
| Headless server, no NM | dnsmasq already direct — check `/etc/resolv.conf` contains `127.0.0.1` |
| No dnsmasq at all | Option C (/etc/hosts) |

---

## Distribution-Specific Notes

### Arch Linux

```bash
# Install
sudo pacman -S dnsmasq

# Config location
sudo tee /etc/dnsmasq.d/chauffeur.conf <<< 'address=/.test/127.0.0.1'

# Enable and start
sudo systemctl enable --now dnsmasq

# If using NetworkManager (common on Arch desktops):
# Add to /etc/NetworkManager/conf.d/dns.conf:
# [main]
# dns=dnsmasq
# Then restart NetworkManager instead of dnsmasq directly
```

### Ubuntu/Debian

```bash
# Install
sudo apt-get install dnsmasq

# Note: Ubuntu uses systemd-resolved, may conflict with dnsmasq
# Check: systemctl status systemd-resolved

# Config location
sudo tee /etc/dnsmasq.d/chauffeur.conf <<< 'address=/.test/127.0.0.1'

# Restart
sudo systemctl restart dnsmasq
```

### Fedora/RHEL

```bash
# Install
sudo dnf install dnsmasq

# Config location
sudo tee /etc/dnsmasq.d/chauffeur.conf <<< 'address=/.test/127.0.0.1'

# Enable and start
sudo systemctl enable --now dnsmasq

# Restart NetworkManager to pick up DNS change
sudo systemctl restart NetworkManager
```

---

## DNS Health Checks

`chauf doctor` validates DNS resolution and offline resilience:

1. Check if dnsmasq is installed: `which dnsmasq`
2. Check if dnsmasq is running: `systemctl is-active dnsmasq`
3. Check if config exists: `stat /etc/dnsmasq.d/chauffeur.conf`
4. Test dnsmasq directly: `dig @127.0.0.1 test.test +short` → must return `127.0.0.1`
5. Test end-to-end system resolution: `dig test.test +short` → must return `127.0.0.1`
6. **Offline resilience check**: Does `/etc/resolv.conf` include `127.0.0.1`? Is there a protection mechanism (NM plugin, systemd-resolved split DNS)?

Checks 4 and 5 can pass while online but fail offline — check 6 catches this gap.

### What to check for offline resilience

```go
// internal/system/dns.go

func CheckOfflineResilience() (protected bool, method string) {
    // Option A: NM dnsmasq plugin
    if fileContains("/etc/NetworkManager/conf.d/chauffeur-dns.conf", "dns=dnsmasq") {
        return true, "NetworkManager dnsmasq plugin"
    }
    // Option B: systemd-resolved split DNS
    if dirContainsMatch("/etc/systemd/resolved.conf.d/", "Domains=~test") {
        return true, "systemd-resolved split DNS"
    }
    // Option C: /etc/hosts has entries for all linked projects
    if allProjectsInHostsFile() {
        return true, "/etc/hosts entries"
    }
    // Fallback: /etc/resolv.conf directly lists 127.0.0.1 — may be overwritten by NM
    if fileContains("/etc/resolv.conf", "nameserver 127.0.0.1") {
        return false, "resolv.conf (not protected — may break offline)"
    }
    return false, "none"
}
```

If DNS check fails, doctor outputs:
- What failed and why (dnsmasq down vs. resolv.conf issue vs. no protection)
- The protection option best suited for the detected system (NM vs. systemd-resolved)
- Per-project `/etc/hosts` lines as a fallback

---

## Retry Logic

DNS resolution can be flaky immediately after dnsmasq restart. Chauffeur implements retry with backoff when checking DNS:

```go
// internal/system/dns.go

func CheckDNSResolution(domain string) error {
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        if attempt > 0 {
            time.Sleep(time.Duration(attempt) * time.Second)
        }
        addrs, err := net.LookupHost(domain)
        if err != nil {
            lastErr = err
            continue
        }
        for _, addr := range addrs {
            if addr == "127.0.0.1" {
                return nil
            }
        }
        lastErr = fmt.Errorf("resolved to %v, expected 127.0.0.1", addrs)
    }
    return fmt.Errorf("dns: %w", lastErr)
}
```

---

## Graceful Failure

dnsmasq is **optional**. Chauffeur works without it — domains won't resolve automatically, but users can add entries to `/etc/hosts` manually:

```
# /etc/hosts
127.0.0.1  my-project.test
127.0.0.1  admin.my-project.test
```

When `chauf doctor` detects no dnsmasq, it shows both options (dnsmasq setup or manual /etc/hosts).

When `chauf link` runs and DNS offline resilience is not configured, it prints a warning with the per-project `/etc/hosts` lines — the project still links successfully. See `docs/commands/project.md` for the warning format.

---

## State Tracking

Chauffeur tracks the dnsmasq config state in `~/.chauffeur/system/dns.json`:

```json
{
  "configured": true,
  "config_path": "/etc/dnsmasq.d/chauffeur.conf",
  "tld": "test",
  "last_checked": "2025-01-01T00:00:00Z"
}
```

This allows `chauf info` to report DNS status without re-checking every time.
