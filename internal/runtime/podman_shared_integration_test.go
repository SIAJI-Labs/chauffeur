package runtime

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPodmanSharedPHPThroughNginxMultipleProjects(t *testing.T) {
	if os.Getenv("CHAUFFEUR_PODMAN_INTEGRATION") != "1" {
		t.Skip("set CHAUFFEUR_PODMAN_INTEGRATION=1 to run Podman integration tests")
	}
	ctx := context.Background()
	runner := ExecRunner{}
	for _, image := range []string{PHP83Image, NginxImage} {
		result, err := runner.Run(ctx, "image", "exists", image)
		if err != nil || result.ExitCode != 0 {
			t.Skipf("required image unavailable: %s", image)
		}
	}
	root := t.TempDir()
	one := filepath.Join(root, "one")
	two := filepath.Join(root, "two")
	if err := os.MkdirAll(one, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(two, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(one, "index.php"), []byte("<?php echo 'one';"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(two, "index.php"), []byte("<?php echo 'two';"), 0644); err != nil {
		t.Fatal(err)
	}
	config, err := RenderNginxPHPConfigForRoutes([]NginxRoute{
		{ServerName: "one.test", DocumentRoot: "/workspace/one", Upstream: FPMContainerName("8.3")},
		{ServerName: "two.test", DocumentRoot: "/workspace/two", Upstream: FPMContainerName("8.3")},
	}, 8080)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	podman := Podman{Runner: runner}
	if err := podman.EnsurePHPContainerWithRoots(ctx, Scope{Version: "8.3"}, PHP83Image, map[string]string{"/workspace/one": one, "/workspace/two": two}); err != nil {
		t.Fatal(err)
	}
	if err := podman.EnsureNginxContainerWithRoots(ctx, configPath, map[string]string{"/workspace/one": one, "/workspace/two": two}, 18082); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx")
		_, _ = runner.Run(ctx, "container", "rm", "-f", FPMContainerName("8.3"))
	})
	client := &http.Client{Timeout: 5 * time.Second}
	for _, tc := range []struct{ host, want string }{{"one.test", "one"}, {"two.test", "two"}} {
		var response *http.Response
		var err error
		for attempt := 0; attempt < 20; attempt++ {
			request, requestErr := http.NewRequest(http.MethodGet, "http://127.0.0.1:18082/", nil)
			if requestErr != nil {
				t.Fatal(requestErr)
			}
			request.Host = tc.host
			response, err = client.Do(request)
			if err == nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil || string(body) != tc.want {
			t.Fatalf("host %s body = %q, err = %v", tc.host, body, readErr)
		}
	}
}
