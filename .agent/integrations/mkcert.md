# mkcert Integration

## Overview

Chauffeur uses `mkcert` to generate locally-trusted SSL certificates for `.test` domains. `mkcert` installs a root CA into the system trust store, making all generated certs trusted by browsers without manual steps.

---

## Prerequisites

mkcert must be installed on the host system. `chauf doctor` checks for it.

**Installation**:
```bash
# Arch Linux
sudo pacman -S mkcert

# Debian/Ubuntu
sudo apt install mkcert

# Using Go (any distro)
go install filippo.io/mkcert@latest

# Manual download from GitHub
curl -Lo mkcert https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-linux-amd64
chmod +x mkcert && sudo mv mkcert /usr/local/bin/
```

---

## CA Initialization

On first use, the user must install the local CA:

```bash
mkcert -install
```

Chauffeur checks if the CA is installed by looking for mkcert's CAROOT directory. If not installed, `chauf secure` prints the command and asks the user to run it.

---

## Certificate Generation

`chauf secure` calls mkcert to generate a certificate:

```go
// internal/lib/ssl.go

func GenerateCert(workspaceRoot string, domains []string) error {
    certPath := filepath.Join(workspaceRoot, "nginx", "certs", domains[0]+".crt")
    keyPath := filepath.Join(workspaceRoot, "nginx", "certs", domains[0]+".key")

    args := []string{
        "-cert-file", certPath,
        "-key-file", keyPath,
    }
    args = append(args, domains...)

    cmd := exec.Command("mkcert", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("mkcert: %w\n%s", err, output)
    }
    return nil
}
```

**SAN Certificate** (Subject Alternative Name) — for projects with aliases:

```bash
mkcert \
  -cert-file ~/.chauffeur/nginx/certs/my-project.test.crt \
  -key-file  ~/.chauffeur/nginx/certs/my-project.test.key \
  my-project.test admin.my-project.test api.my-project.test
```

One certificate covers the primary domain and all aliases.

---

## Certificate Storage

```
~/.chauffeur/nginx/certs/
├── <primary-domain>.crt    # Certificate (includes all SANs)
├── <primary-domain>.key    # Private key
└── default.crt             # Self-signed for catch-all server block
    default.key
```

**Naming convention**: Always named after the primary domain (not aliases).

---

## Certificate Lifecycle

### `chauf secure`

1. Check mkcert is installed
2. Check mkcert CA is installed (`mkcert -install` if not, with user prompt)
3. Generate SAN cert for primary domain + all aliases
4. Store in `~/.chauffeur/nginx/certs/`
5. Update project config: `ssl: true`
6. Regenerate nginx site config with SSL server block
7. Reload nginx (if running)

### `chauf unsecure`

1. Update project config: `ssl: false`
2. Regenerate nginx site config without SSL server block
3. Reload nginx (if running)
4. Optionally remove cert files (ask user)

### `chauf link --alias <domain>`

When adding an alias to an SSL-enabled project:
1. Regenerate SAN cert including the new alias
2. Store with same filename (overwrite previous cert)
3. Reload nginx

### `chauf unlink --alias <domain>`

When removing an alias from an SSL-enabled project:
1. Regenerate SAN cert excluding the removed alias
2. Store with same filename (overwrite previous cert)
3. Reload nginx

---

## File Permissions

```
certs/<domain>.crt   644   # nginx needs to read this
certs/<domain>.key   600   # only the user should read this
```

Set explicitly after generation:
```go
os.Chmod(certPath, 0644)
os.Chmod(keyPath, 0600)
```

---

## `chauf doctor` SSL Checks

1. `mkcert` binary exists: `which mkcert`
2. mkcert CA installed: check `$(mkcert -CAROOT)` directory exists
3. For each SSL-enabled project:
   - Cert file exists: `stat ~/.chauffeur/nginx/certs/<domain>.crt`
   - Key file exists: `stat ~/.chauffeur/nginx/certs/<domain>.key`
   - Cert covers all current aliases (SAN check): `openssl x509 -in <cert> -text -noout`
   - Cert not expired: check `Not After` date

---

## OpenSSL Verification

Use `openssl` to inspect generated certs:

```bash
# View cert details
openssl x509 -in ~/.chauffeur/nginx/certs/my-project.test.crt -text -noout

# Check SAN entries
openssl x509 -in ~/.chauffeur/nginx/certs/my-project.test.crt -text -noout \
  | grep -A1 "Subject Alternative Name"

# Check expiry
openssl x509 -in ~/.chauffeur/nginx/certs/my-project.test.crt -noout -enddate
```

mkcert certificates are valid for 10 years by default — expiry is rarely an issue in development.
