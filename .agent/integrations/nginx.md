# Nginx Integration

## Overview

Chauffeur compiles nginx from source into `~/.chauffeur/nginx/`. A single nginx instance serves all registered projects, routing by `Host` header to the correct project's PHP-FPM pool.

---

## Build Configuration

```bash
./configure \
  --prefix=$HOME/.chauffeur/nginx \
  --sbin-path=$HOME/.chauffeur/nginx/bin/nginx \
  --conf-path=$HOME/.chauffeur/nginx/etc/nginx.conf \
  --http-log-path=$HOME/.chauffeur/nginx/logs/access.log \
  --error-log-path=$HOME/.chauffeur/nginx/logs/error.log \
  --pid-path=$HOME/.chauffeur/nginx/logs/nginx.pid \
  --with-http_ssl_module \
  --with-http_gzip_static_module \
  --with-http_v2_module \
  --with-http_fastcgi_module
```

---

## Directory Layout

```
~/.chauffeur/nginx/
├── bin/
│   └── nginx
├── etc/
│   ├── nginx.conf              # Main config (includes sites-enabled/*)
│   ├── mime.types
│   ├── fastcgi_params          # Standard FastCGI params for PHP
│   ├── sites-available/
│   │   ├── default.conf        # Catch-all → 404 for unlinked domains
│   │   └── <slug>.conf         # Per-project config
│   └── sites-enabled/
│       └── <slug>.conf -> ../sites-available/<slug>.conf  # Symlink
├── certs/
│   ├── <domain>.crt
│   └── <domain>.key
└── logs/
    ├── access.log
    ├── error.log
    └── nginx.pid
```

---

## nginx.conf Template

```nginx
worker_processes auto;
pid {{ .WorkspaceRoot }}/nginx/logs/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include {{ .WorkspaceRoot }}/nginx/etc/mime.types;
    default_type application/octet-stream;

    sendfile on;
    keepalive_timeout 65;

    include {{ .WorkspaceRoot }}/nginx/etc/sites-enabled/*.conf;
}
```

---

## Site Config Template (HTTP)

Generated at `sites-available/<slug>.conf`:

```nginx
server {
    listen 8080;
    server_name {{ .Domain }}{{ range .Aliases }} {{ . }}{{ end }};

    root {{ .ProjectPath }}/public;
    index index.php index.html;

    access_log {{ .WorkspaceRoot }}/nginx/logs/{{ .Slug }}-access.log;
    error_log  {{ .WorkspaceRoot }}/nginx/logs/{{ .Slug }}-error.log;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:{{ .FPMSocket }};
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include {{ .WorkspaceRoot }}/nginx/etc/fastcgi_params;
    }

    location ~ /\.ht {
        deny all;
    }
}
```

---

## Site Config Template (HTTPS)

Generated at `sites-available/<slug>.conf` when SSL is enabled:

```nginx
server {
    listen 8080;
    server_name {{ .Domain }}{{ range .Aliases }} {{ . }}{{ end }};
    return 301 https://$host:8443$request_uri;
}

server {
    listen 8443 ssl;
    http2 on;
    server_name {{ .Domain }}{{ range .Aliases }} {{ . }}{{ end }};

    ssl_certificate     {{ .WorkspaceRoot }}/nginx/certs/{{ .Domain }}.crt;
    ssl_certificate_key {{ .WorkspaceRoot }}/nginx/certs/{{ .Domain }}.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    root {{ .ProjectPath }}/public;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:{{ .FPMSocket }};
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        fastcgi_param HTTPS on;
        include {{ .WorkspaceRoot }}/nginx/etc/fastcgi_params;
    }
}
```

---

## Default Catch-All Config

Prevents traffic from falling through to other projects when a domain is unlinked:

```nginx
# sites-available/default.conf
server {
    listen 8080 default_server;
    listen 8443 ssl default_server;
    server_name _;

    ssl_certificate     {{ .WorkspaceRoot }}/nginx/certs/default.crt;
    ssl_certificate_key {{ .WorkspaceRoot }}/nginx/certs/default.key;

    return 404;
}
```

---

## Process Management

```go
// internal/services/nginx.go

type NginxService struct {
    binaryPath string
    configPath string
    pidPath    string
}

func (n *NginxService) Start() error {
    if n.IsRunning() {
        return nil  // idempotent
    }
    return exec.Command(n.binaryPath, "-c", n.configPath).Run()
}

func (n *NginxService) Stop() error {
    pid, err := n.readPID()
    if err != nil {
        return nil  // not running, nothing to do
    }
    return syscall.Kill(pid, syscall.SIGQUIT)  // graceful shutdown
}

func (n *NginxService) Reload() error {
    pid, err := n.readPID()
    if err != nil {
        return n.Start()
    }
    return syscall.Kill(pid, syscall.SIGHUP)  // reload config without dropping connections
}

func (n *NginxService) IsRunning() bool {
    pid, err := n.readPID()
    if err != nil {
        return false
    }
    return syscall.Kill(pid, 0) == nil  // signal 0 = check if process exists
}
```

---

## Linking Flow

When `chauf link` runs:

1. Generate site config from template → write to `sites-available/<slug>.conf`
2. Create symlink: `sites-enabled/<slug>.conf` → `../sites-available/<slug>.conf`
3. If nginx is running: send `SIGHUP` to reload config
4. If nginx is not running: no action (sites will load on next `chauf start`)

When `chauf unlink` runs:

1. Remove symlink `sites-enabled/<slug>.conf`
2. Remove config `sites-available/<slug>.conf`
3. If nginx is running: send `SIGHUP` to reload config

---

## Config Validation

Before starting nginx, validate the config:

```bash
~/.chauffeur/nginx/bin/nginx -t -c ~/.chauffeur/nginx/etc/nginx.conf
```

If validation fails, abort start and display the nginx error output.

---

## Port Configuration

Default ports: HTTP `8080`, HTTPS `8443`

Override in `~/.chauffeur/config/chauffeur.yaml`:
```yaml
nginx:
  http_port: 8080
  https_port: 8443
```

**Port conflict handling**: Before starting nginx, check if the configured ports are available. If not, report conflict with PID and process name. Do NOT automatically choose a different port — ask the user to resolve the conflict.

**iptables redirect** (optional, requires sudo):
```bash
# Chauffeur prints these commands — never executes them:
iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-port 8080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-port 8443
```
