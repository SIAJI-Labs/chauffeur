# Multi-Domain Specification

## Overview

Each Chauffeur project has one primary domain and can have zero or more alias domains. All domains route to the same project directory.

**Use cases**:
- Multi-tenant apps (one codebase, multiple brand domains)
- Admin panels on a separate subdomain (`admin.project.test`)
- API endpoints on a separate subdomain (`api.project.test`)
- White-label development

---

## Domain Structure

```
Primary:  my-project.test
Aliases:  admin.my-project.test
          api.my-project.test
```

All domains are stored in project config:

```yaml
domain: my-project.test
aliases:
  - admin.my-project.test
  - api.my-project.test
```

---

## Adding Alias Domains

### At Link Time

```bash
chauf link --alias admin.my-project.test --alias api.my-project.test
```

### After Linking (Dynamic Add)

```bash
# In the project directory:
chauf link --alias reporting.my-project.test
```

This adds the alias to the existing project without relinking.

**Flow**:
1. Validate the alias domain format
2. Check alias not already in use by another project
3. Add alias to `config.yaml`
4. Regenerate nginx site config with new `server_name`
5. If SSL: regenerate SAN cert including new alias
6. Reload nginx

---

## Removing Alias Domains

```bash
chauf unlink --alias admin.my-project.test
```

**Flow**:
1. Remove alias from `config.yaml`
2. Regenerate nginx site config without the alias
3. If SSL: regenerate SAN cert without the alias
4. Reload nginx

To remove all aliases:
```bash
chauf unlink --all  # Removes all aliases, then unlinks the project
```

---

## `chauf links` Display

```
 Project        Domain                    PHP    SSL
─────────────────────────────────────────────────────
 my-project     my-project.test           8.3    HTTP
                admin.my-project.test           HTTP (*)
                api.my-project.test             HTTP (*)
```

`(*)` marks alias domains.

---

## nginx Config (Multi-Domain)

All domains are listed in a single `server_name` directive:

```nginx
server {
    listen 8080;
    server_name my-project.test admin.my-project.test api.my-project.test;

    root /home/user/Projects/my-project/public;
    # ...
}
```

For SSL:

```nginx
server {
    listen 8080;
    server_name my-project.test admin.my-project.test api.my-project.test;
    return 301 https://$host:8443$request_uri;
}

server {
    listen 8443 ssl;
    http2 on;
    server_name my-project.test admin.my-project.test api.my-project.test;

    ssl_certificate     .../certs/my-project.test.crt;  # SAN cert covers all
    ssl_certificate_key .../certs/my-project.test.key;
    # ...
}
```

---

## Domain Validation

Valid alias domains must:
- Match pattern: `^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*\.test$`
- Not be in use by another project (check all project configs)
- Not be the primary domain of another project
- Not be identical to the project's own primary domain

```go
// internal/lib/input.go

var domainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*\.test$`)

func IsValidDomain(domain string) bool {
    return domainPattern.MatchString(domain)
}
```

---

## Conflict Detection

Before adding a domain (primary or alias), check all project configs:

```go
// internal/projects/manager.go

func (m *ProjectManager) IsDomainInUse(domain string) (string, bool) {
    projects, _ := m.ListAll()
    for _, p := range projects {
        if p.Domain == domain {
            return p.Slug, true
        }
        for _, alias := range p.Aliases {
            if alias == domain {
                return p.Slug, true
            }
        }
    }
    return "", false
}
```

If conflict found:
```
Error: domain "admin.my-project.test" is already used by project "another-project"

Remove it first:
  chauf unlink --alias admin.my-project.test
  (in the another-project directory)
```

---

## SSL + Multi-Domain

A single SAN certificate covers all domains (primary + aliases). Certificate is regenerated automatically when aliases change.

See [ssl.md](./ssl.md) for full SSL specification.
