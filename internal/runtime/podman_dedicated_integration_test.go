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

func TestPodmanDedicatedPHPThroughNginx(t *testing.T) {
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
	containerName := DedicatedContainerName(filepath.Base(root))
	if err := os.Chmod(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.php"), []byte("<?php echo 'dedicated-ok';"), 0644); err != nil {
		t.Fatal(err)
	}
	config, err := RenderNginxPHPConfigForDomain("/workspace", "dedicated.test", containerName, 8080)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	podman := Podman{Runner: runner}
	scope := Scope{Version: "8.3", Project: root, Dedicated: true}
	if err := podman.EnsurePHPContainerWithRoots(ctx, scope, PHP83Image, map[string]string{"/workspace": root}); err != nil {
		t.Fatal(err)
	}
	if err := podman.EnsureNginxContainerWithRoots(ctx, configPath, map[string]string{"/workspace": root}, 18083); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx")
		_, _ = runner.Run(ctx, "container", "rm", "-f", containerName)
	})
	client := &http.Client{Timeout: 5 * time.Second}
	var response *http.Response
	for attempt := 0; attempt < 20; attempt++ {
		request, requestErr := http.NewRequest(http.MethodGet, "http://127.0.0.1:18083/", nil)
		if requestErr != nil {
			t.Fatal(requestErr)
		}
		request.Host = "dedicated.test"
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
	if response.StatusCode != http.StatusOK || readErr != nil || string(body) != "dedicated-ok" {
		t.Fatalf("status = %s, body = %q, err = %v", response.Status, body, readErr)
	}
}
