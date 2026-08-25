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

func TestPodmanPHPThroughNginx(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(root, "index.php"), []byte("<?php echo 'podman-ok';"), 0644); err != nil {
		t.Fatal(err)
	}
	config, err := RenderNginxPHPConfig("/workspace", "chauf-php83-fpm", 8080)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	podman := Podman{Runner: runner}
	scope := Scope{Version: "8.3", Project: root}
	if err := podman.EnsurePHPContainer(ctx, scope, PHP83Image, root); err != nil {
		t.Fatal(err)
	}
	if err := podman.EnsureNginxContainer(ctx, configPath, root, 18081); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx")
		_, _ = runner.Run(ctx, "container", "rm", "-f", FPMContainerName("8.3"))
	})

	client := &http.Client{Timeout: 5 * time.Second}
	var response *http.Response
	for attempt := 0; attempt < 20; attempt++ {
		response, err = client.Get("http://127.0.0.1:18081/")
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %s", response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "podman-ok" {
		t.Fatalf("body = %q, want podman-ok", body)
	}
}
