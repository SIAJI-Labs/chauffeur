package runtime

import (
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
	if got := scope.Routes[1].DocumentRoot; got != "/workspace" {
		t.Fatalf("dedicated document root = %q", got)
	}
	if got := scope.Roots["/workspace/shop"]; got != filepath.Clean("/home/shop") {
		t.Fatalf("shared mount = %q", got)
	}
	if !strings.Contains(scope.Routes[0].ServerName, "www.shop.test") {
		t.Fatal("aliases were not retained in route")
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
	for _, want := range []string{"listen 8080", "listen 8443 ssl", "ssl_certificate /etc/nginx/certs/secure.test.crt", "ssl_certificate_key /etc/nginx/certs/secure.test.key"} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
}
