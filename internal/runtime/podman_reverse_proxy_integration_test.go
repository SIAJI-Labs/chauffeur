package runtime

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPodmanReverseProxyThroughNginx(t *testing.T) {
	if os.Getenv("CHAUFFEUR_PODMAN_INTEGRATION") != "1" {
		t.Skip("set CHAUFFEUR_PODMAN_INTEGRATION=1 to run Podman integration tests")
	}
	runner := ExecRunner{}
	ctx := context.Background()
	if result, err := runner.Run(ctx, "image", "exists", NginxImage); err != nil || result.ExitCode != 0 {
		t.Skip("nginx image unavailable")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "proxy-ok") })}
	go server.Serve(listener)
	defer server.Close()
	proxyPort := listener.Addr().(*net.TCPAddr).Port

	root := t.TempDir()
	config, err := RenderNginxPHPConfigForRoutes([]NginxRoute{{ServerName: "proxy.test", Upstream: "host.containers.internal", ProxyPort: proxyPort}}, 8080)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	nginxPort := freeTCPPort(t)
	podman := Podman{Runner: runner}
	if err := podman.EnsureNginxContainerWithRootsAndPorts(ctx, configPath, map[string]string{}, nginxPort, 0, ""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx") })

	client := &http.Client{Timeout: 5 * time.Second}
	var response *http.Response
	for attempt := 0; attempt < 20; attempt++ {
		response, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/", nginxPort))
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "proxy-ok" {
		t.Fatalf("response = %s %q", response.Status, body)
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}
