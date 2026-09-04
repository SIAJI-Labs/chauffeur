package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NginxRoute struct {
	ServerName   string
	DocumentRoot string
	Upstream     string
	ProxyPort    int
	DatabaseHost string
	SSL          bool
	CertName     string
}

// ProjectSpec is the runtime-neutral project information needed to assemble a
// workspace. The command layer supplies facts; naming, mounts, and upstream
// selection stay owned by the runtime.
type ProjectSpec struct {
	Slug         string
	Path         string
	Version      string
	Domains      []string
	Dedicated    bool
	SSL          bool
	DocumentRoot string
	CertName     string
	ProjectType  string
	ProxyPort    int
}

// BuildWorkspaceScope resolves stable container names, workspace-local mount
// paths, and nginx upstreams without creating any resources.
func BuildWorkspaceScope(workspace, configPath, certificateDir string, httpPort, httpsPort int, projects []ProjectSpec) (WorkspaceScope, error) {
	if workspace == "" || configPath == "" || len(projects) == 0 {
		return WorkspaceScope{}, fmt.Errorf("invalid runtime workspace scope")
	}
	containers := make([]PHPContainerScope, 0)
	containerIndex := make(map[string]int)
	roots := make(map[string]string)
	routes := make([]NginxRoute, 0, len(projects))
	for _, project := range projects {
		if project.Path == "" || project.Slug == "" || len(project.Domains) == 0 || (project.ProjectType != "reverse-proxy" && project.Version == "") {
			return WorkspaceScope{}, fmt.Errorf("invalid runtime project scope for %q", project.Slug)
		}
		if project.ProjectType == "reverse-proxy" {
			proxyPort := project.ProxyPort
			if proxyPort == 0 {
				proxyPort = 3000
			}
			if proxyPort < 1 || proxyPort > 65535 {
				return WorkspaceScope{}, fmt.Errorf("invalid reverse proxy port for %q", project.Slug)
			}
			routes = append(routes, NginxRoute{ServerName: strings.Join(project.Domains, " "), Upstream: "host.containers.internal", ProxyPort: proxyPort, SSL: project.SSL, CertName: project.CertName})
			continue
		}
		name := FPMContainerName(project.Version)
		mountPath := "/workspace/" + project.Slug
		if project.Dedicated {
			name = DedicatedContainerName(project.Slug)
			// Every project gets a stable, collision-free path. Dedicated
			// containers still mount only their own project, while nginx can
			// safely aggregate several dedicated projects in one workspace.
			mountPath = "/workspace/" + project.Slug
		}
		index, exists := containerIndex[name]
		if !exists {
			index = len(containers)
			containerIndex[name] = index
			containers = append(containers, PHPContainerScope{
				Scope: Scope{Workspace: workspace, Project: project.Path, Slug: project.Slug, Version: project.Version, Dedicated: project.Dedicated},
				Image: PHPImage(project.Version), Roots: make(map[string]string),
			})
		}
		containers[index].Roots[mountPath] = project.Path
		roots[mountPath] = project.Path
		documentRoot := mountPath
		if project.DocumentRoot != "" {
			if relative, err := filepath.Rel(project.Path, project.DocumentRoot); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				documentRoot = filepath.Join(mountPath, relative)
			}
		}
		databaseHost := ""
		if host, ok := loopbackDatabaseHost(project.Path); ok {
			databaseHost = host
		}
		routes = append(routes, NginxRoute{ServerName: strings.Join(project.Domains, " "), DocumentRoot: documentRoot, Upstream: name, DatabaseHost: databaseHost, SSL: project.SSL, CertName: project.CertName})
	}
	return WorkspaceScope{Workspace: workspace, ConfigPath: configPath, CertificateDir: certificateDir, HTTPPort: httpPort, HTTPSPort: httpsPort, Roots: roots, PHPContainers: containers, Routes: routes}, nil
}

func loopbackDatabaseHost(projectPath string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(projectPath, ".env"))
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "DB_HOST=") {
			continue
		}
		host := strings.Trim(strings.TrimPrefix(line, "DB_HOST="), `"'`)
		if host == "127.0.0.1" || host == "localhost" {
			return "host.containers.internal", true
		}
		break
	}
	return "", false
}

