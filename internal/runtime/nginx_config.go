package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
)

type NginxRoute struct {
	ServerName   string
	DocumentRoot string
	Upstream     string
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
		if project.Path == "" || project.Slug == "" || project.Version == "" || len(project.Domains) == 0 {
			return WorkspaceScope{}, fmt.Errorf("invalid runtime project scope for %q", project.Slug)
		}
		name := FPMContainerName(project.Version)
		mountPath := "/workspace/" + project.Slug
		if project.Dedicated {
			name = DedicatedContainerName(project.Slug)
			mountPath = "/workspace"
		}
		index, exists := containerIndex[name]
		if !exists {
			index = len(containers)
			containerIndex[name] = index
			containers = append(containers, PHPContainerScope{
				Scope: Scope{Workspace: workspace, Project: project.Path, Version: project.Version, Dedicated: project.Dedicated},
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
		routes = append(routes, NginxRoute{ServerName: strings.Join(project.Domains, " "), DocumentRoot: documentRoot, Upstream: name, SSL: project.SSL, CertName: project.CertName})
	}
	return WorkspaceScope{Workspace: workspace, ConfigPath: configPath, CertificateDir: certificateDir, HTTPPort: httpPort, HTTPSPort: httpsPort, Roots: roots, PHPContainers: containers, Routes: routes}, nil
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
		if route.DocumentRoot == "" || !filepath.IsAbs(route.DocumentRoot) || route.ServerName == "" || route.Upstream == "" {
			return "", fmt.Errorf("invalid nginx route")
		}
		if route.SSL && (httpsPort == 0 || route.CertName == "") {
			return "", fmt.Errorf("SSL route %s is missing HTTPS port or certificate", route.ServerName)
		}
		listen := fmt.Sprintf("    listen %d;\n", httpPort)
		ssl := ""
		if route.SSL {
			listen += fmt.Sprintf("    listen %d ssl;\n", httpsPort)
			ssl = fmt.Sprintf("    ssl_certificate /etc/nginx/certs/%s.crt;\n    ssl_certificate_key /etc/nginx/certs/%s.key;\n", route.CertName, route.CertName)
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
        fastcgi_param SCRIPT_FILENAME %s$fastcgi_script_name;
        fastcgi_pass %s:9000;
    }
}
`, listen, route.ServerName, route.DocumentRoot, ssl, route.DocumentRoot, route.Upstream)
	}
	return b.String(), nil
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
