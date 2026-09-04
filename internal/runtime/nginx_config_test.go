package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildWorkspaceScopeResolvesSharedAndDedicatedResources(t *testing.T) {
	scope, err := BuildWorkspaceScope("/workspace", "/workspace/nginx/container.conf", "/workspace/nginx/certs", 18080, 18443, []ProjectSpec{
		{Slug: "shop", Path: "/home/shop", Version: "8.3", Domains: []string{"shop.test", "www.shop.test"}, SSL: true, CertName: "shop.test"},
		{Slug: "admin", Path: "/home/admin", Version: "8.3", Domains: []string{"admin.test"}, Dedicated: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(scope.PHPContainers) != 2 || len(scope.Routes) != 2 {
		t.Fatalf("containers=%d routes=%d", len(scope.PHPContainers), len(scope.Routes))
	}
	if got := ContainerName(scope.PHPContainers[0].Scope); got != FPMContainerName("8.3") {
		t.Fatalf("shared container = %q", got)
	}
	if got := ContainerName(scope.PHPContainers[1].Scope); got != DedicatedContainerName("admin") {
		t.Fatalf("dedicated container = %q", got)
	}
	if got := scope.Routes[0].DocumentRoot; got != "/workspace/shop" {
		t.Fatalf("shared document root = %q", got)
	}
	if got := scope.Routes[1].DocumentRoot; got != "/workspace/admin" {
		t.Fatalf("dedicated document root = %q", got)
	}
	if got := scope.Roots["/workspace/shop"]; got != filepath.Clean("/home/shop") {
		t.Fatalf("shared mount = %q", got)
	}
	if !strings.Contains(scope.Routes[0].ServerName, "www.shop.test") {
		t.Fatal("aliases were not retained in route")
	}
}

func TestBuildWorkspaceScopeUsesProjectDocumentRoot(t *testing.T) {
	scope, err := BuildWorkspaceScope("/workspace", "/workspace/nginx/container.conf", "/workspace/nginx/certs", 18080, 18443, []ProjectSpec{
		{Slug: "laravel", Path: "/home/laravel", DocumentRoot: "/home/laravel/public", Version: "8.3", Domains: []string{"laravel.test"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := scope.Routes[0].DocumentRoot; got != "/workspace/laravel/public" {
		t.Fatalf("document root = %q; want /workspace/laravel/public", got)
	}
}

func TestBuildWorkspaceScopeKeepsDedicatedProjectMountsDistinct(t *testing.T) {
	scope, err := BuildWorkspaceScope("/workspace", "/workspace/nginx.conf", "", 8080, 0, []ProjectSpec{
		{Slug: "admin", Path: "/projects/admin", Version: "8.3", Domains: []string{"admin.test"}, Dedicated: true},
		{Slug: "billing", Path: "/projects/billing", Version: "8.3", Domains: []string{"billing.test"}, Dedicated: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(scope.PHPContainers) != 2 || len(scope.Roots) != 2 {
		t.Fatalf("containers=%d roots=%d; want isolated projects", len(scope.PHPContainers), len(scope.Roots))
	}
	if scope.Routes[0].DocumentRoot == scope.Routes[1].DocumentRoot {
		t.Fatal("dedicated projects share a document root")
	}
	if scope.Roots["/workspace/admin"] != "/projects/admin" || scope.Roots["/workspace/billing"] != "/projects/billing" {
		t.Fatalf("roots = %#v", scope.Roots)
	}
}

func TestBuildWorkspaceScopeOverridesLoopbackDatabaseHostForPodman(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, ".env"), []byte("DB_HOST=127.0.0.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	scope, err := BuildWorkspaceScope("/workspace", "/workspace/nginx/container.conf", "/workspace/nginx/certs", 18080, 18443, []ProjectSpec{
		{Slug: "app", Path: project, Version: "8.3", Domains: []string{"app.test"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS(scope.Routes, 18080, 18443)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, "fastcgi_param DB_HOST host.containers.internal;") {
		t.Fatalf("config missing Podman DB host override:\n%s", config)
	}
}

func TestBuildWorkspaceScopeIncludesReverseProxyWithoutPHPContainer(t *testing.T) {
	scope, err := BuildWorkspaceScope("/workspace", "/workspace/nginx/container.conf", "/workspace/nginx/certs", 18080, 18443, []ProjectSpec{
		{Slug: "vite", Path: "/home/vite", ProjectType: "reverse-proxy", ProxyPort: 5173, Domains: []string{"vite.test"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(scope.PHPContainers) != 0 || len(scope.Routes) != 1 {
		t.Fatalf("containers=%d routes=%d", len(scope.PHPContainers), len(scope.Routes))
	}
	if scope.Routes[0].Upstream != "host.containers.internal" || scope.Routes[0].ProxyPort != 5173 {
		t.Fatalf("route = %+v", scope.Routes[0])
	}
}

func TestRenderNginxPHPConfigUsesContainerUpstream(t *testing.T) {
	config, err := RenderNginxPHPConfig("/workspace", "chauf-php83-fpm", 8080)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"listen 8080", "root /workspace", "fastcgi_pass chauf-php83-fpm:9000", "SCRIPT_FILENAME /workspace$fastcgi_script_name"} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
}

func TestRenderNginxPHPConfigForRoutesIncludesEveryProject(t *testing.T) {
	config, err := RenderNginxPHPConfigForRoutes([]NginxRoute{
		{ServerName: "one.test", DocumentRoot: "/workspace/one", Upstream: "chauf-php83-fpm"},
		{ServerName: "two.test", DocumentRoot: "/workspace/two", Upstream: "chauf-php83-fpm"},
	}, 8080)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"server_name one.test", "root /workspace/one", "server_name two.test", "root /workspace/two"} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
}

func TestRenderNginxPHPConfigForRoutesIncludesTLS(t *testing.T) {
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS([]NginxRoute{{
		ServerName:   "secure.test",
		DocumentRoot: "/workspace/secure",
		Upstream:     "chauf-php83-fpm",
		SSL:          true,
		CertName:     "secure.test",
	}}, 8080, 8443)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"listen 8080", "listen 8443 ssl", "return 301 https://$host:8443$request_uri", "ssl_certificate /etc/nginx/certs/secure.test.crt", "ssl_certificate_key /etc/nginx/certs/secure.test.key"} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
}

func TestRenderNginxConfigForReverseProxy(t *testing.T) {
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS([]NginxRoute{{
		ServerName: "vite.test", Upstream: "host.containers.internal", ProxyPort: 5173, SSL: true, CertName: "vite.test",
	}}, 8080, 8443)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"proxy_pass http://host.containers.internal:5173", "proxy_set_header Upgrade $http_upgrade", "return 301 https://$host:8443$request_uri"} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
}