// RenderNginxPHPConfig returns the minimal container-local route used by the
// runtime integration path. Project linking remains responsible for the full
// domain, alias, SSL, and application-specific configuration.
func RenderNginxPHPConfig(documentRoot, upstream string, listenPort int) (string, error) {
	return RenderNginxPHPConfigForDomain(documentRoot, "_", upstream, listenPort)
}

func RenderNginxPHPConfigForDomain(documentRoot, serverName, upstream string, listenPort int) (string, error) {
	return RenderNginxPHPConfigForRoutes([]NginxRoute{{ServerName: serverName, DocumentRoot: documentRoot, Upstream: upstream}}, listenPort)
}

func RenderNginxPHPConfigForRoutes(routes []NginxRoute, listenPort int) (string, error) {
	return RenderNginxPHPConfigForRoutesWithHTTPS(routes, listenPort, 0)
}

func RenderNginxPHPConfigForRoutesWithHTTPS(routes []NginxRoute, httpPort, httpsPort int) (string, error) {
	if len(routes) == 0 || httpPort < 1 || httpPort > 65535 {
		return "", fmt.Errorf("invalid nginx routes or port")
	}
	if httpsPort < 0 || httpsPort > 65535 {
		return "", fmt.Errorf("invalid nginx HTTPS port")
	}
	var b strings.Builder
	for _, route := range routes {
		if route.ServerName == "" || route.Upstream == "" {
			return "", fmt.Errorf("invalid nginx route")
		}
		if route.ProxyPort == 0 && (route.DocumentRoot == "" || !filepath.IsAbs(route.DocumentRoot)) {
			return "", fmt.Errorf("invalid nginx route")
		}
		if route.ProxyPort != 0 && (route.ProxyPort < 1 || route.ProxyPort > 65535) {
			return "", fmt.Errorf("invalid reverse proxy port")
		}
		if route.SSL && (httpsPort == 0 || route.CertName == "") {
			return "", fmt.Errorf("SSL route %s is missing HTTPS port or certificate", route.ServerName)
		}
		renderServer := func(listen, ssl string) {
			if route.ProxyPort != 0 {
				fmt.Fprintf(&b, `server {
  %s
    server_name %s;
%s
    location / {
        proxy_pass http://%s:%d;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`, listen, route.ServerName, ssl, route.Upstream, route.ProxyPort)
				return
			}
			fmt.Fprintf(&b, `server {
  %s
    server_name %s;
    root %s;
%s
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include fastcgi_params;
%s
        fastcgi_param SCRIPT_FILENAME %s$fastcgi_script_name;
        fastcgi_pass %s:9000;
    }
 }

`, listen, route.ServerName, route.DocumentRoot, ssl, fastcgiDatabaseParam(route.DatabaseHost), route.DocumentRoot, route.Upstream)
		}
		if route.SSL {
			fmt.Fprintf(&b, `server {
    listen %d;
    server_name %s;
    return 301 https://$host:%d$request_uri;
}

`, httpPort, route.ServerName, httpsPort)
			renderServer(fmt.Sprintf("    listen %d ssl;", httpsPort), fmt.Sprintf("    ssl_certificate /etc/nginx/certs/%s.crt;\n    ssl_certificate_key /etc/nginx/certs/%s.key;", route.CertName, route.CertName))
		} else {
			renderServer(fmt.Sprintf("    listen %d;", httpPort), "")
		}
	}
	return b.String(), nil
}

func fastcgiDatabaseParam(host string) string {
	if host == "" {
		return ""
	}
	return "        fastcgi_param DB_HOST " + host + ";"
}

func legacyNginxPHPConfig(documentRoot, serverName, upstream string, listenPort int) (string, error) {
	if documentRoot == "" || filepath.IsAbs(documentRoot) == false {
		return "", fmt.Errorf("document root must be an absolute path")
	}
	if serverName == "" || upstream == "" || listenPort < 1 || listenPort > 65535 {
		return "", fmt.Errorf("invalid nginx upstream or port")
	}
	return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME %s$fastcgi_script_name;
        fastcgi_pass %s:9000;
    }
}
`, listenPort, serverName, documentRoot, documentRoot, upstream), nil
}
