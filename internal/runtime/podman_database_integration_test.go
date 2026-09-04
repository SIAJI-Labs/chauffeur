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

func TestPodmanPHPReachesPublishedDatabaseWithLoopbackProjectConfig(t *testing.T) {
	if os.Getenv("CHAUFFEUR_PODMAN_INTEGRATION") != "1" {
		t.Skip("set CHAUFFEUR_PODMAN_INTEGRATION=1 to run Podman integration tests")
	}
	ctx := context.Background()
	runner := ExecRunner{}
	for _, image := range []string{PHP83Image, "docker.io/library/redis:7-alpine"} {
		result, err := runner.Run(ctx, "image", "exists", image)
		if err != nil || result.ExitCode != 0 {
			t.Skipf("required image unavailable: %s", image)
		}
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("DB_HOST=localhost\nDB_PORT=16379\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.php"), []byte("<?php $s=@fsockopen(getenv('DB_HOST'),(int)getenv('DB_PORT'),$e,$m,2); echo $s ? 'db-ok' : 'db-failed';"), 0644); err != nil {
		t.Fatal(err)
	}
	redisName := "chauf-test-db-connectivity"
	_, _ = runner.Run(ctx, "container", "rm", "-f", redisName)
	result, err := runner.Run(ctx, "run", "-d", "--name", redisName, "--network", "chauf-net", "--publish", "16379:6379", "docker.io/library/redis:7-alpine")
	if err != nil || result.ExitCode != 0 {
		t.Fatalf("start database fixture: result=%#v err=%v", result, err)
	}
	t.Cleanup(func() { _, _ = runner.Run(ctx, "container", "rm", "-f", redisName) })

	rt := Podman{Runner: runner}
	if err := rt.EnsurePHPContainer(ctx, Scope{Version: "8.3", Project: root}, PHP83Image, root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = runner.Run(ctx, "container", "rm", "-f", FPMContainerName("8.3")) })

	result, err = rt.Exec(ctx, Scope{Version: "8.3", Project: root}, []string{"php", "-r", "for($i=0;$i<20;$i++){ $s=@fsockopen(getenv('DB_HOST'),(int)getenv('DB_PORT'),$e,$m,1); if($s){fclose($s);exit(0);} usleep(250000); } exit(1);"}, ExecOptions{
		Env: []string{"DB_HOST=host.containers.internal", "DB_PORT=16379"},
	})
	if err != nil || result.ExitCode != 0 {
		t.Fatalf("PHP could not reach published database port: result=%#v err=%v", result, err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "DB_HOST=localhost\nDB_PORT=16379\n" {
		t.Fatalf("project .env was modified: %q", content)
	}

	config, err := RenderNginxPHPConfigForRoutes([]NginxRoute{{ServerName: "db.test", DocumentRoot: "/workspace", Upstream: FPMContainerName("8.3"), DatabaseHost: "host.containers.internal", DatabasePort: 16379}}, 8080)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	if err := rt.EnsureNginxContainerWithRoots(ctx, configPath, map[string]string{"/workspace": root}, 18086); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx") })
	client := http.Client{Timeout: 5 * time.Second}
	for attempt := 0; attempt < 20; attempt++ {
		request, requestErr := http.NewRequest(http.MethodGet, "http://127.0.0.1:18086/", nil)
		if requestErr == nil {
			request.Host = "db.test"
			response, requestErr := client.Do(request)
			if requestErr == nil {
				body, readErr := io.ReadAll(response.Body)
				response.Body.Close()
				if readErr != nil {
					t.Fatal(readErr)
				}
				if string(body) != "db-ok" {
					t.Fatalf("web database check body = %q, status = %s", body, response.Status)
				}
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal("timed out waiting for web database connectivity fixture")
}
