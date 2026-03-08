# Naming Conventions

## File Naming

| Type | Pattern | Example |
|------|---------|---------|
| Command handlers | `commands/<verb>.go` | `commands/link.go` |
| Internal packages | `internal/<domain>/<concern>.go` | `internal/projects/manager.go` |
| Installers | `installers/<service>.go` | `installers/php.go` |
| Library files | `lib/<concern>.go` | `lib/logging.go` |
| Templates | `templates/files/<name>.tmpl` | `templates/files/nginx-site.conf.tmpl` |
| Test files | `<filename>_test.go` | `manager_test.go` |

## Go Identifiers

| Type | Convention | Example |
|------|-----------|---------|
| Packages | lowercase, short, no underscores | `workspace`, `projects`, `lib` |
| Exported types | PascalCase | `ProjectManager`, `PHPInstaller` |
| Unexported types | camelCase | `projectConfig`, `nginxBuilder` |
| Exported functions | PascalCase | `NewProjectManager`, `RunLink` |
| Unexported functions | camelCase | `downloadSource`, `parseVersion` |
| Constants | `ALL_CAPS` or PascalCase (exported) | `DefaultPHPVersion`, `MaxAliases` |
| Interfaces | PascalCase, -er suffix | `Executor`, `Installer`, `Logger` |
| Error variables | `Err`-prefixed | `ErrProjectNotFound`, `ErrVersionUnsupported` |

## Command Naming

CLI commands use kebab-case for multi-word commands:

```
chauf self-update        ✅
chauf selfupdate         ❌
chauf php isolate        ✅  (subcommand, space-separated)
chauf php-isolate        ❌
```

Flags use `--long-form` with hyphens:

```
chauf link --dedicated-fpm    ✅
chauf link --dedicatedFpm     ❌
chauf link --dedicated_fpm    ❌
```

## Project Slugs

Project slugs are derived from the directory name:

- Lowercase
- Hyphens replace spaces and underscores
- Remove special characters
- Max 50 characters

```
~/Projects/my-app          → slug: my-app
~/Projects/My Laravel App  → slug: my-laravel-app
~/Projects/app_v2          → slug: app-v2
```

## Domain Naming

- Always `.test` TLD for local development
- Slug becomes subdomain: `<slug>.test`
- Aliases follow same pattern: `admin.<slug>.test`

```
my-app          → my-app.test
admin panel     → admin-panel.test
```

## Config Keys (YAML)

Use `snake_case` for all YAML config keys:

```yaml
# CORRECT
php_version: "8.3"
dedicated_fpm: false
http_port: 8080
created_at: "2025-01-01T00:00:00Z"

# WRONG
phpVersion: "8.3"
dedicatedFpm: false
```

## Workspace Paths

Path functions in `internal/workspace/paths.go` follow `<Service><Component>` naming:

```go
func (w *Workspace) PHPDir(version string) string         // ~/.chauffeur/php/<version>/
func (w *Workspace) PHPBinary(version string) string      // ~/.chauffeur/php/<version>/bin/php
func (w *Workspace) PHPFPMBinary(version string) string   // ~/.chauffeur/php/<version>/bin/php-fpm
func (w *Workspace) PHPFPMConfig(version string) string   // shared pool config
func (w *Workspace) NginxBinary() string                  // ~/.chauffeur/nginx/bin/nginx
func (w *Workspace) NginxConfig() string                  // ~/.chauffeur/nginx/etc/nginx.conf
func (w *Workspace) NginxSiteConfig(slug string) string   // sites-available/<slug>.conf
func (w *Workspace) NginxSiteLink(slug string) string     // sites-enabled/<slug>.conf
func (w *Workspace) NginxCert(domain string) string       // certs/<domain>.crt
func (w *Workspace) NginxKey(domain string) string        // certs/<domain>.key
func (w *Workspace) ProjectDir(slug string) string        // ~/.chauffeur/projects/<slug>/
func (w *Workspace) ProjectConfig(slug string) string     // ~/.chauffeur/projects/<slug>/config.yaml
func (w *Workspace) ComposerPHAR() string                 // ~/.chauffeur/composer/composer.phar
func (w *Workspace) PHPShim() string                      // ~/.chauffeur/bin/shims/php
func (w *Workspace) ComposerShim() string                 // ~/.chauffeur/bin/shims/composer
func (w *Workspace) CacheDir() string                     // ~/.chauffeur/cache/
```

## Logger Names

Logger names match the command or package they belong to:

```go
lib.NewCommandLogger("link")        // for commands/link.go
lib.NewCommandLogger("install")     // for commands/install.go
lib.NewLogger("projects")           // for internal/projects/
lib.NewLogger("php-installer")      // for internal/installers/php.go
lib.NewLogger("nginx")              // for internal/services/nginx.go
```

## Template Variables

Go templates use PascalCase for all template variables:

```
{{.Domain}}
{{.PHPVersion}}
{{.FPMSocket}}
{{.WorkspaceRoot}}
{{.ProjectSlug}}
{{.SSLCert}}
{{.SSLKey}}
```

## Environment Variables

Chauffeur environment variables use `CHAUFFEUR_` prefix, all caps:

```
CHAUFFEUR_HOME              # Override workspace root (default: ~/.chauffeur)
CHAUFFEUR_PHP_TARBALL       # Override PHP source tarball path (offline install)
CHAUFFEUR_PHP_SIGNATURE     # Override PHP signature file path
CHAUFFEUR_PHP_KEYRING       # Override PHP keyring path
CHAUFFEUR_KEEP_BUILD_DIR    # Set to 1 to keep build dirs for debugging
CHAUFFEUR_LOG_LEVEL         # Override log level (debug, info, warn, error)
```

## Error Messages

Error messages should be:
- Lowercase (no capital first letter, no trailing period)
- Actionable (tell the user what went wrong and what to do)
- Prefixed with package name for wrapped errors

```go
// CORRECT
fmt.Errorf("php install: failed to download PHP %s: %w", version, err)
fmt.Errorf("link: project already linked at %s", existingPath)
fmt.Errorf("invalid php version: %q (supported: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4)", version)

// WRONG
fmt.Errorf("Error: PHP installation failed")
fmt.Errorf("FAILED TO DOWNLOAD")
fmt.Errorf("Something went wrong.")
```
